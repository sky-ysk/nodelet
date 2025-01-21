# 部署使用Apiserver

## 编译apiserver

```bash
#cd到项目根目录下
make all WHAT=./cmd/apiserver FRAMEWORK_VERBOSE=3
```

## 部署apiserver

通过上一步生成的二进制文件apiserver来部署启动apiserver。

通过添加命令行参数来配置apiserver，目前支持的参数：

```bash
      --bind-address ip                 API Server 监听的IP地址 (default 0.0.0.0)
      --bind-port int                   API Server 监听的端口号 (default 10000)
      --delete-collection-workers int   为 DeleteCollection 调用生成的工作线程数。用于加速清理资源集合。 (default 3)
      --enable-garbage-collector        启用泛型垃圾回收器 (default true)
      --etcd-cafile string              用于保护 etcd 通信的 SSL 证书颁发机构文件。
      --etcd-certfile string            用于保护 etcd 通信的 SSL 证书文件。
      --etcd-keyfile string             用于保护 etcd 通信的 SSL 密钥文件。
      --etcd-prefix string              在 etcd 中所有资源路径前附加的前缀 (default "registry/")
      --etcd-servers strings            要连接的 etcd 服务器列表 (格式:ip:port,ip:port), 逗号分隔
      --storage-media-type string       用于在存储中存储对象的媒体类型 支持的媒体类型: [application/json] (default "application/json")
```

示例：

```bash
#本地安装etcd服务，127.0.0.1:2379为etcd服务器地址
./apiserver --etcd-servers=127.0.0.1:2379
```

## 使用apiserver

apiserver通过restful服务来暴露对资源的操作接口

1. 使用client-go作为客户端与apiserver进行交互操作资源
2. 使用restful请求来与apiserver交互。接口文档可在apiserver启动后通过http://ip:port/apidocs.json来获取(ip:port为apiserver部署地址)





# 扩展定制资源

在apiserver中扩展定制资源并装载对应api步骤：

1. 定义资源结构体并实现runtime.Object 接口
2. 将资源注册到注册表中
3. 编写资源的RESTStorage
4. 装载RESTStorage



以添加Sample类型的资源为例

1. 在pkg/apis/cores包下定义资源结构体

   定义资源结构体Sample和SampleList，Sample需包含meta.TypeMeta和meta.ObjectMeta，SampleList需包含meta.TypeMeta和meta.ListMeta，并实现runtime.Object中的DeepCopyObject() 接口

   ```go
   type Sample struct {
   	meta.TypeMeta
   	meta.ObjectMeta
   
   	Spec   SampleSpec   `json:"spec,omitempty" yaml:"spec"`
   	Status SampleStatus `json:"status,omitempty" yaml:"status"`
   }
   
   type SampleSpec struct {
   	SampleName string `json:"sample_name,omitempty" yaml:"name"`
   }
   
   type SampleStatus struct {
   	SampleID string `json:"sample_id,omitempty" yaml:"sample_id"`
   
   	SampleTime Time `json:"sample_time,omitempty" yaml:"sample_time"`
   }
   
   type SampleList struct {
   	meta.TypeMeta
   	meta.ListMeta
   	Items []Sample
   }
   
   func (in *Sample) DeepCopyObject() runtime.Object {
   	if c := in.DeepCopy(); c != nil {
   		return c
   	}
   	return nil
   }
   func (in *Sample) DeepCopy() *Sample {
   	if in == nil {
   		return nil
   	}
   	out := new(Sample)
   	in.DeepCopyInto(out)
   	return out
   }
   func (in *Sample) DeepCopyInto(out *Sample) {
   	*out = *in
   	out.TypeMeta = in.TypeMeta
   	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
   	in.Spec.DeepCopyInto(&out.Spec)
   	in.Status.DeepCopyInto(&out.Status)
   }
   
   func (in *SampleList) DeepCopyInto(out *SampleList) {
   	*out = *in
   	out.TypeMeta = in.TypeMeta
   	in.ListMeta.DeepCopyInto(&out.ListMeta)
   	if in.Items != nil {
   		in, out := &in.Items, &out.Items
   		*out = make([]Sample, len(*in))
   		for i := range *in {
   			(*in)[i].DeepCopyInto(&(*out)[i])
   		}
   	}
   }
   
   func (in *SampleList) DeepCopy() *SampleList {
   	if in == nil {
   		return nil
   	}
   	out := new(SampleList)
   	in.DeepCopyInto(out)
   	return out
   }
   
   func (in *SampleList) DeepCopyObject() runtime.Object {
   	if c := in.DeepCopy(); c != nil {
   		return c
   	}
   	return nil
   }
   
   func (in *SampleSpec) DeepCopyInto(out *SampleSpec) {
   	*out = *in
   }
   func (in *SampleStatus) DeepCopyInto(out *SampleStatus) {
   	*out = *in
   	in.SampleTime.DeepCopyInto(&out.SampleTime)
   }
   ```

2. 将资源注册到注册表中

   在pkg/apis/cores/registry.go的addKnownTypes中，将Sample和SampleList加入到注册表中

   ```go
   func addKnownTypes(scheme *runtime.Scheme) error {
   	...
       scheme.AddKnownTypes(SchemeGroupVersion, &Sample{}, &SampleList{})
       ...
   	return nil
   }
   ```

3. 编写Sample资源的RESTStorage

   仿照pkg/registry/core中各资源的写法，编写Sample资源的RESTStorage

   pkg/registry/core/sample/storage.go:

   ```go
   import (
   	"context"
   	"fmt"
   	"hit.edu/framework/pkg/apimachinery/fields"
   	"hit.edu/framework/pkg/apimachinery/labels"
   	"hit.edu/framework/pkg/apimachinery/runtime"
   	apis "hit.edu/framework/pkg/apis/cores"
   	"hit.edu/framework/pkg/apis/meta"
   	"hit.edu/framework/pkg/apiserver/registry/generic"
   	genericregistry "hit.edu/framework/pkg/apiserver/registry/generic/registry"
   	"hit.edu/framework/pkg/apiserver/registry/rest"
   	"hit.edu/framework/pkg/apiserver/registry/storage"
   )
   
   type REST struct {
   	*genericregistry.Store
   }
   
   type StatusREST struct {
   	*genericregistry.Store
   }
   type SampleStorage struct {
   	Sample *REST
   	Status *StatusREST
   	Spec   *SpecREST
   }
   type SpecREST struct {
   	*genericregistry.Store
   }
   
   func (r *SpecREST) New() runtime.Object {
   	return &apis.Sample{}
   }
   func (r *SpecREST) Destroy() {
   }
   func (r *SpecREST) Get(ctx context.Context, name string, options *meta.GetOptions) (runtime.Object, error) {
   	return r.Store.Get(ctx, name, options)
   }
   func (r *SpecREST) Update(ctx context.Context, name string, objInfo rest.UpdatedObjectInfo, createValidation rest.ValidateObjectFunc, updateValidation rest.ValidateObjectUpdateFunc, forceAllowCreate bool, options *meta.UpdateOptions) (runtime.Object, bool, error) {
   	return r.Store.Update(ctx, name, objInfo, createValidation, updateValidation, false, options)
   }
   
   // 对节点状态进行操作
   func (r *StatusREST) New() runtime.Object {
   	return &apis.Sample{}
   }
   func (r *StatusREST) Destroy() {
   }
   func (r *StatusREST) Get(ctx context.Context, name string, options *meta.GetOptions) (runtime.Object, error) {
   	return r.Store.Get(ctx, name, options)
   }
   func (r *StatusREST) Update(ctx context.Context, name string, objInfo rest.UpdatedObjectInfo, createValidation rest.ValidateObjectFunc, updateValidation rest.ValidateObjectUpdateFunc, forceAllowCreate bool, options *meta.UpdateOptions) (runtime.Object, bool, error) {
   	return r.Store.Update(ctx, name, objInfo, createValidation, updateValidation, false, options)
   }
   
   func NewFunc() runtime.Object {
   	return &apis.Sample{}
   }
   
   func NewListFunc() runtime.Object {
   	return &apis.SampleList{}
   }
   
   // GetAttrs 从传入的资源对象中提取标签和字段
   func GetAttrs(obj runtime.Object) (labels.Set, fields.Set, error) {
   	Sample, ok := obj.(*apis.Sample)
   	if !ok {
   		return nil, nil, fmt.Errorf("not a Sample")
   	}
   	return labels.Set(Sample.ObjectMeta.Labels), generic.ObjectMetaFieldsSet(&Sample.ObjectMeta, true), nil
   }
   func Match(label labels.Selector, field fields.Selector) storage.SelectionPredicate {
   	return storage.SelectionPredicate{
   		Label:    label,
   		Field:    field,
   		GetAttrs: GetAttrs,
   	}
   }
   
   func NewSampleStorage(optsGetter generic.RESTOptionsGetter) (SampleStorage, error) {
   	store := &genericregistry.Store{
   		NewFunc:                   NewFunc,
   		NewListFunc:               NewListFunc,
   		PredicateFunc:             Match,
   		DefaultQualifiedResource:  apis.Resource("samples"),
   		SingularQualifiedResource: apis.Resource("sample"),
   
   		CreateStrategy: thisStrategy,
   		UpdateStrategy: thisStrategy,
   		DeleteStrategy: thisStrategy,
   		//Storage:
   	}
   	options := &generic.StoreOptions{
   		RESTOptions: optsGetter,
   		AttrFunc:    GetAttrs,
   	}
   	if err := store.CompleteWithOptions(options); err != nil {
   		return SampleStorage{}, err
   	}
   	statusStore := *store
   	statusStore.UpdateStrategy = thisStrategy
   
   	SampleREST := &REST{Store: store}
   	statusREST := &StatusREST{Store: &statusStore}
   
   	specStore := *store
   	specStore.UpdateStrategy = thisStrategy
   	specREST := &SpecREST{Store: &specStore}
   	return SampleStorage{
   		Sample: SampleREST,
   		Status: statusREST,
   		Spec:   specREST,
   	}, nil
   }
   ```

   pkg/registry/core/sample/strategy.go:

   ```go
   import (
   	"context"
   	"hit.edu/framework/pkg/apimachinery/runtime"
   	"hit.edu/framework/pkg/apis/legacyscheme"
   	"hit.edu/framework/pkg/apiserver/registry/storage/field"
   )
   
   type Strategy struct {
   	runtime.ObjectTyper
   }
   
   var thisStrategy = &Strategy{legacyscheme.Scheme}
   
   func (t Strategy) PrepareForCreate(ctx context.Context, obj runtime.Object)      {}
   func (t Strategy) PrepareForUpdate(ctx context.Context, obj, old runtime.Object) {}
   func (t Strategy) Validate(ctx context.Context, obj runtime.Object) field.ErrorList {
   	return nil
   }
   func (t Strategy) ValidateUpdate(ctx context.Context, obj, old runtime.Object) field.ErrorList {
   	return nil
   }
   func (t Strategy) Canonicalize(obj runtime.Object) {}
   //在这里指定资源是否必须带有命名空间
   func (t Strategy) NamespaceScoped() bool {
   	return true
   }
   ```

4. 装载RESTStorage

   在pkg/registry/core/rest/storage_core.go中装载SampleRESTStorage

   ```go
   func NewRESTStorage(restOptionsGetter generic.RESTOptionsGetter) (server.APIGroupInfo, error) {
   	...
       //装载SampleRESTStorage
   	sampleStorage, err := samplestore.NewSampleStorage(restOptionsGetter)
   	if err != nil {
   		logs.Error("error occur while create ActionStorage", err)
   		return server.APIGroupInfo{}, err
   	}
   	if resource := "samples"; true {
   		storage[resource] = sampleStorage.Sample
   		storage[resource+"/status"] = sampleStorage.Status
   		storage[resource+"/spec"] = sampleStorage.Spec
   	}
           
   	if len(storage) > 0 {
       ...
   }
   ```



装载SampleRESTStorage后，apiserver便已注册好Sample资源的restful api,编译启动后可通过访问服务器的/apidocs.json路径来查看sample资源的restful接口文档





# **FieldSelector**

fieldSelector根据字段筛选资源，目前支持所有资源通用字段name,namespace

支持=，==，!=判断筛选（= 和 == 意义相同），示例：

```bash
fieldSelector=metadata.name=task-1
fieldSelector=metadata.name==task-2
fieldSelector=metadata.namespace!=default

不同条件之间通过逗号隔开进行链式选择：
fieldSelector=metadata.name=task-1,metadata.namespace!=default

请求示例：
#筛选name为task-1且不在default命名空间下的task资源
http://localhost:10000/apis/resources/v1/tasks?fieldSelector=metadata.name=task-1,metadata.namespace!=default
```





# **LabelSelector**

labelSelector根据用户定义的标签筛选资源

支持基于等值的判断筛选（和fieldSelector相同）和基于集合的判断筛选

labelSelector语法规则：

```bash
<selector-syntax>         ::= <requirement> | <requirement> "," <selector-syntax>
<requirement>             ::= [!] KEY [ <set-based-restriction> | <exact-match-restriction> ]
<set-based-restriction>   ::= "" | <inclusion-exclusion> <value-set>
<inclusion-exclusion>     ::= <inclusion> | <exclusion>
<exclusion>               ::= "notin"
<inclusion>               ::= "in"
<value-set>               ::= "(" <values> ")"
<values>                  ::= VALUE | VALUE "," <values>
<exact-match-restriction> ::= ["="|"=="|"!="] VALUE

示例：
#筛选labels标签中,app标签值等于app1或app2,foo标签值等于bar,存在x标签,y标签值不等于y1且不等于y2的资源
"app in (app1,app2),foo==bar,x,y notin (y1,y2)"

#筛选不存在app标签且foo标签值不等于bar的资源
"!app,foo!=bar" 

请求示例：
http://localhost:10000/apis/resources/v1/tasks?labelSelector=app in (app1,app2),foo==bar,x,y notin (y1,y2)
```

fieldSelector和labelSelector可以结合使用：

```bash
http://localhost:10000/apis/resources/v1/tasks?fieldSelector=metadata.name=my-service,metadata.namespace!=default&labelSelector=app in (app1,app2),foo==bar,x,y notin (y1,y2)
```





# 定制资源字段选择器FieldSelector

所有资源都支持 `metadata.name` 和 `metadata.namespace` 字段选择算符

如果想要根据定制资源的其他字段来筛选定制资源，可通过以下两个步骤完成：

1. 在资源registry包下添加函数设置可筛选字段

2. 在GetAttrs函数中，将ObjectMetaFieldsSet函数替换为1中添加的函数


示例：通过TaskID来筛选Task资源

```go
//1. 在pkg/apiserver/registry/core/task包下定义函数，将需要设置的请求的字符串status.task_id和Task的TaskID字段绑定起来
func ToSelectableFields(task *apis.Task) fields.Set{
    objectMetaFieldsSet := generic.ObjectMetaFieldsSet(&task.ObjectMeta,true)
    specificFieldsSet := fields.Set{
        //在这里定义需要设置的请求的字符串status.task_id
        "status.task_id" : task.Status.TaskID,
    }
    return generic.MergeFieldsSets(objectMetaFieldsSet,specificFieldsSet)
}

//2. 在pkg/apiserver/registry/core/task/storage.go中，将ObjectMetaFieldsSet替换为实现的ToSelectableFields
// GetAttrs 从传入的资源对象中提取标签和字段
func GetAttrs(obj runtime.Object) (labels.Set, fields.Set, error) {
	Task, ok := obj.(*apis.Task)
	if !ok {
		return nil, nil, fmt.Errorf("not a Task")
	}
   	//将原有的ObjectMetaFieldsSet替换为实现的ToSelectableFields
	return labels.Set(Task.ObjectMeta.Labels), ToSelectableFields(Task), nil
}

//3. 在请求中的fieldSelector中通过设置status.task_id来通过TaskID来筛选Task资源
http://localhost:10000/apis/resources/v1/tasks?fieldSelector=status.task_id=={TaskID}
```


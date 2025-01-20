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
3. 编写Sample资源的RESTStorage
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



# 定制资源字段选择器fieldSelector

所有定制资源都支持 `metadata.name` 和 `metadata.namespace` 字段选择算符。如果想要根据定制资源的其他字段来筛选定制资源，可通过以下两个步骤完成：

1. 在pkg/registry/sample包下添加ToSelectableFields函数来设置可筛选字段

   以sample.Status.SampleID为例：

   ```go
   func ToSelectableFields(sample *apis.Sample) fields.Set {
   	objectMetaFieldsSet := generic.ObjectMetaFieldsSet(&sample.ObjectMeta, true)
   	specificFieldsSet := fields.Set{
   		"status.sample_id": sample.Status.SampleID,
   	}
   	return generic.MergeFieldsSets(objectMetaFieldsSet, specificFieldsSet)
   }
   ```

2. 在pkg/registry/sample/storage.go的GetAttrs函数中，将ObjectMetaFieldsSet函数替换为ToSelectableFields

   ```go
   func GetAttrs(obj runtime.Object) (labels.Set, fields.Set, error) {
   	Sample, ok := obj.(*apis.Sample)
   	if !ok {
   		return nil, nil, fmt.Errorf("not a Sample")
   	}
   	return labels.Set(Sample.ObjectMeta.Labels), ToSelectableFields(Sample), nil
   }
   ```

   
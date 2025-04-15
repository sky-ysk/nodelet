### 资源接口的使用

对于etcd的底层调用以接口的形式封装在了结构体store（pkg/apiserver/registry/storage/etcd3）中，通过该结构体可以最终调用etcd的Get、Put等操作。在这个结构体中，保存了etcd客户端的信息，序列化工具等内容，用于实现对etcd存储、数据转换、资源版本等内容的管理。

另外一个Store（pkg/apiserver/registry/generic/registry/store.go），在store的基础上，实现了与存储相关的更多细节操作，以及更加详细的信息，比如对象在etcd中的键、筛选函数、验证函数等，Storage是Store与etcd的核心接口，并且创建了一个store实例，Storage通过这个实例来实现与etcd的实际交互。

而各个资源的最终实现是通过在（pkg/apiserver/registry/core）路径下分别存放的各个文件夹内。在这里为不同的资源创建了Store对象，同时可以为资源的子资源也创建Store对象，比如资源的spec和status。在设置资源的相关信息和操作策略之后，完成Store的创建，即可完成该资源的Storage的创建，并通过这个Storage实现对资源的各项操作。

在调用各个资源的Storage创建之后，即可将这些资源绑定到对应的路径上，返回从路径名到具体的资源的Storage的映射。

目前可以使用的资源包括以下：Node、Workflow、Action、Task、Group、Event、Device、Scene、Resource、Data。

### 深拷贝方法脚本的使用

在新增或者修改了资源之后，到deepcopy.go文件中找到并且修改对应的深拷贝方法很麻烦，目前提供了一个脚本，存放在目录pkg/apis/cores/deepcopy_generate_test下，通过调用k8s的deepcopy_gen工具实现对资源深拷贝方法的自动生成，具体的步骤如下：

1.将所有需要用到的资源，包括资源调用的其余资源，存放在该文件夹中的types.go文件中。

2.根据自动生成工具的要求在对应的结构体前面补充注释，要求如下：

（1）在package apis前面补充注释

// +k8s:deepcopy-gen=package

（2）在需要访问的资源，如Node、NodeList等前面，需要补充注释
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

（3）部分情况下会出现如下问题：引用某些包比如time时，自动生成工具会将该结构体当作自行定义的
针对此问题，目前的解决方案是，通过注释
// +k8s:deepcopy-gen=false
取消存在此问题的结构体的自动生成，并在add_deepcopy.go文件中补充该结构体的深拷贝方法，之后运行单独补充生成。

（4）通过命令运行脚本，生成的深拷贝方法存放在deepcopy.go文件中
./create_deepcopy.sh
（5）将生成的文件替换掉外面的deepcopy.go文件即可

### 添加资源的步骤

（1）在types.go中补充需要使用的资源信息，同时生成他的深拷贝方法。

（2）在register.go文件中将该资源和该资源的列表形态添加到注册列表中。

（3）在pkg/apiserver/registry/core文件夹下新增该资源的文件夹，补充他的Storage创建过程，并根据实际需要补充他的子资源的访问方式，同时补充他的创建策略、验证方式、是否是命名空间级别资源等内容。

（4）到storage_core.go文件中，将该资源和该资源的子资源补充到对应的路由下。

### 修改资源之后要补充的修改

（1）如果仅仅对某项资源新增了string、time等可以直接拷贝的字段，那么无需进行任何修改，只需要对apiserver进行重新的编译运行即可。

（2）如果在某个资源中新增了结构体等无法自动拷贝的字段，那么需要修改deepcopy方法。

（3）如果要新增资源类型，请参考添加资源的步骤。
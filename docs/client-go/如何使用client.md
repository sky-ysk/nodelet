# client-go使用

## 使用client进行增删查改

### 以pkg/client-go/examples/create-update-delete/create-update-delete-node为例:

1、首先需要配置clientSet的参数并创建clientSet，然后使用clientSet创建nodeClient并填入NameSpace，之后可以使用nodesClient的方法进行create、get、update、patch、delete、watch等。

2、create：先创建一个node,并对需要的字段赋值，包括ObjectMeta.Name等。

3、update：先get一个node，再进行update。

4、patch：先创建一个json格式的patchNode，使用patch方法进行补丁更新。

5、list、watch：先创建一个ListOptions，使用FieldSelector或LabelSelector进行筛选。FieldSelector是根据资源某一个字段的值进行筛选；LabelSelector是根据资源的ObjectMeta.Labels进行筛选。

## workqueue使用

### 以pkg/client-go/examples/workqueue/workqueue-node为例：

先配置好options，包括所需资源对应的ListerWatcher，然后创建Informer与indexer并创建controller，最后执行go controller.Run(workers, stopCh)即可开始监测事件变化、增量式添加到workqueue中，并执行业务逻辑处理。例子controller的业务逻辑处理为简单的打印函数syncToStdout，业务逻辑处理方法在main.go中定义。


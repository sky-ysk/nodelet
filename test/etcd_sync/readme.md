### etcd_sync使用方式

假设目前有域A broker需要往域B pve2发送请求

一、前提

broker和pve2之间已经部署了跨域组件

二、在pve2上执行

1.在pve2上启动apiserver，并记录服务暴露的ip和端口，也可以直接使用默认的0.0.0.0:10000

2.进入test/etcd_sync目录，通过make build编译得到可执行文件并存放在/test/etcd_sync/bin/syncserver

3.执行可执行文件，同时输入apiserver的ip和端口，以及希望sync服务暴露的ip和端口

```
./bin/syncserver --apiserver-address=0.0.0.0 --apiserver-port=10000 --bind-address=0.0.0.0 --bind-port=14399
```

这里的四个参数：

apiserver-address和apiserver-port为apiserver的ip和端口号，默认使用0.0.0.0和10000（需要自行启动）

bind-address和bind-port为跨域同步服务的ip和端口号，默认使用0.0.0.0和14399

4.至此pve2上已经启动了跨域同步的服务。

三、在broker上编写代码

示例如etcd_sync/informer/test文件夹下的测试文件，具体使用时

0.通过curl命令首先确保跨域服务正常运行

```shell
curl "http://broker.registry-svc.test.svc.clusterset.local:3001/forward?target=http://172.110.0.120:14399/healthz" -H "FlowType:etcd" -H "ClusterID:pve2"
```

若返回OK说明跨域服务正常运行

1.通过CreateTarget函数创建Target对象，其中需要指定资源的类型，本域ID、对应域ID、跨域服务IP、跨域服务端口号、同步服务运行的ip和端口号

```golang
target := informer.CreateTarget[*apis.Node]("broker", "pve2", "registry-svc.test.svc.clusterset.local", 3001, "172.110.0.120", 14399)
```

2.其余使用过程与域内命令基本一致

```golang
obj, err := target.Create(context.TODO(), node, metav1.CreateOptions{})
obj, err = target.Get(context.TODO(), "Test", "demo-nodes", "nodes", metav1.GetOptions{})
```



注：

1.目前仅支持基本的Create、Get、Update、Delete，其余函数暂不支持

2.目前的版本路径统一使用apis/resource/v1，暂时不支持其余输入

3.这部分代码后面还会调整，不过使用方式尽量不会修改，避免之后更新的代码合并之后出错
**### etcd_sync使用方式**

假设目前有域A broker需要往域B pve2发送请求

一、前提

broker和pve2之间已经部署了跨域组件

二、在pve2上执行

1.在pve2上启动apiserver，并记录服务暴露的ip和端口，也可以直接使用默认的0.0.0.0:10000

三、在broker上编写代码

1.示例如etcd_sync/sync/main.go中的测试文件，具体使用时与域内的etcd通信过程基本一致，只是需要在ContentConfig中注明FlowType和ClusterID，ClusterID为要转发到的域的ID，同时在Host处应该填写本域的跨域转发服务的IP，并将目的域的apiserver的IP写在TargetURL中。需要注意，这三个只能在使用跨域时使用，不进行跨域通信时不要填充这三个值。

```golang
c := &rest.Config{
		//Host: "http://broker.registry-svc.test.svc.clusterset.local:3001/forward?target=",
		//Host:    "http://localhost:10000",
		Host:    "http://broker.registry-svc.test.svc.clusterset.local:3001",
		APIPath: "/apis/resources/v1",
		ContentConfig: rest.ContentConfig{
			AcceptContentTypes: "application/json; charset=UTF-8", //text/plain; charset=UTF-8
			ContentType:        "application/json; charset=UTF-8", //application/json; charset=UTF-8
			GroupVersion: &schema.GroupVersion{
				Group:   "resources",
				Version: "v1",
			},
			NegotiatedSerializer: serializer.NewCodecFactory(scheme),
			TargetURL:            "http://172.110.0.120:10000",
			FlowType:             "etcd",
			ClusterID:            "pve2",
		},
		UserAgent: "defaultUserAgent",
		Transport: &http.Transport{
			MaxIdleConns:        10000,            // 最大空闲连接数
			IdleConnTimeout:     90 * time.Second, // 空闲连接超时时间
			TLSHandshakeTimeout: 10 * time.Second, // TLS 握手超时时间
		},
		Timeout: 1000 * time.Second,
	}
```

同时在创建ClientSet这一步中，引用的包应该是""hit.edu/framework/test/etcd_sync/active/clients""

```golang
clientSet, err := clients.NewForConfig(c)
```

2.通过这种方式创建的Client能够使用原本的api完成各项操作，包括Create、Update、Get、Delete、Patch、List，目前Watch暂时不支持
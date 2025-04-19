package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"hit.edu/framework/pkg/apimachinery/runtime"
	"hit.edu/framework/pkg/apimachinery/runtime/schema"
	"hit.edu/framework/pkg/apimachinery/runtime/serializer"
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/client-go/clients"
	"hit.edu/framework/pkg/client-go/clients/typed/core"
	"hit.edu/framework/pkg/client-go/rest"
	"hit.edu/framework/pkg/component-base/logs"
)

// 创建一个Rest Client
// 验证xxx动词
// 与API Server通信，并执行基础操作

func main() {
	scheme := runtime.NewScheme()
	apis.AddToScheme(scheme)
	fmt.Println(scheme)
	//参数配置
	// TODO: 填写参数
	//部分参数之后可以在core_client等 编写setConfigDefaults函数进行填充
	c := &rest.Config{
		Host:    "http://localhost:10000",
		APIPath: "/apis/resources/v1",
		ContentConfig: rest.ContentConfig{
			AcceptContentTypes: "application/json; charset=UTF-8", //text/plain; charset=UTF-8
			ContentType:        "application/json; charset=UTF-8", //application/json; charset=UTF-8
			GroupVersion: &schema.GroupVersion{
				Group:   "resources",
				Version: "v1",
			},
			NegotiatedSerializer: serializer.NewCodecFactory(scheme),
		},
		UserAgent: "defaultUserAgent",
		Transport: &http.Transport{
			MaxIdleConns:        100,              // 最大空闲连接数
			IdleConnTimeout:     90 * time.Second, // 空闲连接超时时间
			TLSHandshakeTimeout: 10 * time.Second, // TLS 握手超时时间
		},
		Timeout: 10 * time.Second,
	}

	//创建ClientSet
	clientSet, err := clients.NewForConfig(c)
	if err != nil {
		panic(err)
	}
	// 资源定义在 pkg/apis/xxx/type.go 下
	// 这里以访问资源Group为例，
	// 获取访问Group的客户端
	// 默认访问的Namespace是 ""

	eventsClient := clientSet.Core().Events("test")

	// 先删除冗余事件
	clearEvents(eventsClient)

	// Create一个event
	logs.Info("event creating")
	for _, item := range table {
		result, err := eventsClient.Create(context.TODO(), &item, meta.CreateOptions{})
		if err != nil {
			logs.Errorf("Failed to create event: %v", err)
			panic(err)
		}
		logs.Info("Created event : ", result.Name)
	}
}

func clearEvents(client core.EventInterface) {
	logs.Info("event deleting")
	for _, item := range table {
		err := client.Delete(context.TODO(), item.Name, meta.DeleteOptions{})
		if err != nil {
			logs.Errorf("Failed to delete event: %v", err)
			// panic(err)
		}
	}

}

var table = []apis.Event{
	{
		ObjectMeta: meta.ObjectMeta{
			Name:      "test-event",
			Namespace: "test",
		},
		TypeMeta: meta.TypeMeta{
			Kind:       "Event",
			APIVersion: "resources/v1",
		},
		InvolvedObject: apis.ObjectReference{
			Kind:       "Node",
			Name:       "CloudNode1",
			Namespace:  "test",
			UID:        "default",
			APIVersion: "resources/v1",
		},
		Reason:  "Created",
		Message: "Successfully created cloud node 1",
		Source:  apis.EventSource{Component: "test-controller", Host: "127.0.0.1"},
		Count:   1,
		Type:    apis.EventTypeNormal,
	},
}

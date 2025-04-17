package main

import (
	"bufio"
	"context"
	"fmt"
	"hit.edu/framework/pkg/apimachinery/runtime"
	"hit.edu/framework/pkg/apimachinery/runtime/schema"
	"hit.edu/framework/pkg/apimachinery/runtime/serializer"
	apis "hit.edu/framework/pkg/apis/cores"
	metav1 "hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/client-go/clients"
	"hit.edu/framework/pkg/client-go/rest"
	"hit.edu/framework/pkg/component-base/logs"
	"net/http"
	"os"
	"time"
)

// 创建一个Rest Client
// 验证xxx动词
// 与API Server通信，并执行基础操作

func main() {
	logs.Init("create-update-delete-Event")
	scheme := runtime.NewScheme()
	apis.AddToScheme(scheme)
	//fmt.Println(scheme)
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
	// 这里以访问资源Event为例，
	// 获取访问Event的客户端
	// 默认访问的Namespace是 ""

	eventsClient := clientSet.Core().Events("test")

	event1 := &apis.Event{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "demo-events",
			Namespace: "test",
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Event",
			APIVersion: "resources/v1",
		},
	}
	//event2 := &apis.Event{
	//	ObjectMeta: metav1.ObjectMeta{
	//		Name: "demo-event2",
	//	},
	//	TypeMeta: metav1.TypeMeta{
	//		Kind:       "Event",
	//		APIVersion: "resources/v1",
	//	},
	//	Spec: apis.EventSpec{
	//		Name: "demo-event",
	//	},
	//}
	//event3 := &apis.Event{
	//	ObjectMeta: metav1.ObjectMeta{
	//		Name: "demo-event3",
	//	},
	//	TypeMeta: metav1.TypeMeta{
	//		Kind:       "Event",
	//		APIVersion: "resources/v1",
	//	},
	//	Spec: apis.EventSpec{
	//		Name: "demo-event",
	//	},
	//}

	// Create一个Event
	fmt.Println("creating")
	results, err := eventsClient.Create(context.TODO(), event1, metav1.CreateOptions{})

	if err != nil {
		panic(err)
	}
	//_, _ = eventsClient.Create(context.TODO(), event2, metav1.CreateOptions{})
	//_, _ = eventsClient.Create(context.TODO(), event3, metav1.CreateOptions{})
	fmt.Println("Created event ", results)

	prompt()

	// Delete一个Event
	fmt.Println("deleting")
	err = eventsClient.Delete(context.TODO(), "demo-events", metav1.DeleteOptions{})
	if err != nil {
		panic(err)
	}
	fmt.Println("Deleted event...")
	prompt()

	// Delete 之后再次 List所有Event
	fmt.Println("listing")
	lstOpts := metav1.ListOptions{}
	list, err := eventsClient.List(context.TODO(), lstOpts)
	if err != nil {
		panic(err)
	}
	for _, d := range list.Items {
		fmt.Println(d)
	}
	fmt.Println("listing done")

	//DeleteCollection 删除所有Spec.EventName=demo-event的Event
	fmt.Println("deleting collection")
	lstOpts = metav1.ListOptions{
		FieldSelector: "Spec.Name=demo-event",
	}
	err = eventsClient.DeleteCollection(context.TODO(), metav1.DeleteOptions{}, lstOpts)
	if err != nil {
		panic(err)
	}
	fmt.Println("Deleted collection...")

	// DeleteCollection 之后再次 List所有Event
	fmt.Println("listing")
	lstOpts = metav1.ListOptions{}
	list, err = eventsClient.List(context.TODO(), lstOpts)
	if err != nil {
		panic(err)
	}
	for _, d := range list.Items {
		fmt.Println(d)
	}
	fmt.Println("listing done")
}

// From K8s
func prompt() {
	fmt.Printf("-> Press Return key to continue.")
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		break
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}
	fmt.Println()
}

package main

import (
	"context"
	"net/http"
	"time"

	"hit.edu/framework/pkg/apimachinery/runtime"
	"hit.edu/framework/pkg/apimachinery/runtime/schema"
	"hit.edu/framework/pkg/apimachinery/runtime/serializer"
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/client-go/clients"
	"hit.edu/framework/pkg/client-go/rest"
	"hit.edu/framework/pkg/client-go/util/manager"
	"hit.edu/framework/pkg/component-base/logs"
)

var scheme = runtime.NewScheme()

func main() {
	logs.Init("testClear")
	m := manager.NewManager(initClientSet(scheme))
	list, _ := m.GetTasks(apis.NamespaceTest)
	for _, item := range list.Items {
		m.DeleteTask(item.Name, item.Namespace)
	}

	list3, _ := m.GetGroups(apis.NamespaceTest)
	for _, item := range list3.Items {
		m.DeleteGroup(item.Name, item.Namespace)
	}

	list2, _ := m.GetActions(apis.NamespaceTest)
	for _, item := range list2.Items {
		logs.Info(item.Name)
		m.DeleteAction(item.Name, item.Namespace)
	}

	list4, _ := m.GetRuntimes(apis.NamespaceTest)
	for _, item := range list4.Items {
		m.DeleteRuntime(item.Name, item.Namespace)
	}

	list5, _ := m.GetEventClient("test").Client.List(context.TODO(), meta.ListOptions{})
	logs.Infof("event counts: %v", len(list5.Items))
	m.GetEventClient("test").Client.DeleteCollection(context.TODO(), meta.DeleteOptions{}, meta.ListOptions{})

}

func initClientSet(scheme *runtime.Scheme) *clients.ClientSet {
	apis.AddToScheme(scheme)
	logs.Info(scheme)
	// 创建ClientSet
	c := &rest.Config{
		Host:    "http://127.0.0.1:10000",
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
		Timeout: 100 * time.Second,
	}
	clientSet, err := clients.NewForConfig(c)
	if err != nil {
		panic(err)
	}
	return clientSet
}

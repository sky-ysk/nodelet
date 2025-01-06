package main

import (
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
	"time"
)

// 创建一个Rest Client
// 验证xxx动词
// 与API Server通信，并执行基础操作

func main() {
	logs.Init("client-go-examples")
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
			MaxIdleConns:        1000,             // 最大空闲连接数
			IdleConnTimeout:     90 * time.Second, // 空闲连接超时时间
			TLSHandshakeTimeout: 10 * time.Second, // TLS 握手超时时间
		},
		Timeout: 1000 * time.Second,
	}

	//创建ClientSet
	clientSet, err := clients.NewForConfig(c)
	if err != nil {
		logs.Error(err)
	}
	// 资源定义在 pkg/apis/xxx/type.go 下
	// 这里以访问资源Node为例，
	// 获取访问Node的客户端

	nodesClient := clientSet.Core().Nodes("test")
	groupsClient := clientSet.Core().Groups("test")
	workflowsClient := clientSet.Core().Workflows("test")

	node := &apis.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "demo-nodes",
			Namespace: apis.NamespaceAll,
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Node",
			APIVersion: "resources/v1",
		},
		Spec: apis.NodeSpec{
			NodeName: "demo-node",
		},
	}
	group := &apis.Group{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "demo-groups",
			Namespace: "test",
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Group",
			APIVersion: "resources/v1",
		},
		Spec: apis.GroupSpec{
			Name: "demo-group",
		},
	}
	workflow := &apis.Workflow{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "demo-workflows",
			Namespace: "test",
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Workflow",
			APIVersion: "resources/v1",
		},
	}

	// Create一个Node
	fmt.Println("creating")
	_, err = nodesClient.Create(context.TODO(), node, metav1.CreateOptions{})
	_, err = groupsClient.Create(context.TODO(), group, metav1.CreateOptions{})
	_, err = workflowsClient.Create(context.TODO(), workflow, metav1.CreateOptions{})

	result, getErr := nodesClient.Get(context.TODO(), "demo-nodes", metav1.GetOptions{})
	if getErr != nil {
		panic(fmt.Errorf("Failed to get : %v", getErr))
	}
	fmt.Println("get result", result)
	result2, getErr := groupsClient.Get(context.TODO(), "demo-groups", metav1.GetOptions{})
	if getErr != nil {
		panic(fmt.Errorf("Failed to get : %v", getErr))
	}
	fmt.Println("get result", result2)
	result3, getErr := workflowsClient.Get(context.TODO(), "demo-workflows", metav1.GetOptions{})
	if getErr != nil {
		panic(fmt.Errorf("Failed to get : %v", getErr))
	}
	fmt.Println("get result", result3)
}

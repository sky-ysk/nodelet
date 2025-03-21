package main

import (
	"context"
	"net/http"
	"time"

	"hit.edu/framework/pkg/apimachinery/runtime"
	"hit.edu/framework/pkg/apimachinery/runtime/schema"
	"hit.edu/framework/pkg/apimachinery/runtime/serializer"
	apis "hit.edu/framework/pkg/apis/cores"
	metav1 "hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/client-go/clients"
	"hit.edu/framework/pkg/client-go/rest"
	"hit.edu/framework/pkg/component-base/logs"
)

// 测试了一下client-go的selector功能，包括了labelSelector的各种用法，以及labelSelector和fieldSelector一起使用的用法。
func main() {
	logs.Init("main")
	scheme := runtime.NewScheme()
	apis.AddToScheme(scheme)
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
			MaxIdleConns:        10000,            // 最大空闲连接数
			IdleConnTimeout:     90 * time.Second, // 空闲连接超时时间
			TLSHandshakeTimeout: 10 * time.Second, // TLS 握手超时时间
		},
		Timeout: 1000 * time.Second,
	}

	//创建ClientSet
	clientSet, err := clients.NewForConfig(c)
	if err != nil {
		panic(err)
	}
	// 资源定义在 pkg/apis/xxx/type.go 下
	// 这里以访问资源Node为例，
	// 获取访问Node的客户端
	// 默认访问的Namespace是 ""

	nodesClient := clientSet.Core().Nodes("test")

	node := &apis.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "demo-nodes",
			Namespace: "test",
			Labels: map[string]string{
				"environment": "dev",
				"app":         "app1",
			},
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Node",
			APIVersion: "resources/v1",
		},
		Spec: apis.NodeSpec{
			NodeName: "demo-node",
		},
	}
	node2 := &apis.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "demo-node2",
			Namespace: "test",
			Labels: map[string]string{
				"environment": "dev",
				"app":         "app1",
			},
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Node",
			APIVersion: "resources/v1",
		},
		Spec: apis.NodeSpec{
			NodeName: "demo-node",
		},
	}
	node3 := &apis.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "demo-node3",
			Namespace: "test",
			Labels: map[string]string{
				"environment": "qa",
				"app":         "app2",
			},
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Node",
			APIVersion: "resources/v1",
		},
		Spec: apis.NodeSpec{
			NodeName: "demo-node",
		},
	}

	// Create三个Node
	logs.Trace("creating")
	result, err := nodesClient.Create(context.TODO(), node, metav1.CreateOptions{})
	if err != nil {
		logs.Errorf("Failed to create node: %v", err)
	} else {
		logs.Trace("created node", result)
	}
	_, err = nodesClient.Create(context.TODO(), node2, metav1.CreateOptions{})
	_, err = nodesClient.Create(context.TODO(), node3, metav1.CreateOptions{})

	// List 所有Node
	logs.Trace("listing 所有node")
	lstOpts := metav1.ListOptions{}
	list, err := nodesClient.List(context.TODO(), lstOpts)
	if err != nil {
		logs.Error(err)
	}
	for _, d := range list.Items {
		logs.Trace(d)
	}

	// List 筛选的Node

	logs.Trace("listing 有app=app1标签的node")
	lstOpts = metav1.ListOptions{
		LabelSelector: "app=app1",
	}
	list, err = nodesClient.List(context.TODO(), lstOpts)
	if err != nil {
		logs.Error(err)
	}
	for _, d := range list.Items {
		logs.Trace(d)
	}

	logs.Trace("listing done")

	logs.Trace("listing 有app=app1 或 app3标签的node")
	lstOpts = metav1.ListOptions{
		LabelSelector: "app in (app1,app3)",
	}
	list, err = nodesClient.List(context.TODO(), lstOpts)
	if err != nil {
		logs.Error(err)
	}
	for _, d := range list.Items {
		logs.Trace(d)
	}
	logs.Trace("listing done")

	logs.Trace("listing 有app!=app1 或 app3标签的node")
	lstOpts = metav1.ListOptions{
		LabelSelector: "app notin (app1,app3)",
	}
	list, err = nodesClient.List(context.TODO(), lstOpts)
	if err != nil {
		logs.Error(err)
	}
	for _, d := range list.Items {
		logs.Trace(d)
	}
	logs.Trace("listing done")

	logs.Trace("listing 有app!=app1标签的node")
	lstOpts = metav1.ListOptions{
		LabelSelector: "app!=app1",
	}
	list, err = nodesClient.List(context.TODO(), lstOpts)
	if err != nil {
		logs.Error(err)
	}
	for _, d := range list.Items {
		logs.Trace(d)
	}
	logs.Trace("listing done")

	logs.Trace("listing 有environment标签的node")
	lstOpts = metav1.ListOptions{
		LabelSelector: "environment",
	}
	list, err = nodesClient.List(context.TODO(), lstOpts)
	if err != nil {
		logs.Error(err)
	}
	for _, d := range list.Items {
		logs.Trace(d)
	}
	logs.Trace("listing done")

	logs.Trace("listing 没有environment标签的node")
	lstOpts = metav1.ListOptions{
		LabelSelector: "!environment",
	}
	list, err = nodesClient.List(context.TODO(), lstOpts)
	if err != nil {
		logs.Error(err)
	}
	for _, d := range list.Items {
		logs.Trace(d)
	}
	logs.Trace("listing done")

	//测试fieldSelector

	logs.Trace("listing 标签的node")
	lstOpts = metav1.ListOptions{
		FieldSelector: "metadata.name=demo-nodes",
		LabelSelector: "environment=dev",
	}
	list, err = nodesClient.List(context.TODO(), lstOpts)
	if err != nil {
		logs.Error(err)
	}
	for _, d := range list.Items {
		logs.Trace(d)
	}
	logs.Trace("listing done")
}

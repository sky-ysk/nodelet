package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"hit.edu/framework/pkg/apimachinery/runtime"
	"hit.edu/framework/pkg/apimachinery/runtime/schema"
	"hit.edu/framework/pkg/apimachinery/runtime/serializer"
	"hit.edu/framework/pkg/apimachinery/types"
	apis "hit.edu/framework/pkg/apis/cores"
	metav1 "hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/test/etcd_sync/active/clients"

	//"hit.edu/framework/pkg/client-go/clients"
	"hit.edu/framework/pkg/client-go/rest"
	"hit.edu/framework/pkg/component-base/logs"
)

// 用于测试header添加部分

func main() {
	logs.Init("main")
	scheme := runtime.NewScheme()
	apis.AddToScheme(scheme)
	//参数配置
	// TODO: 填写参数
	//部分参数之后可以在core_client等 编写setConfigDefaults函数进行填充
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

	patchNode, err := json.Marshal(map[string]interface{}{
		"objectMeta": map[string]interface{}{
			"namespace": "test",
		},
		"spec": map[string]interface{}{
			"node_name": "patch-node-name",
			"host_name": "master",
		},
	})

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

	//Update一个Node

	logs.Trace("updating")
	// 部分更改一个参数
	// 先Get一个Node ,更改Node的参数, UpdateNode

	result, getErr := nodesClient.Get(context.TODO(), "demo-nodes", metav1.GetOptions{})
	if getErr != nil {
		logs.Info(fmt.Errorf("Failed to get : %v", getErr))
	}

	fmt.Println("get result", result)
	fmt.Println("修改前的result.Spec.NodeName：", result.Spec.NodeName)

	result.Spec.NodeName = "updatedNodeName"
	result, updateErr := nodesClient.Update(context.TODO(), result, metav1.UpdateOptions{})
	if updateErr != nil {
		logs.Error(fmt.Errorf("Update failed: %v", updateErr))
	}

	fmt.Println("修改后的result.Spec.NodeName：", result.Spec.NodeName)
	fmt.Println("Updated node...")

	// List 所有Node
	fmt.Println("listing 筛选的node")
	lstOpts := metav1.ListOptions{
		FieldSelector: "name=demo-node",
	}
	list, err := nodesClient.List(context.TODO(), lstOpts)
	if err != nil {
		logs.Error(err)
	}
	for _, d := range list.Items {
		logs.Trace(d)
	}

	logs.Trace("listing done")

	//Patch 一个Node
	logs.Trace("patching")
	patchResult, err := nodesClient.Patch(context.TODO(), "demo-nodes", types.StrategicMergePatchType, patchNode, metav1.PatchOptions{})
	logs.Trace("patchResult: ", patchResult)
	logs.Trace("patch Done")

	// List 所有Node
	logs.Trace("listing")
	lstOpts = metav1.ListOptions{}
	list, err = nodesClient.List(context.TODO(), lstOpts)
	if err != nil {
		logs.Error(err)
	}
	for _, d := range list.Items {
		logs.Trace(d)
	}

	logs.Trace("listing done")
	// Delete一个Node
	logs.Trace("deleting")
	err = nodesClient.Delete(context.TODO(), "demo-nodes", metav1.DeleteOptions{})
	if err != nil {
		panic(err)
	}
	logs.Trace("Deleted node...")

	// Delete 之后再次 List所有Node
	logs.Trace("listing")
	lstOpts = metav1.ListOptions{}
	list, err = nodesClient.List(context.TODO(), lstOpts)
	if err != nil {
		logs.Error(err)
	}
	for _, d := range list.Items {
		logs.Trace(d)
	}

	fmt.Println("listing done")

	//DeleteCollection 删除所有Spec.NodeName=demo-node的Node
	logs.Info("deleting collection")
	lstOpts = metav1.ListOptions{
		FieldSelector: "Spec.NodeName=demo-node",
	}
	err = nodesClient.DeleteCollection(context.TODO(), metav1.DeleteOptions{}, lstOpts)
	if err != nil {
		logs.Error(err)
	}
	logs.Info("Deleted collection...")

	// DeleteCollection 之后再次 List所有Node
	logs.Info("listing")
	lstOpts = metav1.ListOptions{}
	list, err = nodesClient.List(context.TODO(), lstOpts)
	if err != nil {
		logs.Error(err)
	}
	for _, d := range list.Items {
		logs.Info(d)
	}
	logs.Info("listing done")

	//DeleteCollection 删除所有Spec.NodeName=demo-node的Node
	fmt.Println("deleting collection")
	lstOpts = metav1.ListOptions{
		FieldSelector: "Spec.NodeName=demo-node",
	}
	err = nodesClient.DeleteCollection(context.TODO(), metav1.DeleteOptions{}, lstOpts)
	if err != nil {
		logs.Error(err)
	}
	fmt.Println("Deleted collection...")

	// DeleteCollection 之后再次 List所有Node
	logs.Trace("listing")
	lstOpts = metav1.ListOptions{}
	list, err = nodesClient.List(context.TODO(), lstOpts)
	if err != nil {
		logs.Error(err)
	}
	for _, d := range list.Items {
		logs.Trace(d)
	}
	logs.Trace("listing done")
}

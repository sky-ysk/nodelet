package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"hit.edu/framework/pkg/apimachinery/runtime"
	"hit.edu/framework/pkg/apimachinery/runtime/schema"
	"hit.edu/framework/pkg/apimachinery/runtime/serializer"
	"hit.edu/framework/pkg/apimachinery/types"
	"hit.edu/framework/pkg/apimachinery/watch"
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
	// 这里以访问资源Node为例，
	// 获取访问Node的客户端
	// 默认访问的Namespace是 ""

	nodesClient := clientSet.Core().Nodes("Test")

	node := &apis.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "demo-nodes",
			Namespace: "Test",
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Node",
			APIVersion: "resources/v1",
		},
		Spec: apis.NodeSpec{
			NodeName: "demo-node",
		},
	}

	//node2 := &apis.Node{
	//	ObjectMeta: metav1.ObjectMeta{
	//		Name: "demo-node2",
	//	},
	//	TypeMeta: runtime.TypeMeta{
	//		Kind:       "Node",
	//		APIVersion: "resources/v1",
	//	},
	//	Spec: apis.NodeSpec{
	//		NodeName: "demo-node",
	//	},
	//}
	//node3 := &apis.Node{
	//	ObjectMeta: metav1.ObjectMeta{
	//		Name: "demo-node3",
	//	},
	//	TypeMeta: runtime.TypeMeta{
	//		Kind:       "Node",
	//		APIVersion: "resources/v1",
	//	},
	//	Spec: apis.NodeSpec{
	//		NodeName: "demo-node",
	//	},
	//}
	patchNode, err := json.Marshal(map[string]interface{}{
		"Spec": map[string]interface{}{
			"NodeName": "patch-node-name",
			"HostName": "master",
		},
	})

	//监听事件并打印  监听resources/v1/nodes
	go func() {
		fmt.Println("watching")
		watchOptions := metav1.ListOptions{}

		watcher, err := nodesClient.Watch(context.TODO(), watchOptions)
		if err != nil {
			panic(err)
		}
		defer watcher.Stop() // 确保 watcher 被停止

		// 获取事件通道
		watchChan := watcher.ResultChan()

		for {
			select {
			case event, ok := <-watchChan:
				if !ok {
					fmt.Println("watchChan closed")
					return
				}

				// 打印事件类型和对象的相关信息
				fmt.Printf("接收到事件类型: %v\n", event.Type)
				switch event.Type {
				case watch.Added:
					fmt.Println("资源被添加: ", event.Object)
				case watch.Modified:
					fmt.Println("资源被修改: ", event.Object)
				case watch.Deleted:
					fmt.Println("资源被删除: ", event.Object)
				case watch.Error:
					fmt.Println("发生错误: ", event.Object)
				default:
					fmt.Println("未识别的事件类型: ", event.Type)
				}
			}
		}
	}()

	//如果已经存在，先删掉
	//err = nodesClient.Delete(context.TODO(), "demo-nodes", metav1.DeleteOptions{})

	// Create一个Node
	fmt.Println("creating")
	results, err := nodesClient.Create(context.TODO(), node, metav1.CreateOptions{})

	if err != nil {
		logs.Errorf("Failed to create node: %v", err)
		panic(err)
	}
	//_, _ = nodesClient.Create(context.TODO(), node2, metav1.CreateOptions{})
	//_, _ = nodesClient.Create(context.TODO(), node3, metav1.CreateOptions{})
	fmt.Println("Created node ", results)

	prompt()

	//Update一个Node

	fmt.Println("updating")
	// 部分更改一个参数
	// 先Get一个Node ,更改Node的参数, UpdateNode

	result, getErr := nodesClient.Get(context.TODO(), "demo-nodes", metav1.GetOptions{})
	if getErr != nil {
		panic(fmt.Errorf("Failed to get : %v", getErr))
	}

	fmt.Println("get result", result)
	fmt.Println("修改前的result.Spec.NodeName：", result.Spec.NodeName)

	result.Spec.NodeName = "updatedNodeName"
	_, updateErr := nodesClient.Update(context.TODO(), result, metav1.UpdateOptions{})
	if updateErr != nil {
		panic(fmt.Errorf("Update failed: %v", updateErr))
	}

	fmt.Println("修改后的result.Spec.NodeName：", result.Spec.NodeName)
	fmt.Println("Updated node...")
	prompt()

	// List 所有Node
	fmt.Println("listing")
	lstOpts := metav1.ListOptions{}
	list, err := nodesClient.List(context.TODO(), lstOpts)
	if err != nil {
		panic(err)
	}
	for _, d := range list.Items {
		fmt.Println(d)
	}

	fmt.Println("listing done")
	prompt()

	//Patch 一个Node
	fmt.Println("patching")
	patchResult, err := nodesClient.Patch(context.TODO(), "demo-nodes", types.StrategicMergePatchType, patchNode, metav1.PatchOptions{})
	fmt.Println("patchResult: ", patchResult)
	fmt.Println("patch Done")

	// List 所有Node
	fmt.Println("listing")
	lstOpts = metav1.ListOptions{}
	list, err = nodesClient.List(context.TODO(), lstOpts)
	if err != nil {
		panic(err)
	}
	for _, d := range list.Items {
		fmt.Println(d)
	}

	fmt.Println("listing done")
	prompt()

	// Delete一个Node
	fmt.Println("deleting")
	err = nodesClient.Delete(context.TODO(), "demo-nodes", metav1.DeleteOptions{})
	if err != nil {
		panic(err)
	}
	fmt.Println("Deleted node...")
	prompt()

	// Delete 之后再次 List所有Node
	fmt.Println("listing")
	lstOpts = metav1.ListOptions{}
	list, err = nodesClient.List(context.TODO(), lstOpts)
	if err != nil {
		panic(err)
	}
	for _, d := range list.Items {
		fmt.Println(d)
	}

	fmt.Println("listing done")

	select {}

	////DeleteCollection 删除所有Spec.NodeName=demo-node的Node
	//fmt.Println("deleting collection")
	//lstOpts = metav1.ListOptions{
	//	FieldSelector: "Spec.NodeName=demo-node",
	//}
	//err = nodesClient.DeleteCollection(context.TODO(), metav1.DeleteOptions{}, lstOpts)
	//if err != nil {
	//	panic(err)
	//}
	//fmt.Println("Deleted collection...")
	//prompt()
	//
	//// DeleteCollection 之后再次 List所有Node
	//fmt.Println("listing")
	//lstOpts = metav1.ListOptions{}
	//list, err = nodesClient.List(context.TODO(), lstOpts)
	//if err != nil {
	//	panic(err)
	//}
	//for _, d := range list.Items {
	//	fmt.Println(d)
	//}
	//fmt.Println("listing done")
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

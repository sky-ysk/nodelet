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
	// 这里以访问资源Resource_Node为例，
	// 获取访问Resource_Node的客户端
	// 默认访问的Namespace是 ""

	resource_NodesClient := clientSet.Core().Resource_Nodes("test")

	resource_Node := &apis.Resource_Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "demo-resource_Nodes",
			Namespace: "test",
			Labels: map[string]string{
				"environment": "dev",
			},
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Resource_Node",
			APIVersion: "resources/v1",
		},
		Spec: apis.ResourceSpec{
			Name: "demo-resource_Node",
		},
	}
	resource_Node2 := &apis.Resource_Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "demo-resource_Node2",
			Namespace: "test",
			Labels: map[string]string{
				"environment": "dev",
			},
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Resource_Node",
			APIVersion: "resources/v1",
		},
		Spec: apis.ResourceSpec{
			Name: "demo-resource_Node",
		},
	}
	resource_Node3 := &apis.Resource_Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "demo-resource_Node3",
			Namespace: "test",
			Labels: map[string]string{
				"environment": "qa",
			},
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Resource_Node",
			APIVersion: "resources/v1",
		},
		Spec: apis.ResourceSpec{
			Name: "demo-resource_Node",
		},
	}

	patchResource_Node, err := json.Marshal(map[string]interface{}{
		"objectMeta": map[string]interface{}{
			"namespace": "test",
		},
		"spec": map[string]interface{}{
			"resource_NodeName": "patch-resource_Node-name",
			"hostName":          "master",
		},
	})

	//监听事件并打印  监听resources/v1/resource_Nodes
	go func() {
		logs.Trace("watching")
		var timeoutSeconds int64 = 20
		watchOptions := metav1.ListOptions{
			TimeoutSeconds: &timeoutSeconds,
		}

		watcher, err := resource_NodesClient.Watch(context.TODO(), watchOptions)
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
					logs.Tracef("watchChan closed")
					return
				}

				// 打印事件类型和对象的相关信息
				logs.Tracef("接收到事件类型:", event.Type)
				switch event.Type {
				case watch.Added:
					logs.Tracef("资源被添加: ", event.Object)
				case watch.Modified:
					logs.Tracef("资源被修改: ", event.Object)
				case watch.Deleted:
					logs.Tracef("资源被删除: ", event.Object)
				case watch.Error:
					logs.Tracef("发生错误: ", event.Object)
				case watch.Bookmark:
					logs.Tracef("收到Bookmark", event.Object)

				default:
					logs.Tracef("未识别的事件类型: ", event.Type)
				}
			}
		}
	}()

	// Create三个Resource_Node
	logs.Trace("creating")
	result, err := resource_NodesClient.Create(context.TODO(), resource_Node, metav1.CreateOptions{})
	if err != nil {
		logs.Errorf("Failed to create resource_Node: %v", err)
	} else {
		logs.Trace("created resource_Node", result)
	}
	_, err = resource_NodesClient.Create(context.TODO(), resource_Node2, metav1.CreateOptions{})
	_, err = resource_NodesClient.Create(context.TODO(), resource_Node3, metav1.CreateOptions{})
	prompt()

	//Update一个Resource_Node

	logs.Trace("updating")
	// 部分更改一个参数
	// 先Get一个Resource_Node ,更改Resource_Node的参数, UpdateResource_Node

	result, getErr := resource_NodesClient.Get(context.TODO(), "demo-resource_Nodes", metav1.GetOptions{})
	if getErr != nil {
		logs.Info(fmt.Errorf("Failed to get : %v", getErr))
	}

	logs.Tracef("get result", result)
	logs.Tracef("修改前的result.Spec.Resource_NodeName：", result.Spec.Name)

	result.Spec.Name = "updatedResource_NodeName"
	_, updateErr := resource_NodesClient.Update(context.TODO(), result, metav1.UpdateOptions{})
	if updateErr != nil {
		logs.Error(fmt.Errorf("Update failed: %v", updateErr))
	}

	logs.Tracef("修改后的result.Spec.Resource_NodeName：", result.Spec.Name)
	logs.Tracef("Updated resource_Node...")
	prompt()

	// List 所有Resource_Node
	logs.Tracef("listing 筛选的resource_Node")
	lstOpts := metav1.ListOptions{
		LabelSelector: "environment",
	}
	list, err := resource_NodesClient.List(context.TODO(), lstOpts)
	if err != nil {
		logs.Error(err)
	}
	for _, d := range list.Items {
		logs.Trace(d)
	}

	logs.Trace("listing done")
	prompt()

	//Patch 一个Resource_Node
	logs.Trace("patching")
	patchResult, err := resource_NodesClient.Patch(context.TODO(), "demo-resource_Nodes", types.StrategicMergePatchType, patchResource_Node, metav1.PatchOptions{})
	logs.Trace("patchResult: ", patchResult)
	logs.Trace("patch Done")

	// List 所有Resource_Node
	logs.Trace("listing")
	lstOpts = metav1.ListOptions{}
	list, err = resource_NodesClient.List(context.TODO(), lstOpts)
	if err != nil {
		logs.Error(err)
	}
	for _, d := range list.Items {
		logs.Trace(d)
	}

	logs.Trace("listing done")
	prompt()

	// Delete一个Resource_Node
	logs.Trace("deleting")
	err = resource_NodesClient.Delete(context.TODO(), "demo-resource_Nodes", metav1.DeleteOptions{})
	if err != nil {
		panic(err)
	}
	logs.Trace("Deleted resource_Node...")
	prompt()

	// Delete 之后再次 List所有Resource_Node
	logs.Trace("listing")
	lstOpts = metav1.ListOptions{}
	list, err = resource_NodesClient.List(context.TODO(), lstOpts)
	if err != nil {
		logs.Error(err)
	}
	for _, d := range list.Items {
		logs.Trace(d)
	}

	logs.Tracef("listing done")

	//DeleteCollection 删除所有Spec.Resource_NodeName=demo-resource_Node的Resource_Node
	logs.Tracef("deleting collection")
	lstOpts = metav1.ListOptions{
		FieldSelector: "Spec.Resource_NodeName=demo-resource_Node",
	}
	err = resource_NodesClient.DeleteCollection(context.TODO(), metav1.DeleteOptions{}, lstOpts)
	if err != nil {
		logs.Error(err)
	}
	logs.Tracef("Deleted collection...")
	prompt()

	// DeleteCollection 之后再次 List所有Resource_Node
	logs.Trace("listing")
	lstOpts = metav1.ListOptions{}
	list, err = resource_NodesClient.List(context.TODO(), lstOpts)
	if err != nil {
		logs.Error(err)
	}
	for _, d := range list.Items {
		logs.Trace(d)
	}
	logs.Trace("listing done")

	select {}
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

package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

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
)

// 创建一个Rest Client
// 验证xxx动词
// 与API Server通信，并执行基础操作

func main() {
	logs.Init("main")
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
	// 这里以访问资源Action为例，
	// 获取访问Action的客户端
	// 默认访问的Namespace是 ""

	actionsClient := clientSet.Core().Actions("test")
	action := &apis.Action{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "demo-actions",
			Namespace: "test",
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Action",
			APIVersion: "resources/v1",
		},
		Spec: apis.ActionSpec{
			Name: "demo-action",
		},
	}

	//action2 := &apis.Action{
	//	ObjectMeta: metav1.ObjectMeta{
	//		Name: "demo-action2",
	//	},
	//	TypeMeta: runtime.TypeMeta{
	//		Kind:       "Action",
	//		APIVersion: "resources/v1",
	//	},
	//	Spec: apis.ActionSpec{
	//		ActionName: "demo-action",
	//	},
	//}
	//action3 := &apis.Action{
	//	ObjectMeta: metav1.ObjectMeta{
	//		Name: "demo-action3",
	//	},
	//	TypeMeta: runtime.TypeMeta{
	//		Kind:       "Action",
	//		APIVersion: "resources/v1",
	//	},
	//	Spec: apis.ActionSpec{
	//		ActionName: "demo-action",
	//	},
	//}
	patchAction, err := json.Marshal(map[string]interface{}{
		"Spec": map[string]interface{}{
			"ActionName": "patch-action-name",
			"HostName":   "master",
		},
	})

	//监听事件并打印  监听resources/v1/actions
	go func() {
		fmt.Println("watching")
		watchOptions := metav1.ListOptions{}

		watcher, err := actionsClient.Watch(context.TODO(), watchOptions)
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
	//err = actionsClient.Delete(context.TODO(), "demo-actions", metav1.DeleteOptions{})

	// Create一个Action
	fmt.Println("creating")
	results, err := actionsClient.Create(context.TODO(), action, metav1.CreateOptions{})

	if err != nil {
		logs.Errorf("Failed to create action: %v", err)
		panic(err)
	}
	//_, _ = actionsClient.Create(context.TODO(), action2, metav1.CreateOptions{})
	//_, _ = actionsClient.Create(context.TODO(), action3, metav1.CreateOptions{})
	fmt.Println("Created action ", results)

	prompt()

	//Update一个Action

	fmt.Println("updating")
	// 部分更改一个参数
	// 先Get一个Action ,更改Action的参数, UpdateAction

	result, getErr := actionsClient.Get(context.TODO(), "demo-actions", metav1.GetOptions{})
	if getErr != nil {
		panic(fmt.Errorf("Failed to get : %v", getErr))
	}

	fmt.Println("get result", result)
	fmt.Println("修改前的result.Spec.ActionName：", result.Spec.Name)

	result.Spec.Name = "updatedActionName"
	_, updateErr := actionsClient.Update(context.TODO(), result, metav1.UpdateOptions{})
	if updateErr != nil {
		panic(fmt.Errorf("Update failed: %v", updateErr))
	}

	fmt.Println("修改后的result.Spec.ActionName：", result.Spec.Name)
	fmt.Println("Updated action...")
	prompt()

	// List 所有Action
	fmt.Println("listing")
	lstOpts := metav1.ListOptions{}
	list, err := actionsClient.List(context.TODO(), lstOpts)
	if err != nil {
		panic(err)
	}
	for _, d := range list.Items {
		fmt.Println(d)
	}

	fmt.Println("listing done")
	prompt()

	//Patch 一个Action
	fmt.Println("patching")
	patchResult, err := actionsClient.Patch(context.TODO(), "demo-actions", types.StrategicMergePatchType, patchAction, metav1.PatchOptions{})
	fmt.Println("patchResult: ", patchResult)
	fmt.Println("patch Done")

	// List 所有Action
	fmt.Println("listing")
	lstOpts = metav1.ListOptions{}
	list, err = actionsClient.List(context.TODO(), lstOpts)
	if err != nil {
		panic(err)
	}
	for _, d := range list.Items {
		fmt.Println(d)
	}

	fmt.Println("listing done")
	prompt()

	// Delete一个Action
	fmt.Println("deleting")
	err = actionsClient.Delete(context.TODO(), "demo-actions", metav1.DeleteOptions{})
	if err != nil {
		panic(err)
	}
	fmt.Println("Deleted action...")
	prompt()

	// Delete 之后再次 List所有Action
	fmt.Println("listing")
	lstOpts = metav1.ListOptions{}
	list, err = actionsClient.List(context.TODO(), lstOpts)
	if err != nil {
		panic(err)
	}
	for _, d := range list.Items {
		fmt.Println(d)
	}

	fmt.Println("listing done")

	select {}

	////DeleteCollection 删除所有Spec.ActionName=demo-action的Action
	//fmt.Println("deleting collection")
	//lstOpts = metav1.ListOptions{
	//	FieldSelector: "Spec.ActionName=demo-action",
	//}
	//err = actionsClient.DeleteCollection(context.TODO(), metav1.DeleteOptions{}, lstOpts)
	//if err != nil {
	//	panic(err)
	//}
	//fmt.Println("Deleted collection...")
	//prompt()
	//
	//// DeleteCollection 之后再次 List所有Action
	//fmt.Println("listing")
	//lstOpts = metav1.ListOptions{}
	//list, err = actionsClient.List(context.TODO(), lstOpts)
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

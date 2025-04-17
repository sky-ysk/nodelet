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
	// 这里以访问资源Action为例，
	// 获取访问Action的客户端
	// 默认访问的Namespace是 ""

	actionsClient := clientSet.Core().Actions("test")

	action := &apis.Action{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "demo-actions",
			Namespace: "test",
			Labels: map[string]string{
				"environment": "dev",
			},
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Action",
			APIVersion: "resources/v1",
		},
		Spec: apis.ActionSpec{
			Name: "demo-action",
		},
	}
	action2 := &apis.Action{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "demo-action2",
			Namespace: "test",
			Labels: map[string]string{
				"environment": "dev",
			},
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Action",
			APIVersion: "resources/v1",
		},
		Spec: apis.ActionSpec{
			Name: "demo-action",
		},
	}
	action3 := &apis.Action{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "demo-action3",
			Namespace: "test",
			Labels: map[string]string{
				"environment": "qa",
			},
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Action",
			APIVersion: "resources/v1",
		},
		Spec: apis.ActionSpec{
			Name: "demo-action",
		},
	}

	patchAction, err := json.Marshal(map[string]interface{}{
		"objectMeta": map[string]interface{}{
			"namespace": "test",
		},
		"spec": map[string]interface{}{
			"name":     "patch-action-name",
			"hostName": "master",
		},
	})

	//监听事件并打印  监听resources/v1/actions
	go func() {
		logs.Info("watching")
		var timeoutSeconds int64 = 20
		watchOptions := metav1.ListOptions{
			TimeoutSeconds: &timeoutSeconds,
		}

		watcher, err := actionsClient.Watch(context.TODO(), watchOptions)
		if err != nil {
			logs.Error(err)
		}
		defer watcher.Stop() // 确保 watcher 被停止

		// 获取事件通道
		watchChan := watcher.ResultChan()

		for {
			select {
			case event, ok := <-watchChan:
				if !ok {
					logs.Info("watchChan closed")
					return
				}

				// 打印事件类型和对象的相关信息
				logs.Infof("接收到事件类型:", event.Type)
				switch event.Type {
				case watch.Added:
					logs.Infof("资源被添加: ", event.Object)
				case watch.Modified:
					logs.Infof("资源被修改: ", event.Object)
				case watch.Deleted:
					logs.Infof("资源被删除: ", event.Object)
				case watch.Error:
					logs.Infof("发生错误: ", event.Object)
				case watch.Bookmark:
					logs.Infof("收到Bookmark", event.Object)

				default:
					logs.Infof("未识别的事件类型: ", event.Type)
				}
			}
		}
	}()

	// Create三个Action
	logs.Trace("creating")
	result, err := actionsClient.Create(context.TODO(), action, metav1.CreateOptions{})
	if err != nil {
		logs.Errorf("Failed to create action: %v", err)
	} else {
		logs.Infof("created action", result)
	}
	_, err = actionsClient.Create(context.TODO(), action2, metav1.CreateOptions{})
	_, err = actionsClient.Create(context.TODO(), action3, metav1.CreateOptions{})
	prompt()

	//Update一个Action

	logs.Info("updating")
	// 部分更改一个参数
	// 先Get一个Action ,更改Action的参数, UpdateAction

	result, getErr := actionsClient.Get(context.TODO(), "demo-actions", metav1.GetOptions{})
	if getErr != nil {
		logs.Error(fmt.Errorf("Failed to get : %v", getErr))
	}

	logs.Infof("get result", result)
	logs.Infof("修改前的result.Spec.Name：", result.Spec.Name)

	result.Spec.Name = "updatedName"
	_, updateErr := actionsClient.Update(context.TODO(), result, metav1.UpdateOptions{})
	if updateErr != nil {
		logs.Error(fmt.Errorf("Update failed: %v", updateErr))
	}

	logs.Infof("修改后的result.Spec.Name：", result.Spec.Name)
	logs.Info("Updated action...")
	prompt()

	// List 所有Action
	logs.Info("listing 筛选的action")
	lstOpts := metav1.ListOptions{
		LabelSelector: "environment",
	}
	list, err := actionsClient.List(context.TODO(), lstOpts)
	if err != nil {
		logs.Error(err)
	}
	for _, d := range list.Items {
		logs.Info(d)
	}

	logs.Info("listing done")
	prompt()

	//Patch 一个Action
	logs.Info("patching")
	patchResult, err := actionsClient.Patch(context.TODO(), "demo-actions", types.StrategicMergePatchType, patchAction, metav1.PatchOptions{})
	logs.Infof("patchResult: ", patchResult)
	logs.Info("patch Done")

	// List 所有Action
	logs.Info("listing")
	lstOpts = metav1.ListOptions{}
	list, err = actionsClient.List(context.TODO(), lstOpts)
	if err != nil {
		logs.Error(err)
	}
	for _, d := range list.Items {
		logs.Info(d)
	}

	logs.Info("listing done")
	prompt()

	// Delete一个Action
	logs.Info("deleting")
	err = actionsClient.Delete(context.TODO(), "demo-actions", metav1.DeleteOptions{})
	if err != nil {
		panic(err)
	}
	logs.Info("Deleted action...")
	prompt()

	// Delete 之后再次 List所有Action
	logs.Info("listing")
	lstOpts = metav1.ListOptions{}
	list, err = actionsClient.List(context.TODO(), lstOpts)
	if err != nil {
		logs.Error(err)
	}
	for _, d := range list.Items {
		logs.Info(d)
	}

	logs.Info("listing done")

	//DeleteCollection 删除所有Spec.Name=demo-action的Action
	logs.Info("deleting collection")
	lstOpts = metav1.ListOptions{
		FieldSelector: "Spec.Name=demo-action",
	}
	err = actionsClient.DeleteCollection(context.TODO(), metav1.DeleteOptions{}, lstOpts)
	if err != nil {
		logs.Error(err)
	}
	logs.Info("Deleted collection...")
	prompt()

	// DeleteCollection 之后再次 List所有Action
	logs.Info("listing")
	lstOpts = metav1.ListOptions{}
	list, err = actionsClient.List(context.TODO(), lstOpts)
	if err != nil {
		logs.Error(err)
	}
	for _, d := range list.Items {
		logs.Info(d)
	}
	logs.Info("listing done")

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

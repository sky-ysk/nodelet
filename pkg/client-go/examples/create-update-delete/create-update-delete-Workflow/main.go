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
	// 这里以访问资源Workflow为例，
	// 获取访问Workflow的客户端
	// 默认访问的Namespace是 ""

	workflowsClient := clientSet.Core().Workflows("test")

	workflow := &apis.Workflow{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "demo-workflows",
			Namespace: "test",
			Labels: map[string]string{
				"environment": "dev",
			},
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Workflow",
			APIVersion: "resources/v1",
		},
		Spec: apis.WorkflowSpec{
			Name: "demo-workflow",
		},
	}
	workflow2 := &apis.Workflow{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "demo-workflow2",
			Namespace: "test",
			Labels: map[string]string{
				"environment": "dev",
			},
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Workflow",
			APIVersion: "resources/v1",
		},
		Spec: apis.WorkflowSpec{
			Name: "demo-workflow",
		},
	}
	workflow3 := &apis.Workflow{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "demo-workflow3",
			Namespace: "test",
			Labels: map[string]string{
				"environment": "qa",
			},
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Workflow",
			APIVersion: "resources/v1",
		},
		Spec: apis.WorkflowSpec{
			Name: "demo-workflow",
		},
	}

	patchWorkflow, err := json.Marshal(map[string]interface{}{
		"objectMeta": map[string]interface{}{
			"namespace": "test",
		},
		"spec": map[string]interface{}{
			"workflowName": "patch-workflow-name",
			"hostName":     "master",
		},
	})

	//监听事件并打印  监听resources/v1/workflows
	go func() {
		logs.Trace("watching")
		var timeoutSeconds int64 = 20
		watchOptions := metav1.ListOptions{
			TimeoutSeconds: &timeoutSeconds,
		}

		watcher, err := workflowsClient.Watch(context.TODO(), watchOptions)
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

	// Create三个Workflow
	logs.Trace("creating")
	result, err := workflowsClient.Create(context.TODO(), workflow, metav1.CreateOptions{})
	if err != nil {
		logs.Errorf("Failed to create workflow: %v", err)
	} else {
		logs.Trace("created workflow", result)
	}
	_, err = workflowsClient.Create(context.TODO(), workflow2, metav1.CreateOptions{})
	_, err = workflowsClient.Create(context.TODO(), workflow3, metav1.CreateOptions{})
	prompt()

	//Update一个Workflow

	logs.Trace("updating")
	// 部分更改一个参数
	// 先Get一个Workflow ,更改Workflow的参数, UpdateWorkflow

	result, getErr := workflowsClient.Get(context.TODO(), "demo-workflows", metav1.GetOptions{})
	if getErr != nil {
		logs.Info(fmt.Errorf("Failed to get : %v", getErr))
	}

	logs.Tracef("get result", result)
	logs.Tracef("修改前的result.Spec.WorkflowName：", result.Spec.Name)

	result.Spec.Name = "updatedWorkflowName"
	_, updateErr := workflowsClient.Update(context.TODO(), result, metav1.UpdateOptions{})
	if updateErr != nil {
		logs.Error(fmt.Errorf("Update failed: %v", updateErr))
	}

	logs.Tracef("修改后的result.Spec.WorkflowName：", result.Spec.Name)
	logs.Tracef("Updated workflow...")
	prompt()

	// List 所有Workflow
	logs.Tracef("listing 筛选的workflow")
	lstOpts := metav1.ListOptions{
		LabelSelector: "environment",
	}
	list, err := workflowsClient.List(context.TODO(), lstOpts)
	if err != nil {
		logs.Error(err)
	}
	for _, d := range list.Items {
		logs.Trace(d)
	}

	logs.Trace("listing done")
	prompt()

	//Patch 一个Workflow
	logs.Trace("patching")
	patchResult, err := workflowsClient.Patch(context.TODO(), "demo-workflows", types.StrategicMergePatchType, patchWorkflow, metav1.PatchOptions{})
	logs.Trace("patchResult: ", patchResult)
	logs.Trace("patch Done")

	// List 所有Workflow
	logs.Trace("listing")
	lstOpts = metav1.ListOptions{}
	list, err = workflowsClient.List(context.TODO(), lstOpts)
	if err != nil {
		logs.Error(err)
	}
	for _, d := range list.Items {
		logs.Trace(d)
	}

	logs.Trace("listing done")
	prompt()

	// Delete一个Workflow
	logs.Trace("deleting")
	err = workflowsClient.Delete(context.TODO(), "demo-workflows", metav1.DeleteOptions{})
	if err != nil {
		panic(err)
	}
	logs.Trace("Deleted workflow...")
	prompt()

	// Delete 之后再次 List所有Workflow
	logs.Trace("listing")
	lstOpts = metav1.ListOptions{}
	list, err = workflowsClient.List(context.TODO(), lstOpts)
	if err != nil {
		logs.Error(err)
	}
	for _, d := range list.Items {
		logs.Trace(d)
	}

	logs.Tracef("listing done")

	//DeleteCollection 删除所有Spec.WorkflowName=demo-workflow的Workflow
	logs.Tracef("deleting collection")
	lstOpts = metav1.ListOptions{
		FieldSelector: "Spec.WorkflowName=demo-workflow",
	}
	err = workflowsClient.DeleteCollection(context.TODO(), metav1.DeleteOptions{}, lstOpts)
	if err != nil {
		logs.Error(err)
	}
	logs.Tracef("Deleted collection...")
	prompt()

	// DeleteCollection 之后再次 List所有Workflow
	logs.Trace("listing")
	lstOpts = metav1.ListOptions{}
	list, err = workflowsClient.List(context.TODO(), lstOpts)
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

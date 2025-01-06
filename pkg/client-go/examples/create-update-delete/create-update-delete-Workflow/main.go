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
	logs.Init("create-update-delete-Workflow")
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
	// 这里以访问资源Workflow为例，
	// 获取访问Workflow的客户端
	// 默认访问的Namespace是 ""

	workflowsClient := clientSet.Core().Workflows("test")

	workflow1 := &apis.Workflow{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "demo-workflows",
			Namespace: "test",
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Workflow",
			APIVersion: "resources/v1",
		},
	}
	//workflow2 := &apis.Workflow{
	//	ObjectMeta: metav1.ObjectMeta{
	//		Name: "demo-workflow2",
	//	},
	//	TypeMeta: metav1.TypeMeta{
	//		Kind:       "Workflow",
	//		APIVersion: "resources/v1",
	//	},
	//	Spec: apis.WorkflowSpec{
	//		Name: "demo-workflow",
	//	},
	//}
	//workflow3 := &apis.Workflow{
	//	ObjectMeta: metav1.ObjectMeta{
	//		Name: "demo-workflow3",
	//	},
	//	TypeMeta: metav1.TypeMeta{
	//		Kind:       "Workflow",
	//		APIVersion: "resources/v1",
	//	},
	//	Spec: apis.WorkflowSpec{
	//		Name: "demo-workflow",
	//	},
	//}
	patchWorkflow, err := json.Marshal(map[string]interface{}{
		"Spec": map[string]interface{}{
			"Name": "demo-workflow",
		},
	})

	// Create一个Workflow
	fmt.Println("creating")
	results, err := workflowsClient.Create(context.TODO(), workflow1, metav1.CreateOptions{})

	if err != nil {
		panic(err)
	}
	//_, _ = workflowsClient.Create(context.TODO(), workflow2, metav1.CreateOptions{})
	//_, _ = workflowsClient.Create(context.TODO(), workflow3, metav1.CreateOptions{})
	fmt.Println("Created workflow ", results)

	prompt()

	//Patch 一个Workflow
	fmt.Println("patching")
	patchResult, err := workflowsClient.Patch(context.TODO(), "demo-workflows", types.StrategicMergePatchType, []byte(patchWorkflow), metav1.PatchOptions{})
	fmt.Println("patchResult: ", patchResult)
	fmt.Println("patch Done")

	//Update一个Workflow

	fmt.Println("updating")
	// 部分更改一个参数
	// 先Get一个Workflow ,更改Workflow的参数, UpdateWorkflow

	result, getErr := workflowsClient.Get(context.TODO(), "demo-workflows", metav1.GetOptions{})
	if getErr != nil {
		panic(fmt.Errorf("Failed to get : %v", getErr))
	}

	fmt.Println("get result", result)
	fmt.Println("修改前的result.Spec.Name：", result.Spec.Name)

	result.Spec.Name = "updatedName"
	_, updateErr := workflowsClient.Update(context.TODO(), result, metav1.UpdateOptions{})
	if updateErr != nil {
		panic(fmt.Errorf("Update failed: %v", updateErr))
	}

	fmt.Println("修改后的result.Spec.Name：", result.Spec.Name)
	fmt.Println("Updated workflow...")
	prompt()

	// List所有Workflow
	fmt.Println("listing")
	lstOpts := metav1.ListOptions{}
	list, err := workflowsClient.List(context.TODO(), lstOpts)
	if err != nil {
		panic(err)
	}
	for _, d := range list.Items {
		fmt.Println(d)
	}

	fmt.Println("listing done")
	prompt()

	// Delete一个Workflow
	fmt.Println("deleting")
	err = workflowsClient.Delete(context.TODO(), "demo-workflows", metav1.DeleteOptions{})
	if err != nil {
		panic(err)
	}
	fmt.Println("Deleted workflow...")
	prompt()

	// Delete 之后再次 List所有Workflow
	fmt.Println("listing")
	lstOpts = metav1.ListOptions{}
	list, err = workflowsClient.List(context.TODO(), lstOpts)
	if err != nil {
		panic(err)
	}
	for _, d := range list.Items {
		fmt.Println(d)
	}
	fmt.Println("listing done")

	//DeleteCollection 删除所有Spec.WorkflowName=demo-workflow的Workflow
	fmt.Println("deleting collection")
	lstOpts = metav1.ListOptions{
		FieldSelector: "Spec.Name=demo-workflow",
	}
	err = workflowsClient.DeleteCollection(context.TODO(), metav1.DeleteOptions{}, lstOpts)
	if err != nil {
		panic(err)
	}
	fmt.Println("Deleted collection...")

	// DeleteCollection 之后再次 List所有Workflow
	fmt.Println("listing")
	lstOpts = metav1.ListOptions{}
	list, err = workflowsClient.List(context.TODO(), lstOpts)
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

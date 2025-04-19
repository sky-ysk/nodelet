package main

import (
	"bufio"
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
	"os"
	"time"
)

// 创建一个Rest Client
// 验证xxx动词
// 与API Server通信，并执行基础操作
// 1个group，1个Action，每个Action一个Runtime， 一共1个Runtime，其中第一个group为训练任务（debian1上处理），测试迁移到pve2上
func main() {
	moduleName := "testModule"
	logs.Init(moduleName)
	scheme := runtime.NewScheme()
	apis.AddToScheme(scheme)
	logs.Info(scheme)
	//参数配置
	// TODO: 填写参数
	//部分参数之后可以在core_client等 编写setConfigDefaults函数进行填充
	c := &rest.Config{
		Host:    "http://localhost:10000", //http://suda801.wangwanu.com:11006
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
	// 这里以访问资源Task为例，
	// 获取访问Task的客户端
	// 默认访问的Namespace是 ""

	tasksClient := clientSet.Core().Tasks("test")
	groupsClient := clientSet.Core().Groups("test")
	actionsClient := clientSet.Core().Actions("test")
	runtimesClient := clientSet.Core().Actions("test")
	eventsClient := clientSet.Core().Events("test")

	// Task资源
	list1, err := tasksClient.List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err)
	}
	for _, task := range list1.Items {
		err := tasksClient.Delete(context.TODO(), task.Name, metav1.DeleteOptions{})
		if err != nil {
			panic(err)
		}
		logs.Infof("Task删除成功: %v", task.Name)
	}
	// group资源
	list2, err := groupsClient.List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err)
	}
	for _, group := range list2.Items {
		err := groupsClient.Delete(context.TODO(), group.Name, metav1.DeleteOptions{})
		if err != nil {
			panic(err)
		}
		logs.Infof("Group删除成功: %v", group.Name)
	}
	// action 资源
	list3, err := actionsClient.List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err)
	}
	for _, action := range list3.Items {
		err := actionsClient.Delete(context.TODO(), action.Name, metav1.DeleteOptions{})
		if err != nil {
			panic(err)
		}
		logs.Infof("Action删除成功: %v", action.Name)
	}
	// runtime 资源
	list4, err := runtimesClient.List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err)
	}
	for _, runtime := range list4.Items {
		err := runtimesClient.Delete(context.TODO(), runtime.Name, metav1.DeleteOptions{})
		if err != nil {
			panic(err)
		}
		logs.Infof("Runtime删除成功:%v", runtime.Name)
	}

	// Event资源
	list5, err := eventsClient.List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err)
	}
	for _, event := range list5.Items {
		err := eventsClient.Delete(context.TODO(), event.Name, metav1.DeleteOptions{})
		if err != nil {
			panic(err)
		}
		logs.Infof("Event删除成功: %v", event.Name)
	}

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
	logs.Info()
}

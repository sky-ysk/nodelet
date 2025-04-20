package main

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"hit.edu/framework/pkg/apimachinery/runtime"
	"hit.edu/framework/pkg/apimachinery/runtime/schema"
	"hit.edu/framework/pkg/apimachinery/runtime/serializer"
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
	// 这里以访问资源Runtime为例，
	// 获取访问Runtime的客户端
	// 默认访问的Namespace是 ""

	runtimesClient := clientSet.Core().Runtimes("test")
	runtimesClient2 := clientSet.Core().Runtimes("test2")
	runtimesClientAll := clientSet.Core().Runtimes(metav1.NamespaceAll)

	runtime := &apis.Runtime{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "demo-runtimes",
			Namespace: "test",
			Labels: map[string]string{
				"environment": "dev",
			},
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Runtime",
			APIVersion: "resources/v1",
		},
		Spec: apis.RuntimeSpec{
			Name: "demo-runtime",
		},
	}
	runtime2 := &apis.Runtime{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "demo-runtime2",
			Namespace: "test",
			Labels: map[string]string{
				"environment": "dev",
			},
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Runtime",
			APIVersion: "resources/v1",
		},
		Spec: apis.RuntimeSpec{
			Name: "demo-runtime",
		},
	}
	runtime3 := &apis.Runtime{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "demo-runtime3",
			Namespace: "test2",
			Labels: map[string]string{
				"environment": "qa",
			},
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Runtime",
			APIVersion: "resources/v1",
		},
		Spec: apis.RuntimeSpec{
			Name: "demo-runtime",
		},
	}

	//监听事件并打印  监听resources/v1/runtimes
	go func() {
		logs.Info("watching")
		var timeoutSeconds int64 = 20
		watchOptions := metav1.ListOptions{
			TimeoutSeconds: &timeoutSeconds,
		}

		watcher, err := runtimesClient.Watch(context.TODO(), watchOptions)
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

	// Create三个Runtime
	logs.Trace("creating")
	result, err := runtimesClient.Create(context.TODO(), runtime, metav1.CreateOptions{})
	if err != nil {
		logs.Errorf("Failed to create runtime: %v", err)
	} else {
		logs.Infof("created runtime", result)
	}
	_, err = runtimesClient.Create(context.TODO(), runtime2, metav1.CreateOptions{})
	_, err = runtimesClient2.Create(context.TODO(), runtime3, metav1.CreateOptions{})
	prompt()

	// 使用runtimesClient 进行List 所有namespace为test的Runtime
	logs.Info("listing")
	lstOpts := metav1.ListOptions{}
	list, err := runtimesClient.List(context.TODO(), lstOpts)
	if err != nil {
		logs.Error(err)
	}
	for _, d := range list.Items {
		logs.Info(d)
	}

	logs.Info("listing done")

	// // 使用runtimesClientAll 进行List 所有namespace为的Runtime，包括test和test2
	logs.Info("listing all")
	lstOpts = metav1.ListOptions{}
	list, err = runtimesClientAll.List(context.TODO(), lstOpts)
	if err != nil {
		logs.Error(err)
	}
	for _, d := range list.Items {
		logs.Info(d)
	}

	logs.Info("listing done")
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

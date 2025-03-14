package main

import (
	"bufio"
	"context"
	"fmt"
	"hit.edu/framework/pkg/apimachinery/runtime"
	"hit.edu/framework/pkg/apimachinery/runtime/schema"
	"hit.edu/framework/pkg/apimachinery/runtime/serializer"
	"hit.edu/framework/pkg/apimachinery/util/json"
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

	tasksClient := clientSet.Core().Tasks("")
	groupsClient := clientSet.Core().Groups("")

	group1 := &apis.Group{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "TrainGroup-1",
			Namespace: "test",
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Group",
			APIVersion: "resources/v1",
		},
		Spec: apis.GroupSpec{
			Name: "demo-group",
		},
	}

	//监听事件并打印  监听resources/v1/tasks
	go func() {
		logs.Infof("watching")
		watchOptions := metav1.ListOptions{}

		watcher, err := tasksClient.Watch(context.TODO(), watchOptions)
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
					logs.Infof("watchChan closed")
					return
				}

				// 打印事件类型和对象的相关信息
				logs.Infof("接收到事件类型: %v\n", event.Type)
				switch event.Type {
				case watch.Added:
					logs.Infof("资源被添加: ", event.Object)
				case watch.Modified:
					logs.Infof("资源被修改: ", event.Object)
				case watch.Deleted:
					logs.Infof("资源被删除: ", event.Object)
				case watch.Error:
					logs.Infof("发生错误: ", event.Object)
				default:
					logs.Infof("未识别的事件类型: ", event.Type)
				}
			}
		}
	}()

	//如果已经存在，先删掉
	err1 := groupsClient.Delete(context.TODO(), "TrainGroup-1", metav1.DeleteOptions{})
	err2 := groupsClient.Delete(context.TODO(), "TrainGroup-1-Copy", metav1.DeleteOptions{})

	if err1 != nil {
		logs.Errorf("group1 delete error: %v", err1)
	}
	if err2 != nil {
		logs.Errorf("group1-copy delete error: %v", err2)
	}
	logs.Infof("groupInfo:%v", group1)
	logs.Infof("creating")
	g, err1 := groupsClient.Create(context.TODO(), group1, metav1.CreateOptions{})

	if err1 != nil {
		logs.Errorf("Failed to create group1: %v", err)
		panic(err)
	}
	data, err := json.Marshal(g)
	if err != nil {
		logs.Errorf("Marshal group:%v error:%v", g.Name, err)
	}
	// 反序列化为新的对象
	var copyGroup apis.Group // 非指针
	err = json.Unmarshal(data, &copyGroup)
	if err != nil {
		logs.Errorf("Unmarshal group:%v error:%v", g.Name, err)
	}
	var groupCopy = &copyGroup // 转换为指针
	// 接下来修改这个复制出来的Group信息，首先修改group.Name
	groupCopy.Name = g.Name + "-Copy"
	groupCopy.ResourceVersion = ""

	logs.Infof("groupInfo:%v", groupCopy)
	_, err1 = groupsClient.Create(context.TODO(), groupCopy, metav1.CreateOptions{})
	if err1 != nil {
		logs.Errorf("Failed to create task: %v", err1)
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

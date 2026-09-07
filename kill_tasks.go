package main

import (
	"fmt"
	"net/http"
	"time"

	"hit.edu/framework/pkg/apimachinery/runtime"
	"hit.edu/framework/pkg/apimachinery/runtime/schema"
	"hit.edu/framework/pkg/apimachinery/runtime/serializer"
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/client-go/clients"
	"hit.edu/framework/pkg/client-go/rest"
	"hit.edu/framework/pkg/client-go/util/manager"
	"hit.edu/framework/pkg/component-base/logs"
	"hit.edu/framework/pkg/nodelet/events"
)

// 首先启动一个长时间运行的任务，调用发送kill事件
var scheme = runtime.NewScheme()


// 测试切换
// 1个group，1个Action，每个Action1个Runtime， 一共1个Runtime
func main() {
	moduleName := "killTasksModule"
	logs.Init(moduleName)

	//创建ClientSet
	clientSet := initClientSet(scheme)

	m := manager.NewManager(clientSet)
	tasks, err := m.GetTasks("")
	if err != nil {
		logs.Errorf("GetTasks err: %v", err)
	}
	logs.Infof("tasks num: %+v", len(tasks.Items))

	for i := range tasks.Items {
		task := tasks.Items[i]
		t, err := m.GetTask(task.Name, task.Namespace)
		logs.Infof("GetTask : name %v", t.Name)
		if err != nil {
			logs.Errorf("GetTask err: %v", err)
		}
		m.LogEvent(t, apis.EventTypeNormal, events.KillingCommand, fmt.Sprintf("The task %v is need to kill", t.Name), t.Namespace)
	}

	logs.Info("Kill events sent successfully")
	time.Sleep(1 * time.Second)
}

func initClientSet(scheme *runtime.Scheme) *clients.ClientSet {
	apis.AddToScheme(scheme)
	logs.Info(scheme)
	// 创建ClientSet
	c := &rest.Config{
		Host:    "http://127.0.0.1:8120",
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
	clientSet, err := clients.NewForConfig(c)
	if err != nil {
		panic(err)
	}
	return clientSet
}

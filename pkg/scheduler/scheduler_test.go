package scheduler

import (
	"context"
	"fmt"
	"hit.edu/framework/pkg/apimachinery/runtime"
	"hit.edu/framework/pkg/apimachinery/runtime/schema"
	"hit.edu/framework/pkg/apimachinery/runtime/serializer"
	apis "hit.edu/framework/pkg/apis/cores"
	metav1 "hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/client-go/clients"
	"hit.edu/framework/pkg/client-go/rest"
	"hit.edu/framework/pkg/scheduler/backend/queue"
	"net/http"
	"testing"
	"time"
)

//测试调度框架

func TestScheduler_Run(t *testing.T) {

	//stopEverything := ctx.Done()

	// 配置调度器启动选项，在这里需要定义所需的模块，插件
	//options := defaultSchedulerOptions
	//for _, opt := range opts {
	//	opt(&options)
	//}

	// 配置调度器参数

	// 配置插件模块
	//registry := plugins.NewInTreeRegistry()

	// 配置任务队列git

	// 配置资源监控模块
	//scheduleChan := make(chan internal.ScheduleSignal)
	queue := queue.NewPriorityQueue()
	sched := &Scheduler{
		//StopEverything:  stopEverything,
		//ScheduleSigChan: scheduleChan,
		SchedulingQueue: queue,
	}

	//schedQueue := &queue.PriorityQueue{}

	sched.applyDefaultHandlers()
	sched.ReadyGroup = sched.SchedulingQueue.Pop

	//return sched, nil
}

func TestSendGroupToScheduler(t *testing.T) {
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
	// 这里以访问资源Group为例，
	// 获取访问Group的客户端
	// 默认访问的Namespace是 ""

	groupsClient := clientSet.Core().Groups("")

	group1 := &apis.Group{
		ObjectMeta: metav1.ObjectMeta{
			Name: "demo-group1",
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Group",
			APIVersion: "resources/v1",
		},
	}

	group2 := &apis.Group{
		ObjectMeta: metav1.ObjectMeta{
			Name: "demo-group2",
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Group",
			APIVersion: "resources/v1",
		},
	}

	result1, err := groupsClient.Create(context.TODO(), group1, metav1.CreateOptions{})
	if err != nil {
		panic(err)
	}
	fmt.Println("Created group ", result1)
	result2, err := groupsClient.Create(context.TODO(), group2, metav1.CreateOptions{})
	if err != nil {
		panic(err)
	}
	fmt.Println("Created group ", result2)
}

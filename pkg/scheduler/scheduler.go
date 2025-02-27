/* ==================================================================
* Copyright (c) 2024, HIT Authors
* All rights reserved.
*
* Redistribution and use in source and binary forms, with or without
* modification, are permitted provided that the following conditions
* are met:
*
* 1. Redistributions of source code must retain the above copyright
* notice, this list of conditions and the following disclaimer.
* 2. Redistributions in binary form must reproduce the above copyright
* notice, this list of conditions and the following disclaimer in the
* documentation and/or other materials provided with the
* distribution.
* 3. All advertising materials mentioning features or use of this software
* must display the following acknowledgement:
* This product includes software developed by the xxx Group. and
* its contributors.
* 4. Neither the name of the Group nor the names of its contributors may
* be used to endorse or promote products derived from this software
* without specific prior written permission.
*
* THIS SOFTWARE IS PROVIDED BY Wanyou Wang,GROUP AND CONTRIBUTORS
* ===================================================================
* Author: Wanyou Wang
 */
package scheduler

import (
	"context"
	"fmt"
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/client-go/clients"
	"hit.edu/framework/pkg/component-base/logs"
	"hit.edu/framework/pkg/scheduler/apis/config"
	"hit.edu/framework/pkg/scheduler/backend/queue"
	"hit.edu/framework/pkg/scheduler/framework"
	"hit.edu/framework/pkg/scheduler/framework/plugins"
	"hit.edu/framework/pkg/scheduler/internal"
	"time"

	// extension "hit.edu/framework/pkg/scheduler/schedulechain"
	"hit.edu/framework/pkg/apimachinery/runtime"
	"hit.edu/framework/pkg/apimachinery/runtime/schema"
	"hit.edu/framework/pkg/apimachinery/runtime/serializer"
	"hit.edu/framework/pkg/apimachinery/watch"
	metav1 "hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/client-go/rest"
	schedRuntime "hit.edu/framework/pkg/scheduler/framework/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"net/http"
)

// TODO: 将K8s相关组件替换为我们自己的

//TODO: 调度器配置
//TODO: 服务器，接受参数，对应接口格式
//TODO: 从resource let获取信息，相关接口介入
//TODO: 调度结果确定
//TODO: 插件系统确定，获取算法
//TODO: 大模型相关内容接入
//TODO: 模块定制，裁剪
//TODO: 日志模块替换
//TODO: 不同调度器，如微服务、RMF/机器人类的任务
//TODO: 并发相关组件
//TODO: 事件处理

type ProfileMap map[string]framework.Framework

var ErrNoNodesAvailable = fmt.Errorf("no nodes available to schedule pods")

type Scheduler struct {
	Extenders []framework.Extender

	DefaultFramework framework.Framework

	// Close this to shut down the scheduler.
	StopEverything <-chan struct{}

	SchedulingQueue queue.SchedulingQueue

	ScheduleSigChan chan internal.ScheduleSignal

	// 改自NextPod，阻塞函数
	ReadyGroup func(ctx context.Context) (*config.QueuedGroupInfo, error)

	ScheduleGroup func(ctx context.Context, fwk framework.Framework, state *framework.CycleState, group *apis.Group) (ScheduleResult, error)

	percentageOfNodesToScore int32

	Profiles ProfileMap
	// ScheduleExt     []*extension.ScheduleExtension
	// PreScheduleExt  []*extension.PreScheduleExtension
	// PostScheduleExt []*extension.PostScheduleExtension
}

// Scheduler Options
type schedulerOptions struct {
	//TODO: DSL/大模型支持
	//TODO: ScheduleChain支持，算法插件系统
	parallelism int32
}

// 配置调度器Options
type Option func(*schedulerOptions)

var defaultSchedulerOptions = schedulerOptions{
	// 调度器并行化调度参数
	// 调度器资源请求需要串行保证一致性，其他的算法、大模型、事件请求都可以并行化处理
	parallelism: 2,
}

const (
	SchedulerPipelineNum = 2
)

var defaultScheduler Scheduler

// TODO: 挪到apis中
// func indexHandler(w http.ResponseWriter, r *http.Request) {
// 	fmt.Fprintf(w, "hello and this is the scheduler")
// }

// func (sched *Scheduler) Run(ctx context.Context) {
// 	http.HandleFunc("/test", indexHandler)
// 	http.ListenAndServe(":8000", nil)
// 	go mockScheduleSignal(sched.ScheduleSigChan)

// 	// TODO: 调度器核心逻辑

// 	for i := 0; i < SchedulerPipelineNum; i++ {
// 		go sched.SchedulingPipeline(sched.ScheduleSigChan, ctx)
// 	}
// }

func init() {

}

// 创建新的Scheduler对象
func New(ctx context.Context, opts ...Option) (*Scheduler, error) {
	logs.Info("init scheduler... ")
	stopEverything := ctx.Done()

	// 配置调度器启动选项，在这里需要定义所需的模块，插件
	options := defaultSchedulerOptions
	for _, opt := range opts {
		opt(&options)
	}

	// 配置调度器参数

	// 配置插件模块
	registry := plugins.NewInTreeRegistry()

	// 配置任务队列

	// 配置资源监控模块
	scheduleChan := make(chan internal.ScheduleSignal)
	//queue := internal.NewSchedulingQueue()
	schedQueue := queue.NewPriorityQueue()
	schedQueue.Run(ctx)
	defaultFramework, err := schedRuntime.NewDefaultFramework(ctx, registry, "DefaultScheduler")
	if err != nil {
		return nil, err
	}
	sched := &Scheduler{
		StopEverything:   stopEverything,
		ScheduleSigChan:  scheduleChan,
		SchedulingQueue:  schedQueue,
		DefaultFramework: defaultFramework,
	}

	sched.applyDefaultHandlers()
	sched.ReadyGroup = schedQueue.Pop

	return sched, nil
}

// ScheduleResult represents the result of scheduling a pod.
type ScheduleResult struct {
	// Name of the selected node.
	SuggestedHost string
	// The number of nodes the scheduler evaluated the pod against in the filtering
	// phase and beyond.
	EvaluatedNodes int
	// The number of nodes out of the evaluated ones that fit the pod.
	FeasibleNodes int
	// The nominating info for scheduling cycle.
	nominatingInfo *framework.NominatingInfo

	Group *apis.Group
}

func (sched *Scheduler) Run(ctx context.Context) {
	// 初始化任务优先级队列
	//TODO 做成调度器一个变量来控制
	concurrency := 3
	for i := 0; i < concurrency; i++ {
		go wait.UntilWithContext(ctx, sched.ScheduleOne, 0)
	}
	go sched.monitorWorkflow(ctx)
	<-ctx.Done()

	// TODO: 具体内容实现
}

func (sched *Scheduler) applyDefaultHandlers() {
	sched.ScheduleGroup = sched.scheduleGroup
	//TODO 错误处理handler
	//sched.FailureHandler = sched.s
}

func (sched *Scheduler) monitorWorkflow(ctx context.Context) {
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
		//设置监听通道一小时关闭
		Timeout: 3600 * time.Second,
	}

	//创建ClientSet
	clientSet, err := clients.NewForConfig(c)
	if err != nil {
		panic(err)
	}
	// 资源定义在 pkg/apis/xxx/type.go 下
	// 这里以访问资源Node为例，
	// 获取访问Node的客户端
	// 默认访问的Namespace是 ""

	groupClient := clientSet.Core().Groups("")
	logs.Info("scheduler start watching groups")
	//设置监听通道一小时关闭
	var watchTimeout int64 = 3600
	watchOptions := metav1.ListOptions{
		TimeoutSeconds: &watchTimeout,
	}

	watcher, err := groupClient.Watch(context.TODO(), watchOptions)
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
				logs.Info("watchChan closed")
				return
			}
			// 打印事件类型和对象的相关信息
			msg := fmt.Sprintf("scheduler接收到group事件类型: %v\n", event.Type)
			fmt.Printf(msg)
			switch event.Type {
			case watch.Added:
				{
					addMsg := fmt.Sprintf("资源被添加: %s", event.Object)
					fmt.Println(addMsg)
					sched.handleGroupAdd(ctx, event)
				}
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
}

func (sched *Scheduler) handleGroupAdd(ctx context.Context, event watch.Event) {
	if g, ok := event.Object.(*apis.Group); ok {
		sched.SchedulingQueue.Add(ctx, g)
	} else {
		logs.Error("cannot convert to group")
	}

}

// func mockScheduleSignal(scheChan chan<- internal.ScheduleSignal) {
// 	for {
// 		scheChan <- nil
// 		time.Sleep(500 * time.Millisecond)
// 	}
// }

// func (sched *Scheduler) SchedulingPipeline(scheChan chan internal.ScheduleSignal, ctx context.Context) {
// 	for {
// 		select {
// 		case <-scheChan:
// 			sched.SchedulingCycle(ctx)
// 		}
// 	}
// }

// // TODO: 核心内容移到framework/plugins中
// func (sched *Scheduler) RegisterScheduleExt(ext *extension.ScheduleExtension) {
// 	sched.ScheduleExt = append(sched.ScheduleExt, ext)
// }

// func (sched *Scheduler) RegisterPreScheduleExt(ext *extension.PreScheduleExtension) {
// 	sched.PreScheduleExt = append(sched.PreScheduleExt, ext)
// }

// func (sched *Scheduler) RegisterPostScheduleExt(ext *extension.PostScheduleExtension) {
// 	sched.PostScheduleExt = append(sched.PostScheduleExt, ext)
// }

// func (sched *Scheduler) MatchScheduleExt(group *workflow.Group) *extension.ScheduleExtension {
// 	return nil
// }

// // TODO: 改成Run函数
// func (sched *Scheduler) SchedulingCycle(ctx context.Context) {
// 	holdLock := false

// 	defer func() {
// 		if holdLock {
// 			sched.SchedulingQueue.ReadyGroups.SyncLock.Unlock()
// 			holdLock = false
// 		}
// 	}()
// 	//pop queue
// 	sched.SchedulingQueue.ReadyGroups.SyncLock.IsLocked()
// 	readyGroups := sched.SchedulingQueue.ReadyGroups.GetQueue()
// 	for _, group := range readyGroups {
// 		//TODO match
// 		scheExt := sched.MatchScheduleExt(&group)
// 		(*scheExt).Schedule(&group)
// 	}
// 	//fail -> event
// 	sched.SchedulingQueue.ReadyGroups.Empty()
// 	holdLock = true

// 	//unlock after pop queue to avoid Blocking other pipeline
// 	sched.SchedulingQueue.ReadyGroups.SyncLock.Unlock()
// 	holdLock = false
// 	//TODO: generate

// 	//TODO: schedule

// 	//TODO bind

// 	//enqueue

// }

package task

import (
	"context"
	"fmt"
	apis "hit.edu/framework/pkg/apis/cores"
	metav1 "hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/client-go/clients"
	"hit.edu/framework/pkg/client-go/clients/typed/core"
	"hit.edu/framework/pkg/component-base/logs"
	"hit.edu/framework/pkg/nodelet/events/eventbus"
	"hit.edu/framework/pkg/nodelet/task/group"
	"hit.edu/framework/pkg/nodelet/task/monitor"
	"hit.edu/framework/pkg/nodelet/task/runtime"
	"hit.edu/framework/pkg/nodelet/task/task"
	"hit.edu/framework/pkg/nodelet/task/types"
)

type Exporter interface {
	Run(ctx context.Context) error
}

var (
	updateCh chan types.GroupUpdate
)

type TaskExporter struct {
	// TODO: 增加Client-Go配置  --这块有点不太清楚,应该是为了方便将任务状态存到etcd当中
	nodesClient core.NodeInterface
	gropsClient core.GroupInterface
	tasksClient core.TaskInterface
	// TODO: 增加Event Recorder

	// TODO: 增加Group Lister
	groupLister []*apis.Group

	// 管理所有所有的Group
	groupManager group.Manager
	// 管理所有的Task
	taskManager task.Manager

	// Group的实际执行单元
	groupWorkers group.GroupWorkers

	// 监控正在执行中Group，包括已经部署的Action的执行状态，和未部署的Action的执行条件
	groupMonitor *monitor.GroupMonitor

	// 处理从上游（API-Server）中的Group的更新事件
	groupHandler *monitor.GroupHandler
}

var _ Exporter = &TaskExporter{}

func NewTaskExporter(cfg *Config, clientset *clients.ClientSet) (*TaskExporter, error) {
	// Task Exporter配置 config

	//client配置
	nodeClient := clientset.Core().Nodes("")
	taskClient := clientset.Core().Tasks("")
	groupClient := clientset.Core().Groups("")
	//事件配置
	eb := eventbus.NewEventBus()
	// Manager配置 group
	groupManager := group.NewGroupManager()
	// Manager 配置Task
	taskManager := task.NewTaskManager()
	// lister
	lister := groupManager.GetGroups(nil)
	// runtimeManager的配置
	runtimeManager := runtime.NewRuntimeManager(eb)
	// queue_manager
	groupQueues := group.NewGroupQueues(groupManager)
	// workers
	workers := group.NewGroupWorkers(groupManager, taskManager, groupQueues, runtimeManager, groupClient, taskClient)

	taskExporter := &TaskExporter{
		// Monitor配置
		nodesClient:  nodeClient,
		tasksClient:  taskClient,
		gropsClient:  groupClient,
		groupManager: groupManager,
		taskManager:  taskManager,
		groupLister:  lister,
		groupWorkers: workers,
		groupMonitor: monitor.NewGroupMonitor(groupManager, taskManager, groupQueues, eb, runtimeManager, nodeClient, groupClient, taskClient),
		groupHandler: monitor.NewGroupHandler(groupManager, workers, groupQueues),
	}
	// Client-Go配置

	// 需要一个TaskCache,存储当前节点所有的Task信息 ====这是什么意思,有点没懂 ？-hzy
	logs.Info("init task exporter")
	return taskExporter, nil
}

func (te *TaskExporter) Run(ctx context.Context) error {
	// 部署Client-Go服务(TODO 了解client-go服务的配置)

	//监听Group资源  ***  question: ====监听group任务所需资源还是group任务资源===

	// 当Group bind到当前节点时，开始部署当前Group
	// Group中包含多个Action,Action支持串行和并行执行
	// 任务部署前，需要检查任务的依赖，需要检查的内容包括
	//   资源依赖，任务所需计算、网络、存储或者硬件资源是否就绪
	//   顺序依赖，前序节点是否满足
	//   数据依赖，任务执行所需数据是否准备好
	//   条件依赖，任务执行是否满足条件

	// 任务部署完成后，需要监控任务的执行情况，并通过Client-Go定期更新
	// 任务监控中会产生各类事件，事件也通过Client-Go更新

	// +optional,如果任务开启动态资源调整的需求
	// 任务执行过程中需要动态调整任务进程的资源, 根据当前任务执行的Spec和Status, 通过cGroup动态调整任务执行资源使用情况
	updateCh = make(chan types.GroupUpdate)
	go te.groupHandler.Loop(ctx, updateCh) //主要监控上层发来的消息，主要是启动、停止任务
	go te.groupMonitor.Start()             //主要监控正在启动的任务，获取任务状态信息
	select {
	case <-ctx.Done():
		return ctx.Err() //退出是返回错误
	}
}

// 模拟上层组件发送任务信息给Taskexporter，该方法主要是接受任务信息，并放入管道当中，触发Loop监听
func (te *TaskExporter) ReceiveGroupInfo(updateType string) {
	task := te.GetTask()
	logs.Info("receive task info")
	te.taskManager.AddTask(task)
	for i := range task.Spec.Groups {
		//_, err := te.taskManager.GetTaskByID(task.Status.TaskID)
		if task.Status.Phase == apis.Unknown {
			g := &task.Spec.Groups[i]
			// 这里可以改为从client-go中读取group信息
			if updateType == "create" {
				groupUpdate := types.GroupUpdate{
					Groups: []*apis.Group{g},
					Op:     types.ADD,
				}
				updateCh <- groupUpdate
			} else if updateType == "kill" {
				//TODO
				groupUpdate := types.GroupUpdate{
					Groups: []*apis.Group{g},
					Op:     types.KILL,
				}
				updateCh <- groupUpdate
			}
		}
	}
}

// 从client-go中读取task信息
func (te *TaskExporter) GetTask() *apis.Task {
	result, getErr := te.tasksClient.Get(context.TODO(), "TestTasks", metav1.GetOptions{})
	if getErr != nil {
		panic(fmt.Errorf("Failed to get : %v", getErr))
	}
	return result
}

package task

import (
	"context"
	"hit.edu/framework/pkg/client-go/util/manager"
	"sync"
	"time"

	metav1 "hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/nodelet/task/controller"
	"hit.edu/framework/pkg/nodelet/task/group/dependency"
	"hit.edu/framework/pkg/utils"

	scheme "hit.edu/framework/pkg/apimachinery/runtime"
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/client-go/clients"
	"hit.edu/framework/pkg/client-go/clients/typed/core"
	"hit.edu/framework/pkg/client-go/tools/recorder"
	"hit.edu/framework/pkg/component-base/logs"
	"hit.edu/framework/pkg/nodelet/events/eventbus"
	"hit.edu/framework/pkg/nodelet/task/group"
	"hit.edu/framework/pkg/nodelet/task/monitor"
	"hit.edu/framework/pkg/nodelet/task/runtime"
	"hit.edu/framework/pkg/nodelet/task/types"
)

type Exporter interface {
	Run(ctx context.Context) error
}

type TaskExporter struct {
	// TODO: 增加Client-Go配置
	clientsManager *manager.Manager
	// TODO: 增加Event Broadcaster Recorder
	eventBroadcaster recorder.EventBroadcaster

	// TODO: 增加Group Lister
	groupLister []*apis.Group

	// 管理所有所有的Group
	groupManager group.Manager
	// Group的实际执行单元
	groupWorkers group.GroupWorkers

	// 监控正在执行中Group，包括已经部署的Action的执行状态，和未部署的Action的执行条件
	groupMonitor *monitor.GroupMonitor

	// 处理从上游（API-Server）中的Group的更新事件
	groupHandler *monitor.GroupHandler

	//处理condition
	conditionEngine *utils.ConditionEngine

	// 切换模块
	//groupSwitcher *_switch.GroupSwitch
	migrationController *controller.MigrationController
	nodeMonitor         *controller.NodeMonitor
	// 当前Taskexporter所部署的节点的Name
	nodeName string

	updateCh chan types.GroupUpdate
}

var _ Exporter = &TaskExporter{}

func NewTaskExporter(cfg *Config, clientset *clients.ClientSet) (*TaskExporter, error) {
	// Task Exporter配置 config
	groupTargetMap := cfg.groupTargetMap
	actionTargetMap := cfg.actionTargetMap
	runtimeTargetMap := cfg.runtimeTargetMap
	// Client-Go配置
	//nodeClient := clientset.Core().Nodes("test")
	//taskClient := clientset.Core().Tasks("test")
	//groupClient := clientset.Core().Groups("test")
	eventClient := clientset.Core().Events("test")
	//actionClient := clientset.Core().Actions("test")
	//deviceClient := clientset.Core().Devices("test")
	//runtimeClient := clientset.Core().Runtimes("test")

	clientsManager := manager.NewManager(clientset)
	//事件总线--只使用与Runtime运行时传输状态的
	eb := eventbus.NewEventBus()
	// 全局事件组件的配置
	eventBroadcaster := recorder.NewBroadcaster()
	eventBroadcaster.StartRecordingToSink(context.Background(), &core.EventSinkImpl{Interface: eventClient})
	scheme := scheme.NewScheme()
	apis.AddToScheme(scheme)
	recorder := eventBroadcaster.NewRecorder(scheme, "TaskExporter")

	// Manager配置 group
	groupManager := group.NewGroupManager()

	// lister
	lister := groupManager.GetGroups(nil)
	// runtimeManager的配置
	runtimeManager := runtime.NewRuntimeManager(eb, recorder, clientsManager)
	//dependencyManager配置
	depenManager := dependency.NewDependencyManager()
	//condition engine配置
	conditionEngine := utils.NewConditionEngine()
	// queue_manager
	groupQueues := group.NewGroupQueues(groupManager)
	// workers
	workers := group.NewGroupWorkers(groupManager, groupQueues, runtimeManager, clientsManager)
	// 当前Taskexporter所在节点的NodeName
	nodeName := cfg.NodeName

	taskExporter := &TaskExporter{
		// Monitor配置
		//nodesClient:         nodeClient,
		//tasksClient:         taskClient,
		//gropsClient:         groupClient,
		clientsManager:      clientsManager,
		eventBroadcaster:    eventBroadcaster,
		conditionEngine:     conditionEngine,
		groupManager:        groupManager,
		groupLister:         lister,
		groupWorkers:        workers,
		groupMonitor:        monitor.NewGroupMonitor(groupManager, groupQueues, eb, recorder, runtimeManager, clientsManager, depenManager, groupTargetMap, actionTargetMap, runtimeTargetMap),
		groupHandler:        monitor.NewGroupHandler(groupManager, workers, groupQueues, clientsManager, recorder, eventClient, groupTargetMap, actionTargetMap, runtimeTargetMap),
		migrationController: controller.NewMigrationController(clientset, clientsManager, runtimeManager, groupQueues, recorder, nodeName, groupTargetMap, actionTargetMap, runtimeTargetMap, groupManager),
		nodeMonitor:         controller.NewNodeMonitor(clientset, recorder, nodeName),
		nodeName:            nodeName,
		updateCh:            make(chan types.GroupUpdate),
	}

	// 需要一个TaskCache,存储当前节点所有的Task信息 ====这是什么意思,有点没懂 ？-hzy
	logs.Info("Init task exporter")
	return taskExporter, nil
}

func (te *TaskExporter) Run(ctx context.Context) error {
	//监听Group资源
	// 任务监控中会产生各类事件，事件也通过Client-Go更新
	// +optional,如果任务开启动态资源调整的需求
	// 任务执行过程中需要动态调整任务进程的资源, 根据当前任务执行的Spec和Status, 通过cGroup动态调整任务执行资源使用情况
	defer close(te.updateCh)
	var wg sync.WaitGroup
	wg.Add(3) //等待三个协程
	go func() {
		defer wg.Done()
		te.groupHandler.Loop(ctx, te.updateCh) //主要监控上层发来的消息，主要是启动、停止任务
	}()

	go func() {
		defer wg.Done()
		te.groupMonitor.Start(ctx) //主要监控正在启动的任务，获取任务状态信息
	}()
	go func() {
		defer wg.Done()
		te.ReceiveGroupInfo(ctx) // 持续从etcd当中读取group
	}()

	//go te.migrationController.Run(5, ctx.Done())
	//go te.nodeMonitor.Run(2, ctx.Done())
	<-ctx.Done()
	wg.Wait()
	return ctx.Err()
	//select {
	//case <-ctx.Done():
	//	return ctx.Err() //退出是返回错误
	//}
}

func (te *TaskExporter) ReceiveGroupInfo(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			logs.Info("ReceiveGroupInfo exiting due to context cancel")
			return
		default:
			//读取 etcd当中的group列表
			groupsClient := te.clientsManager.GetGroupClient("test")
			groupList, err := groupsClient.Client.List(context.TODO(), metav1.ListOptions{})
			if err != nil {
				logs.Errorf("List task err:%v", err)
			}
			// 遍历group
			for i := range groupList.Items {
				gr := &groupList.Items[i] // 修改了此处，如果不行的话，改为原来的
				//groupName := gr.Name // 这里是一个坑
				//// 从etcd当中读group的信息
				//gr, err := te.gropsClient.Get(context.TODO(), groupName, metav1.GetOptions{})
				//if err != nil {
				//	logs.Errorf("get group:%s failed", groupName)
				//}
				if gr.Status.Node != nil && *gr.Status.Node == te.nodeName { //gr.Status.Node == "CloudNode1"       gr.Status.Node == "EdgeNode1" || gr.Status.Node == "EndNode1"
					if gr.Status.Phase == apis.ReadyToDeploy {
						groupUpdate := types.GroupUpdate{
							Group: gr,
							Op:    types.ADD,
						}
						te.updateCh <- groupUpdate
					} else if gr.Status.Phase == apis.ReadyToKill {
						groupUpdate := types.GroupUpdate{
							Group: gr,
							Op:    types.KILL,
						}
						te.updateCh <- groupUpdate
					}
				}
				// 读完一个Group暂停一会
				select {
				case <-ctx.Done():
					return
				case <-time.After(100 * time.Millisecond):
				}
			}
		}
		// GroupList 读完暂停一会
		select {
		case <-ctx.Done():
			return
		case <-time.After(100 * time.Millisecond):
		}
	}
}

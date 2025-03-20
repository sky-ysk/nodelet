package monitor

import (
	"context"
	"encoding/json"
	"fmt"
	ty "hit.edu/framework/pkg/apimachinery/types"
	"hit.edu/framework/pkg/apimachinery/watch"
	apis "hit.edu/framework/pkg/apis/cores"
	meta "hit.edu/framework/pkg/apis/meta"
	metav1 "hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/client-go/clients/typed/core"
	"hit.edu/framework/pkg/client-go/tools/recorder"
	"hit.edu/framework/pkg/component-base/logs"
	"hit.edu/framework/pkg/nodelet/events"
	"hit.edu/framework/pkg/nodelet/events/eventbus"
	"hit.edu/framework/pkg/nodelet/task/controller"
	"hit.edu/framework/pkg/nodelet/task/group"
	"hit.edu/framework/pkg/nodelet/task/types"
	"time"
)

type GroupHandler struct {
	// group managers
	groupManager group.Manager

	// group workers, 实际部署任务
	groupWorkers group.GroupWorkers

	groupQueues *group.GroupQueues
	// client -go
	groupClient core.GroupInterface
	//
	eventBus *eventbus.EventBus
	// eventRecorder 记录事件
	recorder    recorder.EventRecorder
	eventClient core.EventInterface

	stopCh chan struct{}
}

func NewGroupHandler(groupManager group.Manager, groupWorkers group.GroupWorkers, groupQueues *group.GroupQueues, groupClient core.GroupInterface, eb *eventbus.EventBus, recorder recorder.EventRecorder, eventClient core.EventInterface) *GroupHandler {
	return &GroupHandler{
		groupManager: groupManager,
		groupWorkers: groupWorkers,
		groupQueues:  groupQueues,
		groupClient:  groupClient,
		eventBus:     eb,
		recorder:     recorder,
		eventClient:  eventClient,
		stopCh:       make(chan struct{}),
	}
}

type Handler interface {
	HandleGroupAdds(groups []*apis.Group)
	HandleGroupUpdates(groups []*apis.Group)
	HandleGroupDeletes(groups []*apis.Group)
}

// 开启监听上层指令
func (gh *GroupHandler) Loop(ctx context.Context, updateCh <-chan types.GroupUpdate) {
	const base = 100 * time.Millisecond //

	logs.Info("GroupHandler component start")
	//duration := base
	for {
		// 檢查Context是否已取消
		select {
		case <-ctx.Done():
			logs.Info("Context canceled, exiting GroupHandler loop")
			return
		case <-time.After(base):
			gh.LoopIteration(ctx, updateCh) //是否采用协程，取决于该函数是否要与Loop方法并行执行
		}
		//time.Sleep(duration)
		//// TODO: 二进制指数退避
	}
}

// 针对不同的指令，对接受到的任务执行不同的操作（开启任务、更新任务、删除任务）
func (gh *GroupHandler) LoopIteration(ctx context.Context, updateCh <-chan types.GroupUpdate) {
	// TODO: Watch Group信息
	for {
		select {
		case <-ctx.Done():
			logs.Error("Context canceled, exiting loop iteration")
			return
		case u, ok := <-updateCh:
			if !ok {
				logs.Error("Update channel is closed, can‘t get group")
				return
			}
			switch u.Op {
			case types.ADD:
				logs.Debug("Add new group")
				gh.HandleGroupAdd(u.Group)
			case types.KILL:
				logs.Debug("Delete group")
				gh.HandleGroupKill(u.Group)
			case types.UPDATE:
				logs.Debug("Update group")
				gh.HandleGroupUpdate(u.Group)
			default:
				logs.Error("Unhandled default case")
			}
		}
	}
}

// 处理Group启动指令  主要内容：检查当前节点是否能执行group，以及检查当前节点是否接收过当前group（得依据groupID）
func (gh *GroupHandler) HandleGroupAdd(gr *apis.Group) {
	// TODO: 对Pod按照优先级排序（目前先按照创建时间） ---目前是处理发过来的单个Group，所以无法做排序工作
	start := time.Now()
	// TODO: 检查任务是否可以在当前节点上运行, 如果不能，则拒绝Group的部署
	//  不能部署的情况包括
	//  1、Group已经在本地部署（是否包括副本）？
	//  2、没有可以执行的资源
	//  ......
	// TODO 这里有一个问题：就是如果节点上部署不了这个group，那么久会导致一直Handler一直接收这个group，解决方法：检查后立即放入
	_, err := gh.groupManager.GetGroupByID(gr.Status.GroupID) // 使用groupID查，因为groupID是唯一分配的
	if err == nil {                                           //err等于nil说明在group_manager当中能找到group信息
		// 1、说明group已经存
		logs.Infof("Group:%s is already put into Deployer", gr.Spec.Name)
		return
	}
	//如果说部署器本地没有改groupID信息的话，就存入group信息
	gh.groupManager.AddGroup(gr)

	// 2、检查资源是否足够并满足部署条件
	if isCanDeploy := gh.checkResource(gr); !isCanDeploy {
		// 如果无法部署，拒绝改Group的部署并通知调度器
		logs.Errorf("Group:%s cannot be deployed,err:%v", gr.Name, err)
		// TODO 这里得直接提交给调度器，告知group无法部署
		return
	}
	// 3、判断group是否需要部署副本，如果需要，在此处往域或者跨域的etcd当中添加副本
	copiesInDomain := gr.Spec.Replicas[0]
	copiesInOtherDomain := gr.Spec.Replicas[1]
	if copiesInDomain > 0 { //如果传进任务的时候该属性没有赋值的话，初始化是为0的
		// 为了适配迁移 ,如果有多个副本要求的话，需要部署多个副本
		for i := 0; i < int(copiesInDomain); i++ {
			// 复制创建一个全新的副本group信息（注意Succeed的Phase不用修改，DeployCheck和Running状态需要修改），另外还需要将副本的groupStatus改为Starting
			groupCopyName := "Reason-Copy"                                        // TODO 这里之后改成随机生成即可源group.Name + 一串随机字符
			groupCopy := controller.NewGroupInfoCopy(gr, true, groupCopyName, "") // 第二个参数为true，表示的是提前写入etcd
			// 将副本group信息写入到etcd当中，目前还只适配本域内迁移
			logs.Infof("group:%v===================", groupCopy.Name)
			_, err = gh.groupClient.Create(context.TODO(), groupCopy, metav1.CreateOptions{})
			if err != nil {
				logs.Errorf("Create group:%s err: %v", groupCopy.Name, err)
			}

			// TODO 这里需要将副本信息写入到Copyinfo当中
			patchGroup, err := json.Marshal(map[string]interface{}{
				"spec": map[string]interface{}{
					"copy_info": map[string]string{groupCopy.Name: "local"}, //value值不同
				},
			})
			if err != nil {
				logs.Errorf("Json Marshal failed, err:%v", err)
			}
			patchResult, err := gh.groupClient.Patch(context.TODO(), groupCopy.Name, ty.StrategicMergePatchType, patchGroup, metav1.PatchOptions{})
			if err != nil {
				logs.Errorf("Patch group error:%v", err)
			}
			logs.Info("Source CopyInfo:[value:%v]", patchResult.Spec.CopyInfo[groupCopy.Name])
		}
	}
	if copiesInOtherDomain > 0 { // 说明有副本需要部署在其他域---后续可能要添加要求：部署在其他哪个域
		for i := 0; i < int(copiesInOtherDomain); i++ {
			// TODO （需要和调度器确认）发送一个事件通知调度器去选择一个域（不能为本域），事件里面放group信息--我已经生成好副本group了，调度器直接把这个group放到别的域即可
			// 复制创建一个全新的副本group信息（注意Succeed的Phase不用修改，DeployCheck和Running状态需要修改），另外还需要将副本的groupStatus改为Starting
			groupCopyName := "Reason-Copy"                                        // TODO 这里之后改成随机生成即可源group.Name + 一串随机字符
			groupCopy := controller.NewGroupInfoCopy(gr, true, groupCopyName, "") // 第二个参数为true，表示的是提前写入etcd
			go gh.CheckEventForSchedulerResult(gr, groupCopyName)
			gh.recorder.Event(groupCopy, apis.EventTypeNormal, events.SelectOtherDomain, fmt.Sprintf("Need Scheduler to choose the domain to cross"))
		}
		// TODO 这里得让调度器那边发送一个事件给我，我在这监听
	}
	// 任务满足条件，提交给GroupWorkers
	gh.groupWorkers.UpdateGroup(
		&group.UpdateGroupOptions{
			Group:      gr,
			StartTime:  start,
			UpdateType: group.GroupCreate,
		},
	)
	// TODO: 监控任务执行状态的组件
	// TODO: Probe Manager
}

// TODO 目前这块的功能还需要讨论，其实group的信息修改，是否调度器就可以修改，就不用让部署器去修改了，有待商榷
func (gh *GroupHandler) HandleGroupUpdate(gr *apis.Group) {
	// 覆盖Manager中对应的group信息
	gh.groupManager.UpdateGroup(gr)
	// TODO 同时修改etcd当中的group信息
}

func (gh *GroupHandler) HandleGroupKill(gr *apis.Group) {
	start := time.Now()
	logs.Infof("Start HandleGroupKill")
	// 遍历所有的Group,创建Group
	// 向 GroupWorkers 提交任务组的删除请求-hzy

	gh.groupWorkers.UpdateGroup(&group.UpdateGroupOptions{
		Group:      gr,
		StartTime:  start,
		UpdateType: group.GroupKill,
	})
	//gh.groupManager.DeleteGroup(gr)
	//gh.groupClient.Delete(context.TODO(), gr.Name, metav1.DeleteOptions{})
}

// TODO 检查本地资源是否可以启动该Group
func (gh *GroupHandler) checkResource(g *apis.Group) bool {
	// 首先检查一下这个group的状态是否为ReadyToDeploy
	logs.Infof("checkResource方法：g.Status.Phase:%v", g.Status.Phase)
	if g.Status.Phase != apis.ReadyToDeploy {
		return false
	}
	//检查当前节点资源是否满足

	//检查当前节点是否满足Group的条件

	return true
}

// TODO 监听事件：当调度器调决定将group放置在哪个域上的时候，这时候需要往本域的etcd发送一个事件，这样的话我这里如果监听到这个事件，将将副本所在域的连接信息写入到源任务的copyInfo当中
func (gh *GroupHandler) CheckEventForSchedulerResult(gr *apis.Group, copyGroupName string) {
	nowtime := time.Now()
	fieldSelector := fmt.Sprintf("reason=%v", events.ScheduledToOtherDomain)
	watchOptions := meta.ListOptions{
		FieldSelector: fieldSelector,
	}
	watcher, err := gh.eventClient.Watch(context.TODO(), watchOptions)
	if err != nil {
		logs.Errorf("Watch group error:%v", err)
	}
	defer watcher.Stop() // 确保 watcher 被停止
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
				newEvent := event.Object.(*apis.Event)
				if newEvent.InvolvedObject.Name == copyGroupName && newEvent.EventTime.Time.After(nowtime) { //前者晚于后者返回true
					message := newEvent.Message
					// 将跨域的连接写入到源的copyInfo当中
					patchGroup, err := json.Marshal(map[string]interface{}{
						"spec": map[string]interface{}{
							"copy_info": map[string]string{copyGroupName: message},
						},
					})
					if err != nil {
						logs.Errorf("Json Marshal failed, err:%v", err)
					}
					_, err = gh.groupClient.Patch(context.TODO(), gr.Name, ty.StrategicMergePatchType, patchGroup, metav1.PatchOptions{})
					if err != nil {
						logs.Errorf("Patch group error:%v", err)
					}
					return
				}
			default:
				logs.Infof("未识别的事件类型: ", event.Type)
			}
		}
	}
}

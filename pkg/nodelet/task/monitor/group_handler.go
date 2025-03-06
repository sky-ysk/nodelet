package monitor

import (
	"context"
	metav1 "hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/client-go/clients/typed/core"
	"hit.edu/framework/pkg/nodelet/task/controller"
	"time"

	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/component-base/logs"
	"hit.edu/framework/pkg/nodelet/task/group"
	"hit.edu/framework/pkg/nodelet/task/types"
)

type GroupHandler struct {
	// group managers
	groupManager group.Manager

	// group workers, 实际部署任务
	groupWorkers group.GroupWorkers

	groupQueues *group.GroupQueues
	// client -go
	groupClient core.GroupInterface

	stopCh chan struct{}
}

func NewGroupHandler(groupManager group.Manager, groupWorkers group.GroupWorkers, groupQueues *group.GroupQueues, groupClient core.GroupInterface) *GroupHandler {
	return &GroupHandler{
		groupManager: groupManager,
		groupWorkers: groupWorkers,
		groupQueues:  groupQueues,
		groupClient:  groupClient,
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
	const (
		// base = 100000 * time.Millisecond //100s
		base = 100 * time.Millisecond //5s
	)
	logs.Info("GroupHandler component start")
	duration := base

	for {
		go gh.LoopIteration(ctx, updateCh) //是否采用协程，取决于该函数是否要与Loop方法并行执行
		time.Sleep(duration)
		// TODO: 二进制指数退避
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
	// 3、判断group是否需要部署副本，如果需要，在此处往域内的etcd当中添加副本
	if gr.Spec.Replicas > 0 { //如果床架任务的时候该属性没有赋值的话，初始化是为0的
		// 为了适配迁移
		// 复制创建一个全新的副本group信息（注意Succeed的Phase不用修改，DeployCheck和Running状态需要修改），另外还需要将副本的groupStatus改为Starting
		groupCopy := controller.NewGroupInfoCopy(gr, true) // 第二个参数为true，表示的是提前写入etcd
		// 将副本group信息写入到etcd当中，目前还只适配本域内迁移
		logs.Infof("group:%v===================", groupCopy.Name)
		_, err = gh.groupClient.Create(context.TODO(), groupCopy, metav1.CreateOptions{})
		if err != nil {
			logs.Errorf("Create group:%s err: %v", groupCopy.Name, err)
		}
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
	gh.groupManager.DeleteGroup(gr)
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

package monitor

import (
	"context"
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

	stopCh chan struct{}
}

func NewGroupHandler(groupManager group.Manager, groupWorkers group.GroupWorkers, groupQueues *group.GroupQueues) *GroupHandler {
	return &GroupHandler{
		groupManager: groupManager,
		groupWorkers: groupWorkers,
		groupQueues:  groupQueues,
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
		base = 5000 * time.Millisecond //5s
	)
	logs.Info("GroupHandler begin")
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
	//获取新的新的Group --这是第二种写法
	logs.Info("GroupHandler-loopIteration- begin")
	for {
		select {
		case <-ctx.Done():
			logs.Info("Context canceled, exiting loop iteration")
			return
		case u, ok := <-updateCh:
			if !ok {
				logs.Error("Update channel is closed, exiting loop")
				return
			}
			switch u.Op {
			case types.ADD:
				// TODO: 输出Group的详细信息
				logs.Info("Add new group")
				gh.HandleGroupAdds(u.Groups)
			case types.UPDATE:
				logs.Info("Update group")
				gh.HandleGroupUpdates(u.Groups)
			case types.KILL:
				logs.Info("Delete group")
				gh.HandleGroupKill(u.Groups)
			default:
				panic("unhandled default case")
			}
		}
	}

	// 更新现有的Group信息

}

// 处理Group启动指令  主要内容：检查当前节点是否能执行group
func (gh *GroupHandler) HandleGroupAdds(groups []*apis.Group) {
	// TODO: 对Pod按照优先级排序（目前先按照创建时间） ---应该是Group吧，目前Group结构体当中好像没有创建时间或者优先级这个参数，不好做排序

	start := time.Now()
	// 遍历所有的Group,将Group对象存入group_manager
	for _, g := range groups {
		// TODO: 检查任务是否可以在当前节点上运行, 如果不能，则拒绝Pod的部署
		//  不能部署的情况包括
		//  1、Group已经在本地部署（是否包括副本）？
		//  2、没有可以执行的资源
		//  ......
		_, err := gh.groupManager.GetGroupByName(g.Name)
		if err == nil { //err等于nil说明在group_manager当中能找到group信息
			// 1、说明group已经存在，且副本数小于0，则拒绝部署，并记录日志
			if g.Spec.Replicas <= 0 {
				logs.Errorf("Group %s is already deployed")
				continue
			}
			//这里可能还得检查，当前任务的部署数量是否小于期望数量，如果数量大于期望的副本数，也不再部署
		}
		// 2、检查资源是否足够并满足部署条件
		canDeploy, err1 := gh.checkResource(g)
		if err1 != nil || !canDeploy {
			// 如果无法部署，拒绝改Group的部署并记录日志
			logs.Errorf("Group %s cannot be deployed: %s\n", g.Name, err)
		}
		// 向 GroupWorkers 提交任务组创建的请求-hzy
		if canDeploy {
			//group_manager 存入group信息
			gh.groupManager.AddGroup(g)
			// 任务满足条件，提交给GroupWorkers
			gh.groupWorkers.UpdateGroup(
				&group.UpdateGroupOptions{
					Group:      g,
					StartTime:  start,
					UpdateType: group.GroupCreate,
				},
			)
		}
	}
	// TODO: 监控任务执行状态的组件
	// TODO: Probe Manager
}

func (gh *GroupHandler) HandleGroupUpdates(groups []*apis.Group) {
	// 覆盖Manager中对应的group信息
	for _, g := range groups {
		gh.groupManager.UpdateGroup(g)
	}
}

func (gh *GroupHandler) HandleGroupKill(groups []*apis.Group) {
	start := time.Now()
	logs.Infof("Start HandleGroupKill")
	// 遍历所有的Group,创建Group
	for _, g := range groups {
		// 向 GroupWorkers 提交任务组的删除请求-hzy
		gh.groupWorkers.UpdateGroup(&group.UpdateGroupOptions{
			Group:      g,
			StartTime:  start,
			UpdateType: group.GroupKill,
		})
		gh.groupManager.DeleteGroup(g)
	}
}

func (gh *GroupHandler) checkResource(g *apis.Group) (bool, error) {
	//检查当前节点资源是否满足

	//检查当前节点是否满足Group的条件

	return true, nil
}

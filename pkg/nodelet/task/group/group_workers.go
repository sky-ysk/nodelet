package group

import (
	"context"
	metav1 "hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/client-go/clients/typed/core"
	"hit.edu/framework/pkg/nodelet/task/task"
	"strings"
	"sync"
	"time"

	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/component-base/logs"
	"hit.edu/framework/pkg/nodelet/task/runtime"
)

// 负责对Group的操作进行管理和调度，比如创建、更新、删除、终止等-hzy
type GroupWorkers interface {
	// 更新Group
	UpdateGroup(options *UpdateGroupOptions)
}

type UpdateGroupType int

const (
	GroupCreate UpdateGroupType = iota
	GroupUpdate
	GroupDelete
	GroupKill //可能暂时没用
)

// 更新Group中的选项内容
type UpdateGroupOptions struct {
	// Group的开始时间
	StartTime time.Time

	// Group To Update
	Group *apis.Group

	//对Group进行了什么操作-标记一下-hzy
	UpdateType UpdateGroupType
}

type groupWorkers struct {
	//锁
	groupLock sync.RWMutex

	// 存储所有Group的GoRoutines
	groupUpdates map[string]chan *UpdateGroupOptions

	//管理所有Group
	groupManager Manager

	//管理所有的Task
	taskManager task.Manager

	//以队列的形式管理监控group，实时反馈给aip-server各个group的状态
	queueManager *GroupQueues

	//group-client
	groupsClient core.GroupInterface
	//task-client
	taskClient core.TaskInterface

	// 管理运行所需的Runtime
	// 存储RuntimeManager
	runtimeManager *runtime.RuntimeManager
}

func NewGroupWorkers(groupManager Manager, taskManager task.Manager, groupQueues *GroupQueues, runtimeManager *runtime.RuntimeManager, groupclient core.GroupInterface, taskclient core.TaskInterface) GroupWorkers {
	//TODO:
	return &groupWorkers{
		runtimeManager: runtimeManager,
		groupManager:   groupManager,
		taskManager:    taskManager,
		queueManager:   groupQueues,
		groupsClient:   groupclient,
		taskClient:     taskclient,
		groupUpdates:   make(map[string]chan *UpdateGroupOptions),
	}
}

var _ GroupWorkers = &groupWorkers{}

// 设计管道应该是为了让同一时刻对任务的执行操作（添加、更新、删除）仅能执行一次   --但是还有一个问题就是，for循环接收管道内容后就清空管道内容，有可能这时候执行的操作还没结束，如果这时候别的操作发起请求，则也会导致请求重叠---应该不会，如果再发起请求，前面的操作应该已经结束了
func (g *groupWorkers) groupWorkerLoop(groupUpdates <-chan *UpdateGroupOptions) {
	for update := range groupUpdates { //从管道中读取内容，只要管道中有内容，就往下执行，没有内容就卡在for循环上
		switch update.UpdateType {
		case GroupCreate:
			g.startGroup(update.Group)
		case GroupUpdate:
			g.UpdateGroup(update)
		case GroupDelete:
			g.deleteGroup(update.Group)
		case GroupKill:
			g.killGroup(update.Group)
			//这里需要关闭for循环，因为任务结束了，这个任务对应的管道需要被关闭，否则会一直开着
			groupID := update.Group.Status.GroupID // 获取groupID
			g.groupLock.Lock()
			defer g.groupLock.Unlock()
			delete(g.groupUpdates, groupID) // 删除对应的channel
			return
		default:
			panic("unhandled default case")
		}
	}
}

func (g *groupWorkers) startGroup(group *apis.Group) {
	// TODO: 检查Group是否可以运行
	// TODO: 检查依赖
	// 首先先将任务放入到pending队列当中
	success := g.queueManager.AddToPending(group.Status.GroupID, group)
	if !success {
		logs.Error("move group into pending queue failed, because groupID has been in pending queue")
		return
	}
	//修改Pennding队列当中改group的信息（同时也同步到group_manager当中），状态都改为penning
	g.handleGroupPenndingUpdate(group) //12.31新增：除了修改group的状态，还需要修改上层Task的状态为Pennding
	//检查依赖，如果满足，则放入running队列，开始执行actions
	if !g.checkGroupDepencies(group) { //12.31：增加检查是否有父亲group
		logs.Infof("The group %s execution dependency is not satisfied", group.Name)
		//继续放在Pennding队列当中，Pennding队列会持续检查依赖，直到依赖满足后，才开始执行，重新将任务group交给group_workers去执行
		return
	}
	// 任务依赖满足后就将任务从Pennding队列转移纸Running队列
	ok := g.queueManager.DeleteFromPending(group.Status.GroupID)
	if !ok {
		logs.Error("delete group from pending queue failed")
	}
	ok = g.queueManager.AddToRunning(group.Status.GroupID, group)
	if !ok {
		logs.Error("move group to running queue failed")
	}
	//此处不用再修改group信息为Running，真正启动任务的时候，会修改phase为running
	// TODO: 检查需要运行的Action
	// TODO: 开始部署
	logs.Infof("start group %s", group.Name)
	//根据group当中的Action开启相应的runtime  group(Spec:Actions)--action（Spec：Runtimes）
	for i := range group.Spec.Actions {
		action := &group.Spec.Actions[i]
		if !g.checkActionDependencies(action, group) {
			logs.Infof("Action %s in group %s waiting for dependencies", action.Name, group.Name)
			action.Status.Waiting = true //第一次执行时发现执行不了，那就交给running队列去检查
			continue
		}
		for j := range action.Spec.Runtimes {
			ru := &action.Spec.Runtimes[j]
			if !g.checkRuntimeDepencies(ru, action) {
				logs.Infof("Runtime %s in group %s waiting for dependencies", ru.Name, group.Name)
				ru.Waiting = true //第一次执行时发现执行不了，那就交给running队列去检查，检查成功才执行
				continue
			}
			err := g.runtimeManager.Run(group, action, ru)
			if err != nil {
				logs.Error("run task err:", err.Error())
			}
		}
	}
}

// 这个groupUpdates-Channel是为了同一时刻，任务只能对应一个操作（增加、删除或者更新），管道满的话说明该Group在进行别的操作当中
// 每个Group的操作通过独立的goroutine管理，通过channel通知Group的更新并执行相应的操作。-hzy
func (g *groupWorkers) UpdateGroup(options *UpdateGroupOptions) {
	g.groupLock.Lock()
	defer g.groupLock.Unlock()
	// 查看GroupUpdate Channel是否存在 ---问题：这里是根据ID来找协程，如果名字重复了怎么办--ID是不会重复的 -解决
	// FIXME: 这里应该换成GroupID，GroupID是Group创建后，由系统分配的唯一的ID

	groupID := options.Group.Status.GroupID
	groupUpdates, exists := g.groupUpdates[groupID] //后期最好将group_workers当中的groupUpdates这个map进行清理（对于已经执行完的group，删除信息）
	if !exists {
		groupUpdates = make(chan *UpdateGroupOptions, 1)
		g.groupUpdates[groupID] = groupUpdates

		go func() {
			g.groupWorkerLoop(groupUpdates) //对每一个任务只开这一个协程，注意了，这个bug找了很久
		}()
	}
	// 通知更新
	select {
	case groupUpdates <- options: //往管道中放入options
		logs.Infof("Group %s: signal sent for %v", groupID, options.UpdateType)
	default:
		logs.Warnf("Group %s: update signal skipped (channel busy)", groupID)
	}
}

// 对于正常完成的group在更新完group_status之后进行delete操作
func (g *groupWorkers) deleteGroup(group *apis.Group) {
	g.queueManager.DeleteGroup(group) //删除group_manager和queue_manager当中的任务
}

func (g *groupWorkers) killGroup(group *apis.Group) {
	if g.runtimeManager == nil {
		logs.Error("runtimeManager is nil")
	}
	for i := range group.Spec.Actions {
		action := &group.Spec.Actions[i]
		for j := range action.Spec.Runtimes {
			ru := &action.Spec.Runtimes[j]
			if action.Status.RuntimeStatus[j].Phase == apis.Successed {
				// runtime已经执行完成，不用再kill了
				continue
			}
			err := g.runtimeManager.Kill(group, action, ru)
			if err != nil {
				logs.Error("kill task err:", err.Error())
			}
		}
	}
}

// 检查Group的依赖是否满足
func (g *groupWorkers) checkGroupDepencies(group *apis.Group) bool {
	//TODO：实现依赖检查逻辑
	if len(group.Spec.Parents) == 0 {
		return true
	} else {
		//检查父亲group是否执行完成
		for i := range group.Spec.Parents {
			parentName := group.Spec.Parents[i]
			// 去client-go当中查group
			result, err := g.groupsClient.Get(context.TODO(), parentName, metav1.GetOptions{})
			if err != nil {
				logs.Errorf("Failed to get group: %s", parentName)
			}
			if result.Status.Phase != apis.Successed {
				return false //说明当前group的付钱group还没完成，直接返回false即可
			}
		}
	}
	return true
}

// 检查Action的依赖是否满足
func (g *groupWorkers) checkActionDependencies(action *apis.Action, group *apis.Group) bool {
	for _, parentName := range action.Spec.Parents {
		for _, ac := range group.Spec.Actions {
			if parentName == ac.Name && ac.Status.Phase != apis.Successed {
				return false
			}
		}
	}
	return true
}

// 检查Runtime的依赖是否满足
func (g *groupWorkers) checkRuntimeDepencies(runtime *apis.Runtime, action *apis.Action) bool {
	//TODO runtime运行之前，需要检查parent的runtime是否正常执行完成
	for _, parentName := range runtime.Parents {
		for _, rs := range action.Status.RuntimeStatus {
			if strings.HasPrefix(rs.RuntimeID, parentName) &&
				rs.Phase != apis.Successed {
				return false
			}
		}
	}
	return true
}

// 修改group下面的所有状态为Pennding  +增加：修改group上层的Task状态为Pennding
func (g *groupWorkers) handleGroupPenndingUpdate(group *apis.Group) {
	// 修改Group上层的Task 的Status状态为penning
	g.handleTaskPenndingUpdate(group)
	// 修改Group层以及Group下面的Action、Runtime的Phase为penning
	groupSpec := &group.Spec
	groupStatus := &group.Status
	//首先标记GroupStatus的Phase为Pennding
	group.Status.Phase = apis.ReadyToDeploy

	// GroupSpec当中的Actions，标记ActionStatus中状态为Pennding
	for i := range groupSpec.Actions { //Actions
		actionStatus := &groupSpec.Actions[i].Status //ActionStatus
		groupSpec.Actions[i].Status.Phase = apis.ReadyToDeploy
		for j := range actionStatus.RuntimeStatus { // RuntimeStatus
			actionStatus.RuntimeStatus[j].Phase = apis.ReadyToDeploy
		}
	}
	//GroupStatus当中的ActionStatus
	for i := range groupStatus.ActionStatus { //ActionStatus
		as := &groupStatus.ActionStatus[i]
		as.Phase = apis.ReadyToDeploy
		for j := range as.RuntimeStatus { //RuntimeStatus
			rs := &as.RuntimeStatus[j]
			rs.Phase = apis.ReadyToDeploy
		}
	}
	//将queue_manager和group_manager的group信息进行更新 ----------有问题
	err := g.queueManager.UpdateGroup(group.Status.GroupID, group)
	if err != nil {
		logs.Error("update group-runtiem-start info error")
	}
}

// 将TaskStatus的phase设置为Pennding， TODO 后续可能还需要修改时间
func (g *groupWorkers) handleTaskPenndingUpdate(group *apis.Group) {
	// 查找该group所属的Task
	taskID := group.Status.Belongs.TaskID
	// client-go 查看task-list
	lstOpts := metav1.ListOptions{}
	list, err := g.taskClient.List(context.TODO(), lstOpts)
	if err != nil {
		logs.Error("list task err:", err.Error())
	}
	for _, t := range list.Items { //遍历etcd当中的所有task
		//logs.Infof("%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%task Pointer address: %p", t)

		if t.Status.TaskID == taskID { // 如果taskId对上了，则就修改该Task的Phase为Pennding
			// 得到该Task的引用
			//task, err := g.taskManager.GetTaskByID(taskID)
			//if err != nil {
			//	logs.Error("get task err:", err.Error())
			//}
			// ceshi
			task1, err := g.taskClient.Get(context.TODO(), "TestTasks", metav1.GetOptions{})
			if err != nil {
				logs.Error("get task err:", err.Error())
			}
			//if task == task1 {
			//	logs.Infof("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!task and task2 point to the same memory location.")
			//} else {
			//	logs.Infof("！！！！！！！！！！！！！！！！！！！！！！！！！！task and task2 point to different memory locations.")
			//}
			task1.Status.Phase = apis.ReadyToDeploy //设置Task的状态为penning
			task1.Status.LastTime = apis.Time{time.Now()}
			for i := range task1.Status.GroupStatus { //同时设置TaskStatus下的GroupStatus的phase为penning
				gs := &task1.Status.GroupStatus[i]
				gs.Phase = apis.ReadyToDeploy
				gs.LastTime = apis.Time{time.Now()}
			}
			// 通过client-go，将信息提交到api-server当中
			g.taskClient.Update(context.TODO(), task1, metav1.UpdateOptions{})
		}
	}
}

//for _, action := range group.Spec.Actions {
//	actionParents := action.Spec.Parents //查看Action的父亲action
//	//判断当前action的parents是否执行完成
//	var acParentsSucceed = true
//	for _, parentName := range actionParents { //parents记录的是parentName数组
//		//能执行到里面第一条语句，说明parents数组是有内容的
//		//去找group下面ActionName等于parentName，查看actionStatus的状态
//		for _, ac := range group.Spec.Actions {
//			if parentName == ac.Name {
//				if ac.Status.Phase != apis.Successed {
//					acParentsSucceed = false
//				}
//			}
//		}
//	}
//	if acParentsSucceed { //说明当前要执行的action，其父亲action都已经执行完了，开始执行Runtime
//		runtimes := action.Spec.Runtimes
//		for _, r := range runtimes {
//			//TODO runtime运行之前，需要检查parent的runtime是否正常执行完成
//			runtimeParents := r.Parents //查看Runtime的父亲Parent
//			//判断当前runtime的parents是否执行完成
//			var ruParentsSucceed = true
//			for _, parentName := range runtimeParents {
//				//能执行到里面第一条语句，说明parents数组是有内容的
//				//去找action下面runtime等于parentName，查看runtimeStatus的状态
//				for _, rs := range action.Status.RuntimeStatus {
//					if strings.HasPrefix(rs.RuntimeID, parentName) { //
//						if rs.Phase != apis.Successed {
//							ruParentsSucceed = false
//						}
//					}
//				}
//			}
//			if ruParentsSucceed { //说明当前要执行的runtime，其父亲runtime都已经执行完成了，开始执行当前runtime任务
//				err := g.runtimeManager.Run(group, &action, &r)
//				if err != nil {
//					logs.Error("run task err:", err.Error())
//				}
//			}
//		}
//	}
//}

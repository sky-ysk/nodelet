package group

import (
	"context"
	metav1 "hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/client-go/clients/typed/core"
	"hit.edu/framework/pkg/nodelet/task/task"
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
	GroupKill
	GroupDelete //可能暂时没用

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
	groupClient core.GroupInterface
	//task-client
	taskClient core.TaskInterface
	//action-Client
	actionClient core.ActionInterface

	// 管理运行所需的Runtime
	// 存储RuntimeManager
	runtimeManager *runtime.RuntimeManager
}

func NewGroupWorkers(groupManager Manager, taskManager task.Manager, groupQueues *GroupQueues, runtimeManager *runtime.RuntimeManager, groupclient core.GroupInterface, taskclient core.TaskInterface, actionClient core.ActionInterface) GroupWorkers {
	//TODO:
	return &groupWorkers{
		runtimeManager: runtimeManager,
		groupManager:   groupManager,
		taskManager:    taskManager,
		queueManager:   groupQueues,
		groupClient:    groupclient,
		taskClient:     taskclient,
		actionClient:   actionClient,
		groupUpdates:   make(map[string]chan *UpdateGroupOptions),
	}
}

var _ GroupWorkers = &groupWorkers{}

// 这个groupUpdates-Channel是为了同一时刻，任务只能对应一个操作（增加、删除或者更新），管道满的话说明该Group在进行别的操作当中
// 每个Group的操作通过独立的goroutine管理，通过channel通知Group的更新并执行相应的操作。-hzy
func (g *groupWorkers) UpdateGroup(options *UpdateGroupOptions) {
	g.groupLock.Lock()
	defer g.groupLock.Unlock()

	groupID := options.Group.Status.GroupID
	groupName := options.Group.Name
	groupUpdates, exists := g.groupUpdates[groupID] //后期最好将group_workers当中的groupUpdates这个map进行清理（对于已经执行完的group，删除信息）
	if !exists {
		groupUpdates = make(chan *UpdateGroupOptions, 1)
		g.groupUpdates[groupID] = groupUpdates

		go func() {
			g.groupWorkerLoop(groupUpdates) //对每一个任务只开这一个协程
		}()
	}

	// 通知更新
	select {
	case groupUpdates <- options: //往管道中放入options
		logs.Debugf("Group:%s signal sent for %v(0:create;1:update;2:kill)", groupName, options.UpdateType)
	default:
		logs.Warnf("Group:%s update signal skipped (channel busy)", groupName)
	}
}

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
			logs.Infof("kill 2...")
			g.killGroup(update.Group)
		default:
			logs.Error("Unhandled default case")
		}
	}
}

func (g *groupWorkers) startGroup(gr *apis.Group) {
	// TODO: 检查Group的运行依赖
	// 当Group bind到当前节点时，开始部署当前Group
	// Group中包含多个Action,Action支持串行和并行执行
	// 任务部署前，需要检查任务的依赖，需要检查的内容包括
	//   资源依赖，任务所需计算、网络、存储或者硬件资源是否就绪
	//   顺序依赖，前序节点是否满足
	//   数据依赖，任务执行所需数据是否准备好
	//   条件依赖，任务执行是否满足条件
	// 首先先将任务放入到checking队列当中  ---也就是对应的
	group, err := g.groupClient.Get(context.TODO(), gr.Name, metav1.GetOptions{}) //因为该startGroup方法当中涉及到参数的更新，所以先从etcd中获取一下
	if err != nil {
		logs.Errorf("Get group err:%v", err)
	}
	success := g.queueManager.AddToChecking(group.Status.GroupID, group)
	if !success {
		// 按理来说不会出现这样的情况，为啥呢，因为如果group_handler.go当中的HandleGroupAdd方法只会执行一次
		logs.Error("Move group into checking queue failed, because groupID has been in checking queue")
		return
	}
	//修改Checking队列当中改group的信息（同时也同步到group_manager当中），状态都改为checking
	g.handleCheckingUpdate(group) //12.31新增：除了修改group的状态，还需要修改上层Task的状态为CheckDeploy
	// 最新：把下面的group的依赖检查放入到group_monitor当中的Checking队列检查，把action、runtime的依赖检查放入group_monitor当中的Running队列检查
	//检查依赖，如果满足，则放入running队列，开始执行actions

	// ************************************把下面这段注释
	//if !g.groupDepenSatisfy(group) { //12.31：增加检查是否有父亲group
	//	logs.Infof("The group:%s execution dependency is not satisfied", group.Name)
	//	//继续放在Checking队列当中，Checking队列会持续检查依赖，直到依赖满足后，才开始执行，重新将任务group交给runtimeManager去执行
	//	return
	//}
	//// 任务依赖满足后就将任务从Checking队列转移纸Running队列
	//ok := g.queueManager.DeleteFromCheckingAndAddToRunning(group.Status.GroupID)
	//if !ok {
	//	logs.Error("Delete group from checking queue and add to running queue failed")
	//}
	// ************************************

	//此处不用再修改group信息为Running，真正启动任务的时候，会修改phase为running
	// TODO: 检查需要运行的Action,开始部署
	//logs.Infof("Ready to start group:%s", group.Name)
	//根据group当中的Action开启相应的runtime  group(Spec:Actions)--action（Spec：Runtimes）
	// 得有一个变量来标记group里的信息是否改变，如果没有改变就不用上传到etcd当中了，因为相同的group应该不能调用update

	// TODO 将下面这段代码合并到了group_monitor当中，还需要测试一下
	//for i := range group.Spec.Actions {
	//	action := &group.Spec.Actions[i]
	//	if !g.actionDepenSatisfy(i, group) {
	//		logs.Infof("Action:%s in group:%s waiting for dependencies", action.Name, group.Name)
	//		//action.Status.Waiting = true //第一次执行时发现执行不了，那就交给running队列去检查
	//		gro, err := g.groupManager.GetGroupByID(group.Status.GroupID)
	//		if err != nil {
	//			logs.Errorf("Get group from group_manager component err:%v", err)
	//		}
	//		gro.Spec.Actions[i].Status.Waiting = true
	//		logs.Debugf("StartGroup method:ActionWaiting:%v, i:%v", gro.Spec.Actions[i].Status.Waiting, i)
	//		//// patch
	//		//patchGroupActions, err4 := json.Marshal(map[string]interface{}{
	//		//	"spec": map[string]interface{}{
	//		//		"actions": group.Spec.Actions,
	//		//	},
	//		//})
	//		//if err4 != nil {
	//		//	logs.Errorf("json marshal:patchGroupActions err:%v", err)
	//		//}
	//		//_, err = g.groupClient.Patch(context.TODO(), group.Name, types.StrategicMergePatchType, patchGroupActions, metav1.PatchOptions{})
	//		//if err != nil {
	//		//	logs.Errorf("patch patchGroupActions:group err:%v", err)
	//		//}
	//		continue
	//	}
	//	for j := range action.Spec.Runtimes {
	//		ru := &action.Spec.Runtimes[j]
	//		if !g.runtimeDepenSatisfy(i, j, group) {
	//			logs.Infof("Runtime:%s in group:%s waiting for dependencies", ru.Name, group.Name)
	//			//ru.Waiting = true //第一次执行时发现执行不了，那就交给running队列去检查，检查成功才执行
	//			gro, err := g.groupManager.GetGroupByID(group.Status.GroupID)
	//			if err != nil {
	//				logs.Errorf("Get group from group_manager err:%v", err)
	//			}
	//			gro.Spec.Actions[i].Spec.Runtimes[j].Waiting = true
	//			logs.Debugf("StartGroup method:runtim:%v, Runtimewaiting:%v, i:%v, j:%v", ru.Name, gro.Spec.Actions[i].Spec.Runtimes[j].Waiting, i, j)
	//			// TODO patch
	//			//logs.Infof("runtime %s in group, waiting:%v,i:%v,j:%v", ru.Name, group.Spec.Actions[i].Spec.Runtimes[j].Waiting, i, j)
	//			//// patch
	//			//patchGroupActionsRuntimes, err4 := json.Marshal(map[string]interface{}{
	//			//	"spec": map[string]interface{}{
	//			//		"actions": group.Spec.Actions,
	//			//	},
	//			//})
	//			//if err4 != nil {
	//			//	logs.Errorf("json marshal:patchGroupActions err:%v", err)
	//			//}
	//			//result, err := g.groupClient.Patch(context.TODO(), group.Name, types.StrategicMergePatchType, patchGroupActionsRuntimes, metav1.PatchOptions{})
	//			//if err != nil {
	//			//	logs.Errorf("patch patchGroupActionsRuntimes:group err:%v", err)
	//			//}
	//			//logs.Info(result)
	//			continue
	//		}
	//		logs.Infof("Run runtime, runtime:%v", ru.Name)
	//		go g.runtimeManager.Run(group, action, ru, i, j) //TODO 考虑这个方法是否使用协程
	//		if err != nil {
	//			logs.Errorf("Run task err:%v", err)
	//		}
	//	}
	//}

	//for i := range group.Status.ActionStatus {
	//	actionStatus := group.Status.ActionStatus[i]
	//	if !g.actionDepenSatisfy(i, group) {
	//		logs.Infof("ActionID %s in group %s waiting for dependencies", actionStatus.ActionID, group.Name)
	//		actionStatus.Waiting = true
	//		groupIsModified = true
	//		continue
	//	}
	//	for j := range actionStatus.RuntimeStatus {
	//		runtimeStatus := actionStatus.RuntimeStatus[j]
	//		if !g.runtimeDepenSatisfy(i, j, group) {
	//			logs.Infof("RuntimeID %s in group %s waiting for dependencies", runtimeStatus.RuntimeID, group.Name)
	//			runtimeStatus.Waiting = true
	//			logs.Infof("runtimeID %s in group, waiting is true", runtimeStatus.RuntimeID)
	//			continue
	//		}
	//		logs.Infof("Run runtimeID:%v", runtimeStatus.RuntimeID)
	//		go g.runtimeManager.Run(group, &group.Spec.Actions[i], &group.Spec.Actions[i].Spec.Runtimes[j], i, j)
	//	}
	//}

	// 将group上传到etcd当中
	//if groupIsModified {
	//	logs.Infof("runtime info updating")
	//	logs.Infof("runtime.Waiting:%v", group.Spec.Actions[0].Spec.Runtimes[1].Waiting)
	//	_, err = g.groupClient.Update(context.TODO(), group, metav1.UpdateOptions{})
	//	if err != nil {
	//		logs.Errorf("update group to etcd err:%v=====123", err)
	//	}
	//	// 也上传一份到group_manager当中
	//	err = g.queueManager.UpdateGroup(group.Status.GroupID, group)
	//	if err != nil {
	//		logs.Errorf("update group to group_manager err:%v", err)
	//	}
	//}
}

// 对于正常完成的group在更新完group_status之后进行delete操作--目前该方法暂未考虑 1.4
func (g *groupWorkers) deleteGroup(group *apis.Group) {
	g.queueManager.DeleteGroup(group) //删除group_manager和queue_manager当中的任务
}

func (g *groupWorkers) killGroup(group *apis.Group) {
	if g.runtimeManager == nil {
		logs.Error("RuntimeManager is nil")
	}

	for i := range group.Spec.Actions {
		logs.Infof("action %v", group.Spec.Actions)
		action := &group.Spec.Actions[i]
		for j := range action.Spec.Runtimes {
			ru := &action.Spec.Runtimes[j]
			if action.Status.RuntimeStatus[j].Phase == apis.Successed || action.Status.RuntimeStatus[j].Phase == apis.DeployCheck || action.Status.RuntimeStatus[j].Phase == apis.Killed {
				// runtime已经执行完成，不用再kill了
				continue
			}
			err := g.runtimeManager.Kill(group, action, ru, i, j)
			if err != nil {
				logs.Errorf("Kill task err:%v", err)
			}
		}
	}
	//这里需要关闭for循环，因为任务结束了，这个任务对应的管道需要被关闭，否则会一直开着
	groupID := group.Status.GroupID // 获取groupID
	g.groupLock.Lock()
	defer g.groupLock.Unlock()
	delete(g.groupUpdates, groupID) // 删除对应的key：group，value：channel
	return
}

// 修改group下面的所有状态为Checking  +增加：修改group上层的Task状态为Checking
func (gw *groupWorkers) handleCheckingUpdate(gr *apis.Group) {
	//gr, err := g.groupClient.Get(context.TODO(), group.Name, metav1.GetOptions{})
	//if err != nil {
	//	logs.Error("get group:%v frm etcd err:", group.Name, err)
	//}
	// 修改Group层以及Group下面的Action、Runtime的Phase为deploychecking
	//times := apis.Time{time.Now()}
	gr.Status.Phase = apis.DeployCheck //首先标记GroupStatus的Phase为DeployCheck
	//gr.Status.LastTime = times  //隐藏
	groupSpec := &gr.Spec
	groupStatus := &gr.Status
	// GroupSpec当中的Actions，需要修改下面的（ActionStatus的Phase以及RuntimeStatus的Phase）
	for i := range groupSpec.Actions { //Actions
		actionStatus := &groupSpec.Actions[i].Status //ActionStatus
		// 为了适配迁移，该Action在A设备上已经执行完成了
		if actionStatus.Phase == apis.Successed {
			continue
		}
		groupSpec.Actions[i].Status.Phase = apis.DeployCheck
		//groupSpec.Actions[i].Status.LastTime = times //隐藏
		for j := range actionStatus.RuntimeStatus { // RuntimeStatus
			// 为了适配迁移，该Runtime在A设备上已经执行完成了
			if actionStatus.RuntimeStatus[j].Phase == apis.Successed {
				continue
			}
			actionStatus.RuntimeStatus[j].Phase = apis.DeployCheck
			//actionStatus.RuntimeStatus[j].LastTime = times // 隐藏
		}
		// 说明Action是第一次启动，这里添加一个操作，将action上传到etcd当中----修改一下改成patch
		action := &groupSpec.Actions[i]
		//action.Status = groupStatus.ActionStatus[i]
		
		_, err := gw.actionClient.Create(context.TODO(), action, metav1.CreateOptions{})
		if err != nil {
			logs.Errorf("Create action failed,err:%v", err)
		}
	}
	//GroupStatus当中的ActionStatus，需要修改（ActionStatus的Phase以及RuntimeStatus的Phase）
	for i := range groupStatus.ActionStatus { //ActionStatus
		as := &groupStatus.ActionStatus[i]
		// 为了适配迁移，该Action在A设备上已经执行完成了
		if as.Phase == apis.Successed {
			continue
		}
		as.Phase = apis.DeployCheck
		//as.LastTime = times // 隐藏
		for j := range as.RuntimeStatus { //RuntimeStatus
			rs := &as.RuntimeStatus[j]
			// 为了适配迁移，该Runtime在A设备上已经执行完成了
			if rs.Phase == apis.Successed {
				continue
			}
			rs.Phase = apis.DeployCheck
			//rs.LastTime = times // 隐藏
		}
	}
	// 修改Group上层的Task 的Status状态为deploychecking
	// 为了适配迁移，副本group在handleCheckingUpdate方法当中无需再将Task的状态设置为DeployChecking，由源任务进行修改
	if !gr.Spec.IsCopy {
		taskID := gr.Status.Belongs.TaskID // 查找该group所属的Task
		// client-go 查看task-list
		list, err := gw.taskClient.List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			logs.Errorf("Get list task err:%v", err)
		}
		for _, t := range list.Items { //遍历etcd当中的所有task，根据taskID取出task
			if t.Status.TaskID == taskID { // 如果taskId对上了，则就修改该Task的Phase为Checking
				taskName := t.Name
				task1, err1 := gw.taskClient.Get(context.TODO(), taskName, metav1.GetOptions{})
				if err1 != nil {
					logs.Errorf("Etcd has group:%v, but not has task:%v, get task err:%v,", gr.Name, taskName, err1)
					return
				}
				logs.Infof("==========================Task的Status.Phase:%v", task1.Status.Phase)
				if task1.Status.Phase == apis.ReadyToDeploy || task1.Status.Phase == apis.Unknown { // TODO 这里为啥要判断是否DeployCheck--因为group被分配到不同的节点上，遍历到group的时候，都需要修改上层Task的信息的话，是重叠的，没必要  这里逻辑错误，如果第一个group遍历到完并且运行了，这里的Task的状态就行Running
					task1.Status.Phase = apis.DeployCheck //首先设置Task的状态为DeployCheck
					logs.Info("=================Task的状态被修改为DeployCheck")
					//task1.Status.LastTime = times //隐藏
				}
				for i := range task1.Spec.Groups { //同时得更新TaskSpec下的Group以及TaskStatus下的GroupStatus为当前的group信息
					if gr.Name == task1.Spec.Groups[i].Name {
						task1.Spec.Groups[i].Spec = gr.Spec
						task1.Spec.Groups[i].Status = gr.Status
						task1.Status.GroupStatus[i] = gr.Status
						break
					}
				}
				// 将task信息提交到etcd上去
				_, err1 = gw.taskClient.Update(context.TODO(), task1, metav1.UpdateOptions{})
				if err1 != nil {
					logs.Errorf("Etcd update task:%v err:%v, now is handing group:%v", taskName, err1, gr.Spec.Name) //这里出错
					// 再次上传
					time.Sleep(200 * time.Millisecond)
					_, err1 = gw.taskClient.Update(context.TODO(), task1, metav1.UpdateOptions{})
				}
				break //后续就不用再遍历Task列表了，直接结束
			}
		}
	}
	//将group信息提交到etcd上去，使用update更新--出现一次报错  TODO 为了适配迁移，如果后面替换为Patch操作，那么要使用gr.ObjectMeta.Name 来进行patch，因为目前规定gr.ObjectMeta.Name为不同group的标识（针对副本、源group）
	_, err := gw.groupClient.Update(context.TODO(), gr, metav1.UpdateOptions{})
	logs.Infof("Group's deployCheck phase submit to etcd, group:%v", gr.Name)
	if err != nil {
		logs.Errorf("Etcd update group:%v err:%v", gr.Name, err)
		// 再次上传
		time.Sleep(200 * time.Millisecond)
		_, err = gw.groupClient.Update(context.TODO(), gr, metav1.UpdateOptions{})
	}
}

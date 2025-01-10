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
		groupClient:    groupclient,
		taskClient:     taskclient,
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
		logs.Infof("Group ID:%s: signal sent for %v(0:create1;1:update;2:kill)", groupID, options.UpdateType)
	default:
		logs.Warnf("Group ID:%s: update signal skipped (channel busy)", groupID)
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
			g.killGroup(update.Group)
		default:
			logs.Error("unhandled default case")
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
		logs.Errorf("get group err:%v", err)
	}
	success := g.queueManager.AddToChecking(group.Status.GroupID, group)
	if !success {
		// 按理来说不会出现这样的情况，为啥呢，因为如果group_handler.go当中的HandleGroupAdd方法只会执行一次
		logs.Error("move group into checking queue failed, because groupID has been in checking queue")
		return
	}

	//修改Checking队列当中改group的信息（同时也同步到group_manager当中），状态都改为checking
	g.handleCheckingUpdate(group) //12.31新增：除了修改group的状态，还需要修改上层Task的状态为CheckDeploy
	//检查依赖，如果满足，则放入running队列，开始执行actions
	if !g.groupDepenSatisfy(group) { //12.31：增加检查是否有父亲group
		logs.Infof("The group %s execution dependency is not satisfied", group.Name)
		//继续放在Checking队列当中，Checking队列会持续检查依赖，直到依赖满足后，才开始执行，重新将任务group交给runtimeManager去执行
		return
	}
	// 任务依赖满足后就将任务从Checking队列转移纸Running队列
	ok := g.queueManager.DeleteFromCheckingAndAddToRunning(group.Status.GroupID)
	if !ok {
		logs.Error("delete group from checking queue and add to running queue failed")
	}

	//此处不用再修改group信息为Running，真正启动任务的时候，会修改phase为running
	// TODO: 检查需要运行的Action,开始部署
	logs.Infof("ready to start group %s", group.Name)
	//根据group当中的Action开启相应的runtime  group(Spec:Actions)--action（Spec：Runtimes）
	for i := range group.Spec.Actions {
		action := &group.Spec.Actions[i]
		if !g.actionDepenSatisfy(i, group) {
			logs.Infof("Action %s in group %s waiting for dependencies", action.Name, group.Name)
			action.Status.Waiting = true //第一次执行时发现执行不了，那就交给running队列去检查
			continue
		}
		for j := range action.Spec.Runtimes {
			ru := &action.Spec.Runtimes[j]
			if !g.runtimeDepenSatisfy(i, j, group) {
				logs.Infof("Runtime %s in group %s waiting for dependencies", ru.Name, group.Name)
				// TODO patch
				ru.Waiting = true //第一次执行时发现执行不了，那就交给running队列去检查，检查成功才执行
				continue
			}
			go g.runtimeManager.Run(group, action, ru, i, j) //TODO 考虑这个方法是否使用协程
			if err != nil {
				logs.Error("run task err:", err.Error())
			}
		}
	}
	// 将group上传到etcd当中，同时顺便也更新一下groupManager
	_, err = g.groupClient.Update(context.TODO(), group, metav1.UpdateOptions{})
	if err != nil {
		logs.Errorf("update group to etcd err:%v", err)
	}
	// 也上传一份到group_manager当中
	err = g.queueManager.UpdateGroup(group.Status.GroupID, group)
	if err != nil {
		logs.Errorf("update group to group_manager err:%v", err)
	}
}

// 对于正常完成的group在更新完group_status之后进行delete操作--目前该方法暂未考虑 1.4
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
	//这里需要关闭for循环，因为任务结束了，这个任务对应的管道需要被关闭，否则会一直开着
	groupID := group.Status.GroupID // 获取groupID
	g.groupLock.Lock()
	defer g.groupLock.Unlock()
	delete(g.groupUpdates, groupID) // 删除对应的key：group，value：channel
	return
}

// 检查Group的依赖是否满足
func (g *groupWorkers) groupDepenSatisfy(group *apis.Group) bool {
	// TODO：实现依赖检查逻辑
	// 目前只是检查Spec当中的Parents选项
	if len(group.Spec.Parents) == 0 {
		return true
	} else {
		//检查父亲group是否执行完成
		for i := range group.Spec.Parents {
			parentGroupID := group.Spec.Parents[i]
			var parentGroupName string
			// 去client-go当中查group
			groupList, err := g.groupClient.List(context.TODO(), metav1.ListOptions{})
			if err != nil {
				logs.Error("get list group from etcd err:", err.Error())
			}
			for _, g := range groupList.Items {
				if g.Status.GroupID == parentGroupID {
					parentGroupName = g.Name
					break
				}
			}
			result, err := g.groupClient.Get(context.TODO(), parentGroupName, metav1.GetOptions{}) //这里查父亲group的状态，得去etcd当中查
			if err != nil {
				logs.Errorf("Failed to get parent group: %s form etcd, err:%v", parentGroupName, err)
			}
			if result.Status.Phase != apis.Successed {
				return false //说明当前group的付钱group还没完成，直接返回false即可
			}
		}
	}
	return true
}

// 检查Action的依赖是否满足
func (g *groupWorkers) actionDepenSatisfy(actionIndex int, group *apis.Group) bool {
	// 得去etcd当中查稳妥一些，还是查action的parents是否完成---已经不用读了，参数传入的已经是从etcd当中读出来的了
	//group, err := g.groupClient.Get(context.TODO(), groupName, metav1.GetOptions{})
	//if err != nil {
	//	logs.Errorf("get group err:%v", err)
	//}
	actionSpec := &group.Spec.Actions[actionIndex].Spec
	for i := range actionSpec.Parents { // 遍历当前Action的所有父亲Action
		actionParentID := actionSpec.Parents[i]
		for j := range group.Status.ActionStatus { // 遍历group当中所有的action，先对actionID，然后看这个action的Phase如何
			as := &group.Status.ActionStatus[j]
			if as.ActionID == actionParentID && as.Phase != apis.Successed { //目前定义，Action的父亲Action必须是成功状态
				return false
			}
		}
	}
	return true
}

// 检查Runtime的依赖是否满足
func (g *groupWorkers) runtimeDepenSatisfy(actionIndex, runtimeIndex int, group *apis.Group) bool {
	//TODO runtime运行之前，需要检查parent的runtime是否正常执行完成
	// 得去etcd当中查稳妥一些，还是查action的parents是否完成 ---已经不用读了，参数传入的已经是从etcd当中读出来的了
	//group, err := g.groupClient.Get(context.TODO(), groupName, metav1.GetOptions{})
	//if err != nil {
	//	logs.Errorf("get group err:%v", err)
	//}
	runtime := &group.Spec.Actions[actionIndex].Spec.Runtimes[runtimeIndex]
	for i := range runtime.Parents { // 遍历当前runtime的父亲
		runtimeParentID := runtime.Parents[i]
		for j := range group.Status.ActionStatus[actionIndex].RuntimeStatus {
			rs := &group.Status.ActionStatus[actionIndex].RuntimeStatus[j]
			if rs.RuntimeID == runtimeParentID && rs.Phase != apis.Successed {
				return false
			}
		}
	}
	return true
}

// 修改group下面的所有状态为Checking  +增加：修改group上层的Task状态为Checking
func (g *groupWorkers) handleCheckingUpdate(gr *apis.Group) {
	//gr, err := g.groupClient.Get(context.TODO(), group.Name, metav1.GetOptions{})
	//if err != nil {
	//	logs.Error("get group:%v frm etcd err:", group.Name, err)
	//}
	// 修改Group层以及Group下面的Action、Runtime的Phase为deploychecking
	time := apis.Time{time.Now()}
	gr.Status.Phase = apis.DeployCheck //首先标记GroupStatus的Phase为DeployCheck
	gr.Status.LastTime = time
	groupSpec := &gr.Spec
	groupStatus := &gr.Status
	// GroupSpec当中的Actions，需要修改下面的（ActionStatus的Phase以及RuntimeStatus的Phase）
	for i := range groupSpec.Actions { //Actions
		actionStatus := &groupSpec.Actions[i].Status //ActionStatus
		groupSpec.Actions[i].Status.Phase = apis.DeployCheck
		groupSpec.Actions[i].Status.LastTime = time
		for j := range actionStatus.RuntimeStatus { // RuntimeStatus
			actionStatus.RuntimeStatus[j].Phase = apis.DeployCheck
			actionStatus.RuntimeStatus[j].LastTime = time
		}
	}
	//GroupStatus当中的ActionStatus，需要修改（ActionStatus的Phase以及RuntimeStatus的Phase）
	for i := range groupStatus.ActionStatus { //ActionStatus
		as := &groupStatus.ActionStatus[i]
		as.Phase = apis.DeployCheck
		as.LastTime = time
		for j := range as.RuntimeStatus { //RuntimeStatus
			rs := &as.RuntimeStatus[j]
			rs.Phase = apis.DeployCheck
			rs.LastTime = time
		}
	}
	// 修改Group上层的Task 的Status状态为deploychecking
	taskID := gr.Status.Belongs.TaskID // 查找该group所属的Task
	// client-go 查看task-list
	list, err := g.taskClient.List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		logs.Errorf("get list task err:%v", err)
	}
	for _, t := range list.Items { //遍历etcd当中的所有task，根据taskID取出task
		if t.Status.TaskID == taskID { // 如果taskId对上了，则就修改该Task的Phase为Checking
			taskName := t.Name
			task1, err1 := g.taskClient.Get(context.TODO(), taskName, metav1.GetOptions{})
			if err1 != nil {
				logs.Errorf("etcd has group: %v, but not has task:%v, get task err: %v,", gr.Name, taskName, err1)
				return
			}
			if task1.Status.Phase != apis.DeployCheck { // TODO 这里为啥要判断是否DeployCheck--因为group被分配到不同的节点上，遍历到group的时候，都需要修改上层Task的信息的话，是重叠的，没必要
				task1.Status.Phase = apis.DeployCheck //首先设置Task的状态为DeployCheck
				task1.Status.LastTime = time
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
			_, err1 = g.taskClient.Update(context.TODO(), task1, metav1.UpdateOptions{})
			if err1 != nil {
				logs.Errorf("etcd update task: %v err:%v", taskName, err1)
			}
			break //后续就不用再遍历Task列表了，直接结束
		}
	}
	//将group信息提交到etcd上去
	_, err = g.groupClient.Update(context.TODO(), gr, metav1.UpdateOptions{})
	if err != nil {
		logs.Errorf("etcd update group: %v err:%v", gr.Name, err)
	}
	//将queue_manager和group_manager的group信息进行更新
	err = g.queueManager.UpdateGroup(gr.Status.GroupID, gr)
	if err != nil {
		logs.Error("update group-DeployChecking info to queue_manager、group_manager error")
	}
}

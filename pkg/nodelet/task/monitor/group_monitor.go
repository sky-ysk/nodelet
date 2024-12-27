package monitor

import (
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/component-base/logs"
	"hit.edu/framework/pkg/nodelet/events"
	"hit.edu/framework/pkg/nodelet/events/eventbus"
	group "hit.edu/framework/pkg/nodelet/task/group"
	"hit.edu/framework/pkg/nodelet/task/runtime"
	"reflect"
	"strings"
	"time"
)

// /* TODO
// 监控任务的执行状态
// 1. 任务相关的进程是否还在
// 2. 实现了特定接口的任务，调用接口查看任务进度
// 3. 进程占用的资源情况

// //TODO: 跟任务相关的接口
// //调用grpc接口获取任务执行状态

type GroupMonitor struct {
	// group managers 存储任务信息
	groupManager group.Manager
	// 存储任务队列
	groupQueues *group.GroupQueues
	eventBus    *eventbus.EventBus
	// 管理运行所需的Runtime
	// 存储RuntimeManager
	runtimeManager *runtime.RuntimeManager
	stopCh         chan struct{}
}

func NewGroupMonitor(groupManager group.Manager, groupQueues *group.GroupQueues, eventbus *eventbus.EventBus, runtimeManager *runtime.RuntimeManager) *GroupMonitor {
	return &GroupMonitor{
		groupManager:   groupManager,
		groupQueues:    groupQueues,
		eventBus:       eventbus,
		runtimeManager: runtimeManager,
		stopCh:         make(chan struct{}),
	}
}

func (gmo *GroupMonitor) Start() {
	//TODO 轮询检查队列当中的内容
	logs.Info("GroupMonitor begin")

	//订阅事件
	chRuntimeStart := make(chan interface{})
	chRuntimeEnd := make(chan interface{})

	gmo.eventBus.Subscribe(reflect.TypeOf(events.RuntimeStartPhaseEvent{}), chRuntimeStart)
	gmo.eventBus.Subscribe(reflect.TypeOf(events.RuntimeEndPhaseEvent{}), chRuntimeEnd)
	//启动监听事件
	go func() {
		for {
			select {
			case event := <-chRuntimeStart:
				RuntimeEvent := event.(events.RuntimeStartPhaseEvent)
				gmo.handleRuntimeStartUpdate(RuntimeEvent)
			case event := <-chRuntimeEnd:
				RuntimeEvent := event.(events.RuntimeEndPhaseEvent)
				gmo.handleRuntimeEndUpdate(RuntimeEvent)
			}
		}
	}()
	gmo.CheckStatus()

}

func (gm *GroupMonitor) Stop() {
	close(gm.stopCh)
	logs.Info("TaskExporter Monitor stopped")
}

//func (gmo *GroupMonitor) PeriodicallyCheckStatus(interval time.Duration) {
//	ticker := time.NewTicker(interval)
//	go func() {
//		for {
//			select {
//			case <-ticker.C: //ticker.C是一个通道，当ticker定时器每隔指定的时间向通道中写入数据
//				gmo.CheckStatus()
//			case <-gmo.stopCh:
//				ticker.Stop()
//				return
//			}
//		}
//	}()
//}

// 检查pennding队列的任务，前置任务是否完成，看是否需要迁移到running队列
// 检查running队列，看任务是否还在执行、是否执行完成、、进程占用资源量    任务完成和任务失败下线该如何判断呢？
// 检查error队列，检查任务是否出错，看能否尝试拉起，多次尝试拉起失败后，重新提交给调度器
func (gmo *GroupMonitor) CheckStatus() {
	go gmo.PendingCheck()
	go gmo.RunningCheck()
	go gmo.CompletedCheck()
	go gmo.ErrorCheck()
}

func (gmo *GroupMonitor) PendingCheck() {
	//// TODO 轮询检查Pending队列，检查任务group的依赖是否满足，如果满足才放入running队列当中
	//for {
	//	logs.Info("pending queue checking")
	//	select {
	//	case <-time.After(time.Second * 1):
	//		penndingGroups := gmo.groupQueues.GetAllPending()
	//		logs.Infof("Pending queue 的group数量：%v", len(penndingGroups))
	//		for _, task := range penndingGroups {
	//			if !gmo.checkGroupDepencies(task) { //再次检查group的执行依赖是否满足了（注意：group_workers当中任务头一次执行前也会检查）
	//				logs.Infof("The group %s execution dependency is not satisfied again", task.Name)
	//				//继续放在Pennding队列当中，Pennding队列会持续检查依赖，直到依赖满足后，才开始执行，重新将任务group交给group_workers去执行
	//				task.Status.CheckDependencyCount++
	//				if task.Status.CheckDependencyCount > 3 { // 当检查依赖的次数大于3次的话，说明依赖还是满足不了，迁移至Error队列
	//					//将任务迁移到Error队列当中
	//					gmo.groupQueues.DeleteFromPending(task.Status.GroupID)
	//					gmo.groupQueues.AddToError(task.Status.GroupID, task)
	//					// TODO 修改queue_manager、group_manager当中的group信息
	//					break
	//				}
	//				continue
	//			} else { //说明group执行的依赖已经满足，接下来开始执行
	//				// 任务依赖满足后就将任务从Pennding队列转移至Running队列
	//				logs.Info("进入else分支---------------------------")
	//				gmo.groupQueues.DeleteFromPending(task.Status.GroupID)
	//				gmo.groupQueues.AddToRunning(task.Status.GroupID, task)
	//				//此处不用再修改group信息为Running，真正启动任务的时候，会修改phase为running
	//				// TODO: 检查需要运行的Action
	//				// TODO: 开始部署
	//				logs.Infof("start group %s", task.Name)
	//				//根据group当中的Action开启相应的runtime  group(Spec:Actions)--action（Spec：Runtimes）
	//				for _, action := range task.Spec.Actions {
	//					if !gmo.checkActionDependencies(&action, task) {
	//						logs.Infof("Action %s in group %s waiting for dependencies", action.Name, task.Name)
	//						continue
	//					}
	//					for _, ru := range action.Spec.Runtimes {
	//						if !gmo.checkRuntimeDepencies(&ru, &action) {
	//							logs.Infof("Runtime %s in group %s waiting for dependencies", ru.Name, task.Name)
	//							continue
	//						}
	//						err := gmo.runtimeManager.Run(task, &action, &ru)
	//						if err != nil {
	//							logs.Error("run task err:", err.Error())
	//						}
	//					}
	//				}
	//			}
	//		}
	//	}
	//}
}

// 检查Running队列，做的事情：①如果发现任务完成，迁移到Completed队列，如果发现任务失败，迁移到Error队列
// ②检查runtime、Action当中的parents是否执行完成，如果完成，则执行
func (gmo *GroupMonitor) RunningCheck() {
	//TODO 监控进程的返回值等判断任务是否正常执行完成，正常则放入completedqueue，否则放入errorqueue(方法待确认)
	logs.Info("running queue checking")
	for {
		select {
		case <-time.After(time.Second * 1):
			runningGroups := gmo.groupQueues.GetAllRunning()
			for _, task := range runningGroups {
				var isSuccess bool
				for _, action := range task.Spec.Actions { //这里改成task.Status下面的Actions
					isSuccess = false
					if action.Status.Phase == apis.Successed {
						isSuccess = true
						continue //说明当前Action执行完成了，接着查看下一个Action的执行情况
					}
					if action.Status.Phase == apis.Failed { //注意:runtime执行失败的时候除了标记Runtime状态为失败，也需要标记Runtime所属的Action状态为失败
						//将任务迁移到Error队列当中
						gmo.groupQueues.DeleteFromRunning(task.Status.GroupID)
						gmo.groupQueues.AddToError(task.Status.GroupID, task)
						break
					}
					if action.Status.Phase == apis.Running {
						continue
					}
					//检查runtime、Action当中的parents是否执行完成，如果父亲节点完成，则让他执行 ---这里有bug，就是任务已经放入running队列，但是还没执行完，这时候runningCheck循环遍历到当前runtime的状态为Pennding，查看是否满足执行条件，发现是满足的，结果有跑起来该任务
					if action.Status.Phase == apis.Pending && action.Status.IsWaiting {
						if !gmo.checkActionDependencies(&action, task) {
							logs.Infof("Action %s depends on parent action, parent not finish ", action.Name)
							continue
						}
						for _, r := range action.Spec.Runtimes {
							if !gmo.checkRuntimeDepencies(&r, &action) {
								logs.Infof("Runtime %s depends on parent runtime", r.Name)
								continue
							}
							//说明runtime可以执行
							err := gmo.runtimeManager.Run(task, &action, &r)
							if err != nil {
								logs.Error("run task err", err.Error())
							}
						}
					}
					//之后这里要考虑迁移的情况
				}
				if isSuccess {
					//将任务迁移到Completed队列当中
					logs.Infof("move to completed queue")
					gmo.groupQueues.DeleteFromRunning(task.Status.GroupID)
					gmo.groupQueues.AddToCompleted(task.Status.GroupID, task)
					continue
				}
			}
		}
	}
}

func (gmo *GroupMonitor) CompletedCheck() {
	logs.Info("Completed queue checking")
	//TODO 可能主要是将信息上传到api-server当中，然后将group_manager中的信息删除
	//for {
	//	select {
	//	case <-time.After(time.Second * 1):
	//
	//	}
	//}
}
func (gmo *GroupMonitor) ErrorCheck() {
	logs.Info("Error queue checking")
	//TODO 可能要做的就是通知调度器，group部署失败
	//for {
	//	select {
	//	case <-time.After(time.Second * 1):
	//
	//	}
	//}
}

// 处理 Runtime运行时启动，如果runtime是action下的首个执行的runtime，同时标记action的状态为Running
func (gmo *GroupMonitor) handleRuntimeStartUpdate(event events.RuntimeStartPhaseEvent) {
	// 更新 Runtime 的状态，依据实际变化更新相应字段
	logs.Info("Handling Runtime Start Status Update")
	g := event.Group
	a := event.Action
	r := event.Runtime
	phase := event.Phase
	startTime := event.StartAt
	lastTime := event.LastTime
	actionID := a.Name + g.Status.GroupID
	runtimeID := r.Name + actionID
	groupSpec := &g.Spec
	groupStatus := &g.Status
	var actionStart = true //action是否需要标记启动（下面的runtime如果都没启动，则说明action要标记Running）
	//修改group下面的groupSpec下面的Actions，Actions下面的ActionStatus，ActionStatus下面的RuntimeStatus
	for i := range groupSpec.Actions { //Action
		actionStatus := &groupSpec.Actions[i].Status //ActionStatus
		if actionStatus.ActionID != actionID {
			continue
		}
		for j := range actionStatus.RuntimeStatus { //RuntimeStatus
			rs := &actionStatus.RuntimeStatus[j]
			if rs.Phase == apis.Running || rs.Phase == apis.Successed {
				actionStart = false
			}
			if rs.RuntimeID == runtimeID {
				rs.Phase = phase
				rs.StartAt = startTime
				rs.LastTime = lastTime
			}
		}
		if actionStart { //为true说明要action还未设置状态为Running
			actionStatus.Phase = apis.Running
			actionStatus.StartAt = startTime
			actionStatus.LastTime = lastTime
		}
	}
	//修改group下面的groupStatus下面的ActionStatus，ActionStatus下面的RuntimeStatus
	for i := range groupStatus.ActionStatus { //ActionStatus
		as := &groupStatus.ActionStatus[i]
		if as.ActionID != actionID {
			continue
		}
		for j := range as.RuntimeStatus { //RuntimeStatus
			rs := &as.RuntimeStatus[j]
			if rs.RuntimeID == runtimeID {
				rs.Phase = phase
				rs.StartAt = startTime
				rs.LastTime = lastTime
			}
		}
		if actionStart { //为true说明要action还未设置状态为Running
			as.Phase = apis.Running
			as.StartAt = startTime
			as.LastTime = lastTime
		}
	}
	//将queue_manager和group_manager的group信息进行更新 ----------有问题
	err := gmo.groupQueues.UpdateGroup(g.Status.GroupID, g)
	if err != nil {
		logs.Error("update group-runtiem-start info error")
	}
	//测试
	p := g.Status.Phase
	actionStatus := g.Status.ActionStatus[0].Phase
	runtimeStatus := g.Status.ActionStatus[0].RuntimeStatus[0].Phase
	logs.Infof("groupStatus:%v, action status: %v, runtime Status: %v", p, actionStatus, runtimeStatus)
	actionStatus1 := g.Spec.Actions[0].Status.Phase
	runtimeStatus1 := g.Spec.Actions[0].Status.RuntimeStatus[0].Phase
	logs.Infof("groupStatus:%v, action status1: %v, runtime Status1: %v", p, actionStatus1, runtimeStatus1)
}

// 处理Runtime运行时结束,如果runtime是最后一个执行完成的，还得同时标记action的phase   总结：所有临时变量赋值时都得使用&
func (gmo *GroupMonitor) handleRuntimeEndUpdate(event events.RuntimeEndPhaseEvent) {
	// 更新 Runtime 的状态，依据实际变化更新相应字段
	logs.Info("Handling Runtime End Status Update")
	g := event.Group     //当前group *apis.Group
	a := event.Action    //当前Action *apis.Action
	r := event.Runtime   //当前runtime *apis.Runtime
	phase := event.Phase //当前phase可能为Succeed、Failed、Migrating、Migrated
	logs.Infof("handleRuntimeEndUpdate方法，收到phase:%s", phase)
	finshTime := event.FinishAt
	lastTime := event.LastTime
	actionID := a.Name + ":" + g.Status.GroupID
	runtimeID := r.Name + ":" + actionID
	groupSpec := &g.Spec
	groupStatus := &g.Status
	//修改group下面的groupSpec下面的Actions，Actions下面的ActionStatus，ActionStatus下面的RuntimeStatus
	var allRuntiemCompleted = true     // 当前action是否已经完成（只有action下面的所有的runtime都执行完成了，也就是最后一个runtime被执行完成了，要标记action的状态为succeed，如果说action下面的某一个runtime执行失败，则要标记action装填为Failed）
	var otherActionCompleted = true    // group下面的其他Action是否都已经完成
	var nowActionCompleted = false     // group下面的当前Action是否已经完成
	for i := range groupSpec.Actions { //Action
		actionStatus := &groupSpec.Actions[i].Status //ActionStatus
		if actionStatus.ActionID != actionID {       //遍历到其他Action，可以顺带看一下别的Action是否都已经完成了
			if actionStatus.Phase == apis.Pending { //其他Action为Pennding状态，说明还有其他的Action没有被遍历到，Group状态为Running状态
				otherActionCompleted = false
			}
			continue
		}
		for j := range actionStatus.RuntimeStatus { //RuntimeStatus
			rs := &actionStatus.RuntimeStatus[j]
			if rs.RuntimeID == runtimeID { //遍历到当前的RUntime，设置RUntime的属性
				rs.Phase = phase
				rs.FinishAt = finshTime
				rs.LastTime = lastTime
			}
			if rs.Phase == apis.Pending { // 遍历所有的Runtime，如果其中一个Runtime状态没有执行完成，说明Action最终不用更新
				allRuntiemCompleted = false
			}
		}
		if allRuntiemCompleted { //如果说ActionStatus下面的RuntimeStatus都被执行了，还得修改ActionStatus的phase状态
			actionStatus.Phase = phase
			//后续可能还要补充:Results
			//actionStatus.Results = results
			actionStatus.FinishAt = finshTime
			actionStatus.LastTime = lastTime
			nowActionCompleted = true //当前Action已经完成
		}
	}
	//如果说GroupStatus下面的ActionStatus都被执行了，还得修改GroupStatus的phase的状态
	if otherActionCompleted && nowActionCompleted { //说明其他Action都执行完成，当前Action也执行完成
		groupStatus.Phase = phase //Group的状态等于当前Action执行完成的状态：Failed  or  Succeed
	}
	//修改group下面的groupStatus下面的ActionStatus，ActionStatus下面的RuntimeStatus
	for i := range groupStatus.ActionStatus { //ActionStatus
		as := &groupStatus.ActionStatus[i]
		if as.ActionID != actionID {
			continue
		}
		for j := range as.RuntimeStatus { //RuntimeStatus
			rs := &as.RuntimeStatus[j]
			if rs.RuntimeID == runtimeID {
				rs.Phase = phase
				rs.FinishAt = finshTime
				rs.LastTime = lastTime
			}
		}
		if allRuntiemCompleted { //如果说ActionStatus下面的RuntimeStatus都是完成的状态，还得修改ActionStatus的phase状态
			as.Phase = phase
			as.FinishAt = finshTime
			as.LastTime = lastTime
		}
	}
	//将queue_manager和group_manager的group信息进行更新 ----------有问题
	err := gmo.groupQueues.UpdateGroup(g.Status.GroupID, g)
	if err != nil {
		logs.Error("update group-runtiem-end info error")
	}
	//测试：
	p := g.Status.Phase
	actionStatus := g.Status.ActionStatus[0].Phase
	runtimeStatus := g.Status.ActionStatus[0].RuntimeStatus[0].Phase
	logs.Infof("groupStatus:%v, action status: %v, runtime Status: %v", p, actionStatus, runtimeStatus)
	actionStatus1 := g.Spec.Actions[0].Status.Phase
	runtimeStatus1 := g.Spec.Actions[0].Status.RuntimeStatus[0].Phase
	logs.Infof("groupStatus:%v, action status1: %v, runtime Status1: %v", p, actionStatus1, runtimeStatus1)
}

// 检查Action的依赖是否满足
func (g *GroupMonitor) checkActionDependencies(action *apis.Action, group *apis.Group) bool {
	for _, parentName := range action.Spec.Parents {
		for _, ac := range group.Spec.Actions {
			if parentName == ac.Name && ac.Status.Phase != apis.Successed { //暂时没有考虑Migrating的状态，后续迁移的时候需要考虑
				return false
			}
		}
	}
	return true
}

// 检查Runtime的依赖是否满足
func (g *GroupMonitor) checkRuntimeDepencies(runtime *apis.Runtime, action *apis.Action) bool {
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

// 检查Group的依赖是否满足
func (g *GroupMonitor) checkGroupDepencies(group *apis.Group) bool {
	//TODO：实现依赖检查逻辑
	return true
}

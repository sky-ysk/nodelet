package monitor

import (
	"context"
	apis "hit.edu/framework/pkg/apis/cores"
	metav1 "hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/client-go/clients/typed/core"
	"hit.edu/framework/pkg/component-base/logs"
	"hit.edu/framework/pkg/nodelet/events"
	"hit.edu/framework/pkg/nodelet/events/eventbus"
	group "hit.edu/framework/pkg/nodelet/task/group"
	"hit.edu/framework/pkg/nodelet/task/runtime"
	"hit.edu/framework/pkg/nodelet/task/task"
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
	// task manager
	taskManager task.Manager
	// 存储任务队列
	groupQueues *group.GroupQueues
	eventBus    *eventbus.EventBus
	// 管理运行所需的Runtime
	// 存储RuntimeManager
	runtimeManager *runtime.RuntimeManager
	//Client-go
	nodesClient core.NodeInterface
	groupClient core.GroupInterface
	taskClient  core.TaskInterface
	stopCh      chan struct{}
}

func NewGroupMonitor(groupManager group.Manager, taskManager task.Manager, groupQueues *group.GroupQueues, eventbus *eventbus.EventBus, runtimeManager *runtime.RuntimeManager, nodeClient core.NodeInterface, groupClient core.GroupInterface, taskClient core.TaskInterface) *GroupMonitor {
	return &GroupMonitor{
		groupManager:   groupManager,
		taskManager:    taskManager,
		groupQueues:    groupQueues,
		eventBus:       eventbus,
		runtimeManager: runtimeManager,
		nodesClient:    nodeClient,
		groupClient:    groupClient,
		taskClient:     taskClient,
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
				gmo.handleRuntimeStartUpdate1(RuntimeEvent)
			case event := <-chRuntimeEnd:
				RuntimeEvent := event.(events.RuntimeEndPhaseEvent)
				gmo.handleRuntimeEndUpdate1(RuntimeEvent)
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
//  ticker := time.NewTicker(interval)
//  go func() {
//     for {
//        select {
//        case <-ticker.C: //ticker.C是一个通道，当ticker定时器每隔指定的时间向通道中写入数据
//           gmo.CheckStatus()
//        case <-gmo.stopCh:
//           ticker.Stop()
//           return
//        }
//     }
//  }()
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

// 都要改成for i：=range
func (gmo *GroupMonitor) PendingCheck() {
	// TODO 轮询检查Pending队列，检查任务group的依赖是否满足，如果满足才放入running队列当中
	logs.Info("pending queue checking")
	for {
		select {
		case <-time.After(time.Second * 1):
			penndingGroups := gmo.groupQueues.GetAllPending()
			for i := range penndingGroups {
				gr := penndingGroups[i]
				if !gmo.checkGroupDepencies(gr) { //再次检查group的执行依赖是否满足了（注意：group_workers当中任务头一次执行前也会检查）
					logs.Debug("The group %s execution dependency is not satisfied again", gr.Name)
					//继续放在Pennding队列当中，Pennding队列会持续检查依赖，直到依赖满足后，才开始执行，重新将任务group交给group_workers去执行
					gr.Status.CheckDependencyCount++
					if gr.Status.CheckDependencyCount > 10000 { // 当检查依赖的次数大于3次的话，说明依赖还是满足不了，迁移至Error队列
						//将任务迁移到Error队列当中
						gmo.groupQueues.DeleteFromPending(gr.Status.GroupID)
						gmo.groupQueues.AddToError(gr.Status.GroupID, gr)
						// TODO 修改queue_manager、group_manager当中的group信息
						gmo.handleStatusUpdate(gr, apis.Failed)
						gmo.groupQueues.UpdateGroup(gr.Status.GroupID, gr)
						_, err := gmo.groupClient.Update(context.TODO(), gr, metav1.UpdateOptions{})
						if err != nil {
							logs.Errorf("Failed to update group %s: %v", gr.Name, err)
						}
						break
					}
					continue
				} else { //说明group执行的依赖已经满足，接下来开始执行
					// 任务依赖满足后就将任务从Pennding队列转移至Running队列
					logs.Info("进入else分支---------------------------")
					gmo.groupQueues.DeleteFromPending(gr.Status.GroupID)
					gmo.groupQueues.AddToRunning(gr.Status.GroupID, gr)
					//此处不用再修改group信息为Running，真正启动任务的时候，会修改phase为running
					// TODO: 检查需要运行的Action
					// TODO: 开始部署
					logs.Infof("start group %s", gr.Name)
					//根据group当中的Action开启相应的runtime  group(Spec:Actions)--action（Spec：Runtimes）
					for j := range gr.Spec.Actions {
						action := &gr.Spec.Actions[j]
						if !gmo.checkActionDependencies(action, gr) {
							logs.Infof("Action %s in group %s waiting for dependencies", action.Name, gr.Name)
							continue
						}
						for k := range action.Spec.Runtimes {
							ru := &action.Spec.Runtimes[k]
							if !gmo.checkRuntimeDepencies(ru, action) {
								logs.Infof("Runtime %s in group %s waiting for dependencies", ru.Name, gr.Name)
								continue
							}
							err := gmo.runtimeManager.Run(gr, action, ru)
							if err != nil {
								logs.Error("run task err:", err.Error())
							}
						}
					}
				}
			}
		}
	}
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
			for i := range runningGroups {
				gro := runningGroups[i] //不用再加&&
				var isSuccess bool
				for j := range gro.Spec.Actions { //这里改成task.Status下面的Actions
					action := gro.Spec.Actions[j]
					isSuccess = false
					if action.Status.Phase == apis.Successed {
						isSuccess = true
						continue //说明当前Action执行完成了，接着查看下一个Action的执行情况
					}
					if action.Status.Phase == apis.Failed { //注意:runtime执行失败的时候除了标记Runtime状态为失败，也需要标记Runtime所属的Action状态为失败
						//将任务迁移到Error队列当中
						gmo.groupQueues.DeleteFromRunning(gro.Status.GroupID)
						gmo.groupQueues.AddToError(gro.Status.GroupID, gro)
						gmo.groupManager.DeleteGroup(gro) //groupManager就删除group的信息，此时group的信息就只存在于etcd当中
						break
					}
					if action.Status.Phase == apis.DeployCheck && action.Status.Waiting {
						if !gmo.checkActionDependencies(&action, gro) {
							logs.Infof("Action %s depends on parent action, parent not finish ", action.Name)
							continue
						}
						for k := range action.Spec.Runtimes {
							r := &action.Spec.Runtimes[k]
							if !gmo.checkRuntimeDepencies(r, &action) {
								logs.Infof("Runtime %s depends on parent runtime", r.Name)
								continue
							}
							//说明runtime可以执行
							//logs.Infof("************************************************************************************************************************")
							err := gmo.runtimeManager.Run(gro, &action, r)
							if err != nil {
								logs.Error("run task err", err.Error())
							}
						}
					}
					//检查runtime、Action当中的parents是否执行完成，如果父亲节点完成，则让他执行 ---这里有bug，就是任务已经放入running队列，但是还没执行完，这时候runningCheck循环遍历到当前runtime的状态为Pennding，查看是否满足执行条件，发现是满足的，结果有跑起来该任务
					if action.Status.Phase == apis.Running {
						for m := range action.Spec.Runtimes {
							r := &action.Spec.Runtimes[m]
							if !gmo.checkRuntimeDepencies(r, &action) {
								logs.Infof("Runtime %s depends on parent runtime", r.Name)
								continue
							}
							if r.Waiting { //如果说runtime也是被标记等待执行的状态，这才能开始执行
								//说明runtime可以执行
								logs.Infof("++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++")
								r.Waiting = false
								err := gmo.runtimeManager.Run(gro, &action, r)
								if err != nil {
									logs.Error("run task err", err.Error())
								}
							}
						}
					}
					//之后这里要考虑迁移的情况
				}
				if isSuccess {
					//将任务迁移到Completed队列当中
					logs.Infof("move to completed queue")
					gmo.groupQueues.DeleteFromRunning(gro.Status.GroupID)
					gmo.groupQueues.AddToCompleted(gro.Status.GroupID, gro)
					gmo.groupManager.DeleteGroup(gro) //groupManager就删除group的信息，此时group的信息就只存在于etcd当中
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
	// select {
	// case <-time.After(time.Second * 1):
	//
	// }
	//}
}
func (gmo *GroupMonitor) ErrorCheck() {
	logs.Info("Error queue checking")
	//TODO 可能要做的就是通知调度器，group部署失败
	//for {
	// select {
	// case <-time.After(time.Second * 1):
	//
	// }
	//}
}

// 修改所有的phase ,例如修改group执行Failed，修改上层Task的phase为Failed，修改下层action、runtime的phase为Failed
func (gmo *GroupMonitor) handleStatusUpdate(group *apis.Group, phase apis.Phase) {

}

// 处理 Runtime运行时启动，如果runtime是action下的首个执行的runtime，同时标记action的状态为Running，如果是第一个action启动，则group的状态也标记为running
func (gmo *GroupMonitor) handleRuntimeStartUpdate(event events.RuntimeStartPhaseEvent) {
	// 更新 Runtime 的状态，依据实际变化更新相应字段
	logs.Info("Handling Runtime Start Status Update")
	g := event.Group
	a := event.Action
	r := event.Runtime
	phase := event.Phase
	startTime := event.StartAt
	lastTime := event.LastTime
	actionID := a.Name + ":" + g.Status.GroupID
	runtimeID := r.Name + ":" + actionID
	groupSpec := &g.Spec
	groupStatus := &g.Status
	var actionStart = true //action是否需要标记启动（下面的runtime如果都没启动，则说明action要标记Running）
	var groupFirstStart = false
	//修改group下面的groupSpec下面的Actions，Actions下面的ActionStatus，ActionStatus下面的RuntimeStatus
	for i := range groupSpec.Actions { //Action
		actionStatus := &groupSpec.Actions[i].Status //ActionStatus
		if actionStatus.ActionID != actionID {
			continue
		}
		if i == 0 {
			groupFirstStart = true //说明Group的Phase也得设置为running
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
		if groupFirstStart {
			groupStatus.Phase = apis.Running
			// 这里遍历Task，Task里面，如果该Group是当前Task下第一个进入running状态的Group，则同时更新Task的状态为running
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
	runtimeStatus1 := g.Status.ActionStatus[0].RuntimeStatus[1].Phase
	logs.Infof("START ：groupStatus:%v, action status: %v, runtime Status1: %v, runtime Status2:%v", p, actionStatus, runtimeStatus, runtimeStatus1)
	actionStatus1 := g.Spec.Actions[0].Status.Phase
	runs := g.Spec.Actions[0].Status.RuntimeStatus[0].Phase
	runs1 := g.Spec.Actions[0].Status.RuntimeStatus[1].Phase
	logs.Infof("START ：groupStatus:%v, action status1: %v, runtime Status1: %v, runtime Status2:%v", p, actionStatus1, runs, runs1)
}

// 这里我打算从Task开始遍历
// 处理 Runtime运行时启动，如果runtime是action下的首个执行的runtime，同时标记action的状态为Running，如果是第一个action启动，则group的状态也标记为running
func (gmo *GroupMonitor) handleRuntimeStartUpdate1(event events.RuntimeStartPhaseEvent) {
	// 更新 Runtime 的状态，依据实际变化更新相应字段
	logs.Info("Handling Runtime Start Status Update")
	g := event.Group
	a := event.Action
	r := event.Runtime
	phase := event.Phase
	startTime := event.StartAt
	lastTime := event.LastTime
	actionID := a.Name + ":" + g.Status.GroupID
	runtimeID := r.Name + ":" + actionID
	groupSpec := &g.Spec
	groupStatus := &g.Status
	var actionStart = true //action是否需要标记启动（下面的runtime如果都没启动，则说明action要标记Running）
	taskID := g.Status.Belongs.TaskID
	groupID := g.Name + ":" + taskID
	//task, err2 := gmo.taskManager.GetTaskByID(taskID) //从etcd当中得到引用
	//if err2 != nil {
	//	logs.Error("Get task by taskID error：", err2)
	//}
	//logs.Infof("%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%task Pointer address: %p", task)
	list, err := gmo.taskClient.List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		logs.Error("list task err:", err.Error())
	}
	var task1 *apis.Task
	var err2 error
	for _, t := range list.Items { //遍历etcd当中的所有task
		if t.Status.TaskID == taskID { // 如果taskId对上了，则就修改该Task的Phase为Pennding
			task1, err2 = gmo.taskClient.Get(context.TODO(), t.Name, metav1.GetOptions{})
			if err2 != nil {
				logs.Error("Get task by taskID error from etcd：", err2)
			}
		}
	}

	//if task == task1 {
	//	logs.Infof("task and task2 point to the same memory location.")
	//} else {
	//	logs.Infof("task and task2 point to different memory locations.")
	//}
	// 修改Task下面的TaskStatus下面的状态  从Task开始遍历的好处是可以修改Task下面的状态
	for i := range task1.Spec.Groups {
		if task1.Spec.Groups[i].Name != groupSpec.Name {
			continue
		}
		if i == 0 && task1.Status.Phase != apis.Running {
			// 说明是Task中的第一个Group启动，那直接标记Task的状态也为Pennding
			task1.Status.Phase = apis.Running
			task1.Status.StartAt = startTime
			task1.Status.LastTime = lastTime
		}
		// TODO
	}
	//修改group下面的groupSpec下面的Actions，Actions下面的ActionStatus，ActionStatus下面的RuntimeStatus
	for j := range groupSpec.Actions { //GroupSpec(Actions)---->ActionStatus----->RuntimeStatus
		actionStatus := &groupSpec.Actions[j].Status //ActionStatus
		if actionStatus.ActionID != actionID {
			continue
		}
		if j == 0 {
			groupStatus.Phase = apis.Running //说明Group的Phase也得设置为running
			groupStatus.StartAt = startTime
			groupStatus.LastTime = lastTime
			//遍历task，同时标记TaskStatus下GroupStatus状态也为running
			for i := range task1.Status.GroupStatus {
				if task1.Status.GroupStatus[i].GroupID == groupID {
					task1.Status.GroupStatus[i].Phase = apis.Running
				}
			}
		}
		for k := range actionStatus.RuntimeStatus { //RuntimeStatus
			rs := &actionStatus.RuntimeStatus[k]
			if rs.Phase == apis.Running || rs.Phase == apis.Successed {
				actionStart = false
			}
			if rs.RuntimeID == runtimeID {
				rs.Phase = phase
				rs.StartAt = startTime
				rs.LastTime = lastTime
			}
		}
		if actionStart { //为true说明要action还未设置状态为Running  TODO 后期可以改为k=0 并且rs.Phase == apis.DeployCheck 进行下述操作
			actionStatus.Phase = apis.Running
			actionStatus.StartAt = startTime
			actionStatus.LastTime = lastTime
		}
	}

	//修改group下面的groupStatus下面的ActionStatus，ActionStatus下面的RuntimeStatus
	for j := range groupStatus.ActionStatus { //GroupStatus--->ActionStatus--->RuntimeStatus
		as := &groupStatus.ActionStatus[j]
		if as.ActionID != actionID {
			continue
		}
		for k := range as.RuntimeStatus { //RuntimeStatus
			rs := &as.RuntimeStatus[k]
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
	// 通过client-go，将信息提交到api-server当中
	gmo.taskClient.Update(context.TODO(), task1, metav1.UpdateOptions{})
	gmo.groupClient.Update(context.TODO(), g, metav1.UpdateOptions{})

	//将queue_manager和group_manager的group信息进行更新 ----------有问题
	err = gmo.groupQueues.UpdateGroup(g.Status.GroupID, g)
	if err != nil {
		logs.Error("update group-runtiem-start info error")
	}
	// 修改task的信息
	gmo.taskManager.UpdateTask(task1)
	//测试
	taskPhase := task1.Status.Phase
	p := g.Status.Phase
	actionStatus := g.Status.ActionStatus[0].Phase
	runtimeStatus := g.Status.ActionStatus[0].RuntimeStatus[0].Phase
	logs.Infof("START ：taskStatus:%v,groupStatus:%v, action status: %v, runtime Status: %v", taskPhase, p, actionStatus, runtimeStatus)
	actionStatus1 := g.Spec.Actions[0].Status.Phase
	runs := g.Spec.Actions[0].Status.RuntimeStatus[0].Phase
	logs.Infof("START ：groupStatus:%v, action status1: %v, runtime Status1: %v", p, actionStatus1, runs)
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
			if actionStatus.Phase == apis.DeployCheck { //其他Action为Pennding状态，说明还有其他的Action没有被遍历到，Group状态为Running状态
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
			if rs.Phase == apis.DeployCheck { // 遍历所有的Runtime，如果其中一个Runtime状态没有执行完成，说明Action最终不用更新
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
	runStatus1 := g.Status.ActionStatus[0].RuntimeStatus[0].Phase
	runStatus2 := g.Status.ActionStatus[0].RuntimeStatus[1].Phase
	logs.Infof("END ：groupStatus:%v, action status: %v, runtime Status1: %v, runtime Status2:%v", p, actionStatus, runStatus1, runStatus2)
	actionStatus1 := g.Spec.Actions[0].Status.Phase
	runtimeStatus1 := g.Spec.Actions[0].Status.RuntimeStatus[0].Phase
	runtimeStatus2 := g.Spec.Actions[0].Status.RuntimeStatus[1].Phase
	logs.Infof("END ：groupStatus:%v, action status1: %v, runtime Status1: %v, runtime Status2:%v", p, actionStatus1, runtimeStatus1, runtimeStatus2)
}

// 处理Runtime运行时结束,如果runtime是最后一个执行完成的，还得同时标记action的phase   总结：所有临时变量赋值时都得使用&
/*func (gmo *GroupMonitor) handleRuntimeEndUpdate2(event events.RuntimeEndPhaseEvent) {
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
	var allRuntiemCompleted = true  // 当前action是否已经完成（只有action下面的所有的runtime都执行完成了，也就是最后一个runtime被执行完成了，要标记action的状态为succeed，如果说action下面的某一个runtime执行失败，则要标记action装填为Failed）
	var otherActionCompleted = true // group下面的其他Action是否都已经完成
	var nowActionCompleted = false  // group下面的当前Action是否已经完成

	taskID := g.Status.Belongs.TaskID
	task, err2 := gmo.taskManager.GetTaskByID(taskID) //从etcd当中得到引用
	if err2 != nil {
		logs.Error("Get task by taskID error：", err2)
	}
	task.Status.Phase = apis.Failed
	gmo.taskClient.Update(context.TODO(), task, metav1.UpdateOptions{})

	for i := range groupSpec.Actions { //Action
		actionStatus := &groupSpec.Actions[i].Status //ActionStatus
		if actionStatus.ActionID != actionID {       //遍历到其他Action，可以顺带看一下别的Action是否都已经完成了
			if actionStatus.Phase == apis.DeployCheck { //其他Action为Pennding状态，说明还有其他的Action没有被遍历到，Group状态为Running状态
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
			if rs.Phase == apis.DeployCheck { // 遍历所有的Runtime，如果其中一个Runtime状态没有执行完成，说明Action最终不用更新
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
	runStatus1 := g.Status.ActionStatus[0].RuntimeStatus[0].Phase
	runStatus2 := g.Status.ActionStatus[0].RuntimeStatus[1].Phase
	logs.Infof("END ：groupStatus:%v, action status: %v, runtime Status1: %v, runtime Status2:%v", p, actionStatus, runStatus1, runStatus2)
	actionStatus1 := g.Spec.Actions[0].Status.Phase
	runtimeStatus1 := g.Spec.Actions[0].Status.RuntimeStatus[0].Phase
	runtimeStatus2 := g.Spec.Actions[0].Status.RuntimeStatus[1].Phase
	logs.Infof("END ：groupStatus:%v, action status1: %v, runtime Status1: %v, runtime Status2:%v", p, actionStatus1, runtimeStatus1, runtimeStatus2)
}*/

// 这里我打算从Task开始遍历
// 处理Runtime运行时结束,如果runtime是最后一个执行完成的，还得同时标记action的phase   总结：所有临时变量赋值时都得使用&
func (gmo *GroupMonitor) handleRuntimeEndUpdate1(event events.RuntimeEndPhaseEvent) {
	// 更新 Runtime 的状态，依据实际变化更新相应字段
	logs.Info("Handling Runtime End Status Update")
	g := event.Group     //当前group *apis.Group
	a := event.Action    //当前Action *apis.Action
	r := event.Runtime   //当前runtime *apis.Runtime
	phase := event.Phase //当前phase可能为Succeed、Failed、Migrating、Migrated
	logs.Infof("handleRuntimeEndUpdate方法，收到phase:%s", phase)
	finshTime := event.FinishAt
	lastTime := event.LastTime
	// Task 信息
	taskID := g.Status.Belongs.TaskID

	groupID := g.Name + ":" + taskID
	actionID := a.Name + ":" + g.Status.GroupID
	runtimeID := r.Name + ":" + actionID
	groupSpec := &g.Spec
	groupStatus := &g.Status
	//修改group下面的groupSpec下面的Actions，Actions下面的ActionStatus，ActionStatus下面的RuntimeStatus
	var allRuntiemCompleted = true  // 当前action是否已经完成（只有action下面的所有的runtime都执行完成了，也就是最后一个runtime被执行完成了，要标记action的状态为succeed，如果说action下面的某一个runtime执行失败，则要标记action装填为Failed）
	var otherActionCompleted = true // group下面的其他Action是否都已经完成
	var nowActionCompleted = false  // group下面的当前Action是否已经完成

	// 修改Task下面的TaskStatus下面的状态  从Task开始遍历的好处是可以修改Task下面的状态
	var otherGroupCompleted = true
	var nowGroupCompleted = false

	//task, err2 := gmo.taskManager.GetTaskByID(taskID) //从etcd当中得到引用
	//if err2 != nil {
	//	logs.Error("Get task by taskID error：", err2)
	//}
	list, err := gmo.taskClient.List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		logs.Error("list task err:", err.Error())
	}
	var task1 *apis.Task
	var err2 error
	for _, t := range list.Items { //遍历etcd当中的所有task
		if t.Status.TaskID == taskID { // 如果taskId对上了，则就修改该Task的Phase为Pennding
			task1, err2 = gmo.taskClient.Get(context.TODO(), t.Name, metav1.GetOptions{})
			if err2 != nil {
				logs.Error("Get task by taskID error from etcd：", err2)
			}
		}
	}
	//if task == task1 {
	//	logs.Infof("task and task2 point to the same memory location.")
	//} else {
	//	logs.Infof("task and task2 point to different memory locations.")
	//}
	//logs.Infof("task point address:%p", task)
	// 检查其他的group是否完成
	for i := range task1.Spec.Groups {
		//grStatus := &task1.Spec.Groups[i].Status
		grStatus := &task1.Status.GroupStatus[i]
		if grStatus.GroupID != groupID {
			//logs.Infof("***********************grStatus.Phase:%v", grStatus.Phase)
			if grStatus.Phase == apis.DeployCheck { //if grStatus.Phase == apis.DeployCheck {
				otherGroupCompleted = false
			}
			continue
		}
	}
	for i := range groupSpec.Actions { //Action
		actionStatus := &groupSpec.Actions[i].Status //ActionStatus
		if actionStatus.ActionID != actionID {       //遍历到其他Action，可以顺带看一下别的Action是否都已经完成了
			if actionStatus.Phase == apis.DeployCheck { //其他Action为Pennding状态，说明还有其他的Action没有被遍历到，Group状态为Running状态
				otherActionCompleted = false
			}
			continue
		}
		for j := range actionStatus.RuntimeStatus { //RuntimeStatus
			rs := &actionStatus.RuntimeStatus[j]
			//logs.Infof("rs.RuntimeID:%v", rs.RuntimeID)
			//logs.Infof("runtimeID:%v", runtimeID)
			if rs.RuntimeID == runtimeID { //遍历到当前的RUntime，设置RUntime的属性
				rs.Phase = phase
				rs.FinishAt = finshTime
				rs.LastTime = lastTime
			}
			if rs.Phase == apis.DeployCheck { // 遍历所有的Runtime，如果其中一个Runtime状态没有执行完成，说明Action最终不用更新
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
	//logs.Infof("oteherActionCompleted:%v", otherActionCompleted)
	//logs.Infof("nowActionCompleted:%v", nowActionCompleted)
	//如果说GroupStatus下面的ActionStatus都被执行了，还得修改GroupStatus的phase的状态
	if otherActionCompleted && nowActionCompleted { //说明其他Action都执行完成，当前Action也执行完成
		groupStatus.Phase = phase //Group的状态等于当前Action执行完成的状态：Failed  or  Succeed
		groupStatus.FinishAt = finshTime
		groupStatus.LastTime = lastTime
		nowGroupCompleted = true
		//遍历task，同时标记TaskStatus下GroupStatus状态也为phase
		for i := range task1.Status.GroupStatus {
			if task1.Status.GroupStatus[i].GroupID == groupID {
				task1.Status.GroupStatus[i].Phase = phase
			}
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
	//logs.Infof("===================otherGroupCompleted:%v", otherGroupCompleted)
	//logs.Infof("==================nowGroupCompleted:%v", nowGroupCompleted)
	//如果说TaskStatus下面的Group都被执行了，还得修改TaskStatus的phase的状态
	if otherGroupCompleted && nowGroupCompleted {
		task1.Status.Phase = phase
		task1.Status.FinishAt = finshTime
		task1.Status.LastTime = lastTime
	}
	//将queue_manager和group_manager的group信息进行更新 ----------有问题
	err = gmo.groupQueues.UpdateGroup(g.Status.GroupID, g)
	if err != nil {
		logs.Error("update group-runtiem-end info error")
	}
	logs.Infof("group:%v", g.Status.Phase)
	// 通过client-go，将信息提交到api-server当中
	gmo.taskClient.Update(context.TODO(), task1, metav1.UpdateOptions{})
	gmo.groupClient.Update(context.TODO(), g, metav1.UpdateOptions{})
	// 修改task的信息
	gmo.taskManager.UpdateTask(task1)
	//测试：
	taskPhase := task1.Status.Phase
	p := g.Status.Phase
	actionStatus := g.Status.ActionStatus[0].Phase
	runStatus1 := g.Status.ActionStatus[0].RuntimeStatus[0].Phase
	logs.Infof("END ：taskStatus:%v, groupStatus:%v, action status: %v, runtime Status: %v", taskPhase, p, actionStatus, runStatus1)
	actionStatus1 := g.Spec.Actions[0].Status.Phase
	runtimeStatus1 := g.Spec.Actions[0].Status.RuntimeStatus[0].Phase
	logs.Infof("END ：groupStatus:%v, action status1: %v, runtime Status: %v", p, actionStatus1, runtimeStatus1)
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

// 改法1：使用i，j下标
func (g *GroupMonitor) checkRuntimeDepencies(runtime *apis.Runtime, action *apis.Action) bool {
	for i := range runtime.Parents {
		parentName := runtime.Parents[i]
		for j := range action.Status.RuntimeStatus {
			rs := action.Status.RuntimeStatus[j]
			if action.Spec.Runtimes[j].Name == parentName &&
				rs.Phase != apis.Successed {
				return false
			}
		}
	}
	return true
}

// 改法2：前缀 + ：前判断
func (g *GroupMonitor) checkRuntimeDepencies1(runtime *apis.Runtime, action *apis.Action) bool {
	for _, parentName := range runtime.Parents {
		for _, rs := range action.Status.RuntimeStatus {
			if strings.Split(rs.RuntimeID, ":")[0] == parentName &&
				rs.Phase != apis.Successed {

				return false
			}
		}
	}
	return true
}

// 检查Runtime的依赖是否满足
func (g *GroupMonitor) checkRuntimeDepencies2(runtime *apis.Runtime, action *apis.Action) bool {
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
	if len(group.Spec.Parents) == 0 {
		return true
	} else {
		//检查父亲group是否执行完成
		for i := range group.Spec.Parents {
			parentName := group.Spec.Parents[i]
			// 去client-go当中查group
			result, err := g.groupClient.Get(context.TODO(), parentName, metav1.GetOptions{})
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

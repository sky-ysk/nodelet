package monitor

import (
	"context"
	"encoding/json"
	"hit.edu/framework/pkg/apimachinery/types"
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
	logs.Info("GroupMonitor component start")

	//订阅事件
	chRuntimeStart := make(chan interface{})
	chRuntimeEnd := make(chan interface{})

	gmo.eventBus.Subscribe(reflect.TypeOf(events.RuntimeStartPhaseEvent1{}), chRuntimeStart)
	gmo.eventBus.Subscribe(reflect.TypeOf(events.RuntimeEndPhaseEvent1{}), chRuntimeEnd)
	//启动监听事件
	go func() {
		for {
			select {
			case event := <-chRuntimeStart:
				RuntimeEvent := event.(events.RuntimeStartPhaseEvent1)
				gmo.handleRuntimeStartUpdate(RuntimeEvent)
			case event := <-chRuntimeEnd:
				RuntimeEvent := event.(events.RuntimeEndPhaseEvent1)
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

// 检查checking队列的任务，前置任务是否完成，看是否需要迁移到running队列
// 检查running队列，看任务是否还在执行、是否执行完成、、进程占用资源量    任务完成和任务失败下线该如何判断呢？
// 检查error队列，检查任务是否出错，看能否尝试拉起，多次尝试拉起失败后，重新提交给调度器
func (gmo *GroupMonitor) CheckStatus() {
	go gmo.CheckingQueueCheck()
	go gmo.RunningQueueCheck()
	go gmo.CompletedQueueCheck()
	go gmo.ErrorQueueCheck()
}

// 都要改成for i：=range
func (gmo *GroupMonitor) CheckingQueueCheck() {
	//主要针对Task下的多个Group在多个设备上运行，group之间有依赖关系，需要检查
	// TODO 轮询检查Checking队列，检查任务group的依赖是否满足，如果满足才放入running队列当中
	logs.Info("Pending queue start checking")
	for {
		select {
		case <-time.After(time.Second * 1):
			checkingGroups := gmo.groupQueues.GetAllChecking()
			for i := range checkingGroups {
				gr := checkingGroups[i]
				// 从etcd当中读取group信息
				get, err := gmo.groupClient.Get(context.TODO(), gr.Name, metav1.GetOptions{})
				if err != nil {
					logs.Errorf("Etcd get group error:%v", err)
				}
				if !gmo.groupDepenSatisfy(get) { //再次检查group的执行依赖是否满足了（注意：group_workers当中任务头一次执行前也会检查）
					//logs.Debugf("The group ：%s execution dependency is not satisfied again, now still in Checking Queue", gr.Name)
					//继续放在checking队列当中，checking队列会持续检查依赖，直到依赖满足后，才开始执行，重新将任务group交给group_workers去执行
					get.Status.CheckDependencyCount++
					//num := get.Status.CheckDependencyCount
					//logs.Infof("Status.CheckDependencyCount:%v", get.Status.CheckDependencyCount)
					if get.Status.CheckDependencyCount > 10000 { // 当检查依赖的次数大于1000次的话，说明依赖还是满足不了，迁移至Error队列---这里其实有问题（group如果有前序依赖，你不知道什么时候其前序依赖能完成），暂停1000s其实也是有问题的
						//将任务迁移到Error队列当中
						gmo.groupQueues.DeleteFromCheckingAndAddToError(get.Status.GroupID)
						// TODO 修改queue_manager、group_manager当中的group信息---这里其实可以不用--继续检查
						gmo.handleStatusUpdate(get, apis.Failed) //===========================================待完成
						continue
					}
					// TODO 这里后期得优化
					patchGroup, err := json.Marshal(map[string]interface{}{
						"status": map[string]interface{}{
							"check_dependency_count": get.Status.CheckDependencyCount,
						},
					})
					if err != nil {
						logs.Errorf("Json Marshal failed, err:%v", err)
					}
					_, err = gmo.groupClient.Patch(context.TODO(), get.Name, types.StrategicMergePatchType, patchGroup, metav1.PatchOptions{})
					if err != nil {
						logs.Errorf("Patch group error-1:%v", err)
					}
					// update的方式
					//_, err := gmo.groupClient.Update(context.TODO(), get, metav1.UpdateOptions{})
					//if err != nil {
					//	logs.Errorf("etcd update group error:%v", err)
					//}
					continue
				} else { //说明group执行的依赖已经满足，接下来开始执行
					// 任务依赖满足后就将任务从checking队列转移至Running队列
					gmo.groupQueues.DeleteFromCheckingAndAddToRunning(get.Status.GroupID)
					//此处不用再修改group信息为Running，真正启动任务的时候，会修改phase为running
					// TODO: 检查需要运行的Action
					// TODO: 开始部署
					logs.Infof("Checking queue start group:%s", get.Spec.Name)
					//根据group当中的Action开启相应的runtime  group(Spec:Actions)--action（Spec：Runtimes）
					grou, err := gmo.groupManager.GetGroupByID(get.Status.GroupID)
					if err != nil {
						logs.Errorf("Get group by id frmo group_manager err:%v", err)
					}
					for j := range get.Spec.Actions {
						action := &get.Spec.Actions[j]
						if !gmo.actionDepenSatisfy(j, get) {
							logs.Infof("Action:%s in group:%s waiting for dependencies", action.Spec.Name, get.Spec.Name)
							//action.Status.Waiting = true
							grou.Spec.Actions[j].Status.Waiting = true
							//logs.Infof("****************CheckingQueueCheck:ActionWaiting:%v,j:%v", grou.Spec.Actions[j].Status.Waiting, j)
							continue
						}
						for k := range action.Spec.Runtimes {
							ru := &action.Spec.Runtimes[k]
							if !gmo.runtimeDepenSatisfy(j, k, get) {
								logs.Infof("Runtime:%s in group:%s waiting for dependencies", ru.Name, get.Spec.Name)
								//ru.Waiting = true
								grou.Spec.Actions[j].Spec.Runtimes[k].Waiting = true
								//logs.Infof("****************CheckingQueueCheck:RuntimeWaiting:%v,j:%v,k:%v", grou.Spec.Actions[j].Spec.Runtimes[k].Waiting, j, k)
								continue
							}
							go gmo.runtimeManager.Run(get, action, ru, j, k)
							if err != nil {
								logs.Error("Run task err:%v", err)
							}
						}
					}
					_, err = gmo.groupClient.Update(context.TODO(), get, metav1.UpdateOptions{})
					if err != nil {
						logs.Errorf("Update task error:%v", err)
					}
				}
				//// 也上传一份到group_manager当中
				//err = gmo.groupQueues.UpdateGroup(get.Status.GroupID, get)
				//if err != nil {
				//	logs.Errorf("update group to group_manager err:%v", err)
				//}
			}
		}
	}
}

// 检查Running队列，做的事情：①如果发现任务完成，迁移到Completed队列，如果发现任务失败，迁移到Error队列
// ②检查runtime、Action当中的parents是否执行完成，如果完成，则执行
func (gmo *GroupMonitor) RunningQueueCheck() {
	//主要针对当前设备上的Group，下面有多个Action，之间有依赖关系，需要检查
	//TODO 监控进程的返回值等判断任务是否正常执行完成，正常则放入completedqueue，否则放入errorqueue(方法待确认)
	logs.Info("Running queue start checking")
	for {
		select {
		case <-time.After(time.Second * 1):
			runningGroups := gmo.groupQueues.GetAllRunning()
			for i := range runningGroups {
				gro := runningGroups[i] //不用再加&&
				// 从etcd当中读取group信息
				group, err := gmo.groupClient.Get(context.TODO(), gro.Name, metav1.GetOptions{})
				if err != nil {
					logs.Errorf("Etcd get group error:%v", err)
				}
				var isSuccess bool                  // 标记group下面的action是否都执行成功
				for j := range group.Spec.Actions { // 遍历group当中的Action
					action := &group.Spec.Actions[j]
					isSuccess = false
					if action.Status.Phase == apis.Successed { // 当前action的状态为Successed
						isSuccess = true
						continue //说明当前Action执行完成了，接着查看下一个Action的执行情况
					}
					if action.Status.Phase == apis.Failed { //注意:runtime执行失败的时候除了标记Runtime状态为失败，也需要标记Runtime所属的Action状态为失败
						//将任务迁移到Error队列当中
						gmo.groupQueues.DeleteFromRunningAndAddToError(group.Status.GroupID)
						gmo.groupManager.DeleteGroup(group) //groupManager就删除group的信息，此时group的信息就只存在于etcd当中
						break
					}
					grou, err := gmo.groupManager.GetGroupByID(group.Status.GroupID)
					if err != nil {
						logs.Errorf("Get group by id from group_manager err:%v", err)
					}
					if action.Status.Phase == apis.DeployCheck && grou.Spec.Actions[j].Status.Waiting { //&& action.Status.Waiting
						if !gmo.actionDepenSatisfy(j, group) {
							//logs.Infof("Action %s depends on parent action, parent not finish ", action.Name)
							continue
						}
						//action.Status.Waiting = false // 说明Action的父亲Action已经执行完成了，那么接下来Action的Runtime必须会被执行（至少会执行一个runtime）
						grou.Spec.Actions[j].Status.Waiting = false
						//logs.Infof("Set-false===============-=====RunningQueueCheck:Action-Waiting:%v,j:%v", grou.Spec.Actions[j].Status.Waiting, j)
						// patch
						//patchGroupActions, err4 := json.Marshal(map[string]interface{}{
						//	"spec": map[string]interface{}{
						//		"actions": group.Spec.Actions,
						//	},
						//})
						//if err4 != nil {
						//	logs.Errorf("json marshal:patchGroupActions err:%v", err)
						//}
						//_, err = gmo.groupClient.Patch(context.TODO(), group.Name, types.StrategicMergePatchType, patchGroupActions, metav1.PatchOptions{})
						//if err != nil {
						//	logs.Errorf("patch patchGroupActions:group err:%v", err)
						//}

						for k := range action.Spec.Runtimes {
							r := &action.Spec.Runtimes[k]
							if !gmo.runtimeDepenSatisfy(j, k, group) {
								//logs.Infof("Runtime %s depends on parent runtime", r.Name)
								//r.Waiting = true
								grou.Spec.Actions[j].Spec.Runtimes[k].Waiting = true
								//logs.Infof("set-true==================RunningQueueCheck:Runtime-Waiting:%v,j:%v,k:%v", grou.Spec.Actions[j].Spec.Runtimes[k].Waiting, j, k)
								// patch
								//patchGroupActionsRuntimes, err4 := json.Marshal(map[string]interface{}{
								//	"spec": map[string]interface{}{
								//		"actions": group.Spec.Actions,
								//	},
								//})
								//if err4 != nil {
								//	logs.Errorf("json marshal:patchGroupActions err:%v", err)
								//}
								//_, err := gmo.groupClient.Patch(context.TODO(), group.Name, types.StrategicMergePatchType, patchGroupActionsRuntimes, metav1.PatchOptions{})
								//if err != nil {
								//	logs.Errorf("patch patchGroupActionsRuntimes:group err:%v", err)
								//}
								continue
							}
							//说明runtime可以执行
							go gmo.runtimeManager.Run(group, action, r, j, k)
							if err != nil {
								logs.Errorf("Run task err:%v", err)
							}
						}

					}
					//检查runtime、Action当中的parents是否执行完成，如果父亲节点完成，则让他执行 ---这里有bug，就是任务已经放入running队列，但是还没执行完，这时候runningCheck循环遍历到当前runtime的状态为checking，查看是否满足执行条件，发现是满足的，结果有跑起来该任务
					if action.Status.Phase == apis.Running {
						for m := range action.Spec.Runtimes {
							r := &action.Spec.Runtimes[m]
							if !gmo.runtimeDepenSatisfy(j, m, group) {
								//logs.Infof("Runtime %s depends on parent runtime", r.Name)
								//r.Waiting = true  // 这里不需要再标记了，因为在DeployCheck阶段就遍历了所有的runtime并标记了
								continue
							}
							grou, err := gmo.groupManager.GetGroupByID(group.Status.GroupID)
							if err != nil {
								logs.Errorf("Get group by id from group_manager err:%v", err)
							}
							//logs.Infof("$$$$$$==================grou.Spec.Actions[j].Spec.Runtimes[m].Waiting:%v,j:%v,m:%v", grou.Spec.Actions[j].Spec.Runtimes[m].Waiting, j, m)
							if grou.Spec.Actions[j].Spec.Runtimes[m].Waiting { //如果说runtime也是被标记等待执行的状态，这才能开始执行  if r.Waiting
								//说明runtime可以执行
								grou.Spec.Actions[j].Spec.Runtimes[m].Waiting = false
								//logs.Infof("=====================RunningQueueCheck:waiting:%v,j:%v,m:%v", grou.Spec.Actions[j].Spec.Runtimes[m].Waiting, j, m)
								// patch
								//patchGroupActionsRuntimes, err4 := json.Marshal(map[string]interface{}{
								//	"spec": map[string]interface{}{
								//		"actions": group.Spec.Actions,
								//	},
								//})
								//if err4 != nil {
								//	logs.Errorf("json marshal:patchGroupActions err:%v", err)
								//}
								//_, err := gmo.groupClient.Patch(context.TODO(), group.Name, types.StrategicMergePatchType, patchGroupActionsRuntimes, metav1.PatchOptions{})
								//if err != nil {
								//	logs.Errorf("patch patchGroupActionsRuntimes:group err:%v", err)
								//}
								err = gmo.runtimeManager.Run(group, action, r, j, m)
								if err != nil {
									logs.Errorf("run task err:%v", err)
								}
							}
						}
					}
					//之后这里要考虑迁移的情况
				}
				if isSuccess {
					//将任务迁移到Completed队列当中
					logs.Info("Move to completed queue")
					gmo.groupQueues.DeleteFromRunningAndAddToCompleted(group.Status.GroupID)
					continue
				}
			}
		}
	}
}

func (gmo *GroupMonitor) CompletedQueueCheck() {
	logs.Info("Completed queue start checking")
	//TODO 可能主要是将信息上传到api-server当中，然后将group_manager中的信息删除
	for {
		select {
		case <-time.After(time.Second * 10):
			//var start = false
			completed := gmo.groupQueues.GetAllCompleted()
			for i := range completed {
				gro := completed[i]
				logs.Infof("Delete group:%v", gro.Spec.Name)
				gmo.groupManager.DeleteGroup(gro)
				gmo.groupQueues.DeleteFromCompleted(gro.Status.GroupID) //得根据groupId进行删除，不是根据groupName
				//start = true
			}

			//
			//if start {
			//	checking := gmo.groupQueues.GetAllChecking()
			//	var checkingNum int
			//	for i := range checking {
			//		checkingNum++
			//		logs.Infof("Check groupName:%v,num=%v", checking[i].Name, checkingNum)
			//	}
			//	var runningNum int
			//	running := gmo.groupQueues.GetAllRunning()
			//	for i := range running {
			//		runningNum++
			//		logs.Infof("running groupName:%v,num=%v", running[i].Name, runningNum)
			//	}
			//	var errorNum int
			//	errored := gmo.groupQueues.GetAllError()
			//	for i := range errored {
			//		errorNum++
			//		logs.Infof("error groupName:%v,num=%v", errored[i].Name, errorNum)
			//	}
			//}

		}
	}
}
func (gmo *GroupMonitor) ErrorQueueCheck() {
	logs.Info("Error queue start checking")
	////TODO 可能要做的就是通知调度器，group部署失败
	//for {
	//	select {
	//	case <-time.After(time.Second * 100):
	//		// 通知调度器
	//	}
	//}
}

// 收到的Phase为：Running or Failed
// 处理 Runtime运行时启动，如果runtime是action下的首个执行的runtime，同时标记action的状态为Running，如果是第一个action启动，则group的状态也标记为running
func (gmo *GroupMonitor) handleRuntimeStartUpdate(event events.RuntimeStartPhaseEvent1) {
	// 更新 Runtime 的状态，依据实际变化更新相应字段
	logs.Info("Handling Runtime Start Status Update")
	groupName := event.GroupName

	get, err := gmo.groupClient.Get(context.TODO(), groupName, metav1.GetOptions{})
	if err != nil {
		logs.Errorf("Failed get group:%v from etcd, err:%v", groupName, err)
	}

	actionIndex := event.ActionIndex
	runtimeIndex := event.RuntimeIndex
	phase := event.Phase // 这里接收的Phase有可能是running，也有可能是Failed
	logs.Infof("Get runtime start event notify, the phase:%s", phase)
	startTime := event.StartAt
	lastTime := event.LastTime
	//actionID := a.Name + ":" + get.Status.GroupID    actionID := get.Status.ActionStatus[actionIndex].ActionID
	//runtimeID := r.Name + ":" + actionID             runtimeID := get.Status.ActionStatus[actionIndex].RuntimeStatus[runtimeID].RuntimeID
	groupSpec := &get.Spec
	groupStatus := &get.Status
	var actionStart = true              //action是否需要标记启动（下面的runtime如果都没启动，则说明action要标记Running）
	taskID := get.Status.Belongs.TaskID // group的Belongs属性当中的TaskID

	// 首先先修改该runtime所对应的group-所对应的Task的phase
	list, err := gmo.taskClient.List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		logs.Errorf("Get list task err:%v", err)
	}
	var groupIndexInTask int //当前group在Task当中的下标
	var task1 *apis.Task     //group所属的Task对象
	var err2 error

	for _, t := range list.Items { //遍历etcd当中的所有task，找到当前group所属的Task
		if t.Status.TaskID == taskID { // 如果taskId对上了，则获取该Task
			task1, err2 = gmo.taskClient.Get(context.TODO(), t.Name, metav1.GetOptions{})
			if err2 != nil {
				logs.Errorf("Get task by taskID error from etcd:%v", err2)
			}
			break
		}
	}
	// ## 处理Task的Phase
	// 遍历Task下面的Group，根据Group的状态来设置Task的Phase
	for i := range task1.Spec.Groups {
		if task1.Status.GroupStatus[i].GroupID != groupStatus.GroupID { // 非当前处理的group，跳过
			continue
		}
		groupIndexInTask = i // 当前处理的group在Task当中的下标
		if i == 0 {          //task1.Status.Phase != apis.Running 主要是为了当group下面的多个runtime进行Start
			// 说明是Task中的第一个Group启动，那说明Task也是第一次启动，要标记状态为Phase（running or Failed）
			if phase == apis.Running && task1.Status.Phase != apis.Running {
				task1.Status.StartAt = startTime
				task1.Status.Phase = phase
				task1.Status.LastTime = lastTime
			} else if phase == apis.Failed {
				task1.Status.StartAt = startTime
				task1.Status.FinishAt = startTime //任务在执行的时候就发送错误，那么开始和结束时间都标记为同一时刻
				task1.Status.Phase = phase
				task1.Status.LastTime = lastTime
			}
		}
		// TODO
	}
	// ## 处理Group的Phase
	//修改group下面的groupSpec下面的Actions，Actions下面的ActionStatus，ActionStatus下面的RuntimeStatus
	for j := range groupSpec.Actions { //GroupSpec(Actions)---->ActionStatus----->RuntimeStatus
		actionStatus := &groupSpec.Actions[j].Status //ActionStatus
		if j != actionIndex {                        // 判断是否是当前处理的Action
			continue
		}
		// 如果当前处理的action为第一个action，同时对所属的group进行状态修改
		if j == 0 {
			if phase == apis.Running && groupStatus.Phase != apis.Running {
				groupStatus.StartAt = startTime
				groupStatus.Phase = phase //说明当前处理的任务是在第一个action当中，所以Group的Phase也得设置为phase（running or Failed）
				groupStatus.LastTime = lastTime
			} else if phase == apis.Failed {
				groupStatus.StartAt = startTime
				groupStatus.FinishAt = startTime //任务在执行的时候就发送错误，那么开始和结束时间都标记为同一时刻
				groupStatus.Phase = phase        //说明当前处理的任务是在第一个action当中，所以Group的Phase也得设置为phase（running or Failed）
				groupStatus.LastTime = lastTime
			}
		}
		for k := range actionStatus.RuntimeStatus { //RuntimeStatus
			rs := &actionStatus.RuntimeStatus[k]
			if rs.Phase == apis.Running || rs.Phase == apis.Successed || rs.Phase == apis.Failed {
				actionStart = false //说明action下面有runtime已经启动过了（状态可能是running、Successed或者Failed）
			}
			if k == runtimeIndex { // 是当前处理的runtime
				if phase == apis.Failed { //对于Phase等于Running，标记startTime
					rs.FinishAt = startTime //任务在执行的时候就发送错误，那么开始和结束时间都标记为同一时刻
				}
				rs.StartAt = startTime
				rs.Phase = phase // 这里的Phase有可能是Failed，也有可能是running
				rs.LastTime = lastTime

			}
		}
		if actionStart { //为true说明要action还未设置状态为Running  TODO 后期可以改为k=0 并且rs.Phase == apis.DeployCheck 进行下述操作
			if phase == apis.Running && actionStatus.Phase != apis.Running { // 说明action当前是DeployCheck状态
				actionStatus.StartAt = startTime
				actionStatus.Phase = phase
				actionStatus.LastTime = lastTime
			} else if phase == apis.Failed {
				actionStatus.StartAt = startTime
				actionStatus.FinishAt = startTime //任务在执行的时候就发送错误，那么开始和结束时间都标记为同一时刻
				actionStatus.Phase = phase
				actionStatus.LastTime = lastTime
			}

		}
	}

	//修改group下面的groupStatus下面的ActionStatus，ActionStatus下面的RuntimeStatus
	for j := range groupStatus.ActionStatus { //GroupStatus--->ActionStatus--->RuntimeStatus
		as := &groupStatus.ActionStatus[j]
		if j != actionIndex {
			continue
		}
		for k := range as.RuntimeStatus { //RuntimeStatus
			rs := &as.RuntimeStatus[k]
			if k == runtimeIndex {
				if phase == apis.Failed {
					rs.FinishAt = startTime
				}
				rs.StartAt = startTime
				rs.Phase = phase
				rs.LastTime = lastTime
			}
		}
		if actionStart { //为true说明要action还未设置状态为Running
			if phase == apis.Running && as.Phase != apis.Running {
				as.StartAt = startTime
				as.Phase = phase
				as.LastTime = lastTime
			} else if phase == apis.Failed {
				as.StartAt = startTime
				as.FinishAt = startTime
				as.Phase = phase
				as.LastTime = lastTime
			}
		}
	}
	// 将修改后的Group状态值赋值给Task
	task1.Status.GroupStatus[groupIndexInTask] = get.Status
	task1.Spec.Groups[groupIndexInTask].Spec = get.Spec
	task1.Spec.Groups[groupIndexInTask].Status = get.Status

	_, err4 := gmo.taskClient.Update(context.TODO(), task1, metav1.UpdateOptions{})
	if err4 != nil {
		logs.Errorf("Update group-runtiem-start info error-1:%v", err4)
		time.Sleep(100 * time.Millisecond)
		_, err4 = gmo.taskClient.Update(context.TODO(), task1, metav1.UpdateOptions{})
	}
	//logs.Info("time----monitor1=============:")
	_, err3 := gmo.groupClient.Update(context.TODO(), get, metav1.UpdateOptions{})
	if err3 != nil {
		logs.Error("Update group-runtiem-start info error-2:", err3)
		time.Sleep(100 * time.Millisecond)
		_, err3 = gmo.groupClient.Update(context.TODO(), get, metav1.UpdateOptions{})
	}
	// 上行代码报错，改为使用patch
	//// patch
	//patchGroupActions, err4 := json.Marshal(map[string]interface{}{
	//	"spec": map[string]interface{}{
	//		"actions": get.Spec.Actions,
	//	},
	//})
	//if err4 != nil {
	//	logs.Errorf("json marshal:patchGroupActions err:%v", err)
	//}
	//_, err = gmo.groupClient.Patch(context.TODO(), get.Name, types.StrategicMergePatchType, patchGroupActions, metav1.PatchOptions{})
	//if err != nil {
	//	logs.Errorf("patch patchGroupActions:group err:%v", err)
	//}
	//patchGroupActionStatus, err5 := json.Marshal(map[string]interface{}{
	//	"status": map[string]interface{}{
	//		"action_status": get.Status.ActionStatus,
	//	},
	//})
	//if err5 != nil {
	//	logs.Errorf("json marshal:patchGroupActions err:%v", err)
	//}
	//_, err = gmo.groupClient.Patch(context.TODO(), get.Name, types.StrategicMergePatchType, patchGroupActionStatus, metav1.PatchOptions{})
	//if err != nil {
	//	logs.Errorf("patch patchGroupActions:group err:%v", err)
	//}

	//logs.Info("time----monitor2=============:")

	//将queue_manager和group_manager的group信息进行更新
	//err = gmo.groupQueues.UpdateGroup(get.Status.GroupID, get)
	//if err != nil {
	//	logs.Error("update group-runtiem-start info error")
	//}
	// 修改task的信息
	gmo.taskManager.UpdateTask(task1)
	//测试
	taskPhase := task1.Status.Phase
	p := get.Status.Phase
	actionStatus := get.Status.ActionStatus[0].Phase
	runtimeStatus := get.Status.ActionStatus[0].RuntimeStatus[0].Phase
	logs.Infof("START ：taskStatus:%v,groupStatus:%v, action status: %v, runtime Status: %v", taskPhase, p, actionStatus, runtimeStatus)
	actionStatus1 := get.Spec.Actions[0].Status.Phase
	runs := get.Spec.Actions[0].Status.RuntimeStatus[0].Phase
	logs.Infof("START ：action status1: %v, runtime Status1: %v", actionStatus1, runs)
}

// 收到的Phase为：Successed or Failed
// 处理Runtime运行时结束,如果runtime是最后一个执行完成的，还得同时标记action的phase   总结：所有临时变量赋值时都得使用&
func (gmo *GroupMonitor) handleRuntimeEndUpdate(event events.RuntimeEndPhaseEvent1) {
	// 更新 Runtime 的状态，依据实际变化更新相应字段
	logs.Info("Handling runtime end status update")
	groupName := event.GroupName
	get, err := gmo.groupClient.Get(context.TODO(), groupName, metav1.GetOptions{})
	if err != nil {
		logs.Errorf("Failed get group:%v from etcd, err:%v", groupName, err)
	}
	actionIndex := event.ActionIndex
	runtimeIndex := event.RuntimeIndex
	phase := event.Phase //当前phase可能为Succeed、Failed、Migrating、Migrated
	logs.Infof("Get runtime finish event notify, the phase:%s", phase)
	finshTime := event.FinishAt
	lastTime := event.LastTime
	// Task 信息
	taskID := get.Status.Belongs.TaskID

	//groupID := get.Name + ":" + taskID
	//actionID := a.Name + ":" + get.Status.GroupID
	//runtimeID := r.Name + ":" + actionID
	groupSpec := &get.Spec
	groupStatus := &get.Status

	//修改group下面的groupSpec下面的Actions，Actions下面的ActionStatus，ActionStatus下面的RuntimeStatus
	var allRuntiemCompleted = true  // 当前action是否已经完成（只有action下面的所有的runtime都执行完成了，也就是最后一个runtime被执行完成了，要标记action的状态为succeed，如果说action下面的某一个runtime执行失败，则要标记action装填为Failed）
	var otherActionCompleted = true // group下面的其他Action是否都已经完成
	var nowActionCompleted = false  // group下面的当前Action是否已经完成

	// 修改Task下面的TaskStatus下面的状态  从Task开始遍历的好处是可以修改Task下面的状态
	var otherGroupCompleted = true //标记其他Group的完成情况
	var nowGroupCompleted = false

	list, err := gmo.taskClient.List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		logs.Errorf("Get list task err:%v", err)
	}
	var groupIndexInTask int
	var task1 *apis.Task
	var err2 error
	for _, t := range list.Items { //遍历etcd当中的所有task
		if t.Status.TaskID == taskID { // 如果taskId对上了，则就修改该Task的Phase为checking
			task1, err2 = gmo.taskClient.Get(context.TODO(), t.Name, metav1.GetOptions{})
			if err2 != nil {
				logs.Error("Get task by taskID error from etcd:%v", err2)
			}
		}
	}
	// 检查其他的group是否完成,修改Task的状态
	for i := range task1.Spec.Groups {
		grStatus := &task1.Status.GroupStatus[i]
		if grStatus.GroupID != groupStatus.GroupID { //遍历到的group的ID不等于当前处理的Group的ID
			//logs.Infof("groupStatus.Phase:%v,groupID:%v", grStatus.Phase, grStatus.GroupID)
			if grStatus.Phase == apis.DeployCheck { // 说明其他Group还未执行---这里出现Bug，就是group3的状态为Unknown
				otherGroupCompleted = false
			}
			continue
		}
		groupIndexInTask = i // 当前处理的Group在Task当中的下标
	}
	// 修改group的状态
	for i := range groupSpec.Actions { //Action
		actionStatus := &groupSpec.Actions[i].Status //ActionStatus
		if i != actionIndex {                        //遍历到其他Action，可以顺带看一下别的Action是否都已经完成了
			if actionStatus.Phase == apis.DeployCheck { //其他Action为checking状态，说明还有其他的Action没有被遍历到，Group状态为Running状态
				otherActionCompleted = false
			}
			continue
		}
		for j := range actionStatus.RuntimeStatus { //RuntimeStatus
			rs := &actionStatus.RuntimeStatus[j]
			if j == runtimeIndex {
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
		groupStatus.FinishAt = finshTime
		groupStatus.LastTime = lastTime
		nowGroupCompleted = true
		//遍历task，同时标记TaskStatus下GroupStatus状态也为phase---这个在下面进行统一处理
	}
	//修改group下面的groupStatus下面的ActionStatus，ActionStatus下面的RuntimeStatus
	for i := range groupStatus.ActionStatus { //ActionStatus
		as := &groupStatus.ActionStatus[i]
		if i != actionIndex {
			continue
		}
		for j := range as.RuntimeStatus { //RuntimeStatus
			rs := &as.RuntimeStatus[j]
			if j == runtimeIndex {
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
	//如果说TaskStatus下面的Group都被执行了，还得修改TaskStatus的phase的状态
	//logs.Infof("otherGroupCompleted:%v", otherGroupCompleted)
	//logs.Infof("nowGroupCompleted:%v", nowGroupCompleted)
	if otherGroupCompleted && nowGroupCompleted {
		task1.Status.Phase = phase
		task1.Status.FinishAt = finshTime
		task1.Status.LastTime = lastTime
	}
	// 将修改后的Group状态值赋值给Task
	task1.Status.GroupStatus[groupIndexInTask] = get.Status
	task1.Spec.Groups[groupIndexInTask].Spec = get.Spec
	task1.Spec.Groups[groupIndexInTask].Status = get.Status

	_, err3 := gmo.taskClient.Update(context.TODO(), task1, metav1.UpdateOptions{})
	if err3 != nil {
		logs.Errorf("Update group-runtiem-start info error-3:%v", err3)
		time.Sleep(100 * time.Millisecond)
		_, err3 = gmo.taskClient.Update(context.TODO(), task1, metav1.UpdateOptions{})
	}
	_, err4 := gmo.groupClient.Update(context.TODO(), get, metav1.UpdateOptions{})
	if err4 != nil {
		logs.Errorf("Update group-runtiem-start info error-4:%v", err4) //报错
		time.Sleep(100 * time.Millisecond)
		_, err4 = gmo.groupClient.Update(context.TODO(), get, metav1.UpdateOptions{})
	}
	//将queue_manager和group_manager的group信息进行更新
	//err = gmo.groupQueues.UpdateGroup(get.Status.GroupID, get)
	//if err != nil {
	//	logs.Error("update group-runtiem-end info error")
	//}
	// 修改task的信息
	gmo.taskManager.UpdateTask(task1)
	//测试：
	taskPhase := task1.Status.Phase
	p := get.Status.Phase
	actionStatus := get.Status.ActionStatus[0].Phase
	runStatus1 := get.Status.ActionStatus[0].RuntimeStatus[0].Phase
	logs.Infof("Runtime END: taskStatus:%v, groupStatus:%v, actionStatus:%v, runtimeStatus:%v", taskPhase, p, actionStatus, runStatus1)
	actionStatus1 := get.Spec.Actions[0].Status.Phase
	runtimeStatus1 := get.Spec.Actions[0].Status.RuntimeStatus[0].Phase
	logs.Infof("Runtime END: actionStatus:%v, runtimeStatus:%v", actionStatus1, runtimeStatus1)
}

// 同group_workers当中的方法
func (gmo *GroupMonitor) groupDepenSatisfy(group *apis.Group) bool {
	// TODO：实现依赖检查逻辑
	// 目前只是检查Spec当中的Parents选项
	if len(group.Spec.Parents) == 0 {
		return true
	} else {
		//检查父亲group是否执行完成
		for i := range group.Spec.Parents {
			//parentGroupID := group.Spec.Parents[i]
			parentName := group.Spec.Parents[i]
			//var parentGroupName string
			//// 去client-go当中查group
			//groupList, err := gmo.groupClient.List(context.TODO(), metav1.ListOptions{})
			//if err != nil {
			//	logs.Error("get list group from etcd err:", err.Error())
			//}
			//for _, g := range groupList.Items {
			//	if g.Status.GroupID == parentGroupID {
			//		parentGroupName = g.Name
			//		break
			//	}
			//}
			// TODO 这里得判断这个group和当前的group是否属于同一个Task
			result, err := gmo.groupClient.Get(context.TODO(), parentName, metav1.GetOptions{}) //这里查父亲group的状态，得去etcd当中查
			if err != nil {
				logs.Errorf("Failed to get parent group:%s form etcd, err:%v", parentName, err)
			}
			if result.Status.Phase != apis.Successed {
				return false //说明当前group的付钱group还没完成，直接返回false即可
			}
		}
	}
	return true
}

// 检查Group的依赖是否满足--存的是Parents的Name
func (gmo *GroupMonitor) checkGroupDepencies(group *apis.Group) bool {
	//TODO：实现依赖检查逻辑
	// 目前只是检查Spec当中的Parents选项
	if len(group.Spec.Parents) == 0 {
		return true
	} else {
		//检查父亲group是否执行完成
		for i := range group.Spec.Parents {
			parentName := group.Spec.Parents[i]
			// 去client-go当中查group
			result, err := gmo.groupClient.Get(context.TODO(), parentName, metav1.GetOptions{})
			if err != nil {
				logs.Errorf("Failed to get group:%s", parentName)
			}
			if result.Status.Phase != apis.Successed {
				return false //说明当前group的付钱group还没完成，直接返回false即可
			}
		}
	}
	return true
}

// 检查Action的依赖是否满足
func (gmo *GroupMonitor) actionDepenSatisfy(actionIndex int, group *apis.Group) bool {
	// 得去etcd当中查稳妥一些，还是查action的parents是否完成
	//group, err := gmo.groupClient.Get(context.TODO(), groupName, metav1.GetOptions{})
	//if err != nil {
	//	logs.Errorf("get group err:%v", err)
	//}
	actionSpec := &group.Spec.Actions[actionIndex].Spec
	for i := range actionSpec.Parents { // 遍历当前Action的所有父亲Action
		//actionParentID := actionSpec.Parents[i]
		actionParentName := actionSpec.Parents[i]
		for j := range group.Status.ActionStatus { // 遍历group当中所有的action，先对actionID，然后看这个action的Phase如何
			as := &group.Status.ActionStatus[j]
			a := &group.Spec.Actions[j]
			if a.Name == actionParentName && as.Phase != apis.Successed { //目前定义，Action的父亲Action必须是成功状态
				return false
			}
		}
	}
	return true
}

// 检查Runtime的依赖是否满足
func (gmo *GroupMonitor) runtimeDepenSatisfy(actionIndex, runtimeIndex int, group *apis.Group) bool {
	//TODO runtime运行之前，需要检查parent的runtime是否正常执行完成
	// 得去etcd当中查稳妥一些，还是查action的parents是否完成
	//group, err := gmo.groupClient.Get(context.TODO(), groupName, metav1.GetOptions{})
	//if err != nil {
	//	logs.Errorf("get group err:%v", err)
	//}
	runtime := &group.Spec.Actions[actionIndex].Spec.Runtimes[runtimeIndex]
	for i := range runtime.Parents { // 遍历当前runtime的父亲
		//runtimeParentID := runtime.Parents[i]
		runtimeParentName := runtime.Parents[i]
		for j := range group.Status.ActionStatus[actionIndex].RuntimeStatus {
			rs := &group.Status.ActionStatus[actionIndex].RuntimeStatus[j]
			r := &group.Spec.Actions[actionIndex].Spec.Runtimes[j]
			if r.Name == runtimeParentName && rs.Phase != apis.Successed {
				return false
			}
		}
	}
	return true
}

// 主要处理group一直未完成的情况
func (gmo *GroupMonitor) handleStatusUpdate(group *apis.Group, failed apis.Phase) {
	// 设置当前group的状态为Failed，同时去找Task，标记其状态也为Failed
	time := apis.Time{time.Now()}
	group.Status.Phase = failed
	group.Status.StartAt = time
	group.Status.FinishAt = time
	group.Status.LastTime = time

	taskID := group.Status.Belongs.TaskID

	list, err := gmo.taskClient.List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		logs.Errorf("Get list task from etcd err:%v", err)
	}
	for _, ta := range list.Items {
		if ta.Status.TaskID == taskID {
			task1, err := gmo.taskClient.Get(context.TODO(), ta.Name, metav1.GetOptions{})
			if err != nil {
				logs.Errorf("Get task from etcd err:%v", err)
			}
			task1.Status.Phase = failed
			task1.Status.StartAt = time
			task1.Status.FinishAt = time
			task1.Status.LastTime = time

			for j := range task1.Spec.Groups {
				gr := &task1.Spec.Groups[j]
				if gr.Status.GroupID == group.Status.GroupID {
					task1.Status.GroupStatus[j] = group.Status
				}
			}
			_, err = gmo.taskClient.Update(context.TODO(), task1, metav1.UpdateOptions{})
			if err != nil {
				logs.Errorf("Update task from etcd err:%v", err)
			}
		}
	}
	_, err = gmo.groupClient.Update(context.TODO(), group, metav1.UpdateOptions{})
	if err != nil {
		logs.Errorf("Update group from etcd err:%v", err)
	}
}

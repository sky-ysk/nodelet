package monitor

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"sync"
	"time"

	"hit.edu/framework/pkg/apimachinery/types"
	apis "hit.edu/framework/pkg/apis/cores"
	metav1 "hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/client-go/clients/typed/core"
	"hit.edu/framework/pkg/client-go/tools/recorder"
	"hit.edu/framework/pkg/component-base/logs"
	"hit.edu/framework/pkg/nodelet/events"
	"hit.edu/framework/pkg/nodelet/events/eventbus"
	group "hit.edu/framework/pkg/nodelet/task/group"
	"hit.edu/framework/pkg/nodelet/task/group/dependency"
	"hit.edu/framework/pkg/nodelet/task/runtime"
	"hit.edu/framework/pkg/nodelet/task/task"
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
	// eventRecorder 记录事件
	recorder recorder.EventRecorder
	// 管理运行所需的Runtime
	// 存储RuntimeManager
	runtimeManager *runtime.RuntimeManager
	//管理依赖
	dependencyManager *dependency.DependencyManager
	//Client-go
	nodesClient  core.NodeInterface //需要查node信息
	groupClient  core.GroupInterface
	taskClient   core.TaskInterface
	actionClient core.ActionInterface
	stopCh       chan struct{}
}

func NewGroupMonitor(groupManager group.Manager, taskManager task.Manager, groupQueues *group.GroupQueues, eventbus *eventbus.EventBus, recorder recorder.EventRecorder, runtimeManager *runtime.RuntimeManager, nodeClient core.NodeInterface, groupClient core.GroupInterface, taskClient core.TaskInterface, dependencyManager *dependency.DependencyManager, actionClient core.ActionInterface) *GroupMonitor {
	return &GroupMonitor{
		groupManager:      groupManager,
		taskManager:       taskManager,
		groupQueues:       groupQueues,
		eventBus:          eventbus,
		recorder:          recorder,
		runtimeManager:    runtimeManager,
		dependencyManager: dependencyManager,
		nodesClient:       nodeClient,
		groupClient:       groupClient,
		taskClient:        taskClient,
		actionClient:      actionClient,
		stopCh:            make(chan struct{}),
	}
}

func (gmo *GroupMonitor) Start(ctx context.Context) {
	//TODO 轮询检查队列当中的内容
	logs.Info("GroupMonitor component start")

	//订阅事件
	chRuntimeStart := make(chan interface{})
	chRuntimeEnd := make(chan interface{})

	gmo.eventBus.Subscribe(reflect.TypeOf(events.RuntimeStartPhaseEvent1{}), chRuntimeStart)
	gmo.eventBus.Subscribe(reflect.TypeOf(events.RuntimeEndPhaseEvent1{}), chRuntimeEnd)

	//启动环境的依赖检查与更新
	depenUpdateDone := make(chan struct{})
	go func() {
		defer close(depenUpdateDone)
		ticker := time.NewTicker(5 * time.Second)
		//循环检查更新依赖，有两个内容要更新：所有虚拟环境的名字；每个虚拟环境所包含的所有包
		for {
			select {
			case <-ctx.Done(): // 如果父进程通知关闭
				logs.Info("依赖检查协程收到关闭通知，正在退出...")
				return // 退出协程
			case <-ticker.C: // 每隔一段时间执行一次更新依赖操作
				logs.Info("定期检查机器的虚拟环境依赖")
				gmo.dependencyManager.UpdateEnvs()
				gmo.dependencyManager.UpdateEnvPackages()
			}
		}
	}()

	//启动监听事件（支持 Context 退出）
	eventLoopDone := make(chan struct{})
	go func() {
		defer close(eventLoopDone)
		for {
			select {
			case <-ctx.Done():
				logs.Info("GroupMonitor exiting due to context cancel")
				return
			case event := <-chRuntimeStart:
				RuntimeEvent := event.(events.RuntimeStartPhaseEvent1)
				gmo.handleRuntimeStartUpdate(RuntimeEvent)
			case event := <-chRuntimeEnd:
				RuntimeEvent := event.(events.RuntimeEndPhaseEvent1)
				gmo.handleRuntimeEndUpdate(RuntimeEvent)
			}
		}
	}()
	// 启动状态检查协程（支持 Context 退出）
	statusCheckDone := make(chan struct{})
	go func() {
		defer close(statusCheckDone)
		gmo.CheckStatus(ctx) // 修改 CheckStatus 方法以接受 Context
	}()
	// 等待 Context 取消或所有协程退出
	select {
	case <-ctx.Done():
		logs.Info("GroupMonitor exiting due to context cancel")
	case <-eventLoopDone:
	case <-statusCheckDone:
	}
	// 等待所有子协程退出
	<-depenUpdateDone
	<-eventLoopDone
	<-statusCheckDone
}

func (gm *GroupMonitor) Stop() {
	close(gm.stopCh)
	logs.Info("TaskExporter Monitor stopped")
}

// 检查checking队列的任务，前置任务是否完成，看是否需要迁移到running队列
// 检车copyPending队列的任务，对于副本任务，Checking完成后进入等待，当副本任务需要启动时，才正式迁移到Running队列
// 检查running队列，看任务是否还在执行、是否执行完成、、进程占用资源量    任务完成和任务失败下线该如何判断呢？
// 检查error队列，检查任务是否出错，看能否尝试拉起，多次尝试拉起失败后，重新提交给调度器
func (gmo *GroupMonitor) CheckStatus(ctx context.Context) {
	var wg sync.WaitGroup
	wg.Add(6)

	// 启动所有队列检查（传递 Context）
	go func() { defer wg.Done(); gmo.CheckingQueueCheck(ctx) }()
	go func() { defer wg.Done(); gmo.RunningQueueCheck(ctx) }()
	go func() { defer wg.Done(); gmo.CopyPendingQueueCheck(ctx) }()
	go func() { defer wg.Done(); gmo.CompletedQueueCheck(ctx) }()
	go func() { defer wg.Done(); gmo.ErrorQueueCheck(ctx) }()
	go func() { defer wg.Done(); gmo.MigratedQueueCheck(ctx) }()

	// 等待所有检查协程退出
	wg.Wait()
}

// 都要改成for i：=range
func (gmo *GroupMonitor) CheckingQueueCheck(ctx context.Context) { //主要针对Task下的多个Group在多个设备上运行，group之间有依赖关系，需要检查
	// TODO 轮询检查Checking队列，检查任务group的依赖是否满足，如果满足才放入running队列当中
	logs.Info("Pending queue start checking")
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Millisecond * 500):
			checkingGroups := gmo.groupQueues.GetAllChecking()
			for i := range checkingGroups {
				gr := checkingGroups[i]
				// 从etcd当中读取group信息
				getGroup, err := gmo.groupClient.Get(context.TODO(), gr.Name, metav1.GetOptions{})
				if err != nil {
					logs.Errorf("Etcd get group error-1:%v", err)
				}
				if !gmo.groupDepenSatisfy(getGroup) { //再次检查group的执行依赖是否满足了（注意：group_workers当中任务头一次执行前也会检查）
					//logs.Debugf("The group ：%s execution dependency is not satisfied again, now still in Checking Queue", gr.Name)
					//继续放在checking队列当中，checking队列会持续检查依赖，直到依赖满足后，才开始执行，重新将任务group交给group_workers去执行
					getGroup.Status.CheckDependencyCount++
					if getGroup.Status.CheckDependencyCount > 10000 { // 当检查依赖的次数大于1000次的话，说明依赖还是满足不了，迁移至Error队列---这里其实有问题（group如果有前序依赖，你不知道什么时候其前序依赖能完成），暂停1000s其实也是有问题的
						//将任务迁移到Error队列当中
						ok := gmo.groupQueues.DeleteFromCheckingAndAddToError(getGroup.Status.GroupID)
						if !ok {
							logs.Error("Delete group from checking queue and add to error queue failed")
						}
						gmo.handleStatusUpdate(getGroup, apis.Failed) //设置group状态为
						continue
					}
					// TODO 这里后期得优化
					patchGroup, err := json.Marshal(map[string]interface{}{
						"status": map[string]interface{}{
							"check_dependency_count": getGroup.Status.CheckDependencyCount,
						},
					})
					if err != nil {
						logs.Errorf("Json Marshal failed, err:%v", err)
					}
					_, err = gmo.groupClient.Patch(context.TODO(), getGroup.Name, types.StrategicMergePatchType, patchGroup, metav1.PatchOptions{})
					if err != nil {
						logs.Errorf("Patch group error-1:%v", err)
					}
					continue
				} else { //说明group执行的依赖已经满足，接下来开始执行
					// 任务依赖满足后就将任务从checking队列转移至Running队列，为了适配迁移，同时适配副本任务,若为副本任务，则转移到CopyPending队列当中
					if getGroup.Spec.IsCopy {
						ok := gmo.groupQueues.DeleteFromCheckingAndAddToCopyPending(getGroup.Status.GroupID)
						if !ok {
							logs.Error("Delete group from checking queue and add to copy pending queue failed")
						}
					} else { //非副本任务，则直接移入到Running队列当中去运行任务
						ok := gmo.groupQueues.DeleteFromCheckingAndAddToRunning(getGroup.Status.GroupID)
						if !ok {
							logs.Error("Delete group from checking queue and add to running queue failed")
						}
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

// 对于副本group，不设置任何时间，除非group启动了，才开始设置时间
// 检查副本任务的队列，做的事情：①如果当前副本任务的状态被标记为Waiting，则副本继续等待，如果说副本任务状态标记为Stating，则副本任务马上启动   ②如果说源任务（Action 、Runtime）Running了，会设置副本任务当中的（Action或Runtime的）copy_status为Running，同理如果为源任务为Failed，设置副本任务为Failed
func (gmo *GroupMonitor) CopyPendingQueueCheck(ctx context.Context) { //TODO 对于专门存放副本的队列，目前暂时用到actionDepenSatisfy方法和runtimeDepencySatisfy方法，默认只要源任务满足，副本任务一定可以满足，后期有的话再加进去
	logs.Info("Copy Pending queue start checking")
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Millisecond * 100):
			copyPendingGroups := gmo.groupQueues.GetAllCopyPending()
			for i := range copyPendingGroups {
				gro := copyPendingGroups[i]
				// 从etcd当中读取group信息
				group, err := gmo.groupClient.Get(context.TODO(), gro.Name, metav1.GetOptions{})
				if err != nil {
					logs.Errorf("Etcd get group error-2:%v", err)
				}
				if group.Status.CopyStatus == "Waiting" { // 说明副本任务是提前部署好的
					// 这里打算Init初始化group,就是提前进行Running步骤  源任务一个Runtime执行完成后，就修改副本runtime的状态即可，Action执行完成后，也会修改副本Runtime的状态
					var isSuccess bool                                   // 标记group下面的action是否都执行成功，如果都执行完了，还没有触发迁移，那么关闭副本即可
					for actionIndex := range group.Status.ActionStatus { // 遍历group当中的Action
						actionStatus := &group.Status.ActionStatus[actionIndex]
						action := &group.Spec.Actions[actionIndex]
						isSuccess = true
						if actionStatus.CopyStatus == "Running" { //说明源任务已经启动了，那么这里需要遍历runtime，找到源任务当中运行的runtime，然后init初始化runtime
							isSuccess = false
							grou, err := gmo.groupManager.GetGroupByID(group.Status.GroupID) // 目前打算把一些小的参数存到本地内存当中的groupManager当中，这样可以减轻访问api-server的压力
							if err != nil {
								logs.Errorf("Get group by id failed, err:%v", err)
							}
							for runtimeIndex := range actionStatus.RuntimeStatus {
								runtimeStatus := &actionStatus.RuntimeStatus[runtimeIndex]
								runtime := &action.Spec.Runtimes[runtimeIndex]
								if runtimeStatus.CopyStatus == "Running" { //说明源任务当中的该runtime已经Running了
									// 这里将副本group中的该runtime进行判断，如果是细粒度控制的，就进行init
									if runtime.EnableFineGrainedControl && !grou.Status.ActionStatus[actionIndex].RuntimeStatus[runtimeIndex].Initing { //细粒度控制 且 还未Init初始化过
										// 启动runtimeStatus的Init方法  --TODO 这里为啥不用依赖检查呢？因为源任务能Running，说明这个runtime是依赖是满足的，所以默认副本对应的runtime依赖也是满足的（所以这里得加一点，就是在init阶段，检查一下依赖再init也OK---最好是这样）
										logs.Info("#############################Init#######################################")
										go gmo.runtimeManager.InitRuntime(group, action, runtime, actionIndex, runtimeIndex) // TODO init方法当中最好也能发送一个事件
										grou.Status.ActionStatus[actionIndex].RuntimeStatus[runtimeIndex].Initing = true     // 标记该runtime是细粒度控制，且开启了Init初始化，因为对于细粒度控制的runtime，如果没有预部署，直接切换的话，不会调用Init方法？
										grou.Status.ActionStatus[actionIndex].RuntimeStatus[runtimeIndex].Waiting = true     // 这个参数标记用来防止接下来进入Running队列的Init检查的时候，Runtime被执行多次，在启动Runtime后，将Waiting属性置为false就能防止Runtime被执行多次了
									} else { // 如果不是细粒度的，那么就标记该runtime的Waiting属性为Waiting（其状态仍然是DeployCheck）
										grou.Status.ActionStatus[actionIndex].RuntimeStatus[runtimeIndex].Waiting = true // TODO应该是为了适配进入到Running队列的DeployCheck检查后，会重复执行（因为当前状态为DeployCheck状态，进入到Running队列，有可能还是DeployCheck状态，对于DeployCheck状态，需要考虑该runtime是否处于等待的过程），这里的runtime，其实就是处于一种等待的过程
									}
								}
								if runtimeStatus.CopyStatus == "Succeeded" { // 说明源任务当中的runtime已经运行完成了
									// 这里将副本group中的该runtime进行判断，如果是细粒度控制的，且进行了初始化的工作话，就关闭Init初始化
									if runtime.EnableFineGrainedControl && grou.Status.ActionStatus[actionIndex].RuntimeStatus[runtimeIndex].Initing {
										// 关闭runtime，直接关闭runtime进程
										logs.Info("(((((((((((((((((((((((((((((((((((((((")
										gmo.runtimeManager.Kill(group, action, runtime, actionIndex, runtimeIndex)
									} else { // 如果说runtime不是细粒度的，那这里源任务完成后，副本runtime的状态是不会主动修改的（因为副本runtime没启动无法调用Kill函数来修改runtime状态），所以这里需要主动修改runtime的状态为Succeed
										gmo.handleRuntimeSucceedUpdate(group, actionIndex, runtimeIndex)
									}
								}
							}
							continue // 这样能快速遍历下一个Action，不然还会进入下面的判断，稍微好一丢丢
						}
						if actionStatus.CopyStatus == "Succeeded" { // 这里有个小插曲，就是对于副本任务里面改Action，其下面的Runtime的状态没有改为Succeed，可以补充进来--  这是为啥呢？因为这里是一层层遍历，虽然说源任务runtime完成、Action完成会同时修改副本的runtime、Action，但是由于这里的逻辑是先遍历到Action，然后再遍历到下面的runtime，这里选择不再遍历下去，这样会很，直接遍历到Action状态为Successed，然后调用一个方法将Action下面的所有runtime的Phase改为Succeed即可
							// 这里有一个问题，就是对于Action下面有细粒度控制的runtime，那么runtimeza在Succeed后也会修改Action的状态为Succeed，所以这里是否要判断一下？其实也没事情，顶多就是重复操作了，问题不大
							// 将etcd当中当前副本的action状态改为Succeed，还需要将action下面的最后一个执行完成的Runtime的Phase修改为Succeed---其实这块也没有完全修改的必要，因为遍历是从Action遍历到Runtime，如果Action的Phase为Succeed，默认下面的所有Runtime的Phase为Succeed，这里以后有时间可以添加上
							//groupCopyName := "Reason-Copy"
							//getCopyGroup, err := gmo.groupClient.Get(context.TODO(), groupCopyName, metav1.GetOptions{})
							//if err != nil {
							//	logs.Errorf("Get group by id failed, err:%v", err)
							//}
							gmo.handleCopyActionSucceedUpdate(group, actionIndex) //修改副本任务的Action、Runtime为Successed的状态（因为源任务的Action、Runtime执行完了，副本任务的Action、Runtime没必要执行，直接设置phase为成功即可）
							continue                                              //说明当前Action执行完成了，接着查看下一个Action的执行情况
						}
						if actionStatus.CopyStatus == "Failed" { //TODO 这里有个小插曲，就是对于副本任务里面的其他Action、Runtime的状态没有改为Failed，以后可以补充进来
							isSuccess = false
							// 修改副本group的状态为Failed
							//groupCopyName := "Reason-Copy"
							//getCopyGroup, err := gmo.groupClient.Get(context.TODO(), groupCopyName, metav1.GetOptions{})
							//if err != nil {
							//	logs.Errorf("Get group by id failed, err:%v", err)
							//}
							gmo.handleCopyRuntimeFailedUpdate(group, actionIndex) // 源任务执行失败了，直接标记副本任务的runtime、action状态为Failed，那么副本group的状态也直接被标记为Failed
							logs.Info("@@@@@@@@@@@@@@@@@@@@@@@@@@@@@!=2")
							ok := gmo.groupQueues.DeleteFromCopyPendingAndAddToCompleted(group.Status.GroupID) // 源任务失败后被移入Error队列来通知调度器，这里副本任务直接移入成功队列即可
							if !ok {
								logs.Error("Delete group from copypending queue and add to completed pending queue failed")
							}
						}
					}
					if isSuccess { //说明源任务，还没有迁移就完成了全部的工作，那么直接将副本迁移到Completed队列就OK了
						// 这里还需要把副本任务的Group状态设置为Succeed
						// 修改副本group的状态为Succeed
						//groupCopyName := "Reason-Copy"
						patchGroup, err := json.Marshal(map[string]interface{}{
							"status": map[string]interface{}{
								"phase": apis.Successed,
							},
						})
						if err != nil {
							logs.Errorf("Json Marshal failed, err:%v", err)
						}
						_, err = gmo.groupClient.Patch(context.TODO(), group.Name, types.StrategicMergePatchType, patchGroup, metav1.PatchOptions{})
						if err != nil {
							logs.Errorf("Patch group error-9:%v", err)
						}
						//将任务迁移到Completed队列当中
						logs.Info("Move to completed queue")
						logs.Info("@@@@@@@@@@@@@@@@@@@@@@@@@@@@@!=1")
						ok := gmo.groupQueues.DeleteFromCopyPendingAndAddToCompleted(group.Status.GroupID)
						if !ok {
							logs.Errorf("Delete group from running queue and add to completed queue failed")
						}
						continue
					}
				} else if group.Status.CopyStatus == "Starting" {
					//logs.Infof("=============CopyPending--Starting")
					ok := gmo.groupQueues.DeleteFromCopyPendingAndAddToRunning(group.Status.GroupID) // 有两种情况，一种是无副本情况的瞬时迁移，另一种是副本任务触发了迁移
					if !ok {
						logs.Error("Delete group from copy checking queue and add to running queue failed")
					}
				}
			}
			//default:
			//case <-ctx.Done():
			//	return
		}
	}
}

// 检查Running队列，做的事情：①如果发现任务完成，迁移到Completed队列，如果发现任务失败，迁移到Error队列
// ②检查runtime、Action当中的parents是否执行完成，如果完成，则执行
func (gmo *GroupMonitor) RunningQueueCheck(ctx context.Context) { //主要针对当前设备上的Group，下面有多个Action，之间有依赖关系，需要检查
	//TODO 监控进程的返回值等判断任务是否正常执行完成，正常则放入completedqueue，否则放入errorqueue(方法待确认)
	logs.Info("Running queue start checking")
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Millisecond * 100):
			runningGroups := gmo.groupQueues.GetAllRunning()
			for i := range runningGroups {
				gro := runningGroups[i] //不用再加&&
				// 从etcd当中读取group信息
				group, err := gmo.groupClient.Get(context.TODO(), gro.Name, metav1.GetOptions{})
				if err != nil {
					logs.Errorf("Etcd get group error-3:%v", err)
				}
				var isSuccess = true                          // 标记group下面的action是否都执行成功
				for actionIndex := range group.Spec.Actions { // 遍历group当中的Action
					action := &group.Spec.Actions[actionIndex]
					actionStatus := &group.Status.ActionStatus[actionIndex]
					// 为了适配迁移，状态为Migrated也说明Action成功结束了，然后接下来就通过Action成功标记Group成功了
					if action.Status.Phase == apis.Successed || action.Status.Phase == apis.Migrated { // 当前action的状态为Successed或者Migrated
						continue //说明当前Action执行完成了，接着查看下一个Action的执行情况
					}
					if action.Status.Phase == apis.Failed { //注意:handleRuntimeEndUpdate方法当中，当Runtime状态失败时，同时也会修改其Action的状态为Failed，同时也会修改Group为Failed、Task为Failed
						isSuccess = false
						//将任务迁移到Error队列当中
						ok := gmo.groupQueues.DeleteFromRunningAndAddToError(group.Status.GroupID)
						if !ok {
							logs.Errorf("Delete group from running queue and add to running queue failed")
						}
						break // 这里直接跳出for循环即可，因为该group已经是Failed了，不用看了
					}
					grou, err := gmo.groupManager.GetGroupByID(group.Status.GroupID)
					if err != nil {
						logs.Errorf("Get group by id from group_manager err:%v", err)
					}
					if action.Status.Phase == apis.DeployCheck { //
						isSuccess = false
						if grou.Status.ActionStatus[actionIndex].Waiting == true { //说明是第二次遍历到了这个Action，第一次遍历到该Action的时候，其依赖没有满足
							if !gmo.actionDepenSatisfy(actionIndex, group) {
								//logs.Infof("Action %s depends on parent action, parent not finish ", action.Name)
								continue
							}
							grou.Status.ActionStatus[actionIndex].Waiting = false // 说明Action的父亲Action已经执行完成了，那么接下来Action的Runtime必须会被执行（至少会执行一个runtime）
						} else { //说明是第一次遍历到了这个Action，当Action的依赖没有满足的时候，要设置Waiting属性为true
							if !gmo.actionDepenSatisfy(actionIndex, group) {
								logs.Infof("Action:%s in group:%s waiting for dependencies", action.Name, group.Name)
								//action.Status.Waiting = true //第一次执行时发现执行不了，那就交给running队列去检查
								grou.Status.ActionStatus[actionIndex].Waiting = true
								logs.Debugf("StartGroup method:ActionWaiting:%v, i:%v", grou.Status.ActionStatus[actionIndex].Waiting, i)
								continue
							}
						}
						for runtimeIndex := range action.Spec.Runtimes {
							runtime := &action.Spec.Runtimes[runtimeIndex]
							runtimeStatus := &actionStatus.RuntimeStatus[runtimeIndex]
							if grou.Status.ActionStatus[actionIndex].RuntimeStatus[runtimeIndex].Waiting == true { // 说明是第二次遍历到这个runtime，第一次遍历到该runtime的时候，其依赖没有满足（判断Runtime是否处于等待）
								if !gmo.runtimeDepenSatisfy(actionIndex, runtimeIndex, group) { // runtime 依赖不满足
									continue
								}
								grou.Status.ActionStatus[actionIndex].RuntimeStatus[runtimeIndex].Waiting = false //runtime依赖已经满足，此时设置为false，就不会继续往下执行，去启动任务了，这里的设置很关键
							} else {
								if !gmo.runtimeDepenSatisfy(actionIndex, runtimeIndex, group) { // runtime 依赖不满足
									grou.Status.ActionStatus[actionIndex].RuntimeStatus[runtimeIndex].Waiting = true // 设置Runtime处于等待
									continue
								}
							}
							//说明runtime可以执行，这里为了适配迁移，如果是副本group，初始化任务的时候直接使用关键状态数据
							// 这块可以执行到，因为group当中有很多action，有多个Action的话，总有没执行的Action，这时候需要判断runtime的启动方式（细粒度的话使用StartingRuntime启动、粗粒度的话使用Run启动）
							if runtime.EnableFineGrainedControl { // 当前group是副本任务，且实现了细粒度控制方法
								logs.Infof("****************************hhhhhhhhhhhhhhhh****************************************")
								if !grou.Status.ActionStatus[actionIndex].RuntimeStatus[runtimeIndex].Starting { //TODO 这个参数好像可以删了，有Waiting是不是就够了？
									logs.Infof("========================runtimeStatus.KeyStatus:%v,runtimeStatus.KeyStatus == \"\"", runtimeStatus.KeyStatus, runtimeStatus.KeyStatus == "")
									if runtimeStatus.KeyStatus == "" { // 说明不是副本任务，还没初始化---TODO 这里需要这个检查的原因：有可能是即时的迁移迁移，那迁移过去的group是没有进入init状态的，所以这边迁移过去的副本是处于DeployCheck的状态开始恢复任务状态
										go gmo.runtimeManager.StartRuntime(group, action, runtime, actionIndex, runtimeIndex)
									} else {
										go gmo.runtimeManager.StartRuntime(group, action, runtime, actionIndex, runtimeIndex) //这句好像会阻塞
										go gmo.runtimeManager.RestoreData(group, action, runtime, actionIndex, runtimeIndex)
									}
									grou.Status.ActionStatus[actionIndex].RuntimeStatus[runtimeIndex].Starting = true
								}
							} else {
								// ①副本任务，但没有细粒度控制 ②原任务（没有副本） 采用Run方式启动任务
								//logs.Infof("****************************ashdkhaskldhklashdk****************************************")
								if !grou.Status.ActionStatus[actionIndex].RuntimeStatus[runtimeIndex].Starting {
									go gmo.runtimeManager.Run(group, action, runtime, actionIndex, runtimeIndex)
									grou.Status.ActionStatus[actionIndex].RuntimeStatus[runtimeIndex].Starting = true
								}
							}
						}
						continue
					}
					//检查runtime、Action当中的parents是否执行完成，如果父亲节点完成，则让他执行 ---这里有bug，就是任务已经放入running队列，但是还没执行完，这时候runningCheck循环遍历到当前runtime的状态为checking，查看是否满足执行条件，发现是满足的，结果有跑起来该任务
					if action.Status.Phase == apis.Running {
						isSuccess = false
						for runtimeIndex := range action.Spec.Runtimes {
							runtime := &action.Spec.Runtimes[runtimeIndex]
							if !gmo.runtimeDepenSatisfy(actionIndex, runtimeIndex, group) {
								//logs.Infof("Runtime %s depends on parent runtime", r.Name)
								//r.Waiting = true  // 这里不需要再标记了，因为在DeployCheck阶段就遍历了所有的runtime并标记了
								continue
							}
							grou, err := gmo.groupManager.GetGroupByID(group.Status.GroupID)
							if err != nil {
								logs.Errorf("Get group by id from group_manager err:%v", err)
							}
							//logs.Infof("$$$$$$==================grou.Spec.Actions[j].Spec.Runtimes[m].Waiting:%v,j:%v,m:%v", grou.Spec.Actions[j].Spec.Runtimes[m].Waiting, j, m)
							if grou.Status.ActionStatus[actionIndex].RuntimeStatus[runtimeIndex].Waiting { //如果说runtime也是被标记等待执行的状态，这才能开始执行，这块是防止runtime重复多次运行
								//说明runtime可以执行
								//grou.Spec.Actions[actionIndex].Spec.Runtimes[runtimeIndex].Waiting = false
								grou.Status.ActionStatus[actionIndex].RuntimeStatus[runtimeIndex].Waiting = false
								if runtime.EnableFineGrainedControl {
									logs.Info("****************************************ABCDSDSADSAD**************************")
									if !grou.Status.ActionStatus[actionIndex].RuntimeStatus[runtimeIndex].Starting {
										go gmo.runtimeManager.StartRuntime(group, action, runtime, actionIndex, runtimeIndex)
										grou.Status.ActionStatus[actionIndex].RuntimeStatus[runtimeIndex].Starting = true
									}
								} else {
									logs.Info("****************************************1234554564**********************")
									if !grou.Status.ActionStatus[actionIndex].RuntimeStatus[runtimeIndex].Starting {
										go gmo.runtimeManager.Run(group, action, runtime, actionIndex, runtimeIndex)
										grou.Status.ActionStatus[actionIndex].RuntimeStatus[runtimeIndex].Starting = true
									}
								}
							}
						}
						continue
					}
					//之后这里要考虑迁移的情况，遇到Action为Init的情况，就将Init的runtime启动run/Start起来
					if action.Status.Phase == apis.Init { //说明当前action下面有Init的runtime了（即：有细粒度控制的runtime）
						isSuccess = false
						for runtimeIndex := range action.Spec.Runtimes {
							runtime := &action.Spec.Runtimes[runtimeIndex]
							runtimeStatus := &action.Status.RuntimeStatus[runtimeIndex]
							grou, err := gmo.groupManager.GetGroupByID(group.Status.GroupID)
							if err != nil {
								logs.Errorf("Get group by id failed, err:%v", err)
							}
							if !gmo.runtimeDepenSatisfy(actionIndex, runtimeIndex, group) { // 这里需要runtime的父亲节点的状态也为Successed，所以源runtime成功后，同样需要将副本runtime的Phase设置为Succeed，这个很关键
								grou.Status.ActionStatus[actionIndex].RuntimeStatus[runtimeIndex].Waiting = true
								continue
							}
							// 当前遍历到的runtime肯定有Init，当然也会有DeployCheck，执行到此处表示满足依赖，在上面加入一个依赖检查就完美了
							if runtimeStatus.Phase == apis.Init {
								if grou.Status.ActionStatus[actionIndex].RuntimeStatus[runtimeIndex].Waiting { // 当前runtime等待启动，这里也是防止多次运行这个runtime，所以上面copyPending队列当中InitRuntime方法下面的Waiting=true是必要的
									grou.Status.ActionStatus[actionIndex].RuntimeStatus[runtimeIndex].Waiting = false // 标记当前runtime马上要启动了，不再Waiting了
									// 将初始化的runtime真正启动 为了适配迁移，这里先回复runtime的状态
									logs.Infof("****************************Restore****************************************")
									go gmo.runtimeManager.RestoreData(group, action, runtime, actionIndex, runtimeIndex)
									//// 将runtime真正的启动
									//time.Sleep(1 * time.Second)
									//logs.Infof("****************************&&&&&&&&&&&&&&&&&&&&&&&&&****************************************")
									//go gmo.runtimeManager.StartRuntime(group, action, runtime, actionIndex, runtimeIndex)
								}
							} else if runtimeStatus.Phase == apis.DeployCheck { // 也有可能碰到的是Init-->Succeed
								if !runtime.EnableFineGrainedControl { // 执行到这块，只能是粗粒度控制的runtime
									if grou.Status.ActionStatus[actionIndex].RuntimeStatus[runtimeIndex].Waiting {
										grou.Status.ActionStatus[actionIndex].RuntimeStatus[runtimeIndex].Waiting = false
										// 如果runtime不是细粒度控制的话，直接调用Run启动
										logs.Infof("****************************mmmmmmmmmmmmmmmmmmmmmmmmmmmmmmm****************************************")
										go gmo.runtimeManager.Run(group, action, runtime, actionIndex, runtimeIndex)
									}
								}
							}
						}
					}
				}
				if isSuccess {
					//将任务迁移到Completed队列当中
					logs.Info("Move to completed queue")
					ok := gmo.groupQueues.DeleteFromRunningAndAddToCompleted(group.Status.GroupID)
					if !ok {
						logs.Errorf("Delete group from running queue and add to completed queue failed")
					}
				}
			}
			//default:
			//case <-ctx.Done():
			//	return
		}
	}
}

// 迁移控制的时候，正在运行的任务触发迁移，会迁移到该队列当中，该队列主要是存放源任务的，去监听副本任务是否执行完成，目前的逻辑可能就是只要有一个副本任务执行完成。就算完成了--TODO 后续可能要改成查Condition，是一个副本完成就OK了，还是说得所有副本完成，原任务才OK
func (gmo *GroupMonitor) MigratedQueueCheck(ctx context.Context) {
	logs.Info("Migrated queue start checking")
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Second * 1):
			//var start = false
			migrated := gmo.groupQueues.GetAllMigrated()
			for i := range migrated {
				gro := migrated[i]
				// 从etcd获取group信息
				group, err := gmo.groupClient.Get(context.TODO(), gro.Name, metav1.GetOptions{})
				if err != nil {
					logs.Errorf("Etcd get group error-4:%v", err)
				}
				for key := range group.Spec.CopyInfo {
					copyGroupName := key
					// 查询副本group的状态是否完成，如果完成了，就将group迁移到Completed队列
					var getGroup *apis.Group
					if group.Spec.CopyInfo[key] == "local" { // 如果副本部署在本域当中
						getGroup, err = gmo.groupClient.Get(context.TODO(), copyGroupName, metav1.GetOptions{})
						if err != nil {
							logs.Errorf("Get copy group err:%v", err)
						}

					} else {
						// TODO 通过跨域的连接取到getGroup

					}
					// 副本任务执行完成，为Succeed状态，则标记源任务copy_status为Succeed，如果说为Failed状态，则标记原任务copy_status为Failed状态
					if getGroup.Status.Phase == apis.Successed {
						gmo.handleTaskSucceedUpdate(group) //设置当前group的状态(copyStatus的值)为Succeed，标记其副本任务执行成功了，同时检查group所属Task下面的所有group（除了当前的group）是否都已经成功了，如果都已经成功了，就修改当前Group所属的Task的phase为Succeed
						logs.Info("@@@@@@@@@@@@@@@@@@@@@@@@@@@@@!=4")
						gmo.groupQueues.DeleteFromMigratedAndAddToCompleted(group.Status.GroupID)
					} else if getGroup.Status.Phase == apis.Failed { // 说明group迁移过去执行失败了，那么需要直接将Task的状态
						// 这里得将group移到error队列当中  注意：如果源任务迁移了，但是迁移的副本任务执行失败了，那么这里将Task的状态改为Failed，同时副本那边会将group移到Error队列上报上去给调度器
						gmo.handleTaskFailedUpdate(group) // 设置当前group的状态（copyStatus的值）为Failed
						logs.Info("@@@@@@@@@@@@@@@@@@@@@@@@@@@@@!=5")
						gmo.groupQueues.DeleteFromMigratedAndAddToCompleted(group.Status.GroupID)
					}
					// 如果上述两个分支都没有执行，说明副本group还在执行
					break // 这里是只针对遍历到的第一个副本group
				}
			}
			//default:
			//case <-ctx.Done():
			//	return
		}
	}
}
func (gmo *GroupMonitor) CompletedQueueCheck(ctx context.Context) {
	logs.Info("Completed queue start checking")
	//TODO 可能主要是将信息上传到api-server当中，然后将group_manager中的信息删除
	for {
		select {
		case <-ctx.Done():
			return
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
			//default:
			//case <-ctx.Done():
			//	return
		}
	}
}
func (gmo *GroupMonitor) ErrorQueueCheck(ctx context.Context) {
	logs.Info("Error queue start checking")
	//TODO 可能要做的就是通知调度器，group部署失败
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Second * 100):
			// 通知调度器
			errored := gmo.groupQueues.GetAllError()
			for i := range errored {
				gro := errored[i]
				// 从etcd获取group信息
				group, err := gmo.groupClient.Get(context.TODO(), gro.Name, metav1.GetOptions{})
				if err != nil {
					logs.Errorf("Etcd get group error-4:%v", err)
				}
				//TODO 通知调度器
				gmo.recorder.Event(group, apis.EventTypeWarning, events.GroupRunError, fmt.Sprintf("Group name:%v run error", group.Name))
				// 删除内存当中group_manager当中的group信息
				gmo.groupManager.DeleteGroup(gro) //groupManager就删除group的信息，此时group的信息就只存在于etcd当中
			}
			//default:
			//case <-ctx.Done():
			//	return
		}
	}
}

// 收到的Phase为：Running or Failed or Init(new add)
// 处理 Runtime运行时启动，如果runtime是action下的首个执行的runtime，同时标记action的状态为Running，如果是第一个action启动，则group的状态也标记为running
func (gmo *GroupMonitor) handleRuntimeStartUpdate(event events.RuntimeStartPhaseEvent1) {
	// 更新 Runtime 的状态，依据实际变化更新相应字段
	logs.Info("Handling runtime start status update")
	groupName := event.GroupName
	processId := event.ProcessId
	getGroup, err := gmo.groupClient.Get(context.TODO(), groupName, metav1.GetOptions{})
	if err != nil {
		logs.Errorf("Failed get group:%v from etcd, err:%v", groupName, err)
	}

	actionIndex := event.ActionIndex
	runtimeIndex := event.RuntimeIndex
	phase := event.Phase // 这里接收的Phase有可能是running，也有可能是Failed
	logs.Infof("Get runtime start event notify, the phase:%s", phase)
	startTime := event.StartAt
	lastTime := event.LastTime
	groupSpec := &getGroup.Spec
	groupStatus := &getGroup.Status
	//var actionStart = true              //action是否需要标记启动（下面的runtime如果都没启动，则说明action要标记Running）
	taskID := getGroup.Status.Belongs.TaskID // group的Belongs属性当中的TaskID
	// 是否为副本任务
	isCopyGroup := getGroup.Spec.IsCopy
	var taskIsFirstSet = false  // 为了适配发送最全的Task信息
	var groupIsFirstSet = false // 为了适配发送最全的Group信息

	var groupIndexInTask int //当前group在Task当中的下标
	var task *apis.Task      //group所属的Task对象
	if !isCopyGroup {        // 不是副本group，要修改所属的Task的状态
		// 首先先修改该runtime所对应的group-所对应的Task的phase
		taskList, err := gmo.taskClient.List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			logs.Errorf("Get list task err:%v", err)
		}
		var err2 error
		for _, t := range taskList.Items { //遍历etcd当中的所有task，找到当前group所属的Task
			if t.Status.TaskID == taskID { // 如果taskId对上了，则获取该Task
				task, err2 = gmo.taskClient.Get(context.TODO(), t.Name, metav1.GetOptions{})
				if err2 != nil {
					logs.Errorf("Get task by taskID error from etcd:%v", err2)
				}
				break
			}
		}
		// ## 处理Task的Phase
		// 遍历Task下面的Group，根据Group的状态来设置Task的Phase
		for i := range task.Spec.Groups {
			if task.Status.GroupStatus[i].GroupID != groupStatus.GroupID { // 非当前处理的group，跳过
				continue
			}
			groupIndexInTask = i // 当前处理的group在Task当中的下标
			//if i == 0 {          // TODO 这里是根据下标=0来判断Task中的group是否执行过，这个逻辑不咋ok，所以后期需要修改
			if task.Status.Phase == apis.DeployCheck { // 说明Task刚从DeployCheck状态边为执行状态
				// 说明是Task中的第一个Group启动，那说明Task也是第一次启动，要标记状态为Phase（running or Failed）
				if phase == apis.Failed {
					task.Status.FinishAt = startTime //任务在执行的时候就发送错误，那么开始和结束时间都标记为同一时刻
				}
				task.Status.StartAt = startTime
				task.Status.Phase = phase
				task.Status.LastTime = lastTime
				taskIsFirstSet = true
			}
		}
	}

	// ## 处理Group的Phase
	//修改group下面的groupSpec下面的Actions，Actions下面的ActionStatus，ActionStatus下面的RuntimeStatus
	for j := range groupSpec.Actions { //GroupSpec(Actions)---->ActionStatus----->RuntimeStatus
		action := &groupSpec.Actions[j]
		actionStatus := &groupSpec.Actions[j].Status //ActionStatus
		if j != actionIndex {                        // 判断是否是当前处理的Action
			continue
		}
		logs.Trace("=======================================================0")
		for k := range actionStatus.RuntimeStatus { //RuntimeStatus
			rs := &actionStatus.RuntimeStatus[k]
			//if rs.Phase == apis.Running || rs.Phase == apis.Successed || rs.Phase == apis.Failed {
			//	actionStart = false //说明action下面有runtime已经启动过了（状态可能是running、Successed或者Failed）
			//}
			if k == runtimeIndex { // 是当前处理的runtime
				if phase == apis.Failed { //对于Phase等于Failed，标记startTime
					rs.FinishAt = startTime //任务在执行的时候就发送错误，那么开始和结束时间都标记为同一时刻
				} else {
					rs.ProcessId = processId
				}
				rs.StartAt = startTime
				rs.Phase = phase // 这里的Phase有可能是Failed，也有可能是running
				rs.LastTime = lastTime
				// 设置副本runtime的状态为Phase（running or failed），如果是跨域的话，这里估计还得再修改
				gmo.updateCopyIngfoForRuntime(groupSpec, phase, actionIndex, runtimeIndex)
				logs.Trace("=======================================================3")
			}
		}
		// TODO 待解决 有Init--时间的问题
		if actionStatus.Phase == apis.DeployCheck || actionStatus.Phase == apis.Init { // 说明action的刚从DeployCheck(Init)切换到启动状态，需要更改状态
			logs.Trace("=======================================================4")
			if phase == apis.Failed {
				actionStatus.FinishAt = startTime //任务在执行的时候就发送错误，那么开始和结束时间都标记为同一时刻
			}
			if phase == apis.Running {
				actionStatus.StartAt = startTime
			}
			actionStatus.Phase = phase
			actionStatus.LastTime = lastTime
			// 当前group有副本，那么需要将该任务对应的副本任务的action的开始装填也设置一下
			gmo.updateCopyIngfoForAction(groupSpec, phase, actionIndex)

			// 修改一下Action的状态
			patchAction, err := json.Marshal(map[string]interface{}{
				"status": actionStatus,
			})
			_, err = gmo.actionClient.Patch(context.TODO(), action.Name, types.StrategicMergePatchType, patchAction, metav1.PatchOptions{})
			if err != nil {
				logs.Errorf("Create action failed,err:%v", err)
			}
			// 发送Action启动的事件
			gmo.recorder.Event(action, apis.EventTypeNormal, events.StartedCommand, fmt.Sprintf("Action Name:\t %s is Running", action.Name))
		}
	}
	// TODO 待解决 有Init--时间的问题
	//修改group下面的groupStatus下面的ActionStatus，ActionStatus下面的RuntimeStatus
	if groupStatus.Phase == apis.DeployCheck || groupStatus.Phase == apis.Init { // 说明Group刚从DeployCheck(Init)状态边为执行状态
		if phase == apis.Failed {
			groupStatus.FinishAt = startTime //任务在执行的时候就发送错误，那么开始和结束时间都标记为同一时刻
		}
		if phase == apis.Running {
			groupStatus.StartAt = startTime
		}
		groupStatus.Phase = phase //说明当前处理的任务是在第一个action当中，所以Group的Phase也得设置为phase（running or Failed）
		groupStatus.LastTime = lastTime
		groupIsFirstSet = true
	}
	// ------------下面可修改为直接赋值 TODO 后期测试
	//groupStatus.ActionStatus[actionIndex] = groupSpec.Actions[actionIndex].Status
	// ------------
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
				rs.ProcessId = processId
			}
		}
		// TODO 待解决 有Init--时间的问题
		if as.Phase == apis.DeployCheck || as.Phase == apis.Init { //说明action还未设置状态为Phase(Running、Failed)
			if phase == apis.Failed {
				as.FinishAt = startTime
			}
			if phase == apis.Running {
				as.StartAt = startTime
			}
			as.Phase = phase
			as.LastTime = lastTime
		}
	}
	// Group信息更新完毕，考虑该Group是否是第一次启动Running，如果是，就发送Running事件
	if groupIsFirstSet {
		// 发送group启动的事件
		gmo.recorder.Event(getGroup, apis.EventTypeNormal, events.StartedCommand, fmt.Sprintf("Group Name:\t %s is Running", getGroup.Name))
	}
	// 修改Task信息
	if !isCopyGroup {
		// 将修改后的Group状态值赋值给Task
		task.Status.GroupStatus[groupIndexInTask] = getGroup.Status
		task.Spec.Groups[groupIndexInTask].Spec = getGroup.Spec
		task.Spec.Groups[groupIndexInTask].Status = getGroup.Status

		_, err4 := gmo.taskClient.Update(context.TODO(), task, metav1.UpdateOptions{})
		if err4 != nil {
			logs.Errorf("Update group-runtiem-start info error-1:%v", err4)
			time.Sleep(100 * time.Millisecond)
			_, err4 = gmo.taskClient.Update(context.TODO(), task, metav1.UpdateOptions{})
		}
		if taskIsFirstSet { // task是头一次Running启动
			// 发送Task启动的事件,这里有个小bug就是这时候上传的task的状态里面有些参数还没有更新 TODO 后期如果有硬性要求，这里可以放到最后修改完Task的状态后，才发送Task事件，就一个参数TaskIsfirstSet
			gmo.recorder.Event(task, apis.EventTypeNormal, events.StartedCommand, fmt.Sprintf("Task Name:\t %s is Running", task.Name))
		}
		// 修改task的信息 ---按理来说删了其实也OK
		gmo.taskManager.UpdateTask(task)
	}

	_, err3 := gmo.groupClient.Update(context.TODO(), getGroup, metav1.UpdateOptions{})
	if err3 != nil {
		logs.Error("Update group-runtiem-start info error-2:", err3)
		time.Sleep(100 * time.Millisecond)
		_, err3 = gmo.groupClient.Update(context.TODO(), getGroup, metav1.UpdateOptions{})
	}

	//将queue_manager和group_manager的group信息进行更新
	//err = gmo.groupQueues.UpdateGroup(get.Status.GroupID, get)
	//if err != nil {
	//	logs.Error("update group-runtiem-start info error")
	//}

	//测试

	p := getGroup.Status.Phase
	actionStatus := getGroup.Status.ActionStatus[actionIndex].Phase
	runtimeStatus := getGroup.Status.ActionStatus[actionIndex].RuntimeStatus[runtimeIndex].Phase
	if !isCopyGroup {
		taskPhase := task.Status.Phase
		logs.Infof("START ：taskStatus:%v,groupStatus:%v, action status: %v, runtime Status: %v", taskPhase, p, actionStatus, runtimeStatus)
	} else {
		logs.Infof("START ：groupStatus:%v, action status: %v, runtime Status: %v", p, actionStatus, runtimeStatus)
	}
	actionStatus1 := getGroup.Spec.Actions[actionIndex].Status.Phase
	runs := getGroup.Spec.Actions[actionIndex].Status.RuntimeStatus[runtimeIndex].Phase
	logs.Infof("START ：action status1: %v, runtime Status1: %v", actionStatus1, runs)
}

// 注意：新增逻辑：如果是副本任务，那么这块对于Task状态的修改，直接跳过
// EndUpdate方法：一个runtime的状态为Failed，则上层Action的状态为Failed，如果说一个runtime的状态为Successed，则上层的的Action状态还不一定是Successed
// TODO 有一个问题，比如一个group的Phase为Migrated，还得查group的副本的状态是否为succeed（目前只适配了本域迁移）
// 收到的Phase为：Successed or Failed or Unknown（Failed、Migrated）,Killed(新增)
// 处理Runtime运行时结束,如果runtime是最后一个执行完成的，还得同时标记action的phase   总结：所有临时变量赋值时都得使用&
func (gmo *GroupMonitor) handleRuntimeEndUpdate(event events.RuntimeEndPhaseEvent1) {
	logs.Info("Handling runtime end status update")
	groupName := event.GroupName
	get, err := gmo.groupClient.Get(context.TODO(), groupName, metav1.GetOptions{})
	if err != nil {
		logs.Errorf("Failed get group:%v from etcd, err:%v", groupName, err)
	}
	actionIndex := event.ActionIndex
	runtimeIndex := event.RuntimeIndex
	phase := event.Phase                                             //当前phase可能为Succeed、Failed、Unknown（Failed、Migrated）
	if phase == apis.Unknown && get.Status.Phase == apis.Migrating { // 这里有三种情况，①主动关闭，则为Failed ②异常退出，也为Failed ③主动迁移关闭，为Migrated
		// 判断为Migrated的情况 查group.Status.Phase，如果为Migrating则为迁移，否则为用户主动关闭任务的操作
		gmo.handleRuntimeMigratedUpdate(get, actionIndex, runtimeIndex)
		return
	}
	logs.Infof("Get runtime finish event notify, the phase:%s", phase)
	finshTime := event.FinishAt
	lastTime := event.LastTime
	// Task 信息
	taskID := get.Status.Belongs.TaskID

	groupSpec := &get.Spec
	groupStatus := &get.Status
	// 是否为副本任务
	isCopyGroup := get.Spec.IsCopy

	//修改group下面的groupSpec下面的Actions，Actions下面的ActionStatus，ActionStatus下面的RuntimeStatus
	var otherActionCompleted = true // group下面的其他Action是否都已经完成
	var nowActionCompleted = false  // group下面的当前Action是否已经完成

	// 修改Task下面的TaskStatus下面的状态  从Task开始遍历的好处是可以修改Task下面的状态
	var otherGroupCompleted = true // 标记其他Group的完成情况
	var nowGroupCompleted = false  // 标记当前Group的完成情况

	var finalGroupIsFailed = false // 仅针对当前group，看其下是否有Action执行失败，如果有，则Group状态必然是Failed
	var finalTaskIsFailed = false  //  仅针对当前Task，看其下是否有Group执行失败，如果有，则Task状态必然是Failed

	var finalGroupIsKilled = false
	var finalTaskIsKilled = false

	var groupIndexInTask int // 当前处理的Group在Task当中的下标
	var task *apis.Task
	if !isCopyGroup {
		list, err := gmo.taskClient.List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			logs.Errorf("Get list task err:%v", err)
		}

		var err2 error
		for _, t := range list.Items { //遍历etcd当中的所有task
			if t.Status.TaskID == taskID {
				task, err2 = gmo.taskClient.Get(context.TODO(), t.Name, metav1.GetOptions{})
				if err2 != nil {
					logs.Error("Get task by taskID error from etcd:%v", err2)
				}
			}
		}
		// 检查其他的group是否完成,修改Task的状态 ---需要适配迁移（目前只适配了本域迁移）
		for i := range task.Spec.Groups {
			grStatus := &task.Status.GroupStatus[i]
			if grStatus.GroupID != groupStatus.GroupID { //遍历到的group的ID不等于当前处理的Group的ID,即遍历兄弟group
				//logs.Infof("groupStatus.Phase:%v,groupID:%v", grStatus.Phase, grStatus.GroupID)
				if grStatus.Phase == apis.DeployCheck || grStatus.Phase == apis.Migrating { // 说明其他Group还未执行或者还没迁移成功
					otherGroupCompleted = false
				}
				if grStatus.Phase == apis.Failed { // 如果有一个Group的状态为Failed，则Task状态必定为Failed
					finalTaskIsFailed = true
				}
				if grStatus.Phase == apis.Killed {
					finalTaskIsKilled = true
				}
				if grStatus.Phase == apis.Migrated { // 还得去查对应副本任务的状态，如果状态为Running（大概率是这个状态）或者是DeployChek（说明迁移过去的group依赖不满足，暂时还不能执行），那么otherGroupCompleted参数也是false
					// 为了适配迁移，目前还是处理同域的迁移,这里怎么根据源任务找到副本任务，还是一个遗留的问题
					//copyGroup, err := gmo.groupClient.Get(context.TODO(), "Reason-Copy", metav1.GetOptions{})
					//if err != nil {
					//	logs.Errorf("Get copy group err:%v", err)
					//}
					if grStatus.CopyStatus == "" { // 副本任务只有Succeed和Failed才会修改源任务的copyStatus，如果说是空，说明副本任务还在运行
						otherGroupCompleted = false
					}
					if grStatus.CopyStatus == "Failed" {
						finalTaskIsFailed = true
					}
				}
				continue
			}
			groupIndexInTask = i // 当前处理的Group在Task当中的下标
		}
	}

	// 修改groupSpec的状态
	for i := range groupSpec.Actions { //Action
		var finalActionIsFailed = false              //标记group里面当前遍历到的Action地下的runtime是否有Failed状态
		var finalActionIsKilled = false              //标价group里面当前遍历到的Action底下的runtime是否有Killed状态
		var allRuntiemCompleted = true               // 当前action是否已经完成（只有action下面的所有的runtime都执行完成了，也就是最后一个runtime被执行完成了，要标记action的状态为succeed，如果说action下面的某一个runtime执行失败，则要标记action装填为Failed）
		actionStatus := &groupSpec.Actions[i].Status //ActionStatus
		if i != actionIndex {                        //遍历到其他Action，可以顺带看一下别的Action是否都已经完成了
			if actionStatus.Phase == apis.DeployCheck { //其他Action为checking状态，说明还有其他的Action没有被遍历到，Group状态为Running状态
				otherActionCompleted = false
			}
			if actionStatus.Phase == apis.Failed {
				finalGroupIsFailed = true
			}
			if actionStatus.Phase == apis.Killed {
				finalGroupIsKilled = true
			}
			continue
		}
		for j := range actionStatus.RuntimeStatus { //RuntimeStatus
			rs := &actionStatus.RuntimeStatus[j]
			if j == runtimeIndex {
				rs.Phase = phase
				rs.FinishAt = finshTime
				rs.LastTime = lastTime
				// 如果说该group有副本，并且该副本group是提前部署副本的，那么这里除了修改源runtime的状态，还得修改副本runtime的状态
				gmo.updateCopyIngfoForRuntime(groupSpec, phase, actionIndex, runtimeIndex)
			}
			if rs.Phase == apis.DeployCheck { // 遍历所有的Runtime，如果其中一个Runtime状态没有执行完成，说明Action最终不用更新
				allRuntiemCompleted = false
			}
			if rs.Phase == apis.Failed {
				finalActionIsFailed = true
			}
			if rs.Phase == apis.Killed {
				finalActionIsKilled = true
			}
		}
		if allRuntiemCompleted { //如果说ActionStatus下面的RuntimeStatus都被执行了，还得修改ActionStatus的phase状态
			//后续可能还要补充:Results
			//actionStatus.Results = results
			actionStatus.FinishAt = finshTime
			actionStatus.LastTime = lastTime
			// 还得发送Action执行完成的事件，同时将etcd当中的Action资源状态进行修改
			action := &groupSpec.Actions[i]
			if finalActionIsFailed {
				actionStatus.Phase = apis.Failed
				gmo.recorder.Event(action, apis.EventTypeWarning, events.ExecuteFailed, fmt.Sprintf("Action Name:\t %s is Failed", action.Name))
			} else if finalActionIsKilled {
				actionStatus.Phase = apis.Killed
				gmo.recorder.Event(action, apis.EventTypeNormal, events.KillingCommand, fmt.Sprintf("Action Name:\t %s is Killed", action.Name))
			} else {
				actionStatus.Phase = phase
				gmo.recorder.Event(action, apis.EventTypeNormal, events.ExecuteSuccessfully, fmt.Sprintf("Action Name:\t %s is Successed", action.Name))
			}
			patchAction, err := json.Marshal(map[string]interface{}{
				"status": actionStatus,
			})
			if err != nil {
				logs.Errorf("Marshal patch group err:%v", err)
			}
			_, err = gmo.actionClient.Patch(context.TODO(), action.Name, types.StrategicMergePatchType, patchAction, metav1.PatchOptions{})
			if err != nil {
				logs.Errorf("Patch action err:%v", err)
			}
			nowActionCompleted = true //当前Action已经完成
			gmo.updateCopyIngfoForAction(groupSpec, phase, actionIndex)
		}
	}
	//如果说GroupStatus下面的ActionStatus都被执行了，还得修改GroupStatus的phase的状态
	if otherActionCompleted && nowActionCompleted { //说明其他Action都执行完成，当前Action也执行完成
		groupStatus.FinishAt = finshTime
		groupStatus.LastTime = lastTime
		if finalGroupIsFailed {
			groupStatus.Phase = apis.Failed
			gmo.recorder.Event(get, apis.EventTypeWarning, events.ExecuteFailed, fmt.Sprintf("Group name:\t %s is failed", get.Name))
		} else if finalGroupIsKilled {
			groupStatus.Phase = apis.Killed
			gmo.recorder.Event(get, apis.EventTypeNormal, events.KillingCommand, fmt.Sprintf("Group name:\t %s is killed", get.Name))
		} else {
			groupStatus.Phase = phase //Group的状态等于当前Action执行完成的状态：Failed  or  Succeed
			gmo.recorder.Event(get, apis.EventTypeNormal, events.ExecuteSuccessfully, fmt.Sprintf("Group name:\t %s is successed", get.Name))
		}

		nowGroupCompleted = true

		//遍历task，同时标记TaskStatus下GroupStatus状态也为phase---这个在下面进行统一处理
	}
	//修改group下面的groupStatus下面的ActionStatus，ActionStatus下面的RuntimeStatus
	// ----------------可以改为赋值 TODO 后期测试  一行替换下面的for循环遍历
	//groupStatus.ActionStatus[actionIndex] = groupSpec.Actions[actionIndex].Status
	// ----------------
	for i := range groupStatus.ActionStatus { //ActionStatus
		var allRuntiemCompleted = true  // 当前action是否已经完成（只有action下面的所有的runtime都执行完成了，也就是最后一个runtime被执行完成了，要标记action的状态为succeed，如果说action下面的某一个runtime执行失败，则要标记action装填为Failed）
		var finalActionIsFailed = false //标记group里面当前遍历到的Action底下的runtime是否有Failed状态，只要有Failed，Action的状态也必为Failed
		var finalActionIsKilled = false
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
			if rs.Phase == apis.DeployCheck {
				allRuntiemCompleted = false
			}
			if rs.Phase == apis.Failed {
				finalActionIsFailed = true
			}
			if rs.Phase == apis.Killed {
				finalActionIsKilled = true
			}
		}
		if allRuntiemCompleted { //如果说ActionStatus下面的RuntimeStatus都是完成的状态，还得修改ActionStatus的phase状态
			if finalActionIsFailed {
				as.Phase = apis.Failed
			} else if finalActionIsKilled {
				as.Phase = apis.Killed
			} else {
				as.Phase = phase
			}
			as.FinishAt = finshTime
			as.LastTime = lastTime
		}
	}
	// 更新Group资源信息，同时还需要判断该group是否执行完成，需要发送事件
	_, err4 := gmo.groupClient.Update(context.TODO(), get, metav1.UpdateOptions{})
	if err4 != nil {
		logs.Errorf("Update group-runtiem-start info error-4:%v", err4) //报错
		time.Sleep(100 * time.Millisecond)
		_, err4 = gmo.groupClient.Update(context.TODO(), get, metav1.UpdateOptions{})
	}

	//如果说TaskStatus下面的Group都被执行了，还得修改TaskStatus的phase的状态
	//logs.Infof("otherGroupCompleted:%v", otherGroupCompleted)
	//logs.Infof("nowGroupCompleted:%v", nowGroupCompleted)
	if !isCopyGroup {
		// 将修改后的Group状态值赋值给Task
		task.Status.GroupStatus[groupIndexInTask] = get.Status
		task.Spec.Groups[groupIndexInTask].Spec = get.Spec
		task.Spec.Groups[groupIndexInTask].Status = get.Status

		if otherGroupCompleted && nowGroupCompleted {
			task.Status.FinishAt = finshTime
			task.Status.LastTime = lastTime
			if finalTaskIsFailed { // 如果说group当中有Failed状态，那么最终Task也是得被标记为Failed
				task.Status.Phase = apis.Failed
				gmo.recorder.Event(task, apis.EventTypeWarning, events.ExecuteFailed, fmt.Sprintf("Task Name:\t %s is Failed", task.Name))
			} else if finalTaskIsKilled {
				task.Status.Phase = apis.Killed
				gmo.recorder.Event(task, apis.EventTypeNormal, events.KillingCommand, fmt.Sprintf("Task Name:\t %s is Killed", task.Name))
			} else {
				task.Status.Phase = phase //表示的是Task下的其他Group都是Successed状态，那么Task的状态取决于当前的Group，如果为Succeed，则Task也为Succeed，反正为Failed
				gmo.recorder.Event(task, apis.EventTypeNormal, events.ExecuteSuccessfully, fmt.Sprintf("Task Name:\t %s is Successed", task.Name))
			}
		}
		_, err3 := gmo.taskClient.Update(context.TODO(), task, metav1.UpdateOptions{})
		if err3 != nil {
			logs.Errorf("Update group-runtiem-start info error-3:%v", err3)
			time.Sleep(100 * time.Millisecond)
			_, err3 = gmo.taskClient.Update(context.TODO(), task, metav1.UpdateOptions{})
		}
		// 修改task的信息 ---这个不要也行
		gmo.taskManager.UpdateTask(task)
	}

	//将queue_manager和group_manager的group信息进行更新
	//err = gmo.groupQueues.UpdateGroup(get.Status.GroupID, get)
	//if err != nil {
	//	logs.Error("update group-runtiem-end info error")
	//}

	//测试：

	p := get.Status.Phase
	actionStatus := get.Status.ActionStatus[actionIndex].Phase
	runStatus1 := get.Status.ActionStatus[actionIndex].RuntimeStatus[runtimeIndex].Phase
	if !isCopyGroup {
		taskPhase := task.Status.Phase
		logs.Infof("Runtime END: taskStatus:%v, groupStatus:%v, actionStatus:%v, runtimeStatus:%v", taskPhase, p, actionStatus, runStatus1)
	} else {
		logs.Infof("Runtime END: groupStatus:%v, actionStatus:%v, runtimeStatus:%v", p, actionStatus, runStatus1)
	}
	actionStatus1 := get.Spec.Actions[actionIndex].Status.Phase
	runtimeStatus1 := get.Spec.Actions[actionIndex].Status.RuntimeStatus[runtimeIndex].Phase
	logs.Infof("Runtime END: actionStatus:%v, runtimeStatus:%v", actionStatus1, runtimeStatus1)
}

// 同group_workers当中的方法
func (gmo *GroupMonitor) groupDepenSatisfy(group *apis.Group) bool {
	// TODO：实现依赖检查逻辑
	// 目前只是检查Spec当中的Parents选项
	if len(group.Spec.Parents) == 0 {
		return true
	} else {
		for _, i := range group.Spec.Conditions.Formulas {
			if i.LeftValue.Name == "NodeDependency" {
				parentName := i.LeftValue.From // 这里要考虑父亲节点有两个的情况吧，还是说弄两个Formulas
				result, err := gmo.groupClient.Get(context.TODO(), parentName, metav1.GetOptions{})
				if err != nil {
					logs.Errorf("Failed to get parent group:%v form etcd, err:%v", parentName, err)
				}
				if result.Status.Phase != apis.Successed {
					if result.Status.Phase == apis.Migrated { // 因为有些group处于DeployCheck状态，不会被迁移走，所以如果该group的父亲Group被迁移走了，需要查父亲Group的copy_status属性，该属性存放副本是否完成
						if result.Status.CopyStatus == "Successed" {
							i.LeftValue.Value = "1"
						}
					} else {
						i.LeftValue.Value = "0"
					}
				} else {
					i.LeftValue.Value = "1"
				}
				if i.LeftValue.Value == i.RightValue.Value {
					i.Result = true
				}
			}
			if !i.Result {
				//logs.Infof("group condition[%v]:%v do not satisfy, groupName:%v", index, i.LeftValue.Name, group.Spec.Name)
				return false
			} else {
				//logs.Infof("group condition[%v]:%v satisfy!", index, i.LeftValue.Name)
			}
		}
	}
	return true
}

// 检查Action的依赖是否满足
func (gmo *GroupMonitor) actionDepenSatisfy(actionIndex int, group *apis.Group) bool {
	actionSpec := &group.Spec.Actions[actionIndex].Spec
	for index, i := range actionSpec.Conditions.Formulas {
		if i.LeftValue.Name == "NodeDependency" {
			actionParentName := i.LeftValue.From
			for j := range group.Status.ActionStatus {
				as := &group.Status.ActionStatus[j]
				a := &group.Spec.Actions[j]
				if a.Name == actionParentName { //目前定义，Action的父亲Action必须是成功状态.更新action的conditions
					if as.Phase != apis.Successed {
						i.LeftValue.Value = "0"
					} else {
						i.LeftValue.Value = "1"
					}
				}
			}
			if i.LeftValue.Value == i.RightValue.Value {
				i.Result = true
			}
		}
		if !i.Result {
			logs.Infof("action condition[%v]:%v do not satisfy, actionName:%v", index, i.LeftValue.Name, actionSpec.Name)
			return false
		} else {
			// logs.Infof("group condition[%v]:%v satisfy!", index, i.LeftValue.Name)
		}
	}
	return true
}

// 检查Runtime的依赖是否满足
func (gmo *GroupMonitor) runtimeDepenSatisfy(actionIndex, runtimeIndex int, group *apis.Group) bool {
	//TODO runtime运行之前，需要检查parent的runtime是否正常执行完成
	runtime := &group.Spec.Actions[actionIndex].Spec.Runtimes[runtimeIndex]
	rtStatus := &group.Spec.Actions[actionIndex].Status.RuntimeStatus[runtimeIndex]
	for _, i := range runtime.Conditions.Formulas {
		if i.LeftValue.Name == "NodeDependency" {
			//正则匹配选择parents的pahse
			runtimeParentName := i.LeftValue.From
			for j := range group.Status.ActionStatus[actionIndex].RuntimeStatus {
				rs := &group.Status.ActionStatus[actionIndex].RuntimeStatus[j]
				r := &group.Spec.Actions[actionIndex].Spec.Runtimes[j]
				if r.Name == runtimeParentName { //目前定义，Action的父亲Action必须是成功状态.更新action的conditions
					if rs.Phase != apis.Successed {
						i.LeftValue.Value = "0"
					} else {
						i.LeftValue.Value = "1"
					}
				}
			}
			if i.LeftValue.Value == i.RightValue.Value {
				i.Result = true
			}
			if !i.Result {
				//logs.Infof("runtime condition[%v]:%v do not satisfy, runtimeName:%v", index, i.LeftValue.Name, runtime.Name)
				return false
			} else {
				// logs.Infof("group condition[%v]:%v satisfy!", index, i.LeftValue.Name)
			}
		} else if i.LeftValue.Name == "ProgramDependency" {
			//runtime运行之前,需要检查程序依赖是不是满足，如果满足则将符合条件的环境变量加入runtime的Env中，方便后续CMD注入环境变量；
			//如果不满足则返回false，开启CMD创建新的程序依赖，等待monitor检查到依赖满足才拉起这个runtime
			//TODO：后续和上面的condition合并进一起，可能是以单独写一个condition函数的形式，然后这里只需要调用统一的condition检查函数即可
			var dependencyFile = i.LeftValue.From

			if !rtStatus.IsParsed {
				// runtimeReqPackages := make([]apis.Requirement, 0)
				runtimeReqPackages, err := gmo.dependencyManager.ParseRequirements(dependencyFile)
				rtStatus.IsParsed = true
				if err != nil {
					logs.Info("parse Runtime Requirements error!")
					rtStatus.IsParsed = false
					return false
				}
				runtime.Packages = runtimeReqPackages
			}
			envName, err := gmo.dependencyManager.CheckEnvironmentSatisfy(runtime.Packages)
			if !err {
				// logs.Info("dependency do not satisfy,runtime name:%v", runtime.Name)
				//需要使用协程，但是还要防止在monitor监控的时候多次创建
				if !rtStatus.DepenPreparing {
					logs.Info("installing dependency for runtime name:%v", runtime.Name)
					rtStatus.DepenPreparing = true
					// go dependency.SetupEnvironment(dependencyFile, runtime.Name)
				} else {
					logs.Info("runtime %v is waiting for installing dependency!", runtime.Name)
				}
				return false
			} else {
				//修改conditionIndex对应的condition结果
				i.LeftValue.Value = "1"
				if i.Signal == apis.Equal {
					if i.LeftValue.Value == i.RightValue.Value {
						i.Result = true
					}
					//暂时没想到 ！= 如何使用，暂定判断条件相等 ==
					// }else {
					// 	if runtime.Conditions.Formulas[conditionIndex].LeftValue.Value != runtime.Conditions.Formulas[conditionIndex].RightValue.Value {
					// 		runtime.Conditions.Formulas[conditionIndex].Result = true
					// 	}
				}
				if !i.Result {
					//logs.Infof("runtime condition[%v]:%v do not satisfy, runtimeName:%v", index, i.LeftValue.Name, runtime.Name)
					return false
				} else {
					// logs.Infof("group condition[%v]:%v satisfy!", index, i.LeftValue.Name)
				}
			}
			path, err := dependency.EnvForInput(envName)
			if !err {
				logs.Infof("%v: envName err:%v", runtime.Name, envName)
				return false
			}
			envVar := []apis.EnvVar{}
			envVar = append(envVar, apis.EnvVar{Name: "PATH", Value: path})

			//TODO 后续将envPath改为数组，返回多种程序依赖
			// envVar = append(envVar, apis.EnvVar{Name: "python", Value: envPath})
			runtime.EnvVar = append(runtime.EnvVar, envVar...)
		}
	}
	return true
}

// 主要处理group一直未完成的情况，这里主要讲group的状态设置为Failed，Task状态设置为Failed，group下面的Action、Runtime状态均为Unknown
func (gmo *GroupMonitor) handleStatusUpdate(group *apis.Group, failed apis.Phase) {
	// 设置当前group的状态为Failed，同时去找Task，标记其状态也为Failed
	time := apis.Time{time.Now()}
	group.Status.Phase = failed
	group.Status.StartAt = time
	group.Status.FinishAt = time
	group.Status.LastTime = time

	taskID := group.Status.Belongs.TaskID // 当前所属的Task的ID

	taskList, err := gmo.taskClient.List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		logs.Errorf("Get list task from etcd err:%v", err)
	}
	for _, ta := range taskList.Items {
		if ta.Status.TaskID == taskID {
			task, err := gmo.taskClient.Get(context.TODO(), ta.Name, metav1.GetOptions{})
			if err != nil {
				logs.Errorf("Get task from etcd err:%v", err)
			}
			task.Status.Phase = failed
			task.Status.StartAt = time
			task.Status.FinishAt = time
			task.Status.LastTime = time

			for j := range task.Spec.Groups {
				gr := &task.Spec.Groups[j]
				if gr.Status.GroupID == group.Status.GroupID {
					task.Status.GroupStatus[j] = group.Status
				}
			}
			_, err = gmo.taskClient.Update(context.TODO(), task, metav1.UpdateOptions{})
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

// 处理runtime是需要迁移的情况，就是将runtime的Running状态改为Migrated，如果为Succeed，这个Phase不修改  最后都修改完后将group修改为Migrated
func (gmo *GroupMonitor) handleRuntimeMigratedUpdate(group *apis.Group, actionIndex, runtimeIndex int) {
	// 将runtime的Phase从Running修改为Migrated
	nowTime := apis.Time{time.Now()}
	groupSpec := &group.Spec
	groupStatus := &group.Status
	// GroupSpec下面的Actions，需要修改想下面的（ActionStatus的Phase以及Runtime的Phase）
	groupSpec.Actions[actionIndex].Status.Phase = apis.Migrated
	groupSpec.Actions[actionIndex].Status.LastTime = nowTime
	groupSpec.Actions[actionIndex].Status.FinishAt = nowTime
	groupSpec.Actions[actionIndex].Status.RuntimeStatus[runtimeIndex].Phase = apis.Migrated
	groupSpec.Actions[actionIndex].Status.RuntimeStatus[runtimeIndex].LastTime = nowTime
	groupSpec.Actions[actionIndex].Status.RuntimeStatus[runtimeIndex].FinishAt = nowTime

	// GroupStatus下面的ActionStatus、RuntimeStatus
	groupStatus.ActionStatus[actionIndex].Phase = apis.Migrated
	groupStatus.ActionStatus[actionIndex].LastTime = nowTime
	groupStatus.ActionStatus[actionIndex].FinishAt = nowTime
	groupStatus.ActionStatus[actionIndex].RuntimeStatus[runtimeIndex].Phase = apis.Migrated
	groupStatus.ActionStatus[actionIndex].RuntimeStatus[runtimeIndex].LastTime = nowTime
	groupStatus.ActionStatus[actionIndex].RuntimeStatus[runtimeIndex].FinishAt = nowTime

	// 检查group下的信息是否都修改完成
	var IsModify = true
	for i := range groupStatus.ActionStatus {
		if i == actionIndex { // 得将当前遍历到的Action排除
			continue
		}
		actionStatus := groupStatus.ActionStatus[i]
		if actionStatus.Phase == apis.Running {
			IsModify = false
		}
	}
	logs.Infof("------------------------------------------IsModify:%v", IsModify)
	if IsModify { // group下面的信息都修改完成，接下来需要将一些DeployCheck的状态进行修改。group修改完后，还需要将group的信息放到Task当中，如果Task下面的所有group都执行成功且都放到了Task下面的话，最后还需要做收尾工作，标记Task状态为Running，当迁移完成后，由Migrated队列轮询检查副本的执行情况，再标记Task的状态为成功或失败
		logs.Info("---------------进入ISModify----------------------------------------")
		// 统一将DeployCheck的Phase都改成Migrate 首先是GroupSpec，如果是成功的Phase就不修改了，其他的都修改，包括running、DeployCheck
		for i := range groupSpec.Actions {
			actionStatus := &groupSpec.Actions[i].Status
			if actionStatus.Phase == apis.DeployCheck || actionStatus.Phase == apis.Migrated {
				for j := range actionStatus.RuntimeStatus {
					runtimeStatus := &actionStatus.RuntimeStatus[j]
					if runtimeStatus.Phase == apis.Migrated {
						continue
					}
					runtimeStatus.Phase = apis.Migrated
					runtimeStatus.LastTime = nowTime
					runtimeStatus.FinishAt = nowTime
				}
				if actionStatus.Phase == apis.Migrated {
					continue
				}
				actionStatus.Phase = apis.Migrated
				actionStatus.LastTime = nowTime
				actionStatus.FinishAt = nowTime
			}
		}
		// 接着是GroupStatus
		for i := range groupStatus.ActionStatus {
			actionStatus := &groupStatus.ActionStatus[i]
			if actionStatus.Phase == apis.DeployCheck || actionStatus.Phase == apis.Migrated {
				for j := range actionStatus.RuntimeStatus {
					runtimeStatus := &actionStatus.RuntimeStatus[j]
					if runtimeStatus.Phase == apis.Migrated {
						continue
					}
					runtimeStatus.Phase = apis.Migrated
					runtimeStatus.LastTime = nowTime
					runtimeStatus.FinishAt = nowTime
				}
				if actionStatus.Phase == apis.Migrated {
					continue
				}
				actionStatus.Phase = apis.Migrated
				actionStatus.LastTime = nowTime
				actionStatus.FinishAt = nowTime
			}
		}
		group.Status.Phase = apis.Migrated

		// 这里检查一下其他Group的状态是否为都成功了，如果都成功了，还需要修改Task的信息
		var task *apis.Task
		var otherGroupCompleted = true
		var groupIndexInTask int
		var finalTaskIsFailed = false
		taskID := group.Status.Belongs.TaskID
		list, err := gmo.taskClient.List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			logs.Errorf("Get list task err:%v", err)
		}
		var err2 error
		for _, t := range list.Items { //遍历etcd当中的所有task
			if t.Status.TaskID == taskID { // 如果taskId对上了，则就修改该Task的Phase为checking
				task, err2 = gmo.taskClient.Get(context.TODO(), t.Name, metav1.GetOptions{})
				if err2 != nil {
					logs.Error("Get task by taskID error from etcd:%v", err2)
				}
			}
		}
		// 检查其他的group是否完成,修改Task的状态 ---需要适配迁移（目前只适配了本域迁移）
		for i := range task.Spec.Groups {
			grStatus := &task.Status.GroupStatus[i]
			if grStatus.GroupID != groupStatus.GroupID { //遍历到的group的ID不等于当前处理的Group的ID
				//logs.Infof("groupStatus.Phase:%v,groupID:%v", grStatus.Phase, grStatus.GroupID)
				if grStatus.Phase == apis.DeployCheck || grStatus.Phase == apis.Migrating { // 说明其他Group还未执行或者还没迁移成功
					otherGroupCompleted = false
				}
				if grStatus.Phase == apis.Failed { // 如果有一个Group的状态为Failed，则Task状态必定为Failed
					finalTaskIsFailed = true
				}
				if grStatus.Phase == apis.Migrated { // 还得去查对应副本任务的状态，如果状态为Running（大概率是这个状态）或者是DeployChek（说明迁移过去的group依赖不满足，暂时还不能执行），那么otherGroupCompleted参数也是false
					// 为了适配迁移，目前还是处理同域的迁移,这里怎么根据源任务找到副本任务，还是一个遗留的问题
					//copyGroup, err := gmo.groupClient.Get(context.TODO(), "Reason-Copy", metav1.GetOptions{})
					//if err != nil {
					//	logs.Errorf("Get copy group err:%v", err)
					//}
					if grStatus.CopyStatus == "" { // 说明该Group的副本group正在运行还没结束
						otherGroupCompleted = false
					}
					if grStatus.CopyStatus == "Failed" {
						finalTaskIsFailed = false
					}
					//if copyGroup.Status.Phase == apis.DeployCheck || copyGroup.Status.Phase == apis.Running {
					//	otherGroupCompleted = false
					//}
					//if copyGroup.Status.Phase == apis.Failed {
					//	finalTaskIsFailed = true
					//}
				}
				continue
			}
			groupIndexInTask = i // 当前处理的Group在Task当中的下标
		}
		if otherGroupCompleted {
			if finalTaskIsFailed { // 如果说group当中有Failed状态，那么最终Task也是得被标记为Failed
				task.Status.Phase = apis.Failed
			} else {
				task.Status.Phase = apis.Running //表示的是Task下的其他Group都是Successed状态，那么Task的状态取决于当前的Group，因为当前的Group正在迁移，只有当前group的副本任务完成了，那么到时候在Migrated队列当中，就会修改Task的状态的，这里设置为Running，合理
			}
			task.Status.FinishAt = nowTime
			task.Status.LastTime = nowTime
		}
		// 将修改后的Group状态值赋值给Task
		task.Status.GroupStatus[groupIndexInTask] = group.Status
		task.Spec.Groups[groupIndexInTask].Spec = group.Spec
		task.Spec.Groups[groupIndexInTask].Status = group.Status

		_, err3 := gmo.taskClient.Update(context.TODO(), task, metav1.UpdateOptions{})
		if err3 != nil {
			logs.Errorf("Update group-runtiem-start info error-88:%v", err3)
			time.Sleep(100 * time.Millisecond)
			_, err3 = gmo.taskClient.Update(context.TODO(), task, metav1.UpdateOptions{})
		}
	}
	// 最终将group信息写入到etcd当中
	_, err4 := gmo.groupClient.Update(context.TODO(), group, metav1.UpdateOptions{})
	if err4 != nil {
		logs.Errorf("Update group-runtiem-start info error-5:%v", err4) //报错
		time.Sleep(100 * time.Millisecond)
		_, err4 = gmo.groupClient.Update(context.TODO(), group, metav1.UpdateOptions{})
	}

}

// 主动将当前指定的runtime状态设置为Successed,如果其兄弟Runtime也都Successed了，那么也需要将Action的状态设置为Succeed---这里调用的情况：对于那些非细粒度控制的runtime，当其源runtime完成后，这里需要修改runtime
func (gmo *GroupMonitor) handleRuntimeSucceedUpdate(group *apis.Group, actionIndex, runtimeIndex int) {
	// 将runtime的Phase从Running修改为Migrated
	groupSpec := &group.Spec
	groupStatus := &group.Status
	// 这里不设置时间，因为这个副本的runtime（粗粒度控制）压根没启动过，只是因为其源runtime执行完成了，所以该runtime也就无需执行了，直接设置Phase即可
	// GroupSpec下面的Actions，需要修改想下面的（ActionStatus的Phase以及Runtime的Phase）
	groupSpec.Actions[actionIndex].Status.Phase = apis.Successed
	groupSpec.Actions[actionIndex].Status.RuntimeStatus[runtimeIndex].Phase = apis.Successed

	// GroupStatus下面的ActionStatus、RuntimeStatus
	groupStatus.ActionStatus[actionIndex].Phase = apis.Successed
	groupStatus.ActionStatus[actionIndex].RuntimeStatus[runtimeIndex].Phase = apis.Successed

	// 是否还需要这一步：检查兄弟Runtime是否都执行完，如果执行完就，还得修改Action的状态为Successed-----最好写上吧，因为对于Action下面有细粒度的runtime，其调用StopRuntime方法的时候，会修改Runtime和上层Action的状态，但是对于粗粒度的Runtime，是没有调用运行时的方法的，所以没有修改上层Action的状态
	// TODO 经过分析，这块不需要，对于action的phase改为Succeed，会在actionStatus.CopyStatus == "Succeeded"这里进行修改
	//var actionIsSuccess = true
	//actionStatus := groupStatus.ActionStatus[actionIndex]
	//for j := range actionStatus.RuntimeStatus {
	//	runtimeStatus := actionStatus.RuntimeStatus[j]
	//	if runtimeStatus.Phase != apis.Successed {
	//		actionIsSuccess = false
	//		break
	//	}
	//}
	//if actionIsSuccess {
	//	groupSpec.Actions[actionIndex].Status.Phase = apis.Successed
	//	groupStatus.ActionStatus[actionIndex].Phase = apis.Successed
	//}
	// 最终将group信息写入到etcd当中
	_, err4 := gmo.groupClient.Update(context.TODO(), group, metav1.UpdateOptions{})
	if err4 != nil {
		logs.Errorf("Update group-runtime-successed info error-9:%v", err4) //报错
		time.Sleep(100 * time.Millisecond)
		_, err4 = gmo.groupClient.Update(context.TODO(), group, metav1.UpdateOptions{})
	}

}

// 将副本的Action以及下面的runtime（还没设置为Succeed）的phase设置为Succeed,时间的话全部改成源任务完成的时间吧
func (gmo *GroupMonitor) handleCopyActionSucceedUpdate(groupCopy *apis.Group, actionIndex int) {
	groupSpec := &groupCopy.Spec
	groupStatus := &groupCopy.Status
	nowTime := apis.Time{time.Now()}
	// groupSpec下面的Action以及runtime的phase修改为Succeed，并且将所有时间修改成当前的时间
	groupSpec.Actions[actionIndex].Status.Phase = apis.Successed
	groupSpec.Actions[actionIndex].Status.LastTime = nowTime
	groupSpec.Actions[actionIndex].Status.FinishAt = nowTime
	groupSpec.Actions[actionIndex].Status.StartAt = nowTime
	for i := range groupSpec.Actions[actionIndex].Status.RuntimeStatus {
		runtimeStatus := &groupSpec.Actions[i].Status
		if runtimeStatus.Phase != apis.Successed {
			runtimeStatus.Phase = apis.Successed
		}
		runtimeStatus.StartAt = nowTime
		runtimeStatus.FinishAt = nowTime
		runtimeStatus.LastTime = nowTime
	}
	// groupStatus下面的Action以及runtime的Phase修改为Succeed
	groupStatus.ActionStatus[actionIndex].Phase = apis.Successed
	groupStatus.ActionStatus[actionIndex].LastTime = nowTime
	groupStatus.ActionStatus[actionIndex].StartAt = nowTime
	groupStatus.ActionStatus[actionIndex].FinishAt = nowTime
	for i := range groupStatus.ActionStatus[actionIndex].RuntimeStatus {
		runtimeStatus := &groupStatus.ActionStatus[actionIndex].RuntimeStatus[i]
		if runtimeStatus.Phase != apis.Successed {
			runtimeStatus.Phase = apis.Successed
		}
		runtimeStatus.LastTime = nowTime
		runtimeStatus.FinishAt = nowTime
		runtimeStatus.StartAt = nowTime
	}
	_, err2 := gmo.groupClient.Update(context.TODO(), groupCopy, metav1.UpdateOptions{})
	if err2 != nil {
		logs.Errorf("Update group-runtime-successed info error-9:%v", err2)
	}
}

// 设置group、action、runtime的phase为failed，单单将当前这个Filed的Action，以及下面Failed的Runtime的状态设置为Failed，对于其他的Action、Runtime不做处理，然后同时将Group的状态设置为Failed，完毕
func (gmo *GroupMonitor) handleCopyRuntimeFailedUpdate(groupCopy *apis.Group, actionIndex int) {
	nowTime := apis.Time{time.Now()}
	groupCopy.Status.Phase = apis.Failed
	groupCopy.Status.LastTime = nowTime
	groupCopy.Status.FinishAt = nowTime
	groupCopy.Status.StartAt = nowTime

	groupSpec := &groupCopy.Spec
	groupStatus := &groupCopy.Status

	// groupSpec下面的Action以及runtime的phase修改为Failed
	groupSpec.Actions[actionIndex].Status.Phase = apis.Failed
	groupSpec.Actions[actionIndex].Status.LastTime = nowTime
	groupSpec.Actions[actionIndex].Status.FinishAt = nowTime
	groupSpec.Actions[actionIndex].Status.StartAt = nowTime
	for i := range groupSpec.Actions[actionIndex].Status.RuntimeStatus {
		runtimeStatus := &groupSpec.Actions[i].Status
		if runtimeStatus.CopyStatus == "Failed" { // 修改造成Failed的Runtime的状态为Failed
			runtimeStatus.Phase = apis.Failed
			runtimeStatus.LastTime = nowTime
			runtimeStatus.FinishAt = nowTime
			runtimeStatus.StartAt = nowTime
		}
	}
	// groupStatus下面的Action以及runtime的Phase修改为Succeed
	groupStatus.ActionStatus[actionIndex].Phase = apis.Failed
	groupStatus.ActionStatus[actionIndex].LastTime = nowTime
	groupStatus.ActionStatus[actionIndex].StartAt = nowTime
	groupStatus.ActionStatus[actionIndex].FinishAt = nowTime
	for i := range groupStatus.ActionStatus[actionIndex].RuntimeStatus {
		runtimeStatus := &groupStatus.ActionStatus[actionIndex].RuntimeStatus[i]
		if runtimeStatus.CopyStatus == "Failed" {
			runtimeStatus.Phase = apis.Failed
			runtimeStatus.LastTime = nowTime
			runtimeStatus.FinishAt = nowTime
			runtimeStatus.StartAt = nowTime
		}
	}
	_, err2 := gmo.groupClient.Update(context.TODO(), groupCopy, metav1.UpdateOptions{})
	if err2 != nil {
		logs.Errorf("Update group-runtime-successed info error-9:%v", err2)
	}
}

// 设置当前group的状态(copyStatus的值)为Failed，标记其副本任务执行失败了，同时修改当前Group所属的Task的phase为Failed
func (gmo *GroupMonitor) handleTaskFailedUpdate(gro *apis.Group) {
	// 设置group的状态(copyStatus的值)为Failed
	patchGroup, err := json.Marshal(map[string]interface{}{
		"status": map[string]interface{}{
			"copy_status": "Failed",
		},
	})
	_, err = gmo.groupClient.Patch(context.TODO(), gro.Name, types.StrategicMergePatchType, patchGroup, metav1.PatchOptions{})
	if err != nil {
		logs.Errorf("Patch task err-14:%v", err)
	}
	// 修改Group所属的Task的Phase为Failed
	nowTime := apis.Time{time.Now()}
	taskID := gro.Status.Belongs.TaskID // group的Belongs属性当中的TaskID
	var task *apis.Task
	list, err := gmo.taskClient.List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		logs.Errorf("Get list task err:%v", err)
	}

	var err2 error
	for _, t := range list.Items { //遍历etcd当中的所有task
		if t.Status.TaskID == taskID { // 如果taskId对上了，则就修改该Task的Phase为checking
			task, err2 = gmo.taskClient.Get(context.TODO(), t.Name, metav1.GetOptions{})
			if err2 != nil {
				logs.Error("Get task by taskID error from etcd:%v", err2)
			}
		}
	}
	if task.Status.Phase == apis.Failed { // 有可能有这种情况，就是有多个迁移的group，可能其中一个已经执行了该方法，将Task的phase改为了Failed了
		return
	}
	patchTask, err := json.Marshal(map[string]interface{}{
		"status": map[string]interface{}{
			"phase":     apis.Failed,
			"finish":    nowTime,
			"last_time": nowTime,
		},
	})
	_, err = gmo.taskClient.Patch(context.TODO(), task.Name, types.StrategicMergePatchType, patchTask, metav1.PatchOptions{})
	if err != nil {
		logs.Errorf("Patch task err:%v", err)
	}
}

// 设置当前group的状态(copyStatus的值)为Succeed，标记其副本任务执行成功了，同时检查group所属Task下面的所有group（除了当前的group）是否都已经成功了，如果都已经成功了，就修改当前Group所属的Task的phase为Succeed
func (gmo *GroupMonitor) handleTaskSucceedUpdate(gro *apis.Group) {
	// 设置group的状态(copyStatus的值)为Succeed
	patchGroup, err := json.Marshal(map[string]interface{}{
		"status": map[string]interface{}{
			"copy_status": "Successed",
		},
	})
	_, err = gmo.groupClient.Patch(context.TODO(), gro.Name, types.StrategicMergePatchType, patchGroup, metav1.PatchOptions{})
	if err != nil {
		logs.Errorf("Patch task err-14:%v", err)
	}
	// 修改Group所属的Task的Phase为Succed
	nowTime := apis.Time{time.Now()}
	taskID := gro.Status.Belongs.TaskID // group的Belongs属性当中的TaskID
	var task *apis.Task
	list, err := gmo.taskClient.List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		logs.Errorf("Get list task err:%v", err)
	}

	var err2 error
	for _, t := range list.Items { //遍历etcd当中的所有task
		if t.Status.TaskID == taskID { // 如果taskId对上了，则就修改该Task的Phase为checking
			task, err2 = gmo.taskClient.Get(context.TODO(), t.Name, metav1.GetOptions{})
			if err2 != nil {
				logs.Error("Get task by taskID error from etcd:%v", err2)
			}
		}
	}
	if task.Status.Phase == apis.Successed { // 有可能有这种情况，就是有多个迁移的group，可能其中一个已经执行了该方法，将Task的phase改为了Succeed了
		return
	}
	// 主要是遍历当前group的兄弟group，看是否执行完成了，如果没有的话就不啥也不做，让最后一个执行完成的group去修改Task状态（注意：啥也不用做，就是说修改Task状态为Failed，也不用做，其他地方有逻辑会完成的）
	var otherGroupIsSuccessed = true
	for j := range task.Status.GroupStatus {
		groupStatus := &task.Status.GroupStatus[j]
		group := &task.Spec.Groups[j]
		if group.Name != gro.Name { // 遍历到当前Group的兄弟group，兄弟group只有Succeed、Failed、Migrated状态
			if groupStatus.Phase == apis.Migrated { //其他group也迁移走了，这里其实得查该group对应副本group的完成情况
				if groupStatus.CopyStatus != "Successed" { // 兄弟group的副本group是Failed或者正在运行，copy_status为""
					otherGroupIsSuccessed = false
				}
				//copyGroup, err := gmo.groupClient.Get(context.TODO(), "Reason-Copy", metav1.GetOptions{})
				//if err != nil {
				//	logs.Errorf("Get copy group err:%v", err)
				//}
				//if copyGroup.Status.Phase == apis.DeployCheck || copyGroup.Status.Phase == apis.Running {
				//	otherGroupIsSuccessed = false
				//}
			}
			if groupStatus.Phase == apis.Successed {
				continue
			}
			if groupStatus.Phase == apis.DeployCheck || groupStatus.Phase == apis.Running {
				otherGroupIsSuccessed = false
			}
			// apis.Failed 这个状态其实不用考虑了，因为当其他group的状态为Failed的话，就会主动该Task的状态为Failed
		}
	}
	if !otherGroupIsSuccessed {
		return
	}
	patchTask, err := json.Marshal(map[string]interface{}{
		"status": map[string]interface{}{
			"phase":     apis.Successed,
			"finish":    nowTime,
			"last_time": nowTime,
		},
	})
	_, err = gmo.taskClient.Patch(context.TODO(), task.Name, types.StrategicMergePatchType, patchTask, metav1.PatchOptions{})
	if err != nil {
		logs.Errorf("Patch task err:%v", err)
	}
}

// 源任务的Runtime已完成，修改副本Group的Runtime的状态，这样副本的Runtime就不用实际去执行，维护这个状态即可
func (gmo *GroupMonitor) updateCopyIngfoForRuntime(groupSpec *apis.GroupSpec, phase apis.Phase, actionIndex, runtimeIndex int) {
	logs.Trace("=======================================================1")
	logs.Trace("#########runtime#############groupSpec.Replicas:%v,groupSpec.Replicas[0] > 0:%v,groupSpec.Replicas[1] > 0:%v", groupSpec.Replicas, groupSpec.Replicas[0] > 0, groupSpec.Replicas[1] > 0)
	if groupSpec.Replicas[0] > 0 { // 当前group有副本，那么需要将该任务对应的副本任务的runtime的结束状态也设置一下
		logs.Info("#######################groupSpec.Replicas > 0#########设置副本runtime的状态为running")
		// 设置副本runtime的状态为Phase（Succeed or Failed）  如果是跨域的话，这里估计还得再修改
		for key := range groupSpec.CopyInfo {
			if groupSpec.CopyInfo[key] != "local" {
				continue
			}
			copyGroupName := key
			patchGroup, err := json.Marshal(map[string]interface{}{
				"op":    "replace",
				"path":  "/status/action_status/" + strconv.Itoa(actionIndex) + "/status/" + strconv.Itoa(runtimeIndex) + "/copy_status",
				"value": phase,
			})
			if err != nil {
				logs.Errorf("Marshal patch group err:%v", err)
			}
			_, err = gmo.groupClient.Patch(context.TODO(), copyGroupName, types.JSONPatchType, patchGroup, metav1.PatchOptions{})
			if err != nil {
				logs.Errorf("Patch group err-7:%v", err)
			}
			logs.Info("#######################groupSpec.Replicas > 0#########设置副本runtime的状态为running----成功")
		}
	}
	if groupSpec.Replicas[1] > 0 {
		for key := range groupSpec.CopyInfo {
			if groupSpec.CopyInfo[key] != "local" { // 说明是跨域的
				// TODO 使用跨域组件连接另一个域，修改group信息
				//copyGroupName := key // 从copyInfo当中获取到副本的名字Name
				//patchGroup, err := json.Marshal(map[string]interface{}{
				//	"op":    "replace",
				//	"path":  "/status/action_status/" + strconv.Itoa(actionIndex) + "/status/" + strconv.Itoa(runtimeIndex) + "/copy_status",
				//	"value": phase,
				//})
				//if err != nil {
				//	logs.Errorf("Marshal patch group err:%v", err)
				//}
			}
		}
	}
}

// 源任务的Action已完成，修改副本Group的Action的状态，这样副本Action就不用实际去执行，维护这个状态即可
func (gmo *GroupMonitor) updateCopyIngfoForAction(groupSpec *apis.GroupSpec, phase apis.Phase, actionIndex int) {
	logs.Trace("Start#########action#############groupSpec.Replicas:%v,groupSpec.Replicas > 0:%v", groupSpec.Replicas[0], groupSpec.Replicas[0] > 0)
	if groupSpec.Replicas[0] > 0 { // 当前group有副本，那么需要将该任务对应的副本任务的action的结束状态也设置一下
		logs.Info("#######################groupSpec.Replicas > 0#########设置副本action的状态为running")
		// 设置副本runtime的状态为Phase（Succeed or Failed）  如果是跨域的话，这里估计还得再修改
		for key := range groupSpec.CopyInfo {
			if groupSpec.CopyInfo[key] != "local" {
				continue
			}
			copyGroupName := key
			patchGroup, err := json.Marshal([]map[string]interface{}{
				{
					"op":    "replace",
					"path":  "/status/action_status/" + strconv.Itoa(actionIndex) + "/copy_status",
					"value": phase,
				},
			})
			if err != nil {
				logs.Errorf("Marshal patch group err:%v", err)
			}
			_, err = gmo.groupClient.Patch(context.TODO(), copyGroupName, types.JSONPatchType, patchGroup, metav1.PatchOptions{})
			if err != nil {
				logs.Errorf("Patch group err-7:%v", err)
			}
			logs.Info("#######################groupSpec.Replicas > 0#########设置副本action的状态为running-----成功")
		}
	}
	if groupSpec.Replicas[1] > 0 {
		for key := range groupSpec.CopyInfo {
			if groupSpec.CopyInfo[key] != "local" { // 说明是跨域的
				// TODO 使用跨域组件连接另一个域，修改group信息
				//copyGroupName := key // 从copyInfo当中获取到副本的名字Name
				//patchGroup, err := json.Marshal([]map[string]interface{}{
				//	{
				//		"op":    "replace",
				//		"path":  "/status/action_status/" + strconv.Itoa(actionIndex) + "/copy_status",
				//		"value": phase,
				//	},
				//})
				//if err != nil {
				//	logs.Errorf("Marshal patch group err:%v", err)
				//}
			}
		}
	}
}

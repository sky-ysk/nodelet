package monitor

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"reflect"
	"sync"
	"time"

	"hit.edu/framework/pkg/client-go/util/manager"
	"hit.edu/framework/pkg/utils"
	cross_core "hit.edu/framework/test/etcd_sync/active/clients/typed/core"

	"hit.edu/framework/pkg/apimachinery/types"
	apis "hit.edu/framework/pkg/apis/cores"
	metav1 "hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/client-go/tools/recorder"
	"hit.edu/framework/pkg/component-base/logs"
	"hit.edu/framework/pkg/nodelet/events"
	"hit.edu/framework/pkg/nodelet/events/eventbus"
	group "hit.edu/framework/pkg/nodelet/task/group"
	"hit.edu/framework/pkg/nodelet/task/group/dependency"
	"hit.edu/framework/pkg/nodelet/task/runtime"
)

// /* TODO
// 监控任务的执行状态-2
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
	// eventRecorder 记录事件
	recorder recorder.EventRecorder
	// 管理运行所需的Runtime
	// 存储RuntimeManager
	runtimeManager *runtime.RuntimeManager
	// 管理依赖
	dependencyManager *dependency.DependencyManager
	// conditionEngine
	conditionEngine *utils.ConditionEngine
	////Client-go
	//nodesClient   core.NodeInterface //需要查node信息
	//groupClient   core.GroupInterface
	//taskClient    core.TaskInterface
	//actionClient  core.ActionInterface
	//runtimeClient core.RuntimeInterface
	clientsManager *manager.Manager
	stopCh         chan struct{}
	// 跨域
	groupTargets   map[string]cross_core.GroupInterface
	actionTargets  map[string]cross_core.ActionInterface
	runtimeTargets map[string]cross_core.RuntimeInterface
	// 维护一个Map存Group的Task，这样可以避免查Group的parents-Group的时候，还得先去查Task
	belongTasks map[string]*apis.Task
}

func NewGroupMonitor(groupManager group.Manager, groupQueues *group.GroupQueues, eventbus *eventbus.EventBus, recorder recorder.EventRecorder,
	runtimeManager *runtime.RuntimeManager, clientsManager *manager.Manager, dependencyManager *dependency.DependencyManager, conditionEngine *utils.ConditionEngine,
	groupTarget map[string]cross_core.GroupInterface, actionTarget map[string]cross_core.ActionInterface, runtimeTarget map[string]cross_core.RuntimeInterface) *GroupMonitor {
	return &GroupMonitor{
		groupManager:      groupManager,
		groupQueues:       groupQueues,
		eventBus:          eventbus,
		recorder:          recorder,
		runtimeManager:    runtimeManager,
		dependencyManager: dependencyManager,
		//nodesClient:       nodeClient,
		//groupClient:       groupClient,
		//taskClient:        taskClient,
		//actionClient:      actionClient,
		//runtimeClient:     runtimeClient,
		clientsManager: clientsManager,
		stopCh:         make(chan struct{}),
		groupTargets:   groupTarget,
		actionTargets:  actionTarget,
		runtimeTargets: runtimeTarget,
		belongTasks:    make(map[string]*apis.Task),
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
				logs.Trace("依赖检查协程收到关闭通知，正在退出...")
				return // 退出协程
			case <-ticker.C: // 每隔一段时间执行一次更新依赖操作
				logs.Trace("定期检查机器的虚拟环境依赖")
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
				getGroup, err := gmo.clientsManager.GetGroup(gr.Name, gr.Namespace)
				if err != nil {
					logs.Errorf("Etcd get group error-1:%v", err)
				}
				task, exists := gmo.belongTasks[gr.Name]
				if exists {
					logs.Tracef("Group:%v's belong Task:%v is already find", gr.Name, task.Name)
				} else {
					logs.Tracef("Group:%v's belong task has't find", gr.Name)
					task, err = gmo.clientsManager.GetTask(gr.Status.Belong.Name, gr.Status.Belong.Namespace)
					if err != nil {
						logs.Errorf("Etcd get task error:%v", err)
					}
					gmo.belongTasks[gr.Name] = task
				}
				if !gmo.groupDepenSatisfy(getGroup, task) { //再次检查group的执行依赖是否满足了（注意：group_workers当中任务头一次执行前也会检查）
					//logs.Debugf("The group ：%s execution dependency is not satisfied again, now still in Checking Queue", gr.Name)
					//继续放在checking队列当中，checking队列会持续检查依赖，直到依赖满足后，才开始执行，重新将任务group交给group_workers去执行
					grou, err := gmo.groupManager.GetGroupByName(getGroup.Name) // 目前打算把一些小的参数存到本地内存当中的groupManager当中，这样可以减轻访问api-server的压力
					if err != nil {
						logs.Errorf("Get group by id failed, err:%v", err)
					}
					grou.Status.CheckDependencyCount++
					if grou.Status.CheckDependencyCount > 10000 { // 当检查依赖的次数大于1000次的话，说明依赖还是满足不了，迁移至Error队列---这里其实有问题（group如果有前序依赖，你不知道什么时候其前序依赖能完成），暂停1000s其实也是有问题的
						//将任务迁移到Error队列当中
						ok := gmo.groupQueues.DeleteFromCheckingAndAddToError(grou.Name)
						if !ok {
							logs.Error("Delete group from checking queue and add to error queue failed")
						}
						gmo.handleStatusUpdate(getGroup, apis.Failed) //设置group状态为
						continue
					}
					//// TODO 这里后期得优化
					//patchGroup, err := json.Marshal(map[string]interface{}{
					//	"status": map[string]interface{}{
					//		"check_dependency_count": getGroup.Status.CheckDependencyCount,
					//	},
					//})
					//if err != nil {
					//	logs.Errorf("Json Marshal failed, err:%v", err)
					//}
					//_, err = gmo.groupClient.Patch(context.TODO(), getGroup.Name, types.StrategicMergePatchType, patchGroup, metav1.PatchOptions{})
					//if err != nil {
					//	logs.Errorf("Patch group error-1:%v", err)
					//}
					continue
				} else { //说明group执行的依赖已经满足，接下来开始执行
					// 任务依赖满足后就将任务从checking队列转移至Running队列，为了适配迁移，同时适配副本任务,若为副本任务，则转移到CopyPending队列当中
					if getGroup.Spec.IsCopy {
						ok := gmo.groupQueues.DeleteFromCheckingAndAddToCopyPending(getGroup.Name)
						if !ok {
							logs.Error("Delete group from checking queue and add to copy pending queue failed")
						}
					} else { //非副本任务，则直接移入到Running队列当中去运行任务
						ok := gmo.groupQueues.DeleteFromCheckingAndAddToRunning(getGroup.Name)
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
				group, err := gmo.clientsManager.GetGroup(gro.Name, gro.Namespace)
				if err != nil {
					logs.Errorf("Etcd get group error-2:%v", err)
				}
				if group.Status.CopyStatus == "Waiting" { // 说明副本任务是提前部署好的
					// 这里打算Init初始化group,就是提前进行Running步骤  源任务一个Runtime执行完成后，就修改副本runtime的状态即可，Action执行完成后，也会修改副本Runtime的状态
					var isSuccess bool                                     // 标记group下面的action是否都执行成功，如果都执行完了，还没有触发迁移，那么关闭副本即可
					for _, actionReference := range group.Status.Actions { // 遍历group当中的Action
						action, err := gmo.clientsManager.GetAction(actionReference.Name, actionReference.Namespace)
						if err != nil {
							logs.Errorf("Etcd get action error-2:%v", err)
						}
						actionStatus := &action.Status
						isSuccess = true
						if actionStatus.CopyStatus == "Running" {
							isSuccess = false
							//logs.Info("***************************************************************Running")
							for _, runtimeReference := range actionStatus.Runtimes {
								runtime, err := gmo.clientsManager.GetRuntime(runtimeReference.Name, runtimeReference.Namespace)
								if err != nil {
									logs.Errorf("Get runtime error-2:%v", err)
								}
								runtimeStatus := &runtime.Status
								if runtimeStatus.CopyStatus == "Running" { //说明源任务当中的该runtime已经Running了
									// 这里将副本group中的该runtime进行判断，如果是细粒度控制的，就进行init
									if runtime.Spec.EnableFineGrainedControl && !runtimeStatus.Initing && gmo.runtimeDepenSatisfy(runtime, action) { //细粒度控制 且 还未Init初始化过
										// 启动runtimeStatus的Init方法  --TODO 这里为啥不用依赖检查呢？因为源任务能Running，说明这个runtime是依赖是满足的，所以默认副本对应的runtime依赖也是满足的（所以这里得加一点，就是在init阶段，检查一下依赖再init也OK---最好是这样）
										logs.Info("#############################Init#######################################")
										go gmo.runtimeManager.InitRuntime(group, action, runtime, action.Spec.Name, runtime.Spec.Name) // TODO init方法当中最好也能发送一个事件
										//grou.Status.ActionStatus[actionIndex].RuntimeStatus[runtimeIndex].Initing = true               // 标记该runtime是细粒度控制，且开启了Init初始化，因为对于细粒度控制的runtime，如果没有预部署，直接切换的话，不会调用Init方法？
										//grou.Status.ActionStatus[actionIndex].RuntimeStatus[runtimeIndex].Waiting = true               // 这个参数标记用来防止接下来进入Running队列的Init检查的时候，Runtime被执行多次，在启动Runtime后，将Waiting属性置为false就能防止Runtime被执行多次了
										patchRuntime, err := json.Marshal(map[string]interface{}{
											"status": map[string]interface{}{
												"initing": true, // 标记该runtime是细粒度控制，且开启了Init初始化，因为对于细粒度控制的runtime，如果没有预部署，直接切换的话，不会调用Init方法？
												"waiting": true, // 这个参数标记用来防止接下来进入Running队列的Init检查的时候，Runtime被执行多次，在启动Runtime后，将Waiting属性置为false就能防止Runtime被执行多次了
											},
										})
										_, err = gmo.clientsManager.PatchRuntime(runtime.Name, runtime.Namespace, patchRuntime)
										if err != nil {
											logs.Errorf("Patch runtime error-2:%v", err)
										}
									} else { // 如果不是细粒度的，那么就标记该runtime的Waiting属性为Waiting（其状态仍然是DeployCheck）
										//grou.Status.ActionStatus[actionIndex].RuntimeStatus[runtimeIndex].Waiting = true // TODO应该是为了适配进入到Running队列的DeployCheck检查后，会重复执行（因为当前状态为DeployCheck状态，进入到Running队列，有可能还是DeployCheck状态，对于DeployCheck状态，需要考虑该runtime是否处于等待的过程），这里的runtime，其实就是处于一种等待的过程
										patchRuntime, err := json.Marshal(map[string]interface{}{
											"status": map[string]interface{}{
												"waiting": true, // TODO应该是为了适配进入到Running队列的DeployCheck检查后，会重复执行（因为当前状态为DeployCheck状态，进入到Running队列，有可能还是DeployCheck状态，对于DeployCheck状态，需要考虑该runtime是否处于等待的过程），这里的runtime，其实就是处于一种等待的过程
											},
										})
										_, err = gmo.clientsManager.PatchRuntime(runtime.Name, runtime.Namespace, patchRuntime)
										if err != nil {
											logs.Errorf("Patch runtime error-2:%v", err)
										}
									}
								}
								if runtimeStatus.CopyStatus == "Succeeded" { // 说明源任务当中的runtime已经运行完成了
									// 这里将副本group中的该runtime进行判断，如果是细粒度控制的，且进行了初始化的工作话，就关闭Init初始化
									if runtime.Spec.EnableFineGrainedControl && runtimeStatus.Initing { // 注意这里是用内存当中的group_manager当中的信息Initing
										// 关闭runtime，直接关闭runtime进程
										logs.Info("(((((((((((((((((((((((((((((((((((((((-1")
										gmo.runtimeManager.Kill(group, action, runtime, action.Spec.Name, runtime.Spec.Name)
									} else { // 如果说runtime不是细粒度的，那这里源任务完成后，副本runtime的状态是不会主动修改的（因为副本runtime没启动无法调用Kill函数来修改runtime状态），所以这里需要主动修改runtime的状态为Succeed
										gmo.handleRuntimeSucceedUpdate(action, runtime)
									}
								}
							}
							continue // 这样能快速遍历下一个Action，不然还会进入下面的判断，稍微好一丢丢
						}
						if actionStatus.CopyStatus == "Succeeded" { // 这里有个小插曲，就是对于副本任务里面改Action，其下面的Runtime的状态没有改为Succeed，可以补充进来--  这是为啥呢？因为这里是一层层遍历，虽然说源任务runtime完成、Action完成会同时修改副本的runtime、Action，但是由于这里的逻辑是先遍历到Action，然后再遍历到下面的runtime，这里选择不再遍历下去，这样会很，直接遍历到Action状态为Successed，然后调用一个方法将Action下面的所有runtime的Phase改为Succeed即可
							logs.Info("***************************************************************Succeeed")
							for _, runtimeReference := range actionStatus.Runtimes {
								runtime, err := gmo.clientsManager.GetRuntime(runtimeReference.Name, runtimeReference.Namespace)
								if err != nil {
									logs.Errorf("Get runtime error-2:%v", err)
								}
								runtimeStatus := &runtime.Status
								if runtime.Spec.EnableFineGrainedControl && runtimeStatus.Initing { // 如果是细粒度控制的，且进行了初始化的工作话，就关闭Init初始化
									logs.Info("(((((((((((((((((((((((((((((((((((((((-2")
									gmo.runtimeManager.Kill(group, action, runtime, action.Spec.Name, runtime.Spec.Name)
								} else { // 如果说runtime不是细粒度的，那这里源任务完成后，副本runtime的状态是不会主动修改的（因为副本runtime没启动无法调用Kill函数来修改runtime状态），所以这里需要主动修改runtime的状态为Succeed
									gmo.handleRuntimeSucceedUpdate(action, runtime)
								}
							}
							//gmo.handleCopyActionSucceedUpdate(group, actionIndex) //修改副本任务的Action、Runtime为Successed的状态（因为源任务的Action、Runtime执行完了，副本任务的Action、Runtime没必要执行，直接设置phase为成功即可）
							continue //说明当前Action执行完成了，接着查看下一个Action的执行情况
						}
						if actionStatus.CopyStatus == "Failed" { //TODO 这里有个小插曲，就是对于副本任务里面的其他Action、Runtime的状态没有改为Failed，以后可以补充进来
							isSuccess = false
							// 修改副本group的状态为Failed
							//groupCopyName := "Reason-Copy"
							//getCopyGroup, err := gmo.groupClient.Get(context.TODO(), groupCopyName, metav1.GetOptions{})
							//if err != nil {
							//	logs.Errorf("Get group by id failed, err:%v", err)
							//}
							gmo.handleCopyRuntimeFailedUpdate(group, action) // 源任务执行失败了，直接标记副本任务的runtime、action状态为Failed，那么副本group的状态也直接被标记为Failed
							logs.Info("@@@@@@@@@@@@@@@@@@@@@@@@@@@@@!=2")
							ok := gmo.groupQueues.DeleteFromCopyPendingAndAddToCompleted(group.Name) // 源任务失败后被移入Error队列来通知调度器，这里副本任务直接移入成功队列即可
							if !ok {
								logs.Error("Delete group from copypending queue and add to completed pending queue failed")
							}
						}
						if actionStatus.CopyStatus == "" {
							isSuccess = false
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
						_, err = gmo.clientsManager.PatchGroup(group.Name, group.Namespace, patchGroup)
						if err != nil {
							logs.Errorf("Patch group error-9:%v", err)
						}
						//将任务迁移到Completed队列当中
						logs.Info("Move to completed queue")
						logs.Info("@@@@@@@@@@@@@@@@@@@@@@@@@@@@@!=1")
						ok := gmo.groupQueues.DeleteFromCopyPendingAndAddToCompleted(group.Name)
						if !ok {
							logs.Errorf("Delete group from running queue and add to completed queue failed")
						}
						continue
					}
				} else if group.Status.CopyStatus == "Starting" {
					logs.Infof("=============CopyPending--Starting")
					ok := gmo.groupQueues.DeleteFromCopyPendingAndAddToRunning(group.Name) // 有两种情况，一种是无副本情况的瞬时迁移，另一种是副本任务触发了迁移
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
				group, err := gmo.clientsManager.GetGroup(gro.Name, gro.Namespace)
				if err != nil {
					logs.Errorf("Etcd get group error-3:%v", err)
				}
				var AllactionisSuccess = true                          // 标记group下面的action是否都执行成功
				for _, actionReference := range group.Status.Actions { // 遍历group当中的Action
					action, err := gmo.clientsManager.GetAction(actionReference.Name, actionReference.Namespace)
					if err != nil {
						logs.Errorf("Etcd get action error-4:%v", err)
					}
					actionStatus := &action.Status
					// 为了适配迁移，状态为Migrated也说明Action成功结束了，然后接下来就通过Action成功标记Group成功了
					if actionStatus.Phase == apis.Unknown {
						AllactionisSuccess = false
					}
					if actionStatus.Phase == apis.Successed || actionStatus.Phase == apis.Migrated { // 当前action的状态为Successed或者Migrated
						continue //说明当前Action执行完成了，接着查看下一个Action的执行情况
					}
					if actionStatus.Phase == apis.Failed { //注意:handleRuntimeEndUpdate方法当中，当Runtime状态失败时，同时也会修改其Action的状态为Failed，同时也会修改Group为Failed、Task为Failed
						AllactionisSuccess = false
						//将任务迁移到Error队列当中
						ok := gmo.groupQueues.DeleteFromRunningAndAddToError(group.Name)
						if !ok {
							logs.Errorf("Delete group from running queue and add to running queue failed")
						}
						logs.Info("put group:%v into error queue", group.Name)
						break // 这里直接跳出for循环即可，因为该group已经是Failed了，不用看了
					}

					if actionStatus.Phase == apis.DeployCheck { //
						AllactionisSuccess = false
						if actionStatus.Waiting == true { //说明是第二次遍历到了这个Action，第一次遍历到该Action的时候，其依赖没有满足
							if !gmo.actionDepenSatisfy(action, group) {
								//logs.Infof("Action %s depends on parent action, parent not finish ", action.Name)
								continue
							}
							patchAction, err := json.Marshal(map[string]interface{}{
								"status": map[string]interface{}{
									"waiting": false, // 说明Action的父亲Action已经执行完成了，那么接下来Action的Runtime必须会被执行（至少会执行一个runtime）
								},
							})
							_, err = gmo.clientsManager.PatchAction(action.Name, action.Namespace, patchAction)
							if err != nil {
								logs.Errorf("Patch action error-21:%v", err)
							}
							//grou.Status.ActionStatus[actionIndex].Waiting = false
						} else { //说明是第一次遍历到了这个Action，当Action的依赖没有满足的时候，要设置Waiting属性为true
							if !gmo.actionDepenSatisfy(action, group) {
								logs.Infof("Action:%s in group:%s waiting for dependencies", action.Name, group.Name)
								//action.Status.Waiting = true //第一次执行时发现执行不了，那就交给running队列去检查
								patchAction, err := json.Marshal(map[string]interface{}{
									"status": map[string]interface{}{
										"waiting": true, // 说明Action的父亲Action已经执行完成了，那么接下来Action的Runtime必须会被执行（至少会执行一个runtime）
									},
								})
								_, err = gmo.clientsManager.PatchAction(action.Name, action.Namespace, patchAction)
								if err != nil {
									logs.Errorf("Patch action error-21:%v", err)
								}
								//grou.Status.ActionStatus[actionIndex].Waiting = true
								logs.Debugf("StartGroup method:ActionWaiting:%v, i:%v", actionStatus.Waiting, i)
								continue
							}
						}
						for _, runtimeReference := range actionStatus.Runtimes {
							logs.Trace("-------------------------------------------get Runtime----------------------------,runtimeReferenceName:%v,runtimeReferenceNamespace:%v", runtimeReference.Name, runtimeReference.Namespace)
							runtime, err := gmo.clientsManager.GetRuntime(runtimeReference.Name, runtimeReference.Namespace)
							if err != nil {
								logs.Errorf("Get runtime error-3:%v", err)
							}
							runtimeStatus := &runtime.Status
							if runtimeStatus.Waiting == true { // 说明是第二次遍历到这个runtime，第一次遍历到该runtime的时候，其依赖没有满足（判断Runtime是否处于等待）
								if !gmo.runtimeDepenSatisfy(runtime, action) { // runtime 依赖不满足
									continue
								}
								patchRuntime, err := json.Marshal(map[string]interface{}{
									"status": map[string]interface{}{
										"waiting": false, //runtime依赖已经满足，此时设置为false，就不会继续往下执行，去启动任务了，这里的设置很关键
									},
								})
								_, err = gmo.clientsManager.PatchRuntime(runtime.Name, runtime.Namespace, patchRuntime)
								if err != nil {
									logs.Errorf("Patch runtime error-2:%v", err)
								}
								//grou.Status.ActionStatus[actionIndex].RuntimeStatus[runtimeIndex].Waiting = false //runtime依赖已经满足，此时设置为false，就不会继续往下执行，去启动任务了，这里的设置很关键
							} else {
								if !gmo.runtimeDepenSatisfy(runtime, action) { // runtime 依赖不满足
									patchRuntime, err := json.Marshal(map[string]interface{}{
										"status": map[string]interface{}{
											"waiting": true, //runtime依赖已经满足，此时设置为false，就不会继续往下执行，去启动任务了，这里的设置很关键
										},
									})
									_, err = gmo.clientsManager.PatchRuntime(runtime.Name, runtime.Namespace, patchRuntime)
									if err != nil {
										logs.Errorf("Patch runtime error-2:%v", err)
									}
									//grou.Status.ActionStatus[actionIndex].RuntimeStatus[runtimeIndex].Waiting = true // 设置Runtime处于等待
									continue
								}
							}
							//说明runtime可以执行，这里为了适配迁移，如果是副本group，初始化任务的时候直接使用关键状态数据
							// 这块可以执行到，因为group当中有很多action，有多个Action的话，总有没执行的Action，这时候需要判断runtime的启动方式（细粒度的话使用StartingRuntime启动、粗粒度的话使用Run启动）
							if runtime.Spec.EnableFineGrainedControl { // 当前group是副本任务，且实现了细粒度控制方法
								if !runtime.Status.Starting { // 还没有调用Start或者Run方法，说明第一次进入  ----Start参数主要解决的问题：Action、Runtime的依赖都满足，且调用了Run、Start方法进入了Runtime的运行时，但是卡在运行时，没有将Runtime、Action的状态设置为Running，导致一直重复进入Action.Status== apis.Deploycheck这个分支
									logs.Infof("****************************hhhhhhhhhhhhhhhh****************************************")
									//if !grou.Status.ActionStatus[actionIndex].RuntimeStatus[runtimeIndex].Starting { //TODO 这个参数好像可以删了，有Waiting是不是就够了？
									logs.Infof("========================runtimeStatus.KeyStatus:%v,runtimeStatus.KeyStatus == \"\"", runtimeStatus.KeyStatus, runtimeStatus.KeyStatus == "")
									if runtimeStatus.KeyStatus == "" { // 说明不是副本任务，还没初始化---TODO 这里需要这个检查的原因：有可能是即时的迁移迁移，那迁移过去的group是没有进入init状态的，所以这边迁移过去的副本是处于DeployCheck的状态开始恢复任务状态
										logs.Info("))))))))))))))))))))))))))))))))))))))))))))))")
										go gmo.runtimeManager.StartRuntime(group, action, runtime, action.Spec.Name, runtime.Spec.Name)
									} else {
										logs.Info("%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%")
										go gmo.runtimeManager.StartRuntime(group, action, runtime, action.Spec.Name, runtime.Spec.Name) //这句好像会阻塞
										go gmo.runtimeManager.RestoreData(group, action, runtime, action.Spec.Name, runtime.Spec.Name)
									}
									patchRuntime, err := json.Marshal(map[string]interface{}{
										"status": map[string]interface{}{
											"starting": true,
										},
									})
									logs.Infof("&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&-1,name:%v,namespace:%v", runtime.Name, runtime.Namespace)
									_, err = gmo.clientsManager.PatchRuntime(runtime.Name, runtime.Namespace, patchRuntime)
									if err != nil {
										logs.Errorf("patch runtimeStatus error")
									}
									logs.Info("&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&-2")
								}
							} else {
								// ①副本任务，但没有细粒度控制 ②原任务（没有副本） 采用Run方式启动任务
								logs.Infof("****************************ashdkhaskldhklashdk****************************************")
								if !runtime.Status.Starting {
									go gmo.runtimeManager.Run(group, action, runtime, action.Spec.Name, runtime.Spec.Name)
									patchRuntime, err := json.Marshal(map[string]interface{}{
										"status": map[string]interface{}{
											"starting": true,
										},
									})
									_, err = gmo.clientsManager.PatchRuntime(runtime.Name, runtime.Namespace, patchRuntime)
									if err != nil {
										logs.Errorf("patch runtimeStatus error")
									}
								}
							}
						}
						continue
					}
					//检查runtime、Action当中的parents是否执行完成，如果父亲节点完成，则让他执行 ---这里有bug，就是任务已经放入running队列，但是还没执行完，这时候runningCheck循环遍历到当前runtime的状态为checking，查看是否满足执行条件，发现是满足的，结果有跑起来该任务
					if action.Status.Phase == apis.Running {
						AllactionisSuccess = false
						var allRuntimeIsSuccessed = true
						for _, runtimeReference := range actionStatus.Runtimes {
							runtime, err := gmo.clientsManager.GetRuntime(runtimeReference.Name, runtimeReference.Namespace)
							if err != nil {
								logs.Errorf("Get runtime error-4:%v", err)
							}
							runtimeStatus := &runtime.Status
							if runtimeStatus.Phase != apis.Successed {
								allRuntimeIsSuccessed = false
							}
							if !gmo.runtimeDepenSatisfy(runtime, action) {
								//logs.Infof("Runtime %s depends on parent runtime", r.Name)
								//r.Waiting = true  // 这里不需要再标记了，因为在DeployCheck阶段就遍历了所有的runtime并标记了
								continue
							}
							//logs.Infof("$$$$$$==================grou.Spec.Actions[j].Spec.Runtimes[m].Waiting:%v,j:%v,m:%v", grou.Spec.Actions[j].Spec.Runtimes[m].Waiting, j, m)
							if runtimeStatus.Waiting { //如果说runtime也是被标记等待执行的状态，这才能开始执行，这块是防止runtime重复多次运行
								//说明runtime可以执行
								//grou.Spec.Actions[actionIndex].Spec.Runtimes[runtimeIndex].Waiting = false
								//grou.Status.ActionStatus[actionIndex].RuntimeStatus[runtimeIndex].Waiting = false
								patchRuntime, err := json.Marshal(map[string]interface{}{
									"status": map[string]interface{}{
										"waiting": false, //runtime依赖已经满足，此时设置为false，就不会继续往下执行，去启动任务了，这里的设置很关键
									},
								})
								_, err = gmo.clientsManager.PatchRuntime(runtime.Name, runtime.Namespace, patchRuntime)
								if err != nil {
									logs.Errorf("Patch runtime error-6:%v", err)
								}
								if runtime.Spec.EnableFineGrainedControl {
									if runtimeStatus.Initing {
										logs.Info("****************************************KKKKKKKKSS**************************")
										go gmo.runtimeManager.RestoreData(group, action, runtime, action.Spec.Name, runtime.Spec.Name)
									} else {
										logs.Info("****************************************ABCDSDSADSAD**************************")
										go gmo.runtimeManager.StartRuntime(group, action, runtime, action.Spec.Name, runtime.Spec.Name)
									}
									//logs.Info("****************************************ABCDSDSADSAD**************************")
									//if !grou.Status.ActionStatus[actionIndex].RuntimeStatus[runtimeIndex].Starting {
									//	go gmo.runtimeManager.StartRuntime(group, action, runtime, actionIndex, runtimeIndex)
									//	grou.Status.ActionStatus[actionIndex].RuntimeStatus[runtimeIndex].Starting = true
									//}
								} else {
									logs.Info("****************************************1234554564**********************")
									//if !grou.Status.ActionStatus[actionIndex].RuntimeStatus[runtimeIndex].Starting {
									go gmo.runtimeManager.Run(group, action, runtime, action.Spec.Name, runtime.Spec.Name)
									//grou.Status.ActionStatus[actionIndex].RuntimeStatus[runtimeIndex].Starting = true
									//}
								}
							}
						}
						if allRuntimeIsSuccessed { //这里修改Action的phase为Succeed，为啥runtime都成功了，Action还是显示Running，可能就是因为runtime同时完成，没有识别到兄弟runtime已经完成，导致action状态没设置为Successed
							// 更新etcd当中Group下面的Action的phase，Action下面的Runtime的信息都是新的，没问题，这里就更新一下Action的Status就OK了
							patchAction, err := json.Marshal(map[string]interface{}{
								"status": map[string]interface{}{
									"phase":     apis.Successed,
									"finish":    apis.Time{time.Now()},
									"last_time": apis.Time{time.Now()},
								},
							})
							_, err = gmo.clientsManager.PatchAction(action.Name, action.Namespace, patchAction)
							if err != nil {
								logs.Errorf("Patch action error-4:%v", err)
							}
						}
						continue
					}
					//之后这里要考虑迁移的情况，遇到Action为Init的情况，就将Init的runtime启动run/Start起来
					if actionStatus.Phase == apis.Init { //说明当前action下面有Init的runtime了（即：有细粒度控制的runtime）
						//logs.Info("========================================Init")
						AllactionisSuccess = false
						for _, runtimeReference := range action.Status.Runtimes {
							runtime, err := gmo.clientsManager.GetRuntime(runtimeReference.Name, runtimeReference.Namespace)
							if err != nil {
								logs.Errorf("Get runtime error-7:%v", err)
							}
							runtimeStatus := &runtime.Status
							//logs.Info("========================================Init----grou")
							if !gmo.runtimeDepenSatisfy(runtime, action) { // 这里需要runtime的父亲节点的状态也为Successed，所以源runtime成功后，同样需要将副本runtime的Phase设置为Succeed，这个很关键
								//grou.Status.ActionStatus[actionIndex].RuntimeStatus[runtimeIndex].Waiting = true
								patchRuntime, err := json.Marshal(map[string]interface{}{
									"status": map[string]interface{}{
										"waiting": true, //runtime依赖已经满足，此时设置为false，就不会继续往下执行，去启动任务了，这里的设置很关键
									},
								})
								_, err = gmo.clientsManager.PatchRuntime(runtime.Name, runtime.Namespace, patchRuntime)
								if err != nil {
									logs.Errorf("Patch runtime error-8:%v", err)
								}
								continue
							}
							//logs.Info("========================================Init----runtimeDepenSatisfy")
							// 当前遍历到的runtime肯定有Init，当然也会有DeployCheck，执行到此处表示满足依赖，在上面加入一个依赖检查就完美了
							if runtimeStatus.Phase == apis.Init {
								if runtimeStatus.Waiting { // 当前runtime等待启动，这里也是防止多次运行这个runtime，所以上面copyPending队列当中InitRuntime方法下面的Waiting=true是必要的
									//grou.Status.ActionStatus[actionIndex].RuntimeStatus[runtimeIndex].Waiting = false
									patchRuntime, err := json.Marshal(map[string]interface{}{
										"status": map[string]interface{}{
											"waiting": false, // 标记当前runtime马上要启动了，不再Waiting了
										},
									})
									_, err = gmo.clientsManager.PatchRuntime(runtime.Name, runtime.Namespace, patchRuntime)
									if err != nil {
										logs.Errorf("Patch runtime error-8:%v", err)
									}
									// 将初始化的runtime真正启动 为了适配迁移，这里先回复runtime的状态
									logs.Infof("****************************Restore****************************************")
									// 注意这里调用RestoreData后直接启动，需要添加一个设置为Running状态的一个步骤
									go gmo.runtimeManager.RestoreData(group, action, runtime, action.Spec.Name, runtime.Spec.Name)
									//// 将runtime真正的启动
									//time.Sleep(1 * time.Second)
									//logs.Infof("****************************&&&&&&&&&&&&&&&&&&&&&&&&&****************************************")
									//go gmo.runtimeManager.StartRuntime(group, action, runtime, actionIndex, runtimeIndex)
								}
							} else if runtimeStatus.Phase == apis.DeployCheck { // 也有可能碰到的是Init-->Succeed
								if !runtime.Spec.EnableFineGrainedControl { // 执行到这块，只能是粗粒度控制的runtime
									if runtimeStatus.Waiting {
										//grou.Status.ActionStatus[actionIndex].RuntimeStatus[runtimeIndex].Waiting = false
										patchRuntime, err := json.Marshal(map[string]interface{}{
											"status": map[string]interface{}{
												"waiting": false, // 标记当前runtime马上要启动了，不再Waiting了
											},
										})
										_, err = gmo.clientsManager.PatchRuntime(runtime.Name, runtime.Namespace, patchRuntime)
										if err != nil {
											logs.Errorf("Patch runtime error-8:%v", err)
										}
										// 如果runtime不是细粒度控制的话，直接调用Run启动
										logs.Infof("****************************mmmmmmmmmmmmmmmmmmmmmmmmmmmmmmm****************************************")
										go gmo.runtimeManager.Run(group, action, runtime, action.Spec.Name, runtime.Spec.Name)
									}
								}
							}
						}
					}
				}
				if AllactionisSuccess {
					//将任务迁移到Completed队列当中
					//检查group状态是否被被设置为成功，这里做这个检查，主要是应对Group下面的Action同时完成，导致读取etcd的时候，兄弟Action还未被标记为Succeed的情况,Group下面的Action状态都是更新的，就是Group的Status.phase、lasttime没更新
					if group.Status.Phase != apis.Successed {
						// 将Group的状态设置为Successed
						patchGroup, err := json.Marshal(map[string]interface{}{
							"status": map[string]interface{}{
								"phase":     apis.Successed,
								"finish":    apis.Time{time.Now()},
								"last_time": apis.Time{time.Now()},
							},
						})
						_, err = gmo.clientsManager.PatchGroup(group.Name, group.Namespace, patchGroup)
						if err != nil {
							logs.Errorf("Patch group error-10:%v", err)
						}
						// 发送事件
						gmo.recorder.Event(group, apis.EventTypeNormal, events.ExecuteSuccessfully, fmt.Sprintf("Group name:\t %s is successed", group.Name))
					}
					logs.Info("Move to completed queue")
					ok := gmo.groupQueues.DeleteFromRunningAndAddToCompleted(group.Name)
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
				group, err := gmo.clientsManager.GetGroup(gro.Name, gro.Namespace)
				if err != nil {
					logs.Errorf("Etcd get group error-4:%v", err)
				}
				for key, value := range group.Spec.CopyInfo {
					copyGroupName := key
					// 查询副本group的状态是否完成，如果完成了，就将group迁移到Completed队列
					var getGroup *apis.Group
					if group.Spec.CopyInfo[key] == "local" { // 如果副本部署在本域当中
						getGroup, err = gmo.clientsManager.GetGroup(copyGroupName, group.Namespace)
						if err != nil {
							logs.Errorf("Get copy group err:%v", err)
						}

					} else {
						// TODO 通过跨域的连接取到getGroup
						groupTarget := gmo.groupTargets[value]
						getGroup, err = groupTarget.Get(context.TODO(), copyGroupName, metav1.GetOptions{})
						if err != nil {
							logs.Errorf("Get copy group err:%v", err)
						}
					}
					// 副本任务执行完成，为Succeed状态，则标记源任务copy_status为Succeed，如果说为Failed状态，则标记原任务copy_status为Failed状态
					if getGroup.Status.Phase == apis.Successed {
						gmo.handleTaskSucceedUpdate(group) //设置当前group的状态(copyStatus的值)为Succeed，标记其副本任务执行成功了，同时检查group所属Task下面的所有group（除了当前的group）是否都已经成功了，如果都已经成功了，就修改当前Group所属的Task的phase为Succeed
						logs.Info("@@@@@@@@@@@@@@@@@@@@@@@@@@@@@!=4")
						gmo.groupQueues.DeleteFromMigratedAndAddToCompleted(group.Name)
					} else if getGroup.Status.Phase == apis.Failed { // 说明group迁移过去执行失败了，那么需要直接将Task的状态
						// 这里得将group移到error队列当中  注意：如果源任务迁移了，但是迁移的副本任务执行失败了，那么这里将Task的状态改为Failed，同时副本那边会将group移到Error队列上报上去给调度器
						gmo.handleTaskFailedUpdate(group) // 设置当前group的状态（copyStatus的值）为Failed
						logs.Info("@@@@@@@@@@@@@@@@@@@@@@@@@@@@@!=5")
						gmo.groupQueues.DeleteFromMigratedAndAddToCompleted(group.Name)
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
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		timeout := time.Duration(500+r.Intn(1501)) * time.Millisecond
		select {
		case <-ctx.Done():
			return
		case <-time.After(timeout):
			//var start = false
			completed := gmo.groupQueues.GetAllCompleted()
			for i := range completed {
				gro := completed[i]
				// 再次检查Group所属的Task，其下面的Group是否都完成，如果完成，就设置Task状态没设置为Succeed，就再设置Task状态为Successed
				taskName := gro.Status.Belong.Name
				var otherGroupCompleted = true
				task, err := gmo.clientsManager.GetTask(taskName, gro.Status.Belong.Namespace)
				if err != nil {
					logs.Errorf("Get task err:%v", err)
				}
				if task.Status.Phase == apis.Successed || task.Status.Phase == apis.Failed {
					logs.Infof("Delete group:%v", gro.Spec.Name)
					gmo.groupManager.DeleteGroup(gro)
					gmo.groupQueues.DeleteFromCompleted(gro.Name)
					// 删除婴喜爱BelongTask当中这个记录
					if _, exists := gmo.belongTasks[gro.Name]; exists {
						delete(gmo.belongTasks, gro.Name)
						logs.Trace("Delete group:%v in belongTasks map", gro.Spec.Name)
					} else {
						logs.Trace("key not exist")
					}
					continue
				}
				for _, groupReference := range task.Status.Groups {
					if groupReference.Name != gro.Name { //遍历到的group的Name不等于当前处理的Group的Name,即遍历兄弟group
						// 这里得去etcd当中读取Group的最新信息
						logs.Infof("brotherGroupName:%v", groupReference.Name)
						brotherGroup, err := gmo.clientsManager.GetGroup(groupReference.Name, groupReference.Namespace)
						if err != nil {
							logs.Errorf("Etcd get group error-4:%v", err)
						}
						//logs.Infof("groupStatus.Phase:%v,groupID:%v", grStatus.Phase, grStatus.GroupID)
						if brotherGroup.Status.Phase == apis.Running || brotherGroup.Status.Phase == apis.DeployCheck || brotherGroup.Status.Phase == apis.Migrating || brotherGroup.Status.Phase == apis.Init { // 说明其他Group还未执行或者还没迁移成功
							otherGroupCompleted = false
							logs.Infof("groupName:%v has't done, otherGroupCompleted:%v", brotherGroup.Name, otherGroupCompleted)
						}
						if brotherGroup.Status.Phase == apis.Migrated { // 还得去查对应副本任务的状态，如果状态为Running（大概率是这个状态）或者是DeployChek（说明迁移过去的group依赖不满足，暂时还不能执行），那么otherGroupCompleted参数也是false
							// 为了适配迁移，目前还是处理同域的迁移,这里怎么根据源任务找到副本任务，还是一个遗留的问题
							if brotherGroup.Status.CopyStatus == "" { // 副本任务只有Succeed和Failed才会修改源任务的copyStatus，如果说是空，说明副本任务还在运行
								otherGroupCompleted = false
							}
						}
					}
				}
				if otherGroupCompleted && task.Status.Phase != apis.Successed { // 该group的兄弟Group都完成了,且所属的Task没有被设置为Successed，这里需要设置Task的状态为Successed。出现这个情况的原因是：两个Group同时完成，导致读取etcd时均发现双方没有Succeed，故没有修改Task的状态为Successed
					patchTask, err := json.Marshal(map[string]interface{}{
						"status": map[string]interface{}{
							"phase":     apis.Successed,
							"finish":    apis.Time{time.Now()},
							"last_time": apis.Time{time.Now()},
						},
					})
					_, err = gmo.clientsManager.PatchTask(task.Name, task.Namespace, patchTask)
					if err != nil {
						logs.Errorf("Patch task err-2223:%v", err)
					}
					// 发送事件
					logs.Infof("Send task Succeed event")
					gmo.recorder.Event(task, apis.EventTypeNormal, events.ExecuteSuccessfully, fmt.Sprintf("Task Name:\t %s is Successed", task.Name))
					// 更新Task下的最新的Group状态---TODO 这里后续补充吧
				}
				logs.Infof("Delete group:%v", gro.Spec.Name)
				gmo.groupManager.DeleteGroup(gro)
				gmo.groupQueues.DeleteFromCompleted(gro.Name) //得根据groupId进行删除，不是根据groupName
				// 删除婴喜爱BelongTask当中这个记录
				if _, exists := gmo.belongTasks[gro.Name]; exists {
					delete(gmo.belongTasks, gro.Name)
					logs.Trace("Delete group:%v in belongTasks map", gro.Spec.Name)
				} else {
					logs.Trace("key not exist")
				}
				continue
			}
		}
	}
}

//	func (gmo *GroupMonitor) CompletedQueueCheck(ctx context.Context) {
//		logs.Info("Completed queue start checking")
//		//TODO 可能主要是将信息上传到api-server当中，然后将group_manager中的信息删除
//		for {
//			select {
//			case <-ctx.Done():
//				return
//			case <-time.After(time.Second * 10):
//				//var start = false
//				completed := gmo.groupQueues.GetAllCompleted()
//				for i := range completed {
//					gro := completed[i]
//					logs.Infof("Delete group:%v", gro.Spec.Name)
//					gmo.groupManager.DeleteGroup(gro)
//					gmo.groupQueues.DeleteFromCompleted(gro.Name) //得根据groupId进行删除，不是根据groupName
//					// 删除一下BelongTask这个map的记录
//					if _, exists := gmo.belongTasks[gro.Name]; exists {
//						delete(gmo.belongTasks, gro.Name) // 2. 执行删除
//						logs.Trace("Delete group:%v in belongTasks map", gro.Spec.Name)
//					} else {
//						logs.Trace("key not exist")
//					}
//				}
//			}
//		}
//	}
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
				group, err := gmo.clientsManager.GetGroup(gro.Name, gro.Namespace)
				if err != nil {
					logs.Errorf("Etcd get group error-4:%v", err)
				}
				//TODO 通知调度器
				gmo.recorder.Event(group, apis.EventTypeWarning, events.GroupRunError, fmt.Sprintf("Group name:%v run error", group.Name))
				// 删除内存当中group_manager当中的group信息
				gmo.groupQueues.DeleteFromError(gro.Name)
				gmo.groupManager.DeleteGroup(gro) //groupManager就删除group的信息，此时group的信息就只存在于etcd当中
				// 删除一下BelongTask这个map的记录
				if _, exists := gmo.belongTasks[gro.Name]; exists {
					delete(gmo.belongTasks, gro.Name) // 2. 执行删除
					logs.Trace("Delete group:%v in belongTasks map", gro.Spec.Name)
				} else {
					logs.Trace("key not exist")
				}
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
	getGroup, err := gmo.clientsManager.GetGroup(groupName, event.GroupNamespace)
	if err != nil {
		logs.Errorf("Failed get group:%v from etcd, err:%v", groupName, err)
	}

	actionSpecName := event.ActionSpecName
	runtimeSpecName := event.RuntimeSpecName
	phase := event.Phase // 这里接收的Phase有可能是running，也有可能是Failed
	logs.Infof("Get runtime start event notify, the phase:%s", phase)
	startTime := event.StartAt
	lastTime := event.LastTime
	groupSpec := &getGroup.Spec
	groupStatus := &getGroup.Status
	//var actionStart = true              //action是否需要标记启动（下面的runtime如果都没启动，则说明action要标记Running）
	taskName := getGroup.Status.Belong.Name // group的Belongs属性当中的TaskName(全局唯一的)
	// 是否为副本任务
	isCopyGroup := getGroup.Spec.IsCopy
	var taskIsFirstSet = false  // 为了适配发送最全的Task信息
	var groupIsFirstSet = false // 为了适配发送最全的Group信息
	var taskIsFailed = false    // 该runtime所属的Task是否为Failed
	var groupIsFailed = false   // 该runtime所属的Group是否为Failed
	var actionIsFailed = false  // 该runtime所属的action是否为Failed

	//var groupIndexInTask int //当前group在Task当中的下标
	var task *apis.Task     //runtime所在的Task对象
	var action *apis.Action //runtime所在的Action对象
	var runtime *apis.Runtime
	if !isCopyGroup { // 不是副本group，要修改所属的Task的状态
		// 首先先修改该runtime所对应的group-所对应的Task的phase
		task, err = gmo.clientsManager.GetTask(taskName, getGroup.Status.Belong.Namespace)
		if err != nil {
			logs.Errorf("Get task error from etcd:%v", err)
		}
		// ## 处理Task的Phase
		// 遍历Task下面的Group，根据Group的状态来设置Task的Phase
		for gSpecName, _ := range task.Status.Groups {
			if gSpecName != groupSpec.Name { // 非当前处理的group，跳过
				continue
			}
			if task.Status.Phase == apis.DeployCheck { // 说明是Task中的第一个Group启动，那说明Task也是第一次启动，要标记状态为Phase（running or Failed）
				if phase == apis.Failed {
					task.Status.FinishAt = &startTime //任务在执行的时候就发送错误，那么开始和结束时间都标记为同一时刻
				}
				task.Status.StartAt = &startTime
				task.Status.Phase = phase
				task.Status.LastTime = &lastTime
				taskIsFirstSet = true
			}
		}
	}

	// ## 处理Group的Phase
	//修改group下面的groupStatus下面的ActionStatus，ActionStatus下面的RuntimeStatus
	for aSpecName, actionReference := range groupStatus.Actions { //GroupSpec(Actions)---->ActionStatus----->RuntimeStatus
		if aSpecName != actionSpecName { // 判断是否是当前处理的Action
			continue // 不是的话跳过
		}
		action, err = gmo.clientsManager.GetAction(actionReference.Name, actionReference.Namespace)
		if err != nil {
			logs.Errorf("Get action error from etcd:%v", err)
		}
		actionStatus := &action.Status //ActionStatus
		logs.Trace("=======================================================0")
		for rSpecName, runtimeReference := range actionStatus.Runtimes { //RuntimeStatus
			if rSpecName == runtimeSpecName { // 是当前处理的runtime
				runtime, err = gmo.clientsManager.GetRuntime(runtimeReference.Name, runtimeReference.Namespace)
				if err != nil {
					logs.Errorf("Get runtime error to etcd-11:%v", err)
				}
				runtimeStatus := &runtime.Status
				if phase == apis.Failed { //对于Phase等于Failed，标记startTime
					runtimeStatus.FinishAt = &startTime //任务在执行的时候就发送错误，那么开始和结束时间都标记为同一时刻
					actionIsFailed = true
					groupIsFailed = true
					taskIsFailed = true
				} else {
					if processId != "" { // 细粒度任务Init已经初始化了，恢复运行的时候，这里没有传入进程id（在Init阶段传的），判断为空的话说明这个processId是有值的，不做覆盖
						runtimeStatus.ProcessId = &processId
					}
				}
				runtimeStatus.StartAt = &startTime
				runtimeStatus.Phase = phase // 这里的Phase有可能是Failed，也有可能是running
				runtimeStatus.LastTime = &lastTime
				// 设置副本runtime的状态为Phase（running or failed），如果是跨域的话，这里估计还得再修改
				gmo.updateCopyIngfoForRuntime(getGroup, phase, actionSpecName, runtimeSpecName)
				// 更新一下etcd当中的Runtime的Status
				err = gmo.UpdateRuntimeStatus(runtime.Namespace, runtimeStatus, runtime.Name)
				if err != nil {
					logs.Errorf("Update runtime status error:%v", err)
				}
				logs.Trace("=======================================================3")
			}
		}
		if actionIsFailed {
			actionStatus.Phase = apis.Failed
			actionStatus.LastTime = &lastTime
			actionStatus.FinishAt = &startTime
		}
		// TODO 待解决 有Init--时间的问题
		if actionStatus.Phase == apis.DeployCheck || actionStatus.Phase == apis.Init { // 说明action的刚从DeployCheck(Init)切换到启动状态，需要更改状态
			logs.Trace("=======================================================4")
			if phase == apis.Failed {
				actionStatus.FinishAt = &startTime //任务在执行的时候就发送错误，那么开始和结束时间都标记为同一时刻
			}
			if phase == apis.Running {
				actionStatus.StartAt = &startTime
			}
			actionStatus.Phase = phase
			actionStatus.LastTime = &lastTime
			// 当前group有副本，那么需要将该任务对应的副本任务的action的开始状态也设置一下
			gmo.updateCopyIngfoForAction(getGroup, phase, actionSpecName)

			// 发送Action启动的事件
			gmo.recorder.Event(action, apis.EventTypeNormal, events.StartedCommand, fmt.Sprintf("Action Name:\t %s is Running", action.Name))
		}
		// 更新一下etcd当中的action的Status
		err = gmo.UpdateActionStatus(action.Namespace, actionStatus, action.Name)
		if err != nil {
			logs.Errorf("Update action status failed,err:%v", err)
		}
	}
	if groupIsFailed {
		groupStatus.LastTime = &startTime
		groupStatus.FinishAt = &startTime
		groupStatus.Phase = apis.Failed
	}
	// TODO 待解决 有Init--时间的问题
	//修改group下面的groupStatus下面的ActionStatus，ActionStatus下面的RuntimeStatus
	if groupStatus.Phase == apis.DeployCheck || groupStatus.Phase == apis.Init { // 说明Group刚从DeployCheck(Init)状态边为执行状态
		if phase == apis.Failed {
			groupStatus.FinishAt = &startTime //任务在执行的时候就发送错误，那么开始和结束时间都标记为同一时刻
		}
		if phase == apis.Running {
			groupStatus.StartAt = &startTime
		}
		groupStatus.Phase = phase //说明当前处理的任务是在第一个action当中，所以Group的Phase也得设置为phase（running or Failed）
		groupStatus.LastTime = &lastTime
		groupIsFirstSet = true
	}

	// Group信息更新完毕，考虑该Group是否是第一次启动Running，如果是，就发送Running事件
	if groupIsFirstSet {
		// 发送group启动的事件
		gmo.recorder.Event(getGroup, apis.EventTypeNormal, events.StartedCommand, fmt.Sprintf("Group Name:\t %s is Running", getGroup.Name))
	}
	// 修改Task信息
	if !isCopyGroup {
		if taskIsFailed {
			task.Status.FinishAt = &startTime
			task.Status.LastTime = &startTime
			task.Status.Phase = apis.Failed
		}
		// 更新一下etcd当中的task的Status
		err := gmo.updateTaskStatus(task.Namespace, &task.Status, task.Name)
		if err != nil {
			logs.Errorf("Update task failed-111,err:%v", err)
		}

		if taskIsFirstSet { // task是头一次Running启动
			// 发送Task启动的事件,这里有个小bug就是这时候上传的task的状态里面有些参数还没有更新 TODO 后期如果有硬性要求，这里可以放到最后修改完Task的状态后，才发送Task事件，就一个参数TaskIsfirstSet
			gmo.recorder.Event(task, apis.EventTypeNormal, events.StartedCommand, fmt.Sprintf("Task Name:\t %s is Running", task.Name))
		}
	}
	// 更新一下etcd当中的Group的Status
	err = gmo.UpdateGroupStatus(getGroup.Namespace, groupStatus, getGroup.Name)
	if err != nil {
		logs.Errorf("Update group failed-111,err:%v", err)
	}

	//测试
	groupStatusPhase := getGroup.Status.Phase
	actionStatusPhase := action.Status.Phase
	runtimeStatusPhase := runtime.Status.Phase
	if !isCopyGroup {
		taskPhase := task.Status.Phase
		logs.Infof("START ：taskStatus:%v,groupStatus:%v, action status: %v, runtime Status: %v", taskPhase, groupStatusPhase, actionStatusPhase, runtimeStatusPhase)
	} else {
		logs.Infof("START ：groupStatus:%v, action status: %v, runtime Status: %v", groupStatusPhase, actionStatusPhase, runtimeStatusPhase)
	}
}

// 注意：新增逻辑：如果是副本任务，那么这块对于Task状态的修改，直接跳过
// EndUpdate方法：一个runtime的状态为Failed，则上层Action的状态为Failed，如果说一个runtime的状态为Successed，则上层的的Action状态还不一定是Successed
// TODO 有一个问题，比如一个group的Phase为Migrated，还得查group的副本的状态是否为succeed（目前只适配了本域迁移）
// 收到的Phase为：Successed or Failed or Unknown（Failed、Migrated）,Killed(新增)
// 处理Runtime运行时结束,如果runtime是最后一个执行完成的，还得同时标记action的phase   总结：所有临时变量赋值时都得使用&
func (gmo *GroupMonitor) handleRuntimeEndUpdate(event events.RuntimeEndPhaseEvent1) {
	logs.Info("Handling runtime end status update")
	groupName := event.GroupName
	get, err := gmo.clientsManager.GetGroup(groupName, event.GroupNamespace)
	if err != nil {
		logs.Errorf("Failed get group:%v from etcd, err:%v", groupName, err)
	}
	groupSpec := &get.Spec
	groupStatus := &get.Status

	actionSpecName := event.ActionSpecName
	runtimeSpecName := event.RuntimeSpecName
	phase := event.Phase //当前phase可能为Succeed、Failed、Unknown（Failed、Migrated）
	logs.Infof("Get runtime finish event notify, the phase:%s", phase)
	finshTime := event.FinishAt
	lastTime := event.LastTime
	// Task 信息
	taskName := get.Status.Belong.Name

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
	var task *apis.Task
	var action *apis.Action
	var runtime *apis.Runtime
	for aSpecName, actionReference := range groupStatus.Actions { // 预先锁定action和Runtime
		if aSpecName != actionSpecName { // 判断是否是当前处理的Action
			continue // 不是的话跳过
		}
		action, err = gmo.clientsManager.GetAction(actionReference.Name, actionReference.Namespace)
		if err != nil {
			logs.Errorf("Get action error from etcd:%v", err)
		}
		actionStatus := &action.Status //ActionStatus
		logs.Trace("=======================================================0")
		for rSpecName, runtimeReference := range actionStatus.Runtimes { //RuntimeStatus
			if rSpecName == runtimeSpecName {
				runtime, err = gmo.clientsManager.GetRuntime(runtimeReference.Name, runtimeReference.Namespace)
				if err != nil {
					logs.Errorf("Get runtime error to etcd-11:%v", err)
				}
			}
		}
	}
	if phase == apis.Unknown { // 这里有三种情况，①主动关闭，则为Killed  ②主动迁移关闭，为Migrated  ③副本任务的runtime，Init初始化了，但是没有迁移过来，最终源任务完成，这里要将init的副本runtime进行关闭，为Succeed状态
		if get.Status.Phase == apis.Migrating {
			// 判断为Migrated的情况 查group.Status.Phase，如果为Migrating则为迁移，否则为用户主动关闭任务的操作
			gmo.handleRuntimeMigratedUpdate(get, action, runtime) //对于②情况
			return
		}
		if runtime.Status.Phase == apis.Init { // 对应③情况,TODO 如果是pod任务的话，那么这里runtime的状态是Running，这块该如何解决呢，哦不对，得去看一下k8s的pod监控，会发送Unknown不
			phase = apis.Successed
		} else {
			//为用户主动关闭，对应①情况
			phase = apis.Killed
		}
	}
	// 如果当前Group不是副本Group，需要修改其所属的Task的信息
	if !isCopyGroup {
		task, err = gmo.clientsManager.GetTask(taskName, get.Status.Belong.Namespace)
		if err != nil {
			logs.Errorf("Get task error from etcd-22:%v", err)
		}
		// 检查其他的group是否完成,修改Task的状态 ---需要适配迁移（目前只适配了本域迁移）
		for gSpecName, groupReference := range task.Status.Groups {
			allgroup, err := gmo.clientsManager.GetGroup(groupReference.Name, groupReference.Namespace)
			if err != nil {
				logs.Errorf("Get group error from etcd-221:%v", err)
			}
			grStatus := &allgroup.Status     // for 循环遍历到的Group
			if gSpecName != groupSpec.Name { //遍历到的group的Spec.Name不等于当前处理的Group的Spec.Name,即遍历兄弟group
				//logs.Infof("groupStatus.Phase:%v,groupID:%v", grStatus.Phase, grStatus.GroupID)
				if grStatus.Phase == apis.DeployCheck || grStatus.Phase == apis.Migrating || grStatus.Phase == apis.Running || grStatus.Phase == apis.Init { // 说明其他Group还未执行或者还没迁移成功
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
					if grStatus.CopyStatus == "" { // 副本任务只有Succeed和Failed才会修改源任务的copyStatus，如果说是空，说明副本任务还在运行
						otherGroupCompleted = false
					}
					if grStatus.CopyStatus == "Failed" {
						finalTaskIsFailed = true
					}
				}
				continue
			}
		}
	}

	// 修改groupStatus的状态
	for aSpecName, actionReference := range groupStatus.Actions { //Action
		var finalActionIsFailed = false //标记group里面当前遍历到的Action地下的runtime是否有Failed状态
		var finalActionIsKilled = false //标价group里面当前遍历到的Action底下的runtime是否有Killed状态
		var allRuntiemCompleted = true  // 当前action是否已经完成（只有action下面的所有的runtime都执行完成了，也就是最后一个runtime被执行完成了，要标记action的状态为succeed，如果说action下面的某一个runtime执行失败，则要标记action装填为Failed）
		allAction, err := gmo.clientsManager.GetAction(actionReference.Name, actionReference.Namespace)
		if err != nil {
			logs.Errorf("Get action error from etcd:%v", err)
		}
		actionStatus := &allAction.Status //ActionStatus
		if aSpecName != actionSpecName {  //遍历到其他Action，可以顺带看一下别的Action是否都已经完成了
			if actionStatus.Phase == apis.DeployCheck || actionStatus.Phase == apis.Running || actionStatus.Phase == apis.Init { //其他Action为checking状态，说明还有其他的Action没有被遍历到，Group状态为Running状态
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
		for rSpecName, runtimeReference := range actionStatus.Runtimes { //RuntimeStatus
			allRuntime, err := gmo.clientsManager.GetRuntime(runtimeReference.Name, runtimeReference.Namespace)
			if err != nil {
				logs.Errorf("Get runtime error to etcd-13:%v", err)
			}
			runtimeStatus := &allRuntime.Status
			if rSpecName == runtimeSpecName {
				runtimeStatus.Phase = phase
				runtime.Status.Phase = phase // 方便最终End 显示状态
				runtimeStatus.FinishAt = &finshTime
				runtimeStatus.LastTime = &lastTime
				// 如果说该group有副本，并且该副本group是提前部署副本的，那么这里除了修改源runtime的状态，还得修改副本runtime的状态
				gmo.updateCopyIngfoForRuntime(get, phase, actionSpecName, runtimeSpecName)
				// 更新一下etcd当中的Runtime的Status
				err := gmo.UpdateRuntimeStatus(allRuntime.Namespace, runtimeStatus, allRuntime.Name)
				if err != nil {
					logs.Errorf("Update runtime error to etcd-222:%v", err)
				}
			}
			if runtimeStatus.Phase == apis.DeployCheck || runtimeStatus.Phase == apis.Running || runtimeStatus.Phase == apis.Init { // 遍历所有的Runtime，如果其中一个Runtime状态没有执行完成，说明Action最终不用更新
				allRuntiemCompleted = false
			}
			if runtimeStatus.Phase == apis.Failed {
				finalActionIsFailed = true
				finalGroupIsFailed = true
				finalTaskIsFailed = true
				logs.Infof("+++++++++++++finalActionIsFailed:%v", finalActionIsFailed)
			}
			if runtimeStatus.Phase == apis.Killed {
				finalActionIsKilled = true
				finalGroupIsKilled = true
				finalTaskIsKilled = true
			}
		}
		// 得加一个逻辑：如果Action下有一个Runtime执行Failed或Killed，在这里得检查一下
		if finalActionIsFailed {
			actionStatus.FinishAt = &finshTime
			actionStatus.LastTime = &lastTime
			actionStatus.Phase = apis.Failed
			action.Status.Phase = apis.Failed // 方便最终End 显示状态
			gmo.recorder.Event(action, apis.EventTypeWarning, events.ExecuteFailed, fmt.Sprintf("Action Name:\t %s is Failed", action.Name))
			gmo.updateCopyIngfoForAction(get, phase, actionSpecName)
		}
		if finalActionIsKilled {
			actionStatus.FinishAt = &finshTime
			actionStatus.LastTime = &lastTime
			actionStatus.Phase = apis.Killed
			action.Status.Phase = apis.Killed // 方便最终End 显示状态
			gmo.recorder.Event(action, apis.EventTypeNormal, events.KillingCommand, fmt.Sprintf("Action Name:\t %s is Killed", action.Name))
			gmo.updateCopyIngfoForAction(get, phase, actionSpecName)
		}
		if allRuntiemCompleted { //如果说ActionStatus下面的RuntimeStatus都被执行了，还得修改ActionStatus的phase状态
			//后续可能还要补充:Results
			//actionStatus.Results = results
			actionStatus.FinishAt = &finshTime
			actionStatus.LastTime = &lastTime
			// 还得发送Action执行完成的事件，同时将etcd当中的Action资源状态进行修改
			if !finalActionIsKilled && !finalActionIsFailed {
				actionStatus.Phase = phase
				action.Status.Phase = phase // 方便最终End 显示状态
				gmo.recorder.Event(action, apis.EventTypeNormal, events.ExecuteSuccessfully, fmt.Sprintf("Action Name:\t %s is Successed", action.Name))
			}
			nowActionCompleted = true //当前Action已经完成
			gmo.updateCopyIngfoForAction(get, phase, actionSpecName)
		}
		// 更新一下etcd当中的Action的Status
		err = gmo.UpdateActionStatus(allAction.Namespace, actionStatus, allAction.Name) //这里======================
		if err != nil {
			logs.Errorf("Update action status err:%v", err)
		}
	}
	// 得加一个逻辑：如果Group下有一个Action执行Failed或Killed，在这里得检查一下
	if finalGroupIsFailed {
		groupStatus.FinishAt = &finshTime
		groupStatus.LastTime = &lastTime
		groupStatus.Phase = apis.Failed
		gmo.recorder.Event(get, apis.EventTypeWarning, events.ExecuteFailed, fmt.Sprintf("Group name:\t %s is failed", get.Name))
	}
	if finalGroupIsKilled {
		groupStatus.FinishAt = &finshTime
		groupStatus.LastTime = &lastTime
		groupStatus.Phase = apis.Killed
		gmo.recorder.Event(get, apis.EventTypeNormal, events.KillingCommand, fmt.Sprintf("Group name:\t %s is killed", get.Name))
	}
	//如果说GroupStatus下面的ActionStatus都被执行了，还得修改GroupStatus的phase的状态
	if otherActionCompleted && nowActionCompleted { //说明其他Action都执行完成，当前Action也执行完成
		groupStatus.FinishAt = &finshTime
		groupStatus.LastTime = &lastTime
		if !finalGroupIsKilled && !finalGroupIsFailed {
			groupStatus.Phase = phase //Group的状态等于当前Action执行完成的状态  Succeed
			gmo.recorder.Event(get, apis.EventTypeNormal, events.ExecuteSuccessfully, fmt.Sprintf("Group name:\t %s is successed", get.Name))
		}
		nowGroupCompleted = true
	}

	// 更新Group资源信息，同时还需要判断该group是否执行完成，需要发送事件
	// 修改Update更新为Patch
	// 更新一下etcd当中的Group的Status
	err = gmo.UpdateGroupStatus(get.Namespace, groupStatus, get.Name)
	if err != nil {
		logs.Errorf("Update group err-222:%v", err)
	}

	//如果说TaskStatus下面的Group都被执行了，还得修改TaskStatus的phase的状态
	//logs.Infof("otherGroupCompleted:%v", otherGroupCompleted)
	//logs.Infof("nowGroupCompleted:%v", nowGroupCompleted)
	if !isCopyGroup {
		// 将修改后的Group状态值赋值给Task
		if finalTaskIsFailed {
			task.Status.FinishAt = &finshTime
			task.Status.LastTime = &lastTime
			task.Status.Phase = apis.Failed
			gmo.recorder.Event(task, apis.EventTypeWarning, events.ExecuteFailed, fmt.Sprintf("Task Name:\t %s is Failed", task.Name))
		}
		if finalTaskIsKilled {
			task.Status.FinishAt = &finshTime
			task.Status.LastTime = &lastTime
			task.Status.Phase = apis.Killed
			gmo.recorder.Event(task, apis.EventTypeNormal, events.KillingCommand, fmt.Sprintf("Task Name:\t %s is Killed", task.Name))
		}
		if otherGroupCompleted && nowGroupCompleted {
			task.Status.FinishAt = &finshTime
			task.Status.LastTime = &lastTime
			if !finalTaskIsFailed && !finalTaskIsKilled {
				task.Status.Phase = phase //表示的是Task下的其他Group都是Successed状态，那么Task的状态取决于当前的Group，如果为Succeed，则Task也为Succeed，反正为Failed
				gmo.recorder.Event(task, apis.EventTypeNormal, events.ExecuteSuccessfully, fmt.Sprintf("Task Name:\t %s is Successed", task.Name))
			}
		}
		// 更新一下etcd当中的Task的Status
		err := gmo.updateTaskStatus(task.Namespace, &task.Status, task.Name)
		if err != nil {
			logs.Errorf("Update task err-333:%v", err)
		}
	}
	logs.Infof("runtime.Spec.Name:%v,runtime.Spec.Type:%v", runtime.Spec.Name, runtime.Spec.Type)
	if runtime.Spec.Type == apis.ByPod { //是pod类型的任务
		logs.Infof("=============删除==============pod、service,runtime.Name:%v", runtime.Name)
		err := gmo.runtimeManager.Kill(get, action, runtime, actionSpecName, runtimeSpecName)
		if err != nil {
			logs.Errorf("Stop runtime error:%v", err)
		}
	}
	//测试：
	GroupStatusPhase := get.Status.Phase
	actionStatusPhase := action.Status.Phase
	runtimeStatusPhase := runtime.Status.Phase
	if !isCopyGroup {
		taskPhase := task.Status.Phase
		logs.Infof("Runtime END: taskStatus:%v, groupStatus:%v, actionStatus:%v, runtimeStatus:%v", taskPhase, GroupStatusPhase, actionStatusPhase, runtimeStatusPhase)
	} else {
		logs.Infof("Runtime END: groupStatus:%v, actionStatus:%v, runtimeStatus:%v", GroupStatusPhase, actionStatusPhase, runtimeStatusPhase)
	}
}
func (gmo *GroupMonitor) UpdateGroup(group *apis.Group, groupStatus *apis.GroupStatus, groupName string) error {
	// 修改Group的update改为Patch
	groupSpecPatch, err := json.Marshal(map[string]interface{}{
		"spec": &group.Spec,
	})
	groupStatusPatch, err2 := json.Marshal(map[string]interface{}{
		"status": groupStatus,
	})
	if err != nil {
		logs.Errorf("Marshal groupSpec failed,err:%v", err)
		return fmt.Errorf("Marshal groupSpec failed, err:%v", err)
	}
	if err2 != nil {
		logs.Errorf("Marshal groupStatus failed,err:%v", err)
		return fmt.Errorf("Marshal groupStatus failed, err:%v", err)
	}

	_, err = gmo.clientsManager.PatchGroup(groupName, group.Namespace, groupSpecPatch)
	if err != nil {
		logs.Errorf("Patch group failed-111,err:%v", err)
		return fmt.Errorf("Patch groupSpec failed, err:%v", err)
	}
	_, err2 = gmo.clientsManager.PatchGroup(groupName, group.Namespace, groupStatusPatch)
	if err2 != nil {
		logs.Errorf("Patch group failed-222,err:%v", err)
		return fmt.Errorf("Patch groupStatus failed, err:%v", err)
	}
	return nil
}
func (gmo *GroupMonitor) UpdateTask(task *apis.Task, taskStatus *apis.TaskStatus, taskName string) error {
	// 修改Task的update改为Patch
	taskSpecPatch, err := json.Marshal(map[string]interface{}{
		"spec": &task.Spec,
	})
	taskStatusPatch, err2 := json.Marshal(map[string]interface{}{
		"status": taskStatus,
	})
	_, err = gmo.clientsManager.PatchTask(taskName, task.Namespace, taskSpecPatch)
	if err != nil {
		logs.Errorf("Patch task failed-333,err:%v", err)
		return fmt.Errorf("Patch taskSpec failed, err:%v", err)
	}
	_, err2 = gmo.clientsManager.PatchTask(taskName, task.Namespace, taskStatusPatch)
	if err2 != nil {
		logs.Errorf("Patch task failed-444,err:%v", err)
		return fmt.Errorf("Patch taskStatus failed, err:%v", err)
	}
	return nil
}
func (gmo *GroupMonitor) updateTaskStatus(taskNamespace string, taskStatus *apis.TaskStatus, taskName string) error {
	taskStatusPatch, err := json.Marshal(map[string]interface{}{
		"status": taskStatus,
	})
	if err != nil {
		logs.Errorf("Marshal task filed,err:%v", err)
		return fmt.Errorf("Marshal taskStatus failed, err:%v", err)
	}
	_, err2 := gmo.clientsManager.PatchTask(taskName, taskNamespace, taskStatusPatch)
	if err2 != nil {
		logs.Errorf("Patch task failed-444,err:%v", err)
		return fmt.Errorf("Patch taskStatus failed, err:%v", err)
	}
	return nil
}
func (gmo *GroupMonitor) UpdateGroupStatus(groupNamespace string, groupStatus *apis.GroupStatus, groupName string) error {
	groupStatusPatch, err := json.Marshal(map[string]interface{}{
		"status": groupStatus,
	})
	if err != nil {
		logs.Errorf("Marshal group filed,err:%v", err)
		return fmt.Errorf("Marshal groupStatus filed, err:%v", err)
	}
	_, err2 := gmo.clientsManager.PatchGroup(groupName, groupNamespace, groupStatusPatch)
	if err2 != nil {
		logs.Errorf("Patch group failed-555,err:%v", err)
		return fmt.Errorf("Patch groupStatus filed, err:%v", err)
	}
	return nil
}
func (gmo *GroupMonitor) UpdateActionStatus(actionNamespace string, actionStatus *apis.ActionStatus, actionName string) error {
	actionStatusPatch, err := json.Marshal(map[string]interface{}{
		"status": actionStatus,
	})
	if err != nil {
		logs.Errorf("Marshal action filed,err:%v", err)
		return fmt.Errorf("Marshal action filed, err:%v", err)
	}
	_, err = gmo.clientsManager.PatchAction(actionName, actionNamespace, actionStatusPatch)
	if err != nil {
		logs.Errorf("Patch action failed-666,err:%v", err)
		return fmt.Errorf("Patch action failed, err:%v", err)
	}
	return nil
}
func (gmo *GroupMonitor) UpdateRuntimeStatus(runtimeNamespace string, runtimeStatus *apis.RuntimeStatus, runtimeName string) error {
	runtimeStatusPatch, err := json.Marshal(map[string]interface{}{
		"status": runtimeStatus,
	})
	if err != nil {
		logs.Errorf("Marshal runtime filed,err:%v", err)
		return fmt.Errorf("Marshal runtime filed, err:%v", err)
	}

	_, err = gmo.clientsManager.PatchRuntime(runtimeName, runtimeNamespace, runtimeStatusPatch)
	if err != nil {
		logs.Errorf("Patch runtime failed-111,err:%v", err)
		return fmt.Errorf("Patch runtime failed-111,err:%v", err)
	}
	return nil
}

// 同group_workers当中的方法
func (gmo *GroupMonitor) groupDepenSatisfy(group *apis.Group, task *apis.Task) bool {
	// TODO：实现依赖检查逻辑
	// 目前只是检查Spec当中的Parents选项
	if group.Spec.Conditions == nil {
		return true
	}
	if len(group.Spec.Conditions.Formulas) == 0 {
		return true
	}
	if len(group.Spec.Parents) == 0 {
		// 没有父节点，直接去检查后续的依赖
	} else {
		for _, parentName := range group.Spec.Parents {
			parentGroupRef := task.Status.Groups[parentName]
			parentGroup, err := gmo.clientsManager.GetGroup(parentGroupRef.Name, parentGroupRef.Namespace)
			if err != nil {
				logs.Errorf("Failed to get parent group:%v form etcd, err:%v", parentName, err)
			}
			if parentGroup.Status.Phase != apis.Successed {
				logs.Trace("group %v's parent group:%v is not completed", group.Name, parentName)
				return false
			}
		}
	}
	// 其他依赖
	res, err := gmo.conditionEngine.CheckConditions(group.Spec.Conditions, group)
	if err != nil {
		logs.Errorf("Check group conditions error:%v", err)
		return false
	}
	switch res {
	case apis.True:
		logs.Infof("group %v's conditions is satisfy", group.Name)
		return true
	case apis.False:
		logs.Tracef("group %v's conditions is not satisfy", group.Name)
		return false
	case apis.NotReady:
		logs.Tracef("group %v's conditions is not ready", group.Name)
		return false
	default:
		return false
	}
}

// 检查Action的依赖是否满足
func (gmo *GroupMonitor) actionDepenSatisfy(action *apis.Action, group *apis.Group) bool {
	actionSpec := &action.Spec
	//检查conditions是否是空指针
	if actionSpec.Conditions == nil {
		return true
	}
	if len(actionSpec.Conditions.Formulas) == 0 {
		return true
	}
	if len(action.Spec.Parents) == 0 {
		// 没有父节点，直接去检查后续的依赖
	} else {
		for _, parentName := range actionSpec.Parents {
			parentActionRef := group.Status.Actions[parentName]
			parentAction, err := gmo.clientsManager.GetAction(parentActionRef.Name, parentActionRef.Namespace)
			if err != nil {
				logs.Errorf("Failed to get parent action:%v form etcd, err:%v", parentName, err)
			}
			if parentAction.Status.Phase != apis.Successed {
				logs.Trace("action %v's parent action:%v is not completed", action.Name, parentName)
				return false
			}
		}
	}
	// 其他依赖
	res, err := gmo.conditionEngine.CheckConditions(actionSpec.Conditions, action)
	if err != nil {
		logs.Errorf("Check action conditions error:%v", err)
		return false
	}
	switch res {
	case apis.True:
		logs.Infof("action %v's conditions is satisfy", action.Name)
		return true
	case apis.False:
		logs.Tracef("action %v's conditions is not satisfy", action.Name)
		return false
	case apis.NotReady:
		logs.Tracef("action %v's conditions is not ready", action.Name)
		return false
	default:
		return false
	}
}

// 检查Runtime的依赖是否满足
func (gmo *GroupMonitor) runtimeDepenSatisfy(runtime *apis.Runtime, action *apis.Action) bool {
	//TODO runtime运行之前，需要检查parent的runtime是否正常执行完成
	runtimeStatus := &runtime.Status
	if runtimeStatus.IsDependencySatisf {
		return true
	}
	if runtime.Spec.Conditions == nil {
		return true
	}
	if len(runtime.Spec.Conditions.Formulas) == 0 {
		return true
	}
	
	// 使用Parents检查Runtime的顺序依赖，如果不满足则返回false
	if len(runtime.Spec.Parents) == 0 {
		// 没有父节点，直接去检查后续的依赖
	} else {
		for _, parentName := range runtime.Spec.Parents {
			parentRuntimeRef := action.Status.Runtimes[parentName]
			parentRuntime, err := gmo.clientsManager.GetRuntime(parentRuntimeRef.Name, parentRuntimeRef.Namespace)
			if err != nil {
				logs.Errorf("Failed to get parent runtime:%v form etcd, err:%v", parentName, err)
			}
			if parentRuntime.Status.Phase != apis.Successed {
				logs.Trace("runtime %v's parent:%v is not completed", runtime.Name, parentName)
				return false
			}
		}
	}

	// 其他依赖
	for _, i := range runtime.Spec.Conditions.Formulas {
		// ProgramDependency暂时不方便直接使用ConditionEngine
		if i.LeftValue.Name == string(apis.ProgramDependency) {
			//runtime运行之前,需要检查程序依赖是不是满足，如果满足则将符合条件的环境变量加入runtime的Env中，方便后续CMD注入环境变量；
			//如果不满足则返回false，开启CMD创建新的程序依赖，等待monitor检查到依赖满足才拉起这个runtime
			//TODO：后续和上面的condition合并进一起，可能是以单独写一个condition函数的形式，然后这里只需要调用统一的condition检查函数即可
			
			// 首先判断这个程序依赖的condition是否已经是满足的，如果是满足的直接跳过
			if i.Result == apis.True {
				// logs.Tracef("runtime %v's program dependency is satisfy", runtime.Name)
				continue
			}

			var dependencyFile = i.LeftValue.From
			if !runtimeStatus.IsParsed {
				// runtimeReqPackages := make([]apis.Requirement, 0)
				runtimeReqPackages, err := gmo.dependencyManager.ParseRequirements(dependencyFile)
				if err != nil {
					logs.Info("parse Runtime Requirements error!")
					runtimeStatus.IsParsed = false
					return false
				}
				runtimeStatus.IsParsed = true
				runtime.Spec.Packages = runtimeReqPackages
				// 这里需要将Package写到etcd上去
				// 为runtime添加一个依赖是否已经满足的参数，如果已经满足的话则标记为true
				//patchGroup, err := json.Marshal([]map[string]interface{}{
				//	{
				//		"op":    "replace",
				//		"path":  "/status/action_status/" + strconv.Itoa(actionIndex) + "/status/" + strconv.Itoa(runtimeIndex) + "/isparsed",
				//		"value": true,
				//	},
				//})
				//if err != nil {
				//	logs.Errorf("Marshal patch group err:%v", err)
				//}
				//_, err = gmo.groupClient.Patch(context.TODO(), group.Name, types.JSONPatchType, patchGroup, metav1.PatchOptions{})
				//if err != nil {
				//	logs.Errorf("Patch group err-32:%v", err)
				//}
				//patchGroup2, err := json.Marshal([]map[string]interface{}{
				//	{
				//		"op":    "replace",
				//		"path":  "/spec/actions/" + strconv.Itoa(actionIndex) + "/spec/runtimes/" + strconv.Itoa(runtimeIndex) + "package",
				//		"value": runtimeReqPackages,
				//	},
				//})
				//if err != nil {
				//	logs.Errorf("Marshal patch group err:%v", err)
				//}
				//_, err = gmo.groupClient.Patch(context.TODO(), group.Name, types.JSONPatchType, patchGroup2, metav1.PatchOptions{})
				//if err != nil {
				//	logs.Errorf("Patch group err-33:%v", err)
				//}
			}
			envName, envPath, err := gmo.dependencyManager.CheckEnvironmentSatisfy(runtime.Spec.Packages)
			if !err {
				// logs.Info("dependency do not satisfy,runtime name:%v", runtime.Name)
				//需要使用协程，但是还要防止在monitor监控的时候多次创建
				if !runtimeStatus.DepenPreparing {
					logs.Trace("installing dependency for runtime name:%v", runtime.Name)
					runtimeStatus.DepenPreparing = true
					// go dependency.SetupEnvironment(dependencyFile, runtime.Name)
				} else {
					logs.Trace("runtime %v is waiting for installing dependency!", runtime.Name)
				}
				logs.Infof("programDependency err, runtime:%v get envName:%v", runtime.Name, envName)
				return false
			}

			i.Result = apis.True

			envVar := []apis.EnvVar{}
			envVar = append(envVar, apis.EnvVar{Name: "PATH", Value: envPath})
			runtime.Spec.EnvVar = append(runtime.Spec.EnvVar, envVar...)
		}
	}
	if !runtimeStatus.IsDependencySatisf {
		// 为runtime添加一个依赖是否已经满足的参数，如果已经满足的话则标记为true
		patchRuntime, err := json.Marshal(map[string]interface{}{
			"staus": map[string]interface{}{
				"dependency_satisf": true,
			},
		})
		if err != nil {
			logs.Errorf("Marshal patch runtime err:%v", err)
		}
		_, err = gmo.clientsManager.PatchRuntime(runtime.Name, runtime.Namespace, patchRuntime)
		if err != nil {
			logs.Errorf("Patch runtime err-23:%v", err)
		}
		logs.Trace("=================set  rtStatus.IsDependencySatisf  = true")
	}
	return true
}

// 主要处理group一直未完成的情况，这里主要讲group的状态设置为Failed，Task状态设置为Failed，group下面的Action、Runtime状态均为Unknown
func (gmo *GroupMonitor) handleStatusUpdate(group *apis.Group, failed apis.Phase) {
	// 设置当前group的状态为Failed，同时去找Task，标记其状态也为Failed
	time := apis.Time{time.Now()}
	group.Status.Phase = failed
	group.Status.StartAt = &time
	group.Status.FinishAt = &time
	group.Status.LastTime = &time
	// 更新一下etcd当中的Group的Status
	err := gmo.UpdateGroupStatus(group.Namespace, &group.Status, group.Name)
	if err != nil {
		logs.Errorf("Update group status err-555:%v", err)
	}
	taskName := group.Status.Belong.Name // 当前所属的Task的Name(全局唯一)
	// 获取这个Task
	task, err := gmo.clientsManager.GetTask(taskName, group.Status.Belong.Namespace)
	if err != nil {
		logs.Errorf("Get task from etcd err:%v", err)
	}
	task.Status.Phase = failed
	task.Status.StartAt = &time
	task.Status.FinishAt = &time
	task.Status.LastTime = &time
	// 更新一下etcd当中的Task的Status
	err = gmo.updateTaskStatus(task.Namespace, &task.Status, task.Name)
	if err != nil {
		logs.Errorf("Update task status err-111:%v", err)
	}
}

// 处理runtime是需要迁移的情况，就是将runtime的Running状态改为Migrated，如果为Succeed，这个Phase不修改  最后都修改完后将group修改为Migrated
func (gmo *GroupMonitor) handleRuntimeMigratedUpdate(group *apis.Group, action *apis.Action, runtime *apis.Runtime) {
	// 将runtime的Phase从Running修改为Migrated
	nowTime := apis.Time{time.Now()}
	groupStatus := &group.Status

	// GroupStatus下面的ActionStatus、RuntimeStatus
	action.Status.Phase = apis.Migrated
	action.Status.LastTime = &nowTime
	action.Status.FinishAt = &nowTime
	runtime.Status.Phase = apis.Migrated
	runtime.Status.LastTime = &nowTime
	runtime.Status.FinishAt = &nowTime

	// 检查group下的信息是否都修改完成
	var IsModify = true
	for aSepecName, actionReference := range groupStatus.Actions {
		if aSepecName == action.Spec.Name { // 得将当前遍历到的Action排除
			continue
		}
		action, err := gmo.clientsManager.GetAction(actionReference.Name, actionReference.Namespace)
		if err != nil {
			logs.Errorf("Get action from etcd err:%v", err)
		}

		if action.Status.Phase == apis.Running {
			IsModify = false
		}
	}
	logs.Infof("------------------------------------------IsModify:%v", IsModify)
	if IsModify { // group下面的信息都修改完成，接下来需要将一些DeployCheck的状态进行修改。group修改完后，还需要将group的信息放到Task当中，如果Task下面的所有group都执行成功且都放到了Task下面的话，最后还需要做收尾工作，标记Task状态为Running，当迁移完成后，由Migrated队列轮询检查副本的执行情况，再标记Task的状态为成功或失败
		logs.Info("---------------进入ISModify----------------------------------------")
		//统一将DeployCheck的Phase都改成Migrate 首先是GroupStatus，如果是成功的Phase就不修改了，其他的都修改，包括running、DeployCheck
		for _, actionReference := range groupStatus.Actions {
			allAction, err := gmo.clientsManager.GetAction(actionReference.Name, actionReference.Namespace)
			if err != nil {
				logs.Errorf("Get action from etcd err:%v", err)
			}
			actionStatus := &allAction.Status
			if actionStatus.Phase == apis.DeployCheck || actionStatus.Phase == apis.Migrated {
				for _, runtimeReference := range actionStatus.Runtimes {
					allRuntime, err := gmo.clientsManager.GetRuntime(runtimeReference.Name, runtimeReference.Namespace)
					if err != nil {
						logs.Errorf("Get runtime from etcd err:%v", err)
					}
					runtimeStatus := &allRuntime.Status
					if runtimeStatus.Phase == apis.Migrated {
						continue
					}
					runtimeStatus.Phase = apis.Migrated
					runtimeStatus.LastTime = &nowTime
					runtimeStatus.FinishAt = &nowTime
				}
				if actionStatus.Phase == apis.Migrated {
					continue
				}
				actionStatus.Phase = apis.Migrated
				actionStatus.LastTime = &nowTime
				actionStatus.FinishAt = &nowTime
			}
		}
		group.Status.Phase = apis.Migrated

		// 这里检查一下其他Group的状态是否为都成功了，如果都成功了，还需要修改Task的信息
		var task *apis.Task
		var err error
		var otherGroupCompleted = true
		var finalTaskIsFailed = false
		taskName := group.Status.Belong.Name
		task, err = gmo.clientsManager.GetTask(taskName, group.Status.Belong.Namespace)
		if err != nil {
			logs.Error("Get task by taskID error from etcd:%v", err)
		}
		// 检查其他的group是否完成,修改Task的状态 ---需要适配迁移（目前只适配了本域迁移）
		for gSpecName, groupReference := range task.Status.Groups {
			if gSpecName != group.Spec.Name { //遍历到的group的Name不等于当前处理的Group的Name
				allGroup, err := gmo.clientsManager.GetGroup(groupReference.Name, groupReference.Namespace)
				if err != nil {
					logs.Errorf("Get group from etcd err:%v", err)
				}
				groupStatus := &allGroup.Status
				//logs.Infof("groupStatus.Phase:%v,groupID:%v", grStatus.Phase, grStatus.GroupID)
				if groupStatus.Phase == apis.DeployCheck || groupStatus.Phase == apis.Migrating || groupStatus.Phase == apis.Init || groupStatus.Phase == apis.Running { // 说明其他Group还未执行或者还没迁移成功
					otherGroupCompleted = false
				}
				if groupStatus.Phase == apis.Failed { // 如果有一个Group的状态为Failed，则Task状态必定为Failed
					finalTaskIsFailed = true
				}
				if groupStatus.Phase == apis.Migrated { // 还得去查对应副本任务的状态，如果状态为Running（大概率是这个状态）或者是DeployChek（说明迁移过去的group依赖不满足，暂时还不能执行），那么otherGroupCompleted参数也是false
					// 为了适配迁移，目前还是处理同域的迁移,这里怎么根据源任务找到副本任务，还是一个遗留的问题
					//copyGroup, err := gmo.groupClient.Get(context.TODO(), "Reason-Copy", metav1.GetOptions{})
					//if err != nil {
					//	logs.Errorf("Get copy group err:%v", err)
					//}
					if groupStatus.CopyStatus == "" { // 说明该Group的副本group正在运行还没结束
						otherGroupCompleted = false
					}
					if groupStatus.CopyStatus == "Failed" {
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
		}
		if otherGroupCompleted {
			if finalTaskIsFailed { // 如果说group当中有Failed状态，那么最终Task也是得被标记为Failed
				task.Status.Phase = apis.Failed
			} else {
				task.Status.Phase = apis.Running //表示的是Task下的其他Group都是Successed状态，那么Task的状态取决于当前的Group，因为当前的Group正在迁移，只有当前group的副本任务完成了，那么到时候在Migrated队列当中，就会修改Task的状态的，这里设置为Running，合理
			}
			task.Status.FinishAt = &nowTime
			task.Status.LastTime = &nowTime
		}
		// 更新一下etcd当中的Task的Status
		err = gmo.updateTaskStatus(task.Namespace, &task.Status, task.Name)
		if err != nil {
			logs.Errorf("Update task err-999:%v", err)
		}

	}
	// 更新一下etcd当中的Group的Status
	err := gmo.UpdateGroupStatus(group.Namespace, groupStatus, group.Name)
	if err != nil {
		logs.Errorf("Update group err-888:%v", err)
	}
	// 更新一下etcd当中的Action的Status
	err = gmo.UpdateActionStatus(action.Namespace, &action.Status, action.Name)
	if err != nil {
		logs.Errorf("Update action err-888:%v", err)
	}
	// 更新一下etcd当中的Runtime的Status
	err = gmo.UpdateRuntimeStatus(runtime.Namespace, &runtime.Status, runtime.Name)
	if err != nil {
		logs.Errorf("Update runtime err-888:%v", err)
	}
}

// 主动将当前指定的runtime状态设置为Successed,如果其兄弟Runtime也都Successed了，那么也需要将Action的状态设置为Succeed---这里调用的情况：对于那些非细粒度控制的runtime，当其源runtime完成后，这里需要修改runtime
func (gmo *GroupMonitor) handleRuntimeSucceedUpdate(action *apis.Action, runtime *apis.Runtime) {
	// 将runtime的Phase从Running修改为Migrated
	// 这里不设置时间，因为这个副本的runtime（粗粒度控制）压根没启动过，只是因为其源runtime执行完成了，所以该runtime也就无需执行了，直接设置Phase即可

	// GroupStatus下面的ActionStatus、RuntimeStatus
	runtime.Status.Phase = apis.Successed
	patchRuntime, err := json.Marshal(map[string]interface{}{
		"status": map[string]interface{}{
			"phase": apis.Successed,
		},
	})
	if err != nil {
		logs.Errorf("Json Marshal failed, err:%v", err)
	}
	_, err = gmo.clientsManager.PatchRuntime(runtime.Name, runtime.Namespace, patchRuntime)
	if err != nil {
		logs.Errorf("Patch runtime error-456:%v", err)
	}
	// 还需要检查兄弟Runtime是否都执行完，如果执行完就，还得修改Action的状态为Successed
	var actionIsSuccess = true
	actionStatus := &action.Status
	for _, runtimeReference := range actionStatus.Runtimes {
		allruntime, err := gmo.clientsManager.GetRuntime(runtimeReference.Name, runtimeReference.Namespace)
		if err != nil {
			logs.Errorf("Get runtime from etcd err:%v", err)
		}
		runtimeStatus := &allruntime.Status
		if runtimeStatus.Phase != apis.Successed {
			actionIsSuccess = false
			break
		}
	}
	if actionIsSuccess { // 说明action下面的所有Runtime都执行完成了，这里要修改Action的状态
		patchAction, err := json.Marshal(map[string]interface{}{
			"status": map[string]interface{}{
				"phase": apis.Successed,
			},
		})
		if err != nil {
			logs.Errorf("Json Marshal failed, err:%v", err)
		}
		_, err = gmo.clientsManager.PatchAction(action.Name, action.Namespace, patchAction)
		if err != nil {
			logs.Errorf("Patch action error-456:%v", err)
		}
	}
}

//// 将副本的Action以及下面的runtime（还没设置为Succeed）的phase设置为Succeed,时间的话全部改成源任务完成的时间吧
//func (gmo *GroupMonitor) handleCopyActionSucceedUpdate(groupCopy *apis.Group, actionIndex int) {
//	groupSpec := &groupCopy.Spec
//	groupStatus := &groupCopy.Status
//	nowTime := apis.Time{time.Now()}
//	// groupSpec下面的Action以及runtime的phase修改为Succeed，并且将所有时间修改成当前的时间
//	groupSpec.Actions[actionIndex].Status.Phase = apis.Successed
//	groupSpec.Actions[actionIndex].Status.LastTime = nowTime
//	groupSpec.Actions[actionIndex].Status.FinishAt = nowTime
//	groupSpec.Actions[actionIndex].Status.StartAt = nowTime
//	for i := range groupSpec.Actions[actionIndex].Status.RuntimeStatus {
//		runtimeStatus := &groupSpec.Actions[i].Status
//		if runtimeStatus.Phase != apis.Successed {
//			runtimeStatus.Phase = apis.Successed
//		}
//		runtimeStatus.StartAt = nowTime
//		runtimeStatus.FinishAt = nowTime
//		runtimeStatus.LastTime = nowTime
//	}
//	// groupStatus下面的Action以及runtime的Phase修改为Succeed
//	groupStatus.ActionStatus[actionIndex].Phase = apis.Successed
//	groupStatus.ActionStatus[actionIndex].LastTime = nowTime
//	groupStatus.ActionStatus[actionIndex].StartAt = nowTime
//	groupStatus.ActionStatus[actionIndex].FinishAt = nowTime
//	for i := range groupStatus.ActionStatus[actionIndex].RuntimeStatus {
//		runtimeStatus := &groupStatus.ActionStatus[actionIndex].RuntimeStatus[i]
//		if runtimeStatus.Phase != apis.Successed {
//			runtimeStatus.Phase = apis.Successed
//		}
//		runtimeStatus.LastTime = nowTime
//		runtimeStatus.FinishAt = nowTime
//		runtimeStatus.StartAt = nowTime
//	}
//	err := gmo.UpdateGroup(groupSpec, groupStatus, groupCopy.Name)
//	if err != nil {
//		logs.Errorf("Update group error-1012:%v", err)
//	}
//	//_, err2 := gmo.groupClient.Update(context.TODO(), groupCopy, metav1.UpdateOptions{})
//	//if err2 != nil {
//	//	logs.Errorf("Update group-runtime-successed info error-9:%v", err2)
//	//}
//}

// 设置group、action、runtime的phase为failed，单单将当前这个Filed的Action，以及下面Failed的Runtime的状态设置为Failed，对于其他的Action、Runtime不做处理，然后同时将Group的状态设置为Failed，完毕
func (gmo *GroupMonitor) handleCopyRuntimeFailedUpdate(groupCopy *apis.Group, action *apis.Action) {
	nowTime := apis.Time{time.Now()}
	groupCopy.Status.Phase = apis.Failed
	groupCopy.Status.LastTime = &nowTime
	groupCopy.Status.FinishAt = &nowTime
	groupCopy.Status.StartAt = &nowTime
	// 更新一下etcd当中的Group的Status
	err := gmo.UpdateGroupStatus(groupCopy.Namespace, &groupCopy.Status, groupCopy.Name)
	if err != nil {
		logs.Errorf("Update group error-2012:%v", err)
	}
	// groupStatus下面的Action以及runtime的Phase修改为Succeed
	action.Status.Phase = apis.Failed
	action.Status.LastTime = &nowTime
	action.Status.FinishAt = &nowTime
	action.Status.StartAt = &nowTime
	// 更新一下etcd当中的Action的Status
	err = gmo.UpdateActionStatus(action.Namespace, &action.Status, action.Name)
	if err != nil {
		logs.Errorf("Update action error-2012:%v", err)
	}

	for _, runtimeReference := range action.Status.Runtimes {
		runtime, err := gmo.clientsManager.GetRuntime(runtimeReference.Name, runtimeReference.Namespace)
		if err != nil {
			logs.Errorf("Get runtime from etcd err:%v", err)
		}
		runtimeStatus := &runtime.Status
		if runtimeStatus.CopyStatus == "Failed" {
			runtimeStatus.Phase = apis.Failed
			runtimeStatus.LastTime = &nowTime
			runtimeStatus.FinishAt = &nowTime
			runtimeStatus.StartAt = &nowTime
			// 更新一下etcd当中的Runtime的Status
			err := gmo.UpdateRuntimeStatus(runtime.Namespace, runtimeStatus, runtime.Name)
			if err != nil {
				logs.Errorf("Update runtime error-2012:%v", err)
			}
		}
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
	_, err = gmo.clientsManager.PatchGroup(gro.Name, gro.Namespace, patchGroup)
	if err != nil {
		logs.Errorf("Patch task err-14:%v", err)
	}
	// 修改Group所属的Task的Phase为Failed
	nowTime := apis.Time{time.Now()}
	taskName := gro.Status.Belong.Name // group的Belongs属性当中的TaskID
	var task *apis.Task
	var err2 error
	task, err2 = gmo.clientsManager.GetTask(taskName, gro.Status.Belong.Namespace)
	if err2 != nil {
		logs.Error("Get task by taskID error from etcd:%v", err2)
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
	_, err = gmo.clientsManager.PatchTask(task.Name, task.Namespace, patchTask)
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
	_, err = gmo.clientsManager.PatchGroup(gro.Name, gro.Namespace, patchGroup)
	if err != nil {
		logs.Errorf("Patch task err-14:%v", err)
	}
	// 修改Group所属的Task的Phase为Succed
	nowTime := apis.Time{time.Now()}
	taskName := gro.Status.Belong.Name // group的Belongs属性当中的TaskID
	var task *apis.Task
	var err2 error
	task, err2 = gmo.clientsManager.GetTask(taskName, gro.Status.Belong.Namespace)
	if err2 != nil {
		logs.Error("Get task by taskID error from etcd:%v", err2)
	}
	if task.Status.Phase == apis.Successed { // 有可能有这种情况，就是有多个迁移的group，可能其中一个已经执行了该方法，将Task的phase改为了Succeed了
		return
	}
	// 主要是遍历当前group的兄弟group，看是否执行完成了，如果没有的话就不啥也不做，让最后一个执行完成的group去修改Task状态（注意：啥也不用做，就是说修改Task状态为Failed，也不用做，其他地方有逻辑会完成的）
	var otherGroupIsSuccessed = true
	for gSpecName, groupReference := range task.Status.Groups {

		if gSpecName != gro.Spec.Name { // 遍历到当前Group的兄弟group，兄弟group只有Succeed、Failed、Migrated状态
			group, err := gmo.clientsManager.GetGroup(groupReference.Name, groupReference.Namespace)
			if err != nil {
				logs.Error("Get group  error from etcd:%v", err2)
			}
			groupStatus := &group.Status

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
	_, err = gmo.clientsManager.PatchTask(taskName, gro.Status.Belong.Namespace, patchTask)
	if err != nil {
		logs.Errorf("Patch task err:%v", err)
	}
}

// 源任务的Runtime已完成，修改副本Group的Runtime的状态，这样副本的Runtime就不用实际去执行，维护这个状态即可
func (gmo *GroupMonitor) updateCopyIngfoForRuntime(group *apis.Group, phase apis.Phase, actionSpecName, runtimeSpecName string) {
	logs.Trace("=======================================================1")
	//logs.Trace("#########runtime#############groupSpec.Replicas:%v,groupSpec.Replicas[0] > 0:%v,groupSpec.Replicas[1] > 0:%v", group.Spec.Replicas, group.Spec.Replicas[0] > 0, group.Spec.Replicas[1] > 0)
	for key, value := range group.Spec.CopyInfo {
		if group.Spec.CopyInfo[key] != "local" { // 副本任务信息存在另一个域当中，为跨域存储
			// TODO 使用跨域组件连接另一个域，修改group信息--暂时这么修改
			//copyGroupName := key // 从copyInfo当中获取到副本的名字Name
			//patchGroup, err := json.Marshal(map[string]interface{}{
			//	"op":    "replace",
			//	"path":  "/status/action_status/" + strconv.Itoa(actionIndex) + "/status/" + strconv.Itoa(runtimeIndex) + "/copy_status",
			//	"value": phase,
			//})
			//if err != nil {
			//	logs.Errorf("Marshal patch group err:%v", err)
			//}
			// 暂时
			copyGroupName := key
			groupTarget := gmo.groupTargets[value]
			actionTarget := gmo.actionTargets[value]   //-=-=-=-=
			runtimeTarget := gmo.runtimeTargets[value] //-=-=-=-=
			getGroup, err := groupTarget.Get(context.TODO(), copyGroupName, metav1.GetOptions{})
			if err != nil {
				logs.Infof("Failed to get group: %v", err)
			}
			for aSpecName, actionReference := range getGroup.Status.Actions {
				if aSpecName == actionSpecName {
					action, err := actionTarget.Get(context.TODO(), actionReference.Name, metav1.GetOptions{})
					if err != nil {
						logs.Infof("Failed to get action: %v", err)
					}
					for rSpecName, runtimeReference := range action.Status.Runtimes {
						if rSpecName == runtimeSpecName {
							patchRuntime, err := json.Marshal(map[string]interface{}{
								"status": map[string]interface{}{
									"copy_status": string(phase),
								},
							})
							//runtime.Status.CopyStatus = string(phase)
							_, err = runtimeTarget.Patch(context.TODO(), runtimeReference.Name, types.StrategicMergePatchType, patchRuntime, metav1.PatchOptions{})
							if err != nil {
								logs.Infof("Failed to Patch runtime: %v", err)
							}
						}
					}
				}

			}
			//getGroup.Status.ActionStatus[actionIndex].RuntimeStatus[runtimeIndex].CopyStatus = string(phase)
			//_, err = groupTarget.Update(context.TODO(), getGroup, metav1.UpdateOptions{})
			//if err != nil {
			//	logs.Errorf("Update group in other domain err:%v", err)
			//}
		} else { // 副本任务信息存在当前域当中
			copyGroupName := key
			getGroup, err := gmo.clientsManager.GetGroup(copyGroupName, group.Namespace) // 副本任务与源任务的namespace一样，所以可以服用
			if err != nil {
				logs.Infof("Failed to get group: %v", err)
			}
			for aSpecName, actionReference := range getGroup.Status.Actions {
				if aSpecName == actionSpecName {
					action, err := gmo.clientsManager.GetAction(actionReference.Name, actionReference.Namespace)
					if err != nil {
						logs.Infof("Failed to get action: %v", err)
					}
					for rSpecName, runtimeReference := range action.Status.Runtimes {
						if rSpecName == runtimeSpecName {
							patchRuntime, err := json.Marshal(map[string]interface{}{
								"status": map[string]interface{}{
									"copy_status": string(phase),
								},
							})
							_, err = gmo.clientsManager.PatchRuntime(runtimeReference.Name, runtimeReference.Namespace, patchRuntime)
							if err != nil {
								logs.Infof("Failed to patch runtime: %v", err)
							}
						}
					}
				}
			}
			//patchGroup, err := json.Marshal([]map[string]interface{}{
			//	{
			//		"op":    "replace",
			//		"path":  "/status/action_status/" + strconv.Itoa(actionIndex) + "/status/" + strconv.Itoa(runtimeIndex) + "/copy_status",
			//		"value": phase,
			//	},
			//})
			//if err != nil {
			//	logs.Errorf("Marshal patch group err:%v", err)
			//}
			//_, err = gmo.groupClient.Patch(context.TODO(), copyGroupName, types.JSONPatchType, patchGroup, metav1.PatchOptions{})
			//if err != nil {
			//	logs.Errorf("Patch group err-8:%v", err)
			//}
			logs.Info("#######################groupSpec.Replicas > 0#########设置副本runtime的状态为running----成功")
		}
	}
}

// 源任务的Action已完成，修改副本Group的Action的状态，这样副本Action就不用实际去执行，维护这个状态即可
func (gmo *GroupMonitor) updateCopyIngfoForAction(group *apis.Group, phase apis.Phase, actionSpecName string) {
	//logs.Trace("Start#########action#############groupSpec.Replicas:%v,groupSpec.Replicas > 0:%v", group.Spec.Replicas[0], group.Spec.Replicas[0] > 0)
	for key, value := range group.Spec.CopyInfo {
		if group.Spec.CopyInfo[key] != "local" { // 任务信息存在别的域当中
			// TODO 使用跨域组件连接另一个域，修改group信息--暂时修改成这样
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

			// 暂时
			copyGroupName := key
			groupTarget := gmo.groupTargets[value]   //-=-=-=-=
			actionTarget := gmo.actionTargets[value] //-=-=-=-=
			getGroup, err := groupTarget.Get(context.TODO(), copyGroupName, metav1.GetOptions{})
			if err != nil {
				logs.Infof("Failed to get group: %v", err)
			}
			for aSpecName, actionReference := range getGroup.Status.Actions {
				if aSpecName == actionSpecName {
					patchAction, err := json.Marshal(map[string]interface{}{
						"status": map[string]interface{}{
							"copy_status": string(phase),
						},
					})
					_, err = actionTarget.Patch(context.TODO(), actionReference.Name, types.StrategicMergePatchType, patchAction, metav1.PatchOptions{})
					if err != nil {
						logs.Infof("Failed to update action: %v", err)
					}
				}
			}
			//getGroup.Status.ActionStatus[actionIndex].CopyStatus = string(phase)
			//_, err = groupTarget.Update(context.TODO(), getGroup, metav1.UpdateOptions{})
			//if err != nil {
			//	logs.Errorf("Update group in other domain err:%v", err)
			//}
		} else { // 任务信息存在本域当中
			copyGroupName := key
			logs.Infof("==================key：%v", key)
			getGroup, err := gmo.clientsManager.GetGroup(copyGroupName, group.Namespace)
			if err != nil {
				logs.Infof("Failed to get group: %v", err)
			}
			for aSpecName, actionReference := range getGroup.Status.Actions {
				if aSpecName == actionSpecName {
					patchAction, err := json.Marshal(map[string]interface{}{
						"status": map[string]interface{}{
							"copy_status": string(phase),
						},
					})
					_, err = gmo.clientsManager.PatchAction(actionReference.Name, actionReference.Namespace, patchAction)
					if err != nil {
						logs.Infof("Failed to patch runtime: %v", err)
					}
				}
			}

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
			//_, err = gmo.groupClient.Patch(context.TODO(), copyGroupName, types.JSONPatchType, patchGroup, metav1.PatchOptions{})
			//if err != nil {
			//	logs.Errorf("Patch group err-7:%v", err)
			//}
			logs.Info("#######################groupSpec.Replicas > 0#########设置副本action的状态为running-----成功")
		}
	}
}

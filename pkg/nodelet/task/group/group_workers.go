package group

import (
	"context"
	"encoding/json"
	"hit.edu/framework/pkg/apimachinery/types"
	metav1 "hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/client-go/clients/typed/core"
	"hit.edu/framework/pkg/nodelet/task/task"
	"strconv"
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
		action := &group.Spec.Actions[i]
		for j := range action.Spec.Runtimes {
			ru := &action.Spec.Runtimes[j]
			if action.Status.RuntimeStatus[j].Phase == apis.Successed || action.Status.RuntimeStatus[j].Phase == apis.DeployCheck {
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
		newAction := &apis.Action{
			ObjectMeta: metav1.ObjectMeta{Name: groupSpec.Actions[i].Name, Namespace: ""},
			TypeMeta:   metav1.TypeMeta{Kind: "Action", APIVersion: "resources/v1"},
			Spec:       groupSpec.Actions[i].Spec,
			Status:     groupSpec.Actions[i].Status,
		}
		//action := &groupSpec.Actions[i]
		_, err := gw.actionClient.Create(context.TODO(), newAction, metav1.CreateOptions{})
		if err != nil {
			logs.Errorf("Create action failed,err-1:%v", err)
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
					//task1.Status.Phase = apis.DeployCheck //首先设置Task的状态为DeployCheck
					logs.Trace("=================Task的状态被修改为DeployCheck")
					//task1.Status.LastTime = times //隐藏
					patchTask, err := json.Marshal(map[string]interface{}{
						"status": map[string]interface{}{
							"phase": apis.DeployCheck, //value值不同
						},
					})
					if err != nil {
						logs.Errorf("Json Marshal failed, err:%v", err)
					}
					_, err = gw.taskClient.Patch(context.TODO(), task1.Name, types.StrategicMergePatchType, patchTask, metav1.PatchOptions{})
					if err != nil {
						logs.Errorf("Patch group error:%v", err)
					}
				}
				for i := range task1.Spec.Groups { //同时得更新TaskSpec下的Group以及TaskStatus下的GroupStatus为当前的group信息
					if gr.Name == task1.Spec.Groups[i].Name {
						//task1.Spec.Groups[i].Spec = gr.Spec
						patchTask1, err := json.Marshal([]map[string]interface{}{
							{
								"op":    "replace",
								"path":  "/spec/groups/" + strconv.Itoa(i) + "/spec",
								"value": gr.Spec, // 这里替换为你需要的 Phase 值
							},
						})
						if err != nil {
							logs.Errorf("Json Marshal failed, err:%v", err)
						}
						_, err = gw.taskClient.Patch(context.TODO(), task1.Name, types.JSONPatchType, patchTask1, metav1.PatchOptions{})
						if err != nil {
							logs.Errorf("Patch group error-5:%v", err)
						}

						//task1.Spec.Groups[i].Status = gr.Status
						patchTask2, err := json.Marshal([]map[string]interface{}{
							{
								"op":    "replace",
								"path":  "/spec/groups/" + strconv.Itoa(i) + "/status",
								"value": gr.Status, // 这里替换为你需要的 Phase 值
							},
						})
						if err != nil {
							logs.Errorf("Json Marshal failed, err:%v", err)
						}
						_, err = gw.taskClient.Patch(context.TODO(), task1.Name, types.JSONPatchType, patchTask2, metav1.PatchOptions{})
						if err != nil {
							logs.Errorf("Patch group error-5:%v", err)
						}

						//task1.Status.GroupStatus[i] = gr.Status
						patchTask3, err := json.Marshal([]map[string]interface{}{
							{
								"op":    "replace",
								"path":  "/status/group_status/" + strconv.Itoa(i),
								"value": gr.Status, // 这里替换为你需要的 Phase 值
							},
						})
						if err != nil {
							logs.Errorf("Json Marshal failed, err:%v", err)
						}
						_, err = gw.taskClient.Patch(context.TODO(), task1.Name, types.JSONPatchType, patchTask3, metav1.PatchOptions{})
						if err != nil {
							logs.Errorf("Patch group error-5:%v", err)
						}
						break
					}
				}
				//// 将task信息提交到etcd上去，这里的话使用Patch，不要使用Update，因为可能一个Task里面有多个Group，如果每个Group都使用Update更新，会有问题
				//_, err1 = gw.taskClient.Update(context.TODO(), task1, metav1.UpdateOptions{})
				//if err1 != nil {
				//	logs.Errorf("Etcd update task:%v err:%v, now is handing group:%v", taskName, err1, gr.Spec.Name) //这里出错
				//	// 再次上传
				//time.Sleep(200 * time.Millisecond)
				//	_, err1 = gw.taskClient.Update(context.TODO(), task1, metav1.UpdateOptions{})
				//}
				break //后续就不用再遍历Task列表了，直接结束
			}
		}
	}
	//将group信息提交到etcd上去，使用update更新--出现一次报错  TODO 为了适配迁移，如果后面替换为Patch操作，那么要使用gr.ObjectMeta.Name 来进行patch，因为目前规定gr.ObjectMeta.Name为不同group的标识（针对副本、源group）
	// 有一个问题，就是怎么直接更新GroupStatus呢
	patchGroup1, err := json.Marshal(map[string]interface{}{
		"status": groupStatus,
	})
	patchGroup2, err2 := json.Marshal(map[string]interface{}{
		"spec": groupSpec,
	})
	_, err = gw.groupClient.Patch(context.TODO(), gr.Name, types.StrategicMergePatchType, patchGroup1, metav1.PatchOptions{})
	if err != nil {
		logs.Errorf("Patch group error222:%v", err)
	}
	_, err2 = gw.groupClient.Patch(context.TODO(), gr.Name, types.StrategicMergePatchType, patchGroup2, metav1.PatchOptions{})
	if err2 != nil {
		logs.Errorf("Patch group error222:%v", err)
	}
}

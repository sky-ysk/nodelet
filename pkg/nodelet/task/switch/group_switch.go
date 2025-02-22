package _switch

import (
	"context"
	"encoding/json"
	"hit.edu/framework/pkg/apimachinery/types"
	apis "hit.edu/framework/pkg/apis/cores"
	metav1 "hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/client-go/clients/typed/core"
	"hit.edu/framework/pkg/component-base/logs"
	"hit.edu/framework/pkg/nodelet/task/group"
	"hit.edu/framework/pkg/nodelet/task/interaction/intwithRuntime"
	"hit.edu/framework/pkg/nodelet/task/runtime"
	"strconv"
	"time"
)

type GroupSwitch struct {
	// 存储任务队列
	groupQueues *group.GroupQueues
	switchCheck *SwitchCheck
	// client-go
	nodesClient  core.NodeInterface
	groupsClient core.GroupInterface
	// 存储RuntimeManager
	runtimeManager *runtime.RuntimeManager
	// grpc-客户端clients管理
	grpcClientsManager *intwithRuntime.ClientsManager
}

func NewSwitchManager(groupQueues *group.GroupQueues, nodesClient core.NodeInterface, groupsClient core.GroupInterface, runtimeManager *runtime.RuntimeManager, clientsManager *intwithRuntime.ClientsManager) *GroupSwitch {
	return &GroupSwitch{
		groupQueues:        groupQueues,
		switchCheck:        NewSwitchCheck(nodesClient),
		groupsClient:       groupsClient,
		nodesClient:        nodesClient,
		runtimeManager:     runtimeManager,
		grpcClientsManager: clientsManager,
	}
}

func (sw *GroupSwitch) StartSwitchService() {
	logs.Info("switch controller model start")
	go sw.startListeningSwitchCondition() //开启监听切换条件
}

// 开启grpc服务端---备注，只有group需要细粒度控制的话，才需要开启grpc服务端 ------切换器作为grpc服务端??
//func (sw *GroupSwitch) startGrpcServer() {
//
//}

// 轮询监听切换条件
func (sw *GroupSwitch) startListeningSwitchCondition() {
	for {
		select {
		case <-time.After(time.Millisecond * 600):
			runningGroups := sw.groupQueues.GetAllRunning()
			for i := range runningGroups {
				group := runningGroups[i]
				// 从etcd当中读取group信息
				newGroup, err := sw.groupsClient.Get(context.TODO(), group.Name, metav1.GetOptions{})
				nodes, err := sw.nodesClient.Get(context.TODO(), "CloudNode1", metav1.GetOptions{})
				if err != nil {
					logs.Errorf("Get node:%s from etcd err: %v", nodes.Name, err)
				}
				if sw.switchCheck.SwitchCondition.CheckCondition(nodes) {
					sw.groupMigration(newGroup)
				}
			}
		}
	}
}

// 任务迁移 ：1）添加group-copy到etcd 2）
func (sw *GroupSwitch) groupMigration(g *apis.Group) {
	// 从etcd当中读取group信息
	group, err := sw.groupsClient.Get(context.TODO(), g.Name, metav1.GetOptions{})
	g = group
	logs.Infof("group:%v migration start", g.Name) //此处作为迁移的开始
	var groupCopyName string
	if group.Spec.Replicas > 0 {
		// 说明当前group已经提前往etcd里写入了副本group，那么此处就不用再写入了，只需要将原先写的副本group信息当中的groupCopy.Status.CopyStatus 改为"Starting"即可-采用patch
		patchGroup, err := json.Marshal(map[string]interface{}{
			"status": map[string]interface{}{
				"copy_status": "Starting",
			},
		})
		if err != nil {
			logs.Errorf("Json Marshal failed, err:%v", err)
		}
		// 这里还没想好怎么解决副本group的名字问题----待解决
		groupCopyName = "Reason-Copy"
		_, err = sw.groupsClient.Patch(context.TODO(), groupCopyName, types.StrategicMergePatchType, patchGroup, metav1.PatchOptions{})
		if err != nil {
			logs.Errorf("Patch group error-6:%v", err)
		}
		logs.Info("===================================将副本任务的copy_Status修改为Starting")
	} else {
		// 复制创建一个全新的副本group信息（注意Succeed的Phase不用修改，DeployCheck和Running状态需要修改），另外还需要将副本的groupStatus改为Starting
		groupCopy := NewGroupInfoCopy(g, false) //第二个参数表示是否为提前写入etcd，这里为否
		groupCopyName = groupCopy.Name
		// 将副本group信息写入到etcd当中，目前还只适配本域内迁移
		logs.Infof("group:%v===================", groupCopy.Name)
		_, err = sw.groupsClient.Create(context.TODO(), groupCopy, metav1.CreateOptions{})
		if err != nil {
			logs.Errorf("Create group:%s err: %v", groupCopy.Name, err)
		}
	}

	// 修改源group的Status.phase为Migrating
	patchGroup, err := json.Marshal(map[string]interface{}{
		"status": map[string]interface{}{
			"phase": apis.Migrating,
		},
	})
	if err != nil {
		logs.Errorf("Json Marshal failed, err:%v", err)
	}
	patchResult, err := sw.groupsClient.Patch(context.TODO(), g.Name, types.StrategicMergePatchType, patchGroup, metav1.PatchOptions{})
	if err != nil {
		logs.Errorf("Patch group error-2:%v", err)
	}
	logs.Infof("**************源任务groupStatus.Phase:%v", patchResult.Status.Phase)
	// 获取任务group中需要迁移的runtime的当前的执行状态，并将该状态写入到副本group上的runtimeStatus中的keyStatus，并关闭源group中Running的runtime   注意对于DeployCheck的任务，就不采用这种读取状态并写入的方式
	// 注意：如果说Runtime本身没有细粒度控制的话，就不用再保存任务状态以及写入到副本任务上去
	for i := range g.Spec.Actions {
		action := &g.Spec.Actions[i]
		actionStatus := g.Spec.Actions[i].Status
		if actionStatus.Phase == apis.Successed {
			continue
		}
		if actionStatus.Phase == apis.Running {
			for j := range actionStatus.RuntimeStatus {
				runtime := &action.Spec.Runtimes[j]
				runtimeStatus := actionStatus.RuntimeStatus[j]
				if runtimeStatus.Phase == apis.Successed {
					continue
				}
				if runtimeStatus.Phase == apis.Running && action.Spec.Runtimes[j].EnableFineGrainedControl { // runtime正在运行，且runtime是细粒度控制的
					logs.Info("!!!!!!!!!!!!!!!!!!!!!!!!!")
					data := sw.runtimeManager.StoreData(g, action, runtime, i, j) // 获取group下的正在执行runtime的关键数据
					// 将获取到的任务关键装填数据写入到本域的etcd上的副本group当中
					patchGroup, err := json.Marshal([]map[string]interface{}{
						{
							"op":    "replace",
							"path":  "/status/action_status/" + strconv.Itoa(i) + "/status/" + strconv.Itoa(j) + "/key_status",
							"value": data, // 这里替换为你需要的 Phase 值
						},
					})
					if err != nil {
						logs.Errorf("Json Marshal failed, err:%v", err)
					}
					patchResult, err = sw.groupsClient.Patch(context.TODO(), groupCopyName, types.JSONPatchType, patchGroup, metav1.PatchOptions{})
					if err != nil {
						logs.Errorf("Patch group error-5:%v", err)
					}
					logs.Infof("*******副本任务runtimeStatus.keyStatus:%v", patchResult.Status.ActionStatus[0].RuntimeStatus[0].KeyStatus)
					// 关闭源任务当中的runtime
					logs.Info("^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^")
					//err = sw.runtimeManager.StopRuntime(g, action, runtime, i, j)
					time.Sleep(1 * time.Second)
					err = sw.runtimeManager.Kill(g, action, runtime)
					if err != nil {
						logs.Errorf("Stop group error:%v", err)
					}
				}
				// 跟python任务建立grpc连接--先隐藏测试
				//port := runtime.EnableFineGrainedControlPort
				//client := grpc_client.NewRuntimeClient(port, runtimeStatus.RuntimeID)
				//success := sw.grpcClientsManager.AddRuntimeClientConnection(client)
				//if !success {
				//	logs.Errorf("GrpcClientManager AddRuntimeClientConnection error:%v", err)
				//}
			}
		}
		// 将源group从Running队列迁移到Completed队列
		logs.Info("--------------DeleteFromRunningAndAddToMigratedQueue=====================")
		ok := sw.groupQueues.DeleteFromRunningAndAddToMigrated(g.Status.GroupID)
		if !ok {
			logs.Error("Delete group from running queue and add to completed queue failed-2")
		}
	}

}

// 新增一个创建一个空白的Group信息，删除不必要的内容（例如Running、DeployCheck的属性都得改为Unknown，时间也得修改）
func NewGroupInfoCopy(g *apis.Group, isAhead bool) *apis.Group {
	// 将原始对象序列化为JSON
	data, err := json.Marshal(g)
	if err != nil {
		logs.Errorf("Marshal group:%v error:%v", g.Name, err)
	}
	// 反序列化为新的对象
	var copyGroup apis.Group // 非指针
	err = json.Unmarshal(data, &copyGroup)
	if err != nil {
		logs.Errorf("Unmarshal group:%v error:%v", g.Name, err)
	}
	var groupCopy = &copyGroup // 转换为指针
	// 接下来修改这个复制出来的Group信息，首先修改group.Name
	groupCopy.Name = "Reason-Copy"
	// 将groupCopy的ResourceVersion置空，注意：这是必须的，否则报错
	groupCopy.ResourceVersion = ""
	// 修改GroupSpec下的Actions数组当中ActionStatus的Phase和time
	groupCopy.Spec.IsCopy = true // 标记改Group为副本group
	// 这个副本group信息当中，其副本数量直接置为0（意思是：不再为副本订制副本）
	groupCopy.Spec.Replicas = 0
	for i := range groupCopy.Spec.Actions {
		action := &groupCopy.Spec.Actions[i]
		if action.Status.Phase == apis.Successed {
			continue
		} else {
			action.Status.Phase = apis.Unknown
			action.Status.StartAt = apis.Time{time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC)}
			action.Status.LastTime = apis.Time{time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC)}
		}
		for j := range action.Status.RuntimeStatus {
			runtimeStatus := &action.Status.RuntimeStatus[j]
			runtimeStatus.ProcessId = ""
			if runtimeStatus.Phase != apis.Successed {
				runtimeStatus.Phase = apis.Unknown
				runtimeStatus.StartAt = apis.Time{time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC)}
				runtimeStatus.LastTime = apis.Time{time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC)}
			}
		}

	}
	// 修改GroupStatus下面的phase、ActionStatus的Phase以及runtimeStatus的Phase
	// 修改副本group信息中的属性来标记副本任务需要马上启动(这个属性会在copyPending队列当中去轮询检查的)
	if isAhead {
		logs.Infof("============预部署副本===========")
		groupCopy.Status.CopyStatus = "Waiting" //注意后面真正切换的时候，需要将这个参数改为Starting
	} else {
		logs.Infof("============直接启动副本===========")
		groupCopy.Status.CopyStatus = "Starting"
	}
	// 将groupStatus下的node属性置空
	groupCopy.Status.Node = ""
	if groupCopy.Status.Phase != apis.Successed {
		groupCopy.Status.Phase = apis.Unknown
	}
	for i := range groupCopy.Status.ActionStatus {
		actionStatus := &groupCopy.Status.ActionStatus[i]
		if actionStatus.Phase == apis.Successed {
			continue
		} else {
			actionStatus.Phase = apis.Unknown
			actionStatus.StartAt = apis.Time{time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC)}
			actionStatus.LastTime = apis.Time{time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC)}
		}
		for j := range actionStatus.RuntimeStatus {
			runtimeStatus := &actionStatus.RuntimeStatus[j]
			runtimeStatus.ProcessId = ""
			if runtimeStatus.Phase != apis.Successed {
				runtimeStatus.Phase = apis.Unknown
				runtimeStatus.StartAt = apis.Time{time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC)}
				runtimeStatus.LastTime = apis.Time{time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC)}
			}
		}
	}
	return groupCopy
}
func newGroupInfoCopy1(g *apis.Group) *apis.Group {

	return nil
}

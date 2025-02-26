package controller

import (
	"context"
	"encoding/json"
	"hit.edu/framework/pkg/apimachinery/types"
	"hit.edu/framework/pkg/apimachinery/util/wait"
	apis "hit.edu/framework/pkg/apis/cores"
	metav1 "hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/client-go/clients/typed/core"
	"hit.edu/framework/pkg/client-go/util/workqueue"
	"hit.edu/framework/pkg/component-base/logs"
	"hit.edu/framework/pkg/nodelet/task/group"
	"hit.edu/framework/pkg/nodelet/task/runtime"
	"k8s.io/client-go/tools/cache"
	"strconv"
	"time"
)

type SwitchController struct {
	nodeInformer    cache.SharedIndexInformer
	groupInformer   cache.SharedIndexInformer
	nodeQueue       workqueue.RateLimitingInterface
	groupQueue      workqueue.RateLimitingInterface
	runtimeManager  *runtime.RuntimeManager
	groupQueues     *group.GroupQueues
	groupsClient    core.GroupInterface
	nodesClient     core.NodeInterface
	thresholdConfig ThresholdConfig
}

// 阈值配置结构
type ThresholdConfig struct {
	CPU     float64
	Memory  float64
	Storage float64
}

func NewSwitchController(
	nodeInformer cache.SharedIndexInformer,
	groupInformer cache.SharedIndexInformer,
	runtimeManager *runtime.RuntimeManager,
	groupQueues *group.GroupQueues,
	threshold ThresholdConfig) *SwitchController {

	// 初始化速率限制队列
	rateLimiter := workqueue.NewItemExponentialFailureRateLimiter(5*time.Millisecond, 30*time.Second)

	c := &SwitchController{
		nodeInformer:    nodeInformer,
		groupInformer:   groupInformer,
		nodeQueue:       workqueue.NewNamedRateLimitingQueue(rateLimiter, "nodeMigration"),
		groupQueue:      workqueue.NewNamedRateLimitingQueue(rateLimiter, "groupMigration"),
		runtimeManager:  runtimeManager,
		groupQueues:     groupQueues,
		thresholdConfig: threshold,
	}
	// 注册节点事件处理器
	nodeInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: func(oldObj, newObj interface{}) {
			if c.isResourceOverThreshold(newObj.(*apis.Node)) {
				key, _ := cache.MetaNamespaceKeyFunc(newObj)
				c.nodeQueue.Add(key)
			}
		},
	})

	// 注册任务组事件处理器
	groupInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: func(oldObj, newObj interface{}) {
			key, _ := cache.MetaNamespaceKeyFunc(newObj)
			c.groupQueue.Add(key)
		},
	})

	return c
}

// 核心运行逻辑
func (c *SwitchController) Run(workers int, stopCh <-chan struct{}) {
	defer c.nodeQueue.ShutDown()
	defer c.groupQueue.ShutDown()

	go c.nodeInformer.Run(stopCh)
	go c.groupInformer.Run(stopCh)

	if !cache.WaitForCacheSync(stopCh,
		c.nodeInformer.HasSynced,
		c.groupInformer.HasSynced) {
		logs.Errorf("Timed out waiting for caches to sync")
	}

	for i := 0; i < workers; i++ {
		go wait.Until(c.runNodeWorker, time.Second, stopCh)
		//go wait.Until(c.runGroupWorker, time.Second, stopCh)
	}

	<-stopCh
}

// 节点工作线程
func (c *SwitchController) runNodeWorker() {
	for c.processNextNodeItem() {
	}
}

// 处理节点事件
func (c *SwitchController) processNextNodeItem() bool {
	key, quit := c.nodeQueue.Get()
	if quit {
		return false
	}
	defer c.nodeQueue.Done(key)

	obj, exists, err := c.nodeInformer.GetIndexer().GetByKey(key.(string))
	if err != nil {
		logs.Errorf("Fetch node %s error: %v", key, err)
		return true
	}

	if !exists {
		logs.Infof("Node %s has been deleted", key)
		return true
	}

	node := obj.(*apis.Node)
	groups := c.getGroupsOnNode(node.Name) // 获取当前node节点上的所有的group

	// 执行迁移逻辑
	for _, g := range groups {
		if err := c.migrateGroup(g); err != nil {
			c.nodeQueue.AddRateLimited(key)
			return true
		}
	}

	c.nodeQueue.Forget(key)
	return true
}

// 迁移任务组核心逻辑
func (c *SwitchController) migrateGroup(g *apis.Group) error {
	//// 添加Finalizer防止资源删除
	//if !containsString(g.Finalizers, "migration-protection") {
	//	g.Finalizers = append(g.Finalizers, "migration-protection")
	//	if _, err := c.groupsClient.Update(context.TODO(), g, metav1.UpdateOptions{}); err != nil {
	//		return err
	//	}
	//}

	// 执行迁移操作（保留原有迁移逻辑）
	groupCopy := NewGroupInfoCopy(g, false)
	if _, err := c.groupsClient.Create(context.TODO(), groupCopy, metav1.CreateOptions{}); err != nil {
		return err
	}

	// 更新原组状态
	patch := []byte(`{"status":{"phase":"Migrating"}}`)
	_, err := c.groupsClient.Patch(context.TODO(), g.Name, types.StrategicMergePatchType, patch, metav1.PatchOptions{})
	return err
}

// 资源阈值检查
func (c *SwitchController) isResourceOverThreshold(node *apis.Node) bool {
	cpu := parseResourceValue(node.Status.Usage["cpu"][0].Values["AveUtil"])
	mem := parseResourceValue(node.Status.Usage["memory"][0].Values["Usage"])
	storage := parseResourceValue(node.Status.Usage["storage"][0].Values["Usage"])

	return cpu > c.thresholdConfig.CPU ||
		mem > c.thresholdConfig.Memory ||
		storage > c.thresholdConfig.Storage
}

func (c *SwitchController) getGroupsOnNode(name string) []*apis.Group {
	groupList, err := c.groupsClient.List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		logs.Errorf("list groups error: %v", err)
	}
	var groups []*apis.Group
	for i := range groupList.Items {
		group := &groupList.Items[i]
		groups = append(groups, group)
	}
	return groups
}

// 辅助函数：解析资源值
func parseResourceValue(s string) float64 {
	value, err := strconv.ParseFloat(s, 64)
	if err != nil {
		logs.Errorf("Convert value failed, err:%v", err)
		return 0
	}
	return value
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

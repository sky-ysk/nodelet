package controller

//import (
//	"context"
//	"hit.edu/framework/pkg/apimachinery/types"
//	"hit.edu/framework/pkg/apimachinery/util/wait"
//	apis "hit.edu/framework/pkg/apis/cores"
//	metav1 "hit.edu/framework/pkg/apis/meta"
//	"hit.edu/framework/pkg/client-go/clients/typed/core"
//	"hit.edu/framework/pkg/client-go/util/workqueue"
//	"hit.edu/framework/pkg/component-base/logs"
//	"hit.edu/framework/pkg/nodelet/task/group"
//	"hit.edu/framework/pkg/nodelet/task/runtime"
//	"k8s.io/client-go/tools/cache"
//	"strconv"
//	"time"
//)
//
//type SwitchController struct {
//	nodeInformer    cache.SharedIndexInformer
//	groupInformer   cache.SharedIndexInformer
//	nodeQueue       workqueue.RateLimitingInterface
//	groupQueue      workqueue.RateLimitingInterface
//	runtimeManager  *runtime.RuntimeManager
//	groupQueues     *group.GroupQueues
//	groupsClient    core.GroupInterface
//	nodesClient     core.NodeInterface
//	thresholdConfig ThresholdConfig
//}
//
//// 阈值配置结构
//type ThresholdConfig struct {
//	CPU     float64
//	Memory  float64
//	Storage float64
//}
//
//func NewSwitchController(
//	nodeInformer cache.SharedIndexInformer,
//	groupInformer cache.SharedIndexInformer,
//	runtimeManager *runtime.RuntimeManager,
//	groupQueues *group.GroupQueues,
//	threshold ThresholdConfig) *SwitchController {
//
//	// 初始化速率限制队列
//	rateLimiter := workqueue.NewItemExponentialFailureRateLimiter(5*time.Millisecond, 30*time.Second)
//
//	c := &SwitchController{
//		nodeInformer:    nodeInformer,
//		groupInformer:   groupInformer,
//		nodeQueue:       workqueue.NewNamedRateLimitingQueue(rateLimiter, "nodeMigration"),
//		groupQueue:      workqueue.NewNamedRateLimitingQueue(rateLimiter, "groupMigration"),
//		runtimeManager:  runtimeManager,
//		groupQueues:     groupQueues,
//		thresholdConfig: threshold,
//	}
//	// 注册节点事件处理器
//	nodeInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
//		UpdateFunc: func(oldObj, newObj interface{}) {
//			if c.isResourceOverThreshold(newObj.(*apis.Node)) {
//				key, _ := cache.MetaNamespaceKeyFunc(newObj)
//				c.nodeQueue.Add(key)
//			}
//		},
//	})
//
//	// 注册任务组事件处理器
//	groupInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
//		UpdateFunc: func(oldObj, newObj interface{}) {
//			key, _ := cache.MetaNamespaceKeyFunc(newObj)
//			c.groupQueue.Add(key)
//		},
//	})
//
//	return c
//}
//
//// 核心运行逻辑
//func (c *SwitchController) Run(workers int, stopCh <-chan struct{}) {
//	defer c.nodeQueue.ShutDown()
//	defer c.groupQueue.ShutDown()
//
//	go c.nodeInformer.Run(stopCh)
//	go c.groupInformer.Run(stopCh)
//
//	if !cache.WaitForCacheSync(stopCh,
//		c.nodeInformer.HasSynced,
//		c.groupInformer.HasSynced) {
//		logs.Errorf("Timed out waiting for caches to sync")
//	}
//
//	for i := 0; i < workers; i++ {
//		go wait.Until(c.runNodeWorker, time.Second, stopCh)
//		//go wait.Until(c.runGroupWorker, time.Second, stopCh)
//	}
//
//	<-stopCh
//}
//
//// 节点工作线程
//func (c *SwitchController) runNodeWorker() {
//	for c.processNextNodeItem() {
//	}
//}
//
//// 处理节点事件
//func (c *SwitchController) processNextNodeItem() bool {
//	key, quit := c.nodeQueue.Get()
//	if quit {
//		return false
//	}
//	defer c.nodeQueue.Done(key)
//
//	obj, exists, err := c.nodeInformer.GetIndexer().GetByKey(key.(string))
//	if err != nil {
//		logs.Errorf("Fetch node %s error: %v", key, err)
//		return true
//	}
//
//	if !exists {
//		logs.Infof("Node %s has been deleted", key)
//		return true
//	}
//
//	node := obj.(*apis.Node)
//	groups := c.getGroupsOnNode(node.Name) // 获取当前node节点上的所有的group
//
//	// 执行迁移逻辑
//	for _, g := range groups {
//		if err := c.migrateGroup(g); err != nil {
//			c.nodeQueue.AddRateLimited(key)
//			return true
//		}
//	}
//
//	c.nodeQueue.Forget(key)
//	return true
//}
//
//// 迁移任务组核心逻辑
//func (c *SwitchController) migrateGroup(g *apis.Group) error {
//	//// 添加Finalizer防止资源删除
//	//if !containsString(g.Finalizers, "migration-protection") {
//	//	g.Finalizers = append(g.Finalizers, "migration-protection")
//	//	if _, err := c.groupsClient.Update(context.TODO(), g, metav1.UpdateOptions{}); err != nil {
//	//		return err
//	//	}
//	//}
//
//	// 执行迁移操作（保留原有迁移逻辑）
//	groupCopy := NewGroupInfoCopy(g, false)
//	if _, err := c.groupsClient.Create(context.TODO(), groupCopy, metav1.CreateOptions{}); err != nil {
//		return err
//	}
//
//	// 更新原组状态
//	patch := []byte(`{"status":{"phase":"Migrating"}}`)
//	_, err := c.groupsClient.Patch(context.TODO(), g.Name, types.StrategicMergePatchType, patch, metav1.PatchOptions{})
//	return err
//}
//
//// 资源阈值检查
//func (c *SwitchController) isResourceOverThreshold(node *apis.Node) bool {
//	cpu := parseResourceValue(node.Status.Usage["cpu"][0].Values["AveUtil"])
//	mem := parseResourceValue(node.Status.Usage["memory"][0].Values["Usage"])
//	storage := parseResourceValue(node.Status.Usage["storage"][0].Values["Usage"])
//
//	return cpu > c.thresholdConfig.CPU ||
//		mem > c.thresholdConfig.Memory ||
//		storage > c.thresholdConfig.Storage
//}
//
//func (c *SwitchController) getGroupsOnNode(name string) []*apis.Group {
//	groupList, err := c.groupsClient.List(context.TODO(), metav1.ListOptions{})
//	if err != nil {
//		logs.Errorf("list groups error: %v", err)
//	}
//	var groups []*apis.Group
//	for i := range groupList.Items {
//		group := &groupList.Items[i]
//		groups = append(groups, group)
//	}
//	return groups
//}
//
//// 辅助函数：解析资源值
//func parseResourceValue(s string) float64 {
//	value, err := strconv.ParseFloat(s, 64)
//	if err != nil {
//		logs.Errorf("Convert value failed, err:%v", err)
//		return 0
//	}
//	return value
//}

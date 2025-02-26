package controller

import (
	"hit.edu/framework/pkg/apimachinery/fields"
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/client-go/clients"
	"hit.edu/framework/pkg/client-go/clients/typed/core"
	"hit.edu/framework/pkg/client-go/tools/cache"
	"hit.edu/framework/pkg/client-go/util/workqueue"
	"hit.edu/framework/pkg/nodelet/task/group"
	"hit.edu/framework/pkg/nodelet/task/runtime"
	"time"
)

const (
	controllerName = "group-switch-controller"
	resyncPeriod   = 30 * time.Second
)

type queueItem struct {
	key  string
	kind string
}

type MigrationController struct { // 自定义的业务控制器（适配迁移触发），依赖与Informer来收集群中资源变化的事件，并可以基于这些事件采取某些操作，如创建、删除或更新某些资源。
	// client-go
	clientSet   *clients.ClientSet
	nodeClient  core.NodeInterface
	groupClient core.GroupInterface

	// Group 相关组件
	groupIndexer  cache.Indexer
	groupInformer cache.Controller // cache.Controller当中包含了cache.Index ,这里我们将cache.Controller中的Indexer拎出来，是为了更好地编写代码而已，其实不要这个Indexer也是OK的，因为cache.Controller当中也是含有Indexer的

	// Node 相关组件
	nodeIndexer  cache.Indexer    // 本地缓存，提供关于资源的快速查询（索引查询）。 informer会调用Indexer的Add、update、delete方法来实现资源的同步于更新
	nodeInformer cache.Controller // cache.Controller 是 k8s中用于控制器模式的核心组件，它封装了资源的监听和事件处理机制，通常用于协调控制循环。，作用：监听资源变化、缓存资源、触发处理逻辑

	// 工作队列
	queue workqueue.TypedRateLimitingInterface[queueItem]

	// 运行时依赖组件
	runtimeManager *runtime.RuntimeManager

	groupQueues *group.GroupQueues
}

func NewMigrationController(clientSet *clients.ClientSet, nodeClient core.NodeInterface, groupClient core.GroupInterface, runtimeManager *runtime.RuntimeManager, groupQueues *group.GroupQueues) *MigrationController {
	//创建Node资源的List Watcher
	nodeListWatcher := cache.NewListWatchFromClient(clientSet.Core().RESTClient(), "nodes", "", fields.Everything())
	groupListWatcher := cache.NewListWatchFromClient(clientSet.Core().RESTClient(), "groups", "", fields.Everything())
	queue := workqueue.NewTypedRateLimitingQueue(workqueue.DefaultTypedControllerRateLimiter[queueItem]())
	nodeOptions := cache.InformerOptions{
		ListerWatcher: nodeListWatcher,
		ObjectType:    &apis.Node{}, // 要监听的资源类型
		Handler:       createEventHandler(queue, "node"),
		ResyncPeriod:  0, // ResyncPeriod，0表示不定期重新同步
		Indexers:      cache.Indexers{},
	}
	groupOptions := cache.InformerOptions{
		ListerWatcher: groupListWatcher,
		ObjectType:    &apis.Group{},
		Handler:       createEventHandler(queue, "group"),
		ResyncPeriod:  0,
		Indexers: cache.Indexers{
			"ByNode": func(obj interface{}) ([]string, error) {
				group := obj.(*apis.Group)
				return []string{group.Status.Node}, nil
			},
		},
	}
	nodeIndexer, nodeInformer := cache.NewInformerWithOptions(nodeOptions)
	groupIndexer, groupInformer := cache.NewInformerWithOptions(groupOptions)
	return &MigrationController{
		clientSet:      clientSet,
		nodeClient:     nodeClient,
		groupClient:    groupClient,
		nodeIndexer:    nodeIndexer,
		nodeInformer:   nodeInformer,
		groupIndexer:   groupIndexer,
		groupInformer:  groupInformer,
		queue:          queue,
		runtimeManager: runtimeManager,
		groupQueues:    groupQueues,
	}
}

// 共享的事件处理器
func createEventHandler(queue workqueue.TypedRateLimitingInterface[queueItem], kind string) cache.ResourceEventHandler {
	return cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			enqueueWithKind(queue, obj, kind)
		},
		UpdateFunc: func(old, new interface{}) {
			enqueueWithKind(queue, new, kind)
		},
		DeleteFunc: func(obj interface{}) {
			enqueueWithKind(queue, obj, kind)
		},
	}
}

func enqueueWithKind(queue workqueue.TypedRateLimitingInterface[queueItem], obj interface{}, kind string) {
	key, err := cache.MetaNamespaceKeyFunc(obj) // 获取对象的唯一标识符，它尝试将对象转化为一个字符串键（结构体包含Name和Namespace），可以用于存储和查询
	if err == nil {
		queue.Add(queueItem{key: key, kind: kind})
	}
}

//
//func (c *MigrationController) syncHandler(item queueItem) error {
//	switch item.kind {
//	case "node":
//		return c.handleNodeEvent(item.key)
//	case "group":
//		return c.handleGroupEvent(item.key)
//	default:
//		return fmt.Errorf("unknown kind: %s", item.kind)
//	}
//}
//
//// 处理Node事件
//func (c *MigrationController) handleNodeEvent(key string) error {
//	obj, exists, err := c.nodeIndexer.GetByKey(key)
//	if err != nil {
//		return fmt.Errorf("error fetching object with key %s from store: %v", key, err)
//	}
//	if !exists {
//		return c.handleDeleteNode(key)
//	}
//	node := obj.(*apis.Node)
//	if c.shouldTriggerMigration(node) {
//		return c.triggerNodeMigration(node)
//	}
//	return nil
//}
//
//// 处理Group事件
//func (c *MigrationController) handleGroupEvnet(key string) error {
//	obj, exists, err := c.groupIndexer.GetByKey(key)
//	if err != nil {
//		return fmt.Errorf("error fetching object with key %s from store: %v", key, err)
//	}
//	if !exists {
//		return c.handleDeleteGroup(key)
//	}
//	group := obj.(*apis.Group)
//	return c.processGroupUpdate(group)
//}
//
//func (c *MigrationController) triggerNodeMigration(node *apis.Node) error {
//	// 通过索引获取关联的Groups
//	groups, err := c.groupIndexer.ByIndex("byNode", node.Name)
//	if err != nil {
//		return fmt.Errorf("Failed to obtain the node association group:%v", err)
//	}
//	for _, obj := range groups {
//		group := obj.(*apis.Group)
//		if err := c.migrateGroup(group); err != nil {
//			logs.Errorf("Failed to migrate group:%v, err:%v", group.Name, err)
//			continue
//		}
//	}
//	return nil
//}
//
//func (c *MigrationController) migrateGroup(group *apis.Group) error {
//	// 1、创建副本group
//
//	// 2、
//
//	// 3、
//
//}
//
//// 入队逻辑（带去重）
//func (c *Controller) enqueueGroup(obj interface{}) {
//	key, err := cache.MetaNamespaceKeyFunc(obj)
//	if err != nil {
//		logs.Errorf("Failed to get key for object: %v", err)
//		return
//	}
//	c.workqueue.Add(key)
//}
//func (c *Controller) Run(workers int, stopCh <-chan struct{}) {
//	defer c.workqueue.ShutDown()
//
//	// 等待缓存同步
//	if !cache.WaitForCacheSync(stopCh, c.groupsSynced) {
//		logs.Error("Timed out waiting for caches to sync")
//		return
//	}
//
//	// 启动 Worker 处理队列
//	for i := 0; i < workers; i++ {
//		go wait.Until(c.runWorker, time.Second, stopCh)
//	}
//
//	<-stopCh
//}
//
//func (c *Controller) runWorker() {
//	for c.processNextWorkItem() {
//	}
//}
//
//func (c *Controller) processNextWorkItem() bool {
//	obj, shutdown := c.workqueue.Get()
//	if shutdown {
//		return false
//	}
//
//	err := func(obj interface{}) error {
//		defer c.workqueue.Done(obj)
//		var key string
//		var ok bool
//		if key, ok = obj.(string); !ok {
//			c.workqueue.Forget(obj)
//			return fmt.Errorf("expected string in workqueue but got %#v", obj)
//		}
//
//		if err := c.syncHandler(key); err != nil {
//			c.workqueue.AddRateLimited(key)
//			return fmt.Errorf("error syncing '%s': %s", key, err.Error())
//		}
//
//		c.workqueue.Forget(obj)
//		return nil
//	}(obj)
//
//	if err != nil {
//		logs.Error(err)
//	}
//	return true
//}

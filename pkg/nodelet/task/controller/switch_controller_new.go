package controller

//
//import (
//	"context"
//	"encoding/json"
//	"fmt"
//	"hit.edu/framework/pkg/apimachinery/fields"
//	"hit.edu/framework/pkg/apimachinery/types"
//	"hit.edu/framework/pkg/apimachinery/util/wait"
//	apis "hit.edu/framework/pkg/apis/cores"
//	metav1 "hit.edu/framework/pkg/apis/meta"
//	"hit.edu/framework/pkg/client-go/clients"
//	"hit.edu/framework/pkg/client-go/clients/typed/core"
//	"hit.edu/framework/pkg/client-go/tools/cache"
//	"hit.edu/framework/pkg/client-go/util/workqueue"
//	"hit.edu/framework/pkg/component-base/logs"
//	"hit.edu/framework/pkg/nodelet/node"
//	"hit.edu/framework/pkg/nodelet/task/group"
//	"hit.edu/framework/pkg/nodelet/task/runtime"
//	"strconv"
//	"sync"
//	"time"
//)
//
//const (
//	resyncPeriod = 30 * time.Second
//)
//
//type MigrationController struct { // 自定义的业务控制器（适配迁移触发），依赖与Informer来收集群中资源变化的事件，并可以基于这些事件采取某些操作，如创建、删除或更新某些资源。
//	// client-go
//	groupClient core.GroupInterface
//
//	// Group 相关组件
//	groupIndexer  cache.Indexer    // 这个参数就是cache.Controller当中的Indexer缓存
//	groupInformer cache.Controller // cache.Controller当中包含了cache.Index ,这里我们将cache.Controller中的Indexer拎出来，是为了更好地编写代码而已，其实不要这个Indexer也是OK的，因为cache.Controller当中也是含有Indexer的
//
//	// Event 相关组件
//	eventIndexer  cache.Indexer    // 这个参数就是cache.Controller当中的Indexer缓存
//	eventInformer cache.Controller // cache.Controller当中包含了cache.Index ,这里我们将cache.Controller中的Indexer拎出来，是为了更好地编写代码而已，其实不要这个Indexer也是OK的，因为cache.Controller当中也是含有Indexer的
//
//	// 工作队列
//	queue workqueue.TypedRateLimitingInterface[string]
//
//	// 运行时依赖组件
//	runtimeManager *runtime.RuntimeManager
//
//	groupQueues *group.GroupQueues
//}
//
//func NewMigrationController(clientSet *clients.ClientSet, groupClient core.GroupInterface, runtimeManager *runtime.RuntimeManager, groupQueues *group.GroupQueues) *MigrationController {
//	// 测试 Group 资源访问
//	_, err := clientSet.Core().Groups("test").List(context.TODO(), metav1.ListOptions{})
//	if err != nil {
//		logs.Fatalf("无法列出 Groups: %v", err) // 输出具体错误
//	}
//	logs.Info("++++++++++++++++++++++++++++")
//	//创建资源的List Watcher
//	eventListWatcher := cache.NewListWatchFromClient(clientSet.Core().RESTClient(), "events", "test", fields.Everything())
//	groupListWatcher := cache.NewListWatchFromClient(clientSet.Core().RESTClient(), "groups", "test", fields.Everything())
//	queue := workqueue.NewTypedRateLimitingQueue(workqueue.DefaultTypedControllerRateLimiter[string]())
//	groupOptions := cache.InformerOptions{
//		ListerWatcher: groupListWatcher,
//		ObjectType:    &apis.Group{},
//		Handler: cache.ResourceEventHandlerFuncs{
//			AddFunc: func(obj interface{}) {
//				key, _ := cache.MetaNamespaceKeyFunc(obj)
//				queue.Add(key)
//			},
//			UpdateFunc: func(oldObj, newObj interface{}) {
//				key, _ := cache.MetaNamespaceKeyFunc(newObj)
//				queue.Add(key)
//			},
//			DeleteFunc: func(obj interface{}) {
//				key, _ := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
//				queue.Add(key)
//			},
//		},
//		ResyncPeriod: 0, //定义全量同步本地缓存的周期，即使没有资源变更事件发生，也会定期重新从 API Server 拉取全量数据。
//		Indexers: cache.Indexers{ // 通过Node名称，找到所有与之关联的Group
//			"ByNode": func(obj interface{}) ([]string, error) {
//				group := obj.(*apis.Group)
//				return []string{group.Status.Node}, nil
//			},
//		},
//	}
//	eventOptions := cache.InformerOptions{
//		ListerWatcher: eventListWatcher,
//		ObjectType:    &apis.Event{},
//		Handler: cache.ResourceEventHandlerFuncs{
//			AddFunc: func(obj interface{}) {
//				if event, ok := obj.(*apis.Event); ok && event.Message == node.NodeName && event.Reason == "MigrationTrigger" { // event关联的节点为本节点，且迁移理由为触发迁移
//					key, _ := cache.MetaNamespaceKeyFunc(obj)
//					queue.Add(key)
//				}
//			},
//		},
//		ResyncPeriod: 0,
//		Indexers:     cache.Indexers{},
//	}
//	eventIndexer, eventInformer := cache.NewInformerWithOptions(eventOptions)
//	//// 在 Informer 启动前预填充本地缓存
//	//eventIndexer.Add(&apis.Event{
//	//	ObjectMeta: metav1.ObjectMeta{
//	//		Name: "demo-event",
//	//	},
//	//})
//	groupIndexer, groupInformer := cache.NewInformerWithOptions(groupOptions)
//	//// 在 Informer 启动前预填充本地缓存
//	//groupIndexer.Add(&apis.Group{
//	//	ObjectMeta: metav1.ObjectMeta{
//	//		Name: "demo-group",
//	//	},
//	//})
//	return &MigrationController{
//		groupClient:    groupClient,
//		eventIndexer:   eventIndexer,
//		eventInformer:  eventInformer,
//		groupIndexer:   groupIndexer,
//		groupInformer:  groupInformer,
//		queue:          queue,
//		runtimeManager: runtimeManager,
//		groupQueues:    groupQueues,
//	}
//}
//
//// Run方法
//func (mc *MigrationController) Run(workers int, stopCh <-chan struct{}) {
//	defer mc.queue.ShutDown()
//
//	var wg sync.WaitGroup
//	wg.Add(2)
//
//	// 启动所有的Informer
//	go func() {
//		defer wg.Done()
//		mc.groupInformer.Run(stopCh)
//	}()
//	go func() {
//		defer wg.Done()
//		mc.eventInformer.Run(stopCh)
//	}()
//
//	//缓存同步仅指初始的列表操作完成，后续的更新是由Informer的Watch机制自动处理的，不需要手动同步。因此，在控制器启动时只需要等待一次初始同步即可，之后Informer会自动维护缓存的更新，不需要循环检查。
//	// 等待缓存同步 启动后第一次将全量数据加载到本地缓存中
//	if !cache.WaitForCacheSync(stopCh, mc.groupInformer.HasSynced, mc.eventInformer.HasSynced) { //WaitForCacheSync:是否同步完成，返回false的话报错
//		logs.Errorf("Timed out waiting for caches to sync")
//		return
//	}
//	logs.Info("缓存同步完成=======================")
//	wg.Add(workers)
//	for i := 0; i < workers; i++ {
//		go func() {
//			defer wg.Done()
//			wait.Until(mc.runWorker, time.Second, stopCh)
//		}()
//	}
//	<-stopCh
//	wg.Wait()
//}
//
//func (mc *MigrationController) runWorker() {
//	for {
//		// 获取队列项
//		item, quit := mc.queue.Get()
//		if quit {
//			return
//		}
//		defer mc.queue.Done(item)
//		if err := mc.handleEventEvent(item); err != nil {
//			// 使用指数退避的重试机制
//			if mc.queue.NumRequeues(item) < 5 {
//				mc.queue.AddRateLimited(item)
//			} else {
//				logs.Errorf("Error syncing item %v: %v", item, err)
//				mc.queue.Forget(item)
//			}
//		}
//	}
//}
//
//// 处理Event事件
//func (c *MigrationController) handleEventEvent(key string) error {
//	obj, exists, err := c.eventIndexer.GetByKey(key)
//	if err != nil {
//		return fmt.Errorf("error fetching object with key %s from store: %v", key, err)
//	}
//	// 情况1：Evnet已删除
//	if !exists {
//		return c.handleDeleteEvent(key)
//	}
//	event := obj.(*apis.Event)
//	// 1、节点资源不足触发的迁移
//	if event.InvolvedObject.Kind == "Node" {
//		nodeName := event.InvolvedObject.Name // TODO 目前还只适配了节点资源不足触发的迁移
//		return c.triggerNodeMigration(nodeName)
//	} else { // 2、人为想触发迁移某个任务
//		groupName := event.InvolvedObject.Name
//		return c.triggerGroupMigration(groupName)
//	}
//
//}
//func (mc *MigrationController) handleDeleteEvent(key string) error {
//	// 这块TODO 如果Event删除了，那就说明会不会是Event上传了又马上撤销了，那不管了
//	return nil
//}
//func (mc *MigrationController) triggerGroupMigration(groupName string) error {
//	group, err := mc.groupClient.Get(context.TODO(), groupName, metav1.GetOptions{})
//	if err != nil {
//		logs.Errorf("Get group %s failed: %v", groupName, err)
//	}
//	// 带基本的重试的逻辑
//	if err := retryOnError(3, func() error {
//		return mc.migrateGroup(group)
//	}); err != nil {
//		logs.Errorf("Migrate group %s failed: %v", groupName, err)
//	}
//	return nil
//}
//
//// 这里有个优化的地方，就是迁移的话，涉及到多个group同时迁移，那么这里普通做法是遍历，但是效果不咋好，所以我们这里可以加一个工作池（协程）
//func (c *MigrationController) triggerNodeMigration(nodeName string) error {
//	logs.Info("-----------------------触发迁移-------------------------")
//	// 标记节点为迁移状态，这样就不要调度到本节点---这个好像没有必要？
//	// 通过索引获取关联的Groups
//	groups, err := c.groupIndexer.ByIndex("ByNode", nodeName)
//	if err != nil {
//		return fmt.Errorf("Failed to obtain the node association group:%v", err)
//	}
//	const maxWorkers = 5
//	var (
//		wg     sync.WaitGroup
//		errMu  sync.Mutex
//		errors []error
//	)
//
//	workCh := make(chan interface{}, len(groups))
//	for _, group := range groups {
//		workCh <- group
//	}
//	close(workCh)
//
//	// 启动工作池
//	for i := 0; i < maxWorkers; i++ {
//		wg.Add(1)
//		go func() { // 相当于每个Workers开一个协程去处理，他们遍历workCh，空闲的Workers就去取
//			defer wg.Done()
//			for obj := range workCh {
//				group, ok := obj.(*apis.Group)
//				if !ok {
//					errMu.Lock()
//					errors = append(errors, fmt.Errorf("无效对象类型：%T", obj))
//					errMu.Unlock()
//					continue
//				}
//				// 带基本的重试的逻辑
//				if err := retryOnError(3, func() error {
//					return c.migrateGroup(group)
//				}); err != nil {
//					errMu.Lock()
//					errors = append(errors, fmt.Errorf("迁移失败[%s]:%w", group.Name, err))
//					errMu.Unlock()
//				}
//			}
//		}()
//	}
//	wg.Wait()
//
//	if len(errors) > 0 {
//		return fmt.Errorf("部分迁移失败(%d/%d): 首个错误: %v", len(errors), len(groups), errors[0])
//	}
//	return nil
//}
//
//func retryOnError(attempts int, fn func() error) error {
//	var err error
//	for i := 0; i < attempts; i++ {
//		if err = fn(); err == nil {
//			return nil
//		}
//		time.Sleep(time.Duration(i*i) * time.Second)
//	}
//	return err
//}
//
//// TODO 任务组迁移核心逻辑---我感觉这里的核心是要改成如果这个方法当中有一步没执行成功，那么如何再次执行，让其成功
//func (mc *MigrationController) migrateGroup(group *apis.Group) error {
//	// 增加一条规则：如果Group不是细粒度控制的，那么就不进行迁移---可能
//	// 1、检查当前状态
//	if group.Status.Phase == apis.Migrating || group.Status.Phase == apis.Migrated {
//		logs.Infof("Group:%v has Migrated", group.Name)
//		return nil
//	}
//	// 2、创建副本group
//	logs.Infof("Group:%v migration start", group.Name) //此处作为迁移的开始
//	var groupCopyName string
//	if group.Spec.Replicas > 0 {
//		// 说明当前group已经提前往etcd里写入了副本group，那么此处就不用再写入了，只需要将原先写的副本group信息当中的groupCopy.Status.CopyStatus 改为"Starting"即可-采用patch
//		patchGroup, err := json.Marshal(map[string]interface{}{
//			"status": map[string]interface{}{
//				"copy_status": "Starting",
//			},
//		})
//		if err != nil {
//			logs.Errorf("Json Marshal failed, err:%v", err)
//		}
//		// 这里还没想好怎么解决副本group的名字问题----待解决
//		groupCopyName = "Reason-Copy"
//		_, err = mc.groupClient.Patch(context.TODO(), groupCopyName, types.StrategicMergePatchType, patchGroup, metav1.PatchOptions{})
//		if err != nil {
//			logs.Errorf("Patch group error-6:%v", err)
//		}
//		logs.Info("Change the copy_Status of the copy task to Starting")
//	} else {
//		// 复制创建一个全新的副本group信息（注意Succeed的Phase不用修改，DeployCheck和Running状态需要修改），另外还需要将副本的groupStatus改为Starting
//		groupCopy := NewGroupInfoCopy(group, false) //第二个参数表示是否为提前写入etcd，这里为否
//		groupCopyName = groupCopy.Name
//		// 将副本group信息写入到etcd当中，目前还只适配本域内迁移
//		logs.Infof("group:%v######################################", groupCopy.Name)
//		_, err := mc.groupClient.Create(context.TODO(), groupCopy, metav1.CreateOptions{})
//		if err != nil {
//			logs.Errorf("Create group:%s err: %v", groupCopy.Name, err)
//		}
//	}
//
//	// 修改源group的Status.phase为Migrating
//	patchGroup, err := json.Marshal(map[string]interface{}{
//		"status": map[string]interface{}{
//			"phase": apis.Migrating,
//		},
//	})
//	if err != nil {
//		logs.Errorf("Json Marshal failed, err:%v", err)
//	}
//	patchResult, err := mc.groupClient.Patch(context.TODO(), group.Name, types.StrategicMergePatchType, patchGroup, metav1.PatchOptions{})
//	if err != nil {
//		logs.Errorf("Patch group error-2:%v", err)
//	}
//	logs.Infof("Source groupStatus.Phase:%v", patchResult.Status.Phase)
//	// 获取任务group中需要迁移的runtime的当前的执行状态，并将该状态写入到副本group上的runtimeStatus中的keyStatus，并关闭源group中Running的runtime   注意对于DeployCheck的任务，就不采用这种读取状态并写入的方式
//	// 注意：如果说Runtime本身没有细粒度控制的话，就不用再保存任务状态以及写入到副本任务上去
//	for i := range group.Spec.Actions {
//		action := &group.Spec.Actions[i]
//		actionStatus := group.Spec.Actions[i].Status
//		if actionStatus.Phase == apis.Successed {
//			continue
//		}
//		if actionStatus.Phase == apis.Running {
//			for j := range actionStatus.RuntimeStatus {
//				runtime := &action.Spec.Runtimes[j]
//				runtimeStatus := actionStatus.RuntimeStatus[j]
//				if runtimeStatus.Phase == apis.Successed {
//					continue
//				}
//				if runtimeStatus.Phase == apis.Running && action.Spec.Runtimes[j].EnableFineGrainedControl { // runtime正在运行，且runtime是细粒度控制的
//					//logs.Info("!!!!!!!!!!!!!!!!!!!!!!!!!")
//					data := mc.runtimeManager.StoreData(group, action, runtime, i, j) // 获取group下的正在执行runtime的关键数据
//					err := mc.runtimeManager.StopRuntime(group, action, runtime, i, j)
//					if err != nil {
//						logs.Errorf("Stop runtime error:%v", err)
//					}
//					// 将获取到的任务关键装填数据写入到本域的etcd上的副本group当中
//					patchGroup, err := json.Marshal([]map[string]interface{}{
//						{
//							"op":    "replace",
//							"path":  "/status/action_status/" + strconv.Itoa(i) + "/status/" + strconv.Itoa(j) + "/key_status",
//							"value": data, // 这里替换为你需要的 Phase 值
//						},
//					})
//					if err != nil {
//						logs.Errorf("Json Marshal failed, err:%v", err)
//					}
//					patchResult, err = mc.groupClient.Patch(context.TODO(), groupCopyName, types.JSONPatchType, patchGroup, metav1.PatchOptions{})
//					if err != nil {
//						logs.Errorf("Patch group error-5:%v", err)
//					}
//					//logs.Infof("*******副本任务runtimeStatus.keyStatus:%v", patchResult.Status.ActionStatus[i].RuntimeStatus[j].KeyStatus)
//					// 关闭源任务当中的runtime
//					logs.Info("^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^")
//					//err = sw.runtimeManager.StopRuntime(g, action, runtime, i, j)
//					time.Sleep(1 * time.Second)
//					err = mc.runtimeManager.Kill(group, action, runtime)
//					if err != nil {
//						logs.Errorf("Stop group error:%v", err)
//					}
//				}
//			}
//		}
//		// 将源group从Running队列迁移到Completed队列
//		logs.Info("--------------DeleteFromRunningAndAddToMigratedQueue=====================")
//		ok := mc.groupQueues.DeleteFromRunningAndAddToMigrated(group.Status.GroupID)
//		if !ok {
//			logs.Error("Delete group from running queue and add to completed queue failed-2")
//		}
//	}
//
//	return nil
//}
//
//// 新增一个创建一个空白的Group信息，删除不必要的内容（例如Running、DeployCheck的属性都得改为Unknown，时间也得修改）
//func NewGroupInfoCopy(g *apis.Group, isAhead bool) *apis.Group {
//	// 将原始对象序列化为JSON
//	data, err := json.Marshal(g)
//	if err != nil {
//		logs.Errorf("Marshal group:%v error:%v", g.Name, err)
//	}
//	// 反序列化为新的对象
//	var copyGroup apis.Group // 非指针
//	err = json.Unmarshal(data, &copyGroup)
//	if err != nil {
//		logs.Errorf("Unmarshal group:%v error:%v", g.Name, err)
//	}
//	var groupCopy = &copyGroup // 转换为指针
//	// 接下来修改这个复制出来的Group信息，首先修改group.Name
//	groupCopy.Name = "Reason-Copy"
//	// 将groupCopy的ResourceVersion置空，注意：这是必须的，否则报错
//	groupCopy.ResourceVersion = ""
//	// 修改GroupSpec下的Actions数组当中ActionStatus的Phase和time
//	groupCopy.Spec.IsCopy = true // 标记改Group为副本group
//	// 这个副本group信息当中，其副本数量直接置为0（意思是：不再为副本订制副本）
//	groupCopy.Spec.Replicas = 0
//	for i := range groupCopy.Spec.Actions {
//		action := &groupCopy.Spec.Actions[i]
//		if action.Status.Phase == apis.Successed {
//			continue
//		} else {
//			action.Status.Phase = apis.Unknown
//			action.Status.StartAt = apis.Time{time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC)}
//			action.Status.LastTime = apis.Time{time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC)}
//		}
//		for j := range action.Status.RuntimeStatus {
//			runtimeStatus := &action.Status.RuntimeStatus[j]
//			runtimeStatus.ProcessId = ""
//			if runtimeStatus.Phase != apis.Successed {
//				runtimeStatus.Phase = apis.Unknown
//				runtimeStatus.StartAt = apis.Time{time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC)}
//				runtimeStatus.LastTime = apis.Time{time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC)}
//			}
//		}
//
//	}
//	// 修改GroupStatus下面的phase、ActionStatus的Phase以及runtimeStatus的Phase
//	// 修改副本group信息中的属性来标记副本任务需要马上启动(这个属性会在copyPending队列当中去轮询检查的)
//	if isAhead {
//		logs.Infof("============预部署副本===========")
//		groupCopy.Status.CopyStatus = "Waiting" //注意后面真正切换的时候，需要将这个参数改为Starting
//	} else {
//		logs.Infof("============直接启动副本===========")
//		groupCopy.Status.CopyStatus = "Starting"
//	}
//	// 将groupStatus下的node属性置空
//	groupCopy.Status.Node = ""
//	if groupCopy.Status.Phase != apis.Successed {
//		groupCopy.Status.Phase = apis.Unknown
//	}
//	for i := range groupCopy.Status.ActionStatus {
//		actionStatus := &groupCopy.Status.ActionStatus[i]
//		if actionStatus.Phase == apis.Successed {
//			continue
//		} else {
//			actionStatus.Phase = apis.Unknown
//			actionStatus.StartAt = apis.Time{time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC)}
//			actionStatus.LastTime = apis.Time{time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC)}
//		}
//		for j := range actionStatus.RuntimeStatus {
//			runtimeStatus := &actionStatus.RuntimeStatus[j]
//			runtimeStatus.ProcessId = ""
//			if runtimeStatus.Phase != apis.Successed {
//				runtimeStatus.Phase = apis.Unknown
//				runtimeStatus.StartAt = apis.Time{time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC)}
//				runtimeStatus.LastTime = apis.Time{time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC)}
//			}
//		}
//	}
//	return groupCopy
//}

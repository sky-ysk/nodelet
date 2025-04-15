package main

import (
	"context"
	"hit.edu/framework/pkg/apimachinery/fields"
	"hit.edu/framework/pkg/apimachinery/runtime"
	"hit.edu/framework/pkg/apimachinery/runtime/schema"
	"hit.edu/framework/pkg/apimachinery/runtime/serializer"
	"hit.edu/framework/pkg/apimachinery/util/wait"
	apis "hit.edu/framework/pkg/apis/cores"
	metav1 "hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/client-go/clients"
	"hit.edu/framework/pkg/client-go/rest"
	"hit.edu/framework/pkg/client-go/tools/cache"
	"hit.edu/framework/pkg/client-go/util/workqueue"
	"hit.edu/framework/pkg/component-base/logs"
	"net/http"
	"time"
)

// 验证Watch功能
// 验证Store功能
// 验证Informer
// 验证Controller

// 应用自定义Controller,实现监控功能
// Controller demonstrates how to implement a controller with client-go.
type Controller struct {
	indexer  cache.Indexer
	queue    workqueue.TypedRateLimitingInterface[string]
	informer cache.Controller
}

// NewController creates a new Controller.
func NewController(queue workqueue.TypedRateLimitingInterface[string], indexer cache.Indexer, informer cache.Controller) *Controller {
	return &Controller{
		informer: informer,
		indexer:  indexer,
		queue:    queue,
	}
}

func (c *Controller) Run(workers int, stopCh chan struct{}) {
	// Let the workers stop when we are done
	defer c.queue.ShutDown()

	go c.informer.Run(stopCh)

	//轮询是否已经 同步缓存
	// Wait for all involved caches to be synced, before processing items from the queue is started
	if !cache.WaitForCacheSync(stopCh, c.informer.HasSynced) {
		logs.Infof("Timed out waiting for caches to sync")
		return
	}
	logs.Trace("缓存同步完成")

	//启动worker
	for i := 0; i < workers; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}
	<-stopCh
}

func (c *Controller) runWorker() {
	for c.processNextItem() {
	}
}

func (c *Controller) processNextItem() bool {
	// Wait until there is a new item in the working queue
	key, quit := c.queue.Get()
	if quit {
		return false
	}
	// Tell the queue that we are done with processing this key. This unblocks the key for other workers
	// This allows safe parallel processing because two pods with the same key are never processed in
	// parallel.
	defer c.queue.Done(key)

	// Invoke the method containing the business logic
	//调用包含业务逻辑的方法
	err := c.syncToStdout(key)
	if err != nil {
		logs.Info(err)
	}
	return true
}

// syncToStdout 是控制器的业务逻辑部分
// 在这个控制器中，它只是打印 有关workflow到stdout的信息
func (c *Controller) syncToStdout(key string) error {
	obj, exists, err := c.indexer.GetByKey(key)
	if err != nil {
		logs.Infof("Fetching object with key %s from store failed with %v", key, err)
		return err
	}

	if !exists {
		// Below we will warm up our cache with a Pod, so that we will see a delete for one pod
		logs.Infof("Source %s does not exist anymore\n", key)
	} else {
		// Note that you also have to check the uid if you have a local controlled resource, which
		// is dependent on the actual instance, to detect that a Pod was recreated with the same name
		logs.Infof("Sync/Add/Update for source:", obj)
	}
	return nil
}

func main() {
	logs.Init("workqueue-main")
	//注册资源
	scheme := runtime.NewScheme()
	apis.AddToScheme(scheme)
	logs.Trace(scheme)

	// 参数配置
	c := &rest.Config{
		Host:    "http://localhost:10000",
		APIPath: "/apis/resources/v1",
		ContentConfig: rest.ContentConfig{
			AcceptContentTypes: "application/json; charset=UTF-8", //text/plain; charset=UTF-8
			ContentType:        "application/json; charset=UTF-8", //application/json; charset=UTF-8
			GroupVersion: &schema.GroupVersion{
				Group:   "resources",
				Version: "v1",
			},
			NegotiatedSerializer: serializer.NewCodecFactory(scheme),
		},
		UserAgent: "defaultUserAgent",
		Transport: &http.Transport{
			MaxIdleConns:        100,              // 最大空闲连接数
			IdleConnTimeout:     90 * time.Second, // 空闲连接超时时间
			TLSHandshakeTimeout: 10 * time.Second, // TLS 握手超时时间
		},
		Timeout: 1000 * time.Second,
	}

	// 创建ClientSet
	clientSet, err := clients.NewForConfig(c)
	if err != nil {
		logs.Error(err)
	}

	workflowsClient := clientSet.Core().Workflows("Test")

	workflow := &apis.Workflow{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "demo-workflows",
			Namespace: "Test",
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Workflow",
			APIVersion: "resources/v1",
		},
		Spec: apis.WorkflowSpec{
			Name: "demo-workflow",
		},
	}

	workflow2 := &apis.Workflow{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "demo-workflow2",
			Namespace: "Test",
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Workflow",
			APIVersion: "resources/v1",
		},
		Spec: apis.WorkflowSpec{
			Name: "demo-workflow",
		},
	}
	workflow3 := &apis.Workflow{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "demo-workflow3",
			Namespace: "Test",
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Workflow",
			APIVersion: "resources/v1",
		},
		Spec: apis.WorkflowSpec{
			Name: "demo-workflow",
		},
	}

	//创建Workflow资源的List Watcher
	workflowListWatcher := cache.NewListWatchFromClient(clientSet.Core().RESTClient(), "workflows", "Test", fields.Everything())

	// 创建WorkQueue
	queue := workqueue.NewTypedRateLimitingQueue(workqueue.DefaultTypedControllerRateLimiter[string]())

	// 创建Indexer和Informer
	// 构造InformerOptions

	sourceEventHandler := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err == nil {
				queue.Add(key)
			}
		},
		UpdateFunc: func(old interface{}, new interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(new)
			if err == nil {
				queue.Add(key)
			}
		},
		DeleteFunc: func(obj interface{}) {
			// IndexerInformer uses a delta queue, therefore for deletes we have to use this
			// key function.
			key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			if err == nil {
				queue.Add(key)
			}
		},
	}

	options := cache.InformerOptions{
		ListerWatcher: workflowListWatcher,
		ObjectType:    &apis.Workflow{},
		Handler:       sourceEventHandler,
		ResyncPeriod:  0,
		Indexers:      cache.Indexers{},
	}

	indexer, informer := cache.NewInformerWithOptions(options)
	// 创建Controller
	controller := NewController(queue, indexer, informer)

	//设置Indexer对象格式
	indexer.Add(&apis.Workflow{
		ObjectMeta: metav1.ObjectMeta{
			Name: "demo-workflows",
		},
	})

	// 设置Indexer对象格式
	indexer.Add(workflow)

	// Now let's start the controller
	stop := make(chan struct{})
	defer close(stop)
	go controller.Run(1, stop)

	// Create一个Workflow
	logs.Trace("creating")
	_, err = workflowsClient.Create(context.TODO(), workflow, metav1.CreateOptions{})
	_, _ = workflowsClient.Create(context.TODO(), workflow2, metav1.CreateOptions{})
	_, _ = workflowsClient.Create(context.TODO(), workflow3, metav1.CreateOptions{})

	if err != nil {
		logs.Infof("Failed to create workflow: %v", err)
	}
	logs.Trace("Created workflow 1")

	//Update一个Workflow
	logs.Trace("updating workflow 1")
	// 部分更改一个参数
	// 先Get一个Workflow ,更改Workflow的参数, UpdateWorkflow
	result, getErr := workflowsClient.Get(context.TODO(), "demo-workflows", metav1.GetOptions{})
	if getErr != nil {
		logs.Errorf("Failed to get : %v", getErr)
	}

	result.Spec.Name = "updatedWorkflowName"
	_, updateErr := workflowsClient.Update(context.TODO(), result, metav1.UpdateOptions{})
	if updateErr != nil {
		logs.Infof("Update failed: %v", updateErr)
	}
	logs.Trace("1 workflow Updated workflow...")

	// List 所有Workflow
	logs.Trace("listing")
	lstOpts := metav1.ListOptions{}
	list, err := workflowsClient.List(context.TODO(), lstOpts)
	if err != nil {
		logs.Info(err)
	}
	for _, d := range list.Items {
		logs.Trace(d)
	}

	logs.Trace("listing done")

	// Delete一个Workflow
	// 删除Workflow后，Indexer就查询不到结点了
	logs.Trace("deleting")
	err = workflowsClient.Delete(context.TODO(), "demo-workflows", metav1.DeleteOptions{})
	if err != nil {
		logs.Info(err)
	}
	logs.Trace("Deleted workflow...")

	//为了验证功能，每5秒删一个Workflow
	time.Sleep(5 * time.Second)
	err = workflowsClient.Delete(context.TODO(), "demo-workflow2", metav1.DeleteOptions{})
	time.Sleep(5 * time.Second)
	err = workflowsClient.Delete(context.TODO(), "demo-workflow3", metav1.DeleteOptions{})

	// Wait forever
	select {}

}

package main

import (
	"fmt"
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
		panic(fmt.Errorf("Timed out waiting for caches to sync"))
		return
	}
	fmt.Println("缓存同步完成")

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
		panic(err)
	}
	return true
}

// syncToStdout 是控制器的业务逻辑部分
// 在这个控制器中，它只是打印 有关node到stdout的信息
func (c *Controller) syncToStdout(key string) error {
	obj, exists, err := c.indexer.GetByKey(key)
	if err != nil {
		fmt.Sprintf("Fetching object with key %s from store failed with %v", key, err)
		return err
	}

	if !exists {
		// Below we will warm up our cache with a Pod, so that we will see a delete for one pod
		fmt.Printf("Source %s does not exist anymore\n", key)
	} else {
		// Note that you also have to check the uid if you have a local controlled resource, which
		// is dependent on the actual instance, to detect that a Pod was recreated with the same name
		fmt.Println("Sync/Add/Update for source:", obj)
	}
	return nil
}

func main() {
	//注册资源
	scheme := runtime.NewScheme()
	apis.AddToScheme(scheme)
	fmt.Println(scheme)

	// 参数配置
	c := &rest.Config{
		Host:    "http://localhost:10000",
		APIPath: "/apis/resources/v1",
		ContentConfig: rest.ContentConfig{
			AcceptContentTypes: "application/json; charset=UTF-8", //text/plain; charset=UTF-8
			ContentType:        "application/json; charset=UTF-8", //application/json; charset=UTF-8
			GroupVersion: &schema.GroupVersion{
				Group:   "hit.edu",
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
		Timeout: 10 * time.Second,
	}

	// 创建ClientSet
	clientSet, err := clients.NewForConfig(c)
	if err != nil {
		panic(err)
	}

	//nodesClient := clientSet.Core().Nodes(apis.NamespaceDefault)
	//
	//node := &apis.Node{
	//	ObjectMeta: metav1.ObjectMeta{
	//		Name: "demo-nodes",
	//	},
	//	TypeMeta: runtime.TypeMeta{
	//		Kind:       "Node",
	//		APIVersion: "v1",
	//	},
	//	Spec: apis.NodeSpec{
	//		NodeName: "demo-node",
	//		HostName: "master",
	//	},
	//}

	//创建Node资源的List Watcher
	nodeListWatcher := cache.NewListWatchFromClient(clientSet.Core().RESTClient(), "nodes", apis.NamespaceDefault, fields.Everything())

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
		ListerWatcher: nodeListWatcher,
		ObjectType:    &apis.Node{},
		Handler:       sourceEventHandler,
		ResyncPeriod:  0,
		Indexers:      cache.Indexers{},
	}

	indexer, informer := cache.NewInformerWithOptions(options)
	// 创建Controller
	controller := NewController(queue, indexer, informer)

	//设置Indexer对象格式
	indexer.Add(&apis.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "demo-nodes",
		},
	})

	//// 设置Indexer对象格式
	//indexer.Add(node)

	// Now let's start the controller
	stop := make(chan struct{})
	workflowStop := make(chan struct{})
	defer close(stop)
	defer close(workflowStop)
	go controller.Run(1, stop)

	////分别对node 与 workflow 资源进行操作
	//// Create两个Node
	//fmt.Println("creating")
	//results, err := nodesClient.Create(context.TODO(), node, metav1.CreateOptions{})
	//if err != nil {
	//	panic(err)
	//}
	//fmt.Println("Created node 1:", results)
	//
	////Update一个Node
	//fmt.Println("updating node 1")
	//// 部分更改一个参数
	//// 先Get一个Node ,更改Node的参数, UpdateNode
	//result, getErr := nodesClient.Get(context.TODO(), "demo-nodes", metav1.GetOptions{})
	//fmt.Println("node get:", result)
	//if getErr != nil {
	//	panic(fmt.Errorf("Failed to get : %v", getErr))
	//}
	//
	//result.Spec.NodeName = "updatedNodeName"
	//_, updateErr := nodesClient.Update(context.TODO(), result, metav1.UpdateOptions{})
	//if updateErr != nil {
	//	panic(fmt.Errorf("Update failed: %v", updateErr))
	//}
	//fmt.Println("1 node Updated node...")
	//
	//// Delete一个Node
	//// 删除Node后，Indexer就查询不到结点了
	//fmt.Println("deleting")
	//err = nodesClient.Delete(context.TODO(), "demo-nodes", metav1.DeleteOptions{})
	//if err != nil {
	//	panic(err)
	//}
	//fmt.Println("Deleted node...")

	// Wait forever
	select {}
}

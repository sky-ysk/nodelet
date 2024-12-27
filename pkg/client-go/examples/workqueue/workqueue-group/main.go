package main

import (
	"context"
	"encoding/json"
	"fmt"
	"hit.edu/framework/pkg/apimachinery/fields"
	"hit.edu/framework/pkg/apimachinery/runtime"
	"hit.edu/framework/pkg/apimachinery/runtime/schema"
	"hit.edu/framework/pkg/apimachinery/runtime/serializer"
	"hit.edu/framework/pkg/apimachinery/util/wait"
	"hit.edu/framework/pkg/apimachinery/watch"
	apis "hit.edu/framework/pkg/apis/cores"
	metav1 "hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/client-go/clients"
	"hit.edu/framework/pkg/client-go/rest"
	"hit.edu/framework/pkg/client-go/tools/cache"
	"hit.edu/framework/pkg/client-go/util/workqueue"
	"net/http"
	"strings"
	"sync"
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
// 在这个控制器中，它只是打印 有关group到stdout的信息
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
	//启动模拟 HTTP 服务器
	go func() {
		server := createMockAPIServer()
		fmt.Println("Starting mock API server on :8080...")
		if err := server.ListenAndServe(); err != nil {
			fmt.Printf("Failed to start mock API server: %v\n", err)
		}
	}()
	time.Sleep(1 * time.Second) // 等待服务器启动
	
	//注册资源
	scheme := runtime.NewScheme()
	apis.AddToScheme(scheme)
	fmt.Println(scheme)
	
	// 参数配置
	c := &rest.Config{
		Host:    "http://localhost:8080",
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
	
	groupsClient := clientSet.Core().Groups(apis.NamespaceDefault)
	
	group := &apis.Group{
		ObjectMeta: metav1.ObjectMeta{
			Name: "demo-groups",
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Group",
			APIVersion: "v1",
		},
		Spec: apis.GroupSpec{
			Name: "demo-group",
		},
	}
	// 创建一个 watch.Bookmark 事件
	bookmarkEvent := watch.Event{
		Type:   watch.Bookmark, // 标记事件类型为 Bookmark
		Object: nil,
	}
	//创建Group资源的List Watcher
	groupListWatcher := cache.NewListWatchFromClient(clientSet.Core().RESTClient(), "groups", apis.NamespaceDefault, fields.Everything())
	
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
		ListerWatcher: groupListWatcher,
		ObjectType:    &apis.Group{},
		Handler:       sourceEventHandler,
		ResyncPeriod:  0,
		Indexers:      cache.Indexers{},
	}
	// TODO: 设置Watch的对象
	
	indexer, informer := cache.NewInformerWithOptions(options)
	// 创建Controller
	controller := NewController(queue, indexer, informer)
	
	//设置Indexer对象格式
	indexer.Add(&apis.Group{
		ObjectMeta: metav1.ObjectMeta{
			Name: "demo-groups",
		},
	})
	
	//// 设置Indexer对象格式
	//indexer.Add(group)
	
	// Now let's start the controller
	stop := make(chan struct{})
	defer close(stop)
	go controller.Run(1, stop)
	
	//对group资源进行操作
	// Create两个Group
	fmt.Println("creating")
	results, err := groupsClient.Create(context.TODO(), group, metav1.CreateOptions{})
	if err != nil {
		panic(err)
	}
	fmt.Println("Created group 1:", results)
	
	//Update一个Group
	fmt.Println("updating group 1")
	// 部分更改一个参数
	// 先Get一个Group ,更改Group的参数, UpdateGroup
	result, getErr := groupsClient.Get(context.TODO(), "demo-groups", metav1.GetOptions{})
	fmt.Println("group get:", result)
	if getErr != nil {
		panic(fmt.Errorf("Failed to get : %v", getErr))
	}
	
	result.Spec.Name = "updatedGroupName"
	_, updateErr := groupsClient.Update(context.TODO(), result, metav1.UpdateOptions{})
	if updateErr != nil {
		panic(fmt.Errorf("Update failed: %v", updateErr))
	}
	fmt.Println("1 group Updated group...")
	
	// Delete一个Group
	// 删除Group后，Indexer就查询不到结点了
	//fmt.Println("deleting")
	//err = groupsClient.Delete(context.TODO(), "demo-groups", meta.DeleteOptions{})
	//if err != nil {
	//	panic(err)
	//}
	//fmt.Println("Deleted group...")
	
	// 将事件发送到 ResultChan
	notifyWatchers(bookmarkEvent)
	
	//Wait 4s
	//time.Sleep(4 * time.Second)
	
	// Wait forever
	select {}
}

var (
	mu sync.Mutex // 用于保护 watchChans 的并发访问
)

func createMockAPIServer() *http.Server {
	mux := http.NewServeMux()
	
	// 模拟存储节点的内存数据库
	groups := make(map[string]apis.Group)
	
	// 处理group的集合操作（POST 创建,List 和 Watch）
	mux.HandleFunc("/apis/resources/v1/defaultNamespace/groups", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		iswatch := query.Get("watch")
		fieldSelector := query.Get("fieldSelector")
		switch r.Method {
		case http.MethodGet:
			if iswatch == "true" { // 判断是否是Watch请求
				// 解析 fieldSelector 并筛选节点
				filteredGroups := make([]apis.Group, 0)
				for _, group := range groups {
					if groupMatchesFieldSelector(group, fieldSelector) {
						filteredGroups = append(filteredGroups, group)
					}
				}
				watchChan := make(chan watch.Event)
				mu.Lock()
				watchChans = append(watchChans, watchChan)
				mu.Unlock()
				
				// 设置响应头以支持流式传输
				w.Header().Set("Content-Type", "application/json; charset=UTF-8")
				w.WriteHeader(http.StatusOK)
				
				// 持续监听通道，发送事件
				encoder := json.NewEncoder(w)
				for event := range watchChan {
					if err := encoder.Encode(event); err != nil {
						fmt.Printf("Error encoding watch event: %v\n", err)
						return
					}
					w.(http.Flusher).Flush()           // 确保事件及时推送到客户端
					time.Sleep(500 * time.Millisecond) // 模拟一定的延迟
				}
				return
			} else { //否则为List请求
				// 解析 fieldSelector 并筛选节点
				filteredGroups := make([]apis.Group, 0)
				for _, group := range groups {
					if groupMatchesFieldSelector(group, fieldSelector) {
						filteredGroups = append(filteredGroups, group)
					}
				}
				
				// 返回筛选后的节点列表
				w.Header().Set("Content-Type", "application/json; charset=UTF-8")
				w.WriteHeader(http.StatusOK)
				if err := json.NewEncoder(w).Encode(filteredGroups); err != nil {
					http.Error(w, "Error encoding response", http.StatusInternalServerError)
				}
			}
			// 处理 GET 请求
		case http.MethodPost:
			// 检查 Content-Type 是否是 application/json; charset=UTF-8
			if r.Header.Get("Content-Type") != "application/json; charset=UTF-8" {
				http.Error(w, "Unsupported Content-Type", http.StatusUnsupportedMediaType)
				return
			}
			
			// 解析请求体中的 Group 数据
			newGroup := &apis.Group{}
			if err := json.NewDecoder(r.Body).Decode(&newGroup); err != nil {
				http.Error(w, "Invalid request body", http.StatusBadRequest)
				return
			}
			
			// 模拟存储节点
			groups[newGroup.ObjectMeta.Name] = *newGroup
			
			// 推送 Watch 事件
			fmt.Println("newGroup:", newGroup)
			event := watch.Event{
				Type:   "ADDED",
				Object: newGroup,
			}
			notifyWatchers(event)
			
			// 返回创建成功的响应
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.WriteHeader(http.StatusCreated) // 状态码 201 Created
			if err := json.NewEncoder(w).Encode(newGroup); err != nil {
				http.Error(w, "Error encoding response", http.StatusInternalServerError)
			}
		default:
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	})
	// 处理group的单个操作（单个的GET 查询和 PUT 更新）
	mux.HandleFunc("/apis/resources/v1/defaultNamespace/groups/demo-groups", func(w http.ResponseWriter, r *http.Request) {
		groupName := "demo-groups" // 固定为 demo-groups
		
		switch r.Method {
		case http.MethodGet: // GET 查询
			// 查询内存数据库中的节点
			group, exists := groups[groupName]
			if !exists {
				http.Error(w, "Group not found", http.StatusNotFound)
				return
			}
			
			// 返回节点信息
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.WriteHeader(http.StatusOK) // 状态码 200 OK
			if err := json.NewEncoder(w).Encode(group); err != nil {
				http.Error(w, "Error encoding response", http.StatusInternalServerError)
			}
		
		case http.MethodPut: // PUT 更新
			// 检查 Content-Type 是否是 application/json; charset=UTF-8
			if r.Header.Get("Content-Type") != "application/json; charset=UTF-8" {
				http.Error(w, "Unsupported Content-Type", http.StatusUnsupportedMediaType)
				return
			}
			
			// 解析请求体中的 Group 数据
			updatedGroup := &apis.Group{}
			if err := json.NewDecoder(r.Body).Decode(&updatedGroup); err != nil {
				http.Error(w, "Invalid request body", http.StatusBadRequest)
				return
			}
			
			// 检查节点是否存在
			_, exists := groups[groupName]
			if !exists {
				http.Error(w, "Group not found", http.StatusNotFound)
				return
			}
			
			// 更新节点信息
			groups[groupName] = *updatedGroup
			
			// 推送 Watch 事件：MODIFIED
			event := watch.Event{
				Type:   "MODIFIED",
				Object: updatedGroup,
			}
			notifyWatchers(event)
			
			// 返回更新成功的响应
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.WriteHeader(http.StatusOK) // 状态码 200 OK
			if err := json.NewEncoder(w).Encode(updatedGroup); err != nil {
				http.Error(w, "Error encoding response", http.StatusInternalServerError)
			}
		case http.MethodDelete: // 删除节点
			deletegroup, exists := groups[groupName]
			if !exists {
				http.Error(w, "Group not found", http.StatusNotFound)
				return
			}
			deletedGroup := &deletegroup
			// 删除节点
			delete(groups, groupName)
			
			// 推送 Watch 事件：DELETED
			event := watch.Event{
				Type:   "DELETED",
				Object: deletedGroup,
			}
			notifyWatchers(event)
			
			// 返回删除成功的响应
			w.WriteHeader(http.StatusNoContent) // 状态码 204 No Content
		
		default:
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
		
	})
	
	// 创建一个 HTTP 服务器
	server := &http.Server{
		Addr:    ":8080", // 监听端口 8080
		Handler: mux,
	}
	
	return server
}

func groupMatchesFieldSelector(group apis.Group, fieldSelector string) bool {
	if fieldSelector == "" {
		return true // 如果没有指定 fieldSelector，匹配所有节点
	}
	
	// 示例：支持解析 "ObjectMeta.Name=demo-groups" 的 fieldSelector
	parts := strings.Split(fieldSelector, "=")
	if len(parts) != 2 {
		return false
	}
	
	key, value := parts[0], parts[1]
	switch key {
	case "ObjectMeta.Name":
		return group.ObjectMeta.Name == value
	// 可扩展其他字段匹配
	default:
		return false
	}
}

var (
	groups     = make(map[string]apis.Group) // 模拟存储节点的内存数据库
	watchChans = make([]chan watch.Event, 0) // 维护所有watch监听的通道
)

// 推送事件
func notifyWatchers(event watch.Event) {
	for _, ch := range watchChans {
		fmt.Println("推送的event:", event)
		ch <- event
	}
}

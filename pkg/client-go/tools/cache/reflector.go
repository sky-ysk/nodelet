package cache

import (
	"fmt"
	"hit.edu/framework/pkg/api/meta"
	"hit.edu/framework/pkg/apimachinery/runtime"
	"hit.edu/framework/pkg/apimachinery/runtime/schema"
	"hit.edu/framework/pkg/apimachinery/util/wait"
	"hit.edu/framework/pkg/apimachinery/watch"
	apis "hit.edu/framework/pkg/apis/cores"
	metav1 "hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/component-base/logs"
	"k8s.io/utils/clock"
	"k8s.io/utils/pointer"
	"k8s.io/utils/ptr"
	"reflect"
	"sync"
	"time"
)

var (
	// We try to spread the load on apiserver by setting timeouts for
	// watch requests - it is random in [minWatchTimeout, 2*minWatchTimeout].
	defaultMinWatchTimeout = 5 * time.Minute
)

// 与API Server建立连接，检查数据是否产生变化
// 将数据存到本地FIFO中
// 或将数据写入API Server的ETCD后端
type Reflector struct {
	//
	name string

	// An example object of the type we expect to place in the store.
	// Only the type needs to be right, except that when that is
	// `unstructured.Unstructured` the object's `"apiVersion"` and
	// `"kind"` must also be right.
	expectedType reflect.Type
	// The GVK of the object we expect to place in the store if unstructured.
	expectedGVK *schema.GroupVersionKind

	// 数据的存储队列，本地DeltaFIFO
	store Store

	// 用于执行lists和watches
	listerWatcher ListerWatcher

	// backoff manages backoff of ListWatch
	backoffManager wait.BackoffManager

	// ShouldResync is invoked periodically and whenever it returns `true` the Store's Resync operation is invoked
	ShouldResync func() bool

	resyncPeriod time.Duration
	// minWatchTimeout defines the minimum timeout for watch requests.

	// lastSyncResourceVersionMutex guards read/write access to lastSyncResourceVersion
	lastSyncResourceVersion              string
	lastSyncResourceVersionMutex         sync.RWMutex
	isLastSyncResourceVersionUnavailable bool

	UseWatchList *bool

	// clock allows tests to manipulate time
	clock clock.Clock
}

// 构造器
func NewReflector() *Reflector {
	return &Reflector{}
}

// TODO: Reflector名称初始化

// ReflectorOptions configures a Reflector.
type ReflectorOptions struct {
	// Name is the Reflector's name. If unset/unspecified, the name defaults to the closest source_file.go:line
	// in the call stack that is outside this package.
	Name string

	//同步周期
	ResyncPeriod time.Duration

	//// MinWatchTimeout, if non-zero, defines the minimum timeout for watch requests send to kube-apiserver.
	//// However, values lower than 5m will not be honored to avoid negative performance impact on controlplane.
	//MinWatchTimeout time.Duration
}

func NewReflectorWithOptions(lw ListerWatcher, expectedType interface{}, store Store, options ReflectorOptions) *Reflector {
	reflectorClock := clock.RealClock{}
	//测试watchList时加的变量useWatchList
	useWatchList := true
	r := &Reflector{
		name:           options.Name,
		listerWatcher:  lw,
		store:          store,
		backoffManager: wait.NewExponentialBackoffManager(800*time.Millisecond, 30*time.Second, 2*time.Minute, 2.0, 1.0, reflectorClock),
		expectedType:   reflect.TypeOf(expectedType),
		UseWatchList:   &useWatchList,
	}
	return r
}

// Reflector运行逻辑
func (r *Reflector) Run(stopCh <-chan struct{}) {
	// List And Watch
	//对于长时间未响应的数据，或者频繁请求的数据，通过指数退避算法将数据后移
	wait.BackoffUntil(func() {
		if err := r.ListAndWatch(stopCh); err != nil {
			logs.Info(err)
		}
	}, r.backoffManager, true, stopCh)
}

// 首先List所有的对象，然后Watch数据的变化
func (r *Reflector) ListAndWatch(stopCh <-chan struct{}) error {
	var err error
	var w watch.Interface
	useWatchList := ptr.Deref(r.UseWatchList, false)
	fallbackToList := !useWatchList

	if useWatchList {
		w, err = r.watchList(stopCh)
		if w == nil && err == nil {
			// stopCh was closed
			return nil
		}
		if err != nil {
			logs.Errorf("The watchlist request ended with an error, falling back to the standard LIST/WATCH semantics because making progress is better than deadlocking, err = %v", err)
			fallbackToList = true
			// ensure that we won't accidentally pass some garbage down the watch.
			w = nil
		}
	}

	if fallbackToList {
		err = r.list(stopCh)
		if err != nil {
			return err
		}
	}

	//todo：这里会再返回一个watchWithResync，确保本地缓存实时更新
	//目前上面的watchList方法正在监听，所以第二次的watch可能建立不了连接
	//return r.watchWithResync(w, stopCh)
	return nil
}

// watchWithResync runs watch with startResync in the background.
func (r *Reflector) watchWithResync(w watch.Interface, stopCh <-chan struct{}) error {
	// 同步数据
	resyncerrc := make(chan error, 1)
	cancelCh := make(chan struct{})
	defer close(cancelCh)
	go r.startResync(stopCh, cancelCh, resyncerrc)
	return r.watch(w, stopCh)
}

// startResync periodically calls r.store.Resync() method.
// Note that this method is blocking and should be
// called in a separate goroutine.
func (r *Reflector) startResync(stopCh <-chan struct{}, cancelCh <-chan struct{}, resyncerrc chan error) {
	resyncCh, cleanup := r.resyncChan()
	defer func() {
		cleanup() // Call the last one written into cleanup
	}()
	for {
		select {
		case <-resyncCh:
		case <-stopCh:
			return
		case <-cancelCh:
			return
		}
		if r.ShouldResync == nil || r.ShouldResync() {
			if err := r.store.Resync(); err != nil {
				resyncerrc <- err
				return
			}
		}
		cleanup()
		resyncCh, cleanup = r.resyncChan()
	}
}

var (
	// nothing will ever be sent down this channel
	neverExitWatch <-chan time.Time = make(chan time.Time)
)

// resyncChan returns a channel which will receive something when a resync is
// required, and a cleanup function.
func (r *Reflector) resyncChan() (<-chan time.Time, func() bool) {
	if r.resyncPeriod == 0 {
		return neverExitWatch, func() bool { return false }
	}
	// The cleanup function is required: imagine the scenario where watches
	// always fail so we end up listing frequently. Then, if we don't
	// manually stop the timer, we could end up with many timers active
	// concurrently.
	t := r.clock.NewTimer(r.resyncPeriod)
	return t.C(), t.Stop
}

// 向API Server发起Watch请求
func (r *Reflector) watch(w watch.Interface, stopCh <-chan struct{}) error {
	for {
		// 监听停止信号
		select {
		case <-stopCh:
			if w != nil {
				w.Stop()
			}
			return nil
		default:
		}

		// 如果当前没有活跃的 Watch 请求，启动新的 Watch
		if w == nil {
			timeoutSeconds := int64(100)
			options := metav1.ListOptions{
				ResourceVersion: r.LastSyncResourceVersion(),
				// We want to avoid situations of hanging watchers. Stop any watchers that do not
				// receive any events within the timeout window.
				TimeoutSeconds: &timeoutSeconds,
				// To reduce load on kube-apiserver on watch restarts, you may enable watch bookmarks.
				// Reflector doesn't assume bookmarks are returned at all (if the server do not support
				// watch bookmarks, it will ignore this field).
				AllowWatchBookmarks: true,
			}

			var err error
			w, err = r.listerWatcher.Watch(options)
			if err != nil {
				return err
			}
		}

		// 处理 Watch 返回的事件流
		err := handleWatch(w, r.store, r.expectedType, r.expectedGVK, r.setLastSyncResourceVersion,
			stopCh)
		if err != nil {
			w.Stop()
			w = nil
			return err
		}
		// 停止当前 Watch，以确保每次都重新创建 Watch
		w.Stop()
		w = nil
	}
}

func (r *Reflector) setLastSyncResourceVersion(v string) {
	r.lastSyncResourceVersionMutex.Lock()
	defer r.lastSyncResourceVersionMutex.Unlock()
	r.lastSyncResourceVersion = v
}

// List所有变量
// list只是列出所有项目，并记录调用时从服务器获得的资源版本。资源版本可用于进一步的 watch。
func (r *Reflector) list(stopCh <-chan struct{}) error {
	var resourceVersion string
	options := metav1.ListOptions{
		ResourceVersion: r.relistResourceVersion(),
	}

	list, err := r.listerWatcher.List(options)
	if err != nil {
		logs.Infof("%s: failed to list: %v", r.name, err)
		return err
	}

	//todo:换成通用的
	// 类型断言为 *NodeList
	nodeList, ok := list.(*apis.NodeList)
	if !ok {
		return fmt.Errorf("unexpected list type %T, expected *NodeList", list)
	}
	// 直接使用 NodeList.Items 转换为 runtime.Object 列表
	items := make([]runtime.Object, len(nodeList.Items))
	for i, node := range nodeList.Items {
		items[i] = node.DeepCopyObject() // 确保每个对象是独立的拷贝
	}

	r.setIsLastSyncResourceVersionUnavailable(false) // list was successful
	//listMetaInterface, err := meta.ListAccessor(list)
	//resourceVersion = listMetaInterface.GetResourceVersion()
	if err := r.syncWith(items, resourceVersion); err != nil {
		return fmt.Errorf("unable to sync list result: %v", err)
	}
	//r.setLastSyncResourceVersion(resourceVersion)

	return nil
}

// syncWith replaces the store's items with the given list.
func (r *Reflector) syncWith(items []runtime.Object, resourceVersion string) error {
	found := make([]interface{}, 0, len(items))
	for _, item := range items {
		found = append(found, item)
	}
	return r.store.Replace(found, resourceVersion)
}

//多层调用有bug，将函数嵌入wachList中
//func (r *Reflector) watchList(stopCh <-chan struct{}) (watch.Interface, error) {
//	var w watch.Interface
//	var err error
//	var temporaryStore Store
//	var resourceVersion string
//
//	// 实现指数回退机制
//	backoff := 1 * time.Second     // 初始等待时间
//	maxBackoff := 30 * time.Second // 最大等待时间
//	retryCount := 0
//
//	for {
//		//这里的resourceVersion 并不是api版本（如resources/v1）
//		resourceVersion = ""
//		lastKnownRV := r.rewatchResourceVersion()
//
//		select {
//		case <-stopCh:
//			return nil, nil
//		default:
//		}
//
//		//这里先用写入temporaryStore，然后通过replace到deltafifo 并且计数器加1，在pop到时会-1，以此判断同步是否完成
//		temporaryStore = NewStore(DeletionHandlingMetaNamespaceKeyFunc)
//		//timeoutSeconds := int64(defaultMinWatchTimeout.Seconds())
//		timeoutSeconds := int64(5)
//		//options 用来控制和定制如何进行资源监听
//		options := metav1.ListOptions{
//			ResourceVersion:      lastKnownRV,
//			AllowWatchBookmarks:  true,
//			SendInitialEvents:    pointer.Bool(true),
//			ResourceVersionMatch: metav1.ResourceVersionMatchNotOlderThan,
//			TimeoutSeconds:       &timeoutSeconds,
//		}
//
//		w, err = r.listerWatcher.Watch(options)
//
//		if err != nil {
//			// 如果请求失败，执行回退
//			time.Sleep(backoff)
//			// 使用指数回退增加等待时间
//			backoff = time.Duration(float64(backoff) * 2.0) // 指数回退
//			if backoff > maxBackoff {
//				backoff = maxBackoff // 达到最大等待时间
//			}
//			retryCount++
//			continue
//		}
//
//		// 重置回退时间
//		backoff = 1 * time.Second
//		retryCount = 0
//
//		watchListBookmarkReceived, err := handleListWatch(
//			w, temporaryStore, r.expectedType, r.expectedGVK,
//			func(rv string) { resourceVersion = rv },
//			stopCh,
//		)
//		if err != nil {
//			w.Stop()
//			continue
//		}
//		if watchListBookmarkReceived {
//			break
//		}
//	}
//	r.setIsLastSyncResourceVersionUnavailable(false)
//
//	if err := r.store.Replace(temporaryStore.List(), resourceVersion); err != nil {
//		logs.Infof("failed to replace temporary store:", err)
//		return nil, fmt.Errorf("unable to sync watch-list result: %w", err)
//	}
//
//	r.setLastSyncResourceVersion(resourceVersion)
//	return w, nil
//}

// 多层调用有bug，将函数嵌入wachList中
func (r *Reflector) watchList(stopCh <-chan struct{}) (watch.Interface, error) {
	var w watch.Interface
	var temporaryStore Store
	var resourceVersion string

	for {
		resourceVersion = ""
		lastKnownRV := r.rewatchResourceVersion()

		select {
		case <-stopCh:
			return nil, nil
		default:
		}

		temporaryStore = NewStore(DeletionHandlingMetaNamespaceKeyFunc)
		timeoutSeconds := int64(10000)
		options := metav1.ListOptions{
			ResourceVersion:      lastKnownRV,
			AllowWatchBookmarks:  true,
			SendInitialEvents:    pointer.Bool(true),
			ResourceVersionMatch: metav1.ResourceVersionMatchNotOlderThan,
			TimeoutSeconds:       &timeoutSeconds,
		}

		w, err := r.listerWatcher.Watch(options)
		if err != nil {
			fmt.Printf("err: %v", err)
			w.Stop()
			continue
		}

		watchListBookmarkReceived := false
	loop:
		for {
			select {
			case <-stopCh:
				return nil, nil
			case event, ok := <-w.ResultChan():
				if !ok {
					break loop
				}
				if event.Type == watch.Error {
					logs.Info("event.Type == watch.Error")
				}
				if r.expectedType != nil && reflect.TypeOf(event.Object) != nil && r.expectedType != reflect.TypeOf(event.Object) {
					logs.Info("类型验证不匹配的事件")
					continue
				}
				if r.expectedGVK != nil && *r.expectedGVK != event.Object.GetObjectKind().GroupVersionKind() {
					logs.Info("GVK验证不匹配的事件")
					continue
				}
				metaObj, err := meta.Accessor(event.Object)
				if err != nil {
					panic(fmt.Errorf("%s: unable to understand watch event %#v", event))
				}
				resourceVersion = metaObj.GetResourceVersion()

				switch event.Type {
				case watch.Added:
					if err := temporaryStore.Add(event.Object); err != nil {
						logs.Errorf("unable to add watch event object: %#v", event.Object)
					}
				case watch.Modified:
					if err := temporaryStore.Update(event.Object); err != nil {
						logs.Errorf("unable to update watch event object: %#v", event.Object)
					}
				case watch.Deleted:
					if err := temporaryStore.Delete(event.Object); err != nil {
						logs.Errorf("unable to delete watch event object: %#v", event.Object)
					}
				case watch.Bookmark:
					if metaObj.GetAnnotations()[metav1.InitialEventsAnnotationKey] == "true" {
						watchListBookmarkReceived = true
					}
					watchListBookmarkReceived = true
					if err := r.store.Replace(temporaryStore.List(), resourceVersion); err != nil {
						logs.Infof("failed to replace temporary store: %v", err)
						return nil, fmt.Errorf("unable to sync watch - list result: %w", err)
					}
				default:
					logs.Errorf("unknown watch event: %#v", event)
				}
				func(rv string) { resourceVersion = rv }(resourceVersion)
			}
		}
		if watchListBookmarkReceived {
			break
		}
	}
	r.setIsLastSyncResourceVersionUnavailable(false)

	r.setLastSyncResourceVersion(resourceVersion)
	return w, nil
}

// handleListWatch consumes events from w, updates the Store, and records the
// last seen ResourceVersion, to allow continuing from that ResourceVersion on
// retry. If successful, the watcher will be left open after receiving the
// initial set of objects, to allow watching for future events.
func handleListWatch(
	w watch.Interface,
	store Store,
	expectedType reflect.Type,
	expectedGVK *schema.GroupVersionKind,
	setLastSyncResourceVersion func(string),
	stopCh <-chan struct{},
) (bool, error) {
	return handleAnyWatch(w, store, expectedType, expectedGVK,
		setLastSyncResourceVersion, stopCh)
}

// handleListWatch consumes events from w, updates the Store, and records the
// last seen ResourceVersion, to allow continuing from that ResourceVersion on
// retry. The watcher will always be stopped on exit.
func handleWatch(
	w watch.Interface,
	store Store,
	expectedType reflect.Type,
	expectedGVK *schema.GroupVersionKind,
	setLastSyncResourceVersion func(string),
	stopCh <-chan struct{},
) error {
	_, err := handleAnyWatch(w, store, expectedType, expectedGVK,
		setLastSyncResourceVersion, stopCh)
	return err
}

func handleAnyWatch(
	w watch.Interface,
	store Store,
	expectedType reflect.Type,
	expectedGVK *schema.GroupVersionKind,
	setLastSyncResourceVersion func(string),
	stopCh <-chan struct{},
) (bool, error) {
	watchListBookmarkReceived := false

loop:
	for {
		select {
		case <-stopCh:
			return false, nil // 停止信号，退出处理
		case event, ok := <-w.ResultChan():
			if !ok {
				break loop
			}
			// 错误事件直接引发 panic
			if event.Type == watch.Error {
				logs.Info("event.Type == watch.Error")
			}
			// 类型验证
			if expectedType != nil && reflect.TypeOf(event.Object) != nil && expectedType != reflect.TypeOf(event.Object) {
				logs.Info("类型验证不匹配的事件")
				continue // 跳过不匹配的事件
			}
			// GVK 验证
			if expectedGVK != nil && *expectedGVK != event.Object.GetObjectKind().GroupVersionKind() {
				logs.Info("GVK验证不匹配的事件")
				continue // 跳过不匹配的事件
			}
			meta, err := meta.Accessor(event.Object)
			if err != nil {
				panic(fmt.Errorf("%s: unable to understand watch event %#v", event))
			}
			resourceVersion := meta.GetResourceVersion()

			// 根据事件类型处理
			switch event.Type {
			case watch.Added:
				if err := store.Add(event.Object); err != nil {
					logs.Errorf("unable to add watch event object: %#v", event.Object)
				}
			case watch.Modified:
				if err := store.Update(event.Object); err != nil {
					logs.Errorf("unable to update watch event object: %#v", event.Object)
				}
			case watch.Deleted:
				// TODO: Will any consumers need access to the "last known
				// state", which is passed in event.Object? If so, may need
				// to change this.
				if err := store.Delete(event.Object); err != nil {
					logs.Errorf("unable to delete watch event object: %#v", event.Object)
				}
			case watch.Bookmark:
				// A `Bookmark` means watch has synced here, just update the resourceVersion
				if meta.GetAnnotations()[metav1.InitialEventsAnnotationKey] == "true" {
					watchListBookmarkReceived = true
				}
				watchListBookmarkReceived = true
			default:
				logs.Errorf("unknown watch event: %#v", event)
			}
			setLastSyncResourceVersion(resourceVersion)
			if watchListBookmarkReceived {
				return watchListBookmarkReceived, nil
			}
		}
	}
	return watchListBookmarkReceived, nil
}

// LastSyncResourceVersion is the resource version observed when last sync with the underlying store
// The value returned is not synchronized with access to the underlying store and is not thread-safe
func (r *Reflector) LastSyncResourceVersion() string {
	r.lastSyncResourceVersionMutex.RLock()
	defer r.lastSyncResourceVersionMutex.RUnlock()
	return r.lastSyncResourceVersion
}

// rewatchResourceVersion determines the resource version the reflector should start streaming from.
func (r *Reflector) rewatchResourceVersion() string {
	r.lastSyncResourceVersionMutex.RLock()
	defer r.lastSyncResourceVersionMutex.RUnlock()
	if r.isLastSyncResourceVersionUnavailable {
		// initial stream should return data at the most recent resource version.
		// the returned data must be consistent i.e. as if served from etcd via a quorum read
		return ""
	}
	return r.lastSyncResourceVersion
}
func (r *Reflector) setIsLastSyncResourceVersionUnavailable(isUnavailable bool) {
	r.lastSyncResourceVersionMutex.Lock()
	defer r.lastSyncResourceVersionMutex.Unlock()
	r.isLastSyncResourceVersionUnavailable = isUnavailable
}

func (r *Reflector) relistResourceVersion() string {
	r.lastSyncResourceVersionMutex.RLock()
	defer r.lastSyncResourceVersionMutex.RUnlock()

	if r.isLastSyncResourceVersionUnavailable {
		// Since this reflector makes paginated list requests, and all paginated list requests skip the watch cache
		// if the lastSyncResourceVersion is unavailable, we set ResourceVersion="" and list again to re-establish reflector
		// to the latest available ResourceVersion, using a consistent read from etcd.
		return ""
	}
	if r.lastSyncResourceVersion == "" {
		// For performance reasons, initial list performed by reflector uses "0" as resource version to allow it to
		// be served from the watch cache if it is enabled.
		return "0"
	}
	return r.lastSyncResourceVersion
}

type WatchErrorHandler func(r *Reflector, err error)

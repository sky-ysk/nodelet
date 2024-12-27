package cache

import (
	"fmt"
	"hit.edu/framework/pkg/apimachinery/runtime"
	"hit.edu/framework/pkg/apimachinery/util/wait"
	"k8s.io/utils/clock"
	"sync"
	"time"
)

// Controller
// 监测数据变化，产生事件交由EventHandler处理

type Config struct {
	// The queue for your objects - has to be a DeltaFIFO due to
	// assumptions in the implementation. Your Process() function
	// should accept the output of this Queue's Pop() method.
	Queue

	// Something that can list and watch your objects.
	ListerWatcher
	ObjectType interface{}
	// Something that can process a popped Deltas.
	Process ProcessFunc
}

// ProcessFunc processes a single object.
type ProcessFunc func(obj interface{}, isInInitialList bool) error

type controller struct {
	config         Config
	reflector      *Reflector
	reflectorMutex sync.RWMutex
	clock          clock.Clock
}

// From K8s
// Controller is a low-level controller that is parameterized by a
// Config and used in sharedIndexInformer.
type Controller interface {
	// Run does two things.  One is to construct and run a Reflector
	// to pump objects/notifications from the Config's ListerWatcher
	// to the Config's Queue and possibly invoke the occasional Resync
	// on that Queue.  The other is to repeatedly Pop from the Queue
	// and process with the Config's ProcessFunc.  Both of these
	// continue until `stopCh` is closed.
	Run(stopCh <-chan struct{})

	// HasSynced delegates to the Config's Queue
	HasSynced() bool
	// LastSyncResourceVersion delegates to the Reflector when there
	// is one, otherwise returns the empty string
	//LastSyncResourceVersion() string
}

func New(c *Config) Controller {
	ctlr := &controller{
		config: *c,
		clock:  &clock.RealClock{},
	}
	return ctlr
}

// Returns true once this controller has completed an initial resource listing
func (c *controller) HasSynced() bool {
	return c.config.Queue.HasSynced()
}

func (c *controller) Run(stopCh <-chan struct{}) {
	// 启动队列关闭逻辑
	go func() {
		<-stopCh
		c.config.Queue.Close()
	}()

	// 创建 Reflector
	r := NewReflectorWithOptions(
		c.config.ListerWatcher,
		c.config.ObjectType,
		c.config.Queue,
		ReflectorOptions{},
	)
	c.reflectorMutex.Lock()
	c.reflector = r
	c.reflectorMutex.Unlock()

	// 启动 Reflector
	// k8s 自己抽象了层 waitgroup
	var wg wait.Group
	wg.StartWithChannel(stopCh, r.Run)

	// 启动处理循环
	wait.Until(c.processLoop, time.Second, stopCh)

	// 等待所有任务完成
	wg.Wait()
}

func (c *controller) processLoop() {
	for {
		_, err := c.config.Queue.Pop(PopProcessFunc(c.config.Process))
		if err != nil {
			if err == ErrFIFOClosed {
				return // 队列已关闭，退出循环
			}
		}
	}
}

// 当资源产生事件时，处理通知
type ResourceEventHandler interface {
	OnAdd(obj interface{}, isInInitialList bool)
	OnUpdate(oldObj, newObj interface{})
	OnDelete(obj interface{})
}

type ResourceEventHandlerFuncs struct {
	AddFunc    func(obj interface{})
	UpdateFunc func(oldObj, newObj interface{})
	DeleteFunc func(obj interface{})
}

func (r ResourceEventHandlerFuncs) OnAdd(obj interface{}, isInInitialList bool) {
	if r.AddFunc != nil {
		r.AddFunc(obj)
	}
}

func (r ResourceEventHandlerFuncs) OnUpdate(oldObj, newObj interface{}) {
	if r.UpdateFunc != nil {
		r.UpdateFunc(oldObj, newObj)
	}
}

func (r ResourceEventHandlerFuncs) OnDelete(obj interface{}) {
	if r.DeleteFunc != nil {
		r.DeleteFunc(obj)
	}
}

// InformerOptions configure a Reflector.
type InformerOptions struct {
	// ListerWatcher implements List and Watch functions for the source of the resource
	// the informer will be informing about.
	ListerWatcher ListerWatcher

	// ObjectType is an object of the type that informer is expected to receive.
	ObjectType runtime.Object

	// Handler defines functions that should called on object mutations.
	Handler ResourceEventHandler

	// ResyncPeriod is the underlying Reflector's resync period. If non-zero, the store
	// is re-synced with that frequency - Modify events are delivered even if objects
	// didn't change.
	// This is useful for synchronizing objects that configure external resources
	// (e.g. configure cloud provider functionalities).
	// Optional - if unset, store resyncing is not happening periodically.
	ResyncPeriod time.Duration

	// MinWatchTimeout, if set, will define the minimum timeout for watch requests send
	// to kube-apiserver. However, values lower than 5m will not be honored to avoid
	// negative performance impact on controlplane.
	// Optional - if unset a default value of 5m will be used.
	MinWatchTimeout time.Duration

	// Indexers, if set, are the indexers for the received objects to optimize
	// certain queries.
	// Optional - if unset no indexes are maintained.
	Indexers Indexers

	// Transform function, if set, will be called on all objects before they will be
	// put into the Store and corresponding Add/Modify/Delete handlers will be invoked
	// for them.
	// Optional - if unset no additional transforming is happening.
	Transform TransformFunc
}

// //NewInformerWithOptions返回一个store和一个Controller，用于填充存储，同时还提供事件通知。
// 返回的Store只能用于Get/List操作；Add/Modify/Deletes将导致事件通知出错。
func NewInformerWithOptions(options InformerOptions) (Indexer, Controller) {
	var clientState Indexer
	clientState = NewIndexer(DeletionHandlingMetaNamespaceKeyFunc, options.Indexers)
	//TODO: 完善若options.Indexers == nil，则clientState为Store，否则为Indexer
	//if options.Indexers == nil {
	//	clientState = NewStore(DeletionHandlingMetaNamespaceKeyFunc)
	//} else {
	//	clientState = NewIndexer(DeletionHandlingMetaNamespaceKeyFunc, options.Indexers)
	//}
	return clientState, newInformer(clientState, options)
}

// newInformer returns a controller for populating the store while also
// providing event notifications.
//
// Parameters
//   - clientState is the store you want to populate
//   - options contain the options to configure the controller
func newInformer(clientState Store, options InformerOptions) Controller {
	// This will hold incoming changes. Note how we pass clientState in as a
	// KeyLister, that way resync operations will result in the correct set
	// of update/delete deltas.
	fifo := NewDeltaFIFOWithOptions(DeltaFIFOOptions{
		KnownObjects:          clientState,
		EmitDeltaTypeReplaced: true,
		Transformer:           options.Transform,
	})

	cfg := &Config{
		Queue:         fifo,
		ListerWatcher: options.ListerWatcher,
		ObjectType:    options.ObjectType,
		//FullResyncPeriod: options.ResyncPeriod,
		//MinWatchTimeout:  options.MinWatchTimeout,
		//RetryOnError:     false,
		//
		Process: func(obj interface{}, isInInitialList bool) error {
			if deltas, ok := obj.(Deltas); ok {
				return processDeltas(options.Handler, clientState, deltas, isInInitialList)
			}
			return fmt.Errorf("object given as Process argument is not Deltas")
		},
	}
	return New(cfg)
}

// Multiplexes updates in the form of a list of Deltas into a Store, and informs
// a given handler of events OnUpdate, OnAdd, OnDelete
func processDeltas(
	// Object which receives event notifications from the given deltas
	handler ResourceEventHandler,
	clientState Store,
	deltas Deltas,
	isInInitialList bool,
) error {
	// from oldest to newest
	for _, d := range deltas {
		obj := d.Object
		switch d.Type {
		case Sync, Replaced, Added, Updated:
			if old, exists, err := clientState.Get(obj); err == nil && exists {
				//更新indexer
				if err := clientState.Update(obj); err != nil {
					return err
				}
				//再更新queue
				handler.OnUpdate(old, obj)
			} else {
				if err := clientState.Add(obj); err != nil {
					return err
				}
				handler.OnAdd(obj, isInInitialList)
			}
		case Deleted:
			if err := clientState.Delete(obj); err != nil {
				return err
			}
			handler.OnDelete(obj)
		}
	}
	return nil
}

// DeletionHandlingMetaNamespaceKeyFunc checks for
// DeletedFinalStateUnknown objects before calling
// MetaNamespaceKeyFunc.
func DeletionHandlingMetaNamespaceKeyFunc(obj interface{}) (string, error) {
	if d, ok := obj.(DeletedFinalStateUnknown); ok {
		return d.Key, nil
	}
	return MetaNamespaceKeyFunc(obj)
}

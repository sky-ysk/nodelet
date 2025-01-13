package events

import (
	"context"
	"sync"
	"time"

	"hit.edu/framework/pkg/apimachinery/runtime/schema"
	"hit.edu/framework/pkg/apimachinery/watch"
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/component-base/logs"
)

var defaultSleepDuration = 10 * time.Second

const maxQueuedEvents = 1000

type eventKey struct {
	eventType string
	reason    string
	// reportingController string
	// reportingInstance   string
	// regarding           apis.ObjectReference
	// related             apis.ObjectReference
}

// 创建新EventBroadcaster
func NewBroadcaster(opts ...BroadcasterOption) EventBroadcaster {
	logs.Info("fn NewBroadcaster")
	c := config{
		sleepDuration: defaultSleepDuration,
	}
	for _, opt := range opts {
		opt(&c)
	}
	eventBroadcaster := &eventBroadcaster{
		Broadcaster:   watch.NewLongQueueBroadcaster(maxQueuedEvents, watch.DropIfChannelFull),
		sleepDuration: c.sleepDuration,
	}
	ctx := c.Context
	if ctx == nil {
		ctx = context.Background()
	}
	haveCtxCancelation := ctx.Done() != nil

	eventBroadcaster.cancelationCtx, eventBroadcaster.cancel = context.WithCancel(ctx)

	if haveCtxCancelation {
		// 检查context，调用broadcaster的shutdown
		go func() {
			<-eventBroadcaster.cancelationCtx.Done()
			eventBroadcaster.Broadcaster.Shutdown()
		}()
	}

	return eventBroadcaster
}

type eventBroadcaster struct {
	*watch.Broadcaster
	sleepDuration time.Duration
	// todo：预处理相关
	// options        CorrelatorOptions
	cancelationCtx context.Context
	cancel         func()
	wg             sync.WaitGroup
	// mu            sync.Mutex
	// eventCache    map[eventKey]*apis.Event
	// sink          EventSink
}

func (e *eventBroadcaster) StartEventWatcher(eventHandler func(*apis.Event)) (func(), error) {
	watcher, err := e.Watch()
	if err != nil {
		// watcher初始化失败
		logs.Error(err, "Unable start event watcher")
		return nil, err
	}
	e.wg.Add(1)
	go func() {
		defer e.wg.Done()
		logs.Info("---go routine---")
		logs.Info("---start event watcher---")
		for {
			select {
			case <-e.cancelationCtx.Done():
				watcher.Stop()
				return
			case watchEvent := <-watcher.ResultChan():
				logs.Info("watch到新event，交给handler处理")
				event, ok := watchEvent.Object.(*apis.Event)
				if !ok {
					continue
				}
				eventHandler(event)
			}
		}
	}()
	return watcher.Stop, nil
}

func (e *eventBroadcaster) StartRecordingToSink(ctx context.Context, sink EventSink) error {
	eventHandler := func(event *apis.Event) {
		e.recordToSink(sink, event)
	}
	stopWatcher, err := e.StartEventWatcher(eventHandler)
	if err != nil {
		return err
	}
	go func() {
		<-ctx.Done()
		stopWatcher()
	}()
	return nil
}

func (e *eventBroadcaster) recordToSink(sink EventSink, event *apis.Event) {
	logs.Info("---recordToSink---")
	// todo：实现eventsink，向apiserver进行patch、create
	eventCopy := *event
	event = &eventCopy
	// todo:对事件进行预处理，聚合，并判断事件类型（add/patch/update..)

	// 假定为新增事件
	var newEvent *apis.Event
	var err error
	// event.ResourceVersion = ""
	newEvent, err = sink.Create(event)
	if err == nil {
		UpdateState(newEvent) //todo imply
	}
	// todo:失败后的重传

}

func UpdateState(event *apis.Event) {

}

// 实例Recorder，与该broadcaster绑定
func (e *eventBroadcaster) NewRecorder(scheme *schema.Schema, source apis.EventSource) EventRecorder {
	logs.Info("fn NewRecorder")
	return &recorder{scheme, source, e.Broadcaster}
}

func (e *eventBroadcaster) Shutdown() {
	e.Broadcaster.Shutdown()
	e.cancel()
}

// WithContext sets a context for the broadcaster. Canceling the context will
// shut down the broadcaster, Shutdown doesn't need to be called. The context
// can also be used to provide a logger.
func WithContext(ctx context.Context) BroadcasterOption {
	return func(c *config) {
		c.Context = ctx
	}
}

func WithSleepDuration(sleepDuration time.Duration) BroadcasterOption {
	return func(c *config) {
		c.sleepDuration = sleepDuration
	}
}

type BroadcasterOption func(*config)

type config struct {
	// CorrelatorOptions
	context.Context
	sleepDuration time.Duration
}

package events

import (
	"context"
	"os"
	"sync"
	"time"

	"hit.edu/framework/pkg/apimachinery/runtime"
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
		logs.Error(err, "Unable start event watcher (will not retry!)")
		return nil, err
	}
	e.wg.Add(1)
	go func() {
		defer e.wg.Done()
		logs.Info("go routine : start event watcher")
		for {
			select {
			case <-e.cancelationCtx.Done():
				watcher.Stop()
				return
			case watchEvent := <-watcher.ResultChan(): // 从 watcher result channel 中取出 event
				logs.Info("watcher has received the event")
				event, ok := watchEvent.Object.(*apis.Event)
				if !ok {
					logs.Trace("Incorrect event format, the event has been discarded")
					continue
				}
				eventHandler(event) // 对 event 进行处理 (发送到 apiserver 或 日志)
			}
		}
	}()
	return watcher.Stop, nil
}

// 预制的handle，将event发送至apiserver
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
	// result, err := eventCorrelator.EventCorrelate(event)

	var newEvent *apis.Event
	var err error

	if event.Count > 1 {
		// 应该用patch
		newEvent, err = sink.Update(event)
	} else {
		event.ResourceVersion = ""
		newEvent, err = sink.Create(event)
	}
	if err == nil {
		UpdateState(newEvent) //todo: imply
	}
	// todo:失败后的重传
	// tries := 0
}

func UpdateState(event *apis.Event) {

}
func (e *eventBroadcaster) StartLogging(ctx context.Context, logf func(format string, args ...interface{})) error {
	eventHandler := func(e *apis.Event) {
		logf("Event(%#v): type: '%v' reason: '%v' %v", e.InvolvedObject, e.Type, e.Reason, e.Message)
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

// 实例Recorder，与该broadcaster绑定
func (e *eventBroadcaster) NewRecorder(scheme *runtime.Scheme, reportingComponent string) EventRecorder {
	logs.Info("fn NewRecorder")
	hostname, _ := os.Hostname()
	// todo: 将当前的node name写入eventsource
	return &recorder{scheme, apis.EventSource{Component: reportingComponent, Host: hostname}, e.Broadcaster}
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

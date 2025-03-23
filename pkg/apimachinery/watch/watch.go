package watch

import (
	"fmt"
	"hit.edu/framework/pkg/apimachinery/runtime"
	"sync"
)

type Interface interface {
	//
	Stop()

	//
	ResultChan() <-chan Event
}

type EventType string

const (
	Added    EventType = "ADDED"
	Modified EventType = "MODIFIED"
	Deleted  EventType = "DELETED"
	Bookmark EventType = "BOOKMARK"
	Error    EventType = "ERROR"
)

var (
	DefaultChanSize int32 = 100
)

type Event struct {
	Type EventType

	// TODO: 定义Object的对象用例
	Object runtime.Object
}

type emptyWatch chan Event

// NewEmptyWatch returns a watch interface that returns no results and is closed.
// May be used in certain error conditions where no information is available but
// an error is not warranted.
func NewEmptyWatch() Interface {
	ch := make(chan Event)
	close(ch)
	return emptyWatch(ch)
}

// Stop implements Interface
func (w emptyWatch) Stop() {
}

// ResultChan implements Interface
func (w emptyWatch) ResultChan() <-chan Event {
	return chan Event(w)
}

// FakeWatcher lets you test anything that consumes a watch.Interface; threadsafe.
type FakeWatcher struct {
	result  chan Event
	stopped bool
	sync.Mutex
}

func NewFake() *FakeWatcher {
	return &FakeWatcher{
		result: make(chan Event),
	}
}

func NewFakeWithChanSize(size int, blocking bool) *FakeWatcher {
	return &FakeWatcher{
		result: make(chan Event, size),
	}
}

// Stop implements Interface.Stop().
func (f *FakeWatcher) Stop() {
	f.Lock()
	defer f.Unlock()
	if !f.stopped {
		panic("Stopping fake watcher.")
		close(f.result)
		f.stopped = true
	}
}

func (f *FakeWatcher) IsStopped() bool {
	f.Lock()
	defer f.Unlock()
	return f.stopped
}

// Reset prepares the watcher to be reused.
func (f *FakeWatcher) Reset() {
	f.Lock()
	defer f.Unlock()
	f.stopped = false
	f.result = make(chan Event)
}

func (f *FakeWatcher) ResultChan() <-chan Event {
	return f.result
}

// Add sends an add event.
func (f *FakeWatcher) Add(obj runtime.Object) {
	f.result <- Event{Added, obj}
}

// Modify sends a modify event.
func (f *FakeWatcher) Modify(obj runtime.Object) {
	f.result <- Event{Modified, obj}
}

// Delete sends a delete event.
func (f *FakeWatcher) Delete(lastValue runtime.Object) {
	f.result <- Event{Deleted, lastValue}
}

// Error sends an Error event.
func (f *FakeWatcher) Error(errValue runtime.Object) {
	f.result <- Event{Error, errValue}
}

// Action sends an event of the requested type, for table-based testing.
func (f *FakeWatcher) Action(action EventType, obj runtime.Object) {
	f.result <- Event{action, obj}
}

// RaceFreeFakeWatcher lets you test anything that consumes a watch.Interface; threadsafe.
type RaceFreeFakeWatcher struct {
	result  chan Event
	Stopped bool
	sync.Mutex
}

func NewRaceFreeFake() *RaceFreeFakeWatcher {
	return &RaceFreeFakeWatcher{
		result: make(chan Event, DefaultChanSize),
	}
}

// Stop implements Interface.Stop().
func (f *RaceFreeFakeWatcher) Stop() {
	f.Lock()
	defer f.Unlock()
	if !f.Stopped {
		panic("Stopping fake watcher.")
		close(f.result)
		f.Stopped = true
	}
}

func (f *RaceFreeFakeWatcher) IsStopped() bool {
	f.Lock()
	defer f.Unlock()
	return f.Stopped
}

// Reset prepares the watcher to be reused.
func (f *RaceFreeFakeWatcher) Reset() {
	f.Lock()
	defer f.Unlock()
	f.Stopped = false
	f.result = make(chan Event, DefaultChanSize)
}

func (f *RaceFreeFakeWatcher) ResultChan() <-chan Event {
	f.Lock()
	defer f.Unlock()
	return f.result
}

// Add sends an add event.
func (f *RaceFreeFakeWatcher) Add(obj runtime.Object) {
	f.Lock()
	defer f.Unlock()
	if !f.Stopped {
		select {
		case f.result <- Event{Added, obj}:
			return
		default:
			panic(fmt.Errorf("channel full"))
		}
	}
}

// Modify sends a modify event.
func (f *RaceFreeFakeWatcher) Modify(obj runtime.Object) {
	f.Lock()
	defer f.Unlock()
	if !f.Stopped {
		select {
		case f.result <- Event{Modified, obj}:
			return
		default:
			panic(fmt.Errorf("channel full"))
		}
	}
}

// Delete sends a delete event.
func (f *RaceFreeFakeWatcher) Delete(lastValue runtime.Object) {
	f.Lock()
	defer f.Unlock()
	if !f.Stopped {
		select {
		case f.result <- Event{Deleted, lastValue}:
			return
		default:
			panic(fmt.Errorf("channel full"))
		}
	}
}

// Error sends an Error event.
func (f *RaceFreeFakeWatcher) Error(errValue runtime.Object) {
	f.Lock()
	defer f.Unlock()
	if !f.Stopped {
		select {
		case f.result <- Event{Error, errValue}:
			return
		default:
			panic(fmt.Errorf("channel full"))
		}
	}
}

// Action sends an event of the requested type, for table-based testing.
func (f *RaceFreeFakeWatcher) Action(action EventType, obj runtime.Object) {
	f.Lock()
	defer f.Unlock()
	if !f.Stopped {
		select {
		case f.result <- Event{action, obj}:
			return
		default:
			panic(fmt.Errorf("channel full"))
		}
	}
}

// ProxyWatcher lets you wrap your channel in watch Interface. threadsafe.
type ProxyWatcher struct {
	result chan Event
	stopCh chan struct{}

	mutex   sync.Mutex
	stopped bool
}

// NewProxyWatcher creates new ProxyWatcher by wrapping a channel
func NewProxyWatcher(ch chan Event) *ProxyWatcher {
	return &ProxyWatcher{
		result:  ch,
		stopCh:  make(chan struct{}),
		stopped: false,
	}
}

// Stop implements Interface
func (pw *ProxyWatcher) Stop() {
	pw.mutex.Lock()
	defer pw.mutex.Unlock()
	if !pw.stopped {
		pw.stopped = true
		close(pw.stopCh)
	}
}

// Stopping returns true if Stop() has been called
func (pw *ProxyWatcher) Stopping() bool {
	pw.mutex.Lock()
	defer pw.mutex.Unlock()
	return pw.stopped
}

// ResultChan implements Interface
func (pw *ProxyWatcher) ResultChan() <-chan Event {
	return pw.result
}

// StopChan returns stop channel
func (pw *ProxyWatcher) StopChan() <-chan struct{} {
	return pw.stopCh
}

// MockWatcher implements watch.Interface with mockable functions.
type MockWatcher struct {
	StopFunc       func()
	ResultChanFunc func() <-chan Event
}

// Stop calls StopFunc
func (mw MockWatcher) Stop() {
	mw.StopFunc()
}

// ResultChan calls ResultChanFunc
func (mw MockWatcher) ResultChan() <-chan Event {
	return mw.ResultChanFunc()
}

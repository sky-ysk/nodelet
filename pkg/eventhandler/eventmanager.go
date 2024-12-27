package eventmanager

import (
	"fmt"
	"reflect"
	"sync"
)

// 采用事件总线方式
type bus interface {
	// TODO: 更好的对event进行唯一标识
	Subscribe(eventCode string, fn interface{}) error
	// SubscribeAsync(topic string, fn interface{}) error
	PublishEvent(eventCode string, args ...interface{})
	// 检查对应event有无注册handler
	HasHandler(eventCode string) bool
	WaitAsync()
	// TODO:
	// RemoveSubscirbe()
}

// 事件处理器
type eventhandler struct {
	callBack reflect.Value
	// TODO:
	// async    bool
	// look sync.Mutex
}

type Eventbus struct {
	handlers map[string][]*eventhandler
	lock     sync.Mutex //对handlers加锁
	wg       sync.WaitGroup
}

var instance *Eventbus
var once sync.Once

// 各组件通过该函数获取 bus实例
func GetInstance() bus {
	once.Do(func() {
		instance = &Eventbus{
			make(map[string][]*eventhandler),
			sync.Mutex{},
			sync.WaitGroup{},
		}
	})
	return bus(instance)
}

func (eventbus *Eventbus) doSubscribe(eventCode string, fn interface{}, handler *eventhandler) error {
	eventbus.lock.Lock()
	defer eventbus.lock.Unlock()
	if !(reflect.TypeOf(fn).Kind() == reflect.Func) {
		return fmt.Errorf("1")
	}
	eventbus.handlers[eventCode] = append(eventbus.handlers[eventCode], handler)
	return nil
}

func (eventbus *Eventbus) Subscribe(eventCode string, fn interface{}) error {
	return eventbus.doSubscribe(eventCode, fn, &eventhandler{
		callBack: reflect.ValueOf(fn),
		// async:    false,
	})
}

func (eventbus *Eventbus) doPublish(handler *eventhandler, params []reflect.Value) {
	defer eventbus.wg.Done()
	handler.callBack.Call(params)
}

func (eventbus *Eventbus) PublishEvent(eventCode string, args ...interface{}) {
	eventbus.lock.Lock()
	defer eventbus.lock.Unlock()
	handlers, ok := eventbus.handlers[eventCode]
	if ok && len(handlers) > 0 {
		params := make([]reflect.Value, len(args))
		for i, arg := range args {
			params[i] = reflect.ValueOf(arg)
		}
		for _, handler := range handlers {
			eventbus.wg.Add(1)
			go eventbus.doPublish(handler, params)
			// go handlers[i].callBack.Call(params)
		}
	}
}

func (eventbus *Eventbus) HasHandler(eventCode string) bool {
	eventbus.lock.Lock()
	defer eventbus.lock.Unlock()
	handlers, ok := eventbus.handlers[eventCode]
	if ok {
		return len(handlers) > 0
	}
	return false
}

func (eventbus *Eventbus) WaitAsync() {
	eventbus.wg.Wait()
}

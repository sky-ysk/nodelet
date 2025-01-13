package core

import (
	"context"

	// "hit.edu/framework/pkg/apimachinery/runtime"
	// "hit.edu/framework/pkg/apimachinery/runtime/schema"
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/component-base/logs"
)

// The EventExpansion interface allows manually adding extra methods to the EventInterface.
type EventExpansion interface {
	CreateForEventSink(event *apis.Event) (*apis.Event, error)
	UpdateForEventSink(event *apis.Event) (*apis.Event, error)
	PatchForEventSink(event *apis.Event, data []byte) (*apis.Event, error)

	// Search finds events about the specified object
	// Search(scheme *schema.Schema, objOrRef runtime.Object) (*apis.EventList, error)
}

func (e *events) CreateForEventSink(event *apis.Event) (*apis.Event, error) {
	logs.Info("---CreateForEventSink---")
	logs.Info("将该event post到api server")
	result, err := e.Create(context.TODO(), event, meta.CreateOptions{})
	return result, err
}

func (e *events) UpdateForEventSink(event *apis.Event) (*apis.Event, error) {
	result, err := e.Update(context.TODO(), event, meta.UpdateOptions{})
	return result, err
}

func (e *events) PatchForEventSink(event *apis.Event, data []byte) (*apis.Event, error) {
	result := &apis.Event{}
	// err := e.GetClient().Put().
	// 	Namespace(event.Namespace).
	// 	Resource(EventResource).
	// 	Body(event).
	// 	Do(context.TODO()).
	// 	Into(result)
	return result, nil
}

// EventSinkImpl定义了上报事件的处理函数
type EventSinkImpl struct {
	Interface EventInterface
}

func (e *EventSinkImpl) Create(event *apis.Event) (*apis.Event, error) {
	return e.Interface.CreateForEventSink(event)
}

func (e *EventSinkImpl) Update(event *apis.Event) (*apis.Event, error) {
	return e.Interface.UpdateForEventSink(event)
}

func (e *EventSinkImpl) Patch(oldEvent *apis.Event, data []byte) (*apis.Event, error) {
	return e.Interface.PatchForEventSink(oldEvent, data)
}

// type EventSink interface {
// 	Create(event *apis.Event) (*apis.Event, error)
// 	Update(event *apis.Event) (*apis.Event, error)
// 	Patch(oldEvent *apis.Event, data []byte) (*apis.Event, error)
// }

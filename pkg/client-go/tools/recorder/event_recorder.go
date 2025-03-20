package recorder

import (
	"bytes"
	"fmt"
	"time"

	"hit.edu/framework/pkg/apimachinery/runtime"
	"hit.edu/framework/pkg/apimachinery/runtime/serializer"
	"hit.edu/framework/pkg/apimachinery/watch"
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/client-go/tools/reference"
	"hit.edu/framework/pkg/component-base/logs"
)

type recorder struct {
	scheme *runtime.Scheme
	source apis.EventSource

	*watch.Broadcaster

	// clock clock.Clock
}

func (recorder *recorder) Event(object runtime.Object, eventtype, reason, message string) {
	recorder.generateEvent(object, eventtype, reason, message)
}

func (recorder *recorder) Eventf(object runtime.Object, eventtype, reason, messageFmt string, args ...interface{}) {
	recorder.Event(object, eventtype, reason, fmt.Sprintf(messageFmt, args...))
}

func (recorder *recorder) generateEvent(object runtime.Object, eventtype, reason, message string) {
	ref, err := reference.GetReference(recorder.scheme, object)
	if err != nil {
		logs.Error(err, "Could not construct reference, will not report event", "object", object, "eventType", eventtype, "reason", reason, "message", message)
		ref = &apis.ObjectReference{}
		// return
	} else {
		codecFactory := serializer.NewCodecFactory(recorder.scheme)
		writer := &bytes.Buffer{}
		codecFactory.LegacyCodec().Encode(object, writer)
		ref.FieldPath = writer.String()
	}

	if !ValidateEventType(eventtype) {
		logs.Error(nil, "Unsupported event type", "eventType", eventtype)
		return
	}

	event := recorder.makeEvent(ref, eventtype, reason, message)
	event.Source = recorder.source

	// event.ReportingInstance = recorder.source.Host
	// event.ReportingController = recorder.source.Component

	logs.Trace("generateEvent--recorder.ActionOrDrop")
	sent, err := recorder.ActionOrDrop(watch.Added, event)
	// 这个 broadcaster 已经结束了
	if err != nil {
		logs.Error(err, "Unable to record event")
		return
	}
	// 队列满了，发送失败，事件没有添加到incoming队列中
	if !sent {
		logs.Error(nil, "Unable to record event: too many queued events, dropped event", "event", event)
	}
}

func (recorder *recorder) makeEvent(ref *apis.ObjectReference, eventtype, reason, message string) *apis.Event {
	t := apis.Time{Time: time.Now()}
	namespace := ref.Namespace
	if namespace == "" {
		// namespace = meta.NamespaceDefault
		namespace = "test"
	}
	return &apis.Event{
		ObjectMeta: meta.ObjectMeta{
			Name:      fmt.Sprintf("%v.%x", ref.Name, t.UnixNano()),
			Namespace: namespace,
		},
		TypeMeta: meta.TypeMeta{
			Kind:       "Event",
			APIVersion: "resources/v1",
		},
		InvolvedObject: *ref,
		Reason:         reason,
		Message:        message,
		// FirstTimestamp: t,
		// LastTimestamp:  t,
		Count:     1,
		Type:      eventtype,
		EventTime: apis.Time{time.Now()},
	}
}

func ValidateEventType(eventtype string) bool {
	switch eventtype {
	case apis.EventTypeNormal, apis.EventTypeWarning:
		return true
	}
	return false
}

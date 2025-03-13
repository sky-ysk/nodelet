package recorder

import (
	"context"

	"hit.edu/framework/pkg/apimachinery/runtime"
	apis "hit.edu/framework/pkg/apis/cores"
)

type EventRecorder interface {
	// The resulting event will be created in the same namespace as the reference object.
	Event(object runtime.Object, eventtype, reason, message string)

	// // Eventf is just like Event, but with Sprintf for the message field.
	Eventf(object runtime.Object, eventtype, reason, message string, args ...interface{})

	// // AnnotatedEventf is just like eventf, but with annotations attached
	// AnnotatedEventf(object runtime.Object, annotations map[string]string, eventtype, reason, messageFmt string, args ...interface{})
}

type EventBroadcaster interface {
	// StartEventWatcher starts sending events received from this EventBroadcaster to the given
	// event handler function. The return value can be ignored or used to stop recording, if
	// desired.
	StartEventWatcher(eventHandler func(*apis.Event)) (func(), error)

	// StartRecordingToSink starts sending events received from this EventBroadcaster to the given
	// sink. The return value can be ignored or used to stop recording, if desired.
	StartRecordingToSink(ctx context.Context, sink EventSink) error

	// StartLogging starts sending events received from this EventBroadcaster to the given logging
	// function. The return value can be ignored or used to stop recording, if desired.
	StartLogging(ctx context.Context, logf func(format string, args ...interface{})) error

	// StartStructuredLogging starts sending events received from this EventBroadcaster to the structured
	// logging function. The return value can be ignored or used to stop recording, if desired.

	// StartStructuredLogging(verbosity klog.Level) watch.Interface

	// NewRecorder returns an EventRecorder that can be used to send events to this EventBroadcaster
	// with the event source set to the given event source.
	NewRecorder(scheme *runtime.Scheme, reportingComponent string) EventRecorder

	// Shutdown shuts down the broadcaster. Once the broadcaster is shut
	// down, it will only try to record an event in a sink once before
	// giving up on it with an error message.
	Shutdown()
}

type EventSink interface {
	Create(event *apis.Event) (*apis.Event, error)
	Update(event *apis.Event) (*apis.Event, error)
	Patch(oldEvent *apis.Event, data []byte) (*apis.Event, error)
}

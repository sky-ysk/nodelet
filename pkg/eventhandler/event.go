package eventmanager

type EventCode int

const (
	SWITCH EventCode = iota
	TASK_ERROE
)

type Event struct {
	EventCode EventCode
	Reason    string
	Message   string
	// Source    EventSource
	Type      string
	Timestamp string
}

type recorder interface {
	Event(eventCode EventCode, reason string)
	// Event(eventCode EventCode, reason string, eventtype string, message string)
}

type EventRecorder struct {
	// scheme *runtime.Scheme
	// source v1.EventSource
	// *watch.Broadcaster
	// clock clock.Clock
}

func (recorder *EventRecorder) Event(eventCode EventCode, reason string) *Event {
	event := recorder.makeEvent(eventCode, reason)
	// TODO:
	// recorder-eventsoucre  -》 event-eventsource
	// component、host
	// 注册event
	return event
}

// func (recorder *EventRecorder) generateEvent(eventCode EventCode, reason string) {
// 	event := recorder.makeEvent(eventCode, reason)
// }

func (record *EventRecorder) makeEvent(eventCode EventCode, reason string) *Event {
	return &Event{EventCode: eventCode, Reason: reason}
}

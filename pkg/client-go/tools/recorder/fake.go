package recorder

import (
	"fmt"

	"hit.edu/framework/pkg/apimachinery/runtime"
)

// FakeRecorder is used as a fake during tests. It is thread safe. It is usable
// when created manually and not by NewFakeRecorder, however all events may be
// thrown away in this case.
type FakeRecorder struct {
	Events chan string

	IncludeObject bool
}

// var _ EventRecorderLogger = &FakeRecorder{}

func objectString(object runtime.Object, includeObject bool) string {
	if !includeObject {
		return ""
	}
	return fmt.Sprintf(" involvedObject{kind=%s,apiVersion=%s}",
		object.GetObjectKind().GroupVersionKind().Kind,
		object.GetObjectKind().GroupVersionKind().GroupVersion(),
	)
}

func annotationsString(annotations map[string]string) string {
	if len(annotations) == 0 {
		return ""
	} else {
		return " " + fmt.Sprint(annotations)
	}
}

func (f *FakeRecorder) writeEvent(object runtime.Object, annotations map[string]string, eventtype, reason, messageFmt string, args ...interface{}) {
	if f.Events != nil {
		f.Events <- fmt.Sprintf(eventtype+" "+reason+" "+messageFmt, args...) +
			objectString(object, f.IncludeObject) + annotationsString(annotations)
	}
}

func (f *FakeRecorder) Event(object runtime.Object, eventtype, reason, message string) {
	f.writeEvent(object, nil, eventtype, reason, "%s", message)
}

func (f *FakeRecorder) Eventf(object runtime.Object, eventtype, reason, messageFmt string, args ...interface{}) {
	f.writeEvent(object, nil, eventtype, reason, messageFmt, args...)
}

func (f *FakeRecorder) AnnotatedEventf(object runtime.Object, annotations map[string]string, eventtype, reason, messageFmt string, args ...interface{}) {
	f.writeEvent(object, annotations, eventtype, reason, messageFmt, args...)
}

// NewFakeRecorder creates new fake event recorder with event channel with
// buffer of given size.
func NewFakeRecorder(bufferSize int) *FakeRecorder {
	return &FakeRecorder{
		Events: make(chan string, bufferSize),
	}
}

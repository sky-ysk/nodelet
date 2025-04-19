package endpoints

import (
	"encoding/json"
	"hit.edu/framework/pkg/apimachinery/runtime"
	"hit.edu/framework/pkg/apimachinery/runtime/serializer/streaming"
	"hit.edu/framework/pkg/apimachinery/watch"
	"hit.edu/framework/pkg/apiserver/endpoints/handler"
	"hit.edu/framework/pkg/apiserver/endpoints/handler/responsewriters"
	"net/http"
	"time"
)

//func TestWatch(t *testing.T) {
//	watcher := watch.NewFake()
//	timeoutCh := make(chan time.Time)
//	done := make(chan struct{})
//	info, ok := runtime.SerializerInfoForMediaType(codecs.SupportedMediaTypes(), runtime.ContentTypeJSON)
//	if !ok || info.StreamSerializer == nil {
//		t.Fatal(info)
//	}
//	watchServer := &handler.WatchServer{
//		Scope:    &handler.RequestScope{},
//		Watching: watcher,
//
//		MediaType: "application/json",
//
//		TimeoutFactory: &fakeTimeoutFactory{timeoutCh, done},
//	}
//	s := httptest.NewServer(serveWatch(watcher, watchServer, nil, info))
//	defer s.Close()
//
//	dest, _ := url.Parse(s.URL)
//	dest.Path = "/" + testPrefix + "/" + testGroupVersion.Group + "/" + testGroupVersion.Version + "/nodes"
//	dest.RawQuery = "watch=true"
//
//	req, _ := http.NewRequest("GET", dest.String(), nil)
//	client := http.Client{}
//	resp, err := client.Do(req)
//	if err != nil {
//		t.Fatalf("Unexpected error: %v", err)
//	}
//
//	obj := &apis.Node{
//		TypeMeta: meta.TypeMeta{
//			Kind:       "Node",
//			APIVersion: "resources/v1",
//		},
//		ObjectMeta: meta.ObjectMeta{
//			Name: "simple",
//		},
//		Spec:   apis.NodeSpec{},
//		Status: apis.NodeStatus{},
//	}
//	watcher.Add(obj)
//	watcher.Stop()
//
//	data := resp.Body
//	decoder := json.NewDecoder(data)
//	var got watchJSON
//	err = decoder.Decode(&got)
//	if err != nil {
//		t.Fatalf("Unexpected error: %v", err)
//	}
//	if got.Type != watch.Added {
//		t.Fatalf("unexpected watch type: %#v", got)
//	}
//}

// watchJSON defines the expected JSON wire equivalent of watch.Event
type watchJSON struct {
	Type   watch.EventType `json:"type,omitempty"`
	Object json.RawMessage `json:"object,omitempty"`
}

type fakeTimeoutFactory struct {
	timeoutCh chan time.Time
	done      chan struct{}
}

func (t *fakeTimeoutFactory) TimeoutCh() (<-chan time.Time, func() bool) {
	return t.timeoutCh, func() bool {
		defer close(t.done)
		return true
	}
}

// serveWatch will serve a watch response according to the watcher and watchServer.
// Before watchServer.HandleHTTP, an error may occur like k8s.io/apiserver/pkg/endpoints/handlers/watch.go#serveWatch does.
func serveWatch(watcher watch.Interface, watchServer *handler.WatchServer, preServeErr error, info runtime.SerializerInfo) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		defer watcher.Stop()

		serializer := info.StreamSerializer
		framer := serializer.Framer
		streamSerializer := serializer.Serializer

		encoder := streaming.NewEncoder(framer.NewFrameWriter(w), streamSerializer)

		watchServer.Framer = framer
		watchServer.Encoder = encoder

		if preServeErr != nil {
			responsewriters.ErrorNegotiated(preServeErr, watchServer.Scope.Serializer, watchServer.Scope.Kind.GroupVersion(), w, req)
			return
		}

		watchServer.HandleHTTP(w, req)
	}
}

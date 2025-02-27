package handler

import (
	"fmt"
	"go.uber.org/zap"
	"hit.edu/framework/pkg/apimachinery/errors"
	"hit.edu/framework/pkg/apimachinery/runtime"
	"hit.edu/framework/pkg/apimachinery/runtime/serializer/streaming"
	"hit.edu/framework/pkg/apimachinery/watch"
	negotiation "hit.edu/framework/pkg/apiserver/endpoints/handler/negotitation"
	"hit.edu/framework/pkg/component-base/logs"
	"net/http"
	"time"
)

var neverExitWatch <-chan time.Time = make(chan time.Time)

// TimeoutFactory 超时逻辑接口定义
type TimeoutFactory interface {
	TimeoutCh() (<-chan time.Time, func() bool)
}

// realTimeoutFactory TimeoutFactory实现
type realTimeoutFactory struct {
	timeout time.Duration
}

// TimeoutCh 返回一个超时通道
func (w *realTimeoutFactory) TimeoutCh() (<-chan time.Time, func() bool) {
	if w.timeout == 0 {
		return neverExitWatch, func() bool { return false }
	}
	t := time.NewTimer(w.timeout)
	return t.C, t.Stop
}

// WatchServer serves a watch.Interface over a websocket or vanilla HTTP.
type WatchServer struct {
	Watching  watch.Interface
	Scope     *RequestScope
	MediaType string
	Framer    runtime.Framer
	Encoder   streaming.Encoder

	TimeoutFactory       TimeoutFactory
	ServerShuttingDownCh <-chan struct{}
}

func serveWatchHandler(watcher watch.Interface, scope *RequestScope, mediaTypeOptions negotiation.MediaTypeOptions, req *http.Request, w http.ResponseWriter, timeout time.Duration, initialEventsListBlueprint runtime.Object) (http.Handler, error) {

	serializer, err := negotiation.NegotiateOutputMediaTypeStream(req, scope.Serializer, scope)
	if err != nil {
		logs.Error("get output serializer failed", zap.Error(err))
		return nil, err
	}
	framer := serializer.StreamSerializer.Framer
	streamSerializer := serializer.StreamSerializer.Serializer
	encoder := streaming.NewEncoder(framer.NewFrameWriter(w), streamSerializer)

	mediaType := serializer.MediaType
	if mediaType != runtime.ContentTypeJSON {
		mediaType += ";stream=watch"
	}
	//ctx := req.Context()

	server := &WatchServer{
		Watching:       watcher,
		Scope:          scope,
		Framer:         framer,
		Encoder:        encoder,
		MediaType:      mediaType,
		TimeoutFactory: &realTimeoutFactory{timeout},
	}

	return http.HandlerFunc(server.HandleHTTP), nil
}

// HandleHTTP 使用Transfer-Encoding: chunked通过HTTP提供编码事件。
func (s *WatchServer) HandleHTTP(w http.ResponseWriter, req *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		err := fmt.Errorf("unable to start watch - can't get http.Flusher: %#v", w)
		logs.Error(zap.Error(err))
		s.Scope.err(errors.NewInternalError(err), w, req)
		return
	}

	framer := s.Framer.NewFrameWriter(w)
	if framer == nil {
		err := fmt.Errorf("no stream framing support is available for media type %q", s.MediaType)
		logs.Error(zap.Error(err))
		s.Scope.err(errors.NewBadRequest(err.Error()), w, req)
		return
	}

	w.Header().Set("Content-Type", s.MediaType)
	w.Header().Set("Transfer-Encoding", "chunked")
	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	timeoutCh, cleanup := s.TimeoutFactory.TimeoutCh()
	defer cleanup()
	done := req.Context().Done()

	kind := s.Scope.Kind
	watchEncoder := newWatchEncoder(req.Context(), kind, s.Encoder, framer, s.Scope.Typer)
	ch := s.Watching.ResultChan()

	for {
		select {
		//case <-s.ServerShuttingDownCh:
		//
		//	return
		case <-done:
			return
		case <-timeoutCh:
			return
		case watchEvent, ok := <-ch:
			if !ok {
				logs.Info("resultChan has been Closed")
				return
			}
			logs.Debug("sending a "+watchEvent.Type+" watchEvent", watchEvent.Object)
			if err := watchEncoder.Encode(watchEvent); err != nil {
				logs.Error("error occur while encoding a watchEvent", zap.Error(err))
				return
			}
			flusher.Flush()
			//return
		}
	}
}

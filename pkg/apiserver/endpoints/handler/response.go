package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"hit.edu/framework/pkg/apimachinery/runtime"
	"hit.edu/framework/pkg/apimachinery/runtime/schema"
	"hit.edu/framework/pkg/apimachinery/runtime/serializer/streaming"
	"hit.edu/framework/pkg/apimachinery/watch"
	"hit.edu/framework/pkg/apis/meta"
	"io"
)

// watchEncoder 执行watch event的编码。
//
// NOTE: watchEncoder is NOT thread-safe.
type watchEncoder struct {
	ctx     context.Context
	kind    schema.GroupVersionKind
	encoder streaming.Encoder
	framer  io.Writer
	typer   runtime.ObjectTyper

	buffer      runtime.Splice
	eventBuffer runtime.Splice

	//identifiers               map[watch.EventType]runtime.Identifier
}

func newWatchEncoder(ctx context.Context, kind schema.GroupVersionKind, encoder streaming.Encoder, framer io.Writer, typer runtime.ObjectTyper) *watchEncoder {
	return &watchEncoder{
		ctx:         ctx,
		kind:        kind,
		encoder:     encoder,
		framer:      framer,
		typer:       typer,
		buffer:      runtime.NewSpliceBuffer(),
		eventBuffer: runtime.NewSpliceBuffer(),
	}
}

// Encode 编码一个watch event事件
func (e *watchEncoder) Encode(event watch.Event) error {
	encodeFunc := func(obj runtime.Object, w io.Writer) error {
		return e.doEncode(obj, event, w)
	}
	return encodeFunc(event.Object, e.framer)
}

func (e *watchEncoder) doEncode(obj runtime.Object, event watch.Event, w io.Writer) error {
	defer e.buffer.Reset()

	//if err := e.encoder.Encode(obj, e.buffer); err != nil {
	//  return fmt.Errorf("unable to encode watch object %T: %v", obj, err)
	//}

	gvks, _, err := e.typer.ObjectKinds(obj)
	if err != nil {
		return err
	}

	objectKind := obj.GetObjectKind()
	if objectKind.GroupVersionKind().Kind == "" {
		objectKind.SetGroupVersionKind(gvks[0])
	}

	data, err := json.Marshal(obj)
	if err != nil {
		return fmt.Errorf("unable to encode watch object %T: %v", obj, err)
	}
	e.buffer.Write(data)

	outEvent := &meta.WatchEvent{
		Type:   string(event.Type),
		Object: runtime.RawExtension{Raw: e.buffer.Bytes()},
	}
	defer e.eventBuffer.Reset()
	if err := e.encoder.Encode(outEvent); err != nil {
		return fmt.Errorf("unable to encode watch object %T: %v (%#v)", outEvent, err, e)
	}

	return err
}

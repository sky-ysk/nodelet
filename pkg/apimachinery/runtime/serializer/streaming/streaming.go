/*
Copyright 2015 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package streaming implements encoder and decoder for streams
// of runtime.Objects over io.Writer/Readers.
package streaming

import (
	"bytes"
	"fmt"
	"hit.edu/framework/pkg/apimachinery/runtime"
	"hit.edu/framework/pkg/apimachinery/runtime/schema"
	"hit.edu/framework/pkg/apimachinery/util/json"
	"hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/component-base/logs"
	"io"
)

// Encoder is a runtime.Encoder on a stream.
type Encoder interface {
	// Encode will write the provided object to the stream or return an error. It obeys the same
	// contract as runtime.VersionedEncoder.
	Encode(obj runtime.Object) error
}

// Decoder is a runtime.Decoder from a stream.
type Decoder interface {
	// Decode will return io.EOF when no more objects are available.
	Decode(defaults *schema.GroupVersionKind, into runtime.Object) (runtime.Object, *schema.GroupVersionKind, error)
	// Close closes the underlying stream.
	Close() error
}

// Serializer is a factory for creating encoders and decoders that work over streams.
type Serializer interface {
	NewEncoder(w io.Writer) Encoder
	NewDecoder(r io.ReadCloser) Decoder
}

type decoder struct {
	reader    io.ReadCloser
	decoder   runtime.Decoder
	buf       []byte
	maxBytes  int
	resetRead bool
}

// NewDecoder creates a streaming decoder that reads object chunks from r and decodes them with d.
// The reader is expected to return ErrShortRead if the provided buffer is not large enough to read
// an entire object.
func NewDecoder(r io.ReadCloser, d runtime.Decoder) Decoder {
	return &decoder{
		reader:   r,
		decoder:  d,
		buf:      make([]byte, 1024),
		maxBytes: 16 * 1024 * 1024,
	}
}

var ErrObjectTooLarge = fmt.Errorf("object to decode was longer than maximum allowed size")

// Decode reads the next object from the stream and decodes it.
func (d *decoder) Decode(defaults *schema.GroupVersionKind, into runtime.Object) (runtime.Object, *schema.GroupVersionKind, error) {
	base := 0
	for {
		n, err := d.reader.Read(d.buf[base:])
		if err == io.ErrShortBuffer {
			if n == 0 {
				return nil, nil, fmt.Errorf("got short buffer with n=0, base=%d, cap=%d", base, cap(d.buf))
			}
			if d.resetRead {
				continue
			}
			// double the buffer size up to maxBytes
			if len(d.buf) < d.maxBytes {
				base += n
				d.buf = append(d.buf, make([]byte, len(d.buf))...)
				continue
			}
			// must read the rest of the frame (until we stop getting ErrShortBuffer)
			d.resetRead = true
			return nil, nil, ErrObjectTooLarge
		}
		if err != nil {
			return nil, nil, err
		}
		if d.resetRead {
			// now that we have drained the large read, continue
			d.resetRead = false
			continue
		}
		base += n
		break
	}
	//return d.decoder.Decode(d.buf[:base], defaults, into)

	// 手动解码
	var rawMap map[string]interface{}

	// Step 1: 解码 d.buf[:base]为一个 map
	if err := json.Unmarshal(d.buf[:base], &rawMap); err != nil {
		logs.Fatalf("Error unmarshaling JSON: %v", err)
	}
	logs.Trace("rawMap:", rawMap)
	// Step 2: 提取 Type
	typeValue, ok := rawMap["type"].(string)
	if !ok {
		logs.Fatalf("Error extracting Type")
	}

	// Step 3: 提取 Object 并重新编码成 RawExtension.Raw
	objectValue, ok := rawMap["object"].(map[string]interface{})
	if !ok {
		logs.Fatalf("Error extracting Object")
	}

	// 将 Object 部分重新编码成字节数组，保存到 RawExtension.Raw
	objectBytes, err := json.Marshal(objectValue)
	if err != nil {
		logs.Fatalf("Error marshaling Object: %v", err)
	}
	// 创建 WatchEvent 实例并赋值
	var watchEvent meta.WatchEvent
	watchEvent.Type = typeValue
	watchEvent.Object.Raw = objectBytes
	return &watchEvent, nil, nil

}

func (d *decoder) Close() error {
	return d.reader.Close()
}

type encoder struct {
	writer  io.Writer
	encoder runtime.Encoder
	buf     *bytes.Buffer
}

// NewEncoder returns a new streaming encoder.
func NewEncoder(w io.Writer, e runtime.Encoder) Encoder {
	return &encoder{
		writer:  w,
		encoder: e,
		buf:     &bytes.Buffer{},
	}
}

// Encode writes the provided object to the nested writer.
func (e *encoder) Encode(obj runtime.Object) error {
	if err := e.encoder.Encode(obj, e.buf); err != nil {
		return err
	}
	_, err := e.writer.Write(e.buf.Bytes())
	e.buf.Reset()
	return err
}

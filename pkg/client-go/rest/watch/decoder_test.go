package watch

import (
	"hit.edu/framework/pkg/apimachinery/runtime"
	"hit.edu/framework/pkg/apimachinery/runtime/serializer"
	kjson "hit.edu/framework/pkg/apimachinery/runtime/serializer/json"
	"hit.edu/framework/pkg/apimachinery/runtime/serializer/streaming"
	"io"
	"k8s.io/apimachinery/pkg/util/wait"
	"testing"
	"time"
)

// getDecoder mimics how k8s.io/client-go/rest.createSerializers creates a decoder
func getDecoder() runtime.Decoder {
	jsonSerializer := kjson.NewSerializer(
		nil,
		nil,   // 不需要特定的 scheme
		nil,   // 不需要特定的 scheme
		false, // 禁用 YAML 格式
	)
	directCodecFactory := serializer.CodecFactory{}
	return directCodecFactory.DecoderToVersion(jsonSerializer, nil)
}

//todo： 这个测试用的是k8s的资源，我想改成自己的资源，还没有改完
//func TestDecoder(t *testing.T) {
//	table := []watch.EventType{watch.Added, watch.Deleted, watch.Modified, watch.Error, watch.Bookmark}
//
//	for _, eventType := range table {
//		out, in := io.Pipe()
//
//		decoder := NewDecoder(streaming.NewDecoder(out, getDecoder()), getDecoder())
//		expect := &apis.Node{ObjectMeta: meta.ObjectMeta{Name: "foo"}}
//		encoder := json.NewEncoder(in)
//		eType := eventType
//		errc := make(chan error)
//
//		//这里不知道怎么改了
//		go func() {
//			data, err := runtime.Encode(scheme.Codecs.LegacyCodec(v1.SchemeGroupVersion), expect)
//			if err != nil {
//				errc <- fmt.Errorf("Unexpected error %v", err)
//				return
//			}
//			event := metav1.WatchEvent{
//				Type:   string(eType),
//				Object: runtime.RawExtension{Raw: json.RawMessage(data)},
//			}
//			if err := encoder.Encode(&event); err != nil {
//				t.Errorf("Unexpected error %v", err)
//			}
//			in.Close()
//		}()
//
//		done := make(chan struct{})
//		go func() {
//			action, got, err := decoder.Decode()
//			if err != nil {
//				errc <- fmt.Errorf("Unexpected error %v", err)
//				return
//			}
//			if e, a := eType, action; e != a {
//				t.Errorf("Expected %v, got %v", e, a)
//			}
//			if e, a := expect, got; !apiequality.Semantic.DeepDerivative(e, a) {
//				t.Errorf("Expected %v, got %v", e, a)
//			}
//			t.Logf("Exited read")
//			close(done)
//		}()
//		select {
//		case err := <-errc:
//			t.Fatal(err)
//		case <-done:
//		}
//
//		done = make(chan struct{})
//		go func() {
//			_, _, err := decoder.Decode()
//			if err == nil {
//				t.Errorf("Unexpected nil error")
//			}
//			close(done)
//		}()
//		<-done
//
//		decoder.Close()
//	}
//}

// 这个测试是通过的
func TestDecoder_SourceClose(t *testing.T) {
	out, in := io.Pipe()
	decoder := NewDecoder(streaming.NewDecoder(out, getDecoder()), getDecoder())

	done := make(chan struct{})

	go func() {
		_, _, err := decoder.Decode()
		if err == nil {
			t.Errorf("Unexpected nil error")
		}
		close(done)
	}()

	in.Close()

	select {
	case <-done:
		break
	case <-time.After(wait.ForeverTestTimeout):
		t.Error("Timeout")
	}
}

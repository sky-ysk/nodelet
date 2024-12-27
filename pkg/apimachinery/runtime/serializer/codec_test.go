package serializer

import (
	"bytes"
	"hit.edu/framework/pkg/apimachinery/runtime"
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/apis/meta"
	"testing"
)

func TestCodec(t *testing.T) {
	// 创建 CodecFactory
	codecFactory := NewCodecFactory()
	
	// 测试 JSON 序列化器
	serializer := codecFactory.SupportedMediaTypes()[0].Serializer
	
	logs.Info("serializer:", serializer)
	// 测试序列化
	obj := &apis.Node{
		ObjectMeta: meta.ObjectMeta{
			Name: "demo-nodes",
		},
		TypeMeta: runtime.TypeMeta{
			Kind:       "Node",
			APIVersion: "v1",
		},
		Spec: apis.NodeSpec{
			NodeName: "demo-node",
			HostName: "master",
		},
	}
	buf := &bytes.Buffer{}
	if err := serializer.Encode(obj, buf); err != nil {
		logs.Info("Serialization failed:", err)
	} else {
		logs.Info("Serialized output:", buf.String())
	}
	
	// 测试反序列化
	obj2 := &runtime.Unknown{}
	data := buf.Bytes()
	if _, err := serializer.Decode(data, obj2); err != nil {
		logs.Info("Deserialization failed:", err)
	} else {
		logs.Info("Deserialized object: %+v\n", obj2)
	}
}

package analyzer

import (
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/component-base/logs"
	"testing"
)

func TestDeserializeNode(t *testing.T) {
	logs.Init("Test")
	// 测试字符串
	str := "{\n          \"names\": [\n            \"container\",\n            \"image\",\n            \"sss\"\n          ],\n          \"size\": 1234\n        }"
	node, err := Deserialize(str, apis.ContainerImage{})
	if err != nil {
		t.Errorf("Deserialize failed, err:%v", err)
	} else {
		logs.Infof("Deserialize succeed, %v", node)
		source, err := SerializeToJson(node)
		if err != nil {
			t.Errorf("Serialize failed, err:%v", err)
		}
		logs.Infof("Serialize succeed: %s", source)
	}
}

package analyzer

import (
	apis "hit.edu/framework/pkg/apis/cores"
	"testing"
)

func TestSerializeToJson(t *testing.T) {
	//node := apis.Node{
	//	Spec: apis.NodeSpec{
	//		NodeName: "TestNode",
	//	},
	//	Status: apis.NodeStatus{},
	//}

	node := apis.ContainerImage{
		Names:     []string{"container", "image", "sss"},
		SizeBytes: 123,
	}

	result, err := SerializeToJson(node)
	if err != nil {
		t.Errorf("Failed to serialize to json, error is \n %v", err)
	} else {
		t.Logf("Success to serialize to json, result is \n %v", result)
	}
}

func TestSerializeToYaml(t *testing.T) {
	node := apis.Node{
		Spec: apis.NodeSpec{
			NodeName: "TestNode",
		},
		Status: apis.NodeStatus{},
	}
	result, err := SerializeToYaml(node)
	if err != nil {
		t.Errorf("Failed to serialize to yaml, error is %v", err)
	} else {
		t.Logf("Success to serialize to yaml, result is \n%v", result)
	}
}

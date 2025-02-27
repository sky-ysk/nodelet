package utils

import (
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/component-base/logs"
	"testing"
)

func TestCheckResource(t *testing.T) {
	logs.Infof("[Test] testing CheckResource.....\n ")
	runtime, action := NewRuntimeAndActionResource()

	err := CheckResource(runtime, action)
	if err != nil {
		logs.Errorf("[Test] CheckResource err: %v\n", err)
	} else {
		logs.Infof("[Test] CheckResource is successful\n")
	}

}

func TestUpdateResourceStatus(t *testing.T) {
	logs.Infof("[Test] testing UpdateResourceStatus.....\n")
	runtime, action := NewRuntimeAndActionResource()
	err := UpdateResourceStatus(runtime, action)
	if err != nil {
		logs.Errorf("[Test] UpdateResourceStatus err: %v\n", err)
	} else {
		logs.Infof("[Test] UpdateResourceStatus is successful\n")
		for name, resource := range action.Status.Resources {
			logs.Infof("[Test] resource name is %s\n", name)
			logs.Infof("[Test] resource reserved value is %v\n", resource.Reserved)
		}
	}

}

func NewRuntimeAndActionResource() (*apis.Runtime, *apis.Action) {
	runtime := apis.Runtime{
		Resources: []apis.ResourceSpec{
			apis.ResourceSpec{
				Name:              "cpu",
				ExpectedValue:     2,
				ExpectedValueUnit: apis.ComputeCPU,
				Type:              apis.Compute,
			},
			apis.ResourceSpec{
				Name:              "memory",
				ExpectedValue:     1024,
				ExpectedValueUnit: apis.StorageMB,
				Type:              apis.Storage,
			},
			apis.ResourceSpec{
				Name:              "disk",
				ExpectedValue:     10,
				ExpectedValueUnit: apis.StorageGB,
				Type:              apis.Storage,
			},
		},
	}
	action := apis.Action{
		Status: apis.ActionStatus{
			Resources: map[string]apis.ResourceStatus{
				"cpu": apis.ResourceStatus{
					Reserved:     10,
					ReservedUnit: apis.ComputeCPU,
				},
				"memory": apis.ResourceStatus{
					Reserved:     4096,
					ReservedUnit: apis.StorageMB,
				},
				"disk": apis.ResourceStatus{
					Reserved:     200,
					ReservedUnit: apis.StorageGB,
				},
			},
		},
	}

	return &runtime, &action
}

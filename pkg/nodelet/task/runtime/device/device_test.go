package device

import (
	"fmt"
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/component-base/logs"
	"testing"
)

func TestRun(t *testing.T) {
	logs.Infof("[test] testing run.....\n")
	dr := NewDeviceRuntime()
	action := NewActionAndRuntime()
	err := dr.Run(&apis.Group{}, action)
	if err != nil {
		fmt.Println(err)
	}
}

func NewActionAndRuntime() *apis.Action {
	action := apis.Action{
		Spec: apis.ActionSpec{
			Name: "ActionTest",
		},
		Status: apis.ActionStatus{
			ActionID: "Action1",
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

			Devices: map[string]apis.DeviceStatus{
				"transferRobot": apis.DeviceStatus{
					Lock: apis.Lock{
						IsLocked: true,
					},
					Status:     "idle",
					Phase:      apis.DeviceIdle,
					InstanceID: "",
					ActionID:   "",
					DeviceID:   "d1",
				},
			},

			Scenes: map[string]apis.SceneStatus{
				"o1": apis.SceneStatus{
					AttachedTask:   "",
					AttachedDevice: "",
					AttachedScene:  "",
					Lock: apis.Lock{
						IsLocked: true,
					},
				},
				"p1": apis.SceneStatus{
					AttachedTask:   "",
					AttachedDevice: "",
					AttachedScene:  "",
					Lock: apis.Lock{
						IsLocked: true,
					},
				},
			},
		},
	}

	runtime := apis.Runtime{
		Image: "Move",
		Name:  "RuntimeTest",
		Devices: []apis.DeviceSpec{
			apis.DeviceSpec{
				Name:               "transferRobot",
				ExpectedProperties: map[string]apis.Property{},
				AccessMethod: apis.AccessMethod{
					Type:  apis.AccessByRmf,
					URL:   "http://192.168.1.225:8000",
					Group: "tinyRobot",
					Alias: "transferRobot",
				},
				Desc: apis.DeviceDesc{
					Label: []string{"Move"},
				},
			},
		},
		Outputs: make([]apis.Output, 0),
		Inputs: []apis.Input{
			apis.Input{
				Type:      apis.LocalData,
				Name:      "dest",
				Value:     "R201",
				ValueType: "string",
			},
			apis.Input{
				Type:      apis.LocalData,
				Name:      "orientation",
				Value:     "-3.12",
				ValueType: "double",
			},
			apis.Input{
				Type:      apis.LocalData,
				Name:      "dock",
				Value:     "true",
				ValueType: "bool",
			},
		},
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
		Scenes: []apis.SceneSpec{
			apis.SceneSpec{
				SceneID:          "o1",
				Type:             apis.ObjectType,
				ExpectedProperty: map[string]apis.Property{"position": apis.Property{Name: "position", Type: apis.StringType, Value: "R201"}},
			},
			apis.SceneSpec{
				SceneID:          "p1",
				Type:             apis.PositionType,
				ExpectedProperty: map[string]apis.Property{"isOccupied": apis.Property{Name: "isOccupied", Type: apis.BoolType, Value: "true"}},
			},
		},
	}
	action.Spec.Runtimes = []apis.Runtime{runtime}
	return &action
}

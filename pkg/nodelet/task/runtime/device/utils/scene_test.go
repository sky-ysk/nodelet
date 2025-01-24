package utils

import (
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/component-base/logs"
	"testing"
)

func TestCheckScene(t *testing.T) {
	logs.Infof("[test] testing Checkscene....\n")
	runtime, action := NewRuntimeAndActionScene()
	err := CheckScene(runtime, action)
	if err != nil {
		logs.Errorf("[test] CheckScene err: %v\n", err)
	} else {
		logs.Infof("[test] CheckScene is successful \n")
	}
}

func TestUpdateSceneStatus(t *testing.T) {
	logs.Infof("[test] testing UpdateSceneStatus...\n")
	runtime, action := NewRuntimeAndActionScene()
	err := UpdateSceneStatus(runtime, action, 1, "taskId_1")
	if err != nil {
		logs.Errorf("[test] UpdateSceneStatus err: %v\n", err)
	} else {
		logs.Infof("[test] UpdateSceneStatus is successful \n")
		for id, scene := range action.Status.Scenes {
			logs.Infof("[test] Scene id: %v\n", id)
			logs.Infof("[test] Scene %v's attached task is %v\n", id, scene.AttachedTask)
			logs.Infof("[Test] Scene %v's attached device is %v\n", id, scene.AttachedDevice)
			logs.Infof("[Test] Scene %v's time is %v\n", id, scene.UpdateTime)

		}
	}
}

func NewRuntimeAndActionScene() (*apis.Runtime, *apis.Action) {
	runtime := apis.Runtime{
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
		Devices: []apis.DeviceSpec{
			apis.DeviceSpec{
				Name: "transferRobot",
			},
		},
	}

	action := apis.Action{
		Status: apis.ActionStatus{
			Scenes: map[string]apis.SceneStatus{
				"o1": apis.SceneStatus{
					AttachedTask:   "",
					AttachedDevice: "",
					AttachedScene:  "",
					Lock: apis.Lock{
						Lock: true,
					},
				},
				"p1": apis.SceneStatus{
					AttachedTask:   "",
					AttachedDevice: "",
					AttachedScene:  "",
					Lock: apis.Lock{
						Lock: true,
					},
				},
			},
			Devices: map[string]apis.DeviceStatus{
				"transferRobot": apis.DeviceStatus{
					DeviceID: "d1",
				},
			},
		},
	}

	return &runtime, &action
}

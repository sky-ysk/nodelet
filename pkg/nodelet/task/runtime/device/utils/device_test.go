package utils

import (
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/component-base/logs"
	"testing"
)

func TestCheckDevice(t *testing.T) {
	logs.Infof("[Test] testing CheckDevice....\n")
	runtime, action := NewRuntimeAndActionDevice()
	err, devices := CheckDevice(runtime, action)
	if err != nil {
		logs.Errorf("[Test] CheckDevice err: %v", err)
	}
	logs.Infof("[Test] CheckDevice is successful\n")
	for name, device := range devices {
		logs.Infof("[Test] name is %v\n", name)
		logs.Infof("[Test] [%v] InstanceID is %v\n", name, device.Status.InstanceID)
		logs.Infof("[Test] [%v] Status is %v\n", name, device.Status.Status)
	}

}

func TestUpdateDeviceStatus(t *testing.T) {
	logs.Infof("[Test] testing UpdateDeviceStatus...\n")
	_, action := NewRuntimeAndActionDevice()
	err := UpdateDeviceStatus(action, "taskId_1")
	if err != nil {
		logs.Errorf("[Test] UpdateDeviceStatus err: %v", err)
	}
	logs.Infof("[Test] UpdateDeviceStatus is successful\n")
	for name, device := range action.Status.Devices {
		logs.Infof("[Test] name is %v\n", name)
		logs.Infof("[Test] [%v] InstanceID is %v\n", name, device.InstanceID)
		logs.Infof("[Test] [%v] Status is %v\n", name, device.Status)
		logs.Infof("[Test] [%v] Phase is %v\n", name, device.Phase)
		logs.Infof("[Test] [%v] ActionId is %v\n", name, device.ActionID)
	}
}

func NewRuntimeAndActionDevice() (*apis.Runtime, *apis.Action) {
	runtime := apis.Runtime{
		Devices: []apis.DeviceSpec{
			apis.DeviceSpec{
				Name: "transferRobot",
			},
		},
	}

	action := apis.Action{
		Status: apis.ActionStatus{
			ActionID: "action1",
			Devices: map[string]apis.DeviceStatus{
				"transferRobot": apis.DeviceStatus{
					Lock: apis.Lock{
						IsLocked: true,
					},
					Status:     "idle",
					Phase:      apis.DeviceIdle,
					InstanceID: "",
					ActionID:   "",
				},
			},
		},
	}
	return &runtime, &action
}

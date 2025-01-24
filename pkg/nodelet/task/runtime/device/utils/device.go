package utils

import (
	"fmt"
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/component-base/logs"
)

type DeviceWorker interface {
	// CheckDevice 检查Device状态，是否具有部署条件
	CheckDevice(runtime *apis.Runtime, action *apis.Action) (error, map[string]apis.Device)
	// UpdateDeviceStatus 更新Device状态
	UpdateDeviceStatus(runtime *apis.Runtime, action *apis.Action) error
}

// CheckDevice 检查设备情况
func CheckDevice(runtime *apis.Runtime, action *apis.Action) (error, map[string]apis.Device) {
	devices := make(map[string]apis.Device)
	for _, spec := range runtime.Devices {
		status := action.Status.Devices[spec.Name]
		// device必须已经被上锁（已经经过检查)
		//TODO:
		if !status.Lock.Lock {
			logs.Errorf("device %s is not locked", spec.Name)
			return fmt.Errorf("device %s is not locked", spec.Name), nil
		}

		// device状态为idle(系统内状态和运行时状态)
		if status.Status != "idle" || status.Phase != apis.DeviceIdle {
			logs.Errorf("device %s is busy", spec.Name)
			return fmt.Errorf("device %s is busy", spec.Name), nil
		}

		// device的Task ID应该为空
		if status.InstanceID != "" || status.ActionID != "" {
			logs.Errorf("device %s's task_id is not null", spec.Name)
			return fmt.Errorf("device %s's task_id is not null", spec.Name), nil
		}
		devices[spec.Name] = apis.Device{Spec: spec, Status: status}
	}

	return nil, devices
}

// UpdateDeviceStatus 更新DeviceStatus
func UpdateDeviceStatus(action *apis.Action, taskId string) error {

	devices := action.Status.Devices
	for name, device := range devices {
		logs.Infof("ddd name is %s", name)
		ds := apis.DeviceStatus{
			Phase:      apis.DeviceRunning,
			InstanceID: taskId,
			Status:     "running",
			ActionID:   action.Status.ActionID,
		}
		action.Status.Devices[name] = ds
		logs.Infof("device id is %s \n", device.DeviceID)
		logs.Infof("update device %s's status", name)
	}
	action.Status.Devices = devices
	return nil
}

package utils

import (
	"fmt"
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/component-base/logs"
	"time"
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
		err := CheckDeviceLock(&status)
		if err != nil {
			return err, nil
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

// GetDevices 获取所有设备
func GetDevices(runtime *apis.Runtime, action *apis.Action) (error, []apis.Device) {
	devices := make([]apis.Device, 0)
	for _, spec := range runtime.Devices {
		status := action.Status.Devices[spec.Name]
		// device必须已经被上锁（已经经过检查)
		//TODO:
		if !status.Lock.IsLocked {
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
		device := apis.Device{Spec: spec, Status: status}
		devices = append(devices, device)
	}

	return nil, devices
}

// UpdateDeviceStatus 更新DeviceStatus
func UpdateDeviceStatus(action *apis.Action, taskId string) error {

	devices := action.Status.Devices
	for name, device := range devices {
		logs.Infof("device name is %s\n", name)
		ds := apis.DeviceStatus{
			// 更新device的相应字段
			Phase:      apis.DeviceRunning,
			InstanceID: taskId,
			Status:     "running",
			ActionID:   action.Status.ActionID,
			Lock:       apis.Lock{Type: device.Lock.Type, IsLocked: true, Ref: device.Lock.Ref + 1},
			LastTime:   apis.Time{Time: time.Now()},

			// 不需要更新的字段直接复制
			DeviceID:   device.DeviceID,
			Events:     device.Events,
			Properties: device.Properties,
		}
		action.Status.Devices[name] = ds
		logs.Infof("update device %s's status\n", name)
	}
	action.Status.Devices = devices
	return nil
}

// RecoverDeviceStatus 恢复DeviceStatus的数据
func RecoverDeviceStatus(action *apis.Action) error {
	devices := action.Status.Devices
	for name, device := range devices {
		logs.Infof("device name is %s\n", name)
		ds := apis.DeviceStatus{
			// 更新device的相应字段
			Phase:      apis.DeviceIdle,
			InstanceID: "",
			Status:     "idle",
			ActionID:   "",
			Lock:       apis.Lock{Type: device.Lock.Type, IsLocked: false, Ref: device.Lock.Ref - 1},
			LastTime:   apis.Time{Time: time.Now()},

			// 不需要更新的字段直接复制
			DeviceID: device.DeviceID,

			// todo
			Events:     device.Events,
			Properties: device.Properties,
		}
		action.Status.Devices[name] = ds
		logs.Infof("update device %s's status\n", name)
	}
	action.Status.Devices = devices
	return nil
}

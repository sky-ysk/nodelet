package utils

import (
	"context"
	"fmt"
	apis "hit.edu/framework/pkg/apis/cores"
	metav1 "hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/client-go/clients/typed/core"
	"hit.edu/framework/pkg/component-base/logs"
	"strings"
	"time"
)

type DeviceWorker interface {
	// CheckDevice 检查Device状态，是否具有部署条件
	CheckDevice(runtime *apis.Runtime, action *apis.Action) (error, map[string]apis.Device)
	// UpdateDeviceStatus 更新Device状态
	UpdateDeviceStatus(runtime *apis.Runtime, action *apis.Action) error
}

// CheckDevice 检查设备情况
func CheckDevice(devices map[string]*apis.Device) error {

	for name, device := range devices {
		logs.Infof("check device %s status", name)
		status := device.Status
		if status.Lock.Lock == false {
			logs.Errorf("Device %s is not locked", device.Name)
			return fmt.Errorf("device %s is not locked", device.Name)
		}

		// device状态为idle(系统内状态和运行时状态)
		if status.Status != "idle" || (status.Phase != apis.DeviceIdle && status.Phase != apis.DeviceInit) {
			logs.Errorf("device %s is busy", device.Name)
			return fmt.Errorf("device %s is busy", device.Name)
		}

		// device的Task ID应该为空
		if status.InstanceID != "" || status.ActionID != "" {
			logs.Errorf("device %s's task_id is not null", device.Name)
			return fmt.Errorf("device %s's task_id is not null", device.Name)
		}
	}

	return nil
}

// GetDevices 获取所有设备
func GetDevices(runtime *apis.Runtime, action *apis.Action) (error, []apis.Device) {
	devices := make([]apis.Device, 0)
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
		device := apis.Device{Spec: spec, Status: status}
		devices = append(devices, device)
	}

	return nil, devices
}

// GetDevices 获取所有设备
func ObtainDevices(runtime *apis.Runtime, action *apis.Action) (error, []apis.Device) {
	devices := make([]apis.Device, 0)
	for _, spec := range runtime.Devices {
		status := action.Status.Devices[spec.Name]
		device := apis.Device{Spec: spec, Status: status}
		devices = append(devices, device)
	}

	return nil, devices
}

// UpdateDeviceStatus 更新DeviceStatus
func UpdateDeviceStatusList(runtime *apis.Runtime, action *apis.Action, taskId string, deviceMap map[string]*apis.Device, deviceClient core.DeviceInterface) error {

	for _, spec := range runtime.Devices {
		status := action.Status.Devices[spec.Name]
		name := spec.Name
		logs.Infof("device name is %s\n", name)
		status.Status = "running"
		status.Phase = apis.DeviceRunning
		status.InstanceID = taskId
		status.Lock = apis.Lock{Type: status.Lock.Type, Lock: true, Ref: status.Lock.Ref}
		status.LastTime = apis.Time{Time: time.Now()}

		action.Status.Devices[spec.Name] = status

		newDevice := &apis.Device{
			ObjectMeta: metav1.ObjectMeta{
				Name:      spec.Name,
				Namespace: "test",
				Labels: map[string]string{
					"environment": "dev",
				},
			},
			TypeMeta: metav1.TypeMeta{
				Kind:       "Device",
				APIVersion: "resources/v1",
			},
			Spec:   spec,
			Status: status,
		}
		_, err := deviceClient.Update(context.TODO(), newDevice, metav1.UpdateOptions{})
		if err != nil {
			logs.Errorf("update device %s status failed, %s", name, err)
			return err
		}
		logs.Infof("update device %s's status\n", name)
	}

	return nil
}

// UpdateDeviceRunning 用于在发布任务指令成功后(但是还不知道业务执行情况)时 更新device的状态
func UpdateDeviceRunning(runtime *apis.Runtime, action *apis.Action, device *apis.Device, deviceClient core.DeviceInterface) error {
	// 绑定ActionID
	device.Status.ActionID = action.Status.ActionID
	// 将phase更改为running
	device.Status.Phase = apis.DeviceRunning
	// 设置更新时间
	device.Status.LastTime = apis.Time{Time: time.Now()}
	_, err := deviceClient.Update(context.TODO(), device, metav1.UpdateOptions{})
	if err != nil {
		logs.Errorf("update device [%s]  failed, %s", device.Name, err)
		return err
	}
	logs.Infof("update device [%s] successfully\n", device.Name)
	return nil
}

func UpdateDeviceSuccess(runtime *apis.Runtime, action *apis.Action, device *apis.Device) error {
	// Lock

	// ActionID更改为空
	device.Status.ActionID = ""
	// 设置更新时间
	device.Status.LastTime = apis.Time{Time: time.Now()}
	// 更新phase
	device.Status.Phase = apis.DeviceIdle

}

func UpdateDevice(runtime *apis.Runtime, device *apis.Device, taskId string, deviceClient core.DeviceInterface) error {
	parts := strings.Split(runtime.Name, "_")
	if parts[0] == "manage" {
		device.Status.Status = "Init"
		device.Status.Phase = apis.DeviceInit
	} else if parts[0] == "service" {
		device.Status.Status = "Running"
		device.Status.Phase = apis.DeviceRunning
	}
	device.Status.InstanceID = taskId

	_, err := deviceClient.Update(context.TODO(), device, metav1.UpdateOptions{})
	if err != nil {
		logs.Errorf("update device %s  failed, %s", device.Name, err)
		return err
	}

	return nil
}

func UpdateDeviceStatusFailed(runtime *apis.Runtime, action *apis.Action, device *apis.Device, deviceClient core.DeviceInterface) error {
	device.Status.Status = "failed"
	device.Status.Phase = apis.DeviceError
	device.Status.LastTime = apis.Time{time.Now()}
	if _, err := deviceClient.Update(context.TODO(), device, metav1.UpdateOptions{}); err != nil {
		logs.Errorf("update device %s status failed, %s", device.Name, err)
		return err
	}

	for _, spec := range runtime.Devices {
		if spec.Name == device.Spec.Name {
			action.Status.Devices[spec.Name] = device.Status

		}
	}
	return nil
}

func UpdateDeviceStatusCompleted(runtime *apis.Runtime, action *apis.Action, device *apis.Device, deviceClient core.DeviceInterface) error {
	device.Status.Status = "completed"
	//device.Status.Phase = apis.DeviceComplete
	device.Status.LastTime = apis.Time{time.Now()}
	if _, err := deviceClient.Update(context.TODO(), device, metav1.UpdateOptions{}); err != nil {
		logs.Errorf("update device %s status failed, %s", device.Name, err)
		return err
	}

	for _, spec := range runtime.Devices {
		if spec.Name == device.Spec.Name {
			action.Status.Devices[spec.Name] = device.Status

		}
	}
	return nil
}

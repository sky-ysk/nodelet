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
func CheckDevice1(devices map[string]*apis.Device) error {

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

func CheckDevice2(devices map[string]*apis.Device, groupID string, deviceClient core.DeviceInterface) error {

	var err error = nil
	for name, device := range devices {
		logs.Infof("check device %s status", name)

		for {

			status := device.Status
			// 首先判断GroupID
			logs.Infof("check device groupID...")
			if status.GroupID != groupID {
				// GroupID 不通过 说明device没有被分配到这个Group中
				logs.Warnf("device %s's GroupID is wrong", device.Name)
				time.Sleep(3 * time.Second)

				// 重新获取device
				device, err = deviceClient.Get(context.TODO(), device.Name, metav1.GetOptions{})
				if err != nil {
					logs.Errorf("get device:%s failed", name)
					return err
				}

			} else {
				// GroupID合格 证明device绑定的Group是对的
				// 判断是否上锁
				if status.Lock.Lock == false {
					logs.Errorf("Device %s is not locked", device.Name)
					return fmt.Errorf("device %s is not locked", device.Name)
				}

				// device状态为idle(系统内状态和运行时状态)
				if status.Phase != apis.DeviceIdle {
					logs.Errorf("device %s is busy", device.Name)
					return fmt.Errorf("device %s is busy", device.Name)
				}

				// device的Task ID应该为空
				if status.ActionID != "" {
					logs.Errorf("device %s's ActionID is not null", device.Name)
					return fmt.Errorf("device %s's ActionID is not null", device.Name)
				}
				break
			}

		}
		devices[name] = device
	}

	return err
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
		logs.Errorf("update device [%s]  failed[stage running], %s", device.Name, err)
		return err
	}
	logs.Infof("update device [%s] successfully[stage running]\n", device.Name)
	return nil
}

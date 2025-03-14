package utils

import (
	"fmt"
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/component-base/logs"
)

type LockManager struct {
}

func NewLockManager() *LockManager {
	return &LockManager{}
}

// CheckDeviceLock 检查Device锁的状态
func (lm *LockManager) CheckDeviceLock(status *apis.DeviceStatus) error {
	if status.Lock.IsLocked == false {
		logs.Errorf("Device %s is not locked", status.DeviceID)
		return fmt.Errorf("Device %s is not locked\n", status.DeviceID)
	}
	return nil
}

// CheckSceneLock 检查Scene锁的状态
func (lm *LockManager) CheckSceneLock(status *apis.SceneStatus) error {
	if status.Lock.IsLocked == false {
		logs.Errorf("Scene is not locked")
		return fmt.Errorf("Deviceis not locked\n")
	}
	return nil
}

// ReleaseSceneLock 释放一次引用 Ref为0解锁
func (lm *LockManager) ReleaseSceneLock(status *apis.SceneStatus) error {
	status.Lock.Ref -= 1
	if status.Lock.Ref == 0 {
		status.Lock.IsLocked = false
	}
	return nil
}

// ReleaseDeviceLock 释放一次引用 Ref为0解锁
func (lm *LockManager) ReleaseDeviceLock(device *apis.Device, action *apis.Action, runtime *apis.Runtime) error {
	device.Status.Lock.Ref -= 1
	for index, spec := range runtime.Devices {
		if device.Name == spec.Name {
			if device.Status.Lock.Ref == 0 {
				// 查看父设备的Ref
				if action.Status.Devices[index].Lock.Ref == 0 {
					device.Status.Lock.IsLocked = false
				}
			}
		}

	}

	return nil
}

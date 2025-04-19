package utils

import (
	"context"
	"fmt"
	apis "hit.edu/framework/pkg/apis/cores"
	metav1 "hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/client-go/clients/typed/core"
	"hit.edu/framework/pkg/component-base/logs"
	"time"
)

func CheckDevices(deviceMap map[string]*apis.Device, specs []apis.DeviceSpec) error {
	for name, device := range deviceMap {
		logs.Infof("[DEVICE RUNTIME] Check Device[%s] ", name)
		// 检查device的GroupID phase
		if device.Status.Phase != apis.DeviceIdle {
			logs.Errorf("[DEVICE RUNTIME] Device[%s] is not IDLE", name)
			return fmt.Errorf("error! Device[%s] is not IDLE", name)
		}
		// 检查是否上锁
		if device.Status.Lock.IsLocked != true {
			logs.Errorf("[DEVICE RUNTIME] Device[%s] is not locked", name)
			return fmt.Errorf("error! Device[%s] is not locked", name)
		}
		// 检查能力的状态
		for _, spec := range specs {
			for _, abilityName := range spec.Abilities {
				if ability, ok := device.Status.Abilities[abilityName]; ok {
					state := apis.AbilityRunning
					if ability.State != state {

					}
				}
			}

		}

	}
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

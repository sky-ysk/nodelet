package utils

import (
	"encoding/json"
	"fmt"
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/client-go/util/manager"
	"hit.edu/framework/pkg/component-base/logs"
	"time"
)

func CheckDevices(deviceMap map[string]*apis.Device, specs []apis.DeviceSpec, m *manager.Manager) error {
	// 遍历device
	for name, device := range deviceMap {
		logs.Infof("[DEVICE RUNTIME] Check Device[%s] ", name)
		// 检查device的GroupID phase
		if device.Status.Phase != apis.DeviceIdle {
			logs.Errorf("[DEVICE RUNTIME] Device[%s] is not IDLE", name)
			return fmt.Errorf("error! Device[%s] is not IDLE", name)
		}
		logs.Infof("[DEVICE RUNTIME] Check Device[%s] Phase is NORMAL", device.Name)
		// 检查是否上锁
		if device.Status.Lock.IsLocked != true {
			logs.Errorf("[DEVICE RUNTIME] Device[%s] is not locked", name)
			return fmt.Errorf("error! Device[%s] is not locked", name)
		}
		logs.Infof("[DEVICE RUNTIME] Check Device[%s] Lock is NORMAL", name)
		// 检查能力的状态
		for _, spec := range specs {
			for _, abilityName := range spec.Abilities {
				if ability, ok := device.Status.Abilities[abilityName]; ok {
					switch ability.Status {
					case apis.AbilityRunning:
						logs.Infof("[DEVICE RUNTIME] Device[%s] Ability[%s] is running, NORMAL", name, ability.Name)
					case apis.AbilityReadyStartUp:
						for {
							logs.Warnf("[DEVICE RUNTIME] Device[%s] Ability[%s] is ready STARTUP, Waiting!", name, ability.Name)
							time.Sleep(1 * time.Second)
							var err error
							device, err = m.GetDevice(name, device.Namespace)
							if err != nil {
								logs.Errorf("[DEVICE RUNTIME] Device[%s] Get Device Error: %v", name, err)
								return err
							}
							if device.Status.Abilities[abilityName].Status == apis.AbilityRunning {
								logs.Infof("[DEVICE RUNTIME] Device[%s] Ability[%s] is running, NORMAL]", name, ability.Name)
								deviceMap[name] = device
								break
							}
						}
					case apis.AbilityError:
						logs.Errorf("[DEVICE RUNTIME] Device[%s] Ability[%s] is ERROR!", name, ability.Name)
						return fmt.Errorf("error! Device[%s] Ability[%s] is ERROR", name, ability.Name)
					}
				}
			}

		}

	}
	return nil
}

// UpdateDeviceRunning 用于在发布任务指令成功后(但是还不知道业务执行情况)时 更新device的状态
func UpdateDeviceRunning(deviceMap map[string]*apis.Device, clientManager *manager.Manager) error {

	for name, device := range deviceMap {
		logs.Infof("[DEVICE RUNTIME] Update Device[%s] stage[RUNNING]", name)
		// 将phase更改为running
		device.Status.Phase = apis.DeviceRunning
		// 设置更新时间
		device.Status.LastTime = apis.Time{Time: time.Now()}
		patchDevice, err := json.Marshal(map[string]interface{}{
			"status": map[string]interface{}{
				"phase":     device.Status.Phase,
				"last_time": device.Status.LastTime,
			},
		})
		_, err = clientManager.PatchDevice(device.Name, device.Namespace, string(patchDevice))
		if err != nil {
			logs.Errorf("[DEVICE RUNTIME] Update Device[%s] Failed stage [RUNNING], err:%s", device.Name, err.Error())
			return err
		}
		logs.Infof("[DEVICE RUNTIME] Update Device[%s] Successfully stage [RUNNING]\n", device.Name)
	}

	return nil
}

// UpdateDeviceFinished 用于在已经获得任务的执行状态 更新device的状态
func UpdateDeviceFinished(deviceMap map[string]*apis.Device, clientManager *manager.Manager) error {

	for name, device := range deviceMap {

		logs.Infof("[DEVICE RUNTIME] Update Device[%s] stage[FINISHED]", name)
		device.Status.Lock.Ref -= 1
		if device.Status.Lock.Ref == 0 {
			device.Status.Lock.IsLocked = false
		}
		// 将phase更改为running
		device.Status.Phase = apis.DeviceIdle
		// 设置更新时间
		device.Status.LastTime = apis.Time{Time: time.Now()}
		patchDevice, err := json.Marshal(map[string]interface{}{
			"status": map[string]interface{}{
				"lock":      device.Status.Lock,
				"phase":     device.Status.Phase,
				"last_time": device.Status.LastTime,
			},
		})

		_, err = clientManager.PatchDevice(device.Name, device.Namespace, string(patchDevice))
		if err != nil {
			logs.Errorf("[DEVICE RUNTIME] Update Device[%s] stage[FINISHED], err:%s", device.Name, err)
			return err
		}
		logs.Infof("[DEVICE RUNTIME] Update Device[%s] successfully stage [FINISHED]\n", device.Name)
	}

	return nil
}

// UpdateDeviceError 用于在已经获得任务的执行状态 更新device的状态
func UpdateDeviceError(deviceMap map[string]*apis.Device, clientManager *manager.Manager) error {

	for name, device := range deviceMap {

		logs.Infof("[DEVICE RUNTIME] Update Device[%s] stage[ERROR]", name)
		device.Status.Lock.Ref -= 1
		if device.Status.Lock.Ref == 0 {
			device.Status.Lock.IsLocked = false
		}
		// 将phase更改为running
		device.Status.Phase = apis.DeviceIdle
		// 设置更新时间
		device.Status.LastTime = apis.Time{Time: time.Now()}
		patchDevice, err := json.Marshal(map[string]interface{}{
			"status": map[string]interface{}{
				"lock":      device.Status.Lock,
				"phase":     device.Status.Phase,
				"last_time": device.Status.LastTime,
			},
		})

		_, err = clientManager.PatchDevice(device.Name, device.Namespace, string(patchDevice))
		if err != nil {
			logs.Errorf("[DEVICE RUNTIME] Update Device[%s] stage[ERROR], err:%s", device.Name, err)
			return err
		}
		logs.Infof("[DEVICE RUNTIME] Update Device[%s] successfully stage [ERROR]\n", device.Name)
	}

	return nil
}

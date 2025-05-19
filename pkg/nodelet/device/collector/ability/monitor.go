package manager

import (
	"encoding/json"
	"fmt"
	apis "hit.edu/framework/pkg/apis/cores"
	m "hit.edu/framework/pkg/client-go/util/manager"
	"hit.edu/framework/pkg/component-base/logs"
	"strconv"
	"sync"
)

type Managers struct {
	AbilityManagers map[string]*ManagerOfAbility
	IDList          map[string]bool
	Mutex           sync.Mutex
}

func NewManagers() *Managers {
	return &Managers{
		AbilityManagers: make(map[string]*ManagerOfAbility),
		IDList:          make(map[string]bool),
	}
}

// MonitorAllDevices 检测所有设备的在线情况
func MonitorAllDevices(clientManager *m.Manager) error {
	logs.Infof("[DEVICE EXPORTER-DEVICE MONITOR] Try to Get All Devices")
	deviceList, err := clientManager.GetDevices("", "test")
	if err != nil {
		logs.Errorf("[DEVICE EXPORTER-DEVICE MONITOR] Get All Devices failed err: %s", err.Error())
		return err
	}
	logs.Infof("[DEVICE EXPORTER-DEVICE MONITOR] Get All Devices Success!")
	logs.Infof("[DEVICE EXPORTER-DEVICE MONITOR] ETCD Has %d Devices", len(deviceList.Items))

	// 并发同步控制
	var wg sync.WaitGroup
	// 错误处理通道
	errChan := make(chan error, 10)

	// 遍历所有的Device
	for _, d := range deviceList.Items {
		device := d
		if device.Name == "deviceLock" {
			return nil
		}
		// Ability类型的device
		if device.Spec.AccessMethod.Type == apis.AccessByAbility {
			wg.Add(1)

			// 每个Device单独开一个协程
			go func() {
				defer wg.Done()
				devicePhase := device.Status.Phase
				switch devicePhase {
				case apis.DeviceDisconnected: // 离线状态
					DeviceDisconnectedHandle(device, clientManager, errChan)

				case apis.DeviceRunning, apis.DeviceIdle: // 运行状态
					DeviceOnlineHandle(device, clientManager, errChan)

				case apis.DeviceError: // 错误状态

				}
			}()
		}

	}
	// 等待所有协程完成
	go func() {
		wg.Wait()
		close(errChan)
	}()
	// 收集所有错误
	var errs []error
	for e := range errChan {
		errs = append(errs, e)
	}

	if len(errs) > 0 {
		return fmt.Errorf("device processing errors: %v", errs)
	}
	return nil
}

func MonitorAllAbilities(clientManager *m.Manager) error {
	// 首先获取所有的Devices
	logs.Tracef("[DEVICE EXPORTER-ABILITY MONITOR] Try to Get All Devices")
	deviceList, err := clientManager.GetDevices("", "test")
	if err != nil {
		logs.Errorf("[DEVICE EXPORTER] Get All Devices failed err: %s", err.Error())
		return err
	}
	logs.Tracef("[DEVICE EXPORTER-ABILITY MONITOR] Get All Devices Success!")
	logs.Tracef("[DEVICE EXPORTER-ABILITY MONITOR] ETCD Has %d Devices", len(deviceList.Items))

	// 外层并发同步控制（用于设备）
	var deviceWG sync.WaitGroup
	// 错误处理通道
	errChan := make(chan error, 10)

	// 遍历所有的Device
	for _, d := range deviceList.Items {
		device := d

		// Ability类型的device
		if device.Spec.AccessMethod.Type == apis.AccessByAbility {
			deviceWG.Add(1)

			go func(device apis.Device) {
				defer deviceWG.Done()

				// 内层并发同步控制（用于设备中的每个能力）
				var abilityWG sync.WaitGroup
				// 用于存储设备中所有能力的最终状态
				updatedAbilities := make(map[string]apis.Ability)

				// 遍历设备的所有能力
				for name, ability := range device.Status.Abilities {
					abilityWG.Add(1)

					go func(name string, ability apis.Ability) {
						defer abilityWG.Done()

						// 处理单个能力
						updatedAbility, err := processAbility(device, name, ability)
						if err != nil {
							errChan <- err
							return
						}
						updatedAbilities[name] = updatedAbility
					}(name, ability)
				}

				// 等待所有能力处理完成
				abilityWG.Wait()

				// 更新设备状态
				if err := patchDeviceStatus(clientManager, device, updatedAbilities); err != nil {
					errChan <- err
				}
			}(device)
		}
	}

	// 等待所有设备处理完成
	go func() {
		deviceWG.Wait()
		close(errChan)
	}()

	// 收集所有错误
	var errs []error
	for e := range errChan {
		errs = append(errs, e)
	}

	if len(errs) > 0 {
		return fmt.Errorf("device processing errors: %v", errs)
	}
	return nil
}

// processAbility 处理单个设备的单个能力
func processAbility(device apis.Device, name string, ability apis.Ability) (apis.Ability, error) {
	// 如果设备处于空闲状态
	if device.Status.Phase == apis.DeviceIdle {
		logs.Tracef("[DEVICE EXPORTER-ABILITY MONITOR] Monitor Device[%s] Ability[%s]", device.Name, ability.Name)
		if ability.Status == apis.AbilityReadyStartUp { // 如果ability需要被拉起
			logs.Infof("[DEVICE EXPORTER-ABILITY MONITOR] Device[%s] Ability[%s] is ReadyStartUp", device.Name, ability.Name)
			am := NewAbilityManager(device.Spec.AccessMethod.URL, ability.Name)
			err := am.BindUUID()
			if err != nil {
				logs.Errorf("[DEVICE EXPORTER-ABILITY MONITOR] Device[%s] Ability[%s] Bind uuid failed, err:%s", device.Name, ability.Name, err.Error())
				return ability, err
			}

			var hb HeartBeat
			hb, err = am.StartupAbility()
			if err != nil {
				logs.Errorf("[DEVICE EXPORTER-ABILITY MONITOR] Device[%s] Ability[%s] Startup failed, err:%s", device.Name, ability.Name, err.Error())
				return ability, err
			}

			logs.Infof("[DEVICE EXPORTER-ABILITY MONITOR] Device[%s] Ability[%s] Startup Success!", device.Name, ability.Name)
			ability.Status = apis.AbilityRunning
			for sn, service := range ability.Services {
				port := strconv.Itoa(hb.AbilityPort)
				service.Port = &port
				ability.Services[sn] = service
			}
		}
	}
	return ability, nil
}

// patchDeviceStatus 更新设备状态
func patchDeviceStatus(clientManager *m.Manager, device apis.Device, abilities map[string]apis.Ability) error {
	device.Status.Abilities = abilities
	patchDevice, err := json.Marshal(map[string]interface{}{
		"status": map[string]interface{}{
			"abilities": device.Status.Abilities,
		},
	})
	if err != nil {
		logs.Errorf("[DEVICE EXPORTER-ABILITY MONITOR] Marshal Device[%s] failed, err:%s", device.Name, err.Error())
		return err
	}

	_, err = clientManager.PatchDevice(device.Name, device.Namespace, string(patchDevice))
	if err != nil {
		logs.Errorf("[DEVICE EXPORTER-ABILITY MONITOR] Patch Device[%s] failed, err:%s", device.Name, err.Error())
		return err
	}
	logs.Tracef("[DEVICE EXPORTER-ABILITY MONITOR] Patch Device[%s] Success!", device.Name)
	return nil
}

// DeviceDisconnectedHandle 设备处于Disconnected状态时的处理
func DeviceDisconnectedHandle(device apis.Device, clientManager *m.Manager, errChan chan error) {
	logs.Infof("[DEVICE EXPORTER-DEVICE MONITOR] Device[%s] is Disconnected...", device.Name)
	logs.Infof("[DEVICE EXPORTER-DEVICE MONITOR] Try to Connect Device[%s]", device.Name)
	url := device.Spec.AccessMethod.URL
	// 尝试连接能力框架
	_, err := GetAbilityInstances(url)
	if err != nil {
		// 连接失败依旧是Disconnected
		logs.Errorf("[DEVICE EXPORTER-DEVICE MONITOR] Connect Device[%s] Failed", device.Name)
		logs.Infof("[DEVICE EXPORTER-DEVICE MONITOR] Device[%s] is still Disconnected...", device.Name)
		return
	}
	// 连接成功 转变为Idle状态
	logs.Infof("[DEVICE EXPORTER-DEVICE MONITOR] Connect Device[%s] Successfully...", device.Name)
	logs.Infof("[DEVICE EXPORTER-DEVICE MONITOR] Device[%s] is Online", device.Name)
	logs.Infof("[DEVICE EXPORTER-DEVICE MONITOR] Try to Update Device[%s]", device.Name)
	device.Status.Phase = apis.DeviceIdle
	var patchDevice []byte
	patchDevice, err = json.Marshal(map[string]interface{}{
		"status": map[string]interface{}{
			"phase": device.Status.Phase,
		},
	})
	if err != nil {
		logs.Errorf("[DEVICE EXPORTER-DEVICE MONITOR] Marshal Device[%s] failed, err:%s", device.Name, err.Error())
		errChan <- err
		return
	}
	_, err = clientManager.PatchDevice(device.Name, device.Namespace, string(patchDevice))
	if err != nil {
		logs.Errorf("[DEVICE EXPORTER-DEVICE MONITOR] Patch Device[%s] failed, err:%s", device.Name, err.Error())
		errChan <- err
		return
	}
	logs.Infof("[DEVICE EXPORTER-DEVICE MONITOR] Patch Device[%s] Successfully", device.Name)
	return
}

// DeviceOnlineHandle 设备处于在线状态时的处理
func DeviceOnlineHandle(device apis.Device, clientManager *m.Manager, errChan chan error) {
	logs.Infof("[DEVICE EXPORTER] Device[%s] is Online...", device.Name)
	url := device.Spec.AccessMethod.URL
	_, err := GetAbilityInstances(url)
	if err != nil {
		logs.Errorf("[DEVICE EXPORTER-DEVICE MONITOR] Connect Device[%s] Failed", device.Name)
		logs.Infof("[DEVICE EXPORTER-DEVICE MONITOR] Now Device[%s] is Disconnected...", device.Name)
		var patchDevice []byte
		patchDevice, err = json.Marshal(map[string]interface{}{
			"status": map[string]interface{}{
				"phase": apis.DeviceDisconnected,
			},
		})
		if err != nil {
			logs.Errorf("[DEVICE EXPORTER-DEVICE MONITOR] Marshal Device[%s] failed, err:%s", device.Name, err.Error())
			errChan <- err
			return
		}
		_, err = clientManager.PatchDevice(device.Name, device.Namespace, string(patchDevice))
		if err != nil {
			logs.Errorf("[DEVICE EXPORTER-DEVICE MONITOR] Patch Device[%s] failed, err:%s", device.Name, err.Error())
			errChan <- err
			return
		}
		logs.Infof("[DEVICE EXPORTER-DEVICE MONITOR] Patch Device[%s] Successfully", device.Name)
		return
	}
	logs.Infof("[DEVICE EXPORTER-DEVICE MONITOR] Device[%s] is still Running...", device.Name)
	return
}

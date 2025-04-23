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
				//case apis.DeviceReadyStartUp: // 准备启动状态
				//	logs.Infof("[DEVICE EXPORTER] Device[%s] is Ready To StartUp...", device.Name)
				//	logs.Infof("[DEVICE EXPORTER] Try to StartUp Device[%s]...", device.Name)
				//	// 获取URL和Name
				//	url := device.Spec.AccessMethod.URL
				//	class := device.Spec.Abilities[0]
				//	abilityName := device.Status.Abilities[class].Name
				//	// 创建AbilityManager
				//	am := NewAbilityManager(url, abilityName)
				//	err = am.BindUUID()
				//	if err != nil {
				//		logs.Errorf("[DEVICE EXPORTER] Bind uuid fail")
				//		errChan <- err
				//		return
				//	}
				//	// 首先先判断一下设备是否在线，不在线才能进行拉起设备的操作
				//	var flag bool
				//	flag, err = am.IsOnline()
				//	if err != nil {
				//		logs.Errorf("[Device EXPORTER] Device[%s] Judge Online fail", device.Name)
				//		return
				//	}
				//	if flag { // 如果设备在线说明错误
				//		logs.Errorf("[DEVICE EXPORTER] Device[%s] is Online", device.Name)
				//		errChan <- fmt.Errorf("[DEVICE EXPORTER] Device[%s] is Online, It should be Offline", device.Name)
				//		return
				//	} else { // 如果设备不在线是正常的
				//		logs.Infof("[DEVICE EXPORTER] Device[%s] is Offline, Normal", device.Name)
				//		logs.Infof("[DEVICE EXPORTER] Device[%s] Try to StartUp...", device.Name)
				//		// 尝试拉起设备
				//		var hb HeartBeat
				//		hb, err = am.StartupAbility()
				//		if err != nil {
				//			logs.Errorf("[DEVICE EXPORTER] Device[%s] Startup fail", device.Name)
				//			errChan <- err
				//			return
				//		}
				//		logs.Infof("[DEVICE EXPORTER] Device[%s] Try to StartUp Successfully", device.Name)
				//		// 拉起成功 更新设备字段
				//		device.Status.Phase = apis.DeviceIdle
				//		for index, service := range device.Status.Abilities[class].Services {
				//			port := new(string)
				//			*port = strconv.Itoa(hb.AbilityPort)
				//			service.Port = port
				//			device.Status.Abilities[class].Services[index] = service
				//		}
				//		_, err = deviceClient.Get(context.TODO(), device.Name, metav1.GetOptions{})
				//		if err != nil {
				//			logs.Errorf("[DEVICE EXPORTER] Get device[%s] fail", device.Name)
				//			errChan <- err
				//			return
				//		}
				//		_, err = deviceClient.Update(context.TODO(), &device, metav1.UpdateOptions{})
				//		if err != nil {
				//			logs.Errorf("[DEVICE EXPORTER] update device fail...")
				//			errChan <- err
				//			return
				//		}
				//	}
				//
				//case apis.DeviceReadyClose: // 准备关闭状态
				//	logs.Infof("[DEVICE EXPORTER] Device[%s] is Ready To Terminate...", device.Name)
				//	logs.Infof("[DEVICE EXPORTER] Try to Terminate Device[%s]...", device.Name)
				//	// 获取URL和Name
				//	url := device.Spec.AccessMethod.URL
				//	class := device.Spec.Abilities[0]
				//	abilityName := device.Status.Abilities[class].Name
				//	// 创建AbilityManager
				//	am := NewAbilityManager(url, abilityName)
				//	err = am.BindUUID()
				//	if err != nil {
				//		logs.Errorf("[DEVICE EXPORTER] Bind uuid fail")
				//		errChan <- err
				//		return
				//	}
				//	// 首先先判断一下设备是否在线，在线才能进行关闭设备的操作
				//	var flag bool
				//	flag, err = am.IsOnline()
				//	if err != nil {
				//		logs.Errorf("[Device EXPORTER] Device[%s] Judge Online fail", device.Name)
				//		return
				//	}
				//	if !flag { // 如果设备不在线说明不正常
				//		logs.Errorf("[DEVICE EXPORTER] Device[%s] is Offline", device.Name)
				//		errChan <- fmt.Errorf("[DEVICE EXPORTER] Device[%s] is Offline, It should be Online", device.Name)
				//		return
				//	} else { // 设备在线说明正常
				//		logs.Infof("[DEVICE EXPORTER] Device[%s] is Online, Normal", device.Name)
				//		logs.Infof("[DEVICE EXPORTER] Device[%s] Try to Terminate...", device.Name)
				//		err = am.TerminateAbility()
				//		if err != nil {
				//			logs.Infof("[DEVICE EXPORTER] Device[%s] Terminate fail", device.Name)
				//			errChan <- err
				//		}
				//		logs.Infof("[DEVICE EXPORTER] Device[%s] Terminate successfully...", device.Name)
				//		// 拉起成功 更新设备字段
				//		device.Status.Phase = apis.DeviceDisconnected
				//		for index, service := range device.Status.Abilities[class].Services {
				//			service.Port = nil
				//			device.Status.Abilities[class].Services[index] = service
				//		}
				//		_, err = deviceClient.Get(context.TODO(), device.Name, metav1.GetOptions{})
				//		if err != nil {
				//			logs.Errorf("[DEVICE EXPORTER] Get device[%s] fail", device.Name)
				//			errChan <- err
				//			return
				//		}
				//		_, err = deviceClient.Update(context.TODO(), &device, metav1.UpdateOptions{})
				//		if err != nil {
				//			logs.Errorf("[DEVICE EXPORTER] update device fail...")
				//			errChan <- err
				//			return
				//		}
				//	}

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
	logs.Infof("[DEVICE EXPORTER-ABILITY MONITOR] Try to Get All Devices")
	deviceList, err := clientManager.GetDevices("", "test")
	if err != nil {
		logs.Errorf("[DEVICE EXPORTER] Get All Devices failed err: %s", err.Error())
		return err
	}
	logs.Infof("[DEVICE EXPORTER-ABILITY MONITOR] Get All Devices Success!")
	logs.Infof("[DEVICE EXPORTER-ABILITY MONITOR] ETCD Has %d Devices", len(deviceList.Items))
	// 并发同步控制
	var wg sync.WaitGroup
	// 错误处理通道
	errChan := make(chan error, 10)

	// 遍历所有的Device
	for _, d := range deviceList.Items {
		device := d
		// Ability类型的device
		if device.Spec.AccessMethod.Type == apis.AccessByAbility {
			wg.Add(1)

			// 每个Device单独开一个协程
			go func() {
				defer wg.Done()
				devicePhase := device.Status.Phase
				if devicePhase == apis.DeviceIdle { // 当phase为IDLE
					url := device.Spec.AccessMethod.URL
					for name, ability := range device.Status.Abilities {
						logs.Infof("[DEVICE EXPORTER-ABILITY MONITOR] Monitor Device[%s] Ability[%s]", device.Name, ability.Name)
						if ability.Status == apis.AbilityReadyStartUp { // 如果ability需要被拉起就给他拉起
							am := NewAbilityManager(url, ability.Name)
							err = am.BindUUID()
							if err != nil {
								logs.Errorf("[DEVICE EXPORTER-ABILITY MONITOR] Device[%s] Ability[%s] Bind uuid failed, err:%s", device.Name, ability.Name, err.Error())
								errChan <- err
								return
							}
							var hb HeartBeat
							hb, err = am.StartupAbility()
							if err != nil {
								logs.Errorf("[DEVICE EXPORTER-ABILITY MONITOR] Device[%s] Ability[%s] Startup failed, err:%s", device.Name, ability.Name, err.Error())
								errChan <- err
								return
							}
							logs.Infof("[DEVICE EXPORTER-ABILITY MONITOR] Device[%s] Ability[%s] Startup Success!", device.Name, ability.Name)
							ability.Status = apis.AbilityRunning
							for sn, service := range ability.Services {
								port := strconv.Itoa(hb.AbilityPort)
								service.Port = &port
								ability.Services[sn] = service
							}
							device.Status.Abilities[name] = ability
						}
					}
					var patchDevice []byte
					// 更新能力状态
					patchDevice, err = json.Marshal(map[string]interface{}{
						"status": map[string]interface{}{
							"abilities": device.Status.Abilities,
						},
					})
					_, err = clientManager.PatchDevice(device.Name, device.Namespace, string(patchDevice))
					if err != nil {
						errChan <- err
						logs.Errorf("[DEVICE EXPORTER-ABILITY MONITOR] Patch Device[%s] failed, err:%s", device.Name, err.Error())
						return
					}
					logs.Infof("[DEVICE EXPORTER-ABILITY MONITOR] Patch Device[%s] Success!", device.Name)
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

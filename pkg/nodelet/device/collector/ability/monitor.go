package manager

import (
	"context"
	"fmt"
	apis "hit.edu/framework/pkg/apis/cores"
	metav1 "hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/client-go/clients/typed/core"
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

func MonitorAllDevicesState(deviceClient core.DeviceInterface) error {
	// 获取全部的device
	deviceList, err := deviceClient.List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		logs.Errorf("[DEVICE EXPORTER] list device err!")
		return err
	}
	logs.Infof("[DEVICE EXPORTER] ETCD has %d Devices", len(deviceList.Items))

	// 并发同步控制
	var wg sync.WaitGroup
	// 错误处理通道
	errChan := make(chan error, len(deviceList.Items))

	// 遍历所有的Device
	for _, d := range deviceList.Items {
		device := d
		// Ability类型的device
		if device.Spec.AccessMethod.Type == apis.AccessByAbility {
			wg.Add(1)
			// 每个Device单独开一个协程
			go func() {
				devicePhase := device.Status.Phase
				switch devicePhase {
				case apis.DeviceDisconnected: // 离线状态
					logs.Infof("[DEVICE EXPORTER] Device[%s] is Disconnected...", device.Name)
					break
				case apis.DeviceRunning: // 运行状态
					logs.Infof("[DEVICE EXPORTER] Device[%s] is Running...", device.Name)
					break
				case apis.DeviceReadyStartUp: // 准备启动状态
					logs.Infof("[DEVICE EXPORTER] Device[%s] is Ready To StartUp...", device.Name)
					logs.Infof("[DEVICE EXPORTER] Try to StartUp Device[%s]...", device.Name)
					// 获取URL和Name
					url := device.Spec.AccessMethod.URL
					class := device.Spec.Abilities[0]
					abilityName := device.Status.Abilities[class].Name
					// 创建AbilityManager
					am := NewAbilityManager(url, abilityName)
					err = am.BindUUID()
					if err != nil {
						logs.Errorf("[DEVICE EXPORTER] Bind uuid fail")
						errChan <- err
						return
					}
					// 首先先判断一下设备是否在线，不在线才能进行拉起设备的操作
					var flag bool
					flag, err = am.IsOnline()
					if err != nil {
						logs.Errorf("[Device EXPORTER] Device[%s] Judge Online fail", device.Name)
						return
					}
					if flag { // 如果设备在线说明错误
						logs.Errorf("[DEVICE EXPORTER] Device[%s] is Online", device.Name)
						errChan <- fmt.Errorf("[DEVICE EXPORTER] Device[%s] is Online, It should be Offline", device.Name)
						return
					} else { // 如果设备不在线是正常的
						logs.Infof("[DEVICE EXPORTER] Device[%s] is Offline, Normal", device.Name)
						logs.Infof("[DEVICE EXPORTER] Device[%s] Try to StartUp...", device.Name)
						// 尝试拉起设备
						var hb HeartBeat
						hb, err = am.StartupAbility()
						if err != nil {
							logs.Errorf("[DEVICE EXPORTER] Device[%s] Startup fail", device.Name)
							errChan <- err
							return
						}
						logs.Infof("[DEVICE EXPORTER] Device[%s] Try to StartUp Successfully", device.Name)
						// 拉起成功 更新设备字段
						device.Status.Phase = apis.DeviceIdle
						for index, service := range device.Status.Abilities[class].Services {
							port := new(string)
							*port = strconv.Itoa(hb.AbilityPort)
							service.Port = port
							device.Status.Abilities[class].Services[index] = service
						}
						_, err = deviceClient.Get(context.TODO(), device.Name, metav1.GetOptions{})
						if err != nil {
							logs.Errorf("[DEVICE EXPORTER] Get device[%s] fail", device.Name)
							errChan <- err
							return
						}
						_, err = deviceClient.Update(context.TODO(), &device, metav1.UpdateOptions{})
						if err != nil {
							logs.Errorf("[DEVICE EXPORTER] update device fail...")
							errChan <- err
							return
						}
					}

				case apis.DeviceReadyClose: // 准备关闭状态
					logs.Infof("[DEVICE EXPORTER] Device[%s] is Ready To Terminate...", device.Name)
					logs.Infof("[DEVICE EXPORTER] Try to Terminate Device[%s]...", device.Name)
					// 获取URL和Name
					url := device.Spec.AccessMethod.URL
					class := device.Spec.Abilities[0]
					abilityName := device.Status.Abilities[class].Name
					// 创建AbilityManager
					am := NewAbilityManager(url, abilityName)
					err = am.BindUUID()
					if err != nil {
						logs.Errorf("[DEVICE EXPORTER] Bind uuid fail")
						errChan <- err
						return
					}
					// 首先先判断一下设备是否在线，在线才能进行关闭设备的操作
					var flag bool
					flag, err = am.IsOnline()
					if err != nil {
						logs.Errorf("[Device EXPORTER] Device[%s] Judge Online fail", device.Name)
						return
					}
					if !flag { // 如果设备不在线说明不正常
						logs.Errorf("[DEVICE EXPORTER] Device[%s] is Offline", device.Name)
						errChan <- fmt.Errorf("[DEVICE EXPORTER] Device[%s] is Offline, It should be Online", device.Name)
						return
					} else { // 设备在线说明正常
						logs.Infof("[DEVICE EXPORTER] Device[%s] is Online, Normal", device.Name)
						logs.Infof("[DEVICE EXPORTER] Device[%s] Try to Terminate...", device.Name)
						err = am.TerminateAbility()
						if err != nil {
							logs.Infof("[DEVICE EXPORTER] Device[%s] Terminate fail", device.Name)
							errChan <- err
						}
						logs.Infof("[DEVICE EXPORTER] Device[%s] Terminate successfully...", device.Name)
						// 拉起成功 更新设备字段
						device.Status.Phase = apis.DeviceDisconnected
						for index, service := range device.Status.Abilities[class].Services {
							service.Port = nil
							device.Status.Abilities[class].Services[index] = service
						}
						_, err = deviceClient.Get(context.TODO(), device.Name, metav1.GetOptions{})
						if err != nil {
							logs.Errorf("[DEVICE EXPORTER] Get device[%s] fail", device.Name)
							errChan <- err
							return
						}
						_, err = deviceClient.Update(context.TODO(), &device, metav1.UpdateOptions{})
						if err != nil {
							logs.Errorf("[DEVICE EXPORTER] update device fail...")
							errChan <- err
							return
						}
					}
				case apis.DeviceIdle: // 空闲状态

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

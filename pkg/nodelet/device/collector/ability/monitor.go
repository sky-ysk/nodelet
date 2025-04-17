package manager

import (
	"context"
	"errors"
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

func MonitorAllDevicesState(deviceClient core.DeviceInterface, managers *Managers) error {
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
				logs.Infof("[DEVICE EXPORTER] Monitor Device[%s]'s Status", device.Name)
				defer wg.Done()
				// 获取uuid url name
				uuid := device.Status.Abilities[0].InstanceID
				url := device.Spec.AccessMethod.URL
				abilityName := device.Status.Abilities[0].Name
				// 首先判断能力框架是否开启
				_, err = GetAbilityInstances(url)
				if err != nil {
					logs.Errorf("[DEVICE EXPORTER] AbilityFramework is closed")
					return
				}
				// 如果开启了
				if uuid == "" { // 如果uuid为空
					// 为这个device绑定一个manager
					am := NewAbilityManager(url, abilityName)
					err = am.BindUUID(managers.IDList)
					if err != nil {
						logs.Errorf("[DEVICE EXPORTER] Bind uuid fail")
						errChan <- err
						return
					}
					managers.Mutex.Lock()
					// 注册到manager中
					managers.IDList[am.UUid] = true
					managers.AbilityManagers[am.UUid] = am
					managers.Mutex.Unlock()
					// 进行能力的声明周期操作
					var heartBeat HeartBeat
					heartBeat, err = am.StartupAbility()
					if err != nil {
						logs.Errorf("[DEVICE EXPORTER] StartupAbility fail")
						errChan <- err
						return
					}
					// 把内容填充到Device中
					for index, a := range device.Status.Abilities {
						for i, s := range a.Services {
							port := strconv.Itoa(heartBeat.AbilityPort)
							s.Port = port
							a.Services[i] = s
						}
						device.Status.Abilities[index] = a
						device.Status.Abilities[index].InstanceID = am.UUid
					}
					device.Status.Phase = apis.DeviceIdle
					_, err = deviceClient.Update(context.TODO(), &device, metav1.UpdateOptions{})
					if err != nil {
						errChan <- err
						logs.Errorf("[DEVICE EXPORTER] update device fail")
					}

				} else { // 如果uuid不为空
					am := managers.AbilityManagers[uuid]
					if am == nil { // 找不到am
						// 尝试重新进行am的绑定
						am = NewAbilityManager(url, abilityName)
						am.UUid = uuid
						managers.Mutex.Lock()
						managers.IDList[am.UUid] = true
						managers.AbilityManagers[am.UUid] = am
						managers.Mutex.Unlock()

						_, err = am.GetHeartBeat()
						if err != nil {
							managers.Mutex.Lock()
							delete(managers.AbilityManagers, am.UUid)
							delete(managers.IDList, am.UUid)
							managers.Mutex.Unlock()
							// 把内容填充到Device中
							for index, a := range device.Status.Abilities {
								for i, s := range a.Services {
									s.Port = ""
									a.Services[i] = s
								}
								device.Status.Abilities[index] = a
								device.Status.Abilities[index].InstanceID = ""
							}
							device.Status.Phase = apis.DeviceDisconnected
							_, err = deviceClient.Update(context.TODO(), &device, metav1.UpdateOptions{})
							if err != nil {
								errChan <- err
								logs.Errorf("[DEVICE EXPORTER] update device fail")
							}
						}
						// 注册到manager中

					} else { // 可以找到am
						var heartbeat HeartBeat
						heartbeat, err = am.GetHeartBeat()
						if err != nil {
							errChan <- err
							logs.Errorf("[DEVICE EXPORTER] get heartbeat fail")
							customErr := errors.New("empty heartbeats")
							if err.Error() == customErr.Error() { // 能力框架重启了
								logs.Errorf("[DEVICE EXPORTER] get empty heartbeat")
								managers.Mutex.Lock()
								delete(managers.AbilityManagers, am.UUid)
								delete(managers.IDList, am.UUid)
								managers.Mutex.Unlock()
								// 把内容填充到Device中
								for index, a := range device.Status.Abilities {
									for i, s := range a.Services {
										s.Port = ""
										a.Services[i] = s
									}
									device.Status.Abilities[index] = a
									device.Status.Abilities[index].InstanceID = ""
								}
								device.Status.Phase = apis.DeviceDisconnected
								_, err = deviceClient.Update(context.TODO(), &device, metav1.UpdateOptions{})
								if err != nil {
									errChan <- err
									logs.Errorf("[DEVICE EXPORTER] update device fail")
								}
							}
							return
						}
						logs.Infof("[DEVICE EXPORTER] get heartbeat success")
						logs.Infof("[DEVICE EXPORTER] heartbeat state is %s", heartbeat.State)
						state := heartbeat.State
						switch state {
						case Running: // running是正常装填
							break
						case Terminated, Inactive: // 不正常状态
							logs.Infof("[DEVICE EXPORTER] state is not running")
							// 更新为离线状态
							device.Status.Phase = apis.DeviceDisconnected
							for index, a := range device.Status.Abilities {
								for i, s := range a.Services {
									s.Port = ""
									a.Services[i] = s
								}
								// uuid和端口设置为空
								device.Status.Abilities[index] = a
								device.Status.Abilities[index].InstanceID = ""
							}
							// 更新到etcd中
							_, err = deviceClient.Update(context.TODO(), &device, metav1.UpdateOptions{})
							if err != nil {
								errChan <- err
								logs.Errorf("[DEVICE EXPORTER] update device fail")
							}
						}
					}

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

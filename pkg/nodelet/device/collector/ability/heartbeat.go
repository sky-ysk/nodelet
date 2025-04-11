package manager

import (
	"context"
	"fmt"
	"hit.edu/framework/pkg/apimachinery/types"
	apis "hit.edu/framework/pkg/apis/cores"
	metav1 "hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/client-go/clients/typed/core"
	"hit.edu/framework/pkg/component-base/logs"
	"strconv"
)

type Managers struct {
	AbilityManagers map[string]*ManagerOfAbility
	IDList          map[string]bool
}

func GetAllDevicesState(deviceClient core.DeviceInterface, managers *Managers) error {
	deviceList, err := deviceClient.List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		logs.Errorf("[DEVICE EXPORTER] list device err!")
		return err
	}

	// 遍历所有的Device
	for _, device := range deviceList.Items {
		// Ability类型的device
		if device.Spec.AccessMethod.Type == apis.AccessByAbility {
			// 每个Device单独开一个协程
			go func() {
				// 获取uuid url name
				uuid := device.Status.Abilities[0].InstanceID
				url := device.Spec.AccessMethod.URL
				abilityName := device.Status.Abilities[0].Name

				if uuid == "" { // 如果uuid为空
					// 为这个device绑定一个manager
					am := NewAbilityManager(url, abilityName)
					err = am.BindUUID(managers.IDList)
					if err != nil {
						logs.Errorf("[DEVICE EXPORTER] Bind uuid fail")
						return
					}
					// 注册到manager中
					managers.IDList[am.UUid] = true
					managers.AbilityManagers[am.UUid] = am
					// 进行能力的声明周期操作
					heartBeat, err := am.StartupAbility()
					if err != nil {
						logs.Errorf("[DEVICE EXPORTER] StartupAbility fail")
						return
					}
					// 把内容填充到Device中
					for index, a := range device.Status.Abilities {
						for index, s := range a.Services {
							port := strconv.Itoa(heartBeat.AbilityPort)
							s.Port = port
							a.Services[index] = s
						}
						device.Status.Abilities[index] = a
						device.Status.Abilities[index].InstanceID = am.UUid
					}
					deviceClient.Update()
				} else { // 如果uuid不为空

				}
			}()
		}

	}
}

package main

import (
	apis "hit.edu/framework/pkg/apis/cores"
	metav1 "hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/client-go/util/manager"
	"hit.edu/framework/pkg/component-base/logs"
	"hit.edu/framework/pkg/scheduler/utils"
)

func main() {
	cs, err := utils.CreateClientSetWithTimeOut(2000)
	if err != nil {
		panic(err)
	}
	m := manager.NewManager(cs)
	// 创建Device
	device1 := &apis.Device{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "device1",
			Namespace: "test",
			Labels: map[string]string{
				"environment": "dev",
			},
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Device",
			APIVersion: "resources/v1",
		},
		Spec: apis.DeviceSpec{
			Name: "device1",
			AccessMethod: &apis.AccessMethod{
				Type: apis.AccessByAbility,
				URL:  "http://192.168.8.197:8080",
			},
			Abilities: []string{
				"Test",
			},
		},
		Status: apis.DeviceStatus{
			Lock: apis.Lock{
				IsLocked: false,
			},
			Abilities: map[string]apis.Ability{
				"Test": {
					Name: "TEST",
					Services: map[string]apis.AbilityService{
						"test": apis.AbilityService{
							Ip:        new(string),
							Interface: new(string),
							Port:      new(string),
						},
					},
					Status: apis.AbilityRunning,
				},
			},
			Phase: apis.DeviceIdle,
		},
	}

	device2 := &apis.Device{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "device2",
			Namespace: "test",
			Labels: map[string]string{
				"environment": "dev",
			},
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Device",
			APIVersion: "resources/v1",
		},
		Spec: apis.DeviceSpec{
			Name: "device2",
			AccessMethod: &apis.AccessMethod{
				Type: apis.AccessByAbility,
				URL:  "http://192.168.8.197:8080",
			},
			Abilities: []string{
				"Test",
			},
		},
		Status: apis.DeviceStatus{
			Lock: apis.Lock{
				IsLocked: false,
			},
			Abilities: map[string]apis.Ability{
				"Test": {
					Name: "TEST",
					Services: map[string]apis.AbilityService{
						"test": apis.AbilityService{
							Ip:        new(string),
							Interface: new(string),
							Port:      new(string),
						},
					},
					Status: apis.AbilityRunning,
				},
			},
			Phase: apis.DeviceIdle,
		},
	}
	_, err = m.CreateDevice(device1, "test")
	if err != nil {
		logs.Errorf("[TEST] Create Device[%s] err:%s", device1.Name, err.Error())
	}
	_, err = m.CreateDevice(device2, "test")
	if err != nil {
		logs.Errorf("[TEST] Create Device[%s] err:%s", device2.Name, err.Error())
	}
}

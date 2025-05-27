package main

import (
	apis "hit.edu/framework/pkg/apis/cores"
	metav1 "hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/client-go/util/manager"
	"hit.edu/framework/pkg/component-base/logs"
	"hit.edu/framework/pkg/scheduler/utils"
)

func main() {

	//长春现场机器人地址
	//初检设备
	device3IP := "192.168.1.237"
	//复检设备
	device4IP := "192.168.1.234"

	////云服务器测试环境地址
	//device1IP := "172.130.0.61"
	//device2IP := "172.130.0.59"
	cs, err := utils.CreateClientSetWithTimeOut(2000)
	if err != nil {
		panic(err)
	}
	m := manager.NewManager(cs)
	device3 := &apis.Device{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "device3",
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
			Name: "device3",
			AccessMethod: &apis.AccessMethod{
				Type: apis.AccessByAbility,
				URL:  "http://" + device3IP + ":8080",
			},
			Abilities: []string{
				"Detect", "Grab",
			},
		},
		Status: apis.DeviceStatus{
			Abilities: map[string]apis.Ability{
				"Detect": {
					Name: "DetectPosition.Galaxea.Guochuang",
					Services: map[string]apis.AbilityService{
						"DetectPosition": apis.AbilityService{
							Ip:        new(string),
							Interface: new(string),
							Port:      new(string),
						},
					},
					Status: apis.AbilityReadyStartUp,
				},
				"Grab": {
					Name: "GrabBall.Galaxea.Guochuang",
					Services: map[string]apis.AbilityService{
						"GrabBall": apis.AbilityService{
							Ip:        new(string),
							Interface: new(string),
							Port:      new(string),
						},
					},
					Status: apis.AbilityReadyStartUp,
				},
			},
			Lock: apis.Lock{
				IsLocked: true,
				Ref:      2,
			},
			Phase: apis.DeviceIdle,
		},
	}
	*deviceGalaxea.Status.Abilities["Detect"].Services["DetectPosition"].Interface = "/api/task/detect"
	*deviceGalaxea.Status.Abilities["Detect"].Services["DetectPosition"].Ip = "192.168.8.197"
	*deviceGalaxea.Status.Abilities["Grab"].Services["GrabBall"].Interface = "/api/task/grab_ball"
	*deviceGalaxea.Status.Abilities["Grab"].Services["GrabBall"].Ip = "192.168.8.197"

	_, err = m.CreateDevice(deviceGalaxea, "test")
	if err != nil {
		logs.Errorf("[TEST] Create Device[%s] err:%s", deviceGalaxea.Name, err.Error())
	}

	deviceLeju := &apis.Device{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "deviceLeju",
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
			Name: "deviceLeju",
			AccessMethod: &apis.AccessMethod{
				Type: apis.AccessByAbility,
				URL:  "http://192.168.8.165:8080",
			},
			Abilities: []string{
				"Detect", "Grab",
			},
		},
		Status: apis.DeviceStatus{
			Abilities: map[string]apis.Ability{
				"Detect": {
					Name: "DetectPosition.Leju.Guochuang",
					Services: map[string]apis.AbilityService{
						"DetectPosition": apis.AbilityService{
							Ip:        new(string),
							Interface: new(string),
							Port:      new(string),
						},
						"Download": apis.AbilityService{
							Ip:        new(string),
							Interface: new(string),
							Port:      new(string),
						},
					},
					Status: apis.AbilityReadyStartUp,
				},
				"Grab": {
					Name: "GrabBall.Leju.Guochuang",
					Services: map[string]apis.AbilityService{
						"GrabBall": apis.AbilityService{
							Ip:        new(string),
							Interface: new(string),
							Port:      new(string),
						},
					},
					Status: apis.AbilityReadyStartUp,
				},
			},
			Lock: apis.Lock{
				IsLocked: true,
				Ref:      2,
			},
			Phase: apis.DeviceIdle,
		},
	}
	*deviceLeju.Status.Abilities["Detect"].Services["DetectPosition"].Interface = "/api/task/detect"
	*deviceLeju.Status.Abilities["Detect"].Services["DetectPosition"].Ip = "192.168.8.165"
	*deviceLeju.Status.Abilities["Detect"].Services["Download"].Interface = "/api/task/down_new_model"
	*deviceLeju.Status.Abilities["Detect"].Services["Download"].Ip = "192.168.8.165"
	*deviceLeju.Status.Abilities["Grab"].Services["GrabBall"].Interface = "/api/task/grab_ball"
	*deviceLeju.Status.Abilities["Grab"].Services["GrabBall"].Ip = "192.168.8.165"
}

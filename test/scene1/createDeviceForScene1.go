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
			Phase: apis.DeviceIdle,
		},
	}
	*device3.Status.Abilities["Detect"].Services["DetectPosition"].Interface = "/api/task/detect"
	*device3.Status.Abilities["Detect"].Services["DetectPosition"].Ip = device3IP
	*device3.Status.Abilities["Grab"].Services["GrabBall"].Interface = "/api/task/grab_ball"
	*device3.Status.Abilities["Grab"].Services["GrabBall"].Ip = device3IP

	_, err = m.CreateDevice(device3, "test")
	if err != nil {
		logs.Errorf("[TEST] Create Device[%s] err:%s", device3.Name, err.Error())
	}

	device4 := &apis.Device{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "device4",
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
			Name: "device4",
			AccessMethod: &apis.AccessMethod{
				Type: apis.AccessByAbility,
				URL:  "http://" + device4IP + ":8080",
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
			Phase: apis.DeviceIdle,
		},
	}
	*device4.Status.Abilities["Detect"].Services["DetectPosition"].Interface = "/api/task/detect"
	*device4.Status.Abilities["Detect"].Services["DetectPosition"].Ip = device4IP
	*device4.Status.Abilities["Detect"].Services["Download"].Interface = "/api/task/down_new_model"
	*device4.Status.Abilities["Detect"].Services["Download"].Ip = device4IP
	*device4.Status.Abilities["Grab"].Services["GrabBall"].Interface = "/api/task/grab_ball"
	*device4.Status.Abilities["Grab"].Services["GrabBall"].Ip = device4IP

	_, err = m.CreateDevice(device4, "test")
	if err != nil {
		logs.Errorf("[TEST] Create Device[%s] err:%s", device4.Name, err.Error())
	}

}

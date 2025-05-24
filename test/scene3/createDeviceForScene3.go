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
	device1IP := "192.168.1.237"
	//复检设备
	device2IP := "192.168.1.234"

	////云服务器测试环境地址
	//device1IP := "172.130.0.61"
	//device2IP := "172.130.0.59"
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
				URL:  "http://" + device1IP + ":8080",
			},
			Abilities: []string{
				"Grab", "Turn", "Detect", "Put",
			},
		},
		Status: apis.DeviceStatus{
			Lock: apis.Lock{
				IsLocked: false,
			},
			Abilities: map[string]apis.Ability{
				"Grab": {
					Name: "GrabObject.Leju.801",
					Services: map[string]apis.AbilityService{
						"GrabWorkpiece": apis.AbilityService{
							Ip:        new(string),
							Interface: new(string),
							Port:      new(string),
						},
						"GrabInit": apis.AbilityService{
							Ip:        new(string),
							Interface: new(string),
							Port:      new(string),
						},
						"ReturnToLevel": apis.AbilityService{
							Ip:        new(string),
							Interface: new(string),
							Port:      new(string),
						},
					},

					Status: apis.AbilityReadyStartUp,
				},
				"Turn": {
					Name: "Turn.Leju.801",
					Services: map[string]apis.AbilityService{
						"Turn": apis.AbilityService{
							Ip:        new(string),
							Interface: new(string),
							Port:      new(string),
						},
					},
					Status: apis.AbilityReadyStartUp,
				},
				"Detect": {
					Name: "DetectObject.Leju.801",
					Services: map[string]apis.AbilityService{
						"DetectWorkpiece": apis.AbilityService{
							Ip:        new(string),
							Interface: new(string),
							Port:      new(string),
						},
					},
					Status: apis.AbilityReadyStartUp,
				},
				"Put": {
					Name: "PutObject.Leju.801",
					Services: map[string]apis.AbilityService{
						"PutWorkpiece": apis.AbilityService{
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
	// 放置工件
	*device1.Status.Abilities["Put"].Services["PutWorkpiece"].Ip = device1IP
	*device1.Status.Abilities["Put"].Services["PutWorkpiece"].Interface = "/api/task/put"
	// 抓取工件
	*device1.Status.Abilities["Grab"].Services["GrabWorkpiece"].Ip = device1IP
	*device1.Status.Abilities["Grab"].Services["GrabWorkpiece"].Interface = "/api/task/grab"
	// 转身
	*device1.Status.Abilities["Turn"].Services["Turn"].Ip = device1IP
	*device1.Status.Abilities["Turn"].Services["Turn"].Interface = "/api/task/turn"
	// 初始化
	*device1.Status.Abilities["Grab"].Services["GrabInit"].Ip = device1IP
	*device1.Status.Abilities["Grab"].Services["GrabInit"].Interface = "/api/task/init"
	// 检测
	*device1.Status.Abilities["Detect"].Services["DetectWorkpiece"].Ip = device1IP
	*device1.Status.Abilities["Detect"].Services["DetectWorkpiece"].Interface = "/api/task/detect"

	// 复位
	*device1.Status.Abilities["Grab"].Services["ReturnToLevel"].Ip = device1IP
	*device1.Status.Abilities["Grab"].Services["ReturnToLevel"].Interface = "/api/task/return-to-level"

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
				URL:  "http://" + device2IP + ":8080",
			},
			Abilities: []string{
				"Grab", "Turn", "Detect", "Put",
			},
		},
		Status: apis.DeviceStatus{
			Lock: apis.Lock{
				IsLocked: false,
			},
			Abilities: map[string]apis.Ability{
				"Grab": {
					Name: "GrabObject.Leju.801",
					Services: map[string]apis.AbilityService{
						"GrabWorkpiece": apis.AbilityService{
							Ip:        new(string),
							Interface: new(string),
							Port:      new(string),
						},
						"GrabInit": apis.AbilityService{
							Ip:        new(string),
							Interface: new(string),
							Port:      new(string),
						},
						"ReturnToLevel": apis.AbilityService{
							Ip:        new(string),
							Interface: new(string),
							Port:      new(string),
						},
					},
					Status: apis.AbilityReadyStartUp,
				},
				"Turn": {
					Name: "Turn.Leju.801",
					Services: map[string]apis.AbilityService{
						"Turn": apis.AbilityService{
							Ip:        new(string),
							Interface: new(string),
							Port:      new(string),
						},
					},
					Status: apis.AbilityReadyStartUp,
				},
				"Detect": {
					Name: "DetectObject.Leju.801",
					Services: map[string]apis.AbilityService{
						"DetectWorkpiece": apis.AbilityService{
							Ip:        new(string),
							Interface: new(string),
							Port:      new(string),
						},
					},
					Status: apis.AbilityReadyStartUp,
				},
				"Put": {
					Name: "PutObject.Leju.801",
					Services: map[string]apis.AbilityService{
						"PutWorkpiece": apis.AbilityService{
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

	// 放置工件
	*device2.Status.Abilities["Put"].Services["PutWorkpiece"].Ip = device2IP
	*device2.Status.Abilities["Put"].Services["PutWorkpiece"].Interface = "/api/task/put"
	// 抓取工件
	*device2.Status.Abilities["Grab"].Services["GrabWorkpiece"].Ip = device2IP
	*device2.Status.Abilities["Grab"].Services["GrabWorkpiece"].Interface = "/api/task/grab"
	// 转身
	*device2.Status.Abilities["Turn"].Services["Turn"].Ip = device2IP
	*device2.Status.Abilities["Turn"].Services["Turn"].Interface = "/api/task/turn"
	// 初始化
	*device2.Status.Abilities["Grab"].Services["GrabInit"].Ip = device2IP
	*device2.Status.Abilities["Grab"].Services["GrabInit"].Interface = "/api/task/init"
	// 检测
	*device2.Status.Abilities["Detect"].Services["DetectWorkpiece"].Ip = device2IP
	*device2.Status.Abilities["Detect"].Services["DetectWorkpiece"].Interface = "/api/task/detect"

	// 复位
	*device2.Status.Abilities["Grab"].Services["ReturnToLevel"].Ip = device2IP
	*device2.Status.Abilities["Grab"].Services["ReturnToLevel"].Interface = "/api/task/return-to-level"
	_, err = m.CreateDevice(device1, "test")
	if err != nil {
		logs.Errorf("[TEST] Create Device[%s] err:%s", device1.Name, err.Error())
	}
	_, err = m.CreateDevice(device2, "test")
	if err != nil {
		logs.Errorf("[TEST] Create Device[%s] err:%s", device2.Name, err.Error())
	}
}

package lib

//
//import (
//	apis "hit.edu/framework/pkg/apis/cores"
//	"hit.edu/framework/pkg/component-base/logs"
//	"hit.edu/framework/pkg/nodelet/task/runtime/device/ability/manager"
//	"strconv"
//	"testing"
//)
//
//func TestSendServiceRequest(t *testing.T) {
//	logs.Init("test")
//	ability := ""
//	// 选择不同的测试device
//	// 在创建device测试函数中填写参数
//	device := CreateArmAngleDevice()
//	//device := CreateLeftArmUpDevice()
//	//device := CreateLeftArmDownDevice()
//	//device := CreatePredictDevice()
//	//device := CreatePredictByUrlDevice()
//
//	// 能力框架的操作
//	managerUrl := "" // 能力框架url
//	abilityName := ""
//
//	am := manager.NewAbilityManager(managerUrl, abilityName)
//	hb, err := am.StartupAbility()
//
//	if err != nil {
//		logs.Errorf(err.Error())
//	}
//	for _, a := range device.Status.Abilities {
//		for _, service := range a.Services {
//			service.Port = strconv.Itoa(hb.AbilityPort)
//		}
//	}
//	output, err := SendServiceRequest(ability, device)
//	if err != nil {
//		logs.Errorf("send service request fail..")
//	} else {
//		logs.Infof("output is %v", output)
//	}
//	//// 终止能力[备用]
//	//err := am.TerminateAbility()
//}
//
//// 创建ArmAngle的测试device
//func CreateArmAngleDevice() *apis.Device {
//
//	url := "" // 填写框架url
//	device := &apis.Device{
//		Spec: apis.DeviceSpec{
//			AccessMethod: apis.AccessMethod{
//				URL: url,
//			},
//			ExpectedProperties: make(map[string]apis.Property),
//		},
//		Status: apis.DeviceStatus{
//			Abilities: make([]apis.AbilityStatus, 0),
//		},
//	}
//	device.Spec.ExpectedProperties["left"] = apis.Property{
//		Value: "-1.221,0.0872,0,0,0,0,0",
//	}
//	device.Spec.ExpectedProperties["right"] = apis.Property{
//		Value: "0,0,0,0,0,0,0",
//	}
//	device.Status.Abilities = append(device.Status.Abilities, apis.AbilityStatus{
//		Name: "Arm",
//		Services: []apis.AbilityServiceStatus{
//			{
//				Name:      "ArmAngle",
//				Ip:        "192.168.8.165",
//				Interface: "/api/control/arm_angle",
//			},
//			{
//				Name:      "LeftArmUp",
//				Ip:        "192.168.8.165",
//				Interface: "/api/control/left_arm_up",
//			},
//			{
//				Name:      "LeftArmDown",
//				Ip:        "192.168.8.165",
//				Interface: "/api/control/left_arm_down",
//			},
//		},
//	})
//	return device
//}
//
//// 创建LeftArmUp的测试device
//func CreateLeftArmUpDevice() *apis.Device {
//	url := "" // 填写框架url
//	device := &apis.Device{
//		Spec: apis.DeviceSpec{
//			AccessMethod: apis.AccessMethod{
//				URL: url,
//			},
//			ExpectedProperties: make(map[string]apis.Property),
//		},
//		Status: apis.DeviceStatus{
//			Abilities: make([]apis.AbilityStatus, 0),
//		},
//	}
//	device.Spec.ExpectedProperties["left"] = apis.Property{
//		Value: "-1.221,0.0872,0,0,0,0,0",
//	}
//	device.Spec.ExpectedProperties["right"] = apis.Property{
//		Value: "0,0,0,0,0,0,0",
//	}
//	device.Status.Abilities = append(device.Status.Abilities, apis.AbilityStatus{
//		Name: "Arm",
//		Services: []apis.AbilityServiceStatus{
//			{
//				Name:      "ArmAngle",
//				Ip:        "192.168.8.165",
//				Interface: "/api/control/arm_angle",
//			},
//			{
//				Name:      "LeftArmUp",
//				Ip:        "192.168.8.165",
//				Interface: "/api/control/left_arm_up",
//			},
//			{
//				Name:      "LeftArmDown",
//				Ip:        "192.168.8.165",
//				Interface: "/api/control/left_arm_down",
//			},
//		},
//	})
//	return device
//}
//
//// // 创建LeftArmDown能力的测试device
//func CreateLeftArmDownDevice() *apis.Device {
//	url := "" // 填写框架url
//	device := &apis.Device{
//		Spec: apis.DeviceSpec{
//			AccessMethod: apis.AccessMethod{
//				URL: url,
//			},
//			ExpectedProperties: make(map[string]apis.Property),
//		},
//		Status: apis.DeviceStatus{
//			Abilities: make([]apis.AbilityStatus, 0),
//		},
//	}
//	device.Spec.ExpectedProperties["left"] = apis.Property{
//		Value: "-1.221,0.0872,0,0,0,0,0",
//	}
//	device.Spec.ExpectedProperties["right"] = apis.Property{
//		Value: "0,0,0,0,0,0,0",
//	}
//	device.Status.Abilities = append(device.Status.Abilities, apis.AbilityStatus{
//		Name: "Arm",
//		Services: []apis.AbilityServiceStatus{
//			{
//				Name:      "ArmAngle",
//				Ip:        "192.168.8.165",
//				Interface: "/api/control/arm_angle",
//			},
//			{
//				Name:      "LeftArmUp",
//				Ip:        "192.168.8.165",
//				Interface: "/api/control/left_arm_up",
//			},
//			{
//				Name:      "LeftArmDown",
//				Ip:        "192.168.8.165",
//				Interface: "/api/control/left_arm_down",
//			},
//		},
//	})
//	return device
//}
//
//// 创建predict能力的测试device
//func CreatePredictDevice() *apis.Device {
//	url := ""
//
//	path := ""
//	device := &apis.Device{
//		Spec: apis.DeviceSpec{
//			AccessMethod: apis.AccessMethod{
//				URL: url,
//			},
//			ExpectedProperties: make(map[string]apis.Property),
//		},
//		Status: apis.DeviceStatus{
//			Abilities: make([]apis.AbilityStatus, 0),
//		},
//	}
//	device.Spec.ExpectedProperties["path"] = apis.Property{Value: path}
//
//	device.Status.Abilities = append(device.Status.Abilities, apis.AbilityStatus{
//		Name: "Predict",
//		Services: []apis.AbilityServiceStatus{
//			{
//				Name:      "Predict",
//				Ip:        "192.168.8.165",
//				Interface: "/predict",
//			},
//			{
//				Name:      "PredictByUrl",
//				Ip:        "192.168.8.165",
//				Interface: "/predict_by_url",
//			},
//		},
//	})
//	return device
//}
//
//// 创建predictByUrl能力的测试device
//func CreatePredictByUrlDevice() *apis.Device {
//	url := ""
//	cameraUrl := ""
//	compressed := "false"
//	imageType := ""
//	position := ""
//	device := &apis.Device{
//		Spec: apis.DeviceSpec{
//			AccessMethod: apis.AccessMethod{
//				URL: url,
//			},
//			ExpectedProperties: make(map[string]apis.Property),
//		},
//		Status: apis.DeviceStatus{
//			Abilities: make([]apis.AbilityStatus, 0),
//		},
//	}
//	device.Spec.ExpectedProperties["cameraUrl"] = apis.Property{Value: cameraUrl}
//	device.Spec.ExpectedProperties["compressed"] = apis.Property{Value: compressed}
//	device.Spec.ExpectedProperties["imageType"] = apis.Property{Value: imageType}
//	device.Spec.ExpectedProperties["position"] = apis.Property{Value: position}
//	device.Status.Abilities = append(device.Status.Abilities, apis.AbilityStatus{
//		Name: "Predict",
//		Services: []apis.AbilityServiceStatus{
//			{
//				Name:      "Predict",
//				Ip:        "192.168.8.165",
//				Interface: "/predict",
//			},
//			{
//				Name:      "PredictByUrl",
//				Ip:        "192.168.8.165",
//				Interface: "/predict_by_url",
//			},
//		},
//	})
//	return device
//}

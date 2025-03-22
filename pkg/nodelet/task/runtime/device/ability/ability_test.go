package ability

import (
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/component-base/logs"
	"testing"
)

func TestPublishAbilityInst(t *testing.T) {

	// 发布指令 service开头为业务请求 manage开头为能力框架操作
	//inst := "service_{具体业务}"
	inst := "manage_{能力名}"

	logs.Init("test")

	// 选择不同的测试device
	// 在创建device测试函数中填写参数
	device := CreateArmAngleDevice()
	//device := CreateLeftArmUpDevice()
	//device := CreateLeftArmDownDevice()
	//device := CreatePredictDevice()
	//device := CreatePredictByUrlDevice()

	output, err := PublishAbilityInst(inst, device, "")

	if err != nil {
		logs.Errorf("publish ability inst err")
		return
	}

	logs.Infof("output is %v", output)
	for _, a := range device.Status.Abilities {
		for _, service := range a.Services {
			service.Port = output.Value
		}
	}

	// 能力框架的终止方式通过操作字段operation来进行
	output, err = PublishAbilityInst(inst, device, "terminate")
}

// 创建ArmAngle的测试device
func CreateArmAngleDevice() *apis.Device {

	url := "" // 填写框架url
	device := &apis.Device{
		Spec: apis.DeviceSpec{
			AccessMethod: apis.AccessMethod{
				URL: url,
			},
			ExpectedProperties: make(map[string]apis.Property),
		},
		Status: apis.DeviceStatus{
			Abilities: make([]apis.AbilityStatus, 0),
		},
	}
	device.Spec.ExpectedProperties["left"] = apis.Property{
		Value: "-1.221,0.0872,0,0,0,0,0",
	}
	device.Spec.ExpectedProperties["right"] = apis.Property{
		Value: "0,0,0,0,0,0,0",
	}
	device.Status.Abilities = append(device.Status.Abilities, apis.AbilityStatus{
		Name: "Arm",
		Services: []apis.AbilityServiceStatus{
			{
				Name:      "ArmAngle",
				Ip:        "192.168.8.165",
				Interface: "/api/control/arm_angle",
			},
			{
				Name:      "LeftArmUp",
				Ip:        "192.168.8.165",
				Interface: "/api/control/left_arm_up",
			},
			{
				Name:      "LeftArmDown",
				Ip:        "192.168.8.165",
				Interface: "/api/control/left_arm_down",
			},
		},
	})
	return device
}

// 创建LeftArmUp的测试device
func CreateLeftArmUpDevice() *apis.Device {
	url := "" // 填写框架url
	device := &apis.Device{
		Spec: apis.DeviceSpec{
			AccessMethod: apis.AccessMethod{
				URL: url,
			},
			ExpectedProperties: make(map[string]apis.Property),
		},
		Status: apis.DeviceStatus{
			Abilities: make([]apis.AbilityStatus, 0),
		},
	}
	device.Spec.ExpectedProperties["left"] = apis.Property{
		Value: "-1.221,0.0872,0,0,0,0,0",
	}
	device.Spec.ExpectedProperties["right"] = apis.Property{
		Value: "0,0,0,0,0,0,0",
	}
	device.Status.Abilities = append(device.Status.Abilities, apis.AbilityStatus{
		Name: "Arm",
		Services: []apis.AbilityServiceStatus{
			{
				Name:      "ArmAngle",
				Ip:        "192.168.8.165",
				Interface: "/api/control/arm_angle",
			},
			{
				Name:      "LeftArmUp",
				Ip:        "192.168.8.165",
				Interface: "/api/control/left_arm_up",
			},
			{
				Name:      "LeftArmDown",
				Ip:        "192.168.8.165",
				Interface: "/api/control/left_arm_down",
			},
		},
	})
	return device
}

// // 创建LeftArmDown能力的测试device
func CreateLeftArmDownDevice() *apis.Device {
	url := "" // 填写框架url
	device := &apis.Device{
		Spec: apis.DeviceSpec{
			AccessMethod: apis.AccessMethod{
				URL: url,
			},
			ExpectedProperties: make(map[string]apis.Property),
		},
		Status: apis.DeviceStatus{
			Abilities: make([]apis.AbilityStatus, 0),
		},
	}
	device.Spec.ExpectedProperties["left"] = apis.Property{
		Value: "-1.221,0.0872,0,0,0,0,0",
	}
	device.Spec.ExpectedProperties["right"] = apis.Property{
		Value: "0,0,0,0,0,0,0",
	}
	device.Status.Abilities = append(device.Status.Abilities, apis.AbilityStatus{
		Name: "Arm",
		Services: []apis.AbilityServiceStatus{
			{
				Name:      "ArmAngle",
				Ip:        "192.168.8.165",
				Interface: "/api/control/arm_angle",
			},
			{
				Name:      "LeftArmUp",
				Ip:        "192.168.8.165",
				Interface: "/api/control/left_arm_up",
			},
			{
				Name:      "LeftArmDown",
				Ip:        "192.168.8.165",
				Interface: "/api/control/left_arm_down",
			},
		},
	})
	return device
}

// 创建predict能力的测试device
func CreatePredictDevice() *apis.Device {
	url := ""

	path := ""
	device := &apis.Device{
		Spec: apis.DeviceSpec{
			AccessMethod: apis.AccessMethod{
				URL: url,
			},
			ExpectedProperties: make(map[string]apis.Property),
		},
		Status: apis.DeviceStatus{
			Abilities: make([]apis.AbilityStatus, 0),
		},
	}
	device.Spec.ExpectedProperties["path"] = apis.Property{Value: path}

	device.Status.Abilities = append(device.Status.Abilities, apis.AbilityStatus{
		Name: "Predict",
		Services: []apis.AbilityServiceStatus{
			{
				Name:      "Predict",
				Ip:        "192.168.8.165",
				Interface: "/predict",
			},
			{
				Name:      "PredictByUrl",
				Ip:        "192.168.8.165",
				Interface: "/predict_by_url",
			},
		},
	})
	return device
}

// 创建predictByUrl能力的测试device
func CreatePredictByUrlDevice() *apis.Device {
	url := ""
	cameraUrl := ""
	compressed := "false"
	imageType := ""
	position := ""
	device := &apis.Device{
		Spec: apis.DeviceSpec{
			AccessMethod: apis.AccessMethod{
				URL: url,
			},
			ExpectedProperties: make(map[string]apis.Property),
		},
		Status: apis.DeviceStatus{
			Abilities: make([]apis.AbilityStatus, 0),
		},
	}
	device.Spec.ExpectedProperties["cameraUrl"] = apis.Property{Value: cameraUrl}
	device.Spec.ExpectedProperties["compressed"] = apis.Property{Value: compressed}
	device.Spec.ExpectedProperties["imageType"] = apis.Property{Value: imageType}
	device.Spec.ExpectedProperties["position"] = apis.Property{Value: position}
	device.Status.Abilities = append(device.Status.Abilities, apis.AbilityStatus{
		Name: "Predict",
		Services: []apis.AbilityServiceStatus{
			{
				Name:      "Predict",
				Ip:        "192.168.8.165",
				Interface: "/predict",
			},
			{
				Name:      "PredictByUrl",
				Ip:        "192.168.8.165",
				Interface: "/predict_by_url",
			},
		},
	})
	return device
}

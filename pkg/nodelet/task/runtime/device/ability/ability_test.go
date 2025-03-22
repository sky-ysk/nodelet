package ability

import (
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/component-base/logs"
	"testing"
)

func TestPublishAbilityInst(t *testing.T) {

	// 发布指令 service开头为业务请求 manage开头为能力框架操作
	//inst := "service_"
	inst := "manage_"

	url := ""

	logs.Init("test")
	ability := ""
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
	}

	logs.Infof("output is %v", output)

}

// 创建ArmAngle的测试device
func CreateArmAngleDevice() *apis.Device {
	url := ""
	device := &apis.Device{
		Spec: apis.DeviceSpec{
			AccessMethod: apis.AccessMethod{
				URL: url,
			},
			ExpectedProperties: make(map[string]apis.Property),
		},
	}
	device.Spec.ExpectedProperties["left"] = apis.Property{
		Value: "-1.221,0.0872,0,0,0,0,0",
	}
	device.Spec.ExpectedProperties["right"] = apis.Property{
		Value: "0,0,0,0,0,0,0",
	}
	return device
}

// 创建LeftArmUp的测试device
func CreateLeftArmUpDevice() *apis.Device {
	url := ""
	device := &apis.Device{
		Spec: apis.DeviceSpec{
			AccessMethod: apis.AccessMethod{
				URL: url,
			},
			ExpectedProperties: make(map[string]apis.Property),
		},
	}

	return device
}

// // 创建LeftArmDown能力的测试device
func CreateLeftArmDownDevice() *apis.Device {
	url := ""
	device := &apis.Device{
		Spec: apis.DeviceSpec{
			AccessMethod: apis.AccessMethod{
				URL: url,
			},
			ExpectedProperties: make(map[string]apis.Property),
		},
	}

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
	}
	device.Spec.ExpectedProperties["path"] = apis.Property{Value: path}
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
	}
	device.Spec.ExpectedProperties["cameraUrl"] = apis.Property{Value: cameraUrl}
	device.Spec.ExpectedProperties["compressed"] = apis.Property{Value: compressed}
	device.Spec.ExpectedProperties["imageType"] = apis.Property{Value: imageType}
	device.Spec.ExpectedProperties["position"] = apis.Property{Value: position}

	return device
}

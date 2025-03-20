package lib

import (
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/component-base/logs"
	"testing"
)

func TestSendServiceRequest(t *testing.T) {
	logs.Init("test")
	//url := ""
	//ability := "ArmAngle"
	//device := &apis.Device{
	//	Spec: apis.DeviceSpec{
	//		AccessMethod: apis.AccessMethod{
	//			URL: url,
	//		},
	//		ExpectedProperties: make(map[string]apis.Property),
	//	},
	//}
	//device.Spec.ExpectedProperties["left"] = apis.Property{
	//	Value: "-1.221,0.0872,0,0,0,0,0",
	//}
	//device.Spec.ExpectedProperties["right"] = apis.Property{
	//	Value: "0,0,0,0,0,0,0",
	//}

	url := ""
	path := ""
	ability := "Predict"
	device := &apis.Device{
		Spec: apis.DeviceSpec{
			AccessMethod: apis.AccessMethod{
				URL: url,
			},
			ExpectedProperties: make(map[string]apis.Property),
		},
	}
	device.Spec.ExpectedProperties["path"] = apis.Property{Value: path}

	output, err := SendServiceRequest(ability, device)
	if err != nil {
		logs.Errorf("send service request fail..")
	} else {
		logs.Infof("output is %v", output)
	}

}

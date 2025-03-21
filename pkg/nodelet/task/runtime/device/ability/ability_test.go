package ability

import (
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/component-base/logs"
	"testing"
)

func TestPublishAbilityInst(t *testing.T) {

	//inst := "service_"

	inst := "manage_"

	url := ""

	device := &apis.Device{
		Spec: apis.DeviceSpec{
			AccessMethod: apis.AccessMethod{
				URL: url,
			},
			ExpectedProperties: make(map[string]apis.Property),
		},
	}

	device.Spec.ExpectedProperties["left"] = apis.Property{Value: "-1.221,0.0872,0,0,0,0,0"}
	device.Spec.ExpectedProperties["right"] = apis.Property{Value: "0,0,0,0,0,0,0"}

	//path := ""
	//device := &apis.Device{
	//	Spec: apis.DeviceSpec{
	//		AccessMethod:       apis.AccessMethod{
	//			URL: url,
	//		},
	//		ExpectedProperties: make(map[string]apis.Property),
	//	},
	//}
	//
	//device.Spec.ExpectedProperties["path"] = apis.Property{Value: path}

	output, err := PublishAbilityInst(inst, device, "")
	if err != nil {
		logs.Errorf("publish ability inst err")
	}

	logs.Infof("output is %v", output)

}

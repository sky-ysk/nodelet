package utils

import (
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/component-base/logs"
	"testing"
)

func TestConstructParam(t *testing.T) {
	logs.Infof("[Test] testing ConstructParam....\n")
	devices, runtime := NewRuntimeAndDevice()
	err := ConstructParam(devices, runtime)
	if err != nil {
		logs.Errorf("[Test] ConstructParam err: %v\n", err)
	}

	logs.Infof("[Test] ConstructParam successful\n")
	logs.Info("[Test] print device's expected properties\n")
	for name, device := range devices {
		logs.Infof("[Test] device name: %s\n", name)
		for propertyName, property := range device.Spec.ExpectedProperties {
			logs.Infof("[Test] \t property name: %s\n", propertyName)
			logs.Infof("[Test] \t property value: %s\n", property.Value)
		}
	}
}

func NewRuntimeAndDevice() (map[string]apis.Device, *apis.Runtime) {

	devices := make(map[string]apis.Device, 1)
	devices["transferRobot"] = apis.Device{
		Spec: apis.DeviceSpec{
			Name:               "transferRobot",
			ExpectedProperties: map[string]apis.Property{},
		},
		Status: apis.DeviceStatus{},
	}
	runtime := apis.Runtime{
		Inputs: []apis.Input{
			apis.Input{
				Type:      apis.LocalData,
				Name:      "dest",
				Value:     "R201",
				ValueType: "string",
			},
			apis.Input{
				Type:      apis.LocalData,
				Name:      "orientation",
				Value:     "-3.12",
				ValueType: "double",
			},
			apis.Input{
				Type:      apis.LocalData,
				Name:      "dock",
				Value:     "true",
				ValueType: "bool",
			},
		},
	}
	return devices, &runtime
}

package device

import (
	"context"
	apis "hit.edu/framework/pkg/apis/cores"
	metav1 "hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/component-base/logs"
	"testing"
)

func CreateDeviceDemo() *apis.Device {
	deviceTest := &apis.Device{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "deviceTest",
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
			Name:               "deviceTest",
			ExpectedProperties: map[string]apis.Property{},
			AccessMethod: apis.AccessMethod{
				Type: apis.AccessByAbility,
				URL:  "http://127.0.0.1:8123",
			},
		},
		Status: apis.DeviceStatus{
			DeviceID: "deviceTest",
			Abilities: []apis.AbilityStatus{
				{
					Name: "Mock",
					Services: []apis.AbilityServiceStatus{
						{
							Name: "test1",
							Ip:   "http://127.0.0.1:8123",
						},
					},
				},
			},
		},
	}
	return deviceTest
}
func TestDeviceExporter(t *testing.T) {
	clientSet, err := InitClient()
	if err != nil {
		logs.Errorf("[DEVICE EXPORTER] init client failed")
		return
	}
	deviceClient := clientSet.Core().Devices("test")
	device := CreateDeviceDemo()
	_, err = deviceClient.Create(context.TODO(), device, metav1.CreateOptions{})
	if err != nil {
		logs.Errorf("create device fail")
	}
	logs.Init("test")
	cfg := &Config{
		EnabledCollectors: make([]string, 0),
	}
	deviceExporter, err := NewDeviceExporter(cfg)
	if err != nil {
		logs.Errorf("new device exporter fail")
		return
	}
	err = deviceExporter.Run()
	if err != nil {
		logs.Errorf("run device exporter fail")
		return
	}
}

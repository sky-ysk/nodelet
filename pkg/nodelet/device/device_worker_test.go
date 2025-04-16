package device

import (
	"hit.edu/framework/pkg/component-base/logs"
	"testing"
)

func TestDeviceWorker(t *testing.T) {
	clientSet, err := InitClient()
	if err != nil {
		logs.Errorf("[DEVICE EXPORTER] init client failed")
		return
	}
	deviceClient := clientSet.Core().Devices("test")
	logs.Init("test")
	deviceWorker := NewDeviceWorker(deviceClient)
	err = deviceWorker.Run()
	if err != nil {
		logs.Errorf("err:%v", err)
	}
}

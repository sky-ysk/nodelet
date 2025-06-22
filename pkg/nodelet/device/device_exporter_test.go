package device

import (
	"hit.edu/framework/pkg/component-base/logs"
	"testing"
)

func TestDeviceExporter(t *testing.T) {
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

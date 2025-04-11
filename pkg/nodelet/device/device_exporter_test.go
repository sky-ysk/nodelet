package device

import (
	"testing"
)

func TestDeviceExporter(t *testing.T) {
	cfg := &Config{
		EnabledCollectors: make([]string, 0),
	}
	NewDeviceExporter(cfg)
}

package plugins

import (
	"context"
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/component-base/logs"
	"hit.edu/framework/pkg/nodelet/device"
	"hit.edu/framework/pkg/scheduler/framework"
)

type DeviceBinder struct {
	deviceWorker *device.DeviceWorker
}

func (db *DeviceBinder) Name() string {
	return "DeviceBinder"
}

func NewDeviceBinder(ctx context.Context, f framework.Handle) (framework.Plugin, error) {
	return &DeviceBinder{
		deviceWorker: device.GetDeviceWorker(),
	}, nil
}

func (db *DeviceBinder) Bind(ctx context.Context, state *framework.CycleState, group *apis.Group, nodeName string) (status *framework.Status) {
	if len(group.Spec.DeviceRequirements) == 0 {
		logs.Infof("group %s have no device spec, skip device binding plugin", group.Name)
		return framework.NewStatus(framework.Skip, "group %s have no device spec, skip device binding plugin", group.Name)
	}

	//TODO

	return nil
}

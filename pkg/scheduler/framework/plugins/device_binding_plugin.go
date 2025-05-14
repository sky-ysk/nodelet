package plugins

import (
	"context"
	apis "hit.edu/framework/pkg/apis/cores"
	m "hit.edu/framework/pkg/client-go/util/manager"
	"hit.edu/framework/pkg/component-base/logs"
	"hit.edu/framework/pkg/nodelet/device"
	"hit.edu/framework/pkg/scheduler/framework"
	"hit.edu/framework/pkg/scheduler/utils"
)

type DeviceBinder struct {
	deviceWorker *device.DeviceWorker
	manager      *m.Manager
}

func (db *DeviceBinder) Name() string {
	return "DeviceBinder"
}

func NewDeviceBinder(ctx context.Context, f framework.Handle) (framework.Plugin, error) {
	cs, err := utils.CreateClientSetWithTimeOut(1000)
	if err != nil {
		logs.Error(err.Error())
		return nil, err
	}
	return &DeviceBinder{
		deviceWorker: device.GetDeviceWorker(),
		manager:      m.NewManager(cs),
	}, nil
}

func (db *DeviceBinder) Bind(ctx context.Context, state *framework.CycleState, group *apis.Group, nodeName string) (status *framework.Status) {
	if len(group.Spec.Devices) == 0 {
		logs.Infof("group %s have no device spec, skip device binding plugin", group.Name)
		return framework.NewStatus(framework.Skip, "group %s have no device spec, skip device binding plugin", group.Name)
	}

	//TODO

	for _, device := range group.Spec.Devices {

	}

	return nil
}

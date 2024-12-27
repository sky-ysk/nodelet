package device

import (
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/component-base/logs"
)

type DeviceRuntime struct {
}

func NewDeviceRuntime() DeviceRuntime {
	return DeviceRuntime{}
}
func (dr DeviceRuntime) Run(group *apis.Group, action *apis.Action, runtime *apis.Runtime) error {
	logs.Infof("device runtime for task: %s", group.Name)
	return nil
}
func (dr DeviceRuntime) Kill(group *apis.Group, action *apis.Action, runtime *apis.Runtime) error {
	logs.Infof("device runtime kill task: %s", group.Name)
	return nil
}
func (dr DeviceRuntime) CheckTaskStatus(group *apis.Group, action *apis.Action, runtime *apis.Runtime) (string, error) {

	return "", nil
}

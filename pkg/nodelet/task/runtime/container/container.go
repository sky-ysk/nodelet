package container

import (
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/component-base/logs"
)

type ContainerRuntime struct {
}

func NewContainerRuntime() ContainerRuntime {
	return ContainerRuntime{}
}

func (dr ContainerRuntime) Run(group *apis.Group, action *apis.Action, runtime *apis.Runtime, actionIndex, runtimeIndex int) error {
	logs.Infof("docker runtime for task:%s", group.Name)
	return nil
}
func (dr ContainerRuntime) Kill(group *apis.Group, action *apis.Action, runtime *apis.Runtime) error {
	logs.Infof("docker runtime kill task:%s", group.Name)
	return nil
}
func (dr ContainerRuntime) CheckRuntimeStatus(group *apis.Group, action *apis.Action, runtime *apis.Runtime) (string, error) {

	return "", nil
}
func (dr ContainerRuntime) StoreData(group *apis.Group, action *apis.Action, runtime *apis.Runtime, actionIndex int, runtimeIndex int) string {

	return ""
}
func (dr ContainerRuntime) RestoreData(group *apis.Group, action *apis.Action, runtime *apis.Runtime, actionIndex int, runtimeIndex int) error {
	return nil
}
func (dr ContainerRuntime) StartRuntime(group *apis.Group, action *apis.Action, runtime *apis.Runtime, actionIndex int, runtimeIndex int) error {

	return nil
}
func (dr ContainerRuntime) InitRuntime(group *apis.Group, action *apis.Action, runtime *apis.Runtime, actionIndex int, runtimeIndex int) error {
	return nil
}
func (dr ContainerRuntime) StopRuntime(group *apis.Group, action *apis.Action, runtime *apis.Runtime, actionIndex int, runtimeIndex int) error {
	return nil
}

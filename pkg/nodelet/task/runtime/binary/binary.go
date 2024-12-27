package binary

import (
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/component-base/logs"
)

type BinaryRuntime struct {
}

func NewBinaryRuntime() BinaryRuntime {
	return BinaryRuntime{}
}
func (br BinaryRuntime) Run(group *apis.Group, action *apis.Action, runtime *apis.Runtime) error {
	logs.Infof("binary runtime for task:%s", group.Name)
	return nil
}
func (br BinaryRuntime) Kill(group *apis.Group, action *apis.Action, runtime *apis.Runtime) error {
	logs.Infof("binary runtime kill task:%s", group.Name)
	return nil
}
func (br BinaryRuntime) CheckTaskStatus(group *apis.Group, action *apis.Action, runtime *apis.Runtime) (string, error) {

	return "", nil
}

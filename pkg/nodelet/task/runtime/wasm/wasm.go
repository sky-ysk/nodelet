package wasm

import (
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/component-base/logs"
)

type WasmRuntime struct {
}

func NewWasmRuntime() WasmRuntime {
	return WasmRuntime{}
}
func (wr WasmRuntime) Run(group *apis.Group, action *apis.Action, runtime *apis.Runtime) error {
	logs.Infof("wasm runtime for task:%s", group.Name)
	return nil
}
func (wr WasmRuntime) Kill(group *apis.Group, action *apis.Action, runtime *apis.Runtime) error {
	logs.Infof("wasm runtime kill task:%s", group.Name)
	return nil
}
func (wr WasmRuntime) CheckTaskStatus(group *apis.Group, action *apis.Action, runtime *apis.Runtime) (string, error) {

	return "", nil
}

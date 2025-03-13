package main

import (
	"time"

	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/nodelet/events/eventbus"
	"hit.edu/framework/pkg/nodelet/task/runtime/command"
	"hit.edu/framework/pkg/nodelet/task/runtime/wasm"

	"hit.edu/framework/pkg/component-base/logs"
)

func main() {
	// fmt.Println("abc")
	moduleName := "testWasmModule"
	logs.Init(moduleName)
	logs.Info("---TestForAIWasm---")

	wasm_runtime := wasm.NewWasmRuntime()
	// 运行时会跑在独立的进程,所以程序结束一定要kill进程,不然始终在运行
	defer wasm_runtime.StopCMD()

	wasm_runtime.Run(&group, &action, &runtimeWasm, 0, 0)
	time.Sleep(5 * time.Second)
	wasm_runtime.Destory()

	// wasm_bycmd()
}

func wasm_bycmd() {
	commandRuntime := command.NewCommandRuntime(eventbus.NewEventBus())

	commandRuntime.Run(&group, &action, &runtimeCommand, 0, 0)
	defer commandRuntime.Kill(&group, &action, &runtimeCommand)

	time.Sleep(15 * time.Second)
}

var runtimeCommand = apis.Runtime{
	Name:    "wasm-cmd",
	Type:    apis.ByCommand,
	Command: []string{"/tmp/wasm/toolchain/wa2x-wasi-nn"},
	Args:    []string{"/tmp/wasm/onnx.so"},
}

var runtimeWasm = apis.Runtime{
	Name:    "wasm-test-ai-task",
	Image:   "/tmp/wasm/onnx.wasm", //暂时以文件本地地址进行测试
	Type:    apis.ByWasm,
	Command: []string{},
	Args:    []string{},
}
var action = apis.Action{
	Spec: apis.ActionSpec{
		Name: "TestAction",
		Runtimes: []apis.Runtime{
			runtimeCommand,
			// runtimeWasm,
		},
	},
}
var group = apis.Group{
	ObjectMeta: meta.ObjectMeta{Name: "wasm_inference"},
	Spec: apis.GroupSpec{
		Name:    "TestGroup",
		Parents: make([]string, 0),
		Actions: []apis.Action{
			action,
		},
	},
	Status: apis.GroupStatus{
		GroupID: "test-wasm-inference-group",
	},
}

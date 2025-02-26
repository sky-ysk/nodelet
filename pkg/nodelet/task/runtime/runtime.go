package runtime

import (
	"fmt"
	"sync"

	"hit.edu/framework/pkg/nodelet/events/eventbus"

	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/component-base/logs"
	"hit.edu/framework/pkg/nodelet/task/runtime/binary"
	"hit.edu/framework/pkg/nodelet/task/runtime/command"
	"hit.edu/framework/pkg/nodelet/task/runtime/container"
	"hit.edu/framework/pkg/nodelet/task/runtime/device"
	"hit.edu/framework/pkg/nodelet/task/runtime/k8s"
	"hit.edu/framework/pkg/nodelet/task/runtime/net"
	"hit.edu/framework/pkg/nodelet/task/runtime/wasm"
)

type Runtime interface {
	Run(group *apis.Group, action *apis.Action, runtime *apis.Runtime, actionIndex int, runtimeIndex int) error
	Kill(group *apis.Group, action *apis.Action, runtime *apis.Runtime) error
	CheckRuntimeStatus(group *apis.Group, action *apis.Action, runtime *apis.Runtime) (string, error)
	StoreData(group *apis.Group, action *apis.Action, runtime *apis.Runtime, actionIndex int, runtimeIndex int) string
	RestoreData(group *apis.Group, action *apis.Action, runtime *apis.Runtime, actionIndex int, runtimeIndex int) error
	StartRuntime(group *apis.Group, action *apis.Action, runtime *apis.Runtime, actionIndex int, runtimeIndex int) error
	InitRuntime(group *apis.Group, action *apis.Action, runtime *apis.Runtime, actionIndex int, runtimeIndex int) error
	StopRuntime(group *apis.Group, action *apis.Action, runtime *apis.Runtime, actionIndex int, runtimeIndex int) error
}

type RuntimeManager struct {
	runtimes map[apis.RuntimeType]Runtime
	eventbus *eventbus.EventBus
	mu       sync.Mutex
}

func NewRuntimeManager(bus *eventbus.EventBus) *RuntimeManager {
	return &RuntimeManager{
		runtimes: make(map[apis.RuntimeType]Runtime),
		eventbus: bus,
	}
}

var _ Runtime = &RuntimeManager{}

func (rm *RuntimeManager) GetRuntime(rt apis.RuntimeType) Runtime {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	runtime, exists := rm.runtimes[rt]
	if !exists {
		switch rt {
		case apis.ByBinary: //将预编译的二进制文件部署
			//TODO
			runtime = binary.NewBinaryRuntime()
			break
		case apis.ByPod, apis.ByDeployment, apis.ByService: //k8s-Pod\k8s-deployment\k8s-service
			//TODO
			runtime = k8s.NewK8sRuntime()
			break
		case apis.ByWasm:
			//TODO
			runtime = wasm.NewWasmRuntime()
			break
		case apis.ByCommand: //任务作为系统命令执行
			runtime = command.NewCommandRuntime(rm.eventbus)
			break
		case apis.ByDocker: //部署在Docker运行时上，非k8s
			runtime = container.NewContainerRuntime()
			break
		case apis.ByDevice: //面向特定的物理设备
			runtime = device.NewDeviceRuntime()
			break
		case apis.ByNet: //基于网络的部署
			runtime = net.NewNetRuntime()
			break
		default:
			logs.Info("unknown runtime type:%s", rt)
			break
		}
		rm.runtimes[rt] = runtime
	}
	return runtime
}

func (rm *RuntimeManager) Run(group *apis.Group, action *apis.Action, runtime *apis.Runtime, actionIndex, runtimeIndex int) error {
	return rm.GetRuntime(runtime.Type).Run(group, action, runtime, actionIndex, runtimeIndex)
}
func (rm *RuntimeManager) Kill(group *apis.Group, action *apis.Action, runtime *apis.Runtime) error {
	if rm == nil {
		logs.Error("runtime manager is nil")
		return fmt.Errorf("RuntimeManager is not initialized")
	}
	return rm.GetRuntime(runtime.Type).Kill(group, action, runtime)
}
func (rm *RuntimeManager) CheckRuntimeStatus(group *apis.Group, action *apis.Action, runtime *apis.Runtime) (string, error) {
	//TODO
	return "", nil
}
func (rm *RuntimeManager) StoreData(group *apis.Group, action *apis.Action, runtime *apis.Runtime, actionIndex int, runtimeIndex int) string {
	return rm.GetRuntime(runtime.Type).StoreData(group, action, runtime, actionIndex, runtimeIndex)
}
func (rm *RuntimeManager) RestoreData(group *apis.Group, action *apis.Action, runtime *apis.Runtime, actionIndex int, runtimeIndex int) error {
	return rm.GetRuntime(runtime.Type).RestoreData(group, action, runtime, actionIndex, runtimeIndex)
}
func (rm *RuntimeManager) StartRuntime(group *apis.Group, action *apis.Action, runtime *apis.Runtime, actionIndex int, runtimeIndex int) error {
	return rm.GetRuntime(runtime.Type).StartRuntime(group, action, runtime, actionIndex, runtimeIndex)
}
func (rm *RuntimeManager) InitRuntime(group *apis.Group, action *apis.Action, runtime *apis.Runtime, actionIndex int, runtimeIndex int) error {
	return rm.GetRuntime(runtime.Type).InitRuntime(group, action, runtime, actionIndex, runtimeIndex)
}
func (rm *RuntimeManager) StopRuntime(group *apis.Group, action *apis.Action, runtime *apis.Runtime, actionIndex int, runtimeIndex int) error {
	return rm.GetRuntime(runtime.Type).StopRuntime(group, action, runtime, actionIndex, runtimeIndex)
}

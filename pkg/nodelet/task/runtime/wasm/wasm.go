package wasm

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"time"

	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/client-go/util/manager"
	"hit.edu/framework/pkg/component-base/logs"
	"hit.edu/framework/pkg/nodelet/events"
	"hit.edu/framework/pkg/nodelet/events/eventbus"
	wasm_client "hit.edu/framework/pkg/nodelet/task/interaction/intwithWasm/grpc-client"
)

type WasmRuntime struct {
	// 配置wasm运行时的基本设置
	config Config
	// 对应运行时所在进程
	cmd *exec.Cmd
	// 对应grpc客户端
	wasmClient *wasm_client.WasmClient
	eventBus   *eventbus.EventBus
	// stopSignals map[string]chan struct{} // 用于标记进程是否被外部停止
	clientsManager *manager.Manager
	ctx            context.Context
	runtimeState   bool
}

type Config struct {
	wasmDir         string
	runtimeExecfile string //wasm runtime server文件地址
	wasmLLVM        string //wasm aot compiler 文件地址
	rpcPort         string
}

// todo:增加config，配置rpc端口和运行时信息
// todo:将config配置和runtime.args组合为启动参数
func NewWasmRuntime(clientsManager *manager.Manager, eventBus *eventbus.EventBus, wasmToolchainDir string, wasmRuntimePort string, ctx context.Context) *WasmRuntime {
	// 请将地址修改到运行时二进制文件的位置，后续考虑将config作为wasm runtime的配置文件  ---是否是可以直接把地址配置到NewWasmRuntime当中，提前加载wasm运行时
	config := Config{
		wasmDir: wasmToolchainDir,
		// runtimeExecfile: wasmToolchainDir + "/toolchain/server",
		// wasmLLVM:        wasmToolchainDir + "/toolchain/wa2xc",
		rpcPort: wasmRuntimePort, //在运行时里暂时写死了rpc端口，后续将rpc端口作为启动参数
	}
	switch runtime.GOOS {
	case "linux":
		config.runtimeExecfile = filepath.Join(wasmToolchainDir, "toolchain", "server")
		config.wasmLLVM = filepath.Join(wasmToolchainDir, "toolchain", "wa2xc")
	case "windows":
		config.runtimeExecfile = filepath.Join(wasmToolchainDir, "toolchain", "server.exe")
		config.wasmLLVM = filepath.Join(wasmToolchainDir, "toolchain", "wa2xc.exe")
	default:
	}
	wr := &WasmRuntime{
		config:   config,
		eventBus: eventBus,
		// stopSignals: make(map[string]chan struct{}),
		clientsManager: clientsManager,
		ctx:            ctx,
	}
	// 拉起运行时
	wr.pullRuntimeProcess()
	// logs.Info("pull wasm runtime")
	// err := wr.startCMD(config.runtimeExecfile, []string{config.rpcPort})
	// if err != nil {
	// 	logs.Errorf("Failed to run cmd: %v", err)
	// 	return nil
	// }
	return wr
}

func (wr *WasmRuntime) pullRuntimeProcess() error {
	if !wr.runtimeState {
		logs.Info("pull wasm runtime")
		err := wr.startCMD(wr.config.runtimeExecfile, []string{wr.config.rpcPort})
		if err != nil {
			logs.Errorf("Failed to run cmd: %v", err)
			return err
		}
		wr.runtimeState = true
		go wr.checkRuntimeTerminal()
	}
	return nil
}

func (wr *WasmRuntime) checkRuntimeTerminal() {
	<-wr.ctx.Done() // 当 channel 关闭时，会立即触发此 case
	logs.Info("wasm runtime terminal")
	wr.StopCMD()
	// wr.cmd.Process.Kill()
}

func ensureFile() error {
	return nil
}

// 启动任务
func (wr *WasmRuntime) Run(group *apis.Group, action *apis.Action, runtime *apis.Runtime, actionSpecName, runtimeSpecName string) error {
	logs.Infof("wasm runtime Run() for task:%s", group.Name)
	wasm_file := runtime.Spec.Image
	if wr.wasmClient == nil {
		wr.wasmClient = wasm_client.NewClient(context.Background(), wr.config.rpcPort, runtime.Name)
	}

	_, err := wr.wasmClient.Deploy(wasm_file, wr.config.wasmDir)
	if err != nil {
		wr.notifyRuntimeStartPhase(group.Name, group.Namespace, actionSpecName, runtimeSpecName, strconv.Itoa(wr.cmd.Process.Pid), apis.Failed, apis.Time{time.Now()}, apis.Time{time.Now()})
		return err
	}
	_, err = wr.wasmClient.Init()
	if err != nil {
		wr.notifyRuntimeStartPhase(group.Name, group.Namespace, actionSpecName, runtimeSpecName, strconv.Itoa(wr.cmd.Process.Pid), apis.Failed, apis.Time{time.Now()}, apis.Time{time.Now()})
		return err
	}
	done := make(chan bool, 1)
	go func() {
		_, err = wr.wasmClient.Start()
		if err != nil {
			done <- false
		}
		done <- true
	}()
	wr.notifyRuntimeStartPhase(group.Name, group.Namespace, actionSpecName, runtimeSpecName, strconv.Itoa(wr.cmd.Process.Pid), apis.Running, apis.Time{time.Now()}, apis.Time{time.Now()})
	if success, ok := <-done; ok {
		if !success {
			wr.notifyRuntimeStartPhase(group.Name, group.Namespace, actionSpecName, runtimeSpecName, strconv.Itoa(wr.cmd.Process.Pid), apis.Failed, apis.Time{time.Now()}, apis.Time{time.Now()})
			return err
		}
	}
	return nil
}

// 关闭任务
func (wr *WasmRuntime) Kill(group *apis.Group, action *apis.Action, runtime *apis.Runtime, actionSpecName, runtimeSpecName string) error {
	// if wr.stopSignals[runtime.Name] == nil {
	// 	logs.Infof("stopSignal for runtime %s already closed", runtime.Name)
	// 	return nil
	// }
	// close(wr.stopSignals[runtime.Name]) // 关闭通道，标记进程被外部停止  这里是一个问题，这个变量全局只能关一次？不然就报错了

	defer wr.StopCMD() // 这里是把运行时进程给关闭了，因为目前运行时提供的destory接口不能关闭正在运行的任务
	logs.Infof("wasm runtime kill for task:%s", runtime.Name)
	if wr.wasmClient == nil {
		wr.wasmClient = wasm_client.NewClient(context.Background(), wr.config.rpcPort, runtime.Name)
	}
	_, err := wr.wasmClient.Destory()
	if err != nil {
		wr.notifyRuntimeEndPhase(group.Name, group.Namespace, actionSpecName, runtimeSpecName, apis.Failed, apis.Time{time.Now()}, apis.Time{time.Now()})
		return err
	}
	wr.notifyRuntimeEndPhase(group.Name, group.Namespace, actionSpecName, runtimeSpecName, apis.Unknown, apis.Time{time.Now()}, apis.Time{time.Now()})
	return nil
}
func (wr *WasmRuntime) Stop(group *apis.Group, action *apis.Action, runtime *apis.Runtime, actionSpecName, runtimeSpecName string) error {
	return nil
}
func (wr *WasmRuntime) Restore(group *apis.Group, action *apis.Action, runtime *apis.Runtime, actionSpecName, runtimeSpecName string) error {
	return nil
}

// 需要保存进程的pid，检查进程是否是正常执行完成
func (wr *WasmRuntime) startCMD(cmd string, args []string) error {
	wr.cmd = exec.Command(cmd, args...)
	wr.cmd.Stdout = os.Stdout
	wr.cmd.Stderr = os.Stderr
	// 设置aot编译器环境变量,打开rust日志信息,设置 推理资源文件夹路径
	llvm := fmt.Sprintf("WASM_LLVM=%s", wr.config.wasmLLVM)
	fixtures := fmt.Sprintf("FIXTURES_DIR=%s/fixtures", wr.config.wasmDir)
	// export LD_LIBRARY_PATH=/tmp/kcm/wasm/toolchain:$LD_LIBRARY_PATH
	lib := filepath.Join(wr.config.wasmDir, "toolchain")
	ldLibraryPath := fmt.Sprintf("LD_LIBRARY_PATH=%s:$LD_LIBRARY_PATH", lib)
	wr.cmd.Env = append(os.Environ(), llvm, "RUST_LOG=info", fixtures, ldLibraryPath)
	logs.Info(cmd)
	logs.Info(llvm)

	err := wr.cmd.Start()
	if err != nil {
		logs.Errorf("Failed to run cmd: %v", err)
		return err
	}
	info := fmt.Sprintf("可执行文件已启动,PID:%d", wr.cmd.Process.Pid)
	logs.Info(info)

	// // 等待命令完成
	// if err := wr.cmd.Wait(); err != nil {
	// 	logs.V2().Errorf("命令执行失败: %v\n", err)
	// 	return err
	// }
	// // fmt.Println("可执行文件已完成")
	return nil
}

// 终止wasm运行时进程(区分关闭wasm任务)
func (wr *WasmRuntime) StopCMD() {
	info := fmt.Sprintf("stop wasm runtime process pid : %d", wr.cmd.Process.Pid)
	logs.Info(info)
	if err := wr.cmd.Process.Kill(); err != nil {
		logs.Error("停止wasm runtime 进程 %d 失败: %v\n", wr.cmd.Process.Pid, err)
	} else {
		logs.Info("wasm runtime 进程已停止")
	}
	wr.runtimeState = false
}
func (wr WasmRuntime) CheckRuntimeStatus(group *apis.Group, action *apis.Action, runtime *apis.Runtime) (string, error) {

	return "", nil
}

func (wr WasmRuntime) InitRuntime(group *apis.Group, action *apis.Action, runtime *apis.Runtime, actionSpecName, runtimeSpecName string) error {
	logs.Infof("wasm runtime InitRuntime() for task:%s", runtime.Name)
	wasm_file := runtime.Spec.Image
	if wr.wasmClient == nil {
		wr.wasmClient = wasm_client.NewClient(context.Background(), wr.config.rpcPort, runtime.Name)
	}

	_, err := wr.wasmClient.Deploy(wasm_file, wr.config.wasmDir)
	if err != nil {
		wr.notifyRuntimeStartPhase(group.Name, group.Namespace, actionSpecName, runtimeSpecName, strconv.Itoa(wr.cmd.Process.Pid), apis.Failed, apis.Time{time.Now()}, apis.Time{time.Now()})
		logs.Errorf("任务启动失败 Deploy: %e", err)
		return err
	}
	_, err = wr.wasmClient.Init()
	if err != nil {
		wr.notifyRuntimeStartPhase(group.Name, group.Namespace, actionSpecName, runtimeSpecName, strconv.Itoa(wr.cmd.Process.Pid), apis.Failed, apis.Time{time.Now()}, apis.Time{time.Now()})
		logs.Errorf("任务启动失败 Init: %e", err)
		return err
	}
	wr.notifyRuntimeStartPhase(group.Name, group.Namespace, actionSpecName, runtimeSpecName, strconv.Itoa(wr.cmd.Process.Pid), apis.Init, apis.Time{time.Now()}, apis.Time{time.Now()})
	return nil
}

func (wr WasmRuntime) StartRuntime(group *apis.Group, action *apis.Action, runtime *apis.Runtime, actionSpecName, runtimeSpecName string) error {
	logs.Infof("wasm runtime StartRuntime() for task:%s", runtime.Name)
	wasm_file := runtime.Spec.Image
	if wr.wasmClient == nil {
		wr.wasmClient = wasm_client.NewClient(context.Background(), wr.config.rpcPort, runtime.Name)
	}

	_, err := wr.wasmClient.Deploy(wasm_file, wr.config.wasmDir)
	if err != nil {
		wr.notifyRuntimeStartPhase(group.Name, group.Namespace, actionSpecName, runtimeSpecName, strconv.Itoa(wr.cmd.Process.Pid), apis.Failed, apis.Time{time.Now()}, apis.Time{time.Now()})
		return err
	}
	_, err = wr.wasmClient.Init()
	if err != nil {
		wr.notifyRuntimeStartPhase(group.Name, group.Namespace, actionSpecName, runtimeSpecName, strconv.Itoa(wr.cmd.Process.Pid), apis.Failed, apis.Time{time.Now()}, apis.Time{time.Now()})
		return err
	}
	done := make(chan bool, 1)
	go func() {
		_, err = wr.wasmClient.Start()
		if err != nil {
			done <- false
		}
		done <- true
	}()
	wr.notifyRuntimeStartPhase(group.Name, group.Namespace, actionSpecName, runtimeSpecName, strconv.Itoa(wr.cmd.Process.Pid), apis.Running, apis.Time{time.Now()}, apis.Time{time.Now()})
	if success, ok := <-done; ok {
		if !success {
			wr.notifyRuntimeStartPhase(group.Name, group.Namespace, actionSpecName, runtimeSpecName, strconv.Itoa(wr.cmd.Process.Pid), apis.Failed, apis.Time{time.Now()}, apis.Time{time.Now()})
			return err
		}
	}
	// todo:后续增加EndPhase(apis.succeed)
	// wr.notifyRuntimeStartPhase(group.Name, group.Namespace, actionSpecName, runtimeSpecName, strconv.Itoa(wr.cmd.Process.Pid), apis.Running, apis.Time{time.Now()}, apis.Time{time.Now()})
	return nil
}

func (wr WasmRuntime) StoreData(group *apis.Group, action *apis.Action, runtime *apis.Runtime, actionSpecName, runtimeSpecName string) string {
	logs.Infof("wasm runtime StoreData() for task:%s", runtime.Name)
	var keyStatus string = ""
	if wr.wasmClient == nil {
		wr.wasmClient = wasm_client.NewClient(context.Background(), wr.config.rpcPort, runtime.Name)
	}
	result, err := wr.wasmClient.Store()
	if err != nil {
		// wr.notifyRuntimeStartPhase(group.Name, group.Namespace, actionSpecName, runtimeSpecName, strconv.Itoa(wr.cmd.Process.Pid), apis.Failed, apis.Time{time.Now()}, apis.Time{time.Now()})
		return ""
	}

	// data := []byte("\x04\x00\x00\x00\xdb\x00\x00\x00") // b"\x04\0\0\0\xdb\0\0\0"
	data := []byte{0x04, 0x00, 0x00, 0x00, 0xC0, 0xc6, 0x2D, 0x00} // [4,3000000]
	keyStatus = base64.StdEncoding.EncodeToString(data)
	if result.StateCode == 0 && result.Data != nil {
		keyStatus = base64.StdEncoding.EncodeToString(result.Data.Data)
	} else {
		logs.Infof("Failed to retrieve status for task:%s", runtime.Name)
	}
	// logs.Info("--WasmRuntime StoreData()--")
	return keyStatus
}

func (wr WasmRuntime) RestoreData(group *apis.Group, action *apis.Action, runtime *apis.Runtime, actionSpecName, runtimeSpecName string) error {
	logs.Infof("wasm runtime RestoreData() for task:%s", runtime.Name)
	if wr.wasmClient == nil {
		wr.wasmClient = wasm_client.NewClient(context.Background(), wr.config.rpcPort, runtime.Name)
	}
	var err error

	// etcdRuntime, err := wr.clientsManager.GetRuntime(runtime.Name, runtime.Namespace)
	// if err != nil {
	// 	logs.Errorf("WasmRuntime: Failed to get runtime '%s': %v", runtime.Name, err)
	// }
	var keyStatus string
	for keyStatus == "" {
		logs.Infof("WasmRuntime: waiting keyStatus ...")
		etcdRuntime, err := wr.clientsManager.GetRuntime(runtime.Name, runtime.Namespace)
		if err != nil {
			logs.Errorf("WasmRuntime: Failed to get runtime '%s': %v", runtime.Name, err)
		}
		keyStatus = etcdRuntime.Status.KeyStatus
	}
	// logs.Infof("WasmRuntime: keyStatus: %s", keyStatus)

	// runtimeStatus := &action.Status.RuntimeStatus[runtimeIndex]
	// keyStatus := runtimeStatus.KeyStatus
	if keyStatus != "" {
		logs.Infof("WasmRuntime: keyStatus: %s", keyStatus)
		keyStatus, _ := base64.StdEncoding.DecodeString(keyStatus)
		_, err := wr.wasmClient.Restore(keyStatus)
		if err != nil {
			return err
		}
	} else {
		logs.Infof("Status data is empty, skip 'restore' and start the task(%s) directly", runtime.Name)
	}

	done := make(chan bool, 1)
	go func() {
		_, err = wr.wasmClient.Start()
		if err != nil {
			done <- false
		}
		done <- true
	}()
	wr.notifyRuntimeStartPhase(group.Name, group.Namespace, actionSpecName, runtimeSpecName, strconv.Itoa(wr.cmd.Process.Pid), apis.Running, apis.Time{time.Now()}, apis.Time{time.Now()})
	if success, ok := <-done; ok {
		if !success {
			wr.notifyRuntimeStartPhase(group.Name, group.Namespace, actionSpecName, runtimeSpecName, strconv.Itoa(wr.cmd.Process.Pid), apis.Failed, apis.Time{time.Now()}, apis.Time{time.Now()})
			return err
		}
	}
	logs.Info("--WasmRuntime RestoreData()--")
	return nil
}

func (wr WasmRuntime) StopRuntime(group *apis.Group, action *apis.Action, runtime *apis.Runtime, actionSpecName, runtimeSpecName string) error {
	defer wr.StopCMD()
	logs.Infof("wasm runtime StopRuntime for task:%s", runtime.Name)
	if wr.wasmClient == nil {
		wr.wasmClient = wasm_client.NewClient(context.Background(), wr.config.rpcPort, runtime.Name)
	}
	_, err := wr.wasmClient.Destory()
	if err != nil {
		return err
	}
	wr.notifyRuntimeEndPhase(group.Name, group.Namespace, actionSpecName, runtimeSpecName, apis.Successed, apis.Time{time.Now()}, apis.Time{time.Now()})
	return nil
}

// 销毁
func (wr *WasmRuntime) Destory() error {
	defer wr.StopCMD()
	logs.Infof("wasm runtime destory for task: wasm-test")
	// if wr.wasmClient == nil {
	// 	// 该方法应增加err返回值
	// 	logs.Error(" no corresponding RPC connection : ")
	// 	return fmt.Errorf(" no corresponding RPC connection : ")
	// }
	if wr.wasmClient == nil {
		// wr.wasmClient = wasm_client.NewClient(context.Background(), wr.config.rpcPort, runtime.Name)
	}
	_, err := wr.wasmClient.Destory()
	if err != nil {
		return err
	}
	return nil
}

// 通过 EventBus 通知 Runtime 状态更新
func (cr *WasmRuntime) notifyRuntimeStartPhase(groupName, groupNamespace string, actionSpeName, runtimeSpecName string, processId string, phase apis.Phase, startAt, lastTime apis.Time) {
	event := events.RuntimeStartPhaseEvent1{
		GroupName:       groupName,
		GroupNamespace:  groupNamespace,
		ActionSpecName:  actionSpeName,
		RuntimeSpecName: runtimeSpecName,
		ProcessId:       processId,
		Phase:           phase,
		StartAt:         startAt,
		LastTime:        lastTime,
	}
	cr.eventBus.Publish(event)
}
func (cr *WasmRuntime) notifyRuntimeEndPhase(groupName, groupNamespace string, actionSpeName, runtimeSpecName string, phase apis.Phase, finishTime, lastTime apis.Time) {
	event := events.RuntimeEndPhaseEvent1{
		GroupName:       groupName,
		GroupNamespace:  groupNamespace,
		ActionSpecName:  actionSpeName,
		RuntimeSpecName: runtimeSpecName,
		Phase:           phase,
		FinishAt:        finishTime,
		LastTime:        lastTime,
	}
	cr.eventBus.Publish(event)
}

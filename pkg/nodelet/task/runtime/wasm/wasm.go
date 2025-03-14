package wasm

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/component-base/logs"
	wasm_client "hit.edu/framework/pkg/nodelet/task/interaction/intwithWasm/grpc-client"
)

type WasmRuntime struct {
	// 配置wasm运行时的基本设置
	config Config
	// 对应运行时所在进程
	cmd *exec.Cmd
	// 对应grpc客户端
	wasmClient *wasm_client.WasmClient
}

type Config struct {
	runtimeExecfile string //wasm runtime server文件地址
	wasmLLVM        string //wasm aot compiler 文件地址
	rpcPort         string
}

// todo:增加config，配置rpc端口和运行时信息
// todo:将config配置和runtime.args组合为启动参数
func NewWasmRuntime() *WasmRuntime {
	// 请将地址修改到运行时二进制文件的位置，后续考虑将config作为wasm runtime的配置文件  ---是否是可以直接把地址配置到NewWasmRuntime当中，提前加载wasm运行时
	config := Config{
		runtimeExecfile: "/tmp/wasm/toolchain/server",
		wasmLLVM:        "/tmp/wasm/toolchain/wasm-llvm",
		rpcPort:         "8080", //在运行时里暂时写死了rpc端口，所以不能改，后续考虑将rpc端口作为启动参数
	}
	wr := &WasmRuntime{config: config}
	// wr.rpcAddr = config.rpcAddr
	// 拉起运行时
	logs.Info("pull wasm runtime")
	err := wr.startCMD(config.runtimeExecfile, []string{})
	if err != nil {
		logs.Errorf("Failed to run cmd: %v", err)
		return nil
	}
	return wr
}
func ensureFile() error {
	return nil
}

// 启动任务
func (wr *WasmRuntime) Run(group *apis.Group, action *apis.Action, runtime *apis.Runtime, actionIndex, runtimeIndex int) error {
	logs.Infof("wasm runtime Run() for task:%s", group.Name)
	wasm_file := runtime.Image
	if wr.wasmClient == nil {
		wr.wasmClient = wasm_client.NewClient(context.Background(), wr.config.rpcPort, "wasm-test-demo")
	}

	_, err := wr.wasmClient.Deploy(wasm_file)
	if err != nil {
		return err
	}
	_, err = wr.wasmClient.Init()
	if err != nil {
		return err
	}
	_, err = wr.wasmClient.Start()
	if err != nil {
		return err
	}
	// time.Sleep(1 * time.Second)
	return nil
}

// 关闭任务
func (wr *WasmRuntime) Kill(group *apis.Group, action *apis.Action, runtime *apis.Runtime, actionIndex, runtimeIndex int) error {
	logs.Infof("wasm runtime kill task:%s", group.Name)
	_, err := wr.wasmClient.Destory()
	if err != nil {
		return err
	}
	// time.Sleep(2 * time.Second)
	return nil
}

// 需要保存进程的pid，检查进程是否是正常执行完成
func (wr *WasmRuntime) startCMD(cmd string, args []string) error {
	wr.cmd = exec.Command(cmd, args...)
	wr.cmd.Stdout = os.Stdout
	wr.cmd.Stderr = os.Stderr
	// 设置aot编译器环境变量,打开rust日志信息,设置 推理资源文件夹路径
	llvm := fmt.Sprintf("WASM_LLVM=%s", wr.config.wasmLLVM)
	fixtures := fmt.Sprintf("FIXTURES_DIR=/tmp/wasm/fixtures")
	wr.cmd.Env = append(os.Environ(), llvm, "RUST_LOG=info", fixtures)
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

// 停止进程，主要是关闭wasm运行时这个进程
func (wr *WasmRuntime) StopCMD() {
	info := fmt.Sprintf("stop cmd process pid : %d", wr.cmd.Process.Pid)
	logs.Info(info)
	if err := wr.cmd.Process.Kill(); err != nil {
		logs.Error("停止进程 %d 失败: %v\n", wr.cmd.Process.Pid, err)
	} else {
		logs.Info("进程已停止")
	}
}
func (wr WasmRuntime) CheckRuntimeStatus(group *apis.Group, action *apis.Action, runtime *apis.Runtime) (string, error) {

	return "", nil
}

func (wr WasmRuntime) InitRuntime(group *apis.Group, action *apis.Action, runtime *apis.Runtime, actionIndex int, runtimeIndex int) error {

	return nil
}

func (wr WasmRuntime) StartRuntime(group *apis.Group, action *apis.Action, runtime *apis.Runtime, actionIndex int, runtimeIndex int) error {
	logs.Infof("wasm runtime StartRuntime() for task:%s", group.Name)
	wasm_file := runtime.Image
	if wr.wasmClient == nil {
		wr.wasmClient = wasm_client.NewClient(context.Background(), wr.config.rpcPort, "wasm-test-demo")
	}

	_, err := wr.wasmClient.Deploy(wasm_file)
	if err != nil {
		logs.Errorf("任务启动失败: %e", err)
		return err
	}
	_, err = wr.wasmClient.Init()
	if err != nil {
		logs.Errorf("任务启动失败: %e", err)
		return err
	}
	_, err = wr.wasmClient.Start()
	if err != nil {
		logs.Errorf("任务启动失败: %e", err)
		return err
	}

	return nil
}

func (wr WasmRuntime) StoreData(group *apis.Group, action *apis.Action, runtime *apis.Runtime, actionIndex int, runtimeIndex int) string {
	if wr.wasmClient == nil {
		// 该方法应增加err返回值
		// return fmt.Errorf(" no corresponding RPC connection : %v", runtimeIndex)
		return ""
	}
	return ""
}

func (wr WasmRuntime) RestoreData(group *apis.Group, action *apis.Action, runtime *apis.Runtime, actionIndex int, runtimeIndex int) error {
	// keyStatus := ""
	//logs.Infof("keyStatus: %s", keyStatus)
	for {
		if wr.wasmClient != nil {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	return nil
}

func (wr WasmRuntime) StopRuntime(group *apis.Group, action *apis.Action, runtime *apis.Runtime, actionIndex int, runtimeIndex int) error {
	logs.Infof("wasm runtime destory for task: wasm-test")
	_, err := wr.wasmClient.Destory()
	if err != nil {
		return err
	}
	return nil
}

// 销毁任务
func (wr *WasmRuntime) Destory() error {
	logs.Infof("wasm runtime destory for task: wasm-test")
	_, err := wr.wasmClient.Destory()
	if err != nil {
		return err
	}
	return nil
}

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
	rpcAddr         string
}

// todo:增加config，配置rpc端口和运行时信息
// todo:将config配置和runtime.args组合为启动参数
func NewWasmRuntime() *WasmRuntime {
	// 请将地址修改到运行时二进制文件的位置，后续考虑将config作为wasm runtime的配置文件  ---是否是可以直接把地址配置到NewWasmRuntime当中，提前加载wasm运行时
	config := Config{
		runtimeExecfile: "/tmp/wasm/toolchain/server",
		wasmLLVM:        "/tmp/wasm/toolchain/wasm-llvm",
		rpcAddr:         "127.0.0.1:8080", //在运行时里暂时写死了rpc端口，所以不能改，后续考虑将rpc端口作为启动参数
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
	logs.Infof("wasm runtime for task:%s", group.Name)
	wasm_file := runtime.Image
	wr.wasmClient = wasm_client.NewClient(context.Background(), wr.config.rpcAddr)

	time.Sleep(1 * time.Second)
	err := wr.wasmClient.Connect()
	if err != nil {
		logs.Error("wasm client 连接失败:", err)
		return err
	}
	_, err = wr.wasmClient.Deploy(wasm_file)
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
	time.Sleep(1 * time.Second)
	return nil
}

// 关闭任务
func (wr *WasmRuntime) Kill(group *apis.Group, action *apis.Action, runtime *apis.Runtime) error {
	logs.Infof("wasm runtime kill task:%s", group.Name)
	_, err := wr.wasmClient.Destory()
	if err != nil {
		return err
	}
	time.Sleep(2 * time.Second)
	return nil
}

// 销毁运行时
func (wr *WasmRuntime) Destory() error {
	logs.Infof("wasm runtime destory for task: wasm-test")
	_, err := wr.wasmClient.Destory()
	if err != nil {
		return err
	}
	time.Sleep(2 * time.Second)
	return nil
}

// 后续改成使用cmd package里的build cmd等
// 需要保存进程的pid，检查进程是否是正常执行完成
func (wr *WasmRuntime) startCMD(cmd string, args []string) error {
	wr.cmd = exec.Command(cmd, args...)
	wr.cmd.Stdout = os.Stdout
	wr.cmd.Stderr = os.Stderr
	// 设置aot编译器环境变量,打开rust日志信息
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
		logs.Error("停止进程失败: %v\n", err)
	} else {
		logs.Info("进程已停止")
	}
}
func (wr WasmRuntime) CheckTaskStatus(group *apis.Group, action *apis.Action, runtime *apis.Runtime) (string, error) {

	return "", nil
}

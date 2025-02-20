package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"time"

	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/component-base/logs"
	"hit.edu/framework/pkg/nodelet/events/eventbus"
	"hit.edu/framework/pkg/nodelet/task/interaction/intwithRuntime"
	grpc_client "hit.edu/framework/pkg/nodelet/task/interaction/intwithRuntime/grpc-client"
	"hit.edu/framework/pkg/nodelet/task/runtime/command"
)

func main() {
	moduleName := "testModule"
	logs.Init(moduleName)

	runtime_test()

}

func rpc_client_test() {
	runtimeId := "runtime-1"
	logs.Infof("runtime for task:%s", runtimeId)
	cmd := pullService()
	defer stopCMD(cmd)

	port := "5123"
	client := grpc_client.NewRuntimeClient(port, runtimeId)

	// rpc调用init()
	_, error := client.RunAppInit()
	if error != nil {
		logs.Infof("任务init失败: %e", error)
	}

	// rpc调用start()
	_, error = client.RunAppStart()
	if error != nil {
		logs.Infof("任务启动失败: %e", error)
	}

	time.Sleep(2 * time.Second)

	// rpc调用store()
	_, error = client.RunAppStore()
	if error != nil {
		logs.Errorf("任务保存状态失败: %e", error)
	}

	time.Sleep(1 * time.Second)

	// rpc调用restore()
	_, error = client.RunAppRestore("{restore data}")
	if error != nil {
		logs.Errorf("任务恢复状态失败: %e", error)
	}

	time.Sleep(1 * time.Second)
}

func runtime_test() {
	commandRuntime := command.NewCommandRuntime(eventbus.NewEventBus(), intwithRuntime.NewClientsManager())
	commandRuntime.Run(newGroup, action, runtime, 0, 0)
	defer commandRuntime.Kill(newGroup, action, runtime)

	commandRuntime.InitRuntime(newGroup, action, runtime, 0, 0)
	commandRuntime.StartRuntime(newGroup, action, runtime, 0, 0)
	time.Sleep(2 * time.Second)
	commandRuntime.StoreData(newGroup, action, runtime, 0, 0)
	time.Sleep(2 * time.Second)
	commandRuntime.RestoreData(newGroup, action, runtime, 0, 0)

	prompt()
}

func prompt() {
	fmt.Printf("-> Press Return key to continue.")
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		break
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}
}

func pullService() *exec.Cmd {
	logs.Infof("拉起任务")
	cmd := exec.Command("python3", "/home/kcm/py_examples/migration-demo-0116/yolo-runner.py")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Start()
	if err != nil {
		logs.Errorf("Failed to run cmd: %v", err)
		return nil
	}
	return cmd
}

func stopCMD(cmd *exec.Cmd) {
	info := fmt.Sprintf("stop cmd process pid : %d", cmd.Process.Pid)
	logs.Info(info)
	if err := cmd.Process.Kill(); err != nil {
		logs.Error("停止进程失败: %v\n", err)
	} else {
		logs.Info("进程已停止")
	}
}

var newGroup = &apis.Group{
	ObjectMeta: meta.ObjectMeta{Name: "migrate-example-1"},
	Spec: apis.GroupSpec{
		Name:    "migrate-example-1",
		Parents: make([]string, 0),
		Actions: []apis.Action{
			apis.Action{
				Spec: apis.ActionSpec{
					Name: "migrate-example-1",
					Runtimes: []apis.Runtime{
						apis.Runtime{
							Name:    "CMD",
							Image:   "",
							Type:    apis.ByCommand,
							Command: []string{"python3"},
							Args:    []string{"/home/kcm/py_examples/migration-demo-0116/yolo-runner.py"},
						},
					},
				},
			},
		},
	},
	Status: apis.GroupStatus{
		GroupID: "migrate-example-1",
	},
}
var groups = []*apis.Group{newGroup}
var action = &newGroup.Spec.Actions[0]
var runtime = &action.Spec.Runtimes[0]

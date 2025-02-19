package command

import (
	"fmt"
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/component-base/logs"
	"hit.edu/framework/pkg/nodelet/events"
	"hit.edu/framework/pkg/nodelet/events/eventbus"
	"hit.edu/framework/pkg/nodelet/task/interaction/intwithRuntime"
	"hit.edu/framework/pkg/nodelet/task/runtime/command/process"
	"os"
	"os/exec"
	rt "runtime"
	"strconv"
	"syscall"
	"time"
)

type CommandRuntime struct {
	processManager *process.ProcessManager
	eventBus       *eventbus.EventBus
	clientsManager *intwithRuntime.ClientsManager
	stopSignals    map[string]chan struct{} // 用于标记进程是否被外部停止
}

func NewCommandRuntime(eventBus *eventbus.EventBus, clients *intwithRuntime.ClientsManager) *CommandRuntime {
	pm := process.NewProcessManager()
	return &CommandRuntime{
		processManager: pm,
		eventBus:       eventBus,
		clientsManager: clients,
		stopSignals:    make(map[string]chan struct{}),
	}
}
func (cr *CommandRuntime) Kill(group *apis.Group, action *apis.Action, runtime *apis.Runtime) error {
	logs.Infof("command runtime kill task:%s", group.Name)
	close(cr.stopSignals[runtime.Name]) // 关闭通道，标记进程被外部停止  这里是一个问题，这个变量全局只能关一次？不然就报错了
	err := cr.stopCMD(runtime.Name)
	if err != nil {
		return err
	}
	return nil
}

func (cr *CommandRuntime) Run(group *apis.Group, action *apis.Action, runtime *apis.Runtime, actionIndex, runtimeIndex int) error {
	logs.Infof("command runtime for runtime task:%s", runtime.Name)
	// 执行时所需命令
	cmd := runtime.Command
	// Command的执行参数, 所有的参数都需要作为执行参数传入系统
	args := runtime.Args

	// 目前只接受Command中第一个元素
	err := cr.startCMD(group.Name, actionIndex, runtimeIndex, runtime, cmd[0], args)
	if err != nil {

		logs.Error("Failed to start action:\t", action.Spec.Name)
		return err
		// TODO: 输出Action的详细信息
	}
	// TODO: 输出Action的详细信息，等级为Debug
	logs.Infof("Action Name:\t %s is Running", action.Spec.Name)
	return nil
}

// 可能需要区分输出output的指定位置, 后续需要改成使用cmd package里的build cmd等
// 需要保存进程的pid，检查进程是否是正常执行完成
func (cr *CommandRuntime) startCMD(groupName string, actionIndex, runtimeIndex int, runtime *apis.Runtime, cmd string, args []string) error {
	// exec.Command可以接受的命令
	// name表示可执行二进制的name
	// ...args表示命令所需的参数

	// TODO: 不同系统平台下的CMD，根据运行平台选择对应路径下的解释器等
	//判断程序所在Linux还是Windows环境，决定python等解释器路径
	sysType := rt.GOOS
	pyPath := "/home/ysk/miniconda3/envs/py38/bin/python" //默认linux
	if sysType == "linux" {
		//可能有其他执行器路径需要指定，后续可加
	} else if sysType == "windows" {
		pyPath = "D:\\Programming\\Anaconda\\envs\\yolo\\python.exe"
	}

	//根据cmd执行类型变更cmd
	if cmd == "python" {
		cmd = pyPath
	}

	//执行参数里注入环境变量
	// args +=

	// 创建命令
	CMD := exec.Command(cmd, args...)

	//注入环境变量，完成指定python的环境变量地址的功能
	//TODO （在runtime检查依赖的时候增加action的envs，指定python等地址）
	envVars := runtime.EnvVar
	env := os.Environ() //获取当前环境的环境变量
	for key, value := range envVars {
		env = append(env, fmt.Sprintf("%s=%s", key, value))
	}
	CMD.Env = env

	// 将子进程的标准输出和标准错误直接重定向到当前程序的输出----后续需要根据group中提供的输出位置重定向？（例如输出到指定文件内）
	//outfile, err := os.Create("/home/ysk/Desktop/outfolder/" + runtime.Name + "out.txt")
	//outfile, err := os.Create("/home/ysk/Desktop/outfolder/out.txt")
	//if err != nil {
	//	panic(err)
	//}
	//defer outfile.Close()
	CMD.Stdout = os.Stdout
	//CMD.Stdout = outfile
	CMD.Stderr = os.Stderr
	// CMD.Env = append(CMD.Env, )

	// 启动命令
	logs.Infof("taskName:\t %s is Running", runtime.Name)

	if err := CMD.Start(); err != nil {
		//通知group_monitor，来修改全局的group信息（其中的runtime属性）
		cr.notifyRuntimeStartPhase(groupName, actionIndex, runtimeIndex, strconv.Itoa(CMD.Process.Pid), apis.Failed, apis.Time{time.Now()}, apis.Time{time.Now()})
		return fmt.Errorf("failed to start command: %w", err)
	}
	//通知group_monitor，来修改全局的group信息（其中的runtime属性）
	cr.notifyRuntimeStartPhase(groupName, actionIndex, runtimeIndex, strconv.Itoa(CMD.Process.Pid), apis.Running, apis.Time{time.Now()}, apis.Time{time.Now()})
	cr.processManager.AddProcess(runtime.Name, CMD)
	cr.stopSignals[runtime.Name] = make(chan struct{})
	logs.Infof("process id:\t %d is Running", CMD.Process.Pid)

	// return nil
	go cr.monitorCMD(groupName, actionIndex, runtimeIndex, runtime, CMD)

	// TODO: 使用进程启动CMD, 异步操作
	// TODO: 返回进程对应的ProcessID
	// TODO: 多个Command拼接---ysk  action包含多个command拼接指？如果一个cmd比较复杂（例如：python predict.py 10 100  data.json）
	//我们如何区分其中的参数和路径（因为执行需要predict.py  data.json的路径，（可以采取固定下载到的数据的路径，采取相对路径的方式？））
	// TODO：注入环境变量---ysk  暂时未确定环境变量的例子
	// TODO: 重定向stdout和stdin, 每个进程都有标准输出，任务监控相关组件需要重定向该输入输出--ysk  目前暂时以文件形式输出
	// TODO: 处理Action的Input和Output---ysk  output的结构后续可能需要调整，目前暂时以文件.txt的输出形式
	// 检查依赖requirements是否满足本地的环境
	return nil
}

// 监控任务的执行状态，并修改Runtime信息并上传（这里是使用事件上传，后续可能要对比传入group_manager，来修改任务信息）
func (cr *CommandRuntime) monitorCMD(groupName string, actionIndex, runtimeIndex int, runtime *apis.Runtime, CMD *exec.Cmd) {
	if err := CMD.Wait(); err != nil {
		// 检查 stopSignal 通道是否被关闭，判断进程是否是外部停止的
		if _, ok := <-cr.stopSignals[runtime.Name]; !ok {
			// 通道已关闭，说明是stopCMD终止的  两种情况：一种是用户想停止任务，一种是需要迁移，从而停止任务
			logs.Info("command killed externally by stopCMD")
			cr.notifyRuntimeEndPhase(groupName, actionIndex, runtimeIndex, apis.Unknown, apis.Time{time.Now()}, apis.Time{time.Now()})
		} else {
			logs.Errorf("command %s finished with error: %s", runtime.Name, err.Error())
			// 修改RuntimeStatus的Phase为Failed，ActionStatus的Phase也为Failed
			cr.notifyRuntimeEndPhase(groupName, actionIndex, runtimeIndex, apis.Failed, apis.Time{time.Now()}, apis.Time{time.Now()})
		}
		cr.processManager.RemoveProcess(runtime.Name)
		delete(cr.stopSignals, runtime.Name)
		return
	}
	logs.Infof("command %s completed", runtime.Name)
	//TODO 正常执行完之后通知修改queues和Manager对应的group信息，group当中Runtime的phase
	cr.processManager.MoveProcessToSucess(runtime.Name) //移入successProcess，同时移出process
	// 修改RuntimeStatus的Phase为Successed，ActionStatus的Phase也为Successed
	cr.notifyRuntimeEndPhase(groupName, actionIndex, runtimeIndex, apis.Successed, apis.Time{time.Now()}, apis.Time{time.Now()})
}

// 停止某个CMD对应的进程
// TODO：保存现场
func (cr *CommandRuntime) stopCMD(taskName string) error {
	CMD, exists := cr.processManager.GetProcess(taskName)
	if !exists {
		logs.Errorf("Failed to find task:\t ", taskName)
		return fmt.Errorf("task '%s' not found", taskName)
	}
	// 终止任务的进程
	if CMD.Process != nil {
		if err := CMD.Process.Kill(); err != nil {
			logs.Errorf("Failed to kill task:\t ", taskName)
			return fmt.Errorf("failed to stop task '%s': %w", taskName, err)
		}
		fmt.Printf("Task '%s' with PID %d has been stopped.\n", taskName, CMD.Process.Pid)
	} else {
		logs.Infof("Task '%s' is already stopped.", taskName)
		return fmt.Errorf("task '%s' process is nil", taskName)
	}
	//cr.processManager.RemoveProcess(taskName)  //这条语句不用了，直接在monitorCMD方法当中会执行
	return nil

}

// 通过 EventBus 通知 Runtime 状态更新
func (cr *CommandRuntime) notifyRuntimeStartPhase(groupName string, actionIndex, runtimeIndex int, processId string, phase apis.Phase, startAt, lastTime apis.Time) {
	event := events.RuntimeStartPhaseEvent1{
		GroupName:    groupName,
		ActionIndex:  actionIndex,
		RuntimeIndex: runtimeIndex,
		ProcessId:    processId,
		Phase:        phase,
		StartAt:      startAt,
		LastTime:     lastTime,
	}
	cr.eventBus.Publish(event)
}
func (cr *CommandRuntime) notifyRuntimeEndPhase(groupName string, actionIndex, runtimeIndex int, phase apis.Phase, finishTime, lastTime apis.Time) {
	event := events.RuntimeEndPhaseEvent1{
		GroupName:    groupName,
		ActionIndex:  actionIndex,
		RuntimeIndex: runtimeIndex,
		Phase:        phase,
		FinishAt:     finishTime,
		LastTime:     lastTime,
	}
	cr.eventBus.Publish(event)
}

//func (cr *CommandRuntime) notifyRuntimeMigratePhase(groupName string, actionIndex, runtimeIndex int, phase apis.Phase, finishTime, lastTime apis.Time) {
//	event := events.RuntimeEndPhaseEvent1{
//		GroupName:    groupName,
//		ActionIndex:  actionIndex,
//		RuntimeIndex: runtimeIndex,
//		Phase:        phase,
//		FinishAt:     finishTime,
//		LastTime:     lastTime,
//	}
//	cr.eventBus.Publish(event)
//}

// 获取任务的执行状态（正在运行or运行失败）---该方法暂时没有用到-先放着
func (cr *CommandRuntime) CheckRuntimeStatus(group *apis.Group, action *apis.Action, runtime *apis.Runtime) (string, error) {
	cmd, exeists := cr.processManager.GetProcess(group.Name)
	if !exeists {
		logs.Error("task %s is not running", group.Name)
		return "", fmt.Errorf("task %s is not running", group.Name)
	}
	//如果说cmd执行了但是中途任务退出了，cmd.Process仍然是有值的（它指向的*os.Process结构体仍然有效，非nil）
	// 如果cmd.Process的值==nil，表示：命令没有成功启动、或者exec.Command调用没有成功创建进程时（比如命令路径错误、权限问题等），说明任务根本没有被运行
	if cmd.Process == nil {
		//需要转移到Error队列当中
		return "Error", fmt.Errorf("task %s process is nil", group.Name)
	}
	// 检查进程是否仍在执行，向进程发送一个空信号，检查进程是否在运行
	if err := cmd.Process.Signal(syscall.Signal(0)); err == nil {
		// 如果发送信号不出错，表示进程仍在运行
		return "Running", nil
	}
	if err := cmd.Wait(); err != nil {
		logs.Error("task %s failed with error: %s", group.Name, err.Error())
		return "Error", err
	}
	return "Completed", nil
}

// 细粒度控制（grpc）：保存任务状态
func (cr *CommandRuntime) StoreData(group *apis.Group, action *apis.Action, runtime *apis.Runtime, actionIndex int, runtimeIndex int) string {
	// 保存任务状态，调用grpc接口获取任务状态，返回任务状态值即可
	//client, success := cr.clientsManager.GetRuntimeConnection(action.Status.RuntimeStatus[runtimeIndex].RuntimeID)
	logs.Infof("********************【模拟】成功保存了任务状态：ABCDEFG")
	return "ABCDEFG"
}

// 细粒度控制（grpc）：恢复任务状态
func (cr *CommandRuntime) RestoreData(group *apis.Group, action *apis.Action, runtime *apis.Runtime, actionIndex int, runtimeIndex int) error {
	// 恢复任务状态，调用grpc接口通知任务恢复任务状态，任务状态存放在etcd当中（group下对应runtime下的runtimeStatus下的keyStatus属性）
	keyStatus := action.Status.RuntimeStatus[runtimeIndex].KeyStatus
	// 下面要使用这个关键状态数据进行恢复，调用Grpc去控制恢复任务状态，这里首先是否需要先启动任务呢？
	//client, success := cr.clientsManager.GetRuntimeConnection(action.Status.RuntimeStatus[runtimeIndex].RuntimeID)
	logs.Infof("keyStatus: %s", keyStatus)
	logs.Info("**********************成功恢复了任务状态****************************")
	return nil
}

// 细粒度控制（grpc）：启动任务状态
func (cr *CommandRuntime) StartRuntime(group *apis.Group, action *apis.Action, runtime *apis.Runtime, actionIndex int, runtimeIndex int) error {
	// 调用grpc接口启动任务
	//client, success := cr.clientsManager.GetRuntimeConnection(action.Status.RuntimeStatus[runtimeIndex].RuntimeID)
	// 下面是为了测试 -----下面的内容是暂时测试的，后面需要替换使用grpc
	logs.Infof("command runtime for runtime task:%s", runtime.Name)
	// 执行时所需命令
	cmd := runtime.Command
	// Command的执行参数, 所有的参数都需要作为执行参数传入系统
	args := runtime.Args

	// 目前只接受Command中第一个元素
	err := cr.startCMD(group.Name, actionIndex, runtimeIndex, runtime, cmd[0], args)
	if err != nil {

		logs.Error("Failed to start action:\t", action.Spec.Name)
		return err
		// TODO: 输出Action的详细信息
	}
	// TODO: 输出Action的详细信息，等级为Debug
	logs.Infof("Action Name:\t %s is Running", action.Spec.Name)
	return nil
}

// 细粒度控制（grpc）：初始化任务
func (cr *CommandRuntime) InitRuntime(group *apis.Group, action *apis.Action, runtime *apis.Runtime, actionIndex int, runtimeIndex int) error {
	logs.Infof("runtime has Init====")
	cr.notifyRuntimeStartPhase(group.Name, actionIndex, runtimeIndex, "", apis.Init, apis.Time{time.Now()}, apis.Time{time.Now()})
	return nil
}

// 细粒度控制（grpc）：关闭任务
func (cr *CommandRuntime) StopRuntime(group *apis.Group, action *apis.Action, runtime *apis.Runtime, actionIndex int, runtimeIndex int) error {
	logs.Infof("runtime has stop====")
	cr.notifyRuntimeEndPhase(group.Name, actionIndex, runtimeIndex, apis.Successed, apis.Time{time.Now()}, apis.Time{time.Now()})
	return nil
}

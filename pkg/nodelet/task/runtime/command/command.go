package command

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"syscall"
	"time"

	"hit.edu/framework/pkg/client-go/util/manager"
	"hit.edu/framework/pkg/nodelet/task/interaction/intwithRuntime/pool"

	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/client-go/tools/recorder"
	"hit.edu/framework/pkg/component-base/logs"
	"hit.edu/framework/pkg/nodelet/events"
	"hit.edu/framework/pkg/nodelet/events/eventbus"
	grpc_client "hit.edu/framework/pkg/nodelet/task/interaction/intwithRuntime/grpc-client"
	"hit.edu/framework/pkg/nodelet/task/runtime/command/process"
)

type CommandRuntime struct {
	processManager *process.ProcessManager
	// 用于传输状态的适配Runtime运行时的事件总线
	eventBus *eventbus.EventBus
	// 全局系统的事件处理
	recorder       recorder.EventRecorder
	connectionPool *pool.ConnectionPool
	//client      *grpc_client.RuntimeClient
	stopSignals    map[string]chan struct{} // 用于标记进程是否被外部停止
	clientsManager *manager.Manager
	mu             sync.Mutex // 保护clients和stopSignals
}

func NewCommandRuntime(clientsManager *manager.Manager, eventBus *eventbus.EventBus, recorder recorder.EventRecorder, pool *pool.ConnectionPool) *CommandRuntime {
	pm := process.NewProcessManager()
	return &CommandRuntime{
		processManager: pm,
		eventBus:       eventBus,
		recorder:       recorder,
		stopSignals:    make(map[string]chan struct{}),
		connectionPool: pool,
		clientsManager: clientsManager,
	}
}
func (cr *CommandRuntime) Kill(group *apis.Group, action *apis.Action, runtime *apis.Runtime, actionSpecName, runtimeSpecName string) error {
	logs.Infof("command runtime kill task:%s", group.Name)
	// 如果 stopSignals[runtime.Name] 已经被关闭，直接返回
	if cr.stopSignals[runtime.Name] == nil {
		logs.Infof("stopSignal for runtime %s already closed", runtime.Name)
		return nil
	}
	close(cr.stopSignals[runtime.Name]) // 关闭通道，标记进程被外部停止,注意这里不发送状态，我们在Monitor部分会监听这个进程的执行，如果监听到执行失败才会发送状态
	err := cr.stopCMD(runtime)
	if err != nil {
		return err
	}
	return nil
}

func (cr *CommandRuntime) Run(group *apis.Group, action *apis.Action, runtime *apis.Runtime, actionSpecName, runtimeSpecName string) error {
	logs.Infof("command runtime for runtime task:%s", runtime.Name)
	// 执行时所需命令
	cmd := runtime.Spec.Command
	// Command的执行参数, 所有的参数都需要作为执行参数传入系统
	args := runtime.Spec.Args
	// 目前只接受Command中第一个元素
	err := cr.startCMD(group.Name, group.Namespace, actionSpecName, runtimeSpecName, runtime, cmd[0], args, false)
	if err != nil {
		logs.Info("Receive info:\t", err)
		return err
	}
	return nil
}

// 可能需要区分输出output的指定位置, 后续需要改成使用cmd package里的build cmd等
// 需要保存进程的pid，检查进程是否是正常执行完成
func (cr *CommandRuntime) startCMD(groupName, groupNamespace string, actionSpeName, runtimeSpecName string, runtime *apis.Runtime, cmd string, args []string, isInit bool) error {
	// exec.Command可以接受的命令
	// name表示可执行二进制的name
	// ...args表示命令所需的参数

	// TODO: 不同系统平台下的CMD，根据运行平台选择对应路径下的解释器等
	//判断程序所在Linux还是Windows环境，决定python等解释器路径
	// 创建命令,填入input的值
	input := runtime.Spec.Inputs
	for i := range input {
		args = append(args, input[i].Value)
	}

	envVars := runtime.Spec.EnvVar
	logs.Infof("command.go: EnvVar:%v", envVars)
	// 处理cmd
	logs.Infof("command.go: cmd:%v", cmd)
	if cmd == "python" {
		for _, value := range envVars {
			if value.Name == "" {
				continue
			}
			logs.Infof("path:%v, value:%v", value.Name, value.Value)
			cmd = value.Value
		}
	}
	logs.Infof("after cmd:%v", cmd)

	// 创建命令
	CMD := exec.Command(cmd, args...)
	env := os.Environ() //获取当前环境的环境变量
	CMD.Env = env

	//defer outfile.Close()
	CMD.Stdout = os.Stdout
	//CMD.Stdout = outfile
	CMD.Stderr = os.Stderr
	// CMD.Env = append(CMD.Env, )

	CMD.Dir = runtime.Spec.Directory
	// 检查工作目录，如果不存在说明数据出问题了
	if _, err := os.Stat(CMD.Dir); os.IsNotExist(err) {
		logs.Errorf("Directory %s does not exist: %v", CMD.Dir, err)
		return fmt.Errorf("directory %s does not exist: %w", CMD.Dir, err)
	}

	// 启动命令
	logs.Infof("runtime Name:\t %s is Running", runtime.Name)

	if err := CMD.Start(); err != nil {
		//通知group_monitor，来修改全局的group信息（其中的runtime属性）
		cr.notifyRuntimeStartPhase(groupName, groupNamespace, actionSpeName, runtimeSpecName, strconv.Itoa(CMD.Process.Pid), apis.Failed, apis.Time{time.Now()}, apis.Time{time.Now()})
		cr.recorder.Event(runtime, apis.EventTypeWarning, events.FailedToStartCommand, fmt.Sprintf("Runtime Name:\t %s start failed", runtime.Name))
		return fmt.Errorf("failed to start command: %w", err)
	}
	//通知group_monitor，来修改全局的group信息（其中的runtime属性）
	if isInit {
		cr.notifyRuntimeStartPhase(groupName, groupNamespace, actionSpeName, runtimeSpecName, strconv.Itoa(CMD.Process.Pid), apis.Init, apis.Time{time.Now()}, apis.Time{time.Now()})
		cr.recorder.Event(runtime, apis.EventTypeNormal, events.StartedCommand, fmt.Sprintf("Runtime Name:\t %s start to init", runtime.Name))
	} else {
		cr.notifyRuntimeStartPhase(groupName, groupNamespace, actionSpeName, runtimeSpecName, strconv.Itoa(CMD.Process.Pid), apis.Running, apis.Time{time.Now()}, apis.Time{time.Now()})
		cr.recorder.Event(runtime, apis.EventTypeNormal, events.StartedCommand, fmt.Sprintf("Runtime Name:\t %s start to Running", runtime.Name))
	}
	cr.processManager.AddProcess(runtime.Name, CMD)
	cr.stopSignals[runtime.Name] = make(chan struct{})
	logs.Infof("process id:\t %d is Running", CMD.Process.Pid)

	// return nil
	if err := CMD.Wait(); err != nil { // err := CMD.Wait()会阻塞
		// 检查 stopSignal 通道是否被关闭，判断进程是否是外部停止的
		select {
		case <-cr.stopSignals[runtime.Name]: // 如果接收到停止信号
			logs.Info("command killed externally by stopCMD")
			cr.notifyRuntimeEndPhase(groupName, groupNamespace, actionSpeName, runtimeSpecName, apis.Unknown, apis.Time{time.Now()}, apis.Time{time.Now()})
			cr.recorder.Event(runtime, apis.EventTypeNormal, events.KillingCommand, fmt.Sprintf("Runtime Name:\t %s start to close", runtime.Name)) // 发送事件：Runtime收到终止信号进行关闭
			cr.processManager.RemoveProcess(runtime.Name)
			delete(cr.stopSignals, runtime.Name)
			return fmt.Errorf("Receive killed command:\t %s is Stopped", runtime.Name)
		default:
			logs.Errorf("command %s finished with error: %s", runtime.Name, err.Error())
			cr.notifyRuntimeEndPhase(groupName, groupNamespace, actionSpeName, runtimeSpecName, apis.Failed, apis.Time{time.Now()}, apis.Time{time.Now()})
			cr.recorder.Event(runtime, apis.EventTypeWarning, events.ExecuteFailed, fmt.Sprintf("Runtime Name:\t %s execution failure", runtime.Name)) // 发送事件，任务执行失败进行关闭
			cr.processManager.RemoveProcess(runtime.Name)
			delete(cr.stopSignals, runtime.Name)
			return fmt.Errorf("Run failure:\t %s is Failed", runtime.Name)
		}
	}
	logs.Infof("command %s completed", runtime.Name)
	//TODO 正常执行完之后通知修改queues和Manager对应的group信息，group当中Runtime的phase
	cr.processManager.MoveProcessToSucess(runtime.Name) //移入successProcess，同时移出process
	// 修改RuntimeStatus的Phase为Successed，ActionStatus的Phase也为Successed
	cr.notifyRuntimeEndPhase(groupName, groupNamespace, actionSpeName, runtimeSpecName, apis.Successed, apis.Time{time.Now()}, apis.Time{time.Now()})

	// TODO: 使用进程启动CMD, 异步操作
	// TODO: 多个Command拼接---ysk  action包含多个command拼接指？如果一个cmd比较复杂（例如：python predict.py 10 100  data.json）
	//我们如何区分其中的参数和路径（因为执行需要predict.py  data.json的路径，（可以采取固定下载到的数据的路径，采取相对路径的方式？））
	// TODO：注入环境变量---ysk  暂时未确定环境变量的例子
	// TODO: 处理Action的Input和Output---ysk  output的结构后续可能需要调整，目前暂时以文件.txt的输出形式
	return nil
}

// 停止某个CMD对应的进程
// TODO：保存现场
func (cr *CommandRuntime) stopCMD(runtime *apis.Runtime) error {
	CMD, exists := cr.processManager.GetProcess(runtime.Name)
	if !exists {
		logs.Errorf("Failed to find task:\t ", runtime.Name)
		return fmt.Errorf("task '%s' not found", runtime.Name)
	}
	// 终止任务的进程
	if CMD.Process != nil {
		if err := CMD.Process.Kill(); err != nil {
			logs.Errorf("Failed to kill task:\t ", runtime.Name)
			return fmt.Errorf("failed to stop task '%s': %w", runtime.Name, err)
		}
		// 这里不用进行事件的发送，在Monitor监控部分会进行
		fmt.Printf("Task '%s' with PID %d has been stopped.\n", runtime.Name, CMD.Process.Pid)
	} else {
		logs.Infof("Task '%s' is already stopped.", runtime.Name)
		return fmt.Errorf("task '%s' process is nil", runtime.Name)
	}
	return nil

}

// 通过 EventBus 通知 Runtime 状态更新
func (cr *CommandRuntime) notifyRuntimeStartPhase(groupName, groupNamespace string, actionSpeName, runtimeSpecName string, processId string, phase apis.Phase, startAt, lastTime apis.Time) {
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
func (cr *CommandRuntime) notifyRuntimeEndPhase(groupName, groupNamespace string, actionSpeName, runtimeSpecName string, phase apis.Phase, finishTime, lastTime apis.Time) {
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
func (cr *CommandRuntime) StoreData(group *apis.Group, action *apis.Action, runtime *apis.Runtime, actionSpecName, runtimeSpecName string) string {
	// 保存任务状态，调用grpc接口获取任务状态，返回任务状态值即可
	// client, success := cr.clientsManager.GetRuntimeConnection(action.Status.RuntimeStatus[runtimeIndex].RuntimeID)
	var port string
	if runtime.Spec.EnableFineGrainedControlPort != nil {
		port = *runtime.Spec.EnableFineGrainedControlPort
		client := cr.getClient(port)
		if client == nil {
			logs.Info("client is nil")
		}
		index, err := client.RunAppStore()
		if err != nil {
			logs.Errorf("任务保存状态失败: %e", err)
		}
		cr.recorder.Event(action, apis.EventTypeNormal, events.StoredCommand, fmt.Sprintf("Runtime Name:\t %s rpc RunAppStore()", runtime.Name))

		return strconv.FormatInt(index, 10)
	} else {
		logs.Errorf("EnableFineGrainedControlPort not provide, failed")
		return ""
	}
}

// 细粒度控制（grpc）：恢复任务状态
func (cr *CommandRuntime) RestoreData(group *apis.Group, action *apis.Action, runtime *apis.Runtime, actionSpecName, runtimeSpecName string) error {
	// 恢复任务状态，调用grpc接口通知任务恢复任务状态，任务状态存放在etcd当中（group下对应runtime下的runtimeStatus下的keyStatus属性）
	nowTime := apis.Time{time.Now()}
	etcdRuntime, err := cr.clientsManager.GetRuntime(runtime.Name, runtime.Namespace)
	if err != nil {
		logs.Errorf("Failed to get runtime '%s': %v", runtime.Name, err)
	}
	var keyStatus string
	for keyStatus == "" {
		etcdRuntime, err = cr.clientsManager.GetRuntime(runtime.Name, runtime.Namespace)
		if err != nil {
			logs.Errorf("Failed to get runtime '%s': %v", runtime.Name, err)
		}
		keyStatus = etcdRuntime.Status.KeyStatus
	}
	logs.Infof("keyStatus: %s", keyStatus)

	//for {
	//	if cr.client != nil {
	//		break
	//	}
	//	time.Sleep(100 * time.Millisecond)
	//}
	// rpc调用restore()
	var port string
	if runtime.Spec.EnableFineGrainedControlPort != nil {
		port = *runtime.Spec.EnableFineGrainedControlPort
		client := cr.getClient(port)
		if client == nil {
			logs.Info("client is nil")
		}
		_, err = client.RunAppRestore(keyStatus)
		if err != nil {
			logs.Errorf("任务启动失败: %e", err)
		}
		cr.notifyRuntimeStartPhase(group.Name, group.Namespace, actionSpecName, runtimeSpecName, "", apis.Running, nowTime, nowTime)
		cr.recorder.Event(action, apis.EventTypeNormal, events.RestoredCommand, fmt.Sprintf("Runtime Name:\t %s rpc RunAppRestore()", runtime.Name))
	} else {
		logs.Errorf("EnableFineGrainedControlPort not provide, failed")
		err = fmt.Errorf("EnableFineGrainedControlPort not provide")
	}
	return err
}

// 细粒度控制（grpc）：启动任务状态
func (cr *CommandRuntime) StartRuntime(group *apis.Group, action *apis.Action, runtime *apis.Runtime, actionSpecName, runtimeSpecName string) error {
	// 运行任务进程
	go cr.Run(group, action, runtime, actionSpecName, runtimeSpecName) // 这里需要加协程进行启动
	// 初始化rpc客户端
	//if cr.client == nil {
	//	cr.client = grpc_client.NewRuntimeClient(runtime.EnableFineGrainedControlPort, runtime.Name)
	//}
	var port string
	var err error
	if runtime.Spec.EnableFineGrainedControlPort != nil {
		port = *runtime.Spec.EnableFineGrainedControlPort
		client := cr.getClient(port)
		if client == nil {
			logs.Info("client is nil")
		}
		_, err = client.RunAppStart()
		if err != nil {
			logs.Errorf("任务启动失败: %e", err)
		}
	} else {
		logs.Errorf("EnableFineGrainedControlPort not provide, failed")
		err = fmt.Errorf("EnableFineGrainedControlPort not provide")
	}
	return err
}

// 细粒度控制（grpc）：初始化任务
func (cr *CommandRuntime) InitRuntime(group *apis.Group, action *apis.Action, runtime *apis.Runtime, actionSpecName, runtimeSpecName string) error {
	// 1、运行任务进程
	cmd := runtime.Spec.Command
	// Command的执行参数, 所有的参数都需要作为执行参数传入系统
	args := runtime.Spec.Args
	// 目前只接受Command中第一个元素
	go func() {
		err := cr.startCMD(group.Name, group.Namespace, actionSpecName, runtimeSpecName, runtime, cmd[0], args, true)
		if err != nil {
			logs.Infof("Receive -1 :%v", err)
			// TODO: 输出Action的详细信息
		}
	}()
	// TODO: 输出Action的详细信息，等级为Debug
	logs.Infof("Action Name:\t %s is Running", action.Spec.Name)

	// 2、初始化rpc客户端（若无初始化）
	//if cr.client == nil {
	//	cr.client = grpc_client.NewRuntimeClient(runtime.EnableFineGrainedControlPort, "")
	//}
	var port string
	var err error
	if runtime.Spec.EnableFineGrainedControlPort != nil {
		port = *runtime.Spec.EnableFineGrainedControlPort
		client := cr.getClient(port)
		// 3、rpc调用init()
		_, err = client.RunAppInit()
		if err != nil {
			logs.Errorf("任务init失败: %e", err)
		}
	} else {
		logs.Errorf("EnableFineGrainedControlPort not provide, failed")
		err = fmt.Errorf("EnableFineGrainedControlPort not provide")
	}
	//logs.Infof("runtime has Init ====")
	return err
}

// 细粒度控制（grpc）：停止任务
func (cr *CommandRuntime) StopRuntime(group *apis.Group, action *apis.Action, runtime *apis.Runtime, actionSpecName, runtimeSpecName string) error {
	logs.Infof("runtime has stop====")

	//---------停止
	// rpc调用restore()
	var port string
	var err error
	if runtime.Spec.EnableFineGrainedControlPort != nil {
		port = *runtime.Spec.EnableFineGrainedControlPort
		client := cr.getClient(port)
		_, err = client.RunAppStop()
		if err != nil {
			logs.Errorf("任务关闭失败: %e", err)
		}
	} else {
		logs.Errorf("EnableFineGrainedControlPort not provide, failed")
		err = fmt.Errorf("EnableFineGrainedControlPort not provide")
	}
	// ------------
	//cr.notifyRuntimeEndPhase(group.Name, actionIndex, runtimeIndex, apis.Successed, apis.Time{time.Now()}, apis.Time{time.Now()})
	return err
}

// 获取或创建指定端口的Client
func (cr *CommandRuntime) getClient(port string) *grpc_client.RuntimeClient {
	return grpc_client.NewRuntimeClient(port, cr.connectionPool)
}

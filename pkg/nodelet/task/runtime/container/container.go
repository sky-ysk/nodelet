package container

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/docker/docker/api/types/container"
	"math"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	client "github.com/docker/docker/client"
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/client-go/util/manager"
	"hit.edu/framework/pkg/component-base/logs"
	"hit.edu/framework/pkg/nodelet/events"
	"hit.edu/framework/pkg/nodelet/events/eventbus"
	"hit.edu/framework/pkg/nodelet/task/interaction/intwithRuntime/pool"
)

type ContainerRuntime struct {
	client           *client.Client // Docker客户端，用于管理容器
	containerManager *ContainerManager
	// imagePuller *ImagePuller
	// 用于传输状态的适配Runtime运行时的事件总线
	eventBus *eventbus.EventBus
	// 全局系统的事件处理
	//recorder       recorder.EventRecorder
	connectionPool *pool.ConnectionPool
	//client      *grpc_client.RuntimeClient
	stopSignals    map[string]chan struct{} // 用于标记进程是否被外部停止
	clientsManager *manager.Manager
	mu             sync.Mutex // 保护clients和stopSignals
}

func NewContainerRuntime(clientsManager *manager.Manager, eventBus *eventbus.EventBus) *ContainerRuntime {
	client, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		logs.Errorf("Failed to create Docker client: %v", err)
		return nil
	}
	return &ContainerRuntime{
		client:           client,
		containerManager: NewContainerManager(client),
		eventBus:         eventBus,
		connectionPool:   pool.NewConnectionPool(),
		stopSignals:      make(map[string]chan struct{}),
		clientsManager:   clientsManager,
	}
}

func (cr *ContainerRuntime) Run(group *apis.Group, action *apis.Action, runtime *apis.Runtime, actionSpecName, runtimeSpecName string) error {
	// 目前docker的运行有两种方式，一直是纯命令行启动docker，
	// 另一种是指定image并且给出了cmd args。
	// 执行的时候区分方式是前者不要填写"image"字段。后者需要填写。
	if runtime.Spec.Image == "" {
		// 纯命令行启动docker，使用go的CMD包
		logs.Infof("cmd and args way to start: docker runtime for group:%s", group.Name)
		cr.RunCMD(group, action, runtime, actionSpecName, runtimeSpecName)
		return nil
	} else {
		logs.Infof("image way to start: docker runtime for group:%s", group.Name)
		// 这里可以添加容器的创建和启动逻辑
		containerName := runtime.Name
		res, containerId := cr.containerManager.CreateContainer(containerName, runtime)
		if !res {
			logs.Errorf("Failed to create container for group: %s", containerName)
			return nil
		}

		// 启动
		err := cr.containerManager.StartContainer(containerId)
		if !err {
			logs.Errorf("Failed to start container for runtime: %s", runtime.Name)
			return nil
		}
		cr.containerManager.AddRuntimeMapping(runtime.Name, containerId)
		// 监控容器状态
		cr.MonitorContainerStatus(group, action, runtime, actionSpecName, runtimeSpecName, containerId)
		logs.Infof("Runtime: %s, Container %s started successfully", runtime.Name, containerId)
		// 通知事件总线，容器已经拉起running
		cr.notifyRuntimeStartPhase(group.Name, group.Namespace, actionSpecName, runtimeSpecName, containerId, apis.Running, apis.Time{time.Now()}, apis.Time{time.Now()})
		cr.clientsManager.LogEvent(runtime, apis.EventTypeNormal, events.StartedCommand, fmt.Sprintf("Runtime Name:\t %s start to Running", runtime.Name), group.Namespace)

		return nil
	}

}

// CMD方式，要求不能填写image字段。目前只能负责启动容器，不能负责容器的监控
// 因为容器的名字固定在了cmd里，后续可以加入根据名字找到ID来监控和自动删除
func (cr *ContainerRuntime) RunCMD(group *apis.Group, action *apis.Action, runtime *apis.Runtime, actionSpecName, runtimeSpecName string) error {
	logs.Infof("RunCMD for group:%s, action:%s, runtime:%s", group.Name, action.Name, runtime.Name)
	cmd := runtime.Spec.Command[0]
	args := runtime.Spec.Args
	// TODO：根据Spec里面的用户提前填写的Env信息一个个查找args里面的字符串，进行替换
	// spec.env
	// 目前只修改$HOME，使用os.ExpandEnv
	for i, arg := range args {
		args[i] = os.ExpandEnv(arg)
	}
	// 查看修改后的args
	logs.Infof("Modified args: %v.cmd :%s", args, cmd)
	// 创建命令
	CMD := exec.Command(cmd, args...)
	env := os.Environ() //获取当前环境的环境变量
	CMD.Env = env
	CMD.Stdout = os.Stdout
	CMD.Stderr = os.Stderr
	CMD.Dir = runtime.Status.Directory
	if _, err := os.Stat(CMD.Dir); os.IsNotExist(err) {
		logs.Errorf("Directory %s does not exist: %v", CMD.Dir, err)
		return fmt.Errorf("directory %s does not exist: %w", CMD.Dir, err)
	}
	// 启动命令
	logs.Infof("runtime %s 's docker is Running", runtime.Name)

	if err := CMD.Start(); err != nil {
		logs.Errorf("error is %s", err.Error())
		//通知group_monitor，来修改全局的group信息（其中的runtime属性）
		cr.notifyRuntimeStartPhase(group.Name, group.Namespace, action.Spec.Name, runtime.Spec.Name, strconv.Itoa(CMD.Process.Pid), apis.Failed, apis.Time{time.Now()}, apis.Time{time.Now()})
		cr.clientsManager.LogEvent(runtime, apis.EventTypeWarning, events.FailedToStartCommand, fmt.Sprintf("Runtime Name:\t %s start failed", runtime.Name), group.Namespace)
		return fmt.Errorf("failed to start docker command: %w", err)
	}
	logs.Infof("Command %s started successfully with PID %d", cmd, CMD.Process.Pid)
	// 通知group_monitor，来修改全局的group信息（其中的runtime属性）

	// TODO:根据容器名字查找ID，然后放入manager，并通过monitor监管
	// 名字是args其中的一个字符串，包含 --name=
	var containerName string
	for i, str := range runtime.Spec.Args {
		if strings.Contains(str, "--name=") {
			containerName = strings.TrimPrefix(str, "--name=")
			logs.Infof("Found container name: %s", containerName)
			break
		}
		// 或者填的是--name 后面跟一个空格 name
		if strings.Contains(str, "--name") {
			if i+1 < len(runtime.Spec.Args) {
				containerName = runtime.Spec.Args[i+1]
				logs.Infof("Found container name: %s", containerName)
				break
			}
		}
	}

	// 根据name查询ID
	// 可能出现docker拉起速度较慢，查不到====有这个问题，需要等待数秒,设置为500ms
	var containerId string
	var cnt int = 0
	for {
		cnt++
		time.Sleep(500 * time.Millisecond)
		containerId = cr.containerManager.GetContainerID(containerName)
		if containerId != "" {
			logs.Infof("Found container ID: %s for name: %s", containerId, containerName)
		} else {
			logs.Infof("Waiting for container ID for name: %s", containerName)
		}
		containerPid, err := cr.getContainerHostPid(containerName)
		if err != nil {
			logs.Errorf("Failed to get host PID for container %s: %v, maybe need time to start", containerName, err)
		} else {
			logs.Infof("Container %s is running with host PID %d", containerName, containerPid)
		}
		if containerId != "" && containerPid != 0 {
			cr.notifyRuntimeStartPhase(group.Name, group.Namespace, action.Spec.Name, runtime.Spec.Name, strconv.Itoa(containerPid), apis.Running, apis.Time{time.Now()}, apis.Time{time.Now()})
			cr.clientsManager.LogEvent(runtime, apis.EventTypeNormal, events.StartedCommand, fmt.Sprintf("Runtime Name:\t %s start to Running", runtime.Name), group.Namespace)
			break
		}
		if cnt >= 200 {
			logs.Errorf("Timeout waiting for container ID for name: %s", containerName)
			return fmt.Errorf("timeout waiting for container ID for name: %s", containerName)
		}
	}

	logs.Infof("Container ID for runtime %s is %s", runtime.Name, containerId)
	cr.containerManager.AddRuntimeMapping(runtime.Name, containerId)
	cr.MonitorContainerStatus(group, action, runtime, actionSpecName, runtimeSpecName, containerId)
	// 启动容器资源监控 <-- 在这里添加
	go cr.monitorContainerResource(containerId, runtime, 500*time.Millisecond)
	return nil
}

func (cr *ContainerRuntime) getContainerHostPid(containerNameOrID string) (int, error) {
	cmd := exec.Command("docker", "inspect", "--format", "{{.State.Pid}}", containerNameOrID)
	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("failed to run docker inspect: %w", err)
	}

	pidStr := strings.TrimSpace(string(output))
	if pidStr == "0" {
		return 0, fmt.Errorf("container is not running or PID is 0")
	}

	pid, err := strconv.Atoi(pidStr)
	if err != nil {
		return 0, fmt.Errorf("failed to parse PID '%s': %w", pidStr, err)
	}

	return pid, nil
}

// 监控容器状态，如果退出等，判断是否是某些异常，然后调用Kill
func (cr *ContainerRuntime) MonitorContainerStatus(group *apis.Group, action *apis.Action, runtime *apis.Runtime, actionSpecName, runtimeSpecName string, containerId string) {
	go func() {
		for {
			select {
			case <-time.After(10 * time.Second):
				logs.Infof("Checking container status for runtime: %s", runtime.Name)
				container, err := cr.containerManager.GetContainer(containerId)
				if err != nil {
					logs.Errorf("Failed to get container info for %s: %v", containerId, err)
					cr.Kill(group, action, runtime, actionSpecName, runtimeSpecName)
					cr.notifyRuntimeEndPhase(group.Name, group.Namespace, actionSpecName, runtimeSpecName, apis.Failed, apis.Time{time.Now()}, apis.Time{time.Now()})
					return
				}
				// String representation of the container state. Can be one of "created", "running", "paused", "restarting", "removing", "exited", or "dead"
				// TODO:根据更细节的stauts区分，执行不同的操作和通知

				if container.State.Status == "exited" {
					switch container.State.ExitCode {
					case 0:
						logs.Infof("Container %s has exited normally", containerId)
						cr.Kill(group, action, runtime, actionSpecName, runtimeSpecName)
						cr.notifyRuntimeEndPhase(group.Name, group.Namespace, actionSpecName, runtimeSpecName, apis.Successed, apis.Time{time.Now()}, apis.Time{time.Now()})
						return
					default:
						logs.Errorf("container exited abnormally with code %d\n", container.State.ExitCode)
						if container.State.Error != "" {
							logs.Errorf("error message: %s\n", container.State.Error)
							cr.Kill(group, action, runtime, actionSpecName, runtimeSpecName)
							cr.notifyRuntimeEndPhase(group.Name, group.Namespace, actionSpecName, runtimeSpecName, apis.Failed, apis.Time{time.Now()}, apis.Time{time.Now()})
						}
						return
					}

				}

			}
		}
	}()
}

// 监控容器资源使用情况
func (cr *ContainerRuntime) monitorContainerResource(containerId string, runtime *apis.Runtime, interval time.Duration) {
	logs.Infof("Starting resource monitoring for container %s (ID: %s)", runtime.Name, containerId)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// 用于跟踪上一次的统计值
	var lastCPUPercent, lastMemUsage float64
	var updateCount int

	for range ticker.C {
		updateCount++

		// 获取容器统计信息
		stats, err := cr.client.ContainerStats(context.Background(), containerId, false)
		if err != nil {
			logs.Warnf("Failed to get container stats for %s: %v", containerId, err)
			continue
		}

		// 使用通用的JSON解析而不是具体类型
		var containerStats map[string]interface{}
		if err := json.NewDecoder(stats.Body).Decode(&containerStats); err != nil {
			logs.Warnf("Failed to decode container stats for %s: %v", containerId, err)
			stats.Body.Close()
			continue
		}
		stats.Body.Close()

		// 计算CPU使用率
		cpuPercent := cr.calculateContainerCPUPercent(containerStats)

		// 计算内存使用量
		memUsage := cr.calculateContainerMemoryUsage(containerStats)

		// 如果值没有变化，且不是第一次更新，可以跳过更新
		if updateCount > 1 &&
			math.Abs(cpuPercent-lastCPUPercent) < 0.01 &&
			math.Abs(memUsage-lastMemUsage) < 0.01 {
			logs.Tracef("[Monitor] No significant change for container %s, skipping update", runtime.Name)
			continue
		}

		lastCPUPercent = cpuPercent
		lastMemUsage = memUsage

		// 记录详细日志
		logs.Debugf("[Monitor] Container %s (ID: %s) CPU: %.2f%%, Memory: %.2f MB",
			runtime.Name, containerId, cpuPercent, memUsage)

		// 更新runtime状态
		resourceItem := make(map[string]apis.Item)
		resourceItem["cpu"] = apis.Item{
			Name:   "cpu",
			Values: map[string]string{"cpu": fmt.Sprintf("%.2f%%", cpuPercent)},
		}
		resourceItem["memory"] = apis.Item{
			Name:   "memory",
			Values: map[string]string{"memory": fmt.Sprintf("%.2f MB", memUsage)},
		}

		patchRuntime, err := json.Marshal(map[string]interface{}{
			"status": map[string]interface{}{
				"resources": resourceItem,
			},
		})
		if err != nil {
			logs.Errorf("Failed to marshal runtime status: %v", err)
			continue
		}

		_, err = cr.clientsManager.PatchRuntime(runtime.Name, runtime.Namespace, patchRuntime)
		if err != nil {
			logs.Errorf("Failed to patch runtime status: %v", err)
		} else {
			logs.Debugf("[Monitor] Successfully updated resource stats for %s", runtime.Name)
		}
	}
}

// 计算容器CPU使用率（修复版本）
func (cr *ContainerRuntime) calculateContainerCPUPercent(stats map[string]interface{}) float64 {
	// 辅助函数：安全提取嵌套字段的值
	extract := func(data map[string]interface{}, path ...string) interface{} {
		current := data
		for i, key := range path {
			if i == len(path)-1 {
				return current[key]
			}
			if next, ok := current[key].(map[string]interface{}); ok {
				current = next
			} else {
				return nil
			}
		}
		return nil
	}

	// 提取必要的CPU数据
	cpuStats, _ := stats["cpu_stats"].(map[string]interface{})
	precpuStats, _ := stats["precpu_stats"].(map[string]interface{})

	// 获取CPU使用量差值
	totalUsage, _ := extract(cpuStats, "cpu_usage", "total_usage").(float64)
	pretotalUsage, _ := extract(precpuStats, "cpu_usage", "total_usage").(float64)
	cpuDelta := totalUsage - pretotalUsage

	// 获取系统时间差值
	systemUsage, _ := extract(cpuStats, "system_cpu_usage").(float64)
	presystemUsage, _ := extract(precpuStats, "system_cpu_usage").(float64)
	systemDelta := systemUsage - presystemUsage

	// 计算CPU使用率
	if systemDelta > 0 && cpuDelta > 0 {
		// 获取CPU核心数
		cores := 1
		if percpu, ok := extract(cpuStats, "cpu_usage", "percpu_usage").([]interface{}); ok {
			cores = len(percpu)
		}
		return (cpuDelta / systemDelta) * float64(cores) * 100.0
	}

	return 0.0
}

// 计算容器内存使用量
func (cr *ContainerRuntime) calculateContainerMemoryUsage(stats map[string]interface{}) float64 {
	if memoryStats, ok := stats["memory_stats"].(map[string]interface{}); ok {
		// 尝试多个可能的内存使用量字段
		for _, field := range []string{"usage", "max_usage", "rss"} {
			if usage, ok := memoryStats[field].(float64); ok && usage > 0 {
				return usage / (1024 * 1024) // 转换为MB
			}
		}
	}
	return 0.0
}

// kill相当于stop然后再remove
func (cr *ContainerRuntime) Kill(group *apis.Group, action *apis.Action, runtime *apis.Runtime, actionSpecName, runtimeSpecName string) error {
	logs.Infof("container runtime kill group:%s, action:%s, runtime:%s", group.Name, action.Name, runtime.Name)

	// 获取容器ID
	containerID, exists := cr.containerManager.getContainerIDByRuntimeName(runtime.Name)
	if !exists {
		logs.Infof("Container for runtime %s not found or already removed", runtime.Name)
		cr.containerManager.DeleteRuntimeMapping(runtime.Name)
		return nil
	}
	// 记录操作开始时间
	startTime := time.Now()
	var stopErr, removeErr error
	// 1. 先尝试优雅停止容器
	stopErr = cr.stopContainer(containerID, runtime.Name)
	if stopErr != nil {
		logs.Warnf("Failed to stop container %s: %v, attempting removal anyway", containerID, stopErr)
	} else {
		logs.Infof("Container %s stopped successfully", containerID)
	}

	// 2. 删除容器（无论停止是否成功都尝试删除）
	removeErr = cr.removeContainer(containerID, runtime.Name)
	if removeErr != nil {
		// 删除失败，记录错误但继续清理映射
		logs.Errorf("Failed to remove container %s: %v", containerID, removeErr)
	} else {
		logs.Infof("Container %s removed successfully", containerID)
	}
	logs.Infof("Container %s removed successfully", cr.containerManager.manager[runtime.Name])
	// 3. 清理运行时映射（必须执行）
	cr.containerManager.DeleteRuntimeMapping(runtime.Name)
	// 4. 汇总错误信息
	if stopErr != nil && removeErr != nil {
		return fmt.Errorf("both stop and remove failed for container %s: stop_err=%v, remove_err=%v",
			containerID, stopErr, removeErr)
	} else if removeErr != nil {
		return fmt.Errorf("remove failed for container %s: %w", containerID, removeErr)
	} else if stopErr != nil {
		// 停止失败但删除成功，只记录警告
		logs.Warnf("Container %s removed but stop had issues: %v", containerID, stopErr)
	}
	//5. 通知部署器修改Runtime、Action、Group的状态为killed
	cr.notifyRuntimeEndPhase(group.Name, group.Namespace, actionSpecName, runtimeSpecName, apis.Unknown, apis.Time{time.Now()}, apis.Time{time.Now()})
	cr.clientsManager.LogEvent(runtime, apis.EventTypeNormal, events.KilledCommand, fmt.Sprintf("Runtime Name:\t %s start to close", runtime.Name), group.Namespace) // 发送事件：Runtime收到终止信号进行关闭

	logs.Infof("Runtime: %s, Container %s cleanup completed in %v",
		runtime.Name, containerID, time.Since(startTime))
	return nil
}

// 增强的停止容器方法
func (cr *ContainerRuntime) stopContainer(containerID, runtimeName string) error {
	if containerID == "" {
		return fmt.Errorf("empty container ID for runtime %s", runtimeName)
	}

	ctx := context.Background()

	// 先检查容器状态
	containerInfo, err := cr.client.ContainerInspect(ctx, containerID)
	if err != nil {
		if client.IsErrNotFound(err) {
			logs.Infof("Container %s not found, may already be removed", containerID)
			return nil
		}
		return fmt.Errorf("failed to inspect container %s: %w", containerID, err)
	}

	if !containerInfo.State.Running {
		logs.Infof("Container %s is not running (status: %s)", containerID, containerInfo.State.Status)
		return nil
	}

	logs.Infof("Stopping container %s (runtime: %s), current status: %s",
		containerID, runtimeName, containerInfo.State.Status)

	// 设置优雅停止的超时时间（30秒）
	timeout := 30
	stopOptions := container.StopOptions{
		Timeout: &timeout,
	}

	// 尝试优雅停止
	logs.Infof("Attempting graceful stop for container %s with %d seconds timeout", containerID, timeout)
	if err := cr.client.ContainerStop(ctx, containerID, stopOptions); err != nil {
		// 检查特定错误类型
		if client.IsErrNotFound(err) {
			logs.Infof("Container %s not found during stop, may already be removed", containerID)
			return nil
		}
		if strings.Contains(err.Error(), "is not running") {
			logs.Infof("Container %s is already stopped", containerID)
			return nil
		}

		// 优雅停止失败，尝试强制杀死
		logs.Warnf("Graceful stop failed for container %s, attempting force kill: %v", containerID, err)

		if killErr := cr.client.ContainerKill(ctx, containerID, "SIGKILL"); killErr != nil {
			if client.IsErrNotFound(killErr) {
				logs.Infof("Container %s not found during force kill, may already be terminated", containerID)
				return nil
			}
			return fmt.Errorf("both graceful stop and force kill failed: stop_err=%v, kill_err=%v", err, killErr)
		}
		logs.Infof("Container %s force killed successfully", containerID)
	} else {
		logs.Infof("Container %s stopped gracefully", containerID)
	}

	return nil
}

// 增强的删除容器方法
func (cr *ContainerRuntime) removeContainer(containerID, runtimeName string) error {
	if containerID == "" {
		return fmt.Errorf("empty container ID for runtime %s", runtimeName)
	}
	logs.Infof("Removing container %s (runtime: %s)", containerID, runtimeName)

	// 使用ContainerManager的RemoveContainer方法
	if success, err := cr.containerManager.RemoveContainer(containerID); success {
		logs.Infof("Container %s removed successfully via ContainerManager", containerID)
		return nil
	} else {
		logs.Errorf("failed to remove container %s: %w", containerID, err)
		return err
	}
}

// 支持暂时停止，后续可以恢复
func (cr *ContainerRuntime) Stop(group *apis.Group, action *apis.Action, runtime *apis.Runtime, actionSpecName, runtimeSpecName string) error {
	logs.Infof("docker runtime Stop task:%s", group.Name)
	return nil
}
func (cr *ContainerRuntime) Restore(group *apis.Group, action *apis.Action, runtime *apis.Runtime, actionSpecName, runtimeSpecName string) error {
	logs.Infof("docker runtime Restore task:%s", group.Name)
	return nil
}
func (cr *ContainerRuntime) CheckRuntimeStatus(group *apis.Group, action *apis.Action, runtime *apis.Runtime) (string, error) {

	return "", nil
}
func (cr *ContainerRuntime) StoreData(group *apis.Group, action *apis.Action, runtime *apis.Runtime, actionSpecName, runtimeSpecName string) string {

	return ""
}
func (cr *ContainerRuntime) RestoreData(group *apis.Group, action *apis.Action, runtime *apis.Runtime, actionSpecName, runtimeSpecName string) error {
	return nil
}
func (cr *ContainerRuntime) StartRuntime(group *apis.Group, action *apis.Action, runtime *apis.Runtime, actionSpecName, runtimeSpecName string) error {

	return nil
}
func (cr *ContainerRuntime) InitRuntime(group *apis.Group, action *apis.Action, runtime *apis.Runtime, actionSpecName, runtimeSpecName string) error {
	return nil
}
func (cr *ContainerRuntime) StopRuntime(group *apis.Group, action *apis.Action, runtime *apis.Runtime, actionSpecName, runtimeSpecName string) error {
	return nil
}

// 通过 EventBus 通知 Runtime 状态更新
func (cr *ContainerRuntime) notifyRuntimeStartPhase(groupName, groupNamespace string, actionSpeName, runtimeSpecName string, processId string, phase apis.Phase, startAt, lastTime apis.Time) {
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
func (cr *ContainerRuntime) notifyRuntimeEndPhase(groupName, groupNamespace string, actionSpeName, runtimeSpecName string, phase apis.Phase, finishTime, lastTime apis.Time) {
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

// // TODO:event需要增加start和end以外的update接口，用于上传除了start和end以外的状态
// func (cr *ContainerRuntime) updateContainerPhase(groupName, groupNamespace string, actionSpeName, runtimeSpecName string, phase apis.Phase, finishTime, lastTime apis.Time) {
// 	event := events.RuntimeEndPhaseEvent1{
// 		GroupName:       groupName,
// 		GroupNamespace:  groupNamespace,
// 		ActionSpecName:  actionSpeName,
// 		RuntimeSpecName: runtimeSpecName,
// 		Phase:           phase,
// 		FinishAt:        finishTime,
// 		LastTime:        lastTime,
// 	}
// 	cr.eventBus.Publish(event)
// }

// 解析dokcer命令行参数给到condig和hostConfig
// 示例命令为：
// 这个方法暂时不使用，命令的格式未统一，无法解析
func parseDockerCommandArgs(args []string) (string, []string) {
	if len(args) == 0 {
		return "", nil
	}
	// 第一个参数是命令，后面的都是参数
	command := args[0]
	var commandArgs []string
	if len(args) > 1 {
		commandArgs = args[1:]
	}
	return command, commandArgs
}

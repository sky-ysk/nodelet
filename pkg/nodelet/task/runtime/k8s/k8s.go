package k8s

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	apis "hit.edu/framework/pkg/apis/cores"
	me "hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/client-go/util/manager"
	"hit.edu/framework/pkg/component-base/logs"
	"hit.edu/framework/pkg/nodelet/events"
	"hit.edu/framework/pkg/nodelet/events/eventbus"
	grpc_client "hit.edu/framework/pkg/nodelet/task/interaction/intwithRuntime/grpc-client"
	"hit.edu/framework/pkg/nodelet/task/interaction/intwithRuntime/pool"
	"hit.edu/framework/pkg/nodelet/task/runtime/k8s/config"
	"hit.edu/framework/pkg/nodelet/task/runtime/k8s/entity"
	"hit.edu/framework/pkg/nodelet/task/runtime/k8s/monitor"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	metricsclientset "k8s.io/metrics/pkg/client/clientset/versioned"
)

type K8sRuntime struct {
	clientset     *kubernetes.Clientset
	metricsClient *metricsclientset.Clientset //go get k8s.io/metrics/pkg/client/clientset/versioned 从 Metrics Server 获取的实时监控数据
	//client         *grpc_client.RuntimeClient
	connectionPool *pool.ConnectionPool
	clientsManager *manager.Manager
	// 全局事件发送的组件
	//recorder recorder.EventRecorder
	monitor  *monitor.Monitor
	eventBus *eventbus.EventBus
}

// 资源占用量
type ResourceUsage struct {
	cpuUsage    float64
	memoryUsage int64
}

// Pod状态，包含Pod的执行状态和Pod的资源占用量
type PodStatus struct {
	PodState       string
	Resource_usage ResourceUsage
}

func NewK8sRuntime(clientsManager *manager.Manager, eventBus *eventbus.EventBus, pool *pool.ConnectionPool, nodeName string) *K8sRuntime {
	clientset := config.LoadConfig()
	metricsClient := config.LoadMcConfig()
	if clientset == nil || metricsClient == nil {
		logs.Error("clientset or metricsClient is nil---")
		return nil
	}
	// 此处需要根据nodeName转换为hostName
	node, err2 := clientsManager.GetNodeClient("test").Client.Get(context.TODO(), nodeName, me.GetOptions{})
	if err2 != nil {
		logs.Errorf("Get node info error1")
	}
	k8sMonitor := monitor.NewMonitor(clientset, eventBus, node.Spec.HostName, clientsManager)

	k8sMonitor.Start()
	return &K8sRuntime{clientset: clientset, metricsClient: metricsClient, connectionPool: pool, monitor: k8sMonitor, eventBus: eventBus, clientsManager: clientsManager}
}
func (k *K8sRuntime) Kill(group *apis.Group, action *apis.Action, runtime *apis.Runtime, actionSpecName, runtimeSpecName string) error {
	logs.Infof("k8s runtime kill runtime: %s", runtime.Name)
	yamlFilePatch := runtime.Spec.Inputs[0].From
	objList, err := entity.ParseK8sResourcesFromFile(yamlFilePatch, *group.Status.Node)
	if err != nil {
		logs.Errorf("Get k8s resources from yaml file failed: %v", err)
	}
	err = DeleteResources(k.clientset, objList)
	if err != nil {
		logs.Errorf("Delete k8s resources from yaml file failed: %v", err)
	}
	return nil
}
func (k *K8sRuntime) Stop(group *apis.Group, action *apis.Action, runtime *apis.Runtime, actionSpecName, runtimeSpecName string) error {

	return nil
}
func (k *K8sRuntime) Restore(group *apis.Group, action *apis.Action, runtime *apis.Runtime, actionSpecName, runtimeSpecName string) error {

	return nil
}

// 粗粒度管理的启动方法
func (k *K8sRuntime) Run(group *apis.Group, action *apis.Action, runtime *apis.Runtime, actionSpecName, runtimeSpecName string) error {
	logs.Infof("k8s runtime for task: %s", group.Name)
	//先执行共同的操作,再各自调用代码
	// 1、首先读取yaml文件，转换为资源
	yamlFilePatch := runtime.Spec.Inputs[0].From
	logs.Infof("group.Status.Node:%v", *group.Status.Node)
	// 根据这个NodeName找到主机名
	get, err2 := k.clientsManager.GetNodeClient("test").Client.Get(context.TODO(), *group.Status.Node, me.GetOptions{})
	if err2 != nil {
		logs.Errorf("Get node info error")
	}
	objList, err := entity.ParseK8sResourcesFromFile(yamlFilePatch, get.Spec.HostName)
	if err != nil {
		logs.Errorf("Get k8s resources from yaml file failed: %v", err)
	}
	for _, obj := range objList {
		objName, objNamespace := getObjectMeta(obj)
		k.monitor.SetState(group, objName, objNamespace, runtime, actionSpecName, runtimeSpecName)
		err := CreateResource(k.clientset, obj)
		if err != nil {
			logs.Errorf("Create k8s resource failed: %v", err)
			return err
		}
	}
	// MonitorPodTimestamp
	// go k.MonitorPodTimestamp(group, podName, namespace)
	go k.MonitorPodTimestamp(group, "", "")

	return nil
}

func (k *K8sRuntime) MonitorPodTimestamp(group *apis.Group, podName string, namespace string) {
	logs.Info("MonitorPodTimestamp Start=======")
	// 配置参数
	if podName == "" {
		podName = "grpc-client-pod-copy"
	}
	if namespace == "" {
		namespace = "switch"
	}
	// logs.Infof("开始监控 Pod %s 的时间戳，命名空间: %s\n", podName, namespace)
	const (
		retryInterval = 1 // Pod 不存在时的重试间隔（秒）
		// logLineMatch  = "Starting server"
		logLineMatch = "restore status successfully"
	)
	cnt := 0
	for {
		// 如果groupName不包含"-copy"子串，则直接返回
		if !strings.Contains(group.Name, "-copy") {
			return
		}

		time.Sleep(time.Duration(retryInterval) * time.Second)
		cnt += 1
		if cnt >= 50 {
			logs.Errorf("monitor times >= 50, can not find the pod%v", podName)
			return
		}
		if !podExists(podName, namespace) {
			logs.Warnf("Pod %s 不存在，等待 %d 秒后重试...\n", podName, retryInterval)
			continue
		}
		// 捕获日志流
		cmd := exec.Command("kubectl", "logs", "-f", podName, "-n", namespace, "--since=0s")
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			logs.Errorf("pod timestamp monitor创建管道失败: %v\n", err)
			return
		}

		if err := cmd.Start(); err != nil {
			logs.Errorf("pod timestamp monitor启动日志捕获失败: %v\n", err)
			return
		}
		// 读取日志流
		reader := bufio.NewReader(stdout)
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				// 日志流中断(暂时不重新获取Logs)
				logs.Errorf("pod timestamp monitor日志流读取失败: %v\n", err)
				cmd.Process.Kill()
				//return
			}

			// 匹配目标日志行
			if strings.Contains(line, logLineMatch) {
				// 提取时间戳（假设时间戳是日志行的前两个字段，格式为 YYYY/MM/DD HH:MM:SS.MICROSECONDS）
				timestampStr := extractTimestamp(line)
				if timestampStr != "" {
					// logs.Infof("已记录事件：%s\n", timestampStr)
					// 上传到etcd
					// 定义时间格式
					layout := "2006/01/02 15:04:05.000000"

					// 解析时间戳字符串
					timestamp, err := time.Parse(layout, timestampStr)
					TimeStamp := apis.Time{timestamp}
					if err != nil {
						logs.Errorf("解析时间戳失败:", err)
						return
					}
					patchGroup, err := json.Marshal(map[string]interface{}{
						"status": map[string]interface{}{
							"serviceRestoreTime": &TimeStamp,
						},
					})
					_, err = k.clientsManager.PatchGroup(group.Name, group.Namespace, patchGroup)
					if err != nil {
						logs.Errorf("Patch group err%v", err)
						return
					}
					logs.Infof("已记录服务重启事件的时间：%s\n", timestampStr)
					return
				}
			}
		}
	}
}

// 检查 Pod 是否存在
func podExists(podName, namespace string) bool {
	cmd := exec.Command("kubectl", "get", "pod", podName, "-n", namespace)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	return err == nil
}

// 提取时间戳（时间戳是日志行的前两个字段）
func extractTimestamp(line string) string {
	parts := strings.Fields(line)
	if len(parts) >= 2 {
		return parts[0] + " " + parts[1]
	}
	return ""
}

// 通用方法：获取资源的 Name 和 Namespace
func getObjectMeta(obj runtime.Object) (name, namespace string) {
	switch t := obj.(type) {
	case metav1.Object: // 所有 Kubernetes 资源都实现了 metav1.Object 接口
		return t.GetName(), t.GetNamespace()
	default:
		logs.Errorf("Resource is not a k8s object")
		return "", ""
	}
}

func (k *K8sRuntime) CheckRuntimeStatus(group *apis.Group, action *apis.Action, runtime *apis.Runtime) (string, error) {

	return "", nil
}
func (k *K8sRuntime) StoreData(group *apis.Group, action *apis.Action, runtime *apis.Runtime, actionSpecName, runtimeSpecName string) string {
	// 保存任务状态，调用grpc接口获取任务状态，返回任务状态值即可
	if runtime.Spec.EnableFineGrainedControlService == nil || runtime.Spec.EnableFineGrainedControlPort == nil {
		logs.Errorf("Need input EnableFineGrainedControlService and EnableFineGrainedControlPort, now all is nil")
		// 发送失败事件
		k.clientsManager.LogEvent(action, apis.EventTypeWarning, events.FailedToStartCommand, fmt.Sprintf("Runtime Name:\t %s failed to use rpc RunAppStore()", runtime.Name), group.Namespace)
		// 同时应该发送失败的Notify
		k.notifyRuntimeEndPhase(group.Name, group.Namespace, action.Spec.Name, runtime.Spec.Name, apis.Failed, apis.Time{time.Now()}, apis.Time{time.Now()})
	}
	client := k.getClient(*runtime.Spec.EnableFineGrainedControlService, *runtime.Spec.EnableFineGrainedControlPort)
	// rpc调用store()
	_, err := client.RunAppStore()
	if err != nil {
		logs.Error(err, "Store application status failed")
	}
	k.clientsManager.LogEvent(action, apis.EventTypeNormal, events.StoredCommand, fmt.Sprintf("Runtime Name:\t %s rpc RunAppStore()", runtime.Name), group.Namespace)
	return "aass"
}
func (k *K8sRuntime) RestoreData(group *apis.Group, action *apis.Action, runtime *apis.Runtime, actionSpecName, runtimeSpecName string) error {
	keyStatus := ""
	if runtime.Spec.EnableFineGrainedControlService == nil || runtime.Spec.EnableFineGrainedControlPort == nil {
		logs.Errorf("Need input EnableFineGrainedControlService and EnableFineGrainedControlPort, now all is nil in RestoreData")
		// 发送失败事件
		k.clientsManager.LogEvent(action, apis.EventTypeWarning, events.FailedToStartCommand, fmt.Sprintf("Runtime Name:\t %s failed to use rpc RestoreData()", runtime.Name), group.Namespace)
		// 同时应该发送失败的Notify
		k.notifyRuntimeEndPhase(group.Name, group.Namespace, action.Spec.Name, runtime.Spec.Name, apis.Failed, apis.Time{time.Now()}, apis.Time{time.Now()})
	}
	go func() {
		nowtime := apis.Time{time.Now()}
		patchGroup, _ := json.Marshal(map[string]interface{}{
			"status": map[string]interface{}{
				"restoreTime": nowtime,
			},
		})
		_, err := k.clientsManager.PatchGroup(group.Name, group.Namespace, patchGroup)
		if err != nil {
			logs.Errorf("Patch group err101:%v", err)
		}
	}()
	client := k.getClient(*runtime.Spec.EnableFineGrainedControlService, *runtime.Spec.EnableFineGrainedControlPort)
	_, err := client.RunAppRestore(keyStatus)
	if err != nil {
		logs.Errorf("Restore application status failed")
	}
	k.clientsManager.LogEvent(action, apis.EventTypeNormal, events.RestoredCommand, fmt.Sprintf("Runtime Name:\t %s rpc RunAppRestore()", runtime.Name), group.Namespace)
	return err
}
func (k *K8sRuntime) StartRuntime(group *apis.Group, action *apis.Action, runtime *apis.Runtime, actionSpecName, runtimeSpecName string) error {
	err := k.Run(group, action, runtime, actionSpecName, runtimeSpecName)
	if err != nil {
		logs.Errorf("StartRuntime failed: %v", err)
		return err
	}
	if runtime.Spec.EnableFineGrainedControlService == nil || runtime.Spec.EnableFineGrainedControlPort == nil {
		logs.Errorf("Need input EnableFineGrainedControlService and EnableFineGrainedControlPort, now all is nil in StartRuntime")
		// 发送失败事件
		k.clientsManager.LogEvent(action, apis.EventTypeWarning, events.FailedToStartCommand, fmt.Sprintf("Runtime Name:\t %s failed to use rpc StartRuntime()", runtime.Name), group.Namespace)
		// 同时应该发送失败的Notify
		k.notifyRuntimeEndPhase(group.Name, group.Namespace, action.Spec.Name, runtime.Spec.Name, apis.Failed, apis.Time{time.Now()}, apis.Time{time.Now()})
	}
	logs.Infof("[Start]开启grpc客户端连接pod当中的grpc服务端，ip：%v,端口：%v", *runtime.Spec.EnableFineGrainedControlService, *runtime.Spec.EnableFineGrainedControlPort)
	client := k.getClient(*runtime.Spec.EnableFineGrainedControlService, *runtime.Spec.EnableFineGrainedControlPort)
	_, err = client.RunAppStart()
	if err != nil {
		logs.Errorf("Start application failed")
	}
	return err
}
func (k *K8sRuntime) InitRuntime(group *apis.Group, action *apis.Action, runtime *apis.Runtime, actionSpecName, runtimeSpecName string) error {
	err := k.Run(group, action, runtime, actionSpecName, runtimeSpecName)
	if err != nil {
		logs.Errorf("InitRuntime failed: %v", err)
		return err
	}
	if runtime.Spec.EnableFineGrainedControlService == nil || runtime.Spec.EnableFineGrainedControlPort == nil {
		logs.Errorf("Need input EnableFineGrainedControlService and EnableFineGrainedControlPort, now all is nil in InitRuntime")
		// 发送失败事件
		k.clientsManager.LogEvent(action, apis.EventTypeWarning, events.FailedToStartCommand, fmt.Sprintf("Runtime Name:\t %s failed to use rpc InitRuntime()", runtime.Name), group.Namespace)
		// 同时应该发送失败的Notify
		k.notifyRuntimeEndPhase(group.Name, group.Namespace, action.Spec.Name, runtime.Spec.Name, apis.Failed, apis.Time{time.Now()}, apis.Time{time.Now()})
	}
	logs.Infof("[Init]开启grpc客户端连接pod当中的grpc服务端，ip：%v,端口：%v", *runtime.Spec.EnableFineGrainedControlService, *runtime.Spec.EnableFineGrainedControlPort)
	client := k.getClient(*runtime.Spec.EnableFineGrainedControlService, *runtime.Spec.EnableFineGrainedControlPort)
	_, err = client.RunAppInit()
	if err != nil {
		logs.Errorf("Init application failed")
	}
	return err
}
func (k *K8sRuntime) StopRuntime(group *apis.Group, action *apis.Action, runtime *apis.Runtime, actionSpecName, runtimeSpecName string) error {
	if runtime.Spec.EnableFineGrainedControlService == nil || runtime.Spec.EnableFineGrainedControlPort == nil {
		logs.Errorf("Need input EnableFineGrainedControlService and EnableFineGrainedControlPort, now all is nil in StopRuntime")
		// 发送失败事件
		k.clientsManager.LogEvent(action, apis.EventTypeWarning, events.FailedToStartCommand, fmt.Sprintf("Runtime Name:\t %s failed to use rpc StopRuntime()", runtime.Name), group.Namespace)
		// 同时应该发送失败的Notify
		k.notifyRuntimeEndPhase(group.Name, group.Namespace, action.Spec.Name, runtime.Spec.Name, apis.Failed, apis.Time{time.Now()}, apis.Time{time.Now()})
	}
	client := k.getClient(*runtime.Spec.EnableFineGrainedControlService, *runtime.Spec.EnableFineGrainedControlPort)
	_, err := client.RunAppStop()
	if err != nil {
		logs.Errorf("Stop application failed")
	}
	return err
}
func (k *K8sRuntime) getClient(service, port string) *grpc_client.RuntimeClient {
	return grpc_client.NewK8sRuntimeClient(service, port, k.connectionPool)
}

// 通过 EventBus 通知 Runtime 状态更新
func (k *K8sRuntime) notifyRuntimeStartPhase(groupName, groupNamespace string, actionSpecName, runtimeSpecName string, processId string, phase apis.Phase, startAt, lastTime apis.Time) {
	event := events.RuntimeStartPhaseEvent1{
		GroupName:       groupName,
		GroupNamespace:  groupNamespace,
		ActionSpecName:  actionSpecName,
		RuntimeSpecName: runtimeSpecName,
		ProcessId:       processId,
		Phase:           phase,
		StartAt:         startAt,
		LastTime:        lastTime,
	}
	k.eventBus.Publish(event)
}
func (k *K8sRuntime) notifyRuntimeEndPhase(groupName, groupNamespace string, actionSpecName, runtimeSpecName string, phase apis.Phase, finishTime, lastTime apis.Time) {
	event := events.RuntimeEndPhaseEvent1{
		GroupName:       groupName,
		GroupNamespace:  groupNamespace,
		ActionSpecName:  actionSpecName,
		RuntimeSpecName: runtimeSpecName,
		Phase:           phase,
		FinishAt:        finishTime,
		LastTime:        lastTime,
	}
	k.eventBus.Publish(event)
}

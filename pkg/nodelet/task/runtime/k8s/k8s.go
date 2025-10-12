package k8s

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"strconv"
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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
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
	monitor   *monitor.Monitor
	eventBus  *eventbus.EventBus
	randomNum map[string]int32
	// 添加node属性--node_exporter
	node *apis.Node
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
	return &K8sRuntime{clientset: clientset, metricsClient: metricsClient, connectionPool: pool, monitor: k8sMonitor, eventBus: eventBus, clientsManager: clientsManager, randomNum: make(map[string]int32), node: node}
}
func (k *K8sRuntime) Kill(group *apis.Group, action *apis.Action, runtime *apis.Runtime, actionSpecName, runtimeSpecName string) error {
	logs.Infof("k8s runtime kill runtime: %s", runtime.Name)
	yamlFilePatch := runtime.Spec.Inputs[0].From
	objList, err := entity.ParseK8sResourcesFromFile(yamlFilePatch, k.node.Spec.HostName, k.randomNum[group.Name])
	if err != nil {
		logs.Errorf("Get k8s resources from yaml file failed: %v", err)
	}
	err = DeleteResources(k.clientset, objList)
	if err != nil {
		logs.Errorf("Delete k8s resources from yaml file failed: %v", err)
	}
	if _, exists := k.randomNum[group.Name]; exists {
		delete(k.randomNum, group.Name)
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
	task, err2 := k.clientsManager.GetTask(group.Labels["belong"], group.Namespace)
	if err2 != nil {
		logs.Errorf("Get Task error")
	}
	randomNum := task.Spec.RandomNum
	k.randomNum[group.Name] = randomNum
	logs.Infof("【【【Run】】】====group.name:%v,randomNum:%v", group.Name, k.randomNum[group.Name])
	logs.Infof("k8s runtime for task: %s", group.Name)
	//先执行共同的操作,再各自调用代码
	// 1、首先读取yaml文件，转换为资源
	// yamlFilePatch := runtime.Spec.Directory + "/" + runtime.Spec.Data[0].Name
	yamlFilePatch := runtime.Spec.Inputs[0].From
	logs.Infof("group.Status.Node:%v", *group.Status.Node)
	// 根据这个NodeName找到主机名
	//get, err2 := k.clientsManager.GetNodeClient("test").Client.Get(context.TODO(), *group.Status.Node, me.GetOptions{})
	if err2 != nil {
		logs.Errorf("Get node info error")
	}
	// 将任务的yaml转为k8s的数据类型
	objList, err := entity.ParseK8sResourcesFromFile(yamlFilePatch, k.node.Spec.HostName, randomNum)
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
		// MonitorPodResources
		//go k.monitorPodResources(objName, objNamespace, runtime, 1*time.Second)
	}
	// MonitorPodTimestamp
	// go k.MonitorPodTimestamp(group, podName, namespace)
	go k.monitorPodTimestamp(group, "", "", randomNum)

	return nil
}

// 监控某一个pod的restore的时间戳并上传
func (k *K8sRuntime) monitorPodTimestamp(group *apis.Group, podName string, namespace string, randomNum int32) {
	logs.Info("MonitorPodTimestamp Start=======")
	// 配置参数
	if podName == "" {
		numStr := strconv.Itoa(int(randomNum))
		podName = "grpc-client-pod-copy-" + numStr
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
		if cnt >= 60*60*10 {
			logs.Errorf("monitor times >= 10h, can not find the pod%v", podName)
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
			if err == io.EOF {
				// 日志流中断(暂时不重新获取Logs)
				logs.Info("Log stream is ni")
				return
			}
			time.Sleep(1 * time.Second)
			if !podExists(podName, namespace) {
				logs.Warnf("Pod %s 已关闭", podName)
				return
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

// 新增的Pod监控方法
func (k *K8sRuntime) monitorPodResources(podName, namespace string, runtime *apis.Runtime, interval time.Duration) {
	logs.Infof("Starting resource monitoring for pod %s in namespace %s", podName, namespace)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		// 获取Pod资源使用情况
		podMetrics, err := k.metricsClient.MetricsV1beta1().PodMetricses(namespace).Get(
			context.TODO(), podName, metav1.GetOptions{})

		if err != nil {
			//logs.Warnf("Failed to get metrics for pod %s: %v", podName, err)
			continue
		}

		var (
			cpuTotal resource.Quantity
			memTotal resource.Quantity
		)

		for _, container := range podMetrics.Containers {
			cpuTotal.Add(container.Usage[corev1.ResourceCPU])
			memTotal.Add(container.Usage[corev1.ResourceMemory])
		}
		// 转换并记录
		cpuMilli := cpuTotal.MilliValue()
		memMB := float64(memTotal.Value()) / (1024 * 1024)
		// // 获取 CPU 核心数,计算总占用的百分比形式需要使用
		// coreCount, err := cpu.Counts(true)
		// if err != nil {
		// 	logs.Fatalf("Failed to get CPU core count: %v", err)
		// }
		logs.Infof("[Monitor] Runtime %s (Pod: %s) - CPU: %dm, Memory: %.2f MB",
			runtime.Name, podName, cpuMilli, memMB)

		//上传etcd
		resourceItem := make(map[string]apis.Item)
		// 填充Values这个字段，这是一个map
		resourceItem["cpu"] = apis.Item{Name: "cpu", Values: map[string]string{"cpu": fmt.Sprintf("%d m", cpuMilli)}}
		resourceItem["memory"] = apis.Item{Name: "memory", Values: map[string]string{"memory": fmt.Sprintf("%.2f MB", float64(memMB))}}
		patchRuntime, err := json.Marshal(map[string]interface{}{
			"status": map[string]interface{}{
				"resources": resourceItem,
			},
		})
		_, err = k.clientsManager.PatchRuntime(runtime.Name, runtime.Namespace, patchRuntime)
		if err != nil {
			logs.Errorf("patch runtimeStatus error")
		}
	}
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
	logs.Infof("【【【StoreData】】】====group.name:%v,randomNum:%v", group.Name, k.randomNum[group.Name])
	randomNum := k.randomNum[group.Name]
	var port string
	if strings.Contains(runtime.Spec.Inputs[0].From, "grpc-client-pod-copy") {
		port = strconv.Itoa(int(randomNum + 2))
	} else if strings.Contains(runtime.Spec.Inputs[0].From, "grpc-client-pod") {
		port = strconv.Itoa(int(randomNum + 1))
	} else if strings.Contains(runtime.Spec.Inputs[0].From, "grpc-server-pod") {
		port = strconv.Itoa(int(randomNum))
	} else {
		port = *runtime.Spec.EnableFineGrainedControlPort
	}
	logs.Infof("[StoreData]开启grpc客户端连接pod当中的grpc服务端，ip：%v,端口：%v", *runtime.Spec.EnableFineGrainedControlService, port)
	logs.Infof("[StoreData-1]开启grpc客户端连接pod当中的grpc服务端，ip：%v,端口：%v", k.node.Spec.HostIp, port)
	//client := k.getClient(*runtime.Spec.EnableFineGrainedControlService, port)
	client := k.getClient(k.node.Spec.HostIp, port)
	// rpc调用store()
	index, err := client.RunAppStore()
	if err != nil {
		logs.Error(err, "Store application status failed")
	}
	k.clientsManager.LogEvent(action, apis.EventTypeNormal, events.StoredCommand, fmt.Sprintf("Runtime Name:\t %s rpc RunAppStore()", runtime.Name), group.Namespace)
	return index
}
func (k *K8sRuntime) RestoreData(group *apis.Group, action *apis.Action, runtime *apis.Runtime, actionSpecName, runtimeSpecName string) error {
	logs.Infof("【【【RestoreData】】】====group.name:%v,randomNum:%v", group.Name, k.randomNum[group.Name])
	keyStatus := ""
	etcdRuntime, err := k.clientsManager.GetRuntime(runtime.Name, runtime.Namespace)
	if err != nil {
		logs.Errorf("Failed to get runtime '%s': %v", runtime.Name, err)
	}
	for keyStatus == "" {
		etcdRuntime, err = k.clientsManager.GetRuntime(runtime.Name, runtime.Namespace)
		if err != nil {
			logs.Errorf("Failed to get runtime '%s': %v", runtime.Name, err)
		}
		keyStatus = etcdRuntime.Status.KeyStatus
		logs.Trace("===================try")
	}
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
	randomNum := k.randomNum[group.Name]
	var port string
	if strings.Contains(runtime.Spec.Inputs[0].From, "grpc-client-pod-copy") {
		port = strconv.Itoa(int(randomNum + 2))
	} else if strings.Contains(runtime.Spec.Inputs[0].From, "grpc-client-pod") {
		port = strconv.Itoa(int(randomNum + 1))
	} else if strings.Contains(runtime.Spec.Inputs[0].From, "grpc-server-pod") {
		port = strconv.Itoa(int(randomNum))
	} else {
		port = *runtime.Spec.EnableFineGrainedControlPort
	}
	logs.Infof("[RestoreData]开启grpc客户端连接pod当中的grpc服务端，ip：%v,端口：%v", *runtime.Spec.EnableFineGrainedControlService, port)
	logs.Infof("[RestoreData-1]开启grpc客户端连接pod当中的grpc服务端，ip：%v,端口：%v", k.node.Spec.HostIp, port)
	//client := k.getClient(*runtime.Spec.EnableFineGrainedControlService, port)
	client := k.getClient(k.node.Spec.HostIp, port)
	_, err = client.RunAppRestore(keyStatus)
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
	var port string
	randomNum := k.randomNum[group.Name]
	if strings.Contains(runtime.Spec.Inputs[0].From, "grpc-client-pod-copy") {
		port = strconv.Itoa(int(randomNum + 2))
	} else if strings.Contains(runtime.Spec.Inputs[0].From, "grpc-client-pod") {
		port = strconv.Itoa(int(randomNum + 1))
	} else if strings.Contains(runtime.Spec.Inputs[0].From, "grpc-server-pod") {
		port = strconv.Itoa(int(randomNum))
	} else {
		port = *runtime.Spec.EnableFineGrainedControlPort
	}
	logs.Infof("[StartRuntime]开启grpc客户端连接pod当中的grpc服务端，ip：%v,端口：%v", *runtime.Spec.EnableFineGrainedControlService, port)
	logs.Infof("[StartRuntime-1]开启grpc客户端连接pod当中的grpc服务端，ip：%v,端口：%v", k.node.Spec.HostIp, port)
	//client := k.getClient(*runtime.Spec.EnableFineGrainedControlService, port)
	client := k.getClient(k.node.Spec.HostIp, port)
	_, err = client.RunAppStart()
	if err != nil {
		logs.Errorf("Start application failed")
	}
	return err
}
func (k *K8sRuntime) InitRuntime(group *apis.Group, action *apis.Action, runtime *apis.Runtime, actionSpecName, runtimeSpecName string) error {
	//task, err2 := k.clientsManager.GetTask(group.Labels["belong"], group.Namespace)
	//if err2 != nil {
	//	logs.Errorf("Get Task error")
	//}
	//randomNum := task.Spec.RandomNum
	//k.randomNum[group.Name] = randomNum
	//logs.Infof("【【【InitRuntime】】】====group.name:%v,randomNum:%v", group.Name, k.randomNum[group.Name])
	logs.Info("++++++++++++++++++++++++++++++++++++k8s的init方法----------------------------------")
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
	var port string
	randomNum := k.randomNum[group.Name]
	if strings.Contains(runtime.Spec.Inputs[0].From, "grpc-client-pod-copy") {
		port = strconv.Itoa(int(randomNum + 2))
	} else if strings.Contains(runtime.Spec.Inputs[0].From, "grpc-client-pod") {
		port = strconv.Itoa(int(randomNum + 1))
	} else if strings.Contains(runtime.Spec.Inputs[0].From, "grpc-server-pod") {
		port = strconv.Itoa(int(randomNum))
	} else {
		port = *runtime.Spec.EnableFineGrainedControlPort
	}
	logs.Infof("[Init]开启grpc客户端连接pod当中的grpc服务端，ip：%v,端口：%v", *runtime.Spec.EnableFineGrainedControlService, port)
	logs.Infof("[Init-1]开启grpc客户端连接pod当中的grpc服务端，ip：%v,端口：%v", k.node.Spec.HostIp, port)
	//client := k.getClient(*runtime.Spec.EnableFineGrainedControlService, port)
	client := k.getClient(k.node.Spec.HostIp, port)
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
	randomNum := k.randomNum[group.Name]
	var port string
	if strings.Contains(runtime.Spec.Inputs[0].From, "grpc-client-pod-copy") {
		port = strconv.Itoa(int(randomNum + 2))
	} else if strings.Contains(runtime.Spec.Inputs[0].From, "grpc-client-pod") {
		port = strconv.Itoa(int(randomNum + 1))
	} else if strings.Contains(runtime.Spec.Inputs[0].From, "grpc-server-pod") {
		port = strconv.Itoa(int(randomNum))
	} else {
		port = *runtime.Spec.EnableFineGrainedControlPort
	}
	logs.Infof("[StopRuntime]开启grpc客户端连接pod当中的grpc服务端，ip：%v,端口：%v", *runtime.Spec.EnableFineGrainedControlService, port)
	logs.Infof("[StopRuntime-1]开启grpc客户端连接pod当中的grpc服务端，ip：%v,端口：%v", k.node.Spec.HostIp, port)
	//client := k.getClient(*runtime.Spec.EnableFineGrainedControlService, port)
	client := k.getClient(k.node.Spec.HostIp, port)
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

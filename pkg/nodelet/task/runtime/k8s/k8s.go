package k8s

import (
	"fmt"
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/client-go/tools/recorder"
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
	"time"
)

type K8sRuntime struct {
	clientset     *kubernetes.Clientset
	metricsClient *metricsclientset.Clientset //go get k8s.io/metrics/pkg/client/clientset/versioned 从 Metrics Server 获取的实时监控数据
	//client         *grpc_client.RuntimeClient
	connectionPool *pool.ConnectionPool
	// 全局事件发送的组件
	recorder recorder.EventRecorder
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

func NewK8sRuntime(clientsManager *manager.Manager, eventBus *eventbus.EventBus, recorder recorder.EventRecorder, pool *pool.ConnectionPool, nodeName string) *K8sRuntime {
	clientset := config.LoadConfig()
	metricsClient := config.LoadMcConfig()
	if clientset == nil || metricsClient == nil {
		logs.Error("clientset or metricsClient is nil---")
		return nil
	}
	k8sMonitor := monitor.NewMonitor(clientset, eventBus, nodeName, clientsManager)
	k8sMonitor.Start()
	return &K8sRuntime{clientset: clientset, metricsClient: metricsClient, connectionPool: pool, recorder: recorder, monitor: k8sMonitor, eventBus: eventBus}
}
func (k *K8sRuntime) Kill(group *apis.Group, action *apis.Action, runtime *apis.Runtime, actionSpecName, runtimeSpecName string) error {
	logs.Infof("k8s runtime kill runtime: %s", runtime.Name)
	//yamlFilePath := runtime.Spec.Inputs[0].From
	yamlFilePath := runtime.Spec.Directory + "/" + runtime.Spec.Data[0].Name
	objList, err := entity.ParseK8sResourcesFromFile(yamlFilePath, *group.Status.Node)
	if err != nil {
		logs.Errorf("Get k8s resources from yaml file failed: %v", err)
	}
	err = DeleteResources(k.clientset, objList)
	if err != nil {
		logs.Errorf("Delete k8s resources from yaml file failed: %v", err)
	}
	return nil
}

// 粗粒度管理的启动方法
func (k *K8sRuntime) Run(group *apis.Group, action *apis.Action, runtime *apis.Runtime, actionSpecName, runtimeSpecName string) error {
	logs.Infof("k8s runtime for task: %s", group.Name)
	//先执行共同的操作,再各自调用代码
	// 1、首先读取yaml文件，转换为资源
	//yamlFilePath := runtime.Spec.Inputs[0].From
	yamlFilePath := runtime.Spec.Directory + "/" + runtime.Spec.Data[0].Name
	logs.Infof("group.Status.Node:%v", *group.Status.Node)
	objList, err := entity.ParseK8sResourcesFromFile(yamlFilePath, *group.Status.Node)
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
	return nil
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
		k.recorder.Event(action, apis.EventTypeWarning, events.FailedToStartCommand, fmt.Sprintf("Runtime Name:\t %s failed to use rpc RunAppStore()", runtime.Name))
		// 同时应该发送失败的Notify
		k.notifyRuntimeEndPhase(group.Name, group.Namespace, action.Spec.Name, runtime.Spec.Name, apis.Failed, apis.Time{time.Now()}, apis.Time{time.Now()})
	}
	client := k.getClient(*runtime.Spec.EnableFineGrainedControlService, *runtime.Spec.EnableFineGrainedControlPort)
	// rpc调用store()
	_, err := client.RunAppStore()
	if err != nil {
		logs.Error(err, "Store application status failed")
	}
	k.recorder.Event(action, apis.EventTypeNormal, events.StoredCommand, fmt.Sprintf("Runtime Name:\t %s rpc RunAppStore()", runtime.Name))
	return "aass"
}
func (k *K8sRuntime) RestoreData(group *apis.Group, action *apis.Action, runtime *apis.Runtime, actionSpecName, runtimeSpecName string) error {
	keyStatus := ""
	if runtime.Spec.EnableFineGrainedControlService == nil || runtime.Spec.EnableFineGrainedControlPort == nil {
		logs.Errorf("Need input EnableFineGrainedControlService and EnableFineGrainedControlPort, now all is nil in RestoreData")
		// 发送失败事件
		k.recorder.Event(action, apis.EventTypeWarning, events.FailedToStartCommand, fmt.Sprintf("Runtime Name:\t %s failed to use rpc RestoreData()", runtime.Name))
		// 同时应该发送失败的Notify
		k.notifyRuntimeEndPhase(group.Name, group.Namespace, action.Spec.Name, runtime.Spec.Name, apis.Failed, apis.Time{time.Now()}, apis.Time{time.Now()})
	}
	client := k.getClient(*runtime.Spec.EnableFineGrainedControlService, *runtime.Spec.EnableFineGrainedControlPort)
	_, err := client.RunAppRestore(keyStatus)
	if err != nil {
		logs.Errorf("Restore application status failed")
	}
	k.recorder.Event(action, apis.EventTypeNormal, events.RestoredCommand, fmt.Sprintf("Runtime Name:\t %s rpc RunAppRestore()", runtime.Name))
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
		k.recorder.Event(action, apis.EventTypeWarning, events.FailedToStartCommand, fmt.Sprintf("Runtime Name:\t %s failed to use rpc StartRuntime()", runtime.Name))
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
		k.recorder.Event(action, apis.EventTypeWarning, events.FailedToStartCommand, fmt.Sprintf("Runtime Name:\t %s failed to use rpc InitRuntime()", runtime.Name))
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
		k.recorder.Event(action, apis.EventTypeWarning, events.FailedToStartCommand, fmt.Sprintf("Runtime Name:\t %s failed to use rpc StopRuntime()", runtime.Name))
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

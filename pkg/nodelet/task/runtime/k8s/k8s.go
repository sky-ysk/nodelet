package k8s

import (
	"fmt"
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/client-go/tools/recorder"
	"hit.edu/framework/pkg/component-base/logs"
	"hit.edu/framework/pkg/nodelet/events"
	"hit.edu/framework/pkg/nodelet/events/eventbus"
	grpc_client "hit.edu/framework/pkg/nodelet/task/interaction/intwithRuntime/grpc-client"
	"hit.edu/framework/pkg/nodelet/task/interaction/intwithRuntime/pool"
	"hit.edu/framework/pkg/nodelet/task/runtime/k8s/config"
	"hit.edu/framework/pkg/nodelet/task/runtime/k8s/entity"
	"hit.edu/framework/pkg/nodelet/task/runtime/k8s/monitor"
	"k8s.io/client-go/kubernetes"
	metricsclientset "k8s.io/metrics/pkg/client/clientset/versioned"
)

type K8sRuntime struct {
	clientset     *kubernetes.Clientset
	metricsClient *metricsclientset.Clientset //go get k8s.io/metrics/pkg/client/clientset/versioned 从 Metrics Server 获取的实时监控数据
	//client         *grpc_client.RuntimeClient
	connectionPool *pool.ConnectionPool
	// 全局事件发送的组件
	recorder recorder.EventRecorder
	monitor  *monitor.Monitor
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

func NewK8sRuntime(eventBus *eventbus.EventBus, recorder recorder.EventRecorder, pool *pool.ConnectionPool) *K8sRuntime {
	clientset := config.LoadConfig()
	metricsClient := config.LoadMcConfig()
	if metricsClient == nil || metricsClient == nil {
		logs.Error("clientset or metricsClient is nil---")
		return nil
	}
	k8sMonitor := monitor.NewMonitor(clientset, eventBus)
	k8sMonitor.Start()
	return &K8sRuntime{clientset: clientset, metricsClient: metricsClient, connectionPool: pool, recorder: recorder, monitor: k8sMonitor}
}
func (k *K8sRuntime) Kill(group *apis.Group, action *apis.Action, runtime *apis.Runtime, actionSpecName, runtimeSpecName string) error {
	logs.Infof("k8s runtime kill task: %s", group.Name)
	return nil
}

// 粗粒度管理的启动方法
func (k *K8sRuntime) Run(group *apis.Group, action *apis.Action, runtime *apis.Runtime, actionSpecName, runtimeSpecName string) error {
	logs.Infof("k8s runtime for task: %s", group.Name)
	//先执行共同的操作,再各自调用代码
	k.monitor.SetState(group, action, runtime, actionIndex, runtimeIndex)
	switch runtime.Spec.Type {
	case apis.ByDeployment:
		deployment1 := entity.GetDeploymentFromParam1(&runtime.Deployment)
		logs.Info("[RUN]-----------------k8s deployment created----------------------------")
		CreateDeployment(k.clientset, deployment1) // 创建Deployment
		return nil
	case apis.ByService:
		service1 := entity.GetServiceFromParam1(&runtime.Service)
		logs.Info("[RUN]-----------------k8s Service created----------------------------")
		CreateService(k.clientset, service1)
		return nil
	case apis.ByPod:
		podInfo1 := entity.GetPodFromParam1(&runtime.Pod)
		logs.Info("[RUN]-----------------k8s Pod created----------------------------")
		CreatePod(k.clientset, podInfo1)
		return nil
	default:
		err := fmt.Errorf("unsupported runtime type: %s", runtime.Type)
		logs.Error(err, "Runtime type not supported")
		return err
	}
}

func (k *K8sRuntime) CheckRuntimeStatus(group *apis.Group, action *apis.Action, runtime *apis.Runtime) (string, error) {

	return "", nil
}
func (k *K8sRuntime) StoreData(group *apis.Group, action *apis.Action, runtime *apis.Runtime, actionSpecName, runtimeSpecName string) string {
	// 保存任务状态，调用grpc接口获取任务状态，返回任务状态值即可
	client := k.getClient(runtime.EnableFineGrainedControlService, runtime.EnableFineGrainedControlPort)
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
	client := k.getClient(runtime.EnableFineGrainedControlService, runtime.EnableFineGrainedControlPort)
	_, error := client.RunAppRestore(keyStatus)
	if error != nil {
		logs.Errorf("Restore application status failed")
	}
	k.recorder.Event(action, apis.EventTypeNormal, events.RestoredCommand, fmt.Sprintf("Runtime Name:\t %s rpc RunAppRestore()", runtime.Name))
	return error
}
func (k *K8sRuntime) StartRuntime(group *apis.Group, action *apis.Action, runtime *apis.Runtime, actionSpecName, runtimeSpecName string) error {
	k.monitor.SetState(group, action, runtime, actionIndex, runtimeIndex)
	switch runtime.Type {
	case apis.ByDeployment:
		deployment1 := entity.GetDeploymentFromParam1(&runtime.Deployment)
		logs.Info("[StartRuntime]-----------------k8s deployment created----------------------------")
		CreateDeployment(k.clientset, deployment1) // 创建Deployment
	case apis.ByPod:
		podInfo1 := entity.GetPodFromParam1(&runtime.Pod)
		logs.Info("[StartRuntime]-----------------k8s Pod created----------------------------")
		CreatePod(k.clientset, podInfo1)
	default:
		err := fmt.Errorf("unsupported runtime type: %s", runtime.Type)
		logs.Error(err, "Runtime type not supported")
	}
	logs.Infof("开启grpc客户端连接pod当中的grpc服务端，ip：%v,端口：%v", runtime.EnableFineGrainedControlService, runtime.EnableFineGrainedControlPort)
	client := k.getClient(runtime.EnableFineGrainedControlService, runtime.EnableFineGrainedControlPort)
	_, error := client.RunAppStart()
	if error != nil {
		logs.Errorf("Start application failed")
	}
	return error
}
func (k *K8sRuntime) InitRuntime(group *apis.Group, action *apis.Action, runtime *apis.Runtime, actionSpecName, runtimeSpecName string) error {
	k.monitor.SetState(group, action, runtime, actionIndex, runtimeIndex)
	switch runtime.Type {
	case apis.ByDeployment:
		deployment1 := entity.GetDeploymentFromParam1(&runtime.Deployment)
		logs.Info("[InitRuntime]-----------------k8s deployment created----------------------------")
		CreateDeployment(k.clientset, deployment1) // 创建Deployment
	case apis.ByPod:
		podInfo1 := entity.GetPodFromParam1(&runtime.Pod)
		logs.Info("[InitRuntime]-----------------k8s Pod created----------------------------")
		CreatePod(k.clientset, podInfo1)
	default:
		err := fmt.Errorf("unsupported runtime type: %s", runtime.Type)
		logs.Error(err, "Runtime type not supported")
	}
	client := k.getClient(runtime.EnableFineGrainedControlService, runtime.EnableFineGrainedControlPort)
	_, error := client.RunAppInit()
	if error != nil {
		logs.Errorf("Init application failed")
	}
	return error
}
func (k *K8sRuntime) StopRuntime(group *apis.Group, action *apis.Action, runtime *apis.Runtime, actionSpecName, runtimeSpecName string) error {
	client := k.getClient(runtime.EnableFineGrainedControlService, runtime.EnableFineGrainedControlPort)
	_, error := client.RunAppStop()
	if error != nil {
		logs.Errorf("Stop application failed")
	}
	return error
}
func (k *K8sRuntime) getClient(service, port string) *grpc_client.RuntimeClient {
	return grpc_client.NewK8sRuntimeClient(service, port, k.connectionPool)
}

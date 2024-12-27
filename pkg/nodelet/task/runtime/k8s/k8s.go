package k8s

import (
	"fmt"
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/component-base/logs"
	"hit.edu/framework/pkg/nodelet/task/runtime/k8s/config"
	"hit.edu/framework/pkg/nodelet/task/runtime/k8s/entity"
	"k8s.io/client-go/kubernetes"
	metricsclientset "k8s.io/metrics/pkg/client/clientset/versioned"
)

type K8sRuntime struct {
	clientset     *kubernetes.Clientset
	metricsClient *metricsclientset.Clientset //go get k8s.io/metrics/pkg/client/clientset/versioned 从 Metrics Server 获取的实时监控数据
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

func NewK8sRuntime() *K8sRuntime {
	clientset := config.LoadConfig()
	metricsClient := config.LoadMcConfig()
	if metricsClient == nil || metricsClient == nil {
		logs.Error("clientset or metricsClient is nil---")
		return nil
	}
	return &K8sRuntime{clientset: clientset, metricsClient: metricsClient}
}
func (k *K8sRuntime) Kill(group *apis.Group, action *apis.Action, runtime *apis.Runtime) error {
	logs.Infof("k8s runtime kill task: %s", group.Name)
	return nil
}

// 目前我把pod、Deployment、Service的name和namespace当成放在action的meta.ObjectMeta当中
func (k *K8sRuntime) Run(group *apis.Group, action *apis.Action, runtime *apis.Runtime) error {
	logs.Infof("k8s runtime for task: %s", group.Name)
	//先执行共同的操作
	//再各自调用代码
	switch runtime.Type {
	case apis.ByDeployment:
		deployment := entity.NewDeployment(group.Name, group.Namespace, group.Spec.Labels, runtime.Replicas, runtime.Name, runtime.Image, runtime.Selector, action.Spec.EnableFineGrainedControl)
		return CreateDeployment(k.clientset, deployment)
		//fmt.Printf("%v\n", deployment.DeploymentName)
		//logs.Info("-----------------k8s deployment created----------------------------")
		//return nil
	case apis.ByService:
		service := entity.NewService(group.Name, group.Namespace, runtime.Selector, runtime.Ports, runtime.ServiceType)
		return CreateService(k.clientset, service)
		//fmt.Printf("%v\n", service.ServiceName)
		//logs.Info("-----------------k8s Service created----------------------------")
		//return nil
	case apis.ByPod:
		podInfo := entity.NewPodFromParam(group.Name, group.Namespace, runtime.Name, runtime.Image, action.Spec.EnableFineGrainedControl)
		return CreatePod(k.clientset, podInfo)
		//fmt.Printf("%v\n", podInfo.PodName)
		//logs.Info("-----------------k8s Pod created----------------------------")
		//return nil
	default:
		err := fmt.Errorf("unsupported runtime type: %s", runtime.Type)
		logs.Error(err, "Runtime type not supported")
		return err
	}
}
func (k *K8sRuntime) CheckTaskStatus(group *apis.Group, action *apis.Action, runtime *apis.Runtime) (string, error) {

	return "", nil
}

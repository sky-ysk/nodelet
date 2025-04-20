package monitor

import (
	"fmt"
	"hit.edu/framework/pkg/apimachinery/labels"
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/client-go/tools/cache"
	"hit.edu/framework/pkg/component-base/logs"
	"hit.edu/framework/pkg/nodelet/events"
	"hit.edu/framework/pkg/nodelet/events/eventbus"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"sync"
	"time"
)

const (
	CreateorLabel = "app.kubernetes.io/created-by"
	SystemName    = "deploy-system"
)

type ResourceState struct {
	group        *apis.Group
	action       *apis.Action
	runtime      *apis.Runtime
	actionIndex  int
	runtimeIndex int
}

//	type DeploymentMonitor struct {
//	   clientset *kubernetes.Clientset
//	}
type Monitor struct {
	clientset kubernetes.Interface
	eventBus  *eventbus.EventBus
	stopChan  chan struct{}
	//stateMap   sync.Map // 保存资源状态的并发安全存储
	eventHooks []func(eventType string, obj interface{})
	infoMap    sync.Map
}

func NewMonitor(clientset kubernetes.Interface, eventBus *eventbus.EventBus) *Monitor {
	return &Monitor{
		clientset: clientset,
		stopChan:  make(chan struct{}),
		eventBus:  eventBus,
	}
}
func (m *Monitor) Start() {
	labelSelector := labels.SelectorFromSet(labels.Set{
		CreateorLabel: SystemName,
	})
	factory := informers.NewSharedInformerFactoryWithOptions(
		m.clientset, 10*time.Minute, informers.WithTweakListOptions(func(options *metav1.ListOptions) {
			options.LabelSelector = labelSelector.String()
		}))
	// 部署监控
	deployInformer := factory.Apps().V1().Deployments().Informer()
	deployInformer.AddEventHandler(m.newEventHandler(apis.ByDeployment))

	// Pod监控
	podInformer := factory.Core().V1().Pods().Informer()
	podInformer.AddEventHandler(m.newEventHandler(apis.ByPod))

	// Service监控
	serviceInformer := factory.Core().V1().Services().Informer()
	serviceInformer.AddEventHandler(m.newEventHandler(apis.ByService))

	go deployInformer.Run(m.stopChan)
	go podInformer.Run(m.stopChan)
	go serviceInformer.Run(m.stopChan)

}

func (m *Monitor) newEventHandler(resType apis.RuntimeType) cache.ResourceEventHandlerFuncs {
	return cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			m.handleEvent("ADDED", resType, nil, obj)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			m.handleEvent("UPDATED", resType, oldObj, newObj)
		},
		DeleteFunc: func(obj interface{}) {
			m.handleEvent("DELETED", resType, obj, nil)
		},
	}
}
func (m *Monitor) handleEvent(eventType string, resType apis.RuntimeType, oldObj, newObj interface{}) {
	logs.Infof("resType：%v======================触发k8s的Event事件处理：%v", resType, eventType)
	//var state ResourceState
	nowTime := apis.Time{time.Now()}
	switch resType {
	case apis.ByDeployment:
		var currentDeploy, previousDeploy *appsv1.Deployment
		// 解析对象
		switch eventType {
		case "ADDED":
			currentDeploy = newObj.(*appsv1.Deployment)
			previousDeploy = nil
		case "UPDATED":
			previousDeploy = oldObj.(*appsv1.Deployment)
			currentDeploy = newObj.(*appsv1.Deployment)
		case "DELETED":
			currentDeploy = oldObj.(*appsv1.Deployment)
			previousDeploy = nil
		}
		if currentDeploy == nil {
			return
		}

		// 获取关联状态
		value, ok := m.infoMap.Load(stateKey(apis.ByDeployment, currentDeploy.Namespace, currentDeploy.Name))
		if !ok {
			logs.Errorf("Deployment %s/%s not found in infoMap", currentDeploy.Namespace, currentDeploy.Name)
			return
		}

		rs := value.(ResourceState)
		// 状态判定逻辑
		currentStatus := getDeploymentStatus(currentDeploy)
		previousStatus := ""
		if previousDeploy != nil {
			previousStatus = getDeploymentStatus(previousDeploy)
		}
		// 确定是否需要通知
		shouldNotify := false
		switch eventType {
		case "ADDED":
			shouldNotify = true
		case "UPDATED":
			shouldNotify = currentStatus != previousStatus
		case "DELETED":
			shouldNotify = true
		}
		if shouldNotify {
			phase := convertDeploymentStatus(currentDeploy, currentStatus, eventType)
			if phase == apis.Killed || phase == apis.Failed || phase == apis.Successed {
				m.notifyRuntimeEndPhase(rs.group.Name, rs.actionIndex, rs.runtimeIndex, phase, nowTime, nowTime)
			} else if phase == apis.Running {
				m.notifyRuntimeStartPhase(rs.group.Name, rs.actionIndex, rs.runtimeIndex, "", phase, nowTime, nowTime)
			}
			logs.Infof("[Deployment]=============发送状态：%v 给事件处理模块===========", phase)
		}
	case apis.ByPod:
		var currentPod, previousPod *corev1.Pod
		var targetPhase corev1.PodPhase
		// 解析对象
		switch eventType {
		case "ADDED":
			currentPod = newObj.(*corev1.Pod)
			previousPod = nil
		case "UPDATED":
			previousPod = oldObj.(*corev1.Pod)
			currentPod = newObj.(*corev1.Pod)
		case "DELETED":
			currentPod = oldObj.(*corev1.Pod) // 删除时从oldObj获取
			previousPod = nil
		}

		// 获取关联状态
		if currentPod == nil {
			return
		}
		// 仅关注最终状态
		targetPhase = currentPod.Status.Phase
		if !isTargetPhase(targetPhase) {
			return
		}
		// 获取关联状态
		value, ok := m.infoMap.Load(stateKey(apis.ByPod, currentPod.Namespace, currentPod.Name))
		if !ok {
			logs.Errorf("Pod %s/%s not found in infoMap", currentPod.Namespace, currentPod.Name)
			return
		}
		rs := value.(ResourceState)

		// 状态变化检测
		shouldNotify := false
		switch eventType {
		case "ADDED":
			// 只有当直接进入目标状态时通知（如已缓存的 Pod）
			shouldNotify = true
		case "UPDATED":
			previousPhase := previousPod.Status.Phase
			shouldNotify = previousPhase != targetPhase && isTargetPhase(targetPhase)
		case "DELETED":
			// 删除时强制上报最终状态
			shouldNotify = true
		}
		if shouldNotify {
			phase := convertPodPhase(targetPhase)
			// 特殊处理删除事件
			if eventType == "DELETED" {
				phase = apis.Killed // 或根据实际状态处理
			}
			if phase == apis.Killed || phase == apis.Failed || phase == apis.Successed {
				m.notifyRuntimeEndPhase(rs.group.Name, rs.actionIndex, rs.runtimeIndex, phase, nowTime, nowTime)
			} else {
				m.notifyRuntimeStartPhase(rs.group.Name, rs.actionIndex, rs.runtimeIndex, "", phase, nowTime, nowTime)
			}
			logs.Infof("[Pod]=============发送状态：%v 给事件处理模块===========", phase)
		}
	case apis.ByService:
		var currentService, previousService *corev1.Service

		// 解析新旧对象
		switch eventType {
		case "ADDED":
			currentService = newObj.(*corev1.Service)
			previousService = nil
		case "UPDATED":
			previousService = oldObj.(*corev1.Service)
			currentService = newObj.(*corev1.Service)
		case "DELETED":
			currentService = oldObj.(*corev1.Service)
			previousService = nil
		}
		if currentService == nil {
			return
		}
		// 获取关联状态
		value, ok := m.infoMap.Load(stateKey(apis.ByService, currentService.Namespace, currentService.Name))
		if !ok {
			logs.Errorf("Service %s/%s not found in infoMap", currentService.Namespace, currentService.Name)
			return
		}

		rs := value.(ResourceState)
		if eventType == "ADDED" {
			m.notifyRuntimeStartPhase(rs.group.Name, rs.actionIndex, rs.runtimeIndex, "", apis.Running, nowTime, nowTime)
		}
		// 状态判定逻辑
		currentStatus := getServiceStatus(currentService)
		previousStatus := ""
		if previousService != nil {
			previousStatus = getServiceStatus(previousService)
		}

		// 确定是否需要通知
		shouldNotify := false
		switch eventType {
		case "ADDED":
			shouldNotify = true // 新增服务总是通知
		case "UPDATED":
			shouldNotify = currentStatus != previousStatus
		case "DELETED":
			shouldNotify = true
		}

		if shouldNotify {
			phase := convertServiceStatus(currentService, currentStatus, eventType)
			if phase == apis.Killed || phase == apis.Failed || phase == apis.Successed {
				m.notifyRuntimeEndPhase(rs.group.Name, rs.actionIndex, rs.runtimeIndex, phase, nowTime, nowTime)
			} else {
				m.notifyRuntimeStartPhase(rs.group.Name, rs.actionIndex, rs.runtimeIndex, "", phase, nowTime, nowTime)
			}
			logs.Infof("[Service]=============发送状态：%v 给事件处理模块===========", phase)
		}
	}

	// 触发事件钩子（根据需求选择使用 oldObj 或 newObj）
	var eventObj interface{}
	if newObj != nil {
		eventObj = newObj
	} else {
		eventObj = oldObj
	}
	for _, hook := range m.eventHooks {
		hook(eventType, eventObj)
	}
}
func stateKey(resType apis.RuntimeType, namespace, name string) string {
	return fmt.Sprintf("%s/%s/%s", resType, namespace, name)
}

// 注册事件回调
func (m *Monitor) RegisterHook(hook func(eventType string, obj interface{})) {
	m.eventHooks = append(m.eventHooks, hook)
}

// 保存资源的信息
func (m *Monitor) SetState(group *apis.Group, action *apis.Action, runtime *apis.Runtime, actionIndex, runtimeIndex int) {
	var state ResourceState
	state = ResourceState{
		group:        group,
		action:       action,
		runtime:      runtime,
		actionIndex:  actionIndex,
		runtimeIndex: runtimeIndex,
	}
	switch runtime.Spec.Type {
	case apis.ByDeployment:
		_, loaded := m.infoMap.LoadOrStore(stateKey(apis.ByDeployment, runtime.Spec.Deployment.Namespace, runtime.Deployment.Name), state)
		if loaded {
			logs.Errorf("Key %s already exists, overwriting", stateKey(apis.ByDeployment, runtime.Deployment.Namespace, runtime.Deployment.Name))
		}
	case apis.ByPod:
		_, loaded := m.infoMap.LoadOrStore(stateKey(apis.ByPod, runtime.Pod.Namespace, runtime.Pod.Name), state)
		if loaded {
			logs.Errorf("Key %s already exists, overwriting", stateKey(apis.ByPod, runtime.Pod.Namespace, runtime.Pod.Name))
		}
	case apis.ByService:
		_, loaded := m.infoMap.LoadOrStore(stateKey(apis.ByService, runtime.Service.Namespace, runtime.Service.Name), state)
		if loaded {
			logs.Errorf("Key %s already exists, overwriting", stateKey(apis.ByService, runtime.Service.Namespace, runtime.Service.Name))
		}
	}
}

// 停止监控
func (m *Monitor) Stop() {
	close(m.stopChan)
}

// 通过 EventBus 通知 Runtime 状态更新
func (m *Monitor) notifyRuntimeStartPhase(groupName string, actionIndex, runtimeIndex int, processId string, phase apis.Phase, startAt, lastTime apis.Time) {
	event := events.RuntimeStartPhaseEvent1{
		GroupName:    groupName,
		ActionIndex:  actionIndex,
		RuntimeIndex: runtimeIndex,
		ProcessId:    processId,
		Phase:        phase,
		StartAt:      startAt,
		LastTime:     lastTime,
	}
	m.eventBus.Publish(event)
}
func (m *Monitor) notifyRuntimeEndPhase(groupName string, actionIndex, runtimeIndex int, phase apis.Phase, finishTime, lastTime apis.Time) {
	event := events.RuntimeEndPhaseEvent1{
		GroupName:    groupName,
		ActionIndex:  actionIndex,
		RuntimeIndex: runtimeIndex,
		Phase:        phase,
		FinishAt:     finishTime,
		LastTime:     lastTime,
	}
	m.eventBus.Publish(event)
}

// 获取 Deployment 状态
func getDeploymentStatus(d *appsv1.Deployment) string {
	for _, cond := range d.Status.Conditions {
		switch cond.Type {
		case appsv1.DeploymentAvailable:
			if cond.Status == corev1.ConditionTrue {
				return "Available"
			}
		case appsv1.DeploymentProgressing:
			if cond.Status == corev1.ConditionFalse {
				return "Failed"
			}
		}
	}
	return "Progressing"
}

// 状态转换逻辑（
func convertDeploymentStatus(deploy *appsv1.Deployment, status string, eventType string) apis.Phase {
	if eventType == "DELETED" {
		return apis.Killed
	}

	switch status {
	case "Available":
		return apis.Successed
	case "Failed":
		return apis.Failed
	case "Progressing":
		// 检查是否超时
		if isDeploymentTimeout(deploy) {
			return apis.Failed
		}
		return apis.Running
	default:
		return apis.Unknown
	}
}

// 超时判断
func isDeploymentTimeout(d *appsv1.Deployment) bool {
	progressDeadline := int32(600) // 默认值
	if d.Spec.ProgressDeadlineSeconds != nil {
		progressDeadline = *d.Spec.ProgressDeadlineSeconds
	}
	return time.Since(d.CreationTimestamp.Time) > time.Duration(progressDeadline)*time.Second
}

// Pod Phase 转换逻辑
func convertPodPhase(phase corev1.PodPhase) apis.Phase {
	switch phase {
	case corev1.PodPending:
		return apis.Running
	case corev1.PodRunning:
		return apis.Running
	case corev1.PodSucceeded:
		return apis.Successed
	case corev1.PodFailed:
		return apis.Failed
	default:
		return apis.Unknown
	}
}

// 判断是否为需要监控的状态
func isTargetPhase(phase corev1.PodPhase) bool {
	return phase == corev1.PodRunning ||
		phase == corev1.PodSucceeded ||
		phase == corev1.PodFailed
}

// 辅助函数：获取服务状态标识
func getServiceStatus(svc *corev1.Service) string {
	// NodePort 类型检查
	if svc.Spec.Type == corev1.ServiceTypeNodePort {
		for _, port := range svc.Spec.Ports {
			if port.NodePort == 0 { // Kubernetes 分配失败
				return "Pending"
			}
		}
		return "Ready"
	}
	// 判断 LoadBalancer 类型
	if svc.Spec.Type == corev1.ServiceTypeLoadBalancer {
		if len(svc.Status.LoadBalancer.Ingress) > 0 {
			return "Ready"
		}
		return "Pending"
	}

	// 其他类型判断 ClusterIP
	if svc.Spec.ClusterIP != "" && svc.Spec.ClusterIP != corev1.ClusterIPNone {
		return "Ready"
	}
	return "Pending"
}

// 添加详细状态判断
func convertServiceStatus(currentService *corev1.Service, status string, eventType string) apis.Phase {
	if eventType == "DELETED" {
		return apis.Killed
	}
	switch status {
	case "Ready":
		return apis.Successed
	case "Pending":
		// 根据超时时间判断是否转为失败
		if isCriticalService(currentService) && isServiceTimeout(currentService) {
			return apis.Failed
		}
		return apis.Running // 新增中间状态
	default:
		return apis.Failed
	}
}

// 判断是否为需要严格超时控制的服务类型
func isCriticalService(svc *corev1.Service) bool {
	return svc.Spec.Type == corev1.ServiceTypeNodePort ||
		svc.Spec.Type == corev1.ServiceTypeLoadBalancer
}

// 超时判断（例如创建超过5分钟）
func isServiceTimeout(svc *corev1.Service) bool {
	return time.Since(svc.CreationTimestamp.Time) > 5*time.Minute
}

//
//func (dm *DeploymentMonitor) WatchDeploymentStatus(namespace, name string) (<-chan string, error) {
//  fieldSelector := fields.OneTermEqualSelector("metadata.name", name).String()
//  watcher, err := dm.clientset.AppsV1().Deployments(namespace).Watch(
//     context.TODO(),
//     metav1.ListOptions{
//        FieldSelector: fieldSelector,
//     },
//  )
//  if err != nil {
//     return nil, err
//  }
//
//  statusChan := make(chan string)
//  go func() {
//     defer close(statusChan)
//     for event := range watcher.ResultChan() {
//        deploy, ok := event.Object.(*appsv1.Deployment)
//        if !ok {
//           continue
//        }
//
//        status := "Pending"
//        for _, cond := range deploy.Status.Conditions {
//           if cond.Type == appsv1.DeploymentAvailable {
//              if cond.Status == corev1.ConditionTrue {
//                 status = "Available"
//              } else {
//                 status = "Unavailable"
//              }
//           }
//        }
//        statusChan <- fmt.Sprintf("[%s/%s] Status: %s", namespace, name, status)
//     }
//  }()
//
//  return statusChan, nil
//}

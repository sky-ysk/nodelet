package monitor

//import (
//	"context"
//	"hit.edu/framework/pkg/component-base/logs"
//	"hit.edu/framework/pkg/nodelet/task/runtime/k8s/config"
//	EM "hit.edu/framework/pkg/nodelet/task/runtime/k8s/entity/entity_Manager"
//	v1 "k8s.io/api/core/v1"
//	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
//	"k8s.io/client-go/kubernetes"
//	metricsclientset "k8s.io/metrics/pkg/client/clientset/versioned"
//	"sync"
//	"time"
//)
//
//type PodMonitor struct {
//	clientset     *kubernetes.Clientset
//	metricsClient *metricsclientset.Clientset //go get k8s.io/metrics/pkg/client/clientset/versioned
//	stopCh        chan struct{}
//	mu            sync.Mutex
//}
//
//func NewPodMonitor() *PodMonitor {
//	clientset := config.LoadConfig()
//	metricsClient := config.LoadMcConfig()
//	if clientset == nil || metricsClient == nil {
//		logs.Error("clientset or metricsClient is nil")
//		return nil
//	}
//	podMonitor := &PodMonitor{
//		clientset:     clientset,
//		metricsClient: metricsClient,
//		stopCh:        make(chan struct{})}
//	podMonitor.Start() //开启监听
//	return podMonitor
//}
//
//func (pm *PodMonitor) Start() {
//	logs.Info("PodMonitor start")
//	pm.PeriodicallyCheckStatus(10 * time.Second)
//}
//
//func (pm *PodMonitor) PeriodicallyCheckStatus(interval time.Duration) {
//	ticker := time.NewTicker(interval)
//	go func() {
//		for {
//			select {
//			case <-ticker.C: //ticker.C是一个通道，当ticker定时器每隔指定的时间，就会忘通道中写入数据，那么通道中就会有数据，就可以被取出
//				pm.CheckStatus()
//				pm.CheckPodResource()
//			case <-pm.stopCh:
//				ticker.Stop()
//				return
//			}
//
//		}
//	}()
//}
//
//func (pm *PodMonitor) CheckStatus() {
//	// 加锁只保护与 pods 相关的操作
//	pm.mu.Lock()
//	defer pm.mu.Unlock()
//	if pods := EM.GetInstance().GetAllPods(); len(pods) != 0 {
//		for podId := range pods {
//			//获取pod
//			pod, err := pm.clientset.CoreV1().Pods(podId.Namespace).Get(context.TODO(), podId.PodName, metav1.GetOptions{})
//			if err != nil {
//				logs.Error("Failed to get entity")
//				continue
//			}
//			if pod.Status.Phase == v1.PodFailed || pod.Status.Phase == v1.PodUnknown { //运行失败、Unknown
//				logs.Info("Pod Failed/Unkown Status")
//				pods[podId] = false
//				//触发事件，向调度器上报
//				continue
//			}
//			if pod.Status.Phase == v1.PodSucceeded {
//				logs.Info("Pod %s has completed. Removing from monitoring.", podId.PodName)
//				EM.GetInstance().RemovePod(podId.Namespace, podId.PodName)
//				continue
//			}
//			if pod.Status.Phase != v1.PodRunning {
//				logs.Info("Pod %s is not running as expected. Status: %s", podId.PodName, pod.Status.Phase)
//			}
//		}
//	}
//}
//func (pm *PodMonitor) CheckPodResource() {
//	// 加锁只保护与 pods 相关的操作
//	pm.mu.Lock()
//	defer pm.mu.Unlock()
//	if pods := EM.GetInstance().GetAllPods(); len(pods) != 0 {
//		for podId := range pods {
//			metrics, err := pm.metricsClient.MetricsV1beta1().PodMetricses(podId.Namespace).Get(context.TODO(), podId.PodName, metav1.GetOptions{})
//			if err != nil {
//				logs.Error("Failed to get entity metrics", err)
//				continue
//			}
//			var totalCPU int64 = 0
//			var totalMemory int64 = 0
//			for _, container := range metrics.Containers {
//				// 打印每个容器的资源占用
//				cpuUsage := container.Usage.Cpu().MilliValue()  // CPU 使用量（毫核）
//				memoryUsage := container.Usage.Memory().Value() // 内存使用量（字节）
//				logs.Info("Pod %s, Container %s is using %d CPU cores and %d memory bytes",
//					podId.PodName, container.Name, cpuUsage, memoryUsage)
//				// 累加每个容器的 CPU 和内存使用量
//				totalCPU += cpuUsage
//				totalMemory += memoryUsage
//			}
//			// 打印整个 Pod 的总资源占用
//			logs.Info("Pod %s is using total %d CPU cores and %d memory bytes",
//				podId.PodName, totalCPU, totalMemory)
//		}
//	}
//}
//
//func (pm *PodMonitor) Stop() {
//	close(pm.stopCh)
//	logs.Info("Kubernetes Monitor stopped")
//}

package main

import (
	"context"
	"fmt"
	"time"

	"hit.edu/framework/pkg/nodelet/task/runtime/k8s/config"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func main() {
	// // 读取 kubeconfig 文件路径
	// kubeconfig := os.Getenv("KUBECONFIG")
	// if kubeconfig == "" {
	//     kubeconfig = "/path/to/kubeconfig"
	// }

	// // 使用 kubeconfig 文件配置客户端
	// config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	// if err != nil {
	//     panic(err)
	// }

	// 创建 Metrics API 客户端
	metricsClient := config.LoadMcConfig()
	if metricsClient == nil {
		fmt.Println("clientset or metricsClient is nil---")
		return
	}
	fmt.Println("clientset or metricsClient is not nil---")
	ticker := time.NewTicker(1 * time.Second)
	podName := "nginx1-deploy-67b489c674-bwkl2"
	defer ticker.Stop()
	for range ticker.C {
		fmt.Println("start to get pod metrics")
		// 获取Pod资源使用情况
		podMetrics, err := metricsClient.MetricsV1beta1().PodMetricses("switch").Get(
			context.TODO(), podName, metav1.GetOptions{})

		if err != nil {
			fmt.Printf("Failed to get metrics for pod %s: %v\n", podName, err)
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

		fmt.Printf("[Monitor] Runtime %s (Pod: %s) - CPU: %dm, Memory: %.2f MB\n",
			"sky-test", podName, cpuMilli, memMB)
	}

	// // 获取指定命名空间中所有 Pod 的资源使用情况
	// podMetrics, err := metricsClient.MetricsV1beta1().PodMetricses("switch").List(context.Background(), metav1.ListOptions{})
	// if err != nil {
	// 	panic(err)
	// }

	// // 打印每个 Pod 的资源使用情况
	// for _, pod := range podMetrics.Items {
	// 	fmt.Printf("Pod: %s\n", pod.Name)
	// 	for _, container := range pod.Containers {
	// 		fmt.Printf("  Container: %s\n", container.Name)
	// 		fmt.Printf("    CPU: %s\n", container.Usage.Cpu().String())
	// 		fmt.Printf("    Memory: %s\n", container.Usage.Memory().String())
	// 	}
	// }
}

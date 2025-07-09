package main

import (
	"context"
	"fmt"

	"hit.edu/framework/pkg/nodelet/task/runtime/k8s/config"
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

	// 获取指定命名空间中所有 Pod 的资源使用情况
	podMetrics, err := metricsClient.MetricsV1beta1().PodMetricses("switch").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		panic(err)
	}

	// 打印每个 Pod 的资源使用情况
	for _, pod := range podMetrics.Items {
		fmt.Printf("Pod: %s\n", pod.Name)
		for _, container := range pod.Containers {
			fmt.Printf("  Container: %s\n", container.Name)
			fmt.Printf("    CPU: %s\n", container.Usage.Cpu().String())
			fmt.Printf("    Memory: %s\n", container.Usage.Memory().String())
		}
	}
}

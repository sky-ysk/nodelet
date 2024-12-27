package config

import (
	"hit.edu/framework/pkg/component-base/logs"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	metricsclientset "k8s.io/metrics/pkg/client/clientset/versioned"
	"os"
)

type K8sConfig struct {
}

const (
	kubeConfig = "/home/public/.kube/config"
)

func getK8sConfig() (*rest.Config, error) {
	if _, err := os.Stat("/var/run/secrets/kubernetes.io/serviceaccount/token"); err == nil {
		logs.Info("Running inside Kubernetes cluster (in-cluster config)")
		return rest.InClusterConfig() // 返回集群内配置
	}
	// 否则，使用 kubeconfig（集群外）
	logs.Info("Running outside Kubernetes cluster (using kubeConfig)")
	return clientcmd.BuildConfigFromFlags("", kubeConfig) // 返回集群外配置
}

func LoadConfig() *kubernetes.Clientset {
	config, err := getK8sConfig()
	if err != nil {
		return nil
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil
	}
	return clientset
}
func LoadMcConfig() *metricsclientset.Clientset {
	config, err := getK8sConfig()
	if err != nil {
		return nil
	}
	metricsClient, err := metricsclientset.NewForConfig(config)
	if err != nil {
		return nil
	}
	return metricsClient
}

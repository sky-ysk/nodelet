package k8s

import (
	"context"
	"hit.edu/framework/pkg/component-base/logs"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// CreateService 将 Service 资源提交到 Kubernetes
func CreateService(clientset *kubernetes.Clientset, service *corev1.Service) {
	svc, err := clientset.CoreV1().Services(service.Namespace).Create(
		context.TODO(),
		service,
		metav1.CreateOptions{},
	)
	if err != nil {
		logs.Error(err, "Failed to create service")
	}
	logs.Info("Service created successfully", "Service:", svc.Name)
}

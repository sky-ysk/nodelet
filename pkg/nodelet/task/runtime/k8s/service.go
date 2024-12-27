package k8s

import (
	"context"
	"hit.edu/framework/pkg/component-base/logs"
	"hit.edu/framework/pkg/nodelet/task/runtime/k8s/entity"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// CreateService 将 Service 资源提交到 Kubernetes
func CreateService(clientset *kubernetes.Clientset, service *entity.Service) error {
	s := service.CreateServiceTemplate()
	svc, err := clientset.CoreV1().Services(service.Namespace).Create(
		context.TODO(),
		&s,
		metav1.CreateOptions{},
	)
	if err != nil {
		logs.Error(err, "Failed to create service")
		return err
	}
	logs.Info("Service created successfully", "Service:", svc.Name)
	return nil
}

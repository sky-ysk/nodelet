package k8s

import (
	"context"
	"hit.edu/framework/pkg/component-base/logs"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// 将 Service 资源提交到 Kubernetes
func CreateService(clientset *kubernetes.Clientset, service *corev1.Service) {
	// 前置检查命名空间
	if err := EnsureNamespace(clientset, service.Namespace); err != nil {
		logs.Errorf("无法确保命名空间存在: %v", err)
		return
	}
	// 创建服务
	svc, err := clientset.CoreV1().Services(service.Namespace).Create(
		context.TODO(),
		service,
		metav1.CreateOptions{},
	)
	if err != nil {
		logs.Errorf("Failed to create service,err:%v", err)
	}
	logs.Info("Service created successfully", "Service:", svc.Name)
}

// 创建命名空间（如果不存在）
func EnsureNamespace(clientset *kubernetes.Clientset, namespace string) error {
	_, err := clientset.CoreV1().Namespaces().Get(
		context.TODO(),
		namespace,
		metav1.GetOptions{},
	)
	if errors.IsNotFound(err) {
		logs.Infof("命名空间 %s 不存在，正在创建...", namespace)
		ns := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: namespace,
			},
		}
		_, createErr := clientset.CoreV1().Namespaces().Create(
			context.TODO(),
			ns,
			metav1.CreateOptions{},
		)
		return createErr
	}
	return err
}

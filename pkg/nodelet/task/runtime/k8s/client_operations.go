package k8s

import (
	"context"
	"hit.edu/framework/pkg/component-base/logs"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
)

// 智能资源创建器
func CreateResources(clientset *kubernetes.Clientset, objects []runtime.Object) error {
	for _, obj := range objects {
		switch resource := obj.(type) {
		case *corev1.Pod:
			if err := createPod(clientset, resource); err != nil {
				return err
			}
		case *corev1.Service:
			if err := createService(clientset, resource); err != nil {
				return err
			}
		default:
			logs.Warnf("不支持的资源类型: %T", obj)
		}
	}
	return nil
}

// 私有化Pod创建逻辑
func createPod(clientset *kubernetes.Clientset, pod *corev1.Pod) error {
	_, err := clientset.CoreV1().Pods(pod.Namespace).Create(
		context.TODO(),
		pod,
		metav1.CreateOptions{FieldValidation: "Strict"},
	)
	return err
}

// 服务创建逻辑
func createService(clientset *kubernetes.Clientset, svc *corev1.Service) error {
	_, err := clientset.CoreV1().Services(svc.Namespace).Create(
		context.TODO(),
		svc,
		metav1.CreateOptions{FieldValidation: "Strict"},
	)
	return err
}

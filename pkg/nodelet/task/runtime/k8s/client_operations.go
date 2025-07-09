package k8s

import (
	"context"
	"fmt"
	"hit.edu/framework/pkg/component-base/logs"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
)

func CreateResource(clientset *kubernetes.Clientset, obj runtime.Object) error {
	// 获取资源命名空间（如果未指定，默认为 default）
	namespace := getObjectNamespace(obj)
	if namespace == "" {
		namespace = "default"
	}
	err := EnsureNamespace(clientset, namespace)
	if err != nil {
		logs.Errorf("无法确保命名空间存在: %v", err)
	}
	// 根据资源类型分发创建逻辑
	switch resource := obj.(type) {
	case *corev1.Pod:
		// 创建Pod
		logs.Info("-----------------k8s Pod created----------------------------")
		_, err := clientset.CoreV1().Pods(namespace).Create(
			context.TODO(), resource, metav1.CreateOptions{},
		)
		if handleCreateError(err, "Pod", resource.Name) != nil {
			return err
		}

		// 监控Pod资源
		// go monitorPodResources(clientset, namespace, resource.Name)

	case *appsv1.Deployment:
		logs.Info("-----------------k8s Deployment created----------------------------")
		_, err := clientset.AppsV1().Deployments(namespace).Create(
			context.TODO(), resource, metav1.CreateOptions{},
		)
		if handleCreateError(err, "Deployment", resource.Name) != nil {
			return err
		}

	case *corev1.Service:
		logs.Info("-----------------k8s Service created----------------------------")
		_, err := clientset.CoreV1().Services(namespace).Create(
			context.TODO(), resource, metav1.CreateOptions{},
		)
		if handleCreateError(err, "Service", resource.Name) != nil {
			return err
		}

	default:
		return fmt.Errorf("Unsupported resource type: %T", obj)
	}
	return nil
}

func CreateResources(clientset *kubernetes.Clientset, objects []runtime.Object) error {
	for _, obj := range objects {
		// 获取资源命名空间（如果未指定，默认为 default）
		namespace := getObjectNamespace(obj)
		if namespace == "" {
			namespace = "default"
		}

		// 根据资源类型分发创建逻辑
		switch resource := obj.(type) {
		case *corev1.Pod:
			_, err := clientset.CoreV1().Pods(namespace).Create(
				context.TODO(), resource, metav1.CreateOptions{},
			)
			if handleCreateError(err, "Pod", resource.Name) != nil {
				return err
			}

		case *appsv1.Deployment:
			_, err := clientset.AppsV1().Deployments(namespace).Create(
				context.TODO(), resource, metav1.CreateOptions{},
			)
			if handleCreateError(err, "Deployment", resource.Name) != nil {
				return err
			}

		case *corev1.Service:
			_, err := clientset.CoreV1().Services(namespace).Create(
				context.TODO(), resource, metav1.CreateOptions{},
			)
			if handleCreateError(err, "Service", resource.Name) != nil {
				return err
			}

		default:
			return fmt.Errorf("unsupported resource type: %T", obj)
		}
	}
	return nil
}

// 新增函数：根据对象列表删除资源
func DeleteResources(clientset *kubernetes.Clientset, objects []runtime.Object) error {
	// 逆序遍历资源（处理依赖关系，如先删 Pod 再删 Service）
	for i := len(objects) - 1; i >= 0; i-- {
		obj := objects[i]
		namespace := getObjectNamespace(obj)
		if namespace == "" {
			namespace = "default"
		}

		// 根据类型分发删除操作
		switch resource := obj.(type) {
		case *corev1.Pod:
			err := clientset.CoreV1().Pods(namespace).Delete(
				context.TODO(), resource.Name, metav1.DeleteOptions{},
			)
			logs.Info("-----------------k8s Pod delete----------------------------")
			if handleDeleteError(err, "Pod", resource.Name) != nil {
				return err
			}

		case *appsv1.Deployment:
			err := clientset.AppsV1().Deployments(namespace).Delete(
				context.TODO(), resource.Name, metav1.DeleteOptions{},
			)
			logs.Info("-----------------k8s Deployment delete----------------------------")
			if handleDeleteError(err, "Deployment", resource.Name) != nil {
				return err
			}

		case *corev1.Service:
			err := clientset.CoreV1().Services(namespace).Delete(
				context.TODO(), resource.Name, metav1.DeleteOptions{},
			)
			logs.Info("-----------------k8s Service delete----------------------------")
			if handleDeleteError(err, "Service", resource.Name) != nil {
				return err
			}

		default:
			return fmt.Errorf("unsupported resource type for deletion: %T", obj)
		}
	}
	return nil
}

// 辅助函数：获取对象的命名空间
func getObjectNamespace(obj runtime.Object) string {
	if meta, ok := obj.(metav1.Object); ok {
		return meta.GetNamespace()
	}
	return ""
}

// 辅助函数：处理创建错误（如资源已存在）
func handleCreateError(err error, resourceType, name string) error {
	if err == nil {
		return nil
	}
	if errors.IsAlreadyExists(err) {
		logs.Errorf("%s %s 已存在，跳过创建\n", resourceType, name)
		return nil // 或返回错误，根据需求决定
	}
	return fmt.Errorf("创建 %s %s 失败: %v", resourceType, name, err)
}

// 新增辅助函数：处理删除错误
func handleDeleteError(err error, resourceType, name string) error {
	if err == nil {
		return nil
	}
	if errors.IsNotFound(err) {
		logs.Warnf("%s %s 不存在，跳过删除", resourceType, name)
		return nil
	}
	return fmt.Errorf("删除 %s %s 失败: %v", resourceType, name, err)
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

//// 智能资源创建器
//func CreateResources(clientset *kubernetes.Clientset, objects []runtime.Object) error {
//	for _, obj := range objects {
//		switch resource := obj.(type) {
//		case *corev1.Pod:
//			if err := createPod(clientset, resource); err != nil {
//				return err
//			}
//		case *corev1.Service:
//			if err := createService(clientset, resource); err != nil {
//				return err
//			}
//		default:
//			logs.Warnf("不支持的资源类型: %T", obj)
//		}
//	}
//	return nil
//}
//
//// 私有化Pod创建逻辑
//func createPod(clientset *kubernetes.Clientset, pod *corev1.Pod) error {
//	_, err := clientset.CoreV1().Pods(pod.Namespace).Create(
//		context.TODO(),
//		pod,
//		metav1.CreateOptions{FieldValidation: "Strict"},
//	)
//	return err
//}
//
//// 服务创建逻辑
//func createService(clientset *kubernetes.Clientset, svc *corev1.Service) error {
//	_, err := clientset.CoreV1().Services(svc.Namespace).Create(
//		context.TODO(),
//		svc,
//		metav1.CreateOptions{FieldValidation: "Strict"},
//	)
//	return err
//}

package entity

import (
	apis "hit.edu/framework/pkg/apis/cores"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// Service 定义
type Service struct {
	Namespace   string               // Service 所在命名空间
	ServiceName string               // Service 名称
	Selector    map[string]string    // 用于选择对应的 Pod/Deployment
	Ports       []corev1.ServicePort // 服务端口
	ServiceType corev1.ServiceType   // 服务类型（ClusterIP, NodePort, LoadBalancer 等）
}

// 将 Runtime 的端口转换为 corev1.ServicePort 格式
func convertPorts(ports []apis.Port) []corev1.ServicePort {
	var servicePorts []corev1.ServicePort
	for _, port := range ports {
		servicePorts = append(servicePorts, corev1.ServicePort{
			Protocol: corev1.Protocol(port.Protocol),
			Port:     int32(port.Port),
			TargetPort: intstr.IntOrString{
				IntVal: int32(port.TargetPort),
			},
		})
	}
	return servicePorts
}

// 将 Runtime 的 ServiceType 字符串转换为 corev1.ServiceType
func convertServiceType(serviceType string) corev1.ServiceType {
	switch serviceType {
	case "ClusterIP":
		return corev1.ServiceTypeClusterIP
	case "NodePort":
		return corev1.ServiceTypeNodePort
	case "LoadBalancer":
		return corev1.ServiceTypeLoadBalancer
	default:
		return corev1.ServiceTypeClusterIP
	}
}

// NewService 创建 Service 的构造函数
func NewService(name, namespace string, selector map[string]string, ports []apis.Port, serviceType string) *Service {
	return &Service{
		Namespace:   namespace,
		ServiceName: name,
		Selector:    selector,
		Ports:       convertPorts(ports),
		ServiceType: convertServiceType(serviceType),
	}
}

// CreateServiceTemplate 使用 Service 对象的信息生成 Kubernetes Service 资源
func (s *Service) CreateServiceTemplate() corev1.Service {
	return corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      s.ServiceName,
			Namespace: s.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Selector: s.Selector,
			Ports:    s.Ports,
			Type:     s.ServiceType,
		},
	}
}

// CreateService 将 Service 资源提交到 Kubernetes
//func CreateService(clientset *kubernetes.Clientset, service *Service) error {
//	s := service.CreateServiceTemplate()
//	svc, err := clientset.CoreV1().Services(service.Namespace).Create(
//		context.TODO(),
//		&s,
//		metav1.CreateOptions{},
//	)
//	if err != nil {
//		logs.Error(err, "Failed to create service")
//		return err
//	}
//	logs.Info("Service created successfully", "Service:", svc.Name)
//	return nil
//}

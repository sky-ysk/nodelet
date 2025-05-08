package entity

//// Service 定义
//type Service struct {
//	Namespace   string               // Service 所在命名空间
//	ServiceName string               // Service 名称
//	Selector    map[string]string    // 用于选择对应的 Pod/Deployment
//	Ports       []corev1.ServicePort // 服务端口
//	ServiceType corev1.ServiceType   // 服务类型（ClusterIP, NodePort, LoadBalancer 等）
//}

//// 将 Runtime 的端口转换为 corev1.ServicePort 格式
//func convertPorts(ports []apis.Port) []corev1.ServicePort {
//	var servicePorts []corev1.ServicePort
//	for _, port := range ports {
//		servicePorts = append(servicePorts, corev1.ServicePort{
//			Protocol: corev1.Protocol(port.Protocol),
//			Port:     int32(port.Port),
//			TargetPort: intstr.IntOrString{
//				IntVal: int32(port.TargetPort),
//			},
//		})
//	}
//	return servicePorts
//}

//// 将 Runtime 的 ServiceType 字符串转换为 corev1.ServiceType
//func convertServiceType(serviceType string) corev1.ServiceType {
//	switch serviceType {
//	case "ClusterIP":
//		return corev1.ServiceTypeClusterIP
//	case "NodePort":
//		return corev1.ServiceTypeNodePort
//	case "LoadBalancer":
//		return corev1.ServiceTypeLoadBalancer
//	default:
//		return corev1.ServiceTypeClusterIP
//	}
//}

/*
func GetServiceFromParam1(customService *apis.Service) *corev1.Service {
	labels := customService.Labels
	if labels == nil {
		labels = make(map[string]string)
	}
	labels[monitor.CreateorLabel] = monitor.SystemName
	service := &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       customService.Kind,
			APIVersion: customService.APIVersion,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      customService.Name,
			Namespace: customService.Namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Type:            corev1.ServiceType(customService.Spec.Type),
			Selector:        customService.Spec.Selector,
			SessionAffinity: corev1.ServiceAffinity(customService.Spec.SessionAffinity),
			ClusterIP:       customService.Spec.ClusterIP,
			ExternalIPs:     customService.Spec.ExternalIPs,
			ExternalName:    customService.Spec.ExternalName,
		},
	}
	// 端口列表转换（至少一个端口）
	var ports []corev1.ServicePort
	for _, p := range customService.Spec.Ports {
		// 处理TargetPort的IntOrString类型
		targetPort := intstr.IntOrString{}
		if p.TargetPort.Type == apis.Int {
			targetPort = intstr.FromInt32(p.TargetPort.IntVal)
		} else {
			targetPort = intstr.FromString(p.TargetPort.StrVal)
		}

		ports = append(ports, corev1.ServicePort{
			Name:       p.Name,
			Protocol:   corev1.Protocol(p.Protocol),
			Port:       p.Port,
			TargetPort: targetPort,
			NodePort:   p.NodePort, // 当Type=NodePort/LoadBalancer时有效
		})
	}
	service.Spec.Ports = ports

	// 特殊字段处理
	if customService.Spec.ExternalTrafficPolicy != "" {
		service.Spec.ExternalTrafficPolicy = corev1.ServiceExternalTrafficPolicyType(
			customService.Spec.ExternalTrafficPolicy)
	}

	return service
}
*/

//func NewService(name, namespace string, selector map[string]string, ports []apis.Port, serviceType string) *Service {
//	return &Service{
//		Namespace:   namespace,
//		ServiceName: name,
//		Selector:    selector,
//		Ports:       convertPorts(ports),
//		ServiceType: convertServiceType(serviceType),
//	}
//}
//
//func (s *Service) CreateServiceTemplate() corev1.Service {
//	return corev1.Service{
//		ObjectMeta: metav1.ObjectMeta{
//			Name:      s.ServiceName,
//			Namespace: s.Namespace,
//		},
//		Spec: corev1.ServiceSpec{
//			Selector: s.Selector,
//			Ports:    s.Ports,
//			Type:     s.ServiceType,
//		},
//	}
//}

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

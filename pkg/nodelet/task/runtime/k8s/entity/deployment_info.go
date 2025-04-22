package entity

// Deployment 定义
//type Deployment struct {
//	Namespace          string                      // Deployment 所在命名空间
//	DeploymentName     string                      // Deployment 名称
//	Labels             map[string]string           // Deployment 的标签
//	Replicas           int32                       // 副本数
//	ContainerName      string                      // 容器名称
//	Image              string                      // 容器镜像
//	Ports              []corev1.ContainerPort      // 容器端口
//	EnvVars            []corev1.EnvVar             // 环境变量
//	Resources          corev1.ResourceRequirements // 资源限制和请求
//	VolumeMounts       []corev1.VolumeMount        // 挂载卷
//	Selector           map[string]string           // 选择器，用于匹配Pod
//	TaskNeedMonitoring bool                        // 是否需要监控
//}

// 创建 Deployment 对象的构造函数
//func NewDeployment(name, namespace string, labels map[string]string, replicas int32, containerName string, image string, selector map[string]string, taskNeedMonitoring bool) *Deployment {
//	return &Deployment{
//		Namespace:          namespace,
//		DeploymentName:     name,
//		Labels:             labels,
//		Image:              image,
//		ContainerName:      containerName,
//		Replicas:           replicas,
//		Selector:           selector,
//		TaskNeedMonitoring: taskNeedMonitoring,
//	}
//}

/*
func GetDeploymentFromParam1(customDeployment *apis.Deployment) *appsv1.Deployment {
	spec := convertDeploymentSpec(&customDeployment.Spec)
	labels := customDeployment.Labels
	if labels == nil {
		labels = make(map[string]string)
	}
	labels[monitor.CreateorLabel] = monitor.SystemName
	deployment := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: customDeployment.APIVersion,
			Kind:       customDeployment.Kind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      customDeployment.Name,
			Namespace: customDeployment.Namespace,
			Labels:    labels,
		},
		Spec: spec,
	}
	return deployment
}

func convertDeploymentSpec(customSpec *apis.DeploymentSpec) appsv1.DeploymentSpec {
	// 处理默认值
	replicas := int32(1)
	if customSpec.Replicas > 1 {
		replicas = customSpec.Replicas
	}
	return appsv1.DeploymentSpec{
		Replicas: &replicas,
		Selector: &metav1.LabelSelector{ // TODO 少一个参数
			MatchLabels:      customSpec.Selector.MatchLabels,
			MatchExpressions: convertLabelSelectorRequirements(customSpec.Selector.MatchExpressions),
		},
		Template:        convertPodTemplateSpec(customSpec.Template),
		Strategy:        convertDeploymentStrategy(customSpec.Strategy),
		MinReadySeconds: customSpec.MinReadySeconds,
		RevisionHistoryLimit: func() *int32 {
			if customSpec.RevisionHistoryLimit != 0 {
				return &customSpec.RevisionHistoryLimit
			}
			return nil
		}(),
		ProgressDeadlineSeconds: func() *int32 {
			if customSpec.ProgressDeadlineSeconds != 0 {
				return &customSpec.ProgressDeadlineSeconds
			}
			return nil
		}(),
	}
}
func convertLabelSelectorRequirements(exprs []meta.LabelSelectorRequirement) []metav1.LabelSelectorRequirement {
	var k8sExprs []metav1.LabelSelectorRequirement
	for _, e := range exprs {
		k8sExprs = append(k8sExprs, metav1.LabelSelectorRequirement{
			Key:      e.Key,
			Operator: metav1.LabelSelectorOperator(e.Operator),
			Values:   e.Values,
		})
	}
	return k8sExprs
}
func convertPodTemplateSpec(template apis.PodTemplateSpec) corev1.PodTemplateSpec {
	return corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels: template.ObjectMeta.Labels,
		},
		Spec: convertPodSpec(template.Spec), // 复用之前实现的 PodSpec 转换函数
	}
}
func convertDeploymentStrategy(strategy apis.DeploymentStrategy) appsv1.DeploymentStrategy {
	// 设置默认策略类型为RollingUpdate
	strategyType := appsv1.RollingUpdateDeploymentStrategyType
	if strategy.Type == apis.RecreateDeploymentStrategyType {
		strategyType = appsv1.RecreateDeploymentStrategyType
	}
	return appsv1.DeploymentStrategy{
		Type:          strategyType,
		RollingUpdate: convertRollingUpdate(strategy.RollingUpdate),
	}
}
func convertRollingUpdate(rolling apis.RollingUpdateDeployment) *appsv1.RollingUpdateDeployment {
	// 处理滚动更新默认值（Kubernetes默认值为25%）
	maxSurge := intstr.FromString("25%")
	if rolling.MaxSurge.Type == apis.String && rolling.MaxSurge.StrVal != "" {
		maxSurge = convertIntOrString(rolling.MaxSurge)
	} else if rolling.MaxSurge.Type == apis.Int && rolling.MaxSurge.IntVal != 0 {
		maxSurge = convertIntOrString(rolling.MaxSurge)
	}
	maxUnavailable := intstr.FromString("25%")
	// 同理处理MaxUnavailable
	if rolling.MaxUnavailable.Type == apis.String && rolling.MaxUnavailable.StrVal != "" {
		maxUnavailable = convertIntOrString(rolling.MaxUnavailable)
	} else if rolling.MaxUnavailable.Type == apis.Int && rolling.MaxUnavailable.IntVal != 0 {
		maxUnavailable = convertIntOrString(rolling.MaxUnavailable)
	}

	return &appsv1.RollingUpdateDeployment{
		MaxSurge:       &maxSurge,
		MaxUnavailable: &maxUnavailable,
	}
}
func convertIntOrString(input apis.IntOrString) intstr.IntOrString {
	switch input.Type {
	case apis.Int:
		return intstr.FromInt32(input.IntVal)
	case apis.String:
		return intstr.FromString(input.StrVal)
	default:
		return intstr.FromInt(0)
	}
}
*/

//// CreateDeploymentTemplate 使用 Deployment 对象信息生成 Kubernetes Deployment 资源
//func (d *Deployment) CreateDeploymentTemplate() appsv1.Deployment {
//	deploymentSpec := appsv1.DeploymentSpec{
//		Replicas: &d.Replicas,
//		Selector: &metav1.LabelSelector{
//			MatchLabels: d.Selector,
//		},
//		Template: corev1.PodTemplateSpec{
//			ObjectMeta: metav1.ObjectMeta{
//				Labels: d.Selector,
//			},
//			Spec: corev1.PodSpec{
//				Containers: []corev1.Container{
//					{
//						Name:         d.ContainerName,
//						Image:        d.Image,
//						Ports:        d.Ports,
//						Env:          d.EnvVars,
//						Resources:    d.Resources,
//						VolumeMounts: d.VolumeMounts,
//					},
//				},
//			},
//		},
//	}
//
//	if d.TaskNeedMonitoring {
//		sidecarContainer := createTaskExporterSidecar()
//		deploymentSpec.Template.Spec.Containers = append(deploymentSpec.Template.Spec.Containers, sidecarContainer)
//	}
//
//	return appsv1.Deployment{
//		ObjectMeta: metav1.ObjectMeta{
//			Name:      d.DeploymentName,
//			Namespace: d.Namespace,
//			Labels:    d.Labels,
//		},
//		Spec: deploymentSpec,
//	}
//}

//// CreateDeployment 负责将 Deployment 资源提交到 Kubernetes
//func CreateDeployment(clientset *kubernetes.Clientset, deployment *Deployment) error {
//	deploy, err := clientset.AppsV1().Deployments(deployment.Namespace).Create(
//		context.TODO(),
//		&CreateDeploymentTemplate(),
//		metav1.CreateOptions{},
//	)
//	if err != nil {
//		logs.Error(err, "Failed to create deployment")
//		return err
//	}
//	logs.Info("Deployment created successfully", "Deployment:", deploy.Name)
//	return nil
//}

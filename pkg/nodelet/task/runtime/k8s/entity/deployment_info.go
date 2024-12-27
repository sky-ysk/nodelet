package entity

import (
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Deployment 定义
type Deployment struct {
	Namespace          string                      // Deployment 所在命名空间
	DeploymentName     string                      // Deployment 名称
	Labels             map[string]string           // Deployment 的标签
	Replicas           int32                       // 副本数
	ContainerName      string                      // 容器名称
	Image              string                      // 容器镜像
	Ports              []corev1.ContainerPort      // 容器端口
	EnvVars            []corev1.EnvVar             // 环境变量
	Resources          corev1.ResourceRequirements // 资源限制和请求
	VolumeMounts       []corev1.VolumeMount        // 挂载卷
	Selector           map[string]string           // 选择器，用于匹配Pod
	TaskNeedMonitoring bool                        // 是否需要监控
}

// 创建 Deployment 对象的构造函数
func NewDeployment(name, namespace string, labels map[string]string, replicas int32, containerName string, image string, selector map[string]string, taskNeedMonitoring bool) *Deployment {
	return &Deployment{
		Namespace:          namespace,
		DeploymentName:     name,
		Labels:             labels,
		Image:              image,
		ContainerName:      containerName,
		Replicas:           replicas,
		Selector:           selector,
		TaskNeedMonitoring: taskNeedMonitoring,
	}
}

// CreateDeploymentTemplate 使用 Deployment 对象信息生成 Kubernetes Deployment 资源
func (d *Deployment) CreateDeploymentTemplate() v1.Deployment {
	deploymentSpec := v1.DeploymentSpec{
		Replicas: &d.Replicas,
		Selector: &metav1.LabelSelector{
			MatchLabels: d.Selector,
		},
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: d.Selector,
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:         d.ContainerName,
						Image:        d.Image,
						Ports:        d.Ports,
						Env:          d.EnvVars,
						Resources:    d.Resources,
						VolumeMounts: d.VolumeMounts,
					},
				},
			},
		},
	}

	if d.TaskNeedMonitoring {
		sidecarContainer := createTaskExporterSidecar()
		deploymentSpec.Template.Spec.Containers = append(deploymentSpec.Template.Spec.Containers, sidecarContainer)
	}

	return v1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      d.DeploymentName,
			Namespace: d.Namespace,
			Labels:    d.Labels,
		},
		Spec: deploymentSpec,
	}
}

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

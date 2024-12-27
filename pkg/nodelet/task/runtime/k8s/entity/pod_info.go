package entity

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	TaskExporterName  = "task-exporter-sidecar"
	TaskExporterImage = "task-exporter-image:latest" //后期换成镜像仓库所在地址
)

type Pod struct {
	Namespace          string                  //Pod所在名称空间
	PodName            string                  //Pod名称
	ContainerName      string                  //Pod内的容器名称
	Image              string                  //Pod内的容器任务镜像   TODO：如果有多个任务镜像还需部署多个任务
	Command            []string                // 启动容器时执行的命令
	Args               []string                // 命令的参数
	WorkingDir         string                  // 容器内的工作目录
	Ports              []v1.ContainerPort      // 容器的端口配置
	EnvVars            []v1.EnvVar             //环境变量
	Resources          v1.ResourceRequirements // 资源限制和请求
	VolumeMounts       []v1.VolumeMount        // 挂载卷
	LivenessProbe      *v1.Probe               // 存活探针
	ReadinessProbe     *v1.Probe               // 就绪探针
	LifeCycle          *v1.Lifecycle           // 生命周期钩子
	SecurityContext    *v1.PodSecurityContext  // 安全上下文
	Volumes            []v1.Volume             // 定义的卷
	RestartPolicy      v1.RestartPolicy        // 重启策略 (Always, OnFailure, Never)
	TaskNeedMonitoring bool                    //是否需要注入Taskexporter

	YamlPath string //Pod的yaml配置文件
}

// NewPod 是一个创建 Pod 对象的构造函数
func NewPodFromParam(podName, namespace, containerName, image string, taskNeedMonitoring bool) *Pod {
	return &Pod{
		Namespace:          namespace,
		PodName:            podName,
		Image:              image,
		ContainerName:      containerName,
		TaskNeedMonitoring: taskNeedMonitoring,
	}
}

// CreatePodTemplate 使用 Pod 对象的信息生成 Kubernetes Pod 资源
func (p *Pod) CreatePodTemplate() (v1.Pod, error) {
	podSpec := v1.PodSpec{
		Containers: []v1.Container{
			{
				Name:      p.ContainerName, //对应的是容器的名称
				Image:     p.Image,
				Env:       p.EnvVars,
				Resources: p.Resources,
			},
		},
	}
	// 如果需要监控，则注入 sidecar 容器
	if p.TaskNeedMonitoring {
		sidecarContainer := createTaskExporterSidecar()
		podSpec.Containers = append(podSpec.Containers, sidecarContainer)
	}

	return v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      p.PodName,
			Namespace: p.Namespace,
		},
		Spec: podSpec,
	}, nil
}
func createTaskExporterSidecar() v1.Container {
	return v1.Container{
		Name:  TaskExporterName,
		Image: TaskExporterImage,
		//Args:  []string{"--monitor"}, //./task-exporter --monitor--目前应该不需要，因为启动这个Taskexporter，自动开启监控协程 如果你的 TaskExporter 接受更多参数，或者参数可以根据任务配置不同，你可以考虑将 args 作为可传递参数
		//这里我觉得应该将别的参数，执行类似Taskexporter的assigner方法，将任务加入到Taskexporter的队列当中
	}
}

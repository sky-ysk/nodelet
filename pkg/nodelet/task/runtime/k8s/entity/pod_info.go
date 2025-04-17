package entity

import (
	"hit.edu/framework/pkg/nodelet/task/runtime/k8s/monitor"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
)

// const (
//
//	TaskExporterName  = "task-exporter-sidecar"
//	TaskExporterImage = "task-exporter-image:latest" //后期换成镜像仓库所在地址
//
// )
func GetPodFromYAML(yamlContent []byte, nodeName string) (*corev1.Pod, error) {
	pod := &corev1.Pod{}
	if err := yaml.Unmarshal(yamlContent, pod); err != nil {
		return nil, err
	}
	// 添加系统标签（保留原有逻辑）
	if pod.Labels == nil {
		pod.Labels = make(map[string]string)
	}
	pod.Labels[monitor.CreateorLabel] = nodeName
	return pod, nil
}

//type Pod struct {
//	Namespace          string                      //Pod所在名称空间
//	PodName            string                      //Pod名称
//	ContainerName      string                      //Pod内的容器名称
//	Image              string                      //Pod内的容器任务镜像   TODO：如果有多个任务镜像还需部署多个任务
//	Command            []string                    // 启动容器时执行的命令
//	Args               []string                    // 命令的参数
//	WorkingDir         string                      // 容器内的工作目录
//	Ports              []corev1.ContainerPort      // 容器的端口配置
//	EnvVars            []corev1.EnvVar             //环境变量
//	Resources          corev1.ResourceRequirements // 资源限制和请求
//	VolumeMounts       []corev1.VolumeMount        // 挂载卷
//	LivenessProbe      *corev1.Probe               // 存活探针
//	ReadinessProbe     *corev1.Probe               // 就绪探针
//	LifeCycle          *corev1.Lifecycle           // 生命周期钩子
//	SecurityContext    *corev1.PodSecurityContext  // 安全上下文
//	Volumes            []corev1.Volume             // 定义的卷
//	RestartPolicy      corev1.RestartPolicy        // 重启策略 (Always, OnFailure, Never)
//	TaskNeedMonitoring bool                        //是否需要注入Taskexporter
//
//	YamlPath string //Pod的yaml配置文件
//}

// // NewPod 是一个创建 Pod 对象的构造函数
//
//	func NewPodFromParam(podName, namespace, containerName, image string, taskNeedMonitoring bool) *Pod {
//		return &Pod{
//			Namespace:          namespace,
//			PodName:            podName,
//			Image:              image,
//			ContainerName:      containerName,
//			TaskNeedMonitoring: taskNeedMonitoring,
//		}
//	}
//func GetPodFromParam1(customPod *apis.Pod) *corev1.Pod {
//	spec := convertPodSpec(customPod.Spec)
//	labels := customPod.ObjectMeta.Labels
//	if labels == nil {
//		labels = make(map[string]string)
//	}
//	labels[monitor.CreateorLabel] = monitor.SystemName
//	pod := &corev1.Pod{
//		TypeMeta: metav1.TypeMeta{
//			APIVersion: customPod.TypeMeta.APIVersion,
//			Kind:       customPod.TypeMeta.Kind,
//		},
//		ObjectMeta: metav1.ObjectMeta{
//			Name:      customPod.ObjectMeta.Name,
//			Namespace: customPod.ObjectMeta.Namespace,
//			Labels:    labels,
//		},
//		Spec: spec,
//	}
//	return pod
//}
//func convertPodSpec(spec apis.PodSpec) corev1.PodSpec {
//	return corev1.PodSpec{
//		Volumes:            convertVolumes(spec.Volumes),
//		InitContainers:     convertContainers(spec.InitContainers),
//		Containers:         convertContainers(spec.Containers),
//		RestartPolicy:      corev1.RestartPolicy(spec.RestartPolicy),
//		DNSPolicy:          corev1.DNSPolicy(spec.DnsPolicy),
//		NodeSelector:       spec.NodeSelector,
//		ServiceAccountName: spec.ServiceAccountName,
//		Affinity:           convertAffinity(spec.Affinity),
//	}
//}
//
//// ----------------------------[ Affinity 转换 ]----------------------------
//func convertAffinity(affinity apis.Affinity) *corev1.Affinity {
//	if isEmptyAffinity(affinity) {
//		return nil
//	}
//
//	return &corev1.Affinity{
//		NodeAffinity:    convertNodeAffinity(affinity.NodeAffinity),
//		PodAffinity:     convertPodAffinity(affinity.PodAffinity),
//		PodAntiAffinity: convertPodAntiAffinity(affinity.PodAntiAffinity),
//	}
//}
//
//// ----------------------------[ 空值判断逻辑 ]----------------------------
//// 总空值判断
//func isEmptyAffinity(a apis.Affinity) bool {
//	return isEmptyNodeAffinity(a.NodeAffinity) &&
//		isEmptyPodAffinity(a.PodAffinity) &&
//		isEmptyPodAntiAffinity(a.PodAntiAffinity)
//}
//
//// NodeAffinity 空值判断
//func isEmptyNodeAffinity(na apis.NodeAffinity) bool {
//	return isEmptyNodeSelector(na.RequiredDuringSchedulingIgnoredDuringExecution) &&
//		len(na.PreferredDuringSchedulingIgnoredDuringExecution) == 0
//}
//
//// NodeSelector 空值判断
//func isEmptyNodeSelector(ns apis.NodeSelector) bool {
//	return len(ns.NodeSelectorTerms) == 0
//}
//
//// PodAffinity/PodAntiAffinity 空值判断（根据你的实际结构补充）
//func isEmptyPodAffinity(pa apis.PodAffinity) bool {
//	// 示例：根据实际字段判断
//	return len(pa.RequiredDuringSchedulingIgnoredDuringExecution) == 0 &&
//		len(pa.PreferredDuringSchedulingIgnoredDuringExecution) == 0
//}
//
//func isEmptyPodAntiAffinity(paa apis.PodAntiAffinity) bool {
//	// 示例：同上
//	return len(paa.RequiredDuringSchedulingIgnoredDuringExecution) == 0 &&
//		len(paa.PreferredDuringSchedulingIgnoredDuringExecution) == 0
//}
//
//// ----------------------------[ NodeAffinity 转换 ]----------------------------
//func convertNodeAffinity(src apis.NodeAffinity) *corev1.NodeAffinity {
//	if isEmptyNodeAffinity(src) {
//		return nil
//	}
//
//	dst := &corev1.NodeAffinity{}
//
//	// 转换 RequiredDuringScheduling...
//	if !isEmptyNodeSelector(src.RequiredDuringSchedulingIgnoredDuringExecution) {
//		dst.RequiredDuringSchedulingIgnoredDuringExecution = convertNodeSelector(src.RequiredDuringSchedulingIgnoredDuringExecution)
//	}
//
//	// 转换 PreferredDuringScheduling...
//	if len(src.PreferredDuringSchedulingIgnoredDuringExecution) > 0 {
//		dst.PreferredDuringSchedulingIgnoredDuringExecution = convertPreferredSchedulingTerms(src.PreferredDuringSchedulingIgnoredDuringExecution)
//	}
//
//	return dst
//}
//
//func convertNodeSelector(src apis.NodeSelector) *corev1.NodeSelector {
//	if isEmptyNodeSelector(src) {
//		return nil
//	}
//
//	return &corev1.NodeSelector{
//		NodeSelectorTerms: convertNodeSelectorTerms(src.NodeSelectorTerms),
//	}
//}
//
//// ----------------------------[ NodeSelectorTerm 转换 ]----------------------------
//func convertNodeSelectorTerms(src []apis.NodeSelectorTerm) []corev1.NodeSelectorTerm {
//	var dst []corev1.NodeSelectorTerm
//	for _, term := range src {
//		if converted := convertNodeSelectorTerm(term); converted != nil {
//			dst = append(dst, *converted)
//		}
//	}
//	return dst
//}
//
//func convertNodeSelectorTerm(src apis.NodeSelectorTerm) *corev1.NodeSelectorTerm {
//	if isEmptyNodeSelectorTerm(src) {
//		return nil
//	}
//
//	return &corev1.NodeSelectorTerm{
//		MatchExpressions: convertNodeSelectorRequirements(src.MatchExpressions),
//		MatchFields:      convertNodeSelectorRequirements(src.MatchFields),
//	}
//}
//
//func isEmptyNodeSelectorTerm(nst apis.NodeSelectorTerm) bool {
//	return len(nst.MatchExpressions) == 0 && len(nst.MatchFields) == 0
//}
//
//// ----------------------------[ NodeSelectorRequirement 转换 ]----------------------------
//func convertNodeSelectorRequirements(src []apis.NodeSelectorRequirement) []corev1.NodeSelectorRequirement {
//	var dst []corev1.NodeSelectorRequirement
//	for _, req := range src {
//		dst = append(dst, corev1.NodeSelectorRequirement{
//			Key:      req.Key,
//			Operator: corev1.NodeSelectorOperator(req.Operator),
//			Values:   req.Values,
//		})
//	}
//	return dst
//}
//
//// ----------------------------[ PreferredSchedulingTerm 转换 ]----------------------------
//func convertPreferredSchedulingTerms(src []apis.PreferredSchedulingTerm) []corev1.PreferredSchedulingTerm {
//	var dst []corev1.PreferredSchedulingTerm
//	for _, term := range src {
//		if converted := convertPreferredSchedulingTerm(term); converted != nil {
//			dst = append(dst, *converted)
//		}
//	}
//	return dst
//}
//
//func convertPreferredSchedulingTerm(src apis.PreferredSchedulingTerm) *corev1.PreferredSchedulingTerm {
//	if src.Weight == 0 || isEmptyNodeSelectorTerm(src.Preference) {
//		return nil
//	}
//
//	return &corev1.PreferredSchedulingTerm{
//		Weight:     src.Weight,
//		Preference: *convertNodeSelectorTerm(src.Preference),
//	}
//}
//
//// ----------------------------[ PodAffinity 转换 ]----------------------------
//func convertPodAffinity(pa apis.PodAffinity) *corev1.PodAffinity {
//	if isEmptyPodAffinity(pa) {
//		return nil
//	}
//
//	return &corev1.PodAffinity{
//		RequiredDuringSchedulingIgnoredDuringExecution:  convertPodAffinityTerms(pa.RequiredDuringSchedulingIgnoredDuringExecution),
//		PreferredDuringSchedulingIgnoredDuringExecution: convertWeightedPodAffinityTerms(pa.PreferredDuringSchedulingIgnoredDuringExecution),
//	}
//}
//
//// ----------------------------[ PodAntiAffinity 转换 ]----------------------------
//func convertPodAntiAffinity(paa apis.PodAntiAffinity) *corev1.PodAntiAffinity {
//	if isEmptyPodAntiAffinity(paa) {
//		return nil
//	}
//
//	return &corev1.PodAntiAffinity{
//		RequiredDuringSchedulingIgnoredDuringExecution:  convertPodAffinityTerms(paa.RequiredDuringSchedulingIgnoredDuringExecution),
//		PreferredDuringSchedulingIgnoredDuringExecution: convertWeightedPodAffinityTerms(paa.PreferredDuringSchedulingIgnoredDuringExecution),
//	}
//}
//
//// ----------------------------[ 通用转换工具函数 ]----------------------------
//// 转换 PodAffinityTerm 列表
//func convertPodAffinityTerms(terms []apis.PodAffinityTerm) []corev1.PodAffinityTerm {
//	var dst []corev1.PodAffinityTerm
//	for _, term := range terms {
//		if converted := convertPodAffinityTerm(term); converted != nil {
//			dst = append(dst, *converted)
//		}
//	}
//	return dst
//}
//
//// 转换单个 PodAffinityTerm
//func convertPodAffinityTerm(term apis.PodAffinityTerm) *corev1.PodAffinityTerm {
//	if isEmptyPodAffinityTerm(term) {
//		return nil
//	}
//
//	return &corev1.PodAffinityTerm{
//		LabelSelector:     convertLabelSelector(term.LabelSelector),
//		Namespaces:        term.Namespaces,
//		TopologyKey:       term.TopologyKey,
//		NamespaceSelector: convertLabelSelector(term.NamespaceSelector),
//	}
//}
//
//// PodAffinityTerm 空值判断
//func isEmptyPodAffinityTerm(term apis.PodAffinityTerm) bool {
//	return term.TopologyKey == "" && // TopologyKey 是必填字段
//		convertLabelSelector(term.LabelSelector) == nil &&
//		convertLabelSelector(term.NamespaceSelector) == nil &&
//		len(term.Namespaces) == 0
//}
//
//// 转换带权重的 PodAffinityTerm
//func convertWeightedPodAffinityTerms(terms []apis.WeightedPodAffinityTerm) []corev1.WeightedPodAffinityTerm {
//	var dst []corev1.WeightedPodAffinityTerm
//	for _, term := range terms {
//		if converted := convertWeightedPodAffinityTerm(term); converted != nil {
//			dst = append(dst, *converted)
//		}
//	}
//	return dst
//}
//
//func convertWeightedPodAffinityTerm(term apis.WeightedPodAffinityTerm) *corev1.WeightedPodAffinityTerm {
//	if term.Weight == 0 || isEmptyPodAffinityTerm(term.PodAffinityTerm) {
//		return nil
//	}
//
//	return &corev1.WeightedPodAffinityTerm{
//		Weight:          term.Weight,
//		PodAffinityTerm: *convertPodAffinityTerm(term.PodAffinityTerm),
//	}
//}
//
//// ----------------------------[ LabelSelector 转换 ]----------------------------
//func convertLabelSelector(selector meta.LabelSelector) *metav1.LabelSelector {
//	if selector.MatchLabels == nil && len(selector.MatchExpressions) == 0 {
//		return nil
//	}
//
//	return &metav1.LabelSelector{
//		MatchLabels:      selector.MatchLabels,
//		MatchExpressions: convertLabelSelectorRequirements(selector.MatchExpressions),
//	}
//}
//
////func convertLabelSelectorRequirements(reqs []meta.LabelSelectorRequirement) []metav1.LabelSelectorRequirement {
////	var dst []metav1.LabelSelectorRequirement
////	for _, req := range reqs {
////		dst = append(dst, metav1.LabelSelectorRequirement{
////			Key:      req.Key,
////			Operator: metav1.LabelSelectorOperator(req.Operator),
////			Values:   req.Values,
////		})
////	}
////	return dst
////}
//
//func convertVolumes(volumes []apis.Volume) []corev1.Volume {
//	var k8sVolumes []corev1.Volume
//	for _, v := range volumes {
//		kv := corev1.Volume{
//			Name: v.Name,
//			VolumeSource: corev1.VolumeSource{
//				ConfigMap:             convertConfigMapSource(v.ConfigMap),
//				Secret:                convertSecretSource(v.Secret),
//				HostPath:              convertHostPathSource(v.HostPath),
//				EmptyDir:              convertEmptyDirSource(v.EmptyDir),
//				PersistentVolumeClaim: convertPersistentVolumeClaimSource(v.PersistentVolumeClaim),
//			},
//		}
//		k8sVolumes = append(k8sVolumes, kv)
//	}
//	return k8sVolumes
//}
//
//func convertConfigMapSource(configMap apis.ConfigMapVolumeSource) *corev1.ConfigMapVolumeSource {
//	if configMap.Name == "" {
//		return nil
//	}
//	return &corev1.ConfigMapVolumeSource{
//		LocalObjectReference: corev1.LocalObjectReference{
//			Name: configMap.Name,
//		},
//		Items: convertKeyToPaths(configMap.Items),
//	}
//}
//func convertKeyToPaths(items []apis.KeyToPath) []corev1.KeyToPath {
//	var k8sItems []corev1.KeyToPath
//	for _, item := range items {
//		k8sItems = append(k8sItems, corev1.KeyToPath{
//			Key:  item.Key,
//			Path: item.Path,
//		})
//	}
//	return k8sItems
//}
//
//func convertSecretSource(secret apis.SecretVolumeSource) *corev1.SecretVolumeSource {
//	if secret.SecretName == "" {
//		return nil
//	}
//	return &corev1.SecretVolumeSource{
//		SecretName: secret.SecretName,
//		Items:      convertKeyToPaths(secret.Items),
//	}
//}
//
//func convertHostPathSource(hostPath apis.HostPathVolumeSource) *corev1.HostPathVolumeSource {
//	if hostPath.Path == "" {
//		return nil
//	}
//	return &corev1.HostPathVolumeSource{
//		Path: hostPath.Path,
//		Type: convertHostPathType(hostPath.Type),
//	}
//}
//func convertHostPathType(hpType apis.HostPathType) *corev1.HostPathType {
//	var k8sType corev1.HostPathType
//	switch hpType {
//	case apis.HostPathDirectoryOrCreate:
//		k8sType = corev1.HostPathDirectoryOrCreate
//	case apis.HostPathDirectory:
//		k8sType = corev1.HostPathDirectory
//	case apis.HostPathFileOrCreate:
//		k8sType = corev1.HostPathFileOrCreate
//	case apis.HostPathFile:
//		k8sType = corev1.HostPathFile
//	default:
//		return nil
//	}
//	return &k8sType
//}
//
//func convertEmptyDirSource(emptyDir apis.EmptyDirVolumeSource) *corev1.EmptyDirVolumeSource {
//	medium := convertStorageMedium(emptyDir.Medium)
//	return &corev1.EmptyDirVolumeSource{
//		Medium: medium,
//	}
//}
//func convertStorageMedium(medium apis.StorageMedium) corev1.StorageMedium {
//	switch medium {
//	case apis.StorageMediumMemory:
//		return corev1.StorageMediumMemory
//	default:
//		return corev1.StorageMediumDefault
//	}
//}
//
//func convertPersistentVolumeClaimSource(persistentVolumeClaim apis.PersistentVolumeClaimVolumeSource) *corev1.PersistentVolumeClaimVolumeSource {
//	if persistentVolumeClaim.ClaimName == "" {
//		return nil
//	}
//	return &corev1.PersistentVolumeClaimVolumeSource{
//		ClaimName: persistentVolumeClaim.ClaimName,
//	}
//}
//
//func convertContainers(containers []apis.Container) []corev1.Container {
//	var k8sContainers []corev1.Container
//	for _, container := range containers {
//		kc := corev1.Container{
//			Name:         container.Name,
//			Image:        container.Image,
//			Command:      container.Command,
//			Args:         container.Args,
//			Ports:        convertContainerPorts(container.Ports),
//			Env:          convertEnvVars(container.Env),
//			VolumeMounts: convertVolumeMounts(container.VolumeMounts),
//		}
//		k8sContainers = append(k8sContainers, kc)
//	}
//	return k8sContainers
//}
//func convertContainerPorts(ports []apis.ContainerPort) []corev1.ContainerPort {
//	var k8sPorts []corev1.ContainerPort
//	for _, port := range ports {
//		k8sPorts = append(k8sPorts, corev1.ContainerPort{
//			ContainerPort: int32(port.ContainerPort),
//			Protocol:      convertProtocol(port.Protocol),
//		})
//	}
//	return k8sPorts
//}
//func convertProtocol(protocol apis.Protocol) corev1.Protocol {
//	switch protocol {
//	case apis.ProtocolTCP:
//		return corev1.ProtocolTCP
//	case apis.ProtocolUDP:
//		return corev1.ProtocolUDP
//	default:
//		return corev1.ProtocolSCTP
//	}
//}
//func convertEnvVars(env []apis.EnvVar) []corev1.EnvVar {
//	var k8sEnvVars []corev1.EnvVar
//	for _, envVar := range env {
//		k8sEnvVars = append(k8sEnvVars, corev1.EnvVar{
//			Name:  envVar.Name,
//			Value: envVar.Value,
//		})
//	}
//	return k8sEnvVars
//}
//func convertVolumeMounts(volumeMounts []apis.VolumeMount) []corev1.VolumeMount {
//	var k8sVolumeMounts []corev1.VolumeMount
//	for _, volumeMount := range volumeMounts {
//		k8sVolumeMounts = append(k8sVolumeMounts, corev1.VolumeMount{
//			Name:      volumeMount.Name,
//			MountPath: volumeMount.MountPath,
//			ReadOnly:  volumeMount.ReadOnly,
//		})
//	}
//	return k8sVolumeMounts
//}

//// CreatePodTemplate 使用 Pod 对象的信息生成 Kubernetes Pod 资源
//func (p *Pod) CreatePodTemplate() (corev1.Pod, error) {
//	podSpec := corev1.PodSpec{
//		Containers: []corev1.Container{
//			{
//				Name:      p.ContainerName, //对应的是容器的名称
//				Image:     p.Image,
//				Env:       p.EnvVars,
//				Resources: p.Resources,
//			},
//		},
//	}
//	// 如果需要监控，则注入 sidecar 容器
//	if p.TaskNeedMonitoring {
//		sidecarContainer := createTaskExporterSidecar()
//		podSpec.Containers = append(podSpec.Containers, sidecarContainer)
//	}
//
//	return corev1.Pod{
//		ObjectMeta: metav1.ObjectMeta{
//			Name:      p.PodName,
//			Namespace: p.Namespace,
//		},
//		Spec: podSpec,
//	}, nil
//}
//func createTaskExporterSidecar() corev1.Container {
//	return corev1.Container{
//		Name:  TaskExporterName,
//		Image: TaskExporterImage,
//		//Args:  []string{"--monitor"}, //./task-exporter --monitor--目前应该不需要，因为启动这个Taskexporter，自动开启监控协程 如果你的 TaskExporter 接受更多参数，或者参数可以根据任务配置不同，你可以考虑将 args 作为可传递参数
//		//这里我觉得应该将别的参数，执行类似Taskexporter的assigner方法，将任务加入到Taskexporter的队列当中
//	}
//}

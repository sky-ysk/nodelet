package entity

import (
	"bytes"
	"fmt"
	"hit.edu/framework/pkg/nodelet/task/runtime/k8s/monitor"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"os"
	"path/filepath"
	"strings"
)

func ParseK8sResourcesFromFile(filePath string, nodeName string, randomNum int32) ([]runtime.Object, error) {
	if ext := filepath.Ext(filePath); ext != ".yaml" && ext != ".yml" {
		return nil, fmt.Errorf("invalid file type: %s", ext)
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("file read error: %v", err)
	}

	return parseK8sResources(content, nodeName, randomNum)
}

func parseK8sResources(yamlContent []byte, nodeName string, randomNum int32) ([]runtime.Object, error) {
	decoder := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
	var objects []runtime.Object

	yamlParts := bytes.Split(yamlContent, []byte("---"))
	for _, part := range yamlParts {
		if len(bytes.TrimSpace(part)) == 0 {
			continue
		}

		obj := &unstructured.Unstructured{}
		_, gvk, err := decoder.Decode(part, nil, obj)
		if err != nil {
			return nil, fmt.Errorf("YAML解析失败: %v", err)
		}

		typedObj, err := convertToTyped(obj, gvk)
		if err != nil {
			return nil, err
		}
		injectNodeSelector(typedObj, nodeName)
		injectLabels(typedObj, nodeName)
		injectName(typedObj, randomNum)
		objects = append(objects, typedObj)
	}

	return objects, nil
}
func injectName(obj runtime.Object, randomNum int32) {
	metaObj, ok := obj.(metav1.Object)
	if !ok {
		return
	}

	// 获取原始名称（修改前）
	originalName := metaObj.GetName()

	// 只处理名称以 "grpc" 开头的情况
	if !strings.HasPrefix(originalName, "grpc") {
		return
	}

	// 生成后缀（-加随机数）
	suffix := fmt.Sprintf("-%d", randomNum)

	// 修改对象名称
	metaObj.SetName(originalName + suffix)

	// 处理标签中的 app
	labels := metaObj.GetLabels()
	if appVal, exists := labels["app"]; exists {
		labels["app"] = appVal + suffix
		metaObj.SetLabels(labels)
	}

	// 根据不同资源类型进行特殊处理
	switch t := obj.(type) {
	case *corev1.Pod:
		handlePod(t, originalName, randomNum)

	case *corev1.Service:
		handleService(t, originalName, randomNum)

	case *appsv1.Deployment:
		handleDeployment(t, suffix)
	}
}

// 处理Pod的特殊修改
func handlePod(pod *corev1.Pod, originalName string, randomNum int32) {
	targetPods := []string{"grpc-client-pod", "grpc-client-pod-copy"}

	// 检查是否是特定的Pod
	for _, name := range targetPods {
		if originalName == name {
			// 更新MY_PORT环境变量
			for i := range pod.Spec.Containers {
				for j := range pod.Spec.Containers[i].Env {
					if pod.Spec.Containers[i].Env[j].Name == "MY_PORT" {
						pod.Spec.Containers[i].Env[j].Value = fmt.Sprintf("%d", randomNum)
					}
				}
			}
			break
		}
	}
}

// 处理Service的特殊修改
func handleService(svc *corev1.Service, originalName string, randomNum int32) {
	// 处理selector中的app
	if appVal, exists := svc.Spec.Selector["app"]; exists {
		svc.Spec.Selector["app"] = appVal + fmt.Sprintf("-%d", randomNum)
	}

	// 处理特定Service的NodePort
	for i := range svc.Spec.Ports {
		port := &svc.Spec.Ports[i]
		if port.NodePort > 0 {
			switch originalName {
			case "grpc-client-service":
				port.NodePort = randomNum + 1
			case "grpc-server-service":
				port.NodePort = randomNum
			case "grpc-client-service-copy":
				port.NodePort = randomNum + 2
			}
		}
	}
}

// 处理Deployment的特殊修改
func handleDeployment(deploy *appsv1.Deployment, suffix string) {
	// 更新spec.selector中的app
	if deploy.Spec.Selector != nil {
		if appVal, exists := deploy.Spec.Selector.MatchLabels["app"]; exists {
			deploy.Spec.Selector.MatchLabels["app"] = appVal + suffix
		}
	}

	// 更新Pod模板中的标签
	labels := deploy.Spec.Template.Labels
	if appVal, exists := labels["app"]; exists {
		labels["app"] = appVal + suffix
		deploy.Spec.Template.Labels = labels
	}
}

func convertToTyped(obj *unstructured.Unstructured, gvk *schema.GroupVersionKind) (runtime.Object, error) {
	switch gvk.Kind {
	case "Pod":
		pod := &corev1.Pod{}
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.UnstructuredContent(), pod); err != nil {
			return nil, err
		}
		return pod, nil
	case "Service":
		svc := &corev1.Service{}
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.UnstructuredContent(), svc); err != nil {
			return nil, err
		}
		return svc, nil
	case "Deployment":
		deployment := &appsv1.Deployment{}
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.UnstructuredContent(), deployment); err != nil {
			return nil, err
		}
		return deployment, nil
	default:
		return nil, fmt.Errorf("unsupported resource type: %s", gvk.Kind)
	}
}
func injectNodeSelector(obj runtime.Object, nodeName string) error {
	switch t := obj.(type) {
	case *corev1.Pod:
		// 处理 Pod 的 nodeSelector
		if t.Spec.NodeSelector == nil {
			t.Spec.NodeSelector = make(map[string]string)
		}
		t.Spec.NodeSelector["kubernetes.io/hostname"] = nodeName

	case *appsv1.Deployment:
		// 处理 Deployment 的 nodeSelector
		if t.Spec.Template.Spec.NodeSelector == nil {
			t.Spec.Template.Spec.NodeSelector = make(map[string]string)
		}
		t.Spec.Template.Spec.NodeSelector["kubernetes.io/hostname"] = nodeName

	case *corev1.Service:
		// Service 不需要处理 nodeSelector
		return nil

	default:
		return fmt.Errorf("unsupported resource type for nodeSelector injection: %T", obj)
	}
	return nil
}
func injectLabels(obj runtime.Object, nodeName string) {
	//metaObj, ok := obj.(metav1.Object)
	//if !ok {
	//	return
	//}
	switch t := obj.(type) {
	case *corev1.Pod:
		addLabelToMeta(&t.ObjectMeta, nodeName)
	case *appsv1.Deployment:
		addLabelToMeta(&t.ObjectMeta, nodeName)
	default:
		return
	}
	//labels := metaObj.GetLabels()
	//if labels == nil {
	//	labels = make(map[string]string)
	//}
	//labels[monitor.CreateorLabel] = nodeName
	//metaObj.SetLabels(labels)
}

// 独立标签注入函数,为了监控
func addLabelToMeta(meta *metav1.ObjectMeta, nodeName string) {
	if meta.Labels == nil {
		meta.Labels = make(map[string]string)
	}
	meta.Labels[monitor.CreateorLabel] = nodeName
}

/*
// 文件路径入口
func ParseK8sResourcesFromFile(filePath string, nodeName string) ([]runtime.Object, error) {
	if ext := filepath.Ext(filePath); ext != ".yaml" && ext != ".yml" {
		return nil, fmt.Errorf("invalid file type: %s", ext)
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("file read error: %v", err)
	}

	return parseK8sResources(content, nodeName)
}

// 核心解析逻辑
func parseK8sResources(yamlContent []byte, nodeName string) ([]runtime.Object, error) {
	decoder := yaml.NewYAMLOrJSONDecoder(bytes.NewReader(yamlContent), 4096)
	var objects []runtime.Object

	for {
		var obj runtime.Object
		err := decoder.Decode(&obj)
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return nil, fmt.Errorf("YAML解析失败: %v", err)
		}

		// 按类型注入标签
		injectLabels(obj, nodeName)
		objects = append(objects, obj)
	}

	return objects, nil
}

// 类型安全标签注入
func injectLabels(obj runtime.Object, nodeName string) {
	switch t := obj.(type) {
	case *corev1.Pod:
		addLabelToMeta(&t.ObjectMeta, nodeName)
	case *appsv1.Deployment:
		addLabelToMeta(&t.ObjectMeta, nodeName)
	}
}

func addLabelToMeta(meta *metav1.ObjectMeta, nodeName string) {
	if meta.Labels == nil {
		meta.Labels = make(map[string]string)
	}
	meta.Labels[monitor.CreateorLabel] = nodeName
}
*/

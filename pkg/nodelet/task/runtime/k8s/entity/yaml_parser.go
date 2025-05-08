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
)

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

func parseK8sResources(yamlContent []byte, nodeName string) ([]runtime.Object, error) {
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
		objects = append(objects, typedObj)
	}

	return objects, nil
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

// 独立标签注入函数
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

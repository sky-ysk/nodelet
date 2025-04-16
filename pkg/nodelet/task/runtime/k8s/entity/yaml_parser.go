package entity

import (
	"bytes"
	"fmt"
	"hit.edu/framework/pkg/nodelet/task/runtime/k8s/monitor"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/yaml"
	"os"
	"path/filepath"
)

// 文件路径入口
func ParseK8sResourcesFromFile(filePath string, nodeName string) ([]runtime.Object, error) {
	if ext := filepath.Ext(filePath); ext != ".yaml" && ext != ".yml" {
		return nil, fmt.Errorf("invalid file type: %s", ext)
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("file read error: %v", err)
	}

	return ParseK8sResources(content, nodeName)
}

// 核心解析逻辑
func ParseK8sResources(yamlContent []byte, nodeName string) ([]runtime.Object, error) {
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

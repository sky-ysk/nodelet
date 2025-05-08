// util/yamltool.go

package util

import (
	"bytes"
	"fmt"
	"gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ================== 通用配置结构体 ==================
type (
	// 节点选择器修改配置
	NodeSelectorConfig struct {
		NodeName string // 要修改的节点名称
	}

	// 名称标签修改配置
	NameLabelConfig struct {
		NameSuffix  string // 资源名称后缀
		LabelSuffix string // 标签值后缀
	}
)

// ================== 核心工具函数 ==================

// 修改YAML中的节点选择器（只处理nodeSelector字段）
func ModifyNodeSelector(yamlContent []byte, config NodeSelectorConfig) ([]byte, error) {
	return modifyYAML(yamlContent, func(obj interface{}) {
		switch resource := obj.(type) {
		case *corev1.Pod:
			updateNodeSelector(resource, config.NodeName)
		}
	})
}

// 修改YAML中的名称和标签（处理Pod/Service的名称和标签）
func ModifyNameAndLabels(yamlContent []byte, config NameLabelConfig) ([]byte, error) {
	return modifyYAML(yamlContent, func(obj interface{}) {
		switch resource := obj.(type) {
		case *corev1.Pod:
			updatePodNameAndLabel(resource, config)
		case *corev1.Service:
			updateServiceNameAndLabel(resource, config)
		}
	})
}

// ================== 私有工具函数 ==================
func modifyYAML(original []byte, modifier func(interface{})) ([]byte, error) {
	decoder := yaml.NewDecoder(bytes.NewReader(original))
	var modifiedBuff bytes.Buffer
	encoder := yaml.NewEncoder(&modifiedBuff)

	for {
		var node yaml.Node
		if err := decoder.Decode(&node); err != nil {
			break // 读取结束
		}

		obj, err := parseKubernetesObject(&node)
		if err != nil {
			return nil, err
		}

		// 执行修改操作
		modifier(obj)

		if err := encoder.Encode(obj); err != nil {
			return nil, fmt.Errorf("编码 YAML 失败: %v", err)
		}
	}

	encoder.Close()
	return modifiedBuff.Bytes(), nil
}

func parseKubernetesObject(node *yaml.Node) (interface{}, error) {
	var typeMeta metav1.TypeMeta
	if err := node.Decode(&typeMeta); err != nil {
		return nil, fmt.Errorf("解析 TypeMeta 失败: %v", err)
	}

	switch typeMeta.Kind {
	case "Pod":
		var pod corev1.Pod
		if err := node.Decode(&pod); err != nil {
			return nil, fmt.Errorf("解析 Pod 失败: %v", err)
		}
		return &pod, nil
	case "Service":
		var svc corev1.Service
		if err := node.Decode(&svc); err != nil {
			return nil, fmt.Errorf("解析 Service 失败: %v", err)
		}
		return &svc, nil
	default:
		return nil, fmt.Errorf("不支持的资源类型: %s", typeMeta.Kind)
	}
}

// ================== 具体修改操作 ==================
func updateNodeSelector(pod *corev1.Pod, nodeName string) {
	if pod.Spec.NodeSelector == nil {
		pod.Spec.NodeSelector = make(map[string]string)
	}
	pod.Spec.NodeSelector["kubernetes.io/hostname"] = nodeName
}

func updatePodNameAndLabel(pod *corev1.Pod, config NameLabelConfig) {
	// 修改名称
	pod.ObjectMeta.Name += config.NameSuffix

	// 修改标签
	if pod.ObjectMeta.Labels == nil {
		pod.ObjectMeta.Labels = make(map[string]string)
	}
	if app, exists := pod.ObjectMeta.Labels["app"]; exists {
		pod.ObjectMeta.Labels["app"] = app + config.LabelSuffix
	}
}

func updateServiceNameAndLabel(svc *corev1.Service, config NameLabelConfig) {
	// 修改名称
	svc.ObjectMeta.Name += config.NameSuffix

	// 修改选择器标签
	if svc.Spec.Selector == nil {
		svc.Spec.Selector = make(map[string]string)
	}
	if app, exists := svc.Spec.Selector["app"]; exists {
		svc.Spec.Selector["app"] = app + config.LabelSuffix
	}
}

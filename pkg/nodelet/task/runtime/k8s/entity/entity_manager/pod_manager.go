package entity_manager

import (
	"fmt"
	"hit.edu/framework/pkg/component-base/logs"
	"sync"
)

type PodIdentifier struct {
	Namespace string
	PodName   string
}
type PodsManager struct {
	pods map[PodIdentifier]bool //后期看map的value是存bool还是pod信息好
	mu   sync.RWMutex
}

var (
	instance *PodsManager
	once     sync.Once
)

func GetInstance() *PodsManager {
	once.Do(func() {
		instance = &PodsManager{
			pods: make(map[PodIdentifier]bool),
		}
	})
	return instance
}

func (pm *PodsManager) AddPod(namespace, podName string) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	podId := PodIdentifier{Namespace: namespace, PodName: podName}
	pm.pods[podId] = true
}
func (pm *PodsManager) RemovePod(namespace, podName string) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	podId := PodIdentifier{Namespace: namespace, PodName: podName}
	if _, ok := pm.pods[podId]; ok {
		delete(pm.pods, podId)
	} else {
		logs.Error("Pod Not Found")
	}
}

// 获取 Pod 的状态 应该
func (pm *PodsManager) GetPodStatus(namespace, podName string) (bool, bool) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	podId := PodIdentifier{Namespace: namespace, PodName: podName}
	status, exists := pm.pods[podId]
	return status, exists
}

// 更新 Pod 的状态
func (pm *PodsManager) UpdatePodStatus(namespace, podName string, status bool) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	podId := PodIdentifier{Namespace: namespace, PodName: podName}
	if _, exists := pm.pods[podId]; exists {
		pm.pods[podId] = status
	}
}

// 获取所有的 Pod
func (pm *PodsManager) GetAllPods() map[PodIdentifier]bool {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.pods
}

// 将 PodIdentifier 转换为字符串表示，用于日志
func PodIdentifierToString(podId PodIdentifier) string {
	return fmt.Sprintf("Namespace:%s,PodName:%s", podId.Namespace, podId.PodName)
}

package manager

import (
	"hit.edu/framework/pkg/client-go/clients"
	"hit.edu/framework/pkg/client-go/clients/typed/core"
	"sync"
)

type Manager struct {
	WorkflowClients map[string]core.WorkflowInterface
	TaskClients     map[string]core.TaskInterface
	GroupClients    map[string]core.GroupInterface
	ActionClients   map[string]core.ActionInterface
	RuntimeClients  map[string]core.RuntimeInterface
	DeviceClients   map[string]core.DeviceInterface
	ClientSet       *clients.ClientSet
	mu              sync.Mutex
}

func NewManager(clientSet *clients.ClientSet) *Manager {
	return &Manager{
		WorkflowClients: make(map[string]core.WorkflowInterface),
		TaskClients:     make(map[string]core.TaskInterface),
		GroupClients:    make(map[string]core.GroupInterface),
		ActionClients:   make(map[string]core.ActionInterface),
		RuntimeClients:  make(map[string]core.RuntimeInterface),
		DeviceClients:   make(map[string]core.DeviceInterface),
		ClientSet:       clientSet,
	}
}

type WorkflowClient struct {
	Client core.WorkflowInterface
}

type TaskClient struct {
	Client core.TaskInterface
}

type GroupClient struct {
	Client core.GroupInterface
}
type ActionClient struct {
	Client core.ActionInterface
}

type RuntimeClient struct {
	Client core.RuntimeInterface
}

type DeviceClient struct {
	Client core.DeviceInterface
}

// 根据 namespace 获取 client，如果不存在则创建
func (m *Manager) GetWorkflowClient(namespace string) *WorkflowClient {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 如果已经存在，直接返回
	if c, exists := m.WorkflowClients[namespace]; exists {
		return &WorkflowClient{
			Client: c,
		}
	}

	// 否则创建新的 client
	newClient := m.ClientSet.Core().Workflows(namespace)
	m.WorkflowClients[namespace] = newClient
	return &WorkflowClient{
		Client: newClient,
	}
}

// 根据 namespace 获取 client，如果不存在则创建
func (m *Manager) GetTaskClient(namespace string) *TaskClient {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 如果已经存在，直接返回
	if c, exists := m.TaskClients[namespace]; exists {
		return &TaskClient{
			Client: c,
		}
	}

	// 否则创建新的 client
	newClient := m.ClientSet.Core().Tasks(namespace)
	m.TaskClients[namespace] = newClient
	return &TaskClient{
		Client: newClient,
	}
}

// 根据 namespace 获取 client，如果不存在则创建
func (m *Manager) GetGroupClient(namespace string) *GroupClient {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 如果已经存在，直接返回
	if c, exists := m.GroupClients[namespace]; exists {
		return &GroupClient{
			Client: c,
		}
	}

	// 否则创建新的 client
	newClient := m.ClientSet.Core().Groups(namespace)
	m.GroupClients[namespace] = newClient
	return &GroupClient{
		Client: newClient,
	}
}

// 根据 namespace 获取 client，如果不存在则创建
func (m *Manager) GetActionClient(namespace string) *ActionClient {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 如果已经存在，直接返回
	if c, exists := m.ActionClients[namespace]; exists {
		return &ActionClient{
			Client: c,
		}
	}

	// 否则创建新的 client
	newClient := m.ClientSet.Core().Actions(namespace)
	m.ActionClients[namespace] = newClient
	return &ActionClient{
		Client: newClient,
	}
}

// 根据 namespace 获取 client，如果不存在则创建
func (m *Manager) GetRuntimeClient(namespace string) *RuntimeClient {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 如果已经存在，直接返回
	if c, exists := m.RuntimeClients[namespace]; exists {
		return &RuntimeClient{
			Client: c,
		}
	}

	// 否则创建新的 client
	newClient := m.ClientSet.Core().Runtimes(namespace)
	m.RuntimeClients[namespace] = newClient
	return &RuntimeClient{
		Client: newClient,
	}
}

// 根据 namespace 获取 client，如果不存在则创建
func (m *Manager) GetDeviceClient(namespace string) *DeviceClient {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 如果已经存在，直接返回
	if c, exists := m.DeviceClients[namespace]; exists {
		return &DeviceClient{
			Client: c,
		}
	}

	// 否则创建新的 client
	newClient := m.ClientSet.Core().Devices(namespace)
	m.DeviceClients[namespace] = newClient
	return &DeviceClient{
		Client: newClient,
	}
}

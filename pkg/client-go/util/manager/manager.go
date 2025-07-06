package manager

import (
	"context"
	"errors"
	scheme "hit.edu/framework/pkg/apimachinery/runtime"
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/client-go/clients"
	"hit.edu/framework/pkg/client-go/clients/typed/core"
	"hit.edu/framework/pkg/client-go/tools/recorder"
	"sync"
)

var (
	NotFound            = errors.New("404 Not Found")
	InternalServerError = errors.New("500 Internal Server Error")
)

type Manager struct {
	WorkflowClients   map[string]core.WorkflowInterface
	TaskClients       map[string]core.TaskInterface
	GroupClients      map[string]core.GroupInterface
	ActionClients     map[string]core.ActionInterface
	RuntimeClients    map[string]core.RuntimeInterface
	DeviceClients     map[string]core.DeviceInterface
	EventClients      map[string]core.EventInterface
	NodeClients       map[string]core.NodeInterface
	DataClients       map[string]core.DataInterface
	SceneClients      map[string]core.SceneInterface
	EventBroadCasters map[string]recorder.EventBroadcaster
	Recoders          map[string]recorder.EventRecorder
	ClientSet         *clients.ClientSet
	mu                sync.Mutex
}

func NewManager(clientSet *clients.ClientSet) *Manager {
	return &Manager{
		WorkflowClients:   make(map[string]core.WorkflowInterface),
		TaskClients:       make(map[string]core.TaskInterface),
		GroupClients:      make(map[string]core.GroupInterface),
		ActionClients:     make(map[string]core.ActionInterface),
		RuntimeClients:    make(map[string]core.RuntimeInterface),
		DeviceClients:     make(map[string]core.DeviceInterface),
		NodeClients:       make(map[string]core.NodeInterface),
		DataClients:       make(map[string]core.DataInterface),
		SceneClients:      make(map[string]core.SceneInterface),
		EventClients:      make(map[string]core.EventInterface),
		EventBroadCasters: make(map[string]recorder.EventBroadcaster),
		Recoders:          make(map[string]recorder.EventRecorder),
		ClientSet:         clientSet,
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

type NodeClient struct {
	Client core.NodeInterface
}

type DataClient struct {
	Client core.DataInterface
}

type SceneClient struct {
	Client core.SceneInterface
}

type EventClient struct {
	Client      core.EventInterface
	Broadcaster recorder.EventBroadcaster
	Recoder     recorder.EventRecorder
}

// GetWorkflowClient 根据 namespace 获取 client，如果不存在则创建
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

// GetTaskClient 根据 namespace 获取 client，如果不存在则创建
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

// GetGroupClient 根据 namespace 获取 client，如果不存在则创建
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

// GetActionClient 根据 namespace 获取 client，如果不存在则创建
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

// GetRuntimeClient 根据 namespace 获取 client，如果不存在则创建
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

// GetDeviceClient 根据 namespace 获取 client，如果不存在则创建
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

func (m *Manager) GetNodeClient(namespace string) *NodeClient {
	m.mu.Lock()
	defer m.mu.Unlock()

	if c, exists := m.NodeClients[namespace]; exists {
		return &NodeClient{
			Client: c,
		}
	}

	newClient := m.ClientSet.Core().Nodes(namespace)
	m.NodeClients[namespace] = newClient
	return &NodeClient{
		Client: newClient,
	}
}

func (m *Manager) GetSceneClient(namespace string) *SceneClient {
	m.mu.Lock()
	defer m.mu.Unlock()

	if c, exists := m.SceneClients[namespace]; exists {
		return &SceneClient{
			Client: c,
		}
	}

	newClient := m.ClientSet.Core().Scenes(namespace)
	m.SceneClients[namespace] = newClient
	return &SceneClient{
		Client: newClient,
	}
}

func (m *Manager) GetDataClient(namespace string) *DataClient {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 如果已经存在，直接返回
	if c, exists := m.DataClients[namespace]; exists {
		return &DataClient{
			Client: c,
		}
	}

	// 否则创建新的 client
	newClient := m.ClientSet.Core().Datas(namespace)
	m.DataClients[namespace] = newClient
	return &DataClient{
		Client: newClient,
	}
}

// GetEventClient 根据 namespace 获取 client，如果不存在则创建
func (m *Manager) GetEventClient(namespace string) *EventClient {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 如果已经存在，直接返回
	if c, exists := m.EventClients[namespace]; exists {
		return &EventClient{
			Client:      c,
			Broadcaster: m.EventBroadCasters[namespace],
			Recoder:     m.Recoders[namespace],
		}
	}

	// 否则创建新的 client
	newClient := m.ClientSet.Core().Events(namespace)
	m.EventClients[namespace] = newClient

	// 创建一个事件
	eventBroadcaster := recorder.NewBroadcaster()
	err := eventBroadcaster.StartRecordingToSink(context.Background(), &core.EventSinkImpl{Interface: newClient})
	if err != nil {
		return nil
	}
	m.EventBroadCasters[namespace] = eventBroadcaster

	s := scheme.NewScheme()
	apis.AddToScheme(s)
	recorder := eventBroadcaster.NewRecorder(s, "Manager")
	m.Recoders[namespace] = recorder

	return &EventClient{
		Client:      newClient,
		Broadcaster: eventBroadcaster,
		Recoder:     recorder,
	}
}

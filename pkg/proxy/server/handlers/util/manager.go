package util

import (
	"context"
	"fmt"
	apis "hit.edu/framework/pkg/apis/cores"
	metav1 "hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/client-go/clients"
	"hit.edu/framework/pkg/client-go/clients/typed/core"
	"hit.edu/framework/pkg/component-base/logs"
	"sync"
	"time"
)

type Manager struct {
	// TODO: 集成客户端
	WorkflowClients map[string]core.WorkflowInterface
	TaskClients     map[string]core.TaskInterface
	GroupClients    map[string]core.GroupInterface
	ActionClients   map[string]core.ActionInterface
	RuntimeClients  map[string]core.RuntimeInterface
	ClientSet       *clients.ClientSet
	mu              sync.Mutex
}

func NewMangerHandler(clientSet *clients.ClientSet) *Manager {
	return &Manager{
		WorkflowClients: make(map[string]core.WorkflowInterface),
		TaskClients:     make(map[string]core.TaskInterface),
		GroupClients:    make(map[string]core.GroupInterface),
		ActionClients:   make(map[string]core.ActionInterface),
		RuntimeClients:  make(map[string]core.RuntimeInterface),
		ClientSet:       clientSet,
	}
}

type WorkflowHandler struct {
	Client core.WorkflowInterface
}

// 根据 namespace 获取 client，如果不存在则创建
func (h *Manager) GetWorkflowClient(namespace string) *WorkflowHandler {
	h.mu.Lock()
	defer h.mu.Unlock()

	// 如果已经存在，直接返回
	if c, exists := h.WorkflowClients[namespace]; exists {
		return &WorkflowHandler{
			Client: c,
		}
	}

	// 否则创建新的 client
	newClient := h.ClientSet.Core().Workflows(namespace)
	h.WorkflowClients[namespace] = newClient
	return &WorkflowHandler{
		Client: newClient,
	}
}

// TODO: GetWorkflow
func (h *Manager) GetWorkflow(name string, namespace string) (*apis.Workflow, error) {
	// 获取workflow的handler
	c := h.GetWorkflowClient(namespace)

	// 获取Workflow
	result, err := c.Client.Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	} else {
		return result, nil
	}
}

// TODO: CreateWorkflow

func (h *Manager) CreateWorkflow(as apis.WorkflowSpec, namespace string, uuid string) (*apis.Workflow, error) {
	// 查看workflow是否存在
	res, err := h.GetWorkflow(as.Name, namespace)
	if err != nil {
		logs.Infof("Get workflow %s error: %v , node not exist ! creat it ", res.Name, err)
	} else {
		logs.Errorf("Create node %s error, node existed: %v", res.Name, res)
		return res, fmt.Errorf("create node %s error, node existed: %v", res.Name, res)
	}

	// workflow 不存在， 创建
	// 临时创建一个Workflow对象
	a := apis.Workflow{}
	// 构造名称
	a.Name = as.Name + uuid
	// 构造Namespace
	if namespace == "" {
		a.Namespace = apis.NamespaceDefault
	} else {
		a.Namespace = namespace
	}

	// 复制Spec
	a.Spec = as

	// 构造Status
	a.Status = apis.WorkflowStatus{}

	// 记录Create时间
	a.Status.CreateAt = apis.Time{time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC)}

	// 初始化状态
	a.Status.Phase = apis.Pending

	// 打上Label
	a.Labels["uuid"] = uuid

	// 根据Spec创建tasks
	tasks, err := CreateTasks(as, namespace, uuid)
	if err != nil {
		return nil, err
	}

	// 根据生成的Runtime修改Action.Status.Runtimes
	for _, r := range tasks {
		a.Status.Tasks[r.Spec.Name] = apis.ObjectReference{
			Name:      r.Name,
			Namespace: r.Namespace,
			Kind:      r.Kind,
		}
	}

	// 获取workflow的handler
	c := h.GetWorkflowClient(a.Namespace)

	// 写入数据库
	wf, err := c.Client.Create(context.TODO(), &a, metav1.CreateOptions{})
	if err != nil {
		logs.Errorf("Create workflow fail , failed write to database ,error: %v", err)
		return nil, err
	}

	return wf, nil
}

// TODO: CreateTasks
func CreateTasks(gs apis.WorkflowSpec, namespace string, uuid string) ([]*apis.Task, error) {
	tasks := []*apis.Task{}
	for _, as := range gs.Tasks {
		a, err := CreateTask(as, &gs, namespace, uuid)
		if err != nil {
			return nil, err
		}
		// TODO: 返回的应该是已经创建的Runtime
		tasks = append(tasks, a) // a替换成实际的fa
	}

	return tasks, nil
}

// TODO: CreateTask
func CreateTask(as apis.TaskSpec, gs *apis.WorkflowSpec, namespace string, uuid string) (*apis.Task, error) {
	// 临时创建一个Task对象
	a := apis.Task{}
	// 构造名称
	a.Name = gs.Name + as.Name + uuid
	// 构造Namespace
	if namespace == "" {
		a.Namespace = apis.NamespaceDefault
	} else {
		a.Namespace = namespace
	}

	// 复制Spec
	a.Spec = as

	// 构造Status
	a.Status = apis.TaskStatus{}

	// 记录Create时间
	a.Status.CreateAt = apis.Time{time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC)}

	// 初始化状态
	a.Status.Phase = apis.Pending

	// 打上Label, 当前任务属于哪个Workflow和uuid域
	if gs != nil {
		a.Labels["task"] = gs.Name + uuid
	}
	a.Labels["uuid"] = uuid

	// 根据Spec创建Groups
	groups, err := CreateGroups(as, namespace, uuid)
	if err != nil {
		return nil, err
	}

	// 根据生成的Runtime修改Task.Status.Groups
	for _, r := range groups {
		a.Status.Groups[r.Spec.Name] = apis.ObjectReference{
			Name:      r.Name,
			Namespace: r.Namespace,
			Kind:      r.Kind,
		}
	}
	// TODO: 写入数据库
	return &a, nil // TODO: 应当返回实际的fa
}

// TODO: CreateGroups
func CreateGroups(gs apis.TaskSpec, namespace string, uuid string) ([]*apis.Group, error) {
	groups := []*apis.Group{}
	for _, as := range gs.Groups {
		a, err := CreateGroup(as, &gs, namespace, uuid)
		if err != nil {
			return nil, err
		}
		// TODO: 返回的应该是已经创建的Runtime
		groups = append(groups, a) // a替换成实际的fa
	}

	return groups, nil
}

// TODO: CreateGroup
func CreateGroup(as apis.GroupSpec, gs *apis.TaskSpec, namespace string, uuid string) (*apis.Group, error) {
	// 临时创建一个Group对象
	a := apis.Group{}
	// 构造名称
	a.Name = as.Name + uuid
	// 构造Namespace
	if namespace == "" {
		a.Namespace = apis.NamespaceDefault
	} else {
		a.Namespace = namespace
	}

	// 复制Spec
	a.Spec = as

	// 构造Status
	a.Status = apis.GroupStatus{}

	// 记录Create时间
	a.Status.CreateAt = apis.Time{time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC)}

	// 初始化状态
	a.Status.Phase = apis.Pending

	// 打上Label, 当前任务属于哪个Task和uuid域
	if gs != nil {
		a.Labels["task"] = gs.Name + uuid
	}
	a.Labels["uuid"] = uuid

	// 根据Spec创建Actions
	actions, err := CreateActions(as, namespace, uuid)
	if err != nil {
		return nil, err
	}

	// 根据生成的Action修改Group.Status.Actions
	for _, r := range actions {
		a.Status.Actions[r.Spec.Name] = apis.ObjectReference{
			Name:      r.Name,
			Namespace: r.Namespace,
			Kind:      r.Kind,
		}
	}
	// TODO: 写入数据库
	return &a, nil // TODO: 应当返回实际的fa
}

// 根据GroupSpec创建Action
func CreateActions(gs apis.GroupSpec, namespace string, uuid string) ([]*apis.Action, error) {
	actions := []*apis.Action{}
	for _, as := range gs.Actions {
		a, err := CreateAction(as, &gs, namespace, uuid)
		if err != nil {
			return nil, err
		}
		// TODO: 返回的应该是已经创建的Runtime
		actions = append(actions, a) // a替换成实际的fa
	}

	return actions, nil
}

func CreateAction(as apis.ActionSpec, gs *apis.GroupSpec, namespace string, uuid string) (*apis.Action, error) {
	// 临时创建一个Action对象
	a := apis.Action{}
	// 构造名称
	a.Name = as.Name + uuid
	// 构造Namespace
	if namespace == "" {
		a.Namespace = apis.NamespaceDefault
	} else {
		a.Namespace = namespace
	}

	// 复制Spec
	a.Spec = as

	// 构造Status
	a.Status = apis.ActionStatus{}

	// 记录Create时间
	a.Status.CreateAt = apis.Time{time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC)}

	// 初始化状态
	a.Status.Phase = apis.Pending

	// 打上Label, 当前任务属于哪个Group和uuid域
	if gs != nil {
		a.Labels["group"] = gs.Name + uuid
	}
	a.Labels["uuid"] = uuid

	// 根据Spec创建Runtimes
	runtimes, err := CreateRuntimes(as, namespace, uuid)
	if err != nil {
		return nil, err
	}

	// 根据生成的Runtime修改Action.Status.Runtimes
	for _, r := range runtimes {
		a.Status.Runtimes[r.Spec.Name] = apis.ObjectReference{
			Name:      r.Name,
			Namespace: r.Namespace,
			Kind:      r.Kind,
		}
	}
	// TODO: 写入数据库
	return &a, nil // TODO: 应当返回实际的fa
}

// 根据ActionSpec创建Runtime
func CreateRuntimes(as apis.ActionSpec, namespace string, uuid string) ([]*apis.Runtime, error) {
	runtimes := []*apis.Runtime{}
	// 遍历所有的Runtime Spec
	for _, rs := range as.Runtimes {
		r, err := CreateRuntime(rs, &as, namespace, uuid)
		if err != nil {
			return nil, err
		}
		// TODO: 返回的应该是已经创建的Runtime
		runtimes = append(runtimes, r) // r替换成实际的fr
	}
	return runtimes, nil
}

// 创建单个Runtime
func CreateRuntime(rs apis.RuntimeSpec, as *apis.ActionSpec, namespace string, uuid string) (*apis.Runtime, error) {
	// 临时创建一个Runtime对象
	r := apis.Runtime{}
	// 构造名称
	r.Name = rs.Name + uuid
	// 构造Namespace
	if namespace == "" {
		r.Namespace = apis.NamespaceDefault
	} else {
		r.Namespace = namespace
	}
	// 复制Spec
	r.Spec = rs

	// 构造Status
	r.Status = apis.RuntimeStatus{}

	// 记录Create时间
	r.Status.CreateAt = apis.Time{time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC)}

	// 初始化状态
	r.Status.Phase = apis.Pending

	// 打上Label, 当前任务属于哪个action和uuid域
	if as != nil {
		r.Labels["action"] = as.Name + uuid
	}
	r.Labels["uuid"] = uuid

	// TODO: 写入Client-Go中, 返回实际的Runtime
	//fr, err := WriteTo

	return &r, nil // TODO: 应当返回fr
}

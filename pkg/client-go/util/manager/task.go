package manager

import (
	"context"
	apis "hit.edu/framework/pkg/apis/cores"
	metav1 "hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/component-base/logs"
	"time"
)

// 根据TaskSpec创建Task
func (m *Manager) CreateTasks(w *apis.Workflow, namespace string, uuid string, prefix string) ([]*apis.Task, error) {
	groups := []*apis.Task{}
	for _, as := range w.Spec.Tasks {
		a, err := m.CreateTask(as, w, namespace, uuid, prefix)
		if err != nil {
			return nil, err
		}
		// TODO: 返回的应该是已经创建的Runtime
		groups = append(groups, a) // a替换成实际的fa
	}

	return groups, nil
}

func (m *Manager) CreateTask(ts apis.TaskSpec, w *apis.Workflow, namespace string, uuid string, prefix string) (*apis.Task, error) {
	// 临时创建一个Task对象
	t := apis.Task{}
	// 构造名称
	if w != nil {
		t.Name = prefix + ts.Name + "-" + uuid
		prefix = prefix + ts.Name + "."
	} else {
		t.Name = ts.Name + "-" + uuid
		prefix = ts.Name + "."
	}

	// 构造Namespace
	if namespace == "" {
		t.Namespace = apis.NamespaceDefault
	} else {
		t.Namespace = namespace
	}

	t.Kind = "Task"
	t.APIVersion = "resources/v1"

	// 构造Labels
	t.Labels = map[string]string{}

	// 复制Spec
	t.Spec = ts

	// 构造Status
	t.Status = apis.TaskStatus{}

	// 记录Create时间
	t.Status.CreateAt = &apis.Time{time.Now()}

	// 初始化状态
	t.Status.Phase = apis.Unknown
	t.Status.Groups = map[string]apis.ObjectReference{}

	// 打上Label, 当前任务属于哪个Task和uuid域
	if w != nil {
		t.Status.Belong = &apis.ObjectReference{
			Name:            w.Name,
			Namespace:       w.Namespace,
			Kind:            w.Kind,
			ResourceVersion: w.ResourceVersion,
			UID:             apis.UID(uuid),
		}
		t.Labels["belong"] = w.Name
	}
	t.Labels["uuid"] = uuid

	// 根据Spec创建Runtimes
	groups, err := m.CreateGroups(&t, namespace, uuid, prefix)
	if err != nil {
		return nil, err
	}

	// 根据生成的Groups修改Task.Status.Groups
	for _, r := range groups {
		t.Status.Groups[r.Spec.Name] = apis.ObjectReference{
			Name:            r.Name,
			Namespace:       r.Namespace,
			Kind:            r.Kind,
			ResourceVersion: r.ResourceVersion,
			UID:             apis.UID(r.UID),
		}
	}
	// 写入Client-Go中, 返回实际的Runtime
	c := m.GetTaskClient(t.Namespace)

	ft, err := c.Client.Create(context.TODO(), &t, metav1.CreateOptions{})
	if err != nil {
		logs.Errorf("Failed to create task: %v", err)
	}
	logs.Debugf("Created task: %v", ft)

	return ft, nil // 应当返回实际的ft
}

func (m *Manager) GetTask(name string, namespace string) (*apis.Task, error) {
	c := m.GetTaskClient(namespace)
	a, err := c.Client.Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	} else {
		return a, nil
	}
}

func (m *Manager) GetTasks(namespace string) (*apis.TaskList, error) {
	c := m.GetTaskClient(namespace)
	g, err := c.Client.List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	} else {
		return g, nil
	}
}

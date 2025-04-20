package manager

import (
	"context"
	"hit.edu/framework/pkg/apimachinery/types"
	apis "hit.edu/framework/pkg/apis/cores"
	metav1 "hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/component-base/logs"
	"time"
)

// CreateTasks 根据TaskSpec创建Group
func (m *Manager) CreateTasks(w *apis.Workflow, namespace string, uuid string, prefix string) ([]*apis.Task, error) {
	var groups []*apis.Task
	for _, as := range w.Spec.Tasks {
		a, err := m.CreateTask(as, w, namespace, uuid, prefix)
		if err != nil {
			return nil, err
		}
		groups = append(groups, a)
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
	t.Status.CreateAt = &apis.Time{Time: time.Now()}

	// 初始化状态
	t.Status.Phase = apis.Pending
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

	// 根据生成的Runtime修改Task.Status.Groups
	for _, r := range groups {
		t.Status.Groups[r.Spec.Name] = apis.ObjectReference{
			Name:            r.Name,
			Namespace:       r.Namespace,
			Kind:            r.Kind,
			ResourceVersion: r.ResourceVersion,
			UID:             apis.UID(r.UID),
		}
	}
	// 写入Client-Go中, 返回实际的Task
	c := m.GetTaskClient(t.Namespace)

	ft, err := c.Client.Create(context.TODO(), &t, metav1.CreateOptions{})
	if err != nil {
		logs.Errorf("Failed to create task: %v", err)
		return nil, err
	}

	logs.Debugf("Created task: %v", ft)
	return ft, nil
}

func (m *Manager) GetTask(name string, namespace string) (*apis.Task, error) {
	c := m.GetTaskClient(namespace)
	a, err := c.Client.Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		logs.Errorf("Failed to get task: %v", err)
		return nil, err
	}

	logs.Debugf("Get task: %v", a)
	return a, nil
}

func (m *Manager) GetTasks(namespace string) (*apis.TaskList, error) {
	c := m.GetTaskClient(namespace)
	g, err := c.Client.List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		logs.Errorf("Failed to get tasks: %v", err)
		return nil, err
	}

	logs.Debugf("Get tasks success.")
	return g, nil
}

func (m *Manager) UpdateTask(namespace string, name string, a *apis.Task) (*apis.Task, error) {
	c := m.GetTaskClient(namespace)

	// 检查task是否存在
	_, err := m.GetTask(name, namespace)
	if err != nil {
		logs.Errorf("Get task %s error: %v , task not exist !", name, err)
		return nil, err
	}

	// 存在更新task
	updatedTask, updateErr := c.Client.Update(context.TODO(), a, metav1.UpdateOptions{})
	if updateErr != nil {
		logs.Errorf("Update task %s error: %v", name, updateErr)
		return nil, updateErr
	}

	logs.Debugf("Update task: %v", updatedTask)
	return updatedTask, nil

}

func (m *Manager) PatchTask(name string, namespace string, patchTask string) (*apis.Task, error) {
	c := m.GetTaskClient(namespace)

	// 检查task是否存在
	_, err := m.GetTask(name, namespace)
	if err != nil {
		logs.Errorf("Get task %s error: %v , task not exist !", name, err)
		return nil, err
	}

	// 部分更新task
	patchedTask, err := c.Client.Patch(context.TODO(), name, types.StrategicMergePatchType, []byte(patchTask), metav1.PatchOptions{})
	if err != nil {
		logs.Errorf("patch task %s error: %v", name, err)
		return nil, err
	}

	logs.Debugf("Patch task: %v", patchedTask)
	return patchedTask, nil
}

func (m *Manager) DeleteTask(name string, namespace string) error {
	c := m.GetTaskClient(namespace)

	// 检查task是否存在
	task, err := m.GetTask(name, namespace)
	if err != nil {
		logs.Errorf("get task %s error: %v , task not exist ", name, err)
		return err
	}

	// 删除task里面的所有group
	for _, v := range task.Status.Groups {
		err := m.DeleteGroup(v.Name, v.Namespace)
		if err != nil {
			return err
		}
	}

	// 存在，删除
	err = c.Client.Delete(context.TODO(), name, metav1.DeleteOptions{})
	if err != nil {
		logs.Errorf("delete task %s error: %v", name, err)
		return err
	}

	logs.Debugf("Delete task: %v", name)
	return nil
}

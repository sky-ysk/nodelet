package manager

import (
	"context"
	"hit.edu/framework/pkg/apimachinery/types"
	apis "hit.edu/framework/pkg/apis/cores"
	metav1 "hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/component-base/logs"
	"net/http"
	"time"
)

// CreateTasks 根据TaskSpec创建Group
func (m *Manager) CreateTasks(w *apis.Workflow, namespace string, uuid string, prefix string) ([]*apis.Task, int, error) {
	var groups []*apis.Task
	for _, as := range w.Spec.Tasks {
		a, code, err := m.CreateTask(as, w, namespace, uuid, prefix)
		if err != nil {
			return nil, code, err
		}
		groups = append(groups, a)
	}

	return groups, http.StatusOK, nil
}

func (m *Manager) CreateTask(ts apis.TaskSpec, w *apis.Workflow, namespace string, uuid string, prefix string) (*apis.Task, int, error) {
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
	groups, code, err := m.CreateGroups(&t, namespace, uuid, prefix)
	if err != nil {
		return nil, code, err
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
		return nil, http.StatusInternalServerError, err
	}

	logs.Debugf("Created task: %v", ft)
	return ft, http.StatusOK, nil
}

// 创建带有label的task
func (m *Manager) CreateTaskWithLabels(ts apis.TaskSpec, w *apis.Workflow, namespace string, uuid string, prefix string, labels map[string]string) (*apis.Task, int, error) {
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
	t.Labels = labels

	// 复制Spec
	t.Spec = ts

	// 构造Status
	t.Status = apis.TaskStatus{}

	// 记录Create时间
	t.Status.CreateAt = &apis.Time{Time: time.Now()}

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
	groups, code, err := m.CreateGroups(&t, namespace, uuid, prefix)
	if err != nil {
		return nil, code, err
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
		return nil, http.StatusInternalServerError, err
	}

	logs.Debugf("Created task: %v", ft)
	return ft, http.StatusOK, nil
}

func (m *Manager) GetTask(name string, namespace string) (*apis.Task, int, error) {
	c := m.GetTaskClient(namespace)
	a, err := c.Client.Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		logs.Errorf("Failed to get task: %v", err)
		return nil, http.StatusNotFound, err
	}

	logs.Debugf("Get task: %v", a)
	return a, http.StatusOK, nil
}

// 根据Label查询Tasks
func (m *Manager) FilterTasks(namespace string, labelSelector string) (*apis.TaskList, int, error) {
	//labelSelector := ""
	//for i, l := range label {
	//	if i == 0 {
	//		labelSelector = l
	//	} else {
	//		labelSelector += "," + l
	//	}
	//}

	listOptions := metav1.ListOptions{
		LabelSelector: labelSelector,
	}

	c := m.GetTaskClient(namespace)
	d, err := c.Client.List(context.TODO(), listOptions)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	return d, http.StatusOK, nil
}

func (m *Manager) GetTasks(namespace string) (*apis.TaskList, int, error) {
	c := m.GetTaskClient(namespace)
	g, err := c.Client.List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		logs.Errorf("Failed to get tasks: %v", err)
		return nil, http.StatusInternalServerError, err
	}

	logs.Debugf("Get tasks success.")
	return g, http.StatusOK, nil
}

func (m *Manager) UpdateTask(name string, namespace string, a *apis.Task) (*apis.Task, int, error) {
	c := m.GetTaskClient(namespace)

	// 检查task是否存在
	_, code, err := m.GetTask(name, namespace)
	if err != nil {
		logs.Errorf("Get task %s error: %v , task not exist !", name, err)
		return nil, code, err
	}

	// 存在更新task
	updatedTask, updateErr := c.Client.Update(context.TODO(), a, metav1.UpdateOptions{})
	if updateErr != nil {
		logs.Errorf("Update task %s error: %v", name, updateErr)
		return nil, http.StatusInternalServerError, updateErr
	}

	logs.Debugf("Update task: %v", updatedTask)
	return updatedTask, http.StatusOK, nil

}

func (m *Manager) PatchTask(name string, namespace string, patchTask []byte) (*apis.Task, int, error) {
	c := m.GetTaskClient(namespace)

	// 检查task是否存在
	_, code, err := m.GetTask(name, namespace)
	if err != nil {
		logs.Errorf("Get task %s error: %v , task not exist !", name, err)
		return nil, code, err
	}

	// 部分更新task
	patchedTask, err := c.Client.Patch(context.TODO(), name, types.StrategicMergePatchType, patchTask, metav1.PatchOptions{})
	if err != nil {
		logs.Errorf("patch task %s error: %v", name, err)
		return nil, http.StatusInternalServerError, err
	}

	logs.Debugf("Patch task: %v", patchedTask)
	return patchedTask, http.StatusOK, nil
}

func (m *Manager) DeleteTask(name string, namespace string) (int, error) {
	c := m.GetTaskClient(namespace)

	// 检查task是否存在
	task, code, err := m.GetTask(name, namespace)
	if err != nil {
		logs.Errorf("get task %s error: %v , task not exist ", name, err)
		return code, err
	}

	// 删除task里面的所有group
	for _, v := range task.Status.Groups {
		code, err := m.DeleteGroup(v.Name, v.Namespace)
		if err != nil {
			return code, err
		}
	}

	// 存在，删除
	err = c.Client.Delete(context.TODO(), name, metav1.DeleteOptions{})
	if err != nil {
		logs.Errorf("delete task %s error: %v", name, err)
		return http.StatusInternalServerError, err
	}

	logs.Debugf("Delete task: %v", name)
	return http.StatusOK, nil
}

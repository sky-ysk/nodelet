package manager

import (
	"context"
	"errors"
	"fmt"
	"time"

	"hit.edu/framework/pkg/apimachinery/types"
	apis "hit.edu/framework/pkg/apis/cores"
	metav1 "hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/component-base/logs"
)

// CreateTasks 根据TaskSpec创建Group
func (m *Manager) CreateTasks(w *apis.Workflow, namespace string, uuid string, prefix string) ([]*apis.Task, error) {
	start := time.Now() // 记录开始时间
	var groups []*apis.Task

	for _, as := range w.Spec.Tasks {
		a, err := m.CreateTask(as, w, namespace, uuid, prefix)
		if err != nil {
			logs.Errorf("create actions in groupSpec failed , err: %v", err)
			return nil, err
		}
		groups = append(groups, a)
	}

	elapsed := time.Since(start) // 计算耗时
	logs.Tracef("-------------------------test time : CreateTasks took %s", elapsed)

	return groups, nil
}

func (m *Manager) CreateTask(ts apis.TaskSpec, w *apis.Workflow, namespace string, uuid string, prefix string) (*apis.Task, error) {
	start := time.Now() // 记录开始时间
	// TODO：需要检查一下Spec里面的东西 1.循环依赖  2.也不要允许创建空的任务？没有意义
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

	// TODO: 创建无需提供namespace ， 可以使用默认的namespace
	// 构造Namespace
	if namespace == "" {
		t.Namespace = apis.NamespaceDefault
	} else {
		t.Namespace = namespace
	}

	t.Kind = "Task"
	t.APIVersion = "resources/v1"

	// 构造Labels
	if ts.Desc != nil && ts.Desc.Label != nil && len(ts.Desc.Label) > 0 {
		t.Labels = ts.Desc.Label
	} else {
		t.Labels = map[string]string{}
	}

	// 检查是否有scheduler标签，有的话不管，没有的话增加标记为cloud
	//if _, ok := t.Labels["scheduler"]; !ok {
	//	t.Labels["scheduler"] = "cloud"
	//}

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

	elapsed := time.Since(start) // 计算耗时
	logs.Tracef("-------------------------test time : CreateTask took %s", elapsed)
	return ft, nil
}

func (m *Manager) GetTask(name string, namespace string) (*apis.Task, error) {
	start := time.Now() // 记录开始时间
	c := m.GetTaskClient(namespace)

	a, err := c.Client.Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		logs.Errorf("Failed to get task: %v", err)
		return nil, fmt.Errorf("%w-%v", NotFound, err)
	}

	logs.Debugf("Get task: %v", a)

	elapsed := time.Since(start) // 计算耗时
	logs.Tracef("-------------------------test time : CreateTask took %s", elapsed)
	return a, nil
}

func (m *Manager) GetTasks(namespace string) (*apis.TaskList, error) {
	start := time.Now() // 记录开始时间
	c := m.GetTaskClient(namespace)

	g, err := c.Client.List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		logs.Errorf("Failed to get tasks: %v", err)
		return nil, err
	}

	logs.Debugf("Get tasks success.")

	elapsed := time.Since(start) // 计算耗时
	logs.Tracef("-------------------------test time : GetTasks took %s", elapsed)

	return g, nil
}

// FilterTasks 根据Label查询Tasks
func (m *Manager) FilterTasks(namespace string, labelSelector string) (*apis.TaskList, error) {
	start := time.Now() // 记录开始时间
	c := m.GetTaskClient(namespace)

	listOptions := metav1.ListOptions{
		LabelSelector: labelSelector,
	}

	d, err := c.Client.List(context.TODO(), listOptions)
	if err != nil {
		logs.Errorf("Failed to get tasks with labelselector: %s , error %v ", labelSelector, err)
		return nil, fmt.Errorf("%v-%w", InternalServerError, err)
	}

	logs.Infof("Get tasks with label success.")

	elapsed := time.Since(start) // 计算耗时
	logs.Tracef("-------------------------test time : FilterTasks took %s", elapsed)
	return d, nil
}

func (m *Manager) UpdateTask(name string, namespace string, a *apis.Task) (*apis.Task, error) {
	start := time.Now() // 记录开始时间
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
		return nil, fmt.Errorf("%w-%v", InternalServerError, updateErr)
	}

	logs.Debugf("Update task: %v", updatedTask)

	elapsed := time.Since(start) // 计算耗时
	logs.Tracef("-------------------------test time : UpdateTask took %s", elapsed)
	return updatedTask, nil

}

func (m *Manager) PatchTask(name string, namespace string, patchTask []byte) (*apis.Task, error) {
	start := time.Now() // 记录开始时间
	c := m.GetTaskClient(namespace)

	// 检查task是否存在
	_, err := m.GetTask(name, namespace)
	if err != nil {
		logs.Errorf("Get task %s error: %v , task not exist !", name, err)
		return nil, err
	}

	// 部分更新task
	patchedTask, err := c.Client.Patch(context.TODO(), name, types.StrategicMergePatchType, patchTask, metav1.PatchOptions{})
	if err != nil {
		logs.Errorf("patch task %s error: %v", name, err)
		return nil, err
	}

	logs.Debugf("Patch task: %v", patchedTask)

	elapsed := time.Since(start) // 计算耗时
	logs.Tracef("-------------------------test time : PatchTask took %s", elapsed)
	return patchedTask, nil
}

func (m *Manager) DeleteTask(name string, namespace string) error {
	start := time.Now() // 记录开始时间
	c := m.GetTaskClient(namespace)

	// 检查task是否存在
	task, err := m.GetTask(name, namespace)
	if err != nil {
		logs.Errorf("get task %s error: %v , task not exist ", name, err)
		return err
	}

	// if err != nil && !errors.Is(err, NotFound)

	// 删除task里面的所有groups
	for _, v := range task.Status.Groups {
		err := m.DeleteGroup(v.Name, v.Namespace)
		if err != nil && !errors.Is(err, NotFound) {
			logs.Errorf("delete groups in task error: %v", err)
			return err
		}
	}

	// 存在，删除
	err = c.Client.Delete(context.TODO(), name, metav1.DeleteOptions{})
	if err != nil {
		logs.Errorf("delete task %s error: %v", name, err)
		return fmt.Errorf("%w-%v", InternalServerError, err)
	}

	logs.Debugf("Delete task: %v", name)

	elapsed := time.Since(start) // 计算耗时
	logs.Tracef("-------------------------test time : DeleteTask took %s", elapsed)
	return nil
}

// TODO: 任务从task开始创建，应该提供一个递归删除task的接口

// CreateTaskWithLabels 创建带有label的task
// TODO: 将这部分全部替换为根据Spec里面的label创建，删除这部分
func (m *Manager) CreateTaskWithLabels(ts apis.TaskSpec, w *apis.Workflow, namespace string, uuid string, prefix string, labels map[string]string) (*apis.Task, error) {
	start := time.Now() // 记录开始时间
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

	elapsed := time.Since(start) // 计算耗时
	logs.Tracef("-------------------------test time : CreateTaskWithLabels took %s", elapsed)
	return ft, nil
}

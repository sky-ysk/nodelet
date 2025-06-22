package manager

import (
	"context"
	"errors"
	"fmt"
	"hit.edu/framework/pkg/apimachinery/types"
	apis "hit.edu/framework/pkg/apis/cores"
	metav1 "hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/component-base/logs"
	"time"
)

func (m *Manager) CreateWorkflow(ts apis.WorkflowSpec, namespace string, uuid string) (*apis.Workflow, error) {
	// 临时创建一个Workflow对象
	w := apis.Workflow{}

	// 构造名称
	w.Name = ts.Name + "-" + uuid
	prefix := ts.Name + "."

	// 构造Namespace
	if namespace == "" {
		w.Namespace = apis.NamespaceDefault
	} else {
		w.Namespace = namespace
	}

	w.Kind = "Workflow"
	w.APIVersion = "resources/v1"

	// 构造Labels
	if ts.Desc != nil && ts.Desc.Label != nil && len(ts.Desc.Label) > 0 {
		w.Labels = ts.Desc.Label
	} else {
		w.Labels = map[string]string{}
	}

	// 复制Spec
	w.Spec = ts

	// 构造Status
	w.Status = apis.WorkflowStatus{}

	// 记录Create时间
	w.Status.CreateAt = &apis.Time{Time: time.Now()}

	// 初始化状态
	// TODO: 检查创建状态Pending？unknown？
	// w.Status.Phase = apis.Pending
	w.Status.Phase = apis.Unknown
	w.Status.Tasks = map[string]apis.ObjectReference{}

	w.Labels["uuid"] = uuid

	// 根据Spec创建Runtimes
	tasks, err := m.CreateTasks(&w, namespace, uuid, prefix)
	if err != nil {
		return nil, err
	}

	// 根据生成的Runtime修改Workflow.Status.Tasks
	for _, r := range tasks {
		w.Status.Tasks[r.Spec.Name] = apis.ObjectReference{
			Name:            r.Name,
			Namespace:       r.Namespace,
			Kind:            r.Kind,
			ResourceVersion: r.ResourceVersion,
			UID:             apis.UID(r.UID),
		}
	}
	// 写入Client-Go中, 返回实际的Runtime
	c := m.GetWorkflowClient(w.Namespace)

	fw, err := c.Client.Create(context.TODO(), &w, metav1.CreateOptions{})
	if err != nil {
		logs.Errorf("Failed to create workflow: %v", err)
		return nil, fmt.Errorf("%w-%v", InternalServerError, err)
	}

	logs.Debugf("Created workflow: %v", fw)
	return fw, nil
}

func (m *Manager) GetWorkflow(name string, namespace string) (*apis.Workflow, error) {
	c := m.GetWorkflowClient(namespace)

	a, err := c.Client.Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		logs.Errorf("Failed to get workflow: %v", err)
		return nil, fmt.Errorf("%w-%v", NotFound, err)
	}

	logs.Debugf("Get workflow: %v", a)
	return a, nil
}

func (m *Manager) GetWorkflows(namespace string) (*apis.WorkflowList, error) {
	c := m.GetWorkflowClient(namespace)

	g, err := c.Client.List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		logs.Errorf("Failed to get workflows: %v", err)
		return nil, fmt.Errorf("%w-%v", InternalServerError, err)
	}

	logs.Debugf("Get workflows success. ")
	return g, nil
}

// FilterWorkflows 根据Label查询Workflows
func (m *Manager) FilterWorkflows(namespace string, labelSelector string) (*apis.WorkflowList, error) {
	c := m.GetWorkflowClient(namespace)

	listOptions := metav1.ListOptions{
		LabelSelector: labelSelector,
	}

	d, err := c.Client.List(context.TODO(), listOptions)
	if err != nil {
		logs.Errorf("Failed to get workflowss with labelselector: %s , error %v ", labelSelector, err)
		return nil, fmt.Errorf("%v-%w", InternalServerError, err)
	}

	logs.Infof("Get actions with label success.")
	return d, nil
}

func (m *Manager) UpdateWorkflow(name string, namespace string, a *apis.Workflow) (*apis.Workflow, error) {
	c := m.GetWorkflowClient(namespace)

	// 检查workflow是否存在
	_, err := m.GetWorkflow(name, namespace)
	if err != nil {
		logs.Errorf("Get workflow %s error: %v , workflow not exist !", name, err)
		return nil, err
	}

	// 存在更新workflow
	updatedWorkflow, updateErr := c.Client.Update(context.TODO(), a, metav1.UpdateOptions{})
	if updateErr != nil {
		logs.Errorf("Update workflow %s error: %v", name, updateErr)
		return nil, fmt.Errorf("%w-%v", InternalServerError, updateErr)
	}

	logs.Debugf("Update workflow: %v", updatedWorkflow)
	return updatedWorkflow, nil

}

func (m *Manager) PatchWorkflow(name string, namespace string, patchWorkflow []byte) (*apis.Workflow, error) {
	c := m.GetWorkflowClient(namespace)

	// 检查workflow是否存在
	_, err := m.GetWorkflow(name, namespace)
	if err != nil {
		logs.Errorf("Get workflow %s error: %v , workflow not exist !", name, err)
		return nil, err
	}

	// 部分更新workflow
	patchedWorkflow, err := c.Client.Patch(context.TODO(), name, types.StrategicMergePatchType, []byte(patchWorkflow), metav1.PatchOptions{})
	if err != nil {
		logs.Errorf("Patch workflow %s error: %v", name, err)
		return nil, fmt.Errorf("%w-%v", InternalServerError, err)
	}

	logs.Debugf("Patch workflow: %v", patchedWorkflow)
	return patchedWorkflow, nil
}

func (m *Manager) DeleteWorkflow(name string, namespace string) error {
	c := m.GetWorkflowClient(namespace)

	// 检查workflow是否存在
	workflow, err := m.GetWorkflow(name, namespace)
	if err != nil {
		logs.Errorf("get workflow %s error: %v , workflow not exist ", name, err)
		return err
	}

	// TODO：循环删除，里面部分资源已删除的错误处理
	// 删除workflow里面的所有task
	for _, v := range workflow.Status.Tasks {
		err := m.DeleteTask(v.Name, v.Namespace)
		if err != nil && !errors.Is(err, NotFound) {
			logs.Errorf("delete tasks in workflow error: %v", err)
			return err
		}
	}

	// 存在，删除
	err = c.Client.Delete(context.TODO(), name, metav1.DeleteOptions{})
	if err != nil {
		logs.Errorf("delete workflow %s error: %v", name, err)
		return fmt.Errorf("%w-%v", InternalServerError, err)
	}

	logs.Debugf("Delete workflow: %v", name)
	return nil
}

func (m *Manager) DeleteWorkflows(namespace string) error {
	// c := m.GetWorkflowClient(namespace)

	// 不知道怎么调用，示例只给了用Name
	//str := "NameSpace=" + namespace
	//lstOpts := metav1.ListOptions{
	//	FieldSelector: str,
	//}
	//
	//err := c.Client.DeleteCollection(context.TODO(), metav1.DeleteOptions{}, lstOpts)
	//if err != nil {
	//	logs.Errorf("Delete workflows failed: %v", err)
	//	return err
	//}

	// TODO：迭代删除，处理资源已经不存在的错误

	// 获取 workflows
	list, err := m.GetWorkflows(namespace)
	if err != nil {
		logs.Errorf("get workflows error: %v ", err)
		return err
	}

	// 循环删除
	for _, v := range list.Items {
		err := m.DeleteWorkflow(v.Name, v.Namespace)
		if err != nil {
			logs.Errorf("delete workflow %s error: %v", v.Name, err)
			return err
		}
	}

	logs.Debugf("Delete workflows in %s success.", namespace)
	return nil
}

// CreateWorkflowWithLabels 创建带有label的task
// TODO: 将这部分全部替换为根据Spec里面的label创建，删除这部分
func (m *Manager) CreateWorkflowWithLabels(ts apis.WorkflowSpec, namespace string, uuid string, labels map[string]string) (*apis.Workflow, error) {
	// 临时创建一个Workflow对象
	w := apis.Workflow{}
	// 构造名称

	w.Name = ts.Name + "-" + uuid
	prefix := ts.Name + "."

	// 构造Namespace
	if namespace == "" {
		w.Namespace = apis.NamespaceDefault
	} else {
		w.Namespace = namespace
	}

	w.Kind = "Workflow"
	w.APIVersion = "resources/v1"

	// 构造Labels
	w.Labels = labels

	// 复制Spec
	w.Spec = ts

	// 构造Status
	w.Status = apis.WorkflowStatus{}

	// 记录Create时间
	w.Status.CreateAt = &apis.Time{Time: time.Now()}

	// 初始化状态
	w.Status.Phase = apis.Pending
	w.Status.Tasks = map[string]apis.ObjectReference{}

	w.Labels["uuid"] = uuid

	// 根据Spec创建Runtimes
	tasks, err := m.CreateTasks(&w, namespace, uuid, prefix)
	if err != nil {
		return nil, err
	}

	// 根据生成的Runtime修改Workflow.Status.Tasks
	for _, r := range tasks {
		w.Status.Tasks[r.Spec.Name] = apis.ObjectReference{
			Name:            r.Name,
			Namespace:       r.Namespace,
			Kind:            r.Kind,
			ResourceVersion: r.ResourceVersion,
			UID:             apis.UID(r.UID),
		}
	}
	// 写入Client-Go中, 返回实际的Runtime
	c := m.GetWorkflowClient(w.Namespace)

	fw, err := c.Client.Create(context.TODO(), &w, metav1.CreateOptions{})
	if err != nil {
		logs.Errorf("Failed to create workflow: %v", err)
		return nil, err
	}

	logs.Debugf("Created workflow: %v", fw)
	return fw, nil
}

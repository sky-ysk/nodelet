package manager

import (
	"context"
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
	w.Labels = map[string]string{}
	
	// 复制Spec
	w.Spec = ts
	
	// 构造Status
	w.Status = apis.WorkflowStatus{}
	
	// 记录Create时间
	w.Status.CreateAt = &apis.Time{time.Now()}
	
	// 初始化状态
	w.Status.Phase = apis.Pending
	w.Status.Tasks = map[string]apis.ObjectReference{}
	
	w.Labels["uuid"] = uuid
	
	// 根据Spec创建Runtimes
	groups, err := m.CreateTasks(&w, namespace, uuid, prefix)
	if err != nil {
		return nil, err
	}
	
	// 根据生成的Runtime修改Workflow.Status.Tasks
	for _, r := range groups {
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
	}
	logs.Debugf("Created workflow: %v", fw)
	
	return fw, nil // 应当返回实际的ft
}

func (m *Manager) GetWorkflow(name string, namespace string) (*apis.Workflow, error) {
	c := m.GetWorkflowClient(namespace)
	a, err := c.Client.Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	} else {
		return a, nil
	}
}

func (m *Manager) GetWorkflows(namespace string) (*apis.WorkflowList, error) {
	c := m.GetWorkflowClient(namespace)
	g, err := c.Client.List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	} else {
		return g, nil
	}
}

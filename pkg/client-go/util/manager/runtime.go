package manager

import (
	"context"
	"hit.edu/framework/pkg/apimachinery/types"
	apis "hit.edu/framework/pkg/apis/cores"
	metav1 "hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/component-base/logs"
	"time"
)

// CreateRuntimes 根据ActionSpec创建Runtime
func (m *Manager) CreateRuntimes(a *apis.Action, namespace string, uuid string, prefix string) ([]*apis.Runtime, error) {
	var runtimes []*apis.Runtime
	// 遍历所有的Runtime Spec
	for _, rs := range a.Spec.Runtimes {
		r, err := m.CreateRuntime(rs, a, namespace, uuid, prefix)
		if err != nil {
			return nil, err
		}
		runtimes = append(runtimes, r)
	}
	return runtimes, nil
}

// CreateRuntime 创建单个Runtime
func (m *Manager) CreateRuntime(rs apis.RuntimeSpec, a *apis.Action, namespace string, uuid string, prefix string) (*apis.Runtime, error) {
	// 临时创建一个Runtime对象
	r := apis.Runtime{}

	// 构造名称
	if a != nil {
		r.Name = prefix + rs.Name + "-" + uuid
		// 有父亲节点，则需要继承Prefix
		prefix = prefix + rs.Name + "."
	} else {
		r.Name = rs.Name + "-" + uuid
		// 没有父亲节点，则需要本地构造prefix
		prefix = rs.Name + "."
	}

	// 构造Namespace
	if namespace == "" {
		r.Namespace = apis.NamespaceDefault
	} else {
		r.Namespace = namespace
	}

	r.Kind = "Runtime"
	r.APIVersion = "resources/v1"

	// 构造Labels
	r.Labels = make(map[string]string)

	// 复制Spec
	r.Spec = rs

	// 构造Status
	r.Status = apis.RuntimeStatus{}

	// 记录Create时间
	r.Status.CreateAt = &apis.Time{Time: time.Now()}

	// 初始化状态
	r.Status.Phase = apis.Unknown

	// 打上Label, 当前任务属于哪个action和uuid域
	if a != nil {
		r.Status.Belong = &apis.ObjectReference{
			Name:            a.Name,
			Namespace:       a.Namespace,
			Kind:            a.Kind,
			ResourceVersion: a.ResourceVersion,
			UID:             apis.UID(uuid),
		}
		r.Labels["belong"] = a.Name
	}
	r.Labels["uuid"] = uuid

	// 写入Client-Go中, 返回实际的Runtime
	c := m.GetRuntimeClient(r.Namespace)

	fr, err := c.Client.Create(context.TODO(), &r, metav1.CreateOptions{})
	if err != nil {
		logs.Errorf("Failed to create runtime: %v", err)
		return nil, err
	}
	logs.Debugf("Created runtime: %v", fr)

	//
	return fr, nil
}

func (m *Manager) GetRuntime(name string, namespace string) (*apis.Runtime, error) {
	c := m.GetRuntimeClient(namespace)
	a, err := c.Client.Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		logs.Errorf("Failed to get runtime: %v", err)
		return nil, err
	}

	//
	logs.Debugf("Get runtime: %v", a)
	return a, nil
}

func (m *Manager) GetRuntimes(namespace string) (*apis.RuntimeList, error) {
	c := m.GetRuntimeClient(namespace)
	g, err := c.Client.List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		logs.Errorf("Failed to get runtime: %v", err)
		return nil, err
	}

	logs.Debugf("Get runtimes success.")
	return g, nil
}

func (m *Manager) UpdateRuntime(name string, namespace string, a *apis.Runtime) (*apis.Runtime, error) {
	c := m.GetRuntimeClient(namespace)

	// 检查runtime是否存在
	_, err := m.GetRuntime(name, namespace)
	if err != nil {
		logs.Errorf("Get runtime %s error: %v , runtime not exist !", name, err)
		return nil, err
	}

	// 存在更新runtime
	updatedRuntime, updateErr := c.Client.Update(context.TODO(), a, metav1.UpdateOptions{})
	if updateErr != nil {
		logs.Errorf("Update runtime %s error: %v", name, updateErr)
		return nil, updateErr
	}

	logs.Debugf("Update runtime: %v", updatedRuntime)
	return updatedRuntime, nil

}

func (m *Manager) PatchRuntime(name string, namespace string, patchRuntime string) (*apis.Runtime, error) {
	c := m.GetRuntimeClient(namespace)

	// 检查runtime是否存在
	_, err := m.GetRuntime(name, namespace)
	if err != nil {
		logs.Errorf("Get runtime %s error: %v , runtime not exist !", name, err)
		return nil, err
	}

	// 部分更新runtime
	patchedRuntime, err := c.Client.Patch(context.TODO(), name, types.StrategicMergePatchType, []byte(patchRuntime), metav1.PatchOptions{})
	if err != nil {
		logs.Errorf("patch runtime %s error: %v", name, err)
		return nil, err
	}

	logs.Debugf("Patch runtime: %v", patchedRuntime)
	return patchedRuntime, nil
}

func (m *Manager) DeleteRuntime(name string, namespace string) error {
	c := m.GetRuntimeClient(namespace)

	// 检查runtime是否存在
	_, err := m.GetRuntime(name, namespace)
	if err != nil {
		logs.Errorf("get runtime %s error: %v , runtime not exist ", name, err)
		return err
	}

	// 存在，删除
	err = c.Client.Delete(context.TODO(), name, metav1.DeleteOptions{})
	if err != nil {
		logs.Errorf("delete runtime %s error: %v", name, err)
		return err
	}

	logs.Debugf("Delete runtime: %v", name)
	return nil
}

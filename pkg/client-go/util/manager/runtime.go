package manager

import (
	"context"
	apis "hit.edu/framework/pkg/apis/cores"
	metav1 "hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/component-base/logs"
	"time"
)

// 根据ActionSpec创建Runtime
func (m *Manager) CreateRuntimes(a *apis.Action, namespace string, uuid string, prefix string) ([]*apis.Runtime, error) {
	runtimes := []*apis.Runtime{}
	// 遍历所有的Runtime Spec
	for _, rs := range a.Spec.Runtimes {
		r, err := m.CreateRuntime(rs, a, namespace, uuid, prefix)
		if err != nil {
			return nil, err
		}
		// TODO: 返回的应该是已经创建的Runtime
		runtimes = append(runtimes, r) // r替换成实际的fr
	}
	return runtimes, nil
}

// 创建单个Runtime
func (m *Manager) CreateRuntime(rs apis.RuntimeSpec, a *apis.Action, namespace string, uuid string, prefix string) (*apis.Runtime, error) {
	// 临时创建一个Runtime对象
	r := apis.Runtime{}
	// 构造名称
	if a != nil {
		r.Name = prefix + rs.Name + "-" + uuid
	} else {
		r.Name = rs.Name + "-" + uuid
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
	r.Status.CreateAt = &apis.Time{time.Now()}

	// 初始化状态
	r.Status.Phase = apis.Pending

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
	}
	logs.Debugf("Created runtime: %v", fr)

	//
	return fr, nil // 返回fr
}

func (m *Manager) GetRuntime(name string, namespace string) (*apis.Runtime, error) {
	c := m.GetRuntimeClient(namespace)
	a, err := c.Client.Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	} else {
		return a, nil
	}
}

func (m *Manager) GetRuntimes(namespace string) (*apis.RuntimeList, error) {
	c := m.GetRuntimeClient(namespace)
	g, err := c.Client.List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	} else {
		return g, nil
	}
}

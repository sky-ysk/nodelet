package manager

import (
	"context"
	apis "hit.edu/framework/pkg/apis/cores"
	metav1 "hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/component-base/logs"
	"strings"
	"time"
)

// 根据GroupSpec创建Action
func (m *Manager) CreateGroups(g *apis.Task, namespace string, uuid string, prefix string) ([]*apis.Group, error) {
	groups := []*apis.Group{}
	for _, as := range g.Spec.Groups {
		// 默认情况下，创建Groups不填充Actions
		a, err := m.CreateGroupWithoutActions(as, g, namespace, uuid, prefix)
		//a, err := m.CreateGroup(as, g, namespace, uuid, prefix)
		if err != nil {
			return nil, err
		}
		// TODO: 返回的应该是已经创建的Runtime
		groups = append(groups, a) // a替换成实际的fa
	}

	return groups, nil
}

// 创建Group，不创建Actions
func (m *Manager) CreateGroupWithoutActions(gs apis.GroupSpec, t *apis.Task, namespace string, uuid string, prefix string) (*apis.Group, error) {
	// 临时创建一个Group对象
	g := apis.Group{}
	// 构造名称
	if t != nil {
		g.Name = prefix + gs.Name + "-" + uuid
		prefix = prefix + gs.Name + "."
	} else {
		g.Name = gs.Name + "-" + uuid
		prefix = gs.Name + "."
	}

	// 构造Namespace
	if namespace == "" {
		g.Namespace = apis.NamespaceDefault
	} else {
		g.Namespace = namespace
	}

	g.Kind = "Group"
	g.APIVersion = "resources/v1"

	// 构造Labels
	g.Labels = map[string]string{}

	// 复制Spec
	g.Spec = gs

	// 构造Status
	g.Status = apis.GroupStatus{}

	// 记录Create时间
	g.Status.CreateAt = &apis.Time{time.Now()}

	// 初始化状态
	g.Status.Phase = apis.Pending

	// 打上Label, 当前任务属于哪个Group和uuid域
	if t != nil {
		g.Status.Belong = &apis.ObjectReference{
			Name:            t.Name,
			Namespace:       t.Namespace,
			Kind:            t.Kind,
			ResourceVersion: t.ResourceVersion,
			UID:             apis.UID(uuid),
		}
		g.Labels["belong"] = t.Name
	}
	g.Labels["uuid"] = uuid

	// 写入Client-Go中, 返回实际的Runtime
	c := m.GetGroupClient(g.Namespace)

	fg, err := c.Client.Create(context.TODO(), &g, metav1.CreateOptions{})
	if err != nil {
		logs.Errorf("Failed to create group: %v", err)
	}
	logs.Debugf("Created group: %v", fg)

	return fg, nil // TODO: 应当返回实际的fa
}

// 填充Group的Actions
func (m *Manager) FillGroupWithActions(g *apis.Group) (*apis.Group, error) {
	// 根据当前Group的Name来获取Prefix
	// Prefix固定为g.Name - uuid的部分
	suffix := "-" + g.Labels["uuid"]
	prefix := strings.TrimSuffix(g.Name, suffix) + "."

	// 根据Spec创建Runtimes
	actions, err := m.CreateActions(g, g.Namespace, g.Labels["uuid"], prefix)
	if err != nil {
		return nil, err
	}
	g.Status.Actions = map[string]apis.ObjectReference{}

	// 根据生成的Runtime修改Group.Status.Actions
	for _, r := range actions {
		g.Status.Actions[r.Spec.Name] = apis.ObjectReference{
			Name:            r.Name,
			Namespace:       r.Namespace,
			Kind:            r.Kind,
			ResourceVersion: r.ResourceVersion,
			UID:             apis.UID(r.UID),
		}
	}

	// 写入Client-Go中, 返回实际的Runtime
	c := m.GetGroupClient(g.Namespace)

	fg, err := c.Client.Update(context.TODO(), g, metav1.UpdateOptions{})
	if err != nil {
		logs.Errorf("Failed to fill group: %v", err)
	}
	logs.Debugf("Created fill: %v", fg)

	return fg, nil // TODO: 应当返回实际的fa
}

// 创建完整的Group
func (m *Manager) CreateGroup(gs apis.GroupSpec, t *apis.Task, namespace string, uuid string, prefix string) (*apis.Group, error) {
	// 临时创建一个Group对象
	g := apis.Group{}
	// 构造名称
	if t != nil {
		g.Name = prefix + gs.Name + "-" + uuid
		prefix = prefix + gs.Name + "."
	} else {
		g.Name = gs.Name + "-" + uuid
		prefix = gs.Name + "."
	}

	// 构造Namespace
	if namespace == "" {
		g.Namespace = apis.NamespaceDefault
	} else {
		g.Namespace = namespace
	}

	g.Kind = "Group"
	g.APIVersion = "resources/v1"

	// 构造Labels
	g.Labels = map[string]string{}

	// 复制Spec
	g.Spec = gs

	// 构造Status
	g.Status = apis.GroupStatus{}

	// 记录Create时间
	g.Status.CreateAt = &apis.Time{time.Now()}

	// 初始化状态
	g.Status.Phase = apis.Pending
	g.Status.Actions = map[string]apis.ObjectReference{}

	// 打上Label, 当前任务属于哪个Group和uuid域
	if t != nil {
		g.Status.Belong = &apis.ObjectReference{
			Name:            t.Name,
			Namespace:       t.Namespace,
			Kind:            t.Kind,
			ResourceVersion: t.ResourceVersion,
			UID:             apis.UID(uuid),
		}
		g.Labels["belong"] = t.Name
	}
	g.Labels["uuid"] = uuid

	// 根据Spec创建Runtimes
	actions, err := m.CreateActions(&g, namespace, uuid, prefix)
	if err != nil {
		return nil, err
	}

	// 根据生成的Runtime修改Group.Status.Actions
	for _, r := range actions {
		g.Status.Actions[r.Spec.Name] = apis.ObjectReference{
			Name:            r.Name,
			Namespace:       r.Namespace,
			Kind:            r.Kind,
			ResourceVersion: r.ResourceVersion,
			UID:             apis.UID(r.UID),
		}
	}
	// 写入Client-Go中, 返回实际的Runtime
	c := m.GetGroupClient(g.Namespace)

	fg, err := c.Client.Create(context.TODO(), &g, metav1.CreateOptions{})
	if err != nil {
		logs.Errorf("Failed to create group: %v", err)
	}
	logs.Debugf("Created group: %v", fg)

	return fg, nil // TODO: 应当返回实际的fa
}

func (m *Manager) GetGroup(name string, namespace string) (*apis.Group, error) {
	c := m.GetGroupClient(namespace)
	a, err := c.Client.Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	} else {
		return a, nil
	}
}

func (m *Manager) GetGroups(namespace string) (*apis.GroupList, error) {
	c := m.GetGroupClient(namespace)
	g, err := c.Client.List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	} else {
		return g, nil
	}
}

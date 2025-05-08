package manager

import (
	"context"
	"hit.edu/framework/pkg/apimachinery/types"
	apis "hit.edu/framework/pkg/apis/cores"
	metav1 "hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/component-base/logs"
	"strings"
	"time"
)

// 根据TaskSpec创建Groups
func (m *Manager) CreateGroups(g *apis.Task, namespace string, uuid string, prefix string) ([]*apis.Group, error) {
	var groups []*apis.Group
	for _, as := range g.Spec.Groups {
		// 默认情况下，创建Groups不填充Actions
		// a, err := m.CreateGroupWithoutActions(as, g, namespace, uuid, prefix)
		a, err := m.CreateGroup(as, g, namespace, uuid, prefix)
		if err != nil {
			return nil, err
		}
		groups = append(groups, a)
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
	g.Status.CreateAt = &apis.Time{Time: time.Now()}

	// 初始化状态
	g.Status.Phase = apis.Pending

	// 打上Label, 当前任务属于哪个task和uuid域
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
		return nil, err
	}

	logs.Infof("Created group: %v", fg)
	return fg, nil
}

// 创建带有label的Group
func (m *Manager) CreateGroupWithLabels(gs apis.GroupSpec, t *apis.Task, namespace string, uuid string, prefix string, labels map[string]string) (*apis.Group, error) {
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
	g.Labels = labels

	// 复制Spec
	g.Spec = gs

	// 构造Status
	g.Status = apis.GroupStatus{}

	// 记录Create时间
	g.Status.CreateAt = &apis.Time{Time: time.Now()}

	// 初始化状态
	g.Status.Phase = apis.Unknown
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

	// 根据Spec创建Actions
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
		return nil, err
	}

	logs.Debugf("Created group: %v", fg)
	return fg, nil
}

// 填充Group的Actions
func (m *Manager) FillGroupWithActions(g *apis.Group) (*apis.Group, error) {
	// 根据当前Group的Name来获取Prefix
	// Prefix固定为g.Name - uuid的部分
	suffix := "-" + g.Labels["uuid"]
	prefix := strings.TrimSuffix(g.Name, suffix) + "."

	// 根据Spec创建Actions
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

	// 写入Client-Go中, 返回实际的 Group
	c := m.GetGroupClient(g.Namespace)

	fg, err := c.Client.Update(context.TODO(), g, metav1.UpdateOptions{})
	if err != nil {
		logs.Errorf("Failed to fill group: %v", err)
		return nil, err
	}

	logs.Infof("Created fill: %v", fg)
	return fg, nil
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
	g.Status.Phase = apis.Unknown
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

	// 根据Spec创建Actions
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
		return nil, err
	}

	logs.Debugf("Created group: %v", fg)
	return fg, nil
}

func (m *Manager) GetGroup(name string, namespace string) (*apis.Group, error) {
	c := m.GetGroupClient(namespace)
	a, err := c.Client.Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		logs.Errorf("Failed to get group: %v", err)
		return nil, err
	}

	logs.Debugf("Get group: %v", a)
	return a, nil
}

func (m *Manager) GetGroups(namespace string) (*apis.GroupList, error) {
	c := m.GetGroupClient(namespace)
	g, err := c.Client.List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		logs.Errorf("Failed to get groups: %v", err)
		return nil, err
	}

	logs.Debugf("Get groups success")
	return g, nil
}

// 根据Label查询Groups
func (m *Manager) FilterGroups(namespace string, labelSelector string) (*apis.GroupList, error) {
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

	c := m.GetGroupClient(namespace)
	d, err := c.Client.List(context.TODO(), listOptions)
	if err != nil {
		return nil, err
	}
	return d, nil
}
func (m *Manager) UpdateGroup(name string, namespace string, a *apis.Group) (*apis.Group, error) {
	c := m.GetGroupClient(namespace)

	// 检查group是否存在
	_, err := m.GetGroup(name, namespace)
	if err != nil {
		logs.Errorf("Get group %s error: %v , group not exist !", name, err)
		return nil, err
	}

	// 存在更新group
	updatedGroup, updateErr := c.Client.Update(context.TODO(), a, metav1.UpdateOptions{})
	if updateErr != nil {
		logs.Errorf("Update group %s error: %v", name, updateErr)
		return nil, updateErr
	}

	logs.Debugf("Update group: %v", updatedGroup)
	return updatedGroup, nil

}

func (m *Manager) PatchGroup(name string, namespace string, patchGroup []byte) (*apis.Group, error) {
	c := m.GetGroupClient(namespace)

	// 检查group是否存在
	_, err := m.GetGroup(name, namespace)
	if err != nil {
		logs.Errorf("Get group %s error: %v , group not exist !", name, err)
		return nil, err
	}

	// 部分更新group
	patchedGroup, err := c.Client.Patch(context.TODO(), name, types.StrategicMergePatchType, patchGroup, metav1.PatchOptions{})
	if err != nil {
		logs.Errorf("Patch group %s error: %v", name, err)
		return nil, err
	}

	logs.Debugf("Patch group: %v", patchedGroup)
	return patchedGroup, nil
}

func (m *Manager) DeleteGroup(name string, namespace string) error {
	c := m.GetGroupClient(namespace)

	// 检查group是否存在
	group, err := m.GetGroup(name, namespace)
	if err != nil {
		logs.Errorf("get group %s error: %v , group not exist ", name, err)
		return err
	}

	// 删除group里面的所有action
	for _, v := range group.Status.Actions {
		err := m.DeleteAction(v.Name, v.Namespace)
		if err != nil {
			return err
		}
	}

	// 存在，删除
	err = c.Client.Delete(context.TODO(), name, metav1.DeleteOptions{})
	if err != nil {
		logs.Errorf("delete group %s error: %v", name, err)
		return err
	}

	logs.Debugf("delete group: %v", name)
	return nil
}

package manager

import (
	"context"
	"hit.edu/framework/pkg/apimachinery/types"
	apis "hit.edu/framework/pkg/apis/cores"
	metav1 "hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/component-base/logs"
	"time"
)

// CreateActions 根据GroupSpec创建Action
func (m *Manager) CreateActions(g *apis.Group, namespace string, uuid string, prefix string) ([]*apis.Action, error) {
	var actions []*apis.Action
	for _, as := range g.Spec.Actions {
		a, err := m.CreateAction(as, g, namespace, uuid, prefix)
		if err != nil {
			return nil, err
		}
		actions = append(actions, a)
	}
	return actions, nil
}

// CreateActionWithLabels 创建带有label的Action
func (m *Manager) CreateActionWithLabels(as apis.ActionSpec, g *apis.Group, namespace string, uuid string, prefix string, labels map[string]string) (*apis.Action, error) {
	// 临时创建一个Action对象
	a := apis.Action{}

	// 构造名称
	if g != nil {
		a.Name = prefix + as.Name + "-" + uuid
		// 有父亲节点，则需要继承Prefix
		prefix = prefix + as.Name + "."
	} else {
		a.Name = as.Name + "-" + uuid
		// 没有父亲节点，则需要本地构造prefix
		prefix = as.Name + "."
	}

	// 构造Namespace
	if namespace == "" {
		a.Namespace = apis.NamespaceDefault
	} else {
		a.Namespace = namespace
	}

	a.Kind = "Action"
	a.APIVersion = "resources/v1"

	// 构造Labels
	a.Labels = labels

	// 复制Spec
	a.Spec = as

	// 构造Status
	a.Status = apis.ActionStatus{}

	// 记录Create时间
	a.Status.CreateAt = &apis.Time{Time: time.Now()}

	// 初始化状态
	a.Status.Phase = apis.Unknown
	a.Status.Runtimes = map[string]apis.ObjectReference{}

	// 打上Label, 当前任务属于哪个Group和uuid域
	if g != nil {
		a.Labels["belong"] = g.Name
		a.Status.Belong = &apis.ObjectReference{
			Name:            g.Name,
			Namespace:       g.Namespace,
			Kind:            g.Kind,
			ResourceVersion: g.ResourceVersion,
			UID:             apis.UID(uuid),
		}
	}
	a.Labels["uuid"] = uuid

	// 根据Spec创建Runtimes
	runtimes, err := m.CreateRuntimes(&a, namespace, uuid, prefix)
	if err != nil {
		return nil, err
	}

	// 根据生成的Runtime修改Action.Status.Runtimes
	for _, r := range runtimes {
		a.Status.Runtimes[r.Spec.Name] = apis.ObjectReference{
			Name:            r.Name,
			Namespace:       r.Namespace,
			Kind:            r.Kind,
			ResourceVersion: r.ResourceVersion,
			UID:             apis.UID(r.UID),
		}
	}

	// 写入Client-Go中, 返回实际的Action
	c := m.GetActionClient(a.Namespace)

	fa, err := c.Client.Create(context.TODO(), &a, metav1.CreateOptions{})
	if err != nil {
		logs.Errorf("Failed to create action: %v", err)
		return nil, err
	}

	//
	logs.Debugf("Created action: %v", fa)
	return fa, nil
}

func (m *Manager) CreateAction(as apis.ActionSpec, g *apis.Group, namespace string, uuid string, prefix string) (*apis.Action, error) {
	// 临时创建一个Action对象
	a := apis.Action{}

	// 构造名称
	if g != nil {
		a.Name = prefix + as.Name + "-" + uuid
		// 有父亲节点，则需要继承Prefix
		prefix = prefix + as.Name + "."
	} else {
		a.Name = as.Name + "-" + uuid
		// 没有父亲节点，则需要本地构造prefix
		prefix = as.Name + "."
	}

	// 构造Namespace
	if namespace == "" {
		a.Namespace = apis.NamespaceDefault
	} else {
		a.Namespace = namespace
	}

	a.Kind = "Action"
	a.APIVersion = "resources/v1"

	// 构造Labels
	if as.Desc != nil && as.Desc.Label != nil && len(as.Desc.Label) > 0 {
		a.Labels = as.Desc.Label
	} else {
		a.Labels = map[string]string{}
	}

	// 复制Spec
	a.Spec = as

	// 构造Status
	a.Status = apis.ActionStatus{}

	// 记录Create时间
	a.Status.CreateAt = &apis.Time{time.Now()}

	// 初始化状态
	a.Status.Phase = apis.Unknown
	a.Status.Runtimes = map[string]apis.ObjectReference{}

	// 打上Label, 当前任务属于哪个Group和uuid域
	if g != nil {
		a.Labels["belong"] = g.Name
		a.Status.Belong = &apis.ObjectReference{
			Name:            g.Name,
			Namespace:       g.Namespace,
			Kind:            g.Kind,
			ResourceVersion: g.ResourceVersion,
			UID:             apis.UID(uuid),
		}
	}
	a.Labels["uuid"] = uuid

	// 根据Spec创建Runtimes
	runtimes, err := m.CreateRuntimes(&a, namespace, uuid, prefix)
	if err != nil {
		return nil, err
	}

	// 根据生成的Runtime修改Action.Status.Runtimes
	for _, r := range runtimes {
		a.Status.Runtimes[r.Spec.Name] = apis.ObjectReference{
			Name:            r.Name,
			Namespace:       r.Namespace,
			Kind:            r.Kind,
			ResourceVersion: r.ResourceVersion,
			UID:             apis.UID(r.UID),
		}
	}

	// 写入Client-Go中, 返回实际的Action
	c := m.GetActionClient(a.Namespace)

	fa, err := c.Client.Create(context.TODO(), &a, metav1.CreateOptions{})
	if err != nil {
		logs.Errorf("Failed to create action: %v", err)
		return nil, err
	}

	//
	logs.Debugf("Created action: %v", fa)
	return fa, nil
}

func (m *Manager) GetAction(name string, namespace string) (*apis.Action, error) {
	c := m.GetActionClient(namespace)
	a, err := c.Client.Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		logs.Errorf("Failed to get action: %v", err)
		return nil, err
	}

	//
	logs.Debugf("Get action: %v", a)
	return a, nil
}

func (m *Manager) GetActions(namespace string) (*apis.ActionList, error) {
	c := m.GetActionClient(namespace)
	a, err := c.Client.List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		logs.Errorf("Failed to get actions: %v", err)
		return nil, err
	}

	//
	logs.Debugf("Get actions success.")
	return a, nil
}

// 根据Label查询Actions
func (m *Manager) FilterActions(namespace string, labelSelector string) (*apis.ActionList, error) {
	//labelSelector := ""
	//for i, l := range label {
	//	if i == 0 {
	//		labelSelector = l
	//	} else {
	//		labelSelector += "," + l
	//	}
	//}

	logs.Infof("Filter actions by label selector: %v", labelSelector)

	listOptions := metav1.ListOptions{
		LabelSelector: labelSelector,
	}

	c := m.GetActionClient(namespace)
	d, err := c.Client.List(context.TODO(), listOptions)
	if err != nil {
		logs.Errorf("Failed to get actions with labelselector: %s , error %v ", labelSelector, err)
		return nil, err
	}
	logs.Infof("Get actionslist : %v", d)
	logs.Infof("Get actions with label success.")
	return d, nil
}

func (m *Manager) UpdateAction(name string, namespace string, a *apis.Action) (*apis.Action, error) {
	c := m.GetActionClient(namespace)

	// 检查action是否存在
	_, err := m.GetAction(name, namespace)
	if err != nil {
		logs.Errorf("Get action %s error: %v , action not exist !", name, err)
		return nil, err
	}

	// 存在更新action
	updatedAction, updateErr := c.Client.Update(context.TODO(), a, metav1.UpdateOptions{})
	if updateErr != nil {
		logs.Errorf("Update action %s error: %v", name, updateErr)
		return nil, updateErr
	}

	//
	logs.Debugf("Update action: %v", updatedAction)
	return updatedAction, nil

}

func (m *Manager) PatchAction(name string, namespace string, patchAction []byte) (*apis.Action, error) {
	c := m.GetActionClient(namespace)

	// 检查action是否存在
	_, err := m.GetAction(name, namespace)
	if err != nil {
		logs.Errorf("Get action %s error: %v , action not exist !", name, err)
		return nil, err
	}

	// 部分更新action
	patchedAction, err := c.Client.Patch(context.TODO(), name, types.StrategicMergePatchType, patchAction, metav1.PatchOptions{})
	if err != nil {
		logs.Errorf("patch action %s error: %v", name, err)
		return nil, err
	}

	//
	logs.Debugf("patched action : %v ", patchedAction)
	return patchedAction, nil
}

func (m *Manager) DeleteAction(name string, namespace string) error {
	c := m.GetActionClient(namespace)

	// 检查action是否存在
	action, err := m.GetAction(name, namespace)
	if err != nil {
		logs.Errorf("get action %s error: %v , action not exist ", name, err)
		return err
	}

	// 删除action里面的所有group
	for _, v := range action.Status.Runtimes {
		err := m.DeleteRuntime(v.Name, v.Namespace)
		if err != nil {
			return err
		}
	}

	// 存在，删除
	err = c.Client.Delete(context.TODO(), name, metav1.DeleteOptions{})
	if err != nil {
		logs.Errorf("delete action %s error: %v", name, err)
		return err
	}

	//
	logs.Debugf("Delete action %v ", err)
	return nil
}

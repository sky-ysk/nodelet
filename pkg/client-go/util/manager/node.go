package manager

import (
	"context"
	"fmt"
	"hit.edu/framework/pkg/apimachinery/types"
	apis "hit.edu/framework/pkg/apis/cores"
	metav1 "hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/component-base/logs"
)

func (m *Manager) CreateNode(ns apis.NodeSpec, namespace string, uuid string) (*apis.Node, error) {
	// 临时创建一个Node对象
	n := apis.Node{}

	// 构造名称
	n.Name = ns.NodeName

	// 构造Namespace
	if namespace == "" {
		n.Namespace = apis.NamespaceDefault
	} else {
		n.Namespace = namespace
	}

	n.Kind = "Node"
	n.APIVersion = "resources/v1"

	// 构造Labels
	if ns.Desc != nil && ns.Desc.Label != nil && len(ns.Desc.Label) > 0 {
		n.Labels = ns.Desc.Label
	} else {
		n.Labels = map[string]string{}
	}

	// 复制Spec
	n.Spec = ns

	// 构造Status
	n.Status = apis.NodeStatus{}

	// n.Labels["uuid"] = uuid

	c := m.GetNodeClient(n.Namespace)

	node, err := c.Client.Create(context.TODO(), &n, metav1.CreateOptions{})
	if err != nil {
		logs.Errorf("Failed to create node: %v", err)
		return nil, err
	}

	logs.Debugf("Created node: %v", node)
	return node, nil
}

func (m *Manager) GetNode(name string, namespace string) (*apis.Node, error) {
	c := m.GetNodeClient(namespace)

	n, err := c.Client.Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		logs.Errorf("Failed to get node: %v", err)
		return nil, fmt.Errorf("%w-%v", NotFound, err)
	}

	//
	logs.Debugf("Get node: %v", n)
	return n, nil
}

func (m *Manager) GetNodes(namespace string) (*apis.NodeList, error) {
	c := m.GetNodeClient(namespace)

	n, err := c.Client.List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		logs.Errorf("Failed to get nodes: %v", err)
		return nil, fmt.Errorf("%w-%v", InternalServerError, err)
	}

	logs.Debugf("Get nodes success. ")
	return n, nil
}

// FilterNodes 根据Label查询nodes
func (m *Manager) FilterNodes(namespace string, labelSelector string) (*apis.NodeList, error) {
	c := m.GetNodeClient(namespace)

	listOptions := metav1.ListOptions{
		LabelSelector: labelSelector,
	}

	ns, err := c.Client.List(context.TODO(), listOptions)
	if err != nil {
		logs.Errorf("Failed to get nodes with labelselector: %s , error %v ", labelSelector, err)
		return nil, fmt.Errorf("%v-%w", InternalServerError, err)
	}

	logs.Infof("Get nodes with label success.")
	return ns, nil
}

func (m *Manager) UpdateNode(name string, namespace string, n *apis.Node) (*apis.Node, error) {
	c := m.GetNodeClient(namespace)

	// 检查node是否存在
	_, err := m.GetNode(name, namespace)
	if err != nil {
		logs.Errorf("Get node %s error: %v , node not exist !", name, err)
		return nil, err
	}

	// 存在更新node
	updatedNode, updateErr := c.Client.Update(context.TODO(), n, metav1.UpdateOptions{})
	if updateErr != nil {
		logs.Errorf("Update node %s error: %v", name, updateErr)
		return nil, fmt.Errorf("%w-%v", InternalServerError, updateErr)
	}

	//
	logs.Debugf("Update node: %v", updatedNode)
	return updatedNode, nil

}

func (m *Manager) PatchNode(name string, namespace string, patchNode []byte) (*apis.Node, error) {
	c := m.GetNodeClient(namespace)

	// 检查node是否存在
	_, err := m.GetNode(name, namespace)
	if err != nil {
		logs.Errorf("Get node %s error: %v , node not exist !", name, err)
		return nil, err
	}

	// 部分更新node
	patchedNode, updateErr := c.Client.Patch(context.TODO(), name, types.StrategicMergePatchType, []byte(patchNode), metav1.PatchOptions{})
	if updateErr != nil {
		logs.Errorf("Update node %s error: %v", name, updateErr)
		return nil, fmt.Errorf("%w-%v", InternalServerError, updateErr)
	}

	//
	logs.Debugf("Patch node: %v", patchedNode)
	return patchedNode, nil

}

func (m *Manager) DeleteNode(name string, namespace string) error {
	c := m.GetNodeClient(namespace)

	// 检查node是否存在
	_, err := m.GetNode(name, namespace)
	if err != nil {
		logs.Errorf("get node %s error: %v , node not exist ", name, err)
		return err
	}

	// 存在，删除
	err = c.Client.Delete(context.TODO(), name, metav1.DeleteOptions{})
	if err != nil {
		logs.Errorf("delete node %s error: %v", name, err)
		return fmt.Errorf("%w-%v", InternalServerError, err)
	}

	logs.Debugf("Delete node: %v", name)
	return nil
}

package manager

import (
	"context"
	"hit.edu/framework/pkg/apimachinery/types"
	apis "hit.edu/framework/pkg/apis/cores"
	metav1 "hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/component-base/logs"
)

func (m *Manager) CreateScene(ss apis.SceneSpec, namespace string, uuid string) (*apis.Scene, error) {
	// 临时创建一个Scene对象
	s := apis.Scene{}

	// 构造名称
	s.Name = uuid

	// 构造Namespace
	if namespace == "" {
		s.Namespace = apis.NamespaceDefault
	} else {
		s.Namespace = namespace
	}

	s.Kind = "Scene"
	s.APIVersion = "resources/v1"

	// 构造Labels
	if ss.Desc != nil && ss.Desc.Label != nil && len(ss.Desc.Label) > 0 {
		s.Labels = ss.Desc.Label
	} else {
		s.Labels = map[string]string{}
	}

	// 复制Spec
	s.Spec = ss

	// 构造Status
	s.Status = apis.SceneStatus{}

	s.Labels["uuid"] = uuid

	c := m.GetSceneClient(s.Namespace)

	scene, err := c.Client.Create(context.TODO(), &s, metav1.CreateOptions{})
	if err != nil {
		logs.Errorf("Failed to create scene: %v", err)
		return nil, err
	}

	logs.Debugf("Created scene: %v", scene)
	return scene, nil
}

func (m *Manager) GetScene(name string, namespace string) (*apis.Scene, error) {
	c := m.GetSceneClient(namespace)
	n, err := c.Client.Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		logs.Errorf("Failed to get Scene: %v", err)
		return nil, err
	}

	//
	logs.Debugf("Get scene: %v", n)
	return n, nil
}

func (m *Manager) GetScenes(namespace string) (*apis.SceneList, error) {
	c := m.GetSceneClient(namespace)
	n, err := c.Client.List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		logs.Errorf("Failed to get scenes: %v", err)
		return nil, err
	}

	logs.Debugf("Get scenes success. ")
	return n, nil
}

func (m *Manager) UpdateScene(name string, namespace string, n *apis.Scene) (*apis.Scene, error) {
	c := m.GetSceneClient(namespace)

	// 检查scene是否存在
	_, err := m.GetScene(name, namespace)
	if err != nil {
		logs.Errorf("Get scene %s error: %v , scene not exist !", name, err)
		return nil, err
	}

	// 存在更新scene
	updatedScene, updateErr := c.Client.Update(context.TODO(), n, metav1.UpdateOptions{})
	if updateErr != nil {
		logs.Errorf("Update scene %s error: %v", name, updateErr)
		return nil, updateErr
	}

	//
	logs.Debugf("Update scene: %v", updatedScene)
	return updatedScene, nil

}

func (m *Manager) PatchScene(name string, namespace string, patchScene []byte) (*apis.Scene, error) {
	c := m.GetSceneClient(namespace)

	// 检查scene是否存在
	_, err := m.GetScene(name, namespace)
	if err != nil {
		logs.Errorf("Get scene %s error: %v , scene not exist !", name, err)
		return nil, err
	}

	// 部分更新scene
	patchedScene, updateErr := c.Client.Patch(context.TODO(), name, types.StrategicMergePatchType, []byte(patchScene), metav1.PatchOptions{})
	if updateErr != nil {
		logs.Errorf("Update scene %s error: %v", name, updateErr)
		return nil, updateErr
	}

	//
	logs.Debugf("Patch Scene: %v", patchedScene)
	return patchedScene, nil

}

func (m *Manager) DeleteScene(name string, namespace string) error {
	c := m.GetSceneClient(namespace)

	// 检查scene是否存在
	_, err := m.GetScene(name, namespace)
	if err != nil {
		logs.Errorf("get scene %s error: %v , scene not exist ", name, err)
		return err
	}

	// 存在，删除
	err = c.Client.Delete(context.TODO(), name, metav1.DeleteOptions{})
	if err != nil {
		logs.Errorf("delete scene %s error: %v", name, err)
		return err
	}

	logs.Debugf("Delete scene: %v", name)
	return nil
}

// FilterScenes 根据Label查询scenes
func (m *Manager) FilterScenes(namespace string, labelSelector string) (*apis.SceneList, error) {

	listOptions := metav1.ListOptions{
		LabelSelector: labelSelector,
	}

	c := m.GetSceneClient(namespace)
	ns, err := c.Client.List(context.TODO(), listOptions)
	if err != nil {
		return nil, err
	}
	return ns, nil
}

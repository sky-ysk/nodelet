package manager

import (
	"context"
	"fmt"
	"hit.edu/framework/pkg/apimachinery/types"
	apis "hit.edu/framework/pkg/apis/cores"
	metav1 "hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/component-base/logs"
	"time"
)

func (m *Manager) CreateData(as *apis.Data, namespace string) (*apis.Data, error) {
	a := apis.Data{}

	// TODO:data中spec提供的名称是否唯一，名称构造
	a = *as
	a.Name = as.Name

	// 构造Namespace
	if namespace == "" {
		a.Namespace = apis.NamespaceDefault
	} else {
		a.Namespace = namespace
	}

	a.Kind = "Data"
	a.APIVersion = "resources/v1"

	// 复制Spec
	a.Spec = as.Spec

	// 构造Status
	a.Status = apis.DataStatus{}

	// 记录Create时间
	a.Status.CreateAt = &apis.Time{Time: time.Now()}

	// 写入Client-Go中, 返回实际的Data
	c := m.GetDataClient(a.Namespace)

	fa, err := c.Client.Create(context.TODO(), &a, metav1.CreateOptions{})
	if err != nil {
		logs.Errorf("Failed to create data: %v", err)
		return nil, fmt.Errorf("%w-%v", InternalServerError, err)
	}

	//
	logs.Debugf("Created data: %v", fa)
	return fa, nil
}

func (m *Manager) GetData(name string, namespace string) (*apis.Data, error) {
	c := m.GetDataClient(namespace)

	a, err := c.Client.Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		logs.Errorf("Failed to get data: %v", err)
		return nil, fmt.Errorf("%w-%v", NotFound, err)
	}

	//
	logs.Debugf("Get data: %v", a)
	return a, nil
}

func (m *Manager) GetDatas(namespace string) (*apis.DataList, error) {
	c := m.GetDataClient(namespace)

	a, err := c.Client.List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		logs.Errorf("Failed to get datas: %v", err)
		return nil, fmt.Errorf("%w-%v", InternalServerError, err)
	}

	//
	logs.Debugf("Get datas success.")
	return a, nil
}

func (m *Manager) UpdateData(name string, namespace string, a *apis.Data) (*apis.Data, error) {
	c := m.GetDataClient(namespace)

	// 检查data是否存在
	_, err := m.GetData(name, namespace)
	if err != nil {
		logs.Errorf("Get data %s error: %v , data not exist !", name, err)
		return nil, err
	}

	// 存在更新data
	updatedData, updateErr := c.Client.Update(context.TODO(), a, metav1.UpdateOptions{})
	if updateErr != nil {
		logs.Errorf("Update data %s error: %v", name, updateErr)
		return nil, fmt.Errorf("%w-%v", InternalServerError, updateErr)
	}

	//
	logs.Debugf("Update data: %v", updatedData)
	return updatedData, nil

}

func (m *Manager) PatchData(name string, namespace string, patchData []byte) (*apis.Data, error) {
	c := m.GetDataClient(namespace)

	// 检查data是否存在
	_, err := m.GetData(name, namespace)
	if err != nil {
		logs.Errorf("Get data %s error: %v , data not exist !", name, err)
		return nil, err
	}

	// 部分更新data
	patchedData, err := c.Client.Patch(context.TODO(), name, types.StrategicMergePatchType, patchData, metav1.PatchOptions{})
	if err != nil {
		logs.Errorf("patch data %s error: %v", name, err)
		return nil, fmt.Errorf("%w-%v", InternalServerError, err)
	}

	//
	logs.Debugf("patched data : %v ", patchedData)
	return patchedData, nil
}

func (m *Manager) DeleteData(name string, namespace string) error {
	c := m.GetDataClient(namespace)

	// 检查data是否存在
	_, err := m.GetData(name, namespace)
	if err != nil {
		logs.Errorf("get data %s error: %v , data not exist ", name, err)
		return err
	}

	// 存在，删除data
	err = c.Client.Delete(context.TODO(), name, metav1.DeleteOptions{})
	if err != nil {
		logs.Errorf("delete data %s error: %v", name, err)
		return fmt.Errorf("%w-%v", InternalServerError, err)
	}

	//
	logs.Debugf("Delete data")
	return nil
}

func (m *Manager) DeleteDatas(namespace string, filedSelector string) error {
	c := m.GetDataClient(namespace)

	lstOpts := metav1.ListOptions{
		FieldSelector: filedSelector,
	}

	// 存在，删除data
	err := c.Client.DeleteCollection(context.TODO(), metav1.DeleteOptions{}, lstOpts)
	if err != nil {
		logs.Errorf("delete datas error: %v", err)
		return fmt.Errorf("%w-%v", InternalServerError, err)
	}

	//
	logs.Debugf("Delete datas")
	return nil
}

// FilterDevices 根据Label查询Device
func (m *Manager) FilterDatas(namespace string, labelSelector string) (*apis.DataList, error) {
	c := m.GetDataClient(namespace)

	listOptions := metav1.ListOptions{
		LabelSelector: labelSelector,
	}

	d, err := c.Client.List(context.TODO(), listOptions)
	if err != nil {
		logs.Errorf("Failed to get devices with labelselector: %s , error %v ", labelSelector, err)
		return nil, fmt.Errorf("%v-%w", InternalServerError, err)
	}

	logs.Infof("Get devices with label success.")
	return d, nil
}

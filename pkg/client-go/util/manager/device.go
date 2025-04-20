package manager

import (
	"context"
	"errors"
	"hit.edu/framework/pkg/apimachinery/types"
	apis "hit.edu/framework/pkg/apis/cores"
	metav1 "hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/component-base/logs"
)

// 创建Device
func (m *Manager) CreateDevice(d *apis.Device, namespace string) (*apis.Device, error) {
	// TODO: 写入Device的相关信息
	// 使用参数Namespace覆盖
	d.Namespace = namespace

	d.Labels = map[string]string{}

	// 根据设备能力打上Label
	for _, a := range d.Spec.Abilities {
		d.Labels[a] = a
	}
	// 根据设备的厂商打上Label
	if d.Spec.Desc != nil && d.Spec.Desc.Maker != nil {
		d.Labels["Maker"] = *d.Spec.Desc.Maker
	}
	//
	c := m.GetDeviceClient(namespace)

	fd, err := c.Client.Create(context.TODO(), d, metav1.CreateOptions{})
	if err != nil {
		return nil, errors.New("Failed to create device: " + err.Error())
	}
	return fd, nil
}

func (m *Manager) GetDevice(name string, namespace string) (*apis.Device, error) {
	c := m.GetDeviceClient(namespace)
	d, err := c.Client.Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return d, nil
}

// 查询所有Device
func (m *Manager) GetDevices(name string, namespace string) (*apis.DeviceList, error) {
	c := m.GetDeviceClient(namespace)
	d, err := c.Client.List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return d, nil
}

// 根据Label查询Device
func (m *Manager) FilterDevice(namespace string, label []string) (*apis.DeviceList, error) {
	labelSelector := ""
	for i, l := range label {
		if i == 0 {
			labelSelector = l
		} else {
			labelSelector += "," + l
		}
	}

	listOptions := metav1.ListOptions{
		LabelSelector: labelSelector,
	}

	c := m.GetDeviceClient(namespace)
	d, err := c.Client.List(context.TODO(), listOptions)
	if err != nil {
		return nil, err
	}
	return d, nil
}

func (m *Manager) UpdateDevice(namespace string, name string, a *apis.Device) (*apis.Device, error) {
	c := m.GetDeviceClient(namespace)

	// 检查device是否存在
	_, err := m.GetDevice(name, namespace)
	if err != nil {
		logs.Errorf("Get device %s error: %v , device not exist !", name, err)
		return nil, err
	}

	// 存在更新device
	updatedDevice, updateErr := c.Client.Update(context.TODO(), a, metav1.UpdateOptions{})
	if updateErr != nil {
		logs.Errorf("Update device %s error: %v", name, updateErr)
		return nil, updateErr
	}

	//
	logs.Debugf("Update device: %v", updatedDevice)
	return updatedDevice, nil

}

func (m *Manager) PatchDevice(name string, namespace string, patchDevice string) (*apis.Device, error) {
	c := m.GetDeviceClient(namespace)

	// 检查device是否存在
	_, err := m.GetDevice(name, namespace)
	if err != nil {
		logs.Errorf("Get device %s error: %v , device not exist !", name, err)
		return nil, err
	}

	// 部分更新device
	patchedDevice, err := c.Client.Patch(context.TODO(), name, types.StrategicMergePatchType, []byte(patchDevice), metav1.PatchOptions{})
	if err != nil {
		logs.Errorf("patch device %s error: %v", name, err)
		return nil, err
	}

	//
	logs.Debugf("patched device : %v ", patchedDevice)
	return patchedDevice, nil
}

func (m *Manager) DeleteDevice(name string, namespace string) error {
	c := m.GetDeviceClient(namespace)

	// 检查device是否存在
	_, err := m.GetDevice(name, namespace)
	if err != nil {
		logs.Errorf("get device %s error: %v , device not exist ", name, err)
		return err
	}

	// 存在，删除
	err = c.Client.Delete(context.TODO(), name, metav1.DeleteOptions{})
	if err != nil {
		logs.Errorf("delete device %s error: %v", name, err)
		return err
	}

	//
	logs.Debugf("Delete device %v ", err)
	return nil
}

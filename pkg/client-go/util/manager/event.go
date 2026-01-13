package manager

import (
	"context"
	"fmt"

	"hit.edu/framework/pkg/apimachinery/types"
	"hit.edu/framework/pkg/component-base/logs"

	"hit.edu/framework/pkg/apimachinery/runtime"
	apis "hit.edu/framework/pkg/apis/cores"
	metav1 "hit.edu/framework/pkg/apis/meta"
)

func (m *Manager) CreateEvent(e *apis.Event, namespace string) (*apis.Event, error) {
	c := m.GetEventClient(namespace)

	fe, err := c.Client.Create(context.TODO(), e, metav1.CreateOptions{})
	if err != nil {
		logs.Errorf("Failed to create event: %v", err)
		return nil, fmt.Errorf("%w-%v", InternalServerError, err)
	}

	logs.Debugf("Created event: %v", fe)
	return fe, nil
}

func (m *Manager) GetEvent(name string, namespace string) (*apis.EventList, error) {
	//c := m.GetEventClient(namespace)
	//
	//e, err := c.Client.Get(context.TODO(), name, metav1.GetOptions{})
	//if err != nil {
	//	logs.Errorf("Failed to get event: %v", err)
	//	return nil, fmt.Errorf("%w-%v", NotFound, err)
	//}
	//
	////
	//logs.Debugf("Get event: %v", e)
	//return e, nil

	fieldSelector := fmt.Sprintf("involvedObject.name=%s", name)

	listOptions := metav1.ListOptions{
		FieldSelector: fieldSelector,
	}

	c := m.GetEventClient(namespace)

	e, err := c.Client.List(context.TODO(), listOptions)
	if err != nil {
		logs.Errorf("Failed to get events: %v", err)
		return nil, fmt.Errorf("%w-%v", InternalServerError, err)
	}
	return e, nil
}

// GetEvents 查询一个变量所有的相关事件
func (m *Manager) GetEvents(namespace string) (*apis.EventList, error) {
	// fieldSelector := fmt.Sprintf("involvedObject.name=%s", name)

	listOptions := metav1.ListOptions{}

	c := m.GetEventClient(namespace)

	e, err := c.Client.List(context.TODO(), listOptions)
	if err != nil {
		logs.Errorf("Failed to get events: %v", err)
		return nil, fmt.Errorf("%w-%v", InternalServerError, err)
	}
	return e, nil
}

func (m *Manager) LogEvent(object runtime.Object, eventtype, reason, message string, namespace string) error {
	logs.Infof("create event")
	c := m.GetEventClient(namespace)
	c.Recoder.Event(object, eventtype, reason, message)
	return nil
}

func (m *Manager) LogEventForMigration(object runtime.Object, eventtype, reason, message string, migrationTarget string, namespace string) error {
	c := m.GetEventClient(namespace)
	c.Recoder.EventForMigration(object, eventtype, reason, message, migrationTarget)
	return nil
}

func (m *Manager) UpdateEvent(name string, namespace string, a *apis.Event) (*apis.Event, error) {
	c := m.GetEventClient(namespace)

	// 检查event是否存在
	_, err := m.GetEvent(name, namespace)
	if err != nil {
		logs.Errorf("Get event %s error: %v , event not exist !", name, err)
		return nil, err
	}

	// 存在更新event
	updatedEvent, updateErr := c.Client.Update(context.TODO(), a, metav1.UpdateOptions{})
	if updateErr != nil {
		logs.Errorf("Update event %s error: %v", name, updateErr)
		return nil, fmt.Errorf("%w-%v", InternalServerError, updateErr)
	}

	//
	logs.Debugf("Update event: %v", updatedEvent)
	return updatedEvent, nil

}

func (m *Manager) PatchEvent(name string, namespace string, patchEvent []byte) (*apis.Event, error) {
	c := m.GetEventClient(namespace)

	// 检查event是否存在
	_, err := m.GetEvent(name, namespace)
	if err != nil {
		logs.Errorf("Get event %s error: %v , event not exist !", name, err)
		return nil, err
	}

	// 部分更新event
	patchedEvent, err := c.Client.Patch(context.TODO(), name, types.StrategicMergePatchType, patchEvent, metav1.PatchOptions{})
	if err != nil {
		logs.Errorf("patch event %s error: %v", name, err)
		return nil, fmt.Errorf("%w-%v", InternalServerError, err)
	}

	//
	logs.Debugf("patched event : %v ", patchedEvent)
	return patchedEvent, nil
}

func (m *Manager) DeleteEvent(name string, namespace string) error {
	c := m.GetEventClient(namespace)

	// 检查event是否存在
	_, err := m.GetEvent(name, namespace)
	if err != nil {
		logs.Errorf("get event %s error: %v , event not exist ", name, err)
		return err
	}

	// 存在，删除event
	err = c.Client.Delete(context.TODO(), name, metav1.DeleteOptions{})
	if err != nil {
		logs.Errorf("delete event %s error: %v", name, err)
		return fmt.Errorf("%w-%v", InternalServerError, err)
	}

	//
	logs.Debugf("Delete event %v ", err)
	return nil
}

func (m *Manager) DeleteEvents(namespace string, Selector string) error {
	// "Spec.Name=demo-event"

	fieldSelector := fmt.Sprintf("involvedObject.name=%s", Selector)
	listOptions := metav1.ListOptions{
		FieldSelector: fieldSelector,
	}

	c := m.GetEventClient(namespace)

	err := c.Client.DeleteCollection(context.TODO(), metav1.DeleteOptions{}, listOptions)
	if err != nil {
		logs.Errorf("Failed to get events: %v", err)
		return fmt.Errorf("%w-%v", InternalServerError, err)
	}
	return nil
}

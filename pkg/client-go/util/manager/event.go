package manager

import (
	"context"
	"errors"
	"fmt"
	"hit.edu/framework/pkg/apimachinery/runtime"
	apis "hit.edu/framework/pkg/apis/cores"
	metav1 "hit.edu/framework/pkg/apis/meta"
)

func (m *Manager) CreateEvent(e *apis.Event, namespace string) (*apis.Event, error) {
	c := m.GetEventClient(namespace)

	fe, err := c.Client.Create(context.TODO(), e, metav1.CreateOptions{})
	if err != nil {
		return nil, errors.New("Failed to create device: " + err.Error())
	}
	return fe, nil
}

func (m *Manager) LogEvent(object runtime.Object, eventtype, reason, message string, namespace string) error {
	c := m.GetEventClient(namespace)
	c.Recoder.Event(object, eventtype, reason, message)
	return nil
}

// 查询一个变量所有的相关事件
func (m *Manager) GetEvents(name string, namespace string) (*apis.EventList, error) {
	fieldSelector := fmt.Sprintf("involvedObject.name=%s", name)

	listOptions := metav1.ListOptions{
		FieldSelector: fieldSelector,
	}

	c := m.GetEventClient(namespace)

	e, err := c.Client.List(context.TODO(), listOptions)
	if err != nil {
		return nil, err
	}
	return e, nil
}

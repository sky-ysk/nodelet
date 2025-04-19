package manager

import (
	"fmt"
	apis "hit.edu/framework/pkg/apis/cores"
	metav1 "hit.edu/framework/pkg/apis/meta"
	"testing"
)

func TestCreateDevice(t *testing.T) {
	namespace := "Guochuang"
	clientset, err := CreateClientSet()
	manager := NewManager(clientset)
	if err != nil {
		panic(err)
	}

	ip := "127.0.0.1"
	inter := "api/control/start_task"
	port := "2387"
	spec := apis.DeviceSpec{
		Name: "Robot",
		Abilities: []string{
			"Move", "Grab",
		},
	}

	status := apis.DeviceStatus{
		Abilities: map[string]apis.Ability{
			"Move": apis.Ability{
				Name: "Move.Leju.Guochuang",
				Services: map[string]apis.AbilityService{
					"Start": apis.AbilityService{
						Ip:        &ip,
						Port:      &port,
						Interface: &inter,
					},
				},
			},
			"Grab": apis.Ability{
				Name: "Grab.Leju.Guochuang",
				Services: map[string]apis.AbilityService{
					"Start": apis.AbilityService{
						Ip:        &ip,
						Port:      &port,
						Interface: &inter,
					},
				},
			},
		},
	}
	device := apis.Device{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "Robot",
			Namespace: "Guochuang",
		},
		Spec:   spec,
		Status: status,
	}

	spec.Abilities = []string{
		"Move",
	}
	device2 := apis.Device{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "Robot1",
			Namespace: "Guochuang",
		},
		Spec:   spec,
		Status: status,
	}

	d, err := manager.CreateDevice(&device, namespace)
	if err != nil {
		panic(err)
	}
	fmt.Println(d)

	d, err = manager.CreateDevice(&device2, namespace)
	if err != nil {
		panic(err)
	}
	fmt.Println(d)
}

func TestGetDevice(t *testing.T) {
	name := "Robot"
	namespace := "Guochuang"

	clientset, err := CreateClientSet()
	manager := NewManager(clientset)
	d, err := manager.GetDevice(name, namespace)
	if err != nil {
		panic(err)
	}
	fmt.Println(d)
}

func TestFilterDevice(t *testing.T) {
	clientset, err := CreateClientSet()
	manager := NewManager(clientset)

	namespace := "Guochuang"
	label := []string{
		"Move",
	}

	d, err := manager.FilterDevice(namespace, label)
	if err != nil {
		panic(err)
	}

	for _, j := range d.Items {
		fmt.Println(j)
	}

	label = []string{
		"Move", "Grab",
	}

	d, err = manager.FilterDevice(namespace, label)
	if err != nil {
		panic(err)
	}

	fmt.Println("-------------")
	for _, j := range d.Items {
		fmt.Println(j)
	}
}

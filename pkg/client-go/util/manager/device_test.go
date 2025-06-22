package manager

import (
	"fmt"
	apis "hit.edu/framework/pkg/apis/cores"
	metav1 "hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/component-base/analyzer"
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

	label := "Move==Move"

	d, err := manager.FilterDevices(namespace, label)
	if err != nil {
		panic(err)
	}

	for _, j := range d.Items {
		fmt.Println(j)
	}

	label = "Move==Move,Grab==Grab"

	d, err = manager.FilterDevices(namespace, label)
	if err != nil {
		panic(err)
	}

	fmt.Println("-------------")
	for _, j := range d.Items {
		fmt.Println(j)
	}
}

func TestGetDevices(t *testing.T) {
	clientset, err := CreateClientSet()
	manager := NewManager(clientset)

	lst, err := manager.FilterDevices("Guochuang", "")
	if err != nil {
		panic(err)
	} else {
		str, err := analyzer.SerializeToJson(lst)
		if err != nil {
			return
		}
		fmt.Println(str)
	}
}

func TestPatchDevices(t *testing.T) {
	namespace := "Guochuang"
	clientset, err := CreateClientSet()
	manager := NewManager(clientset)
	if err != nil {
		panic(err)
	}

	str1 := "device before update"

	desc := &apis.DeviceDesc{
		Docs: &str1,
	}

	ip := "127.0.0.1"
	inter := "api/control/start_task"
	port := "2387"
	spec := apis.DeviceSpec{
		Name: "testRobot3",
		Abilities: []string{
			"Move", "Grab",
		},
		Desc: desc,
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
			Name:      "testRobot3",
			Namespace: "Guochuang",
		},
		Spec:   spec,
		Status: status,
	}

	spec.Abilities = []string{
		"Move",
	}

	d, err := manager.CreateDevice(&device, namespace)
	if err != nil {
		panic(err)
	} else {
		str, err := analyzer.SerializeToJson(d)
		if err != nil {
			return
		}
		fmt.Println(str)
	}

	patchData := "{\n  \"spec\": {\n      \"desc\": {\n          \"docs\": \"node after patch\"\n      }\n  }\n\n}"
	d, err = manager.PatchDevice(d.Name, d.Namespace, []byte(patchData))
	if err != nil {
		panic(err)
	} else {
		str, err := analyzer.SerializeToJson(d)
		if err != nil {
			return
		}
		fmt.Println(str)
	}
}

func TestDeleteDevice(t *testing.T) {
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
		Name: "testRobot4",
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
			Name:      "testRobot4",
			Namespace: "Guochuang",
		},
		Spec:   spec,
		Status: status,
	}

	spec.Abilities = []string{
		"Move",
	}

	d, err := manager.CreateDevice(&device, namespace)
	if err != nil {
		panic(err)
	} else {
		str, err := analyzer.SerializeToJson(d)
		if err != nil {
			return
		}
		fmt.Println(str)
	}

	err = manager.DeleteDevice(d.Name, d.Namespace)
	if err != nil {
		panic(err)
	}

}

func TestUpdateDevice(t *testing.T) {
	namespace := "Guochuang"
	clientset, err := CreateClientSet()
	manager := NewManager(clientset)
	if err != nil {
		panic(err)
	}

	str1 := "device before update"
	str2 := "device after update"

	desc := &apis.DeviceDesc{
		Docs: &str1,
	}

	desc2 := &apis.DeviceDesc{
		Docs: &str2,
	}

	ip := "127.0.0.1"
	inter := "api/control/start_task"
	port := "2387"
	spec := apis.DeviceSpec{
		Name: "testRobot2",
		Abilities: []string{
			"Move", "Grab",
		},
		Desc: desc,
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
			Name:      "testRobot2",
			Namespace: "Guochuang",
		},
		Spec:   spec,
		Status: status,
	}

	spec.Abilities = []string{
		"Move",
	}

	d, err := manager.CreateDevice(&device, namespace)
	if err != nil {
		panic(err)
	} else {
		str, err := analyzer.SerializeToJson(d)
		if err != nil {
			return
		}
		fmt.Println(str)
	}

	d.Spec.Desc = desc2
	d, err = manager.UpdateDevice(d.Name, d.Namespace, d)
	if err != nil {
		panic(err)
	} else {
		str, err := analyzer.SerializeToJson(d)
		if err != nil {
			return
		}
		fmt.Println(str)
	}
}

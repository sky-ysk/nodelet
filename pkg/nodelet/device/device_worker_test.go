package device

import (
	apis "hit.edu/framework/pkg/apis/cores"
	metav1 "hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/client-go/util/manager"
	"hit.edu/framework/pkg/component-base/logs"
	"hit.edu/framework/pkg/scheduler/utils"
	"testing"
)

func TestDeviceWorker(t *testing.T) {
	logs.Init("test")
	dw := GetDeviceWorker()
	/* 单例测试 */
	//dw1 := GetDeviceWorker()
	//if dw1 == dw {
	//	logs.Infof("single!")
	//}

	/* demo1 */
	flag, devices := dw.CheckSatisfaction([]string{"A", "B", "E"})
	if flag {
		logs.Infof("CheckSatisfaction successfully device is %v", devices)
		//g := CreateGroup()
		//g1 := dw.LockDevices(g, devices)
		//dw.Manager.CreateGroup(g1.Spec, nil, "test", "", "")
	} else {
		logs.Infof("CheckSatisfaction unsuccessfully device is %v", devices)
	}

}

func TestDevice(t *testing.T) {
	cs, err := utils.CreateClientSetWithTimeOut(2000)
	if err != nil {
		panic(err)
	}
	m := manager.NewManager(cs)
	// 创建Device
	device1 := &apis.Device{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "device1",
			Namespace: "test",
			Labels: map[string]string{
				"environment": "dev",
			},
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Device",
			APIVersion: "resources/v1",
		},
		Spec: apis.DeviceSpec{
			Name: "device1",
			AccessMethod: &apis.AccessMethod{
				Type: apis.AccessByAbility,
				URL:  "http://192.168.8.197:8080",
			},
			Abilities: []string{
				"A", "B", "C",
			},
		},
		Status: apis.DeviceStatus{
			Abilities: map[string]apis.Ability{
				"A": {
					Name: "test",
					Services: map[string]apis.AbilityService{
						"test": apis.AbilityService{
							Ip:        new(string),
							Interface: new(string),
							Port:      new(string),
						},
					},
					Status: apis.AbilityRunning,
				},
				"B": {
					Name: "test",
					Services: map[string]apis.AbilityService{
						"test": apis.AbilityService{
							Ip:        new(string),
							Interface: new(string),
							Port:      new(string),
						},
					},
					Status: apis.AbilityRunning,
				},
				"C": {
					Name: "test",
					Services: map[string]apis.AbilityService{
						"test": apis.AbilityService{
							Ip:        new(string),
							Interface: new(string),
							Port:      new(string),
						},
					},
					Status: apis.AbilityRunning,
				},
			},
			Phase: apis.DeviceIdle,
		},
	}

	device2 := &apis.Device{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "device2",
			Namespace: "test",
			Labels: map[string]string{
				"environment": "dev",
			},
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Device",
			APIVersion: "resources/v1",
		},
		Spec: apis.DeviceSpec{
			Name: "device2",
			AccessMethod: &apis.AccessMethod{
				Type: apis.AccessByAbility,
				URL:  "http://192.168.8.197:8080",
			},
			Abilities: []string{
				"A", "D", "C",
			},
		},
		Status: apis.DeviceStatus{
			Abilities: map[string]apis.Ability{
				"A": {
					Name: "test",
					Services: map[string]apis.AbilityService{
						"test": apis.AbilityService{
							Ip:        new(string),
							Interface: new(string),
							Port:      new(string),
						},
					},
					Status: apis.AbilityRunning,
				},
				"D": {
					Name: "test",
					Services: map[string]apis.AbilityService{
						"test": apis.AbilityService{
							Ip:        new(string),
							Interface: new(string),
							Port:      new(string),
						},
					},
					Status: apis.AbilityRunning,
				},
				"C": {
					Name: "test",
					Services: map[string]apis.AbilityService{
						"test": apis.AbilityService{
							Ip:        new(string),
							Interface: new(string),
							Port:      new(string),
						},
					},
					Status: apis.AbilityRunning,
				},
			},
			Phase: apis.DeviceIdle,
		},
	}
	_, err = m.CreateDevice(device1, "test")
	if err != nil {
		logs.Errorf("[TEST] Create Device[%s] err:%s", device1.Name, err.Error())
	}
	_, err = m.CreateDevice(device2, "test")
	if err != nil {
		logs.Errorf("[TEST] Create Device[%s] err:%s", device2.Name, err.Error())
	}
}

func CreateGroup() *apis.Group {

	runtime1 := &apis.Runtime{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "R1",
			Namespace: "test",
			Labels: map[string]string{
				"environment": "dev",
			},
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Runtime",
			APIVersion: "resources/v1",
		},
		Spec: apis.RuntimeSpec{
			Name: "R1",
			Type: apis.ByDevice,
			Devices: []apis.DeviceSpec{
				apis.DeviceSpec{
					Name:               "deviceLock",
					ExpectedProperties: map[string]apis.Property{},
					Abilities: []string{
						"A",
					},
				},
			},
			Image: "",
		},
	}

	// 乐聚检测
	runtime2 := &apis.Runtime{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "R2",
			Namespace: "test",
			Labels: map[string]string{
				"environment": "dev",
			},
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Runtime",
			APIVersion: "resources/v1",
		},
		Spec: apis.RuntimeSpec{
			Name: "R2",
			Type: apis.ByDevice,

			Devices: []apis.DeviceSpec{
				apis.DeviceSpec{
					Name:               "deviceLock",
					ExpectedProperties: map[string]apis.Property{},
					Abilities: []string{
						"D",
					},
				},
			},
			Image: "",
		},
	}

	// 星海图抓取
	runtime3 := &apis.Runtime{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "R3",
			Namespace: "test",
			Labels: map[string]string{
				"environment": "dev",
			},
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Runtime",
			APIVersion: "resources/v1",
		},
		Spec: apis.RuntimeSpec{
			Name: "R3",
			Type: apis.ByDevice,
			Devices: []apis.DeviceSpec{
				apis.DeviceSpec{
					Name:               "deviceLock",
					ExpectedProperties: map[string]apis.Property{},
					Abilities: []string{
						"C",
					},
				},
			},
			Image: "",
		},
	}

	// 星海图检测
	action1 := &apis.Action{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "A1",
			Namespace: "test",
			Labels: map[string]string{
				"environment": "dev",
			},
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Action",
			APIVersion: "resources/v1",
		},
		Spec: apis.ActionSpec{
			Desc: &apis.Description{
				Docs: "启始任务",
			},
			Name: "A1",
			Runtimes: []apis.RuntimeSpec{
				runtime1.Spec,
			},
		},
	}

	// 乐聚检测
	action2 := &apis.Action{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "A2",
			Namespace: "test",
			Labels: map[string]string{
				"environment": "dev",
			},
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Action",
			APIVersion: "resources/v1",
		},
		Spec: apis.ActionSpec{
			Desc: &apis.Description{
				Docs: "并行任务1",
			},
			Name: "A2",
			Runtimes: []apis.RuntimeSpec{
				runtime2.Spec,
			},
		},
	}

	// 星海图抓取
	action3 := &apis.Action{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "A3",
			Namespace: "test",
			Labels: map[string]string{
				"environment": "dev",
			},
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Action",
			APIVersion: "resources/v1",
		},
		Spec: apis.ActionSpec{
			Desc: &apis.Description{
				Docs: "并行任务2",
			},
			Name: "A3",
			Runtimes: []apis.RuntimeSpec{
				runtime3.Spec,
			},
		},
	}

	group1 := &apis.Group{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "G1",
			Namespace: "test",
			Labels: map[string]string{
				"environment": "dev",
			},
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Group",
			APIVersion: "resources/v1",
		},
		Spec: apis.GroupSpec{
			Desc: &apis.Description{
				Docs: "测试Group",
			},
			Name: "G1",
			Actions: []apis.ActionSpec{
				action1.Spec, action2.Spec, action3.Spec,
			},
			Replicas: []int32{0, 0},
		},
	}

	return group1
}

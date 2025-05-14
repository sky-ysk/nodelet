package device

import (
	"fmt"
	json "github.com/json-iterator/go"
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
	g := CreateGroup()

	flag, deviceTable := dw.ChooseDevices(&g.Spec)
	go func() {
		if flag {
			logs.Infof("ChooseDevices successfully")
			logs.Infof("deviceTable is %v", deviceTable)
			flag, g = dw.LockDevices(g, deviceTable)
			if flag {
				logs.Infof("lock device successfully")
				task1 := &apis.Task{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "T1",
						Namespace: "test",
						Labels: map[string]string{
							"environment": "dev",
						},
					},
					TypeMeta: metav1.TypeMeta{
						Kind:       "Task",
						APIVersion: "resources/v1",
					},
					Spec: apis.TaskSpec{
						Desc: &apis.Description{
							Docs: "星海图&乐聚的检测和抓取",
						},
						Name: "T1",
						Groups: []apis.GroupSpec{
							g.Spec,
						},
					},
				}
				_, err := dw.Manager.CreateTask(task1.Spec, nil, task1.Namespace, "", "")
				if err != nil {
					logs.Errorf("create task fail")
				}
			} else {
				logs.Errorf("lock device unsuccessfully")
			}
		} else {
			logs.Errorf("ChooseDevices unsuccessfully")
		}
	}()

	go func() {
		if flag {
			logs.Infof("ChooseDevices successfully")
			logs.Infof("deviceTable is %v", deviceTable)
			flag, g = dw.LockDevices(g, deviceTable)
			if flag {
				logs.Infof("lock device successfully")
				task1 := &apis.Task{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "T1",
						Namespace: "test",
						Labels: map[string]string{
							"environment": "dev",
						},
					},
					TypeMeta: metav1.TypeMeta{
						Kind:       "Task",
						APIVersion: "resources/v1",
					},
					Spec: apis.TaskSpec{
						Desc: &apis.Description{
							Docs: "星海图&乐聚的检测和抓取",
						},
						Name: "T1",
						Groups: []apis.GroupSpec{
							g.Spec,
						},
					},
				}
				_, err := dw.Manager.CreateTask(task1.Spec, nil, task1.Namespace, "", "")
				if err != nil {
					logs.Errorf("create task fail")
				}
			} else {
				logs.Errorf("lock device unsuccessfully")
			}
		} else {
			logs.Errorf("ChooseDevices unsuccessfully")
		}
	}()

	select {}
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
				"Test",
			},
		},
		Status: apis.DeviceStatus{
			Lock: apis.Lock{
				IsLocked: false,
			},
			Abilities: map[string]apis.Ability{
				"Test": {
					Name: "TEST",
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
				"Test",
			},
		},
		Status: apis.DeviceStatus{
			Lock: apis.Lock{
				IsLocked: false,
			},
			Abilities: map[string]apis.Ability{
				"Test": {
					Name: "TEST",
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
					Name: "DEVICE_A",
					ExpectedProperties: map[string]apis.Property{
						"name": apis.Property{},
					},
					Abilities: []string{
						"Test",
					},
				},
			},
			Image: "Device{DEVICE_A}.Ability{Test}.Skill{test}",
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
					Name: "DEVICE_B",
					ExpectedProperties: map[string]apis.Property{
						"name": apis.Property{},
					},
					Abilities: []string{
						"Test",
					},
				},
			},
			Image: "Device{DEVICE_B}.Ability{Test}.Skill{test}",
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
				Docs: "",
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
				Docs: "",
			},
			Name: "A2",
			Runtimes: []apis.RuntimeSpec{
				runtime2.Spec,
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
			Devices: []apis.DeviceSpec{
				{
					Name: "DEVICE_A",
					Abilities: []string{
						"Test",
					},
					ExpectedProperties: map[string]apis.Property{
						"name": apis.Property{},
					},
				},
				{
					Name: "DEVICE_B",
					Abilities: []string{
						"Test",
					},
					ExpectedProperties: map[string]apis.Property{
						"name": apis.Property{},
					},
				},
			},
			Desc: &apis.Description{
				Docs: "测试Group",
			},
			Name: "G1",
			Actions: []apis.ActionSpec{
				action1.Spec, action2.Spec,
			},
			Replicas: []int32{0, 0},
		},
	}

	return group1
}

func Test1(t *testing.T) {
	group := &apis.GroupSpec{
		Devices: []apis.DeviceSpec{
			{
				Name: "DEVICE_A",
				Abilities: []string{
					"A",
				},
			},
			{
				Name: "DEVICE_B",
				Abilities: []string{
					"B",
				},
			},
		},
	}

	//runtime1 := &apis.Runtime{
	//	Spec: apis.RuntimeSpec{
	//		Name: "runtime1",
	//		Type: apis.ByDevice,
	//		Devices: []apis.DeviceSpec{
	//			apis.DeviceSpec{
	//				Name:               "DEVICE_A",
	//				ExpectedProperties: map[string]apis.Property{},
	//				Abilities: []string{
	//					"A",
	//				},
	//			},
	//		},
	//		Image: "Device{DEVICE_A}.Ability{A}.Service{TEST}",
	//	},
	//}
	//
	//runtime2 := &apis.Runtime{
	//	Spec: apis.RuntimeSpec{
	//		Name: "runtime2",
	//		Type: apis.ByDevice,
	//
	//		Devices: []apis.DeviceSpec{
	//			apis.DeviceSpec{
	//				Name:               "DEVICE_B",
	//				ExpectedProperties: map[string]apis.Property{},
	//				Abilities: []string{
	//					"B",
	//				},
	//			},
	//		},
	//		Image: "Device{DEVICE_B}.Ability{B}.Service{TEST}",
	//	},
	//}
	//runtimeList := []*apis.Runtime{runtime1, runtime2}
	b, _ := json.Marshal(group)
	fmt.Printf("%s", string(b))
}

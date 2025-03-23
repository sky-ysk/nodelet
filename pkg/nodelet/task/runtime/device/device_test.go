package device

import (
	"context"
	apis "hit.edu/framework/pkg/apis/cores"
	metav1 "hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/component-base/logs"
	"hit.edu/framework/pkg/nodelet/events/eventbus"
	"strconv"
	"testing"
)

func NewActionAndRuntimeAbility() (*apis.Action, *apis.Runtime) {
	var url string = "http://127.0.0.1:8123"
	var abilityName string = "manage_predict"
	action := apis.Action{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "actionTest",
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
			Name: "ActionTest",
		},
		Status: apis.ActionStatus{
			ActionID: "Action1",

			Devices: []apis.DeviceStatus{
				apis.DeviceStatus{
					Lock: apis.Lock{
						IsLocked: true,
					},
					Status:     "idle",
					Phase:      apis.DeviceIdle,
					InstanceID: "",
					ActionID:   "",
					DeviceID:   "ability framework test",
				},
			},
		},
	}

	runtime := apis.Runtime{
		Image: abilityName,
		Name:  "RuntimeTest",
		Devices: []apis.DeviceSpec{
			apis.DeviceSpec{
				Name:               "ability framework test",
				ExpectedProperties: map[string]apis.Property{},
				AccessMethod: apis.AccessMethod{
					Type:  apis.AccessByAbility,
					URL:   url,
					Group: "tinyRobot",
					Alias: "transferRobot",
				},
				Desc: apis.DeviceDesc{
					Label: []string{"Move"},
				},
			},
		},
		Outputs: apis.Output{},
	}
	action.Spec.Runtimes = []apis.Runtime{runtime}
	return &action, &runtime
}

func NewInstancePredict(url string, abilityName string, path string) (*apis.Action, *apis.Runtime) {
	action := apis.Action{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "actionTest",
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
			Name: "ActionTest",
		},
		Status: apis.ActionStatus{
			ActionID: "Action1",

			Devices: []apis.DeviceStatus{
				apis.DeviceStatus{
					Lock: apis.Lock{
						IsLocked: true,
					},
					Status:     "idle",
					Phase:      apis.DeviceIdle,
					InstanceID: "",
					ActionID:   "",
					DeviceID:   "ability framework test",
				},
			},
		},
	}

	runtime := apis.Runtime{
		Image: abilityName,
		Name:  "RuntimeTest",
		Devices: []apis.DeviceSpec{
			apis.DeviceSpec{
				Name:               "ability framework test",
				ExpectedProperties: map[string]apis.Property{},
				AccessMethod: apis.AccessMethod{
					Type:  apis.AccessByAbility,
					URL:   url,
					Group: "tinyRobot",
					Alias: "transferRobot",
				},
				Desc: apis.DeviceDesc{
					Label: []string{"Move"},
				},
			},
		},
		Outputs: apis.Output{},
		Inputs: []apis.Input{
			apis.Input{
				Type:      apis.LocalData,
				Name:      "path",
				Value:     path,
				ValueType: "string",
			},
		},
	}
	action.Spec.Runtimes = []apis.Runtime{runtime}
	return &action, &runtime
}

func NewInstanceArmAngle(url string, abilityName string) (*apis.Action, *apis.Runtime) {
	action := apis.Action{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "actionTest",
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
			Name: "ActionTest",
		},
		Status: apis.ActionStatus{
			ActionID: "Action1",

			Devices: []apis.DeviceStatus{
				apis.DeviceStatus{
					Lock: apis.Lock{
						IsLocked: true,
					},
					Status:     "idle",
					Phase:      apis.DeviceIdle,
					InstanceID: "",
					ActionID:   "",
					DeviceID:   "ability framework test",
				},
			},
		},
	}

	runtime := apis.Runtime{
		Image: abilityName,
		Name:  "RuntimeTest",
		Devices: []apis.DeviceSpec{
			apis.DeviceSpec{
				Name:               "ability framework test",
				ExpectedProperties: map[string]apis.Property{},
				AccessMethod: apis.AccessMethod{
					Type:  apis.AccessByAbility,
					URL:   url,
					Group: "tinyRobot",
					Alias: "transferRobot",
				},
				Desc: apis.DeviceDesc{
					Label: []string{"Move"},
				},
			},
		},
		Outputs: apis.Output{},
		Inputs: []apis.Input{
			apis.Input{
				Type:      apis.LocalData,
				Name:      "left",
				Value:     "-1.221,0.0872,0,0,0,0,0",
				ValueType: "string",
			},
			apis.Input{
				Type:      apis.LocalData,
				Name:      "right",
				Value:     "0,0,0,0,0,0,0",
				ValueType: "string",
			},
		},
	}
	action.Spec.Runtimes = []apis.Runtime{runtime}
	return &action, &runtime
}

func NewActionAndRuntimeRMF() (*apis.Action, *apis.Runtime) {
	action := apis.Action{
		Spec: apis.ActionSpec{
			Name: "ActionTest",
		},
		Status: apis.ActionStatus{
			ActionID: "Action1",
			Resources: []apis.ResourceStatus{
				apis.ResourceStatus{
					Name:         "cpu",
					Reserved:     10,
					ReservedUnit: apis.ComputeCPU,
				},
				apis.ResourceStatus{
					Name:         "memory",
					Reserved:     4096,
					ReservedUnit: apis.StorageMB,
				},
				apis.ResourceStatus{
					Name:         "disk",
					Reserved:     200,
					ReservedUnit: apis.StorageGB,
				},
			},

			Devices: []apis.DeviceStatus{
				apis.DeviceStatus{
					Lock: apis.Lock{
						IsLocked: true,
					},
					Status:     "idle",
					Phase:      apis.DeviceIdle,
					InstanceID: "",
					ActionID:   "",
					DeviceID:   "transferRobot",
				},
			},

			//Scenes: map[string]apis.SceneStatus{
			//	"o1": apis.SceneStatus{
			//		AttachedTask:   "",
			//		AttachedDevice: "",
			//		AttachedScene:  "",
			//		Lock: apis.Lock{
			//			IsLocked: true,
			//		},
			//	},
			//	"p1": apis.SceneStatus{
			//		AttachedTask:   "",
			//		AttachedDevice: "",
			//		AttachedScene:  "",
			//		Lock: apis.Lock{
			//			IsLocked: true,
			//		},
			//	},
			//},
		},
	}

	runtime := apis.Runtime{
		Image: "Move",
		Name:  "RuntimeTest",
		Devices: []apis.DeviceSpec{
			apis.DeviceSpec{
				Name:               "transferRobot",
				ExpectedProperties: map[string]apis.Property{},
				AccessMethod: apis.AccessMethod{
					Type:  apis.AccessByRmf,
					URL:   "http://192.168.1.225:8000",
					Group: "tinyRobot",
					Alias: "transferRobot",
				},
				Desc: apis.DeviceDesc{
					Label: []string{"Move"},
				},
			},
		},
		Outputs: apis.Output{},
		Inputs: []apis.Input{
			apis.Input{
				Type:      apis.LocalData,
				Name:      "dest",
				Value:     "R201",
				ValueType: "string",
			},
			apis.Input{
				Type:      apis.LocalData,
				Name:      "orientation",
				Value:     "-3.12",
				ValueType: "double",
			},
			apis.Input{
				Type:      apis.LocalData,
				Name:      "dock",
				Value:     "true",
				ValueType: "bool",
			},
		},
		//Resources: []apis.ResourceSpec{
		//	apis.ResourceSpec{
		//		Name:              "cpu",
		//		ExpectedValue:     2,
		//		ExpectedValueUnit: apis.ComputeCPU,
		//		Type:              apis.Compute,
		//	},
		//	apis.ResourceSpec{
		//		Name:              "memory",
		//		ExpectedValue:     1024,
		//		ExpectedValueUnit: apis.StorageMB,
		//		Type:              apis.Storage,
		//	},
		//	apis.ResourceSpec{
		//		Name:              "disk",
		//		ExpectedValue:     10,
		//		ExpectedValueUnit: apis.StorageGB,
		//		Type:              apis.Storage,
		//	},
		//},
		//Scenes: []apis.SceneSpec{
		//	apis.SceneSpec{
		//		SceneID:          "o1",
		//		Type:             apis.ObjectType,
		//		ExpectedProperty: map[string]apis.Property{"position": apis.Property{Name: "position", Type: apis.StringType, Value: "R201"}},
		//	},
		//	apis.SceneSpec{
		//		SceneID:          "p1",
		//		Type:             apis.PositionType,
		//		ExpectedProperty: map[string]apis.Property{"isOccupied": apis.Property{Name: "isOccupied", Type: apis.BoolType, Value: "true"}},
		//	},
		//},
	}
	action.Spec.Runtimes = []apis.Runtime{runtime}
	return &action, &runtime
}

func createDemoDevice(url string) *apis.Device {
	deviceTest := &apis.Device{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "deviceTest",
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
			Name:               "mock test",
			ExpectedProperties: map[string]apis.Property{},
			AccessMethod: apis.AccessMethod{
				Type:  apis.AccessByAbility,
				URL:   url,
				Group: "tinyRobot",
				Alias: "patrolRobot",
			},
			Desc: apis.DeviceDesc{
				Label: []string{"Move"},
			},
		},
	}
	return deviceTest
}

func TestRun(t *testing.T) {
	moduleName := "testModule"
	logs.Init(moduleName)
	logs.Infof("[test] testing run.....\n")
	clientSet, err := InitClient()
	if err != nil {
		logs.Errorf("[test] init clientSet failed: %v", err)
	}
	deviceClient := clientSet.Core().Devices("test")
	actionClient := clientSet.Core().Actions("test")
	url := ""
	ability := ""
	path := ""
	action, runtimeForTest := NewInstancePredict(url, ability, path)
	//action, runtimeForTest := NewInstanceArmAngle(url, ability)
	deviceDemo := createDemoDevice(url)
	_, err = actionClient.Create(context.TODO(), action, metav1.CreateOptions{})
	_, err = deviceClient.Create(context.TODO(), deviceDemo, metav1.CreateOptions{})
	if err != nil {
		logs.Errorf("%v", err)
		logs.Errorf("create device demo fail..")
	}
	eb := eventbus.NewEventBus()
	dr := NewDeviceRuntime(deviceClient, actionClient, eb)
	//action, runtimeForTest := NewActionAndRuntimeRMF()

	err = dr.Run(&apis.Group{}, action, runtimeForTest, 0, 0)
	if err != nil {
		logs.Errorf("run fail %v", err)
	}
}

//func NewActionAndRuntimeForKill() (*apis.Action, *apis.Runtime) {
//	taskId := ""
//	action := apis.Action{
//		Spec: apis.ActionSpec{
//			Name: "ActionTest",
//		},
//		Status: apis.ActionStatus{
//			Phase:    apis.Running,
//			ActionID: "Action1",
//			Resources: []apis.ResourceStatus{
//				apis.ResourceStatus{
//					Name:         "cpu",
//					Reserved:     10,
//					ReservedUnit: apis.ComputeCPU,
//				},
//				apis.ResourceStatus{
//					Name:         "memory",
//					Reserved:     4096,
//					ReservedUnit: apis.StorageMB,
//				},
//				apis.ResourceStatus{
//					Name:         "disk",
//					Reserved:     200,
//					ReservedUnit: apis.StorageGB,
//				},
//			},
//
//			Devices: []apis.DeviceStatus{
//				apis.DeviceStatus{
//					Lock: apis.Lock{
//						IsLocked: true,
//					},
//					Status:     "idle",
//					Phase:      apis.DeviceRunning,
//					InstanceID: "",
//					ActionID:   "",
//					DeviceID:   "transferRobot",
//				},
//			},
//		},
//	}
//
//	runtime := apis.Runtime{
//		Image: "Move",
//		Name:  "RuntimeTest",
//
//		Devices: []apis.DeviceSpec{
//			apis.DeviceSpec{
//				Name:               "transferRobot",
//				ExpectedProperties: map[string]apis.Property{},
//				AccessMethod: apis.AccessMethod{
//					Type:  apis.AccessByRmf,
//					URL:   "http://192.168.1.225:8000",
//					Group: "tinyRobot",
//					Alias: "transferRobot",
//				},
//				Desc: apis.DeviceDesc{
//					Label: []string{"Move"},
//				},
//			},
//		},
//		Outputs: apis.Output{
//			Type:  apis.LocalData,
//			Name:  "taskId",
//			Value: taskId,
//		},
//
//		Inputs: []apis.Input{
//			apis.Input{
//				Type:      apis.LocalData,
//				Name:      "dest",
//				Value:     "R201",
//				ValueType: "string",
//			},
//			apis.Input{
//				Type:      apis.LocalData,
//				Name:      "orientation",
//				Value:     "-3.12",
//				ValueType: "double",
//			},
//			apis.Input{
//				Type:      apis.LocalData,
//				Name:      "dock",
//				Value:     "true",
//				ValueType: "bool",
//			},
//		},
//	}
//	action.Spec.Runtimes = []apis.Runtime{runtime}
//	return &action, &runtime
//}
//
//func TestKill(t *testing.T) {
//	moduleName := "testModule"
//	logs.Init(moduleName)
//	logs.Infof("[test] testing kill.....\n")
//	clientSet, err := InitClient()
//	if err != nil {
//		logs.Errorf("[test] init clientSet failed: %v", err)
//	}
//	deviceClient := clientSet.Core().Devices("test")
//	actionClient := clientSet.Core().Actions("test")
//	eb := eventbus.NewEventBus()
//	dr := NewDeviceRuntime(deviceClient, actionClient, eb)
//	action, runtimeForTest := NewActionAndRuntimeForKill()
//	err = dr.Kill(&apis.Group{}, action, runtimeForTest)
//	if err != nil {
//		fmt.Println(err)
//	}
//}

// predict 的测试
func CreateGroupActionRuntimePredict() (*apis.Group, *apis.Action, *apis.Runtime, *apis.Device) {

	// 能力框架的url
	manageUrl := ""
	imageType := ""
	cameraUrl := ""
	position := ""
	compressed := false
	path := ""
	// 创建device
	device := &apis.Device{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "devicePredict",
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
			Name: "devicePredict",
			AccessMethod: apis.AccessMethod{
				Type: apis.AccessByAbility,
				URL:  manageUrl,
			},
			ExpectedProperties: map[string]apis.Property{},
			Abilities:          make([]apis.AbilitySpec, 0),
		},
		Status: apis.DeviceStatus{
			DeviceID: "devicePredict",
			Phase:    apis.DeviceIdle,
			Status:   "idle",
			ActionID: "",
			Lock: apis.Lock{
				IsLocked: true,
			},
			Abilities: make([]apis.AbilityStatus, 0),
		},
	}

	device.Status.Abilities = append(device.Status.Abilities, apis.AbilityStatus{
		Name: "Predict",
		Services: []apis.AbilityServiceStatus{
			{
				Name:      "Predict",
				Ip:        "192.168.8.165",
				Interface: "/predict",
			},
			{
				Name:      "PredictByUrl",
				Ip:        "192.168.8.165",
				Interface: "/predict_by_url",
			},
		},
	})

	runtime := &apis.Runtime{
		Image: "manage_Detect",
		Name:  "RuntimeTest",
		Devices: []apis.DeviceSpec{
			device.Spec,
		},
		Outputs: apis.Output{},
		Inputs: []apis.Input{
			{
				Name:  "imageType",
				Type:  "string",
				Value: imageType,
			},
			{
				Name:  "position",
				Type:  "string",
				Value: position,
			},
			{
				Name:  "cameraUrl",
				Type:  "string",
				Value: cameraUrl,
			},
			{
				Name:  "compressed",
				Type:  "bool",
				Value: strconv.FormatBool(compressed),
			},
			{
				Name:  "path",
				Type:  "string",
				Value: path,
			},
		},
	}

	action := &apis.Action{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "action",
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
			Name: "action",
			Runtimes: []apis.Runtime{
				*runtime,
			},
		},
		Status: apis.ActionStatus{
			ActionID: "Action1",
			Devices:  []apis.DeviceStatus{},
		},
	}

	group := &apis.Group{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "group",
			Namespace: "test",
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Group",
			APIVersion: "resources/v1",
		},
		Spec: apis.GroupSpec{
			Name: "group",
			Actions: []apis.Action{
				*action,
			},
		},
		Status: apis.GroupStatus{
			GroupID: "group",
			ActionStatus: []apis.ActionStatus{
				action.Status,
			},
		},
	}

	return group, action, runtime, device
}

func CreateGroupActionRuntimeArm() (*apis.Group, *apis.Action, *apis.Runtime, *apis.Device) {

	// 能力框架的url
	manageUrl := ""
	left := "-1.221,0.0872,0,0,0,0,0"
	right := "-0,0,0,0,0,0,0"
	// 创建device
	device := &apis.Device{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "deviceArm",
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
			Name: "deviceArm",
			AccessMethod: apis.AccessMethod{
				Type: apis.AccessByAbility,
				URL:  manageUrl,
			},
			ExpectedProperties: map[string]apis.Property{},
			Abilities:          make([]apis.AbilitySpec, 0),
		},
		Status: apis.DeviceStatus{
			DeviceID: "deviceArm",
			Phase:    apis.DeviceIdle,
			Status:   "idle",
			ActionID: "",
			Lock: apis.Lock{
				IsLocked: true,
			},
			Abilities: make([]apis.AbilityStatus, 0),
		},
	}

	device.Status.Abilities = append(device.Status.Abilities, apis.AbilityStatus{
		Name: "Arm",
		Services: []apis.AbilityServiceStatus{
			{
				Name:      "ArmAngle",
				Ip:        "192.168.8.165",
				Interface: "/api/control/arm_angle",
			},
			{
				Name:      "LeftArmUp",
				Ip:        "192.168.8.165",
				Interface: "/api/control/left_arm_up",
			},
			{
				Name:      "LeftArmDown",
				Ip:        "192.168.8.165",
				Interface: " /api/control/left_arm_down",
			},
		},
	})

	runtime := &apis.Runtime{
		Image: "manage_ArmControl.Leju.Guochuang",
		Name:  "RuntimeTest",
		Devices: []apis.DeviceSpec{
			device.Spec,
		},
		Outputs: apis.Output{},
		Inputs: []apis.Input{
			{
				Name:  "left",
				Value: left,
			},
			{
				Name:  "right",
				Value: right,
			},
		},
	}

	action := &apis.Action{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "action",
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
			Name: "action",
			Runtimes: []apis.Runtime{
				*runtime,
			},
		},
		Status: apis.ActionStatus{
			ActionID: "Action1",
			Devices:  []apis.DeviceStatus{},
		},
	}

	group := &apis.Group{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "group",
			Namespace: "test",
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Group",
			APIVersion: "resources/v1",
		},
		Spec: apis.GroupSpec{
			Name: "group",
			Actions: []apis.Action{
				*action,
			},
		},
		Status: apis.GroupStatus{
			GroupID: "group",
			ActionStatus: []apis.ActionStatus{
				action.Status,
			},
		},
	}

	return group, action, runtime, device
}

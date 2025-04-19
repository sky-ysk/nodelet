package device

import (
	"context"
	apis "hit.edu/framework/pkg/apis/cores"
	metav1 "hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/component-base/logs"
	"hit.edu/framework/pkg/nodelet/events/eventbus"
	"hit.edu/framework/pkg/nodelet/task/runtime/device/ability/manager"
	"strconv"
	"testing"
)

//func NewActionAndRuntimeAbility() (*apis.Action, *apis.Runtime) {
//	var url string = "http://127.0.0.1:8123"
//	var abilityName string = "manage_predict"
//	action := apis.Action{
//		ObjectMeta: metav1.ObjectMeta{
//			Name:      "actionTest",
//			Namespace: "test",
//			Labels: map[string]string{
//				"environment": "dev",
//			},
//		},
//		TypeMeta: metav1.TypeMeta{
//			Kind:       "Action",
//			APIVersion: "resources/v1",
//		},
//		Spec: apis.ActionSpec{
//			Name: "ActionTest",
//		},
//		Status: apis.ActionStatus{
//			ActionID: "Action1",
//
//			Devices: []apis.DeviceStatus{
//				apis.DeviceStatus{
//					Lock: apis.Lock{
//						IsLocked: true,
//					},
//					Status:     "idle",
//					Phase:      apis.DeviceIdle,
//					InstanceID: "",
//					ActionID:   "",
//					DeviceID:   "ability framework test",
//				},
//			},
//		},
//	}
//
//	runtime := apis.Runtime{
//		Image: abilityName,
//		Name:  "RuntimeTest",
//		Devices: []apis.DeviceSpec{
//			apis.DeviceSpec{
//				Name:               "ability framework test",
//				ExpectedProperties: map[string]apis.Property{},
//				AccessMethod: apis.AccessMethod{
//					Type:  apis.AccessByAbility,
//					URL:   url,
//					Group: "tinyRobot",
//					Alias: "transferRobot",
//				},
//				Desc: apis.DeviceDesc{
//					Label: []string{"Move"},
//				},
//			},
//		},
//		Outputs: apis.Output{},
//	}
//	action.Spec.Runtimes = []apis.Runtime{runtime}
//	return &action, &runtime
//}

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
	groupClient := clientSet.Core().Groups("test")

	//g, a, r, d := CreateGroupActionRuntimePredict()
	g, a, r, d := CreateGroupActionRuntimeArm()
	groupClient.Delete(context.TODO(), g.Name, metav1.DeleteOptions{})
	actionClient.Delete(context.TODO(), a.Name, metav1.DeleteOptions{})
	deviceClient.Delete(context.TODO(), d.Name, metav1.DeleteOptions{})
	_, err = actionClient.Create(context.TODO(), a, metav1.CreateOptions{})
	if err != nil {
		logs.Errorf("[test] CreateAction failed: %v", err)
	}
	_, err = deviceClient.Create(context.TODO(), d, metav1.CreateOptions{})
	if err != nil {
		logs.Errorf("[test] CreateDevice failed: %v", err)
	}
	_, err = groupClient.Create(context.TODO(), g, metav1.CreateOptions{})
	if err != nil {
		logs.Errorf("[test] CreateGroup failed: %v", err)
	}
	if err != nil {
		logs.Errorf("%v", err)
		logs.Errorf("create device demo fail..")
	}
	eb := eventbus.NewEventBus()
	dr := NewDeviceRuntime(deviceClient, actionClient, eb)
	//action, runtimeForTest := NewActionAndRuntimeRMF()

	err = dr.Run(g, a, r, 0, 0)
	if err != nil {
		logs.Errorf("run fail %v", err)
	}
	r.Image = "service_LeftArmDown"
	err = dr.Run(g, a, r, 0, 0)
}

func TestTerminate(t *testing.T) {
	moduleName := "testModule"
	logs.Init(moduleName)
	url := "http://192.168.8.165:8080"
	ability := "ArmControl.Leju.Guochuang"
	am := manager.NewAbilityManager(url, ability)
	err := am.TerminateAbility()
	if err != nil {
		logs.Errorf("[test] terminate failed: %v", err)
	}
}

// predict 的测试
func CreateGroupActionRuntimePredict() (*apis.Group, *apis.Action, *apis.Runtime, *apis.Device) {

	// 能力框架的url
	manageUrl := "http://192.168.8.165:8080"
	imageType := "rgb"
	cameraUrl := "http://127.0.0.1:51169/api/status/camera"
	position := "head"
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
				Lock: true,
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
		Outputs: make([]apis.Output, 1),
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
			ActionID:      "Action1",
			Devices:       make(map[string]apis.DeviceStatus),
			RuntimeStatus: make([]apis.RuntimeStatus, 1),
		},
	}
	action.Status.Devices["devicePredict"] = device.Status

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
			Name:    "group",
			Actions: make([]apis.Action, 1),
		},
		Status: apis.GroupStatus{
			GroupID: "group",
			ActionStatus: []apis.ActionStatus{
				action.Status,
			},
		},
	}
	group.Spec.Actions[0] = *action
	return group, action, runtime, device
}

func CreateGroupActionRuntimeArm() (*apis.Group, *apis.Action, *apis.Runtime, *apis.Device) {

	// 能力框架的url
	manageUrl := "http://192.168.8.165:8080"
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
				Lock: true,
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
				Interface: "/api/control/left_arm_down",
			},
		},
	})

	runtime := &apis.Runtime{
		Image: "manage_ArmControl.Leju.Guochuang",
		Name:  "RuntimeTest",
		Devices: []apis.DeviceSpec{
			device.Spec,
		},
		Outputs: make([]apis.Output, 1),
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
			ActionID:      "Action1",
			Devices:       make(map[string]apis.DeviceStatus),
			RuntimeStatus: make([]apis.RuntimeStatus, 1),
		},
	}
	action.Status.Devices["deviceArm"] = device.Status
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
			Name:    "group",
			Actions: make([]apis.Action, 1),
		},
		Status: apis.GroupStatus{
			GroupID: "group",
			ActionStatus: []apis.ActionStatus{
				action.Status,
			},
		},
	}
	group.Spec.Actions[0] = *action
	return group, action, runtime, device
}

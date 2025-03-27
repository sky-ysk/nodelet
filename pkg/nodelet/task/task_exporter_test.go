package task

import (
	"context"
	"fmt"
	"hit.edu/framework/pkg/apimachinery/runtime"
	"hit.edu/framework/pkg/apimachinery/runtime/schema"
	"hit.edu/framework/pkg/apimachinery/runtime/serializer"
	metav1 "hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/client-go/clients"
	"hit.edu/framework/pkg/client-go/rest"
	"hit.edu/framework/pkg/nodelet/task/runtime/device/ability/manager"
	"net/http"
	"strconv"
	"testing"
	"time"

	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/component-base/logs"
)

//test 111

func yoloTrainTaskGroup() []*apis.Group {
	newGroup := apis.Group{
		ObjectMeta: meta.ObjectMeta{Name: "cmd_yolo_train"},
		Spec: apis.GroupSpec{
			Name:    "TestGroup",
			Parents: make([]string, 0),
			Actions: []apis.Action{
				apis.Action{
					ObjectMeta: meta.ObjectMeta{Name: "cmd_yolo_train_action"},
					Spec: apis.ActionSpec{
						Name: "TestAction",
						Runtimes: []apis.Runtime{
							apis.Runtime{
								Name:    "CMD",
								Image:   "",
								Type:    apis.ByCommand,
								Command: []string{"D:\\Programming\\Anaconda\\envs\\yolo\\python.exe"},
								Args:    []string{"D:\\Programming\\GoLand\\goProject\\all\\adaptive-scheduling-framework\\test\\nodelet\\task_exporter\\cmd_yolo\\yolo_task\\train.py"},
							},
						},
					},
					Status: apis.ActionStatus{
						ActionID: "cmd_yolo_train_action:test-group",
						Phase:    apis.Unknown,
						RuntimeStatus: []apis.RuntimeStatus{
							apis.RuntimeStatus{
								RuntimeID: "CMD:cmd_yolo_train_action:test-group",
								Phase:     apis.Unknown,
							},
						},
					},
				},
			},
		},
		Status: apis.GroupStatus{
			GroupID: "test-group",
			ActionStatus: []apis.ActionStatus{
				apis.ActionStatus{
					ActionID: "cmd_yolo_train_action:test-group",
					RuntimeStatus: []apis.RuntimeStatus{
						apis.RuntimeStatus{
							RuntimeID: "CMD:cmd_yolo_train_action:test-group",
							Phase:     apis.Unknown,
						},
					},
					Phase: apis.Unknown,
				},
			},
		},
	}
	groups := []*apis.Group{&newGroup}
	return groups
}

func yoloPredictAndTrainTaskGroup() []*apis.Group {
	newGroup := apis.Group{
		ObjectMeta: meta.ObjectMeta{Name: "cmd_yolo_train"},
		Spec: apis.GroupSpec{
			Name:    "TestGroup",
			Parents: make([]string, 0),
			Actions: []apis.Action{
				apis.Action{
					ObjectMeta: meta.ObjectMeta{Name: "cmd_yolo_train_action"},
					Spec: apis.ActionSpec{
						Name: "TestAction",
						Runtimes: []apis.Runtime{
							apis.Runtime{
								Name:    "CMD",
								Image:   "",
								Type:    apis.ByCommand,
								Command: []string{"D:\\Programming\\Anaconda\\envs\\yolo\\python.exe"},
								Args:    []string{"D:\\Programming\\GoLand\\goProject\\all\\adaptive-scheduling-framework\\test\\nodelet\\task_exporter\\cmd_yolo\\yolo_task\\predict.py"},
							},
							apis.Runtime{
								Name:    "ABC",
								Image:   "",
								Type:    apis.ByCommand,
								Command: []string{"D:\\Programming\\Anaconda\\envs\\yolo\\python.exe"},
								Args:    []string{"D:\\Programming\\GoLand\\goProject\\all\\adaptive-scheduling-framework\\test\\nodelet\\task_exporter\\cmd_yolo\\yolo_task\\train.py"},
								Parents: []string{"CMD"},
							},
						},
					},
					Status: apis.ActionStatus{
						ActionID: "cmd_yolo_train_action:test-group",
						Phase:    apis.Unknown,
						RuntimeStatus: []apis.RuntimeStatus{
							apis.RuntimeStatus{
								RuntimeID: "CMD:cmd_yolo_train_action:test-group", //RuntimeName +":"+ ActionID
								Phase:     apis.Unknown,
							},
							apis.RuntimeStatus{
								RuntimeID: "ABC:cmd_yolo_train_action:test-group", //RuntimeName +":"+ ActionID
								Phase:     apis.Unknown,
							},
						},
					},
				},
			},
		},
		Status: apis.GroupStatus{
			GroupID: "test-group",
			ActionStatus: []apis.ActionStatus{
				apis.ActionStatus{
					ActionID: "cmd_yolo_train_action:test-group",
					RuntimeStatus: []apis.RuntimeStatus{
						apis.RuntimeStatus{
							RuntimeID: "CMD:cmd_yolo_train_action:test-group",
							Phase:     apis.Unknown,
						},
						apis.RuntimeStatus{
							RuntimeID: "ABC:cmd_yolo_train_action:test-group", //RuntimeName +":"+ ActionID
							Phase:     apis.Unknown,
						},
					},
					Phase: apis.Unknown,
				},
			},
		},
	}
	groups := []*apis.Group{&newGroup}
	return groups
}

func yoloTrainTaskGroupInlinux() []*apis.Group {
	newGroup := apis.Group{
		ObjectMeta: meta.ObjectMeta{Name: "cmd_yolo_train"},
		Spec: apis.GroupSpec{
			Name:    "TestGroup",
			Parents: make([]string, 0),
			Actions: []apis.Action{
				apis.Action{
					ObjectMeta: meta.ObjectMeta{Name: "cmd_yolo_train_action"},
					Spec: apis.ActionSpec{
						Name: "TestAction",
						Runtimes: []apis.Runtime{
							apis.Runtime{
								Name:    "CMD",
								Image:   "",
								Type:    apis.ByCommand,
								Command: []string{"/home/public/anaconda3/envs/yolo/bin/python"},
								Args:    []string{"/home/public/workspace/heongtong_yolo_linux/train.py"},
							},
						},
					},
					Status: apis.ActionStatus{
						ActionID: "cmd_yolo_train_action:test-group",
						Phase:    apis.Unknown,
						RuntimeStatus: []apis.RuntimeStatus{
							apis.RuntimeStatus{
								RuntimeID: "CMD:cmd_yolo_train_action:test-group",
								Phase:     apis.Unknown,
							},
						},
					},
				},
			},
		},
		Status: apis.GroupStatus{
			GroupID: "test-group",
			ActionStatus: []apis.ActionStatus{
				apis.ActionStatus{
					ActionID: "cmd_yolo_train_action:test-group",
					RuntimeStatus: []apis.RuntimeStatus{
						apis.RuntimeStatus{
							RuntimeID: "CMD:cmd_yolo_train_action:test-group",
							Phase:     apis.Unknown,
						},
					},
					Phase: apis.Unknown,
				},
			},
		},
	}
	groups := []*apis.Group{&newGroup}
	return groups
}

func yoloPredictTaskGroup() []*apis.Group {
	newGroup := apis.Group{
		ObjectMeta: meta.ObjectMeta{Name: "cmd_yolo_train"},
		Spec: apis.GroupSpec{
			Name:    "TestGroup",
			Parents: make([]string, 0),
			Actions: []apis.Action{
				apis.Action{
					Spec: apis.ActionSpec{
						Name: "TestAction",
						Runtimes: []apis.Runtime{
							apis.Runtime{
								Name:    "CMD",
								Image:   "",
								Type:    apis.ByCommand,
								Command: []string{"D:\\Programming\\Anaconda\\envs\\yolo\\python.exe"},
								Args:    []string{"D:\\Programming\\GoLand\\goProject\\new2-task\\resourcelet\\test\\nodelet\\task_exporter\\cmd_yolo\\yolo_task\\predict.py"},
							},
						},
					},
				},
			},
		},
		Status: apis.GroupStatus{
			GroupID: "test-group",
		},
	}
	groups := []*apis.Group{&newGroup}
	return groups
}

func simpleTaskGroup() []*apis.Group {
	// 测试任务是否正确部署
	// 创建一个任务
	newGroup := apis.Group{
		ObjectMeta: meta.ObjectMeta{Name: "cmd_test"},
		Spec: apis.GroupSpec{
			Name:    "TestGroup",
			Parents: make([]string, 0), // 当前Group没有Parents
			Actions: []apis.Action{
				apis.Action{
					Spec: apis.ActionSpec{
						Name: "TestAction",
						Runtimes: []apis.Runtime{
							apis.Runtime{
								Name: "CMD",
								Type: apis.ByCommand,
								Command: []string{
									// "python",
									"ls",
								},
								Args: []string{
									// "/home/ysk/Desktop/datafolder/predict.py",
									"-a",
								},
							},
						},
					},
				},
			},
		},
		Status: apis.GroupStatus{
			GroupID: "test-group",
		},
	}
	groups := []*apis.Group{&newGroup}
	return groups
}

// predict 的测试
func CreateGroupActionRuntimePredict() (*apis.Group, *apis.Action, *apis.Runtime, *apis.Device) {

	// 能力框架的url
	manageUrl := "http://192.168.8.165:8080"
	imageType := "rgb"
	cameraUrl := "http://127.0.0.1:54533/api/status/camera"
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

	runtime1 := &apis.Runtime{
		Type:  apis.ByDevice,
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

	runtime2 := &apis.Runtime{
		Type:  apis.ByDevice,
		Image: "service_PredictByUrl",
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

	action1 := &apis.Action{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "action1p",
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
			Name: "action1p",
			Runtimes: []apis.Runtime{
				*runtime1,
			},
		},
		Status: apis.ActionStatus{
			ActionID:      "Action1p",
			Devices:       make(map[string]apis.DeviceStatus),
			RuntimeStatus: make([]apis.RuntimeStatus, 2),
		},
	}
	action2 := &apis.Action{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "action2p",
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
			Name: "action2p",
			Runtimes: []apis.Runtime{
				*runtime2,
			},
			Conditions: apis.Conditions{
				Formulas: []apis.ConditionFormula{
					apis.ConditionFormula{
						LeftValue: apis.ConditionValue{
							Type:      apis.ResultsData,
							Name:      "NodeDependency",
							Value:     "0",
							ValueType: "string",
							From:      action1.Name,
						},
						RightValue: apis.ConditionValue{
							Type:      apis.ConstData,
							Name:      "NodeDependency",
							Value:     "1",
							ValueType: "string",
							From:      "",
						},
						Signal: apis.Equal,
						Join:   "",
						Result: false,
					},
				},
			},
		},
		Status: apis.ActionStatus{
			ActionID:      "Action1p",
			Devices:       make(map[string]apis.DeviceStatus),
			RuntimeStatus: make([]apis.RuntimeStatus, 2),
		},
	}

	action1.Status.Devices["devicePredict"] = device.Status
	action2.Status.Devices["devicePredict"] = device.Status
	group := &apis.Group{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "groupp",
			Namespace: "test",
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Group",
			APIVersion: "resources/v1",
		},
		Spec: apis.GroupSpec{
			Name:     "groupp",
			Actions:  make([]apis.Action, 2),
			Replicas: []int32{0, 0},
		},
		Status: apis.GroupStatus{
			Node:    "test-node",
			GroupID: "groupp",
			Phase:   apis.ReadyToDeploy,
			ActionStatus: []apis.ActionStatus{
				action1.Status,
				action2.Status,
			},
		},
	}
	group.Spec.Actions[0] = *action1
	group.Spec.Actions[1] = *action2
	return group, action1, runtime1, device
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

	runtime1 := &apis.Runtime{
		Image: "manage_ArmControl.Leju.Guochuang",
		Name:  "RuntimeTest",
		Type:  apis.ByDevice,
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
	runtime2 := &apis.Runtime{
		Image: "service_LeftArmUp",
		Name:  "RuntimeTest",
		Type:  apis.ByDevice,
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
	runtime3 := &apis.Runtime{
		Image: "service_Sleep",
		Name:  "RuntimeTest",
		Type:  apis.ByDevice,
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
			{
				Name:  "sleepTime",
				Value: "5",
			},
		},
	}
	runtime4 := &apis.Runtime{
		Image: "service_LeftArmDown",
		Name:  "RuntimeTest",
		Type:  apis.ByDevice,
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

	action1 := &apis.Action{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "action1arm",
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
			Name: "action1arm",
			Runtimes: []apis.Runtime{
				*runtime1,
			},
		},
		Status: apis.ActionStatus{
			ActionID:      "Action1arm",
			Devices:       make(map[string]apis.DeviceStatus),
			RuntimeStatus: make([]apis.RuntimeStatus, 2),
		},
	}

	action2 := &apis.Action{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "action2arm",
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
			Name: "action2arm",
			Runtimes: []apis.Runtime{
				*runtime2,
			},
			Conditions: apis.Conditions{
				Formulas: []apis.ConditionFormula{
					apis.ConditionFormula{
						LeftValue: apis.ConditionValue{
							Type:      apis.ResultsData,
							Name:      "NodeDependency",
							Value:     "0",
							ValueType: "string",
							From:      action1.Name,
						},
						RightValue: apis.ConditionValue{
							Type:      apis.ConstData,
							Name:      "NodeDependency",
							Value:     "1",
							ValueType: "string",
							From:      "",
						},
						Signal: apis.Equal,
						Join:   "",
						Result: false,
					},
				},
			},
		},
		Status: apis.ActionStatus{
			RuntimeStatus: make([]apis.RuntimeStatus, 2),
			ActionID:      "Action2arm",
			Devices:       make(map[string]apis.DeviceStatus),
		},
	}

	action3 := &apis.Action{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "action3arm",
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
			Name: "action3arm",
			Runtimes: []apis.Runtime{
				*runtime3,
			},
			Conditions: apis.Conditions{
				Formulas: []apis.ConditionFormula{
					apis.ConditionFormula{
						LeftValue: apis.ConditionValue{
							Type:      apis.ResultsData,
							Name:      "NodeDependency",
							Value:     "0",
							ValueType: "string",
							From:      action2.Name,
						},
						RightValue: apis.ConditionValue{
							Type:      apis.ConstData,
							Name:      "NodeDependency",
							Value:     "1",
							ValueType: "string",
							From:      "",
						},
						Signal: apis.Equal,
						Join:   "",
						Result: false,
					},
				},
			},
		},
		Status: apis.ActionStatus{
			RuntimeStatus: make([]apis.RuntimeStatus, 2),
			ActionID:      "Action3arm",
			Devices:       make(map[string]apis.DeviceStatus),
		},
	}

	action4 := &apis.Action{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "action4arm",
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
			Name: "action4arm",
			Runtimes: []apis.Runtime{
				*runtime4,
			},
			Conditions: apis.Conditions{
				Formulas: []apis.ConditionFormula{
					apis.ConditionFormula{
						LeftValue: apis.ConditionValue{
							Type:      apis.ResultsData,
							Name:      "NodeDependency",
							Value:     "0",
							ValueType: "string",
							From:      action3.Name,
						},
						RightValue: apis.ConditionValue{
							Type:      apis.ConstData,
							Name:      "NodeDependency",
							Value:     "1",
							ValueType: "string",
							From:      "",
						},
						Signal: apis.Equal,
						Join:   "",
						Result: false,
					},
				},
			},
		},
		Status: apis.ActionStatus{
			RuntimeStatus: make([]apis.RuntimeStatus, 2),
			ActionID:      "Action4arm",
			Devices:       make(map[string]apis.DeviceStatus),
		},
	}

	action1.Status.Devices["deviceArm"] = device.Status
	action2.Status.Devices["deviceArm"] = device.Status
	action3.Status.Devices["deviceArm"] = device.Status
	action4.Status.Devices["deviceArm"] = device.Status

	group := &apis.Group{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "grouparm",
			Namespace: "test",
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Group",
			APIVersion: "resources/v1",
		},
		Spec: apis.GroupSpec{
			Replicas: []int32{0, 0},
			Name:     "grouparm",
			Actions:  make([]apis.Action, 4),
		},
		Status: apis.GroupStatus{
			Node:    "test-node",
			GroupID: "grouparm",
			Phase:   apis.ReadyToDeploy,
			ActionStatus: []apis.ActionStatus{
				action1.Status,
				action2.Status,
				action3.Status,
				action4.Status,
			},
		},
	}
	group.Spec.Actions[0] = *action1
	group.Spec.Actions[1] = *action2
	group.Spec.Actions[2] = *action3
	group.Spec.Actions[3] = *action4

	return group, action1, runtime1, device
}

func InitClient() (*clients.ClientSet, error) {
	//初始化ClientSet客户端
	scheme := runtime.NewScheme()
	apis.AddToScheme(scheme)
	c := &rest.Config{
		Host:    "http://localhost:10000",
		APIPath: "/apis/resources/v1",
		ContentConfig: rest.ContentConfig{
			AcceptContentTypes: "application/json; charset=UTF-8", //text/plain; charset=UTF-8
			ContentType:        "application/json; charset=UTF-8", //application/json; charset=UTF-8
			GroupVersion: &schema.GroupVersion{
				Group:   "resources",
				Version: "v1",
			},
			NegotiatedSerializer: serializer.NewCodecFactory(scheme),
		},
		UserAgent: "defaultUserAgent",
		Transport: &http.Transport{
			MaxIdleConns:        100,              // 最大空闲连接数
			IdleConnTimeout:     90 * time.Second, // 空闲连接超时时间
			TLSHandshakeTimeout: 10 * time.Second, // TLS 握手超时时间
		},
		Timeout: 10 * time.Second,
	}
	clientSet, err := clients.NewForConfig(c)
	if err != nil {
		return nil, fmt.Errorf("Failed to initialize clientSet: %v", err)
	}
	return clientSet, nil
}

func TestTaskExporter(t *testing.T) {

	// 初始化logs
	moduleName := "testModule"
	logs.Init(moduleName)

	ctx, _ := context.WithCancel(context.Background())

	// 构造Task Exporter
	tc := NewConfig("test-node")
	clientSet, err := InitClient()
	te, err := NewTaskExporter(tc, clientSet)
	if err != nil {
		panic(err)
	}

	// 创建测试用的group
	//testGroup := testTaskDeviceCreateGroup_RMF()

	g, a, _, d := CreateGroupActionRuntimeGrab()
	actionClient := clientSet.Core().Actions("test")
	deviceClient := clientSet.Core().Devices("test")

	err = actionClient.Delete(context.TODO(), a.Name, metav1.DeleteOptions{})
	if err != nil {
		logs.Errorf("Failed to delete action: %v", err)
	}
	err = actionClient.Delete(context.TODO(), "action2", metav1.DeleteOptions{})
	if err != nil {
		logs.Errorf("Failed to delete action: %v", err)
	}
	err = deviceClient.Delete(context.TODO(), d.Name, metav1.DeleteOptions{})
	err = actionClient.Delete(context.TODO(), "action3", metav1.DeleteOptions{})
	if err != nil {
		logs.Errorf("Failed to delete action: %v", err)
	}
	err = deviceClient.Delete(context.TODO(), d.Name, metav1.DeleteOptions{})
	err = actionClient.Delete(context.TODO(), "action4", metav1.DeleteOptions{})
	if err != nil {
		logs.Errorf("Failed to delete action: %v", err)
	}
	err = deviceClient.Delete(context.TODO(), d.Name, metav1.DeleteOptions{})
	if err != nil {
		logs.Errorf("Failed to delete device: %v", err)
	}
	err = te.gropsClient.Delete(ctx, g.Name, metav1.DeleteOptions{})
	if err != nil {
		logs.Errorf("Failed to delete group: %v", err)
	}
	//_, err = actionClient.Create(context.TODO(), a, metav1.CreateOptions{})
	//if err != nil {
	// logs.Errorf("create error %v", err)
	//}
	//testGroup := CreateTest()

	if err != nil {
		logs.Errorf("create error %v", err)
	}
	_, err = deviceClient.Create(context.TODO(), d, metav1.CreateOptions{})
	if err != nil {
		logs.Errorf("create error %v", err)
	}
	//将group存到数据总线中
	_, err = te.gropsClient.Create(ctx, g, metav1.CreateOptions{})
	//if err != nil {
	// logs.Errorf("Create group failed: %v", err)
	//
	//}

	// 开一个协程去更改数据总线中的testgroup的状态，ReceiveGroupInfo会自动检测并执行kill

	//go func() {
	// time.Sleep(15 * time.Second)
	// _, err = testTaskDeviceGroupKill(ctx, te.gropsClient, *testGroup)
	// if err != nil {
	//    logs.Error(err)
	// }
	//}()

	// 部署一个任务
	go func() {
		err2 := te.Run(ctx)
		if err2 != nil {
			logs.Error("fail to run task exporter")
		}
	}()
	//time.Sleep(2 * time.Second)
	//te.ReceiveGroupInfo(ctx)
	//Groups := yoloPredictAndTrainTaskGroup()
	//ReceiveGroupInfo(Groups, "create")
	//time.Sleep(60 * time.Second)
	//ReceiveGroupInfo(Groups, "kill")
	select {}
	// 部署多个任务
	//Groups[0].Spec.Name = "Test-Group2"
	//Groups[0].Status.GroupID = "Test-Group2"
	//Groups[0].Spec.Actions[0].Name = "Test-Actions"
	//Groups[0].Spec.Actions[0].Spec.Name = "Test-Actions"
	//Groups[0].Spec.Actions[0].Status.ActionID = "Test-Actions"
	//
	//time.Sleep(5 * time.Second)
	//ReceiveGroupInfo(Groups, "create")
}

func CreateGroupActionRuntimeGrab() (*apis.Group, *apis.Action, *apis.Runtime, *apis.Device) {
	// 能力框架的url
	manageUrl := "http://192.168.8.197:8080"
	taskTypeGrab := "0"
	// 创建device
	device := &apis.Device{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "deviceGrab",
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
			Name: "deviceGrab",
			AccessMethod: apis.AccessMethod{
				Type: apis.AccessByAbility,
				URL:  manageUrl,
			},
			ExpectedProperties: map[string]apis.Property{},
			Abilities:          make([]apis.AbilitySpec, 0),
		},
		Status: apis.DeviceStatus{
			DeviceID: "deviceGrab",
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
		Name: "Grab",
		Services: []apis.AbilityServiceStatus{
			{
				Name:      "TaskState",
				Ip:        "192.168.8.197",
				Interface: "/api/task_state",
			},
			{
				Name:      "GoStandBy",
				Ip:        "192.168.8.197",
				Interface: "/api/go_standby",
			},
			{
				Name:      "StartTask",
				Ip:        "192.168.8.197",
				Interface: "/api/start_task",
			},
			{
				Name:      "GoInit",
				Ip:        "192.168.8.197",
				Interface: "/api/go_initial",
			},
		},
	})

	runtime1 := &apis.Runtime{
		Image: "manage_ActInferenceAbility",
		Name:  "RuntimeTest",
		Type:  apis.ByDevice,
		Devices: []apis.DeviceSpec{
			device.Spec,
		},
		Outputs: make([]apis.Output, 1),
		Inputs: []apis.Input{
			{
				Name:  "taskType",
				Value: taskTypeGrab,
			},
		},
	}
	runtime2 := &apis.Runtime{
		Image: "service_GoInit",
		Name:  "RuntimeTest",
		Type:  apis.ByDevice,
		Devices: []apis.DeviceSpec{
			device.Spec,
		},
		Outputs: make([]apis.Output, 1),
		Inputs: []apis.Input{
			{
				Name:  "taskType",
				Value: taskTypeGrab,
			},
		},
	}

	runtime3 := &apis.Runtime{
		Image: "service_TaskState",
		Name:  "RuntimeTest",
		Type:  apis.ByDevice,
		Devices: []apis.DeviceSpec{
			device.Spec,
		},
		Outputs: make([]apis.Output, 1),
		Inputs: []apis.Input{
			{
				Name:  "taskType",
				Value: taskTypeGrab,
			},
		},
	}

	runtime4 := &apis.Runtime{
		Image: "service_GoStandBy",
		Name:  "RuntimeTest",
		Type:  apis.ByDevice,
		Devices: []apis.DeviceSpec{
			device.Spec,
		},
		Outputs: make([]apis.Output, 1),
		Inputs: []apis.Input{
			{
				Name:  "taskType",
				Value: taskTypeGrab,
			},
		},
	}

	runtime5 := &apis.Runtime{
		Image: "service_TaskState",
		Name:  "RuntimeTest",
		Type:  apis.ByDevice,
		Devices: []apis.DeviceSpec{
			device.Spec,
		},
		Outputs: make([]apis.Output, 1),
		Inputs: []apis.Input{
			{
				Name:  "taskType",
				Value: taskTypeGrab,
			},
		},
	}

	runtime6 := &apis.Runtime{
		Image: "service_StartTask",
		Name:  "RuntimeTest",
		Type:  apis.ByDevice,
		Devices: []apis.DeviceSpec{
			device.Spec,
		},
		Outputs: make([]apis.Output, 1),
		Inputs: []apis.Input{
			{
				Name:  "taskType",
				Value: taskTypeGrab,
			},
		},
	}

	runtime7 := &apis.Runtime{
		Image: "service_TaskState",
		Name:  "RuntimeTest",
		Type:  apis.ByDevice,
		Devices: []apis.DeviceSpec{
			device.Spec,
		},
		Outputs: make([]apis.Output, 1),
		Inputs: []apis.Input{
			{
				Name:  "taskType",
				Value: taskTypeGrab,
			},
		},
	}

	runtime8 := &apis.Runtime{
		Image: "manage_ActInferenceAbility_terminate",
		Name:  "RuntimeTest",
		Type:  apis.ByDevice,
		Devices: []apis.DeviceSpec{
			device.Spec,
		},
		Outputs: make([]apis.Output, 1),
		Inputs: []apis.Input{
			{
				Name:  "taskType",
				Value: taskTypeGrab,
			},
		},
	}

	action1 := &apis.Action{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "action1grab",
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
			Name: "action1grab",
			Runtimes: []apis.Runtime{
				*runtime1,
			},
		},
		Status: apis.ActionStatus{
			ActionID:      "Action1grab",
			Devices:       make(map[string]apis.DeviceStatus),
			RuntimeStatus: make([]apis.RuntimeStatus, 2),
		},
	}

	action2 := &apis.Action{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "action2grab",
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
			Name: "action2grab",
			Runtimes: []apis.Runtime{
				*runtime2,
			},
			Conditions: apis.Conditions{
				Formulas: []apis.ConditionFormula{
					apis.ConditionFormula{
						LeftValue: apis.ConditionValue{
							Type:      apis.ResultsData,
							Name:      "NodeDependency",
							Value:     "0",
							ValueType: "string",
							From:      action1.Name,
						},
						RightValue: apis.ConditionValue{
							Type:      apis.ConstData,
							Name:      "NodeDependency",
							Value:     "1",
							ValueType: "string",
							From:      "",
						},
						Signal: apis.Equal,
						Join:   "",
						Result: false,
					},
				},
			},
		},
		Status: apis.ActionStatus{
			RuntimeStatus: make([]apis.RuntimeStatus, 2),
			ActionID:      "Action2grab",
			Devices:       make(map[string]apis.DeviceStatus),
		},
	}

	action3 := &apis.Action{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "action3grab",
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
			Name: "action3grab",
			Runtimes: []apis.Runtime{
				*runtime3,
			},
			Conditions: apis.Conditions{
				Formulas: []apis.ConditionFormula{
					apis.ConditionFormula{
						LeftValue: apis.ConditionValue{
							Type:      apis.ResultsData,
							Name:      "NodeDependency",
							Value:     "0",
							ValueType: "string",
							From:      action2.Name,
						},
						RightValue: apis.ConditionValue{
							Type:      apis.ConstData,
							Name:      "NodeDependency",
							Value:     "1",
							ValueType: "string",
							From:      "",
						},
						Signal: apis.Equal,
						Join:   "",
						Result: false,
					},
				},
			},
		},
		Status: apis.ActionStatus{
			RuntimeStatus: make([]apis.RuntimeStatus, 2),
			ActionID:      "Action3grab",
			Devices:       make(map[string]apis.DeviceStatus),
		},
	}
	action4 := &apis.Action{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "action4grab",
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
			Name: "action4grab",
			Runtimes: []apis.Runtime{
				*runtime4,
			},
			Conditions: apis.Conditions{
				Formulas: []apis.ConditionFormula{
					apis.ConditionFormula{
						LeftValue: apis.ConditionValue{
							Type:      apis.ResultsData,
							Name:      "NodeDependency",
							Value:     "0",
							ValueType: "string",
							From:      action3.Name,
						},
						RightValue: apis.ConditionValue{
							Type:      apis.ConstData,
							Name:      "NodeDependency",
							Value:     "1",
							ValueType: "string",
							From:      "",
						},
						Signal: apis.Equal,
						Join:   "",
						Result: false,
					},
				},
			},
		},
		Status: apis.ActionStatus{
			RuntimeStatus: make([]apis.RuntimeStatus, 2),
			ActionID:      "Action4grab",
			Devices:       make(map[string]apis.DeviceStatus),
		},
	}

	action5 := &apis.Action{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "action5grab",
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
			Name: "action5grab",
			Runtimes: []apis.Runtime{
				*runtime5,
			},
			Conditions: apis.Conditions{
				Formulas: []apis.ConditionFormula{
					apis.ConditionFormula{
						LeftValue: apis.ConditionValue{
							Type:      apis.ResultsData,
							Name:      "NodeDependency",
							Value:     "0",
							ValueType: "string",
							From:      action4.Name,
						},
						RightValue: apis.ConditionValue{
							Type:      apis.ConstData,
							Name:      "NodeDependency",
							Value:     "1",
							ValueType: "string",
							From:      "",
						},
						Signal: apis.Equal,
						Join:   "",
						Result: false,
					},
				},
			},
		},
		Status: apis.ActionStatus{
			RuntimeStatus: make([]apis.RuntimeStatus, 2),
			ActionID:      "Action5grab",
			Devices:       make(map[string]apis.DeviceStatus),
		},
	}

	action6 := &apis.Action{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "action6grab",
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
			Name: "action6grab",
			Runtimes: []apis.Runtime{
				*runtime6,
			},
			Conditions: apis.Conditions{
				Formulas: []apis.ConditionFormula{
					apis.ConditionFormula{
						LeftValue: apis.ConditionValue{
							Type:      apis.ResultsData,
							Name:      "NodeDependency",
							Value:     "0",
							ValueType: "string",
							From:      action5.Name,
						},
						RightValue: apis.ConditionValue{
							Type:      apis.ConstData,
							Name:      "NodeDependency",
							Value:     "1",
							ValueType: "string",
							From:      "",
						},
						Signal: apis.Equal,
						Join:   "",
						Result: false,
					},
				},
			},
		},
		Status: apis.ActionStatus{
			RuntimeStatus: make([]apis.RuntimeStatus, 2),
			ActionID:      "Action6grab",
			Devices:       make(map[string]apis.DeviceStatus),
		},
	}

	action7 := &apis.Action{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "action7grab",
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
			Name: "action7grab",
			Runtimes: []apis.Runtime{
				*runtime7,
			},
			Conditions: apis.Conditions{
				Formulas: []apis.ConditionFormula{
					apis.ConditionFormula{
						LeftValue: apis.ConditionValue{
							Type:      apis.ResultsData,
							Name:      "NodeDependency",
							Value:     "0",
							ValueType: "string",
							From:      action6.Name,
						},
						RightValue: apis.ConditionValue{
							Type:      apis.ConstData,
							Name:      "NodeDependency",
							Value:     "1",
							ValueType: "string",
							From:      "",
						},
						Signal: apis.Equal,
						Join:   "",
						Result: false,
					},
				},
			},
		},
		Status: apis.ActionStatus{
			RuntimeStatus: make([]apis.RuntimeStatus, 2),
			ActionID:      "Action7grab",
			Devices:       make(map[string]apis.DeviceStatus),
		},
	}

	action8 := &apis.Action{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "action8grab",
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
			Name: "action8grab",
			Runtimes: []apis.Runtime{
				*runtime8,
			},
			Conditions: apis.Conditions{
				Formulas: []apis.ConditionFormula{
					apis.ConditionFormula{
						LeftValue: apis.ConditionValue{
							Type:      apis.ResultsData,
							Name:      "NodeDependency",
							Value:     "0",
							ValueType: "string",
							From:      action7.Name,
						},
						RightValue: apis.ConditionValue{
							Type:      apis.ConstData,
							Name:      "NodeDependency",
							Value:     "1",
							ValueType: "string",
							From:      "",
						},
						Signal: apis.Equal,
						Join:   "",
						Result: false,
					},
				},
			},
		},
		Status: apis.ActionStatus{
			RuntimeStatus: make([]apis.RuntimeStatus, 2),
			ActionID:      "action8grab",
			Devices:       make(map[string]apis.DeviceStatus),
		},
	}

	action1.Status.Devices["deviceArm"] = device.Status
	action2.Status.Devices["deviceArm"] = device.Status
	action3.Status.Devices["deviceArm"] = device.Status
	action4.Status.Devices["deviceArm"] = device.Status
	action5.Status.Devices["deviceArm"] = device.Status
	action6.Status.Devices["deviceArm"] = device.Status
	action7.Status.Devices["deviceArm"] = device.Status
	action8.Status.Devices["deviceArm"] = device.Status

	group := &apis.Group{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "groupgrab",
			Namespace: "test",
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Group",
			APIVersion: "resources/v1",
		},
		Spec: apis.GroupSpec{
			Replicas: []int32{0, 0},
			Name:     "groupgrab",
			Actions:  make([]apis.Action, 10),
		},
		Status: apis.GroupStatus{
			Node:    "test-node",
			GroupID: "group",
			Phase:   apis.ReadyToDeploy,
			ActionStatus: []apis.ActionStatus{
				action1.Status,
				action2.Status,
				action3.Status,
				action4.Status,
				action5.Status,
				action6.Status,
				action7.Status,
				action8.Status,
			},
		},
	}
	group.Spec.Actions[0] = *action1
	group.Spec.Actions[1] = *action2
	group.Spec.Actions[2] = *action3
	group.Spec.Actions[3] = *action4
	group.Spec.Actions[4] = *action5
	group.Spec.Actions[5] = *action6
	group.Spec.Actions[6] = *action7
	group.Spec.Actions[7] = *action8

	return group, action1, runtime1, device
}

func CreateGroupActionRuntimePut() (*apis.Group, *apis.Action, *apis.Runtime, *apis.Device) {
	// 能力框架的url
	manageUrl := "http://192.168.8.197:8080"
	taskTypePut := "1"
	// 创建device
	device := &apis.Device{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "deviceGrab",
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
			Name: "deviceGrab",
			AccessMethod: apis.AccessMethod{
				Type: apis.AccessByAbility,
				URL:  manageUrl,
			},
			ExpectedProperties: map[string]apis.Property{},
			Abilities:          make([]apis.AbilitySpec, 0),
		},
		Status: apis.DeviceStatus{
			DeviceID: "deviceGrab",
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
		Name: "Grab",
		Services: []apis.AbilityServiceStatus{
			{
				Name:      "TaskState",
				Ip:        "192.168.8.197",
				Interface: "/api/task_state",
			},
			{
				Name:      "GoStandBy",
				Ip:        "192.168.8.197",
				Interface: "/api/go_standby",
			},
			{
				Name:      "StartTask",
				Ip:        "192.168.8.197",
				Interface: "/api/start_task",
			},
			{
				Name:      "GoInit",
				Ip:        "192.168.8.197",
				Interface: "/api/go_initial",
			},
		},
	})

	runtime1 := &apis.Runtime{
		Image: "manage_ActInferenceAbility",
		Name:  "RuntimeTest",
		Type:  apis.ByDevice,
		Devices: []apis.DeviceSpec{
			device.Spec,
		},
		Outputs: make([]apis.Output, 1),
		Inputs: []apis.Input{
			{
				Name:  "taskType",
				Value: taskTypePut,
			},
		},
	}
	runtime2 := &apis.Runtime{
		Image: "service_GoInit",
		Name:  "RuntimeTest",
		Type:  apis.ByDevice,
		Devices: []apis.DeviceSpec{
			device.Spec,
		},
		Outputs: make([]apis.Output, 1),
		Inputs: []apis.Input{
			{
				Name:  "taskType",
				Value: taskTypePut,
			},
		},
	}

	runtime3 := &apis.Runtime{
		Image: "service_TaskState",
		Name:  "RuntimeTest",
		Type:  apis.ByDevice,
		Devices: []apis.DeviceSpec{
			device.Spec,
		},
		Outputs: make([]apis.Output, 1),
		Inputs: []apis.Input{
			{
				Name:  "taskType",
				Value: taskTypePut,
			},
		},
	}

	runtime4 := &apis.Runtime{
		Image: "service_GoStandBy",
		Name:  "RuntimeTest",
		Type:  apis.ByDevice,
		Devices: []apis.DeviceSpec{
			device.Spec,
		},
		Outputs: make([]apis.Output, 1),
		Inputs: []apis.Input{
			{
				Name:  "taskType",
				Value: taskTypePut,
			},
		},
	}

	runtime5 := &apis.Runtime{
		Image: "service_TaskState",
		Name:  "RuntimeTest",
		Type:  apis.ByDevice,
		Devices: []apis.DeviceSpec{
			device.Spec,
		},
		Outputs: make([]apis.Output, 1),
		Inputs: []apis.Input{
			{
				Name:  "taskType",
				Value: taskTypePut,
			},
		},
	}

	runtime6 := &apis.Runtime{
		Image: "service_StartTask",
		Name:  "RuntimeTest",
		Type:  apis.ByDevice,
		Devices: []apis.DeviceSpec{
			device.Spec,
		},
		Outputs: make([]apis.Output, 1),
		Inputs: []apis.Input{
			{
				Name:  "taskType",
				Value: taskTypePut,
			},
		},
	}

	runtime7 := &apis.Runtime{
		Image: "service_TaskState",
		Name:  "RuntimeTest",
		Type:  apis.ByDevice,
		Devices: []apis.DeviceSpec{
			device.Spec,
		},
		Outputs: make([]apis.Output, 1),
		Inputs: []apis.Input{
			{
				Name:  "taskType",
				Value: taskTypePut,
			},
		},
	}

	runtime8 := &apis.Runtime{
		Image: "manage_ActInferenceAbility_terminate",
		Name:  "RuntimeTest",
		Type:  apis.ByDevice,
		Devices: []apis.DeviceSpec{
			device.Spec,
		},
		Outputs: make([]apis.Output, 1),
		Inputs: []apis.Input{
			{
				Name:  "taskType",
				Value: taskTypePut,
			},
		},
	}

	action1 := &apis.Action{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "action1put",
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
			Name: "action1put",
			Runtimes: []apis.Runtime{
				*runtime1,
			},
		},
		Status: apis.ActionStatus{
			ActionID:      "Action1put",
			Devices:       make(map[string]apis.DeviceStatus),
			RuntimeStatus: make([]apis.RuntimeStatus, 2),
		},
	}

	action2 := &apis.Action{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "action2put",
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
			Name: "action2put",
			Runtimes: []apis.Runtime{
				*runtime2,
			},
			Conditions: apis.Conditions{
				Formulas: []apis.ConditionFormula{
					apis.ConditionFormula{
						LeftValue: apis.ConditionValue{
							Type:      apis.ResultsData,
							Name:      "NodeDependency",
							Value:     "0",
							ValueType: "string",
							From:      action1.Name,
						},
						RightValue: apis.ConditionValue{
							Type:      apis.ConstData,
							Name:      "NodeDependency",
							Value:     "1",
							ValueType: "string",
							From:      "",
						},
						Signal: apis.Equal,
						Join:   "",
						Result: false,
					},
				},
			},
		},
		Status: apis.ActionStatus{
			RuntimeStatus: make([]apis.RuntimeStatus, 2),
			ActionID:      "Action2put",
			Devices:       make(map[string]apis.DeviceStatus),
		},
	}

	action3 := &apis.Action{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "action3put",
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
			Name: "action3put",
			Runtimes: []apis.Runtime{
				*runtime3,
			},
			Conditions: apis.Conditions{
				Formulas: []apis.ConditionFormula{
					apis.ConditionFormula{
						LeftValue: apis.ConditionValue{
							Type:      apis.ResultsData,
							Name:      "NodeDependency",
							Value:     "0",
							ValueType: "string",
							From:      action2.Name,
						},
						RightValue: apis.ConditionValue{
							Type:      apis.ConstData,
							Name:      "NodeDependency",
							Value:     "1",
							ValueType: "string",
							From:      "",
						},
						Signal: apis.Equal,
						Join:   "",
						Result: false,
					},
				},
			},
		},
		Status: apis.ActionStatus{
			RuntimeStatus: make([]apis.RuntimeStatus, 2),
			ActionID:      "Action3put",
			Devices:       make(map[string]apis.DeviceStatus),
		},
	}
	action4 := &apis.Action{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "action4put",
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
			Name: "action4put",
			Runtimes: []apis.Runtime{
				*runtime4,
			},
			Conditions: apis.Conditions{
				Formulas: []apis.ConditionFormula{
					apis.ConditionFormula{
						LeftValue: apis.ConditionValue{
							Type:      apis.ResultsData,
							Name:      "NodeDependency",
							Value:     "0",
							ValueType: "string",
							From:      action3.Name,
						},
						RightValue: apis.ConditionValue{
							Type:      apis.ConstData,
							Name:      "NodeDependency",
							Value:     "1",
							ValueType: "string",
							From:      "",
						},
						Signal: apis.Equal,
						Join:   "",
						Result: false,
					},
				},
			},
		},
		Status: apis.ActionStatus{
			RuntimeStatus: make([]apis.RuntimeStatus, 2),
			ActionID:      "Action4put",
			Devices:       make(map[string]apis.DeviceStatus),
		},
	}

	action5 := &apis.Action{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "action5put",
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
			Name: "action5put",
			Runtimes: []apis.Runtime{
				*runtime5,
			},
			Conditions: apis.Conditions{
				Formulas: []apis.ConditionFormula{
					apis.ConditionFormula{
						LeftValue: apis.ConditionValue{
							Type:      apis.ResultsData,
							Name:      "NodeDependency",
							Value:     "0",
							ValueType: "string",
							From:      action4.Name,
						},
						RightValue: apis.ConditionValue{
							Type:      apis.ConstData,
							Name:      "NodeDependency",
							Value:     "1",
							ValueType: "string",
							From:      "",
						},
						Signal: apis.Equal,
						Join:   "",
						Result: false,
					},
				},
			},
		},
		Status: apis.ActionStatus{
			RuntimeStatus: make([]apis.RuntimeStatus, 2),
			ActionID:      "Action5put",
			Devices:       make(map[string]apis.DeviceStatus),
		},
	}

	action6 := &apis.Action{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "action6put",
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
			Name: "action6put",
			Runtimes: []apis.Runtime{
				*runtime6,
			},
			Conditions: apis.Conditions{
				Formulas: []apis.ConditionFormula{
					apis.ConditionFormula{
						LeftValue: apis.ConditionValue{
							Type:      apis.ResultsData,
							Name:      "NodeDependency",
							Value:     "0",
							ValueType: "string",
							From:      action5.Name,
						},
						RightValue: apis.ConditionValue{
							Type:      apis.ConstData,
							Name:      "NodeDependency",
							Value:     "1",
							ValueType: "string",
							From:      "",
						},
						Signal: apis.Equal,
						Join:   "",
						Result: false,
					},
				},
			},
		},
		Status: apis.ActionStatus{
			RuntimeStatus: make([]apis.RuntimeStatus, 2),
			ActionID:      "Action6put",
			Devices:       make(map[string]apis.DeviceStatus),
		},
	}

	action7 := &apis.Action{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "action7put",
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
			Name: "action7put",
			Runtimes: []apis.Runtime{
				*runtime7,
			},
			Conditions: apis.Conditions{
				Formulas: []apis.ConditionFormula{
					apis.ConditionFormula{
						LeftValue: apis.ConditionValue{
							Type:      apis.ResultsData,
							Name:      "NodeDependency",
							Value:     "0",
							ValueType: "string",
							From:      action6.Name,
						},
						RightValue: apis.ConditionValue{
							Type:      apis.ConstData,
							Name:      "NodeDependency",
							Value:     "1",
							ValueType: "string",
							From:      "",
						},
						Signal: apis.Equal,
						Join:   "",
						Result: false,
					},
				},
			},
		},
		Status: apis.ActionStatus{
			RuntimeStatus: make([]apis.RuntimeStatus, 2),
			ActionID:      "Action7put",
			Devices:       make(map[string]apis.DeviceStatus),
		},
	}

	action8 := &apis.Action{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "action8put",
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
			Name: "action8put",
			Runtimes: []apis.Runtime{
				*runtime8,
			},
			Conditions: apis.Conditions{
				Formulas: []apis.ConditionFormula{
					apis.ConditionFormula{
						LeftValue: apis.ConditionValue{
							Type:      apis.ResultsData,
							Name:      "NodeDependency",
							Value:     "0",
							ValueType: "string",
							From:      action7.Name,
						},
						RightValue: apis.ConditionValue{
							Type:      apis.ConstData,
							Name:      "NodeDependency",
							Value:     "1",
							ValueType: "string",
							From:      "",
						},
						Signal: apis.Equal,
						Join:   "",
						Result: false,
					},
				},
			},
		},
		Status: apis.ActionStatus{
			RuntimeStatus: make([]apis.RuntimeStatus, 2),
			ActionID:      "action8put",
			Devices:       make(map[string]apis.DeviceStatus),
		},
	}

	action1.Status.Devices["deviceArm"] = device.Status
	action2.Status.Devices["deviceArm"] = device.Status
	action3.Status.Devices["deviceArm"] = device.Status
	action4.Status.Devices["deviceArm"] = device.Status
	action5.Status.Devices["deviceArm"] = device.Status
	action6.Status.Devices["deviceArm"] = device.Status
	action7.Status.Devices["deviceArm"] = device.Status
	action8.Status.Devices["deviceArm"] = device.Status

	group := &apis.Group{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "groupput",
			Namespace: "test",
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Group",
			APIVersion: "resources/v1",
		},
		Spec: apis.GroupSpec{
			Replicas: []int32{0, 0},
			Name:     "groupput",
			Actions:  make([]apis.Action, 10),
		},
		Status: apis.GroupStatus{
			Node:    "test-node",
			GroupID: "group",
			Phase:   apis.ReadyToDeploy,
			ActionStatus: []apis.ActionStatus{
				action1.Status,
				action2.Status,
				action3.Status,
				action4.Status,
				action5.Status,
				action6.Status,
				action7.Status,
				action8.Status,
			},
		},
	}
	group.Spec.Actions[0] = *action1
	group.Spec.Actions[1] = *action2
	group.Spec.Actions[2] = *action3
	group.Spec.Actions[3] = *action4
	group.Spec.Actions[4] = *action5
	group.Spec.Actions[5] = *action6
	group.Spec.Actions[6] = *action7
	group.Spec.Actions[7] = *action8

	return group, action1, runtime1, device
}

// go test -run TestWorkFlow -v
func TestWorkFlow(t *testing.T) {
	// 初始化logs
	moduleName := "testModule"
	logs.Init(moduleName)

	ctx, _ := context.WithCancel(context.Background())

	// 构造Task Exporter
	tc := NewConfig("test-node")
	clientSet, err := InitClient()
	te, err := NewTaskExporter(tc, clientSet)
	if err != nil {
		panic(err)
	}
	te.gropsClient.Delete(context.TODO(), "groupgrab", metav1.DeleteOptions{})
	te.gropsClient.Delete(context.TODO(), "groupput", metav1.DeleteOptions{})
	te.gropsClient.Delete(context.TODO(), "grouparm", metav1.DeleteOptions{})
	te.gropsClient.Delete(context.TODO(), "groupp", metav1.DeleteOptions{})

	// 先把task exporter拉起来
	go func() {
		err2 := te.Run(ctx)
		if err2 != nil {
			logs.Error("fail to run task exporter")
		}
	}()

	// 测试predict
	//构造任务和设备
	predicrGroup, predictAction, _, predictDevice := CreateGroupActionRuntimePredict()
	actionClient := clientSet.Core().Actions("test")
	deviceClient := clientSet.Core().Devices("test")

	// 删除put的8个action
	actionClient.Delete(ctx, "action8put", metav1.DeleteOptions{})
	actionClient.Delete(ctx, "action7put", metav1.DeleteOptions{})
	actionClient.Delete(ctx, "action6put", metav1.DeleteOptions{})
	actionClient.Delete(ctx, "action5put", metav1.DeleteOptions{})
	actionClient.Delete(ctx, "action4put", metav1.DeleteOptions{})
	actionClient.Delete(ctx, "action3put", metav1.DeleteOptions{})
	actionClient.Delete(ctx, "action2put", metav1.DeleteOptions{})
	actionClient.Delete(ctx, "action1put", metav1.DeleteOptions{})

	// 删除grab的8个action
	actionClient.Delete(ctx, "action8grab", metav1.DeleteOptions{})
	actionClient.Delete(ctx, "action7grab", metav1.DeleteOptions{})
	actionClient.Delete(ctx, "action6grab", metav1.DeleteOptions{})
	actionClient.Delete(ctx, "action5grab", metav1.DeleteOptions{})
	actionClient.Delete(ctx, "action4grab", metav1.DeleteOptions{})
	actionClient.Delete(ctx, "action3grab", metav1.DeleteOptions{})
	actionClient.Delete(ctx, "action2grab", metav1.DeleteOptions{})
	actionClient.Delete(ctx, "action1grab", metav1.DeleteOptions{})

	// 删除arm的4个action
	actionClient.Delete(ctx, "action1arm", metav1.DeleteOptions{})
	actionClient.Delete(ctx, "action2arm", metav1.DeleteOptions{})
	actionClient.Delete(ctx, "action3arm", metav1.DeleteOptions{})
	actionClient.Delete(ctx, "action4arm", metav1.DeleteOptions{})

	// 删除predict的2个action
	actionClient.Delete(ctx, "action1p", metav1.DeleteOptions{})
	actionClient.Delete(ctx, "action2p", metav1.DeleteOptions{})

	// 删除三个device
	deviceClient.Delete(ctx, "deviceArm", metav1.DeleteOptions{})
	deviceClient.Delete(ctx, "deviceGrab", metav1.DeleteOptions{})
	deviceClient.Delete(ctx, "devicePredict", metav1.DeleteOptions{})

	//删除历史遗留的设备和任务（如果有的话）
	err = deviceClient.Delete(ctx, predictDevice.Name, metav1.DeleteOptions{})
	if err != nil {
		logs.Errorf("fail to delete predictDevice: %v", err)
	}
	err = actionClient.Delete(ctx, predictAction.Name, metav1.DeleteOptions{})
	if err != nil {
		logs.Errorf("fail to delete predictAction: %v", err)
	}
	err = te.gropsClient.Delete(ctx, predicrGroup.Name, metav1.DeleteOptions{})
	if err != nil {
		logs.Errorf("Failed to delete predicrGroup: %v", err)
	}
	err = actionClient.Delete(ctx, "action2", metav1.DeleteOptions{})
	if err != nil {
		logs.Errorf("fail to delete predictAction: %v", err)
	}

	//添加任务和设备
	_, err = deviceClient.Create(context.TODO(), predictDevice, metav1.CreateOptions{})
	if err != nil {
		logs.Errorf("create error %v", err)
	}

	//将group存到数据总线中
	_, err = te.gropsClient.Create(ctx, predicrGroup, metav1.CreateOptions{})
	if err != nil {
		logs.Errorf("create error %v", err)
	}

	logs.Info("create predict group done , wait for 10s")

	time.Sleep(10 * time.Second)

	//查询predict任务完成状态
	pg, err := te.gropsClient.Get(ctx, predicrGroup.Spec.Name, metav1.GetOptions{})
	if err != nil {
		logs.Errorf("get error %v", err)
		return
	}
	if pg.Status.Phase != apis.Successed {
		logs.Error("predict group is not succeed")
	}

	//再把东西删一遍
	//err = deviceClient.Delete(ctx, predictDevice.Name, metav1.DeleteOptions{})
	//if err != nil {
	// logs.Errorf("fail to delete predictDevice: %v", err)
	//}
	//err = actionClient.Delete(ctx, predictAction.Name, metav1.DeleteOptions{})
	//if err != nil {
	// logs.Errorf("fail to delete predictAction: %v", err)
	//}
	//err = te.gropsClient.Delete(ctx, predicrGroup.Name, metav1.DeleteOptions{})
	//if err != nil {
	// logs.Errorf("Failed to delete predicrGroup: %v", err)
	//}

	logs.Infof("run grab....")
	ggrab, _, _, dgrab := CreateGroupActionRuntimeGrab()
	_, err = deviceClient.Create(context.TODO(), dgrab, metav1.CreateOptions{})
	_, err = te.gropsClient.Create(context.TODO(), ggrab, metav1.CreateOptions{})
	time.Sleep(10 * time.Second)

	logs.Info("run arms ....")
	garm, _, _, darm := CreateGroupActionRuntimeArm()
	_, err = deviceClient.Create(context.TODO(), darm, metav1.CreateOptions{})
	_, err = te.gropsClient.Create(context.TODO(), garm, metav1.CreateOptions{})
	time.Sleep(10 * time.Second)

	logs.Info("run put....")
	gput, _, _, dput := CreateGroupActionRuntimePut()
	_, err = deviceClient.Create(context.TODO(), dput, metav1.CreateOptions{})
	_, err = te.gropsClient.Create(context.TODO(), gput, metav1.CreateOptions{})

	select {}
}

// go test -run TestTerminate -v
func TestTerminate(t *testing.T) {
	url1 := "http://192.168.8.165:8080"
	url2 := "http://192.168.8.197:8080"
	abilityName1 := "ArmControl.Leju.Guochuang"
	abilityName2 := "Detect"
	abilityName3 := "ActInferenceAbility"
	err := manager.NewAbilityManager(url1, abilityName1).TerminateAbility()
	if err != nil {
		logs.Errorf("fail to create AbilityManager: %v", err)
	}
	err = manager.NewAbilityManager(url1, abilityName2).TerminateAbility()
	if err != nil {
		logs.Errorf("fail to create AbilityManager: %v", err)
	}
	err = manager.NewAbilityManager(url2, abilityName3).TerminateAbility()
	if err != nil {
		logs.Errorf("fail to create AbilityManager: %v", err)
	}
}

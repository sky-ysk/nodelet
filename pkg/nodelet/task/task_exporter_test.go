package task

import (
	"context"
	"fmt"
	"hit.edu/framework/pkg/apimachinery/runtime"
	"hit.edu/framework/pkg/apimachinery/runtime/schema"
	"hit.edu/framework/pkg/apimachinery/runtime/serializer"
	metav1 "hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/client-go/clients"
	"hit.edu/framework/pkg/client-go/clients/typed/core"
	"hit.edu/framework/pkg/client-go/rest"
	"net/http"
	"strconv"
	"testing"
	"time"

	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/component-base/logs"
)

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
	cameraUrl := "http://127.0.0.1:33107/api/status/camera"
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
				*runtime1,
			},
		},
		Status: apis.ActionStatus{
			ActionID:      "Action1",
			Devices:       make(map[string]apis.DeviceStatus),
			RuntimeStatus: make([]apis.RuntimeStatus, 2),
		},
	}
	action2 := &apis.Action{
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
			ActionID:      "Action1",
			Devices:       make(map[string]apis.DeviceStatus),
			RuntimeStatus: make([]apis.RuntimeStatus, 2),
		},
	}

	action1.Status.Devices["devicePredict"] = device.Status
	action2.Status.Devices["devicePredict"] = device.Status
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
			Name:     "group",
			Actions:  make([]apis.Action, 1),
			Replicas: []int32{0, 0},
		},
		Status: apis.GroupStatus{
			Node:    "test-node",
			GroupID: "group",
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
			ActionID: "Action1",
			Devices:  make(map[string]apis.DeviceStatus),
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
			Node:    "test-node",
			GroupID: "group",
			ActionStatus: []apis.ActionStatus{
				action.Status,
			},
		},
	}
	group.Spec.Actions[0] = *action
	return group, action, runtime, device
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

func testTaskDeviceGroupKill(ctx context.Context, groupClient core.GroupInterface, group apis.Group) (*apis.Group, error) {
	g, err := groupClient.Get(ctx, group.Name, meta.GetOptions{})
	if err != nil {
		logs.Error(err)
		return nil, err
	}
	g.Status.Phase = apis.ReadyToKill
	g.Spec.Actions[0].Status.RuntimeStatus[0].Phase = apis.Running
	_, err = groupClient.Update(ctx, g, meta.UpdateOptions{})
	if err != nil {
		logs.Error(err)
		return nil, err
	}
	return g, nil
}
func createDemoDevice() *apis.Device {
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
				URL:   "http://192.168.1.225:8000",
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

	g, a, _, d := CreateGroupActionRuntimePredict()
	actionClient := clientSet.Core().Actions("test")
	deviceClient := clientSet.Core().Devices("test")

	err = actionClient.Delete(context.TODO(), a.Name, metav1.DeleteOptions{})
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
	//	logs.Errorf("create error %v", err)
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
	//	logs.Errorf("Create group failed: %v", err)
	//
	//}

	// 开一个协程去更改数据总线中的testgroup的状态，ReceiveGroupInfo会自动检测并执行kill

	//go func() {
	//	time.Sleep(15 * time.Second)
	//	_, err = testTaskDeviceGroupKill(ctx, te.gropsClient, *testGroup)
	//	if err != nil {
	//		logs.Error(err)
	//	}
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

	time.Sleep(3 * time.Second)

	//查询predict任务完成状态
	pg, err := te.gropsClient.Get(ctx, predicrGroup.Spec.Name, metav1.GetOptions{})
	if err != nil {
		logs.Errorf("get error %v", err)
		return
	}
	if pg.Status.Phase != apis.Successed {
		logs.Fatal("predict group is not succeed")
	}

	//再把东西删一遍
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

	//TODO 测试抬手臂

	//TODO 测试放手臂

	//TODO 测试夹爪

}

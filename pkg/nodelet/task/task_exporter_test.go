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

func testTaskDeviceCreateGroup_RMF() *apis.Group {
	// 测试任务是否正确部署
	// 创建一个任务
	newGroup := apis.Group{
		ObjectMeta: meta.ObjectMeta{Name: "device_test"},
		Spec: apis.GroupSpec{
			Name:    "Test-Group-Device",
			Parents: make([]string, 0), // 当前Group没有Parents
			Actions: []apis.Action{
				apis.Action{
					Spec: apis.ActionSpec{
						Name: "Test-Action",
						Runtimes: []apis.Runtime{
							{
								Name: "device_test",
								Type: apis.ByDevice,
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
								Image: "Move",
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
					ActionID: "test-action-device",
					Phase:    apis.ReadyToDeploy,
					Devices: []apis.DeviceStatus{
						apis.DeviceStatus{
							Lock: apis.Lock{
								IsLocked: true,
								Ref:      1,
							},
							Status:     "idle",
							Phase:      apis.DeviceIdle,
							InstanceID: "",
							ActionID:   "",
							DeviceID:   "transferRobot",
						},
					},
				},
			},
			Phase: apis.ReadyToDeploy,
		},
	}

	return &newGroup
}

// testTaskDeviceCreateGroup_Ability 创建ability的测试用例
func testTaskDeviceCreateGroup_Ability() *apis.Group {

	// 填写url和image
	var url string = "http://127.0.0.1:8123"
	var image string = "Mock"
	newGroup := apis.Group{
		ObjectMeta: meta.ObjectMeta{Name: "device_test"},
		Spec: apis.GroupSpec{
			Replicas: []int32{0, 0},
			Name:     "device_test",
			Parents:  make([]string, 0), // 当前Group没有Parents
			Actions: []apis.Action{
				apis.Action{

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
					Status: apis.ActionStatus{
						RuntimeStatus: []apis.RuntimeStatus{
							apis.RuntimeStatus{},
						},
						ActionID: "test-action-device",
						Phase:    apis.ReadyToDeploy,
						Devices: []apis.DeviceStatus{
							apis.DeviceStatus{
								Lock: apis.Lock{
									IsLocked: true,
									Ref:      1,
								},
								Status:     "idle",
								Phase:      apis.DeviceIdle,
								InstanceID: "",
								ActionID:   "",
								DeviceID:   "transferRobot",
							},
						},
					},
					Spec: apis.ActionSpec{
						Name: "Test-Action",
						Runtimes: []apis.Runtime{
							{
								Name: "device_test",
								Type: apis.ByDevice,
								Devices: []apis.DeviceSpec{
									apis.DeviceSpec{
										Name:               "mock test",
										ExpectedProperties: map[string]apis.Property{},
										AccessMethod: apis.AccessMethod{
											Type:  apis.AccessByAbility,
											URL:   url,
											Group: "tinyRobot",
											Alias: "transferRobot",
										},
										Desc: apis.DeviceDesc{
											Label: []string{"Mock"},
										},
									},
								},
								Outputs: apis.Output{},
								Image:   image,
							},
						},
					},
				},
			},
		},
		Status: apis.GroupStatus{
			Node:    "test-node",
			GroupID: "device_test",
			ActionStatus: []apis.ActionStatus{
				apis.ActionStatus{
					RuntimeStatus: []apis.RuntimeStatus{
						apis.RuntimeStatus{},
					},
					ActionID: "test-action-device",
					Phase:    apis.ReadyToDeploy,
					Devices: []apis.DeviceStatus{
						apis.DeviceStatus{
							Lock: apis.Lock{
								IsLocked: true,
								Ref:      1,
							},
							Status:     "idle",
							Phase:      apis.DeviceIdle,
							InstanceID: "",
							ActionID:   "",
							DeviceID:   "transferRobot",
						},
					},
				},
			},
			Phase: apis.ReadyToDeploy,
		},
	}

	return &newGroup
}

func CreateTest() apis.Group {

	// Task  总共1个Task、1个Group、1个Action、1个runtime
	task1Name := "Test-Task-Device" // 第一个Task的Name
	task1ID := "Test-Task-DeviceID" // 第一个Task的ID

	// group
	group1_1Name := "Test-Group-Device" // 第一个Task下的第一个GroupName
	//group1_2Name := "ReasonGroup-2" // 第一个Task下的第二个GroupName
	group1_1ID := "Test-Group-DeviceID" // 第一个Task下的第一个GroupID
	//group1_2ID := "GroupID-2"       // 第一个Task下的第二个GroupID

	// action
	action1_1_1Name := "Action1-1" // 第一个Task下的第一个Group下的第一个ActionName  "cmd_yolo_train_action"
	//action1_2_1Name := "Action2-1" // 第一个Task下的第二个Group下的第一个ActionName
	action1_1_1ID := "ActionID1-1" // 第一个Task下的第一个Group下的第一个ActionID
	//action1_2_1ID := "ActionID2-1" // 第一个Task下的第二个Group下的第一个ActionID

	// runtime
	runtime1_1_1_1Name := "Runtime1-1-1" // 第一个Task下的第一个Group下的第一个ActionName下的第一个RuntimeName
	//runtime1_2_1_1Name := "Runtime2-1-1" // 第一个Task下的第二个Group下的第一个ActionName下的第一个RuntimeName

	runtime1_1_1_1ID := "RuntimeID1-1-1" // 第一个Task下的第一个Group下的第一个ActionName下的第一个RuntimeID
	//runtime1_2_1_1ID := "RuntimeID2-1-1" // 第一个Task下的第二个Group下的第一个ActionName下的第一个RuntimeID

	// runtime是否细粒度控制
	runtime1_1_1_1FineGrainedControl := false
	//runtime1_2_1_1FineGrainedControl := false

	// 统一地规定： Belongs：填的是ID
	//            Parents: 填的也是ID吧--改为Name
	runtime1_1_1_1Condition := apis.Conditions{
		Formulas: []apis.ConditionFormula{},
	}
	//runtime1_1_1_1Input := []apis.Input{
	//	apis.Input{
	//		Type:      apis.LocalData,
	//		Name:      "dest",
	//		Value:     "R201",
	//		ValueType: "string",
	//	},
	//	apis.Input{
	//		Type:      apis.LocalData,
	//		Name:      "orientation",
	//		Value:     "-3.12",
	//		ValueType: "double",
	//	},
	//	apis.Input{
	//		Type:      apis.LocalData,
	//		Name:      "dock",
	//		Value:     "true",
	//		ValueType: "bool",
	//	},
	//}

	//runtime1_2_1_1Condition := apis.Conditions{
	//	Formulas: []apis.ConditionFormula{
	//		apis.ConditionFormula{
	//			LeftValue: apis.ConditionValue{
	//				Type:      apis.ResultsData,
	//				Name:      "ProgramDependency",
	//				Value:     "0",
	//				ValueType: "string",
	//				From:      "/home/l1hy/workspace/task_input/requirements.txt",
	//			},
	//			RightValue: apis.ConditionValue{
	//				Type:      apis.ConstData,
	//				Name:      "ProgramDependency",
	//				Value:     "1",
	//				ValueType: "string",
	//				From:      "",
	//			},
	//			Signal: apis.Equal,
	//			Join:   "",
	//			Result: false,
	//		},
	//	},
	//}

	g1 := apis.Group{
		ObjectMeta: metav1.ObjectMeta{Name: group1_1Name, Namespace: ""},
		TypeMeta:   metav1.TypeMeta{Kind: "Group", APIVersion: "resources/v1"},
		Spec: apis.GroupSpec{
			Name:    group1_1Name,
			Parents: make([]string, 0),
			Actions: []apis.Action{
				apis.Action{
					ObjectMeta: metav1.ObjectMeta{Name: action1_1_1Name},
					Spec: apis.ActionSpec{
						Name: action1_1_1Name,
						Runtimes: []apis.Runtime{
							apis.Runtime{
								Name: runtime1_1_1_1Name,
								Type: apis.ByDevice,
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
								Parents:    make([]string, 0), // 加入Parents
								Conditions: runtime1_1_1_1Condition,
								Image:      "Move",
								//Inputs:                   runtime1_1_1_1Input,
								EnvVar:                   []apis.EnvVar{apis.EnvVar{Name: "", Value: ""}},
								EnableFineGrainedControl: runtime1_1_1_1FineGrainedControl,
							},
						},
					},
					Status: apis.ActionStatus{
						ActionID: action1_1_1ID, // ActionID =ActionName + GroupID
						Phase:    apis.Unknown,
						RuntimeStatus: []apis.RuntimeStatus{
							apis.RuntimeStatus{
								RuntimeID: runtime1_1_1_1ID, // RuntimeID = RuntimeName + ActionID
								Phase:     apis.Unknown,
							},
						},
					},
				},
			},
		},
		Status: apis.GroupStatus{
			GroupID: group1_1ID,
			ActionStatus: []apis.ActionStatus{
				apis.ActionStatus{
					ActionID: action1_1_1ID,
					Devices: []apis.DeviceStatus{
						apis.DeviceStatus{
							Lock: apis.Lock{
								IsLocked: true,
								Ref:      1,
							},
							Status:     "idle",
							Phase:      apis.DeviceIdle,
							InstanceID: "",
							ActionID:   "",
							DeviceID:   "transferRobot",
						},
					},
					RuntimeStatus: []apis.RuntimeStatus{
						apis.RuntimeStatus{
							RuntimeID: runtime1_1_1_1ID,
							Phase:     apis.Unknown,
						},
					},
					Phase: apis.ReadyToDeploy,
				},
			},
			Belongs: apis.IDRef{TaskID: task1ID},
			Phase:   apis.Unknown,
		},
	}
	//group1 := &g1

	//g2 := apis.Group{
	//	ObjectMeta: metav1.ObjectMeta{Name: group1_2Name, Namespace: ""},
	//	TypeMeta:   metav1.TypeMeta{Kind: "Group", APIVersion: "resources/v1"},
	//	Spec: apis.GroupSpec{
	//		Name:    group1_2Name,
	//		Parents: []string{},
	//		Actions: []apis.Action{
	//			apis.Action{
	//				ObjectMeta: metav1.ObjectMeta{Name: action1_2_1Name},
	//				Spec: apis.ActionSpec{
	//					Name: action1_2_1Name,
	//					Runtimes: []apis.Runtime{
	//						apis.Runtime{
	//							Name:                     runtime1_2_1_1Name,
	//							Type:                     apis.ByCommand,
	//							Command:                  []string{"python"},
	//							Args:                     []string{"/home/l1hy/workspace/heongtong_yolo_linux/predict.py"},
	//							Parents:                  make([]string, 0), // 加入Parents
	//							Conditions:               runtime1_2_1_1Condition,
	//							Image:                    "/home/l1hy/workspace/heongtong_yolo_linux/predict.py",
	//							EnvVar:                   []apis.EnvVar{apis.EnvVar{Name: "", Value: ""}},
	//							EnableFineGrainedControl: runtime1_2_1_1FineGrainedControl,
	//						},
	//					},
	//				},
	//				Status: apis.ActionStatus{
	//					ActionID: action1_2_1ID,
	//					Phase:    apis.Unknown,
	//					RuntimeStatus: []apis.RuntimeStatus{
	//						apis.RuntimeStatus{
	//							RuntimeID: runtime1_2_1_1ID,
	//							Phase:     apis.Unknown,
	//						},
	//					},
	//				},
	//			},
	//		},
	//	},
	//	Status: apis.GroupStatus{
	//		GroupID: group1_2ID,
	//		ActionStatus: []apis.ActionStatus{
	//			apis.ActionStatus{
	//				ActionID: action1_2_1ID,
	//				RuntimeStatus: []apis.RuntimeStatus{
	//					apis.RuntimeStatus{
	//						RuntimeID: runtime1_2_1_1ID,
	//						Phase:     apis.Unknown,
	//					},
	//				},
	//				Phase: apis.Unknown,
	//			},
	//		},
	//		Belongs: apis.IDRef{TaskID: task1ID},
	//		Phase:   apis.Unknown,
	//	},
	//}
	//group2 := &g2

	_ = &apis.Task{
		ObjectMeta: metav1.ObjectMeta{
			Name:      task1Name,
			Namespace: "",
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Task",
			APIVersion: "resources/v1",
		},
		Spec: apis.TaskSpec{
			Name: task1Name,
			Groups: []apis.Group{
				g1,
			},
		},
		Status: apis.TaskStatus{
			TaskID: task1ID,
			Phase:  apis.Unknown,
			GroupStatus: []apis.GroupStatus{
				apis.GroupStatus{
					GroupID: group1_1ID,
					ActionStatus: []apis.ActionStatus{
						apis.ActionStatus{
							ActionID: action1_1_1ID,
							RuntimeStatus: []apis.RuntimeStatus{
								apis.RuntimeStatus{
									RuntimeID: runtime1_1_1_1ID,
									Phase:     apis.Unknown,
								},
							},
							Phase: apis.Unknown,
						},
					},
					Belongs: apis.IDRef{TaskID: task1ID},
					Phase:   apis.Unknown,
				},
			},
		},
	}
	return g1
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

	testGroup := testTaskDeviceCreateGroup_Ability()

	//testGroup := CreateTest()
	deviceClient := clientSet.Core().Devices("test")
	deviceDemo := createDemoDevice()
	_, err = deviceClient.Create(context.TODO(), deviceDemo, metav1.CreateOptions{})

	// 将group存到数据总线中
	_, err = te.gropsClient.Create(ctx, testGroup, metav1.CreateOptions{})
	if err != nil {
		logs.Errorf("Create group failed: %v", err)
		return
	}

	// 开一个协程去更改数据总线中的testgroup的状态，ReceiveGroupInfo会自动检测并执行kill

	go func() {
		time.Sleep(15 * time.Second)
		_, err = testTaskDeviceGroupKill(ctx, te.gropsClient, *testGroup)
		if err != nil {
			logs.Error(err)
		}
	}()

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

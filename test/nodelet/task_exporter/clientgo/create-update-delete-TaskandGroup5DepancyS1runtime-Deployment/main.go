package main

import (
	"bufio"
	"context"
	"fmt"
	"hit.edu/framework/pkg/apimachinery/runtime"
	"hit.edu/framework/pkg/apimachinery/runtime/schema"
	"hit.edu/framework/pkg/apimachinery/runtime/serializer"
	"hit.edu/framework/pkg/apimachinery/watch"
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/apis/meta"
	metav1 "hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/client-go/clients"
	"hit.edu/framework/pkg/client-go/rest"
	"hit.edu/framework/pkg/component-base/logs"
	"hit.edu/framework/pkg/nodelet/task/runtime/k8s/config"
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/utils/pointer"
	"net/http"
	"os"
	"time"
)

// 创建一个Rest Client
// 验证xxx动词
// 与API Server通信，并执行基础操作
// 1个group，1个Action，每个Action一个Runtime， 一共1个Runtime，runtime为Deployment
func main() {
	moduleName := "testModule"
	logs.Init(moduleName)
	scheme := runtime.NewScheme()
	apis.AddToScheme(scheme)
	logs.Info(scheme)
	//参数配置
	// TODO: 填写参数
	//部分参数之后可以在core_client等 编写setConfigDefaults函数进行填充
	c := &rest.Config{
		Host:    "http://localhost:10000", //http://suda801.wangwanu.com:11006
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

	//创建ClientSet
	clientSet, err := clients.NewForConfig(c)
	if err != nil {
		panic(err)
	}

	tasksClient := clientSet.Core().Tasks("test")
	groupsClient := clientSet.Core().Groups("test")
	actionClient := clientSet.Core().Actions("test")
	eventclient := clientSet.Core().Events("test")

	// Task  总共1个Task、3个Group、3个Action、6个runtime
	task1Name := "ServiceTask" // 第一个Task的Name
	task1ID := "ServiceTaskID" // 第一个Task的ID

	// group
	group1_1Name := "TrainGroup-1" // 第一个Task下的第一个GroupName
	//group1_2Name := "ReasonGroup-2"           // 第一个Task下的第二个GroupName
	//group1_3Name := "RobotDestinationGroup-3" // 第一个Task下的第三个GroupName  "RobotDestinationGroup"
	group1_1ID := "GroupID-1" // 第一个Task下的第一个GroupID
	//group1_2ID := "GroupID-2"                 // 第一个Task下的第二个GroupID
	//group1_3ID := "GroupID-3"                 // 第一个Task下的第三个GroupID

	// action
	action1_1_1Name := "Action1-1" // 第一个Task下的第一个Group下的第一个ActionName  "cmd_yolo_train_action"
	//action1_2_1Name := "Action2-1" // 第一个Task下的第二个Group下的第一个ActionName
	//action1_3_1Name := "Action3-1" // 第一个Task下的第三个Group下的第一个ActionName
	action1_1_1ID := "ActionID1-1" // 第一个Task下的第一个Group下的第一个ActionID
	//action1_2_1ID := "ActionID2-1" // 第一个Task下的第二个Group下的第一个ActionID
	//action1_3_1ID := "ActionID3-1" // 第一个Task下的第三个Group下的第一个ActionID

	// runtime
	runtime1_1_1_1Name := "Runtime1-1-1" // 第一个Task下的第一个Group下的第一个ActionName下的第一个RuntimeName
	//runtime1_1_1_2Name := "Runtime1-1-2" // 第一个Task下的第一个Group下的第一个ActionName下的第二个RuntimeName
	//runtime1_2_1_1Name := "Runtime2-1-1" // 第一个Task下的第二个Group下的第一个ActionName下的第一个RuntimeName
	//runtime1_2_1_2Name := "Runtime2-1-2" // 第一个Task下的第二个Group下的第一个ActionName下的第二个RuntimeName
	//runtime1_3_1_1Name := "Runtime3-1-1" // 第一个Task下的第三个Group下的第一个ActionName下的第一个RuntimeName
	//runtime1_3_1_2Name := "Runtime3-1-2" // 第一个Task下的第三个Group下的第一个ActionName下的第二个RuntimeName

	runtime1_1_1_1ID := "RuntimeID1-1-1" // 第一个Task下的第一个Group下的第一个ActionName下的第一个RuntimeID
	//runtime1_1_1_2ID := "RuntimeID1-1-2" // 第一个Task下的第一个Group下的第一个ActionName下的第二个RuntimeID
	//runtime1_2_1_1ID := "RuntimeID2-1-1" // 第一个Task下的第二个Group下的第一个ActionName下的第一个RuntimeID
	//runtime1_2_1_2ID := "RuntimeID2-1-2" // 第一个Task下的第二个Group下的第一个ActionName下的第二个RuntimeID
	//runtime1_3_1_1ID := "RuntimeID3-1-1" // 第一个Task下的第三个Group下的第一个ActionName下的第一个RuntimeID
	//runtime1_3_1_2ID := "RuntimeID3-1-2" // 第一个Task下的第三个Group下的第一个ActionName下的第二个RuntimeID

	// runtime是否细粒度控制
	runtime1_1_1_1FineGrainedControl := false
	//runtime1_1_1_2FineGrainedControl := true
	//runtime1_2_1_1FineGrainedControl := false
	//runtime1_2_1_2FineGrainedControl := false
	//runtime1_3_1_1FineGrainedControl := false
	//runtime1_3_1_2FineGrainedControl := false
	// 统一地规定： Belongs：填的是ID
	//            Parents: 填的也是ID吧--改为Name
	runtime1_1_1_1Condition := apis.Conditions{
		Formulas: []apis.ConditionFormula{
			//apis.ConditionFormula{
			//	LeftValue: apis.ConditionValue{
			//		Type:      apis.ResultsData,
			//		Name:      "ProgramDependency",
			//		Value:     "0",
			//		ValueType: "string",
			//		From:      "/home/l1hy/workspace/task_input/requirements.txt",
			//	},
			//	RightValue: apis.ConditionValue{
			//		Type:      apis.ConstData,
			//		Name:      "ProgramDependency",
			//		Value:     "1",
			//		ValueType: "string",
			//		From:      "",
			//	},
			//	Signal: apis.Equal,
			//	Join:   "",
			//	Result: false,
			//},
		},
	}
	//runtime1_1_1_2Condition := apis.Conditions{
	//	Formulas: []apis.ConditionFormula{
	//		apis.ConditionFormula{
	//			LeftValue: apis.ConditionValue{
	//				Type:      apis.ResultsData,
	//				Name:      "NodeDependency",
	//				Value:     "0",
	//				ValueType: "string",
	//				From:      runtime1_1_1_1Name,
	//			},
	//			RightValue: apis.ConditionValue{
	//				Type:      apis.ConstData,
	//				Name:      "NodeDependency",
	//				Value:     "1",
	//				ValueType: "string",
	//				From:      "",
	//			},
	//			Signal: apis.Equal,
	//			Join:   "",
	//			Result: false,
	//		},
	//		//apis.ConditionFormula{
	//		//	LeftValue: apis.ConditionValue{
	//		//		Type:      apis.ResultsData,
	//		//		Name:      "ProgramDependency",
	//		//		Value:     "0",
	//		//		ValueType: "string",
	//		//		From:      "/home/l1hy/workspace/task_input/requirements.txt",
	//		//	},
	//		//	RightValue: apis.ConditionValue{
	//		//		Type:      apis.ConstData,
	//		//		Name:      "ProgramDependency",
	//		//		Value:     "1",
	//		//		ValueType: "string",
	//		//		From:      "",
	//		//	},
	//		//	Signal: apis.Equal,
	//		//	Join:   "",
	//		//	Result: false,
	//		//},
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
	//runtime1_2_1_2Condition := apis.Conditions{
	//	Formulas: []apis.ConditionFormula{
	//		apis.ConditionFormula{
	//			LeftValue: apis.ConditionValue{
	//				Type:      apis.ResultsData,
	//				Name:      "NodeDependency",
	//				Value:     "0",
	//				ValueType: "strting",
	//				From:      runtime1_2_1_1Name,
	//			},
	//			RightValue: apis.ConditionValue{
	//				Type:      apis.ConstData,
	//				Name:      "NodeDependency",
	//				Value:     "1",
	//				ValueType: "string",
	//				From:      "",
	//			},
	//			Signal: apis.Equal,
	//			Join:   "",
	//			Result: false,
	//		},
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
	//
	//runtime1_3_1_1Condition := apis.Conditions{
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
	//runtime1_3_1_2Condition := apis.Conditions{
	//	Formulas: []apis.ConditionFormula{
	//		apis.ConditionFormula{
	//			LeftValue: apis.ConditionValue{
	//				Type:      apis.ResultsData,
	//				Name:      "NodeDependency",
	//				Value:     "0",
	//				ValueType: "strting",
	//				From:      runtime1_3_1_1Name,
	//			},
	//			RightValue: apis.ConditionValue{
	//				Type:      apis.ConstData,
	//				Name:      "NodeDependency",
	//				Value:     "1",
	//				ValueType: "string",
	//				From:      "",
	//			},
	//			Signal: apis.Equal,
	//			Join:   "",
	//			Result: false,
	//		},
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

	//group1_2Condition := apis.Conditions{
	//	Formulas: []apis.ConditionFormula{
	//		apis.ConditionFormula{
	//			LeftValue: apis.ConditionValue{
	//				Type:      apis.ResultsData,
	//				Name:      "NodeDependency",
	//				Value:     "0",
	//				ValueType: "strting",
	//				From:      group1_1Name,
	//			},
	//			RightValue: apis.ConditionValue{
	//				Type:      apis.ConstData,
	//				Name:      "NodeDependency",
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
	//
	//group1_3Condition := apis.Conditions{
	//	Formulas: []apis.ConditionFormula{
	//		apis.ConditionFormula{
	//			LeftValue: apis.ConditionValue{
	//				Type:      apis.ResultsData,
	//				Name:      "NodeDependency",
	//				Value:     "0",
	//				ValueType: "strting",
	//				From:      group1_2Name,
	//			},
	//			RightValue: apis.ConditionValue{
	//				Type:      apis.ConstData,
	//				Name:      "NodeDependency",
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
	deploymentName := "grpc-server-deploy"
	deploymentNameNamespace := "switch"
	deploymentReplicas := int32(1)

	g1 := apis.Group{
		ObjectMeta: metav1.ObjectMeta{Name: group1_1Name, Namespace: ""},
		TypeMeta:   metav1.TypeMeta{Kind: "Group", APIVersion: "resources/v1"},
		Spec: apis.GroupSpec{
			Name:     group1_1Name,
			Parents:  make([]string, 0),
			Replicas: []int32{0, 0},
			Actions: []apis.Action{
				apis.Action{
					ObjectMeta: metav1.ObjectMeta{Name: action1_1_1Name},
					Spec: apis.ActionSpec{
						Name: action1_1_1Name,
						Runtimes: []apis.Runtime{
							apis.Runtime{
								Name:                     runtime1_1_1_1Name,
								Type:                     apis.ByDeployment,
								Command:                  []string{},
								Args:                     []string{},
								Parents:                  []string{runtime1_1_1_1Name}, // 加入Parents
								Conditions:               runtime1_1_1_1Condition,
								Image:                    "",
								EnvVar:                   []apis.EnvVar{apis.EnvVar{Name: "", Value: ""}},
								EnableFineGrainedControl: runtime1_1_1_1FineGrainedControl,
								Deployment: apis.Deployment{
									TypeMeta:   metav1.TypeMeta{Kind: "Deployment", APIVersion: "v1"},
									ObjectMeta: metav1.ObjectMeta{Name: deploymentName, Namespace: deploymentNameNamespace},
									Spec: apis.DeploymentSpec{
										Replicas: deploymentReplicas,
										Selector: metav1.LabelSelector{
											MatchLabels: map[string]string{"app": "grpc-server"},
										},
										Template: apis.PodTemplateSpec{
											ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"app": "grpc-server"}},
											Spec: apis.PodSpec{
												Containers: []apis.Container{
													apis.Container{
														Name:  "grpc-server",
														Image: "registry.cn-hangzhou.aliyuncs.com/sudasuzhou/grpc_service:grpc-server-pod-v1.0",
														Ports: []apis.ContainerPort{
															apis.ContainerPort{
																ContainerPort: 50051,
															},
														},
													},
												},
											},
										},
									},
								},
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
							//apis.RuntimeStatus{
							//	RuntimeID: runtime1_1_1_2ID, // RuntimeID = RuntimeName + ActionID
							//	Phase:     apis.Unknown,
							//},
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
					RuntimeStatus: []apis.RuntimeStatus{
						apis.RuntimeStatus{
							RuntimeID: runtime1_1_1_1ID,
							Phase:     apis.Unknown,
						},
						//apis.RuntimeStatus{
						//	RuntimeID: runtime1_1_1_2ID,
						//	Phase:     apis.Unknown,
						//},
					},
					Phase: apis.Unknown,
				},
			},
			Belongs: apis.IDRef{TaskID: task1ID},
			Phase:   apis.Unknown,
		},
	}
	group1 := &g1
	//
	//g2 := apis.Group{
	//	ObjectMeta: metav1.ObjectMeta{Name: group1_2Name, Namespace: ""},
	//	TypeMeta:   metav1.TypeMeta{Kind: "Group", APIVersion: "resources/v1"},
	//	Spec: apis.GroupSpec{
	//		Name:       group1_2Name,
	//		Parents:    []string{group1_1Name}, // 加入Parents
	//		Conditions: group1_2Condition,
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
	//						apis.Runtime{
	//							Name:                     runtime1_2_1_2Name,
	//							Type:                     apis.ByCommand,
	//							Command:                  []string{"python"},
	//							Args:                     []string{"/home/l1hy/workspace/heongtong_yolo_linux/predict.py"},
	//							Parents:                  []string{runtime1_2_1_1Name}, // 加入Parents
	//							Conditions:               runtime1_2_1_2Condition,
	//							Image:                    "/home/l1hy/workspace/heongtong_yolo_linux/predict.py",
	//							EnvVar:                   []apis.EnvVar{apis.EnvVar{Name: "", Value: ""}},
	//							EnableFineGrainedControl: runtime1_2_1_2FineGrainedControl,
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
	//						apis.RuntimeStatus{
	//							RuntimeID: runtime1_2_1_2ID,
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
	//					apis.RuntimeStatus{
	//						RuntimeID: runtime1_2_1_2ID,
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
	//
	//g3 := apis.Group{
	//	ObjectMeta: metav1.ObjectMeta{Name: group1_3Name, Namespace: ""},
	//	TypeMeta:   metav1.TypeMeta{Kind: "Group", APIVersion: "resources/v1"},
	//	Spec: apis.GroupSpec{
	//		Name:       group1_3Name,
	//		Parents:    []string{group1_1Name, group1_2Name}, // 加入Parents ID
	//		Conditions: group1_3Condition,
	//		Actions: []apis.Action{
	//			apis.Action{
	//				ObjectMeta: metav1.ObjectMeta{Name: action1_3_1Name},
	//				Spec: apis.ActionSpec{
	//					Name: action1_3_1Name,
	//					Runtimes: []apis.Runtime{
	//						apis.Runtime{
	//							Name:                     runtime1_3_1_1Name,
	//							Type:                     apis.ByCommand,
	//							Command:                  []string{"python"},
	//							Args:                     []string{"/home/l1hy/workspace/heongtong_yolo_linux/predict.py"},
	//							Parents:                  make([]string, 0), // 加入Parents
	//							Conditions:               runtime1_3_1_1Condition,
	//							Image:                    "/home/l1hy/workspace/heongtong_yolo_linux/predict.py",
	//							EnvVar:                   []apis.EnvVar{apis.EnvVar{Name: "", Value: ""}},
	//							EnableFineGrainedControl: runtime1_3_1_1FineGrainedControl,
	//						},
	//						apis.Runtime{
	//							Name:                     runtime1_3_1_2Name,
	//							Type:                     apis.ByCommand,
	//							Command:                  []string{"python"},
	//							Args:                     []string{"/home/l1hy/workspace/heongtong_yolo_linux/predict.py"},
	//							Parents:                  []string{runtime1_3_1_1Name}, // 加入Parents
	//							Conditions:               runtime1_3_1_2Condition,
	//							Image:                    "/home/l1hy/workspace/heongtong_yolo_linux/predict.py",
	//							EnvVar:                   []apis.EnvVar{apis.EnvVar{Name: "", Value: ""}},
	//							EnableFineGrainedControl: runtime1_3_1_2FineGrainedControl,
	//						},
	//					},
	//				},
	//				Status: apis.ActionStatus{
	//					ActionID: action1_3_1ID, // ActionID =ActionName + GroupID
	//					Phase:    apis.Unknown,
	//					RuntimeStatus: []apis.RuntimeStatus{
	//						apis.RuntimeStatus{
	//							RuntimeID: runtime1_3_1_1ID, // RuntimeID = RuntimeName + ActionID
	//							Phase:     apis.Unknown,
	//						},
	//						apis.RuntimeStatus{
	//							RuntimeID: runtime1_3_1_2ID, // RuntimeID = RuntimeName + ActionID
	//							Phase:     apis.Unknown,
	//						},
	//					},
	//				},
	//			},
	//		},
	//	},
	//	Status: apis.GroupStatus{
	//		GroupID: group1_3ID,
	//		ActionStatus: []apis.ActionStatus{
	//			apis.ActionStatus{
	//				ActionID: action1_3_1ID,
	//				RuntimeStatus: []apis.RuntimeStatus{
	//					apis.RuntimeStatus{
	//						RuntimeID: runtime1_3_1_1ID,
	//						Phase:     apis.Unknown,
	//					},
	//					apis.RuntimeStatus{
	//						RuntimeID: runtime1_3_1_2ID,
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
	//group3 := &g3

	task := &apis.Task{
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
								//apis.RuntimeStatus{
								//	RuntimeID: runtime1_1_1_2ID,
								//	Phase:     apis.Unknown,
								//},
							},
							Phase: apis.Unknown,
						},
					},
					Belongs: apis.IDRef{TaskID: task1ID},
					Phase:   apis.Unknown,
				},
				//apis.GroupStatus{
				//	GroupID: group1_2ID,
				//	ActionStatus: []apis.ActionStatus{
				//		apis.ActionStatus{
				//			ActionID: action1_2_1ID,
				//			RuntimeStatus: []apis.RuntimeStatus{
				//				apis.RuntimeStatus{
				//					RuntimeID: runtime1_2_1_1ID,
				//					Phase:     apis.Unknown,
				//				},
				//				apis.RuntimeStatus{
				//					RuntimeID: runtime1_2_1_2ID,
				//					Phase:     apis.Unknown,
				//				},
				//			},
				//			Phase: apis.Unknown,
				//		},
				//	},
				//	Belongs: apis.IDRef{TaskID: task1ID},
				//	Phase:   apis.Unknown,
				//},
				//apis.GroupStatus{
				//	GroupID: group1_3ID,
				//	ActionStatus: []apis.ActionStatus{
				//		apis.ActionStatus{
				//			ActionID: action1_3_1ID,
				//			RuntimeStatus: []apis.RuntimeStatus{
				//				apis.RuntimeStatus{
				//					RuntimeID: runtime1_3_1_1ID,
				//					Phase:     apis.Unknown,
				//				},
				//				apis.RuntimeStatus{
				//					RuntimeID: runtime1_3_1_2ID,
				//					Phase:     apis.Unknown,
				//				},
				//			},
				//			Phase: apis.Unknown,
				//		},
				//	},
				//	Belongs: apis.IDRef{TaskID: task1ID},
				//	Phase:   apis.Unknown,
				//},
			},
		},
	}

	//监听事件并打印  监听resources/v1/tasks
	go func() {
		logs.Infof("watching")
		watchOptions := metav1.ListOptions{}

		watcher, err := tasksClient.Watch(context.TODO(), watchOptions)
		if err != nil {
			panic(err)
		}
		defer watcher.Stop() // 确保 watcher 被停止

		// 获取事件通道
		watchChan := watcher.ResultChan()

		for {
			select {
			case event, ok := <-watchChan:
				if !ok {
					logs.Infof("watchChan closed")
					return
				}

				// 打印事件类型和对象的相关信息
				logs.Infof("接收到事件类型: %v\n", event.Type)
				switch event.Type {
				case watch.Added:
					logs.Trace("资源被添加: ", event.Object)
				case watch.Modified:
					logs.Trace("资源被修改: ", event.Object)
				case watch.Deleted:
					logs.Trace("资源被删除: ", event.Object)
				case watch.Error:
					logs.Trace("发生错误: ", event.Object)
				default:
					logs.Infof("未识别的事件类型: ", event.Type)
				}
			}
		}
	}()
	// 删除事件
	list, err := eventclient.List(context.TODO(), meta.ListOptions{})
	if err != nil {
		logs.Error(err.Error())
	}
	for _, item := range list.Items {
		err := eventclient.Delete(context.TODO(), item.Name, meta.DeleteOptions{})
		if err != nil {
			logs.Error(err.Error())
		}
	}

	//如果已经存在，先删掉
	////err = tasksClient.Delete(context.TODO(), "TrainInferTask", metav1.DeleteOptions{})
	clientset := config.LoadConfig()
	DeleteDeployment(clientset, deploymentName, deploymentNameNamespace, deploymentReplicas)
	err = tasksClient.Delete(context.TODO(), task1Name, metav1.DeleteOptions{})
	err1 := groupsClient.Delete(context.TODO(), group1_1Name, metav1.DeleteOptions{})
	err2 := actionClient.Delete(context.TODO(), action1_1_1Name, metav1.DeleteOptions{})
	//err2 := groupsClient.Delete(context.TODO(), "ReasonGroup-2", metav1.DeleteOptions{})
	//err3 := groupsClient.Delete(context.TODO(), "RobotDestinationGroup-3", metav1.DeleteOptions{})
	if err != nil {
		logs.Errorf("task delete error: %v", err)
	}
	if err1 != nil {
		logs.Errorf("group1 delete error: %v", err1)
	}
	if err2 != nil {
		logs.Errorf("action1 delete error: %v", err2)
	}
	//if err3 != nil {
	//	logs.Errorf("group3 delete error: %v", err3)
	//}
	// Create一个Task
	logs.Infof("creating")
	_, err = tasksClient.Create(context.TODO(), task, metav1.CreateOptions{})
	_, err1 = groupsClient.Create(context.TODO(), group1, metav1.CreateOptions{})
	//results2, err2 := groupsClient.Create(context.TODO(), group2, metav1.CreateOptions{})
	//results3, err3 := groupsClient.Create(context.TODO(), group3, metav1.CreateOptions{})

	if err != nil {
		logs.Errorf("Failed to create task: %v", err)
		panic(err)
	}
	if err1 != nil {
		logs.Errorf("Failed to create group1: %v", err1)
		panic(err)
	}
	//if err2 != nil {
	//	logs.Errorf("Failed to create group2: %v", err)
	//	panic(err)
	//}
	//if err3 != nil {
	//	logs.Errorf("Failed to create group3: %v", err)
	//	panic(err)
	//}
	//_, _ = tasksClient.Create(context.TODO(), task2, metav1.CreateOptions{})
	//_, _ = tasksClient.Create(context.TODO(), task3, metav1.CreateOptions{})
	//logs.Infof("Created task ", results)
	//logs.Infof("Created group1 ", results1)
	//logs.Infof("Created group2 ", results2)
	//logs.Infof("Created group3 ", results3)
	//prompt()

	//Update一个Task

	//logs.Info("updating")
	//// 部分更改一个参数
	//// 先Get一个Task ,更改Task的参数, UpdateTask
	//
	//result, getErr := tasksClient.Get(context.TODO(), "TrainInferTask", metav1.GetOptions{})
	//if getErr != nil {
	//	panic(fmt.Errorf("Failed to get : %v", getErr))
	//}
	//
	//logs.Infof("get result", result)
	//logs.Infof("修改前的result.Spec.Name：", result.Spec.Name)
	//
	//result.Spec.Name = "updatedTaskName"
	//_, updateErr := tasksClient.Update(context.TODO(), result, metav1.UpdateOptions{})
	//if updateErr != nil {
	//	panic(fmt.Errorf("Update failed: %v", updateErr))
	//}
	//
	//logs.Infof("修改后的result.Spec.Name：", result.Spec.Name)
	//logs.Info("Updated task...")
	//prompt()

	//// List 所有Task
	//logs.Info("listing")
	//lstOpts := metav1.ListOptions{}
	//list, err := tasksClient.List(context.TODO(), lstOpts)
	//list1, err1 := groupsClient.List(context.TODO(), lstOpts)
	//if err != nil {
	//	panic(err)
	//}
	//if err1 != nil {
	//	panic(err)
	//}
	//for _, d := range list.Items {
	//	logs.Info(d)
	//}
	//for _, d := range list1.Items {
	//	logs.Info(d)
	//}
	//logs.Infof("listing done")
	//prompt()

	////Patch 一个Task
	//logs.Infof("patching")
	//patchResult, err := tasksClient.Patch(context.TODO(), "TrainInferTask", types.StrategicMergePatchType, patchTask, metav1.PatchOptions{})
	//logs.Infof("patchResult: ", patchResult)
	//logs.Infof("patch Done")
	//
	//// List 所有Task
	//logs.Info("listing")
	//lstOpts = metav1.ListOptions{}
	//list, err = tasksClient.List(context.TODO(), lstOpts)
	//if err != nil {
	//	panic(err)
	//}
	//for _, d := range list.Items {
	//	logs.Info(d)
	//}
	//
	//logs.Infof("listing done")
	//prompt()

	// Delete一个Task 和两个Group
	//logs.Info("deleting")
	//err = tasksClient.Delete(context.TODO(), "TrainInferTask", metav1.DeleteOptions{})
	//err1 = groupsClient.Delete(context.TODO(), "demo-group1", metav1.DeleteOptions{})
	//err2 = groupsClient.Delete(context.TODO(), "demo-group2", metav1.DeleteOptions{})
	//if err != nil {
	//	panic(err)
	//}
	//if err1 != nil {
	//	panic(err)
	//}
	//if err2 != nil {
	//	panic(err)
	//}
	//logs.Info("Deleted task...")
	//logs.Info("Deleted group1、group2...")
	//prompt()
	//
	//// Delete 之后再次 List所有Task
	//logs.Info("listing")
	//lstOpts = metav1.ListOptions{}
	//list, err = tasksClient.List(context.TODO(), lstOpts)
	//if err != nil {
	//	panic(err)
	//}
	//for _, d := range list.Items {
	//	logs.Info(d)
	//}
	//
	//logs.Infof("listing done")
	//
	//select {}

	////DeleteCollection 删除所有Spec.TaskName=TrainInferTask的Task
	//logs.Infof("deleting collection")
	//lstOpts = metav1.ListOptions{
	//	FieldSelector: "Spec.TaskName=TrainInferTask",
	//}
	//err = tasksClient.DeleteCollection(context.TODO(), metav1.DeleteOptions{}, lstOpts)
	//if err != nil {
	//	panic(err)
	//}
	//logs.Infof("Deleted collection...")
	//prompt()
	//
	//// DeleteCollection 之后再次 List所有Task
	//logs.Infof("listing")
	//lstOpts = metav1.ListOptions{}
	//list, err = tasksClient.List(context.TODO(), lstOpts)
	//if err != nil {
	//	panic(err)
	//}
	//for _, d := range list.Items {
	//	logs.Infof(d)
	//}
	//logs.Infof("listing done")
}

// From K8s
func prompt() {
	fmt.Printf("-> Press Return key to continue.")
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		break
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}
	logs.Info()
}
func DeletePod(clientset *kubernetes.Clientset, podName, podNamespace string) {
	// 使用与创建时一致的日志记录风格
	err := clientset.CoreV1().Pods(podNamespace).Delete(
		context.TODO(),
		podName, // 直接从Pod对象获取名称
		k8smetav1.DeleteOptions{
			GracePeriodSeconds: pointer.Int64Ptr(5), // 可选：优雅删除等待时间
		},
	)

	if err != nil {
		// 带上下文的错误日志，保持与你的风格一致
		logs.Error(err, "删除Pod失败", "Pod名称", podName, "命名空间", podNamespace)
	} else {
		// 成功日志包含结构化参数
		logs.Info("Pod删除成功",
			"Pod名称", podName,
			"命名空间", podNamespace,
			"删除时间", time.Now().Format(time.RFC3339))
	}

	// 如果EM需要清理，可以在此处调用
	// EM.GetInstance().RemovePod(pod.Namespace, pod.Name)
}

// 删除 Deployment
func DeleteDeployment(clientset *kubernetes.Clientset, deploymentName, deploymentNamespace string, replicas int32) {
	err := clientset.AppsV1().Deployments(deploymentNamespace).Delete(
		context.TODO(),
		deploymentName,
		k8smetav1.DeleteOptions{
			PropagationPolicy: func() *k8smetav1.DeletionPropagation {
				policy := k8smetav1.DeletePropagationBackground
				return &policy
			}(), // 确保级联删除关联的 Pod
		},
	)

	if err != nil {
		logs.Error(err, "删除Deployment失败",
			"Deployment名称", deploymentName,
			"命名空间", deploymentNamespace)
	} else {
		logs.Info("Deployment删除成功",
			"Deployment名称", deploymentName,
			"副本数", replicas) // 显示有意义的调试信息
	}
}

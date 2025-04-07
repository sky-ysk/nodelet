package main

import (
	"bufio"
	"context"
	"fmt"
	"hit.edu/framework/pkg/apimachinery/runtime"
	"hit.edu/framework/pkg/apimachinery/runtime/schema"
	"hit.edu/framework/pkg/apimachinery/runtime/serializer"
	apis "hit.edu/framework/pkg/apis/cores"
	metav1 "hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/client-go/clients"
	"hit.edu/framework/pkg/client-go/rest"
	"hit.edu/framework/pkg/component-base/logs"
	"math/rand"
	"net/http"
	"os"
	"time"
)

func randomSuffix(length int) string {
	rand.Seed(time.Now().UnixNano())
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

// 测试部署-123-12
// 3个group，3个Action，每个Action两个Runtime， 一共6个Runtime，其中第一个group为训练任务（debian1上处理），第二个任务为推理任务（pve2上处理），第三个任务为机器人任务（pve2上处理）
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

	// Task  总共1个Task、3个Group、3个Action、6个runtime
	task1Name := fmt.Sprintf("Task1-%s", randomSuffix(10))  // 第一个Task的Name
	task1ID := fmt.Sprintf("TaskID-1-%s", randomSuffix(10)) // 第一个Task的ID

	// group
	group1_1Name := fmt.Sprintf("Group1-%s", randomSuffix(10)) // 第一个Task下的第一个GroupName
	group1_2Name := fmt.Sprintf("Group2-%s", randomSuffix(10)) // 第一个Task下的第二个GroupName
	group1_3Name := fmt.Sprintf("Group3-%s", randomSuffix(10))
	group1_4Name := fmt.Sprintf("Group4-%s", randomSuffix(10))
	group1_5Name := fmt.Sprintf("Group5-%s", randomSuffix(10))
	group1_6Name := fmt.Sprintf("Group6-%s", randomSuffix(10))
	group1_1ID := fmt.Sprintf("GroupID-1-%s", randomSuffix(10)) // 第一个Task下的第一个GroupID
	group1_2ID := fmt.Sprintf("GroupID-2-%s", randomSuffix(10)) // 第一个Task下的第二个GroupID
	group1_3ID := fmt.Sprintf("GroupID-3-%s", randomSuffix(10)) // 第一个Task下的第三个GroupID
	group1_4ID := fmt.Sprintf("GroupID-4-%s", randomSuffix(10)) // 第一个Task下的第四个GroupID
	group1_5ID := fmt.Sprintf("GroupID-5-%s", randomSuffix(10)) // 第一个Task下的第五个GroupID
	group1_6ID := fmt.Sprintf("GroupID-6-%s", randomSuffix(10)) // 第一个Task下的第五个GroupID
	group1_1Replicas := []int32{0, 0}
	group1_2Replicas := []int32{0, 0}
	group1_3Replicas := []int32{0, 0}
	group1_4Replicas := []int32{0, 0}
	group1_5Replicas := []int32{0, 0}
	group1_6Replicas := []int32{0, 0}

	// action
	action1_1_1Name := fmt.Sprintf("Action1-1-%s", randomSuffix(10)) // 第一个Task下的第一个Group下的第一个ActionName  "cmd_yolo_train_action"
	action1_2_1Name := fmt.Sprintf("Action2-1-%s", randomSuffix(10)) // 第一个Task下的第二个Group下的第一个ActionName
	action1_3_1Name := fmt.Sprintf("Action3-1-%s", randomSuffix(10)) // 第一个Task下的第三个Group下的第一个ActionName
	action1_4_1Name := fmt.Sprintf("Action4-1-%s", randomSuffix(10))
	action1_5_1Name := fmt.Sprintf("Action5-1-%s", randomSuffix(10))
	action1_6_1Name := fmt.Sprintf("Action6-1-%s", randomSuffix(10))
	action1_1_1ID := fmt.Sprintf("ActionID1-1-%s", randomSuffix(10)) // 第一个Task下的第一个Group下的第一个ActionID
	action1_2_1ID := fmt.Sprintf("ActionID2-1-%s", randomSuffix(10)) // 第一个Task下的第二个Group下的第一个ActionID
	action1_3_1ID := fmt.Sprintf("ActionID3-1-%s", randomSuffix(10)) // 第一个Task下的第三个Group下的第一个ActionID
	action1_4_1ID := fmt.Sprintf("ActionID4-1-%s", randomSuffix(10))
	action1_5_1ID := fmt.Sprintf("ActionID5-1-%s", randomSuffix(10))
	action1_6_1ID := fmt.Sprintf("ActionID6-1-%s", randomSuffix(10))

	// runtime
	runtime1_1_1_1Name := fmt.Sprintf("Runtime1-1-1-%s", randomSuffix(10)) // 第一个Task下的第一个Group下的第一个ActionName下的第一个RuntimeName
	runtime1_1_1_2Name := fmt.Sprintf("Runtime1-1-2-%s", randomSuffix(10)) // 第一个Task下的第一个Group下的第一个ActionName下的第二个RuntimeName
	runtime1_2_1_1Name := fmt.Sprintf("Runtime2-1-1-%s", randomSuffix(10)) // 第一个Task下的第二个Group下的第一个ActionName下的第一个RuntimeName
	runtime1_2_1_2Name := fmt.Sprintf("Runtime2-1-2-%s", randomSuffix(10)) // 第一个Task下的第二个Group下的第一个ActionName下的第二个RuntimeName
	runtime1_3_1_1Name := fmt.Sprintf("Runtime3-1-1-%s", randomSuffix(10)) // 第一个Task下的第三个Group下的第一个ActionName下的第一个RuntimeName
	runtime1_3_1_2Name := fmt.Sprintf("Runtime3-1-2-%s", randomSuffix(10)) // 第一个Task下的第三个Group下的第一个ActionName下的第二个RuntimeName
	runtime1_4_1_1Name := fmt.Sprintf("Runtime4-1-1-%s", randomSuffix(10)) // 第一个Task下的第4个Group下的第一个ActionName下的第一个RuntimeName
	runtime1_4_1_2Name := fmt.Sprintf("Runtime4-1-2-%s", randomSuffix(10)) // 第一个Task下的第4个Group下的第一个ActionName下的第二个RuntimeName
	runtime1_5_1_1Name := fmt.Sprintf("Runtime5-1-1-%s", randomSuffix(10)) // 第一个Task下的第5个Group下的第一个ActionName下的第一个RuntimeName
	runtime1_5_1_2Name := fmt.Sprintf("Runtime5-1-2-%s", randomSuffix(10)) // 第一个Task下的第5个Group下的第一个ActionName下的第二个RuntimeName
	runtime1_6_1_1Name := fmt.Sprintf("Runtime6-1-1-%s", randomSuffix(10)) // 第一个Task下的第6个Group下的第一个ActionName下的第一个RuntimeName
	runtime1_6_1_2Name := fmt.Sprintf("Runtime6-1-2-%s", randomSuffix(10)) // 第一个Task下的第6个Group下的第一个ActionName下的第二个RuntimeName

	runtime1_1_1_1ID := fmt.Sprintf("RuntimeID1-1-1-%s", randomSuffix(10)) // 第一个Task下的第一个Group下的第一个ActionName下的第一个RuntimeID
	runtime1_1_1_2ID := fmt.Sprintf("RuntimeID1-1-2-%s", randomSuffix(10)) // 第一个Task下的第一个Group下的第一个ActionName下的第二个RuntimeID
	runtime1_2_1_1ID := fmt.Sprintf("RuntimeID2-1-1-%s", randomSuffix(10)) // 第一个Task下的第二个Group下的第一个ActionName下的第一个RuntimeID
	runtime1_2_1_2ID := fmt.Sprintf("RuntimeID2-1-2-%s", randomSuffix(10)) // 第一个Task下的第二个Group下的第一个ActionName下的第二个RuntimeID
	runtime1_3_1_1ID := fmt.Sprintf("RuntimeID3-1-1-%s", randomSuffix(10)) // 第一个Task下的第三个Group下的第一个ActionName下的第一个RuntimeID
	runtime1_3_1_2ID := fmt.Sprintf("RuntimeID3-1-2-%s", randomSuffix(10)) // 第一个Task下的第三个Group下的第一个ActionName下的第二个RuntimeID
	runtime1_4_1_1ID := fmt.Sprintf("RuntimeID4-1-1-%s", randomSuffix(10)) // 第一个Task下的第4个Group下的第一个ActionName下的第一个RuntimeID
	runtime1_4_1_2ID := fmt.Sprintf("RuntimeID4-1-2-%s", randomSuffix(10)) // 第一个Task下的第4个Group下的第一个ActionName下的第二个RuntimeID
	runtime1_5_1_1ID := fmt.Sprintf("RuntimeID5-1-1-%s", randomSuffix(10)) // 第一个Task下的第5个Group下的第一个ActionName下的第一个RuntimeID
	runtime1_5_1_2ID := fmt.Sprintf("RuntimeID5-1-2-%s", randomSuffix(10)) // 第一个Task下的第5个Group下的第一个ActionName下的第二个RuntimeID
	runtime1_6_1_1ID := fmt.Sprintf("RuntimeID6-1-1-%s", randomSuffix(10)) // 第一个Task下的第6个Group下的第一个ActionName下的第一个RuntimeID
	runtime1_6_1_2ID := fmt.Sprintf("RuntimeID6-1-2-%s", randomSuffix(10)) // 第一个Task下的第6个Group下的第一个ActionName下的第二个RuntimeID

	// runtime是否细粒度控制
	runtime1_1_1_1FineGrainedControl := false
	runtime1_1_1_2FineGrainedControl := false
	runtime1_2_1_1FineGrainedControl := false
	runtime1_2_1_2FineGrainedControl := false
	runtime1_3_1_1FineGrainedControl := false
	runtime1_3_1_2FineGrainedControl := false
	runtime1_4_1_1FineGrainedControl := false
	runtime1_4_1_2FineGrainedControl := false
	runtime1_5_1_1FineGrainedControl := false
	runtime1_5_1_2FineGrainedControl := false
	runtime1_6_1_1FineGrainedControl := false
	runtime1_6_1_2FineGrainedControl := false
	// 统一地规定： Belongs：填的是ID
	//            Parents: 填的也是ID吧--改为Name

	// 程序依赖（requirements.txt）
	ProgramDependencyConditionFormula := apis.ConditionFormula{
		LeftValue: apis.ConditionValue{
			Type:      apis.ResultsData,
			Name:      "ProgramDependency",
			Value:     "0",
			ValueType: "string",
			From:      "/home/public/goprojects/score/test/nodelet/task_exporter/dependency/requirements.txt",
		},
		RightValue: apis.ConditionValue{
			Type:      apis.ConstData,
			Name:      "ProgramDependency",
			Value:     "1",
			ValueType: "string",
			From:      "",
		},
		Signal: apis.Equal,
		Join:   "",
		Result: apis.False,
	}

	runtime1_1_1_1Condition := apis.Conditions{
		Formulas: []apis.ConditionFormula{
			ProgramDependencyConditionFormula,
		},
	}
	runtime1_1_1_2Condition := apis.Conditions{
		Formulas: []apis.ConditionFormula{
			GetNodeDepencyConditionFormula(runtime1_1_1_1Name),
			ProgramDependencyConditionFormula,
		},
	}

	runtime1_2_1_1Condition := apis.Conditions{
		Formulas: []apis.ConditionFormula{
			ProgramDependencyConditionFormula,
		},
	}
	runtime1_2_1_2Condition := apis.Conditions{
		Formulas: []apis.ConditionFormula{
			GetNodeDepencyConditionFormula(runtime1_2_1_1Name),
			ProgramDependencyConditionFormula,
		},
	}

	runtime1_3_1_1Condition := apis.Conditions{
		Formulas: []apis.ConditionFormula{
			ProgramDependencyConditionFormula,
		},
	}
	runtime1_3_1_2Condition := apis.Conditions{
		Formulas: []apis.ConditionFormula{
			GetNodeDepencyConditionFormula(runtime1_3_1_1Name),
			ProgramDependencyConditionFormula,
		},
	}
	runtime1_4_1_1Condition := apis.Conditions{
		Formulas: []apis.ConditionFormula{
			ProgramDependencyConditionFormula,
		},
	}
	runtime1_4_1_2Condition := apis.Conditions{
		Formulas: []apis.ConditionFormula{
			GetNodeDepencyConditionFormula(runtime1_4_1_1Name),
			ProgramDependencyConditionFormula,
		},
	}
	runtime1_5_1_1Condition := apis.Conditions{
		Formulas: []apis.ConditionFormula{
			ProgramDependencyConditionFormula,
		},
	}
	runtime1_5_1_2Condition := apis.Conditions{
		Formulas: []apis.ConditionFormula{
			GetNodeDepencyConditionFormula(runtime1_5_1_1Name),
			ProgramDependencyConditionFormula,
		},
	}
	runtime1_6_1_1Condition := apis.Conditions{
		Formulas: []apis.ConditionFormula{
			ProgramDependencyConditionFormula,
		},
	}
	runtime1_6_1_2Condition := apis.Conditions{
		Formulas: []apis.ConditionFormula{
			GetNodeDepencyConditionFormula(runtime1_6_1_1Name),
			ProgramDependencyConditionFormula,
		},
	}

	group1_1Condition := apis.Conditions{
		Formulas: []apis.ConditionFormula{},
	}

	group1_2Condition := apis.Conditions{
		Formulas: []apis.ConditionFormula{},
	}

	group1_3Condition := apis.Conditions{
		Formulas: []apis.ConditionFormula{},
	}

	group1_4Condition := apis.Conditions{
		Formulas: []apis.ConditionFormula{},
	}

	group1_5Condition := apis.Conditions{
		Formulas: []apis.ConditionFormula{},
	}

	group1_6Condition := apis.Conditions{
		Formulas: []apis.ConditionFormula{
			GetNodeDepencyConditionFormula(group1_2Name),
		},
	}

	g1 := apis.Group{
		ObjectMeta: metav1.ObjectMeta{Name: group1_1Name, Namespace: ""},
		TypeMeta:   metav1.TypeMeta{Kind: "Group", APIVersion: "resources/v1"},
		Spec: apis.GroupSpec{
			ResourceRequirements: []apis.ResourceRequirement{
				apis.ResourceRequirement{
					Name:       "CPU",
					Lowbound:   "2",
					Upperbound: "4",
				},
				apis.ResourceRequirement{
					Name:       "RAM",
					Lowbound:   "2",
					Upperbound: "4",
				},
			},
			Replicas:   group1_1Replicas,
			Name:       group1_1Name,
			Parents:    make([]string, 0),
			Conditions: group1_1Condition,
			Actions: []apis.Action{
				apis.Action{
					ObjectMeta: metav1.ObjectMeta{Name: action1_1_1Name},
					Spec: apis.ActionSpec{
						Name: action1_1_1Name,
						Runtimes: []apis.Runtime{
							apis.Runtime{
								Name:                     runtime1_1_1_1Name,
								Type:                     apis.ByCommand,
								Command:                  []string{"python"},
								Args:                     []string{"/home/public/workspace/heongtong_yolo_linux/train.py"}, //20s
								Parents:                  make([]string, 0),                                                // 加入Parents
								Conditions:               runtime1_1_1_1Condition,
								EnvVar:                   []apis.EnvVar{apis.EnvVar{Name: "", Value: ""}},
								EnableFineGrainedControl: runtime1_1_1_1FineGrainedControl,
							},
							apis.Runtime{
								Name:                     runtime1_1_1_2Name,
								Type:                     apis.ByCommand,
								Command:                  []string{"python"},
								Args:                     []string{"/home/public/workspace/modelfortest/model_torch_absolute/wine.py"}, //8s
								Parents:                  []string{runtime1_1_1_1Name},                                                 // 加入Parents
								Conditions:               runtime1_1_1_2Condition,
								EnvVar:                   []apis.EnvVar{apis.EnvVar{Name: "", Value: ""}},
								EnableFineGrainedControl: runtime1_1_1_2FineGrainedControl,
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
							apis.RuntimeStatus{
								RuntimeID: runtime1_1_1_2ID, // RuntimeID = RuntimeName + ActionID
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
					RuntimeStatus: []apis.RuntimeStatus{
						apis.RuntimeStatus{
							RuntimeID: runtime1_1_1_1ID,
							Phase:     apis.Unknown,
						},
						apis.RuntimeStatus{
							RuntimeID: runtime1_1_1_2ID,
							Phase:     apis.Unknown,
						},
					},
					Phase: apis.Unknown,
				},
			},
			Belongs: apis.IDRef{TaskID: task1ID},
			Phase:   apis.Unknown,
		},
	}
	group1 := &g1

	g2 := apis.Group{
		ObjectMeta: metav1.ObjectMeta{Name: group1_2Name, Namespace: ""},
		TypeMeta:   metav1.TypeMeta{Kind: "Group", APIVersion: "resources/v1"},
		Spec: apis.GroupSpec{
			ResourceRequirements: []apis.ResourceRequirement{
				apis.ResourceRequirement{
					Name:       "CPU",
					Lowbound:   "0",
					Upperbound: "2",
				},
				apis.ResourceRequirement{
					Name:       "RAM",
					Lowbound:   "0",
					Upperbound: "2",
				},
			},
			Replicas:   group1_2Replicas,
			Name:       group1_2Name,
			Parents:    []string{}, // 加入Parents
			Conditions: group1_2Condition,
			Actions: []apis.Action{
				apis.Action{
					ObjectMeta: metav1.ObjectMeta{Name: action1_2_1Name},
					Spec: apis.ActionSpec{
						Name: action1_2_1Name,
						Runtimes: []apis.Runtime{
							apis.Runtime{
								Name:                     runtime1_2_1_1Name,
								Type:                     apis.ByCommand,
								Command:                  []string{"python"},
								Args:                     []string{"/home/public/workspace/modelfortest/model_torch_absolute/wine.py"}, //8s
								Parents:                  make([]string, 0),                                                            // 加入Parents
								Conditions:               runtime1_2_1_1Condition,
								EnvVar:                   []apis.EnvVar{apis.EnvVar{Name: "", Value: ""}},
								EnableFineGrainedControl: runtime1_2_1_1FineGrainedControl,
							},
							apis.Runtime{
								Name:                     runtime1_2_1_2Name,
								Type:                     apis.ByCommand,
								Command:                  []string{"python"},
								Args:                     []string{"/home/public/workspace/modelfortest/model_torch_absolute/wine.py"},
								Parents:                  []string{runtime1_2_1_1Name}, // 加入Parents
								Conditions:               runtime1_2_1_2Condition,
								EnvVar:                   []apis.EnvVar{apis.EnvVar{Name: "", Value: ""}},
								EnableFineGrainedControl: runtime1_2_1_2FineGrainedControl,
							},
						},
					},
					Status: apis.ActionStatus{
						ActionID: action1_2_1ID,
						Phase:    apis.Unknown,
						RuntimeStatus: []apis.RuntimeStatus{
							apis.RuntimeStatus{
								RuntimeID: runtime1_2_1_1ID,
								Phase:     apis.Unknown,
							},
							apis.RuntimeStatus{
								RuntimeID: runtime1_2_1_2ID,
								Phase:     apis.Unknown,
							},
						},
					},
				},
			},
		},
		Status: apis.GroupStatus{
			GroupID: group1_2ID,
			ActionStatus: []apis.ActionStatus{
				apis.ActionStatus{
					ActionID: action1_2_1ID,
					RuntimeStatus: []apis.RuntimeStatus{
						apis.RuntimeStatus{
							RuntimeID: runtime1_2_1_1ID,
							Phase:     apis.Unknown,
						},
						apis.RuntimeStatus{
							RuntimeID: runtime1_2_1_2ID,
							Phase:     apis.Unknown,
						},
					},
					Phase: apis.Unknown,
				},
			},
			Belongs: apis.IDRef{TaskID: task1ID},
			Phase:   apis.Unknown,
		},
	}
	group2 := &g2

	g3 := apis.Group{
		ObjectMeta: metav1.ObjectMeta{Name: group1_3Name, Namespace: ""},
		TypeMeta:   metav1.TypeMeta{Kind: "Group", APIVersion: "resources/v1"},
		Spec: apis.GroupSpec{
			ResourceRequirements: []apis.ResourceRequirement{
				apis.ResourceRequirement{
					Name:       "CPU",
					Lowbound:   "3",
					Upperbound: "6",
				},
				apis.ResourceRequirement{
					Name:       "RAM",
					Lowbound:   "3",
					Upperbound: "6",
				},
			},
			Replicas:   group1_3Replicas,
			Name:       group1_3Name,
			Parents:    []string{},
			Conditions: group1_3Condition,
			Actions: []apis.Action{
				apis.Action{
					ObjectMeta: metav1.ObjectMeta{Name: action1_3_1Name},
					Spec: apis.ActionSpec{
						Name: action1_3_1Name,
						Runtimes: []apis.Runtime{
							apis.Runtime{
								Name:                     runtime1_3_1_1Name,
								Type:                     apis.ByCommand,
								Command:                  []string{"python"},
								Args:                     []string{"/home/public/workspace/modelfortest/model_torch_absolute/mnist.py"}, //24s
								Parents:                  make([]string, 0),                                                             // 加入Parents
								Conditions:               runtime1_3_1_1Condition,
								EnvVar:                   []apis.EnvVar{apis.EnvVar{Name: "", Value: ""}},
								EnableFineGrainedControl: runtime1_3_1_1FineGrainedControl,
							},
							apis.Runtime{
								Name:                     runtime1_3_1_2Name,
								Type:                     apis.ByCommand,
								Command:                  []string{"python"},
								Args:                     []string{"/home/public/workspace/modelfortest/model_torch_absolute/mnist.py"},
								Parents:                  []string{runtime1_3_1_1Name}, // 加入Parents
								Conditions:               runtime1_3_1_2Condition,
								EnvVar:                   []apis.EnvVar{apis.EnvVar{Name: "", Value: ""}},
								EnableFineGrainedControl: runtime1_3_1_2FineGrainedControl,
							},
						},
					},
					Status: apis.ActionStatus{
						ActionID: action1_3_1ID, // ActionID =ActionName + GroupID
						Phase:    apis.Unknown,
						RuntimeStatus: []apis.RuntimeStatus{
							apis.RuntimeStatus{
								RuntimeID: runtime1_3_1_1ID, // RuntimeID = RuntimeName + ActionID
								Phase:     apis.Unknown,
							},
							apis.RuntimeStatus{
								RuntimeID: runtime1_3_1_2ID, // RuntimeID = RuntimeName + ActionID
								Phase:     apis.Unknown,
							},
						},
					},
				},
			},
		},
		Status: apis.GroupStatus{
			GroupID: group1_3ID,
			ActionStatus: []apis.ActionStatus{
				apis.ActionStatus{
					ActionID: action1_3_1ID,
					RuntimeStatus: []apis.RuntimeStatus{
						apis.RuntimeStatus{
							RuntimeID: runtime1_3_1_1ID,
							Phase:     apis.Unknown,
						},
						apis.RuntimeStatus{
							RuntimeID: runtime1_3_1_2ID,
							Phase:     apis.Unknown,
						},
					},
					Phase: apis.Unknown,
				},
			},
			Belongs: apis.IDRef{TaskID: task1ID},
			Phase:   apis.Unknown,
		},
	}
	group3 := &g3
	g4 := apis.Group{
		ObjectMeta: metav1.ObjectMeta{Name: group1_4Name, Namespace: ""},
		TypeMeta:   metav1.TypeMeta{Kind: "Group", APIVersion: "resources/v1"},
		Spec: apis.GroupSpec{
			ResourceRequirements: []apis.ResourceRequirement{
				apis.ResourceRequirement{
					Name:       "CPU",
					Lowbound:   "1",
					Upperbound: "2",
				},
				apis.ResourceRequirement{
					Name:       "RAM",
					Lowbound:   "1",
					Upperbound: "2",
				},
			},
			Replicas:   group1_4Replicas,
			Name:       group1_4Name,
			Parents:    []string{}, // 加入Parents
			Conditions: group1_4Condition,
			Actions: []apis.Action{
				apis.Action{
					ObjectMeta: metav1.ObjectMeta{Name: action1_4_1Name},
					Spec: apis.ActionSpec{
						Name: action1_4_1Name,
						Runtimes: []apis.Runtime{
							apis.Runtime{
								Name:                     runtime1_4_1_1Name,
								Type:                     apis.ByCommand,
								Command:                  []string{"python"},
								Args:                     []string{"/home/public/workspace/modelfortest/model_torch_absolute/Iris.py"}, //11s
								Parents:                  make([]string, 0),                                                            // 加入Parents
								Conditions:               runtime1_4_1_1Condition,
								EnvVar:                   []apis.EnvVar{apis.EnvVar{Name: "", Value: ""}},
								EnableFineGrainedControl: runtime1_4_1_1FineGrainedControl,
							},
							apis.Runtime{
								Name:                     runtime1_4_1_2Name,
								Type:                     apis.ByCommand,
								Command:                  []string{"python"},
								Args:                     []string{"/home/public/workspace/modelfortest/model_torch_absolute/Iris.py"},
								Parents:                  []string{runtime1_4_1_1Name}, // 加入Parents
								Conditions:               runtime1_4_1_2Condition,
								EnvVar:                   []apis.EnvVar{apis.EnvVar{Name: "", Value: ""}},
								EnableFineGrainedControl: runtime1_4_1_2FineGrainedControl,
							},
						},
					},
					Status: apis.ActionStatus{
						ActionID: action1_4_1ID,
						Phase:    apis.Unknown,
						RuntimeStatus: []apis.RuntimeStatus{
							apis.RuntimeStatus{
								RuntimeID: runtime1_4_1_1ID,
								Phase:     apis.Unknown,
							},
							apis.RuntimeStatus{
								RuntimeID: runtime1_4_1_2ID,
								Phase:     apis.Unknown,
							},
						},
					},
				},
			},
		},
		Status: apis.GroupStatus{
			GroupID: group1_4ID,
			ActionStatus: []apis.ActionStatus{
				apis.ActionStatus{
					ActionID: action1_4_1ID,
					RuntimeStatus: []apis.RuntimeStatus{
						apis.RuntimeStatus{
							RuntimeID: runtime1_4_1_1ID,
							Phase:     apis.Unknown,
						},
						apis.RuntimeStatus{
							RuntimeID: runtime1_4_1_2ID,
							Phase:     apis.Unknown,
						},
					},
					Phase: apis.Unknown,
				},
			},
			Belongs: apis.IDRef{TaskID: task1ID},
			Phase:   apis.Unknown,
		},
	}
	group4 := &g4

	g5 := apis.Group{
		ObjectMeta: metav1.ObjectMeta{Name: group1_5Name, Namespace: ""},
		TypeMeta:   metav1.TypeMeta{Kind: "Group", APIVersion: "resources/v1"},
		Spec: apis.GroupSpec{
			ResourceRequirements: []apis.ResourceRequirement{
				apis.ResourceRequirement{
					Name:       "CPU",
					Lowbound:   "1",
					Upperbound: "3",
				},
				apis.ResourceRequirement{
					Name:       "RAM",
					Lowbound:   "1",
					Upperbound: "3",
				},
			},
			Replicas:   group1_5Replicas,
			Name:       group1_5Name,
			Parents:    make([]string, 0),
			Conditions: group1_5Condition,
			Actions: []apis.Action{
				apis.Action{
					ObjectMeta: metav1.ObjectMeta{Name: action1_5_1Name},
					Spec: apis.ActionSpec{
						Name: action1_5_1Name,
						Runtimes: []apis.Runtime{
							apis.Runtime{
								Name:                     runtime1_5_1_1Name,
								Type:                     apis.ByCommand,
								Command:                  []string{"python"},
								Args:                     []string{"/home/public/workspace/modelfortest/model_torch_absolute/Diabetes.py"}, //14s
								Parents:                  make([]string, 0),                                                                // 加入Parents
								Conditions:               runtime1_5_1_1Condition,
								EnvVar:                   []apis.EnvVar{apis.EnvVar{Name: "", Value: ""}},
								EnableFineGrainedControl: runtime1_5_1_1FineGrainedControl,
							},
							apis.Runtime{
								Name:                     runtime1_5_1_2Name,
								Type:                     apis.ByCommand,
								Command:                  []string{"python"},
								Args:                     []string{"/home/public/workspace/modelfortest/model_torch_absolute/Diabetes.py"},
								Parents:                  []string{runtime1_5_1_1Name}, // 加入Parents
								Conditions:               runtime1_5_1_2Condition,
								EnvVar:                   []apis.EnvVar{apis.EnvVar{Name: "", Value: ""}},
								EnableFineGrainedControl: runtime1_5_1_2FineGrainedControl,
							},
						},
					},
					Status: apis.ActionStatus{
						ActionID: action1_5_1ID, // ActionID =ActionName + GroupID
						Phase:    apis.Unknown,
						RuntimeStatus: []apis.RuntimeStatus{
							apis.RuntimeStatus{
								RuntimeID: runtime1_5_1_1ID, // RuntimeID = RuntimeName + ActionID
								Phase:     apis.Unknown,
							},
							apis.RuntimeStatus{
								RuntimeID: runtime1_5_1_2ID, // RuntimeID = RuntimeName + ActionID
								Phase:     apis.Unknown,
							},
						},
					},
				},
			},
		},
		Status: apis.GroupStatus{
			GroupID: group1_5ID,
			ActionStatus: []apis.ActionStatus{
				apis.ActionStatus{
					ActionID: action1_5_1ID,
					RuntimeStatus: []apis.RuntimeStatus{
						apis.RuntimeStatus{
							RuntimeID: runtime1_5_1_1ID,
							Phase:     apis.Unknown,
						},
						apis.RuntimeStatus{
							RuntimeID: runtime1_5_1_2ID,
							Phase:     apis.Unknown,
						},
					},
					Phase: apis.Unknown,
				},
			},
			Belongs: apis.IDRef{TaskID: task1ID},
			Phase:   apis.Unknown,
		},
	}
	group5 := &g5

	g6 := apis.Group{
		ObjectMeta: metav1.ObjectMeta{Name: group1_6Name, Namespace: ""},
		TypeMeta:   metav1.TypeMeta{Kind: "Group", APIVersion: "resources/v1"},
		Spec: apis.GroupSpec{
			ResourceRequirements: []apis.ResourceRequirement{
				apis.ResourceRequirement{
					Name:       "CPU",
					Lowbound:   "1",
					Upperbound: "2",
				},
				apis.ResourceRequirement{
					Name:       "RAM",
					Lowbound:   "1",
					Upperbound: "2",
				},
			},
			Replicas:   group1_6Replicas,
			Name:       group1_6Name,
			Parents:    []string{group1_2Name}, // 加入Parents
			Conditions: group1_6Condition,
			Actions: []apis.Action{
				apis.Action{
					ObjectMeta: metav1.ObjectMeta{Name: action1_6_1Name},
					Spec: apis.ActionSpec{
						Name: action1_6_1Name,
						Runtimes: []apis.Runtime{
							apis.Runtime{
								Name:                     runtime1_6_1_1Name,
								Type:                     apis.ByCommand,
								Command:                  []string{"python"},
								Args:                     []string{"/home/public/workspace/modelfortest/model_torch_absolute/breast_cancer.py"}, //10s
								Parents:                  make([]string, 0),                                                                     // 加入Parents
								Conditions:               runtime1_6_1_1Condition,
								Image:                    "/home/public/workspace/heongtong_yolo_linux/predict.py",
								EnvVar:                   []apis.EnvVar{apis.EnvVar{Name: "", Value: ""}},
								EnableFineGrainedControl: runtime1_6_1_1FineGrainedControl,
							},
							apis.Runtime{
								Name:                     runtime1_6_1_2Name,
								Type:                     apis.ByCommand,
								Command:                  []string{"python"},
								Args:                     []string{"/home/public/workspace/modelfortest/model_torch_absolute/breast_cancer.py"},
								Parents:                  []string{runtime1_6_1_1Name}, // 加入Parents
								Conditions:               runtime1_6_1_2Condition,
								Image:                    "/home/public/workspace/heongtong_yolo_linux/predict.py",
								EnvVar:                   []apis.EnvVar{apis.EnvVar{Name: "", Value: ""}},
								EnableFineGrainedControl: runtime1_6_1_2FineGrainedControl,
							},
						},
					},
					Status: apis.ActionStatus{
						ActionID: action1_6_1ID,
						Phase:    apis.Unknown,
						RuntimeStatus: []apis.RuntimeStatus{
							apis.RuntimeStatus{
								RuntimeID: runtime1_6_1_1ID,
								Phase:     apis.Unknown,
							},
							apis.RuntimeStatus{
								RuntimeID: runtime1_6_1_2ID,
								Phase:     apis.Unknown,
							},
						},
					},
				},
			},
		},
		Status: apis.GroupStatus{
			GroupID: group1_6ID,
			ActionStatus: []apis.ActionStatus{
				apis.ActionStatus{
					ActionID: action1_6_1ID,
					RuntimeStatus: []apis.RuntimeStatus{
						apis.RuntimeStatus{
							RuntimeID: runtime1_6_1_1ID,
							Phase:     apis.Unknown,
						},
						apis.RuntimeStatus{
							RuntimeID: runtime1_6_1_2ID,
							Phase:     apis.Unknown,
						},
					},
					Phase: apis.Unknown,
				},
			},
			Belongs: apis.IDRef{TaskID: task1ID},
			Phase:   apis.Unknown,
		},
	}
	group6 := &g6
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
				g1, g2, g3, g4, g5, g6,
			},
		},
		Status: apis.TaskStatus{
			TaskID: task1ID,
			Phase:  apis.Unknown,
			GroupStatus: []apis.GroupStatus{
				g1.Status, g2.Status, g3.Status, g4.Status, g5.Status, g6.Status,
			},
		},
	}
	// Create一个Task、6个Group
	logs.Infof("creating task and groups")
	_, err = tasksClient.Create(context.TODO(), task, metav1.CreateOptions{})
	_, err1 := groupsClient.Create(context.TODO(), group1, metav1.CreateOptions{})
	_, err2 := groupsClient.Create(context.TODO(), group2, metav1.CreateOptions{})
	_, err3 := groupsClient.Create(context.TODO(), group3, metav1.CreateOptions{})
	_, err4 := groupsClient.Create(context.TODO(), group4, metav1.CreateOptions{})
	_, err5 := groupsClient.Create(context.TODO(), group5, metav1.CreateOptions{})
	_, err6 := groupsClient.Create(context.TODO(), group6, metav1.CreateOptions{})

	if err != nil {
		logs.Errorf("Failed to create task: %v", err)
	}
	if err1 != nil {
		logs.Errorf("Failed to create group1: %v", err)
	}
	if err2 != nil {
		logs.Errorf("Failed to create group2: %v", err)
	}
	if err3 != nil {
		logs.Errorf("Failed to create group3: %v", err)
	}
	if err4 != nil {
		logs.Errorf("Failed to create group4: %v", err)
	}
	if err5 != nil {
		logs.Errorf("Failed to create group5: %v", err)
	}
	if err6 != nil {
		logs.Errorf("Failed to create group6: %v", err)

	}
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

func GetNodeDepencyConditionFormula(parentName string) apis.ConditionFormula {
	return apis.ConditionFormula{
		LeftValue: apis.ConditionValue{
			Type:      apis.ResultsData,
			Name:      "NodeDependency",
			Value:     "0",
			ValueType: "string",
			From:      parentName,
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
		Result: apis.False,
	}
}

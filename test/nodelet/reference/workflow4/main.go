package main

import (
	"bufio"
	"fmt"
	"github.com/google/uuid"
	"hit.edu/framework/pkg/apimachinery/runtime"
	"hit.edu/framework/pkg/apimachinery/runtime/schema"
	"hit.edu/framework/pkg/apimachinery/runtime/serializer"
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/client-go/clients"
	"hit.edu/framework/pkg/client-go/rest"
	"hit.edu/framework/pkg/client-go/util/manager"
	"hit.edu/framework/pkg/component-base/analyzer"
	"hit.edu/framework/pkg/component-base/logs"
	"net/http"
	"os"
	"time"
)

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

	//tasksClient := clientSet.Core().Tasks("test")
	//groupsClient := clientSet.Core().Groups("test")

	// Task  总共1个Task、3个Group、3个Action、6个runtime
	task1Name := "T4" // 第一个Task的Name

	// group
	group1_1Name := "G1" // 第一个Task下的第一个GroupName
	group1_2Name := "G2" // 第一个Task下的第二个GroupName
	group1_3Name := "G3"
	group1_4Name := "G4"
	group1_5Name := "G5"
	group1_6Name := "G6"

	group1_1Replicas := []int32{0, 0}
	group1_2Replicas := []int32{0, 0}
	group1_3Replicas := []int32{0, 0}
	group1_4Replicas := []int32{0, 0}
	group1_5Replicas := []int32{0, 0}
	group1_6Replicas := []int32{0, 0}

	// action
	action1_1_1Name := "A1" // 第一个Task下的第一个Group下的第一个ActionName  "cmd_yolo_train_action"
	action1_2_1Name := "A1" // 第一个Task下的第二个Group下的第一个ActionName
	action1_3_1Name := "A1" // 第一个Task下的第三个Group下的第一个ActionName
	action1_4_1Name := "A1"
	action1_5_1Name := "A1"
	action1_6_1Name := "A1"

	// runtime
	runtime1_1_1_1Name := "R1" // 第一个Task下的第一个Group下的第一个ActionName下的第一个RuntimeName
	runtime1_1_1_2Name := "R2" // 第一个Task下的第一个Group下的第一个ActionName下的第二个RuntimeName
	runtime1_2_1_1Name := "R1" // 第一个Task下的第二个Group下的第一个ActionName下的第一个RuntimeName
	runtime1_2_1_2Name := "R2" // 第一个Task下的第二个Group下的第一个ActionName下的第二个RuntimeName
	runtime1_3_1_1Name := "R1" // 第一个Task下的第三个Group下的第一个ActionName下的第一个RuntimeName
	runtime1_3_1_2Name := "R2" // 第一个Task下的第三个Group下的第一个ActionName下的第二个RuntimeName
	runtime1_4_1_1Name := "R1" // 第一个Task下的第4个Group下的第一个ActionName下的第一个RuntimeName
	runtime1_4_1_2Name := "R2" // 第一个Task下的第4个Group下的第一个ActionName下的第二个RuntimeName
	runtime1_5_1_1Name := "R1" // 第一个Task下的第5个Group下的第一个ActionName下的第一个RuntimeName
	runtime1_5_1_2Name := "R2" // 第一个Task下的第5个Group下的第一个ActionName下的第二个RuntimeName
	runtime1_6_1_1Name := "R1" // 第一个Task下的第6个Group下的第一个ActionName下的第一个RuntimeName
	runtime1_6_1_2Name := "R2" // 第一个Task下的第6个Group下的第一个ActionName下的第二个RuntimeName

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
		ConditionType: apis.ProgramDependency,
		LeftValue: apis.Value{
			Type:      apis.ResultsData,
			Name:      "ProgramDependency",
			Value:     "0",
			ValueType: "string",
			From:      "/home/public/goprojects/reference/test/nodelet/task_exporter/dependency/requirements.txt",
		},
		RightValue: apis.Value{
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
			//GetNodeDepencyConditionFormula(runtime1_1_1_1Name),
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
			//GetNodeDepencyConditionFormula(runtime1_2_1_1Name),
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
			//GetNodeDepencyConditionFormula(runtime1_3_1_1Name),
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
			//GetNodeDepencyConditionFormula(runtime1_4_1_1Name),
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
			//GetNodeDepencyConditionFormula(runtime1_5_1_1Name),
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
			//GetNodeDepencyConditionFormula(runtime1_6_1_1Name),
			ProgramDependencyConditionFormula,
		},
	}

	group1_1Condition := apis.Conditions{
		Formulas: []apis.ConditionFormula{
			//GetNodeDepencyConditionFormula(group1_2Name),
		},
	}

	group1_2Condition := apis.Conditions{
		Formulas: []apis.ConditionFormula{},
	}

	group1_3Condition := apis.Conditions{
		Formulas: []apis.ConditionFormula{},
	}
	group1_4Condition := apis.Conditions{
		Formulas: []apis.ConditionFormula{
			//GetNodeDepencyConditionFormula(group1_2Name),
		},
	}

	group1_5Condition := apis.Conditions{
		Formulas: []apis.ConditionFormula{},
	}

	group1_6Condition := apis.Conditions{
		Formulas: []apis.ConditionFormula{
			//GetNodeDepencyConditionFormula(group1_5Name),
		},
	}

	gs1 := apis.GroupSpec{
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
		Desc: &apis.Description{
			Label: map[string]string{
				"type": "Infer",
			},
		},
		Replicas:   group1_1Replicas,
		Name:       group1_1Name,
		Parents:    []string{group1_2Name},
		Conditions: &group1_1Condition,
		Actions: []apis.ActionSpec{
			apis.ActionSpec{
				Name: action1_1_1Name,
				Runtimes: []apis.RuntimeSpec{
					apis.RuntimeSpec{
						Name:                     runtime1_1_1_1Name,
						Type:                     apis.ByCommand,
						Command:                  []string{"python"},
						Args:                     []string{"/home/public/workspace/heongtong_yolo_linux/train.py"}, //20s
						Parents:                  make([]string, 0),                                                // 加入Parents
						Conditions:               &runtime1_1_1_1Condition,
						EnvVar:                   []apis.EnvVar{apis.EnvVar{Name: "", Value: ""}},
						EnableFineGrainedControl: runtime1_1_1_1FineGrainedControl,
					},
					apis.RuntimeSpec{
						Name:                     runtime1_1_1_2Name,
						Type:                     apis.ByCommand,
						Command:                  []string{"python"},
						Args:                     []string{"/home/public/workspace/heongtong_yolo_linux/predict.py"}, //8s
						Parents:                  []string{runtime1_1_1_1Name},                                       // 加入Parents
						Conditions:               &runtime1_1_1_2Condition,
						EnvVar:                   []apis.EnvVar{apis.EnvVar{Name: "", Value: ""}},
						EnableFineGrainedControl: runtime1_1_1_2FineGrainedControl,
					},
				},
			},
		},
	}

	gs2 := apis.GroupSpec{
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
		Desc: &apis.Description{
			Label: map[string]string{
				"type": "Train",
			},
		},
		Replicas:   group1_2Replicas,
		Name:       group1_2Name,
		Parents:    make([]string, 0), // 加入Parents
		Conditions: &group1_2Condition,
		Actions: []apis.ActionSpec{
			apis.ActionSpec{
				Name: action1_2_1Name,
				Runtimes: []apis.RuntimeSpec{
					apis.RuntimeSpec{
						Name:                     runtime1_2_1_1Name,
						Type:                     apis.ByCommand,
						Command:                  []string{"python"},
						Args:                     []string{"/home/public/workspace/modelfortest/model_torch_absolute/wine.py"}, //8s
						Parents:                  make([]string, 0),                                                            // 加入Parents
						Conditions:               &runtime1_2_1_1Condition,
						EnvVar:                   []apis.EnvVar{apis.EnvVar{Name: "", Value: ""}},
						EnableFineGrainedControl: runtime1_2_1_1FineGrainedControl,
					},
					apis.RuntimeSpec{
						Name:                     runtime1_2_1_2Name,
						Type:                     apis.ByCommand,
						Command:                  []string{"python"},
						Args:                     []string{"/home/public/workspace/modelfortest/model_torch_absolute/wine.py"},
						Parents:                  []string{runtime1_2_1_1Name}, // 加入Parents
						Conditions:               &runtime1_2_1_2Condition,
						EnvVar:                   []apis.EnvVar{apis.EnvVar{Name: "", Value: ""}},
						EnableFineGrainedControl: runtime1_2_1_2FineGrainedControl,
					},
				},
			},
		},
	}

	gs3 := apis.GroupSpec{
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
		Desc: &apis.Description{
			Label: map[string]string{
				"type": "Train",
			},
		},
		Replicas:   group1_3Replicas,
		Name:       group1_3Name,
		Parents:    []string{},
		Conditions: &group1_3Condition,
		Actions: []apis.ActionSpec{
			apis.ActionSpec{
				Name: action1_3_1Name,
				Runtimes: []apis.RuntimeSpec{
					apis.RuntimeSpec{
						Name:                     runtime1_3_1_1Name,
						Type:                     apis.ByCommand,
						Command:                  []string{"python"},
						Args:                     []string{"/home/public/workspace/modelfortest/model_torch_absolute/mnist.py"}, //24s
						Parents:                  make([]string, 0),                                                             // 加入Parents
						Conditions:               &runtime1_3_1_1Condition,
						EnvVar:                   []apis.EnvVar{apis.EnvVar{Name: "", Value: ""}},
						EnableFineGrainedControl: runtime1_3_1_1FineGrainedControl,
					},
					apis.RuntimeSpec{
						Name:                     runtime1_3_1_2Name,
						Type:                     apis.ByCommand,
						Command:                  []string{"python"},
						Args:                     []string{"/home/public/workspace/modelfortest/model_torch_absolute/mnist.py"},
						Parents:                  []string{runtime1_3_1_1Name}, // 加入Parents
						Conditions:               &runtime1_3_1_2Condition,
						EnvVar:                   []apis.EnvVar{apis.EnvVar{Name: "", Value: ""}},
						EnableFineGrainedControl: runtime1_3_1_2FineGrainedControl,
					},
				},
			},
		},
	}

	gs4 := apis.GroupSpec{
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
		Desc: &apis.Description{
			Label: map[string]string{
				"type": "Train",
			},
		},
		Replicas:   group1_4Replicas,
		Name:       group1_4Name,
		Parents:    []string{group1_2Name}, // 加入Parents
		Conditions: &group1_4Condition,
		Actions: []apis.ActionSpec{
			apis.ActionSpec{
				Name: action1_4_1Name,
				Runtimes: []apis.RuntimeSpec{
					apis.RuntimeSpec{
						Name:                     runtime1_4_1_1Name,
						Type:                     apis.ByCommand,
						Command:                  []string{"python"},
						Args:                     []string{"/home/public/workspace/modelfortest/model_torch_absolute/Iris.py"}, //11s
						Parents:                  make([]string, 0),                                                            // 加入Parents
						Conditions:               &runtime1_4_1_1Condition,
						EnvVar:                   []apis.EnvVar{apis.EnvVar{Name: "", Value: ""}},
						EnableFineGrainedControl: runtime1_4_1_1FineGrainedControl,
					},
					apis.RuntimeSpec{
						Name:                     runtime1_4_1_2Name,
						Type:                     apis.ByCommand,
						Command:                  []string{"python"},
						Args:                     []string{"/home/public/workspace/modelfortest/model_torch_absolute/Iris.py"},
						Parents:                  []string{runtime1_4_1_1Name}, // 加入Parents
						Conditions:               &runtime1_4_1_2Condition,
						EnvVar:                   []apis.EnvVar{apis.EnvVar{Name: "", Value: ""}},
						EnableFineGrainedControl: runtime1_4_1_2FineGrainedControl,
					},
				},
			},
		},
	}

	gs5 := apis.GroupSpec{
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
		Desc: &apis.Description{
			Label: map[string]string{
				"type": "Train",
			},
		},
		Replicas:   group1_5Replicas,
		Name:       group1_5Name,
		Parents:    make([]string, 0),
		Conditions: &group1_5Condition,
		Actions: []apis.ActionSpec{
			apis.ActionSpec{
				Name: action1_5_1Name,
				Runtimes: []apis.RuntimeSpec{
					apis.RuntimeSpec{
						Name:                     runtime1_5_1_1Name,
						Type:                     apis.ByCommand,
						Command:                  []string{"python"},
						Args:                     []string{"/home/public/workspace/modelfortest/model_torch_absolute/Diabetes.py"}, //14s
						Parents:                  make([]string, 0),                                                                // 加入Parents
						Conditions:               &runtime1_5_1_1Condition,
						EnvVar:                   []apis.EnvVar{apis.EnvVar{Name: "", Value: ""}},
						EnableFineGrainedControl: runtime1_5_1_1FineGrainedControl,
					},
					apis.RuntimeSpec{
						Name:                     runtime1_5_1_2Name,
						Type:                     apis.ByCommand,
						Command:                  []string{"python"},
						Args:                     []string{"/home/public/workspace/modelfortest/model_torch_absolute/Diabetes.py"},
						Parents:                  []string{runtime1_5_1_1Name}, // 加入Parents
						Conditions:               &runtime1_5_1_2Condition,
						EnvVar:                   []apis.EnvVar{apis.EnvVar{Name: "", Value: ""}},
						EnableFineGrainedControl: runtime1_5_1_2FineGrainedControl,
					},
				},
			},
		},
	}

	gs6 := apis.GroupSpec{
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
		Desc: &apis.Description{
			Label: map[string]string{
				"type": "Train",
			},
		},
		Replicas:   group1_6Replicas,
		Name:       group1_6Name,
		Parents:    []string{group1_5Name}, // 加入Parents
		Conditions: &group1_6Condition,
		Actions: []apis.ActionSpec{
			apis.ActionSpec{

				Name: action1_6_1Name,
				Runtimes: []apis.RuntimeSpec{
					apis.RuntimeSpec{
						Name:                     runtime1_6_1_1Name,
						Type:                     apis.ByCommand,
						Command:                  []string{"python"},
						Args:                     []string{"/home/public/workspace/modelfortest/model_torch_absolute/breast_cancer.py"}, //10s
						Parents:                  make([]string, 0),                                                                     // 加入Parents
						Conditions:               &runtime1_6_1_1Condition,
						Image:                    "/home/public/workspace/heongtong_yolo_linux/predict.py",
						EnvVar:                   []apis.EnvVar{apis.EnvVar{Name: "", Value: ""}},
						EnableFineGrainedControl: runtime1_6_1_1FineGrainedControl,
					},
					apis.RuntimeSpec{
						Name:                     runtime1_6_1_2Name,
						Type:                     apis.ByCommand,
						Command:                  []string{"python"},
						Args:                     []string{"/home/public/workspace/modelfortest/model_torch_absolute/breast_cancer.py"},
						Parents:                  []string{runtime1_6_1_1Name}, // 加入Parents
						Conditions:               &runtime1_6_1_2Condition,
						Image:                    "/home/public/workspace/heongtong_yolo_linux/predict.py",
						EnvVar:                   []apis.EnvVar{apis.EnvVar{Name: "", Value: ""}},
						EnableFineGrainedControl: runtime1_6_1_2FineGrainedControl,
					},
				},
			},
		},
	}

	ts := apis.TaskSpec{
		Name: task1Name,
		Groups: []apis.GroupSpec{
			gs1, gs2, gs3, gs4, gs5, gs6,
		},
	}
	// 生成UUID
	u := uuid.Must(uuid.NewV7())
	m := manager.NewManager(clientSet)
	task, err := m.CreateTask(ts, nil, "test", u.String(), "")
	if err != nil {
		panic(err)
	}
	str, err := analyzer.SerializeToJson(task)
	if err != nil {
		return
	}
	fmt.Println(str)
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

//func GetNodeDepencyConditionFormula(parentName string) apis.ConditionFormula {
//	return apis.ConditionFormula{
//		ConditionType: apis.NodeDependency,
//		LeftValue: apis.Value{
//			Type:      apis.ResultsData,
//			Name:      "NodeDependency",
//			Value:     "0",
//			ValueType: "string",
//			From:      parentName,
//		},
//		RightValue: apis.Value{
//			Type:      apis.ConstData,
//			Name:      "NodeDependency",
//			Value:     "1",
//			ValueType: "string",
//			From:      "",
//		},
//		Signal: apis.Equal,
//		Join:   "",
//		Result: apis.False,
//	}
//}

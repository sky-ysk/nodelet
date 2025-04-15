package main

import (
	"bufio"
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

// 测试部署
// 2个group，2个Action，每个Action1个Runtime， 一共2个Runtime
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

	_ = clientSet.Core().Tasks("test")
	//groupsClient := clientSet.Core().Groups("test")
	//actionsClient := clientSet.Core().Actions("test")
	//eventsClient := clientSet.Core().Events("test")

	// Task  总共1个Task、3个Group、3个Action、6个runtime
	task1Name := fmt.Sprintf("Task1-%s", randomSuffix(5))  // 第一个Task的Name
	task1ID := fmt.Sprintf("TaskID-1-%s", randomSuffix(5)) // 第一个Task的ID

	// group
	group1_1Name := fmt.Sprintf("Group1-%s", randomSuffix(5)) // 第一个Task下的第一个GroupName
	group1_2Name := fmt.Sprintf("Group2-%s", randomSuffix(5)) // 第一个Task下的第二个GroupName

	group1_1ID := fmt.Sprintf("GroupID-1-%s", randomSuffix(5)) // 第一个Task下的第一个GroupID
	group1_2ID := fmt.Sprintf("GroupID-2-%s", randomSuffix(5)) // 第一个Task下的第二个GroupID

	group1_1Replicas := []int32{0, 0}
	group1_2Replicas := []int32{0, 0}

	// action
	action1_1_1Name := "Action1-1" // 第一个Task下的第一个Group下的第一个ActionName  "cmd_yolo_train_action"
	action1_2_1Name := "Action2-1" // 第一个Task下的第二个Group下的第一个ActionName

	action1_1_1ID := "ActionID1-1" // 第一个Task下的第一个Group下的第一个ActionID
	action1_2_1ID := "ActionID2-1" // 第一个Task下的第二个Group下的第一个ActionID

	// runtime
	runtime1_1_1_1Name := "Runtime1-1-1" // 第一个Task下的第一个Group下的第一个ActionName下的第一个RuntimeName
	runtime1_2_1_1Name := "Runtime2-1-1" // 第一个Task下的第二个Group下的第一个ActionName下的第一个RuntimeName

	runtime1_1_1_1ID := "RuntimeID1-1-1" // 第一个Task下的第一个Group下的第一个ActionName下的第一个RuntimeID
	runtime1_2_1_1ID := "RuntimeID2-1-1" // 第一个Task下的第二个Group下的第一个ActionName下的第一个RuntimeID

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

	runtime1_2_1_1Condition := apis.Conditions{
		Formulas: []apis.ConditionFormula{
			ProgramDependencyConditionFormula,
		},
	}

	group1_1Condition := apis.Conditions{
		Formulas: []apis.ConditionFormula{},
	}

	group1_2Condition := apis.Conditions{
		Formulas: []apis.ConditionFormula{
			GetNodeDepencyConditionFormula(group1_1Name),
		},
	}

	g1 := apis.Group{
		ObjectMeta: metav1.ObjectMeta{Name: group1_1Name, Namespace: ""},
		TypeMeta:   metav1.TypeMeta{Kind: "Group", APIVersion: "resources/v1"},
		Spec: apis.GroupSpec{
			ResourceRequirements: []apis.ResourceRequirement{
				apis.ResourceRequirement{
					Name:       "CPU",
					Lowbound:   "3",
					Upperbound: "4",
				},
				apis.ResourceRequirement{
					Name:       "RAM",
					Lowbound:   "3",
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
								Name:       runtime1_1_1_1Name,
								Type:       apis.ByCommand,
								Command:    []string{"python"},
								Args:       []string{"/home/public/workspace/modelfortest/model_torch_absolute/wine.py"},
								Parents:    make([]string, 0), // 加入Parents
								Conditions: runtime1_1_1_1Condition,
								EnvVar:     []apis.EnvVar{apis.EnvVar{Name: "", Value: ""}},
							},
						},
					},
					Status: apis.ActionStatus{
						ActionID: action1_1_1ID,
						Phase:    apis.Unknown,
						RuntimeStatus: []apis.RuntimeStatus{
							apis.RuntimeStatus{
								RuntimeID: runtime1_1_1_1ID,
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
					},
					Phase: apis.Unknown,
				},
			},
			Belongs: apis.IDRef{TaskID: task1ID},
			Phase:   apis.Unknown,
		},
	}

	g2 := apis.Group{
		ObjectMeta: metav1.ObjectMeta{Name: group1_2Name, Namespace: ""},
		TypeMeta:   metav1.TypeMeta{Kind: "Group", APIVersion: "resources/v1"},
		Spec: apis.GroupSpec{
			ResourceRequirements: []apis.ResourceRequirement{
				apis.ResourceRequirement{
					Name:       "CPU",
					Lowbound:   "3",
					Upperbound: "4",
				},
				apis.ResourceRequirement{
					Name:       "RAM",
					Lowbound:   "3",
					Upperbound: "4",
				},
			},
			Replicas:   group1_2Replicas,
			Name:       group1_2Name,
			Parents:    []string{group1_1Name}, // 加入Parents
			Conditions: group1_2Condition,
			Actions: []apis.Action{
				apis.Action{
					ObjectMeta: metav1.ObjectMeta{Name: action1_2_1Name},
					Spec: apis.ActionSpec{
						Name: action1_2_1Name,
						Runtimes: []apis.Runtime{
							apis.Runtime{
								Name:       runtime1_2_1_1Name,
								Type:       apis.ByCommand,
								Command:    []string{"python"},
								Args:       []string{"/home/public/workspace/modelfortest/model_torch_absolute/wine.py"},
								Parents:    make([]string, 0), // 加入Parents
								Conditions: runtime1_2_1_1Condition,
								EnvVar:     []apis.EnvVar{apis.EnvVar{Name: "", Value: ""}},
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
					},
					Phase: apis.Unknown,
				},
			},
			Belongs: apis.IDRef{TaskID: task1ID},
			Phase:   apis.Unknown,
		},
	}
	//group2 := &g2

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
				g1, g2,
			},
		},
		Status: apis.TaskStatus{
			TaskID: task1ID,
			Phase:  apis.Unknown,
			GroupStatus: []apis.GroupStatus{
				g1.Status, g2.Status, //g3.Status, g4.Status, g5.Status, g6.Status,
			},
		},
	}
	logs.Info("task:%v", task)

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

func randomSuffix(length int) string {
	rand.Seed(time.Now().UnixNano())
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

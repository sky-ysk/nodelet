package analyzer

import (
	"fmt"
	apis "hit.edu/framework/pkg/apis/cores"
	metav1 "hit.edu/framework/pkg/apis/meta"
	"math/rand"
	"testing"
	"time"
)

func TestSerializeToJson(t *testing.T) {
	// Task  总共1个Task、3个Group、3个Action、6个runtime
	task1Name := fmt.Sprintf("Task1-%s", randomSuffix(5))  // 第一个Task的Name
	task1ID := fmt.Sprintf("TaskID-1-%s", randomSuffix(5)) // 第一个Task的ID

	// group
	group1_1Name := fmt.Sprintf("Group1-%s", randomSuffix(5)) // 第一个Task下的第一个GroupName
	//group1_2Name := fmt.Sprintf("Group2-%s", randomSuffix(5)) // 第一个Task下的第二个GroupName
	//group1_3Name := fmt.Sprintf("Group3-%s", randomSuffix(5))
	//group1_4Name := fmt.Sprintf("Group4-%s", randomSuffix(5))
	//group1_5Name := fmt.Sprintf("Group5-%s", randomSuffix(5))
	//group1_6Name := fmt.Sprintf("Group6-%s", randomSuffix(5))
	group1_1ID := fmt.Sprintf("GroupID-1-%s", randomSuffix(5)) // 第一个Task下的第一个GroupID
	//group1_2ID := fmt.Sprintf("GroupID-2-%s", randomSuffix(5)) // 第一个Task下的第二个GroupID
	//group1_3ID := fmt.Sprintf("GroupID-3-%s", randomSuffix(5)) // 第一个Task下的第三个GroupID
	//group1_4ID := fmt.Sprintf("GroupID-4-%s", randomSuffix(5)) // 第一个Task下的第四个GroupID
	//group1_5ID := fmt.Sprintf("GroupID-5-%s", randomSuffix(5)) // 第一个Task下的第五个GroupID
	//group1_6ID := fmt.Sprintf("GroupID-6-%s", randomSuffix(5)) // 第一个Task下的第五个GroupID
	group1_1Replicas := []int32{0, 0}
	//group1_2Replicas := []int32{0, 0}
	//group1_3Replicas := []int32{0, 0}
	//group1_4Replicas := []int32{0, 0}
	//group1_5Replicas := []int32{0, 0}
	//group1_6Replicas := []int32{0, 0}

	// action
	action1_1_1Name := "Action1-1" // 第一个Task下的第一个Group下的第一个ActionName  "cmd_yolo_train_action"
	//action1_2_1Name := "Action2-1" // 第一个Task下的第二个Group下的第一个ActionName
	//action1_3_1Name := "Action3-1" // 第一个Task下的第三个Group下的第一个ActionName
	//action1_4_1Name := "Action4-1"
	//action1_5_1Name := "Action5-1"
	//action1_6_1Name := "Action6-1"
	action1_1_1ID := "ActionID1-1" // 第一个Task下的第一个Group下的第一个ActionID
	//action1_2_1ID := "ActionID2-1" // 第一个Task下的第二个Group下的第一个ActionID
	//action1_3_1ID := "ActionID3-1" // 第一个Task下的第三个Group下的第一个ActionID
	//action1_4_1ID := "ActionID4-1"
	//action1_5_1ID := "ActionID5-1"
	//action1_6_1ID := "ActionID6-1"

	// runtime
	runtime1_1_1_1Name := "Runtime1-1-1" // 第一个Task下的第一个Group下的第一个ActionName下的第一个RuntimeName
	//runtime1_1_1_2Name := "Runtime1-1-2" // 第一个Task下的第一个Group下的第一个ActionName下的第二个RuntimeName
	//runtime1_2_1_1Name := "Runtime2-1-1" // 第一个Task下的第二个Group下的第一个ActionName下的第一个RuntimeName
	//runtime1_2_1_2Name := "Runtime2-1-2" // 第一个Task下的第二个Group下的第一个ActionName下的第二个RuntimeName
	//runtime1_3_1_1Name := "Runtime3-1-1" // 第一个Task下的第三个Group下的第一个ActionName下的第一个RuntimeName
	//runtime1_3_1_2Name := "Runtime3-1-2" // 第一个Task下的第三个Group下的第一个ActionName下的第二个RuntimeName
	//runtime1_4_1_1Name := "Runtime4-1-1" // 第一个Task下的第4个Group下的第一个ActionName下的第一个RuntimeName
	//runtime1_4_1_2Name := "Runtime4-1-2" // 第一个Task下的第4个Group下的第一个ActionName下的第二个RuntimeName
	//runtime1_5_1_1Name := "Runtime5-1-1" // 第一个Task下的第5个Group下的第一个ActionName下的第一个RuntimeName
	//runtime1_5_1_2Name := "Runtime5-1-2" // 第一个Task下的第5个Group下的第一个ActionName下的第二个RuntimeName
	//runtime1_6_1_1Name := "Runtime6-1-1" // 第一个Task下的第6个Group下的第一个ActionName下的第一个RuntimeName
	//runtime1_6_1_2Name := "Runtime6-1-2" // 第一个Task下的第6个Group下的第一个ActionName下的第二个RuntimeName

	runtime1_1_1_1ID := "RuntimeID1-1-1" // 第一个Task下的第一个Group下的第一个ActionName下的第一个RuntimeID
	//runtime1_1_1_2ID := "RuntimeID1-1-2" // 第一个Task下的第一个Group下的第一个ActionName下的第二个RuntimeID
	//runtime1_2_1_1ID := "RuntimeID2-1-1" // 第一个Task下的第二个Group下的第一个ActionName下的第一个RuntimeID
	//runtime1_2_1_2ID := "RuntimeID2-1-2" // 第一个Task下的第二个Group下的第一个ActionName下的第二个RuntimeID
	//runtime1_3_1_1ID := "RuntimeID3-1-1" // 第一个Task下的第三个Group下的第一个ActionName下的第一个RuntimeID
	//runtime1_3_1_2ID := "RuntimeID3-1-2" // 第一个Task下的第三个Group下的第一个ActionName下的第二个RuntimeID
	//runtime1_4_1_1ID := "RuntimeID4-1-1" // 第一个Task下的第4个Group下的第一个ActionName下的第一个RuntimeID
	//runtime1_4_1_2ID := "RuntimeID4-1-2" // 第一个Task下的第4个Group下的第一个ActionName下的第二个RuntimeID
	//runtime1_5_1_1ID := "RuntimeID5-1-1" // 第一个Task下的第5个Group下的第一个ActionName下的第一个RuntimeID
	//runtime1_5_1_2ID := "RuntimeID5-1-2" // 第一个Task下的第5个Group下的第一个ActionName下的第二个RuntimeID
	//runtime1_6_1_1ID := "RuntimeID6-1-1" // 第一个Task下的第6个Group下的第一个ActionName下的第一个RuntimeID
	//runtime1_6_1_2ID := "RuntimeID6-1-2" // 第一个Task下的第6个Group下的第一个ActionName下的第二个RuntimeID

	// runtime是否细粒度控制
	runtime1_1_1_1FineGrainedControl := false
	//runtime1_1_1_2FineGrainedControl := false
	//runtime1_2_1_1FineGrainedControl := false
	//runtime1_2_1_2FineGrainedControl := false
	//runtime1_3_1_1FineGrainedControl := false
	//runtime1_3_1_2FineGrainedControl := false
	//runtime1_4_1_1FineGrainedControl := false
	//runtime1_4_1_2FineGrainedControl := false
	//runtime1_5_1_1FineGrainedControl := false
	//runtime1_5_1_2FineGrainedControl := false
	//runtime1_6_1_1FineGrainedControl := false
	//runtime1_6_1_2FineGrainedControl := false
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

	group1_1Condition := apis.Conditions{
		Formulas: []apis.ConditionFormula{},
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
								Name:                     runtime1_1_1_1Name,
								Type:                     apis.ByCommand,
								Command:                  []string{"python"},
								Args:                     []string{"/home/public/workspace/modelfortest/model_torch_absolute/wine.py"},
								Parents:                  make([]string, 0), // 加入Parents
								Conditions:               runtime1_1_1_1Condition,
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
	//group1 := &g1

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
				g1.Status, //g3.Status, g4.Status, g5.Status, g6.Status,
			},
		},
	}

	result, err := SerializeToJson(task)
	if err != nil {
		t.Errorf("Failed to serialize to json, error is \n %v", err)
	} else {
		t.Logf("Success to serialize to json, result is \n %v", result)
	}
}

func TestSerializeToYaml(t *testing.T) {
	node := apis.Node{
		Spec: apis.NodeSpec{
			NodeName: "TestNode",
		},
		Status: apis.NodeStatus{},
	}
	result, err := SerializeToYaml(node)
	if err != nil {
		t.Errorf("Failed to serialize to yaml, error is %v", err)
	} else {
		t.Logf("Success to serialize to yaml, result is \n%v", result)
	}
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

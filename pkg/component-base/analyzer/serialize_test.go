package analyzer

import (
	apis "hit.edu/framework/pkg/apis/cores"
	metav1 "hit.edu/framework/pkg/apis/meta"
	"testing"
)

func TestSerializeToJson(t *testing.T) {
	//node := apis.Node{
	//	Spec: apis.NodeSpec{
	//		NodeName: "TestNode",
	//	},
	//	Status: apis.NodeStatus{},
	//}
	g1 := apis.Group{
		ObjectMeta: metav1.ObjectMeta{Name: "TrainGroup", Namespace: ""},
		TypeMeta:   metav1.TypeMeta{Kind: "Group", APIVersion: "resources/v1"},
		Spec: apis.GroupSpec{
			Name:    "TrainGroup",
			Parents: make([]string, 0),
			Actions: []apis.Action{
				apis.Action{
					ObjectMeta: metav1.ObjectMeta{Name: "cmd_yolo_train_action"},
					Spec: apis.ActionSpec{
						Name: "cmd_yolo_train_action",
						Runtimes: []apis.Runtime{
							apis.Runtime{
								Name:    "CMD",
								Type:    apis.ByCommand,
								Command: []string{"/home/public/anaconda3/envs/yolo/bin/python"},
								Args:    []string{"/home/public/workspace/heongtong_yolo_linux/train.py"},
							},
						},
					},
					Status: apis.ActionStatus{
						ActionID: "cmd_yolo_train_action:TrainGroup:test-task", // ActionID =ActionName + GroupID
						Phase:    apis.Unknown,
						RuntimeStatus: []apis.RuntimeStatus{
							apis.RuntimeStatus{
								RuntimeID: "CMD:cmd_yolo_train_action:TrainGroup:test-task", // RuntimeID = RuntimeName + ActionID
								Phase:     apis.Unknown,
							},
						},
					},
				},
			},
		},
		Status: apis.GroupStatus{
			GroupID: "TrainGroup:test-task",
			ActionStatus: []apis.ActionStatus{
				apis.ActionStatus{
					ActionID: "cmd_yolo_train_action:TrainGroup:test-task",
					RuntimeStatus: []apis.RuntimeStatus{
						apis.RuntimeStatus{
							RuntimeID: "CMD:cmd_yolo_train_action:TrainGroup:test-task",
							Phase:     apis.Unknown,
						},
					},
					Phase: apis.Unknown,
				},
			},
			Belongs: apis.IDRef{TaskID: "test-task"},
		},
	}
	//group1 := &g1
	g2 := apis.Group{
		ObjectMeta: metav1.ObjectMeta{Name: "ReasonGroup", Namespace: ""},
		TypeMeta:   metav1.TypeMeta{Kind: "Group", APIVersion: "resources/v1"},
		Spec: apis.GroupSpec{
			Name:    "ReasonGroup",
			Parents: []string{"TrainGroup"}, // 加入Parents
			Actions: []apis.Action{
				apis.Action{
					ObjectMeta: metav1.ObjectMeta{Name: "cmd_yolo_predict_action"},
					Spec: apis.ActionSpec{
						Name: "cmd_yolo_predict_action",
						Runtimes: []apis.Runtime{
							apis.Runtime{
								Name:    "ABC",
								Type:    apis.ByCommand,
								Command: []string{"/home/public/anaconda3/envs/yolo/bin/python"},
								Args:    []string{"/home/public/workspace/heongtong_yolo_linux/predict.py"},
							},
						},
					},
					Status: apis.ActionStatus{
						ActionID: "cmd_yolo_predict_action:ReasonGroup:test-task",
						Phase:    apis.Unknown,
						RuntimeStatus: []apis.RuntimeStatus{
							apis.RuntimeStatus{
								RuntimeID: "ABC:cmd_yolo_predict_action:ReasonGroup:test-task",
								Phase:     apis.Unknown,
							},
						},
					},
				},
			},
		},
		Status: apis.GroupStatus{
			GroupID: "ReasonGroup:test-task",
			ActionStatus: []apis.ActionStatus{
				apis.ActionStatus{
					ActionID: "cmd_yolo_predict_action:ReasonGroup:test-task",
					RuntimeStatus: []apis.RuntimeStatus{
						apis.RuntimeStatus{
							RuntimeID: "ABC:cmd_yolo_predict_action:ReasonGroup:test-task",
							Phase:     apis.Unknown,
						},
					},
					Phase: apis.Unknown,
				},
			},
			Belongs: apis.IDRef{TaskID: "test-task"},
		},
	}
	//group2 := &g2
	task := apis.Task{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "TrainInferTask",
			Namespace: "",
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Task",
			APIVersion: "resources/v1",
		},
		Spec: apis.TaskSpec{
			Name: "demo-task",
			Groups: []apis.Group{
				g1, g2,
			},
		},
		Status: apis.TaskStatus{
			TaskID: "test-task",
			Phase:  apis.Unknown,
			GroupStatus: []apis.GroupStatus{
				apis.GroupStatus{
					GroupID: "TrainGroup:test-task",
					ActionStatus: []apis.ActionStatus{
						apis.ActionStatus{
							ActionID: "cmd_yolo_train_action:TrainGroup:test-task",
							RuntimeStatus: []apis.RuntimeStatus{
								apis.RuntimeStatus{
									RuntimeID: "CMD:cmd_yolo_train_action:TrainGroup:test-task",
									Phase:     apis.Unknown,
								},
							},
							Phase: apis.Unknown,
						},
					},
					Belongs: apis.IDRef{TaskID: "test-task"},
				},
				apis.GroupStatus{
					GroupID: "ReasonGroup:test-task",
					ActionStatus: []apis.ActionStatus{
						apis.ActionStatus{
							ActionID: "cmd_yolo_predict_action:ReasonGroup:test-task",
							RuntimeStatus: []apis.RuntimeStatus{
								apis.RuntimeStatus{
									RuntimeID: "ABC:cmd_yolo_predict_action:ReasonGroup:test-task",
									Phase:     apis.Unknown,
								},
							},
							Phase: apis.Unknown,
						},
					},
					Belongs: apis.IDRef{TaskID: "test-task"},
				},
			},
		},
	}

	//node := apis.ContainerImage{
	//	Names:     []string{"container", "image", "sss"},
	//	SizeBytes: 123,
	//}

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

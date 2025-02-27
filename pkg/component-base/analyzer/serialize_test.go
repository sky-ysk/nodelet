package analyzer

import (
	apis "hit.edu/framework/pkg/apis/cores"
	metav1 "hit.edu/framework/pkg/apis/meta"
	"testing"
)

func TestSerializeToJson(t *testing.T) {
	// Task  总共1个Task、3个Group、3个Action、12个runtime
	task1Name := "TrainInferTask-1" // 第一个Task的Name
	task1ID := "TrainInferTaskID-1" // 第一个Task的ID

	// group
	group1_1Name := "TrainGroup-1"            // 第一个Task下的第一个GroupName
	group1_2Name := "ReasonGroup-2"           // 第一个Task下的第二个GroupName
	group1_3Name := "RobotDestinationGroup-3" // 第一个Task下的第三个GroupName  "RobotDestinationGroup"
	group1_1ID := "GroupID-1"                 // 第一个Task下的第一个GroupID
	group1_2ID := "GroupID-2"                 // 第一个Task下的第二个GroupID
	group1_3ID := "GroupID-3"                 // 第一个Task下的第三个GroupID

	// action
	action1_1_1Name := "Action1-1" // 第一个Task下的第一个Group下的第一个ActionName  "cmd_yolo_train_action"
	action1_1_2Name := "Action1-2" // 第一个Task下的第一个Group下的第二个ActionName  "cmd_yolo_train_action"
	action1_2_1Name := "Action2-1" // 第一个Task下的第二个Group下的第一个ActionName
	action1_2_2Name := "Action2-2" // 第一个Task下的第二个Group下的第二个ActionName
	action1_3_1Name := "Action3-1" // 第一个Task下的第三个Group下的第一个ActionName
	action1_3_2Name := "Action3-2" // 第一个Task下的第三个Group下的第二个ActionName
	action1_1_1ID := "ActionID1-1" // 第一个Task下的第一个Group下的第一个ActionID
	action1_1_2ID := "ActionID1-2" // 第一个Task下的第一个Group下的第二个ActionID
	action1_2_1ID := "ActionID2-1" // 第一个Task下的第二个Group下的第一个ActionID
	action1_2_2ID := "ActionID2-2" // 第一个Task下的第二个Group下的第二个ActionID
	action1_3_1ID := "ActionID3-1" // 第一个Task下的第三个Group下的第一个ActionID
	action1_3_2ID := "ActionID3-2" // 第一个Task下的第三个Group下的第二个ActionID

	// runtime
	runtime1_1_1_1Name := "Runtime1-1-1" // 第一个Task下的第一个Group下的第一个ActionName下的第一个RuntimeName
	runtime1_1_1_2Name := "Runtime1-1-2" // 第一个Task下的第一个Group下的第一个ActionName下的第二个RuntimeName
	runtime1_1_2_1Name := "Runtime1-2-1" // 第一个Task下的第一个Group下的第二个ActionName下的第一个RuntimeName
	runtime1_1_2_2Name := "Runtime1-2-2" // 第一个Task下的第一个Group下的第二个ActionName下的第二个RuntimeName
	runtime1_2_1_1Name := "Runtime2-1-1" // 第一个Task下的第二个Group下的第一个ActionName下的第一个RuntimeName
	runtime1_2_1_2Name := "Runtime2-1-2" // 第一个Task下的第二个Group下的第一个ActionName下的第二个RuntimeName
	runtime1_2_2_1Name := "Runtime2-2-1" // 第一个Task下的第二个Group下的第二个ActionName下的第一个RuntimeName
	runtime1_2_2_2Name := "Runtime2-2-2" // 第一个Task下的第二个Group下的第二个ActionName下的第二个RuntimeName
	runtime1_3_1_1Name := "Runtime3-1-1" // 第一个Task下的第三个Group下的第一个ActionName下的第一个RuntimeName
	runtime1_3_1_2Name := "Runtime3-1-2" // 第一个Task下的第三个Group下的第一个ActionName下的第二个RuntimeName
	runtime1_3_2_1Name := "Runtime3-2-1" // 第一个Task下的第三个Group下的第二个ActionName下的第一个RuntimeName
	runtime1_3_2_2Name := "Runtime3-2-2" // 第一个Task下的第三个Group下的第二个ActionName下的第二个RuntimeName

	runtime1_1_1_1ID := "RuntimeID1-1-1" // 第一个Task下的第一个Group下的第一个ActionName下的第一个RuntimeID
	runtime1_1_1_2ID := "RuntimeID1-1-2" // 第一个Task下的第一个Group下的第一个ActionName下的第二个RuntimeID
	runtime1_1_2_1ID := "RuntimeID1-2-1" // 第一个Task下的第一个Group下的第二个ActionName下的第一个RuntimeID
	runtime1_1_2_2ID := "RuntimeID1-2-2" // 第一个Task下的第一个Group下的第二个ActionName下的第二个RuntimeID
	runtime1_2_1_1ID := "RuntimeID2-1-1" // 第一个Task下的第二个Group下的第一个ActionName下的第一个RuntimeID
	runtime1_2_1_2ID := "RuntimeID2-1-2" // 第一个Task下的第二个Group下的第一个ActionName下的第二个RuntimeID
	runtime1_2_2_1ID := "RuntimeID2-2-1" // 第一个Task下的第二个Group下的第二个ActionName下的第一个RuntimeID
	runtime1_2_2_2ID := "RuntimeID2-2-2" // 第一个Task下的第二个Group下的第二个ActionName下的第二个RuntimeID
	runtime1_3_1_1ID := "RuntimeID3-1-1" // 第一个Task下的第三个Group下的第一个ActionName下的第一个RuntimeID
	runtime1_3_1_2ID := "RuntimeID3-1-2" // 第一个Task下的第三个Group下的第一个ActionName下的第二个RuntimeID
	runtime1_3_2_1ID := "RuntimeID3-2-1" // 第一个Task下的第三个Group下的第二个ActionName下的第一个RuntimeID
	runtime1_3_2_2ID := "RuntimeID3-2-2" // 第一个Task下的第三个Group下的第二个ActionName下的第二个RuntimeID

	// 统一地规定： Belongs：填的是ID
	//            Parents: 填的也是ID吧--该为Name

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
						Name:    action1_1_1Name,
						Parents: make([]string, 0),
						Runtimes: []apis.Runtime{
							apis.Runtime{
								Name:    runtime1_1_1_1Name,
								Type:    apis.ByCommand,
								Command: []string{"/home/public/anaconda3/envs/yolo/bin/python"},
								Args:    []string{"/home/public/workspace/heongtong_yolo_linux/train.py"},
								Parents: make([]string, 0), // 加入Parents
								Image:   "/home/public/workspace/heongtong_yolo_linux/train.py",
								EnvVar:  []apis.EnvVar{apis.EnvVar{Name: "", Value: ""}},
							},
							apis.Runtime{
								Name:    runtime1_1_1_2Name,
								Type:    apis.ByCommand,
								Command: []string{"/home/public/anaconda3/envs/yolo/bin/python"},
								Args:    []string{"/home/public/workspace/heongtong_yolo_linux/train.py"},
								Parents: []string{runtime1_1_1_1Name}, // 加入Parents
								Image:   "/home/public/workspace/heongtong_yolo_linux/train.py",
								EnvVar:  []apis.EnvVar{apis.EnvVar{Name: "", Value: ""}},
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
				apis.Action{
					ObjectMeta: metav1.ObjectMeta{Name: action1_1_2Name},
					Spec: apis.ActionSpec{
						Name:    action1_1_2Name,
						Parents: []string{action1_1_1Name},
						Runtimes: []apis.Runtime{
							apis.Runtime{
								Name:    runtime1_1_2_1Name,
								Type:    apis.ByCommand,
								Command: []string{"/home/public/anaconda3/envs/yolo/bin/python"},
								Args:    []string{"/home/public/workspace/heongtong_yolo_linux/train.py"},
								Parents: make([]string, 0), // 加入Parents
								Image:   "/home/public/workspace/heongtong_yolo_linux/train.py",
								EnvVar:  []apis.EnvVar{apis.EnvVar{Name: "", Value: ""}},
							},
							apis.Runtime{
								Name:    runtime1_1_2_2Name,
								Type:    apis.ByCommand,
								Command: []string{"/home/public/anaconda3/envs/yolo/bin/python"},
								Args:    []string{"/home/public/workspace/heongtong_yolo_linux/train.py"},
								Parents: []string{runtime1_1_2_1Name}, // 加入Parents
								Image:   "/home/public/workspace/heongtong_yolo_linux/train.py",
								EnvVar:  []apis.EnvVar{apis.EnvVar{Name: "", Value: ""}},
							},
						},
					},
					Status: apis.ActionStatus{
						ActionID: action1_1_2ID, // ActionID =ActionName + GroupID
						Phase:    apis.Unknown,
						RuntimeStatus: []apis.RuntimeStatus{
							apis.RuntimeStatus{
								RuntimeID: runtime1_1_2_1ID, // RuntimeID = RuntimeName + ActionID
								Phase:     apis.Unknown,
							},
							apis.RuntimeStatus{
								RuntimeID: runtime1_1_2_2ID, // RuntimeID = RuntimeName + ActionID
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
				apis.ActionStatus{
					ActionID: action1_1_2ID,
					RuntimeStatus: []apis.RuntimeStatus{
						apis.RuntimeStatus{
							RuntimeID: runtime1_1_2_1ID,
							Phase:     apis.Unknown,
						},
						apis.RuntimeStatus{
							RuntimeID: runtime1_1_2_2ID,
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

	g2 := apis.Group{
		ObjectMeta: metav1.ObjectMeta{Name: group1_2Name, Namespace: ""},
		TypeMeta:   metav1.TypeMeta{Kind: "Group", APIVersion: "resources/v1"},
		Spec: apis.GroupSpec{
			Name:    group1_2Name,
			Parents: []string{group1_1Name}, // 加入Parents
			Actions: []apis.Action{
				apis.Action{
					ObjectMeta: metav1.ObjectMeta{Name: action1_2_1Name},
					Spec: apis.ActionSpec{
						Name:    action1_2_1Name,
						Parents: make([]string, 0),
						Runtimes: []apis.Runtime{
							apis.Runtime{
								Name:    runtime1_2_1_1Name,
								Type:    apis.ByCommand,
								Command: []string{"/home/public/anaconda3/envs/yolo/bin/python"},
								Args:    []string{"/home/public/workspace/heongtong_yolo_linux/predict.py"},
								Parents: make([]string, 0), // 加入Parents
								Image:   "/home/public/workspace/heongtong_yolo_linux/predict.py",
								EnvVar:  []apis.EnvVar{apis.EnvVar{Name: "", Value: ""}},
							},
							apis.Runtime{
								Name:    runtime1_2_1_2Name,
								Type:    apis.ByCommand,
								Command: []string{"/home/public/anaconda3/envs/yolo/bin/python"},
								Args:    []string{"/home/public/workspace/heongtong_yolo_linux/predict.py"},
								Parents: []string{runtime1_2_1_1Name}, // 加入Parents
								Image:   "/home/public/workspace/heongtong_yolo_linux/predict.py",
								EnvVar:  []apis.EnvVar{apis.EnvVar{Name: "", Value: ""}},
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
				apis.Action{
					ObjectMeta: metav1.ObjectMeta{Name: action1_2_2Name},
					Spec: apis.ActionSpec{
						Name:    action1_2_2Name,
						Parents: []string{action1_2_1Name},
						Runtimes: []apis.Runtime{
							apis.Runtime{
								Name:    runtime1_2_2_1Name,
								Type:    apis.ByCommand,
								Command: []string{"/home/public/anaconda3/envs/yolo/bin/python"},
								Args:    []string{"/home/public/workspace/heongtong_yolo_linux/predict.py"},
								Parents: make([]string, 0), // 加入Parents
								Image:   "/home/public/workspace/heongtong_yolo_linux/predict.py",
								EnvVar:  []apis.EnvVar{apis.EnvVar{Name: "", Value: ""}},
							},
							apis.Runtime{
								Name:    runtime1_2_2_2Name,
								Type:    apis.ByCommand,
								Command: []string{"/home/public/anaconda3/envs/yolo/bin/python"},
								Args:    []string{"/home/public/workspace/heongtong_yolo_linux/predict.py"},
								Parents: []string{runtime1_2_2_1Name}, // 加入Parents
								Image:   "/home/public/workspace/heongtong_yolo_linux/predict.py",
								EnvVar:  []apis.EnvVar{apis.EnvVar{Name: "", Value: ""}},
							},
						},
					},
					Status: apis.ActionStatus{
						ActionID: action1_2_2ID,
						Phase:    apis.Unknown,
						RuntimeStatus: []apis.RuntimeStatus{
							apis.RuntimeStatus{
								RuntimeID: runtime1_2_2_1ID,
								Phase:     apis.Unknown,
							},
							apis.RuntimeStatus{
								RuntimeID: runtime1_2_2_2ID,
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
				apis.ActionStatus{
					ActionID: action1_2_2ID,
					RuntimeStatus: []apis.RuntimeStatus{
						apis.RuntimeStatus{
							RuntimeID: runtime1_2_2_1ID,
							Phase:     apis.Unknown,
						},
						apis.RuntimeStatus{
							RuntimeID: runtime1_2_2_2ID,
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

	g3 := apis.Group{
		ObjectMeta: metav1.ObjectMeta{Name: group1_3Name, Namespace: ""},
		TypeMeta:   metav1.TypeMeta{Kind: "Group", APIVersion: "resources/v1"},
		Spec: apis.GroupSpec{
			Name:    group1_3Name,
			Parents: []string{group1_1Name, group1_2Name}, // 加入Parents ID
			Actions: []apis.Action{
				apis.Action{
					ObjectMeta: metav1.ObjectMeta{Name: action1_3_1Name},
					Spec: apis.ActionSpec{
						Name:    action1_3_1Name,
						Parents: make([]string, 0),
						Runtimes: []apis.Runtime{
							apis.Runtime{
								Name:    runtime1_3_1_1Name,
								Type:    apis.ByCommand,
								Command: []string{"/home/public/anaconda3/envs/yolo/bin/python"},
								Args:    []string{"/home/public/workspace/heongtong_yolo_linux/predict.py"},
								Parents: make([]string, 0), // 加入Parents
								Image:   "/home/public/workspace/heongtong_yolo_linux/predict.py",
								EnvVar:  []apis.EnvVar{apis.EnvVar{Name: "", Value: ""}},
							},
							apis.Runtime{
								Name:    runtime1_3_1_2Name,
								Type:    apis.ByCommand,
								Command: []string{"/home/public/anaconda3/envs/yolo/bin/python"},
								Args:    []string{"/home/public/workspace/heongtong_yolo_linux/predict.py"},
								Parents: []string{runtime1_3_1_1Name}, // 加入Parents
								Image:   "/home/public/workspace/heongtong_yolo_linux/predict.py",
								EnvVar:  []apis.EnvVar{apis.EnvVar{Name: "", Value: ""}},
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
				apis.Action{
					ObjectMeta: metav1.ObjectMeta{Name: action1_3_2Name},
					Spec: apis.ActionSpec{
						Name:    action1_3_2Name,
						Parents: []string{action1_3_1Name},
						Runtimes: []apis.Runtime{
							apis.Runtime{
								Name:    runtime1_3_2_1Name,
								Type:    apis.ByCommand,
								Command: []string{"/home/public/anaconda3/envs/yolo/bin/python"},
								Args:    []string{"/home/public/workspace/heongtong_yolo_linux/predict.py"},
								Parents: make([]string, 0), // 加入Parents
								Image:   "/home/public/workspace/heongtong_yolo_linux/predict.py",
								EnvVar:  []apis.EnvVar{apis.EnvVar{Name: "", Value: ""}},
							},
							apis.Runtime{
								Name:    runtime1_3_2_2Name,
								Type:    apis.ByCommand,
								Command: []string{"/home/public/anaconda3/envs/yolo/bin/python"},
								Args:    []string{"/home/public/workspace/heongtong_yolo_linux/predict.py"},
								Parents: []string{runtime1_3_2_1Name}, // 加入Parents
								Image:   "/home/public/workspace/heongtong_yolo_linux/predict.py",
								EnvVar:  []apis.EnvVar{apis.EnvVar{Name: "", Value: ""}},
							},
						},
					},
					Status: apis.ActionStatus{
						ActionID: action1_3_2ID, // ActionID =ActionName + GroupID
						Phase:    apis.Unknown,
						RuntimeStatus: []apis.RuntimeStatus{
							apis.RuntimeStatus{
								RuntimeID: runtime1_3_2_1ID, // RuntimeID = RuntimeName + ActionID
								Phase:     apis.Unknown,
							},
							apis.RuntimeStatus{
								RuntimeID: runtime1_3_2_2ID, // RuntimeID = RuntimeName + ActionID
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
				apis.ActionStatus{
					ActionID: action1_3_2ID,
					RuntimeStatus: []apis.RuntimeStatus{
						apis.RuntimeStatus{
							RuntimeID: runtime1_3_2_1ID,
							Phase:     apis.Unknown,
						},
						apis.RuntimeStatus{
							RuntimeID: runtime1_3_2_2ID,
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
	//group3 := &g3

	task := apis.Task{
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
				g1, g2, g3,
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
								apis.RuntimeStatus{
									RuntimeID: runtime1_1_1_2ID,
									Phase:     apis.Unknown,
								},
							},
							Phase: apis.Unknown,
						},
						apis.ActionStatus{
							ActionID: action1_1_2ID,
							RuntimeStatus: []apis.RuntimeStatus{
								apis.RuntimeStatus{
									RuntimeID: runtime1_1_2_1ID,
									Phase:     apis.Unknown,
								},
								apis.RuntimeStatus{
									RuntimeID: runtime1_1_2_2ID,
									Phase:     apis.Unknown,
								},
							},
							Phase: apis.Unknown,
						},
					},
					Belongs: apis.IDRef{TaskID: task1ID},
					Phase:   apis.Unknown,
				},
				apis.GroupStatus{
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
						apis.ActionStatus{
							ActionID: action1_2_2ID,
							RuntimeStatus: []apis.RuntimeStatus{
								apis.RuntimeStatus{
									RuntimeID: runtime1_2_2_1ID,
									Phase:     apis.Unknown,
								},
								apis.RuntimeStatus{
									RuntimeID: runtime1_2_2_2ID,
									Phase:     apis.Unknown,
								},
							},
							Phase: apis.Unknown,
						},
					},
					Belongs: apis.IDRef{TaskID: task1ID},
					Phase:   apis.Unknown,
				},
				apis.GroupStatus{
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
						apis.ActionStatus{
							ActionID: action1_3_2ID,
							RuntimeStatus: []apis.RuntimeStatus{
								apis.RuntimeStatus{
									RuntimeID: runtime1_3_2_1ID,
									Phase:     apis.Unknown,
								},
								apis.RuntimeStatus{
									RuntimeID: runtime1_3_2_2ID,
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

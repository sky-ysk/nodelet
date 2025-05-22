package main

import (
	"github.com/google/uuid"
	apis "hit.edu/framework/pkg/apis/cores"
	metav1 "hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/client-go/util/manager"
	"hit.edu/framework/pkg/component-base/logs"
	"hit.edu/framework/pkg/nodelet/device"
	"time"
)

func main() {
	//path := "/home/public"

	//场景三 服务器路径
	path := "/home/smj"

	// 手臂初始化
	runtime0 := &apis.Runtime{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "R0",
			Namespace: "test",
			Labels: map[string]string{
				"environment": "dev",
			},
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Runtime",
			APIVersion: "resources/v1",
		},
		Spec: apis.RuntimeSpec{
			Name: "R0",
			Type: apis.ByDevice,
			Inputs: []apis.Value{
				{
					Name:      "label",
					Value:     "up",
					ValueType: apis.StringType,
					Type:      apis.ConstData,
				},
			},
			Image: "Device{Robot1}.Ability{Grab}.Service{GrabInit}",
			Devices: []apis.DeviceSpec{
				apis.DeviceSpec{
					Name: "Robot1",
					ExpectedProperties: map[string]apis.Property{
						"name": apis.Property{},
					},
					Abilities: []string{
						"Grab",
					},
				},
			},
		},
	}

	// 抓取流水线工件
	runtime1 := &apis.Runtime{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "R1",
			Namespace: "test",
			Labels: map[string]string{
				"environment": "dev",
			},
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Runtime",
			APIVersion: "resources/v1",
		},
		Spec: apis.RuntimeSpec{
			Parents: []string{"R0"},
			Name:    "R1",
			Type:    apis.ByDevice,
			Inputs: []apis.Value{
				{
					Name:      "label",
					Value:     "belt",
					ValueType: apis.StringType,
					Type:      apis.ConstData,
				},
			},
			Image: "Device{Robot1}.Ability{Grab}.Service{GrabWorkpiece}",
			Devices: []apis.DeviceSpec{
				apis.DeviceSpec{
					Name: "Robot1",
					ExpectedProperties: map[string]apis.Property{
						"name": apis.Property{},
					},
					Abilities: []string{
						"Grab",
					},
				},
			},
		},
	}

	// 转向复检台
	runtime2 := &apis.Runtime{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "R2",
			Namespace: "test",
			Labels: map[string]string{
				"environment": "dev",
			},
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Runtime",
			APIVersion: "resources/v1",
		},
		Spec: apis.RuntimeSpec{
			Name: "R2",
			Type: apis.ByDevice,
			Inputs: []apis.Value{
				{
					Name:      "label",
					Value:     "table",
					ValueType: apis.StringType,
					Type:      apis.ConstData,
				},
			},
			Image: "Device{Robot1}.Ability{Turn}.Service{Turn}",
			Devices: []apis.DeviceSpec{
				apis.DeviceSpec{
					Name: "Robot1",
					ExpectedProperties: map[string]apis.Property{
						"name": apis.Property{},
					},
					Abilities: []string{
						"Turn",
					},
				},
			},
		},
	}

	// 检测
	runtime3 := &apis.Runtime{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "R3",
			Namespace: "test",
			Labels: map[string]string{
				"environment": "dev",
			},
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Runtime",
			APIVersion: "resources/v1",
		},
		Spec: apis.RuntimeSpec{
			Name:    "R3",
			Type:    apis.ByDevice,
			Outputs: []apis.Value{},
			Image:   "Device{Robot1}.Ability{Detect}.Service{DetectWorkpiece}",
			Devices: []apis.DeviceSpec{
				apis.DeviceSpec{
					Name: "Robot1",
					ExpectedProperties: map[string]apis.Property{
						"name": apis.Property{},
					},
					Abilities: []string{
						"Detect",
					},
				},
			},
		},
	}

	// 放入合格框
	runtime4 := &apis.Runtime{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "R4",
			Namespace: "test",
			Labels: map[string]string{
				"environment": "dev",
			},
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Runtime",
			APIVersion: "resources/v1",
		},
		Spec: apis.RuntimeSpec{
			Parents: []string{"R3"},
			Conditions: &apis.Conditions{
				Formulas: []apis.ConditionFormula{
					{
						Signal:        apis.Equal,
						ConditionType: apis.DataDependency,
						LeftValue: apis.Value{
							NameSpace: apis.NamespaceTest,
							Type:      apis.LocalData,
							Value:     "",
							ValueType: apis.BoolType,
							From:      "Runtime{R3}.Outputs{isQualified}",
						},
						RightValue: apis.Value{
							NameSpace: apis.NamespaceTest,
							Type:      apis.ConstData,
							Value:     "true",
							ValueType: apis.BoolType,
						},
					},
				},
			},
			Name: "R4",
			Type: apis.ByDevice,
			Inputs: []apis.Value{
				{
					Name:      "label",
					Value:     "finished_bin",
					ValueType: apis.StringType,
					Type:      apis.ConstData,
				},
			},
			Image: "Device{Robot1}.Ability{Put}.Service{PutWorkpiece}",
			Devices: []apis.DeviceSpec{
				apis.DeviceSpec{
					Name: "Robot1",
					ExpectedProperties: map[string]apis.Property{
						"name": apis.Property{},
					},
					Abilities: []string{
						"Put",
					},
				},
			},
		},
	}

	// 放置到复检台
	runtime5 := &apis.Runtime{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "R5",
			Namespace: "test",
			Labels: map[string]string{
				"environment": "dev",
			},
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Runtime",
			APIVersion: "resources/v1",
		},
		Spec: apis.RuntimeSpec{
			Parents: []string{"R3"},
			Conditions: &apis.Conditions{
				Formulas: []apis.ConditionFormula{
					{
						Signal:        apis.Equal,
						ConditionType: apis.DataDependency,
						LeftValue: apis.Value{
							NameSpace: apis.NamespaceTest,
							Type:      apis.LocalData,
							Value:     "",
							ValueType: apis.BoolType,
							From:      "Runtime{R3}.Outputs{isQualified}",
						},
						RightValue: apis.Value{
							NameSpace: apis.NamespaceTest,
							Type:      apis.ConstData,
							Value:     "false",
							ValueType: apis.BoolType,
						},
					},
				},
			},
			Name: "R5",
			Type: apis.ByDevice,
			Inputs: []apis.Value{
				{
					Name:      "label",
					Value:     "table",
					ValueType: apis.StringType,
					Type:      apis.ConstData,
				},
			},
			Image: "Device{Robot1}.Ability{Put}.Service{PutWorkpiece}",
			Devices: []apis.DeviceSpec{
				apis.DeviceSpec{
					Name: "Robot1",
					ExpectedProperties: map[string]apis.Property{
						"name": apis.Property{},
					},
					Abilities: []string{
						"Put",
					},
				},
			},
		},
	}

	//提交复检任务
	runtime6 := &apis.Runtime{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "R6",
			Namespace: "test",
			Labels: map[string]string{
				"environment": "dev",
			},
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Runtime",
			APIVersion: "resources/v1",
		},
		Spec: apis.RuntimeSpec{
			Parents: []string{"R5"},
			Name:    "R6",
			Type:    apis.ByCommand,
			Command: []string{"python3"}, // 801的机器 要用python3
			Args: []string{"" +
				path + "/adaptive-scheduling-framework/test/scene3/recheck1.py",
				path + "/adaptive-scheduling-framework/test/scene3/data.txt"},
			// 这个是smj
			//Args: []string{"" +
			//	"/home/public/lock_test/test/scene3/recheck1.py",
			//	"/home/public/lock_test/test/scene3/data.txt",
			//	"device1"},
			// 这个是801的路径
		},
	}

	// 转向流水线
	runtime7 := &apis.Runtime{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "R7",
			Namespace: "test",
			Labels: map[string]string{
				"environment": "dev",
			},
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Runtime",
			APIVersion: "resources/v1",
		},
		Spec: apis.RuntimeSpec{
			Name: "R7",
			Type: apis.ByDevice,
			Inputs: []apis.Value{
				{
					Name:      "label",
					Value:     "belt",
					ValueType: apis.StringType,
					Type:      apis.ConstData,
				},
			},
			Image: "Device{Robot1}.Ability{Turn}.Service{Turn}",
			Devices: []apis.DeviceSpec{
				apis.DeviceSpec{
					Name: "Robot1",
					ExpectedProperties: map[string]apis.Property{
						"name": apis.Property{},
					},
					Abilities: []string{
						"Turn",
					},
				},
			},
		},
	}

	// runtime0 和 runtime1
	action1 := &apis.Action{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "A1",
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
			Desc: &apis.Description{
				Docs: "抓取物品",
			},
			Name: "A1",
			Runtimes: []apis.RuntimeSpec{
				runtime0.Spec, runtime1.Spec,
			},
		},
	}

	// runtime2
	action2 := &apis.Action{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "A2",
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
			Desc: &apis.Description{
				Docs: "转身复检台",
			},
			Name: "A2",
			Runtimes: []apis.RuntimeSpec{
				runtime2.Spec,
			},
			Parents: []string{"A1"},
		},
	}

	// runtime3 runtime4 runtime5 runtime6
	action3 := &apis.Action{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "A3",
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
			Desc: &apis.Description{
				Docs: "检测/提交复检任务",
			},
			Name: "A3",
			Runtimes: []apis.RuntimeSpec{
				runtime3.Spec, runtime4.Spec, runtime5.Spec, runtime6.Spec,
			},
			Parents: []string{"A2"},
		},
	}

	// runtime7
	action4 := &apis.Action{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "A4",
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
			Desc: &apis.Description{
				Docs: "转向流水线",
			},
			Name: "A4",
			Runtimes: []apis.RuntimeSpec{
				runtime7.Spec,
			},
			Parents: []string{"A3"},
		},
	}

	group1 := &apis.Group{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "G1",
			Namespace: "test",
			Labels: map[string]string{
				"environment": "dev",
			},
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Group",
			APIVersion: "resources/v1",
		},
		Spec: apis.GroupSpec{

			Devices: []apis.DeviceSpec{
				{
					Name: "Robot1",
					ExpectedProperties: map[string]apis.Property{
						"name": apis.Property{
							Value: "device1",
						},
						"strategy": apis.Property{
							Value: "nominate",
						},
					},
					Abilities: []string{
						"Turn", "Grab", "Detect", "Put",
					},
				},
			},
			Desc: &apis.Description{
				Docs:  "场景三的初检group",
				Label: []string{"scene3"},
			},
			Name: "G1",
			Actions: []apis.ActionSpec{
				action1.Spec, action2.Spec, action3.Spec, action4.Spec,
			},
			Replicas: []int32{0, 0},
		},
	}
	task1 := &apis.Task{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "T1",
			Namespace: "test",
			Labels: map[string]string{
				"environment": "dev",
			},
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Task",
			APIVersion: "resources/v1",
		},
		Spec: apis.TaskSpec{
			Desc: &apis.Description{
				Docs: "初检任务task",
			},
			Name: "T1",
			Groups: []apis.GroupSpec{
				group1.Spec,
			},
		},
	}
	clientSet, _ := device.InitClient()
	m := manager.NewManager(clientSet)
	timestamp := time.Now().Format("20060102T150405")
	randomStr := uuid.New().String()[:5]

	UUID := timestamp + "-" + randomStr

	_, err := m.CreateTask(task1.Spec, nil, apis.NamespaceTest, UUID, "")
	if err != nil {
		logs.Errorf("create task1 fail")
	}
}

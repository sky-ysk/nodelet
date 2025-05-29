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
	//星海图下载
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
			Name: "R1",
			Type: apis.ByDevice,
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
			Image: "Device{Robot1}.Ability{Detect}.Service{Download}",
			Inputs: []apis.Value{
				{
					Name:      "user_id",
					Type:      apis.ConstData,
					ValueType: apis.StringType,
					Value:     "1",
				},
				{
					Name:      "model_id",
					Type:      apis.ConstData,
					ValueType: apis.StringType,
					Value:     "2",
				},
				{
					Name:      "path",
					Type:      apis.ConstData,
					ValueType: apis.StringType,
					Value:     "",
				},
				{
					Name:      "filename",
					Type:      apis.ConstData,
					ValueType: apis.StringType,
					Value:     "ball.onnx",
				},
			},
		},
	}

	// 星海图检测
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
			Parents: []string{"R1"},
			Name:    "R2",
			Type:    apis.ByDevice,
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
			Image: "Device{Robot1}.Ability{Detect}.Service{DetectPosition}",
			Outputs: []apis.Value{
				{
					Name:      "worldPoints",
					Type:      apis.ConstData,
					ValueType: apis.ComposeType,
				},
				{
					Name:      "Success",
					Type:      apis.LocalData,
					ValueType: apis.BoolType,
				},
			},
		},
	}

	// 星海图抓取
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
			Name: "R3",
			Type: apis.ByDevice,
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
			Image: "Device{Robot}.Ability{Grab}.Service{GrabBall}",
			Inputs: []apis.Value{
				apis.Value{
					Name: "worldPoints",
					Type: apis.LocalData,
					From: "Action{A1}.Runtime{R2}.Outputs{worldPoints}", // TODO
				},
			},
		},
	}

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
			Name: "R4",
			Type: apis.ByDevice,
			Devices: []apis.DeviceSpec{
				apis.DeviceSpec{
					Name: "Robot2",
					ExpectedProperties: map[string]apis.Property{
						"name": apis.Property{},
					},
					Abilities: []string{
						"Detect",
					},
				},
			},
			Image: "Device{Robot2}.Ability{Detect}.Service{Download}",
			Inputs: []apis.Value{
				{
					Name:      "user_id",
					Type:      apis.ConstData,
					ValueType: apis.StringType,
					Value:     "1",
				},
				{
					Name:      "model_id",
					Type:      apis.ConstData,
					ValueType: apis.StringType,
					Value:     "2",
				},
				{
					Name:      "path",
					Type:      apis.ConstData,
					ValueType: apis.StringType,
					Value:     "",
				},
				{
					Name:      "filename",
					Type:      apis.ConstData,
					ValueType: apis.StringType,
					Value:     "ball.onnx",
				},
			},
			Outputs: []apis.Value{
				{
					Name:      "Success",
					Type:      apis.LocalData,
					ValueType: apis.BoolType,
				},
			},
		},
	}

	// 乐聚检测
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
			Name:    "R5",
			Type:    apis.ByDevice,
			Parents: []string{"R4"},
			Devices: []apis.DeviceSpec{
				apis.DeviceSpec{
					Name: "Robot2",
					ExpectedProperties: map[string]apis.Property{
						"name": apis.Property{},
					},
					Abilities: []string{
						"Detect",
					},
				},
			},
			Image: "Device{Robot2}.Ability{Detect}.Service{DetectPosition}",
			Outputs: []apis.Value{
				{
					Name:      "worldPoints",
					Type:      apis.ConstData,
					ValueType: apis.ComposeType,
				},
				{
					Name:      "Success",
					Type:      apis.LocalData,
					ValueType: apis.BoolType,
				},
			},
		},
	}

	// 乐聚抓取
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
			Name: "R6",
			Type: apis.ByDevice,
			Devices: []apis.DeviceSpec{
				apis.DeviceSpec{
					Name: "Robot2",
					ExpectedProperties: map[string]apis.Property{
						"name": apis.Property{},
					},
					Abilities: []string{
						"Grab",
					},
				},
			},
			Image: "Device{Robot2}.Ability{Grab}.Service{GrabBall}",
			Inputs: []apis.Value{
				apis.Value{
					Name: "worldPoints",
					Type: apis.LocalData,
					From: "Action{A2}.Runtime{R5}.Outputs{worldPoints}", // TODO
				},
			},
		},
	}

	// 星海图模型下载和检测
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
				Docs: "星海图检测",
			},
			Name: "A1",
			Runtimes: []apis.RuntimeSpec{
				runtime1.Spec,
				runtime2.Spec,
			},
		},
	}

	// 乐聚模型下载和检测
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
				Docs: "乐聚检测",
			},
			Name: "A2",
			Runtimes: []apis.RuntimeSpec{
				runtime4.Spec,
				runtime5.Spec,
			},
		},
	}

	// 星海图抓取
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
				Docs: "星海图抓取",
			},
			Name: "A3",
			Runtimes: []apis.RuntimeSpec{
				runtime3.Spec,
			},
			Parents: []string{"A1"},
		},
	}

	// 乐聚抓取
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
				Docs: "乐聚抓取",
			},
			Name: "A4",
			Runtimes: []apis.RuntimeSpec{
				runtime6.Spec,
			},
			Parents: []string{"A2"},
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
					Abilities: []string{
						"Grab", "Detect",
					},
					ExpectedProperties: map[string]apis.Property{
						"name": apis.Property{
							Value: "deviceGalaxea",
						},
						"strategy": apis.Property{
							Value: "nominate",
						},
					},
				},
			},
			Desc: &apis.Description{
				Label: map[string]string{
					"scene": "scene1",
				},
				Docs: "星海图检测和抓取",
			},
			Name: "G1",
			Actions: []apis.ActionSpec{
				action1.Spec, action3.Spec,
			},
			Parents:  []string{"G2"},
			Replicas: []int32{0, 0},
		},
	}

	group2 := &apis.Group{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "G2",
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
					Name: "Robot2",
					Abilities: []string{
						"Grab", "Detect",
					},
					ExpectedProperties: map[string]apis.Property{
						"name": apis.Property{
							Value: "deviceLeju",
						},
						"strategy": apis.Property{
							Value: "nominate",
						},
					},
				},
			},
			Desc: &apis.Description{
				Docs: "乐聚检测和抓取",
				Label: map[string]string{
					"scene": "scene1",
				},
			},
			Name: "G2",
			Actions: []apis.ActionSpec{
				action2.Spec, action4.Spec,
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
				Docs: "星海图&乐聚的检测和抓取",
			},
			Name: "T1",
			Groups: []apis.GroupSpec{
				group1.Spec, group2.Spec,
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

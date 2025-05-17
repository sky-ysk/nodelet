package scheduler

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	apis "hit.edu/framework/pkg/apis/cores"
	metav1 "hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/client-go/util/manager"
	"hit.edu/framework/pkg/component-base/logs"
	"hit.edu/framework/pkg/nodelet/device"
	"hit.edu/framework/pkg/scheduler/utils"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

func createConditionTask() apis.Task {

	group1Action1Runtime1 := apis.RuntimeSpec{
		Name: "ConditionGroup1Action1Runtime1",
	}

	group1Action1 := apis.ActionSpec{
		Name: "ConditionGroup1Action1",
		Runtimes: []apis.RuntimeSpec{
			group1Action1Runtime1,
		},
	}

	group1Spec := apis.GroupSpec{
		Name: "ConditionGroup1",
		Actions: []apis.ActionSpec{
			group1Action1,
		},
	}

	group2Action1Runtime1 := apis.RuntimeSpec{
		Name: "ConditionGroup2Action1Runtime1",
	}

	group2Action1 := apis.ActionSpec{
		Name: "ConditionGroup2Action1",
		Runtimes: []apis.RuntimeSpec{
			group2Action1Runtime1,
		},
	}

	group2Spec := apis.GroupSpec{
		Name: "ConditionGroup2",
		Actions: []apis.ActionSpec{
			group2Action1,
		},
	}

	task := apis.Task{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Task",
			APIVersion: "resources/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ConditionTask",
			Namespace: apis.NamespaceTest,
		},
		Spec: apis.TaskSpec{
			Name: "ConditionTask",
			Groups: []apis.GroupSpec{
				group1Spec,
				group2Spec,
			},
		},
		Status: apis.TaskStatus{
			Groups: map[string]apis.ObjectReference{
				"ConditionGroup1": {
					Kind: "Group",
					Name: "ConditionGroup1",
				},
				"ConditionGroup2": {
					Kind: "Group",
					Name: "ConditionGroup2",
				},
			},
		},
	}
	return task
}

func createOrangeTask() apis.Task {
	task := apis.Task{}
	return task
}

// 新的场景一的task
func createSceneOneTask() apis.Task {

	kuavoDetectRuntime := apis.RuntimeSpec{
		Name: "KuavoDetectRuntime",
	}

	kuavoDetectAction := apis.ActionSpec{
		Name: "KuavoDetectAction",
		Runtimes: []apis.RuntimeSpec{
			kuavoDetectRuntime,
		},
	}

	kuavoDetectGroup := apis.GroupSpec{
		Name: "KuavoDetectGroup",
		Actions: []apis.ActionSpec{
			kuavoDetectAction,
		},
	}

	kuavoGrabRuntime := apis.RuntimeSpec{
		Name: "KuavoGrabRuntime",
	}

	kuavoGrabAction := apis.ActionSpec{
		Name: "KuavoGrabAction",
		Runtimes: []apis.RuntimeSpec{
			kuavoGrabRuntime,
		},
	}

	kuavoGrabGroup := apis.GroupSpec{
		Name: "kuavoGrabGroup",
		Actions: []apis.ActionSpec{
			kuavoGrabAction,
		},
		Parents: []string{"KuavoDetectGroup"},
		Conditions: &apis.Conditions{
			Formulas: []apis.ConditionFormula{
				{},
			},
		},
	}

	galaxeaDetectRuntime := apis.RuntimeSpec{
		Name: "GalaxeaDetectRuntime",
	}

	galaxeaDetectAction := apis.ActionSpec{
		Name: "GalaxeaDetectAction",
		Runtimes: []apis.RuntimeSpec{
			galaxeaDetectRuntime,
		},
	}

	galaxeaDetectGroup := apis.GroupSpec{
		Name: "GalaxeaDetectGroup",
		Actions: []apis.ActionSpec{
			galaxeaDetectAction,
		},
		Parents: []string{"KuavoDetectGroup"},
	}

	galaxeaGrabRuntime := apis.RuntimeSpec{
		Name: "GalaxeaGrabRuntime",
	}

	galaxeaGrabAction := apis.ActionSpec{
		Name: "GalaxeaGrabAction",
		Runtimes: []apis.RuntimeSpec{
			galaxeaGrabRuntime,
		},
	}

	galaxeaGrabGroup := apis.GroupSpec{
		Name: "GalaxeaGrabGroup",
		Actions: []apis.ActionSpec{
			galaxeaGrabAction,
		},
		Parents: []string{"GalaxeaDetectGroup"},
		Conditions: &apis.Conditions{
			Formulas: []apis.ConditionFormula{
				{},
			},
		},
	}

	task := apis.Task{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Task",
			APIVersion: "resources/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "Scene1Task",
			Namespace: apis.NamespaceTest,
		},
		Spec: apis.TaskSpec{
			Name: "ConditionTask",
			Groups: []apis.GroupSpec{
				kuavoDetectGroup,
				kuavoGrabGroup,
				galaxeaDetectGroup,
				galaxeaGrabGroup,
			},
		},
		Status: apis.TaskStatus{
			Groups: map[string]apis.ObjectReference{
				"ConditionGroup1": {
					Kind: "Group",
					Name: "ConditionGroup1",
				},
				"ConditionGroup2": {
					Kind: "Group",
					Name: "ConditionGroup2",
				},
			},
		},
	}
	return task
}

// go test -run TestAddGroup -v
func TestAddGroup(t *testing.T) {
	group := apis.Group{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Group",
			APIVersion: "resources/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "testGroup1",
			Namespace: apis.NamespaceTest,
		},
		Spec: apis.GroupSpec{
			Name: "testGroup1",
		},
		Status: apis.GroupStatus{
			Phase: apis.Unknown,
		},
	}

	cs, err := utils.CreateClientSet()
	if err != nil {
		t.Fatalf("%v", err)
	}
	ctx := context.Background()
	gc := cs.Core().Groups(apis.NamespaceTest)
	_, err = gc.Create(ctx, &group, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("%v", err)
		return
	}
}

// go test -run TestSendToProxy -v
func TestSendToProxy(t *testing.T) {
	logs.Init("testModule")
	orange, err := os.ReadFile("orange.json")
	client := &http.Client{}

	url := "http://192.168.8.176:8899/framework/v1/task?Name=Scene1Task&&Namesapce=test"
	logs.Info(url)
	req, err := http.NewRequest("POST", url, strings.NewReader(string(orange)))
	if err != nil {
		logs.Fatal(err)
	}
	//Content-Type很重要，下文解释
	//req.Header.Set("Content-Type", "application/x-www")
	req.Header.Set("Content-Type", "application/json")
	//req.Header.Set("Content-Type", "multipart/form-data")

	rep, err := client.Do(req)
	if err != nil {
		logs.Fatal(err.Error())
	}
	data, err := io.ReadAll(rep.Body)
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			logs.Error(err)
		}
	}(rep.Body)
	if err != nil {
		logs.Fatal(err)
	}
	logs.Infof("resp is : %s", string(data))
}

func TestDevice(t *testing.T) {
	cs, err := utils.CreateClientSetWithTimeOut(2000)
	if err != nil {
		panic(err)
	}
	m := manager.NewManager(cs)
	// 创建Device
	deviceGalaxea := &apis.Device{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "deviceGalaxea",
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
			Name: "deviceGalaxea",
			AccessMethod: &apis.AccessMethod{
				Type: apis.AccessByAbility,
				URL:  "http://192.168.8.197:8080",
			},
			Abilities: []string{
				"Detect", "Grab",
			},
		},
		Status: apis.DeviceStatus{
			Abilities: map[string]apis.Ability{
				"Detect": {
					Name: "DetectPosition.Galaxea.Guochuang",
					Services: map[string]apis.AbilityService{
						"DetectPosition": apis.AbilityService{
							Ip:        new(string),
							Interface: new(string),
							Port:      new(string),
						},
						"Download": apis.AbilityService{
							Ip:        new(string),
							Interface: new(string),
							Port:      new(string),
						},
					},
					Status: apis.AbilityReadyStartUp,
				},
				"Grab": {
					Name: "GrabBall.Galaxea.Guochuang",
					Services: map[string]apis.AbilityService{
						"GrabBall": apis.AbilityService{
							Ip:        new(string),
							Interface: new(string),
							Port:      new(string),
						},
					},
					Status: apis.AbilityReadyStartUp,
				},
			},
			Lock: apis.Lock{
				IsLocked: true,
				Ref:      2,
			},
			Phase: apis.DeviceIdle,
		},
	}
	*deviceGalaxea.Status.Abilities["Detect"].Services["DetectPosition"].Interface = "/api/task/detect"
	*deviceGalaxea.Status.Abilities["Detect"].Services["DetectPosition"].Ip = "192.168.8.197"
	*deviceGalaxea.Status.Abilities["Detect"].Services["Download"].Interface = "/api/task/down_new_model"
	*deviceGalaxea.Status.Abilities["Detect"].Services["Download"].Ip = "192.168.8.197"
	*deviceGalaxea.Status.Abilities["Grab"].Services["GrabBall"].Interface = "/api/task/grab_ball"
	*deviceGalaxea.Status.Abilities["Grab"].Services["GrabBall"].Ip = "192.168.8.197"

	_, err = m.CreateDevice(deviceGalaxea, "test")
	if err != nil {
		logs.Errorf("[TEST] Create Device[%s] err:%s", deviceGalaxea.Name, err.Error())
	}

	deviceLeju := &apis.Device{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "deviceLeju",
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
			Name: "deviceLeju",
			AccessMethod: &apis.AccessMethod{
				Type: apis.AccessByAbility,
				URL:  "http://192.168.8.165:8080",
			},
			Abilities: []string{
				"Detect", "Grab",
			},
		},
		Status: apis.DeviceStatus{
			Abilities: map[string]apis.Ability{
				"Detect": {
					Name: "DetectPosition.Leju.Guochuang",
					Services: map[string]apis.AbilityService{
						"DetectPosition": apis.AbilityService{
							Ip:        new(string),
							Interface: new(string),
							Port:      new(string),
						},
						"Download": apis.AbilityService{
							Ip:        new(string),
							Interface: new(string),
							Port:      new(string),
						},
					},
					Status: apis.AbilityReadyStartUp,
				},
				"Grab": {
					Name: "GrabBall.Leju.Guochuang",
					Services: map[string]apis.AbilityService{
						"GrabBall": apis.AbilityService{
							Ip:        new(string),
							Interface: new(string),
							Port:      new(string),
						},
					},
					Status: apis.AbilityReadyStartUp,
				},
			},
			Lock: apis.Lock{
				IsLocked: true,
				Ref:      2,
			},
			Phase: apis.DeviceIdle,
		},
	}
	*deviceLeju.Status.Abilities["Detect"].Services["DetectPosition"].Interface = "/api/task/detect"
	*deviceLeju.Status.Abilities["Detect"].Services["DetectPosition"].Ip = "192.168.8.165"
	*deviceLeju.Status.Abilities["Detect"].Services["Download"].Interface = "/api/task/down_new_model"
	*deviceLeju.Status.Abilities["Detect"].Services["Download"].Ip = "192.168.8.165"
	*deviceLeju.Status.Abilities["Grab"].Services["GrabBall"].Interface = "/api/task/grab_ball"
	*deviceLeju.Status.Abilities["Grab"].Services["GrabBall"].Ip = "192.168.8.165"
	_, err = m.CreateDevice(deviceLeju, "test")
	if err != nil {
		logs.Errorf("[TEST] Create Device[%s] err:%s", deviceLeju.Name, err.Error())
	}
}

// go test -run TestCreateWorkFlow -v
func TestCreateWorkFlow(t *testing.T) {
	cs, err := utils.CreateClientSetWithTimeOut(2000)
	if err != nil {
		panic(err)
	}
	m := manager.NewManager(cs)
	// 创建Device
	deviceGalaxea := &apis.Device{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "deviceGalaxea",
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
			Name: "deviceGalaxea",
			AccessMethod: &apis.AccessMethod{
				Type: apis.AccessByAbility,
				URL:  "http://192.168.8.197:8080",
			},
			Abilities: []string{
				"Detect", "Grab",
			},
		},
		Status: apis.DeviceStatus{
			Abilities: map[string]apis.Ability{
				"Detect": {
					Name: "DetectPosition.Galaxea.Guochuang",
					Services: map[string]apis.AbilityService{
						"DetectPosition": apis.AbilityService{
							Ip:        new(string),
							Interface: new(string),
							Port:      new(string),
						},
					},
					Status: apis.AbilityReadyStartUp,
				},
				"Grab": {
					Name: "GrabBall.Galaxea.Guochuang",
					Services: map[string]apis.AbilityService{
						"GrabBall": apis.AbilityService{
							Ip:        new(string),
							Interface: new(string),
							Port:      new(string),
						},
					},
					Status: apis.AbilityReadyStartUp,
				},
			},
			Lock: apis.Lock{
				IsLocked: true,
				Ref:      2,
			},
			Phase: apis.DeviceIdle,
		},
	}
	*deviceGalaxea.Status.Abilities["Detect"].Services["DetectPosition"].Interface = "/api/task/detect"
	*deviceGalaxea.Status.Abilities["Detect"].Services["DetectPosition"].Ip = "192.168.8.197"
	*deviceGalaxea.Status.Abilities["Grab"].Services["GrabBall"].Interface = "/api/task/grab_ball"
	*deviceGalaxea.Status.Abilities["Grab"].Services["GrabBall"].Ip = "192.168.8.197"

	_, err = m.CreateDevice(deviceGalaxea, "test")
	if err != nil {
		logs.Errorf("[TEST] Create Device[%s] err:%s", deviceGalaxea.Name, err.Error())
	}

	deviceLeju := &apis.Device{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "deviceLeju",
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
			Name: "deviceLeju",
			AccessMethod: &apis.AccessMethod{
				Type: apis.AccessByAbility,
				URL:  "http://192.168.8.165:8080",
			},
			Abilities: []string{
				"Detect", "Grab",
			},
		},
		Status: apis.DeviceStatus{
			Abilities: map[string]apis.Ability{
				"Detect": {
					Name: "DetectPosition.Leju.Guochuang",
					Services: map[string]apis.AbilityService{
						"DetectPosition": apis.AbilityService{
							Ip:        new(string),
							Interface: new(string),
							Port:      new(string),
						},
						"Download": apis.AbilityService{
							Ip:        new(string),
							Interface: new(string),
							Port:      new(string),
						},
					},
					Status: apis.AbilityReadyStartUp,
				},
				"Grab": {
					Name: "GrabBall.Leju.Guochuang",
					Services: map[string]apis.AbilityService{
						"GrabBall": apis.AbilityService{
							Ip:        new(string),
							Interface: new(string),
							Port:      new(string),
						},
					},
					Status: apis.AbilityReadyStartUp,
				},
			},
			Lock: apis.Lock{
				IsLocked: true,
				Ref:      2,
			},
			Phase: apis.DeviceIdle,
		},
	}
	*deviceLeju.Status.Abilities["Detect"].Services["DetectPosition"].Interface = "/api/task/detect"
	*deviceLeju.Status.Abilities["Detect"].Services["DetectPosition"].Ip = "192.168.8.165"
	*deviceLeju.Status.Abilities["Detect"].Services["Download"].Interface = "/api/task/down_new_model"
	*deviceLeju.Status.Abilities["Detect"].Services["Download"].Ip = "192.168.8.165"
	*deviceLeju.Status.Abilities["Grab"].Services["GrabBall"].Interface = "/api/task/grab_ball"
	*deviceLeju.Status.Abilities["Grab"].Services["GrabBall"].Ip = "192.168.8.165"
	_, err = m.CreateDevice(deviceLeju, "test")
	if err != nil {
		logs.Errorf("[TEST] Create Device[%s] err:%s", deviceLeju.Name, err.Error())
	}

	time.Sleep(5 * time.Second)

	// 星海图检测
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
			Conditions: &apis.Conditions{
				Formulas: []apis.ConditionFormula{
					{
						LeftValue: apis.Value{
							Type:      apis.ResultsData,
							Name:      "NodeDependency",
							Value:     "0",
							ValueType: "string",
							From:      "R6",
						},
						RightValue: apis.Value{
							Type:      apis.ConstData,
							Name:      "NodeDependency",
							Value:     "1",
							ValueType: "string",
							From:      "R6",
						},
					},
				},
			},
			Devices: []apis.DeviceSpec{
				apis.DeviceSpec{
					Name: "Detector1",
					ExpectedProperties: map[string]apis.Property{
						"name": apis.Property{
							Value: "deviceGalaxea",
						},
					},
					Abilities: []string{
						"Detect",
					},
				},
			},
			Image: "Device{Detector1}.Ability{Detect}.Service{DetectPosition}",
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
		Status: apis.RuntimeStatus{
			Devices: map[string]apis.ObjectReference{
				"deviceGalaxea": apis.ObjectReference{
					Name:      "deviceGalaxea",
					Namespace: "test",
					Kind:      "Device",
				},
			},
		},
	}

	// 乐聚检测
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
			Conditions: &apis.Conditions{
				Formulas: []apis.ConditionFormula{
					{
						LeftValue: apis.Value{
							Type:      apis.ResultsData,
							Name:      "NodeDependency",
							Value:     "0",
							ValueType: "string",
							From:      "R5",
						},
						RightValue: apis.Value{
							Type:      apis.ConstData,
							Name:      "NodeDependency",
							Value:     "1",
							ValueType: "string",
							From:      "R5",
						},
					},
				},
			},
			Devices: []apis.DeviceSpec{
				apis.DeviceSpec{
					Name: "Detector2",
					ExpectedProperties: map[string]apis.Property{
						"name": apis.Property{
							Value: "deviceLeju",
						},
					},
					Abilities: []string{
						"Detect",
					},
				},
			},
			Image: "Device{Detector2}.Ability{Detect}.Service{DetectPosition}",
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
		Status: apis.RuntimeStatus{
			Devices: map[string]apis.ObjectReference{
				"deviceLeju": apis.ObjectReference{
					Name:      "deviceLeju",
					Namespace: "test",
					Kind:      "Device",
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
					Name: "deviceGalaxea",
					ExpectedProperties: map[string]apis.Property{
						"name": apis.Property{
							Value: "deviceGalaxea",
						},
					},
					Abilities: []string{
						"Grab",
					},
				},
			},
			Image: "Device{deviceGalaxea}.Ability{Grab}.Service{GrabBall}",
			Inputs: []apis.Value{
				apis.Value{
					Name: "worldPoints",
					Type: apis.LocalData,
					From: "Action{A1}.Runtime{R1}.Outputs{worldPoints}", // TODO
				},
			},
		},
		Status: apis.RuntimeStatus{
			Devices: map[string]apis.ObjectReference{
				"deviceGalaxea": apis.ObjectReference{
					Name:      "deviceGalaxea",
					Namespace: "test",
					Kind:      "Device",
				},
			},
		},
	}

	// 乐聚抓取
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
					Name: "deviceLeju",
					ExpectedProperties: map[string]apis.Property{
						"name": apis.Property{
							Value: "deviceLeju",
						},
					},
					Abilities: []string{
						"Grab",
					},
				},
			},
			Image: "Device{deviceLeju}.Ability{Grab}.Service{GrabBall}",
			Inputs: []apis.Value{
				apis.Value{
					Name: "worldPoints",
					Type: apis.LocalData,
					From: "Action{A2}.Runtime{R2}.Outputs{worldPoints}", // TODO
				},
			},
		},
		Status: apis.RuntimeStatus{
			Devices: map[string]apis.ObjectReference{
				"deviceLeju": apis.ObjectReference{
					Name:      "deviceLeju",
					Namespace: "test",
					Kind:      "Device",
				},
			},
		},
	}

	//乐聚下载
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
			Name: "R5",
			Type: apis.ByDevice,
			Devices: []apis.DeviceSpec{
				apis.DeviceSpec{
					Name: "Detector2",
					ExpectedProperties: map[string]apis.Property{
						"name": apis.Property{
							Value: "deviceLeju",
						},
					},
					Abilities: []string{
						"Detect",
					},
				},
			},
			Image: "Device{Detector2}.Ability{Detect}.Service{Download}",
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
		Status: apis.RuntimeStatus{
			Devices: map[string]apis.ObjectReference{
				"deviceLeju": apis.ObjectReference{
					Name:      "deviceLeju",
					Namespace: "test",
					Kind:      "Device",
				},
			},
		},
	}

	//星海图下载
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
					Name: "Detector1",
					ExpectedProperties: map[string]apis.Property{
						"name": apis.Property{
							Value: "deviceGalaxea",
						},
					},
					Abilities: []string{
						"Detect",
					},
				},
			},
			Image: "Device{Detector1}.Ability{Detect}.Service{Download}",
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
		Status: apis.RuntimeStatus{
			Devices: map[string]apis.ObjectReference{
				"deviceGalaxea": apis.ObjectReference{
					Name:      "deviceGalaxea",
					Namespace: "test",
					Kind:      "Device",
				},
			},
		},
	}

	// 星海图检测
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
				runtime6.Spec,
			},
		},
	}

	// 乐聚检测
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
				runtime5.Spec,
				runtime2.Spec,
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
			Conditions: &apis.Conditions{
				Formulas: []apis.ConditionFormula{
					{
						LeftValue: apis.Value{
							Type:      apis.ResultsData,
							Name:      "NodeDependency",
							Value:     "0",
							ValueType: "string",
							From:      "A1",
						},
						RightValue: apis.Value{
							Type:      apis.ConstData,
							Name:      "NodeDependency",
							Value:     "1",
							ValueType: "string",
							From:      "A1",
						},
					},
				},
			},
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
				runtime4.Spec,
			},
			Parents: []string{"A2"},
			Conditions: &apis.Conditions{
				Formulas: []apis.ConditionFormula{
					{
						LeftValue: apis.Value{
							Type:      apis.ResultsData,
							Name:      "NodeDependency",
							Value:     "0",
							ValueType: "string",
							From:      "A2",
						},
						RightValue: apis.Value{
							Type:      apis.ConstData,
							Name:      "NodeDependency",
							Value:     "1",
							ValueType: "string",
							From:      "A2",
						},
					},
				},
			},
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
			Desc: &apis.Description{
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
			Desc: &apis.Description{
				Docs: "乐聚检测和抓取",
			},
			Name: "G2",
			Actions: []apis.ActionSpec{
				action2.Spec, action4.Spec,
			},
			Replicas: []int32{0, 0},
			//Parents: []string{
			//	"G1",
			//},
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

	u := uuid.Must(uuid.NewV7())
	_, err = m.CreateTask(task1.Spec, nil, task1.Namespace, u.String(), "")
	if err != nil {
		logs.Errorf("[TEST] create task err:%v", err)
	}
}

func TestLockDevice(t *testing.T) {
	cs, err := utils.CreateClientSetWithTimeOut(2000)
	if err != nil {
		panic(err)
	}
	m := manager.NewManager(cs)
	// 创建Device
	deviceLock := &apis.Device{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "deviceLock",
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
			Name: "deviceLock",
			AccessMethod: &apis.AccessMethod{
				Type: apis.AccessByAbility,
				URL:  "http://192.168.8.197:8080",
			},
			Abilities: []string{
				"Test1", "Test2",
			},
		},
		Status: apis.DeviceStatus{
			Abilities: map[string]apis.Ability{
				"Test1": {
					Name: "test",
					Services: map[string]apis.AbilityService{
						"test": apis.AbilityService{
							Ip:        new(string),
							Interface: new(string),
							Port:      new(string),
						},
					},
					Status: apis.AbilityRunning,
					Lock: apis.Lock{
						Ref:      1,
						IsLocked: false,
					},
				},
				"Test2": {
					Name: "test",
					Services: map[string]apis.AbilityService{
						"test": apis.AbilityService{
							Ip:        new(string),
							Interface: new(string),
							Port:      new(string),
						},
					},
					Status: apis.AbilityRunning,
					Lock: apis.Lock{
						Ref:      2,
						IsLocked: false,
					},
				},
			},
			Lock: apis.Lock{
				IsLocked: true,
				Ref:      3,
			},
			Phase: apis.DeviceIdle,
		},
	}
	*deviceLock.Status.Abilities["Test1"].Services["test"].Interface = "/api/task/detect"
	*deviceLock.Status.Abilities["Test1"].Services["test"].Ip = "192.168.8.197"
	*deviceLock.Status.Abilities["Test1"].Services["test"].Port = "22"
	*deviceLock.Status.Abilities["Test2"].Services["test"].Interface = "/api/task/detect"
	*deviceLock.Status.Abilities["Test2"].Services["test"].Ip = "192.168.8.197"
	*deviceLock.Status.Abilities["Test2"].Services["test"].Port = "22"

	_, err = m.CreateDevice(deviceLock, "test")
	if err != nil {
		logs.Errorf("[TEST] Create Device[%s] err:%s", deviceLock.Name, err.Error())
	}

}

// go test -run TestCreateWorkFlow -v
func TestCreateLockTest(t *testing.T) {
	cs, err := utils.CreateClientSetWithTimeOut(2000)
	if err != nil {
		panic(err)
	}
	m := manager.NewManager(cs)

	// 星海图检测
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
					Name: "deviceLock",
					ExpectedProperties: map[string]apis.Property{
						"name": apis.Property{
							Value: "deviceLock",
						},
					},
					Abilities: []string{
						"Test1",
					},
				},
			},
			Image: "Device{deviceLock}.Ability{Test1}.Service{test}",
		},
	}

	// 乐聚检测
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

			Devices: []apis.DeviceSpec{
				apis.DeviceSpec{
					Name: "deviceLock",
					ExpectedProperties: map[string]apis.Property{
						"name": apis.Property{
							Value: "deviceLock",
						},
					},
					Abilities: []string{
						"Test2",
					},
				},
			},
			Image: "Device{deviceLock}.Ability{Test2}.Service{test}",
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
					Name: "deviceLock",
					ExpectedProperties: map[string]apis.Property{
						"name": apis.Property{
							Value: "deviceLock",
						},
					},
					Abilities: []string{
						"test2",
					},
				},
			},
			Image: "Device{deviceLock}.Ability{Test2}.Service{test}",
		},
	}

	// 星海图检测
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
				Docs: "启始任务",
			},
			Name: "A1",
			Runtimes: []apis.RuntimeSpec{
				runtime1.Spec,
			},
		},
	}

	// 乐聚检测
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
				Docs: "并行任务1",
			},
			Name: "A2",
			Runtimes: []apis.RuntimeSpec{
				runtime2.Spec,
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
				Docs: "并行任务2",
			},
			Name: "A3",
			Runtimes: []apis.RuntimeSpec{
				runtime3.Spec,
			},
			Parents: []string{"A1"},
			//Conditions: &apis.Conditions{
			//	Formulas: []apis.ConditionFormula{
			//		{
			//			LeftValue: apis.Value{
			//				Type:      apis.ResultsData,
			//				Name:      "NodeDependency",
			//				Value:     "0",
			//				ValueType: "string",
			//				From:      "A1",
			//			},
			//			RightValue: apis.Value{
			//				Type:      apis.ConstData,
			//				Name:      "NodeDependency",
			//				Value:     "1",
			//				ValueType: "string",
			//				From:      "A1",
			//			},
			//		},
			//	},
			//},
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
			Desc: &apis.Description{
				Docs: "测试Group",
			},
			Name: "G1",
			Actions: []apis.ActionSpec{
				action1.Spec, action2.Spec, action3.Spec,
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
				Docs: "测试Task",
			},
			Name: "T1",
			Groups: []apis.GroupSpec{
				group1.Spec,
			},
		},
	}

	u := uuid.Must(uuid.NewV7())
	_, err = m.CreateTask(task1.Spec, nil, task1.Namespace, u.String(), "")
	if err != nil {
		logs.Errorf("[TEST] create task err:%v", err)
	}
}

// go test -run TestSendScene1ToProxy -v
func TestSendScene1ToProxy(t *testing.T) {
	logs.Init("testModule")
	task := createOrangeTask()
	client := &http.Client{}

	taskBytes, err := json.Marshal(task)

	url := "http://192.168.8.176:8899/framework/v1/task?Name=T1&&Namesapce=test"
	logs.Info(url)
	req, err := http.NewRequest("POST", url, strings.NewReader(string(taskBytes)))
	if err != nil {
		logs.Fatal(err)
	}
	//Content-Type很重要，下文解释
	//req.Header.Set("Content-Type", "application/x-www")
	req.Header.Set("Content-Type", "application/json")
	//req.Header.Set("Content-Type", "multipart/form-data")

	rep, err := client.Do(req)
	if err != nil {
		logs.Fatal(err.Error())
	}
	data, err := io.ReadAll(rep.Body)
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			logs.Error(err)
		}
	}(rep.Body)
	if err != nil {
		logs.Fatal(err)
	}
	logs.Infof("resp is : %s", string(data))
}

// go test -run TestAddDevice -v
func TestAddDevice(t *testing.T) {
	logs.Init("testModule")
	ctx := context.Background()
	cs, err := createClientSet()
	if err != nil {
		logs.Error(err)
		return
	}
	dc := cs.Core().Nodes(apis.NamespaceTest)
	node := apis.Node{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Node",
			APIVersion: "resources/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "TestEnvNode1",
			Namespace: apis.NamespaceTest,
		},
		Spec: apis.NodeSpec{
			NodeName: "TestEnvNode1",
			HostName: "192.168.8.176",
		},
		Status: apis.NodeStatus{},
	}
	_, err = dc.Create(ctx, &node, metav1.CreateOptions{})
	if err != nil {
		logs.Error(err.Error())
		return
	}

}

func TestCreateJson1(t *testing.T) {
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
						"name": apis.Property{
							Value: "device1",
						},
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
			Name: "R1",
			Type: apis.ByDevice,
			Inputs: []apis.Value{
				{
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
						"name": apis.Property{
							Value: "device1",
						},
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
						"name": apis.Property{
							Value: "device1",
						},
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
						"name": apis.Property{
							Value: "device1",
						},
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
			Name: "R4",
			Type: apis.ByDevice,
			Inputs: []apis.Value{
				{
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
						"name": apis.Property{
							Value: "device1",
						},
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
			Name: "R5",
			Type: apis.ByDevice,
			Inputs: []apis.Value{
				{
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
						"name": apis.Property{
							Value: "device1",
						},
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
			Name:    "R6",
			Type:    apis.ByCommand,
			Command: []string{"python3"}, // 801的机器 要用python3
			Args: []string{"" +
				"/home/smj/adaptive-scheduling-framework/test/scene3/recheck1.py",
				"/home/smj/adaptive-scheduling-framework/test/scene3/data.txt",
				"device1"},
			// 这个是smj
			//Args: []string{"" +
			//	"/home/public/lock_test/test/scene3/recheck1.py",
			//	"/home/public/lock_test/test/scene3/data.txt",
			//	"device1"},
			// 这个是801的路径
			Parents: make([]string, 0),
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
						"name": apis.Property{
							Value: "device1",
						},
					},
					Abilities: []string{
						"Turn",
					},
				},
			},
		},
	}

	// 星海图检测
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

	// 乐聚检测
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
				Docs: "检测/提交复检任务",
			},
			Name: "A3",
			Runtimes: []apis.RuntimeSpec{
				runtime3.Spec, runtime4.Spec, runtime5.Spec, runtime6.Spec,
			},
			Parents: []string{"A2"},
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
					Abilities: []string{
						"Turn", "Grab", "Detect", "Put",
					},
				},
			},
			Desc: &apis.Description{
				Docs: "场景三的初检group",
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
	b, _ := json.Marshal(task1)
	fmt.Printf("%s", string(b))
}

// 复检
func TestCreateJson2(t *testing.T) {

	// 转向table
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
					Type:      apis.ConstData,
					Value:     "table",
					ValueType: apis.StringType,
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

	// 初始化
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
			Inputs: []apis.Value{
				{
					Type:      apis.ConstData,
					Value:     "up",
					ValueType: apis.StringType,
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

	// 抓取
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
			Inputs: []apis.Value{
				{
					Type:      apis.ConstData,
					Value:     "table",
					ValueType: apis.StringType,
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
			Name:    "R4",
			Type:    apis.ByDevice,
			Inputs: []apis.Value{
				{
					Type:      apis.ConstData,
					Value:     "finished_bin",
					ValueType: apis.StringType,
				},
			},
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

	// 放到废品框
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
					Type:      apis.ConstData,
					Value:     "reject_bin",
					ValueType: apis.StringType,
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
			Inputs: []apis.Value{
				{
					Type:      apis.ConstData,
					Value:     "belt",
					ValueType: apis.StringType,
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

	// 星海图检测
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
				Docs: "转身",
			},
			Name: "A1",
			Runtimes: []apis.RuntimeSpec{
				runtime0.Spec,
			},
		},
	}

	// 乐聚检测
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
				Docs: "抓取物品",
			},
			Name: "A2",
			Runtimes: []apis.RuntimeSpec{
				runtime1.Spec, runtime2.Spec,
			},
			Parents: []string{"A1"},
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
				Docs: "检测",
			},
			Name: "A3",
			Runtimes: []apis.RuntimeSpec{
				runtime3.Spec, runtime4.Spec, runtime5.Spec,
			},
			Parents: []string{"A2"},
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
				Docs: "转向流水线",
			},
			Name: "A4",
			Runtimes: []apis.RuntimeSpec{
				runtime6.Spec,
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
					Abilities: []string{
						"Turn", "Grab", "Detect", "Put",
					},
					ExpectedProperties: map[string]apis.Property{
						"name": apis.Property{
							Value: "{which_device}",
						},
						"strategy": apis.Property{
							Value: "nominate",
						},
					},
				},
			},
			Desc: &apis.Description{
				Docs: "场景三的复检group",
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
			Name:      "T2",
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
				Docs: "复检task",
			},
			Name: "T2",
			Groups: []apis.GroupSpec{
				group1.Spec,
			},
		},
	}
	b, _ := json.Marshal(task1)
	fmt.Printf("%s", string(b))
}

func TestCreateRecheck(t *testing.T) {
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
			Name:    "R1",
			Type:    apis.ByCommand,
			Command: []string{"python"},
			Args: []string{"" +
				"/home/weland/workspace/Resource-Based-Framework/adaptive-scheduling-framework/test/scene3/recheck1.py",
				"/home/weland/workspace/Resource-Based-Framework/adaptive-scheduling-framework/test/scene3/data.txt",
				"device1"},
			Parents: make([]string, 0),
		},
	}

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
				Docs: "检测",
			},
			Name: "A1",
			Runtimes: []apis.RuntimeSpec{
				runtime1.Spec,
			},
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
			Desc: &apis.Description{
				Docs: "场景三的初检group",
			},
			Name: "G1",
			Actions: []apis.ActionSpec{
				action1.Spec,
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
				group1.Spec,
			},
		},
	}
	clientSet, _ := device.InitClient()
	m := manager.NewManager(clientSet)
	timestamp := time.Now().Format("20060102T150405")
	randomStr := uuid.New().String()[:5]
	UUID := timestamp + "-" + randomStr
	m.CreateTask(task1.Spec, nil, task1.Namespace, UUID, "")
}

func TestCreateTwoTaskForLock(t *testing.T) {
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
					Name: "DEVICE_A",
					ExpectedProperties: map[string]apis.Property{
						"name": apis.Property{},
					},
					Abilities: []string{
						"Test",
					},
				},
			},
			Image: "Device{DEVICE_A}.Ability{Test}.Service{test}",
		},
	}

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
			Name:    "R2",
			Type:    apis.ByCommand,
			Command: []string{"python3"}, // 801的机器 要用python3
			Args: []string{"" +
				"/home/smj/adaptive-scheduling-framework/test/scene3/recheck1.py",
				"/home/smj/adaptive-scheduling-framework/test/scene3/data.txt",
				"device1"},
			// 这个是smj
			//Args: []string{"" +
			//	"/home/public/lock_test/test/scene3/recheck1.py",
			//	"/home/public/lock_test/test/scene3/data.txt",
			//	"device1"},
			// 这个是801的路径
			Parents: make([]string, 0),
		},
	}

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
				Docs: "",
			},
			Name: "A1",
			Runtimes: []apis.RuntimeSpec{
				runtime1.Spec, runtime2.Spec,
			},
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
					Name: "DEVICE_A",
					Abilities: []string{
						"Test",
					},
					ExpectedProperties: map[string]apis.Property{
						"name": apis.Property{
							Value: "device1",
						},
						"strategy": apis.Property{
							Value: "nominate",
						},
					},
				},
			},
			Desc: &apis.Description{
				Docs: "测试Group",
			},
			Name: "G1",
			Actions: []apis.ActionSpec{
				action1.Spec,
			},
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
					Name: "DEVICE_A",
					Abilities: []string{
						"Test",
					},
					ExpectedProperties: map[string]apis.Property{
						"name": apis.Property{
							Value: "device1",
						},
						"strategy": apis.Property{
							Value: "nominate",
						},
					},
				},
			},
			Desc: &apis.Description{
				Docs: "测试Group",
			},
			Name: "G1",
			Actions: []apis.ActionSpec{
				action1.Spec,
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
				Docs: "test",
			},
			Name: "T1",
			Groups: []apis.GroupSpec{
				group1.Spec,
			},
		},
	}

	task2 := &apis.Task{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "T2",
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
				Docs: "test",
			},
			Name: "T2",
			Groups: []apis.GroupSpec{
				group2.Spec,
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

	timestamp1 := time.Now().Format("20060102T150405")
	randomStr1 := uuid.New().String()[:5]
	UUID1 := timestamp1 + "-" + randomStr1
	_, err = m.CreateTask(task2.Spec, nil, apis.NamespaceTest, UUID1, "")
	if err != nil {
		logs.Errorf("create task2 fail")
	}
}

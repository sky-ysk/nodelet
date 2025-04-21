package main

import (
	"github.com/google/uuid"
	"hit.edu/framework/pkg/apimachinery/runtime"
	"hit.edu/framework/pkg/apimachinery/runtime/schema"
	"hit.edu/framework/pkg/apimachinery/runtime/serializer"
	apis "hit.edu/framework/pkg/apis/cores"
	metav1 "hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/client-go/clients"
	"hit.edu/framework/pkg/client-go/rest"
	"hit.edu/framework/pkg/client-go/util/manager"
	"hit.edu/framework/pkg/component-base/logs"
	"net/http"
	"time"
)

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
		logs.Errorf("[TEST] Create clientSet err:%v", err.Error())
		return
	}
	u := uuid.Must(uuid.NewV7())
	m := manager.NewManager(clientSet)
	task, err := m.CreateTask(ts, nil, "test", u.String(), "")

}

func CreateWorkFlow(m *manager.Manager) {
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
					Status: apis.AbilityRunning,
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
					Status: apis.AbilityRunning,
				},
			},
			Lock: apis.Lock{
				IsLocked: true,
				Ref:      0,
			},
			Phase: apis.DeviceIdle,
		},
	}
	*deviceGalaxea.Status.Abilities["Detect"].Services["DetectPosition"].Interface = "/api/task/detect"
	*deviceGalaxea.Status.Abilities["Detect"].Services["DetectPosition"].Ip = "192.168.8.197"
	*deviceGalaxea.Status.Abilities["Detect"].Services["DetectPosition"].Port = "" // 填写这个端口
	*deviceGalaxea.Status.Abilities["Grab"].Services["GrabBall"].Interface = "/api/task/grab_ball"
	*deviceGalaxea.Status.Abilities["Grab"].Services["GrabBall"].Ip = "192.168.8.197"
	*deviceGalaxea.Status.Abilities["Grab"].Services["GrabBall"].Port = ""

	_, err := m.CreateDevice(deviceGalaxea, "test")
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
				URL:  "http://192.168.8.197:8080",
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
					},
					Status: apis.AbilityRunning,
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
					Status: apis.AbilityRunning,
				},
			},
			Lock: apis.Lock{
				IsLocked: true,
				Ref:      1,
			},
			Phase: apis.DeviceIdle,
		},
	}
	*deviceLeju.Status.Abilities["Detect"].Services["DetectPosition"].Interface = "/api/task/detect"
	*deviceLeju.Status.Abilities["Detect"].Services["DetectPosition"].Ip = "192.168.8.165"
	*deviceLeju.Status.Abilities["Detect"].Services["DetectPosition"].Port = "" // 填写这个端口
	*deviceLeju.Status.Abilities["Grab"].Services["GrabBall"].Interface = "/api/task/grab_ball"
	*deviceLeju.Status.Abilities["Grab"].Services["GrabBall"].Ip = "192.168.8.165"
	*deviceLeju.Status.Abilities["Grab"].Services["GrabBall"].Port = ""
	_, err = m.CreateDevice(deviceLeju, "test")
	if err != nil {
		logs.Errorf("[TEST] Create Device[%s] err:%s", deviceLeju.Name, err.Error())
	}

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
					Name: "deviceGalaxea",
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
			Image: "Device{deviceGalaxea}.Ability{Detect}.Skill{DetectPosition}",
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
			Devices: []apis.DeviceSpec{
				apis.DeviceSpec{
					Name: "deviceLeju",
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
			Image: "Device{deviceLeju}.Ability{Detect}.Skill{DetectPosition}",
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
			Image: "Device{deviceGalaxea}.Ability{Grab}.Skill{GrabBall}",
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
			Image: "Device{deviceLeju}.Ability{Grab}.Skill{GrabBall}",
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
			Name: "A1",
			Runtimes: []apis.RuntimeSpec{
				runtime1.Spec,
			},
		},
		Status: apis.ActionStatus{
			Runtimes: map[string]apis.ObjectReference{
				"R1": apis.ObjectReference{
					Name:      "R1",
					Namespace: "test",
					Kind:      "Runtime",
				},
			},
		},
	}

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
			Name: "A2",
			Runtimes: []apis.RuntimeSpec{
				runtime2.Spec,
			},
		},
		Status: apis.ActionStatus{
			Runtimes: map[string]apis.ObjectReference{
				"R2": apis.ObjectReference{
					Name:      "R2",
					Namespace: "test",
					Kind:      "Runtime",
				},
			},
		},
	}

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
			Name: "A3",
			Runtimes: []apis.RuntimeSpec{
				runtime3.Spec,
			},
		},
		Status: apis.ActionStatus{
			Runtimes: map[string]apis.ObjectReference{
				"R3": apis.ObjectReference{
					Name:      "R3",
					Namespace: "test",
					Kind:      "Runtime",
				},
			},
		},
	}

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
			Name: "A4",
			Runtimes: []apis.RuntimeSpec{
				runtime4.Spec,
			},
		},
		Status: apis.ActionStatus{
			Runtimes: map[string]apis.ObjectReference{
				"R4": apis.ObjectReference{
					Name:      "R4",
					Namespace: "test",
					Kind:      "Runtime",
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
			Kind:       "Action",
			APIVersion: "resources/v1",
		},
		Spec: apis.GroupSpec{
			Name: "G1",
			Actions: []apis.ActionSpec{
				action1.Spec, action2.Spec, action3.Spec, action4.Spec,
			},
		},
		Status: apis.GroupStatus{
			Actions: map[string]apis.ObjectReference{
				"A1": apis.ObjectReference{
					Name:      "A1",
					Namespace: "test",
					Kind:      "Action",
				},
				"A2": apis.ObjectReference{
					Name:      "A2",
					Namespace: "test",
					Kind:      "Action",
				},
				"A3": apis.ObjectReference{
					Name:      "A3",
					Namespace: "test",
					Kind:      "Action",
				},
				"A4": apis.ObjectReference{
					Name:      "A4",
					Namespace: "test",
					Kind:      "Action",
				},
			},
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
			Name: "T1",
			Groups: []apis.GroupSpec{
				group1.Spec,
			},
		},
		Status: apis.TaskStatus{
			Groups: map[string]apis.ObjectReference{
				"G1": apis.ObjectReference{
					Name:      "G1",
					Namespace: "test",
					Kind:      "Task",
				},
			},
		},
	}
	u := uuid.Must(uuid.NewV7())
	_, err = m.CreateRuntime(runtime1.Spec, action1, runtime1.Namespace, u.String(), "")
	if err != nil {
		logs.Errorf("[TEST] create runtime err:%v", err)
	}

	_, err = m.CreateRuntime(runtime2.Spec, action2, runtime2.Namespace, u.String(), "")
	if err != nil {
		logs.Errorf("[TEST] create runtime err:%v", err)
	}

	_, err = m.CreateRuntime(runtime3.Spec, action3, runtime3.Namespace, u.String(), "")
	if err != nil {
		logs.Errorf("[TEST] create runtime err:%v", err)
	}

	_, err = m.CreateRuntime(runtime4.Spec, action4, runtime4.Namespace, u.String(), "")
	if err != nil {
		logs.Errorf("[TEST] create runtime err:%v", err)
	}

	_, err = m.CreateAction(action1.Spec, group1, action1.Namespace, u.String(), "")
	if err != nil {
		logs.Errorf("[TEST] create action err:%v", err)
	}
	_, err = m.CreateAction(action2.Spec, group1, action2.Namespace, u.String(), "")
	if err != nil {
		logs.Errorf("[TEST] create action err:%v", err)
	}
	_, err = m.CreateAction(action3.Spec, group1, action3.Namespace, u.String(), "")
	if err != nil {
		logs.Errorf("[TEST] create action err:%v", err)
	}
	_, err = m.CreateAction(action4.Spec, group1, action4.Namespace, u.String(), "")
	if err != nil {
		logs.Errorf("[TEST] create action err:%v", err)
	}
	_, err = m.CreateGroup(group1.Spec, task1, task1.Namespace, u.String(), "")
	if err != nil {
		logs.Errorf("[TEST] create group err:%v", err)
	}
	_, err = m.CreateTask(task1.Spec, nil, task1.Namespace, u.String(), "")
	if err != nil {
		logs.Errorf("[TEST] create task err:%v", err)
	}
}

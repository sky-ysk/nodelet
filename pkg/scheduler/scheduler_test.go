package scheduler

import (
	"context"
	"encoding/json"
	"fmt"
	"hit.edu/framework/pkg/apimachinery/runtime"
	"hit.edu/framework/pkg/apimachinery/runtime/schema"
	"hit.edu/framework/pkg/apimachinery/runtime/serializer"
	apis "hit.edu/framework/pkg/apis/cores"
	metav1 "hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/client-go/clients"
	"hit.edu/framework/pkg/client-go/rest"
	"hit.edu/framework/pkg/component-base/logs"
	"hit.edu/framework/pkg/nodelet/task/runtime/device/ability/manager"
	"hit.edu/framework/pkg/scheduler/transport"
	"net/http"
	"testing"
	"time"
)

//测试调度框架

//func TestScheduler_Run(t *testing.T) {
//
//	//stopEverything := ctx.Done()
//
//	// 配置调度器启动选项，在这里需要定义所需的模块，插件
//	//options := defaultSchedulerOptions
//	//for _, opt := range opts {
//	//	opt(&options)
//	//}
//
//	// 配置调度器参数
//
//	// 配置插件模块
//	//registry := plugins.NewInTreeRegistry()
//
//	// 配置任务队列git
//
//	// 配置资源监控模块
//	//scheduleChan := make(chan internal.ScheduleSignal)
//	queue := queue.NewPriorityQueue()
//	sched := &Scheduler{
//		//StopEverything:  stopEverything,
//		//ScheduleSigChan: scheduleChan,
//		SchedulingQueue: queue,
//	}
//
//	//schedQueue := &queue.PriorityQueue{}
//
//	sched.applyDefaultHandlers()
//	sched.ReadyGroup = sched.SchedulingQueue.Pop
//
//	//return sched, nil
//}

func TestSendGroupToScheduler(t *testing.T) {
	scheme := runtime.NewScheme()
	apis.AddToScheme(scheme)
	//fmt.Println(scheme)
	//参数配置
	// TODO: 填写参数
	//部分参数之后可以在core_client等 编写setConfigDefaults函数进行填充

	c := &rest.Config{
		Host:    "http://localhost:10000",
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
	// 资源定义在 pkg/apis/xxx/type.go 下
	// 这里以访问资源Group为例，
	// 获取访问Group的客户端
	// 默认访问的Namespace是 ""

	groupsClient := clientSet.Core().Groups("")

	group1 := &apis.Group{
		ObjectMeta: metav1.ObjectMeta{
			Name: "demo-group1",
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Group",
			APIVersion: "resources/v1",
		},
	}

	group2 := &apis.Group{
		ObjectMeta: metav1.ObjectMeta{
			Name: "demo-group2",
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Group",
			APIVersion: "resources/v1",
		},
	}

	result1, err := groupsClient.Create(context.TODO(), group1, metav1.CreateOptions{})
	if err != nil {
		panic(err)
	}
	fmt.Println("Created group ", result1)
	result2, err := groupsClient.Create(context.TODO(), group2, metav1.CreateOptions{})
	if err != nil {
		panic(err)
	}
	fmt.Println("Created group ", result2)
}

func TestCreateNode(t *testing.T) {
	scheme := runtime.NewScheme()
	apis.AddToScheme(scheme)
	fmt.Println(scheme)
	//参数配置
	// TODO: 填写参数
	//部分参数之后可以在core_client等 编写setConfigDefaults函数进行填充
	c := &rest.Config{
		Host:    "http://localhost:10000",
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
	// 资源定义在 pkg/apis/xxx/type.go 下
	// 这里以访问资源Node为例，
	// 获取访问Node的客户端
	// 默认访问的Namespace是 ""

	nodesClient := clientSet.Core().Nodes("")
	node1 := &apis.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "demo-nodes1",
			Namespace: "",
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Node",
			APIVersion: "resources/v1",
		},
		Spec: apis.NodeSpec{
			NodeName: "demo-node1",
		},
	}
	node2 := &apis.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "demo-nodes2",
			Namespace: "",
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Node",
			APIVersion: "resources/v1",
		},
		Spec: apis.NodeSpec{
			NodeName: "demo-node2",
		},
	}
	fmt.Println("creating")
	_, err = nodesClient.Create(context.TODO(), node1, metav1.CreateOptions{})

	if err != nil {
		logs.Errorf("Failed to create node: %v", err)
		panic(err)
	}
	fmt.Println("creating")
	_, err = nodesClient.Create(context.TODO(), node2, metav1.CreateOptions{})
	if err != nil {
		logs.Errorf("Failed to create node: %v", err)
		panic(err)
	}
}

func TestListGroup(t *testing.T) {
	clientSet, err := createClientSet()
	if err != nil {
		logs.Error(err)
		return
	}
	groupsClient := clientSet.Core().Groups("")

	lstOpts := metav1.ListOptions{
		//FieldSelector: "ObjectMeta.Name=demo-groups",
	}
	list, err := groupsClient.List(context.TODO(), lstOpts)
	if err != nil {
		panic(err)
	}
	for _, d := range list.Items {
		fmt.Println(d.ObjectMeta.Name)
		fmt.Println(d.Status.Phase)
		fmt.Println(d.Status.Node)
	}
}

// 清空etcd 里面的groups Actions devices events
// go test -run TestClearEtcd -v
func TestClearEtcd(t *testing.T) {
	moduleName := "testModule"
	logs.Init(moduleName)
	ctx := context.Background()
	cs, err := createClientSet()
	if err != nil {
		logs.Error(err)
		return
	}

	//删group
	groupClient := cs.Core().Groups(apis.NamespaceTest)
	groups, err := groupClient.List(ctx, metav1.ListOptions{})
	if err != nil {
		return
	}
	for _, g := range groups.Items {
		logs.Infof("delete group %s ", g.Name)
		err := groupClient.Delete(ctx, g.Name, metav1.DeleteOptions{})
		if err != nil {
			logs.Error(err.Error())
			return
		}
	}

	//删actions
	actionClient := cs.Core().Actions(apis.NamespaceTest)
	acts, err := actionClient.List(ctx, metav1.ListOptions{})
	if err != nil {
		return
	}
	for _, a := range acts.Items {
		logs.Infof("delete act %s ", a.Name)
		err := actionClient.Delete(ctx, a.Name, metav1.DeleteOptions{})
		if err != nil {
			logs.Error(err)
			return
		}
	}

	////删devices
	//deviceClient := cs.Core().Devices("test")
	//devices, err := deviceClient.List(ctx, metav1.ListOptions{})
	//if err != nil {
	//	return
	//}
	//for _, d := range devices.Items {
	//	logs.Infof("delete act %s ", d.Name)
	//	err := deviceClient.Delete(ctx, d.Spec.Name, metav1.DeleteOptions{})
	//	if err != nil {
	//		logs.Error(err)
	//		return
	//	}
	//}

	//删events
	eventClient := cs.Core().Events(apis.NamespaceTest)
	events, err := eventClient.List(ctx, metav1.ListOptions{})
	if err != nil {
		return
	}
	for _, e := range events.Items {
		logs.Infof("delete event %s ", e.Name)
		err := eventClient.Delete(ctx, e.Name, metav1.DeleteOptions{})
		if err != nil {
			logs.Error(err)
			return
		}
	}

	//删task
	taskClient := cs.Core().Tasks(apis.NamespaceTest)
	tasks, err := taskClient.List(ctx, metav1.ListOptions{})
	if err != nil {
		return
	}
	for _, t := range tasks.Items {
		logs.Infof("delete task %s ", t.Name)
		err := taskClient.Delete(ctx, t.Name, metav1.DeleteOptions{})
		if err != nil {
			logs.Error(err)
			return
		}
	}
}

func createClientSet() (*clients.ClientSet, error) {
	scheme := runtime.NewScheme()
	apis.AddToScheme(scheme)

	c := &rest.Config{
		Host:    "http://localhost:10000",
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
	return clients.NewForConfig(c)
}

//func CreateOrangeGroupActionRuntimePredict() (*apis.Group, *apis.Action, *apis.Runtime, *apis.Device) {
//
//	// 能力框架的url
//	manageUrl := "http://192.168.8.165:8080"
//	imageType := "rgb"
//	cameraUrl := "http://127.0.0.1:54533/api/status/camera"
//	position := "head"
//	compressed := false
//	path := ""
//	// 创建device
//	device := &apis.Device{
//		ObjectMeta: metav1.ObjectMeta{
//			Name:      "devicePredict",
//			Namespace: "test",
//			Labels: map[string]string{
//				"environment": "dev",
//			},
//		},
//		TypeMeta: metav1.TypeMeta{
//			Kind:       "Device",
//			APIVersion: "resources/v1",
//		},
//		Spec: apis.DeviceSpec{
//			Name: "devicePredict",
//			AccessMethod: apis.AccessMethod{
//				Type: apis.AccessByAbility,
//				URL:  manageUrl,
//			},
//			ExpectedProperties: map[string]apis.Property{},
//			Abilities:          make([]apis.AbilitySpec, 0),
//		},
//		Status: apis.DeviceStatus{
//			DeviceID: "devicePredict",
//			Phase:    apis.DeviceIdle,
//			Status:   "idle",
//			ActionID: "",
//			Lock: apis.Lock{
//				Lock: true,
//			},
//			Abilities: make([]apis.AbilityStatus, 0),
//		},
//	}
//
//	device.Status.Abilities = append(device.Status.Abilities, apis.AbilityStatus{
//		Name: "Predict",
//		Services: []apis.AbilityServiceStatus{
//			{
//				Name:      "Predict",
//				Ip:        "192.168.8.165",
//				Interface: "/predict",
//			},
//			{
//				Name:      "PredictByUrl",
//				Ip:        "192.168.8.165",
//				Interface: "/predict_by_url",
//			},
//		},
//	})
//
//	runtime1 := &apis.Runtime{
//		Type:  apis.ByDevice,
//		Image: "manage_Detect",
//		Name:  "RuntimeTest",
//		Devices: []apis.DeviceSpec{
//			device.Spec,
//		},
//		Outputs: make([]apis.Output, 1),
//		Inputs: []apis.Input{
//			{
//				Name:  "imageType",
//				Type:  "string",
//				Value: imageType,
//			},
//			{
//				Name:  "position",
//				Type:  "string",
//				Value: position,
//			},
//			{
//				Name:  "cameraUrl",
//				Type:  "string",
//				Value: cameraUrl,
//			},
//			{
//				Name:  "compressed",
//				Type:  "bool",
//				Value: strconv.FormatBool(compressed),
//			},
//			{
//				Name:  "path",
//				Type:  "string",
//				Value: path,
//			},
//		},
//	}
//
//	runtime2 := &apis.Runtime{
//		Type:  apis.ByDevice,
//		Image: "service_PredictByUrl",
//		Name:  "RuntimeTest",
//		Devices: []apis.DeviceSpec{
//			device.Spec,
//		},
//		Outputs: make([]apis.Output, 1),
//		Inputs: []apis.Input{
//			{
//				Name:  "imageType",
//				Type:  "string",
//				Value: imageType,
//			},
//			{
//				Name:  "position",
//				Type:  "string",
//				Value: position,
//			},
//			{
//				Name:  "cameraUrl",
//				Type:  "string",
//				Value: cameraUrl,
//			},
//			{
//				Name:  "compressed",
//				Type:  "bool",
//				Value: strconv.FormatBool(compressed),
//			},
//			{
//				Name:  "path",
//				Type:  "string",
//				Value: path,
//			},
//		},
//	}
//
//	action1 := &apis.Action{
//		ObjectMeta: metav1.ObjectMeta{
//			Name:      "action1p",
//			Namespace: "test",
//			Labels: map[string]string{
//				"environment": "dev",
//			},
//		},
//		TypeMeta: metav1.TypeMeta{
//			Kind:       "Action",
//			APIVersion: "resources/v1",
//		},
//		Spec: apis.ActionSpec{
//			Name: "action1p",
//			Runtimes: []apis.Runtime{
//				*runtime1,
//			},
//		},
//		Status: apis.ActionStatus{
//			ActionID:      "Action1p",
//			Devices:       make(map[string]apis.DeviceStatus),
//			RuntimeStatus: make([]apis.RuntimeStatus, 1),
//		},
//	}
//	action2 := &apis.Action{
//		ObjectMeta: metav1.ObjectMeta{
//			Name:      "action2p",
//			Namespace: "test",
//			Labels: map[string]string{
//				"environment": "dev",
//			},
//		},
//		TypeMeta: metav1.TypeMeta{
//			Kind:       "Action",
//			APIVersion: "resources/v1",
//		},
//		Spec: apis.ActionSpec{
//			Name: "action2p",
//			Runtimes: []apis.Runtime{
//				*runtime2,
//			},
//			Conditions: apis.Conditions{
//				Formulas: []apis.ConditionFormula{
//					apis.ConditionFormula{
//						LeftValue: apis.ConditionValue{
//							Type:      apis.ResultsData,
//							Name:      "NodeDependency",
//							Value:     "0",
//							ValueType: "string",
//							From:      action1.Name,
//						},
//						RightValue: apis.ConditionValue{
//							Type:      apis.ConstData,
//							Name:      "NodeDependency",
//							Value:     "1",
//							ValueType: "string",
//							From:      "",
//						},
//						Signal: apis.Equal,
//						Join:   "",
//						Result: apis.False,
//					},
//				},
//			},
//		},
//		Status: apis.ActionStatus{
//			ActionID:      "Action1p",
//			Devices:       make(map[string]apis.DeviceStatus),
//			RuntimeStatus: make([]apis.RuntimeStatus, 1),
//		},
//	}
//
//	action1.Status.Devices["devicePredict"] = device.Status
//	action2.Status.Devices["devicePredict"] = device.Status
//	group := &apis.Group{
//		ObjectMeta: metav1.ObjectMeta{
//			Name:      "groupp",
//			Namespace: "test",
//		},
//		TypeMeta: metav1.TypeMeta{
//			Kind:       "Group",
//			APIVersion: "resources/v1",
//		},
//		Spec: apis.GroupSpec{
//			Name:     "groupp",
//			Actions:  make([]apis.Action, 2),
//			Replicas: []int32{0, 0},
//		},
//		Status: apis.GroupStatus{
//			Node:    "test-node",
//			GroupID: "groupp",
//			Phase:   apis.Pending,
//			ActionStatus: []apis.ActionStatus{
//				action1.Status,
//				action2.Status,
//			},
//		},
//	}
//	group.Spec.Actions[0] = *action1
//	group.Spec.Actions[1] = *action2
//	return group, action1, runtime1, device
//}

//func CreateOrangeGroupActionRuntimeArm() (*apis.Group, *apis.Action, *apis.Runtime, *apis.Device) {
//
//	// 能力框架的url
//	manageUrl := "http://192.168.8.165:8080"
//	left := "-1.221,0.0872,0,0,0,0,0"
//	right := "-0,0,0,0,0,0,0"
//	// 创建device
//	device := &apis.Device{
//		ObjectMeta: metav1.ObjectMeta{
//			Name:      "deviceArm",
//			Namespace: "test",
//			Labels: map[string]string{
//				"environment": "dev",
//			},
//		},
//		TypeMeta: metav1.TypeMeta{
//			Kind:       "Device",
//			APIVersion: "resources/v1",
//		},
//		Spec: apis.DeviceSpec{
//			Name: "deviceArm",
//			AccessMethod: apis.AccessMethod{
//				Type: apis.AccessByAbility,
//				URL:  manageUrl,
//			},
//			ExpectedProperties: map[string]apis.Property{},
//			Abilities:          make([]apis.AbilitySpec, 0),
//		},
//		Status: apis.DeviceStatus{
//			DeviceID: "deviceArm",
//			Phase:    apis.DeviceIdle,
//			Status:   "idle",
//			ActionID: "",
//			Lock: apis.Lock{
//				Lock: true,
//			},
//			Abilities: make([]apis.AbilityStatus, 0),
//		},
//	}
//
//	device.Status.Abilities = append(device.Status.Abilities, apis.AbilityStatus{
//		Name: "Arm",
//		Services: []apis.AbilityServiceStatus{
//			{
//				Name:      "ArmAngle",
//				Ip:        "192.168.8.165",
//				Interface: "/api/control/arm_angle",
//			},
//			{
//				Name:      "LeftArmUp",
//				Ip:        "192.168.8.165",
//				Interface: "/api/control/left_arm_up",
//			},
//			{
//				Name:      "LeftArmDown",
//				Ip:        "192.168.8.165",
//				Interface: "/api/control/left_arm_down",
//			},
//		},
//	})
//
//	runtime1 := &apis.Runtime{
//		Image: "manage_ArmControl.Leju.Guochuang",
//		Name:  "RuntimeTest",
//		Type:  apis.ByDevice,
//		Devices: []apis.DeviceSpec{
//			device.Spec,
//		},
//		Outputs: make([]apis.Output, 1),
//		Inputs: []apis.Input{
//			{
//				Name:  "left",
//				Value: left,
//			},
//			{
//				Name:  "right",
//				Value: right,
//			},
//		},
//	}
//	runtime2 := &apis.Runtime{
//		Image: "service_LeftArmUp",
//		Name:  "RuntimeTest",
//		Type:  apis.ByDevice,
//		Devices: []apis.DeviceSpec{
//			device.Spec,
//		},
//		Outputs: make([]apis.Output, 1),
//		Inputs: []apis.Input{
//			{
//				Name:  "left",
//				Value: left,
//			},
//			{
//				Name:  "right",
//				Value: right,
//			},
//		},
//	}
//	runtime3 := &apis.Runtime{
//		Image: "service_Sleep",
//		Name:  "RuntimeTest",
//		Type:  apis.ByDevice,
//		Devices: []apis.DeviceSpec{
//			device.Spec,
//		},
//		Outputs: make([]apis.Output, 1),
//		Inputs: []apis.Input{
//			{
//				Name:  "left",
//				Value: left,
//			},
//			{
//				Name:  "right",
//				Value: right,
//			},
//			{
//				Name:  "sleepTime",
//				Value: "5",
//			},
//		},
//	}
//	runtime4 := &apis.Runtime{
//		Image: "service_LeftArmDown",
//		Name:  "RuntimeTest",
//		Type:  apis.ByDevice,
//		Devices: []apis.DeviceSpec{
//			device.Spec,
//		},
//		Outputs: make([]apis.Output, 1),
//		Inputs: []apis.Input{
//			{
//				Name:  "left",
//				Value: left,
//			},
//			{
//				Name:  "right",
//				Value: right,
//			},
//		},
//	}
//
//	action1 := &apis.Action{
//		ObjectMeta: metav1.ObjectMeta{
//			Name:      "action1arm",
//			Namespace: "test",
//			Labels: map[string]string{
//				"environment": "dev",
//			},
//		},
//		TypeMeta: metav1.TypeMeta{
//			Kind:       "Action",
//			APIVersion: "resources/v1",
//		},
//		Spec: apis.ActionSpec{
//			Name: "action1arm",
//			Runtimes: []apis.Runtime{
//				*runtime1,
//			},
//		},
//		Status: apis.ActionStatus{
//			ActionID:      "Action1arm",
//			Phase:         apis.Unknown,
//			Devices:       make(map[string]apis.DeviceStatus),
//			RuntimeStatus: make([]apis.RuntimeStatus, 1),
//		},
//	}
//
//	action2 := &apis.Action{
//		ObjectMeta: metav1.ObjectMeta{
//			Name:      "action2arm",
//			Namespace: "test",
//			Labels: map[string]string{
//				"environment": "dev",
//			},
//		},
//		TypeMeta: metav1.TypeMeta{
//			Kind:       "Action",
//			APIVersion: "resources/v1",
//		},
//		Spec: apis.ActionSpec{
//			Name: "action2arm",
//			Runtimes: []apis.Runtime{
//				*runtime2,
//			},
//			Conditions: apis.Conditions{
//				Formulas: []apis.ConditionFormula{
//					apis.ConditionFormula{
//						LeftValue: apis.ConditionValue{
//							Type:      apis.ResultsData,
//							Name:      "NodeDependency",
//							Value:     "0",
//							ValueType: "string",
//							From:      action1.Name,
//						},
//						RightValue: apis.ConditionValue{
//							Type:      apis.ConstData,
//							Name:      "NodeDependency",
//							Value:     "1",
//							ValueType: "string",
//							From:      "",
//						},
//						Signal: apis.Equal,
//						Join:   "",
//						Result: apis.False,
//					},
//				},
//			},
//		},
//		Status: apis.ActionStatus{
//			Phase:         apis.Unknown,
//			RuntimeStatus: make([]apis.RuntimeStatus, 1),
//			ActionID:      "Action2arm",
//			Devices:       make(map[string]apis.DeviceStatus),
//		},
//	}
//
//	action3 := &apis.Action{
//		ObjectMeta: metav1.ObjectMeta{
//			Name:      "action3arm",
//			Namespace: "test",
//			Labels: map[string]string{
//				"environment": "dev",
//			},
//		},
//		TypeMeta: metav1.TypeMeta{
//			Kind:       "Action",
//			APIVersion: "resources/v1",
//		},
//		Spec: apis.ActionSpec{
//			Name: "action3arm",
//			Runtimes: []apis.Runtime{
//				*runtime3,
//			},
//			Conditions: apis.Conditions{
//				Formulas: []apis.ConditionFormula{
//					apis.ConditionFormula{
//						LeftValue: apis.ConditionValue{
//							Type:      apis.ResultsData,
//							Name:      "NodeDependency",
//							Value:     "0",
//							ValueType: "string",
//							From:      action2.Name,
//						},
//						RightValue: apis.ConditionValue{
//							Type:      apis.ConstData,
//							Name:      "NodeDependency",
//							Value:     "1",
//							ValueType: "string",
//							From:      "",
//						},
//						Signal: apis.Equal,
//						Join:   "",
//						Result: apis.False,
//					},
//				},
//			},
//		},
//		Status: apis.ActionStatus{
//			Phase:         apis.Unknown,
//			RuntimeStatus: make([]apis.RuntimeStatus, 1),
//			ActionID:      "Action3arm",
//			Devices:       make(map[string]apis.DeviceStatus),
//		},
//	}
//
//	action4 := &apis.Action{
//		ObjectMeta: metav1.ObjectMeta{
//			Name:      "action4arm",
//			Namespace: "test",
//			Labels: map[string]string{
//				"environment": "dev",
//			},
//		},
//		TypeMeta: metav1.TypeMeta{
//			Kind:       "Action",
//			APIVersion: "resources/v1",
//		},
//		Spec: apis.ActionSpec{
//			Name: "action4arm",
//			Runtimes: []apis.Runtime{
//				*runtime4,
//			},
//			Conditions: apis.Conditions{
//				Formulas: []apis.ConditionFormula{
//					apis.ConditionFormula{
//						LeftValue: apis.ConditionValue{
//							Type:      apis.ResultsData,
//							Name:      "NodeDependency",
//							Value:     "0",
//							ValueType: "string",
//							From:      action3.Name,
//						},
//						RightValue: apis.ConditionValue{
//							Type:      apis.ConstData,
//							Name:      "NodeDependency",
//							Value:     "1",
//							ValueType: "string",
//							From:      "",
//						},
//						Signal: apis.Equal,
//						Join:   "",
//						Result: apis.False,
//					},
//				},
//			},
//		},
//		Status: apis.ActionStatus{
//			Phase:         apis.Unknown,
//			RuntimeStatus: make([]apis.RuntimeStatus, 1),
//			ActionID:      "Action4arm",
//			Devices:       make(map[string]apis.DeviceStatus),
//		},
//	}
//
//	action1.Status.Devices["deviceArm"] = device.Status
//	action2.Status.Devices["deviceArm"] = device.Status
//	action3.Status.Devices["deviceArm"] = device.Status
//	action4.Status.Devices["deviceArm"] = device.Status
//
//	group := &apis.Group{
//		ObjectMeta: metav1.ObjectMeta{
//			Name:      "grouparm",
//			Namespace: "test",
//		},
//		TypeMeta: metav1.TypeMeta{
//			Kind:       "Group",
//			APIVersion: "resources/v1",
//		},
//		Spec: apis.GroupSpec{
//			Replicas: []int32{0, 0},
//			Name:     "grouparm",
//			Actions:  make([]apis.Action, 4),
//		},
//		Status: apis.GroupStatus{
//			Node:    "test-node",
//			GroupID: "grouparm",
//			Phase:   apis.Unknown,
//			ActionStatus: []apis.ActionStatus{
//				action1.Status,
//				action2.Status,
//				action3.Status,
//				action4.Status,
//			},
//		},
//	}
//	group.Spec.Actions[0] = *action1
//	group.Spec.Actions[1] = *action2
//	group.Spec.Actions[2] = *action3
//	group.Spec.Actions[3] = *action4
//
//	return group, action1, runtime1, device
//}

//func CreateOrangeGroupActionRuntimeGrab() (*apis.Group, *apis.Action, *apis.Runtime, *apis.Device) {
//	// 能力框架的url
//	manageUrl := "http://192.168.8.197:8080"
//	taskTypeGrab := "0"
//	taskTypePut := "1"
//	// 创建device
//	device := &apis.Device{
//		ObjectMeta: metav1.ObjectMeta{
//			Name:      "deviceGrab",
//			Namespace: "test",
//			Labels: map[string]string{
//				"environment": "dev",
//			},
//		},
//		TypeMeta: metav1.TypeMeta{
//			Kind:       "Device",
//			APIVersion: "resources/v1",
//		},
//		Spec: apis.DeviceSpec{
//			Name: "deviceGrab",
//			AccessMethod: apis.AccessMethod{
//				Type: apis.AccessByAbility,
//				URL:  manageUrl,
//			},
//			ExpectedProperties: map[string]apis.Property{},
//			Abilities:          make([]apis.AbilitySpec, 0),
//		},
//		Status: apis.DeviceStatus{
//			DeviceID: "deviceGrab",
//			Phase:    apis.DeviceIdle,
//			Status:   "idle",
//			ActionID: "",
//			Lock: apis.Lock{
//				Lock: true,
//			},
//			Abilities: make([]apis.AbilityStatus, 0),
//		},
//	}
//
//	device.Status.Abilities = append(device.Status.Abilities, apis.AbilityStatus{
//		Name: "Grab",
//		Services: []apis.AbilityServiceStatus{
//			{
//				Name:      "TaskState",
//				Ip:        "192.168.8.197",
//				Interface: "/api/task_state",
//			},
//			{
//				Name:      "GoStandBy",
//				Ip:        "192.168.8.197",
//				Interface: "/api/go_standby",
//			},
//			{
//				Name:      "StartTask",
//				Ip:        "192.168.8.197",
//				Interface: "/api/start_task",
//			},
//			{
//				Name:      "GoInit",
//				Ip:        "192.168.8.197",
//				Interface: "/api/go_initial",
//			},
//		},
//	})
//
//	runtime1 := &apis.Runtime{
//		Image: "manage_ActInferenceAbility",
//		Name:  "RuntimeTest",
//		Type:  apis.ByDevice,
//		Devices: []apis.DeviceSpec{
//			device.Spec,
//		},
//		Outputs: make([]apis.Output, 1),
//		Inputs: []apis.Input{
//			{
//				Name:  "taskType",
//				Value: taskTypeGrab,
//			},
//		},
//	}
//	runtime2 := &apis.Runtime{
//		Image: "service_GoStandBy",
//		Name:  "RuntimeTest",
//		Type:  apis.ByDevice,
//		Devices: []apis.DeviceSpec{
//			device.Spec,
//		},
//		Outputs: make([]apis.Output, 1),
//		Inputs: []apis.Input{
//			{
//				Name:  "taskType",
//				Value: taskTypeGrab,
//			},
//		},
//	}
//
//	runtime3 := &apis.Runtime{
//		Image: "service_TaskState",
//		Name:  "RuntimeTest",
//		Type:  apis.ByDevice,
//		Devices: []apis.DeviceSpec{
//			device.Spec,
//		},
//		Outputs: make([]apis.Output, 1),
//		Inputs: []apis.Input{
//			{
//				Name:  "taskType",
//				Value: taskTypeGrab,
//			},
//		},
//	}
//
//	runtime4 := &apis.Runtime{
//		Image: "service_StartTask",
//		Name:  "RuntimeTest",
//		Type:  apis.ByDevice,
//		Devices: []apis.DeviceSpec{
//			device.Spec,
//		},
//		Outputs: make([]apis.Output, 1),
//		Inputs: []apis.Input{
//			{
//				Name:  "taskType",
//				Value: taskTypeGrab,
//			},
//		},
//	}
//
//	runtime5 := &apis.Runtime{
//		Image: "service_TaskState",
//		Name:  "RuntimeTest",
//		Type:  apis.ByDevice,
//		Devices: []apis.DeviceSpec{
//			device.Spec,
//		},
//		Outputs: make([]apis.Output, 1),
//		Inputs: []apis.Input{
//			{
//				Name:  "taskType",
//				Value: taskTypeGrab,
//			},
//		},
//	}
//
//	runtime6 := &apis.Runtime{
//		Image: "service_GoInit",
//		Name:  "RuntimeTest",
//		Type:  apis.ByDevice,
//		Devices: []apis.DeviceSpec{
//			device.Spec,
//		},
//		Outputs: make([]apis.Output, 1),
//		Inputs: []apis.Input{
//			{
//				Name:  "taskType",
//				Value: taskTypePut,
//			},
//		},
//	}
//
//	runtime7 := &apis.Runtime{
//		Image: "service_TaskState",
//		Name:  "RuntimeTest",
//		Type:  apis.ByDevice,
//		Devices: []apis.DeviceSpec{
//			device.Spec,
//		},
//		Outputs: make([]apis.Output, 1),
//		Inputs: []apis.Input{
//			{
//				Name:  "taskType",
//				Value: taskTypePut,
//			},
//		},
//	}
//
//	runtime8 := &apis.Runtime{
//		Image: "service_GoStandBy",
//		Name:  "RuntimeTest",
//		Type:  apis.ByDevice,
//		Devices: []apis.DeviceSpec{
//			device.Spec,
//		},
//		Outputs: make([]apis.Output, 1),
//		Inputs: []apis.Input{
//			{
//				Name:  "taskType",
//				Value: taskTypePut,
//			},
//		},
//	}
//
//	runtime9 := &apis.Runtime{
//		Image: "service_TaskState",
//		Name:  "RuntimeTest",
//		Type:  apis.ByDevice,
//		Devices: []apis.DeviceSpec{
//			device.Spec,
//		},
//		Outputs: make([]apis.Output, 1),
//		Inputs: []apis.Input{
//			{
//				Name:  "taskType",
//				Value: taskTypePut,
//			},
//		},
//	}
//
//	runtime10 := &apis.Runtime{
//		Image: "service_StartTask",
//		Name:  "RuntimeTest",
//		Type:  apis.ByDevice,
//		Devices: []apis.DeviceSpec{
//			device.Spec,
//		},
//		Outputs: make([]apis.Output, 1),
//		Inputs: []apis.Input{
//			{
//				Name:  "taskType",
//				Value: taskTypePut,
//			},
//		},
//	}
//
//	action1 := &apis.Action{
//		ObjectMeta: metav1.ObjectMeta{
//			Name:      "action1grab",
//			Namespace: "test",
//			Labels: map[string]string{
//				"environment": "dev",
//			},
//		},
//		TypeMeta: metav1.TypeMeta{
//			Kind:       "Action",
//			APIVersion: "resources/v1",
//		},
//		Spec: apis.ActionSpec{
//			Name: "action1grab",
//			Runtimes: []apis.Runtime{
//				*runtime1,
//			},
//		},
//		Status: apis.ActionStatus{
//			Phase:         apis.Unknown,
//			ActionID:      "Action1grab",
//			Devices:       make(map[string]apis.DeviceStatus),
//			RuntimeStatus: make([]apis.RuntimeStatus, 1),
//		},
//	}
//
//	action2 := &apis.Action{
//		ObjectMeta: metav1.ObjectMeta{
//			Name:      "action2grab",
//			Namespace: "test",
//			Labels: map[string]string{
//				"environment": "dev",
//			},
//		},
//		TypeMeta: metav1.TypeMeta{
//			Kind:       "Action",
//			APIVersion: "resources/v1",
//		},
//		Spec: apis.ActionSpec{
//			Name: "action2grab",
//			Runtimes: []apis.Runtime{
//				*runtime2,
//			},
//			Conditions: apis.Conditions{
//				Formulas: []apis.ConditionFormula{
//					apis.ConditionFormula{
//						LeftValue: apis.ConditionValue{
//							Type:      apis.ResultsData,
//							Name:      "NodeDependency",
//							Value:     "0",
//							ValueType: "string",
//							From:      action1.Name,
//						},
//						RightValue: apis.ConditionValue{
//							Type:      apis.ConstData,
//							Name:      "NodeDependency",
//							Value:     "1",
//							ValueType: "string",
//							From:      "",
//						},
//						Signal: apis.Equal,
//						Join:   "",
//						Result: apis.False,
//					},
//				},
//			},
//		},
//		Status: apis.ActionStatus{
//			Phase:         apis.Unknown,
//			RuntimeStatus: make([]apis.RuntimeStatus, 1),
//			ActionID:      "Action2grab",
//			Devices:       make(map[string]apis.DeviceStatus),
//		},
//	}
//
//	action3 := &apis.Action{
//		ObjectMeta: metav1.ObjectMeta{
//			Name:      "action3grab",
//			Namespace: "test",
//			Labels: map[string]string{
//				"environment": "dev",
//			},
//		},
//		TypeMeta: metav1.TypeMeta{
//			Kind:       "Action",
//			APIVersion: "resources/v1",
//		},
//		Spec: apis.ActionSpec{
//			Name: "action3grab",
//			Runtimes: []apis.Runtime{
//				*runtime3,
//			},
//			Conditions: apis.Conditions{
//				Formulas: []apis.ConditionFormula{
//					apis.ConditionFormula{
//						LeftValue: apis.ConditionValue{
//							Type:      apis.ResultsData,
//							Name:      "NodeDependency",
//							Value:     "0",
//							ValueType: "string",
//							From:      action2.Name,
//						},
//						RightValue: apis.ConditionValue{
//							Type:      apis.ConstData,
//							Name:      "NodeDependency",
//							Value:     "1",
//							ValueType: "string",
//							From:      "",
//						},
//						Signal: apis.Equal,
//						Join:   "",
//						Result: apis.False,
//					},
//				},
//			},
//		},
//		Status: apis.ActionStatus{
//			Phase:         apis.Unknown,
//			RuntimeStatus: make([]apis.RuntimeStatus, 1),
//			ActionID:      "Action3grab",
//			Devices:       make(map[string]apis.DeviceStatus),
//		},
//	}
//	action4 := &apis.Action{
//		ObjectMeta: metav1.ObjectMeta{
//			Name:      "action4grab",
//			Namespace: "test",
//			Labels: map[string]string{
//				"environment": "dev",
//			},
//		},
//		TypeMeta: metav1.TypeMeta{
//			Kind:       "Action",
//			APIVersion: "resources/v1",
//		},
//		Spec: apis.ActionSpec{
//			Name: "action4grab",
//			Runtimes: []apis.Runtime{
//				*runtime4,
//			},
//			Conditions: apis.Conditions{
//				Formulas: []apis.ConditionFormula{
//					apis.ConditionFormula{
//						LeftValue: apis.ConditionValue{
//							Type:      apis.ResultsData,
//							Name:      "NodeDependency",
//							Value:     "0",
//							ValueType: "string",
//							From:      action3.Name,
//						},
//						RightValue: apis.ConditionValue{
//							Type:      apis.ConstData,
//							Name:      "NodeDependency",
//							Value:     "1",
//							ValueType: "string",
//							From:      "",
//						},
//						Signal: apis.Equal,
//						Join:   "",
//						Result: apis.False,
//					},
//				},
//			},
//		},
//		Status: apis.ActionStatus{
//			Phase:         apis.Unknown,
//			RuntimeStatus: make([]apis.RuntimeStatus, 1),
//			ActionID:      "Action4grab",
//			Devices:       make(map[string]apis.DeviceStatus),
//		},
//	}
//
//	action5 := &apis.Action{
//		ObjectMeta: metav1.ObjectMeta{
//			Name:      "action5grab",
//			Namespace: "test",
//			Labels: map[string]string{
//				"environment": "dev",
//			},
//		},
//		TypeMeta: metav1.TypeMeta{
//			Kind:       "Action",
//			APIVersion: "resources/v1",
//		},
//		Spec: apis.ActionSpec{
//			Name: "action5grab",
//			Runtimes: []apis.Runtime{
//				*runtime5,
//			},
//			Conditions: apis.Conditions{
//				Formulas: []apis.ConditionFormula{
//					apis.ConditionFormula{
//						LeftValue: apis.ConditionValue{
//							Type:      apis.ResultsData,
//							Name:      "NodeDependency",
//							Value:     "0",
//							ValueType: "string",
//							From:      action4.Name,
//						},
//						RightValue: apis.ConditionValue{
//							Type:      apis.ConstData,
//							Name:      "NodeDependency",
//							Value:     "1",
//							ValueType: "string",
//							From:      "",
//						},
//						Signal: apis.Equal,
//						Join:   "",
//						Result: apis.False,
//					},
//				},
//			},
//		},
//		Status: apis.ActionStatus{
//			RuntimeStatus: make([]apis.RuntimeStatus, 1),
//			ActionID:      "Action5grab",
//			Devices:       make(map[string]apis.DeviceStatus),
//		},
//	}
//
//	action6 := &apis.Action{
//		ObjectMeta: metav1.ObjectMeta{
//			Name:      "action6grab",
//			Namespace: "test",
//			Labels: map[string]string{
//				"environment": "dev",
//			},
//		},
//		TypeMeta: metav1.TypeMeta{
//			Kind:       "Action",
//			APIVersion: "resources/v1",
//		},
//		Spec: apis.ActionSpec{
//			Name: "action6grab",
//			Runtimes: []apis.Runtime{
//				*runtime6,
//			},
//			Conditions: apis.Conditions{
//				Formulas: []apis.ConditionFormula{
//					apis.ConditionFormula{
//						LeftValue: apis.ConditionValue{
//							Type:      apis.ResultsData,
//							Name:      "NodeDependency",
//							Value:     "0",
//							ValueType: "string",
//							From:      action5.Name,
//						},
//						RightValue: apis.ConditionValue{
//							Type:      apis.ConstData,
//							Name:      "NodeDependency",
//							Value:     "1",
//							ValueType: "string",
//							From:      "",
//						},
//						Signal: apis.Equal,
//						Join:   "",
//						Result: apis.False,
//					},
//				},
//			},
//		},
//		Status: apis.ActionStatus{
//			Phase:         apis.Unknown,
//			RuntimeStatus: make([]apis.RuntimeStatus, 1),
//			ActionID:      "Action6grab",
//			Devices:       make(map[string]apis.DeviceStatus),
//		},
//	}
//
//	action7 := &apis.Action{
//		ObjectMeta: metav1.ObjectMeta{
//			Name:      "action7grab",
//			Namespace: "test",
//			Labels: map[string]string{
//				"environment": "dev",
//			},
//		},
//		TypeMeta: metav1.TypeMeta{
//			Kind:       "Action",
//			APIVersion: "resources/v1",
//		},
//		Spec: apis.ActionSpec{
//			Name: "action7grab",
//			Runtimes: []apis.Runtime{
//				*runtime7,
//			},
//			Conditions: apis.Conditions{
//				Formulas: []apis.ConditionFormula{
//					apis.ConditionFormula{
//						LeftValue: apis.ConditionValue{
//							Type:      apis.ResultsData,
//							Name:      "NodeDependency",
//							Value:     "0",
//							ValueType: "string",
//							From:      action6.Name,
//						},
//						RightValue: apis.ConditionValue{
//							Type:      apis.ConstData,
//							Name:      "NodeDependency",
//							Value:     "1",
//							ValueType: "string",
//							From:      "",
//						},
//						Signal: apis.Equal,
//						Join:   "",
//						Result: apis.False,
//					},
//				},
//			},
//		},
//		Status: apis.ActionStatus{
//			Phase:         apis.Unknown,
//			RuntimeStatus: make([]apis.RuntimeStatus, 1),
//			ActionID:      "Action7grab",
//			Devices:       make(map[string]apis.DeviceStatus),
//		},
//	}
//
//	action8 := &apis.Action{
//		ObjectMeta: metav1.ObjectMeta{
//			Name:      "action8grab",
//			Namespace: "test",
//			Labels: map[string]string{
//				"environment": "dev",
//			},
//		},
//		TypeMeta: metav1.TypeMeta{
//			Kind:       "Action",
//			APIVersion: "resources/v1",
//		},
//		Spec: apis.ActionSpec{
//			Name: "action8grab",
//			Runtimes: []apis.Runtime{
//				*runtime8,
//			},
//			Conditions: apis.Conditions{
//				Formulas: []apis.ConditionFormula{
//					apis.ConditionFormula{
//						LeftValue: apis.ConditionValue{
//							Type:      apis.ResultsData,
//							Name:      "NodeDependency",
//							Value:     "0",
//							ValueType: "string",
//							From:      action7.Name,
//						},
//						RightValue: apis.ConditionValue{
//							Type:      apis.ConstData,
//							Name:      "NodeDependency",
//							Value:     "1",
//							ValueType: "string",
//							From:      "",
//						},
//						Signal: apis.Equal,
//						Join:   "",
//						Result: apis.False,
//					},
//				},
//			},
//		},
//		Status: apis.ActionStatus{
//			Phase:         apis.Unknown,
//			RuntimeStatus: make([]apis.RuntimeStatus, 1),
//			ActionID:      "Action8grab",
//			Devices:       make(map[string]apis.DeviceStatus),
//		},
//	}
//
//	action9 := &apis.Action{
//		ObjectMeta: metav1.ObjectMeta{
//			Name:      "action9grab",
//			Namespace: "test",
//			Labels: map[string]string{
//				"environment": "dev",
//			},
//		},
//		TypeMeta: metav1.TypeMeta{
//			Kind:       "Action",
//			APIVersion: "resources/v1",
//		},
//		Spec: apis.ActionSpec{
//			Name: "action9grab",
//			Runtimes: []apis.Runtime{
//				*runtime9,
//			},
//			Conditions: apis.Conditions{
//				Formulas: []apis.ConditionFormula{
//					apis.ConditionFormula{
//						LeftValue: apis.ConditionValue{
//							Type:      apis.ResultsData,
//							Name:      "NodeDependency",
//							Value:     "0",
//							ValueType: "string",
//							From:      action8.Name,
//						},
//						RightValue: apis.ConditionValue{
//							Type:      apis.ConstData,
//							Name:      "NodeDependency",
//							Value:     "1",
//							ValueType: "string",
//							From:      "",
//						},
//						Signal: apis.Equal,
//						Join:   "",
//						Result: apis.False,
//					},
//				},
//			},
//		},
//		Status: apis.ActionStatus{
//			Phase:         apis.Unknown,
//			RuntimeStatus: make([]apis.RuntimeStatus, 1),
//			ActionID:      "Action9grab",
//			Devices:       make(map[string]apis.DeviceStatus),
//		},
//	}
//
//	action10 := &apis.Action{
//		ObjectMeta: metav1.ObjectMeta{
//			Name:      "action10grab",
//			Namespace: "test",
//			Labels: map[string]string{
//				"environment": "dev",
//			},
//		},
//		TypeMeta: metav1.TypeMeta{
//			Kind:       "Action",
//			APIVersion: "resources/v1",
//		},
//		Spec: apis.ActionSpec{
//			Name: "action10grab",
//			Runtimes: []apis.Runtime{
//				*runtime10,
//			},
//			Conditions: apis.Conditions{
//				Formulas: []apis.ConditionFormula{
//					apis.ConditionFormula{
//						LeftValue: apis.ConditionValue{
//							Type:      apis.ResultsData,
//							Name:      "NodeDependency",
//							Value:     "0",
//							ValueType: "string",
//							From:      action9.Name,
//						},
//						RightValue: apis.ConditionValue{
//							Type:      apis.ConstData,
//							Name:      "NodeDependency",
//							Value:     "1",
//							ValueType: "string",
//							From:      "",
//						},
//						Signal: apis.Equal,
//						Join:   "",
//						Result: apis.False,
//					},
//				},
//			},
//		},
//		Status: apis.ActionStatus{
//			Phase:         apis.Unknown,
//			RuntimeStatus: make([]apis.RuntimeStatus, 1),
//			ActionID:      "Action10grab",
//			Devices:       make(map[string]apis.DeviceStatus),
//		},
//	}
//
//	action1.Status.Devices["deviceArm"] = device.Status
//	action2.Status.Devices["deviceArm"] = device.Status
//	action3.Status.Devices["deviceArm"] = device.Status
//	action4.Status.Devices["deviceArm"] = device.Status
//	action5.Status.Devices["deviceArm"] = device.Status
//	action6.Status.Devices["deviceArm"] = device.Status
//	action7.Status.Devices["deviceArm"] = device.Status
//	action8.Status.Devices["deviceArm"] = device.Status
//	action9.Status.Devices["deviceArm"] = device.Status
//	action10.Status.Devices["deviceArm"] = device.Status
//
//	group := &apis.Group{
//		ObjectMeta: metav1.ObjectMeta{
//			Name:      "groupgrab",
//			Namespace: "test",
//		},
//		TypeMeta: metav1.TypeMeta{
//			Kind:       "Group",
//			APIVersion: "resources/v1",
//		},
//		Spec: apis.GroupSpec{
//			Replicas: []int32{0, 0},
//			Name:     "groupgrab",
//			Actions:  make([]apis.Action, 10),
//			Parents:  []string{"grouparm"},
//		},
//		Status: apis.GroupStatus{
//			Node:    "test-node",
//			GroupID: "group",
//			Phase:   apis.Unknown,
//			ActionStatus: []apis.ActionStatus{
//				action1.Status,
//				action2.Status,
//				action3.Status,
//				action4.Status,
//				action5.Status,
//				action6.Status,
//				action7.Status,
//				action8.Status,
//				action9.Status,
//				action10.Status,
//			},
//		},
//	}
//	group.Spec.Actions[0] = *action1
//	group.Spec.Actions[1] = *action2
//	group.Spec.Actions[2] = *action3
//	group.Spec.Actions[3] = *action4
//	group.Spec.Actions[4] = *action5
//	group.Spec.Actions[5] = *action6
//	group.Spec.Actions[6] = *action7
//	group.Spec.Actions[7] = *action8
//	group.Spec.Actions[8] = *action9
//	group.Spec.Actions[9] = *action10
//	return group, action1, runtime1, device
//}
//
//func GenerateOrangeTask() apis.Task {
//	//predicrGroup, _, _, _ := CreateOrangeGroupActionRuntimePredict()
//	garm, _, _, _ := CreateOrangeGroupActionRuntimeArm()
//	ggrab, _, _, _ := CreateOrangeGroupActionRuntimeGrab()
//	orangeTask := apis.Task{
//		TypeMeta: metav1.TypeMeta{
//			Kind:       "Task",
//			APIVersion: "resources/v1",
//		},
//		ObjectMeta: metav1.ObjectMeta{
//			Name:      "orange_task",
//			Namespace: "test",
//		},
//		Spec: apis.TaskSpec{
//			Name: "orange_task",
//			Desc: apis.Description{
//				Docs: "国创中心抓橙子演示任务",
//			},
//			Type:   apis.Norm,
//			Groups: []apis.Group{*garm, *ggrab},
//		},
//		Status: apis.TaskStatus{
//			GroupStatus: []apis.GroupStatus{
//				{
//					GroupID: "",
//				},
//				{},
//				{},
//			},
//			Phase: apis.Unknown,
//		},
//	}
//	return orangeTask
//}

//// go test -run TestAddInDataBus -v
//func TestAddInDataBus(t *testing.T) {
//	logs.Init("testModule")
//	ctx := context.Background()
//	task := GenerateOrangeTask()
//	clientSet, _ := utils.CreateClientSet()
//	tclient := clientSet.Core().Tasks("test")
//	gclient := clientSet.Core().Groups("test")
//	_, _ = tclient.Create(ctx, &task, metav1.CreateOptions{})
//	for _, group := range task.Spec.Groups {
//		gclient.Create(ctx, &group, metav1.CreateOptions{})
//	}
//}

// go test -run TestSendToProxy -v
//func TestSendToProxy(t *testing.T) {
//	logs.Init("testModule")
//	task := GenerateOrangeTask()
//	client := &http.Client{}
//	marshal, err := json.Marshal(task)
//	if err != nil {
//		logs.Error(err)
//		return
//	}
//	url := "http://192.168.8.191:8899/framework/v1/task?Name=orange_task"
//	logs.Info(url)
//	req, err := http.NewRequest("POST", url, strings.NewReader(string(marshal)))
//	if err != nil {
//		logs.Fatal(err)
//	}
//	//Content-Type很重要，下文解释
//	//req.Header.Set("Content-Type", "application/x-www")
//	req.Header.Set("Content-Type", "application/json")
//	//req.Header.Set("Content-Type", "multipart/form-data")
//
//	rep, err := client.Do(req)
//	if err != nil {
//		logs.Fatal(err)
//	}
//	data, err := io.ReadAll(rep.Body)
//	defer func(Body io.ReadCloser) {
//		err := Body.Close()
//		if err != nil {
//			logs.Error(err)
//		}
//	}(rep.Body)
//	if err != nil {
//		logs.Fatal(err)
//	}
//	logs.Infof("resp is : %s", string(data))
//}

//// go test -run TestGenerateOrangeTask -v
//func TestGenerateOrangeTask(t *testing.T) {
//	predicrGroup, _, _, _ := CreateOrangeGroupActionRuntimePredict()
//	garm, _, _, _ := CreateOrangeGroupActionRuntimeArm()
//	ggrab, _, _, _ := CreateOrangeGroupActionRuntimeGrab()
//	orangeTask := apis.Task{
//		TypeMeta: metav1.TypeMeta{
//			Kind:       "Task",
//			APIVersion: "resources/v1",
//		},
//		ObjectMeta: metav1.ObjectMeta{
//			Name:      "orange_task",
//			Namespace: "test",
//		},
//		Spec: apis.TaskSpec{
//			Name: "orange_task",
//			Desc: apis.Description{
//				Docs: "国创中心抓橙子演示任务",
//			},
//			Type:   apis.Norm,
//			Groups: []apis.Group{*predicrGroup, *garm, *ggrab},
//		},
//		Status: apis.TaskStatus{
//			Phase: apis.Unknown,
//		},
//	}
//	marshal, err := json.Marshal(orangeTask)
//	if err != nil {
//		logs.Error(err)
//		return
//	}
//	fmt.Println(string(marshal))
//}

//// go test -run TestGenerateOrangeTask -v
//func TestGenerateSimpleTask(t *testing.T) {
//	orangeTask := apis.Task{
//		TypeMeta: metav1.TypeMeta{
//			Kind:       "Task",
//			APIVersion: "resources/v1",
//		},
//		ObjectMeta: metav1.ObjectMeta{
//			Name:      "testtask",
//			Namespace: "test",
//		},
//		Spec: apis.TaskSpec{
//			Name:   "testtask",
//			Type:   apis.Norm,
//			Groups: []apis.Group{},
//		},
//		Status: apis.TaskStatus{
//			Phase: apis.Unknown,
//		},
//	}
//	marshal, err := json.Marshal(orangeTask)
//	if err != nil {
//		logs.Error(err)
//		return
//	}
//	fmt.Println(string(marshal))
//}

// go test -run TestAddDevice -v

//func TestAddDevice(t *testing.T) {
//	_, _, _, predictDevice := CreateOrangeGroupActionRuntimePredict()
//	_, _, _, armDevice := CreateOrangeGroupActionRuntimeArm()
//	_, _, _, grabDevice := CreateOrangeGroupActionRuntimeGrab()
//
//	moduleName := "testModule"
//	logs.Init(moduleName)
//	ctx := context.Background()
//	cs, err := createClientSet()
//	if err != nil {
//		logs.Error(err)
//		return
//	}
//
//	deviceClient := cs.Core().Devices("test")
//	_, err = deviceClient.Create(ctx, predictDevice, metav1.CreateOptions{})
//	if err != nil {
//		logs.Error(err)
//	}
//	_, err = deviceClient.Create(ctx, armDevice, metav1.CreateOptions{})
//	if err != nil {
//		logs.Error(err)
//	}
//	_, err = deviceClient.Create(ctx, grabDevice, metav1.CreateOptions{})
//	if err != nil {
//		logs.Error(err)
//	}
//}

// go test -run TestEnd -v
func TestEnd(t *testing.T) {
	url1 := "http://192.168.8.165:8080"
	url2 := "http://192.168.8.197:8080"
	abilityName1 := "ArmControl.Leju.Guochuang"
	abilityName2 := "Detect"
	abilityName3 := "ActInferenceAbility"
	err := manager.NewAbilityManager(url1, abilityName1).TerminateAbility()
	if err != nil {
		logs.Errorf("fail to create AbilityManager: %v", err)
	}
	err = manager.NewAbilityManager(url1, abilityName2).TerminateAbility()
	if err != nil {
		logs.Errorf("fail to create AbilityManager: %v", err)
	}
	err = manager.NewAbilityManager(url2, abilityName3).TerminateAbility()
	if err != nil {
		logs.Errorf("fail to create AbilityManager: %v", err)
	}
}

// go test -run TestEnd -v
func TestParse(t *testing.T) {
	logs.Init("ttt")
	mess := "{\"groupID\" : \"afsd\", \"c\", \"d\"}"

	var resp transport.ScoreRespData
	err := json.Unmarshal([]byte(mess), &resp)
	if err != nil {
		logs.Error(err)
	}
	fmt.Println(resp.Score)
}

//func TestFile(t *testing.T) {
//	task := GenerateOrangeTask()
//	mashral, err := json.Marshal(task)
//	if err != nil {
//		fmt.Println(err)
//	}
//	// 将JSON数据写入文件
//	err = os.WriteFile("task.txt", mashral, 0644)
//	if err != nil {
//		fmt.Println("文件写入错误:", err)
//		return
//	}
//}

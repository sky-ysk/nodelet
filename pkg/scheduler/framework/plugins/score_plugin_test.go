package plugins

import (
	"context"
	"fmt"
	"hit.edu/framework/pkg/apimachinery/runtime"
	"hit.edu/framework/pkg/apimachinery/runtime/schema"
	"hit.edu/framework/pkg/apimachinery/runtime/serializer"
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/apis/meta"
	metav1 "hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/client-go/clients"
	"hit.edu/framework/pkg/client-go/rest"
	"hit.edu/framework/pkg/component-base/logs"
	"io"
	"net/http"
	"testing"
	"time"
)

func TestBuildGroupsRequest(t *testing.T) {
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
	clientSet, err := clients.NewForConfig(c)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	// 资源定义在 pkg/apis/xxx/type.go 下
	// 这里以访问资源Task为例，
	// 获取访问Task的客户端
	// 默认访问的Namespace是 ""

	tasksClient := clientSet.Core().Tasks("test")
	task := mockGetTask()
	results, err := tasksClient.Create(context.TODO(), &task, metav1.CreateOptions{})
	if err != nil {
		logs.Errorf("Failed to create task: %v", err)
		fmt.Println(err)
		panic(err)
	}
	fmt.Println("Created task ", results)

}

func TestScore(t *testing.T) {
	task := mockGetTask()
	ctx := context.Background()
	plugin := &ScorePluginDBY{
		pluginClient: NewScorePluginClient(),
	}
	//plugin.SendGroups(ctx, &task)
	plugin.Score(ctx, &task.Spec.Groups[0], "EdgeNode2")
}

func TestListAllNodes(t *testing.T) {
	scheme := runtime.NewScheme()
	apis.AddToScheme(scheme)
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
			MaxIdleConns:        10000,            // 最大空闲连接数
			IdleConnTimeout:     90 * time.Second, // 空闲连接超时时间
			TLSHandshakeTimeout: 10 * time.Second, // TLS 握手超时时间
		},
		Timeout: 1000 * time.Second,
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

	nodesClient := clientSet.Core().Nodes("test")
	logs.Info("listing 筛选的node")
	lstOpts := metav1.ListOptions{}
	list, err := nodesClient.List(context.TODO(), lstOpts)
	if err != nil {
		logs.Error(err)
	}
	for _, d := range list.Items {
		fmt.Println(d)
	}
}

func TestSendGroupsRequest(t *testing.T) {
	task := mockGetTask()
	ctx := context.Background()
	req := BuildSendGroupsRequest(ctx, &task)
	plugin := &ScorePluginDBY{
		pluginClient: NewScorePluginClient(),
	}
	//NewScorePluginClient()
	plugin.SendGroups(ctx, &task)
	fmt.Println(req)
	fmt.Println(len(req.TopInfo))
	//TODO Fill the node name
	//plugin.Score(ctx, &task.Spec.Groups[0], "")
}

func mockGetTask() apis.Task {
	reqs := make([]apis.ResourceRequirement, 0)
	req1 := apis.ResourceRequirement{
		Name:       "CPU",
		Lowbound:   "1",
		Upperbound: "1",
	}
	req2 := apis.ResourceRequirement{
		Name:       "RAM",
		Lowbound:   "1",
		Upperbound: "1",
	}
	req3 := apis.ResourceRequirement{
		Name:       "IO",
		Lowbound:   "1",
		Upperbound: "1",
	}
	reqs = append(reqs, req1, req2, req3)
	g1 := apis.Group{
		ObjectMeta: meta.ObjectMeta{
			Name: "testGroup1",
		},
		Spec: apis.GroupSpec{
			ResourceRequirements: reqs,
		},
		Status: apis.GroupStatus{
			GroupID: "testGroup1",
			Belongs: apis.IDRef{
				TaskID: "task1",
			},
		},
	}
	g2 := apis.Group{
		ObjectMeta: meta.ObjectMeta{
			Name: "testGroup2",
		},
		Spec: apis.GroupSpec{
			ResourceRequirements: reqs,
			Parents:              []string{"testGroup1"},
		},
		Status: apis.GroupStatus{
			GroupID: "testGroup2",
			Belongs: apis.IDRef{
				TaskID: "task1",
			},
		},
	}
	g3 := apis.Group{
		ObjectMeta: meta.ObjectMeta{
			Name: "testGroup3",
		},
		Spec: apis.GroupSpec{
			ResourceRequirements: reqs,
			Parents:              []string{"testGroup2"},
		},
		Status: apis.GroupStatus{
			GroupID: "testGroup3",
			Belongs: apis.IDRef{
				TaskID: "task1",
			},
		},
	}

	return apis.Task{
		ObjectMeta: meta.ObjectMeta{
			Name: "testTask",
		},
		Spec: apis.TaskSpec{
			Groups: []apis.Group{g1, g2, g3},
		},
		Status: apis.TaskStatus{
			TaskID: "task1",
		},
	}
}

// go test -run TestAddNode -v
func TestAddNode(t *testing.T) {
	logs.Init("main")
	scheme := runtime.NewScheme()
	apis.AddToScheme(scheme)
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
			MaxIdleConns:        10000,            // 最大空闲连接数
			IdleConnTimeout:     90 * time.Second, // 空闲连接超时时间
			TLSHandshakeTimeout: 10 * time.Second, // TLS 握手超时时间
		},
		Timeout: 1000 * time.Second,
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

	nodesClient := clientSet.Core().Nodes("test")

	node := &apis.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "demo-nodes",
			Namespace: "test",
			Labels: map[string]string{
				"environment": "dev",
			},
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Node",
			APIVersion: "resources/v1",
		},
		Spec: apis.NodeSpec{
			NodeName: "EndNode1",
		},
	}
	node2 := &apis.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "demo-node2",
			Namespace: "test",
			Labels: map[string]string{
				"environment": "dev",
			},
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Node",
			APIVersion: "resources/v1",
		},
		Spec: apis.NodeSpec{
			NodeName: "EndNode2",
		},
	}
	node3 := &apis.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "demo-node3",
			Namespace: "test",
			Labels: map[string]string{
				"environment": "qa",
			},
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Node",
			APIVersion: "resources/v1",
		},
		Spec: apis.NodeSpec{
			NodeName: "EndNode3",
		},
	}

	logs.Trace("creating")
	result, err := nodesClient.Create(context.TODO(), node, metav1.CreateOptions{})
	if err != nil {
		logs.Errorf("Failed to create node: %v", err)
	} else {
		logs.Info("created node", result)
	}
	_, err = nodesClient.Create(context.TODO(), node2, metav1.CreateOptions{})
	if err != nil {
		logs.Errorf("Failed to create node: %v", err)
	} else {
		logs.Info("created node", result)
	}
	_, err = nodesClient.Create(context.TODO(), node3, metav1.CreateOptions{})
	if err != nil {
		logs.Errorf("Failed to create node: %v", err)
	} else {
		logs.Info("created node", result)
	}

	if err != nil {
		logs.Error(err)
	}
	logs.Info("listing 筛选的node")
	lstOpts := metav1.ListOptions{
		//LabelSelector: "environment",
	}
	list, err := nodesClient.List(context.TODO(), lstOpts)
	if err != nil {
		logs.Error(err)
	}
	for _, d := range list.Items {
		logs.Info(d)
	}
}

// 机器人A拿盘子到远处 -> 机器人B放橙子 -> 机器人A拿盘子到近处 -> 机器人B拿出橙子 (loop)
//func mockGetOrangeTask() apis.Task {
//
//	g1 := apis.Group{
//		TypeMeta: meta.TypeMeta{
//			Kind:       "Group",
//			APIVersion: "resources/v1",
//		},
//		ObjectMeta: meta.ObjectMeta{
//			Name: "PlacePlateToFarPosition",
//		},
//		Spec: apis.GroupSpec{
//			Name: "PlacePlateToFarPosition",
//			Desc: apis.Description{
//				Docs: "把盘子放到远处位置",
//			},
//			Type:          apis.Norm,
//			AffinityNodes: []string{"RobotA"},
//			Actions: []apis.Action{
//				{
//					TypeMeta: meta.TypeMeta{
//						Kind:       "Action",
//						APIVersion: "resources/v1",
//					},
//					ObjectMeta: meta.ObjectMeta{
//						Name: "Action-PlacePlateToFarPosition",
//					},
//					Spec: apis.ActionSpec{
//						Name: "",
//						Runtimes: []apis.Runtime{
//							{
//								Name:   "Runtime-PlacePlateToFarPosition",
//								Inputs: apis.Input{},
//							},
//						},
//						Type: apis.Norm,
//						Desc: apis.Description{},
//					},
//					Status: apis.ActionStatus{},
//				},
//			},
//		},
//		Status: apis.GroupStatus{
//			GroupID: "testGroup1",
//			Belongs: apis.IDRef{
//				TaskID: "task1",
//			},
//		},
//	}
//	g2 := apis.Group{
//		TypeMeta: meta.TypeMeta{
//			Kind:       "Group",
//			APIVersion: "resources/v1",
//		},
//		ObjectMeta: meta.ObjectMeta{
//			Name: "PlaceOrangeToPlate-Far",
//		},
//		Spec: apis.GroupSpec{
//			Name: "PlaceOrangeToPlate-Far",
//			Desc: apis.Description{
//				Docs: "把橙子放到盘子里面",
//			},
//			Parents: []string{"PlacePlateToFarPosition"},
//			Conditions: apis.Conditions{
//				Formulas: []apis.ConditionFormula{
//					{
//						LeftValue: apis.ConditionValue{
//							Type:      apis.ResultsData,
//							Name:      "testGroup1_execution_status",
//							ValueType: "bool",
//							From:      "resource_bus",
//						},
//						RightValue: apis.ConditionValue{
//							Type:      apis.ConstData,
//							Name:      "const_data",
//							Value:     "true",
//							ValueType: "bool",
//						},
//						Signal: apis.Equal,
//					},
//				},
//			},
//			Type:          apis.Norm,
//			AffinityNodes: []string{"RobotB"},
//			Actions: []apis.Action{
//				{
//					TypeMeta: meta.TypeMeta{
//						Kind:       "Action",
//						APIVersion: "resources/v1",
//					},
//					ObjectMeta: meta.ObjectMeta{
//						Name: "Action-PlaceOrangeToPlate",
//					},
//					Spec: apis.ActionSpec{
//						Name: "Action-PlaceOrangeToPlate",
//						Runtimes: []apis.Runtime{
//							{
//								Name:    "Runtime-PlaceOrangeToPlate",
//								Inputs:  apis.Input{},
//								Outputs: apis.Output{},
//							},
//						},
//						Type: apis.Norm,
//						Desc: apis.Description{},
//					},
//					Status: apis.ActionStatus{},
//				},
//			},
//		},
//		Status: apis.GroupStatus{
//			GroupID: "testGroup2",
//			Belongs: apis.IDRef{
//				TaskID: "task1",
//			},
//		},
//	}
//	g3 := apis.Group{
//		TypeMeta: meta.TypeMeta{
//			Kind:       "Group",
//			APIVersion: "resources/v1",
//		},
//		ObjectMeta: meta.ObjectMeta{
//			Name: "PlacePlateToClosePosition",
//		},
//		Spec: apis.GroupSpec{
//			Name: "PlacePlateToClosePosition",
//			Desc: apis.Description{
//				Docs: "把盘子放到近处",
//			},
//			Parents: []string{"PlaceOrangeToPlate-Far"},
//			Conditions: apis.Conditions{
//				Formulas: []apis.ConditionFormula{
//					{
//						LeftValue: apis.ConditionValue{
//							Type:      apis.ResultsData,
//							Name:      "testGroup2_execution_status",
//							ValueType: "bool",
//							From:      "resource_bus",
//						},
//						RightValue: apis.ConditionValue{
//							Type:      apis.ConstData,
//							Name:      "const_data",
//							Value:     "true",
//							ValueType: "bool",
//						},
//						Signal: apis.Equal,
//					},
//				},
//			},
//			Type:          apis.Norm,
//			AffinityNodes: []string{"RobotA"},
//			Actions: []apis.Action{
//				{
//					TypeMeta: meta.TypeMeta{
//						Kind:       "Action",
//						APIVersion: "resources/v1",
//					},
//					ObjectMeta: meta.ObjectMeta{
//						Name: "Action-PlacePlateToClosePosition",
//					},
//					Spec: apis.ActionSpec{
//						Name: "Action-PlacePlateToClosePosition",
//						Runtimes: []apis.Runtime{
//							{
//								Name:    "Runtime-PlacePlateToClosePosition",
//								Inputs:  apis.Input{},
//								Outputs: apis.Output{},
//							},
//						},
//						Type: apis.Norm,
//						Desc: apis.Description{},
//					},
//					Status: apis.ActionStatus{},
//				},
//			},
//		},
//		Status: apis.GroupStatus{
//			GroupID: "testGroup3",
//			Belongs: apis.IDRef{
//				TaskID: "task1",
//			},
//		},
//	}
//
//	g4 := apis.Group{
//		TypeMeta: meta.TypeMeta{
//			Kind:       "Group",
//			APIVersion: "resources/v1",
//		},
//		ObjectMeta: meta.ObjectMeta{
//			Name: "TakeOrangeFromPlate",
//		},
//		Spec: apis.GroupSpec{
//			Name: "TakeOrangeFromPlate",
//			Desc: apis.Description{
//				Docs: "把橙子从盘子里拿出",
//			},
//			Parents: []string{"PlacePlateToClosePosition"},
//			Conditions: apis.Conditions{
//				Formulas: []apis.ConditionFormula{
//					{
//						LeftValue: apis.ConditionValue{
//							Type:      apis.ResultsData,
//							Name:      "testGroup3_execution_status",
//							ValueType: "bool",
//							From:      "resource_bus",
//						},
//						RightValue: apis.ConditionValue{
//							Type:      apis.ConstData,
//							Name:      "const_data",
//							Value:     "true",
//							ValueType: "bool",
//						},
//						Signal: apis.Equal,
//					},
//				},
//			},
//			Type:          apis.Norm,
//			AffinityNodes: []string{"RobotB"},
//			Actions: []apis.Action{
//				{
//					TypeMeta: meta.TypeMeta{
//						Kind:       "Action",
//						APIVersion: "resources/v1",
//					},
//					ObjectMeta: meta.ObjectMeta{
//						Name: "Action-TakeOrangeFromPlate",
//					},
//					Spec: apis.ActionSpec{
//						Name: "Action-TakeOrangeFromPlate",
//						Runtimes: []apis.Runtime{
//							{
//								Name:    "Runtime-TakeOrangeFromPlate",
//								Inputs:  apis.Input{},
//								Outputs: apis.Output{},
//							},
//						},
//						Type: apis.Norm,
//						Desc: apis.Description{},
//					},
//					Status: apis.ActionStatus{},
//				},
//			},
//		},
//		Status: apis.GroupStatus{
//			GroupID: "testGroup4",
//			Belongs: apis.IDRef{
//				TaskID: "task1",
//			},
//		},
//	}
//
//	return apis.Task{
//		TypeMeta: meta.TypeMeta{
//			Kind:       "task",
//			APIVersion: "resources/v1",
//		},
//		ObjectMeta: meta.ObjectMeta{
//			Name: "testOrangeWorkflowTask",
//		},
//		Spec: apis.TaskSpec{
//			Groups: []apis.Group{g1, g2, g3, g4},
//		},
//		Status: apis.TaskStatus{
//			TaskID: "task1",
//		},
//	}
//}

//func TestTask(t *testing.T) {
//	task := mockGetOrangeTask()
//	data, err := json.Marshal(task)
//	if err != nil {
//		t.Fatal(err)
//	}
//	fmt.Println(string(data))
//}

func TestServer(t *testing.T) {
	http.HandleFunc("/schedule/postGroup", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("request received is ")
		fmt.Println(r)
		// 获取请求报文的内容长度
		length := r.ContentLength
		// 新建一个字节切片，长度与请求报文的内容长度相同
		fmt.Println("lem is")
		fmt.Println(length)
		body, err := io.ReadAll(r.Body)
		// 读取 r 的请求主体，并将具体内容读入 body 中
		if err != nil {
			fmt.Println(err)
			return
		}
		// 将字节切片内容写入相应报文
		fmt.Println("body is: ", string(body))
	})
	port := ":5000"
	fmt.Printf("Starting server on port %s...\n", port)
	err := http.ListenAndServe(port, nil)
	if err != nil {
		fmt.Println("Error starting server: ", err)
		return
	}

}

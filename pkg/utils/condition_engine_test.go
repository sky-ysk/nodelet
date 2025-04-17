package utils

import (
	"context"
	"fmt"
	"hit.edu/framework/pkg/apimachinery/runtime"
	"hit.edu/framework/pkg/apimachinery/runtime/schema"
	"hit.edu/framework/pkg/apimachinery/runtime/serializer"
	apis "hit.edu/framework/pkg/apis/cores"
	metav1 "hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/client-go/clients"
	"hit.edu/framework/pkg/client-go/rest"
	"hit.edu/framework/pkg/component-base/logs"
	"net/http"
	"testing"
	"time"
)

func TestAddAction(t *testing.T) {
	logs.Init("main")
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
	clientSet, err := clients.NewForConfig(c)
	if err != nil {
		panic(err)
	}
	actionsClient := clientSet.Core().Actions("test")
	action := &apis.Action{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "testmeta",
			Namespace: "test",
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Action",
			APIVersion: "resources/v1",
		},
		Spec: apis.ActionSpec{
			Name: "testspec",
		},
	}
	fmt.Println("creating")
	_, err = actionsClient.Create(context.TODO(), action, metav1.CreateOptions{})
	if err != nil {
		logs.Errorf("Failed to create action: %v", err)
		panic(err)
	}
	//git config  credential.helper store
	result, getErr := actionsClient.Get(context.TODO(), "testspec", metav1.GetOptions{})
	if getErr != nil {
		panic(fmt.Errorf("Failed to get : %v", getErr))
	}
	fmt.Println("get result ", result)

}

//单独测试解析功能
func TestParseFunction(t *testing.T) {
	//创建一个group并用client提交，附带一个需要查找的condition，打印解析的结果
	// Task  总共1个Task、3个Group、3个Action、6个runtime
	// task1Name := "TrainInferTask-1" // 第一个Task的Name
	task1ID := "TrainInferTaskID-1" // 第一个Task的ID
	group1_1Name := "TrainGroup-1"  // 第一个Task下的第一个GroupName
	group1_1ID := "GroupID-1"
	group1_1Replicas := []int32{0, 0}
	action1_1_1Name := "Action1-1"
	action1_1_1ID := "ActionID1-1"
	runtime1_1_1_1Name := "Runtime1-1-1"
	runtime1_1_1_2Name := "Runtime1-1-2"
	runtime1_1_1_1ID := "RuntimeID1-1-1" // 第一个Task下的第一个Group下的第一个ActionName下的第一个RuntimeID
	runtime1_1_1_2ID := "RuntimeID1-1-2"
	runtime1_1_1_1FineGrainedControl := false
	runtime1_1_1_2FineGrainedControl := false

	runtime1_1_1_1Condition := apis.Conditions{
		Formulas: []apis.ConditionFormula{
			apis.ConditionFormula{
				LeftValue: apis.ConditionValue{
					Type:      apis.ResultsData,
					Name:      "ProgramDependency",
					Value:     "0",
					ValueType: "string",
					From:      "/home/public/goprojects/myProject/test/nodelet/task_exporter/dependency/requirements.txt",
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
			},
		},
	}
	runtime1_1_1_2Condition := apis.Conditions{
		Formulas: []apis.ConditionFormula{
			apis.ConditionFormula{
				LeftValue: apis.ConditionValue{
					Type:      apis.ResultsData,
					Name:      "NodeDependency",
					Value:     "0",
					ValueType: "string",
					From:      "Task{task1}.Group{TrainGroup-1}.Action{action1}.Runtime{runtime0}",
					Field:     "Status{}.ActionStatus{0}.RuntimeStatus{0}.Phase{}",
				},
				RightValue: apis.ConditionValue{
					Type:      apis.ConstData,
					Name:      "NodeDependency",
					Value:     "1",
					ValueType: "string",
					From:      "Group{TrainGroup-1}",
					Field:     "Status{}.ActionStatus{0}.RuntimeStatus{0}.Phase{}",
				},
				Signal: apis.Equal,
				Join:   "",
				Result: apis.False,
			},
		},
	}

	g1 := apis.Group{
		ObjectMeta: metav1.ObjectMeta{Name: group1_1Name, Namespace: ""},
		TypeMeta:   metav1.TypeMeta{Kind: "Group", APIVersion: "resources/v1"},
		Spec: apis.GroupSpec{
			Replicas: group1_1Replicas,
			Name:     group1_1Name,
			Parents:  make([]string, 0),
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
								Args:                     []string{"/home/public/workspace/heongtong_yolo_linux/predict.py"},
								Parents:                  make([]string, 0), // 加入Parents
								Conditions:               runtime1_1_1_1Condition,
								Image:                    "/home/l1hy/workspace/heongtong_yolo_linux/predict.py",
								EnvVar:                   []apis.EnvVar{apis.EnvVar{Name: "", Value: ""}},
								EnableFineGrainedControl: runtime1_1_1_1FineGrainedControl,
							},
							apis.Runtime{
								Name:                     runtime1_1_1_2Name,
								Type:                     apis.ByCommand,
								Command:                  []string{"python"},
								Args:                     []string{"/home/public/workspace/heongtong_yolo_linux/predict.py"},
								Parents:                  []string{runtime1_1_1_1Name}, // 加入Parents
								Conditions:               runtime1_1_1_2Condition,
								Image:                    "/home/l1hy/workspace/heongtong_yolo_linux/predict.py",
								EnvVar:                   []apis.EnvVar{apis.EnvVar{Name: "", Value: ""}},
								EnableFineGrainedControl: runtime1_1_1_2FineGrainedControl,
							},
						},
					},
					Status: apis.ActionStatus{
						ActionID: action1_1_1ID, // ActionID =ActionName + GroupID
						Phase:    apis.Unknown,
						RuntimeStatus: []apis.RuntimeStatus{
							apis.RuntimeStatus{
								RuntimeID: runtime1_1_1_1ID, // RuntimeID = RuntimeName + ActionID
								Phase:     apis.Successed,
							},
							apis.RuntimeStatus{
								RuntimeID: runtime1_1_1_2ID, // RuntimeID = RuntimeName + ActionID
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
			},
			Belongs: apis.IDRef{TaskID: task1ID},
			Phase:   apis.Unknown,
		},
	}
	group1 := &g1


	// 解析字段路径并访问对应的字段
	FromItem, err := ParseFrom(group1.Spec.Actions[0].Spec.Runtimes[1].Conditions.Formulas[0].LeftValue.From)
	if err != nil {
		fmt.Printf("Parse FromItemInfo err")
	}
	fmt.Printf("FromItem:%v", FromItem)
	getItem := group1
	Value, Type, err := ParseField(*getItem, group1.Spec.Actions[0].Spec.Runtimes[1].Conditions.Formulas[0].LeftValue.Field)
	if err != nil {
		fmt.Printf("Parse Field err")
	}
	fmt.Println("Value:", Value, "Type:", Type)
}

func TestGetValue(t *testing.T) {
	logs.Init("main")
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
	clientSet, err := clients.NewForConfig(c)
	if err != nil {
		panic(err)
	}
	actionsClient := clientSet.Core().Actions("test")
	action := &apis.Action{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "testmeta",
			Namespace: "test",
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Action",
			APIVersion: "resources/v1",
		},
		Spec: apis.ActionSpec{
			Name: "testspec",
		},
	}
	fmt.Println("creating")
	_, err = actionsClient.Create(context.TODO(), action, metav1.CreateOptions{})
	if err != nil {
		logs.Errorf("Failed to create action: %v", err)
		panic(err)
	}
	//git config  credential.helper store
	result, getErr := actionsClient.Get(context.TODO(), "testspec", metav1.GetOptions{})
	if getErr != nil {
		panic(fmt.Errorf("Failed to get : %v", getErr))
	}
	fmt.Println("get result ", result)

}
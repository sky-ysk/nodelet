package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"hit.edu/framework/pkg/apimachinery/runtime"
	"hit.edu/framework/pkg/apimachinery/runtime/schema"
	"hit.edu/framework/pkg/apimachinery/runtime/serializer"
	apis "hit.edu/framework/pkg/apis/cores"
	metav1 "hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/client-go/clients"
	"hit.edu/framework/pkg/client-go/rest"
	Manager "hit.edu/framework/pkg/client-go/util/manager"
	"hit.edu/framework/pkg/component-base/logs"
	value "hit.edu/framework/pkg/utils/value"
)

func CreateClientSet() (*clients.ClientSet, error) {
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
		return nil, err
	}
	return clientSet, nil
}

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

// TODO 测试NodeDependency
func TestNodeDependency(t *testing.T) {
	//参数配置
	CreateLocalTest()
}

// TODO 测试DataDependency
func TestDataDependency(t *testing.T) {
	//创建condition Engine
	client, _ := CreateClientSet()
	ce := NewConditionEngine(client)
	fmt.Println(ce)
	logs.Init("====test condition engine init====")

	testOption := []apis.DataType{apis.FileData, apis.LocalData, apis.DeviceData, apis.ResultsData}

	option := testOption[1] //选择测试的某个功能

	switch option {
	case apis.FileData:

	case apis.LocalData:
		_, runtime := CreateLocalTest()
		fmt.Println("====After create Local Test:====")
		// 数据依赖（../tmp/testFolder）
		res, err := ce.CheckConditions(runtime.Spec.Conditions, runtime)
		if err != nil {
			fmt.Println("err!")
		}
		fmt.Println("res:", res)
	case apis.DeviceData:

	case apis.ResultsData:

	default:
		fmt.Println("err test option type!")
	}

}

func CreateLocalTest() (apis.Action, apis.Runtime) {
	clientset, err := CreateClientSet()
	if err != nil {
		panic(err)
	}
	// 构造Manager
	m := Manager.NewManager(clientset)
	engine := value.NewEngine(clientset)

	logs.Init("main")

	// 生成UUID
	u := uuid.Must(uuid.NewV7())
	namespace := "sky-test"

	action1_1_1Name := "A1"
	runtime1_1_1_1Name := "R1"
	// 数据依赖（../tmp/testFolder）
	DataDependencyConditionFormula := apis.ConditionFormula{
		ConditionType: apis.DataDependency,
		LeftValue: apis.Value{
			Type:      apis.LocalData,
			Name:      "sky-test",
			Value:     "0",
			ValueType: "string",
			From:      "Action{A1}.Runtime{R1}.Outputs{Output1}",
		},
		RightValue: apis.Value{
			Type:      apis.ConstData,
			Name:      "asdasd",
			Value:     "123",
			ValueType: "string",
			From:      "",
		},
		Signal: apis.Equal,
		Join:   "",
		Result: apis.False,
	}
	runtime1_1_1_1Condition := apis.Conditions{
		Formulas: []apis.ConditionFormula{
			DataDependencyConditionFormula,
		},
	}
	Action := apis.ActionSpec{
		Name: action1_1_1Name,
		Runtimes: []apis.RuntimeSpec{
			apis.RuntimeSpec{
				Name:       runtime1_1_1_1Name,
				Type:       apis.ByCommand,
				Command:    []string{"python"},
				Args:       []string{"upload.py", "test.txt", "uotput.txt"}, // 10s
				Inputs:     []apis.Value{apis.Value{Value: "test.txt"}},     // 加入Parents
				Conditions: &runtime1_1_1_1Condition,
			},
		},
	}

	a, err := m.CreateAction(Action, nil, namespace, u.String(), "")
	if err != nil {
		panic(err)
	}

	actionName := a.Name
	action, err := m.GetAction(actionName, namespace)
	if err != nil {
		panic(err)
	}
	runtimeref := action.Status.Runtimes["R1"]
	runtimeName := runtimeref.Name
	runtime, err := m.GetRuntime(runtimeName, namespace)
	runtimeOutputs := map[string]apis.Value{
		"Output1": {
			Value: "123",
		},
		"Output2": {
			Value: "456",
		},
	}
	patchRuntime, err := json.Marshal(map[string]interface{}{
		"status": map[string]interface{}{
			"outputs": &runtimeOutputs,
		},
	})
	_, err = m.PatchRuntime(runtime.Name, namespace, patchRuntime)
	if err != nil {
		logs.Errorf("Patch runtime error-2:%v", err)
	}

	actionOutputs := map[string]apis.Value{
		"Output1": {
			Value: "789",
		},
		"Output2": {
			Value: "10 11 12",
		},
	}
	patchAction, err := json.Marshal(map[string]interface{}{
		"status": map[string]interface{}{
			"outputs": &actionOutputs,
		},
	})
	_, err = m.PatchAction(actionName, namespace, patchAction)
	if err != nil {
		logs.Errorf("Patch action error-2:%v", err)
	}

	// 睡眠1秒
	time.Sleep(time.Duration(time.Second * 1))
	actionNew, _ := m.GetAction(actionName, namespace)
	runtimeNew, _ := m.GetRuntime(runtime.Name, namespace)
	fmt.Println("==========")
	fmt.Println(actionNew.Status.Outputs)
	fmt.Println("==========")
	fmt.Println(runtimeNew.Status.Outputs)

	actionfrom := "Action{A1}.Outputs{Output1}"
	runtimefrom := "Runtime{R1}.Outputs{Output1}"
	ActionRefValue := apis.Value{
		From:  actionfrom,
		Value: "actionsuccess!",
	}
	RuntimeRefValue := apis.Value{
		From:  runtimefrom,
		Value: "runtimesuccess!",
	}
	ActionValue, err := engine.ExtractLocalValue(&ActionRefValue, *actionNew)

	fmt.Println("++++++++")
	fmt.Println(ActionValue.Value)

	RuntimeValue, err := engine.ExtractLocalValue(&RuntimeRefValue, *runtimeNew)
	fmt.Println("-----------------")
	fmt.Println(RuntimeValue.Value)
	return *actionNew, *runtimeNew
}

// TODO 测试ResourceDependency
func TestResourceDependency(t *testing.T) {
	//参数配置
}

// TODO 测试ProcessDependency
func TestProcessDependency(t *testing.T) {
	//参数配置
}

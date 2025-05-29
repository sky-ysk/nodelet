package value

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
	"hit.edu/framework/pkg/client-go/util/manager"
	"hit.edu/framework/pkg/component-base/logs"
)

// 创建Task
func TestCreateTask(t *testing.T) {
	clientset, err := CreateClientSet()
	if err != nil {
		panic(err)
	}
	// 构造Manager
	m := manager.NewManager(clientset)

	logs.Init("main")

	// 生成UUID
	u := uuid.Must(uuid.NewV7())
	rs1 := apis.RuntimeSpec{
		Name:  "R1",
		Type:  apis.ByDevice,
		Image: "xxxxx",
	}
	rs2 := apis.RuntimeSpec{
		Name:  "R2",
		Type:  apis.ByDevice,
		Image: "xxxxx",
	}

	as1 := apis.ActionSpec{
		Name: "A1",
		Runtimes: []apis.RuntimeSpec{
			rs1,
			rs2,
		},
	}

	as2 := apis.ActionSpec{
		Name:     "A2",
		Runtimes: []apis.RuntimeSpec{},
	}

	gs1 := apis.GroupSpec{
		Name: "G1",
		Actions: []apis.ActionSpec{
			as1,
			as2,
		},
	}

	gs2 := apis.GroupSpec{
		Name:    "G2",
		Actions: []apis.ActionSpec{},
	}

	ts := apis.TaskSpec{
		Name: "T1",
		Groups: []apis.GroupSpec{
			gs1,
			gs2,
		},
	}

	task, err := m.CreateTask(ts, nil, "Guochuang", u.String(), "")
	if err != nil {
		panic(err)
	}

	fmt.Println(task)

}

// 测试发现问题是status里面的reference信息必须填完整才能从Task Group Action Runtime中获取到信息，否则找不到
// 但是在创建Task的时候并没有填充这些Status的信息
func TestValueExtract(t *testing.T) {
	fmt.Println("TestValueExtract")
	clientSet, err := CreateClientSet()
	if err != nil {
		panic(err)
	}

	engine := NewEngine(clientSet)

	name := "T1.G1-20250529T190834-c7db3"
	namespace := "Test"

	fmt.Println("start get group")
	g, err := engine.manager.GetGroup(name, namespace)
	if err != nil {

		fmt.Println("get group err:", err)
		panic(err)
	}

	// value := apis.Value{
	// 	Name:      "Test",
	// 	Type:      apis.LocalData,
	// 	From:      "Group{G1}.Action{A1}.Status{phase}",
	// 	ValueType: apis.StringType,
	// }

	// v, err := engine.ExtractLocalValue(&value, *g)
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println("-----------")
	// fmt.Println(v)

	value := apis.Value{
		NameSpace: "Test",
		Name:      "Task",
		Type:      apis.LocalData,
		From:      "Group{G2}.Action{A1}.Runtime{R1}.Status{belong}",
		ValueType: apis.StringType,
	}

	v, err := engine.ExtractLocalValue(&value, *g)
	if err != nil {
		panic(err)
	}
	fmt.Println("-----------")
	fmt.Println(v)
}

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

func TestCreateDevice(t *testing.T) {
	clientset, err := CreateClientSet()
	if err != nil {
		panic(err)
	}

	ip := "127.0.0.1"
	inter := "api/control/start_task"
	port := "2387"
	spec := apis.DeviceSpec{
		Name: "Robot",
	}
	status := apis.DeviceStatus{
		Abilities: map[string]apis.Ability{
			"Move": apis.Ability{
				Name: "Move.Leju.Guochuang",
				Services: map[string]apis.AbilityService{
					"Start": apis.AbilityService{
						Ip:        &ip,
						Port:      &port,
						Interface: &inter,
					},
				},
			},
		},
	}
	device := apis.Device{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "Robot",
			Namespace: "Guochuang",
		},
		Spec:   spec,
		Status: status,
	}

	client := clientset.Core().Devices(device.Namespace)
	client.Create(context.TODO(), &device, metav1.CreateOptions{})
}

//func TestGetDeviceImage(t *testing.T) {
//	clientSet, err := CreateClientSet()
//	if err != nil {
//		panic(err)
//	}
//
//	engine := NewEngine(clientSet)
//
//	namespace := "Guochuang"
//
//	from := "Device{Robot}.Ability{Move}.Service{Start}"
//
//	v, err := engine.ExtractDeviceValue(from, namespace)
//	if err != nil {
//		panic(err)
//	}
//	fmt.Println("-----------")
//	fmt.Println(v)
//}

func TestCreateTaskForOutput(t *testing.T) {
	clientset, err := CreateClientSet()
	if err != nil {
		panic(err)
	}
	// 构造Manager
	m := manager.NewManager(clientset)
	engine := NewEngine(clientset)

	logs.Init("main")

	// 生成UUID
	u := uuid.Must(uuid.NewV7())
	namespace := "sky-test"
	rs1 := apis.RuntimeSpec{
		Name:  "R1",
		Type:  apis.ByDevice,
		Image: "xxxxx",
	}
	as1 := apis.ActionSpec{
		Name: "A1",
		Runtimes: []apis.RuntimeSpec{
			rs1,
		},
	}

	action, err := m.CreateAction(as1, nil, namespace, u.String(), "")
	if err != nil {
		panic(err)
	}

	fmt.Println(action)

	actionName := action.Name
	a, err := engine.manager.GetAction(actionName, namespace)
	if err != nil {
		panic(err)
	}
	runtimeref := a.Status.Runtimes["R1"]
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

	// 睡眠三秒
	time.Sleep(time.Duration(time.Second * 1))
	actionNew, _ := engine.manager.GetAction(actionName, namespace)
	runtimeNew, _ := engine.manager.GetRuntime(runtime.Name, namespace)
	fmt.Println("==========")
	fmt.Println(actionNew.Status.Outputs)
	fmt.Println("==========")
	fmt.Println(runtimeNew.Status.Outputs)

	actionfrom := "Action{A1}.Outputs{Output1}"
	runtimefrom := "Runtime{R1}.Outputs{Output1}"
	ActionRefValue := apis.Value{
		From:  actionfrom,
		Value: "success!",
	}
	RuntimeRefValue := apis.Value{
		From:  runtimefrom,
		Value: "success!",
	}
	ActionValue, err := engine.ExtractLocalValue(&ActionRefValue, *actionNew)

	fmt.Println("++++++++")
	fmt.Println(ActionValue)

	RuntimeValue, err := engine.ExtractLocalValue(&RuntimeRefValue, *runtimeNew)
	fmt.Println("-----------------")
	fmt.Println(RuntimeValue)

}

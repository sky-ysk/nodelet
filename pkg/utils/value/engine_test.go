package value

import (
	"context"
	"fmt"
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
	"testing"
	"time"
)

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

func TestValueExtract(t *testing.T) {
	clientSet, err := CreateClientSet()
	if err != nil {
		panic(err)
	}
	
	engine := NewEngine(clientSet)
	
	name := "T1.G1-0196434c-279b-7e67-8ac6-0bf756da1c7f"
	namespace := "Guochuang"
	
	g, err := engine.manager.GetGroup(name, namespace)
	if err != nil {
		panic(err)
	}
	
	value := apis.Value{
		Name:      "Test",
		Type:      apis.LocalData,
		From:      "Group{G1}.Action{A1}.Status{phase}",
		ValueType: apis.StringType,
	}
	
	v, err := engine.ExtractLocalValue(&value, *g)
	if err != nil {
		panic(err)
	}
	fmt.Println("-----------")
	fmt.Println(v)
	
	value = apis.Value{
		Name:      "Test",
		Type:      apis.LocalData,
		From:      "Group{G1}.Status{phase}",
		ValueType: apis.StringType,
	}
	
	v, err = engine.ExtractLocalValue(&value, *g)
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

func TestGetDeviceImage(t *testing.T) {
	clientSet, err := CreateClientSet()
	if err != nil {
		panic(err)
	}
	
	engine := NewEngine(clientSet)
	
	namespace := "Guochuang"
	from := "Device{Robot}.Ability{Move}.Service{Start}"
	
	v, err := engine.ExtractDeviceValue(from, namespace)
	if err != nil {
		panic(err)
	}
	fmt.Println("-----------")
	fmt.Println(v)
}

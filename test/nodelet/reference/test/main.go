package main

import (
	"hit.edu/framework/pkg/apimachinery/runtime"
	"hit.edu/framework/pkg/apimachinery/runtime/schema"
	"hit.edu/framework/pkg/apimachinery/runtime/serializer"
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/client-go/clients"
	"hit.edu/framework/pkg/client-go/rest"
	"hit.edu/framework/pkg/client-go/util/manager"
	"hit.edu/framework/pkg/component-base/logs"
	"net/http"
	"time"
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

//func main() {
//	// create Group
//	clientset, err := CreateClientSet()
//	if err != nil {
//		panic(err)
//	}
//	// 构造Manager
//	m := manager.NewManager(clientset)
//
//	logs.Init("main")
//
//	// 生成UUID
//	u := uuid.Must(uuid.NewV7())
//	rs1 := apis.RuntimeSpec{
//		Name:  "R1",
//		Type:  apis.ByDevice,
//		Image: "xxxxx",
//	}
//	rs2 := apis.RuntimeSpec{
//		Name:  "R2",
//		Type:  apis.ByDevice,
//		Image: "xxxxx",
//	}
//
//	as1 := apis.ActionSpec{
//		Name: "A1",
//		Runtimes: []apis.RuntimeSpec{
//			rs1,
//			rs2,
//		},
//	}
//
//	as2 := apis.ActionSpec{
//		Name:     "A2",
//		Runtimes: []apis.RuntimeSpec{},
//	}
//
//	gs := apis.GroupSpec{
//		Name: "G1",
//		Actions: []apis.ActionSpec{
//			as1,
//			as2,
//		},
//	}
//
//	g, err := m.CreateGroup(gs, nil, "Guochuang", u.String(), "")
//	if err != nil {
//		panic(err)
//	}
//
//	fmt.Println(g)
//}

func main() {
	//Get Group
	moduleName := "testModule"
	logs.Init(moduleName)
	scheme := runtime.NewScheme()
	apis.AddToScheme(scheme)
	logs.Info(scheme)

	logs.Info("11111")
	clientset, err := CreateClientSet()
	if err != nil {
		panic(err)
	}
	// 构造Manager
	m := manager.NewManager(clientset)
	_, err = m.GetGroup("G1-01964c27-7c19-7d3d-9ef2-78a20ffd707a", "Guochuang")
	if err != nil {
		logs.Errorf("get group err:%v", err)
	}
	logs.Info("get group success")
}

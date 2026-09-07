package main

import (
	"fmt"
	"net/http"
	"time"

	"hit.edu/framework/pkg/apimachinery/runtime"
	"hit.edu/framework/pkg/apimachinery/runtime/schema"
	"hit.edu/framework/pkg/apimachinery/runtime/serializer"
	apis "hit.edu/framework/pkg/apis/cores"
	metav1 "hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/client-go/clients"
	"hit.edu/framework/pkg/client-go/rest"
	"hit.edu/framework/pkg/client-go/util/manager"
)

func CreateClientSet() (*clients.ClientSet, error) {
	scheme := runtime.NewScheme()
	apis.AddToScheme(scheme)
	//参数配置
	// TODO: 填写参数
	//部分参数之后可以在core_client等 编写setConfigDefaults函数进行填充
	c := &rest.Config{
		Host:    "http://120.220.95.189:48120",
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

func main() {
	namespace1 := "HenanEP"

	clientset, err := CreateClientSet()
	manager := manager.NewManager(clientset)
	if err != nil {
		panic(err)
	}
	data := &apis.Data{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "wasm-windows-HenanEP",
			Namespace: "HenanEP",
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Data",
			APIVersion: "resources/v1",
		},
		Spec: apis.DataSpec{
			Name:       "电网wasm-windows-HenanEP测试",
			BelongNode: "n20",
			FilePath:   "/root/goprojects/workflow/wasm-windows-HenanEP.json", //这个就是咱们放工作流文件的地方，然后前端读到这个文件名（注意是文件名），然后根据文件名再去到文件仓库当中读取这个json文件，最后展示到前端
			FileFormat: "json",
			Desc: &apis.Description{
				Docs: "电网wasm-windows-HenanEP测试",
			},
		},
		Status: apis.DataStatus{
			CreateAt: &apis.Time{time.Now()},
			LastTime: &apis.Time{time.Now()},
		},
	}

	// data1 := &apis.Data{
	// 	ObjectMeta: metav1.ObjectMeta{
	// 		Name:      "Dianli",
	// 		Namespace: "HenanEP",
	// 	},
	// 	TypeMeta: metav1.TypeMeta{
	// 		Kind:       "Data",
	// 		APIVersion: "resources/v1",
	// 	},
	// 	Spec: apis.DataSpec{
	// 		Name:       "Dianli",
	// 		BelongNode: "n20",
	// 		FilePath:   "/root/goprojects/workflow/Dianli.json", //这个就是咱们放工作流文件的地方，然后前端读到这个文件名（注意是文件名），然后根据文件名再去到文件仓库当中读取这个json文件，最后展示到前端
	// 		FileFormat: "json",
	// 		Desc: &apis.Description{
	// 			Docs: "电力拉起服务",
	// 		},
	// 	},
	// 	Status: apis.DataStatus{
	// 		CreateAt: &apis.Time{time.Now()},
	// 		LastTime: &apis.Time{time.Now()},
	// 	},
	// }

	d, err := manager.CreateData(data, namespace1)
	if err != nil {
		panic(err)
	}
	fmt.Println(d)

	// d, err = manager.CreateData(data1, namespace1)
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println(d)

}

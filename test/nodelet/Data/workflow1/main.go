package main

import (
	"fmt"
	"hit.edu/framework/pkg/apimachinery/runtime"
	"hit.edu/framework/pkg/apimachinery/runtime/schema"
	"hit.edu/framework/pkg/apimachinery/runtime/serializer"
	apis "hit.edu/framework/pkg/apis/cores"
	metav1 "hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/client-go/clients"
	"hit.edu/framework/pkg/client-go/rest"
	"hit.edu/framework/pkg/client-go/util/manager"
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

func main() {
	namespace1 := "HenanEP"
	namespace2 := "Cosmo"
	namespace3 := "ShandongHS"
	clientset, err := CreateClientSet()
	manager := manager.NewManager(clientset)
	if err != nil {
		panic(err)
	}

	//data := &apis.Data{
	//	ObjectMeta: metav1.ObjectMeta{
	//		Name:      "HenanEP-workflow1",
	//		Namespace: "HenanEP",
	//	},
	//	TypeMeta: metav1.TypeMeta{
	//		Kind:       "Data",
	//		APIVersion: "resources/v1",
	//	},
	//	Spec: apis.DataSpec{
	//		Name:       "HenanEP-workflow1",
	//		BelongNode: "n19",
	//		FilePath:   "/root/goprojects/workflow/dianwang.json",
	//		FileFormat: "json",
	//		Desc: &apis.Description{
	//			Docs: "北航-电网工作流-Command",
	//		},
	//	},
	//	Status: apis.DataStatus{
	//		CreateAt: &apis.Time{time.Now()},
	//		LastTime: &apis.Time{time.Now()},
	//	},
	//}
	//
	//data2 := &apis.Data{
	//	ObjectMeta: metav1.ObjectMeta{
	//		Name:      "Cosmo-workflow1",
	//		Namespace: "Cosmo",
	//	},
	//	TypeMeta: metav1.TypeMeta{
	//		Kind:       "Data",
	//		APIVersion: "resources/v1",
	//	},
	//	Spec: apis.DataSpec{
	//		Name:       "Cosmo-workflow1",
	//		BelongNode: "n19",
	//		FilePath:   "/root/goprojects/workflow/zhizao.json",
	//		FileFormat: "json",
	//		Desc: &apis.Description{
	//			Docs: "北航-卡奥斯工作流-Command",
	//		},
	//	},
	//	Status: apis.DataStatus{
	//		CreateAt: &apis.Time{time.Now()},
	//		LastTime: &apis.Time{time.Now()},
	//	},
	//}
	//data3 := &apis.Data{
	//	ObjectMeta: metav1.ObjectMeta{
	//		Name:      "ShandongHS-workflow1",
	//		Namespace: "ShandongHS",
	//	},
	//	TypeMeta: metav1.TypeMeta{
	//		Kind:       "Data",
	//		APIVersion: "resources/v1",
	//	},
	//	Spec: apis.DataSpec{
	//		Name:       "ShandongHS-workflow1",
	//		BelongNode: "n19",
	//		FilePath:   "/root/goprojects/workflow/jiaotong.json",
	//		FileFormat: "json",
	//		Desc: &apis.Description{
	//			Docs: "北航-高速工作流-Command",
	//		},
	//	},
	//	Status: apis.DataStatus{
	//		CreateAt: &apis.Time{time.Now()},
	//		LastTime: &apis.Time{time.Now()},
	//	},
	//}
	//
	//d, err := manager.CreateData(data, namespace1)
	//if err != nil {
	//	panic(err)
	//}
	//fmt.Println(d)
	//
	//d, err = manager.CreateData(data2, namespace2)
	//if err != nil {
	//	panic(err)
	//}
	//fmt.Println(d)
	//
	//d, err = manager.CreateData(data3, namespace3)
	//if err != nil {
	//	panic(err)
	//}
	//fmt.Println(d)

	//data := &apis.Data{
	//	ObjectMeta: metav1.ObjectMeta{
	//		Name:      "HenanEP-train2",
	//		Namespace: "HenanEP",
	//	},
	//	TypeMeta: metav1.TypeMeta{
	//		Kind:       "Data",
	//		APIVersion: "resources/v1",
	//	},
	//	Spec: apis.DataSpec{
	//		Name:       "HenanEP-train2",
	//		BelongNode: "n19",
	//		FilePath:   "/root/goprojects/workflow/train_dianwang_pod.json",
	//		FileFormat: "json",
	//		Desc: &apis.Description{
	//			Docs: "模型训练工作流-Pod",
	//		},
	//	},
	//	Status: apis.DataStatus{
	//		CreateAt: &apis.Time{time.Now()},
	//		LastTime: &apis.Time{time.Now()},
	//	},
	//}
	//
	//data2 := &apis.Data{
	//	ObjectMeta: metav1.ObjectMeta{
	//		Name:      "Cosmo-train2",
	//		Namespace: "Cosmo",
	//	},
	//	TypeMeta: metav1.TypeMeta{
	//		Kind:       "Data",
	//		APIVersion: "resources/v1",
	//	},
	//	Spec: apis.DataSpec{
	//		Name:       "Cosmo-train2",
	//		BelongNode: "n19",
	//		FilePath:   "/root/goprojects/workflow/train_zhizao_pod.json",
	//		FileFormat: "json",
	//		Desc: &apis.Description{
	//			Docs: "模型训练工作流-Pod",
	//		},
	//	},
	//	Status: apis.DataStatus{
	//		CreateAt: &apis.Time{time.Now()},
	//		LastTime: &apis.Time{time.Now()},
	//	},
	//}
	//data3 := &apis.Data{
	//	ObjectMeta: metav1.ObjectMeta{
	//		Name:      "ShandongHS-train2",
	//		Namespace: "ShandongHS",
	//	},
	//	TypeMeta: metav1.TypeMeta{
	//		Kind:       "Data",
	//		APIVersion: "resources/v1",
	//	},
	//	Spec: apis.DataSpec{
	//		Name:       "ShandongHS-train2",
	//		BelongNode: "n19",
	//		FilePath:   "/root/goprojects/workflow/train_jiaotong_pod.json",
	//		FileFormat: "json",
	//		Desc: &apis.Description{
	//			Docs: "模型训练工作流-Pod",
	//		},
	//	},
	//	Status: apis.DataStatus{
	//		CreateAt: &apis.Time{time.Now()},
	//		LastTime: &apis.Time{time.Now()},
	//	},
	//}
	//
	//d, err := manager.CreateData(data, namespace1)
	//if err != nil {
	//	panic(err)
	//}
	//fmt.Println(d)
	//
	//d, err = manager.CreateData(data2, namespace2)
	//if err != nil {
	//	panic(err)
	//}
	//fmt.Println(d)
	//
	//d, err = manager.CreateData(data3, namespace3)
	//if err != nil {
	//	panic(err)
	//}
	//fmt.Println(d)

	//data := &apis.Data{
	//	ObjectMeta: metav1.ObjectMeta{
	//		Name:      "HenanEP-switch1",
	//		Namespace: "HenanEP",
	//	},
	//	TypeMeta: metav1.TypeMeta{
	//		Kind:       "Data",
	//		APIVersion: "resources/v1",
	//	},
	//	Spec: apis.DataSpec{
	//		Name:       "HenanEP-switch1",
	//		BelongNode: "n19",
	//		FilePath:   "/root/goprojects/workflow/switch_dianwang.json",
	//		FileFormat: "json",
	//		Desc: &apis.Description{
	//			Docs: "推理迁移工作流-Command",
	//		},
	//	},
	//	Status: apis.DataStatus{
	//		CreateAt: &apis.Time{time.Now()},
	//		LastTime: &apis.Time{time.Now()},
	//	},
	//}
	//
	//data2 := &apis.Data{
	//	ObjectMeta: metav1.ObjectMeta{
	//		Name:      "Cosmo-switch1",
	//		Namespace: "Cosmo",
	//	},
	//	TypeMeta: metav1.TypeMeta{
	//		Kind:       "Data",
	//		APIVersion: "resources/v1",
	//	},
	//	Spec: apis.DataSpec{
	//		Name:       "Cosmo-switch1",
	//		BelongNode: "n19",
	//		FilePath:   "/root/goprojects/workflow/switch_zhizao.json",
	//		FileFormat: "json",
	//		Desc: &apis.Description{
	//			Docs: "推理迁移工作流-Command",
	//		},
	//	},
	//	Status: apis.DataStatus{
	//		CreateAt: &apis.Time{time.Now()},
	//		LastTime: &apis.Time{time.Now()},
	//	},
	//}
	//data3 := &apis.Data{
	//	ObjectMeta: metav1.ObjectMeta{
	//		Name:      "ShandongHS-switch1",
	//		Namespace: "ShandongHS",
	//	},
	//	TypeMeta: metav1.TypeMeta{
	//		Kind:       "Data",
	//		APIVersion: "resources/v1",
	//	},
	//	Spec: apis.DataSpec{
	//		Name:       "ShandongHS-switch1",
	//		BelongNode: "n19",
	//		FilePath:   "/root/goprojects/workflow/switch_jiaotong.json",
	//		FileFormat: "json",
	//		Desc: &apis.Description{
	//			Docs: "推理迁移工作流-Command",
	//		},
	//	},
	//	Status: apis.DataStatus{
	//		CreateAt: &apis.Time{time.Now()},
	//		LastTime: &apis.Time{time.Now()},
	//	},
	//}
	//
	//data4 := &apis.Data{
	//	ObjectMeta: metav1.ObjectMeta{
	//		Name:      "HenanEP-switch2",
	//		Namespace: "HenanEP",
	//	},
	//	TypeMeta: metav1.TypeMeta{
	//		Kind:       "Data",
	//		APIVersion: "resources/v1",
	//	},
	//	Spec: apis.DataSpec{
	//		Name:       "HenanEP-switch2",
	//		BelongNode: "n19",
	//		FilePath:   "/root/goprojects/workflow/switch_grpc_dianwang.json",
	//		FileFormat: "json",
	//		Desc: &apis.Description{
	//			Docs: "grpc服务迁移工作流-Pod",
	//		},
	//	},
	//	Status: apis.DataStatus{
	//		CreateAt: &apis.Time{time.Now()},
	//		LastTime: &apis.Time{time.Now()},
	//	},
	//}
	//
	//data5 := &apis.Data{
	//	ObjectMeta: metav1.ObjectMeta{
	//		Name:      "Cosmo-switch2",
	//		Namespace: "Cosmo",
	//	},
	//	TypeMeta: metav1.TypeMeta{
	//		Kind:       "Data",
	//		APIVersion: "resources/v1",
	//	},
	//	Spec: apis.DataSpec{
	//		Name:       "Cosmo-switch2",
	//		BelongNode: "n19",
	//		FilePath:   "/root/goprojects/workflow/switch_grpc_zhizao.json",
	//		FileFormat: "json",
	//		Desc: &apis.Description{
	//			Docs: "grpc服务迁移工作流-Pod",
	//		},
	//	},
	//	Status: apis.DataStatus{
	//		CreateAt: &apis.Time{time.Now()},
	//		LastTime: &apis.Time{time.Now()},
	//	},
	//}
	//data6 := &apis.Data{
	//	ObjectMeta: metav1.ObjectMeta{
	//		Name:      "ShandongHS-switch2",
	//		Namespace: "ShandongHS",
	//	},
	//	TypeMeta: metav1.TypeMeta{
	//		Kind:       "Data",
	//		APIVersion: "resources/v1",
	//	},
	//	Spec: apis.DataSpec{
	//		Name:       "ShandongHS-switch2",
	//		BelongNode: "n19",
	//		FilePath:   "/root/goprojects/workflow/switch_grpc_jiaotong.json",
	//		FileFormat: "json",
	//		Desc: &apis.Description{
	//			Docs: "grpc服务迁移工作流-Pod",
	//		},
	//	},
	//	Status: apis.DataStatus{
	//		CreateAt: &apis.Time{time.Now()},
	//		LastTime: &apis.Time{time.Now()},
	//	},
	//}
	//
	//data7 := &apis.Data{
	//	ObjectMeta: metav1.ObjectMeta{
	//		Name:      "HenanEP-train1",
	//		Namespace: "HenanEP",
	//	},
	//	TypeMeta: metav1.TypeMeta{
	//		Kind:       "Data",
	//		APIVersion: "resources/v1",
	//	},
	//	Spec: apis.DataSpec{
	//		Name:       "HenanEP-train1",
	//		BelongNode: "n19",
	//		FilePath:   "/root/goprojects/workflow/infer_dianwang.json",
	//		FileFormat: "json",
	//		Desc: &apis.Description{
	//			Docs: "训练工作流-Command",
	//		},
	//	},
	//	Status: apis.DataStatus{
	//		CreateAt: &apis.Time{time.Now()},
	//		LastTime: &apis.Time{time.Now()},
	//	},
	//}
	//
	//data8 := &apis.Data{
	//	ObjectMeta: metav1.ObjectMeta{
	//		Name:      "Cosmo-train1",
	//		Namespace: "Cosmo",
	//	},
	//	TypeMeta: metav1.TypeMeta{
	//		Kind:       "Data",
	//		APIVersion: "resources/v1",
	//	},
	//	Spec: apis.DataSpec{
	//		Name:       "Cosmo-train1",
	//		BelongNode: "n19",
	//		FilePath:   "/root/goprojects/workflow/infer_zhizao.json",
	//		FileFormat: "json",
	//		Desc: &apis.Description{
	//			Docs: "训练工作流-Command",
	//		},
	//	},
	//	Status: apis.DataStatus{
	//		CreateAt: &apis.Time{time.Now()},
	//		LastTime: &apis.Time{time.Now()},
	//	},
	//}
	//data9 := &apis.Data{
	//	ObjectMeta: metav1.ObjectMeta{
	//		Name:      "ShandongHS-train1",
	//		Namespace: "ShandongHS",
	//	},
	//	TypeMeta: metav1.TypeMeta{
	//		Kind:       "Data",
	//		APIVersion: "resources/v1",
	//	},
	//	Spec: apis.DataSpec{
	//		Name:       "ShandongHS-train1",
	//		BelongNode: "n19",
	//		FilePath:   "/root/goprojects/workflow/infer_jiaotong.json",
	//		FileFormat: "json",
	//		Desc: &apis.Description{
	//			Docs: "训练工作流-Command",
	//		},
	//	},
	//	Status: apis.DataStatus{
	//		CreateAt: &apis.Time{time.Now()},
	//		LastTime: &apis.Time{time.Now()},
	//	},
	//}
	//
	//d, err := manager.CreateData(data, namespace1)
	//if err != nil {
	//	panic(err)
	//}
	//fmt.Println(d)
	//
	//d, err = manager.CreateData(data2, namespace2)
	//if err != nil {
	//	panic(err)
	//}
	//fmt.Println(d)
	//
	//d, err = manager.CreateData(data3, namespace3)
	//if err != nil {
	//	panic(err)
	//}
	//fmt.Println(d)
	//d, err = manager.CreateData(data4, namespace1)
	//if err != nil {
	//	panic(err)
	//}
	//fmt.Println(d)
	//
	//d, err = manager.CreateData(data5, namespace2)
	//if err != nil {
	//	panic(err)
	//}
	//fmt.Println(d)
	//
	//d, err = manager.CreateData(data6, namespace3)
	//if err != nil {
	//	panic(err)
	//}
	//fmt.Println(d)
	//d, err = manager.CreateData(data7, namespace1)
	//if err != nil {
	//	panic(err)
	//}
	//fmt.Println(d)
	//
	//d, err = manager.CreateData(data8, namespace2)
	//if err != nil {
	//	panic(err)
	//}
	//fmt.Println(d)
	//
	//d, err = manager.CreateData(data9, namespace3)
	//if err != nil {
	//	panic(err)
	//}
	//fmt.Println(d)

	// 卡奥斯-docker
	//data := &apis.Data{
	//	ObjectMeta: metav1.ObjectMeta{
	//		Name:      "Cosmo-task1",
	//		Namespace: "HenanEP",
	//	},
	//	TypeMeta: metav1.TypeMeta{
	//		Kind:       "Data",
	//		APIVersion: "resources/v1",
	//	},
	//	Spec: apis.DataSpec{
	//		Name:       "Cosmo-task1",
	//		BelongNode: "n19",
	//		FilePath:   "/home/lkcoffee/Desktop/proj/reference/bin/tmp/data/docker.json",
	//		FileFormat: "json",
	//		Desc: &apis.Description{
	//			Docs: "相机识别工作流-Docker",
	//		},
	//	},
	//	Status: apis.DataStatus{
	//		CreateAt: &apis.Time{time.Now()},
	//		LastTime: &apis.Time{time.Now()},
	//	},
	//}
	//
	//d, err := manager.CreateData(data, namespace2)
	//if err != nil {
	//	panic(err)
	//}
	//fmt.Println(d)

	data := &apis.Data{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "HenanEP-infer1",
			Namespace: "HenanEP",
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Data",
			APIVersion: "resources/v1",
		},
		Spec: apis.DataSpec{
			Name:       "HenanEP-infer1",
			BelongNode: "n19",
			FilePath:   "/root/goprojects/workflow/dianwang_infer.json",
			FileFormat: "json",
			Desc: &apis.Description{
				Docs: "推理工作流-Command",
			},
		},
		Status: apis.DataStatus{
			CreateAt: &apis.Time{time.Now()},
			LastTime: &apis.Time{time.Now()},
		},
	}

	data2 := &apis.Data{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "Cosmo-infer1",
			Namespace: "Cosmo",
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Data",
			APIVersion: "resources/v1",
		},
		Spec: apis.DataSpec{
			Name:       "Cosmo-infer1",
			BelongNode: "n19",
			FilePath:   "/root/goprojects/workflow/zhizao_infer.json",
			FileFormat: "json",
			Desc: &apis.Description{
				Docs: "推理工作流-Command",
			},
		},
		Status: apis.DataStatus{
			CreateAt: &apis.Time{time.Now()},
			LastTime: &apis.Time{time.Now()},
		},
	}
	data3 := &apis.Data{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ShandongHS-infer1",
			Namespace: "ShandongHS",
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Data",
			APIVersion: "resources/v1",
		},
		Spec: apis.DataSpec{
			Name:       "ShandongHS-infer1",
			BelongNode: "n19",
			FilePath:   "/root/goprojects/workflow/jiaotong_infer.json",
			FileFormat: "json",
			Desc: &apis.Description{
				Docs: "推理工作流-Command",
			},
		},
		Status: apis.DataStatus{
			CreateAt: &apis.Time{time.Now()},
			LastTime: &apis.Time{time.Now()},
		},
	}

	d, err := manager.CreateData(data, namespace1)
	if err != nil {
		panic(err)
	}
	fmt.Println(d)

	d, err = manager.CreateData(data2, namespace2)
	if err != nil {
		panic(err)
	}
	fmt.Println(d)

	d, err = manager.CreateData(data3, namespace3)
	if err != nil {
		panic(err)
	}
	fmt.Println(d)
}

//func TestGetData(t *testing.T) {
//	name := "demo-data2"
//	namespace := "Guochuang"
//
//	clientset, err := CreateClientSet()
//	manager := manager.NewManager(clientset)
//	d, err := manager.GetData(name, namespace)
//	if err != nil {
//		panic(err)
//	}
//	fmt.Println(d)
//}
//
//func TestFilterData(t *testing.T) {
//	clientset, err := CreateClientSet()
//	manager := manager.NewManager(clientset)
//
//	namespace := "Guochuang"
//
//	label := "environment=main"
//
//	d, err := manager.FilterDatas(namespace, label)
//	if err != nil {
//		panic(err)
//	}
//
//	for _, j := range d.Items {
//		fmt.Println(j)
//	}
//
//	label = "environment=dev"
//
//	d, err = manager.FilterDatas(namespace, label)
//	if err != nil {
//		panic(err)
//	}
//
//	fmt.Println("-------------")
//	for _, j := range d.Items {
//		fmt.Println(j)
//	}
//}
//
//func TestGetDatas(t *testing.T) {
//	clientset, err := CreateClientSet()
//	manager := manager.NewManager(clientset)
//
//	lst, err := manager.FilterDatas("Guochuang", "")
//	if err != nil {
//		panic(err)
//	} else {
//		str, err := analyzer.SerializeToJson(lst)
//		if err != nil {
//			return
//		}
//		fmt.Println(str)
//	}
//}
//
//func TestPatchData(t *testing.T) {
//	namespace := "Guochuang"
//	clientset, err := CreateClientSet()
//	manager := manager.NewManager(clientset)
//	if err != nil {
//		panic(err)
//	}
//
//	data := &apis.Data{
//		ObjectMeta: metav1.ObjectMeta{
//			Name:      "demo-datas-patch",
//			Namespace: "test",
//			Labels: map[string]string{
//				"environment": "main",
//			},
//		},
//		TypeMeta: metav1.TypeMeta{
//			Kind:       "Data",
//			APIVersion: "resources/v1",
//		},
//	}
//
//	d, err := manager.CreateData(data, namespace)
//	if err != nil {
//		panic(err)
//	} else {
//		str, err := analyzer.SerializeToJson(d)
//		if err != nil {
//			return
//		}
//		fmt.Println(str)
//	}
//
//	patchData := "{\n    \"labels\" : {\n        \"environments\" : \"testPatch\"\n    }\n}"
//	d, err = manager.PatchData(d.Name, d.Namespace, []byte(patchData))
//	if err != nil {
//		panic(err)
//	} else {
//		str, err := analyzer.SerializeToJson(d)
//		if err != nil {
//			return
//		}
//		fmt.Println(str)
//	}
//}
//
//func TestDeleteData(t *testing.T) {
//	namespace := "Guochuang"
//	clientset, err := CreateClientSet()
//	manager := manager.NewManager(clientset)
//	if err != nil {
//		panic(err)
//	}
//
//	data := &apis.Data{
//		ObjectMeta: metav1.ObjectMeta{
//			Name:      "demo-datas-delete",
//			Namespace: "test",
//			Labels: map[string]string{
//				"environment": "main",
//			},
//		},
//		TypeMeta: metav1.TypeMeta{
//			Kind:       "Data",
//			APIVersion: "resources/v1",
//		},
//	}
//
//	d, err := manager.CreateData(data, namespace)
//	if err != nil {
//		panic(err)
//	} else {
//		str, err := analyzer.SerializeToJson(d)
//		if err != nil {
//			return
//		}
//		fmt.Println(str)
//	}
//
//	err = manager.DeleteData(d.Name, d.Namespace)
//	if err != nil {
//		panic(err)
//	}
//
//}
//
//func TestUpdateData(t *testing.T) {
//	namespace := "Guochuang"
//	clientset, err := CreateClientSet()
//	manager := manager.NewManager(clientset)
//	if err != nil {
//		panic(err)
//	}
//
//	data := &apis.Data{
//		ObjectMeta: metav1.ObjectMeta{
//			Name:      "demo-datas-delete",
//			Namespace: "test",
//			Labels: map[string]string{
//				"environment": "main",
//			},
//		},
//		TypeMeta: metav1.TypeMeta{
//			Kind:       "Data",
//			APIVersion: "resources/v1",
//		},
//	}
//
//	d, err := manager.CreateData(data, namespace)
//	if err != nil {
//		panic(err)
//	} else {
//		str, err := analyzer.SerializeToJson(d)
//		if err != nil {
//			return
//		}
//		fmt.Println(str)
//	}
//
//	d.Labels["testUpdate"] = "update"
//
//	d, err = manager.UpdateData(d.Name, d.Namespace, d)
//	if err != nil {
//		panic(err)
//	} else {
//		str, err := analyzer.SerializeToJson(d)
//		if err != nil {
//			return
//		}
//		fmt.Println(str)
//	}
//}

//var scheme = runtime.NewScheme()
//
//func main() {
//	moduleName := "testModule"
//	logs.Init(moduleName)
//
//	//创建ClientSet
//	clientSet := initClientSet(scheme)
//	m := manager.NewManager(clientSet)
//	data := &apis.Data{
//		Spec: apis.DataSpec{
//			Name:       "",
//			BelongNode: "n19",
//			FilePath:   "/root/goprojects/workflow",
//			FileFormat: "",
//			SizeBytes:  0,
//			AccessMode: "",
//		},
//		Status: apis.DataStatus{
//			CreateAt: &apis.Time{time.Now()},
//			LastTime: &apis.Time{time.Now()},
//		},
//	}
//	data, err := m.CreateData(data, "test")
//	if err != nil {
//		logs.Error("error")
//	}
//	str, err := analyzer.SerializeToJson(&data)
//	if err != nil {
//		return
//	}
//	fmt.Println(str)
//
//}
//func initClientSet(scheme *runtime.Scheme) *clients.ClientSet {
//	apis.AddToScheme(scheme)
//	logs.Info(scheme)
//	// 创建ClientSet
//	c := &rest.Config{
//		Host:    "http://127.0.0.1:10000",
//		APIPath: "/apis/resources/v1",
//		ContentConfig: rest.ContentConfig{
//			AcceptContentTypes: "application/json; charset=UTF-8", //text/plain; charset=UTF-8
//			ContentType:        "application/json; charset=UTF-8", //application/json; charset=UTF-8
//			GroupVersion: &schema.GroupVersion{
//				Group:   "resources",
//				Version: "v1",
//			},
//			NegotiatedSerializer: serializer.NewCodecFactory(scheme),
//		},
//		UserAgent: "defaultUserAgent",
//		Transport: &http.Transport{
//			MaxIdleConns:        100,              // 最大空闲连接数
//			IdleConnTimeout:     90 * time.Second, // 空闲连接超时时间
//			TLSHandshakeTimeout: 10 * time.Second, // TLS 握手超时时间
//		},
//		Timeout: 10 * time.Second,
//	}
//	clientSet, err := clients.NewForConfig(c)
//	if err != nil {
//		panic(err)
//	}
//	return clientSet
//}


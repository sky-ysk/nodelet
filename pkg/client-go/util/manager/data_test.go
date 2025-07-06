package manager

import (
	"fmt"
	apis "hit.edu/framework/pkg/apis/cores"
	metav1 "hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/component-base/analyzer"
	"testing"
)

func TestCreateData(t *testing.T) {
	namespace := "Guochuang"
	clientset, err := CreateClientSet()
	manager := NewManager(clientset)
	if err != nil {
		panic(err)
	}

	data := &apis.Data{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "demo-datas-main",
			Namespace: "test",
			Labels: map[string]string{
				"environment": "main",
			},
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Data",
			APIVersion: "resources/v1",
		},
	}

	data2 := &apis.Data{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "demo-data2-main",
			Namespace: "test",
			Labels: map[string]string{
				"environment": "main",
			},
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Data",
			APIVersion: "resources/v1",
		},
	}

	d, err := manager.CreateData(data, namespace)
	if err != nil {
		panic(err)
	}
	fmt.Println(d)

	d, err = manager.CreateData(data2, namespace)
	if err != nil {
		panic(err)
	}
	fmt.Println(d)
}

func TestGetData(t *testing.T) {
	name := "demo-data2"
	namespace := "Guochuang"

	clientset, err := CreateClientSet()
	manager := NewManager(clientset)
	d, err := manager.GetData(name, namespace)
	if err != nil {
		panic(err)
	}
	fmt.Println(d)
}

func TestFilterData(t *testing.T) {
	clientset, err := CreateClientSet()
	manager := NewManager(clientset)

	namespace := "Guochuang"

	label := "environment=main"

	d, err := manager.FilterDatas(namespace, label)
	if err != nil {
		panic(err)
	}

	for _, j := range d.Items {
		fmt.Println(j)
	}

	label = "environment=dev"

	d, err = manager.FilterDatas(namespace, label)
	if err != nil {
		panic(err)
	}

	fmt.Println("-------------")
	for _, j := range d.Items {
		fmt.Println(j)
	}
}

func TestGetDatas(t *testing.T) {
	clientset, err := CreateClientSet()
	manager := NewManager(clientset)

	lst, err := manager.FilterDatas("Guochuang", "")
	if err != nil {
		panic(err)
	} else {
		str, err := analyzer.SerializeToJson(lst)
		if err != nil {
			return
		}
		fmt.Println(str)
	}
}

func TestPatchData(t *testing.T) {
	namespace := "Guochuang"
	clientset, err := CreateClientSet()
	manager := NewManager(clientset)
	if err != nil {
		panic(err)
	}

	data := &apis.Data{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "demo-datas-patch",
			Namespace: "test",
			Labels: map[string]string{
				"environment": "main",
			},
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Data",
			APIVersion: "resources/v1",
		},
	}

	d, err := manager.CreateData(data, namespace)
	if err != nil {
		panic(err)
	} else {
		str, err := analyzer.SerializeToJson(d)
		if err != nil {
			return
		}
		fmt.Println(str)
	}

	patchData := "{\n    \"labels\" : {\n        \"environments\" : \"testPatch\"\n    }\n}"
	d, err = manager.PatchData(d.Name, d.Namespace, []byte(patchData))
	if err != nil {
		panic(err)
	} else {
		str, err := analyzer.SerializeToJson(d)
		if err != nil {
			return
		}
		fmt.Println(str)
	}
}

func TestDeleteData(t *testing.T) {
	namespace := "Guochuang"
	clientset, err := CreateClientSet()
	manager := NewManager(clientset)
	if err != nil {
		panic(err)
	}

	data := &apis.Data{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "demo-datas-delete",
			Namespace: "test",
			Labels: map[string]string{
				"environment": "main",
			},
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Data",
			APIVersion: "resources/v1",
		},
	}

	d, err := manager.CreateData(data, namespace)
	if err != nil {
		panic(err)
	} else {
		str, err := analyzer.SerializeToJson(d)
		if err != nil {
			return
		}
		fmt.Println(str)
	}

	err = manager.DeleteData(d.Name, d.Namespace)
	if err != nil {
		panic(err)
	}

}

func TestUpdateData(t *testing.T) {
	namespace := "Guochuang"
	clientset, err := CreateClientSet()
	manager := NewManager(clientset)
	if err != nil {
		panic(err)
	}

	data := &apis.Data{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "demo-datas-delete",
			Namespace: "test",
			Labels: map[string]string{
				"environment": "main",
			},
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Data",
			APIVersion: "resources/v1",
		},
	}

	d, err := manager.CreateData(data, namespace)
	if err != nil {
		panic(err)
	} else {
		str, err := analyzer.SerializeToJson(d)
		if err != nil {
			return
		}
		fmt.Println(str)
	}

	d.Labels["testUpdate"] = "update"

	d, err = manager.UpdateData(d.Name, d.Namespace, d)
	if err != nil {
		panic(err)
	} else {
		str, err := analyzer.SerializeToJson(d)
		if err != nil {
			return
		}
		fmt.Println(str)
	}
}

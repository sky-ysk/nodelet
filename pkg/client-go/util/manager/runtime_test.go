package manager

import (
	"fmt"
	"github.com/google/uuid"
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/component-base/analyzer"
	"hit.edu/framework/pkg/component-base/logs"
	"testing"
)

func TestCreateRuntime(t *testing.T) {

	clientset, err := CreateClientSet()
	if err != nil {
		panic(err)
	}
	// 构造Manager
	m := NewManager(clientset)

	logs.Init("main")
	rs := apis.RuntimeSpec{
		Name:  "R1",
		Type:  apis.ByDevice,
		Image: "xxxxx",
	}

	// 生成UUID
	u := uuid.Must(uuid.NewV7())

	runtime, err := m.CreateRuntime(rs, nil, "Guochuang", u.String(), "")
	if err != nil {
		return
	}
	str, err := analyzer.SerializeToJson(runtime)
	if err != nil {
		return
	}

	fmt.Println(str)
}

func TestCreateRuntimes(t *testing.T) {
	clientset, err := CreateClientSet()
	if err != nil {
		panic(err)
	}
	// 构造Manager
	m := NewManager(clientset)

	logs.Init("main")

	// 生成UUID
	u := uuid.Must(uuid.NewV7())

	//

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
	a := apis.Action{
		ObjectMeta: meta.ObjectMeta{
			Name: "A1",
		},
		Spec: apis.ActionSpec{
			Runtimes: []apis.RuntimeSpec{
				rs1,
				rs2,
			},
		},
	}

	runtimes, err := m.CreateRuntimes(&a, "Guochaung", u.String(), "A1.")
	if err != nil {
		panic(err)
	}
	fmt.Println(runtimes)
}

func TestGetRuntime(t *testing.T) {
	clientset, err := CreateClientSet()
	if err != nil {
		panic(err)
	}
	// 构造Manager
	m := NewManager(clientset)

	logs.Init("main")
	rs := apis.RuntimeSpec{
		Name:  "R1",
		Type:  apis.ByDevice,
		Image: "xxxxx",
	}

	// 生成UUID
	u := uuid.Must(uuid.NewV7())

	runtime, err := m.CreateRuntime(rs, nil, "Guochuang", u.String(), "")
	if err != nil {
		return
	}

	a1, err := m.GetRuntime(runtime.Name, "Guochuang")
	if err != nil {
		panic(err)
	} else {
		str, err := analyzer.SerializeToJson(&a1)
		if err != nil {
			return
		}
		fmt.Println(str)
	}
}

func TestGetRuntimes(t *testing.T) {
	clientset, err := CreateClientSet()
	if err != nil {
		panic(err)
	}
	// 构造Manager
	m := NewManager(clientset)

	logs.Init("main")

	a, err := m.GetRuntimes("Guochuang")
	if err != nil {
		panic(err)
	} else {
		str, err := analyzer.SerializeToJson(a)
		if err != nil {
			return
		}
		fmt.Println(str)
	}
}

func TestUpdateRuntime(t *testing.T) {
	clientset, err := CreateClientSet()
	if err != nil {
		panic(err)
	}
	// 构造Manager
	m := NewManager(clientset)

	parent := []string{"runtimeTest1"}

	logs.Init("main")
	rs := apis.RuntimeSpec{
		Name:    "R1",
		Parents: parent,
		Type:    apis.ByDevice,
		Image:   "xxxxx",
	}

	// 生成UUID
	u := uuid.Must(uuid.NewV7())

	runtime, err := m.CreateRuntime(rs, nil, "Guochuang", u.String(), "")
	if err != nil {
		return
	}
	str, err := analyzer.SerializeToJson(runtime)
	if err != nil {
		return
	}

	fmt.Println(str)

	parent1 := []string{"runtimeTest1", "runtimeTest2"}
	runtime.Spec.Parents = parent1

	// 修改parent
	patched, err := m.UpdateRuntime(runtime.Name, "Guochuang", runtime)
	if err != nil {
		panic(err)
	} else {
		str, err := analyzer.SerializeToJson(&patched)
		if err != nil {
			return
		}
		fmt.Println(str)
	}
}

func TestPatchRuntime(t *testing.T) {
	clientset, err := CreateClientSet()
	if err != nil {
		panic(err)
	}
	// 构造Manager
	m := NewManager(clientset)

	parent := []string{"runtimeTest1"}

	logs.Init("main")
	rs := apis.RuntimeSpec{
		Name:    "R1",
		Parents: parent,
		Type:    apis.ByDevice,
		Image:   "xxxxx",
	}

	// 生成UUID
	u := uuid.Must(uuid.NewV7())

	runtime, err := m.CreateRuntime(rs, nil, "Guochuang", u.String(), "")
	if err != nil {
		return
	}
	str, err := analyzer.SerializeToJson(runtime)
	if err != nil {
		return
	}

	fmt.Println(str)

	// 修改parent
	patchData := "{\"spec\": {\n    \"parents\": [\n      \"runtimeTest1\",\n      \"runtimeTest2\"\n    ]\n  }}"
	patched, err := m.PatchRuntime(runtime.Name, "Guochuang", []byte(patchData))
	if err != nil {
		panic(err)
	} else {
		str, err := analyzer.SerializeToJson(&patched)
		if err != nil {
			return
		}
		fmt.Println(str)
	}

}

func TestDeleteRuntime(t *testing.T) {

	clientset, err := CreateClientSet()
	if err != nil {
		panic(err)
	}
	// 构造Manager
	m := NewManager(clientset)

	logs.Init("main")
	rs := apis.RuntimeSpec{
		Name:  "R1",
		Type:  apis.ByDevice,
		Image: "xxxxx",
	}

	// 生成UUID
	u := uuid.Must(uuid.NewV7())

	runtime, err := m.CreateRuntime(rs, nil, "Guochuang", u.String(), "")
	if err != nil {
		return
	}

	err = m.DeleteRuntime(runtime.Name, "Guochuang")
	if err != nil {
		panic(err)
	} else {
		fmt.Println("Delete runtime success")
	}
}

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

func TestCreateAction(t *testing.T) {

	clientset, err := CreateClientSet()
	if err != nil {
		panic(err)
	}
	// 构造Manager
	m := NewManager(clientset)

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

	as := apis.ActionSpec{
		Name: "A1",
		Runtimes: []apis.RuntimeSpec{
			rs1,
			rs2,
		},
	}

	a, err := m.CreateAction(as, nil, "Guochuang", u.String(), "")
	if err != nil {
		panic(err)
	}

	str, err := analyzer.SerializeToJson(a)
	if err != nil {
		return
	}

	fmt.Println(str)
}

func TestCreateActions(t *testing.T) {
	clientset, err := CreateClientSet()
	if err != nil {
		panic(err)
	}
	// 构造Manager
	m := NewManager(clientset)

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

	gs := apis.GroupSpec{
		Name: "G1",
		Actions: []apis.ActionSpec{
			as1,
			as2,
		},
	}

	g := apis.Group{
		ObjectMeta: meta.ObjectMeta{
			Name: "G1",
		},
		Spec: gs,
	}

	actions, err := m.CreateActions(&g, "Guochaung", u.String(), "G1.")
	if err != nil {
		panic(err)
	}
	fmt.Println(actions)
}

func TestGetActions(t *testing.T) {
	clientset, err := CreateClientSet()
	if err != nil {
		panic(err)
	}
	// 构造Manager
	m := NewManager(clientset)

	logs.Init("main")

	a, err := m.GetActions("Guochuang")
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

func TestGetAction(t *testing.T) {

	clientset, err := CreateClientSet()
	if err != nil {
		panic(err)
	}
	// 构造Manager
	m := NewManager(clientset)

	logs.Init("main")

	// 生成UUID
	u := uuid.Must(uuid.NewV7())
	rs1 := apis.RuntimeSpec{
		Name:  "R10",
		Type:  apis.ByDevice,
		Image: "xxxxx",
	}
	rs2 := apis.RuntimeSpec{
		Name:  "R20",
		Type:  apis.ByDevice,
		Image: "xxxxx",
	}

	as := apis.ActionSpec{
		Name: "A10",
		Runtimes: []apis.RuntimeSpec{
			rs1,
			rs2,
		},
	}

	a, err := m.CreateAction(as, nil, "Guochuang", u.String(), "")
	if err != nil {
		panic(err)
	}
	actionName := a.Name

	//str, err := analyzer.SerializeToJson(a)
	//if err != nil {
	//	return
	//}
	//
	//fmt.Println(str)

	a1, err := m.GetAction(actionName, "Guochuang")
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

func TestDeleteAction(t *testing.T) {

	clientset, err := CreateClientSet()
	if err != nil {
		panic(err)
	}
	// 构造Manager
	m := NewManager(clientset)

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

	as := apis.ActionSpec{
		Name: "A1",
		Runtimes: []apis.RuntimeSpec{
			rs1,
			rs2,
		},
	}

	a, err := m.CreateAction(as, nil, "Guochuang", u.String(), "")
	if err != nil {
		panic(err)
	}
	actionName := a.Name

	err = m.DeleteAction(actionName, "Guochuang")
	if err != nil {
		panic(err)
	} else {
		fmt.Println("Delete action success")
	}
}

func TestUpdateAction(t *testing.T) {
	clientset, err := CreateClientSet()
	if err != nil {
		panic(err)
	}
	// 构造Manager
	m := NewManager(clientset)

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

	desc := &apis.Description{
		Docs: "test before update",
	}

	as := apis.ActionSpec{
		Name: "A1",
		Desc: desc,
		Runtimes: []apis.RuntimeSpec{
			rs1,
			rs2,
		},
	}

	a, err := m.CreateAction(as, nil, "Guochuang", u.String(), "")
	if err != nil {
		panic(err)
	}

	str, err := analyzer.SerializeToJson(a)
	if err != nil {
		return
	}

	fmt.Println(str)

	desc1 := &apis.Description{
		Docs: "test after update",
	}

	a.Spec.Desc = desc1

	updatedAction, err := m.UpdateAction(a.Name, "Guochuang", a)
	if err != nil {
		panic(err)
	} else {
		str, err := analyzer.SerializeToJson(updatedAction)
		if err != nil {
			return
		}
		fmt.Println(str)
	}
}

func TestPatchAction(t *testing.T) {
	clientset, err := CreateClientSet()
	if err != nil {
		panic(err)
	}
	// 构造Manager
	m := NewManager(clientset)

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

	desc := &apis.Description{
		Docs: "抬手",
	}

	as := apis.ActionSpec{
		Name: "A1",
		Desc: desc,
		Runtimes: []apis.RuntimeSpec{
			rs1,
			rs2,
		},
	}

	a, err := m.CreateAction(as, nil, "Guochuang", u.String(), "")
	if err != nil {
		panic(err)
	}
	actionName := a.Name

	// 把抬手改成放下
	patchData := "{\n  \"spec\": {\n      \"desc\": {\n          \"docs\": \"放下\"\n      }\n  }\n\n}"

	patched, err := m.PatchAction(actionName, "Guochuang", []byte(patchData))
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

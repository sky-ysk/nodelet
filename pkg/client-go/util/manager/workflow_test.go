package manager

import (
	"fmt"
	"github.com/google/uuid"
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/component-base/analyzer"
	"hit.edu/framework/pkg/component-base/logs"
	"testing"
)

func TestCreateWorkflow(t *testing.T) {
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

	ts1 := apis.TaskSpec{
		Name: "T1",
		Groups: []apis.GroupSpec{
			gs1, gs2,
		},
	}

	ts2 := apis.TaskSpec{
		Name:   "T2",
		Groups: []apis.GroupSpec{},
	}

	ws := apis.WorkflowSpec{
		Name: "W1",
		Tasks: []apis.TaskSpec{
			ts1,
			ts2,
		},
	}

	w, err := m.CreateWorkflow(ws, "Guochuang", u.String())
	if err != nil {
		panic(err)
	}

	fmt.Println(w)
}

func TestGetWorkflows(t *testing.T) {
	clientset, err := CreateClientSet()
	if err != nil {
		panic(err)
	}
	// 构造Manager
	m := NewManager(clientset)

	logs.Init("main")

	a, err := m.GetWorkflows("Guochuang")
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

func TestGetWorkflow(t *testing.T) {

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

	ts1 := apis.TaskSpec{
		Name: "T1",
		Groups: []apis.GroupSpec{
			gs1, gs2,
		},
	}

	ts2 := apis.TaskSpec{
		Name:   "T2",
		Groups: []apis.GroupSpec{},
	}

	ws := apis.WorkflowSpec{
		Name: "W1",
		Tasks: []apis.TaskSpec{
			ts1,
			ts2,
		},
	}

	w, err := m.CreateWorkflow(ws, "Guochuang", u.String())
	if err != nil {
		panic(err)
	}

	fmt.Println(w)

	a1, err := m.GetWorkflow(w.Name, "Guochuang")
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

func TestDeleteWorkflow(t *testing.T) {

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

	ts1 := apis.TaskSpec{
		Name: "T1",
		Groups: []apis.GroupSpec{
			gs1, gs2,
		},
	}

	ts2 := apis.TaskSpec{
		Name:   "T2",
		Groups: []apis.GroupSpec{},
	}

	ws := apis.WorkflowSpec{
		Name: "W1",
		Tasks: []apis.TaskSpec{
			ts1,
			ts2,
		},
	}

	w, err := m.CreateWorkflow(ws, "Guochuang", u.String())
	if err != nil {
		panic(err)
	}

	fmt.Println(w)

	err = m.DeleteWorkflow(w.Name, "Guochuang")
	if err != nil {
		panic(err)
	} else {
		fmt.Println("Delete action success")
	}
}

func TestDeleteWorkflows(t *testing.T) {
	clientset, err := CreateClientSet()
	if err != nil {
		panic(err)
	}
	// 构造Manager
	m := NewManager(clientset)

	logs.Init("main")

	err = m.DeleteWorkflows("Guochuang")
	if err != nil {
		panic(err)
	} else {
		fmt.Println("Delete all workflows success")
	}
}

func TestUpdateWorkflow(t *testing.T) {
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

	ts1 := apis.TaskSpec{
		Name: "T1",
		Groups: []apis.GroupSpec{
			gs1, gs2,
		},
	}

	ts2 := apis.TaskSpec{
		Name:   "T2",
		Groups: []apis.GroupSpec{},
	}

	desc := &apis.Description{
		Docs: "test before update",
	}

	ws := apis.WorkflowSpec{
		Name: "W1",
		Desc: desc,
		Tasks: []apis.TaskSpec{
			ts1,
			ts2,
		},
	}

	w, err := m.CreateWorkflow(ws, "Guochuang", u.String())
	if err != nil {
		panic(err)
	}
	str, err := analyzer.SerializeToJson(&w)
	if err != nil {
		return
	}
	fmt.Println(str)

	desc1 := &apis.Description{
		Docs: "test after update",
	}

	w.Spec.Desc = desc1

	updated, err := m.UpdateWorkflow(w.Name, w.Namespace, w)
	if err != nil {
		panic(err)
	} else {
		str, err := analyzer.SerializeToJson(&updated)
		if err != nil {
			return
		}
		fmt.Println(str)
	}
}

func TestPatchWorkflow(t *testing.T) {
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

	ts1 := apis.TaskSpec{
		Name: "T1",
		Groups: []apis.GroupSpec{
			gs1, gs2,
		},
	}

	ts2 := apis.TaskSpec{
		Name:   "T2",
		Groups: []apis.GroupSpec{},
	}

	desc := &apis.Description{
		Docs: "test before patch",
	}

	ws := apis.WorkflowSpec{
		Name: "W1",
		Desc: desc,
		Tasks: []apis.TaskSpec{
			ts1,
			ts2,
		},
	}

	w, err := m.CreateWorkflow(ws, "Guochuang", u.String())
	if err != nil {
		panic(err)
	}
	str, err := analyzer.SerializeToJson(&w)
	if err != nil {
		return
	}

	fmt.Println(str)

	patchData := "{\n  \"spec\": {\n      \"desc\": {\n          \"docs\": \"test after patch\"\n      }\n  }\n\n}"

	patched, err := m.PatchWorkflow(w.Name, "Guochuang", []byte(patchData))
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

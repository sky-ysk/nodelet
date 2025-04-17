package manager

import (
	"fmt"
	"github.com/google/uuid"
	apis "hit.edu/framework/pkg/apis/cores"
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

package manager

import (
	"fmt"
	"github.com/google/uuid"
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/component-base/analyzer"
	"hit.edu/framework/pkg/component-base/logs"
	"testing"
	"time"
)

func TestCreateGroupWithoutActions(t *testing.T) {
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
	
	g, err := m.CreateGroupWithoutActions(gs, nil, "Guochuang", u.String(), "")
	if err != nil {
		panic(err)
	}
	
	time.Sleep(10 * time.Second)
	g, err = m.FillGroupWithActions(g)
	if err != nil {
		panic(err)
	}
	
	fmt.Println(g)
}

func TestCreateGroup(t *testing.T) {
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
	
	g, err := m.CreateGroup(gs, nil, "Guochuang", u.String(), "")
	if err != nil {
		panic(err)
	}
	
	fmt.Println(g)
}

func TestCreateGroups(t *testing.T) {
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
	
	ts := apis.TaskSpec{
		Name: "T1",
		Groups: []apis.GroupSpec{
			gs1, gs2,
		},
	}
	
	task := apis.Task{
		ObjectMeta: meta.ObjectMeta{
			Name: "T1",
		},
		Spec: ts,
	}
	
	g, err := m.CreateGroups(&task, "Guochuang", u.String(), "T1.")
	if err != nil {
		panic(err)
	}
	
	fmt.Println(g)
}

func TestGetGroups(t *testing.T) {
	clientset, err := CreateClientSet()
	if err != nil {
		panic(err)
	}
	// 构造Manager
	m := NewManager(clientset)
	
	logs.Init("main")
	
	alist, err := m.GetActions("Guochuang")
	if err != nil {
		panic(err)
	} else {
		
		for _, item := range alist.Items {
			str, err := analyzer.SerializeToJson(item)
			if err != nil {
				return
			}
			fmt.Println(str)
		}
	}
}

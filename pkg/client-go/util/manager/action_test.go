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

	a, err := m.CreateAction(as, nil, "Guochaung", u.String(), "")
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

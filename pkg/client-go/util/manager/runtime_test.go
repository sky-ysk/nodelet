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

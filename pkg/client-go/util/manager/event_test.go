package manager

import (
	"fmt"
	"github.com/google/uuid"
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/nodelet/events"
	"testing"
	"time"
)

func TestCreateEvent(t *testing.T) {
	namespace := "Guochuang"
	clientSet, _ := CreateClientSet()
	m := NewManager(clientSet)

	// 创建一个Action
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

	a, err := m.CreateAction(as, nil, namespace, u.String(), "")
	if err != nil {
		panic(err)
	}

	// 创建一个事件
	//eventBroadcaster := recorder.NewBroadcaster()
	//err = eventBroadcaster.StartRecordingToSink(context.Background(), &core.EventSinkImpl{Interface: m.GetEventClient(namespace).Client})
	//if err != nil {
	//	panic(err)
	//}
	//
	//scheme := scheme.NewScheme()
	//apis.AddToScheme(scheme)
	//recorder := eventBroadcaster.NewRecorder(scheme, "Task-Exporter")

	err = m.LogEvent(a, apis.EventTypeNormal, events.SelectOtherDomain, fmt.Sprintf("Create Event"), namespace)
	if err != nil {
		return
	}
	time.Sleep(1 * time.Second)

	// 开始查询
	fmt.Println("Get Event")
	e, err := m.GetEvents(a.Name, a.Namespace)
	if err != nil {
		panic(err)
	}
	fmt.Println(e)
}

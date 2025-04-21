package scheduler

import (
	"context"
	apis "hit.edu/framework/pkg/apis/cores"
	metav1 "hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/component-base/logs"
	"hit.edu/framework/pkg/scheduler/utils"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
)

func createConditionTask() apis.Task {

	group1Action1Runtime1 := apis.RuntimeSpec{
		Name: "ConditionGroup1Action1Runtime1",
	}

	group1Action1 := apis.ActionSpec{
		Name: "ConditionGroup1Action1",
		Runtimes: []apis.RuntimeSpec{
			group1Action1Runtime1,
		},
	}

	group1Spec := apis.GroupSpec{
		Name: "ConditionGroup1",
		Actions: []apis.ActionSpec{
			group1Action1,
		},
	}

	group2Action1Runtime1 := apis.RuntimeSpec{
		Name: "ConditionGroup2Action1Runtime1",
	}

	group2Action1 := apis.ActionSpec{
		Name: "ConditionGroup2Action1",
		Runtimes: []apis.RuntimeSpec{
			group2Action1Runtime1,
		},
	}

	group2Spec := apis.GroupSpec{
		Name: "ConditionGroup2",
		Actions: []apis.ActionSpec{
			group2Action1,
		},
	}

	task := apis.Task{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Task",
			APIVersion: "resources/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ConditionTask",
			Namespace: apis.NamespaceTest,
		},
		Spec: apis.TaskSpec{
			Name: "ConditionTask",
			Groups: []apis.GroupSpec{
				group1Spec,
				group2Spec,
			},
		},
		Status: apis.TaskStatus{
			Groups: map[string]apis.ObjectReference{
				"ConditionGroup1": {
					Kind: "Group",
					Name: "ConditionGroup1",
				},
				"ConditionGroup2": {
					Kind: "Group",
					Name: "ConditionGroup2",
				},
			},
		},
	}
	return task
}

func createOrangeTask() apis.Task {
	task := apis.Task{}
	return task
}

// 新的场景一的task
func createSceneOneTask() apis.Task {

	kuavoDetectRuntime := apis.RuntimeSpec{
		Name: "KuavoDetectRuntime",
	}

	kuavoDetectAction := apis.ActionSpec{
		Name: "KuavoDetectAction",
		Runtimes: []apis.RuntimeSpec{
			kuavoDetectRuntime,
		},
	}

	kuavoDetectGroup := apis.GroupSpec{
		Name: "KuavoDetectGroup",
		Actions: []apis.ActionSpec{
			kuavoDetectAction,
		},
	}

	kuavoGrabRuntime := apis.RuntimeSpec{
		Name: "KuavoGrabRuntime",
	}

	kuavoGrabAction := apis.ActionSpec{
		Name: "KuavoGrabAction",
		Runtimes: []apis.RuntimeSpec{
			kuavoGrabRuntime,
		},
	}

	kuavoGrabGroup := apis.GroupSpec{
		Name: "kuavoGrabGroup",
		Actions: []apis.ActionSpec{
			kuavoGrabAction,
		},
		Parents: []string{"KuavoDetectGroup"},
		Conditions: &apis.Conditions{
			Formulas: []apis.ConditionFormula{
				{},
			},
		},
	}

	galaxeaDetectRuntime := apis.RuntimeSpec{
		Name: "GalaxeaDetectRuntime",
	}

	galaxeaDetectAction := apis.ActionSpec{
		Name: "GalaxeaDetectAction",
		Runtimes: []apis.RuntimeSpec{
			galaxeaDetectRuntime,
		},
	}

	galaxeaDetectGroup := apis.GroupSpec{
		Name: "GalaxeaDetectGroup",
		Actions: []apis.ActionSpec{
			galaxeaDetectAction,
		},
		Parents: []string{"KuavoDetectGroup"},
	}

	galaxeaGrabRuntime := apis.RuntimeSpec{
		Name: "GalaxeaGrabRuntime",
	}

	galaxeaGrabAction := apis.ActionSpec{
		Name: "GalaxeaGrabAction",
		Runtimes: []apis.RuntimeSpec{
			galaxeaGrabRuntime,
		},
	}

	galaxeaGrabGroup := apis.GroupSpec{
		Name: "GalaxeaGrabGroup",
		Actions: []apis.ActionSpec{
			galaxeaGrabAction,
		},
		Parents: []string{"GalaxeaDetectGroup"},
		Conditions: &apis.Conditions{
			Formulas: []apis.ConditionFormula{
				{},
			},
		},
	}

	task := apis.Task{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Task",
			APIVersion: "resources/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "Scene1Task",
			Namespace: apis.NamespaceTest,
		},
		Spec: apis.TaskSpec{
			Name: "ConditionTask",
			Groups: []apis.GroupSpec{
				kuavoDetectGroup,
				kuavoGrabGroup,
				galaxeaDetectGroup,
				galaxeaGrabGroup,
			},
		},
		Status: apis.TaskStatus{
			Groups: map[string]apis.ObjectReference{
				"ConditionGroup1": {
					Kind: "Group",
					Name: "ConditionGroup1",
				},
				"ConditionGroup2": {
					Kind: "Group",
					Name: "ConditionGroup2",
				},
			},
		},
	}
	return task
}

// go test -run TestAddGroup -v
func TestAddGroup(t *testing.T) {
	group := apis.Group{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Group",
			APIVersion: "resources/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "testGroup1",
			Namespace: apis.NamespaceTest,
		},
		Spec: apis.GroupSpec{
			Name: "testGroup1",
		},
		Status: apis.GroupStatus{
			Phase: apis.Unknown,
		},
	}

	cs, err := utils.CreateClientSet()
	if err != nil {
		t.Fatalf("%v", err)
	}
	ctx := context.Background()
	gc := cs.Core().Groups(apis.NamespaceTest)
	_, err = gc.Create(ctx, &group, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("%v", err)
		return
	}
}

// go test -run TestSendToProxy -v
func TestSendToProxy(t *testing.T) {
	logs.Init("testModule")
	orange, err := os.ReadFile("orange.json")
	client := &http.Client{}

	url := "http://192.168.8.176:8899/framework/v1/task?Name=T1&&Namesapce=test"
	logs.Info(url)
	req, err := http.NewRequest("POST", url, strings.NewReader(string(orange)))
	if err != nil {
		logs.Fatal(err)
	}
	//Content-Type很重要，下文解释
	//req.Header.Set("Content-Type", "application/x-www")
	req.Header.Set("Content-Type", "application/json")
	//req.Header.Set("Content-Type", "multipart/form-data")

	rep, err := client.Do(req)
	if err != nil {
		logs.Fatal(err.Error())
	}
	data, err := io.ReadAll(rep.Body)
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			logs.Error(err)
		}
	}(rep.Body)
	if err != nil {
		logs.Fatal(err)
	}
	logs.Infof("resp is : %s", string(data))
}

// go test -run TestAddDevice -v
func TestAddDevice(t *testing.T) {
	logs.Init("testModule")
	ctx := context.Background()
	cs, err := createClientSet()
	if err != nil {
		logs.Error(err)
		return
	}
	dc := cs.Core().Nodes(apis.NamespaceTest)
	node := apis.Node{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Node",
			APIVersion: "resources/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "TestEnvNode1",
			Namespace: apis.NamespaceTest,
		},
		Spec: apis.NodeSpec{
			NodeName: "TestEnvNode1",
			HostName: "192.168.8.176",
		},
		Status: apis.NodeStatus{},
	}
	_, err = dc.Create(ctx, &node, metav1.CreateOptions{})
	if err != nil {
		logs.Error(err.Error())
		return
	}

}

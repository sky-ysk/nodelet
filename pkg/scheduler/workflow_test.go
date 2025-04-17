package scheduler

import (
	"context"
	apis "hit.edu/framework/pkg/apis/cores"
	metav1 "hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/scheduler/utils"
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
			Namespace: apis.NamespaceAll,
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
	task := apis.Task{}
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
			Namespace: apis.NamespaceAll,
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
	gc := cs.Core().Groups(apis.NamespaceAll)
	_, err = gc.Create(ctx, &group, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("%v", err)
		return
	}
}

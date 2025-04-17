package scheduler

import (
	apis "hit.edu/framework/pkg/apis/cores"
	metav1 "hit.edu/framework/pkg/apis/meta"
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

package plugins

import (
	"context"
	"fmt"
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/apis/meta"
	"testing"
)

func TestBuildGroupsRequest(t *testing.T) {
	task := mockGetTask()
	ctx := context.Background()
	req := BuildSendGroupsRequest(ctx, &task)
	fmt.Println(req)
	fmt.Println(len(req.TopInfo))
}

func TestSendGroupsRequest(t *testing.T) {
	task := mockGetTask()
	ctx := context.Background()
	req := BuildSendGroupsRequest(ctx, &task)
	plugin := &ScorePluginDBY{
		pluginClient: NewScorePluginClient(),
	}
	//NewScorePluginClient()
	plugin.SendGroups(ctx, &task)
	fmt.Println(req)
	fmt.Println(len(req.TopInfo))
	//TODO Fill the node name
	//plugin.Score(ctx, &task.Spec.Groups[0], "")
}

func mockGetTask() apis.Task {
	reqs := make([]apis.ResourceRequirement, 0)
	req1 := apis.ResourceRequirement{
		Name:       "CPU",
		Lowbound:   "1",
		Upperbound: "1",
	}
	req2 := apis.ResourceRequirement{
		Name:       "RAM",
		Lowbound:   "1",
		Upperbound: "1",
	}
	req3 := apis.ResourceRequirement{
		Name:       "IO",
		Lowbound:   "1",
		Upperbound: "1",
	}
	reqs = append(reqs, req1, req2, req3)
	g1 := apis.Group{
		ObjectMeta: meta.ObjectMeta{
			Name: "testGroup1",
		},
		Spec: apis.GroupSpec{
			ResourceRequirements: reqs,
			//Parents:              [] string {1,2,3 },
		},
		Status: apis.GroupStatus{
			GroupID: "testGroup1",
			Belongs: apis.IDRef{
				TaskID: "task1",
			},
		},
	}
	g2 := apis.Group{
		ObjectMeta: meta.ObjectMeta{
			Name: "testGroup2",
		},
		Spec: apis.GroupSpec{
			ResourceRequirements: reqs,
			Parents:              []string{"testGroup1"},
		},
		Status: apis.GroupStatus{
			GroupID: "testGroup2",
			Belongs: apis.IDRef{
				TaskID: "task1",
			},
		},
	}
	g3 := apis.Group{
		ObjectMeta: meta.ObjectMeta{
			Name: "testGroup3",
		},
		Spec: apis.GroupSpec{
			ResourceRequirements: reqs,
			Parents:              []string{"testGroup2"},
		},
		Status: apis.GroupStatus{
			GroupID: "testGroup3",
			Belongs: apis.IDRef{
				TaskID: "task1",
			},
		},
	}

	return apis.Task{
		Spec: apis.TaskSpec{
			Groups: []apis.Group{g1, g2, g3},
		},
		Status: apis.TaskStatus{
			TaskID: "task1",
		},
	}
}

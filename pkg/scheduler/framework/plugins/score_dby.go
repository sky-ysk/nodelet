package plugins

import (
	"context"
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/scheduler/framework"
	"hit.edu/framework/pkg/scheduler/transport"
	"hit.edu/framework/pkg/scheduler/transport/client"
	"net/http"
)

type ScorePluginDBY struct {
	pluginClient ScorePluginClient
}

type ScorePluginClient struct {
	client *http.Client
}

func (sp *ScorePluginDBY) Name() string {
	return "ScorePluginForDuBoyu"
}

func (sp *ScorePluginDBY) Score(ctx context.Context, group *apis.Group, nodeName string) (int64, *framework.Status) {
	//TODO 没测过
	request := transport.ScoreRequest{
		Group:  group,
		NodeID: nodeName,
	}
	resp := sp.SendScoreRequest(request)
	if resp.BaseResp.Success {
		return resp.Data.Score, framework.NewStatus(framework.Success, "default success")
	}
	return -1, framework.NewStatus(framework.Error, resp.BaseResp.Reason)
}

func (sp *ScorePluginDBY) SendGroups(ctx context.Context, task *apis.Task) (int64, *framework.Status) {
	//TODO 没测过

	request := buildSendGroupsRequest(ctx, task)
	resp := sp.SendScoreRequest(request)
	if resp.BaseResp.Success {
		return resp.Data.Score, framework.NewStatus(framework.Success, "default success")
	}
	return -1, framework.NewStatus(framework.Error, resp.BaseResp.Reason)
}

func NewScorePluginDBY(ctx context.Context, f framework.Handle) (framework.Plugin, error) {
	return &ScorePluginDBY{
		SchedulerClient: client.GetHClient(),
	}, nil
}

type GroupTopInfo struct {
	Parent string `json:"parent"`
	Child  string `json:"child"`
}

type SendGroupsRequest struct {
	Groups              []string                              `json:"groups"`
	TaskId              string                                `json:"taskId"`
	ResourceRequirement map[string][]apis.ResourceRequirement `json:"resourceRequirement"`
	TopInfo             []GroupTopInfo                        `json:"topInfo"`
}

func buildSendGroupsRequest(ctx context.Context, task *apis.Task) *SendGroupsRequest {
	topInfo := make([]GroupTopInfo, 0)
	groupsID := make([]string, 0)
	resourcesMap := make(map[string][]apis.ResourceRequirement)
	taskID := task.UID
	//TODO 可能需要做深复制 @lbh
	for _, group := range task.Spec.Groups {
		topInfo = append(topInfo, GroupTopInfo{})
		groupsID = append(groupsID, string(group.ObjectMeta.UID))
		resources := make([]apis.ResourceRequirement, 0)
		for _, requirement := range group.Spec.ResourceRequirements {
			resources = append(resources, requirement)
		}
		if len(resources) > 0 {
			resourcesMap[string(group.ObjectMeta.UID)] = resources
		}
		for _, parent := range group.Spec.Parents {
			topInfo = append(topInfo, GroupTopInfo{
				Child: string(group.ObjectMeta.UID),
				//TODO 这个里面放的估计不是uid。。
				Parent: parent,
			})
		}
	}
	return &SendGroupsRequest{
		Groups:              groupsID,
		TaskId:              string(taskID),
		ResourceRequirement: resourcesMap,
		TopInfo:             topInfo,
	}
}

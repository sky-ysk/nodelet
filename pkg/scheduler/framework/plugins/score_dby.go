package plugins

import (
	"context"
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/scheduler/framework"
	"hit.edu/framework/pkg/scheduler/transport"
	"hit.edu/framework/pkg/scheduler/transport/client"
)

type ScorePluginDBY struct {
	SchedClient transport.SchedulerClient
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
	resp := sp.SchedClient.SendScoreRequest(request)
	if resp.BaseResp.Success {
		return resp.Data.Score, framework.NewStatus(framework.Success, "default success")
	}
	return -1, framework.NewStatus(framework.Error, resp.BaseResp.Reason)
}

func NewScorePluginDBY(ctx context.Context, f framework.Handle) (framework.Plugin, error) {
	return &ScorePluginDBY{
		SchedClient: client.GetHClient(),
	}, nil
}

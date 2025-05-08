package plugins

import (
	"context"
	apis "hit.edu/framework/pkg/apis/cores"
	metav1 "hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/client-go/clients"
	"hit.edu/framework/pkg/client-go/clients/typed/core"
	"hit.edu/framework/pkg/component-base/logs"
	"hit.edu/framework/pkg/scheduler/framework"
	"hit.edu/framework/pkg/scheduler/utils"
	"strconv"
)

//该插件主要适配@Duboyu的需求，作为DTS插件的benchmark对比
//根据@Hezhangyi设计的测试工作流特点（吃CPU不吃内存），该插件为CPU占用低的Node打高分（分数1-10）

type GreedyScorePlugin struct {
	clientSet  *clients.ClientSet
	nodeClient core.NodeInterface
}

func NewGreedyScorePlugin(ctx context.Context, f framework.Handle) (framework.Plugin, error) {
	cs, err := utils.CreateClientSetWithTimeOut(3600)
	if err != nil {
		return nil, err
	}
	nc := cs.Core().Nodes(apis.NamespaceTest)
	return &GreedyScorePlugin{
		clientSet:  cs,
		nodeClient: nc,
	}, nil
}

func (sp *GreedyScorePlugin) Name() string {
	return "GreedyScore"
}

func (sp *GreedyScorePlugin) Score(ctx context.Context, group *apis.Group, nodeName string) (int64, *framework.Status) {
	return sp.greedyScore(ctx, group, nodeName)
}

func (sp *GreedyScorePlugin) greedyScore(ctx context.Context, group *apis.Group, nodeName string) (int64, *framework.Status) {
	node, err := sp.nodeClient.Get(ctx, nodeName, metav1.GetOptions{})
	if err != nil {
		logs.Error(err.Error())
		return 0, framework.NewStatus(framework.Error, err.Error())
	}
	if _, ok := node.Status.Usage["cpu"]; !ok {
		return 0, framework.NewStatus(framework.Error, "CPU usage info is nil")
	}
	if _, ok := node.Status.Usage["memory"]; !ok {
		return 0, framework.NewStatus(framework.Error, "RAM usage info is nil")
	}
	if len(node.Status.Usage["memory"]) == 0 || len(node.Status.Usage["cpu"]) == 0 {
		return 0, framework.NewStatus(framework.Error, " usage len  is 0")
	}
	//内存使用率数据解析出来了，暂时放这里没用到，后续可以修改
	cpuValue := node.Status.Usage["cpu"][0]
	ramValue := node.Status.Usage["memory"][0]
	if _, ok := cpuValue.Values["AveUtil"]; !ok {
		return 0, framework.NewStatus(framework.Error, "CPU value is nil")
	}
	if _, ok := ramValue.Values["Usage"]; !ok {
		return 0, framework.NewStatus(framework.Error, "ram value is nil")
	}
	cpuAvgUse, err := strconv.ParseFloat(cpuValue.Values["AveUtil"], 64)
	if err != nil {
		return 0, framework.NewStatus(framework.Error, err.Error())
	}
	ramUse, err := strconv.ParseFloat(ramValue.Values["Usage"], 64)
	if err != nil {
		return 0, framework.NewStatus(framework.Error, err.Error())
	}
	score := int64((100 - cpuAvgUse) / 10)
	if score <= 0 {
		//logs.Warnf("invalid score %d, group %s, node %s", score, group.Name, nodeName)
		score = 1
	}
	if score > 100 {
		//logs.Warnf("score exceed %d, group %s, node %s", score, group.Name, nodeName)
		score = 10
	}
	logs.Infof("node %s CPU use %f, mem use %f, score %d, group %s", nodeName, cpuAvgUse, ramUse, score, group.Name)
	return score, framework.NewStatus(framework.Success, "")
}

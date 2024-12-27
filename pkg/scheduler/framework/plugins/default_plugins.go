package plugins

import (
	"context"
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/scheduler/apis/config"
	"hit.edu/framework/pkg/scheduler/framework"
	"math/rand"
	"time"
)

type DefaultFilter struct {
}

func (fl *DefaultFilter) Name() string {
	return "DefaultFilter"
}

// 插件接口 字段
// TODO @linbohai GO 单元测试
// 过滤插件 返回Group是否能在对应的设备节点上运行的信息
// 可以运行返回success
func (fl *DefaultFilter) Filter(ctx context.Context, group *apis.Group, node *config.NodeInfo) *framework.Status {
	status := framework.NewStatus(framework.Success, "default success")
	return status
}

type DefaultScorePlugin struct {
	r *rand.Rand
}

func (sp *DefaultScorePlugin) Name() string {
	return "DefaultScorePlugin"
}

// 默认打分 随机生成一个0-10的整数
func (sp *DefaultScorePlugin) Score(ctx context.Context, group *apis.Group, nodeName string) (int64, *framework.Status) {
	status := framework.NewStatus(framework.Success, "default success")
	return int64(sp.r.Intn(11)), status
}

func NewDefaultFilterPlugin(ctx context.Context, f framework.Handle) (framework.Plugin, error) {
	return &DefaultFilter{}, nil
}

func NewDefaultScorePlugin(ctx context.Context, f framework.Handle) (framework.Plugin, error) {
	return &DefaultScorePlugin{
		r: rand.New(rand.NewSource(time.Now().UnixNano())),
	}, nil
}

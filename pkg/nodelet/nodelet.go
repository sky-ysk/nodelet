package nodelet

import (
	"context"
	"hit.edu/framework/pkg/nodelet/node/collector"
	"hit.edu/framework/pkg/nodelet/task"
)

// Nodelet,部署在每个节点上，管理当前节点上的所有资源
// Node中存在以下类型的exporter
//
//	NodeExporter,监控当前节点的资源情况
//	TaskExporter，部署任务，监控当前任务的执行情况和资源使用情况，分配资源
//	AbilityExporter, 拉起能力
//
// Nodelet管理当前节点上的所有Images,Image类型包括Docker镜像，
// Nodelet管理当前节点上的所有Docker，包括Task Exporter直接拉起和通过Pod间接部署的Docker

type Nodelet struct {
	cfg   *Config //全局config，包含下级的exporter config
	cache map[string]collector.Metric
	// Close this to shut down the resourcelet.
	StopEverything <-chan struct{}
}

func New(ctx context.Context) (*Nodelet, error) {
	cfg := NewConfig()
	stopEverything := ctx.Done()

	nl := &Nodelet{
		cfg:            cfg,
		StopEverything: stopEverything,
	}

	// TODO: Nodelet注册, 注册自己的Node资源
	return nl, nil
}

func (nl *Nodelet) Run(ctx context.Context) {
	// 构造Node Exporter
	//ne, err := node.NewNodeExporter(nl.cfg.nc)
	//if err != nil {
	//	panic(err)
	//}
	//go ne.Run(ctx)

	//构造Task Exporter
	te, err := task.NewTaskExporter(nl.cfg.tc)
	if err != nil {
		panic(err)
	}
	go te.Run(ctx)

	// 构造Ability Exporter

	// TODO：配置不同的Channel

	// TODO: 运行不同的模块,多进程？
	for {

	}
}

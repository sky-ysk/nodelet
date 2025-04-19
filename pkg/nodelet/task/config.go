package task

import (
	cross_core "hit.edu/framework/test/etcd_sync/active/clients/typed/core"
)

type Config struct {
	// Node Name
	NodeName         string
	groupTargetMap   map[string]cross_core.GroupInterface
	actionTargetMap  map[string]cross_core.ActionInterface
	runtimeTargetMap map[string]cross_core.RuntimeInterface
}

func NewConfig(name string, grouptargetMap map[string]cross_core.GroupInterface, actiontargetMap map[string]cross_core.ActionInterface, runtimetargetMap map[string]cross_core.RuntimeInterface) *Config {
	return &Config{
		NodeName:         name,
		groupTargetMap:   grouptargetMap,
		actionTargetMap:  actiontargetMap,
		runtimeTargetMap: runtimetargetMap,
	}
}

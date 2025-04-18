package task

import (
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/test/etcd_sync/informer"
)

type Config struct {
	// Node Name
	NodeName         string
	groupTargetMap   map[string]*informer.Target[*apis.Group]
	actionTargetMap  map[string]*informer.Target[*apis.Action]
	runtimeTargetMap map[string]*informer.Target[*apis.Runtime]
}

func NewConfig(name string, grouptargetMap map[string]*informer.Target[*apis.Group], actiontargetMap map[string]*informer.Target[*apis.Action], runtimetargetMap map[string]*informer.Target[*apis.Runtime]) *Config {
	return &Config{
		NodeName:         name,
		groupTargetMap:   grouptargetMap,
		actionTargetMap:  actiontargetMap,
		runtimeTargetMap: runtimetargetMap,
	}
}

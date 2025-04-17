package task

import (
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/test/etcd_sync/informer"
)

type Config struct {
	// Node Name
	NodeName  string
	TargetMap map[string]*informer.Target[*apis.Group]
}

func NewConfig(name string, targetMap map[string]*informer.Target[*apis.Group]) *Config {
	return &Config{
		NodeName:  name,
		TargetMap: targetMap,
	}
}

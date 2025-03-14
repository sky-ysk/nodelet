package nodelet

import (
	"hit.edu/framework/pkg/nodelet/node"
	"hit.edu/framework/pkg/nodelet/task"
)

type Config struct {
	// TODO：整合子模块
	nc *node.Config
	tc *task.Config
}

func NewConfig() *Config {
	nodeName := "CloudNode1" // 如果是边集群，这里修改为  "EdgeNode1"
	return &Config{
		//需要修改成从配置文件中读取内容 例如：config.json
		nc: node.NewConfig([]string{"CPU", "Memory", "Storage"}, "", nodeName, "Cloud"),
		tc: task.NewConfig(nodeName),
	}
}

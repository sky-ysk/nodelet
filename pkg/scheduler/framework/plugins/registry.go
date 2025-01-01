package plugins

import "hit.edu/framework/pkg/scheduler/framework/runtime"

// TODO: 实现模型加载器，加载模型类算法插件
// TODO: 实现服务访问器，访问远程服务类算法插件
// TODO: 实现任务生成类插件，访问大模型
// TODO: 实现DSL解析类插件, 解析DSL
func NewInTreeRegistry() runtime.Registry {
	registry := runtime.Registry{
		//将实现的插件放入这个位置
		"DefaultFilter":      NewDefaultFilterPlugin,
		"DefaultScorePlugin": NewDefaultScorePlugin,
		//"ScorePluginForDBY":  NewScorePluginDBY,
	}
	return registry
}

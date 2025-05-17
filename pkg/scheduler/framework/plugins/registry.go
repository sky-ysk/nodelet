package plugins

import "hit.edu/framework/pkg/scheduler/framework/runtime"

// TODO: 实现模型加载器，加载模型类算法插件
// TODO: 实现服务访问器，访问远程服务类算法插件
// TODO: 实现任务生成类插件，访问大模型
// TODO: 实现DSL解析类插件, 解析DSL
func NewInTreeRegistry() runtime.Registry {
	registry := runtime.Registry{
		//将实现的插件放入这个位置
		"DefaultFilter": NewDefaultFilterPlugin,

		//下面这仨是打分插件，测试性能的时候，同一时间只开一个，用不到的注释掉
		//贪心调度策略（CPU占用率低优先调度）
		"GreedyScore": NewGreedyScorePlugin,
		//随机调度
		"DefaultScorePlugin": NewDefaultScorePlugin,
		//DTS插件策略, 注意！使用这个策略必须要配合代码 go sched.monitorTask(ctx)使用（全局搜索把这个地方打开注释）
		//"ScorePluginForDBY": NewScorePluginDBY,

		"DefaultBindPlugin": NewDefaultBindPlugin,

		"DeviceBinder": NewDeviceBinder,
	}
	return registry
}

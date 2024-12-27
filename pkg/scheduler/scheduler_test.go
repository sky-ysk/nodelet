package scheduler

import (
	"hit.edu/framework/pkg/scheduler/backend/queue"
	"testing"
)

//测试调度框架

func TestScheduler_Run(t *testing.T) {

	//stopEverything := ctx.Done()

	// 配置调度器启动选项，在这里需要定义所需的模块，插件
	//options := defaultSchedulerOptions
	//for _, opt := range opts {
	//	opt(&options)
	//}

	// 配置调度器参数

	// 配置插件模块
	//registry := plugins.NewInTreeRegistry()

	// 配置任务队列git

	// 配置资源监控模块
	//scheduleChan := make(chan internal.ScheduleSignal)
	queue := queue.NewPriorityQueue()
	sched := &Scheduler{
		//StopEverything:  stopEverything,
		//ScheduleSigChan: scheduleChan,
		SchedulingQueue: queue,
	}

	//schedQueue := &queue.PriorityQueue{}

	sched.applyDefaultHandlers()
	sched.ReadyGroup = sched.SchedulingQueue.Pop

	//return sched, nil
}

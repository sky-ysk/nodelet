package resource

// 动态资源分配组件
// 监控所有运行过程中的Group
// 检查Group使用的各类资源
// 对于计算、网络、存储资源，根据资源的配额和Group的QoS,根据Group的执行进度，动态分配资源的使用情况
// 需要区分资源充足和资源不足的情况
// TODO: 需要整合QoS动态资源分配算法，结合预测的模块，更改Group对应的资源配额，由Task Exporter完成资源的实际分配

// Group的获取方式，从API-Server中获取所有的Group
// 我们自己的API-Server和Client-Go, 使用方法类似Informer和Lister
// TODO: 后续可能会合并到调度器中

// 检测所有Group的资源分配
// 检测是否有因为Group执行异常而未被释放的资源
// 如果有，尝试释放这些资源

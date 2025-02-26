package controller

//// 添加 Node Informer 到 Controller
//type Controller struct {
//	// 原有字段...
//	nodeLister  corelisters.NodeLister
//	nodesSynced cache.InformerSynced
//}
//
//// 在 NewController 中初始化 Node Informer
//nodeInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
//UpdateFunc: func(oldObj, newObj interface{}) {
//newNode := newObj.(*corev1.Node)
//if isNodeOverThreshold(newNode) {
//// 触发关联 Group 的重新同步
//groups := getGroupsOnNode(newNode.Name)
//for _, g := range groups {
//c.enqueueGroup(g)
//}
//}
//},
//})
//
//// 阈值检查函数
//func isNodeOverThreshold(node *corev1.Node) bool {
//	cpuUsage := getNodeMetric(node, "cpu")
//	memUsage := getNodeMetric(node, "memory")
//	return cpuUsage > thresholdCPU || memUsage > thresholdMemory
//}

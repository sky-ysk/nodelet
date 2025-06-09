package scheduler

import (
	"container/heap"
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"hit.edu/framework/pkg/apimachinery/runtime"
	"hit.edu/framework/pkg/apimachinery/runtime/schema"
	"hit.edu/framework/pkg/apimachinery/runtime/serializer"
	apis "hit.edu/framework/pkg/apis/cores"
	metav1 "hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/client-go/clients"
	"hit.edu/framework/pkg/client-go/rest"
	"hit.edu/framework/pkg/component-base/logs"
	"hit.edu/framework/pkg/scheduler/apis/config"

	"hit.edu/framework/pkg/scheduler/framework"
	"hit.edu/framework/pkg/scheduler/workflow"
)

//TODO: 调度相关参数

var clearNominatedNode = &framework.NominatingInfo{NominatingMode: framework.ModeOverride, NominatedNodeName: ""}

const (
	// Percentage of plugin metrics to be sampled.
	pluginMetricsSamplePercent = 10
	// minFeasibleNodesToFind is the minimum number of nodes that would be scored
	// in each scheduling cycle. This is a semi-arbitrary value to ensure that a
	// certain minimum of nodes are checked for feasibility. This in turn helps
	// ensure a minimum level of spreading.
	minFeasibleNodesToFind = 100
	// minFeasibleNodesPercentageToFind is the minimum percentage of nodes that
	// would be scored in each scheduling cycle. This is a semi-arbitrary value
	// to ensure that a certain minimum of nodes are checked for feasibility.
	// This in turn helps ensure a minimum level of spreading.
	minFeasibleNodesPercentageToFind = 5
	// numberOfHighestScoredNodesToReport is the number of node scores
	// to be included in ScheduleResult.
	numberOfHighestScoredNodesToReport = 3
)

func (sched *Scheduler) ScheduleOne(ctx context.Context) {
	//TODO @linbohai 从调度队列中获取待调度的Group
	logs.Trace("now schedule one running")
	groupInfo, err := sched.ReadyGroup(ctx)
	//msg := fmt.Sprintf("ready group info: %v", groupInfo.Group.ObjectMeta.Name)
	//logs.Info(msg)
	//groupInfo := groupInfos[0]
	if err != nil {
		logs.Error(err.Error())
		//sched.SchedulingQueue.Done(groupInfo.Group.GroupId)
		return
	}
	// // TODO: 检查是否跳过Group
	//if sched.skipPodSchedule(ctx, fwk, pod) {
	//	// We don't put this Pod back to the queue, but we have to cleanup the in-flight pods/events.
	//	sched.SchedulingQueue.Done(pod.UID)
	//	return
	//}

	// // 开始调度过程
	start := time.Now()
	group := groupInfo.Group
	fwk, err := sched.frameworkForGroup(group)

	//schedulingCycleCtx, cancel := context.WithCancel(ctx)
	//defer cancel()

	//scheduleResults, assumedPodInfo, status := sched.schedulingCycle(ctx, start, nil, groupInfos)
	//TODO: 解决调度失败的任务

	//初始化的时候就把Group里面需要跳过的插件写进去
	state := framework.NewCycleState(group)

	scheduleResult, _, _ := sched.schedulingCycle(ctx, fwk, start, state, *groupInfo)
	//logs.Info("sched result is")
	//jsonData, err := json.Marshal(scheduleResult)
	//if err != nil {
	//	logs.Error(err.Error())
	//}
	//logs.Info(string(jsonData))
	msg := fmt.Sprintf("schedule result : group %s on node %s", scheduleResult.Group.ObjectMeta.Name, scheduleResult.SuggestedHost)
	logs.Info(msg)

	//TODO: 部署任务/Bind相关接口
	go func() {
		bindingCycleCtx, cancel := context.WithCancel(ctx)
		defer cancel()

		//metrics.Goroutines.WithLabelValues(metrics.Binding).Inc()
		//defer metrics.Goroutines.WithLabelValues(metrics.Binding).Dec()

		status := sched.bindingCycle(bindingCycleCtx, fwk, state, scheduleResult, start)
		//TODO 错误处理
		if !status.IsSuccess() {
			//sched.handleBindingCycleError(bindingCycleCtx, state, fwk, assumedPodInfo, start, scheduleResult, status)
			return
		}
	}()

}

func (sched *Scheduler) frameworkForGroup(group *apis.Group) (framework.Framework, error) {
	if group.Spec.SchedulerName == nil {
		return sched.DefaultFramework, nil
	}
	fwk, ok := sched.Profiles[*group.Spec.SchedulerName]
	if !ok {
		return sched.DefaultFramework, nil
		//return nil, fmt.Errorf("profile not found for scheduler name %q", group.Spec.SchedulerName)
	}
	return fwk, nil
}

//ctx context.Context,
//state *framework.CycleState,
//fwk framework.Framework,
//podInfo *framework.QueuedPodInfo,
//start time.Time,
//podsToActivate *framework.PodsToActivate,

// schedulingCycle tries to schedule a single Pod.
func (sched *Scheduler) schedulingCycle(
	ctx context.Context,
	fwk framework.Framework,
	start time.Time,
	state *framework.CycleState,
	//TODO: 等待调度的Pod
	groupInfo config.QueuedGroupInfo,
	// TODO: 需要跳过的插件
) (ScheduleResult, *config.QueuedGroupInfo, *framework.Status) {
	//logger := klog.FromContext(ctx)

	group := groupInfo.Group
	scheduleResult, err := sched.ScheduleGroup(ctx, fwk, state, group)
	//TODO 错误处理，目前先fail fast
	if err != nil {
		logs.Error(err.Error())
		//TODO 算法时间监控模块，后面再做

		// 没有可调度的节点，后续放入Unschedulable队列
		//TODO 这部分队列转移？
		if errors.Is(err, ErrNoNodesAvailable) {
			status := framework.NewStatus(framework.UnschedulableAndUnresolvable).WithError(err)
			return scheduleResult, nil, status
		}
		//TODO 执行后续插件，细化错误信息，后续再做
		return scheduleResult, nil, framework.NewStatus(framework.Unschedulable).WithError(err)
	}
	//TODO 记录时间指标
	//TODO 做一些字段级别的资源预留？？ 目前是浅复制
	//TODO 填写results
	if sts := fwk.RunReservePluginsReserve(ctx, state, scheduleResult.Group, scheduleResult.SuggestedHost); !sts.IsSuccess() {
		// trigger un-reserve to clean up state associated with the reserved Pod
		fwk.RunReservePluginsUnreserve(ctx, state, scheduleResult.Group, scheduleResult.SuggestedHost)
		if sts.IsRejected() {
		}
	}

	//TODO 后续结合下层框架看看

	//TODO Run "permit" plugins.

	return scheduleResult, &groupInfo, nil
}

// TODO @linbohai 部署 写到资源视图
// 将任务与节点绑定，后续开始部署流程
func (sched *Scheduler) bindingCycle(
	ctx context.Context,
	fwk framework.Framework,
	state *framework.CycleState,
	scheduleResult ScheduleResult,
	start time.Time,
	// TODO: 待部署的Group
) *framework.Status {
	// start := time.Now()
	//TODO @linbohai 资源预留

	status := fwk.RunBindPlugins(ctx, state, scheduleResult.Group, scheduleResult.SuggestedHost)
	// {
	// 	return status
	// }
	return status

	//TODO: 将Group标记为调度完成
	//TODO: 将Group迁移到待部署队列
}

// assume signals to the cache that a pod is already in the cache, so that binding can be asynchronous.
// assume modifies `assumed`.
func (sched *Scheduler) assume(assumed *workflow.Group, host string) error {
	// Optimistically assume that the binding will succeed and send it to apiserver
	// in the background.
	// If the binding fails, scheduler will release resources allocated to assumed pod
	// immediately.
	assumed.Spec.NodeName = host

	//TODO 实现一下对应的功能，调用事件，发送绑定消息
	//if err := sched.Cache.AssumePod(logger, assumed); err != nil {
	//	logger.Error(err, "Scheduler cache AssumePod failed")
	//	return err
	//}
	// if "assumed" is a nominated pod, we should remove it from internal cache
	if sched.SchedulingQueue != nil {
		//sched.SchedulingQueue.DeleteNominatedGroupIfExists(assumed)
	}

	return nil
}

// schedulePod tries to schedule the given pod to one of the nodes in the node list.
// If it succeeds, it will return the name of the node.
// If it fails, it will return a FitError with reasons.
func (sched *Scheduler) scheduleGroup(ctx context.Context,
	fwk framework.Framework,
	state *framework.CycleState,
	group *apis.Group) (result ScheduleResult, err error) {
	//TODO 这个调用链追踪后面再说
	//trace := utiltrace.New("Scheduling", utiltrace.Field{Key: "namespace", Value: pod.Namespace}, utiltrace.Field{Key: "name", Value: pod.Name})
	//defer trace.LogIfLong(100 * time.Millisecond)
	//TODO 刷新缓存，可以先写成同步调用
	//if err := sched.Cache.UpdateSnapshot(klog.FromContext(ctx), sched.nodeInfoSnapshot); err != nil {
	//	return result, err
	//}

	//trace.Step("Snapshotting scheduler cache and node infos done")

	//TODO 拿到nodeinfo
	//if sched.nodeInfoSnapshot.NumNodes() == 0 {
	//	return result, ErrNoNodesAvailable

	feasibleNodes, err := sched.findNodesThatFitGroup(ctx, fwk, state, group)
	if err != nil {
		return result, err
	}
	//trace.Step("Computing predicates done")

	//TODO 无节点 返回
	if len(feasibleNodes) == 0 {
		//return result, &framework.FitError{
		//	Pod:         pod,
		//	NumAllNodes: sched.nodeInfoSnapshot.NumNodes(),
		//	Diagnosis:   diagnosis,
		//}
		return result, errors.New("no nodes found")
	}

	//TODO  When only one node after predicate, just use it.
	if len(feasibleNodes) == 1 {
		return ScheduleResult{
			SuggestedHost: feasibleNodes[0].Node().Name,
			//EvaluatedNodes: 1 + diagnosis.NodeToStatus.Len(),
			FeasibleNodes: 1,
			Group:         group,
		}, nil
	}
	// 排序
	priorityList, err := prioritizeNodes(ctx, fwk, state, group, feasibleNodes)
	if err != nil {
		return result, err
	}
	//TODO out-tree input nodes + group
	// 筛选
	//host, _, err := selectHost(priorityList, numberOfHighestScoredNodesToReport)
	host, err := selectHostByProbability(priorityList)
	logs.Infof("host select by probability is %s, group %s", host, group.ObjectMeta.Name)
	if err != nil {
		logs.Error(err.Error())
	}

	//前端演示页面特判逻辑
	//if group.Spec.Desc != nil && len(group.Spec.Desc.Label) != 0 {
	//	if strings.Contains(group.Spec.Desc.Label[0], "Infer") {
	//		host = "EdgeNode1"
	//	} else {
	//		host = "CloudNode1"
	//	}
	//}
	//}
	//if strings.Contains(group.ObjectMeta.Name, "Reason") {
	//	host = "EdgeNode1"
	//}
	//if strings.Contains(group.ObjectMeta.Name, "Robot") {
	//	host = "EdgeNode1"
	//}
	return ScheduleResult{
		SuggestedHost: host,
		Group:         group,
		//TODO 后面需要了再做
		//EvaluatedNodes: len(feasibleNodes) + diagnosis.NodeToStatus.Len(),
		FeasibleNodes: len(feasibleNodes),
	}, err
}

// Filters the nodes to find the ones that fit the pod based on the framework
// filter plugins and filter extenders.
func (sched *Scheduler) findNodesThatFitGroup(ctx context.Context, fwk framework.Framework, state *framework.CycleState,
	group *apis.Group) ([]*config.NodeInfo, error) {

	//diagnosis := framework.Diagnosis{
	//	NodeToStatus: framework.NewDefaultNodeToStatus(),
	//}
	//TODO 拿到所有节点，走下层同步调用接口 没有可用节点最好抛个错误出来
	//目前走的一个mock方法
	allNodes, err := sched.getAllNodes()
	if err != nil {
		return nil, err
	}
	//logs.Infof("found %d nodes", len(allNodes))
	//TODO Run "prefilter" plugins. 这个先不做
	//preRes, s, unscheduledPlugins := fwk.RunPreFilterPlugins(ctx, state, pod)

	//preRes, s, unscheduledPlugins := fwk.RunPreSchedulePlugins(ctx, group)

	//diagnosis.UnschedulablePlugins = unscheduledPlugins
	//if !s.IsSuccess() {
	//	if !s.IsRejected() {
	//		return nil, diagnosis, s.AsError()
	//	}
	//	// All nodes in NodeToStatus will have the same status so that they can be handled in the preemption.
	//	diagnosis.NodeToStatus.SetAbsentNodesStatus(s)
	//
	//	// Record the messages from PreFilter in Diagnosis.PreFilterMsg.
	//	msg := s.Message()
	//	diagnosis.PreFilterMsg = msg
	//	logger.V(5).Info("Status after running PreFilter plugins for pod", "pod", klog.KObj(pod), "status", msg)
	//	diagnosis.AddPluginStatus(s)
	//	return nil, diagnosis, nil
	//}

	// "NominatedNodeName" can potentially be set in a previous scheduling cycle as a result of preemption.
	// This node is likely the only candidate that will fit the pod, and hence we try it first before iterating over all nodes.
	//TODO nominate 后面再说
	//if len(group.Status.NominatedNodeName) > 0 {
	//	feasibleNodes, err := sched.evaluateNominatedNode(ctx, pod, fwk, state, diagnosis)
	//	if err != nil {
	//		logger.Error(err, "Evaluation failed on nominated node", "pod", klog.KObj(pod), "node", pod.Status.NominatedNodeName)
	//	}
	//	// Nominated node passes all the filters, scheduler is good to assign this node to the pod.
	//	if len(feasibleNodes) != 0 {
	//		return feasibleNodes, diagnosis, nil
	//	}
	//}

	nodes := allNodes
	//if !preRes.AllNodes() {
	//	nodes = make([]*config.NodeInfo, 0, len(preRes.NodeNames))
	//	for nodeName := range preRes.NodeNames {
	//		// PreRes may return nodeName(s) which do not exist; we verify
	//		// node exists in the Snapshot.
	//		if nodeInfo, err := sched.nodeInfoSnapshot.Get(nodeName); err == nil {
	//			nodes = append(nodes, nodeInfo)
	//		}
	//	}
	//	diagnosis.NodeToStatus.SetAbsentNodesStatus(framework.NewStatus(framework.UnschedulableAndUnresolvable, fmt.Sprintf("node(s) didn't satisfy plugin(s) %v", sets.List(unscheduledPlugins))))
	//}
	feasibleNodes, err := sched.findNodesThatPassFilters(ctx, fwk, state, group, nodes)
	// always try to update the sched.nextStartNodeIndex regardless of whether an error has occurred
	// this is helpful to make sure that all the nodes have a chance to be searched
	//processedNodes := len(allNodes)
	//sched.nextStartNodeIndex = (sched.nextStartNodeIndex + processedNodes) % len(allNodes)
	//if err != nil {
	//	return nil, diagnosis, err
	//}
	//
	//feasibleNodesAfterExtender, err := findNodesThatPassExtenders(ctx, sched.Extenders, pod, feasibleNodes, diagnosis.NodeToStatus)
	//if err != nil {
	//	return nil, diagnosis, err
	//}
	//if len(feasibleNodesAfterExtender) != len(feasibleNodes) {
	//	// Extenders filtered out some nodes.
	//	//
	//	// Extender doesn't support any kind of requeueing feature like EnqueueExtensions in the scheduling framework.
	//	// When Extenders reject some Nodes and the pod ends up being unschedulable,
	//	// we put framework.ExtenderName to pInfo.UnschedulablePlugins.
	//	// This Pod will be requeued from unschedulable pod pool to activeQ/backoffQ
	//	// by any kind of cluster events.
	//	// https://github.com/kubernetes/kubernetes/issues/122019
	//	if diagnosis.UnschedulablePlugins == nil {
	//		diagnosis.UnschedulablePlugins = sets.New[string]()
	//	}
	//	diagnosis.UnschedulablePlugins.Insert(framework.ExtenderName)
	//}

	return feasibleNodes, nil
}

// findNodesThatPassFilters finds the nodes that fit the filter plugins.
func (sched *Scheduler) findNodesThatPassFilters(
	ctx context.Context,
	fwk framework.Framework,
	state *framework.CycleState,
	group *apis.Group,
	// TODO diagnosis *framework.Diagnosis,
	nodes []*config.NodeInfo) ([]*config.NodeInfo, error) {

	//没有插件 直接返回
	if !fwk.HasFilterPlugins() {
		return nodes, nil
	}

	feasibleNodes := make([]*config.NodeInfo, 0)
	//串行检查所有的设备节点是否要被过滤
	//TODO 并行优化 错误信息不全
	for _, node := range nodes {
		filterState := fwk.RunFilterPlugins(ctx, node, state, group)
		if filterState.Code() == framework.Error {
			logs.Error(filterState.AsError())
			continue
		}
		if filterState.IsSuccess() {
			feasibleNodes = append(feasibleNodes, node)
		} else {
			logs.Info(filterState.Message())
		}
	}
	return feasibleNodes, nil
}

func (sched *Scheduler) getAllNodes() ([]*config.NodeInfo, error) {
	return sched.getNodeFromApiServer(), nil
}

func (sched *Scheduler) getNodeFromApiServer() []*config.NodeInfo {
	scheme := runtime.NewScheme()
	apis.AddToScheme(scheme)
	c := &rest.Config{
		Host:    "http://localhost:10000",
		APIPath: "/apis/resources/v1",
		ContentConfig: rest.ContentConfig{
			AcceptContentTypes: "application/json; charset=UTF-8", //text/plain; charset=UTF-8
			ContentType:        "application/json; charset=UTF-8", //application/json; charset=UTF-8
			GroupVersion: &schema.GroupVersion{
				Group:   "resources",
				Version: "v1",
			},
			NegotiatedSerializer: serializer.NewCodecFactory(scheme),
		},
		UserAgent: "defaultUserAgent",
		Transport: &http.Transport{
			MaxIdleConns:        100,              // 最大空闲连接数
			IdleConnTimeout:     90 * time.Second, // 空闲连接超时时间
			TLSHandshakeTimeout: 10 * time.Second, // TLS 握手超时时间
		},
		Timeout: 10 * time.Second,
	}

	//创建ClientSet
	clientSet, err := clients.NewForConfig(c)
	if err != nil {
		panic(err)
	}
	// 资源定义在 pkg/apis/xxx/type.go 下
	// 这里以访问资源Node为例，
	// 获取访问Node的客户端
	// 默认访问的Namespace是 ""

	nodesClient := clientSet.Core().Nodes(sched.Namespace)
	lstOpts := metav1.ListOptions{}
	list, err := nodesClient.List(context.TODO(), lstOpts)
	if err != nil {
		panic(err)
	}
	nodes := make([]*config.NodeInfo, 0)
	for _, n := range list.Items {
		info := config.NewNodeInfo(&n)
		nodes = append(nodes, info)
	}
	return nodes
}

//func mockGetAllNodes() []*config.NodeInfo {
//	nodes := make([]*config.NodeInfo, 0)
//	node1 := &config.Node{}
//	node2 := &config.Node{}
//	info1 := &config.NodeInfo{}
//	info2 := &config.NodeInfo{}
//	info1.SetNode(node1)
//	info2.SetNode(node2)
//	nodes = append(nodes, info1, info2)
//	return nodes
//}

func prioritizeNodes(
	ctx context.Context,
	//先不考虑extender
	//extenders []framework.Extender,
	fwk framework.Framework,
	state *framework.CycleState,
	group *apis.Group,
	nodes []*config.NodeInfo,
) ([]framework.NodePluginScores, error) {
	// If no priority configs are provided, then all nodes will have a score of one.
	// This is required to generate the priority list in the required format
	if !fwk.HasScorePlugins() {
		logs.Warnf("no score plugins, use default score, group name %s", group.ObjectMeta.Name)
		result := make([]framework.NodePluginScores, 0, len(nodes))
		for i := range nodes {
			result = append(result, framework.NodePluginScores{
				Name:       nodes[i].Node().Name,
				TotalScore: 1,
			})
		}
		return result, nil
	}
	nodesScores, scoreStatus := fwk.RunScorePlugins(ctx, state, group, nodes)
	if !scoreStatus.IsSuccess() {
		return nil, scoreStatus.AsError()
	}
	return nodesScores, nil
}

type nodeScoreHeap []framework.NodePluginScores

func (h nodeScoreHeap) Len() int           { return len(h) }
func (h nodeScoreHeap) Less(i, j int) bool { return h[i].TotalScore > h[j].TotalScore }
func (h nodeScoreHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *nodeScoreHeap) Push(x interface{}) {
	*h = append(*h, x.(framework.NodePluginScores))
}

func (h *nodeScoreHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

func selectHost(nodeScoreList []framework.NodePluginScores, count int) (string, []framework.NodePluginScores, error) {
	if len(nodeScoreList) == 0 {
		return "", nil, errors.New("no nodes found")
	}

	var h nodeScoreHeap = nodeScoreList
	heap.Init(&h)
	cntOfMaxScore := 1
	selectedIndex := 0
	// The top of the heap is the NodeScoreResult with the highest score.
	sortedNodeScoreList := make([]framework.NodePluginScores, 0, count)
	sortedNodeScoreList = append(sortedNodeScoreList, heap.Pop(&h).(framework.NodePluginScores))

	// This for-loop will continue until all Nodes with the highest scores get checked for a reservoir sampling,
	// and sortedNodeScoreList gets (count - 1) elements.
	for ns := heap.Pop(&h).(framework.NodePluginScores); ; ns = heap.Pop(&h).(framework.NodePluginScores) {
		if ns.TotalScore != sortedNodeScoreList[0].TotalScore && len(sortedNodeScoreList) == count {
			break
		}

		if ns.TotalScore == sortedNodeScoreList[0].TotalScore {
			cntOfMaxScore++
			if rand.Intn(cntOfMaxScore) == 0 {
				// Replace the candidate with probability of 1/cntOfMaxScore
				selectedIndex = cntOfMaxScore - 1
			}
		}

		sortedNodeScoreList = append(sortedNodeScoreList, ns)

		if h.Len() == 0 {
			break
		}
	}

	if selectedIndex != 0 {
		// replace the first one with selected one
		previous := sortedNodeScoreList[0]
		sortedNodeScoreList[0] = sortedNodeScoreList[selectedIndex]
		sortedNodeScoreList[selectedIndex] = previous
	}

	if len(sortedNodeScoreList) > count {
		sortedNodeScoreList = sortedNodeScoreList[:count]
	}

	return sortedNodeScoreList[0].Name, sortedNodeScoreList, nil
}

func selectHostByProbability(nodeScoreList []framework.NodePluginScores) (string, error) {

	if len(nodeScoreList) == 0 {
		return "", errors.New("node score list is empty")
	}

	// 计算总分数
	total := int64(0)
	for _, score := range nodeScoreList {
		total += score.TotalScore
	}

	if total == 0 {
		return "", errors.New("all scores are zero, cannot select a host")
	}

	// 初始化随机数种子
	rand.Seed(time.Now().UnixNano())

	// 生成一个 0 到 total-1 之间的随机数
	randomNum := int64(rand.Intn(int(total)))
	// 根据随机数选择节点
	currentSum := int64(0)
	for _, score := range nodeScoreList {
		currentSum += score.TotalScore
		if randomNum < currentSum {
			return score.Name, nil
		}
	}

	return "", errors.New("unexpected error occurred while selecting a host")
}

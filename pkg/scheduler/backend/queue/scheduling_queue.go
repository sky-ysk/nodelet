package queue

import (
	"context"
	"fmt"
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/component-base/logs"
	"hit.edu/framework/pkg/scheduler/apis/config"
	sutils "hit.edu/framework/pkg/scheduler/utils"
	"hit.edu/framework/pkg/utils/value"
	"strconv"
	"strings"

	//scheutils "hit.edu/framework/pkg/scheduler/utils"
	"hit.edu/framework/pkg/utils"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/wait"
	"sync"
	"time"

	"hit.edu/framework/pkg/scheduler/workflow"
)

type SchedulingQueue interface {
	// TODO: 将Ready状态的任务迁移到Scheduling Queue中的组件

	Add(ctx context.Context, group *apis.Group)

	//Activate(groups *workflow.Group)

	//TODO: 添加到Unschedulable队列中

	//SchdulingCycle() int64

	//Done()

	Pop(ctx context.Context) (*config.QueuedGroupInfo, error)

	//Update()

	//Delete()

	//Close()

	Run(ctx context.Context)
}

// 优先级队列
// 优先级队列中头部表示最高优先级的PendingGroup
// 优先级队列下包含几个附属队列
//

type PriorityQueue struct {
	stop chan struct{}

	readyQ *readyQueue

	pendingQueue PendingGroups

	lock sync.RWMutex

	conditionEngine *utils.ConditionEngine

	valueEngine *value.Engine
}

func NewPriorityQueue() *PriorityQueue {
	logs.Info("NewPriorityQueue method")
	cs, err := sutils.CreateClientSetWithTimeOut(3600)
	if err != nil {
		logs.Error(err.Error())
		return nil
	}
	return &PriorityQueue{
		stop:            make(chan struct{}),
		readyQ:          newReadyQueue(),
		pendingQueue:    *newPendingQueue(),
		lock:            sync.RWMutex{},
		conditionEngine: utils.NewConditionEngine(),
		valueEngine:     value.NewEngine(cs),
	}
}

func (p *PriorityQueue) Pop(ctx context.Context) (*config.QueuedGroupInfo, error) {
	return p.readyQ.pop(ctx)
}

func (p *PriorityQueue) Add(ctx context.Context, group *apis.Group) {
	p.AddToPending(ctx, group)
}

// newQueuedPodInfo builds a QueuedPodInfo object.
func (p *PriorityQueue) newQueuedGroupInfo(group *apis.Group, plugins ...string) *config.QueuedGroupInfo {
	//now := p.clock.Now()
	groupInfo, _ := config.NewGroupInfo(group)
	return &config.QueuedGroupInfo{
		GroupInfo: groupInfo,
		//Timestamp:               now,
		InitialAttemptTimestamp: nil,
		UnschedulablePlugins:    sets.New(plugins...),
	}
}

// Add adds a pod to the active queue. It should be called only when a new pod
// is added so there is no chance the pod is already in active/unschedulable/backoff queues
func (p *PriorityQueue) AddToActive(ctx context.Context, group *apis.Group) {
	p.lock.Lock()
	defer p.lock.Unlock()
	gInfo := p.newQueuedGroupInfo(group)
	if added := p.moveToActiveQ(ctx, gInfo); added {
		msg := fmt.Sprintf("group %s now in active queue", gInfo.Group.ObjectMeta.Name)
		logs.Info(msg)
		p.readyQ.broadcast()
	}
}

// Add adds a pod to the active queue. It should be called only when a new pod
// is added so there is no chance the pod is already in active/unschedulable/backoff queues
func (p *PriorityQueue) AddToPending(ctx context.Context, group *apis.Group) {
	p.lock.Lock()
	defer p.lock.Unlock()
	gInfo := p.newQueuedGroupInfo(group)
	if added := p.moveToPendingQ(ctx, gInfo); added {
		//msg := fmt.Sprintf("group %s now in pending queue", gInfo.Group.ObjectMeta.Name)
		//logs.Info(msg)
		p.readyQ.broadcast()
	}
}

func (p *PriorityQueue) Run(ctx context.Context) {
	go wait.Until(func() {
		p.flushPendingQueue(ctx)
	}, 1.0*time.Second, p.stop)
}

func (p *PriorityQueue) flushPendingQueue(ctx context.Context) {
	p.lock.Lock()
	defer p.lock.Unlock()
	removeGroupss := make([]*config.QueuedGroupInfo, 0)
	//logs.Debug("now run the flush method")
	//fmt.Println("now run the flush method")
	for _, v := range p.pendingQueue.groupInfoMap {
		readyRes, err := p.checkGroupReady(ctx, v)
		if err != nil {
			logs.Error(err.Error())
			continue
		}
		if readyRes == apis.True {
			removeGroupss = append(removeGroupss, v)
			p.moveToActiveQ(ctx, v)
			p.readyQ.cond.Signal()
			//msg := fmt.Sprintf("group %s is ready , move to active queue", k)
			//fmt.Println(msg)
			//logs.Info(msg)
		}
	}
	for _, group := range removeGroupss {
		p.pendingQueue.underLockDelete(group)
		//msg := fmt.Sprintf("group %s is removed from pending queue", group.Group.Name)
		//fmt.Println(msg)
		//logs.Info(msg)
	}
}

// TODO 这个方法目前不完善，只检查了父母节点的依赖
func (p *PriorityQueue) checkGroupReady(ctx context.Context, gInfo *config.QueuedGroupInfo) (apis.ResultType, error) {
	//检查父节点完成情况
	for _, par := range gInfo.Group.Spec.Parents {
		fromStr := fmt.Sprintf("Group{%s}.Status{phase}", par)
		valueTmp := apis.Value{
			Name:      "",
			Type:      apis.LocalData,
			ValueType: apis.StringType,
			From:      fromStr,
		}
		parentValue, err := p.valueEngine.ExtractLocalValue(&valueTmp, *(gInfo.Group))
		if err != nil {
			logs.Error(err.Error())
			return apis.False, err
		}

		if parentValue.Value != string(apis.Successed) {
			//logs.Infof("parent group is not ready %s, Phase : %s", par, parentValue.Value)
			return apis.NotReady, nil
		}

	}

	return apis.True, nil
	//return p.conditionEngine.CheckConditions(gInfo.Group.Spec.Conditions)
}

func (p *PriorityQueue) moveToActiveQ(ctx context.Context, gInfo *config.QueuedGroupInfo) bool {
	//gatedBefore := pInfo.Gated
	//pInfo.Gated = !p.runPreEnqueuePlugins(context.Background(), pInfo)
	added := false
	p.readyQ.underLock(func(unlockedActiveQ unlockedReadyQueuer) {
		//TODO @linbohai time
		//if gInfo.InitialAttemptTimestamp == nil {
		//	now := p.clock.Now()
		//	gInfo.InitialAttemptTimestamp = &now
		//}
		unlockedActiveQ.AddOrUpdate(gInfo)
		added = true
		//p.unschedulablePods.delete(pInfo.Pod, gatedBefore)
		//_ = p.podBackoffQ.Delete(pInfo) // Don't need to react when pInfo is not found.
		//logger.V(5).Info("Pod moved to an internal scheduling queue", "pod", klog.KObj(pInfo.Pod), "event", event, "queue", activeQ)
		//metrics.SchedulerQueueIncomingPods.WithLabelValues("active", event).Inc()
		//if event == framework.EventUnscheduledPodAdd.Label() || event == framework.EventUnscheduledPodUpdate.Label() {
		//	p.AddNominatedPod(logger, pInfo.PodInfo, nil)
		//}
	})
	return added
}

func (p *PriorityQueue) moveToPendingQ(ctx context.Context, gInfo *config.QueuedGroupInfo) bool {
	return p.pendingQueue.underLockAddOrUpdate(gInfo)
}

// PodNominator abstracts operations to maintain nominated Pods.
type PodNominator interface {
	// AddNominatedPod adds the given pod to the nominator or
	// updates it if it already exists.
	//AddNominatedPod(logger klog.Logger, pod *PodInfo, nominatingInfo *NominatingInfo)
	// DeleteNominatedPodIfExists deletes nominatedPod from internal cache. It's a no-op if it doesn't exist.
	DeleteNominatedGroupIfExists(group *workflow.Group)
	// UpdateNominatedPod updates the <oldPod> with <newPod>.
	//UpdateNominatedPod(logger klog.Logger, oldPod *v1.Pod, newPodInfo *PodInfo)
	// NominatedPodsForNode returns nominatedPods on the given node.
	//NominatedPodsForNode(nodeName string) []*PodInfo
}

func groupInfoKeyFunc(gInfo *config.QueuedGroupInfo) string {
	key := fmt.Sprintf("%s-%s", gInfo.GroupInfo.Group.Name, gInfo.GroupInfo.Group.ObjectMeta.Name)
	return key
}

func groupInfoLessFunc(i, j *config.QueuedGroupInfo) bool {
	if i == nil || j == nil || i.Group == nil || j.Group == nil {
		return false
	}
	iScore := parseWeightFromQGroupInfo(i)
	jScore := parseWeightFromQGroupInfo(j)
	return iScore >= jScore
}

func parseWeightFromQGroupInfo(g *config.QueuedGroupInfo) int64 {
	if g == nil || g.Group == nil {
		return 0
	}
	if g.Group.Spec.Desc == nil || len(g.Group.Spec.Desc.Label) == 0 {
		return 1
	}

	val, ok := g.Group.Spec.Desc.Label["weight"]
	if !ok {
		return 1
	}
	valueStr := strings.TrimSpace(val)

	ret, err := strconv.ParseInt(valueStr, 10, 64)
	if err != nil {
		logs.Error(err.Error())
		return 1
	}
	return ret
}

func groupKeyFunc(g *apis.Group) string {
	key := fmt.Sprintf("%s-%s", g.Name, g.ObjectMeta.Name)
	return key
}

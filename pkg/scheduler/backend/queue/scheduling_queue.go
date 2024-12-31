package queue

import (
	"context"
	"fmt"
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/component-base/logs"
	"hit.edu/framework/pkg/scheduler/apis/config"
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
}

func NewPriorityQueue() *PriorityQueue {
	logs.Info("NewPriorityQueue method")
	return &PriorityQueue{
		stop:         make(chan struct{}),
		readyQ:       newReadyQueue(),
		pendingQueue: *newPendingQueue(),
		lock:         sync.RWMutex{},
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
		msg := fmt.Sprintf("group %s now in active queue", gInfo.Group.Spec.Name)
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
		msg := fmt.Sprintf("group %s now in pending queue", gInfo.Group.Spec.Name)
		logs.Info(msg)
		fmt.Println(msg)
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
	logs.Info("now run the flush method")
	//fmt.Println("now run the flush method")
	for k, v := range p.pendingQueue.groupInfoMap {
		if checkGroupReady(v) {
			removeGroupss = append(removeGroupss, v)
			p.moveToActiveQ(ctx, v)
			p.readyQ.cond.Signal()
			msg := fmt.Sprintf("group %s is ready , move to active queue", k)
			fmt.Println(msg)
			logs.Info(msg)
		}
	}
	for _, group := range removeGroupss {
		p.pendingQueue.underLockDelete(group)
		msg := fmt.Sprintf("group %s is removed from pending queue", group.Group.Spec.Name)
		fmt.Println(msg)
		logs.Info(msg)
	}
}

// TODO 这个方法等待后续完善
func checkGroupReady(gInfo *config.QueuedGroupInfo) bool {
	return true
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
	return gInfo.GroupInfo.Group.Name
}

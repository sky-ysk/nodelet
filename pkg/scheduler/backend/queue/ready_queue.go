package queue

import (
	"container/list"
	"context"
	"hit.edu/framework/pkg/scheduler/apis/config"
	"hit.edu/framework/pkg/scheduler/backend/heap"
	"sync"
)

// activeQueue implements activeQueuer. All of the fields have to be protected using the lock.
type readyQueue struct {
	// lock synchronizes all operations related to activeQ.
	// It protects activeQ, inFlightPods, inFlightEvents, schedulingCycle and closed fields.
	// Caution: DO NOT take "SchedulingQueue.lock" after taking "lock".
	// You should always take "SchedulingQueue.lock" first, otherwise the queue could end up in deadlock.
	// "lock" should not be taken after taking "nLock".
	// Correct locking order is: SchedulingQueue.lock > lock > nominator.nLock.
	lock sync.RWMutex

	// activeQ is heap structure that scheduler actively looks at to find pods to
	// schedule. Head of heap is the highest priority pod.
	queue *heap.Heap[*config.QueuedGroupInfo]

	// cond is a condition that is notified when the pod is added to activeQ.
	// It is used with lock.
	cond sync.Cond

	// inFlightPods holds the UID of all pods which have been popped out for which Done
	// hasn't been called yet - in other words, all pods that are currently being
	// processed (being scheduled, in permit, or in the binding cycle).
	//
	// The values in the map are the entry of each pod in the inFlightEvents list.
	// The value of that entry is the *v1.Pod at the time that scheduling of that
	// pod started, which can be useful for logging or debugging.
	inFlightPods map[string]*list.Element

	// inFlightEvents holds the events received by the scheduling queue
	// (entry value is clusterEvent) together with in-flight pods (entry
	// value is *v1.Pod). Entries get added at the end while the mutex is
	// locked, so they get serialized.
	//
	// The pod entries are added in Pop and used to track which events
	// occurred after the pod scheduling attempt for that pod started.
	// They get removed when the scheduling attempt is done, at which
	// point all events that occurred in the meantime are processed.
	//
	// After removal of a pod, events at the start of the list are no
	// longer needed because all of the other in-flight pods started
	// later. Those events can be removed.
	inFlightEvents *list.List

	// schedCycle represents sequence number of scheduling cycle and is incremented
	// when a pod is popped.
	schedCycle int64

	// closed indicates that the queue is closed.
	// It is mainly used to let Pop() exit its control loop while waiting for an item.
	closed bool

	// isSchedulingQueueHintEnabled indicates whether the feature gate for the scheduling queue is enabled.
	isSchedulingQueueHintEnabled bool
}

func newReadyQueue() *readyQueue {
	//TODO @linbohai 写个真的方法
	lessFn := func(i, j *config.QueuedGroupInfo) bool {
		return false
	}
	queue := heap.New(groupInfoKeyFunc, lessFn)
	rq := &readyQueue{
		queue:                        queue,
		inFlightPods:                 make(map[string]*list.Element),
		inFlightEvents:               list.New(),
		closed:                       false,
		isSchedulingQueueHintEnabled: false,
	}
	rq.cond.L = &rq.lock
	return rq
}

func (rq *readyQueue) underLock(fn func(unlockedActiveQ unlockedReadyQueuer)) {
	rq.lock.Lock()
	defer rq.lock.Unlock()
	fn(rq.queue)
}

// pop removes the head of the queue and returns it.
// It blocks if the queue is empty and waits until a new item is added to the queue.
// It increments scheduling cycle when a pod is popped.
// TODO @linbohai log
func (rq *readyQueue) pop(ctx context.Context) (*config.QueuedGroupInfo, error) {
	rq.lock.Lock()
	defer rq.lock.Unlock()
	return rq.unlockedPop()
}

// broadcast notifies the pop() operation that new pod(s) was added to the activeQueue.
func (rq *readyQueue) broadcast() {
	rq.cond.Broadcast()
}

// unlockedActiveQueueReader defines activeQ read-only methods that are not protected by the lock itself.
// underLock() or underRLock() method should be used to protect these methods.
type unlockedReadyQueueReader interface {
	Get(gInfo *config.QueuedGroupInfo) (*config.QueuedGroupInfo, bool)
	Has(gInfo *config.QueuedGroupInfo) bool
}

type unlockedReadyQueuer interface {
	unlockedReadyQueueReader
	AddOrUpdate(gInfo *config.QueuedGroupInfo)
}

func (rq *readyQueue) unlockedPop() (*config.QueuedGroupInfo, error) {
	for rq.queue.Len() == 0 {
		// When the queue is empty, invocation of Pop() is blocked until new item is enqueued.
		// When Close() is called, the p.closed is set and the condition is broadcast,
		// which causes this loop to continue and return from the Pop().
		if rq.closed {
			//logger.V(2).Info("Scheduling queue is closed")
			return nil, nil
		}
		rq.cond.Wait()
	}
	pInfo, err := rq.queue.Pop()
	if err != nil {
		return nil, err
	}
	pInfo.Attempts++
	// In flight, no concurrent events yet.
	if rq.isSchedulingQueueHintEnabled {
		// If the pod is already in the map, we shouldn't overwrite the inFlightPods otherwise it'd lead to a memory leak.
		// https://github.com/kubernetes/kubernetes/pull/127016
		if _, ok := rq.inFlightPods[string(pInfo.Group.UID)]; ok {
			// Just report it as an error, but no need to stop the scheduler
			// because it likely doesn't cause any visible issues from the scheduling perspective.
			//logger.Error(nil, "the same pod is tracked in multiple places in the scheduler, and just discard it", "pod", klog.KObj(pInfo.Pod))
			// Just ignore/discard this duplicated pod and try to pop the next one.
			return rq.unlockedPop()
		}

		//aq.metricsRecorder.ObserveInFlightEventsAsync(metrics.PodPoppedInFlightEvent, 1, false)
		rq.inFlightPods[string(pInfo.Group.UID)] = rq.inFlightEvents.PushBack(pInfo.Group)
	}
	rq.schedCycle++
	pInfo.UnschedulablePlugins.Clear()
	pInfo.PendingPlugins.Clear()

	return pInfo, nil
}

// delete deletes the pod info from activeQ.
//func (aq *activeQueue) delete(pInfo *framework.QueuedPodInfo) error {
//	aq.lock.Lock()
//	defer aq.lock.Unlock()
//
//	return aq.queue.Delete(pInfo)
//}

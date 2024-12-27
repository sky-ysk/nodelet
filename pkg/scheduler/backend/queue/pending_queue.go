package queue

import (
	"fmt"
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/component-base/logs"
	"hit.edu/framework/pkg/scheduler/apis/config"
	"sync"
)

type unlockedPendingQueuer interface {
	unlockedReadyQueueReader
	AddOrUpdate(gInfo *config.QueuedGroupInfo)
}

// unlockedActiveQueueReader defines activeQ read-only methods that are not protected by the lock itself.
// underLock() or underRLock() method should be used to protect these methods.
type unlockedPendingQueueReader interface {
	Get(gInfo *config.QueuedGroupInfo) (*config.QueuedGroupInfo, bool)
	Has(gInfo *config.QueuedGroupInfo) bool
}

type PendingGroups struct {
	lock         sync.Mutex
	groupInfoMap map[string]*config.QueuedGroupInfo
	keyFunc      func(group *apis.Group) string
	// unschedulableRecorder/gatedRecorder updates the counter when elements of an unschedulablePodsMap
	// get added or removed, and it does nothing if it's nil.
	//unschedulableRecorder, gatedRecorder metrics.MetricRecorder
}

func newPendingQueue() *PendingGroups {
	pg := &PendingGroups{
		groupInfoMap: make(map[string]*config.QueuedGroupInfo),
		keyFunc:      func(group *apis.Group) string { return group.Spec.Name },
	}
	return pg
}

func (pg *PendingGroups) underLockAddOrUpdate(gInfo *config.QueuedGroupInfo) bool {
	pg.lock.Lock()
	defer pg.lock.Unlock()
	key := pg.keyFunc(gInfo.Group)
	if _, ok := pg.groupInfoMap[key]; ok {
		msg := fmt.Sprintf("group %s is already in pending queue", gInfo.Group.Spec.Name)
		logs.Info(msg)
	}
	pg.groupInfoMap[key] = gInfo
	return true
}

func (pg *PendingGroups) underLockDelete(gInfo *config.QueuedGroupInfo) {
	pg.lock.Lock()
	defer pg.lock.Unlock()
	key := pg.keyFunc(gInfo.Group)
	delete(pg.groupInfoMap, key)
}

package group

import (
	"fmt"
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/component-base/logs"
	"sync"
)

// 队列的key目前都填groupID；groupStatus-GroupID
type GroupQueues struct {
	pendingQueue   map[string]*apis.Group
	runningQueue   sync.Map
	errorQueue     map[string]*apis.Group
	completedQueue map[string]*apis.Group
	groupManager   Manager
}

func NewGroupQueues(groupManager Manager) *GroupQueues {
	return &GroupQueues{
		pendingQueue:   make(map[string]*apis.Group),
		runningQueue:   sync.Map{},
		errorQueue:     make(map[string]*apis.Group),
		completedQueue: make(map[string]*apis.Group),
		groupManager:   groupManager,
	}
}

// add
func (gq *GroupQueues) AddToPending(key string, value *apis.Group) bool {
	if _, exists := gq.pendingQueue[key]; exists {
		logs.Info("%v has been added to pending queue", key)
		return false
	}
	gq.pendingQueue[key] = value
	return true
}
func (gq *GroupQueues) AddToRunning(key string, value *apis.Group) bool {
	if _, exists := gq.runningQueue.Load(key); exists {
		logs.Info("%v has been added to running queue", key)
		return false
	}
	gq.runningQueue.Store(key, value)
	return true
}
func (gq *GroupQueues) AddToError(key string, value *apis.Group) bool {
	if _, exists := gq.errorQueue[key]; exists {
		logs.Info("%v has been added to error queue", key)
		return false
	}
	gq.errorQueue[key] = value
	return true
}
func (gq *GroupQueues) AddToCompleted(key string, value *apis.Group) bool {
	if _, exists := gq.completedQueue[key]; exists {
		logs.Info("%v has been added to completed queue", key)
		return false
	}
	gq.completedQueue[key] = value
	return true
}

// delete
func (gq *GroupQueues) DeleteFromPending(key string) bool {
	if _, exists := gq.pendingQueue[key]; !exists {
		logs.Info("%v not in pending queue", key)
		return false
	}
	delete(gq.pendingQueue, key)
	return true
}
func (gq *GroupQueues) DeleteFromRunning(key string) bool {
	if _, exists := gq.runningQueue.Load(key); !exists {
		logs.Info("%v not in running queue", key)
		return false
	}
	gq.runningQueue.Delete(key)
	return true
}
func (gq *GroupQueues) DeleteFromError(key string) bool {
	if _, exists := gq.errorQueue[key]; !exists {
		logs.Info("%v not in Error queue", key)
		return false
	}
	delete(gq.errorQueue, key)
	return true
}
func (gq *GroupQueues) DeleteFromCompleted(key string) bool {
	if _, exists := gq.completedQueue[key]; !exists {
		logs.Info("%v not in Completed queue", key)
		return false
	}
	delete(gq.completedQueue, key)
	return true
}

// get
func (gq *GroupQueues) GetFromPending(key string) (*apis.Group, bool) {
	value, ok := gq.pendingQueue[key]
	return value, ok
}
func (gq *GroupQueues) GetFromRunning(key string) (*apis.Group, bool) {
	value, ok := gq.runningQueue.Load(key)
	if !ok {
		return nil, false
	}
	return value.(*apis.Group), ok
}
func (gq *GroupQueues) GetFromError(key string) (*apis.Group, bool) {
	value, ok := gq.errorQueue[key]
	return value, ok
}

func (gq *GroupQueues) GetFromCompleted(key string) (*apis.Group, bool) {
	value, ok := gq.completedQueue[key]
	return value, ok
}

func (gq *GroupQueues) GetAllRunning() []*apis.Group {
	values := make([]*apis.Group, 0)
	gq.runningQueue.Range(func(key, value interface{}) bool {
		values = append(values, value.(*apis.Group))
		return true
	})
	return values
}

func (gq *GroupQueues) GetAllPending() []*apis.Group {
	values := make([]*apis.Group, 0)
	for _, value := range gq.pendingQueue {
		values = append(values, value)
	}
	return values
}

func (gq *GroupQueues) GetAllError() []*apis.Group {
	values := make([]*apis.Group, 0)
	for _, value := range gq.errorQueue {
		values = append(values, value)
	}
	return values
}

func (gq *GroupQueues) GetAllCompleted() []*apis.Group {
	values := make([]*apis.Group, 0)
	for _, value := range gq.completedQueue {
		values = append(values, value)
	}
	return values
}

// 删除任务信息从queue_manager，同时同步group_manager
func (gq *GroupQueues) DeleteGroup(group *apis.Group) {
	groupID := group.Status.GroupID
	var err error
	if state, ok := gq.getFromQueue(groupID); ok {
		switch state {
		case "pending":
			gq.DeleteFromPending(groupID)
		case "running":
			gq.DeleteFromRunning(groupID)
		case "error":
			gq.DeleteFromPending(groupID)
		case "completed":
			gq.DeleteFromCompleted(groupID)
		default:
			err = fmt.Errorf("Unfind group named %s", groupID)
		}
	}
	if err == nil {
		gq.groupManager.DeleteGroup(group)
	}
}

// 更新任务状态并同步
func (gq *GroupQueues) UpdateGroup(groupID string, group *apis.Group) error {
	//更新队列中的任务状态
	var err error
	if state, ok := gq.getFromQueue(groupID); ok {
		switch state {
		case "pending":
			logs.Info("pending group update")
			gq.pendingQueue[groupID] = group
		case "running":
			logs.Info("running group update")
			gq.runningQueue.Store(groupID, group)
		case "error":
			logs.Info("error group update")
			gq.errorQueue[groupID] = group
		case "completed":
			logs.Info("completed group update")
			gq.completedQueue[groupID] = group
		default:
			err = fmt.Errorf("Unfind group named %s", groupID)
		}
	}
	if err == nil {
		gq.groupManager.UpdateGroup(group)
	}
	return err
}

// 从queue队列中找到group所属的队列
func (gq *GroupQueues) getFromQueue(groupID string) (string, bool) {
	if _, ok := gq.pendingQueue[groupID]; ok {
		return "pending", true
	}
	if _, ok := gq.runningQueue.Load(groupID); ok {
		return "running", true
	}
	if _, ok := gq.completedQueue[groupID]; ok {
		return "completed", true
	}
	if _, ok := gq.errorQueue[groupID]; ok {
		return "error", true
	}
	return "", false
}

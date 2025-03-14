package task

import (
	"fmt"
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/component-base/logs"
	"sync"
)

type Manager interface {
	// 获取所有的Group
	GetTasks(Map map[string]*apis.Task) []*apis.Task

	//设置Group
	SetTasks([]*apis.Task)

	//增加Group
	AddTask(*apis.Task)

	//更新Group
	UpdateTask(*apis.Task)

	//删除Group
	DeleteTask(*apis.Task)

	// 获取Task通过ID
	GetTaskByID(taskID string) (*apis.Task, error) // 添加这个方法
}
type taskManager struct {
	// lock
	lock       sync.RWMutex
	modifyLock sync.Mutex
	// 存储所有的Task, 按照Task Name进行索引
	// TODO: TaskName并不是唯一的，TaskID是唯一的，是调度后由调度框架分配的
	tasksByName map[string]*apis.Task

	// TODO: 存储所有的Task, 按照TaskID进行索引
	tasksByID map[string]*apis.Task
}

func NewTaskManager() Manager {
	gm := &taskManager{
		tasksByName: make(map[string]*apis.Task),
		tasksByID:   make(map[string]*apis.Task),
	}
	gm.SetTasks(nil)
	return gm
}

func (gm *taskManager) GetTaskByID(taskID string) (*apis.Task, error) {
	gm.lock.RLock()
	defer gm.lock.RUnlock()

	task, exists := gm.tasksByID[taskID]
	if !exists {
		return nil, fmt.Errorf("task with id %s not found", taskID)
	}
	return task, nil
}
func (gm *taskManager) GetTasks(Map map[string]*apis.Task) []*apis.Task {
	gm.lock.RLock()
	defer gm.lock.RUnlock()

	tasks := make([]*apis.Task, 0, len(Map))
	for _, task := range Map {
		tasks = append(tasks, task)
	}
	return tasks
}

func (gm *taskManager) SetTasks(tasks []*apis.Task) {
	gm.modifyLock.Lock()
	defer gm.modifyLock.Unlock()

	for _, g := range tasks {
		gm.tasksByName[g.Name] = g
	}
}

func (gm *taskManager) AddTask(task *apis.Task) {
	//TODO implement me
	gm.modifyLock.Lock()
	defer gm.modifyLock.Unlock()
	//检查TaskID是否已经存在
	if _, exists := gm.tasksByID[task.Status.TaskID]; exists {
		logs.Errorf("Task With ID %s is existed", task.Status.TaskID)
		return
	}
	gm.tasksByID[task.Status.TaskID] = task
	gm.tasksByName[task.Name] = task
}

func (gm *taskManager) UpdateTask(task *apis.Task) {
	//TODO implement me
	gm.modifyLock.Lock()
	defer gm.modifyLock.Unlock()
	gm.tasksByID[task.Status.TaskID] = task
	gm.tasksByName[task.Name] = task
}

func (gm *taskManager) DeleteTask(task *apis.Task) {
	//TODO implement me
	gm.modifyLock.Lock()
	defer gm.modifyLock.Unlock()
	delete(gm.tasksByID, task.Status.TaskID)
	delete(gm.tasksByName, task.Name)
}

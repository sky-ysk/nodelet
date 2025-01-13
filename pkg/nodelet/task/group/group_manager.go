package group

import (
	"fmt"
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/component-base/logs"
	"sync"
)

// 管理当前节点上运行的所有Group
// TODO: Group包含Actions, Action相关的运行方式

type Manager interface {
	// 获取所有的Group
	GetGroups(Map map[string]*apis.Group) []*apis.Group

	//设置Group
	SetGroups([]*apis.Group)

	//增加Group
	AddGroup(*apis.Group)

	//更新Group
	UpdateGroup(*apis.Group)

	//删除Group
	DeleteGroup(*apis.Group)

	// 获取Group通过Name
	GetGroupByID(groupID string) (*apis.Group, error) // 添加这个方法
}

type groupManager struct {
	// lock
	lock       sync.RWMutex
	modifyLock sync.Mutex
	// 存储所有的Group, 按照Group Name进行索引
	// TODO: GroupName并不是唯一的，GroupID是唯一的，是调度后由调度框架分配的
	groupsByName map[string]*apis.Group

	// TODO: 存储所有的Group, 按照GroupID进行索引
	groupsByID map[string]*apis.Group
}

func NewGroupManager() Manager {
	gm := &groupManager{
		groupsByName: make(map[string]*apis.Group),
		groupsByID:   make(map[string]*apis.Group),
	}
	gm.SetGroups(nil)
	return gm
}

func (gm *groupManager) GetGroups(Map map[string]*apis.Group) []*apis.Group {
	gm.lock.RLock()
	defer gm.lock.RUnlock()

	groups := make([]*apis.Group, 0, len(Map))
	for _, group := range Map {
		groups = append(groups, group)
	}
	return groups
}

func (gm *groupManager) GetGroupByID(groupID string) (*apis.Group, error) {
	gm.lock.RLock()
	defer gm.lock.RUnlock()

	group, exists := gm.groupsByID[groupID]
	if !exists {
		return nil, fmt.Errorf("group with ID： %s not found", groupID)
	}
	return group, nil
}

func (gm *groupManager) SetGroups(groups []*apis.Group) {
	gm.modifyLock.Lock()
	defer gm.modifyLock.Unlock()

	for _, g := range groups {
		gm.groupsByName[g.Name] = g
	}
}

func (gm *groupManager) AddGroup(group *apis.Group) {
	//TODO implement me
	gm.modifyLock.Lock()
	defer gm.modifyLock.Unlock()
	//检查GroupID是否已经存在
	if _, exists := gm.groupsByID[group.Status.GroupID]; exists {
		logs.Errorf("Group With ID %s is existed", group.Status.GroupID)
		return
	}
	gm.groupsByID[group.Status.GroupID] = group
	gm.groupsByName[group.Name] = group
}

func (gm *groupManager) UpdateGroup(group *apis.Group) {
	//TODO implement me
	gm.modifyLock.Lock()
	defer gm.modifyLock.Unlock()
	gm.groupsByID[group.Status.GroupID] = group
	gm.groupsByName[group.Name] = group
}

func (gm *groupManager) DeleteGroup(group *apis.Group) {
	//TODO implement me
	gm.modifyLock.Lock()
	defer gm.modifyLock.Unlock()
	delete(gm.groupsByID, group.Status.GroupID)
	delete(gm.groupsByName, group.Name)
}

package util

import (
	apis "hit.edu/framework/pkg/apis/cores"
	"time"
)

type Manager struct {
	// TODO: 集成客户端

}

// 创建一个随机串UUID，该串由时间+10位随机串构成
// UUID应该唯一
// 同一个Workflow/Task/Group/Action/Runtime的UUID应该相同

// 根据GroupSpec创建Action
func CreateActions(gs apis.GroupSpec, namespace string, uuid string) ([]*apis.Action, error) {
	actions := []*apis.Action{}
	for _, as := range gs.Actions {
		a, err := CreateAction(as, &gs, namespace, uuid)
		if err != nil {
			return nil, err
		}
		// TODO: 返回的应该是已经创建的Runtime
		actions = append(actions, a) // a替换成实际的fa
	}

	return actions, nil
}

func CreateAction(as apis.ActionSpec, gs *apis.GroupSpec, namespace string, uuid string) (*apis.Action, error) {
	// 临时创建一个Action对象
	a := apis.Action{}
	// 构造名称
	a.Name = as.Name + uuid
	// 构造Namespace
	if namespace == "" {
		a.Namespace = apis.NamespaceDefault
	} else {
		a.Namespace = namespace
	}

	// 复制Spec
	a.Spec = as

	// 构造Status
	a.Status = apis.ActionStatus{}

	// 记录Create时间
	a.Status.CreateAt = apis.Time{time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC)}

	// 初始化状态
	a.Status.Phase = apis.Pending

	// 打上Label, 当前任务属于哪个Group和uuid域
	if gs != nil {
		a.Labels["group"] = gs.Name + uuid
	}
	a.Labels["uuid"] = uuid

	// 根据Spec创建Runtimes
	runtimes, err := CreateRuntimes(as, namespace, uuid)
	if err != nil {
		return nil, err
	}

	// 根据生成的Runtime修改Action.Status.Runtimes
	for _, r := range runtimes {
		a.Status.Runtimes[r.Spec.Name] = apis.ObjectReference{
			Name:      r.Name,
			Namespace: r.Namespace,
			Kind:      r.Kind,
		}
	}
	// TODO: 写入数据库
	return &a, nil // TODO: 应当返回实际的fa
}

// 根据ActionSpec创建Runtime
func CreateRuntimes(as apis.ActionSpec, namespace string, uuid string) ([]*apis.Runtime, error) {
	runtimes := []*apis.Runtime{}
	// 遍历所有的Runtime Spec
	for _, rs := range as.Runtimes {
		r, err := CreateRuntime(rs, &as, namespace, uuid)
		if err != nil {
			return nil, err
		}
		// TODO: 返回的应该是已经创建的Runtime
		runtimes = append(runtimes, r) // r替换成实际的fr
	}
	return runtimes, nil
}

// 创建单个Runtime
func CreateRuntime(rs apis.RuntimeSpec, as *apis.ActionSpec, namespace string, uuid string) (*apis.Runtime, error) {
	// 临时创建一个Runtime对象
	r := apis.Runtime{}
	// 构造名称
	r.Name = rs.Name + uuid
	// 构造Namespace
	if namespace == "" {
		r.Namespace = apis.NamespaceDefault
	} else {
		r.Namespace = namespace
	}
	// 复制Spec
	r.Spec = rs

	// 构造Status
	r.Status = apis.RuntimeStatus{}

	// 记录Create时间
	r.Status.CreateAt = apis.Time{time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC)}

	// 初始化状态
	r.Status.Phase = apis.Pending

	// 打上Label, 当前任务属于哪个action和uuid域
	if as != nil {
		r.Labels["action"] = as.Name + uuid
	}
	r.Labels["uuid"] = uuid

	// TODO: 写入Client-Go中, 返回实际的Runtime
	//fr, err := WriteTo

	return &r, nil // TODO: 应当返回fr
}

package ability

import (
	"fmt"
	"hit.edu/framework/pkg/component-base/logs"
	"hit.edu/framework/pkg/nodelet/task/runtime/device/ability/inst"
	"time"
)

type Manager struct {
	State  inst.AbilityState // 能力状态
	UUid   string
	TaskId string // 能力对应的taskId
	Url    string // 能力访问的rl
	Name   string // 能力名字
}

type Ability interface {
	StartupAbility() error
	PauseAbility()
	TerminateAbility() error
}

func NewAbilityManager(url string, name string) *Manager {
	return &Manager{
		Url:  url,
		Name: name,
	}
}
func (am *Manager) GetUUID() (string, error) {
	// 首先获取所有的ability信息
	instances, err := inst.GetAbilityInstances(am.Url)
	if err != nil {
		return "", err
	}
	// 通过AbilityName找到对应的uuid
	id, err := inst.FindIdByAbilityName(am.Name, instances)
	am.UUid = id
	if err != nil {
		logs.Info("can not find the id by name\n")
		return "", err
	}
	logs.Info("find the id by name\n")
	return id, nil
}

// StartupAbility 启动一个能力
func (am *Manager) StartupAbility() error {
	// 获取能力的uuid
	id, err := am.GetUUID()
	if err != nil {
		logs.Info("can not get uuid\n")
		return err
	}
	logs.Info("find ability's uuid\n")
	// 获取能力的taskId
	taskId, err := inst.PostLifeCycleRequest(id, inst.Start, am.Url)
	if err != nil {
		logs.Info("can not post lifecycle request and obtain taskId\n")
		return err
	}
	am.TaskId = taskId
	fmt.Println("taskId is ", taskId)
	for {
		time.Sleep(2000 * time.Millisecond)
		var state inst.AbilityState
		logs.Infof("getting ability state\n")
		state, err = inst.GetAbilityState(am.Url, am.UUid)
		if err != nil {
			logs.Info("can not get ability state\n")
			return err
		}
		switch state {
		case inst.Standby: // 进入standby状态说明启动成功，可以connect了
			am.State = state
			logs.Info("the ability state is standby, publish connect command...\n")
			taskId, err = inst.PostLifeCycleRequest(id, inst.Connect, am.Url)
			if err != nil {
				logs.Info("can not post lifecycle request and obtain taskId\n")
				return err
			}
			logs.Info("successfully post lifecycle request and obtain taskId\n")

		case inst.Running: // 进入running状态说明程序正在运行了
			logs.Info("the ability state is running, startup ability successfully\n")
			am.State = state
			return nil
		case inst.Error: // 进入error状态说明程序进入错误
			logs.Info("the ability state is error, startup ability fail\n")
			am.State = state
			return fmt.Errorf("ability is in error state")
		default:
			logs.Infof("the ability state is %v", state)
			am.State = state

		}
	}

}

// PauseAbility 暂停一个能力
func PauseAbility() {

}

// TerminateAbility 终止一个能力
func (am *Manager) TerminateAbility() error {

	// 发送terminate指令
	_, err := inst.PostLifeCycleRequest(am.UUid, inst.Terminate, am.Url)
	if err != nil {
		logs.Info("can not post lifecycle request and obtain taskId\n")
		return err
	}

	for {
		time.Sleep(200 * time.Millisecond)
		var state inst.AbilityState
		logs.Infof("getting ability state\n")
		state, err = inst.GetAbilityState(am.Url, am.UUid)
		if err != nil {
			return err
		}
		logs.Info("successfully get ability state\n")
		switch state {
		case inst.Terminating:
			logs.Info("the ability state is terminating,waiting....\n")
			am.State = state
		case inst.Inactive:
			logs.Info("the ability state is inactive, terminate successfully\n")
			am.State = state
			return nil
		case inst.Terminated:
			logs.Info("the ability state is terminated, terminate successfully\n")
			am.State = state
			return nil
		}
	}
}

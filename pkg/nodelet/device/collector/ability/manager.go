package manager

import (
	"fmt"
	"hit.edu/framework/pkg/component-base/logs"
	"time"
)

type ManagerOfAbility struct {
	State  AbilityState // 能力状态
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

func NewAbilityManager(url string, name string) *ManagerOfAbility {
	return &ManagerOfAbility{
		Url:  url,
		Name: name,
	}
}

// BindUUID 进行ManagerOfAbility与UUID的绑定
func (am *ManagerOfAbility) BindUUID() error {
	// 首先获取所有的ability信息
	instances, err := GetAbilityInstances(am.Url)
	if err != nil {
		logs.Errorf("[DEVICE EXPORTER] get ability instances fail")
		return fmt.Errorf("[DEVICE EXPORTER] get ability instances fail")
	}

	// 通过AbilityName找到所有对应的uuid
	var id string
	id, err = FindIdByAbilityName(am.Name, instances)
	if err != nil {
		logs.Errorf("[DEVICE EXPORTER] Find UUId By AbilityName fail")
		return err
	}
	// 填写相应字段
	am.UUid = id
	return nil
}

func (am *ManagerOfAbility) GetUUID() (string, error) {
	return am.UUid, nil
}

// GetHeartBeat 获取对应的心跳包
func (am *ManagerOfAbility) GetHeartBeat() (HeartBeat, error) {
	heartBeat, err := GetAbilityState(am.Url, am.UUid)
	if err != nil {
		logs.Errorf("[DEVICE EXPORTER] get heartbeat fail")
		return HeartBeat{}, err
	}
	return heartBeat, nil
}

func (am *ManagerOfAbility) IsOnline() (bool, error) {
	// 获取全部的心跳包
	logs.Infof("[DEVICE EXPORTER] Try to get heart beat......\n")
	hearBeats, err := GetAbilityHeartBeat(am.Url)
	if err != nil {
		logs.Error("[DEVICE EXPORTER] Get heart beats error\n")
		return false, err
	}

	// 遍历全部的心跳包
	for _, hearBeat := range hearBeats {
		// 如果找到了对应的id 说明在线
		if hearBeat.ID == am.UUid {
			return true, nil
		}
	}

	// 没找到返回false
	return false, nil
}

// StartupAbility 启动一个能力
func (am *ManagerOfAbility) StartupAbility() (HeartBeat, error) {
	// 获取能力的uuid
	id, err := am.GetUUID()
	if err != nil {
		logs.Info("can not get uuid\n")
		return HeartBeat{}, err
	}
	logs.Info("find ability's uuid\n")
	// 获取能力的taskId
	taskId, err := PostLifeCycleRequest(id, Start, am.Url)
	if err != nil {
		logs.Info("can not post lifecycle request and obtain taskId\n")
		return HeartBeat{}, err
	}
	am.TaskId = taskId
	fmt.Println("taskId is ", taskId)
	for {
		time.Sleep(5000 * time.Millisecond)
		var state AbilityState
		var heartBeat HeartBeat
		logs.Infof("getting ability state\n")
		logs.Infof("uuid is%v", am.UUid)
		heartBeat, err = GetAbilityState(am.Url, am.UUid)
		state = heartBeat.State
		if err != nil {
			logs.Info("can not get ability state\n")
			return HeartBeat{}, err
		}
		switch state {
		case Standby: // 进入standby状态说明启动成功，可以connect了
			am.State = state
			logs.Info("the ability state is standby, publish connect command...\n")
			taskId, err = PostLifeCycleRequest(id, Connect, am.Url)
			if err != nil {
				logs.Info("can not post lifecycle request and obtain taskId\n")
				return HeartBeat{}, err
			}
			logs.Info("successfully post lifecycle request and obtain taskId\n")

		case Running: // 进入running状态说明程序正在运行了
			logs.Info("the ability state is running, startup ability successfully\n")
			am.State = state
			return heartBeat, nil
		case Error: // 进入error状态说明程序进入错误
			logs.Info("the ability state is error, startup ability fail\n")
			am.State = state
			return HeartBeat{}, fmt.Errorf("ability is in error state")
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
func (am *ManagerOfAbility) TerminateAbility() error {

	// 获取能力的uuid
	id, err := am.GetUUID()
	if err != nil {
		logs.Info("can not get uuid\n")
		return err
	}
	logs.Info("find ability's uuid\n")
	am.UUid = id
	// 发送terminate指令
	_, err = PostLifeCycleRequest(am.UUid, Terminate, am.Url)
	if err != nil {
		logs.Info("can not post lifecycle request and obtain taskId\n")
		return err
	}

	for {
		time.Sleep(200 * time.Millisecond)
		var state AbilityState
		var heartBeat HeartBeat
		logs.Infof("getting ability state\n")
		heartBeat, err = GetAbilityState(am.Url, am.UUid)
		state = heartBeat.State
		if err != nil {
			return err
		}
		logs.Info("successfully get ability state\n")
		switch state {
		case Terminating:
			logs.Info("the ability state is terminating,waiting....\n")
			am.State = state
		case Inactive:
			logs.Info("the ability state is inactive, terminate successfully\n")
			am.State = state
			return nil
		case Terminated:
			logs.Info("the ability state is terminated, terminate successfully\n")
			am.State = state
			return nil
		default:
			return fmt.Errorf("can not get ability state")
		}

	}
}

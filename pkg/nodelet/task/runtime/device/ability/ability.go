package ability

import (
	"fmt"
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/component-base/logs"
	"hit.edu/framework/pkg/nodelet/task/runtime/device/ability/lib"
	"hit.edu/framework/pkg/nodelet/task/runtime/device/ability/manager"
	"strconv"
	"strings"
	"time"
)

func PublishAbilityInst(Inst string, device *apis.Device, operation string) (apis.Output, error) {
	// 能力操作的适配分为两种
	parts := strings.Split(Inst, "_")
	if parts[0] == "manage" { // 以manage开头的是 拉起 暂停 终止能力等操作
		abilityManager := manager.NewAbilityManager(device.Spec.AccessMethod.URL, parts[1])

		if operation == "terminate" {
			err := abilityManager.TerminateAbility()
			if err != nil {
				logs.Errorf("terminate ability %v failed", parts[1])
				return apis.Output{}, err
			}
			logs.Infof("ability%s is terminated......", abilityManager.Name)
			return apis.Output{}, err
		}

		heartBeat, err := abilityManager.StartupAbility()
		if err != nil {
			logs.Errorf("start ability frame fail...")
			return apis.Output{}, err
		}
		logs.Info("start AbilityFramework successfully")
		// 单独开一个协程来监控运行结果
		go monitorDeviceAbility(abilityManager)
		for index, a := range device.Status.Abilities {
			for index, s := range a.Services {
				port := strconv.Itoa(heartBeat.AbilityPort)
				s.Port = port
				a.Services[index] = s
			}
			device.Status.Abilities[index] = a
		}
		return apis.Output{
			Value:     strconv.Itoa(heartBeat.AbilityPort),
			Name:      "port",
			ValueType: "int",
			Type:      apis.ResultsData,
		}, nil

	} else if parts[0] == "service" { // 以service开头的是每一次业务逻辑
		logs.Infof("publish service %s inst ......", parts[1])
		output, err := lib.SendServiceRequest(parts[1], device)
		if err != nil {
			return apis.Output{}, err
		}
		return output, nil
	}
	logs.Errorf("inst is invalid")
	return apis.Output{}, fmt.Errorf("inst is invalid")
}

func monitorDeviceAbility(am *manager.ManagerOfAbility) {
	logs.Infof("monitoring the AbilityFrameWork status....")
	go func() {
		for {
			hb, err := manager.GetAbilityHeartBeat(am.Url)
			if err != nil {
				logs.Errorf("get heartbeats fail")
			}
			logs.Infof("[AbilityFrameWork Monitor]ability: %v  status:%v", am.Name, hb[0].State)
			time.Sleep(time.Millisecond * 1500)
		}
	}()
	logs.Infof("monitoring the mock ability status....")
	go func() {
		for {
			as, err := manager.GetAbilityExeStatus(am.UUid, am.Url)
			if err != nil {
				logs.Errorf("get mock status fail")
			}
			logs.Infof("[AbilityExe Monitor]ability: %v status %v", am.Name, as)
			time.Sleep(time.Millisecond * 1500)
		}
	}()

}

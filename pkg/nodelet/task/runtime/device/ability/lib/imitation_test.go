package lib

import (
	"fmt"
	"hit.edu/framework/pkg/component-base/logs"
	"hit.edu/framework/pkg/nodelet/task/runtime/device/ability/manager"
	"strconv"
	"testing"
	"time"
)

func TestPublishStartTask(t *testing.T) {
	logs.Init("test")
	managerUrl := "http://192.168.8.197:8080" // 能力框架url
	abilityName := "ActInferenceAbility"
	url := "http://192.168.8.197" // 业务url
	param := 0
	am := manager.NewAbilityManager(managerUrl, abilityName)
	heartBeat, err := am.StartupAbility()
	if err != nil {
		logs.Errorf("start up %s fail", abilityName)
	} else {
		logs.Infof("heartBeat is %v", heartBeat)

		// 通过heartbeat中的abilityPort拼接成新的url

		serviceUrl := fmt.Sprintf("%s:%s%s", url, strconv.Itoa(heartBeat.AbilityPort), "/api/start_task")
		resp, err := PublishStartTask(param, serviceUrl)
		if err != nil {
			logs.Errorf("error publish start_task inst %v", err)
		}
		logs.Infof("resp is %v", resp)
	}
	time.Sleep(15 * time.Second)
	//终止能力[备用]
	err = am.TerminateAbility()
	if err != nil {
		logs.Errorf("error is %v", err)
	}
}

func TestPublishGoStandBy(t *testing.T) {
	logs.Init("test")
	managerUrl := "http://192.168.8.197:8080"
	abilityName := "ActInferenceAbility"
	url := "http://192.168.8.197" // 业务url

	am := manager.NewAbilityManager(managerUrl, abilityName)
	heartBeat, err := am.StartupAbility()

	if err != nil {
		logs.Errorf("start up %s fail", abilityName)
	} else {
		logs.Infof("heartBeat is %v", heartBeat)

		// 通过heartbeat中的abilityPort拼接成新的url

		param := 0
		serviceUrl1 := fmt.Sprintf("%s:%s%s", url, strconv.Itoa(heartBeat.AbilityPort), "/api/go_standby")
		resp1, err := PublishGoStandBy(param, serviceUrl1)
		if err != nil {
			logs.Errorf("error publish go_standby inst %v", err)
		}
		logs.Infof("resp is %v", resp1)

		time.Sleep(4 * time.Second)

		param = 0
		serviceUrl2 := fmt.Sprintf("%s:%s%s", url, strconv.Itoa(heartBeat.AbilityPort), "/api/start_task")
		resp2, err := PublishStartTask(param, serviceUrl2)
		if err != nil {
			logs.Errorf("error publish start_task inst %v", err)
		}
		logs.Infof("resp is %v", resp2)

	}

	////终止能力[备用]
	//err = am.TerminateAbility()
	//if err != nil {
	//	logs.Errorf("error is %v", err)
	//}
}

func TestPublishTaskState(t *testing.T) {
	logs.Init("test")
	managerUrl := "http://192.168.8.197:8080"
	abilityName := "ActInferenceAbility"
	//url := "http://192.168.8.197" // 业务url
	am := manager.NewAbilityManager(managerUrl, abilityName)
	//heartBeat, err := am.StartupAbility()
	//if err != nil {
	//	logs.Errorf("start up %s fail", abilityName)
	//} else {
	//	logs.Infof("heartBeat is %v", heartBeat)
	//
	//	// 通过heartbeat中的abilityPort拼接成新的url
	//
	//	serviceUrl := fmt.Sprintf("%s:%s%s", url, strconv.Itoa(heartBeat.AbilityPort), "/api/task_state")
	//	resp, err := PublishTaskState(serviceUrl)
	//	if err != nil {
	//		logs.Errorf("error publish LeftArmUp inst %v", err)
	//	}
	//	logs.Infof("resp is %v", resp)
	//}
	//time.Sleep(15 * time.Second)

	//终止能力[备用]
	err := am.TerminateAbility()
	if err != nil {
		logs.Errorf("error is %v", err)
	}

}

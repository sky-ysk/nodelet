package lib

import (
	"fmt"
	"hit.edu/framework/pkg/component-base/logs"
	"hit.edu/framework/pkg/nodelet/task/runtime/device/ability/manager"
	"strconv"
	"testing"
)

func TestPublishPredictInst(t *testing.T) {
	logs.Init("test")
	managerUrl := ""
	abilityName := ""
	am := manager.NewAbilityManager(managerUrl, abilityName)
	heartBeat, err := am.StartupAbility()

	//// 终止能力[备用]
	//err =am.TerminateAbility()
	//if err!=nil{
	//	logs.Errorf("error is %v",err)
	//}
	if err != nil {
		logs.Errorf("start up %s fail", abilityName)
	} else {
		imagePath := ""

		// 通过heartbeat中的abilityPort拼接成新的url
		url := ""
		serviceUrl := fmt.Sprintf("%s:%s", url, strconv.Itoa(heartBeat.AbilityPort))
		resp, err := PublishPredictInst(imagePath, serviceUrl)
		if err != nil {
			logs.Errorf("error publish ArmAngle inst %v", err)
		} else {
			logs.Infof("predict resp is %v", resp)
		}

	}

}

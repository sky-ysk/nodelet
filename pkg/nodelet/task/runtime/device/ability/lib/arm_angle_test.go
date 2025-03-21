package lib

import (
	"fmt"
	"hit.edu/framework/pkg/component-base/logs"
	"hit.edu/framework/pkg/nodelet/task/runtime/device/ability/manager"
	"strconv"
	"testing"
)

// go test -run TestPublishArmAngleTerminate  -v
func TestPublishArmAngleTerminate(t *testing.T) {
	logs.Init("test")
	managerUrl := "http://192.168.8.165:8080" // 能力框架url
	abilityName := "ArmControl.Leju.Guochuang"

	am := manager.NewAbilityManager(managerUrl, abilityName)

	// 终止能力[备用]
	err := am.TerminateAbility()
	if err != nil {
		logs.Errorf("error is %v", err)
	}
}

// go test -run TestPublishArmAngleInst  -v
func TestPublishArmAngleInst(t *testing.T) {
	logs.Init("test")
	managerUrl := "http://192.168.8.165:8080" // 能力框架url
	abilityName := "ArmControl.Leju.Guochuang"
	url := "http://192.168.8.165" // 业务url
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
		logs.Infof("heartBeat is %v", heartBeat)
		var params map[string][]float64 = make(map[string][]float64)
		params["left"] = []float64{-1.221, 0.0872, 0, 0, 0, 0, 0}
		params["right"] = []float64{0, 0, 0, 0, 0, 0, 0}

		// 通过heartbeat中的abilityPort拼接成新的url

		serviceUrl := fmt.Sprintf("%s:%s", url, strconv.Itoa(heartBeat.AbilityPort))
		err := PublishArmAngleInst(params, serviceUrl)
		if err != nil {
			logs.Errorf("error publish ArmAngle inst %v", err)
		}
	}

}

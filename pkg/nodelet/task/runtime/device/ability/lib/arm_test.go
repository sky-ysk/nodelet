package lib

import (
	"fmt"
	"hit.edu/framework/pkg/component-base/logs"
	"hit.edu/framework/pkg/nodelet/task/runtime/device/ability/manager"
	"strconv"
	"testing"
	"time"
)

func TestPublishGrabWorkpieceInst(t *testing.T) {
	logs.Init("test")
	ip := "http://172.130.0.61:53051"
	api := "/api/task/grab"
	url := fmt.Sprintf("%s%s", ip, api)
	taskId, err := PublishGrabWorkpieceInst("belt", url)
	if err != nil {
		logs.Errorf("err is %v", err)
	}

	time.Sleep(time.Second * 10)
	tr, err := GetTaskStatus(taskId, "http://172.130.0.61:8080")
	if err != nil {
		logs.Errorf("tr is %v", tr)
	}
}
func TestPublishPutWorkpieceInst(t *testing.T) {
	logs.Init("test")
	ip := "http://172.130.0.61:55607"
	api := "/api/task/put"
	url := fmt.Sprintf("%s%s", ip, api)
	taskId, err := PublishPutWorkpieceInst("reject_bin", url)
	if err != nil {
		logs.Errorf("err is %v", err)
	}

	time.Sleep(time.Second * 10)
	tr, err := GetTaskStatus(taskId, "http://172.130.0.61:8080")
	if err != nil {
		logs.Errorf("tr is %v", tr)
	}
}
func TestPublishGrabInitInst(t *testing.T) {
	logs.Init("test")
	ip := "http://172.130.0.61:53051"
	api := "/api/task/init"
	url := fmt.Sprintf("%s%s", ip, api)
	taskId, err := PublishGrabInitInst("up", url)
	if err != nil {
		logs.Errorf("err is %v", err)
	}

	time.Sleep(time.Second * 10)
	tr, err := GetTaskStatus(taskId, "http://172.130.0.61:8080")
	if err != nil {
		logs.Errorf("tr is %v", tr)
	}
}

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
	//url := "http://192.168.8.165" // 业务url
	am := manager.NewAbilityManager(managerUrl, abilityName)
	// 终止能力[备用]
	err := am.TerminateAbility()
	if err != nil {
		logs.Errorf("error is %v", err)
	}
	//heartBeat, err := am.StartupAbility()
	//
	//if err != nil {
	//	logs.Errorf("start up %s fail", abilityName)
	//} else {
	//	logs.Infof("heartBeat is %v", heartBeat)
	//	var params map[string][]float64 = make(map[string][]float64)
	//	params["left"] = []float64{-1.221, 0.0872, 0, 0, 0, 0, 0}
	//	params["right"] = []float64{0, 0, 0, 0, 0, 0, 0}
	//
	//	// 通过heartbeat中的abilityPort拼接成新的url
	//
	//	serviceUrl := fmt.Sprintf("%s:%s%s", url, strconv.Itoa(heartBeat.AbilityPort), "/api/control/arm_angle")
	//	err := PublishArmAngleInst(params, serviceUrl)
	//	if err != nil {
	//		logs.Errorf("error publish ArmAngle inst %v", err)
	//	}
	//}

}

// go test -run TestPublishLeftArmDownInst -v
func TestPublishLeftArmDownInst(t *testing.T) {
	logs.Init("test")
	managerUrl := "http://192.168.8.165:8080" // 能力框架url
	abilityName := "ArmControl.Leju.Guochuang"
	url := "http://192.168.8.165" // 业务url
	am := manager.NewAbilityManager(managerUrl, abilityName)
	heartBeat, err := am.StartupAbility()
	if err != nil {
		logs.Errorf("start up %s fail", abilityName)
	} else {
		logs.Infof("heartBeat is %v", heartBeat)
		// 通过heartbeat中的abilityPort拼接成新的url

		serviceUrl := fmt.Sprintf("%s:%s%s", url, strconv.Itoa(heartBeat.AbilityPort), "/api/control/left_arm_down")
		logs.Infof("go into down ...")
		err := PublishLeftArmDownInst(serviceUrl)
		if err != nil {
			logs.Errorf("error publish LeftArmDown inst %v", err)
		}
	}
	// 终止能力[备用]
	time.Sleep(3 * time.Second)

	err = am.TerminateAbility()
	if err != nil {
		logs.Errorf("error is %v", err)
	}
}

// go test -run TestPublishLeftArmUpInst -v
func TestPublishLeftArmUpInst(t *testing.T) {
	logs.Init("test")
	managerUrl := "http://192.168.8.165:8080" // 能力框架url
	abilityName := "ArmControl.Leju.Guochuang"
	url := "http://192.168.8.165" // 业务url
	am := manager.NewAbilityManager(managerUrl, abilityName)
	heartBeat, err := am.StartupAbility()
	if err != nil {
		logs.Errorf("start up %s fail", abilityName)
	} else {
		logs.Infof("heartBeat is %v", heartBeat)

		// 通过heartbeat中的abilityPort拼接成新的url

		serviceUrl := fmt.Sprintf("%s:%s%s", url, strconv.Itoa(heartBeat.AbilityPort), "/api/control/left_arm_up")
		err := PublishLeftArmUpInst(serviceUrl)
		if err != nil {
			logs.Errorf("error publish LeftArmUp inst %v", err)
		}
	}
	time.Sleep(3 * time.Second)
	//终止能力[备用]
	err = am.TerminateAbility()
	if err != nil {
		logs.Errorf("error is %v", err)
	}
}

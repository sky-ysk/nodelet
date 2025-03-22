package lib

import (
	"fmt"
	"hit.edu/framework/pkg/component-base/logs"
	"hit.edu/framework/pkg/nodelet/task/runtime/device/ability/manager"
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

// go test -run TestTerminate  -v
func TestTerminate(t *testing.T) {
	logs.Init("test")
	managerUrl := "http://192.168.8.165:8080" // 能力框架url
	abilityName := "Detect"
	am := manager.NewAbilityManager(managerUrl, abilityName)

	// 终止能力[备用]
	err := am.TerminateAbility()
	if err != nil {
		logs.Errorf("error is %v", err)
	}
}

// go test -run TestPublishPredictInst  -v
func TestPublishPredictInst(t *testing.T) {
	logs.Init("test")
	managerUrl := "http://192.168.8.165:8080" // 能力框架url
	// 获取当前测试文件的目录
	testDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("无法获取当前工作目录: %v", err)
	}

	// 构建相对路径
	imagePath := filepath.Join(testDir, "testpics", "orange0.jpg")
	abilityName := "Detect"
	url := "http://192.168.8.165" // 业务的url
	am := manager.NewAbilityManager(managerUrl, abilityName)
	heartBeat, err := am.StartupAbility()

	//// 终止能力[备用]
	//err =am.TerminateAbility()
	//if err!=nil{
	//	logs.Errorf("error is %v",err)
	//}
	if err != nil {
		logs.Errorf("start up %s fail", abilityName)
		fmt.Println(err)
	} else {
		fmt.Println("heart beat is ", heartBeat)

		// 通过heartbeat中的abilityPort拼接成新的url

		serviceUrl := fmt.Sprintf("%s:%s", url, strconv.Itoa(heartBeat.AbilityPort))
		resp, err := PublishPredictInst(imagePath, serviceUrl)
		if err != nil {
			logs.Errorf("error publish predict inst %v", err)
		} else {
			fmt.Println("predict resp is ", resp)
			logs.Infof("predict resp is %v", resp)
		}
	}
}

// go test -run TestPostPic  -v

func TestPostPic(t *testing.T) {
	managerUrl := "http://192.168.8.165:8080" // 能力框架url
	// 获取当前测试文件的目录
	testDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("无法获取当前工作目录: %v", err)
	}

	// 构建相对路径
	imagePath := filepath.Join(testDir, "testpics", "test.json")
	_, err = PublishPredictInst(imagePath, managerUrl)

}

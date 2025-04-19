package manager

import (
	"fmt"
	"hit.edu/framework/pkg/component-base/logs"
	"testing"
)

func TestGetUUID(t *testing.T) {
	moduleName := "testModule"
	logs.Init(moduleName)
	logs.Infof("[test] testing run.....\n")
	var url string = "http://127.0.0.1:8123"
	var name string = "Mock"
	abilityManager := NewAbilityManager(url, name)
	uuid, err := abilityManager.GetUUID()
	if err != nil {
		t.Error(err)
	}
	fmt.Println(uuid)
}

// go test -run TestStartupAbility -v
func TestStartupAbility(t *testing.T) {
	moduleName := "testModule"
	logs.Init(moduleName)
	logs.Infof("[test] testing run.....\n")
	var url string = "http://192.168.8.165:8080" // 更改ip和端口
	var name string = "GrabBall.Leju.Guochuang"  // 更改能力名字
	//var name string = "DetectPosition.Leju.Guochuang" // 更改能力名字
	abilityManager := NewAbilityManager(url, name)
	heartBeat, err := abilityManager.StartupAbility()
	if err != nil {
		t.Error(err.Error())
	}
	logs.Infof("heartbeat is %v", heartBeat)
	logs.Infof("ability port is%v", heartBeat.AbilityPort)
	logs.Info("start mock ability successfully...")
	logs.Infof("terminate mock ability....")
	//err = abilityManager.TerminateAbility()
	//if err != nil {
	//	t.Error(err)
	//}
	logs.Info("terminate mock ability successfully")
}

// go test -run TestTerminateAbility -v
func TestTerminateAbility(t *testing.T) {
	moduleName := "testModule"
	logs.Init(moduleName)
	logs.Infof("[test] testing run.....\n")
	var url string = "http://192.168.8.165:8080"      // 更改ip和端口
	var name string = "DetectPosition.Leju.Guochuang" // 更改能力名字
	abilityManager := NewAbilityManager(url, name)
	err := abilityManager.TerminateAbility()
	if err != nil {
		t.Error(err)
	}
}

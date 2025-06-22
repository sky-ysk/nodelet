package lib

import (
	"hit.edu/framework/pkg/component-base/logs"
	"hit.edu/framework/pkg/nodelet/task/runtime/device/ability/manager"
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

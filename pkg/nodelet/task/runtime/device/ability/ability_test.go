package ability

import (
	"fmt"
	"hit.edu/framework/pkg/component-base/logs"
	"testing"
)

func TestGetUUID(t *testing.T) {
	moduleName := "testModule"
	logs.Init(moduleName)
	logs.Infof("[test] testing run.....\n")
	var url string = ""
	var name string = ""
	abilityManager := NewAbilityManager(url, name)
	uuid, err := abilityManager.GetUUID()
	if err != nil {
		t.Error(err)
	}
	fmt.Println(uuid)
}

func TestStartupAbility(t *testing.T) {
	moduleName := "testModule"
	logs.Init(moduleName)
	logs.Infof("[test] testing run.....\n")
	var url string = ""
	var name string = ""
	abilityManager := NewAbilityManager(url, name)
	err := abilityManager.StartupAbility()
	if err != nil {
		t.Error(err)
	}
}

func TestTerminateAbility(t *testing.T) {
	moduleName := "testModule"
	logs.Init(moduleName)
	logs.Infof("[test] testing run.....\n")
	var url string = ""
	var name string = ""
	abilityManager := NewAbilityManager(url, name)
	err := abilityManager.TerminateAbility()
	if err != nil {
		t.Error(err)
	}
}

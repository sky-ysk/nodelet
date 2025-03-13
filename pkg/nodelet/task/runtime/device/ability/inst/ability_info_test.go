package inst

import (
	"encoding/json"
	"fmt"
	"hit.edu/framework/pkg/component-base/logs"
	"testing"
)

func TestGetAbilityInstances(t *testing.T) {
	moduleName := "testModule"
	logs.Init(moduleName)
	logs.Infof("[test] testing run.....\n")
	var url string = "http://127.0.0.1:8123"

	var abilityInstances []AbilityInstance

	abilityInstances, err := GetAbilityInstances(url)
	if err != nil {
		fmt.Println(err)
	}
	responseByte, err := json.Marshal(abilityInstances)
	if err != nil {
		logs.Errorf("Marshal failed")
	}
	responseStr := string(responseByte)
	logs.Infof("%v", responseStr)
}

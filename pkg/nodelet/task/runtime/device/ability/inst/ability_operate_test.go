package inst

import (
	"encoding/json"
	"fmt"
	"hit.edu/framework/pkg/component-base/logs"
	"testing"
)

func TestGetHeartBeats(t *testing.T) {
	moduleName := "testModule"
	logs.Init(moduleName)
	logs.Infof("[test] testing run.....\n")
	var url string
	var heartBeats []HeartBeat
	heartBeats, err := GetAbilityHeartBeat(url)

	if err != nil {
		t.Error(err)
	}

	fmt.Println(heartBeats[2].AbilityName)
	fmt.Println(heartBeats[2].ID)
	fmt.Println(heartBeats[2].State)
}

func TestNewOperatorStr(t *testing.T) {
	moduleName := "testModule"
	logs.Init(moduleName)
	logs.Infof("[test] testing run.....\n")
	var abilityOperate = AbilityOperate{
		Command:           "start",
		AbilityInstanceId: "this is a test id",
	}
	// 序列化
	operate, err := json.Marshal(abilityOperate)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(operate))
}

func TestGetTaskState(t *testing.T) {
	moduleName := "testModule"
	logs.Init(moduleName)
	logs.Infof("[test] testing run.....\n")
	var taskState TaskState
	var url string
	var taskId string

	taskState, err := GetTaskState(taskId, url)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(taskState.State)
}

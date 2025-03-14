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
	var url string = "http://127.0.0.1:8123"
	var heartBeats []HeartBeat
	heartBeats, err := GetAbilityHeartBeat(url)

	if err != nil {
		t.Error(err)
	}

	fmt.Println(heartBeats)
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
	var url string = "http://127.0.0.1:8123"
	var taskId string = "d26a86e4-83c3-44af-97fc-d65fa033adbd"

	taskState, err := GetTaskState(taskId, url)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(taskState.State)
}

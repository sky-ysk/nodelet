package inst

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestGetHeartBeats(t *testing.T) {

	responseBody := `[
	{
	"IPCPort": 38379,
	"IPCProtocol": "http",
	"abilityName": "Robot.MonitorAbility",
	"abilityPort": 0,
	"id": "20d3ff60-8520-442f-b228-3d4ab25477d2",
	"instanceName": "Robot.MonitorAbility",
	"state": "Standby",
	"version": "0.1.0"
	},
	{
	"IPCPort": 44639,
	"IPCProtocol": "http",
	"abilityName": "Fixed.MonitorAbility",
	"abilityPort": 0,
	"id": "48780900-98d5-4332-abcf-e8996e3b599a",
	"instanceName": "Fixed.MonitorAbility",
	"state": "Standby",
	"version": "0.1.0"
	},
	{
	"IPCPort": 34355,
	"IPCProtocol": "http",
	"abilityName": "Abstract.MonitorAbility",
	"abilityPort": 0,
	"id": "9ef71d93-f575-4cdb-8e6b-ec30d3dfb72c",
	"instanceName": "abstract-monitor-1",
	"state": "Standby",
	"version": "0.1.0"
	}
	]`

	var heartBeats []HeartBeat
	err := json.Unmarshal([]byte(responseBody), &heartBeats)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(heartBeats[2].AbilityName)
	fmt.Println(heartBeats[2].ID)
	fmt.Println(heartBeats[2].State)
}

func TestNewOperatorStr(t *testing.T) {

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
	jsonStr := `{
	"name": "任务名称",
	"type": "任务类型",
	"state": "<任务执行状态>"
	}`

	var taskState TaskState
	err := json.Unmarshal([]byte(jsonStr), &taskState)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(taskState.State)
}

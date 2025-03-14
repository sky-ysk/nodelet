package inst

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"hit.edu/framework/pkg/component-base/logs"
	"io"
	"net/http"
)

type AbilityOperate struct {
	AbilityInstanceId string  `json:"abilityInstanceId"`
	Command           Command `json:"command"`
}

type Command string

const (
	Start      Command = "start"
	Connect    Command = "connect"
	Disconnect Command = "disconnect"
	Terminate  Command = "terminate"
)

type AbilityState string

const (
	Standby     AbilityState = "Standby"
	Running     AbilityState = "Running"
	Suspend     AbilityState = "Suspend"
	Terminating AbilityState = "Terminating"
	Terminated  AbilityState = "Terminated"
	Error       AbilityState = "Error"
	Inactive    AbilityState = "Inactive"
)

type HeartBeat struct {
	IPCPort      int          `json:"IPCPort"`
	IPCProtocol  string       `json:"IPCProtocol"`
	AbilityName  string       `json:"abilityName"`
	AbilityPort  int          `json:"abilityPort"`
	ID           string       `json:"id"`
	InstanceName string       `json:"instanceName"`
	State        AbilityState `json:"state"`
	Version      string       `json:"version"`
}

type LifeCycleResponse struct {
	TaskID string `json:"taskId"`
}

type TaskState struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	State string `json:"state"`
}

// PostLifeCycleRequest 向能力框架发送post请求，操作生命周期，返回taskId
func PostLifeCycleRequest(Id string, command Command, url string) (string, error) {
	// 构造playload
	var abilityOperate = AbilityOperate{
		Command:           command,
		AbilityInstanceId: Id,
	}
	// 序列化
	operate, err := json.Marshal(abilityOperate)
	if err != nil {
		logs.Error("marshal AbilityOperate error\n")
		return "", err
	}
	logs.Info("post request construct successfully\n")
	// 构造post请求
	operatorStr := string(operate)
	requestForPost := NewPostRequest(fmt.Sprintf("%s/api/lifecycle-request", url), operatorStr)
	// 发送post亲求
	client := &http.Client{}
	fmt.Println(requestForPost.Payload)
	response, err := client.Post(requestForPost.Url, "application/json", bytes.NewBuffer([]byte(requestForPost.Payload)))
	if err != nil {
		logs.Error("post request error\n")
		return "", err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		logs.Errorf("response code is %v\n", response.StatusCode)
		return "", fmt.Errorf("response code is %v\n", response.StatusCode)
	}
	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return "", err
	}
	var lifeCycleResponse LifeCycleResponse
	err = json.Unmarshal(responseBody, &lifeCycleResponse)
	if err != nil {
		logs.Error("unmarshal LifeCycleResponse error\n")
		return "", err
	}
	return lifeCycleResponse.TaskID, nil
}

// GetAbilityHeartBeat 向能力框架发送get请求，返回所有的能力状态
func GetAbilityHeartBeat(url string) ([]HeartBeat, error) {
	requestForGet := NewGetRequest(fmt.Sprintf("%s/api/ability-heartbeat", url))
	// 创建HTTP client
	logs.Info("publish get request\n")
	client := &http.Client{}
	response, err := client.Get(requestForGet.Url)
	if err != nil {
		logs.Error("get request error\n")
		return []HeartBeat{}, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		logs.Error("response code is %v\n", response.StatusCode)
		return []HeartBeat{}, fmt.Errorf("response code is %v\n", response.StatusCode)
	}
	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return []HeartBeat{}, err
	}
	var heartBeats []HeartBeat
	err = json.Unmarshal(responseBody, &heartBeats)
	if err != nil {
		logs.Error("unmarshal HeartBeat error\n")
		return []HeartBeat{}, err
	}
	return heartBeats, nil
}

// GetTaskState 获取一个能力的状态
func GetTaskState(taskId string, url string) (TaskState, error) {
	requestForGet := NewGetRequest(fmt.Sprintf("%s/api/task/%s", url, taskId))
	// 创建HTTP client
	client := &http.Client{}

	response, err := client.Get(requestForGet.Url)
	if err != nil {
		return TaskState{}, err
	}
	if response.StatusCode != http.StatusOK {
		return TaskState{}, errors.New(response.Status)
	}
	responseBody, err := io.ReadAll(response.Body)
	defer response.Body.Close()
	if err != nil {
		return TaskState{}, err
	}
	var taskState TaskState
	err = json.Unmarshal(responseBody, &taskState)
	if err != nil {
		return TaskState{}, err
	}
	return taskState, nil
}

func FindStateByUUID(hearBeats []HeartBeat, uuid string) (AbilityState, error) {
	for _, hearBeat := range hearBeats {
		if hearBeat.ID == uuid {
			return hearBeat.State, nil
		}
	}
	return "", errors.New(uuid)
}

func GetAbilityState(url string, uuid string) (AbilityState, error) {
	logs.Infof("getting heart beat......\n")
	hearBeats, err := GetAbilityHeartBeat(url)
	if err != nil {
		logs.Error("get heart beats error\n")
		return "", err
	}
	logs.Info("get heart beats successfully\n")

	logs.Info("try to find state by UUID\n")
	state, err := FindStateByUUID(hearBeats, uuid)
	if err != nil {
		logs.Error("find state by UUID error\n")
		return "", err
	}
	logs.Infof("find state by UUID successfully\n")
	return state, nil
}

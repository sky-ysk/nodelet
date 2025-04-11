package manager

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"hit.edu/framework/pkg/component-base/logs"
	"io"
	"net/http"
	"time"
)

type AdjustStatus struct {
	status interface{}
}
type AbilityInstance struct {
	Id           string       `json:"id"`
	Kind         string       `json:"kind"`
	MetaData     MetaData     `json:"metadata"`
	Owner        Owner        `json:"owner"`
	RunInfo      RunInfo      `json:"runInfo"`
	Shares       interface{}  `json:"sharers"`
	Spec         Spec         `json:"spec"`
	Status       interface{}  `json:"status"`
	Tag          Tag          `json:"tag"`
	Subabilities []SubAbility `json:"subabilities"`
}

type MetaData struct {
	Labels interface{} `json:"labels"`
	Name   string      `json:"name"`
}

type Owner struct {
	AbilityInstanceId string `json:"abilityInstanceId"`
	Position          string `json:"position"`
}

type RunInfo struct {
	LastConnect    int    `json:"lastConnect"`
	LastUpdate     int    `json:"lastUpdate"`
	LifeCycleState string `json:"lifecycleState"`
}

type Spec struct {
	AbilityName       string            `json:"abilityName"`
	ActivityCondition ActivityCondition `json:"activityCondition"`
	Config            interface{}       `json:"config"`
	Package           string            `json:"package"`
	SubAbilities      []Spec            `json:"subabilities"`
	Position          string            `json:"position"`
	Priority          int               `json:"priority"`
	Version           string            `json:"version"`
}

type ActivityCondition struct {
	JQ string `json:"jq"`
}
type Tag struct {
	Parent string `json:"parent"`
	Source string `json:"source"`
}

type Request struct {
	Url     string
	Payload string
}

type SubAbility struct {
	Id       string `json:"id"`
	Position string `json:"position"`
}

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
		// 读取响应的 Body 内容
		body, err := io.ReadAll(response.Body)
		if err != nil {
			fmt.Printf("读取响应 Body 时出错: %v\n", err)
			return "", err
		}

		// 输出响应的 Body 内容
		logs.Errorf("response code is %v\n", response.StatusCode)
		fmt.Println("response code is", string(body))

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

func FindHeartBeatByUUID(hearBeats []HeartBeat, uuid string) (HeartBeat, error) {
	for _, hearBeat := range hearBeats {
		if hearBeat.ID == uuid {
			fmt.Println("now find ", hearBeat)
			return hearBeat, nil
		}
	}
	return HeartBeat{}, errors.New(uuid)
}

// GetAbilityState 获取所有心跳包 根据uuid找到对应的心跳包
func GetAbilityState(url string, uuid string) (HeartBeat, error) {
	logs.Infof("getting heart beat......\n")
	hearBeats, err := GetAbilityHeartBeat(url)
	if err != nil {
		logs.Error("get heart beats error\n")
		return HeartBeat{}, err
	}
	logs.Info("get heart beats successfully\n")
	logs.Info("try to find state by UUID\n")
	heartBeat, err := FindHeartBeatByUUID(hearBeats, uuid)
	if err != nil {
		logs.Error("find state by UUID error\n")
		return HeartBeat{}, err
	}
	logs.Info("find state by UUID successfully  heart beat :  ", heartBeat)
	fmt.Println(heartBeat)
	return heartBeat, nil
}

func NewPostRequest(url string, payload string) *Request {
	return &Request{Url: url, Payload: payload}
}
func NewGetRequest(url string) *Request {
	return &Request{Url: url}
}

func GetAbilityInstances(url string) ([]AbilityInstance, error) {
	requestForGet := NewGetRequest(fmt.Sprintf("%s/api/cr", url))
	// 创建HTTP client
	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	logs.Info("publish get request\n")
	response, err := client.Get(requestForGet.Url)
	if err != nil {
		logs.Error("get response error: %v\n", err)
		return []AbilityInstance{}, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		logs.Error("response code is %v\n", response.StatusCode)
		return []AbilityInstance{}, errors.New(response.Status)
	}
	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return []AbilityInstance{}, err
	}
	var abilityInstances []AbilityInstance
	err = json.Unmarshal(responseBody, &abilityInstances)
	if err != nil {
		logs.Error("unmarshal response error: %v\n", err)
		return []AbilityInstance{}, err
	}
	logs.Info("unmarshal response successfully\n")
	return abilityInstances, nil
}

// FindIdByAbilityName 根据能力名字寻找uuid
func FindIdByAbilityName(abilityName string, abilityInstances []AbilityInstance) ([]string, error) {
	var idList []string
	for _, abilityInstance := range abilityInstances {
		fmt.Println("instance is", abilityInstance, abilityInstance.MetaData.Name)
		fmt.Println("hh", abilityInstance.Spec.AbilityName)

		if abilityInstance.Spec.AbilityName == abilityName {
			idList = append(idList, abilityInstance.Id)
		}

	}
	return idList, errors.New(abilityName + " is not found")
}

func GetAbilityExeStatus(uuid string, url string) (AdjustStatus, error) {
	requestForGet := NewGetRequest(fmt.Sprintf("%s/ability/%s/adjust-status", url, uuid))
	// 创建HTTP client
	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	logs.Info("publish get request\n")
	response, err := client.Get(requestForGet.Url)
	if err != nil {
		logs.Error("get response error: %v\n", err)
		return AdjustStatus{}, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		logs.Error("response code is %v\n", response.StatusCode)
		return AdjustStatus{}, errors.New(response.Status)
	}
	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return AdjustStatus{}, err
	}
	var adjustStatus AdjustStatus
	err = json.Unmarshal(responseBody, &adjustStatus)
	if err != nil {
		logs.Error("unmarshal response error: %v\n", err)
		return AdjustStatus{}, err
	}
	logs.Info("unmarshal response successfully\n")
	return adjustStatus, nil
}

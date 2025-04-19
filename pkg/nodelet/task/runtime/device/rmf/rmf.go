package rmf

import (
	"bytes"
	"encoding/json"
	"fmt"
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/component-base/logs"
	inst "hit.edu/framework/pkg/nodelet/task/runtime/device/rmf/inst"
	"io"
	"net/http"
)

/*
	1.rmf.go主要负责post发布指令、get获取任务信息
	2.发布指令
*/

// StatusResponse 解析设备状态
type StatusResponse struct {
	Status    string `json:"status"`
	Completed []int  `json:"completed"`
}
type Worker interface {
	// PublishAbilityInstruction 发布能力指令
	PublishAbilityInstruction(device apis.Device, ability string) (string, error)
	// PublishCancelTaskInstruction 发布取消任务指令
	PublishCancelTaskInstruction(device apis.Device, taskId string) (bool, error)
	// PublishCancelPhaseInstruction 发布取消Phase指令
	PublishCancelPhaseInstruction(device apis.Device, phaseId int, taskId string) (bool, error)
	// GetTaskState 获取任务状态
	GetTaskState(device apis.Device, taskId string) (*TaskStateSuccessResponse, error)
}

// PublishAbilityInstruction 发布指令，成功返回任务ID，失败返回错误信息
func PublishAbilityInstruction(device *apis.Device, ability string) (string, error) {

	// 判断参数的合法性
	if ability == "" {
		return "", fmt.Errorf("empty ability\n")
	}
	if device.Spec.AccessMethod.URL == "" {
		logs.Errorf("Error lack of url, can not new a instruction\n")
		err := fmt.Errorf("Error lack of url, can not new a instruction\n")
		return "", err
	}

	// 构造并发布指令
	instruction, err := inst.NewAbilityInstruction(*device, ability)
	if err != nil {
		logs.Errorf(err.Error())
		logs.Errorf("Ability instruction create failed\n")
		return "", err
	}
	logs.Infof("Ability instruction has been created\n")

	// 发布指令
	requestForPost := NewPostRequest(fmt.Sprintf("%s/tasks/robot_task", device.Spec.AccessMethod.URL), instruction)
	client := &http.Client{}
	fmt.Println(requestForPost.Payload)

	logs.Info("publishing ability instruction")
	response, err := client.Post(requestForPost.Url, "application/json", bytes.NewBuffer([]byte(requestForPost.Payload)))
	if err != nil {
		logs.Errorf("Error client post :%v\n", err)
		return "", err
	}
	logs.Info("Successfully publish ability instruction\n")
	defer response.Body.Close()

	// 读取response的body
	logs.Info("Start to parse response\n")
	body, err := io.ReadAll(response.Body)
	if err != nil {
		logs.Errorf("Error read response body: %v\n", err)
		return "", err
	}

	fmt.Println(string(body))

	switch response.StatusCode {

	case http.StatusOK:
		logs.Infof("The http's status code is %v\n", response.StatusCode)
		abilitySuccessResponse, err := inst.ParseAbilitySuccessResponse(body)
		if err != nil {
			logs.Errorf("Error parse response body: %v\n", err)
		}
		taskId := abilitySuccessResponse.State.Booking.ID
		logs.Infof("taskId: %v\n", taskId)
		return taskId, nil

	case http.StatusUnprocessableEntity:
		logs.Infof("The http's status code is %v\n", response.StatusCode)
		_, err := inst.ParseAbilityValidationErrorResponse(body)
		if err != nil {
			logs.Errorf("Error parse response body: %v\n", err)
		}
		//TODO: ValidationError的处理
		return string(body), err

	default:
		logs.Infof("The http's status code is %v\n", response.StatusCode)
		return string(body), fmt.Errorf("http status code is %v\n", response.StatusCode)

	}
}

// PublishCancelTaskInstruction 发布取消任务的指令，返回是否成功
func PublishCancelTaskInstruction(device apis.Device, taskId string) (bool, error) {
	// 构建取消任务的指令
	instruction, err := inst.NewCancelTaskInst(taskId)
	logs.Infof("cancel task instruction has been created\n")
	if err != nil {
		logs.Errorf("Error new cancel task manager: %v\n", err)
		return false, err
	}
	logs.Info("cancel instruction has been created\n")
	fmt.Println(instruction)
	logs.Info(instruction)
	requestForPost := NewPostRequest(fmt.Sprintf("%s/tasks/cancel_task", device.Spec.AccessMethod.URL), instruction)
	// 创建HTTP client
	client := &http.Client{}

	// 发送POST请求并错误处理
	response, err := client.Post(requestForPost.Url, "application/json", bytes.NewBuffer([]byte(requestForPost.Payload)))
	if err != nil {
		logs.Errorf("Error client post :%v\n", err)
		return false, err
	}
	defer response.Body.Close() // 确保在函数返回时关闭响应体
	logs.Info("Successfully publish cancel task instruction\n")
	logs.Info("Start to parse response\n")
	/* 第二部分：解析response */
	// 读取response的body
	body, err := io.ReadAll(response.Body)
	if err != nil {
		logs.Errorf("Error read response body: %v\n", err)
		return false, err
	}
	logs.Info("Start to parse response\n")
	switch response.StatusCode {
	case http.StatusOK:
		logs.Infof("The http's status code is %v\n", response.StatusCode)
		cancelTaskResponse, err := inst.ParseCancelTaskSuccessResponse(body)
		if err != nil {
			logs.Errorf("Error parse response body: %v\n", err)
			return false, err
		}
		logs.Infof("taskId: %v is canceled\n", taskId)
		return cancelTaskResponse.Success, err
	case http.StatusUnprocessableEntity:
		logs.Infof("The http's status code is %v\n", response.StatusCode)
		//TODO:对错误信息进行处理
		_, err := inst.ParseCancelTaskValidationErrorResponse(body)
		if err != nil {
			logs.Errorf("Error parse response body: %v\n", err)
			return false, err
		}
		logs.Infof("cancel taskId: %v is fail\n", taskId)

		return false, err
	}
	logs.Errorf("Error response code: %v\n", response.StatusCode)
	return false, nil
}

// PublishCancelPhaseInstruction 发布取消phase的指令，返回是否成功
func PublishCancelPhaseInstruction(device apis.Device, phaseId int, taskId string) (bool, error) {
	instruction, err := inst.NewCancelPhaseInst(taskId, phaseId)
	logs.Infof(" cancel phase instruction has been created\n")
	if err != nil {
		logs.Errorf("Error new cancel phase manager: %v\n", err)
		return false, err
	}
	requestForPost := NewPostRequest(fmt.Sprintf("%s/tasks/skip_phase", device.Spec.AccessMethod.URL), instruction)
	// 创建HTTP client
	client := &http.Client{}

	//fmt.Println(instruction)

	// 发送POST请求并错误处理
	response, err := client.Post(requestForPost.Url, "application/json", bytes.NewBuffer([]byte(requestForPost.Payload)))
	if err != nil {
		logs.Errorf("Error client post :%v\n", err)
		return false, err
	}
	logs.Info("Successfully publish cancel phase instruction\n")
	defer response.Body.Close() // 确保在函数返回时关闭响应体

	/* 第二部分：解析response */
	// 读取response的body
	body, err := io.ReadAll(response.Body)
	if err != nil {
		logs.Errorf("Error read response body: %v\n", err)
		return false, err
	}
	logs.Info("Start to parse response\n")
	switch response.StatusCode {
	case http.StatusOK:
		logs.Infof("The http's status code is %v\n", response.StatusCode)
		cancelPhaseResponse, err := inst.ParseCancelPhaseSuccessResponse(body)
		if err != nil {
			logs.Errorf("Error parse response body: %v\n", err)
			return false, err
		}
		logs.Infof("phaseId: %v is canceled\n", taskId)
		return cancelPhaseResponse.Success, err

	case http.StatusUnprocessableEntity:
		logs.Infof("The http's status code is %v\n", response.StatusCode)
		_, err := inst.ParseCancelPhaseValidationErrorResponse(body)
		// TODO:进行错误处理
		if err != nil {
			logs.Errorf("Error parse response body: %v\n", err)
			return false, err
		}
		logs.Infof("cancel phaseId: %v is fail\n", phaseId)

		return false, err
	}
	logs.Errorf("Error response code: %v\n", response.StatusCode)
	return false, nil
}

// GetTaskState 返回一个TaskStateSuccessResponse，失败返回的空的TaskStateSuccessResponse
func GetTaskState(device *apis.Device, taskId string) (*TaskStateSuccessResponse, error) {

	requestForGet := NewGetRequest(fmt.Sprintf("%s/tasks/%s/state", device.Spec.AccessMethod.URL, taskId))
	// 创建HTTP client
	client := &http.Client{}
	response, err := client.Get(requestForGet.Url)
	if err != nil {
		logs.Errorf("Error client get task state: %v\n", err)
		return &TaskStateSuccessResponse{}, err
	}
	logs.Info("Successfully get task state\n")
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		logs.Errorf("Error read response body: %v\n", err)
		return &TaskStateSuccessResponse{}, err
	}
	logs.Info("Start to parse response\n")
	// 根据状态码进行switch判断
	switch response.StatusCode {
	case http.StatusOK:
		logs.Infof("The http's status code is %v\n", response.StatusCode)
		var taskStateSuccessResponse TaskStateSuccessResponse
		err = json.Unmarshal(body, &taskStateSuccessResponse)
		if err != nil {
			logs.Errorf("Error parse response body: %v\n", err)
			return &TaskStateSuccessResponse{}, err
		}
		return &taskStateSuccessResponse, nil

	case http.StatusUnprocessableEntity:
		//TODO:错误处理，返回类型改为interface
		logs.Infof("The http's status code is %v\n", response.StatusCode)
		var taskValidationErrorResponse TaskStateValidationErrorResponse
		err = json.Unmarshal(body, &taskValidationErrorResponse)
		if err != nil {
			logs.Errorf("Error parse response body: %v\n", err)
			return &TaskStateSuccessResponse{}, err
		}
	default:
		logs.Infof("Error response code: %v\n", response.StatusCode)
	}
	logs.Errorf("Error response code: %v\n", response.StatusCode)
	return &TaskStateSuccessResponse{}, err
}

type TaskStateValidationErrorResponse struct {
	Detail []inst.ValidationErrorDetail `json:"detail"`
}

// 定义最外层的结构体
type TaskStateSuccessResponse struct {
	Booking                Booking                 `json:"booking"`
	Category               string                  `json:"category"`
	Detail                 []string                `json:"detail"`
	UnixMillisStartTime    int                     `json:"unix_millis_start_time"`
	UnixMillisFinishTime   int                     `json:"unix_millis_finish_time"`
	OriginalEstimateMillis int                     `json:"original_estimate_millis"`
	EstimateMillis         int                     `json:"estimate_millis"`
	AssignedTo             AssignedTo              `json:"assigned_to"`
	Status                 string                  `json:"status"`
	Dispatch               Dispatch                `json:"dispatch"`
	Phases                 map[string]Phase        `json:"phases"`
	Completed              []int                   `json:"completed"`
	Active                 int                     `json:"active"`
	Pending                []int                   `json:"pending"`
	Intruptions            map[string]Interruption `json:"interruptions"`
	Cancellation           Cancellation            `json:"cancellation"`
	Killed                 Cancellation            `json:"killed"`
}

// 定义Booking结构体
type Booking struct {
	ID                          string          `json:"id"`
	UnixMillisEarliestStartTime int             `json:"unix_millis_earliest_start_time"`
	UnixMillisRequestTime       int             `json:"unix_millis_request_time"`
	Priority                    json.RawMessage `json:"priority"` // 使用json.RawMessage处理未知结构
	Labels                      []string        `json:"labels"`
	Requester                   string          `json:"requester"`
}

// 定义AssignedTo结构体
type AssignedTo struct {
	Group string `json:"group"`
	Name  string `json:"name"`
}

// 定义Dispatch结构体
type Dispatch struct {
	Status     string     `json:"status"`
	Assignment Assignment `json:"assignment"`
	Errors     []Error    `json:"errors"`
}

// 定义Assignment结构体
type Assignment struct {
	FleetName         string `json:"fleet_name"`
	ExpectedRobotName string `json:"expected_robot_name"`
}

// 定义Error结构体
type Error struct {
	Code     int    `json:"code"`
	Category string `json:"category"`
	Detail   string `json:"detail"`
}

// 定义Phase结构体
type Phase struct {
	ID                     int                    `json:"id"`
	Category               string                 `json:"category"`
	Detail                 []string               `json:"detail"`
	UnixMillisStartTime    int                    `json:"unix_millis_start_time"`
	UnixMillisFinishTime   int                    `json:"unix_millis_finish_time"`
	OriginalEstimateMillis int                    `json:"original_estimate_millis"`
	EstimateMillis         int                    `json:"estimate_millis"`
	FinalEventID           int                    `json:"final_event_id"`
	Events                 map[string]Event       `json:"events"`
	SkipRequests           map[string]SkipRequest `json:"skip_requests"`
}

// 定义Event结构体
type Event struct {
	ID     int      `json:"id"`
	Status string   `json:"status"`
	Name   string   `json:"name"`
	Detail []string `json:"detail"`
	Deps   []int    `json:"deps"`
}

// 定义SkipRequest结构体
type SkipRequest struct {
	UnixMillisRequestTime int                `json:"unix_millis_request_time"`
	Labels                []string           `json:"labels"`
	Undo                  SkipRequestDetails `json:"undo"`
}

// 定义SkipRequestDetails结构体
type SkipRequestDetails struct {
	UnixMillisRequestTime int      `json:"unix_millis_request_time"`
	Labels                []string `json:"labels"`
}

// 定义Interruption结构体
type Interruption struct {
	UnixMillisRequestTime int       `json:"unix_millis_request_time"`
	Labels                []string  `json:"labels"`
	ResumedBy             ResumedBy `json:"resumed_by"`
}

// 定义ResumedBy结构体
type ResumedBy struct {
	UnixMillisRequestTime int      `json:"unix_millis_request_time"`
	Labels                []string `json:"labels"`
}

// 定义Cancellation结构体
type Cancellation struct {
	UnixMillisRequestTime int      `json:"unix_millis_request_time"`
	Labels                []string `json:"labels"`
}
type Request struct {
	Url     string
	Payload string
}

func NewPostRequest(url string, payload string) *Request {
	return &Request{Url: url, Payload: payload}
}
func NewGetRequest(url string) *Request {
	return &Request{Url: url}
}

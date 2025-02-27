package inst

import (
	"encoding/json"
	"fmt"
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/component-base/logs"
	"strconv"
	"strings"
)

// 常用接口适配，这里只放入已经确定的接口
//
//	目前已经适配的接口包括
//	   移动指令
//	   抓取指令
/*
	inst主要进行常用指令的构建
	1.目前分为两类：cancel类和动作类
	2.其中动作类，例如grab需要grab.go和inst.go组合而成
	3.而cancel类不需要依赖inst.go就可以构建，但是为了统一性，依然在inst.go中进行一次封装
*/

type InstructionConstructor interface {
	// GetCancelPhaseInst 构建CancelPhase指令
	GetCancelPhaseInst(taskId string, phaseId int) (string, error)

	// GetCancelTaskInst 构建CancelTask指令
	GetCancelTaskInst(taskId string) (string, error)

	// GetMoveInst 构建Move指令
	GetMoveInst(device apis.Device, ability string) (string, error)

	// GetArmInst 构建Grab指令
	GetArmInst(device apis.Device, ability string) (string, error)

	// NewAbilityInstruction 根据ability构建动作指令，相当于对GetMoveInst和GetGrabInst的再一次封装
	NewAbilityInstruction(device apis.Device, ability string) (string, error)

	// ParseAbilityValidationErrorResponse 解析ValidationError信息
	ParseAbilityValidationErrorResponse(body []byte) (*AbilityValidationErrorResponse, error)

	// ParseAbilitySuccessResponse 解析Success信息
	ParseAbilitySuccessResponse(body []byte) (*AbilitySuccessResponse, error)
}

type Instruction struct {
	Type    string  `json:"type"`
	Fleet   string  `json:"fleet"`
	Robot   string  `json:"robot"`
	Request Request `json:"request"`
}

// NewAbilityInstruction 根据传入参数ability，构建动作指令
func NewAbilityInstruction(device apis.Device, ability string) (string, error) {
	if device.Spec.AccessMethod.URL == "" {
		return "", fmt.Errorf("device %s does not have accessmethod", device.Name)
	}
	// 检查是否为RMF设备
	if device.Spec.AccessMethod.Type != apis.AccessByRmf {
		return "", fmt.Errorf("device %s does not support access by rmf", device.Name)
	}
	// 检查设备能力是否支持
	isSupported := false
	for _, a := range device.Spec.Desc.Label {
		if a == ability {
			isSupported = true
			break
		}
	}

	if !isSupported {
		return "", fmt.Errorf("device %s does not support ability %s", device.Name, ability)
	}

	// 支持的能力
	switch ability {
	case "Move":
		return GetMoveInst(device, ability)
	case "Arm":
		return GetArmInst(device, ability)
	case "Lift":
		return GetLiftInst(device, ability)
	}
	return "", fmt.Errorf("framework %s does not support ability %s", device.Name, ability)
}
func GetLiftInst(device apis.Device, ability string) (string, error) {
	robotName := device.Spec.AccessMethod.Alias
	robotId := device.Spec.Name
	group := device.Spec.AccessMethod.Group
	ep := device.Spec.ExpectedProperties
	heightp, ok := ep["height"]
	if !ok {
		return "", fmt.Errorf("height not found in expected properties")
	}
	height := heightp.Value
	return NewLiftInst(robotName, group, height, robotId)
}
func GetMoveInst(device apis.Device, ability string) (string, error) {
	// 构造参数
	robot := device.Spec.AccessMethod.Alias
	group := device.Spec.AccessMethod.Group
	// 获取期望属性
	ep := device.Spec.ExpectedProperties

	destp, ok := ep["dest"]
	if !ok {
		return "", fmt.Errorf("device %s does not have dest property", device.Name)
	}
	dest := destp.Value

	orientationp, ok := ep["orientation"]
	if !ok {
		return "", fmt.Errorf("device %s does not have orientation property", device.Name)
	}
	orientation, err := strconv.ParseFloat(orientationp.Value, 32)
	if err != nil {
		return "", fmt.Errorf("device %s orientation property is not valid", device.Name)
	}

	dockp, ok := ep["dock"]
	dock := false
	// 允许没有dock这个字段
	if ok && dockp.Value == "true" {
		dock = true
	}

	waitp, ok := ep["wait"]
	var wait int
	if !ok {
		// 没有wait这个字段，就默认wait为100
		wait = 100
	} else {
		wait, err = strconv.Atoi(waitp.Value)
		if err != nil {
			return "", fmt.Errorf("device %s wait property is not valid", device.Name)
		}
	}
	return NewMoveInst(robot, group, dest, orientation, dock, wait)
}
func GetArmInst(device apis.Device, ability string) (string, error) {
	robotId := device.Spec.Name
	robotName := device.Spec.AccessMethod.Alias
	group := device.Spec.AccessMethod.Group
	ep := device.Spec.ExpectedProperties

	// 获取destination
	destp, ok := ep["dest"]
	if !ok {
		return "", fmt.Errorf("device %s does not have dest property", device.Name)
	}
	dest := destp.Value
	// 获取x、y、z、rx、ry、rz、speed 7个属性，重复度较高，集成为一个函数统一操作，返回一个map
	armParam, err := ObtainArmParam(ep)
	if err != nil {
		return "", fmt.Errorf("ObtainArmParam is failed err:%s", err.Error())
	}
	return NewArmInst(robotName, group, dest, armParam, robotId, strings.ToLower(ability))
}

// ObtainArmParam 用于从期望的属性中解析出构造json的具体参数，返回一个map
func ObtainArmParam(ep map[string]apis.Property) (map[string]float64, error) {
	result := make(map[string]float64)
	paramList := []string{"x", "y", "z", "rx", "ry", "rz", "speed"}
	for _, param := range paramList {
		property, ok := ep[param]
		if !ok {
			return nil, fmt.Errorf("does not have %s property", param)
		}
		value, err := strconv.ParseFloat(property.Value, 32)
		if err != nil {
			return nil, fmt.Errorf("%s property value is not valid", param)
		}
		result[param] = value
	}
	return result, nil
}

func GetCancelPhaseInst(taskId string, phaseId int) (string, error) {
	return NewCancelPhaseInst(taskId, phaseId)
}

func GetCancelTaskInst(taskId string) (string, error) {
	return NewCancelTaskInst(taskId)
}

func ParseAbilitySuccessResponse(body []byte) (*AbilitySuccessResponse, error) {
	var response AbilitySuccessResponse
	err := json.Unmarshal(body, &response)
	if err != nil {
		logs.Errorf("parse ability success response failed: %v\n", err)
		return nil, fmt.Errorf("unmarshal response failed: %v", err)
	}
	return &response, nil
}

func ParseAbilityValidationErrorResponse(body []byte) (*AbilityValidationErrorResponse, error) {
	var response AbilityValidationErrorResponse
	err := json.Unmarshal(body, &response)
	if err != nil {
		logs.Errorf("parse ability success response failed: %v\n", err)
		return nil, fmt.Errorf("unmarshal response failed: %v", err)
	}
	return &response, nil
}

type AbilityValidationErrorResponse struct {
	Detail []ValidationErrorDetail `json:"detail"`
}

// 定义结构体以匹配JSON中的字段
type AbilitySuccessBooking struct {
	ID                          string `json:"id"`
	UnixMillisEarliestStartTime int64  `json:"unix_millis_earliest_start_time"`
	//UnixMillisRequestTime       int64           `json:"unix_millis_request_time"`
	Priority json.RawMessage `json:"priority"`
	Labels   []string        `json:"labels"`
	//Requester                   string          `json:"requester"`
}

type Assignment struct {
	FleetName         string `json:"fleet_name"`
	ExpectedRobotName string `json:"expected_robot_name"`
}

type Error struct {
	Code     int    `json:"code"`
	Category string `json:"category"`
	Detail   string `json:"detail"`
}

type Dispatch struct {
	Status     string     `json:"status"`
	Assignment Assignment `json:"assignment"`
	Errors     []Error    `json:"errors"`
}

type Event struct {
	ID     int           `json:"id"`
	Status string        `json:"status"`
	Name   string        `json:"name"`
	Detail []interface{} `json:"detail"`
	Deps   []int         `json:"deps"`
}

type SkipRequest struct {
	UnixMillisRequestTime int64    `json:"unix_millis_request_time"`
	Labels                []string `json:"labels"`
	Undo                  struct {
		UnixMillisRequestTime int64    `json:"unix_millis_request_time"`
		Labels                []string `json:"labels"`
	} `json:"undo"`
}

type AbilityPhase struct {
	ID                     int64                  `json:"id"`
	Category               string                 `json:"category"`
	Detail                 []interface{}          `json:"detail"`
	UnixMillisStartTime    int64                  `json:"unix_millis_start_time"`
	UnixMillisFinishTime   int64                  `json:"unix_millis_finish_time"`
	OriginalEstimateMillis int64                  `json:"original_estimate_millis"`
	EstimateMillis         int64                  `json:"estimate_millis"`
	FinalEventID           int64                  `json:"final_event_id"`
	Events                 map[string]Event       `json:"events"`
	SkipRequests           map[string]SkipRequest `json:"skip_requests"`
}

type Interruption struct {
	UnixMillisRequestTime int64    `json:"unix_millis_request_time"`
	Labels                []string `json:"labels"`
	ResumedBy             struct {
		UnixMillisRequestTime int64    `json:"unix_millis_request_time"`
		Labels                []string `json:"labels"`
	} `json:"resumed_by"`
}

type Cancellation struct {
	UnixMillisRequestTime int64    `json:"unix_millis_request_time"`
	Labels                []string `json:"labels"`
}

type Killed struct {
	UnixMillisRequestTime int64    `json:"unix_millis_request_time"`
	Labels                []string `json:"labels"`
}

type AbilitySuccessState struct {
	Booking                AbilitySuccessBooking `json:"booking"`
	Category               string                `json:"category"`
	Detail                 interface{}           `json:"detail"`
	UnixMillisStartTime    int64                 `json:"unix_millis_start_time"`
	UnixMillisFinishTime   int64                 `json:"unix_millis_finish_time"`
	OriginalEstimateMillis int64                 `json:"original_estimate_millis"`
	EstimateMillis         int64                 `json:"estimate_millis"`
	AssignedTo             struct {
		Group string `json:"group"`
		Name  string `json:"name"`
	} `json:"assigned_to"`
	Status        string                  `json:"status"`
	Dispatch      Dispatch                `json:"dispatch"`
	Phases        map[string]AbilityPhase `json:"phases"`
	Completed     []int                   `json:"completed"`
	Active        int64                   `json:"active"`
	Pending       []int                   `json:"pending"`
	Interruptions map[string]Interruption `json:"interruptions"`
	Cancellation  Cancellation            `json:"cancellation"`
	Killed        Killed                  `json:"killed"`
}

// Response is the top-level struct that contains the "success" field and the "state" struct
type AbilitySuccessResponse struct {
	Success bool                `json:"success"`
	State   AbilitySuccessState `json:"state"`
}

type Request struct {
	Category                    string      `json:"category"`
	Description                 RequestDesc `json:"description"`
	UnixMillisEarliestStartTime int         `json:"unix_millis_earliest_start_time"`
	Priority                    Priority    `json:"priority"`
}

type RequestDesc struct {
	Category string  `json:"category"`
	Phases   []Phase `json:"phases"`
}

type Phase struct {
	Activity Activity `json:"activity"`
}

type Activity struct {
	Category    string       `json:"category"`
	Description ActivityDesc `json:"description"`
}

type ActivityDesc struct {
	// 存放实际的指令
	// 非结构化数据，下层构造完成后，集成到Activities
	Activities []AcitivityInterface `json:"activities"`
}

type Priority struct {
	Type  string `json:"type"`
	Value int    `json:"value"`
}

type AcitivityInterface interface {
}

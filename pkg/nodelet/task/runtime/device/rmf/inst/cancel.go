package inst

import (
	"encoding/json"
	"fmt"
	"hit.edu/framework/pkg/component-base/logs"
)

/*
	cancel主要有两部分：cancelTask 和 cancelPhase
	1.这两部分有各自的Request和Response
	2.Request主要用来构建rmf指令
	3.Response主要用来解析http的返回信息，判断rmf指令是否成功
	4.根据http的状态码来判断选用哪个结构体进行解析
*/

type CancelTaskWorker interface {
	// NewCancelTaskInst 创建CancelTask指令
	NewCancelTaskInst(taskId string) (string, error)
	// ParseCancelTaskSuccessResponse 解析CancelTask成功信息
	ParseCancelTaskSuccessResponse(body []byte) (*CancelTaskSuccessResponse, error)
	// ParseCancelTaskValidationErrorResponse 解析CancelTask失败信息
	ParseCancelTaskValidationErrorResponse(body []byte) (*CancelTaskValidationErrorResponse, error)
}

type CancelPhaseWorker interface {
	// NewCancelPhaseInst 创建CancelPhase指令
	NewCancelPhaseInst(taskId string) (string, error)
	// ParseCancelPhaseSuccessResponse 解析CancelPhase成功信息
	ParseCancelPhaseSuccessResponse(body []byte) (*CancelPhaseSuccessResponse, error)
	// ParseCancelPhaseValidationErrorResponse 解析CancelPhase失败信息
	ParseCancelPhaseValidationErrorResponse(body []byte) (*CancelPhaseValidationErrorResponse, error)
}

// CancelTaskRequest 用来构建取消任务的指令
type CancelTaskRequest struct {
	Type   string   `json:"type"`
	TaskId string   `json:"task_id"`
	Label  []string `json:"labels"`
}

// CancelTaskSuccessResponse 用来解析成功的信息
type CancelTaskSuccessResponse struct {
	Success bool `json:"success"`
}

// CancelTaskValidationErrorResponse 用来解析ValidationError的信息
type CancelTaskValidationErrorResponse struct {
	Detail []ValidationErrorDetail `json:"detail"`
}

type ValidationErrorDetail struct {
	Loc  []interface{} `json:"loc"`
	Msg  string        `json:"msg"`
	Type string        `json:"type"`
}

// CancelPhaseRequest 用来构建取消等待的指令
type CancelPhaseRequest struct {
	Type    string   `json:"type"`
	TaskId  string   `json:"task_id"`
	PhaseId int      `json:"phase_id"`
	Labels  []string `json:"labels"`
}

// CancelPhaseSuccessResponse 用来解析成功消息
type CancelPhaseSuccessResponse struct {
	Success bool   `json:"success"`
	Token   string `json:"token"`
}

// CancelPhaseValidationErrorResponse 解析ValidationError的信息
type CancelPhaseValidationErrorResponse struct {
	Detail []ValidationErrorDetail `json:"detail"`
}

// NewCancelTaskInst 创建取消任务指令
func NewCancelTaskInst(taskId string) (string, error) {
	cancelTask := CancelTaskRequest{
		TaskId: taskId,
		Type:   "cancel_task_request",
		Label:  []string{"string"},
	}
	cancelTaskStr, err := json.Marshal(cancelTask)
	if err != nil {
		return "", fmt.Errorf("marshal cancel task manager failed: %v", err)
	}
	return string(cancelTaskStr), nil
}

// NewCancelPhaseInst 构建取消Phase的指令
func NewCancelPhaseInst(taskId string, phaseId int) (string, error) {
	cancelWait := CancelPhaseRequest{
		Type:    "skip_phase_request",
		TaskId:  taskId,
		PhaseId: phaseId,
		Labels:  []string{"string"},
	}
	cancelPhaseStr, err := json.Marshal(cancelWait)
	if err != nil {
		return "", fmt.Errorf("marshal go_to_place failed: %v", err)
	}
	return string(cancelPhaseStr), nil
}

// ParseCancelTaskSuccessResponse 解析cancelTask成功的response
func ParseCancelTaskSuccessResponse(body []byte) (*CancelTaskSuccessResponse, error) {
	var cancelTaskSuccessResponse CancelTaskSuccessResponse
	err := json.Unmarshal(body, &cancelTaskSuccessResponse)
	if err != nil {
		logs.Errorf("Error unmarshal CancelTaskSuccessResponse body: %v\n", err)
		return nil, err
	}
	return &cancelTaskSuccessResponse, nil
}

// ParseCancelPhaseSuccessResponse 解析cancelPhase成功的response
func ParseCancelPhaseSuccessResponse(body []byte) (*CancelPhaseSuccessResponse, error) {
	var cancelPhaseSuccessResponse CancelPhaseSuccessResponse
	err := json.Unmarshal(body, &cancelPhaseSuccessResponse)
	if err != nil {
		logs.Errorf("Error unmarshal CancelPhaseSuccessResponse body: %v\n", err)
		return nil, err
	}
	return &cancelPhaseSuccessResponse, nil
}

// ParseCancelTaskValidationErrorResponse 解析cancelTask validation error的response
func ParseCancelTaskValidationErrorResponse(body []byte) (*CancelTaskValidationErrorResponse, error) {
	var cancelTaskValidationErrorResponse CancelTaskValidationErrorResponse
	err := json.Unmarshal(body, &cancelTaskValidationErrorResponse)
	if err != nil {
		logs.Errorf("Error unmarshal CancelTaskValidationErrorResponse body: %v\n", err)
		return nil, err
	}
	return &cancelTaskValidationErrorResponse, nil
}

// ParseCancelPhaseValidationErrorResponse 解析cancelPhase validationError的response
func ParseCancelPhaseValidationErrorResponse(body []byte) (*CancelPhaseValidationErrorResponse, error) {
	var cancelPhaseValidationErrorResponse CancelPhaseValidationErrorResponse
	err := json.Unmarshal(body, &cancelPhaseValidationErrorResponse)
	if err != nil {
		logs.Errorf("Error unmarshal CancelPhaseValidationErrorResponse body: %v\n", err)
	}
	return &cancelPhaseValidationErrorResponse, nil
}

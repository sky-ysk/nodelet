package inst

import (
	"hit.edu/framework/pkg/component-base/logs"
	"testing"
)

func TestNewCancelTaskInst(t *testing.T) {
	// TODO:获取taskId
	taskId := ""
	cancelTaskInstStr, err := NewCancelTaskInst(taskId)
	if err != nil {
		t.Error(err)
	}
	logs.Infof("cancelTaskInstStr is :%s", cancelTaskInstStr)
}

func TestNewCancelPhaseInst(t *testing.T) {
	// TODO:获取taskId和phaseId
	taskId := ""
	phaseId := 2
	CancelWaitInstStr, err := NewCancelPhaseInst(taskId, phaseId)
	if err != nil {
		t.Error(err)
	}
	logs.Infof("cancelTaskInstStr is :\n%s", CancelWaitInstStr)
}

func TestParseCancelTaskSuccessResponse(t *testing.T) {
	jsonData := `{"success": true}`

	cancelTaskSuccessResponse, err := ParseCancelTaskSuccessResponse([]byte(jsonData))
	if err != nil {
		t.Error(err)
		logs.Errorf("parse cancel task success response error: %s", err.Error())
	}
	logs.Infof("parse cancel task success response: %v", cancelTaskSuccessResponse.Success)
}

func TestParseCancelTaskValidationErrorResponse(t *testing.T) {

	jsonData := `{
  	"detail": [
	{	
	"loc": ["string",0],
	"msg": "string",
	"type": "string"
    }]
	}`
	cancelTaskValidationErrorResponse, err := ParseCancelTaskValidationErrorResponse([]byte(jsonData))
	if err != nil {
		t.Error(err)
		logs.Errorf("parse cancel task success response error: %s", err.Error())
	}
	logs.Infof("detail.msg is :%s", cancelTaskValidationErrorResponse.Detail[0].Msg)
}

func TestParseCancelPhaseSuccessResponse(t *testing.T) {}

func TestParseCancelPhaseValidationErrorResponse(t *testing.T) {
	jsonData := `{
  	"detail": [
	{	
	"loc": ["string",0],
	"msg": "string",
	"type": "string"
    }]
	}`
	cancelPhaseValidationErrorResponse, err := ParseCancelPhaseValidationErrorResponse([]byte(jsonData))
	if err != nil {
		t.Error(err)
		logs.Errorf("parse cancel phase validation error response error: %s", err.Error())
	}

	logs.Infof("detail.msg is:%s", cancelPhaseValidationErrorResponse.Detail[0].Msg)
}

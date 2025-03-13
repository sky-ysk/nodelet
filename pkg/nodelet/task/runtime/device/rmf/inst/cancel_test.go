package inst

import (
	"fmt"
	"hit.edu/framework/pkg/component-base/logs"
	"testing"
)

func TestNewCancelTaskInst(t *testing.T) {
	// TODO:获取taskId
	taskId := "1a828a92-44da-4865-97ae-634eeb0ebae3"
	cancelTaskInstStr, err := NewCancelTaskInst(taskId)
	fmt.Println(cancelTaskInstStr)
	logs.Infof("cancelTaskInstStr is :%s", cancelTaskInstStr)
	if err != nil {
		t.Error(err)
	}
	logs.Infof("cancelTaskInstStr is :%s", cancelTaskInstStr)
}

func TestNewCancelPhaseInst(t *testing.T) {
	// TODO:获取taskId和phaseId
	taskId := "1a828a92-44da-4865-97ae-634eeb0ebae3"
	phaseId := 2
	CancelWaitInstStr, err := NewCancelPhaseInst(taskId, phaseId)
	fmt.Println(CancelWaitInstStr)
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
	fmt.Println("parse cancel task success response: ", cancelTaskSuccessResponse.Success)
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
	fmt.Println("loc is ", cancelTaskValidationErrorResponse.Detail[0].Loc[0])
	fmt.Println("detail.msg is ", cancelTaskValidationErrorResponse.Detail[0].Msg)
}

func TestParseCancelPhaseSuccessResponse(t *testing.T) {
	jsonData := `{
  	"success": true,
	"token": "string"
	}`
	cancelPhaseSuccessResponse, err := ParseCancelPhaseSuccessResponse([]byte(jsonData))
	if err != nil {
		t.Error(err)
	}
	fmt.Println("success is ", cancelPhaseSuccessResponse.Success)
	fmt.Println("token is ", cancelPhaseSuccessResponse.Token)
}

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
	fmt.Println("loc is ", cancelPhaseValidationErrorResponse.Detail[0].Loc[0])
	fmt.Println("detail.msg is ", cancelPhaseValidationErrorResponse.Detail[0].Msg)

}

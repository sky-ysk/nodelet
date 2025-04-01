package transport

type ScoreRequest struct {
	TaskID     string `json:"task_id"`
	PipelineId string `json:"pipeline_id"`
	SecID      string `json:"sec_id"`
	GroupID    string `json:"group_id"`
	//Group      *apis.Group `json:"group"`
	NodeID string `json:"nodeId"`
	Host   string `json:"host"`
}

type ScoreResponse struct {
	BaseResp BaseResponse
	Data     ScoreRespData `json:"data"`
}

type SendGroupsResponse struct {
	BaseResponse
	TaskID string `json:"task_id"`
}

type BaseResponse struct {
	Success bool
	Code    string `json:"code"`
	Reason  string `json:"reason"`
}

func NewFailSendScoreResponse(taskID string, err error) SendGroupsResponse {
	return SendGroupsResponse{
		BaseResponse: NewFailBaseResponse(err),
		TaskID:       taskID,
	}
}

func NewFailBaseResponse(err error) BaseResponse {
	return NewFailBaseResponseWithCode("400", err)
}

func NewFailBaseResponseWithCode(code string, err error) BaseResponse {
	return BaseResponse{
		Success: false,
		Code:    code,
		Reason:  err.Error(),
	}
}

func NewSuccessBaseResponseWithCode() BaseResponse {
	return BaseResponse{
		Success: true,
		Code:    "200",
	}
}

type ScoreRespData struct {
	PipelineId string `json:"pipelineId"`
	SecID      string `json:"secId"`
	GroupID    string `json:"groupID"`
	NodeID     string `json:"nodeId"`
	Score      int64  `json:"score"` //nodeID - score
}

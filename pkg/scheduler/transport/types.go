package transport

import (
	apis "hit.edu/framework/pkg/apis/cores"
)

type ScoreRequest struct {
	PipelineId string      `json:"pipelineId"`
	SecID      string      `json:"secId"`
	Group      *apis.Group `json:"group"`
	NodeID     string      `json:"nodeId"`
	Host       string      `json:"host"`
}

type ScoreResponse struct {
	BaseResp BaseResponse
	Data     ScoreRespData `json:"data"`
}

type BaseResponse struct {
	Success bool
	Code    string `json:"code"`
	Reason  string `json:"reason"`
}

type ScoreRespData struct {
	PipelineId string `json:"pipelineId"`
	SecID      string `json:"secId"`
	GroupID    string `json:"group"`
	Score      int64  `json:"score"` //nodeID - score
}

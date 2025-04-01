package plugins

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/component-base/logs"
	"hit.edu/framework/pkg/scheduler/framework"
	"hit.edu/framework/pkg/scheduler/transport"
	"io"
	"net/http"
	"time"
)

type ScorePluginDBY struct {
	pluginClient ScorePluginClient
}

type ScorePluginClient struct {
	client *http.Client
}

func (client *ScorePluginClient) SendData(data []byte, path string) ([]byte, error) {
	// 自动处理 Content-Length 和 Body 封装
	httpReq, err := http.NewRequest("POST", "http://172.150.0.11:5000"+path, bytes.NewBuffer(data))
	if err != nil {
		panic(err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpRes, err := client.client.Do(httpReq)
	if err != nil {
		fmt.Println(err)
		logs.Fatal(err)
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			logs.Fatal(err)
		}
	}(httpRes.Body)
	//TODO 后续再确定下返回的细节
	res, err := io.ReadAll(httpRes.Body)
	if err != nil {
		logs.Fatal(err)
		return nil, err
	}
	return res, nil
}

func (client *ScorePluginClient) SendGroups(request *SendGroupsRequest) transport.SendGroupsResponse {
	jsonData, err := json.Marshal(request)
	if err != nil {
		logs.Fatal(err)
		return transport.NewFailSendScoreResponse(request.TaskId, err)
	}
	data, err := client.SendData(jsonData, "/schedule/postGroup")
	if err != nil {
		return transport.NewFailSendScoreResponse(request.TaskId, err)
	}
	fmt.Println("raw resp is like")
	fmt.Println(string(data))
	//TODO 确认下返回细节
	var resp SendGroupsResponse
	err = json.Unmarshal(data, &resp)
	if err != nil {
		logs.Fatal(err)
		return transport.NewFailSendScoreResponse(request.TaskId, err)
	}
	if resp.State != 200 {
		logs.Error("send groups to DTS plugin fail", resp.Reason)
	}
	return transport.SendGroupsResponse{
		TaskID: request.TaskId,
		BaseResponse: transport.BaseResponse{
			Code:    "200",
			Success: true,
		},
	}
}

func NewScorePluginClient() ScorePluginClient {
	return ScorePluginClient{
		&http.Client{
			Timeout: time.Second * 1200,
			//走http1
			Transport: &http.Transport{
				//AllowHTTP: true, // 允许非加密的HTTP/2连接（测试环境可用，生产环境建议使用TLS加密）
			},
		},
	}
}

func (sp *ScorePluginDBY) Name() string {
	return "ScorePluginForDuBoyu"
}

func (sp *ScorePluginDBY) Score(ctx context.Context, group *apis.Group, nodeName string) (int64, *framework.Status) {
	//TODO 没测过
	logs.Infof("use DTS plugin to generate a score on %s", nodeName)
	request := transport.ScoreRequest{
		GroupID: group.Status.GroupID,
		TaskID:  group.Status.Belongs.TaskID,
		NodeID:  nodeName,
	}
	jsonData, err := json.Marshal(request)
	if err != nil {
		logs.Fatal(err)
		return 0, framework.NewStatus(framework.Error, err.Error())
	}
	fmt.Println("request is")
	fmt.Println(string(jsonData))
	data, err := sp.pluginClient.SendData(jsonData, "/schedule/getSchedule")
	if err != nil {
		return 0, framework.NewStatus(framework.Error, err.Error())
	}
	fmt.Println(data)
	var resp transport.ScoreRespData
	err = json.Unmarshal(data, &resp)
	if err != nil {
		logs.Error(err)
		return 0, framework.NewStatus(framework.Error, err.Error())
	}
	return resp.Score, framework.NewStatus(framework.Success)
}

func (sp *ScorePluginDBY) SendGroups(ctx context.Context, task *apis.Task) (bool, *framework.Status) {
	request := buildSendGroupsRequest(ctx, task)
	resp := sp.pluginClient.SendGroups(request)
	if resp.BaseResponse.Success {
		return true, framework.NewStatus(framework.Success, "default success")
	}
	return false, framework.NewStatus(framework.Error, resp.BaseResponse.Reason)
}

func NewScorePluginDBY(ctx context.Context, f framework.Handle) (framework.Plugin, error) {
	return &ScorePluginDBY{
		pluginClient: NewScorePluginClient(),
	}, nil
}

type GroupTopInfo struct {
	Parent string `json:"parent"`
	Child  string `json:"child"`
}

type SendGroupsRequest struct {
	Groups              []string                              `json:"groups"`
	TaskId              string                                `json:"taskId"`
	ResourceRequirement map[string][]apis.ResourceRequirement `json:"resourceRequirement"`
	TopInfo             []GroupTopInfo                        `json:"topInfo"`
}

type SendGroupsResponse struct {
	Reason string `json:"Reason"`
	State  int64  `json:"state"`
}

func BuildSendGroupsRequest(ctx context.Context, task *apis.Task) *SendGroupsRequest {
	return buildSendGroupsRequest(ctx, task)
}

func buildSendGroupsRequest(ctx context.Context, task *apis.Task) *SendGroupsRequest {
	topInfo := make([]GroupTopInfo, 0)
	groupsID := make([]string, 0)
	resourcesMap := make(map[string][]apis.ResourceRequirement)
	taskID := task.Status.TaskID
	//TODO 可能需要做深复制 @lbh
	for _, group := range task.Spec.Groups {
		groupsID = append(groupsID, group.Status.GroupID)
		resources := make([]apis.ResourceRequirement, 0)
		for _, requirement := range group.Spec.ResourceRequirements {
			resources = append(resources, requirement)
		}
		if len(resources) > 0 {
			resourcesMap[group.Status.GroupID] = resources
		}
		for _, parent := range group.Spec.Parents {
			//fmt.Println("parent : ", parent, " child ", group.Status.GroupID)
			topInfo = append(topInfo, GroupTopInfo{
				Child:  group.Status.GroupID,
				Parent: parent,
			})
		}
	}
	return &SendGroupsRequest{
		Groups:              groupsID,
		TaskId:              taskID,
		ResourceRequirement: resourcesMap,
		TopInfo:             topInfo,
	}
}

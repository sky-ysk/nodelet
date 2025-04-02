package plugins

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"hit.edu/framework/pkg/apimachinery/runtime"
	"hit.edu/framework/pkg/apimachinery/runtime/schema"
	"hit.edu/framework/pkg/apimachinery/runtime/serializer"
	apis "hit.edu/framework/pkg/apis/cores"
	metav1 "hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/client-go/clients"
	"hit.edu/framework/pkg/client-go/clients/typed/core"
	"hit.edu/framework/pkg/client-go/rest"
	"hit.edu/framework/pkg/component-base/logs"
	"hit.edu/framework/pkg/scheduler/framework"
	"hit.edu/framework/pkg/scheduler/transport"
	"io"
	"net/http"
	"time"
)

type ScorePluginDBY struct {
	pluginClient ScorePluginClient
	clientSet    *clients.ClientSet
	taskClient   core.TaskInterface
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
	logs.Infof("the groups request send to dts is %s", string(jsonData))
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

func (sp *ScorePluginDBY) getTaskNameByID(ctx context.Context, taskID string) string {
	list, err := sp.taskClient.List(ctx, metav1.ListOptions{})
	if err != nil {
		logs.Error(err.Error())
		return ""
	}
	for _, item := range list.Items {
		if item.Status.TaskID == taskID {
			logs.Info("get taskname %s by id %s", item.Name, taskID)
			return item.Name
		}
	}
	return ""
}

func (sp *ScorePluginDBY) Score(ctx context.Context, group *apis.Group, nodeName string) (int64, *framework.Status) {
	//TODO 没测过
	logs.Infof("use DTS plugin to generate a score on %s", nodeName)
	taskName := sp.getTaskNameByID(ctx, group.Status.Belongs.TaskID)

	request := transport.ScoreRequest{
		GroupID: group.Spec.Name,
		TaskID:  taskName,
		NodeID:  nodeName,
	}
	jsonData, err := json.Marshal(request)
	logs.Infof("the request send to dts is %s", string(jsonData))
	if err != nil {
		logs.Fatal(err)
		return 0, framework.NewStatus(framework.Error, err.Error())
	}
	data, err := sp.pluginClient.SendData(jsonData, "/schedule/getSchedule")
	time.Sleep(5 * time.Second)
	logs.Info("sleep 5s to get dts score")
	data, err = sp.pluginClient.SendData(jsonData, "/schedule/getSchedule")
	if err != nil {
		return 0, framework.NewStatus(framework.Error, err.Error())
	}
	logs.Infof("raw result given by dts is %s ", string(data))
	var resp transport.ScoreRespData
	err = json.Unmarshal(data, &resp)
	if err != nil {
		logs.Error(err)
		return 0, framework.NewStatus(framework.Error, err.Error())
	}
	logs.Infof("score given by dts plugin : %d, group : %s", resp.Score, resp.GroupID)
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
	scheme := runtime.NewScheme()
	apis.AddToScheme(scheme)
	c := &rest.Config{
		Host:    "http://localhost:10000",
		APIPath: "/apis/resources/v1",
		ContentConfig: rest.ContentConfig{
			AcceptContentTypes: "application/json; charset=UTF-8", //text/plain; charset=UTF-8
			ContentType:        "application/json; charset=UTF-8", //application/json; charset=UTF-8
			GroupVersion: &schema.GroupVersion{
				Group:   "resources",
				Version: "v1",
			},
			NegotiatedSerializer: serializer.NewCodecFactory(scheme),
		},
		UserAgent: "defaultUserAgent",
		Transport: &http.Transport{
			MaxIdleConns:        100,              // 最大空闲连接数
			IdleConnTimeout:     90 * time.Second, // 空闲连接超时时间
			TLSHandshakeTimeout: 10 * time.Second, // TLS 握手超时时间
		},
		Timeout: 3600 * time.Second,
	}

	cs, err := clients.NewForConfig(c)
	if err != nil {
		panic(err)
	}

	tc := cs.Core().Tasks("test")
	return &ScorePluginDBY{
		clientSet:    cs,
		taskClient:   tc,
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
	taskID := task.Spec.Name
	//TODO 可能需要做深复制 @lbh
	for _, group := range task.Spec.Groups {
		groupsID = append(groupsID, group.Spec.Name)
		resources := make([]apis.ResourceRequirement, 0)
		for _, requirement := range group.Spec.ResourceRequirements {
			resources = append(resources, requirement)
		}
		if len(resources) > 0 {
			resourcesMap[group.Spec.Name] = resources
		}
		for _, parent := range group.Spec.Parents {
			//fmt.Println("parent : ", parent, " child ", group.Status.GroupID)
			topInfo = append(topInfo, GroupTopInfo{
				Child:  group.Spec.Name,
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

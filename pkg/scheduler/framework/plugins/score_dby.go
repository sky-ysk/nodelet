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
	"hit.edu/framework/pkg/client-go/clients"
	"hit.edu/framework/pkg/client-go/clients/typed/core"
	"hit.edu/framework/pkg/client-go/rest"
	"hit.edu/framework/pkg/component-base/logs"
	"hit.edu/framework/pkg/scheduler/framework"
	"hit.edu/framework/pkg/scheduler/transport"
	"hit.edu/framework/pkg/utils/value"
	"io"
	"math/rand"
	"net/http"
	"time"
)

type ScorePluginDBY struct {
	pluginClient ScorePluginClient
	clientSet    *clients.ClientSet
	taskClient   core.TaskInterface
	valueEngine  *value.Engine
	r            *rand.Rand
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
		logs.Error(err.Error())
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			logs.Error(err.Error())
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
	logs.Infof("the group request send to dts is task %s, time %s", request.TaskId, time.Now().String())
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
			Timeout: time.Second * 3600 * 24,
			//走http1
			Transport: &http.Transport{
				MaxIdleConns:        100,              // 最大空闲连接数
				MaxIdleConnsPerHost: 10,               // 每个主机的最大空闲连接数
				IdleConnTimeout:     30 * time.Second, // 空闲连接的超时时间
				//AllowHTTP: true, // 允许非加密的HTTP/2连接（测试环境可用，生产环境建议使用TLS加密）
			},
		},
	}
}

func (sp *ScorePluginDBY) Name() string {
	return "ScorePluginForDuBoyu"
}

//func (sp *ScorePluginDBY) getTaskNameByID(ctx context.Context, taskID string) string {
//	list, err := sp.taskClient.List(ctx, metav1.ListOptions{})
//	if err != nil {
//		logs.Error(err.Error())
//		return ""
//	}
//	for _, item := range list.Items {
//		if item.Status.TaskID == taskID {
//			logs.Info("get taskname %s by id %s", item.Name, taskID)
//			return item.Name
//		}
//	}
//	return ""
//}

func (sp *ScorePluginDBY) Score(ctx context.Context, group *apis.Group, nodeName string) (int64, *framework.Status) {
	//TODO 没测过
	//logs.Infof("use DTS plugin to generate a score on %s", nodeName)
	taskName := group.Status.Belong.Name
	randScore := int64(sp.r.Intn(3))
	request := transport.ScoreRequest{
		GroupID: group.ObjectMeta.Name,
		TaskID:  taskName,
		NodeID:  nodeName,
	}
	jsonData, err := json.Marshal(request)
	//logs.Infof("the request send to dts is %s", string(jsonData))
	if err != nil {
		logs.Error(err.Error())
		return randScore, framework.NewStatus(framework.Error, err.Error())
	}
	time.Sleep(8 * time.Second)
	logs.Infof("now send schedule request to dts, group %s , time %s", request.GroupID, time.Now().String())
	var resp transport.ScoreRespData
	//先发第一次 理论上第一次是收不到的
	data, err := sp.pluginClient.SendData(jsonData, "/schedule/getSchedule")
	if err != nil {
		logs.Error(err.Error())
		return randScore, framework.NewStatus(framework.Error, err.Error())
	}
	//logs.Infof("first time raw result given by dts is %s ,\n group : %s, node %s\"", string(data), group.Name, nodeName)
	err = json.Unmarshal(data, &resp)
	if err != nil {
		logs.Error(err.Error())
		return randScore, framework.NewStatus(framework.Error, err.Error())
	}
	//logs.Infof("raw result given by dts is %s ,\n group : %s, node %s\"", string(data), group.Name, nodeName)
	if resp.GroupID == group.ObjectMeta.Name {
		return resp.Score, framework.NewStatus(framework.Success, "")
	}

	//理论上第二次才能收到分数
	time.Sleep(8 * time.Second)
	data, err = sp.pluginClient.SendData(jsonData, "/schedule/getSchedule")
	if err != nil {
		logs.Error(err.Error())
		return randScore, framework.NewStatus(framework.Error, err.Error())
	}
	err = json.Unmarshal(data, &resp)
	if err != nil {
		logs.Error(err.Error())
		return randScore, framework.NewStatus(framework.Error, err.Error())
	}
	logs.Infof("raw result given by dts is %s ,\n group : %s, node %s\"", string(data), group.Name, nodeName)
	if resp.GroupID == group.ObjectMeta.Name {
		logs.Infof("score by dts is %d ,\n group : %s, node %s\"", resp.Score, group.Name, nodeName)
		return resp.Score, framework.NewStatus(framework.Success, "")
	}

	//大约20%的请求会走到这里
	logs.Warnf("group %s get schedule fail, node %s", group.Name, nodeName)
	for {
		time.Sleep(5 * time.Second)
		data, err = sp.pluginClient.SendData(jsonData, "/schedule/getSchedule")
		if err != nil {
			logs.Error(err.Error())
			return randScore, framework.NewStatus(framework.Error, err.Error())
		}
		logs.Infof("loop raw result given by dts is %s ,\n group : %s, node %s\"", string(data), group.Name, nodeName)
		err = json.Unmarshal(data, &resp)
		if err != nil {
			logs.Error(err.Error())
			return randScore, framework.NewStatus(framework.Error, err.Error())
		}
		//logs.Infof("raw result given by dts is %s ,\n group : %s, node %s\"", string(data), group.Name, nodeName)
		if resp.GroupID == group.ObjectMeta.Name {
			break
		}
		logs.Warnf("group %s get schedule fail again, node %s", group.Name, nodeName)
	}
	logs.Infof("group %s finally get the score ! Node %s, Score %d", group.Name, nodeName, resp.Score)
	return resp.Score, framework.NewStatus(framework.Success, "")
}

func (sp *ScorePluginDBY) SendGroups(ctx context.Context, task *apis.Task) (bool, *framework.Status) {
	request := sp.buildSendGroupsRequest(ctx, task)
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
		r:            rand.New(rand.NewSource(time.Now().UnixNano())),
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

//func BuildSendGroupsRequest(ctx context.Context, task *apis.Task) *SendGroupsRequest {
//	return buildSendGroupsRequest(ctx, task)
//}

func (sp *ScorePluginDBY) buildSendGroupsRequest(ctx context.Context, task *apis.Task) *SendGroupsRequest {
	topInfo := make([]GroupTopInfo, 0)
	groupsID := make([]string, 0)
	resourcesMap := make(map[string][]apis.ResourceRequirement)
	taskID := task.ObjectMeta.Name
	//TODO 可能需要做深复制 @lbh
	for _, group := range task.Spec.Groups {
		groupID := task.Status.Groups[group.Name].Name
		groupsID = append(groupsID, groupID)
		resources := make([]apis.ResourceRequirement, 0)
		for _, requirement := range group.ResourceRequirements {
			resources = append(resources, requirement)
		}
		if len(resources) > 0 {
			resourcesMap[groupID] = resources
		}
		for _, parent := range group.Parents {
			//fmt.Println("parent : ", parent, " child ", group.Status.GroupID)
			parID := task.Status.Groups[parent].Name
			topInfo = append(topInfo, GroupTopInfo{
				Child:  groupID,
				Parent: parID,
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

package lib

import (
	"bytes"
	"encoding/json"
	"fmt"
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/component-base/logs"
	"hit.edu/framework/pkg/utils/value"
	"io"
	"net/http"
)

type DownloadParam struct {
	UserId   string `json:"user_id"`
	ModelId  string `json:"model_id"`
	Path     string `json:"path"`
	Filename string `json:"filename"`
}

type DownloadModelStrategy struct{}

func (dms *DownloadModelStrategy) Execute(url string, params []apis.Value, engine *value.Engine, runtime *apis.Runtime, action *apis.Action) (string, error) {
	var userId string
	var modelId string
	var path string
	var filename string
	for _, param := range params {
		if param.Name == "user_id" {
			if param.Type == apis.ConstData {
				userId = param.Value
			}
		} else if param.Name == "model_id" {
			if param.Type == apis.ConstData {
				modelId = param.Value
			}
		} else if param.Name == "path" {
			if param.Type == apis.ConstData {
				path = param.Value
			}
		} else if param.Name == "filename" {
			if param.Type == apis.ConstData {
				filename = param.Value
			}
		}
	}
	taskId, err := PublishDownloadModelInst(userId, modelId, path, filename, url)
	if err != nil {
		logs.Errorf("[DEVICE RUNTIME] PublishDownloadInst fail")
		return "", err
	}
	logs.Infof("[DEVICE RUNTIME] Task ID is %s", taskId)
	return taskId, nil
}

// PublishDownloadModelInst 发送下载模型的任务请求
func PublishDownloadModelInst(userId string, modelId string, path string, filename string, url string) (string, error) {
	// 构建请求体
	logs.Infof("download url %s", url)
	requestBody := DownloadParam{
		UserId:   userId,
		ModelId:  modelId,
		Path:     path,
		Filename: filename,
	}
	// 将请求体编码为 JSON
	//logs.Infof("req body is %v", requestBody)
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("无法编码任务数据: %v", err)
	}
	//logs.Infof("download req body %s", string(jsonData))

	// 创建一个 POST 请求
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %v", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")

	// 发起 HTTP 请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("发起请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 检查状态码
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		logs.Errorf("download not ok %s", string(bodyBytes))
		return "", fmt.Errorf("请求失败，状态码: %d，响应体: %s", resp.StatusCode, string(bodyBytes))
	}

	// 读取并解析响应体
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应体失败: %v", err)
	}

	var taskResponse AbilityInstResponse
	if err := json.Unmarshal(bodyBytes, &taskResponse); err != nil {
		return "", fmt.Errorf("解析响应体失败: %v", err)
	}

	// 返回任务 ID
	return taskResponse.TaskId, nil
}

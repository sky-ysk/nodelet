package lib

import (
	"bytes"
	"encoding/json"
	"fmt"
	"hit.edu/framework/pkg/component-base/logs"
	"io"
	"net/http"
)

type DownloadParam struct {
	UserId   string `json:"user_id"`
	ModelId  string `json:"model_id"`
	Path     string `json:"path"`
	Filename string `json:"filename"`
}

// PublishGrabBallInst 向指定的 API 发送抓取小球的任务请求
func PublishDownloadModelInst(user_id string, model_id string, path string, filename string, url string) (string, error) {
	// 构建请求体
	logs.Infof("download url %s", url)
	requestBody := DownloadParam{
		UserId:   user_id,
		ModelId:  model_id,
		Path:     path,
		Filename: filename,
	}
	// 将请求体编码为 JSON
	logs.Infof("req body is %v", requestBody)
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("无法编码任务数据: %v", err)
	}
	logs.Infof("download req body %s", string(jsonData))

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

	var taskResponse GrabBallResponse
	if err := json.Unmarshal(bodyBytes, &taskResponse); err != nil {
		return "", fmt.Errorf("解析响应体失败: %v", err)
	}

	// 返回任务 ID
	return taskResponse.TaskId, nil
}

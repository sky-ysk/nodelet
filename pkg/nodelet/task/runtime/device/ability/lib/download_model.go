package lib

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type DownloadParam struct {
	user_id  string `json:"user_id"`
	model_id string `json:"model_id"`
	path     string `json:"path"`
	filename string `json:"filename"`
}

// PublishGrabBallInst 向指定的 API 发送抓取小球的任务请求
func PublishDownloadModelInst(user_id string, model_id string, path string, filename string, url string) (string, error) {
	// 构建请求体
	requestBody := DownloadParam{
		user_id:  user_id,
		model_id: model_id,
		path:     path,
		filename: filename,
	}
	// 将请求体编码为 JSON
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("无法编码任务数据: %v", err)
	}

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

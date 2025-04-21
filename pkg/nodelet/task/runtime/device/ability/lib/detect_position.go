package lib

import (
	"encoding/json"
	"fmt"
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/component-base/logs"
	"hit.edu/framework/pkg/utils/value"
	"io"
	"net/http"
)

type DetectPositionStrategy struct{}

func (dps *DetectPositionStrategy) Execute(url string, params []apis.Value, engine *value.Engine, runtime *apis.Runtime) (string, error) {
	taskId, err := PublishDetectPositionInst(url)
	if err != nil {
		logs.Errorf("[DEVICE RUNTIME] PublishDetectPosition fail")
		return "", err
	}
	logs.Infof("[DEVICE RUNTIME] Task ID is %s", taskId)
	return taskId, nil
}

// DetectPositionResponse 定义了预期的响应体结构
type DetectPositionResponse struct {
	TaskId string `json:"taskId"`
}

// PublishDetectPositionInst 向指定的 API 发送任务检测请求
func PublishDetectPositionInst(url string) (string, error) {
	// 创建一个 POST 请求，无需 payload
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %v", err)
	}

	// 设置请求头，指定 Content-Type
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

	var taskResponse DetectPositionResponse
	if err := json.Unmarshal(bodyBytes, &taskResponse); err != nil {
		return "", fmt.Errorf("解析响应体失败: %v", err)
	}

	// 返回任务 ID
	return taskResponse.TaskId, nil
}

package lib

import (
	"encoding/json"
	"fmt"
	"hit.edu/framework/pkg/component-base/logs"
	"io"
	"net/http"
)

// TaskResponse 定义了任务响应的结构
type TaskResponse struct {
	EndTime      string      `json:"end_time"`      // 结束时间
	ExecutorID   string      `json:"executor_id"`   // 执行器 ID
	ExecutorType string      `json:"executor_type"` // 执行器类型
	ID           string      `json:"id"`            // 任务 ID
	Message      string      `json:"message"`       // 消息
	Payload      interface{} `json:"payload"`       // 负载数据，可能有多种结构
	StartTime    string      `json:"start_time"`    // 开始时间
	State        TaskState   `json:"state"`         // 状态
	Timeout      int         `json:"timeout"`       // 超时时间
}

type TaskState string

const (
	Unstarted TaskState = "unstarted"
	Running   TaskState = "running"
	Finished  TaskState = "finished"
	Error     TaskState = "error"
	Cancelled TaskState = "cancelled"
)

// GetTaskStatus 向指定的 API 发送 GET 请求以查询任务状态
func GetTaskStatus(taskId string, url string) (TaskResponse, error) {
	// 构建完整的 API URL
	apiURL := fmt.Sprintf("%s/api/task/%s/status", url, taskId)

	// 创建一个 GET 请求
	logs.Infof("Create Get Request...")
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		logs.Errorf("Create Get Request fail: %v", err)
		return TaskResponse{}, fmt.Errorf("Create Get Request fail: %v", err)
	}
	logs.Infof("Create Get Request Successfully")

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")

	// 发起 HTTP 请求
	logs.Infof("Publish Request...")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logs.Errorf("Publish Request fail: %v", err)
		return TaskResponse{}, fmt.Errorf("Publish Request fail: %v", err)
	}
	logs.Infof("Publish Request successfully")
	defer resp.Body.Close()

	// 检查状态码
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return TaskResponse{}, fmt.Errorf("Request fail, code: %d，response: %s", resp.StatusCode, string(bodyBytes))
	}

	// 读取并解析响应体
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		logs.Errorf(":Read Response fail %v", err)
		return TaskResponse{}, fmt.Errorf(":Read Response fail %v", err)
	}

	var taskResponse TaskResponse
	if err := json.Unmarshal(bodyBytes, &taskResponse); err != nil {
		logs.Errorf("Unmarshal Response fail: %v", err)
		return TaskResponse{}, fmt.Errorf("unmarshal Response fail: %v", err)
	}

	return taskResponse, nil
}

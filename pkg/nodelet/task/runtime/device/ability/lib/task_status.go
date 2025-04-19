package lib

import (
	"encoding/json"
	"fmt"
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/component-base/logs"
	"io"
	"net/http"
	"strings"
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

// parseWorldPoints 专门将 payload 中的 world_points 解析为 [][]float64 类型
func parseWorldPoints(payload interface{}) ([][]float64, error) {
	// 断言 payload 是一个 map[string]interface{}
	worldPointsMap, ok := payload.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("payload 类型断言失败，不是 map[string]interface{} 类型")
	}

	// 获取 world_points 对应的值
	worldPointsInterface, ok := worldPointsMap["world_points"]
	if !ok {
		return nil, fmt.Errorf("world_points 字段不存在")
	}

	// 断言 world_points 是一个 []interface{}
	worldPointsSlice, ok := worldPointsInterface.([]interface{})
	if !ok {
		return nil, fmt.Errorf("world_points 类型断言失败，不是 []interface{} 类型")
	}

	var result [][]float64

	// 遍历 world_points 的每一行
	for _, rowInterface := range worldPointsSlice {
		// 断言每一行是 []interface{} 类型
		rowSlice, ok := rowInterface.([]interface{})
		if !ok {
			return nil, fmt.Errorf("world_points 中的行类型断言失败，不是 []interface{} 类型")
		}

		var floatRow []float64

		// 遍历行中的每个元素，将其断言为 float64 类型
		for _, valueInterface := range rowSlice {
			valueFloat64, ok := valueInterface.(float64)
			if !ok {
				return nil, fmt.Errorf("world_points 中的元素类型断言失败，不是 float64 类型")
			}
			floatRow = append(floatRow, valueFloat64)
		}

		result = append(result, floatRow)
	}

	return result, nil
}

// worldPointsToString 将worldPoints转换为string
func worldPointsToString(worldPoints [][]float64) string {
	var builder strings.Builder
	for i, row := range worldPoints {
		// 将一行中的元素转换为字符串，并用逗号分隔
		var rowStrings []string
		for _, value := range row {
			rowStrings = append(rowStrings, fmt.Sprintf("%f", value))
		}
		builder.WriteString(strings.Join(rowStrings, ","))

		// 如果不是最后一行，添加分号
		if i < len(worldPoints)-1 {
			builder.WriteString(";")
		}
	}

	return builder.String()
}

// ParsePayLoad 将payload进行解析
func ParsePayLoad(inst string, payload interface{}) ([]apis.Value, error) {
	switch inst {
	case "DetectPosition":
		// 使用对应的函数进行解析
		worldPoints, err := parseWorldPoints(payload)
		if err != nil {
			logs.Errorf("[DEVICE RUNTIME] parse world points fail: %s", err.Error())
			return []apis.Value{}, err
		}
		// 转换为字符串
		str := worldPointsToString(worldPoints)
		// 放到Value中
		outputs := []apis.Value{
			{
				ValueType: apis.ComposeType,
				Value:     str,
				Name:      "worldPoints",
				Type:      apis.ConstData,
			},
		}
		return outputs, nil

	case "":

	default:

	}
	return []apis.Value{}, nil
}

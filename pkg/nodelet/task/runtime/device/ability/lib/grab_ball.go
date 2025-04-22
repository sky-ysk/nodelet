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
	"strconv"
	"strings"
)

type GrabBallStrategy struct{}

func (gbs *GrabBallStrategy) Execute(url string, params []apis.Value, engine *value.Engine, runtime *apis.Runtime) (string, error) {
	var worldPoints [][]float64
	for _, param := range params {
		if param.Name == "worldPoints" {
			if param.Type == apis.ConstData {
				worldPoints = GetWorldPoints(param)
			} else if param.Type == apis.LocalData {
				worldPointValue, err := engine.ExtractLocalValue(&param, *runtime)
				if err != nil {
					logs.Errorf("[DEVICE RUNTIME] Engine ExtractLocalValue fail")
					return "", err
				}
				worldPoints = GetWorldPoints(*worldPointValue)
			}
		}
	}

	taskId, err := PublishGrabBallInst(worldPoints, url)
	if err != nil {
		logs.Errorf("[DEVICE RUNTIME] PublishGrabBallInst fail")
		return "", err
	}
	logs.Infof("[DEVICE RUNTIME] Task ID is %s", taskId)
	return taskId, nil
}

type WorldPoints struct {
	WorldPoints [][]float64 `json:"world_points"`
}

func GetWorldPoints(param apis.Value) [][]float64 {

	if param.Name == "worldPoints" {
		result, err := parseCustomFormat(param.Value)
		if err != nil {
			logs.Error("Parse world points failed: %v", err.Error())
			return nil
		}
		return result
	}

	return nil
}

// 如果你的 param.Value 是其他格式的字符串（例如自定义格式），可以使用下面的解析方法
func parseCustomFormat(value string) ([][]float64, error) {
	var result [][]float64
	// 假设字符串格式是 "1.1,2.2;3.3,4.4" 这样的形式
	parts := strings.Split(value, ";")
	for _, part := range parts {
		if part == "" {
			continue
		}
		point := make([]float64, 0)
		coords := strings.Split(part, ",")
		for _, coord := range coords {
			if coord == "" {
				continue
			}
			f, err := strconv.ParseFloat(coord, 64)
			if err != nil {
				return nil, fmt.Errorf("failed to parse coordinate %s: %v", coord, err)
			}
			point = append(point, f)
		}
		result = append(result, point)
	}
	return result, nil
}

// PublishGrabBallInst 向指定的 API 发送抓取小球的任务请求
func PublishGrabBallInst(worldPoints [][]float64, url string) (string, error) {
	// 构建请求体
	requestBody := WorldPoints{
		worldPoints,
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

	var taskResponse AbilityInstResponse
	if err := json.Unmarshal(bodyBytes, &taskResponse); err != nil {
		return "", fmt.Errorf("解析响应体失败: %v", err)
	}

	// 返回任务 ID
	return taskResponse.TaskId, nil
}

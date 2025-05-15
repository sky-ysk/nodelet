package lib

import (
	"encoding/json"
	"fmt"
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/component-base/logs"
	"hit.edu/framework/pkg/utils/value"
	"io"
	"net/http"
	"strings"
)

/*
	这个文件实现了检测操作的策略
	1.检测小球位置
	2.检测工件是否合格
*/
/* -----------------------------------------发布指令部分---------------------------------------------- */

// 1.检测小球位置

// DetectPositionStrategy 策略
type DetectPositionStrategy struct{}

func (dps *DetectPositionStrategy) Execute(url string, params []apis.Value, engine *value.Engine, runtime *apis.Runtime, action *apis.Action) (string, error) {
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

// 2. 检测工件是否合格

type DetectWorkpieceStrategy struct{}

func (dws *DetectWorkpieceStrategy) Execute(url string, params []apis.Value, engine *value.Engine, runtime *apis.Runtime, action *apis.Action) (string, error) {
	taskId, err := PublishDetectWorkpieceInst(url)
	if err != nil {
		logs.Errorf("[DEVICE RUNTIME] PublishDetectPosition fail")
		return "", err
	}
	logs.Infof("[DEVICE RUNTIME] Task ID is %s", taskId)
	return taskId, nil
}

func PublishDetectWorkpieceInst(url string) (string, error) {
	// 创建一个 GET 请求
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %v", err)
	}

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

/* -----------------------------------------指令解析部分---------------------------------------------- */

// 1.检测小球位置解析策略

type DetectPositionParseStrategy struct{}

func (dpps *DetectPositionParseStrategy) Execute(payload interface{}) ([]apis.Value, error) {
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
		{
			Value:     "true",
			Name:      "success",
			Type:      apis.LocalData,
			ValueType: apis.BoolType,
		},
	}
	return outputs, nil
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

// 2.检测工件合格解析策略

type DetectWorkpieceParseStrategy struct{}

func (dwps *DetectWorkpieceParseStrategy) Execute(payload interface{}) ([]apis.Value, error) {
	// 使用对应的函数进行解析
	isQualified, err := parseQualified(payload)
	if err != nil {
		logs.Errorf("[DEVICE RUNTIME] parse world points fail: %s", err.Error())
		return []apis.Value{}, err
	}

	var outputs = make([]apis.Value, 1)
	// 放到Value中
	if isQualified == "qualified" {
		outputs[0] = apis.Value{
			Value:     "true",
			Name:      "isQualified",
			Type:      apis.LocalData,
			ValueType: apis.BoolType,
		}
	} else if isQualified == "unqualified" {
		outputs[0] = apis.Value{
			Value:     "true",
			Name:      "isQualified",
			Type:      apis.LocalData,
			ValueType: apis.BoolType,
		}
	} else {
		logs.Errorf("[DEVICE RUNTIME] PARSE DetectWorkpiece payload fail")
		return nil, fmt.Errorf("[DEVICE RUNTIME] PARSE DetectWorkpiece payload fail")

	}

	return outputs, nil
}

// parseQualified 专门将 payload 中的label 解析为string类型
func parseQualified(payload interface{}) (string, error) {
	// 将 payload 转换为 map[string]interface{}
	payloadMap, ok := payload.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("payload is not a map[string]interface{}")
	}

	// 从 map 中提取 "label" 字段
	label, ok := payloadMap["label"].(string)
	if !ok {
		return "", fmt.Errorf("label is not a string")
	}

	return label, nil
}

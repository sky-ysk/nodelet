package lib

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/component-base/logs"
	"hit.edu/framework/pkg/utils/value"
	"io"
	"net/http"
	"strconv"
	"strings"
)

/*
	这个文件实现了机器人进行躯干转动的策略

*/

type TurnLeftStrategy struct{}

func (tls *TurnLeftStrategy) Execute(url string, params []apis.Value, engine *value.Engine, runtime *apis.Runtime, action *apis.Action) (string, error) {

	// 创建 HTTP GET 请求
	resp, err := http.Get(url)
	if err != nil {
		logs.Errorf("发送请求失败: %v", err)
		return "", err
	}
	defer resp.Body.Close()

	// 检查 HTTP 响应状态码
	if resp.StatusCode != http.StatusOK {
		logs.Errorf("Turn Left response: %d", resp.StatusCode)
		return "", fmt.Errorf("turn left response: %d", resp.StatusCode)
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

type TurnRightStrategy struct{}

func (trs *TurnRightStrategy) Execute(url string, params []apis.Value, engine *value.Engine, runtime *apis.Runtime, action *apis.Action) (string, error) {

	// 创建 HTTP GET 请求
	resp, err := http.Get(url)
	if err != nil {
		logs.Errorf("发送请求失败: %v", err)
		return "", err
	}
	defer resp.Body.Close()

	// 检查 HTTP 响应状态码
	if resp.StatusCode != http.StatusOK {
		logs.Errorf("Turn Right response: %d", resp.StatusCode)
		return "", fmt.Errorf("turn right response: %d", resp.StatusCode)
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

type TurnStrategy struct{}

func (ts *TurnStrategy) Execute(url string, params []apis.Value, engine *value.Engine, runtime *apis.Runtime, action *apis.Action) (string, error) {
	var torsoAngle []float64
	for _, param := range params {
		logs.Infof("[DEVICE RUNTIME] param is %v, name %s, type%s", param, param.Name, param.Type)
		if param.Name == "worldPoints" {
			if param.Type == apis.ConstData {
				torsoAngle = getTorsoAngle(param)
			} else if param.Type == apis.LocalData {
				torsoAngleValue, err := engine.ExtractLocalValue(&param, *action)
				if err != nil {
					logs.Errorf("[DEVICE RUNTIME] Engine ExtractLocalValue fail")
					return "", err
				}
				torsoAngle = getTorsoAngle(*torsoAngleValue)
			}
		}
	}

	taskId, err := PublishTurnAngleInst(torsoAngle, url)
	if err != nil {
		logs.Errorf("[DEVICE RUNTIME] PublishTurnAngleInst fail")
		return "", err
	}
	logs.Infof("[DEVICE RUNTIME] Task ID is %s", taskId)
	return taskId, nil
}

// TorsoAngle 描述躯干转动角度的数据类型
type TorsoAngle struct {
	Angle []float64 `json:"angle"`
}

// getTorsoAngle 从param中获取躯干转动的角度
func getTorsoAngle(param apis.Value) []float64 {

	if param.Name == "torsoAngle" {
		result, err := parseTorsoAngle(param.Value)
		if err != nil {
			logs.Error("Parse world points failed: %v", err.Error())
			return nil
		}
		return result
	}

	return nil
}

// parseTorsoAngle 将string类型的数据进行解析，转换为TorsoAngle类型
func parseTorsoAngle(value string) ([]float64, error) {
	var result []float64
	coords := strings.Split(value, ",")
	if len(coords) < 2 {
		return nil, errors.New("invalid input format")
	}
	for _, coord := range coords {
		if coord == "" {
			return nil, errors.New("invalid input format")
		}
		f, err := strconv.ParseFloat(coord, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse coordinate %s: %v", coord, err)
		}
		result = append(result, f)
	}
	return result, nil
}

// PublishTurnAngleInst 向指定的 API 发送转动躯体（指定角度）的任务请求
func PublishTurnAngleInst(torsoAngle []float64, url string) (string, error) {
	// 构建请求体
	requestBody := TorsoAngle{
		torsoAngle,
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

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

/*
	这个文件实现了机器人进行躯干转动的策略
*/

type TurnLabel struct {
	Label string `json:"label"`
}

type TurnStrategy struct{}

func (ts *TurnStrategy) Execute(url string, params []apis.Value, engine *value.Engine, runtime *apis.Runtime, action *apis.Action) (string, error) {
	var label string
	for _, param := range params {
		if param.Name == "label" {
			label = param.Value
			logs.Infof("[DEVICE RUNTIME] get param: label")
		}
	}
	taskId, err := PublishTurnInst(label, url)
	if err != nil {
		logs.Errorf("[DEVICE RUNTIME] PublishTurnInst fail, err:%v", err)
		return "", err
	}
	return taskId, nil
}

// PublishTurnInst 使用PublishPredictByUrl接口进行预测
func PublishTurnInst(label string, url string) (string, error) {

	// 构建请求体
	requestBody := TurnLabel{
		label,
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

// todo 以角度为参数的接口

type TurnAngleStrategy struct{}

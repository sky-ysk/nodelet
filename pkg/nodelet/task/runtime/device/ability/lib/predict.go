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
)

/* -----------------------------------------发布指令部分---------------------------------------------- */

// PredictByUrlReq PublishPredictByUrlInst的参数
type PredictByUrlReq struct {
	Url        string `json:"url"`
	Position   string `json:"position"`
	Type       string `json:"type"`
	Compressed bool   `json:"compressed"`
}
type PreByUrlStrategy struct{}

func (pbus *PreByUrlStrategy) Execute(url string, params []apis.Value, engine *value.Engine, runtime *apis.Runtime) (string, error) {
	// 需要用到的参数
	var cameraUrl string
	var position string
	var imageType string
	var compressed bool

	for _, param := range params {
		if param.Name == "url" { //TODO 更改解析方式
			if param.Type == apis.ConstData {
				cameraUrl = param.Value
			}
		} else if param.Name == "position" {
			if param.Type == apis.ConstData {
				position = param.Value
			}
		} else if param.Name == "type" {
			if param.Type == apis.ConstData {
				imageType = param.Value
			}
		} else if param.Name == "compressed" {
			if param.Type == apis.ConstData {
				var err error
				compressed, err = strconv.ParseBool(param.Value)
				if err != nil {
					logs.Errorf("[DEVICE RUNTIME] parse param[compressed] error:%v", err.Error())
				}
			}
		}
	}
	taskId, err := PublishPredictByUrlInst(compressed, cameraUrl, position, imageType, url)
	if err != nil {
		logs.Errorf("[DEVICE RUNTIME] PublishPredictByUrlInst fail")
		return "", err
	}
	logs.Infof("[DEVICE RUNTIME] Task ID is %s", taskId)
	return taskId, nil
}

// PublishPredictByUrlInst 使用PublishPredictByUrl接口进行预测
func PublishPredictByUrlInst(compressed bool, cameraUrl string, position string, imageType string, url string) (string, error) {

	// 构建请求体
	requestBody := PredictByUrlReq{
		Url:        cameraUrl,
		Position:   position,
		Type:       imageType,
		Compressed: compressed,
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

/* -----------------------------------------解析指令部分---------------------------------------------- */

// PreByUrlParseStrategy 是PredictByUrl接口的解析策略
type PreByUrlParseStrategy struct{}

//func (pbups *PreByUrlParseStrategy) Execute(payload interface{}) ([]apis.Value, error) {
//
//	// 放到Value中
//	outputs := []apis.Value{
//		{
//			ValueType: apis.ComposeType,
//			Value:     str,
//			Name:      "worldPoints",
//			Type:      apis.ConstData,
//		},
//		{
//			Value:     "true",
//			Name:      "success",
//			Type:      apis.LocalData,
//			ValueType: apis.BoolType,
//		},
//	}
//	return outputs, nil
//}

// parsePrediction 用于从payload中解析预测结果
func parsePrediction(data interface{}) (int, error) {
	// 将 payload 转换为 map[string]interface{}
	payload, ok := data.(map[string]interface{})
	if !ok {
		return 0, fmt.Errorf("[DEVICE RUNTIME] payload parse fail")
	}

	// 获取 prediction 的值
	prediction, ok := payload["prediction"]
	if !ok {
		return 0, fmt.Errorf("[DEVICE RUNTIME] prediction is not exist")
	}

	// 将 prediction 转换为整数
	predictionInt, ok := prediction.(float64) // json 解码时整数可能会被解析为 float64
	if !ok {
		return 0, fmt.Errorf("prediction 转换为整数失败")
	}

	return int(predictionInt), nil
}

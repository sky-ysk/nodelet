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
	这个文件实现了手臂操作的策略
	1.抓取工件
	2.放置工件
	3.手臂初始化
*/

// 1.抓取工件

type GrabWorkpieceLabel struct {
	Label string `json:"label"`
}

type GrabWorkpieceStrategy struct{}

func (gws *GrabWorkpieceStrategy) Execute(url string, params []apis.Value, engine *value.Engine, runtime *apis.Runtime, action *apis.Action) (string, error) {
	var label string
	for _, param := range params {
		if param.Name == "label" {
			label = param.Value
			logs.Infof("[DEVICE RUNTIME] get param: label")
		}
	}
	taskId, err := PublishGrabWorkpieceInst(label, url)
	if err != nil {
		logs.Errorf("[DEVICE RUNTIME] PublishGrabWorkpieceInst fail, err:%v", err)
		return "", err
	}
	return taskId, nil
}
func PublishGrabWorkpieceInst(label string, url string) (string, error) {
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

// 2.放置工件

type PutWorkpieceLabel struct {
	Label string `json:"label"`
}
type PutWorkpieceStrategy struct{}

func (pws *PutWorkpieceStrategy) Execute(url string, params []apis.Value, engine *value.Engine, runtime *apis.Runtime, action *apis.Action) (string, error) {
	var label string
	for _, param := range params {
		if param.Name == "label" {
			label = param.Value
			logs.Infof("[DEVICE RUNTIME] get param: label")
		}
	}
	taskId, err := PublishPutWorkpieceInst(label, url)
	if err != nil {
		logs.Errorf("[DEVICE RUNTIME] PublishPutWorkpieceInst fail, err:%v", err)
		return "", err
	}
	return taskId, nil
}
func PublishPutWorkpieceInst(label string, url string) (string, error) {
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

// 3.初始化

type GrabInitStrategy struct{}

func (gis *GrabInitStrategy) Execute(url string, params []apis.Value, engine *value.Engine, runtime *apis.Runtime, action *apis.Action) (string, error) {
	var label string
	for _, param := range params {
		if param.Name == "label" {
			label = param.Value
			logs.Infof("[DEVICE RUNTIME] get param: label")
		}
	}
	taskId, err := PublishGrabInitInst(label, url)
	if err != nil {
		logs.Errorf("[DEVICE RUNTIME] PublishPutWorkpieceInst fail, err:%v", err)
		return "", err
	}
	return taskId, nil
}
func PublishGrabInitInst(label string, url string) (string, error) {
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

type ArmAngle struct {
	Left  []float64 `json:"left"`
	Right []float64 `json:"right"`
}

func PublishArmAngleInst(params map[string][]float64, url string) error {

	armAngle := ArmAngle{
		Left:  params["left"],
		Right: params["right"],
	}
	// 序列化 armAngle 为 JSON
	jsonData, err := json.Marshal(armAngle)
	if err != nil {
		logs.Errorf("JSON 序列化错误: %v\n", err)
		return err
	}
	logs.Infof("the json data is %s", string(jsonData))
	// 创建 POST 请求
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		logs.Errorf("创建请求错误: %v\n", err)
		return err
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logs.Errorf("发送请求错误: %v\n", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		logs.Infof("arm angle control successfully")
		return nil
	} else {
		return fmt.Errorf("StatusCode is %v", resp.StatusCode)
	}
}

func PublishLeftArmUpInst(url string) error {

	// 创建 HTTP GET 请求
	resp, err := http.Get(url)
	if err != nil {
		logs.Errorf("发送请求失败: %v", err)
		return err
	}
	defer resp.Body.Close()

	// 检查 HTTP 响应状态码
	if resp.StatusCode != http.StatusOK {
		logs.Errorf("请求失败，状态码: %d", resp.StatusCode)
		//body, err1 := io.ReadAll(resp.Body)
		//if err1 != nil {
		//	fmt.Printf("读取响应 Body 时出错: %v\n", err)
		//	return err1
		//}
		//fmt.Println(string(body))
		return err
	}
	logs.Infof("left arm up successfully")
	return nil
}

func PublishLeftArmDownInst(url string) error {
	// 创建 HTTP GET 请求
	logs.Infof(url)
	resp, err := http.Get(url)
	if err != nil {
		logs.Errorf("发送请求失败: %v", err)
		return err
	}
	defer resp.Body.Close()
	// 检查 HTTP 响应状态码
	if resp.StatusCode != http.StatusOK {
		//body, err1 := io.ReadAll(resp.Body)
		//if err1 != nil {
		//	fmt.Printf("读取响应 Body 时出错: %v\n", err)
		//	return err1
		//}
		//logs.Errorf("body : %d", resp.StatusCode)
		//fmt.Println("body is ")
		//fmt.Println(string(body))
		logs.Errorf("请求失败，状态码: %d", resp.StatusCode)
		return err
	}

	logs.Infof("left arm down successfully")
	return nil
}

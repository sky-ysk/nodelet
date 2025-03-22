package lib

import (
	"bytes"
	"encoding/json"
	"fmt"
	"hit.edu/framework/pkg/component-base/logs"
	"net/http"
)

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

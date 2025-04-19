package lib

import (
	"bytes"
	"encoding/json"
	"fmt"
	"hit.edu/framework/pkg/component-base/logs"
	"io"
	"net/http"
	"time"
)

type TaskStateResp struct {
	TaskState int `json:"task_state"`
}

type GoStandByRequest struct {
	TaskType int `json:"task_type"`
}
type GoInitRequest struct {
	TaskType int `json:"task_type"`
}

type GoInitResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}
type GoStandByResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type StartTaskRequest struct {
	TaskType int `json:"task_type"`
}

type StartTaskResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

func PublishTaskState(url string) (TaskStateResp, error) {
	// 创建 HTTP GET 请求
	logs.Infof(url)
	for {
		time.Sleep(time.Second * 1)
		resp, err := http.Get(url)
		if err != nil {
			logs.Errorf("发送请求失败: %v", err)
			return TaskStateResp{}, err
		}
		defer resp.Body.Close()
		// 检查 HTTP 响应状态码
		if resp.StatusCode != http.StatusOK {
			logs.Errorf("请求失败，状态码: %d", resp.StatusCode)
			return TaskStateResp{}, err
		}
		respData, err := io.ReadAll(resp.Body)
		var taskStateResp TaskStateResp
		err = json.Unmarshal(respData, &taskStateResp)
		if err != nil {
			logs.Errorf("unmarshal task state response failed: %v", err)
		}

		logs.Infof("get task state successfully")
		if taskStateResp.TaskState == 0 {
			logs.Infof("state is 0, robot is idle")
			return taskStateResp, nil
		} else {
			logs.Infof("state is %d", taskStateResp.TaskState)
			logs.Infof("waiting ...")
		}
	}

}

func PublishGoStandBy(param int, url string) (GoStandByResponse, error) {

	taskType := GoStandByRequest{
		TaskType: param,
	}

	// 序列化 armAngle 为 JSON
	jsonData, err := json.Marshal(taskType)
	if err != nil {
		logs.Errorf("JSON 序列化错误: %v\n", err)
		return GoStandByResponse{}, err
	}
	logs.Infof("the json data is %s", string(jsonData))
	// 创建 POST 请求
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		logs.Errorf("创建请求错误: %v\n", err)
		return GoStandByResponse{}, err
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logs.Errorf("发送请求错误: %v\n", err)
		return GoStandByResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		logs.Infof("arm go stand by successfully")
		var goStandByResp GoStandByResponse
		respData, err := io.ReadAll(resp.Body)
		err = json.Unmarshal(respData, &goStandByResp)
		if err != nil {
			logs.Errorf("unmarshal go stand by response failed: %v", err)
			return GoStandByResponse{}, err
		}
		return goStandByResp, nil
	} else {
		return GoStandByResponse{}, fmt.Errorf("StatusCode is %v", resp.StatusCode)
	}
}

func PublishStartTask(param int, url string) (StartTaskResponse, error) {

	taskType := StartTaskRequest{
		TaskType: param,
	}

	// 序列化 armAngle 为 JSON
	jsonData, err := json.Marshal(taskType)
	if err != nil {
		logs.Errorf("JSON 序列化错误: %v\n", err)
		return StartTaskResponse{}, err
	}
	logs.Infof("the json data is %s", string(jsonData))
	// 创建 POST 请求
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		logs.Errorf("创建请求错误: %v\n", err)
		return StartTaskResponse{}, err
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logs.Errorf("发送请求错误: %v\n", err)
		return StartTaskResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		logs.Infof("start task successfully")
		var startTaskResponse StartTaskResponse
		respData, err := io.ReadAll(resp.Body)
		err = json.Unmarshal(respData, &startTaskResponse)
		if err != nil {
			logs.Errorf("unmarshal go start task response failed: %v", err)
			return StartTaskResponse{}, err
		}
		return startTaskResponse, nil
	} else {
		return StartTaskResponse{}, fmt.Errorf("StatusCode is %v", resp.StatusCode)
	}
}

func PublishGoInit(param int, url string) (GoInitResponse, error) {
	taskType := GoInitRequest{
		TaskType: param,
	}

	// 序列化 armAngle 为 JSON
	jsonData, err := json.Marshal(taskType)
	if err != nil {
		logs.Errorf("JSON 序列化错误: %v\n", err)
		return GoInitResponse{}, err
	}
	logs.Infof("the json data is %s", string(jsonData))
	// 创建 POST 请求
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		logs.Errorf("创建请求错误: %v\n", err)
		return GoInitResponse{}, err
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logs.Errorf("发送请求错误: %v\n", err)
		return GoInitResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		logs.Infof("go init successfully")
		var initResponse GoInitResponse
		respData, err := io.ReadAll(resp.Body)
		err = json.Unmarshal(respData, &initResponse)
		if err != nil {
			logs.Errorf("unmarshal go start task response failed: %v", err)
			return GoInitResponse{}, err
		}
		return initResponse, nil
	} else {
		return GoInitResponse{}, fmt.Errorf("StatusCode is %v", resp.StatusCode)
	}
}

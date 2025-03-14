package rmf

import (
	"encoding/json"
	"fmt"
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/component-base/logs"
	"io"
	"net/http"
)

type Request struct {
	//
	URL string

	//
	Fleets string

	//
	Params []string

	//
	Name string
}

type Response struct {
	Name   string           `json:"name"`
	Robots map[string]Robot `json:"robots"`
}

type Robot struct {
	Name           string `json:"name"`
	Status         string `json:"status"`
	TaskID         string `json:"task_id"`
	UnixMillisTime int    `json:"unix_millis_time"`
	Location       struct {
		Map string  `json:"map"`
		X   float64 `json:"x"`
		Y   float64 `json:"y"`
		Yaw float64 `json:"yaw"`
	} `json:"location"`
	Battery float64 `json:"battery"`
	Issues  []struct {
		Category string `json:"category"`
		Detail   struct {
		} `json:"detail"`
	} `json:"issues"`
}

// TODO: 实现查询逻辑
func GetDeviceInfo(request Request) (*DeviceInfo, error) {
	// 发送RMF格式请求，查询设备信息
	// TODO: 对RMF接口的封装
	// URL/fleets/{fleet}name
	accessURL := fmt.Sprintf("%s/%s/%s/%s", request.URL, "fleets", request.Fleets, "state")
	//fmt.Println(accessURL)
	client := &http.Client{}
	response, err := client.Get(accessURL)
	if err != nil {
		// TODO: 输出错误信息
		logs.V2().Errorf("Error client get :%v\n", err)
		return nil, err
	}
	// 读取Response
	body, err := io.ReadAll(response.Body)
	if err != nil {
		logs.V2().Errorf("Error reading response body: %v\n", err)
		return nil, err
	}
	// 反序列化JSON数据到结构体
	var resp Response
	err = json.Unmarshal(body, &resp)
	if err != nil {
		logs.V2().Errorf("Error unmarshalling JSON: %v\n", err)
		return nil, err
	}

	// 从Fleet中获取Device对应的Value
	if robot, ok := resp.Robots[request.Name]; ok {

		// 只关心其中部分字段
		// 位置
		x := apis.SubProperty{
			Name:  "x",
			Type:  apis.DoubleType,
			Value: fmt.Sprintf("%.2f", robot.Location.X),
		}

		y := apis.SubProperty{
			Name:  "y",
			Type:  apis.DoubleType,
			Value: fmt.Sprintf("%.2f", robot.Location.Y),
		}

		location := apis.Property{
			Name:        "location",
			Type:        apis.ComposeType,
			SubProperty: []apis.SubProperty{x, y},
		}

		// 电池电量
		battery := apis.Property{
			Name:  "battery",
			Type:  apis.DoubleType,
			Value: fmt.Sprintf("%.2f", robot.Battery),
		}

		// 运行状态
		status := apis.Property{
			Name:  "status",
			Type:  apis.StringType,
			Value: robot.Status,
		}

		// TaskID
		taskID := apis.Property{
			Name:  "task_id",
			Type:  apis.StringType,
			Value: robot.TaskID,
		}

		// 设备信息
		deviceInfo := &DeviceInfo{
			Infos: []apis.Property{
				location, battery, status, taskID,
			},
		}
		return deviceInfo, nil
	}

	return nil, fmt.Errorf("robot %s is not exsited", resp.Name)
}

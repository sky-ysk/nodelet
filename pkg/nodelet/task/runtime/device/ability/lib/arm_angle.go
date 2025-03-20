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

func PublishArmAngleInst(params map[string][]float64, baseUrl string) error {

	url := fmt.Sprintf("%s/api/control/arm_angle", baseUrl)

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

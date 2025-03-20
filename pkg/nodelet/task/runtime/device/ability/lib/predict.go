package lib

import (
	"bytes"
	"encoding/json"
	"fmt"
	"hit.edu/framework/pkg/component-base/logs"
	"io"
	"mime/multipart"
	"net/http"
	"os"
)

type PredictSuccessRes struct {
	Prediction int `json:"prediction"`
}

type PredictFailRes struct {
	Error string `json:"error"`
}

func PublishPredictInst(imagePath, baseUrl string) (PredictSuccessRes, error) {

	// 打开文件
	file, err := os.Open(imagePath)
	if err != nil {
		logs.Errorf("Open file error: %v\n", err)
		return PredictSuccessRes{}, err
	}
	defer file.Close()

	// 创建缓冲区
	var buffer bytes.Buffer
	writer := multipart.NewWriter(&buffer)

	// 添加文件到请求
	part, err := writer.CreateFormFile("file", imagePath)
	if err != nil {
		logs.Errorf("Create form file error: %v\n", err)
		return PredictSuccessRes{}, err
	}
	_, err = io.Copy(part, file)
	if err != nil {
		fmt.Printf("Copy file error: %v\n", err)
		return PredictSuccessRes{}, err
	}

	// 关闭 writer，确保数据写入缓冲区
	err = writer.Close()
	if err != nil {
		fmt.Printf("Close writer error: %v\n", err)
		return PredictSuccessRes{}, err
	}

	url := fmt.Sprintf("%s/predict", baseUrl)
	// 创建 POST 请求
	req, err := http.NewRequest("POST", url, &buffer)
	if err != nil {
		fmt.Printf("Create request error: %v\n", err)
		return PredictSuccessRes{}, err
	}

	// 设置请求头
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Send request error: %v\n", err)
		return PredictSuccessRes{}, err
	}
	defer resp.Body.Close()

	// 读取响应
	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Read response error: %v\n", err)
		return PredictSuccessRes{}, err
	}

	if resp.StatusCode == http.StatusOK {
		var successRes PredictSuccessRes
		err = json.Unmarshal(respData, &successRes)
		if err != nil {
			return PredictSuccessRes{}, err
		}
		return successRes, err

	} else {
		var failRes PredictFailRes
		err = json.Unmarshal(respData, &failRes)
		if err != nil {
			return PredictSuccessRes{}, err
		}
		logs.Errorf("predict failed, the reason is %v", failRes.Error)
		return PredictSuccessRes{}, err
	}
}

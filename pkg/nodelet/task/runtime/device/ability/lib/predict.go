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
type PredictByUrlReq struct {
	Url        string `json:"url"`
	Position   string `json:"position"`
	Type       string `json:"type"`
	Compressed bool   `json:"compressed"`
}

func PublishPredictInst(imagePath, url string) (PredictSuccessRes, error) {

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
	part, err := writer.CreateFormFile("image", imagePath)
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

//
//func main() {
//	// Image file path to test with
//	imagePath := "test_image.jpg" // Replace with your image path
//
//	// API endpoint URL
//	url := "http://localhost:8080/predict" // Adjust port if needed
//
//	// Create buffer for multipart form data
//	var body bytes.Buffer
//	writer := multipart.NewWriter(&body)
//
//	// Open image file
//	file, err := os.Open(imagePath)
//	if err != nil {
//		log.Fatalf("Failed to open image: %v", err)
//	}
//	defer file.Close()
//
//	// Create form file field
//	part, err := writer.CreateFormFile("image", filepath.Base(imagePath))
//	if err != nil {
//		log.Fatalf("Failed to create form file: %v", err)
//	}
//
//	// Copy image content to form
//	_, err = io.Copy(part, file)
//	if err != nil {
//		log.Fatalf("Failed to copy image content: %v", err)
//	}
//
//	// Close writer to finalize multipart form
//	err = writer.Close()
//	if err != nil {
//		log.Fatalf("Failed to close writer: %v", err)
//	}
//
//	// Create HTTP request
//	req, err := http.NewRequest("POST", url, &body)
//	if err != nil {
//		log.Fatalf("Failed to create request: %v", err)
//	}
//
//	// Set content type with boundary
//
//	// Send request
//	client := &http.Client{}
//	resp, err := client.Do(req)
//	if err != nil {
//		log.Fatalf("Failed to send request: %v", err)
//	}
//	defer resp.Body.Close()
//
//	// Read response
//	respBody, err := io.ReadAll(resp.Body)
//	if err != nil {
//		log.Fatalf("Failed to read response: %v", err)
//	}
//
//	// Print results
//	log.Printf("Status: %d", resp.StatusCode)
//	log.Printf("Response: %s", string(respBody))
//}

func PublishPredictByUrlInst(compressed bool, cameraUrl string, position string, imageType string, url string) (PredictSuccessRes, error) {

	// 序列化 predictByUrl 为 JSON
	predictByUrl := PredictByUrlReq{
		Url:        cameraUrl,
		Position:   position,
		Type:       imageType,
		Compressed: compressed,
	}
	jsonData, err := json.Marshal(predictByUrl)
	if err != nil {
		logs.Errorf("JSON 序列化错误: %v\n", err)
		return PredictSuccessRes{}, err
	}
	logs.Infof("the json data is %s", string(jsonData))
	// 创建 POST 请求
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		logs.Errorf("创建请求错误: %v\n", err)
		return PredictSuccessRes{}, err
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logs.Errorf("发送请求错误: %v\n", err)
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

package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"hit.edu/framework/pkg/component-base/logs"

	utils "hit.edu/framework/pkg/nodelet/registry/Utils"
)

func main() {
	// url := "http://localhost:8919/download?filename=server"
	// savePath := "./"

	// err := utils.DownloadFile(url, savePath)
	// if err != nil {
	// 	fmt.Println("Download failed:", err)
	// } else {
	// 	fmt.Println("Download successful!")
	// }

	url := "http://172.150.0.24:8919/upload"
	filePath := "/home/public/goprojects/ysk-1110/reference-626/test/registry/clientUpload.go"
	fmt.Println("aaaaaUploading file:", filePath)
	err := utils.UploadFile(filePath, "v1.0.0", url)
	fmt.Println("bbbbUploading file:", filePath)
	if err != nil {
		fmt.Println("Upload failed:", err)
	} else {
		fmt.Println("Upload successful!")
	}

	// url := "http://localhost:8888/upload"
	// filePath := "/home/public/goprojects/Combine-ysk-0102/tmp/ForUploadServerRegistry/test.txt"
	// err := utils.UploadFile(filePath, "v1.0.0", url)
	// if err != nil {
	// 	fmt.Println("Upload failed:", err)
	// } else {
	// 	fmt.Println("Upload successful!")
	// }

	// url := "http://localhost:8919/upload?filename=test"
	// filePath1 := "/home/public/registry/Registry/cmd/registry/test"
	// err := utils.Traverse(filePath1, url)
	// if err != nil {
	// 	fmt.Println("Upload failed:", err)
	// } else {
	// 	fmt.Println("Upload successful!")
	// }

	// url := "http://localhost:8081/download?filename=test"

	// // 创建 GET 请求
	// req, err := http.NewRequest("GET", url, nil)
	// if err != nil {
	// 	log.Fatal("Failed to create request:", err)
	// }

	// // 设置请求头（这里是示例）
	// req.Header.Set("FileType", "folder")

	// // 发送请求
	// client := &http.Client{}
	// resp, err := client.Do(req)
	// if err != nil {
	// 	log.Fatal("Request failed:", err)
	// }
	// defer resp.Body.Close()

	// filePath := "./"
	// http.HandleFunc("/receive", func(w http.ResponseWriter, r *http.Request) {
	// 	utils.ReceiveDir(w, r, filePath)
	// })

	// port := ":8080"
	// log.Printf("Server is running on port %s", port)
	// if err := http.ListenAndServe(port, nil); err != nil {
	// 	log.Fatalf("Failed to start server: %v", err)
	// }
	// log.Printf("end waiting for close")

	// fileName := "yolo_projects_asy"
	// filePath := "./"
	// DownloadFileDir(fileName, filePath)
}

func DownloadFileDir(fileName, savePath string) (string, error) {
	// 1. 创建HTTP服务器
	server := &http.Server{Addr: ":8920"}
	done := make(chan bool)

	// 2. 设置处理函数（使用闭包捕获savePath）
	http.HandleFunc("/receive", func(w http.ResponseWriter, r *http.Request) {
		utils.ReceiveDir(w, r, savePath)

		// 检查传输完成信号
		if fileType := r.Header.Get("FileType"); fileType == "completion" {
			logs.Info("File transfer completed")
			close(done)
		}
	})

	// 3. 在goroutine中启动服务器
	go func() {
		log.Println("启动文件接收服务...")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("服务器启动失败: %v", err)
		}
	}()

	// 4. 短时等待确保服务器已启动
	time.Sleep(3 * time.Second)

	// 5. 发送下载请求
	downloadURL := "http://localhost:8919/download?filename=" + fileName

	// 创建 GET 请求
	req, err := http.NewRequest("GET", downloadURL, nil)
	if err != nil {
		log.Fatal("Failed to create request:", err)
	}

	// 设置请求头（这里是示例）
	req.Header.Set("FileType", "folder")
	// req.Header.Set("User-Agent", "MyClient/1.0")

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("Request failed:", err)
	}
	defer resp.Body.Close()

	// 6. 等待传输完成或超时
	select {
	case <-done:
		logs.Info("Prepare to shut down the server...")
		server.Shutdown(context.Background())
		return "文件接收成功并保存至 " + savePath, nil
	case <-time.After(5 * time.Minute):
		server.Shutdown(context.Background())
		return "", errors.New("文件接收超时")
	}
}

func hello() {
	fmt.Println("Hello, World!")
}

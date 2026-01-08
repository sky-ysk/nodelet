package main

import (
	"fmt"

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
	filePath := "/home/public/goprojects/ysk-1106/reference-626/test/registry/clientUpload.go"
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

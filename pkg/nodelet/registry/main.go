package main

import (
	utils "hit.edu/framework/pkg/nodelet/registry/Utils"
	"log"
	"net/http"
)

func main() {
	//url := "http://localhost:8081/download?filename=test.txt&tag=v1.2.0"
	//savePath := "./test"
	//
	//err := utils.DownloadFile(url, savePath)
	//if err != nil {
	//	fmt.Println("Download failed:", err)
	//} else {
	//	fmt.Println("Download successful!")
	//}

	//url := "http://localhost:8081/upload"
	//filePath := "./downloads/test.txt"
	//err := utils.UploadFile(filePath, "v1.0.0", url)
	//if err != nil {
	//	fmt.Println("Upload failed:", err)
	//} else {
	//	fmt.Println("Upload successful!")
	//}

	//url := "http://localhost:8081/upload?filename=downloads"
	//filePath := "./downloads"
	//err := utils.Traverse(filePath, url)
	//if err != nil {
	//	fmt.Println("Upload failed:", err)
	//} else {
	//	fmt.Println("Upload successful!")
	//}

	//url := "http://localhost:8081/download?filename=1.txt"
	//
	//// 发送 GET 请求
	//resp, err := http.Get(url)
	//if err != nil {
	//	fmt.Println("Request failed:", err)
	//	return
	//}
	//defer resp.Body.Close() // 确保响应体被关闭

	filePath := "./test"
	http.HandleFunc("/receive", func(w http.ResponseWriter, r *http.Request) {
		utils.ReceiveDir(w, r, filePath)
	})

	port := ":8080"
	log.Printf("Server is running on port %s", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

}

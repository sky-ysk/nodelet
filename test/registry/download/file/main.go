package main

import (
	"fmt"

	utils "hit.edu/framework/pkg/nodelet/registry/Utils"
)

func main() {
	url := "http://120.220.95.189:48121/download?filename=clientUpload.go"
	savePath := "./"

	err := utils.DownloadFile(url, savePath)
	if err != nil {
		fmt.Println("Download failed:", err)
	} else {
		fmt.Println("Download successful!")
	}

}

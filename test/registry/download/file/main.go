package main

import (
	"fmt"

	utils "hit.edu/framework/pkg/nodelet/registry/Utils"
)

func main() {
	url := "http://localhost:8919/download?filename=server"
	savePath := "./"

	err := utils.DownloadFile(url, savePath)
	if err != nil {
		fmt.Println("Download failed:", err)
	} else {
		fmt.Println("Download successful!")
	}

}

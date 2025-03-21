package lib

import (
	"fmt"
	"hit.edu/framework/pkg/component-base/logs"
	"os"
	"path/filepath"
	"testing"
)

func TestPublishPredictInst(t *testing.T) {
	logs.Init("test")

	// 获取当前测试文件的目录
	testDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("无法获取当前工作目录: %v", err)
	}

	// 构建相对路径
	imagePath := filepath.Join(testDir, "testpics", "orange.jpg")
	url := "http://192.168.8.165:48479"
	resp, err := PublishPredictInst(imagePath, url)
	if err != nil {
		logs.Errorf("error publish predict inst %v", err)
	} else {
		logs.Infof("predict result is %v", resp)
		fmt.Println(resp)
	}
}

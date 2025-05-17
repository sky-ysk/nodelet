package lib

import (
	"fmt"
	"hit.edu/framework/pkg/component-base/logs"
	"testing"
	"time"
)

func TestPublishDetectWorkpieceInst(t *testing.T) {
	logs.Init("test")
	ip := "http://172.130.0.61:48227"
	api := "/api/task/detect"
	url := fmt.Sprintf("%s%s", ip, api)
	taskId, err := PublishDetectWorkpieceInst(url)
	if err != nil {
		logs.Errorf("err is %v", err)
	}

	time.Sleep(time.Second * 10)
	tr, err := GetTaskStatus(taskId, "http://172.130.0.61:8080")
	if err != nil {
		logs.Errorf("tr is %v", tr)
	}
}

// go test -run TestPublishDetectPositionInst -v
func TestPublishDetectPositionInst(t *testing.T) {
	logs.Init("test")
	url1 := "http://192.168.8.197:60489/api/task/detect" // 把ip和port填写完整
	taskId, err := PublishDetectPositionInst(url1)
	if err != nil {
		logs.Errorf("PublishDetectPositionInst fail %s", err.Error())
		return
	}
	logs.Infof("PublishDetectPositionInst successfully")
	url2 := "http://192.168.8.197:8080" // 填写ip和端口
	for {
		time.Sleep(time.Second * 2)
		response, err := GetTaskStatus(taskId, url2)
		if err != nil {
			logs.Errorf("GetTaskStatus fail")
			return
		}
		logs.Infof("GetTaskStatus successfully")
		switch response.State {
		case Running:
			logs.Infof("DetectPosition is running")
		case Finished:
			logs.Infof("DetectPosition is finished")
			worldPoints := response.Payload
			logs.Infof("World Point is: [%v]", worldPoints)
			return
		case Error:
			logs.Infof("DetectPosition is error")
			errMessage := response.Message
			logs.Errorf("err message is %s", errMessage)
			return
		}
	}

}

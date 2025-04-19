package lib

import (
	"hit.edu/framework/pkg/component-base/logs"
	"testing"
	"time"
)

func TestPublishDetectPositionInst(t *testing.T) {
	logs.Init("test")
	url1 := "/api/task/detect" // 把ip和port填写完整
	taskId, err := PublishDetectPositionInst(url1)
	if err != nil {
		logs.Errorf("PublishDetectPositionInst fail")
		return
	}
	logs.Infof("PublishDetectPositionInst successfully")
	url2 := "" // 填写ip和端口
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

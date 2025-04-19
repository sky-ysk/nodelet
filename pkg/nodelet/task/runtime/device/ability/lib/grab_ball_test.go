package lib

import (
	"hit.edu/framework/pkg/component-base/logs"
	"testing"
	"time"
)

func TestPublishGrabBallInst(t *testing.T) {
	logs.Init("test")
	url1 := "/api/task/grab_ball" // 把ip和port填写完整
	worldPoints := [][]float64{
		{},
		{},
	}
	taskId, err := PublishGrabBallInst(worldPoints, url1)
	if err != nil {
		logs.Errorf("PublishGrabBallInst fail")
		return
	}
	logs.Infof("PublishGrabBallInst successfully")
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
			logs.Infof("GrabBall is running")
		case Finished:
			logs.Infof("GrabBall is finished")
			return
		case Error:
			logs.Infof("GrabBall is error")
			errMessage := response.Message
			logs.Errorf("err message is %s", errMessage)
			return
		}
	}
}

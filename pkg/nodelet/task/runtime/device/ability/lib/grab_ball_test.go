package lib

import (
	"hit.edu/framework/pkg/component-base/logs"
	"testing"
	"time"
)

// go test -run TestPublishGrabBallInst -v
func TestPublishGrabBallInst(t *testing.T) {
	logs.Init("test")
	url1 := "http://192.168.8.197:49761/api/task/grab_ball" // 把ip和port填写完整
	worldPoints := [][]float64{
		{835.1900024414062, -141.73117065429688, 306.4671630859375},
	}
	taskId, err := PublishGrabBallInst(worldPoints, url1)
	if err != nil {
		logs.Errorf("PublishGrabBallInst fail")
		return
	}
	logs.Infof("PublishGrabBallInst successfully")
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

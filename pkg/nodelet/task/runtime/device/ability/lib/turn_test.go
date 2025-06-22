package lib

import (
	"fmt"
	"hit.edu/framework/pkg/component-base/logs"
	"testing"
	"time"
)

func TestPublishTurnInst(t *testing.T) {
	logs.Init("test")
	ip := "http://172.130.0.61:51759"
	api := "/api/task/turn"
	url := fmt.Sprintf("%s%s", ip, api)
	taskId, err := PublishTurnInst("table", url)
	if err != nil {
		logs.Errorf("err is %v", err)
	}

	time.Sleep(time.Second * 10)
	tr, err := GetTaskStatus(taskId, "http://172.130.0.61:8080")
	if err != nil {
		logs.Errorf("tr is %v", tr)
	}
}

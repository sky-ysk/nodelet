package lib

import (
	"hit.edu/framework/pkg/component-base/logs"
	"testing"
)

func TestPublishPredictInst(t *testing.T) {
	logs.Init("test")
	imagePath := ""
	url := "http://127.0.0.1:8123"
	resp, err := PublishPredictInst(imagePath, url)
	if err != nil {
		logs.Errorf("error publish predict inst %v", err)
	} else {
		logs.Infof("predict result is %v", resp)
	}
}

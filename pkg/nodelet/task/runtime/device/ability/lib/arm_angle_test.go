package lib

import (
	"hit.edu/framework/pkg/component-base/logs"
	"testing"
)

func TestPublishArmAngleInst(t *testing.T) {
	logs.Init("test")
	var params map[string][]float64 = make(map[string][]float64)
	params["left"] = []float64{-1.221, 0.0872, 0, 0, 0, 0, 0}
	params["right"] = []float64{0, 0, 0, 0, 0, 0, 0}
	url := ""
	err := PublishArmAngleInst(params, url)
	if err != nil {
		logs.Errorf("error publish ArmAngle inst %v", err)
	}
}

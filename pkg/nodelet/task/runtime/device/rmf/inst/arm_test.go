package inst

import (
	"hit.edu/framework/pkg/component-base/logs"
	"testing"
)

func TestArm(t *testing.T) {
	logs.Infof("[Test] testing NewArmInst....\n")
	// 构建几个示例
	robotName := "mobileArm"
	group := "tinyRobot"
	dest := "arm1-1-1"
	armMap := map[string]float64{
		"x":     297.787,
		"y":     -325.616,
		"z":     670.058,
		"rx":    -89.181,
		"ry":    -0.486,
		"rz":    -137.505,
		"speed": 30.000,
	}
	armOperator := []string{"grab", "loosen", "go_up_and_down"}
	robotId := "105"
	for _, operator := range armOperator {
		JsonStr, err := NewArmInst(robotName, group, dest, armMap, robotId, operator)
		if err != nil {
			logs.Fatal("call NewArmInst fail: %v\n", err)
		}
		t.Logf("[test] Arm instruction is :\n")
		t.Logf("%s\n", JsonStr)
	}

}

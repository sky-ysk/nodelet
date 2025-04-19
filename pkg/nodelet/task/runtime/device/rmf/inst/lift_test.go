package inst

import (
	"hit.edu/framework/pkg/component-base/logs"
	"testing"
)

func TestLift(t *testing.T) {
	logs.Infof("[Test] testing NewArmInst....\n")
	// 构建几个示例
	robotName := "patrolRobot"
	group := "tinyRobot"
	robotId := "107"
	height := "low"
	JsonStr, err := NewLiftInst(robotName, group, height, robotId)
	if err != nil {
		logs.Fatal("call NewLiftInst fail: %v\n", err)
	}
	t.Logf("[test] Lift instruction is :\n")
	t.Logf("%s\n", JsonStr)

}

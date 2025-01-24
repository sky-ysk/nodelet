package inst

import (
	"log"
	"testing"
)

func TestMove(t *testing.T) {
	// 构建几个示例
	robot := "transferRobot"
	group := "tinyRobot"
	dest := "R201"
	orientation := -3.12
	isDock := false
	timeout := 100
	jsonstr, err := NewMoveInst(robot, group, dest, orientation, isDock, timeout)
	if err != nil {
		log.Fatal(err)
	}
	t.Logf("move instruction is :\n")
	t.Logf("%s\n", jsonstr)
}

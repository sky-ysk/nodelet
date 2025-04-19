package ability

import (
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/component-base/logs"
	"hit.edu/framework/pkg/nodelet/task/runtime/device/ability/lib"
	"hit.edu/framework/pkg/nodelet/task/runtime/device/ability/manager"
	"time"
)

func PublishAbilityInst(inst string, device *apis.Device) (string, error) {
	switch inst {
	case "DetectPosition":
		url := device.Status.Abilities
		taskId, err := lib.PublishDetectPositionInst()
		if err != nil {
			logs.Errorf("[DEVICE RUNTIME] PublishDetectPosition fail")
			return "", err
		}
		logs.Infof("[DEVICE RUNTIME] Task ID is %s", taskId)
		return taskId, nil
	case "GrabBall":
		taskId, err := lib.PublishGrabBallInst()
	}
}

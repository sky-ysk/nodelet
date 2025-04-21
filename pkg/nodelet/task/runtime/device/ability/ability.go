package ability

import (
	"fmt"
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/component-base/logs"
	"hit.edu/framework/pkg/nodelet/task/runtime/device/ability/lib"
	"hit.edu/framework/pkg/utils/value"
)

// PublishAbilityInst 根据传入的指令来发布对应的指令
func PublishAbilityInst(inst string, device *apis.Device, params []apis.Value, engine *value.Engine, runtime *apis.Runtime,
	action *apis.Action) (string, error) {
	// 构造URL
	ability := GetAbilityByService(device, inst)
	ip := device.Status.Abilities[ability].Services[inst].Ip
	port := device.Status.Abilities[ability].Services[inst].Port
	api := device.Status.Abilities[ability].Services[inst].Interface
	url := fmt.Sprintf("http://%s:%s%s", *ip, *port, *api)

	switch inst {
	case "DetectPosition":
		taskId, err := lib.PublishDetectPositionInst(url)
		if err != nil {
			logs.Errorf("[DEVICE RUNTIME] PublishDetectPosition fail")
			return "", err
		}
		logs.Infof("[DEVICE RUNTIME] Task ID is %s", taskId)
		return taskId, nil
	case "GrabBall":
		// 获取参数
		var worldPoints [][]float64
		for _, param := range params {
			if param.Name == "worldPoints" {
				if param.Type == apis.ConstData {
					worldPoints = lib.GetWorldPoints(param)
				} else if param.Type == apis.LocalData {
					logs.Infof("[DEVICE RUNTIME] WorldPoints param is %v", param)
					logs.Infof("[DEVICE RUNTIME] WorldPoints param is %v", action)
					worldPointValue, err := engine.ExtractLocalValue(&param, *action)
					if err != nil {
						logs.Errorf("[DEVICE RUNTIME] Engine ExtractLocalValue fail, %s ", err.Error())
						return "", err
					}
					logs.Infof("[DEVICE RUNTIME] WorldPoints param is %v", worldPointValue)
					worldPoints = lib.GetWorldPoints(*worldPointValue)
				}
			}
		}

		taskId, err := lib.PublishGrabBallInst(worldPoints, url)
		if err != nil {
			logs.Errorf("[DEVICE RUNTIME] PublishGrabBallInst fail")
			return "", err
		}
		logs.Infof("[DEVICE RUNTIME] Task ID is %s", taskId)
		return taskId, nil
	case "":
	default:
		logs.Errorf("[DEVICE RUNTIME] Unknown Ability")
		return "", fmt.Errorf("unknow ability")
	}
	return "", nil
}

// GetAbilityByService 通过service得到相应的能力名称
func GetAbilityByService(device *apis.Device, service string) string {
	for nameAbility, a := range device.Status.Abilities {
		for nameService, _ := range a.Services {
			if nameService == service {
				return nameAbility
			}
		}
	}
	return ""
}

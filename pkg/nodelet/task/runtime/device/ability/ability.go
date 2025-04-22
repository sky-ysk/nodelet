package ability

import (
	"fmt"
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/component-base/logs"
	"hit.edu/framework/pkg/nodelet/task/runtime/device/ability/lib"
	"hit.edu/framework/pkg/utils/value"
)

type AbilityStrategy interface {
	Execute(url string, params []apis.Value, engine *value.Engine, runtime *apis.Runtime, action *apis.Action) (string, error)
}
type AbilityContext struct {
	strategy AbilityStrategy
}

func (ac *AbilityContext) SetStrategy(strategy AbilityStrategy) {
	ac.strategy = strategy
}

func (ac *AbilityContext) Execute(url string, params []apis.Value, engine *value.Engine, runtime *apis.Runtime, action *apis.Action) (string, error) {
	return ac.strategy.Execute(url, params, engine, runtime, action)
}

// PublishAbilityInst 根据传入的指令来发布对应的指令
func PublishAbilityInst(inst string, device *apis.Device, params []apis.Value, engine *value.Engine, runtime *apis.Runtime, action *apis.Action) (string, error) {
	// 构造URL
	ability := GetAbilityByService(device, inst)
	ip := device.Status.Abilities[ability].Services[inst].Ip
	port := device.Status.Abilities[ability].Services[inst].Port
	api := device.Status.Abilities[ability].Services[inst].Interface
	url := fmt.Sprintf("http://%s:%s%s", *ip, *port, *api)

	var strategy AbilityStrategy
	switch inst {
	case "DetectPosition":
		strategy = &lib.DetectPositionStrategy{}
	case "GrabBall":
		strategy = &lib.GrabBallStrategy{}
	case "Download":
		strategy = &lib.DownloadModelStrategy{}
	case "PredictByUrl":
		strategy = &lib.PreByUrlStrategy{}
	default:
		logs.Errorf("[DEVICE RUNTIME] Unknown Ability")
		return "", fmt.Errorf("unknow ability")
	}

	context := &AbilityContext{}
	context.SetStrategy(strategy)
	return context.Execute(url, params, engine, runtime, action)

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

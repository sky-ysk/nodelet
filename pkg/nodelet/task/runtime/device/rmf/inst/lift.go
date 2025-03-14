package inst

import (
	"encoding/json"
	"fmt"
	"hit.edu/framework/pkg/component-base/logs"
)

type LiftWorker interface {
	// NewLiftInst 构建go_up_and_down指令【和inst组合使用】
	NewLiftInst(robotName string, group string, height string, robotId string) (string, error)
}

func NewLiftInst(robotName string, group string, height string, robotId string) (string, error) {
	if height == "" {
		logs.Errorf("height is empty\n")
		return "", fmt.Errorf("height is empty")
	}
	var phases []Phase
	// Phase1放入GoToPlace和Dock
	phase1 := Phase{
		Activity: Activity{
			Category: "sequence",
			Description: ActivityDesc{
				Activities: make([]AcitivityInterface, 0),
			},
		},
	}
	liftParam := LiftParamJSON{
		Action: "go_up_and_down",
		Param: LiftParam{
			Height: height,
		},
	}

	liftParamStr, err := json.Marshal(liftParam)
	if err != nil {
		logs.Errorf("Failed to marshal grabParam")
	}

	lift := Lift{
		Category: "device_action",
		Description: ArmDesc{
			DeviceID:     robotId,
			ParamJSONStr: string(liftParamStr),
		},
	}
	phase1.Activity.Description.Activities = append(phase1.Activity.Description.Activities, lift)
	phases = append(phases, phase1)

	inst := Instruction{
		Type:  "robot_task_request",
		Fleet: group,
		Robot: robotName,
		Request: Request{
			Category: "compose",
			Description: RequestDesc{
				Category: "go_up_and_down",
				Phases:   phases,
			},
			UnixMillisEarliestStartTime: 0,
			Priority: Priority{
				Type:  "binary",
				Value: 0,
			},
		},
	}
	instStr, err := json.Marshal(inst)
	if err != nil {
		return "", fmt.Errorf("marshal instruction failed: %v", err)
	}
	// TODO: 格式检查
	return string(instStr), nil
}

type LiftParamJSON struct {
	Action string    `json:"action"`
	Param  LiftParam `json:"param"`
}

type LiftParam struct {
	Height string `json:"height"`
}

type Lift struct {
	Category    string  `json:"category"`
	Description ArmDesc `json:"description"`
}

type LiftDesc struct {
	DeviceID     string `json:"device_id"`
	ParamJSONStr string `json:"param_json_str"`
}

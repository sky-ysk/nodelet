package inst

import (
	"encoding/json"
	"fmt"
	"hit.edu/framework/pkg/component-base/logs"
)

/*
	grab主要负责两部分：发布抓取命令和http返回消息的解析
	1.发布抓取命令通过grab.go和inst.go协作拼接完成
	2.http返回消息解析一共有三种successful、badRequest【?】和ValidationError
*/

type ArmWorker interface {
	// NewArmInst 构建arm指令【和inst组合使用】
	NewArmInst(robotName string, group string, dest string, armMap map[string]float64, robotId string) (string, error)
}

func NewArmInst(robotName string, group string, dest string, armMap map[string]float64, robotId string, operator string) (string, error) {
	if dest == "" {
		return "", fmt.Errorf("empty dest")
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
	armParam := GrabParamJSON{
		Action: operator,
		Param: ArmParam{
			Position: dest,
			X:        armMap["x"],
			Y:        armMap["y"],
			Z:        armMap["z"],
			Rx:       armMap["rx"],
			Ry:       armMap["ry"],
			Rz:       armMap["rz"],
			Speed:    armMap["speed"],
		},
	}

	armParamStr, err := json.Marshal(armParam)
	if err != nil {
		logs.Errorf("Failed to marshal grabParam")
	}

	arm := Arm{
		Category: "device_action",
		Description: ArmDesc{
			DeviceID:     robotId,
			ParamJSONStr: string(armParamStr),
		},
	}
	phase1.Activity.Description.Activities = append(phase1.Activity.Description.Activities, arm)
	phases = append(phases, phase1)

	inst := Instruction{
		Type:  "robot_task_request",
		Fleet: group,
		Robot: robotName,
		Request: Request{
			Category: "compose",
			Description: RequestDesc{
				Category: operator,
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

/*
	构造指令所需的结构体
*/

type GrabParamJSON struct {
	Action string   `json:"action"`
	Param  ArmParam `json:"param"`
}

type ArmParam struct {
	Position string  `json:"position"`
	X        float64 `json:"x"`
	Y        float64 `json:"y"`
	Z        float64 `json:"z"`
	Rx       float64 `json:"rx"`
	Ry       float64 `json:"ry"`
	Rz       float64 `json:"rz"`
	Speed    float64 `json:"speed"`
}
type Arm struct {
	Category    string  `json:"category"`
	Description ArmDesc `json:"description"`
}

type ArmDesc struct {
	DeviceID     string `json:"device_id"`
	ParamJSONStr string `json:"param_json_str"`
}

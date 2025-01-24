package inst

import (
	"encoding/json"
	"fmt"
	"hit.edu/framework/pkg/component-base/logs"
)

// 移动任务目前支持
//    go_to_place
//    wait_for(只允许放到第二个Phase)
//    dock 特殊点指令

type Move struct {
	// 移动的目标位置
	//   RMF移动的参数为目标点位
	dest string

	// 目标角度
	orientation float32

	// 是否需要支持Dock
	isDock bool

	// RMF等待时长，用于锁定机器人，避免机器人被其他任务中断
	// 单位为秒
	Timeout int
}

type MoveWorker interface {
	// NewMoveInst 构建move指令
	NewMoveInst(robot string, group string, dest string, orientation float64, isDock bool, timeout int) (string, error)
}

func NewMoveInst(robot string, group string, dest string, orientation float64, isDock bool, timeout int) (string, error) {
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

	if dest == "" {
		return "", fmt.Errorf("empty dest")
	}

	goToPlace := GoToPlace{
		Category: "go_to_place",
		Description: GoToPlaceDesc{
			Waypoint:    dest,
			Orientation: orientation,
		},
	}

	phase1.Activity.Description.Activities = append(phase1.Activity.Description.Activities, goToPlace)

	if isDock {
		paramJson := ParamJSON{
			Action: "dock",
			Param: Param{
				TargetDock: dest,
			},
		}
		paramJsonStr, err := json.Marshal(paramJson)
		if err != nil {
			logs.Errorf("Failed to marshal paramJson")
		}

		dock := Dock{
			Category: "device_action",
			Description: DockDesc{
				DeviceID:     robot,
				ParamJSONStr: string(paramJsonStr),
			},
		}
		phase1.Activity.Description.Activities = append(phase1.Activity.Description.Activities, dock)
	}
	phases = append(phases, phase1)

	if timeout >= 0 {
		// Phase2 放入Wait函数
		phase2 := Phase{
			Activity: Activity{
				Category: "sequence",
				Description: ActivityDesc{
					Activities: make([]AcitivityInterface, 0),
				},
			},
		}

		wait := Wait{
			Category:    "wait_for",
			Description: WaitDesc{Duration: timeout * 1000},
		}
		phase2.Activity.Description.Activities = append(phase2.Activity.Description.Activities, wait)
		phases = append(phases, phase2)
	}

	inst := Instruction{
		Type:  "robot_task_request",
		Fleet: group,
		Robot: robot,
		Request: Request{
			Category: "compose",
			Description: RequestDesc{
				Category: "move",
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

// RMF go_to_place 指令
type GoToPlace struct {
	Category    string        `json:"category"`
	Description GoToPlaceDesc `json:"description"`
}

type GoToPlaceDesc struct {
	Waypoint    string  `json:"waypoint"`
	Orientation float64 `json:"orientation"`
}

// RMF wait_for 指令
type Wait struct {
	Category    string   `json:"category"`
	Description WaitDesc `json:"description"`
}

type WaitDesc struct {
	Duration int `json:"duration"`
}

// RMF Dock指令
type Dock struct {
	Category    string   `json:"category"`
	Description DockDesc `json:"description"`
}

type DockDesc struct {
	DeviceID     string `json:"device_id"`
	ParamJSONStr string `json:"param_json_str"`
}

type ParamJSON struct {
	Action string `json:"action"`
	Param  Param  `json:"param"`
}

type Param struct {
	TargetDock string `json:"target_dock"`
}

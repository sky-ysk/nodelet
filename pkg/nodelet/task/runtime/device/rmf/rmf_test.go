package rmf

import (
	"encoding/json"
	"fmt"
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/component-base/logs"
	"testing"
)

// move指令的测试
func TestRmfPublishMoveInst(t *testing.T) {
	// 构造示例
	logs.Infof("start to test move inst......\n")
	device := apis.Device{
		Spec: apis.DeviceSpec{
			AccessMethod: apis.AccessMethod{
				Type:  apis.AccessByRmf,
				URL:   "http://192.168.1.225:8000",
				Group: "tinyRobot",
				Alias: "patrolRobot",
			},
			ExpectedProperties: map[string]apis.Property{
				"dest":        apis.Property{Value: "R203"},
				"orientation": apis.Property{Value: "-3.12"},
				"dock":        apis.Property{Value: "false"},
			},
			Desc: apis.DeviceDesc{
				Label: []string{"Move"},
			},
			Name: "patrolRobot",
		},
	}
	abilityList := []string{"Move"}
	for _, ability := range abilityList {
		taskId, err := PublishAbilityInstruction(device, ability)
		if err != nil {
			t.Fatal(err)
		}
		t.Log(taskId)
		logs.Infof("taskId: %s\n", taskId)
	}
}

// grab指令的测试
func TestRmfPublishArmInst(t *testing.T) {
	logs.Infof("start to test grab inst......\n")
	// 构造示例
	device := apis.Device{
		Spec: apis.DeviceSpec{
			AccessMethod: apis.AccessMethod{
				Type:  apis.AccessByRmf,
				URL:   "http://192.168.1.225:8000",
				Group: "tinyRobot",
				Alias: "fixedArm",
			},
			ExpectedProperties: map[string]apis.Property{
				"x":     apis.Property{Value: "597.995"},
				"y":     apis.Property{Value: "191.819"},
				"z":     apis.Property{Value: "385.029"},
				"rx":    apis.Property{Value: "178.406"},
				"ry":    apis.Property{Value: "0.329"},
				"rz":    apis.Property{Value: "-136.163"},
				"speed": apis.Property{Value: "50.000"},
				"dest":  apis.Property{Value: "arm2belt01"},
			},
			Desc: apis.DeviceDesc{
				Label: []string{"Grab", "Loosen"},
			},
			Name: "106",
		},
	}
	abilityList := []string{"Grab", "Loosen"}
	for _, ability := range abilityList {
		taskId, err := PublishAbilityInstruction(device, ability)
		if err != nil {
			t.Fatal(err)
		}
		t.Log(taskId)
	}
}

func TestRmfPublishLiftInst(t *testing.T) {
	logs.Infof("start to test lift inst......\n")
	device := apis.Device{
		Spec: apis.DeviceSpec{
			AccessMethod: apis.AccessMethod{
				Type:  apis.AccessByRmf,
				URL:   "http://192.168.1.225:8000",
				Group: "tinyRobot",
				Alias: "patrolRobot",
			},
			ExpectedProperties: map[string]apis.Property{
				"height": apis.Property{Value: "low"},
			},
			Desc: apis.DeviceDesc{
				Label: []string{"Lift"},
			},
			Name: "107",
		},
	}
	abilityList := []string{"Lift"}
	for _, ability := range abilityList {
		taskId, err := PublishAbilityInstruction(device, ability)
		if err != nil {
			t.Fatal(err)
		}
		t.Log(taskId)
	}
}

func TestPublishCancelTaskInstruction(t *testing.T) {
	// 保存旧的标准输入
	logs.Infof("start to test cancel task inst......\n")
	device := apis.Device{
		Spec: apis.DeviceSpec{
			AccessMethod: apis.AccessMethod{
				Type: apis.AccessByRmf,
				URL:  "http://192.168.1.225:8000",
			},
		},
	}
	var taskId string
	taskId = "d3b06ef4-96c7-4302-a9b8-a77987d8014c"
	fmt.Println(taskId)
	isSuccessful, err := PublishCancelTaskInstruction(device, taskId)
	if err != nil {
		t.Error(err)
		logs.Errorf("PublishCancelTaskInstruction err: %s\n", err.Error())
	}
	if !isSuccessful {
		logs.Errorf("PublishCancelTaskInstruction is not successful: %s \n", taskId)
		t.Error("PublishCancelTaskInstruction is not successful\n")
	}
	logs.Info("publish cancel task inst is successful\n")
}

func TestPublishCancelPhaseInstruction(t *testing.T) {
	logs.Infof("start to test cancel phase inst......\n")
	device := apis.Device{
		Spec: apis.DeviceSpec{
			AccessMethod: apis.AccessMethod{
				Type: apis.AccessByRmf,
				URL:  "http://192.168.1.225:8000",
			},
		},
	}
	var taskId string
	taskId = "1a828a92-44da-4865-97ae-634eeb0ebae3"
	fmt.Println(taskId)
	var phaseId int = 2
	fmt.Println(phaseId)

	isSuccessful, err := PublishCancelPhaseInstruction(device, phaseId, taskId)
	if err != nil {
		t.Error(err)
		logs.Errorf("PublishCancelPhaseInstruction err: %s\n", err.Error())
	}
	if !isSuccessful {
		logs.Errorf("PublishCancelPhaseInstruction is not successful: %s \n", taskId)
		t.Error("PublishCancelPhaseInstruction is not successful\n")
	}
	logs.Info("publish cancel phase inst is successful\n")
}

func TestGetTaskState(t *testing.T) {
	//logs.Infof("start to test get task state......\n")
	//device := apis.Device{
	//	Spec: apis.DeviceSpec{
	//		AccessMethod: apis.AccessMethod{
	//			Type: apis.AccessByRmf,
	//			URL:  "http://192.168.1.225:8000",
	//		},
	//	},
	//}
	//var taskId string
	//fmt.Println(taskId)
	//taskStateResponse, err := GetTaskState(device, taskId)
	var taskStateResponse TaskStateSuccessResponse
	jsonData := `{
  "booking": {
    "id": "string",
    "unix_millis_earliest_start_time": 0,
    "unix_millis_request_time": 0,
    "priority": {},
    "labels": [
      "string"
    ],
    "requester": "string"
  },
  "category": "string",
  "detail": [],
  "unix_millis_start_time": 0,
  "unix_millis_finish_time": 0,
  "original_estimate_millis": 0,
  "estimate_millis": 0,
  "assigned_to": {
    "group": "string",
    "name": "string"
  },
  "status": "uninitialized",
  "dispatch": {
    "status": "queued",
    "assignment": {
      "fleet_name": "string",
      "expected_robot_name": "string"
    },
    "errors": [
      {
        "code": 0,
        "category": "string",
        "detail": "string"
      }
    ]
  },
  "phases": {
    "additionalProp1": {
      "id": 0,
      "category": "string",
      "detail": [],
      "unix_millis_start_time": 0,
      "unix_millis_finish_time": 0,
      "original_estimate_millis": 0,
      "estimate_millis": 0,
      "final_event_id": 0,
      "events": {
        "additionalProp1": {
          "id": 0,
          "status": "uninitialized",
          "name": "string",
          "detail": [],
          "deps": [
            0
          ]
        },
        "additionalProp2": {
          "id": 0,
          "status": "uninitialized",
          "name": "string",
          "detail": [],
          "deps": [
            0
          ]
        },
        "additionalProp3": {
          "id": 0,
          "status": "uninitialized",
          "name": "string",
          "detail": [],
          "deps": [
            0
          ]
        }
      },
      "skip_requests": {
        "additionalProp1": {
          "unix_millis_request_time": 0,
          "labels": [
            "string"
          ],
          "undo": {
            "unix_millis_request_time": 0,
            "labels": [
              "string"
            ]
          }
        },
        "additionalProp2": {
          "unix_millis_request_time": 0,
          "labels": [
            "string"
          ],
          "undo": {
            "unix_millis_request_time": 0,
            "labels": [
              "string"
            ]
          }
        },
        "additionalProp3": {
          "unix_millis_request_time": 0,
          "labels": [
            "string"
          ],
          "undo": {
            "unix_millis_request_time": 0,
            "labels": [
              "string"
            ]
          }
        }
      }
    },
    "additionalProp2": {
      "id": 0,
      "category": "string",
      "detail": [],
      "unix_millis_start_time": 0,
      "unix_millis_finish_time": 0,
      "original_estimate_millis": 0,
      "estimate_millis": 0,
      "final_event_id": 0,
      "events": {
        "additionalProp1": {
          "id": 0,
          "status": "uninitialized",
          "name": "string",
          "detail": [],
          "deps": [
            0
          ]
        },
        "additionalProp2": {
          "id": 0,
          "status": "uninitialized",
          "name": "string",
          "detail": [],
          "deps": [
            0
          ]
        },
        "additionalProp3": {
          "id": 0,
          "status": "uninitialized",
          "name": "string",
          "detail": [],
          "deps": [
            0
          ]
        }
      },
      "skip_requests": {
        "additionalProp1": {
          "unix_millis_request_time": 0,
          "labels": [
            "string"
          ],
          "undo": {
            "unix_millis_request_time": 0,
            "labels": [
              "string"
            ]
          }
        },
        "additionalProp2": {
          "unix_millis_request_time": 0,
          "labels": [
            "string"
          ],
          "undo": {
            "unix_millis_request_time": 0,
            "labels": [
              "string"
            ]
          }
        },
        "additionalProp3": {
          "unix_millis_request_time": 0,
          "labels": [
            "string"
          ],
          "undo": {
            "unix_millis_request_time": 0,
            "labels": [
              "string"
            ]
          }
        }
      }
    },
    "additionalProp3": {
      "id": 0,
      "category": "string",
      "detail": [],
      "unix_millis_start_time": 0,
      "unix_millis_finish_time": 0,
      "original_estimate_millis": 0,
      "estimate_millis": 0,
      "final_event_id": 0,
      "events": {
        "additionalProp1": {
          "id": 0,
          "status": "uninitialized",
          "name": "string",
          "detail": [],
          "deps": [
            0
          ]
        },
        "additionalProp2": {
          "id": 0,
          "status": "uninitialized",
          "name": "string",
          "detail": [],
          "deps": [
            0
          ]
        },
        "additionalProp3": {
          "id": 0,
          "status": "uninitialized",
          "name": "string",
          "detail": [],
          "deps": [
            0
          ]
        }
      },
      "skip_requests": {
        "additionalProp1": {
          "unix_millis_request_time": 0,
          "labels": [
            "string"
          ],
          "undo": {
            "unix_millis_request_time": 0,
            "labels": [
              "string"
            ]
          }
        },
        "additionalProp2": {
          "unix_millis_request_time": 0,
          "labels": [
            "string"
          ],
          "undo": {
            "unix_millis_request_time": 0,
            "labels": [
              "string"
            ]
          }
        },
        "additionalProp3": {
          "unix_millis_request_time": 0,
          "labels": [
            "string"
          ],
          "undo": {
            "unix_millis_request_time": 0,
            "labels": [
              "string"
            ]
          }
        }
      }
    }
  },
  "completed": [
    0
  ],
  "active": 0,
  "pending": [
    0
  ],
  "interruptions": {
    "additionalProp1": {
      "unix_millis_request_time": 0,
      "labels": [
        "string"
      ],
      "resumed_by": {
        "unix_millis_request_time": 0,
        "labels": [
          "string"
        ]
      }
    },
    "additionalProp2": {
      "unix_millis_request_time": 0,
      "labels": [
        "string"
      ],
      "resumed_by": {
        "unix_millis_request_time": 0,
        "labels": [
          "string"
        ]
      }
    },
    "additionalProp3": {
      "unix_millis_request_time": 0,
      "labels": [
        "string"
      ],
      "resumed_by": {
        "unix_millis_request_time": 0,
        "labels": [
          "string"
        ]
      }
    }
  },
  "cancellation": {
    "unix_millis_request_time": 0,
    "labels": [
      "string"
    ]
  },
  "killed": {
    "unix_millis_request_time": 0,
    "labels": [
      "string"
    ]
  }
}`
	if err := json.Unmarshal([]byte(jsonData), &taskStateResponse); err != nil {
		t.Error(err)
		logs.Errorf("GetTaskState err: %s\n", err.Error())
	}
	logs.Init("this is a test log")
	//logs.Infof("taskStateResponse: %+v\n", taskStateResponse)
	if data, err := json.Marshal(taskStateResponse); err != nil {
		logs.Errorf("Marshal err: %s\n", err.Error())
	} else {
		logs.Infof("data: %s\n", string(data))
	}
}

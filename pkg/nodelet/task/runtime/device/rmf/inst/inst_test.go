package inst

import (
	"fmt"
	"hit.edu/framework/pkg/component-base/logs"
	"testing"
)

func TestParseAbilitySuccessResponse(t *testing.T) {
	abilitySuccessResponseStr := `{
		  "success": true,
		  "state": {
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
		  }
		}`

	abilitySuccessResponse, err := ParseAbilitySuccessResponse([]byte(abilitySuccessResponseStr))
	if err != nil {
		logs.Errorf("Parse AbilitySuccessResponseStr err:%v", err)
		t.Error(err)
	}
	fmt.Println("abilitySuccessResponse ", abilitySuccessResponse.State.Interruptions["additionalProp1"].UnixMillisRequestTime)
}

func TestParseAbilityValidationErrorResponse(t *testing.T) {
	jsonData := `{
  	"detail": [
	{	
	"loc": ["string",0],
	"msg": "string",
	"type": "string"
    }]
	}`
	AbilityValidationError, err := ParseAbilityValidationErrorResponse([]byte(jsonData))
	if err != nil {
		logs.Errorf("Parse AbilityValidationErrorResponse err:%v", err)
		t.Error(err)
	}
	fmt.Println("AbilityValidationError is ", AbilityValidationError.Detail[0].Type)
}

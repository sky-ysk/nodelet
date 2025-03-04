package events

import apis "hit.edu/framework/pkg/apis/cores"

// 定义Nodelet中所有的事件

// Image相关

// Container相关

// Ability相关

// Wasm相关

// command相关
//type ActionStatusUpdateEvent struct {
//	//ActionID      string // Action的ID，用于标识该Action
//	//Belongs       apis.IDRef
//	Phase         apis.Phase            // 可选字段，表示Action的生命周期阶段
//	Resources     []apis.ResourceStatus // 可选字段，表示当前Action的资源状态
//	Devices       []apis.DeviceStatus   // 可选字段，表示当前Action的设备状态
//	Data          []apis.DataStatus     // 可选字段，表示当前Action的数据状态
//	Scenes        []apis.SceneStatus    // 可选字段，表示当前Action的场景状态
//	RuntimeStatus []apis.RuntimeStatus  // 可选字段，表示当前Action的Runtime状态
//	Results       []apis.Result         // 可选字段，表示任务执行结果
//	StartTime     apis.Time
//	FinishTime    apis.Time
//	LastTime      apis.Time // 可选字段，表示最后更新时间
//}

// RuntimeStatusUpdateEvent 用于表示Runtime状态更新的事件
//type RuntimeStartPhaseEvent struct {
//	ProcessId string
//	Phase     apis.Phase
//	StartAt   apis.Time
//	LastTime  apis.Time  // 可选字段，表示最后更新时间
//}

type ActionResourceEvent struct {
	Resources []apis.ResourceStatus
	Devices   []apis.DeviceStatus
	Data      []apis.DataStatus
	Scenes    []apis.SceneStatus
	LastTime  apis.Time // 可选字段，表示最后更新时间
}

//// RuntimeStatusUpdateEvent 用于表示Runtime状态更新的事件
//type RuntimeStartPhaseEvent struct {
//	Group     *apis.Group
//	Action    *apis.Action
//	Runtime   *apis.Runtime
//	ProcessId string
//	Phase     apis.Phase
//	StartAt   apis.Time
//	LastTime  apis.Time // 可选字段，表示最后更新时间
//}
//
//// RuntimeStatusUpdateEvent 用于表示Runtime状态更新的事件
//type RuntimeEndPhaseEvent struct {
//	Group    *apis.Group
//	Action   *apis.Action
//	Runtime  *apis.Runtime
//	Phase    apis.Phase
//	FinishAt apis.Time
//	LastTime apis.Time // 可选字段，表示最后更新时间
//}

// RuntimeStatusUpdateEvent 用于表示Runtime状态更新的事件
type RuntimeStartPhaseEvent1 struct {
	GroupName    string
	ActionIndex  int
	RuntimeIndex int
	ProcessId    string
	Phase        apis.Phase
	StartAt      apis.Time
	LastTime     apis.Time // 可选字段，表示最后更新时间
}

// RuntimeStatusUpdateEvent 用于表示Runtime状态更新的事件
type RuntimeEndPhaseEvent1 struct {
	GroupName    string
	ActionIndex  int
	RuntimeIndex int
	Phase        apis.Phase
	FinishAt     apis.Time
	LastTime     apis.Time // 可选字段，表示最后更新时间
}

const (
	ReadyToMigrate = "ReadyToMigrate"
)

// Wasm event reason list
const (
	CreatedWasm         = "Created"
	StartedWasm         = "Started"
	FailedToCreateWasm  = "Failed"
	FailedToStartWasm   = "Failed"
	KillingWasm         = "Killing"
	PreemptWasm         = "Preempting"
	BackOffStartWasm    = "BackOff"
	ExceededGracePeriod = "ExceededGracePeriod"
)

// Command event reason list
const (
	CreatedCommand        = "Created"
	StartedCommand        = "Started"
	StoredCommand         = "Stored"
	RestoredCommand       = "Restored"
	FailedToCreateCommand = "Failed"
	FailedToStartCommand  = "Failed"
	KillingCommand        = "Killing"
	PreemptCommand        = "Preempting"
	BackOffStartCommand   = "BackOff"
	TriggerMigration      = "TriggerMigration"
)

package apis

import (
	"hit.edu/framework/pkg/apis/meta"
	"time"
)

const (
	NamespaceDefault string = "defaultNamespace"
	NamespaceAll     string = ""
)

// 定义框架基础资源
// 节点资源信息
// TODO: 接口版本

// TODO: 独立配置
type Time struct {
	time.Time `json:"time" yaml:"time"`
}

// TODO: 独立配置
type Event struct {
	//TODO: 定义Event
	//TODO: ObjectReference设计
}

// Node
type Node struct {
	//
	meta.TypeMeta

	//
	meta.ObjectMeta

	// 定义Node行为
	Spec NodeSpec `json:"spec,omitempty" yaml:"spec"`

	// 定义Node的当前状态
	Status NodeStatus `json:"status,omitempty" yaml:"status"`
}

type NodeSpec struct {
	// 调度时需要使用，将任务调度到该节点
	NodeName string `json:"node_name,omitempty" yaml:"node_name"`

	// 节点的HostName
	HostName string `json:"host_name,omitempty" yaml:"host_name"`

	// 不能被调度的节点
	Unschedulable bool `json:"unschedulable,omitempty" yaml:"unschedulable"`
	// TODO: 节点Label
}

type NodeList struct {
	meta.TypeMeta

	meta.ListMeta
	// TODO: List Options

	Items []Node `json:"items" yaml:"items"`
}

type WorkflowList struct {
	meta.TypeMeta
	meta.ListMeta
	Items []Workflow
}
type TaskList struct {
	meta.TypeMeta
	meta.ListMeta
	Items []Task
}
type GroupList struct {
	meta.TypeMeta
	meta.ListMeta
	Items []Group
}
type ActionList struct {
	meta.TypeMeta
	meta.ListMeta
	Items []Action
}

// todo device, data, scen,resource 需要list吗？
type DeviceList struct {
	meta.TypeMeta
	meta.ListMeta
	Items []Device
}

type Resource_NodeList struct {
	meta.TypeMeta
	meta.ListMeta
	Items []Resource_Node
}

type DataList struct {
	meta.TypeMeta
	meta.ListMeta
	Items []Data
}
type SceneList struct {
	meta.TypeMeta
	meta.ListMeta
	Items []Scene
}

// 计算、网络、存储等定量资源
// 资源名称 => 定量资源描述
type ResourceList map[string]Quantity

type NodeStatus struct {
	// 节点上的所有物理资源
	Capacity ResourceList `json:"capacity,omitempty" yaml:"capacity"`

	// 节点上当前可以分配的资源
	Allocatable ResourceList `json:"allocatable,omitempty" yaml:"allocatable"`

	// 这里只表示连接关系，对硬件的调用放到能力中
	// TODO: 节点上的硬件资源

	// TODO：节点上部署的任务

	// 节点上的Docker镜像
	Images []ContainerImage `json:"images,omitempty" yaml:"images"`

	// 节点上的Wasm镜像
	Wasms []WasmImage `json:"wasms,omitempty" yaml:"wasms"`

	// TODO：节点上的依赖情况

	// 节点的物理地址，可以有多个
	Addresses NodeAddress `json:"addresses,omitempty" yaml:"addresses"`

	// 节点系统信息
	NodeInfo NodeSystemInfo `json:"info,omitempty" yaml:"info"`
}

// From K8s
type NodeSystemInfo struct {
	// MachineID reported by the node. For unique machine identification
	// in the cluster this field is preferred. Learn more from man(5)
	MachineID string `json:"machine_id,omitempty" yaml:"machine_id"`
	// SystemUUID reported by the node. For unique machine identification
	// MachineID is preferred. This field is specific to Red Hat hosts
	SystemUUID string `json:"system_uuid,omitempty" yaml:"system_uuid"`
	// Boot ID reported by the node.
	BootID string `json:"boot_id,omitempty" yaml:"boot_id"`
	// Kernel Version reported by the node.
	KernelVersion string `json:"kernel_version,omitempty" yaml:"kernel_version"`
	// OS Image reported by the node.
	OSImage string `json:"os_image,omitempty" yaml:"os_image"`
	// ContainerRuntime Version reported by the node.
	ContainerRuntimeVersion string `json:"container_runtime_version,omitempty" yaml:"container_runtime_version"`
	// NodeLet
	NodeletVersion string `json:"nodelet_version,omitempty" yaml:"nodelet_version"`
	// The Operating System reported by the node
	OperatingSystem string `json:"os,omitempty" yaml:"os"`
	// The Architecture reported by the node
	Architecture string `json:"architecture,omitempty" yaml:"architecture"`
}

// From K8s
// NodeAddress represents node's address
type NodeAddress struct {
	Type    string `json:"type,omitempty" yaml:"type"`
	Address string `json:"address,omitempty" yaml:"address"`
}

// From K8s
type ContainerImage struct {
	// Names by which this image is known.
	Names []string `json:"names,omitempty" yaml:"names"`
	// The size of the image in bytes.
	SizeBytes int64 `json:"size,omitempty" yaml:"size"`
}

// Wasm镜像
type WasmImage struct {
	// Names by which this image is known.
	Names []string `json:"names,omitempty" yaml:"names"`
	// The size of the image in bytes.
	SizeBytes int64 `json:"size,omitempty" yaml:"size"`
}

// 设备资源信息
// Device
// Device具有Ability

// 能力资源信息
// Ability
// TODO: 待增加

// 工作流相关资源，表示一次部署执行的内容
// 框架中工作流包含四级,Workflow => Task => Group => Action
// 其中Group和Action具体部署在节点上
// Task和Workflow表示逻辑结构,Group,Action表示具体的执行关系
// 对工作流的描述
type Description struct {
	// TODO: Label单独字段
	Label []string `json:"label,omitempty" yaml:"label"`
	// 用户对工作流行为的描述
	// +Optional
	Docs string `json:"docs,omitempty" yaml:"docs"`
}

// 工作流节点执行需要满足以下条件
// 1. 节点依赖，所有前序节点都需要执行完成
// 2. 数据依赖，所需的数据都下载到对应的节点上
// 3. 资源依赖，所需的资源都已经满足
// 4. 程序依赖，程序依赖已经安装完成

// TODO: 增加Label及相关选择器

// ------- Workflow

// 工作流相关生命周期
type Phase string

const (
	// 任务相关状态
	Pending   Phase = "Pending"
	Running   Phase = "Running"
	Successed Phase = "Succeeded"
	Failed    Phase = "Failed"
	Unknown   Phase = "Unknown"
	// 迁移相关状态
	Migrating Phase = "Migrating"
	Migrated  Phase = "Migrated"
)

// 定义流程类型，用于表示有条件的DAG
// 支持顺序、分支和循环（有限展开)
type ProcessType string

const (
	// 普通类型的节点，默认为该节点
	Norm ProcessType = "Normal"
	// 需要判断执行条件的Action
	// 用于分支类型的节点
	Cond ProcessType = "Condition"
	// 带有循环生成器的Action
	// 检查Condition是否满足所需条件
	// 如果不满足条件，或者未达到最大循环次数，则继续生成Action
	// TODO: 生成的ActionID需要以Loop+{Count}为后缀
	// TODO: 循环计数器Count
	Loop ProcessType = "Loop"
	//...待后续拓展
)

// 流程条件

type ConditionValueType string

const (
	Constants ConditionValueType = "Constants"
	Dynamic   ConditionValueType = "Dynamic"
)

// TODO: 参考Inputs,重新定义
type ConditionValue struct {
	// Condition的变量有以下类型
	// 	Constants, Value则为常数，From中信息为空
	//  Dynamic, Value需要根据From中的信息从Results中获取
	//  	比如前序任务的执行状态
	//		或者循环计数器的数量
	// 		或者任务的执行结果
	// TODO: 使用DataType代替，更新相关文档
	Type DataType `json:"type,omitempty" yaml:"type"` //ConditionValueType
	//
	Name string `json:"name,omitempty" yaml:"name"`
	// 实际的值
	Value string `json:"value,omitempty" yaml:"value"`
	// TODO: From
	// TODO: 动态类型的Value,数据来源,需要对应的Controller Watch相关变量
	// +Optional
	ValueType string `json:"value_type,omitempty" yaml:"value_type"`
}

// 条件连接符，支持大小写
type JoinType string

const (
	and JoinType = "and"
	or  JoinType = "or"
	And JoinType = "and"
	Or  JoinType = "or"
)

// 符号判断
type SignalType string

const (
	Equal    SignalType = "=="
	NotEqual SignalType = "!="
)

// 流程执行条件
// LeftValue ==或!= RightValue
// 输出结果为Bool类型的值
// TODO: Value格式检查和调整，比如存在空格的情况
type ConditionFormula struct {
	LeftValue  ConditionValue `json:"left_value,omitempty" yaml:"left_value"`
	RightValue ConditionValue `json:"right_value,omitempty" yaml:"right_value"`
	// == 或 !=
	Signal SignalType `json:"signal,omitempty" yaml:"signal"`
	// 在条件串中的期望结果
	// 类型包含 and 或者 or
	Join JoinType `json:"join,omitempty" yaml:"join"`
	// TODO: 符号判断结果
}

// 不建议使用过于复杂的逻辑
// 只支持逻辑的串行连接
// 如果没有Condition,Conditions默认值为true
// Example: Condition[1] and/or Condition[2] and/or Condition[3] ......
type Conditions struct {
	Formulas []ConditionFormula `json:"formulas,omitempty" yaml:"formulas"`
}

// 工作流相关Ref
// 表示一个工作流属于哪个ID
// 唯一标识
type IDRef struct {
	WorkflowID string `json:"workflow_id,omitempty" yaml:"workflow_id"`
	TaskID     string `json:"task_id,omitempty" yaml:"task_id"`
	GroupID    string `json:"group_id,omitempty" yaml:"group_id"`
	ActionID   string `json:"action_id,omitempty" yaml:"action_id"`
}

type Workflow struct {
	//
	meta.TypeMeta

	//
	meta.ObjectMeta

	//
	Spec WorkflowSpec `json:"spec,omitempty" yaml:"spec"`

	//
	Status WorkflowStatus `json:"status,omitempty" yaml:"status"`
}

type WorkflowSpec struct {
	// Workflow Name, 用户提交时提供的Name
	// FIXME: Name应当唯一，否则不允许提交
	Name string `json:"name,omitempty" yaml:"name"`

	// 对工作流的描述
	Desc Description `json:"desc,omitempty" yaml:"desc"`

	// 工作流下有哪些任务
	// 目前只支持Task=>Group=>Action
	// TODO: 更为宽松的Task关系定义
	// 目前与Action的定义方式类似
	Tasks []Task `json:"tasks,omitempty" yaml:"tasks"`
	// TODO: Tasks的拓扑关系
}

type WorkflowStatus struct {
	// Workflow ID，系统为根据Desc中的Name为Workflow分配的唯一标识, 用户填写时应该为空
	WorkflowID string `json:"workflow_id,omitempty" yaml:"workflow_id"`

	//
	Phase Phase `json:"phase,omitempty" yaml:"phase"`

	//
	TaskStatus []TaskStatus `json:"status,omitempty" yaml:"status"`

	// 执行时间
	StartAt Time `json:"start,omitempty" yaml:"start"`
	// 结束时间
	FinishAt Time `json:"finish,omitempty" yaml:"finish"`
	// 最新获取状态的时间
	LastTime Time `json:"last_time,omitempty" yaml:"last_time"`
}

// --------- Task
type Task struct {
	//
	meta.TypeMeta

	//
	meta.ObjectMeta

	Spec   TaskSpec   `json:"spec,omitempty" yaml:"spec"`
	Status TaskStatus `json:"status,omitempty" yaml:"status"`
}

type TaskTemplate struct {
	Spec TaskSpec `json:"spec,omitempty" yaml:"spec"`
}

type TaskSpec struct {
	// Task Name
	Name string `json:"name,omitempty" yaml:"name"`

	// Parents Name
	Parents []string `json:"parents,omitempty" yaml:"parents"`

	// Task描述
	Desc Description `json:"desc,omitempty" yaml:"desc"`

	// Task类型
	Type ProcessType `json:"type,omitempty" yaml:"type"`

	// Task条件
	// +Optional
	Conditions Conditions `json:"conditions,omitempty" yaml:"conditions"`

	// 存储当前Task中的所有Group
	// Group之间没有严格的依赖限制
	// 可以有多个Group作为Group的入口，支持多个图形结构
	// 支持并发执行多组Group
	// TODO: 增加Level支持
	Groups []Group `json:"groups,omitempty" yaml:"groups"`
}

type TaskStatus struct {
	//
	TaskID string `json:"task_id,omitempty" yaml:"task_id"`

	//
	Belongs IDRef `json:"belongs,omitempty" yaml:"belongs"`

	//
	Phase Phase `json:"phase,omitempty" yaml:"phase"`

	//
	GroupStatus []GroupStatus `json:"group_status,omitempty" yaml:"group_status"`

	// TODO: Events定义

	// 执行时间
	StartAt Time `json:"start,omitempty" yaml:"start"`
	// 结束时间
	FinishAt Time `json:"finish,omitempty" yaml:"finish"`
	// 最新获取状态的时间
	LastTime Time `json:"last_time,omitempty" yaml:"last_time"`
}

// ---------- Group
type Group struct {
	//
	meta.TypeMeta

	//
	meta.ObjectMeta

	//
	Spec GroupSpec `json:"spec,omitempty" yaml:"spec"`

	//
	Status GroupStatus `json:"status,omitempty" yaml:"status"`
}

type GroupTemplate struct {
	Spec GroupSpec `json:"spec,omitempty" yaml:"spec"`
}

type GroupSpec struct {
	// Group Name
	Name string `json:"name,omitempty" yaml:"name"`

	//
	SchedulerName string `json:"scheduler_name,omitempty" yaml:"name"`

	// Parents Name, 一个Group可以有多个Parents
	Parents []string `json:"parents,omitempty" yaml:"parents"`

	// Group描述
	Desc Description `json:"desc,omitempty" yaml:"desc"`

	// Group类型
	Type ProcessType `json:"type,omitempty" yaml:"type"`

	// Group条件
	// +Optional
	Conditions Conditions `json:"conditions,omitempty" yaml:"conditions"`

	// Group所需资源需要先遍历自己的Action
	//   Ref类型的指针需要根据Action中的资源需求计算
	//   TODO: 对于Condition类的节点，使用资源的预估
	//   TODO: 资源需求

	// 存储当前Group所属的所有Action
	// Group中可以只有一个Action, 简化实现逻辑
	Actions []Action `json:"actions,omitempty" yaml:"actions"`
	// TODO: 生成时是否可以直接分析依赖关系？
	// TODO: Action的依赖关系描述, 需要先遍历Actions构建DAG图
	// Group中的Action需要有较严格的依赖顺序，可以支持分支,条件,循环
	// 只能有一个Action作为入口Action,图形结构

	SkipScorePlugins []string `json:"skip_score_plugins,omitempty" yaml:"skip_score_plugins"`

	SkipFilterPlugins []string `json:"skip_filter_plugins,omitempty" yaml:"skip_filter_plugins"`
	//添加-hzy
	Replicas int32 `json:"replicas,omitempty" yaml:"replicas"`
}

type GroupStatus struct {
	//
	GroupID string `json:"group_id,omitempty" yaml:"group_id"`

	//
	Belongs IDRef `json:"belongs,omitempty" yaml:"belongs"`

	//
	Phase Phase `json:"phase,omitempty" yaml:"phase"`

	// TODO: 所属Actions的状态
	ActionStatus []ActionStatus `json:"action_status,omitempty" yaml:"action_status"`

	// TODO: 整体资源使用情况

	// TODO: Events定义

	// 执行时间
	StartAt Time `json:"start,omitempty" yaml:"start"`
	// 结束时间
	FinishAt Time `json:"finish,omitempty" yaml:"finish"`
	// 最新获取状态的时间
	LastTime Time `json:"last_time,omitempty" yaml:"last_time"`
	//添加-hzy
	CheckDependencyCount int32 `json:"check_dependency_count,omitempty" yaml:"check_dependency_count"`
}

// ---------- Action

type Action struct {
	//
	meta.TypeMeta

	//
	meta.ObjectMeta

	//
	Spec ActionSpec `json:"spec,omitempty" yaml:"spec"`

	//
	Status ActionStatus `json:"status,omitempty" yaml:"status"`
}

// 创建Action所需要的字段
type ActionTemplate struct {
	Spec ActionSpec `json:"spec,omitempty" yaml:"spec"`
}

type ActionSpec struct {
	// Action Name
	Name string `json:"name,omitempty" yaml:"name"`

	// Parents Name, 一个节点可以有多个Parents
	Parents []string `json:"parents,omitempty" yaml:"parents"`

	// Action描述
	Desc Description `json:"desc,omitempty" yaml:"desc"`

	// Action类型
	Type ProcessType `json:"type,omitempty" yaml:"type"`

	// Action条件
	// +Optional
	Conditions Conditions `json:"conditions,omitempty" yaml:"conditions"`

	// 需要的执行环境
	// 串行执行Runtime中的运行环境
	// 如果使用Pod方式部署的任务，建议只部署一个阻塞的Runtime
	// 可以用于执行命令，环境准备等操作
	Runtimes []Runtime `json:"runtimes,omitempty" yaml:"runtimes"`

	// 其他配置/选项
	// 细粒度任务控制
	// TODO: RPC相关接口实现
	EnableFineGrainedControl bool `json:"enable_control,omitempty" yaml:"enable_control"`
}

type RuntimeType string

const (
	ByDevice     RuntimeType = "device"
	ByNet        RuntimeType = "net"
	ByCommand    RuntimeType = "command"
	ByBinary     RuntimeType = "binary"
	ByDocker     RuntimeType = "docker"
	ByService    RuntimeType = "service"
	ByDeployment RuntimeType = "deployment"
	ByPod        RuntimeType = "pod"
)

// 环境变量
// TODO: 环境变量格式定义
type EnvVar struct {
	Name  string `json:"name,omitempty" yaml:"name"`
	Value string `json:"value,omitempty" yaml:"value"`
	// TODO: 动态获取相关字段
}

type Resource_Node struct {
	meta.TypeMeta

	meta.ObjectMeta

	Spec ResourceSpec `json:"spec,omitempty" yaml:"spec"`

	Status ResourceStatus `json:"status,omitempty" yaml:"status"`
}

type Data struct {
	meta.TypeMeta

	meta.ObjectMeta

	Spec DataSpec `json:"spec,omitempty" yaml:"spec"`

	Status DataStatus `json:"status,omitempty" yaml:"status"`
}

type Scene struct {
	meta.TypeMeta

	meta.ObjectMeta

	Spec SceneSpec `json:"spec,omitempty" yaml:"spec"`

	Status SceneStatus `json:"status,omitempty" yaml:"status"`
}

// TODO: 后续补充完整

// TODO:node字段
type ResourceSpec struct {
	// 描述期待占用多少资源 资源的单位是什么
	ExpectedValue     float64
	ExpectedValueUnit ResourceUnit

	// 描述资源的固有属性
	Type     ResourceType   // 类型：网络、计算、存储
	Name     string         // 名称：cpu、gpu、内存、硬盘
	Unit     ResourceUnit   // 资源的单位
	Detail   ResourceDetail // 资源的细致描述
	Capacity float64        // 资源的总量
	NodeId   string         // 资源属于哪个node
}
type ResourceDetail struct {
	Type string
}
type ResourceStatus struct {
	// 资源的使用量和他的单位
	Usage     float64
	UsageUnit ResourceUnit

	// 剩余的资源和单位
	Reserved     float64
	ReservedUnit ResourceUnit
}

type ResourceType string

const (
	Compute ResourceType = "compute"
	Network ResourceType = "network"
	Storage ResourceType = "storage"
)

type ResourceUnit string

const (
	StorageKB       ResourceUnit = "KB"
	StorageMB       ResourceUnit = "MB"
	StorageGB       ResourceUnit = "GB"
	StorageTB       ResourceUnit = "TB"
	ComputeCPU      ResourceUnit = "cpu"
	ComputeMilliCPU ResourceUnit = "milliCPU"
	ComputeNanoCPU  ResourceUnit = "nanocpu"
	ComputeGPU      ResourceUnit = "gpu"
	NetworkGbps     ResourceUnit = "Gbps"
	NetworkMbps     ResourceUnit = "Mbps"
	NetworkKbps     ResourceUnit = "Kbps"
	Networkbps      ResourceUnit = "bps"
)

// 增加设备定义
type Device struct {
	//
	meta.TypeMeta

	//
	meta.ObjectMeta

	//
	Spec DeviceSpec

	//
	Status DeviceStatus
}

type DevicePhase string

const (
	DeviceInit         DevicePhase = "Init"
	DeviceRunning      DevicePhase = "Running"
	DeviceIdle         DevicePhase = "Idle"
	DeviceError        DevicePhase = "Error"
	DeviceDisconnected DevicePhase = "Disconnected"
)

// 对设备能力的描述
type DeviceDesc struct {
	// 在任务部署时，根据Label的内容查找所需设备
	Label []string
	// +Optional
	Docs string
}

type AccessType string

const (
	AccessByAbility   AccessType = "ByAbility"
	AccessByRmf       AccessType = "ByRmf"
	AccessByCustomize AccessType = "ByCustomize"
)

// 设备的访问方式
type AccessMethod struct {
	// Type
	Type AccessType

	// 访问方式对应的URL
	//   对于Customize类型，URL对应部署脚本的地址
	URL string

	// 设备组
	//  对于RMF类型，Group对应RMF Fleets
	Group string

	// 设备别名
	//   部分情况下，RMF中的设备名与系统中设备名不一致
	//   默认情况下，Alias应该与Name相同
	Alias string
}

type PropertyType string

const (
	BoolType    PropertyType = "bool"
	StringType  PropertyType = "String"
	IntegerType PropertyType = "Integer"
	DoubleType  PropertyType = "Double"
	URLType     PropertyType = "URL"
	ComposeType PropertyType = "Compose"
)

// 设备的属性
// TODO: 任务部署时，通过类似Device.Property的方式来寻址
type Property struct {
	// Name
	Name string

	// Type
	Type PropertyType

	// TODO: 格式校验
	Value string

	SubProperty []SubProperty
}

// Compose类型的设备属性
// 设备属性的子属性，如Location类型，具有子属性Location.X
type SubProperty struct {
	// Name
	Name string

	// Type
	Type PropertyType

	// TODO: 格式校验
	Value string
}

type LockType string

const (
	SharedLock LockType = "Shared"
	MutexLock  LockType = "Mutex"
	NoneLock   LockType = "None"
)

// 设备资源锁
type Lock struct {
	// 锁类型
	Type LockType

	// 调度时 ref为0时释放
	Lock bool

	// 资源引用数 部署时
	Ref int
}

// 设备事件描述
type DeviceEvent struct {
	// EventCode, 时间码，对应事件处理的方案
	Code int

	// 事件描述
	Desc string

	// TODO: 额外参数, 设备的上下文状态
}

type DeviceSpec struct {
	// 设备名称，每个Node上的设备，名称应该唯一
	Name string

	// 对设备的描述
	Desc DeviceDesc

	// 设备的关联Node，每个设备需要与一个Node相关联
	Node string

	// 设备的访问方式
	AccessMethod AccessMethod

	// 父设备
	AttachedDevice string

	// 子设备
	SubDevices []string

	// 设备的期望属性
	ExpectedProperties map[string]Property
}

type DeviceStatus struct {
	// DeviceID, Name+NodeID
	DeviceID string

	// 正在使用Device的ActionID
	ActionID string

	// 设备的实例ID
	//  对于Ability来说，InstanceID对应Ability的InstanceID
	//  对于RMF来说，InstanceID对应RMF的TaskID
	InstanceID string

	// 设备的运行阶段
	Phase DevicePhase

	// 运行时中，设备的实际状态
	// 当Phase与Status不一致时，机器人出现运行错误
	Status string

	// 设备的实际属性
	Properties map[string]Property

	// 设备资源锁状态
	Lock Lock

	// 设备事件描述
	Events []DeviceEvent

	// 上次成功获取设备状态的时间
	// 如果长时间不能获取设备的状态，则认为设备离线
	LastTime Time
}

type DataSpec struct {
	// 对于文件类型的Data
	// 文件格式
	// 文件大小
	// SHA文件校验
}
type DataStatus struct{}

// SceneSpec 描述scene的固有属性和期待属性
type SceneSpec struct {
	// 每一个scene的标识
	SceneID string

	// scene的类型 是一个地点还是一个物品
	Type SceneType

	// 期待属性
	ExpectedProperty map[string]Property

	// 场景的描述（不可变属性）
	Desc SceneDesc
}

type SceneDesc struct {
	Label []string
	Value map[string]string
}

type SceneType string

const (
	ObjectType   SceneType = "Object"
	PositionType SceneType = "Position"
)

/*
	Object的位置信息存储在SceneStatus.Property中
	Position的位置信息存储在SceneSpec.SceneDesc中
*/

// SceneStatus 描述scene的动态属性
type SceneStatus struct {
	// 更新的方式和时间
	UpdateMethod string
	UpdateTime   Time

	// 关联的场景
	AttachedScene string

	// 关联的设备
	AttachedDevice string

	// 关联的任务
	AttachedTask string

	// 实时属性
	Property map[string]Property

	// 锁
	Lock Lock
}

// Action所需执行环境
type Runtime struct {
	// 定义Action所需资源
	// 在Group调度完成后，Action需要被分配和保留，直到Action开始执行
	// 动态资源分配所需字段

	// 运行时,需要检查以下四类资源，只有资源都满足时，才可以继续执行
	// 需要的计算、网络、存储资源
	Resources []ResourceSpec `json:"resources,omitempty" yaml:"resources"`

	// 需要的硬件资源
	Devices []DeviceSpec `json:"devices,omitempty" yaml:"devices"`

	// 需要使用的数据
	Data []DataSpec `json:"data,omitempty" yaml:"data"`

	// 需要使用的场景数据
	Scenes []SceneSpec `json:"scenes,omitempty" yaml:"scenes"`

	// 运行时类型
	// 执行环境,有如下种类：
	//      Device, 端侧需要控制设备和能力
	//      Net, 需要发起网络请求
	//      Command, 需要执行脚本命令
	// 		Binary, 需要检查环境依赖，二进制所在位置
	// 		Docker, 需要下载镜像到对应节点，启动Docker并注入环境依赖
	// 		Service, 通过K8s部署Service到节点上
	// 		Deployment, 通过K8s部署Deployment到节点上
	// 		Pod, 通过K8s部署Pod到节点上
	// 其中端侧只能使用Device,Net,Command,Binary或者Docker的方式部署运行环境,使用Docker环境部署需要注意网络环境
	// 云边侧支持除Device外的其他方式,使用Docker部署需要注意网络环境
	// 通过Pod方式部署的Action,如果不启用细粒度任务控制，需要通过K8s相关接口获取运行状态
	// TODO: 如果使用Pod方式部署，需要记录Pod对应的ID,如果生命周期内Pod存在改动，该需要对应的更新
	// TODO: 实现一个Controller,专门监控对应的Pod的变化
	// TODO: 使用Pod方式部署，可以同时部署Action的多个副本
	Type RuntimeType `json:"type,omitempty" yaml:"type"`

	// 标识运行时的名称
	Name string `json:"name,omitempty" yaml:"name"`

	// 镜像
	//  对于Command,Net类型，Image应当为空
	// 	对于Binary或者Script，Image对应二进制上传的位置，部署时需要检查依赖环境
	// 	对于Docker,Image对应镜像的存储位置
	//  对于Service,Deployment和Pod,对应相关文件的存储位置，需要使用网络链接
	// 需要先检查本地是否有对应版本的镜像
	// +Optional
	Image string `json:"image,omitempty" yaml:"image"`

	// TODO: 软件依赖如何表示

	// 执行参数
	// 程序的入口函数，一般不做更改
	Command []string `json:"command,omitempty" yaml:"command"`

	// 如果程序要注入其他的运行参数，则放到这里
	// +Optional
	Args []string `json:"args,omitempty" yaml:"args"`

	// 环境变量
	// +Optional
	EnvVar []EnvVar `json:"env_var,omitempty" yaml:"env_var"`

	// 需要的数据
	// 输入数据
	// 	输入数据作为参数注入到命令参数中
	Inputs Input `json:"inputs,omitempty" yaml:"inputs"`

	// 输出数据
	//  输出数据作为参数注入到命令参数中
	Outputs Output `json:"outputs,omitempty" yaml:"outputs"`

	//添加-hzy
	Parents []string `json:"parents,omitempty" yaml:"parents"`
	Waiting bool     `json:"waiting,omitempty" yaml:"waiting"`
}

//	 输入的数据有以下几类
//			常量类型的数据
//			从Results中获取数据
//			从本地获取的资源中获取数据
type DataType string

const (
	ConstData   DataType = "constants"
	ResultsData DataType = "results"
	LocalData   DataType = "local"
)

type Input struct {
	// 类型
	Type DataType `json:"type,omitempty" yaml:"type"`

	// 名称
	Name string `json:"name,omitempty" yaml:"name"`

	// 值
	//   对应常量类型，ValueType对应常量的类型，Value对应常量的值，类型与值应该对应，支持的类型
	// 		包括：整数、浮点数、布尔值
	//   对于Results类型，Value对应从哪个节点的Results中获取数据，ValueType对应从Results中获取的ResultType
	//      Results的访问格式对应
	// 			TODO: 正则表达式
	//			Action{ID}.Results.{Name}， 缺省访问本Group对应的Action
	//			Group{ID}.Action{ID}.Results.{Name}， 访问对应Group的对应Action
	// 			Task{ID}.Group{ID}.Action{ID}.Results.{Name}, 访问对应Task的对应Group的对应Action
	//          目前不支持跨Workflow获取数据
	//   对于Local类型，Value对应从本地资源或者数据节点获取的数据，ValueType对应从资源或者数据中获取的数据类型
	//   	Local的访问格式对应
	//          Resources.{Name}: 从本地资源中获取
	//			Devices.{Name}: 从本地设备里列表中获取
	//			Scenes.{Name}： 从本地场景中获取
	//			Data.{Name}： 从本地数据中获取
	//      Local类型的数据对其他节点不可见
	Value     string `json:"value,omitempty" yaml:"value"`
	ValueType string `json:"value_type,omitempty" yaml:"value_type"`
}

// TODO: 数据格式后续还需要调整
type Output struct {
	//
	Type DataType `json:"type,omitempty" yaml:"type"`

	Name string `json:"name,omitempty" yaml:"name"`

	// TODO:
	Value     string `json:"value,omitempty" yaml:"value"`
	ValueType string `json:"value_type,omitempty" yaml:"value_type"`
}

type ActionStatus struct {
	// 在开始执行时，确定Action所属的Workflow,Task,Group
	// ID组成格式为ActionName+GroupID
	ActionID string `json:"action_id,omitempty" yaml:"action_id"`
	// Action所属标识
	Belongs IDRef `json:"belongs,omitempty" yaml:"belongs"`
	// 生命周期
	Phase Phase `json:"phase,omitempty" yaml:"phase"`
	// 当前资源使用情况
	Resources []ResourceStatus `json:"resources,omitempty" yaml:"resources"`
	// 当前设备使用情况
	Devices []DeviceStatus `json:"devices,omitempty" yaml:"devices"`
	// 当前数据使用情况
	Data []DataStatus `json:"data,omitempty" yaml:"data"`
	// 当前场景更新情况
	Scenes []SceneStatus `json:"scenes,omitempty" yaml:"scenes"`
	// 当前Runtime执行状态
	RuntimeStatus []RuntimeStatus `json:"status,omitempty" yaml:"status"`
	// 任务执行结果
	Results []Result `json:"results,omitempty" yaml:"results"`
	// TODO: Events定义

	// 执行时间
	StartAt Time `json:"start,omitempty" yaml:"start"`
	// 结束时间
	FinishAt Time `json:"finish,omitempty" yaml:"finish"`
	// 最新获取状态的时间
	LastTime Time `json:"last_time,omitempty" yaml:"last_time"`

	//增加一个参数-hzy
	Waiting bool `json:"waiting,omitempty" yaml:"waiting"`
}

type RuntimeStatus struct {
	// 当前任务的执行情况
	// 任务在哪里执行,进程ID
	NodeName  string `json:"name,omitempty" yaml:"name"`
	ProcessId string `json:"process_id,omitempty" yaml:"process_id"`

	// 执行状态
	Phase Phase `json:"phase,omitempty" yaml:"phase"`

	// 执行时间
	StartAt Time `json:"start,omitempty" yaml:"start"`
	// 结束时间
	FinishAt Time `json:"finish,omitempty" yaml:"finish"`
	// 最新获取状态的时间
	LastTime Time `json:"last_time,omitempty" yaml:"last_time"`

	//增加一个参数0hzy
	RuntimeID string `json:"runtime_id,omitempty" yaml:"runtime_id"`
}

// 任务的输出结果
// 单独作为一个资源，方便其他节点获取该资源
// 数据结果的类型
type Result struct {
	//Result唯一表示，由ActionID
	ResultID string `json:"result_id,omitempty" yaml:"result_id"`
	// Result属于哪个Action/Group/Task/Workflow
	Belongs IDRef `json:"belongs,omitempty" yaml:"belongs"`
	// Results类型，有如下的类型
	// 		存在内存中的结果,可以通过etcd模块同步
	// 		比较大的数据，需要提供资源的访问方式，资源同步模块需要该字段，由该模块实现资源的点对点同步
	Type string `json:"type,omitempty" yaml:"type"`
	// TODO: Results存储位置，访问方式
	// Results可以有备份位置
	// 需要构建Ownership表，任务执行节点可以获取所需数据的位置
}

// 用于端口的结构体 ---hzy暂时增加
type Port struct {
	Protocol   string
	Port       int
	TargetPort int
}

// 系统运行时相关信息，表示长时间部署执行的内容
// 云边侧模块使用相关字段
// TODO: Pod From K8s
type Pod struct{}

// TODO: Service From K8s
type Service struct{}

// TODO: Deployment From K8s
type Deployment struct{}

// TODO: VM From K8s
type VM struct{}

// 后续需要可以扩展

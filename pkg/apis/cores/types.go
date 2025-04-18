// +k8s:deepcopy-gen=package
package apis

import (
	"hit.edu/framework/pkg/apis/meta"
	"time"
)

const (
	NamespaceDefault string = "defaultNamespace"
	NamespaceTest           = "test"
	NamespaceAll     string = ""
)

// 定义框架基础资源
// 节点资源信息
// TODO: 接口版本

type Item struct {
	Name   string            `json:"name,omitempty" yaml:"name"`
	Desc   string            `json:"desc,omitempty" yaml:"desc"`
	Labels []string          `json:"labels,omitempty" yaml:"labels"`
	Values map[string]string `json:"values,omitempty" yaml:"values"`
}

type Quantity struct {
	// 定量数据
	i int64

	// 单位
	format string
}

// +k8s:deepcopy-gen=false
type Time struct {
	time.Time `json:"time" yaml:"time"`
}

//type Time time.Time

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type NodeList struct {
	meta.TypeMeta
	meta.ListMeta
	// TODO: List Options
	Items []Node `json:"items" yaml:"items"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type WorkflowList struct {
	meta.TypeMeta
	meta.ListMeta
	Items []Workflow `json:"items" yaml:"items"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type TaskList struct {
	meta.TypeMeta
	meta.ListMeta
	Items []Task `json:"items" yaml:"items"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type GroupList struct {
	meta.TypeMeta
	meta.ListMeta
	Items []Group `json:"items" yaml:"items"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type ActionList struct {
	meta.TypeMeta
	meta.ListMeta
	Items []Action `json:"items" yaml:"items"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type DataList struct {
	meta.TypeMeta
	meta.ListMeta
	Items []Data `json:"items" yaml:"items"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type SceneList struct {
	meta.TypeMeta
	meta.ListMeta
	Items []Scene `json:"items" yaml:"items"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type DeviceList struct {
	meta.TypeMeta
	meta.ListMeta
	Items []Device `json:"items" yaml:"items"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type Resource_NodeList struct {
	meta.TypeMeta
	meta.ListMeta
	Items []Resource_Node `json:"items" yaml:"items"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type Event struct {
	//TODO: 定义Event
	//TODO: ObjectReference设计
	meta.TypeMeta
	meta.ObjectMeta
	InvolvedObject ObjectReference
	// 事件产生原因，机器可读，供handler判断
	Reason string
	// 描述，应有用户可读性
	Message string
	// 事件产生来源
	Source    EventSource
	EventTime Time
	Count     int32
	Type      string // EventTypeNormal or EventTypeWarning
	// todo: 补充 action、reporting controller 、 instance
}

type EventSource struct {
	// 事件产生组件
	Component string
	// 事件产生节点
	Host string
}

// event type 常量
const (
	EventTypeNormal    string = "Normal"
	EventTypeWarning   string = "Warning"
	EventTypeMigration string = "Migration"
)

// todo:改objereference
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type ObjectReference struct {
	// GVK
	// +Optional
	APIVersion string
	// +Optional
	Kind string
	// Name
	Namespace string
	Name      string
	// +Optional
	UID UID
	// +Optional
	ResourceVersion string
	// +Optional
	FieldPath string
}
type UID string

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type EventList struct {
	meta.TypeMeta

	meta.ListMeta

	Items []Event `json:"items" yaml:"items"`
}

// Node
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
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

	// 设备固有资源
	Resource map[string][]Item `json:"resource,omitempty" yaml:"resource"`
	// TODO: 节点Label

	//  +个字段（Cloud、Edge、End）
	ClusterCategory string `json:"clusterCategory,omitempty" yaml:"clusterCategory"` //该节点所在的集群类别：1、云集群 2、边集群 3、端集群
}

// 计算、网络、存储等定量资源
// 资源名称 => 定量资源描述
type ResourceList map[string]Quantity

type NodeStatus struct {
	// 节点上的所有物理资源
	Capacity ResourceList `json:"capacity,omitempty" yaml:"capacity"`

	// 节点上当前可以分配的资源
	Allocatable ResourceList `json:"allocatable,omitempty" yaml:"allocatable"`

	// 节点上的资源的动态资源
	Usage map[string][]Item `json:"usage,omitempty" yaml:"usage"`
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
	Pending       Phase = "Pending"
	Running       Phase = "Running"
	Successed     Phase = "Succeeded"
	Failed        Phase = "Failed"
	Unknown       Phase = "Unknown"
	ReadyToDeploy Phase = "ReadyToDeploy"
	DeployCheck   Phase = "DeployCheck"
	ReadyToKill   Phase = "ReadyToKill"
	Killed        Phase = "Killed"
	Terminated    Phase = "Terminated"
	// 迁移相关状态
	CopyPending Phase = "CopyPending" //副本就绪状态-B
	Restoring   Phase = "Restoring"   //副本恢复任务状态-B
	Migrating   Phase = "Migrating"   //迁移状态-A
	Migrated    Phase = "Migrated"    //迁移完成状态-A
	Init        Phase = "Init"
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

type ValueType string

const (
	BoolType    ValueType = "bool"
	StringType  ValueType = "string"
	IntegerType ValueType = "integer"
	DoubleType  ValueType = "double"
	URLType     ValueType = "url"
	ComposeType ValueType = "compose"
)

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

// 符号判断
type ResultType string

const (
	True     ResultType = "True"
	False    ResultType = "False"
	NotReady ResultType = "NotReady"
)

// 流程执行条件
// LeftValue ==或!= RightValue
// 输出结果为Bool类型的值
// TODO: Value格式检查和调整，比如存在空格的情况
type ConditionFormula struct {
	LeftValue  Value `json:"left_value,omitempty" yaml:"left_value"`
	RightValue Value `json:"right_value,omitempty" yaml:"right_value"`
	// == 或 !=
	Signal SignalType `json:"signal,omitempty" yaml:"signal"`
	// 在条件串中的期望结果
	// 类型包含 and 或者 or
	Join JoinType `json:"join,omitempty" yaml:"join"`
	// 符号判断结果
	Result ResultType `json:"result_type,omitempty" yaml:"result_type"`
}

// 不建议使用过于复杂的逻辑
// 只支持逻辑的串行连接
// 如果没有Condition,Conditions默认值为true
// Example: Condition[1] and/or Condition[2] and/or Condition[3] ......
type Conditions struct {
	Formulas []ConditionFormula `json:"formulas,omitempty" yaml:"formulas"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
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
	// 该Name不会改变
	// 如果TypeMeta.Name提供，则使用TypeMeta.Name
	// 否则，TypeMeta.Name = Spec.Name + UUID
	Name string `json:"name,omitempty" yaml:"name"`

	// 对工作流的描述
	// +optional
	Desc *Description `json:"desc,omitempty" yaml:"desc"`

	// 工作流下有哪些任务
	// 目前只支持Task=>Group=>Action
	// TODO: 更为宽松的Task关系定义
	// 目前与Action的定义方式类似
	Tasks []TaskSpec `json:"tasks,omitempty" yaml:"tasks"`
	// TODO: Tasks的拓扑关系
}

type WorkflowStatus struct {
	//
	Phase Phase `json:"phase,omitempty" yaml:"phase"`

	//
	Tasks map[string]ObjectReference `json:"tasks,omitempty" yaml:"tasks"`

	// 创建时间
	CreateAt *Time `json:"create,omitempty" yaml:"create"`
	// 执行时间
	StartAt *Time `json:"start,omitempty" yaml:"start"`
	// 结束时间
	FinishAt *Time `json:"finish,omitempty" yaml:"finish"`
	// 最新获取状态的时间
	LastTime *Time `json:"last_time,omitempty" yaml:"last_time"`
}

// --------- Task
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
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
	// +optional
	Desc *Description `json:"desc,omitempty" yaml:"desc"`

	// Task类型
	// +optional
	Type *ProcessType `json:"type,omitempty" yaml:"type"`

	// Task条件
	// +Optional
	Conditions *Conditions `json:"conditions,omitempty" yaml:"conditions"`

	// 存储当前Task中的所有Group
	// Group之间没有严格的依赖限制
	// 可以有多个Group作为Group的入口，支持多个图形结构
	// 支持并发执行多组Group
	// TODO: 增加Level支持
	Groups []GroupSpec `json:"groups,omitempty" yaml:"groups"`
}

type TaskStatus struct {
	//
	Phase Phase `json:"phase,omitempty" yaml:"phase"`
	//
	Belong *ObjectReference `json:"belong,omitempty" yaml:"belong"`

	// Prefix
	Prefix string `json:"prefix,omitempty" yaml:"prefix"`

	//
	Groups map[string]ObjectReference `json:"groups,omitempty" yaml:"groups"`

	// TODO: Events定义
	// 创建时间
	CreateAt *Time `json:"create,omitempty" yaml:"create"`
	// 执行时间
	StartAt *Time `json:"start,omitempty" yaml:"start"`
	// 结束时间
	FinishAt *Time `json:"finish,omitempty" yaml:"finish"`
	// 最新获取状态的时间
	LastTime *Time `json:"last_time,omitempty" yaml:"last_time"`
}

// ---------- Group
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
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

	// +Optional
	SchedulerName *string `json:"scheduler_name,omitempty" yaml:"name"`

	// Parents Name, 一个Group可以有多个Parents
	Parents []string `json:"parents,omitempty" yaml:"parents"`

	// Group描述
	// +Optional
	Desc *Description `json:"desc,omitempty" yaml:"desc"`

	// Group类型
	// +Optional
	Type *ProcessType `json:"type,omitempty" yaml:"type"`

	// Group条件
	// +Optional
	Conditions *Conditions `json:"conditions,omitempty" yaml:"conditions"`

	// Group所需资源需要先遍历自己的Action
	//   Ref类型的指针需要根据Action中的资源需求计算
	//   TODO: 对于Condition类的节点，使用资源的预估
	//   TODO: 资源需求

	// 存储当前Group所属的所有Action
	// Group中可以只有一个Action, 简化实现逻辑
	Actions []ActionSpec `json:"actions,omitempty" yaml:"actions"`
	// TODO: 生成时是否可以直接分析依赖关系？
	// TODO: Action的依赖关系描述, 需要先遍历Actions构建DAG图
	// Group中的Action需要有较严格的依赖顺序，可以支持分支,条件,循环
	// 只能有一个Action作为入口Action,图形结构

	ResourceRequirements []ResourceRequirement `json:"resource_requirements,omitempty" yaml:"resource_requirements"`

	SkipScorePlugins []string `json:"skip_score_plugins,omitempty" yaml:"skip_score_plugins"`

	SkipFilterPlugins []string `json:"skip_filter_plugins,omitempty" yaml:"skip_filter_plugins"`
	//添加-hzy
	Replicas []int32 `json:"replicas,omitempty" yaml:"replicas"` //group的副本数量，用户需要输入,例如：[2,0] 第一个值表示本域想部署的副本数量，第二个值表示其他域想部署的副本数量

	// +Optional
	IsCopy   bool              `json:"is_copy,omitempty" yaml:"is_copy"`     //标记当前group是否是副本
	CopyInfo map[string]string `json:"copy_info,omitempty" yaml:"copy_info"` //存放任务的副本信息的  key：副本的ObjectMeta.Name  value:副本在本域还是在哪个域  如果是本域为："local" ,如果是跨域，则为连接那个域的ip或者是XX（待定）

	//亲和节点，如果该字段不为空的话，那么group就必须放在这些节点上执行
	AffinityNodes []string `json:"affinity_nodes,omitempty" yaml:"affinity_nodes"`
}

type ResourceRequirement struct {
	Name       string `json:"name,omitempty" yaml:"name"`
	Lowbound   string `json:"lowbound,omitempty" yaml:"lowbound"`
	Upperbound string `json:"upperbound,omitempty" yaml:"upperbound"`
}

type GroupStatus struct {
	//
	Phase Phase `json:"phase,omitempty" yaml:"phase"`

	Belong *ObjectReference `json:"belong,omitempty" yaml:"belong"`

	// TODO: 所属Actions的状态
	Actions map[string]ObjectReference `json:"actions,omitempty" yaml:"actions"`

	// TODO: 整体资源使用情况

	// TODO: Events定义
	//部署在哪个节点
	Node *string `json:"node,omitempty" yaml:"node"`
	// 创建时间
	CreateAt *Time `json:"create,omitempty" yaml:"create"`
	// 执行时间
	StartAt *Time `json:"start,omitempty" yaml:"start"`
	// 结束时间
	FinishAt *Time `json:"finish,omitempty" yaml:"finish"`
	// 最新获取状态的时间
	LastTime *Time `json:"last_time,omitempty" yaml:"last_time"`
	//添加-hzy
	CheckDependencyCount int32  `json:"check_dependency_count,omitempty" yaml:"check_dependency_count"`
	CopyStatus           string `json:"copy_status,omitempty" yaml:"copy_status"` //如果是源任务，这个参数可以标记其副本任务的执行状态    如果是副本任务，这个参数可以标记其是预部署还是说直接切换
}

// ---------- Action
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
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
	// +Optional
	Desc *Description `json:"desc,omitempty" yaml:"desc"`

	// Action类型
	// +Optional
	Type *ProcessType `json:"type,omitempty" yaml:"type"`

	// Action条件
	// +Optional
	Conditions *Conditions `json:"conditions,omitempty" yaml:"conditions"`

	// 需要的执行环境
	// 串行执行Runtime中的运行环境
	// 如果使用Pod方式部署的任务，建议只部署一个阻塞的Runtime
	// 可以用于执行命令，环境准备等操作
	Runtimes []RuntimeSpec `json:"runtimes,omitempty" yaml:"runtimes"`

	// 其他配置/选项
	// 细粒度任务控制
	// TODO: RPC相关接口实现
	EnableFineGrainedControl *bool `json:"enable_control,omitempty" yaml:"enable_control"`
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
	//添加-hzy ---这个要讨论是否有该选项，被删除了？
	ByWasm RuntimeType = "wasm"
	ByK8s  RuntimeType = "k8s"
)

// 环境变量
// TODO: 环境变量格式定义
type EnvVar struct {
	Name  string `json:"name,omitempty" yaml:"name"`
	Value string `json:"value,omitempty" yaml:"value"`
	// TODO: 动态获取相关字段
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type Resource_Node struct {
	meta.TypeMeta

	meta.ObjectMeta

	Spec ResourceSpec `json:"spec,omitempty" yaml:"spec"`

	Status ResourceStatus `json:"status,omitempty" yaml:"status"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type Data struct {
	meta.TypeMeta

	meta.ObjectMeta

	Spec DataSpec `json:"spec,omitempty" yaml:"spec"`

	Status DataStatus `json:"status,omitempty" yaml:"status"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
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
	ExpectedValue     *float64      `json:"expected_value,omitempty" yaml:"expected_value"`
	ExpectedValueUnit *ResourceUnit `json:"expected_value_unit,omitempty" yaml:"expected_value_unit"`

	// 描述资源的固有属性
	Type     *ResourceType   `json:"type,omitempty" yaml:"type"`         // 类型：网络、计算、存储
	Name     *string         `json:"name,omitempty" yaml:"name"`         // 名称：cpu、gpu、内存、硬盘
	Unit     *ResourceUnit   `json:"unit,omitempty" yaml:"unit"`         // 资源的单位
	Detail   *ResourceDetail `json:"detail,omitempty" yaml:"detail"`     // 资源的细致描述
	Capacity *float64        `json:"capacity,omitempty" yaml:"capacity"` // 资源的总量
	NodeId   *string         `json:"node_id,omitempty" yaml:"node_id"`   // 资源属于哪个node
}

type ResourceDetail struct {
	Type string
}

type ResourceStatus struct {
	Name string
	// 资源的使用量和他的单位
	// TODO
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
	CPUPercentage   ResourceUnit = "percent"
)

// 增加设备定义
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
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
	DeviceComplete     DevicePhase = "Complete"
	DeviceDisconnected DevicePhase = "Disconnected"
)

// 对设备能力的描述
type DeviceDesc struct {
	Label []string `json:"label,omitempty" yaml:"label"`
	// +Optional
	Docs *string `json:"docs,omitempty" yaml:"docs"`
	// +Optional
	// 厂家
	Maker *string `json:"maker,omitempty" yaml:"maker"`
	// +Optional
	// 型号
	Model *string `json:"model,omitempty" yaml:"model"`
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
	Type AccessType `json:"type,omitempty" yaml:"type"`

	// 访问方式对应的URL
	//   对于Customize类型，URL对应部署脚本的地址
	URL string `json:"url,omitempty" yaml:"url"`

	// 设备组
	//  对于RMF类型，Group对应RMF Fleets
	Group *string `json:"group,omitempty" yaml:"group"`

	// 设备别名
	//   部分情况下，RMF中的设备名与系统中设备名不一致
	//   默认情况下，Alias应该与Name相同
	Alias *string `json:"alias,omitempty" yaml:"alias"`
}

// 设备的属性
// TODO: 任务部署时，通过类似Device.Property的方式来寻址
type Property struct {
	// Type
	// 缺省值为string
	Type *ValueType `json:"type,omitempty" yaml:"type"`

	Value string `json:"value,omitempty" yaml:"value"`

	// 子属性
	// Key: 属性名称 Value: 属性值
	SubProperty map[string]SubProperty `json:"sub_property,omitempty" yaml:"sub_property"`
}

// Compose类型的设备属性
// 设备属性的子属性，如Location类型，具有子属性Location.X
type SubProperty struct {
	// Type
	Type ValueType `json:"type,omitempty" yaml:"type"`

	// TODO: 格式校验
	Value string `json:"value,omitempty" yaml:"value"`
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
	Type LockType `json:"type,omitempty" yaml:"type"`

	// 调度时 ref为0时释放
	IsLocked bool `json:"is_locked,omitempty" yaml:"is_locked"`

	// 资源引用数 部署时
	Ref int `json:"ref,omitempty" yaml:"ref"`
}

// 设备事件描述
type DeviceEvent struct {
	// EventCode, 时间码，对应事件处理的方案
	Code int `json:"code,omitempty" yaml:"code"`

	// 事件描述
	Desc string `json:"desc,omitempty" yaml:"desc"`

	// TODO: 额外参数, 设备的上下文状态
}

type DeviceSpec struct {
	// 设备名称，每个Node上的设备，名称应该唯一
	Name string `json:"name,omitempty" yaml:"name"`

	// 对设备的描述
	Desc *DeviceDesc `json:"desc,omitempty" yaml:"desc"`

	// 设备的关联Node，每个设备需要与一个Node相关联
	Node *string `json:"node,omitempty" yaml:"node"`

	// 设备的访问方式
	AccessMethod *AccessMethod `json:"access_method,omitempty" yaml:"access_method"`

	// 父设备
	AttachedDevice *string `json:"attached_device,omitempty" yaml:"attached_device"`

	// 子设备
	SubDevices []string `json:"sub_devices,omitempty" yaml:"sub_devices"`

	// 设备的期望属性
	// Key: 属性名 Value: 实际的参数值
	ExpectedProperties map[string]Property `json:"expected_properties,omitempty" yaml:"expected_properties"`

	// 设备具有的能力
	// 这里的Name是统一的Name,比如Grab、Move等
	Abilities []string `json:"abilities,omitempty" yaml:"abilities"`
}

// Ability 描述一个能力
type Ability struct {
	// 这里的Name是实际的Name，包括Ability的Domain，如Grab.Leju.Guochuang
	// 如Move.Leju.Guochuang
	Name string `json:"name,omitempty" yaml:"name"`
	Desc *string
	// 一个能力对应的多个业务（技能）
	Services   map[string]AbilityService `json:"services,omitempty" yaml:"services"`
	InstanceID *string                   `json:"instance_id,omitempty" yaml:"instance_id"`
	State      *AbilityState             `json:"state,omitempty" yaml:"state"`
	Status     *string                   `json:"status,omitempty" yaml:"status"`
}

// AbilityService 描述一个能力的具体业务（技能）
type AbilityService struct {
	Desc      *string `json:"desc,omitempty" yaml:"desc"`
	Ip        *string `json:"ip,omitempty" yaml:"ip"`
	Port      *string `json:"port,omitempty" yaml:"port"`
	Interface *string `json:"interface,omitempty" yaml:"interface"`
	Model     *string `json:"model,omitempty" yaml:"model"`
}

// TODO: 增加具体的值限制
type AbilityState int

type DeviceStatus struct {
	Abilities map[string]Ability `json:"abilities,omitempty" yaml:"abilities"`

	// 正在使用Device的Runtime
	Runtime ObjectReference `json:"runtime,omitempty" yaml:"runtime"`

	// 设备的运行阶段
	Phase DevicePhase `json:"phase,omitempty" yaml:"phase"`

	// 运行时中，设备的实际状态
	// 当Phase与Status不一致时，机器人出现运行错误
	Status string `json:"status,omitempty" yaml:"status"`

	// 设备的实际属性
	Properties map[string]Property `json:"properties,omitempty" yaml:"properties"`

	// 设备资源锁状态
	Lock Lock `json:"lock,omitempty" yaml:"lock"`

	// 设备事件描述
	Events []DeviceEvent `json:"events,omitempty" yaml:"events"`

	// 上次成功获取设备状态的时间
	// 如果长时间不能获取设备的状态，则认为设备离线
	LastTime Time `json:"last_time,omitempty" yaml:"last_time"`
}

// SceneSpec 描述scene的固有属性和期待属性
type SceneSpec struct {
	// 每一个scene的标识
	SceneID string `json:"scene_id,omitempty" yaml:"scene_id"`

	// scene的类型 是一个地点还是一个物品
	Type SceneType `json:"type,omitempty" yaml:"type"`

	// 期待属性
	ExpectedProperty map[string]Property `json:"expected_property,omitempty" yaml:"expected_property"`

	// 场景的描述（不可变属性）
	Desc SceneDesc `json:"desc,omitempty" yaml:"desc"`
}

type SceneDesc struct {
	Label []string          `json:"label,omitempty" yaml:"label"`
	Value map[string]string `json:"value,omitempty" yaml:"value"`
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
	UpdateMethod string `json:"update_method,omitempty" yaml:"update_method"`
	UpdateTime   Time   `json:"update_time,omitempty" yaml:"update_time"`

	// 关联的场景
	AttachedScene string `json:"attached_scene,omitempty" yaml:"attached_scene"`

	// 关联的设备
	AttachedDevice string `json:"attached_device,omitempty" yaml:"attached_device"`

	// 关联的任务
	AttachedTask string `json:"attached_task,omitempty" yaml:"attached_task"`

	// 实时属性
	Property map[string]Property `json:"property,omitempty" yaml:"property"`

	// 锁
	Lock Lock `json:"lock,omitempty" yaml:"lock"`
}
type DataSpec struct {
	// 对于文件类型的Data
	// 文件格式
	// 文件大小
	// SHA文件校验
}

type DataStatus struct{}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type Runtime struct {
	//
	meta.TypeMeta

	//
	meta.ObjectMeta

	//
	Spec RuntimeSpec `json:"spec,omitempty" yaml:"spec"`

	//
	Status RuntimeStatus `json:"status,omitempty" yaml:"status"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type RuntimeList struct {
	meta.TypeMeta
	meta.ListMeta
	Items []Runtime `json:"items" yaml:"items"`
}

// Action所需执行环境
type RuntimeSpec struct {
	// 标识运行时的名称
	Name string `json:"name,omitempty" yaml:"name"`

	//
	Parents []string `json:"parents,omitempty" yaml:"parents"`

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

	// Runtime条件
	// +Optional
	Conditions *Conditions `json:"conditions,omitempty" yaml:"conditions"`

	// 需要的数据
	// 输入数据
	// 	输入数据作为参数注入到命令参数中
	Inputs []Value `json:"inputs,omitempty" yaml:"inputs"`

	// 输出数据
	//  输出数据作为参数注入到命令参数中
	Outputs []Value `json:"outputs,omitempty" yaml:"outputs"`

	//Waiting                      bool     `json:"waiting" yaml:"waiting"`
	EnableFineGrainedControl        bool    `json:"enable_control,omitempty" yaml:"enable_control"`
	EnableFineGrainedControlService *string `json:"enable_control_service,omitempty" yaml:"enable_control_service"`
	EnableFineGrainedControlPort    *string `json:"enable_control_port,omitempty" yaml:"enable_control_port"`
	//-hzy暂时添加
	//Labels      map[string]string `json:"labels,omitempty" yaml:"labels"`           // 用于模板的 labels 配置
	//Selector    map[string]string `json:"selector,omitempty" yaml:"selector"`       // Deployment/Service 选择器
	//Ports       []Port            `json:"ports,omitempty" yaml:"ports"`             // 容器/服务端口
	//ServiceType *string           `json:"serviceType,omitempty" yaml:"serviceType"` // 服务类型，例如 ClusterIP
	//TargetPorts []int             `json:"targetPorts,omitempty" yaml:"targetPorts"` // 目标端口映射
	//Replicas    *int32            `json:"replicas,omitempty" yaml:"replicas"`       // 用于 Deployment 副本数量
	//Pod         *Pod              `json:"pod,omitempty" yaml:"pod"`                 // 如果是Pod，则放入该参数
	//Service     *Service          `json:"service,omitempty" yaml:"service"`
	//Deployment  *Deployment       `json:"deployment,omitempty" yaml:"deployment"`
	//ysk添加
	Dependency *string       `json:"dependency,omitempty" yaml:"dependency"` //依赖文件的地址，后续改成多种依赖
	Packages   []Requirement `json:"package,omitempty" yaml:"package"`       //解析之后的包
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
	DeviceData  DataType = "device"
)

// 值类型，表示数据使用
type Value struct {
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
	Value     string    `json:"value,omitempty" yaml:"value"`
	ValueType ValueType `json:"value_type,omitempty" yaml:"value_type"`
	From      string    `json:"from,omitempty" yaml:"from"`
}

type ActionStatus struct {
	Belong *ObjectReference `json:"belong,omitempty" yaml:"belong"`

	// 生命周期
	Phase Phase `json:"phase,omitempty" yaml:"phase"`
	// 当前Runtime执行状态
	// Key是Action中Runtime.Spec.Name，可能与真实的名字不同
	Runtimes map[string]ObjectReference `json:"runtimes,omitempty" yaml:"runtimes"` // TODO: 修改为Map
	// 任务实际的执行结果
	// TODO: 后续增加单独字段定义，使用Results来替代该部分内容
	Outputs map[string]Value `json:"results,omitempty" yaml:"results"`
	// TODO: Events定义
	CreateAt *Time `json:"create,omitempty" yaml:"create"`
	// 执行时间
	StartAt *Time `json:"start,omitempty" yaml:"start"`
	// 结束时间
	FinishAt *Time `json:"finish,omitempty" yaml:"finish"`
	// 最新获取状态的时间
	LastTime *Time `json:"last_time,omitempty" yaml:"last_time"`
	//增加一个参数-hzy
	Waiting    bool   `json:"waiting,omitempty" yaml:"waiting"`
	CopyStatus string `json:"copy_status,omitempty" yaml:"copy_status"`
}

type RuntimeStatus struct {
	// 输出的结果
	Outputs map[string]Value `json:"outputs,omitempty" yaml:"outputs"`
	//
	Belong *ObjectReference `json:"belong,omitempty" yaml:"belong"`
	// 当前资源使用情况
	Resources map[string]ObjectReference `json:"resources,omitempty" yaml:"resources"` // TODO: 修改为Map
	// 当前设备使用情况
	Devices map[string]ObjectReference `json:"devices,omitempty" yaml:"devices"`
	// 当前数据使用情况
	Data map[string]ObjectReference `json:"data,omitempty" yaml:"data"`
	// 当前场景更新情况
	Scenes map[string]ObjectReference `json:"scenes,omitempty" yaml:"scenes"`
	// 当前任务的执行情况
	// 任务在哪里执行,进程ID
	NodeName  *string `json:"node_name,omitempty" yaml:"node_name"`
	ProcessId *string `json:"process_id,omitempty" yaml:"process_id"`

	// 执行状态
	Phase Phase `json:"phase,omitempty" yaml:"phase"`

	// 创建时间
	CreateAt *Time `json:"create,omitempty" yaml:"create"`
	// 执行时间
	StartAt *Time `json:"start,omitempty" yaml:"start"`
	// 结束时间
	FinishAt *Time `json:"finish,omitempty" yaml:"finish"`
	// 最新获取状态的时间
	LastTime *Time `json:"last_time,omitempty" yaml:"last_time"`
	//增加一个参数0hzy
	Waiting            bool   `json:"waiting,omitempty" yaml:"waiting"`
	Initing            bool   `json:"initing,omitempty" yaml:"initing"`
	KeyStatus          string `json:"key_status,omitempty" yaml:"key_status"`
	CopyStatus         string `json:"copy_status,omitempty" yaml:"copy_status"`
	IsDependencySatisf bool   `json:"dependency_satisf,omitempty" yaml:"dependency_satisf"`
	IsParsed           bool   `json:"isparsed,omitempty" yaml:"isparsed"`             //是否已经被解析过
	DepenPreparing     bool   `json:"sepenPreparing,omitempty" yaml:"DepenPreparing"` //是否正在创建虚拟环境，防止多次创建
}

// 任务的输出结果
// 单独作为一个资源，方便其他节点获取该资源
// 数据结果的类型---
type Result struct {
	//Result唯一表示，由ActionID
	ResultID string `json:"result_id,omitempty" yaml:"result_id"`
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
	Protocol   string `json:"protocol,omitempty" yaml:"protocol"`
	Port       int    `json:"port,omitempty" yaml:"port"`
	TargetPort int    `json:"target_port,omitempty" yaml:"target_port"`
}

// 依赖解析成包的结构体
type Requirement struct {
	Name    string
	Version string
}

type conditionType string

const (
	NodeDependency     conditionType = "NodeDependency"
	DataDependency     conditionType = "DataDependency"
	ResourceDependency conditionType = "ResourceDependency"
	ProgramDependency  conditionType = "ProgramDependency"
)

//// 系统运行时相关信息，表示长时间部署执行的内容
//// 云边侧模块使用相关字段
//// TODO: Pod From K8s
//type Pod struct {
//	meta.TypeMeta
//	meta.ObjectMeta
//	Spec PodSpec `json:"spec,omitempty" yaml:"spec"`
//}
//type PodSpec struct {
//	Volumes            []Volume          `json:"volumes,omitempty" yaml:"volumes"`                           // Pod挂载的存储卷列表（如ConfigMap、Secret、PVC等）
//	InitContainers     []Container       `json:"init_containers,omitempty" yaml:"init_containers"`           // 初始化容器（主容器启动前运行）
//	Containers         []Container       `json:"containers" yaml:"containers"`                               // 应用容器列表（必须至少一个）
//	RestartPolicy      RestartPolicy     `json:"restart_policy,omitempty" yaml:"restart_policy"`             // 容器失败重启策略（Always/OnFailure/Never）
//	DnsPolicy          DNSPolicy         `json:"dns_policy,omitempty" yaml:"dns_policy"`                     // DNS解析策略（ClusterFirst/Default/None）
//	NodeSelector       map[string]string `json:"node_selector,omitempty" yaml:"node_selector"`               // 节点标签选择器（强制调度到匹配节点）
//	ServiceAccountName string            `json:"service_account_name,omitempty" yaml:"service_account_name"` // 关联的ServiceAccount名称
//	Affinity           Affinity          `json:"affinity,omitempty" yaml:"affinity"`
//}
//type Volume struct {
//	Name         string                          `json:"name" yaml:"name"` // 存储卷名称
//	VolumeSource `json:",inline" yaml:",inline"` // 存储卷来源（如ConfigMap、Secret）
//}
//
//// VolumeSource 定义存储卷的数据来源（必须且常用的类型）
//type VolumeSource struct {
//	ConfigMap             ConfigMapVolumeSource             `json:"config_map,omitempty" yaml:"config_map,omitempty"`                           // 从ConfigMap挂载键值对到文件
//	Secret                SecretVolumeSource                `json:"secret,omitempty" yaml:"secret,omitempty"`                                   // 从Secret挂载敏感数据到文件
//	HostPath              HostPathVolumeSource              `json:"host_path,omitempty" yaml:"host_path,omitempty"`                             // 挂载宿主机目录或文件
//	EmptyDir              EmptyDirVolumeSource              `json:"empty_dir,omitempty" yaml:"empty_dir,omitempty"`                             // 临时空目录（Pod生命周期内有效）
//	PersistentVolumeClaim PersistentVolumeClaimVolumeSource `json:"persistent_volume_claim,omitempty" yaml:"persistent_volume_claim,omitempty"` // 挂载持久化存储卷（PVC）
//}
//
//// ConfigMapVolumeSource ConfigMap类型的存储卷配置
//type ConfigMapVolumeSource struct {
//	Name  string      `json:"name" yaml:"name"`                       // ConfigMap名称（必须字段）
//	Items []KeyToPath `json:"items,omitempty" yaml:"items,omitempty"` // 选择特定键挂载为文件
//}
//
//// SecretVolumeSource Secret类型的存储卷配置
//type SecretVolumeSource struct {
//	SecretName string      `json:"secret_name" yaml:"secret_name"`         // Secret名称（必须字段）
//	Items      []KeyToPath `json:"items,omitempty" yaml:"items,omitempty"` // 选择特定键挂载为文件
//}
//
//// HostPathVolumeSource 宿主机目录挂载配置
//type HostPathVolumeSource struct {
//	Path string       `json:"path" yaml:"path"`                     // 宿主机绝对路径（必须字段）
//	Type HostPathType `json:"type,omitempty" yaml:"type,omitempty"` // 路径类型检查（如DirectoryOrCreate）
//}
//
//// EmptyDirVolumeSource 临时空目录配置
//type EmptyDirVolumeSource struct {
//	Medium StorageMedium `json:"medium,omitempty" yaml:"medium,omitempty"` // 存储介质（Memory/默认空为磁盘）
//}
//
//// PersistentVolumeClaimVolumeSource 持久化存储卷声明配置
//type PersistentVolumeClaimVolumeSource struct {
//	ClaimName string `json:"claim_name" yaml:"claim_name"` // PVC名称（必须字段）
//}
//
//// 辅助结构体
//type KeyToPath struct {
//	Key  string `json:"key" yaml:"key"`   // ConfigMap/Secret中的键名
//	Path string `json:"path" yaml:"path"` // 挂载目标路径（如"config.yaml"）
//}
//
//// 枚举类型
//type HostPathType string
//
//const (
//	HostPathDirectoryOrCreate HostPathType = "DirectoryOrCreate" // 目录不存在则创建
//	HostPathDirectory         HostPathType = "Directory"         // 必须存在目录
//	HostPathFileOrCreate      HostPathType = "FileOrCreate"      // 文件不存在则创建空文件
//	HostPathFile              HostPathType = "File"              // 必须存在文件
//)
//
//type StorageMedium string
//
//const (
//	StorageMediumDefault StorageMedium = ""       // 默认磁盘存储
//	StorageMediumMemory  StorageMedium = "Memory" // 内存临时存储（tmpfs）
//)
//
//type Container struct {
//	Name         string          `json:"name" yaml:"name"`                             // 容器名称（必须唯一）
//	Image        string          `json:"image" yaml:"image"`                           // 容器镜像地址（必须）
//	Command      []string        `json:"command,omitempty" yaml:"command"`             // 容器启动命令（覆盖镜像默认值）
//	Args         []string        `json:"args,omitempty" yaml:"args"`                   // 容器启动参数
//	Ports        []ContainerPort `json:"ports,omitempty" yaml:"ports"`                 // 容器暴露的端口
//	Env          []EnvVar        `json:"env,omitempty" yaml:"env"`                     // 容器环境变量
//	VolumeMounts []VolumeMount   `json:"volume_mounts,omitempty" yaml:"volume_mounts"` // 存储卷挂载配置
//}
//
//// 存储卷挂载点配置
//type VolumeMount struct {
//	Name      string `json:"name" yaml:"name"`                     // 引用的存储卷名称（必须与volumes[*].name对应）
//	MountPath string `json:"mount_path" yaml:"mount_path"`         // 容器内的挂载路径（必须）
//	ReadOnly  bool   `json:"read_only,omitempty" yaml:"read_only"` // 是否以只读方式挂载（默认false）
//}
//
//// 容器端口配置
//type ContainerPort struct {
//	ContainerPort int32    `json:"container_port" yaml:"container_port"` // 容器内监听端口（必须）
//	Protocol      Protocol `json:"protocol,omitempty" yaml:"protocol"`   // 协议类型（TCP/UDP，默认TCP）
//}
//type Protocol string
//
//const (
//	// ProtocolTCP is the TCP protocol.
//	ProtocolTCP Protocol = "TCP"
//	// ProtocolUDP is the UDP protocol.
//	ProtocolUDP Protocol = "UDP"
//	// ProtocolSCTP is the SCTP protocol.
//	ProtocolSCTP Protocol = "SCTP"
//)
//
//// 污点容忍规则
//type Toleration struct {
//	Key      string             `json:"key,omitempty" yaml:"key"`       // 污点键名
//	Operator TolerationOperator `json:"operator" yaml:"operator"`       // 匹配操作符（Exists/Equal）
//	Effect   TaintEffect        `json:"effect,omitempty" yaml:"effect"` // 污点效果（NoSchedule/NoExecute）
//}
//type TaintEffect string
//
//const (
//	// Do not allow new pods to schedule onto the node unless they tolerate the taint,
//	// but allow all pods submitted to Kubelet without going through the scheduler
//	// to start, and allow all already-running pods to continue running.
//	// Enforced by the scheduler.
//	TaintEffectNoSchedule TaintEffect = "NoSchedule"
//	// Like TaintEffectNoSchedule, but the scheduler tries not to schedule
//	// new pods onto the node, rather than prohibiting new pods from scheduling
//	// onto the node entirely. Enforced by the scheduler.
//	TaintEffectPreferNoSchedule TaintEffect = "PreferNoSchedule"
//	// NOT YET IMPLEMENTED. TODO: Uncomment field once it is implemented.
//	// Like TaintEffectNoSchedule, but additionally do not allow pods submitted to
//	// Kubelet without going through the scheduler to start.
//	// Enforced by Kubelet and the scheduler.
//	// TaintEffectNoScheduleNoAdmit TaintEffect = "NoScheduleNoAdmit"
//
//	// Evict any already-running pods that do not tolerate the taint.
//	// Currently enforced by NodeController.
//	TaintEffectNoExecute TaintEffect = "NoExecute"
//)
//
//// A toleration operator is the set of operators that can be used in a toleration.
//// +enum
//type TolerationOperator string
//
//const (
//	TolerationOpExists TolerationOperator = "Exists"
//	TolerationOpEqual  TolerationOperator = "Equal"
//)
//
//type RestartPolicy string
//
//const (
//	RestartPolicyAlways    RestartPolicy = "Always"
//	RestartPolicyOnFailure RestartPolicy = "OnFailure"
//	RestartPolicyNever     RestartPolicy = "Never"
//)
//
//type DNSPolicy string
//
//const (
//	// DNSClusterFirstWithHostNet indicates that the pod should use cluster DNS
//	// first, if it is available, then fall back on the default
//	// (as determined by kubelet) DNS settings.
//	DNSClusterFirstWithHostNet DNSPolicy = "ClusterFirstWithHostNet"
//
//	// DNSClusterFirst indicates that the pod should use cluster DNS
//	// first unless hostNetwork is true, if it is available, then
//	// fall back on the default (as determined by kubelet) DNS settings.
//	DNSClusterFirst DNSPolicy = "ClusterFirst"
//
//	// DNSDefault indicates that the pod should use the default (as
//	// determined by kubelet) DNS settings.
//	DNSDefault DNSPolicy = "Default"
//
//	// DNSNone indicates that the pod should use empty DNS settings. DNS
//	// parameters such as nameservers and search paths should be defined via
//	// DNSConfig.
//	DNSNone DNSPolicy = "None"
//)
//
//type Affinity struct {
//	NodeAffinity    NodeAffinity    `json:"node_affinity,omitempty" yaml:"node_affinity"`
//	PodAffinity     PodAffinity     `json:"pod_affinity,omitempty" yaml:"pod_affinity"`
//	PodAntiAffinity PodAntiAffinity `json:"pod_anti_affinity,omitempty" yaml:"pod_anti_affinity"`
//}
//
//// Pod亲和性调度规则
//type PodAffinity struct {
//	RequiredDuringSchedulingIgnoredDuringExecution  []PodAffinityTerm         `json:"required_during_scheduling_ignored_during_execution,omitempty" yaml:"required_during_scheduling_ignored_during_execution"`
//	PreferredDuringSchedulingIgnoredDuringExecution []WeightedPodAffinityTerm `json:"preferred_during_scheduling_ignored_during_execution,omitempty" yaml:"preferred_during_scheduling_ignored_during_execution"`
//}
//
//// Pod反亲和性调度规则
//type PodAntiAffinity struct {
//	RequiredDuringSchedulingIgnoredDuringExecution  []PodAffinityTerm         `json:"required_during_scheduling_ignored_during_execution,omitempty" yaml:"required_during_scheduling_ignored_during_execution"`
//	PreferredDuringSchedulingIgnoredDuringExecution []WeightedPodAffinityTerm `json:"preferred_during_scheduling_ignored_during_execution,omitempty" yaml:"preferred_during_scheduling_ignored_during_execution"`
//}
//
//// 带权重的Pod亲和性规则
//type WeightedPodAffinityTerm struct {
//	Weight          int32           `json:"weight" yaml:"weight"`
//	PodAffinityTerm PodAffinityTerm `json:"pod_affinity_term" yaml:"pod_affinity_term"`
//}
//
//// Pod亲和性规则条件
//type PodAffinityTerm struct {
//	LabelSelector     meta.LabelSelector `json:"label_selector,omitempty" yaml:"label_selector"`
//	Namespaces        []string           `json:"namespaces,omitempty" yaml:"namespaces"`
//	TopologyKey       string             `json:"topology_key" yaml:"topology_key"`
//	NamespaceSelector meta.LabelSelector `json:"namespace_selector,omitempty" yaml:"namespace_selector"`
//}
//
//// 节点亲和性调度规则
//type NodeAffinity struct {
//	RequiredDuringSchedulingIgnoredDuringExecution  NodeSelector              `json:"required_during_scheduling_ignored_during_execution,omitempty" yaml:"required_during_scheduling_ignored_during_execution"`
//	PreferredDuringSchedulingIgnoredDuringExecution []PreferredSchedulingTerm `json:"preferred_during_scheduling_ignored_during_execution,omitempty" yaml:"preferred_during_scheduling_ignored_during_execution"`
//}
//type PreferredSchedulingTerm struct {
//	// Weight associated with matching the corresponding nodeSelectorTerm, in the range 1-100.
//	Weight int32 `json:"weight,omitempty" yaml:"weight"`
//	// A node selector term, associated with the corresponding weight.
//	Preference NodeSelectorTerm `json:"preference,omitempty" yaml:"preference"`
//}
//
//// 节点选择器
//type NodeSelector struct {
//	NodeSelectorTerms []NodeSelectorTerm `json:"node_selector_terms" yaml:"node_selector_terms"`
//}
//
//// 节点选择条件
//type NodeSelectorTerm struct {
//	MatchExpressions []NodeSelectorRequirement `json:"match_expressions,omitempty" yaml:"match_expressions"`
//	MatchFields      []NodeSelectorRequirement `json:"match_fields,omitempty" yaml:"match_fields"`
//}
//
//// 节点选择器条件
//type NodeSelectorRequirement struct {
//	Key      string               `json:"key" yaml:"key"`
//	Operator NodeSelectorOperator `json:"operator" yaml:"operator"`
//	Values   []string             `json:"values,omitempty" yaml:"values"`
//}
//type NodeSelectorOperator string
//
//const (
//	NodeSelectorOpIn           NodeSelectorOperator = "In"
//	NodeSelectorOpNotIn        NodeSelectorOperator = "NotIn"
//	NodeSelectorOpExists       NodeSelectorOperator = "Exists"
//	NodeSelectorOpDoesNotExist NodeSelectorOperator = "DoesNotExist"
//	NodeSelectorOpGt           NodeSelectorOperator = "Gt"
//	NodeSelectorOpLt           NodeSelectorOperator = "Lt"
//)
//
//// TODO: Service From K8s
//type Service struct {
//	meta.TypeMeta
//	meta.ObjectMeta
//	Spec ServiceSpec `json:"spec,omitempty" yaml:"spec"`
//}
//type ServiceSpec struct {
//	Type     ServiceType       `json:"type,omitempty" yaml:"type"`         // 服务类型：ClusterIP/NodePort/LoadBalancer/ExternalName
//	Selector map[string]string `json:"selector,omitempty" yaml:"selector"` // 后端Pod标签选择器（必需字段）
//
//	Ports []ServicePort `json:"ports,omitempty" yaml:"ports"` // 服务端口映射列表（至少一个）
//
//	SessionAffinity SessionAffinity `json:"sessionAffinity,omitempty" yaml:"session_affinity"` // 会话亲和性（None/ClientIP）
//	ClusterIP       string          `json:"clusterIP,omitempty" yaml:"cluster_ip"`             // 虚拟IP地址（留空自动分配）
//
//	ExternalIPs  []string `json:"externalIPs,omitempty" yaml:"external_ips"`   // 外部可达的IP地址列表（非云环境使用）
//	ExternalName string   `json:"externalName,omitempty" yaml:"external_name"` // ExternalName类型时指向的外部服务域名
//
//	ExternalTrafficPolicy string `json:"externalTrafficPolicy,omitempty" yaml:"external_traffic_policy"` // 外部流量策略（Local/Cluster）
//}
//
//// 服务类型枚举
//type ServiceType string
//
//const (
//	ServiceTypeClusterIP    ServiceType = "ClusterIP"    // 集群内部访问（默认）
//	ServiceTypeNodePort     ServiceType = "NodePort"     // 通过节点端口暴露
//	ServiceTypeLoadBalancer ServiceType = "LoadBalancer" // 云厂商负载均衡器
//	ServiceTypeExternalName ServiceType = "ExternalName" // 映射到外部服务
//)
//
//type ServicePort struct {
//	Name       string      `json:"name,omitempty" yaml:"name"`              // 端口名称（DNS_LABEL格式）
//	Protocol   Protocol    `json:"protocol,omitempty" yaml:"protocol"`      // 协议类型（TCP/UDP，默认TCP）
//	Port       int32       `json:"port" yaml:"port"`                        // 服务暴露端口（必需）
//	TargetPort IntOrString `json:"targetPort,omitempty" yaml:"target_port"` // 容器监听端口（默认与Port相同）
//	NodePort   int32       `json:"nodePort,omitempty" yaml:"node_port"`     // NodePort类型时分配的节点端口
//}
//type IntOrString struct {
//	Type   Type   `protobuf:"varint,1,opt,name=type,casttype=Type"`
//	IntVal int32  `protobuf:"varint,2,opt,name=intVal"`
//	StrVal string `protobuf:"bytes,3,opt,name=strVal"`
//}
//type Type int64
//
//const (
//	Int    Type = iota // The IntOrString holds an int.
//	String             // The IntOrString holds a string.
//)
//
//// 会话亲和性类型
//type SessionAffinity string
//
//const (
//	SessionAffinityNone     SessionAffinity = "None"     // 不保持会话
//	SessionAffinityClientIP SessionAffinity = "ClientIP" // 基于客户端IP保持会话
//)
//
//// TODO: Deployment From K8s
//type Deployment struct {
//	meta.TypeMeta
//	meta.ObjectMeta
//	Spec DeploymentSpec `json:"spec,omitempty" yaml:"spec"`
//}
//type DeploymentSpec struct {
//	// 副本数量，指定期望运行的 Pod 副本数
//	Replicas int32 `json:"replicas,omitempty" yaml:"replicas,omitempty"`
//
//	// 标签选择器，用于匹配要管理的 Pod
//	Selector meta.LabelSelector `json:"selector" yaml:"selector"`
//
//	// Pod 模板定义（必须字段）
//	Template PodTemplateSpec `json:"template" yaml:"template"`
//
//	// 更新策略（默认 RollingUpdate）
//	Strategy DeploymentStrategy `json:"strategy,omitempty" yaml:"strategy,omitempty"`
//
//	// 新 Pod 就绪后需等待的秒数（默认 0）
//	MinReadySeconds int32 `json:"minReadySeconds,omitempty" yaml:"minReadySeconds,omitempty"`
//
//	// 保留的历史版本数量（用于回滚，默认 10）
//	RevisionHistoryLimit int32 `json:"revisionHistoryLimit,omitempty" yaml:"revisionHistoryLimit,omitempty"`
//
//	// 部署进度超时时间（秒，默认 600）
//	ProgressDeadlineSeconds int32 `json:"progressDeadlineSeconds,omitempty" yaml:"progressDeadlineSeconds,omitempty"`
//}
//type PodTemplateSpec struct {
//	meta.ObjectMeta
//	Spec PodSpec `json:"spec,omitempty" yaml:"spec"`
//}
//
//// 更新策略结构体
//type DeploymentStrategy struct {
//	// 策略类型：RollingUpdate 或 Recreate
//	Type DeploymentStrategyType `json:"type,omitempty" yaml:"type,omitempty"`
//
//	// 滚动更新配置
//	RollingUpdate RollingUpdateDeployment `json:"rollingUpdate,omitempty" yaml:"rollingUpdate,omitempty"`
//}
//type DeploymentStrategyType string
//
//const (
//	// Kill all existing pods before creating new ones.
//	RecreateDeploymentStrategyType DeploymentStrategyType = "Recreate"
//
//	// Replace the old ReplicaSets by new one using rolling update i.e gradually scale down the old ReplicaSets and scale up the new one.
//	RollingUpdateDeploymentStrategyType DeploymentStrategyType = "RollingUpdate"
//)
//
//// 滚动更新配置结构体
//type RollingUpdateDeployment struct {
//	// 最大超量 Pod 数（如 25% 或绝对数 2）
//	MaxSurge IntOrString `json:"maxSurge,omitempty" yaml:"maxSurge,omitempty"`
//
//	// 最大不可用 Pod 数（如 25% 或绝对数 1）
//	MaxUnavailable IntOrString `json:"maxUnavailable,omitempty" yaml:"maxUnavailable,omitempty"`
//}
//
//// TODO: VM From K8s
//type VM struct{}
//
//// 后续需要可以扩展

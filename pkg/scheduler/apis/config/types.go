package config

import (
	apis "hit.edu/framework/pkg/apis/cores"
	"k8s.io/apimachinery/pkg/util/sets"
	"time"
)

//TODO: 将K8s相关组件替换为我们自己的

const (
	DefaultSchedulerPort = 10000
)

type SchedulerConfiguration struct {
	Parallelism int32

	// TODO: Leader选举相关内容

	// 调度器配置
	Profiles []SchedulerProfile

	Extenders []Extender
}

type SchedulerProfile struct {
	// 框架中可能会有多个调度器，每个调度器分管不同的资源
	SchedulerName string

	// TODO: 打分相关参数
	Plugins *Plugins
}

type Plugins struct {
	PreSchedule  PluginSet
	Generate     PluginSet
	Schedule     PluginSet
	PostSchedule PluginSet
}

type PluginSet struct {
	Enabled  []Plugin
	Disabled []Plugin
}

type Plugin struct {
	Name string
}

type PluginConfig struct {
	Name string
	// Args runtime.Object
}

// 返回所有Enable的插件
func (p *Plugins) Names() []string {
	if p == nil {
		return nil
	}
	extensions := []PluginSet{
		p.PreSchedule,
		p.Generate,
		p.Schedule,
		p.PostSchedule,
	}
	n := sets.New[string]()
	for _, e := range extensions {
		for _, pg := range e.Enabled {
			n.Insert(pg.Name)
		}
	}
	return sets.List(n)
}

type Extender struct {
}

// TODO: 加密认证

// 源自PodInfo
type GroupInfo struct {
	Group *apis.Group
	//RequiredAffinityTerms      []AffinityTerm
	//RequiredAntiAffinityTerms  []AffinityTerm
	//PreferredAffinityTerms     []WeightedAffinityTerm
	//PreferredAntiAffinityTerms []WeightedAffinityTerm
}

func (gi *GroupInfo) DeepCopy() *GroupInfo {
	return &GroupInfo{
		Group: gi.Group,
	}
}

type QueuedGroupInfo struct {
	*GroupInfo
	// The time pod added to the scheduling queue.
	Timestamp time.Time
	// Number of schedule attempts before successfully scheduled.
	// It's used to record the # attempts metric and calculate the backoff time this Pod is obliged to get before retrying.
	Attempts int
	// The time when the pod is added to the queue for the first time. The pod may be added
	// back to the queue multiple times before it's successfully scheduled.
	// It shouldn't be updated once initialized. It's used to record the e2e scheduling
	// latency for a pod.
	InitialAttemptTimestamp *time.Time
	// UnschedulablePlugins records the plugin names that the Pod failed with Unschedulable or UnschedulableAndUnresolvable status
	// at specific extension points: PreFilter, Filter, Reserve, Permit (WaitOnPermit), or PreBind.
	// If Pods are rejected at other extension points,
	// they're assumed to be unexpected errors (e.g., temporal network issue, plugin implementation issue, etc)
	// and retried soon after a backoff period.
	// That is because such failures could be solved regardless of incoming cluster events (registered in EventsToRegister).
	UnschedulablePlugins sets.Set[string]
	// PendingPlugins records the plugin names that the Pod failed with Pending status.
	PendingPlugins sets.Set[string]
	// Whether the Pod is scheduling gated (by PreEnqueuePlugins) or not.
	Gated bool
}

// NewPodInfo returns a new PodInfo.
func NewGroupInfo(group *apis.Group) (*GroupInfo, error) {
	gInfo := &GroupInfo{
		Group: group,
	}
	//TODO 看看这里 @linbohai
	//err := pInfo.Update(pod)
	return gInfo, nil
}

// NodeInfo is node level aggregated information.
type NodeInfo struct {
	// Overall node information.
	node *apis.Node

	// Pods running on the node.
	Groups []*GroupInfo

	// The subset of pods with affinity.
	//GroupsWithAffinity []*PodInfo

	// The subset of pods with required anti-affinity.
	//PodsWithRequiredAntiAffinity []*PodInfo

	// Ports allocated on the node.
	//UsedPorts HostPortInfo

	// Total requested resources of all pods on this node. This includes assumed
	// pods, which scheduler has sent for binding, but may not be scheduled yet.
	//Requested *Resource
	// Total requested resources of all pods on this node with a minimum value
	// applied to each container's CPU and memory requests. This does not reflect
	// the actual resource requests for this node, but is used to avoid scheduling
	// many zero-request pods onto one node.
	//NonZeroRequested *Resource
	// We store allocatedResources (which is Node.Status.Allocatable.*) explicitly
	// as int64, to avoid conversions and accessing map.
	//Allocatable *Resource

	// ImageStates holds the entry of an image if and only if this image is on the node. The entry can be used for
	// checking an image's existence and advanced usage (e.g., image locality scheduling policy) based on the image
	// state information.
	//ImageStates map[string]*ImageStateSummary

	// PVCRefCounts contains a mapping of PVC names to the number of pods on the node using it.
	// Keys are in the format "namespace/name".
	PVCRefCounts map[string]int

	// Whenever NodeInfo changes, generation is bumped.
	// This is used to avoid cloning it if the object didn't change.
	Generation int64
}

func NewNodeInfo(n *apis.Node) *NodeInfo {
	ret := &NodeInfo{
		node: n,
	}
	return ret
}

// Node returns overall information about this node.
func (n *NodeInfo) Node() *apis.Node {
	if n == nil {
		return nil
	}
	return n.node
}

func (n *NodeInfo) SetNode(node *apis.Node) {
	n.node = node
	//n.Allocatable = NewResource(node.Status.Allocatable)
	//n.Generation = nextGeneration()
}

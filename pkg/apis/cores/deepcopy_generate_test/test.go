// +k8s:deepcopy-gen=package
package deepcopy

import meta "k8s.io/apimachinery/pkg/apis/meta/v1"

type Node1 struct {
	meta.TypeMeta
	meta.ObjectMeta
	Spec   NodeSpec1   `json:"spec,omitempty" yaml:"spec"`
	Status NodeStatus1 `json:"status,omitempty" yaml:"status"`
}

type NodeSpec1 struct {
	NodeName      string `json:"node_name,omitempty" yaml:"node_name"`
	HostName      string `json:"host_name,omitempty" yaml:"host_name"`
	Unschedulable bool   `json:"unschedulable,omitempty" yaml:"unschedulable"`
}

type NodeStatus1 struct {
	NodeName      string `json:"node_name,omitempty" yaml:"node_name"`
	HostName      string `json:"host_name,omitempty" yaml:"host_name"`
	Unschedulable bool   `json:"unschedulable,omitempty" yaml:"unschedulable"`
}

type NodeList1 struct {
	meta.TypeMeta
	meta.ListMeta
	Items []Node1
}

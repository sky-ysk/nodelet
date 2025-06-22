package manager

import (
	"fmt"
	"github.com/google/uuid"
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/component-base/analyzer"
	"hit.edu/framework/pkg/component-base/logs"
	"testing"
)

func TestCreateNode(t *testing.T) {
	clientset, err := CreateClientSet()
	if err != nil {
		panic(err)
	}
	// 构造Manager
	m := NewManager(clientset)

	logs.Init("main")

	// 生成UUID
	u := uuid.Must(uuid.NewV7())

	str := "id"

	nodeSpec := apis.NodeSpec{
		NodeName:        "CloudNode1",
		HostName:        "publicServer0",
		ClusterCategory: "Cloud",
		ClusterID:       &str,
		Resource: map[string][]apis.Item{
			"cpu": {
				{
					Name:   "node.CPU.Info",
					Desc:   "CPU-0-Info",
					Labels: []string{"ModelName", "Core", "BaseFreq"},
					Values: map[string]string{
						"ModelName": "Common KVM processor",
						"Core":      "20",
						"BaseFreq":  "2.45 GHz",
					},
				},
			},
			"memory": {
				{
					Name:   "node.Memory.Info",
					Desc:   "Free and total memory space",
					Labels: []string{"Total", "Available"},
					Values: map[string]string{
						"Total":     "25194901504",
						"Available": "5332951040",
					},
				},
			},
			"storage": {
				{
					Name:   "node.Storage.Info",
					Desc:   "/dev/dm-0-Info",
					Labels: []string{"Device", "MountPoint", "Total", "Used", "Free"},
					Values: map[string]string{
						"Device":     "/dev/dm-0",
						"MountPoint": "/",
						"Total":      "132833185792",
						"Used":       "103099092992",
						"Free":       "23629647872",
					},
				},
			},
		},
	}
	// (ns apis.NodeSpec, namespace string, uuid string)
	g, err := m.CreateNode(nodeSpec, "test", u.String())
	if err != nil {
		panic(err)
	}

	fmt.Println(g)
}

func TestGetNodes(t *testing.T) {
	clientset, err := CreateClientSet()
	if err != nil {
		panic(err)
	}
	// 构造Manager
	m := NewManager(clientset)

	logs.Init("main")

	alist, err := m.GetNodes("test")
	if err != nil {
		panic(err)
	} else {

		for _, item := range alist.Items {
			str, err := analyzer.SerializeToJson(item)
			if err != nil {
				return
			}
			fmt.Println(str)
		}
	}
}

func TestGetNode(t *testing.T) {
	clientset, err := CreateClientSet()
	if err != nil {
		panic(err)
	}
	// 构造Manager
	m := NewManager(clientset)

	logs.Init("main")

	// 生成UUID
	u := uuid.Must(uuid.NewV7())

	str := "id"

	nodeSpec := apis.NodeSpec{
		NodeName:        "Node1",
		HostName:        "publicServer0",
		ClusterCategory: "Cloud",
		ClusterID:       &str,
		Resource: map[string][]apis.Item{
			"cpu": {
				{
					Name:   "node.CPU.Info",
					Desc:   "CPU-0-Info",
					Labels: []string{"ModelName", "Core", "BaseFreq"},
					Values: map[string]string{
						"ModelName": "Common KVM processor",
						"Core":      "20",
						"BaseFreq":  "2.45 GHz",
					},
				},
			},
			"memory": {
				{
					Name:   "node.Memory.Info",
					Desc:   "Free and total memory space",
					Labels: []string{"Total", "Available"},
					Values: map[string]string{
						"Total":     "25194901504",
						"Available": "5332951040",
					},
				},
			},
			"storage": {
				{
					Name:   "node.Storage.Info",
					Desc:   "/dev/dm-0-Info",
					Labels: []string{"Device", "MountPoint", "Total", "Used", "Free"},
					Values: map[string]string{
						"Device":     "/dev/dm-0",
						"MountPoint": "/",
						"Total":      "132833185792",
						"Used":       "103099092992",
						"Free":       "23629647872",
					},
				},
			},
		},
	}
	// (ns apis.NodeSpec, namespace string, uuid string)
	g, err := m.CreateNode(nodeSpec, "test", u.String())
	if err != nil {
		panic(err)
	}

	node, err := m.GetNode(g.Name, g.Namespace)
	if err != nil {
		panic(err)
	}
	fmt.Println(node)

}

func TestUpdateNode(t *testing.T) {

	clientset, err := CreateClientSet()
	if err != nil {
		panic(err)
	}
	// 构造Manager
	m := NewManager(clientset)

	logs.Init("main")

	// 生成UUID
	u := uuid.Must(uuid.NewV7())

	desc := &apis.Description{
		Docs: "node before update",
	}

	str := "id"

	nodeSpec := apis.NodeSpec{
		NodeName:        "Node2",
		HostName:        "publicServer0",
		ClusterCategory: "Cloud",
		ClusterID:       &str,
		Desc:            desc,
		Resource: map[string][]apis.Item{
			"cpu": {
				{
					Name:   "node.CPU.Info",
					Desc:   "CPU-0-Info",
					Labels: []string{"ModelName", "Core", "BaseFreq"},
					Values: map[string]string{
						"ModelName": "Common KVM processor",
						"Core":      "20",
						"BaseFreq":  "2.45 GHz",
					},
				},
			},
			"memory": {
				{
					Name:   "node.Memory.Info",
					Desc:   "Free and total memory space",
					Labels: []string{"Total", "Available"},
					Values: map[string]string{
						"Total":     "25194901504",
						"Available": "5332951040",
					},
				},
			},
			"storage": {
				{
					Name:   "node.Storage.Info",
					Desc:   "/dev/dm-0-Info",
					Labels: []string{"Device", "MountPoint", "Total", "Used", "Free"},
					Values: map[string]string{
						"Device":     "/dev/dm-0",
						"MountPoint": "/",
						"Total":      "132833185792",
						"Used":       "103099092992",
						"Free":       "23629647872",
					},
				},
			},
		},
	}
	// (ns apis.NodeSpec, namespace string, uuid string)
	g, err := m.CreateNode(nodeSpec, "test", u.String())
	if err != nil {
		panic(err)
	}

	str, err = analyzer.SerializeToJson(&g)
	if err != nil {
		return
	}
	fmt.Println(str)

	desc1 := &apis.Description{
		Docs: "node after update",
	}

	g.Spec.Desc = desc1

	patched, err := m.UpdateNode(g.Name, g.Namespace, g)
	if err != nil {
		panic(err)
	} else {
		str, err := analyzer.SerializeToJson(&patched)
		if err != nil {
			return
		}
		fmt.Println(str)
	}
}

func TestPatchNode(t *testing.T) {
	clientset, err := CreateClientSet()
	if err != nil {
		panic(err)
	}
	// 构造Manager
	m := NewManager(clientset)

	logs.Init("main")

	// 生成UUID
	u := uuid.Must(uuid.NewV7())

	desc := &apis.Description{
		Docs: "node before update",
	}

	str := "id"

	nodeSpec := apis.NodeSpec{
		NodeName:        "Node3",
		HostName:        "publicServer0",
		ClusterCategory: "Cloud",
		ClusterID:       &str,
		Desc:            desc,
		Resource: map[string][]apis.Item{
			"cpu": {
				{
					Name:   "node.CPU.Info",
					Desc:   "CPU-0-Info",
					Labels: []string{"ModelName", "Core", "BaseFreq"},
					Values: map[string]string{
						"ModelName": "Common KVM processor",
						"Core":      "20",
						"BaseFreq":  "2.45 GHz",
					},
				},
			},
			"memory": {
				{
					Name:   "node.Memory.Info",
					Desc:   "Free and total memory space",
					Labels: []string{"Total", "Available"},
					Values: map[string]string{
						"Total":     "25194901504",
						"Available": "5332951040",
					},
				},
			},
			"storage": {
				{
					Name:   "node.Storage.Info",
					Desc:   "/dev/dm-0-Info",
					Labels: []string{"Device", "MountPoint", "Total", "Used", "Free"},
					Values: map[string]string{
						"Device":     "/dev/dm-0",
						"MountPoint": "/",
						"Total":      "132833185792",
						"Used":       "103099092992",
						"Free":       "23629647872",
					},
				},
			},
		},
	}
	// (ns apis.NodeSpec, namespace string, uuid string)
	g, err := m.CreateNode(nodeSpec, "test", u.String())
	if err != nil {
		panic(err)
	}

	str, err = analyzer.SerializeToJson(&g)
	if err != nil {
		return
	}
	fmt.Println(str)

	patchData := "{\n  \"spec\": {\n      \"desc\": {\n          \"docs\": \"node after patch\"\n      }\n  }\n\n}"

	patched, err := m.PatchNode(g.Name, g.Namespace, []byte(patchData))
	if err != nil {
		panic(err)
	} else {
		str, err := analyzer.SerializeToJson(&patched)
		if err != nil {
			return
		}
		fmt.Println(str)
	}
}

func TestDeleteNode(t *testing.T) {
	clientset, err := CreateClientSet()
	if err != nil {
		panic(err)
	}
	// 构造Manager
	m := NewManager(clientset)

	logs.Init("main")

	// 生成UUID
	u := uuid.Must(uuid.NewV7())

	desc := &apis.Description{
		Docs: "node before update",
	}

	str := "id"

	nodeSpec := apis.NodeSpec{
		NodeName:        "Node5",
		HostName:        "publicServer0",
		ClusterCategory: "Cloud",
		ClusterID:       &str,
		Desc:            desc,
		Resource: map[string][]apis.Item{
			"cpu": {
				{
					Name:   "node.CPU.Info",
					Desc:   "CPU-0-Info",
					Labels: []string{"ModelName", "Core", "BaseFreq"},
					Values: map[string]string{
						"ModelName": "Common KVM processor",
						"Core":      "20",
						"BaseFreq":  "2.45 GHz",
					},
				},
			},
			"memory": {
				{
					Name:   "node.Memory.Info",
					Desc:   "Free and total memory space",
					Labels: []string{"Total", "Available"},
					Values: map[string]string{
						"Total":     "25194901504",
						"Available": "5332951040",
					},
				},
			},
			"storage": {
				{
					Name:   "node.Storage.Info",
					Desc:   "/dev/dm-0-Info",
					Labels: []string{"Device", "MountPoint", "Total", "Used", "Free"},
					Values: map[string]string{
						"Device":     "/dev/dm-0",
						"MountPoint": "/",
						"Total":      "132833185792",
						"Used":       "103099092992",
						"Free":       "23629647872",
					},
				},
			},
		},
	}
	// (ns apis.NodeSpec, namespace string, uuid string)
	g, err := m.CreateNode(nodeSpec, "test", u.String())
	if err != nil {
		panic(err)
	}

	err = m.DeleteNode(g.Name, g.Namespace)
	if err != nil {
		panic(err)
	} else {
		fmt.Println("Delete group success")
	}
}

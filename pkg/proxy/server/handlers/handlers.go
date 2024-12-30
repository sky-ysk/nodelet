package handlers

import (
	"github.com/emicklei/go-restful/v3"
	"hit.edu/framework/pkg/client-go/clients"
	"hit.edu/framework/pkg/proxy/server/handlers/group"
	"hit.edu/framework/pkg/proxy/server/handlers/node"
	"hit.edu/framework/pkg/proxy/server/handlers/task"
	"hit.edu/framework/pkg/proxy/server/handlers/workflow"
)

type Handlers struct {
	// ClientSets
	ClientSet *clients.ClientSet
}

func NewHandlers(clientSets *clients.ClientSet) *Handlers {
	return &Handlers{
		ClientSet: clientSets,
	}
}

func (h *Handlers) InstallWorkflowHandlers(container *restful.Container) {
	// Workflows相关
	wsh := workflow.NewWorkflowsHandler(h.ClientSet)
	// 查询Workflows
	container.Add(wsh.NewGetWebService())

	// Workflow相关
	// 查询单个Workflow
	wh := workflow.NewWorkflowHandler(h.ClientSet)
	container.Add(wh.NewGetWebService())
}

func (h *Handlers) InstallTaskHandlers(container *restful.Container) {
	// Tasks相关
	tsh := task.NewTasksHandler(h.ClientSet)
	// 查询Tasks
	container.Add(tsh.NewGetWebService())

	// Task相关
	// 查询单个Task
	th := task.NewTaskHandler(h.ClientSet)
	container.Add(th.NewGetWebService())
}

func (h *Handlers) InstallNodeHandlers(container *restful.Container) {
	// Nodes相关
	nsh := node.NewNodesHandler(h.ClientSet)
	// 查询Nodes
	container.Add(nsh.NewGetWebService())

	// Node相关
	// 查询单个Node
	nh := node.NewNodeHandler(h.ClientSet)
	container.Add(nh.NewGetWebService())
}

func (h *Handlers) InstallGroupHandlers(container *restful.Container) {
	// Groups相关
	gsh := group.NewGroupsHandler(h.ClientSet)
	// 查询Groups
	container.Add(gsh.NewGetWebService())

	// Group相关
	// 查询单个Group
	gh := group.NewGroupHandler(h.ClientSet)
	container.Add(gh.NewGetWebService())
}

package workflow

import "github.com/emicklei/go-restful/v3"

const (
	NAMESPACE     = "framework"
	GROUP         = "v1"
	TAG           = "Workflow"
	API_PREFIX    = "/" + NAMESPACE + "/" + GROUP
	WorkflowsPath = API_PREFIX + "/workflows"
	WorkflowPath  = API_PREFIX + "/workflow"
	WorkflowName  = "Name"
)

type Handler interface {
	// 查询
	NewGetWebService() *restful.WebService
}

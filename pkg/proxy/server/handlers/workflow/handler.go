package workflow

import "github.com/emicklei/go-restful/v3"

const (
	NAMESPACE      = "framework"
	GROUP          = "v1"
	TAG            = "Workflow"
	API_PREFIX     = "/" + NAMESPACE + "/" + GROUP
	WORKFLOWS_PATH = API_PREFIX + "/workflows"
	WORKFLOW_PATH  = API_PREFIX + "/workflow"
	WORKFLOW_NAME  = "Name"
	NAME_SPACE     = "Namespace"
)

type Handler interface {
	// 查询
	NewGetWebService() *restful.WebService
}

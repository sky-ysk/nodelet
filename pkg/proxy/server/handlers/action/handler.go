package action

import "github.com/emicklei/go-restful/v3"

const (
	NAMESPACE    = "framework"
	GROUP        = "v1"
	TAG          = "Action"
	API_PREFIX   = "/" + NAMESPACE + "/" + GROUP
	ACTIONS_PATH = API_PREFIX + "/actions"
	ACTION_PATH  = API_PREFIX + "/action"
	ACTION_NAME  = "Name"
	NAME_SPACE   = "Namespace"
)

type Handler interface {
	// NewGetWebService 查询
	NewGetWebService() *restful.WebService
}

package action

import "github.com/emicklei/go-restful/v3"

const (
	//NAMESPACE    = "framework"
	//GROUP        = "v1"
	//TAG          = "Action"
	//API_PREFIX   = "/" + NAMESPACE + "/" + GROUP
	//ACTIONS_PATH = API_PREFIX + "/actions"
	//ACTION_PATH  = API_PREFIX + "/action"
	//ACTION_NAME  = "Name"
	//GROUP = "v1"
	//API_PREFIX = "/" + NAMESPACE + "/" + GROUP
	TAG          = "Action"
	ACTIONS_PATH = "/apis/resources/v1/namespaces/actions"
	ACTION_PATH  = "/apis/resources/v1/namespaces/action"
	ACTION_NAME  = "Name"
	NAMESPACE    = "Namespace"
)

type Handler interface {
	// 查询
	NewGetWebService() *restful.WebService
}

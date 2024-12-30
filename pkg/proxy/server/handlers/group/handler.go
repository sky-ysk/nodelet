package group

import (
	"github.com/emicklei/go-restful/v3"
)

const (
	NAMESPACE   = "framework"
	GROUP       = "v1"
	TAG         = "Group"
	API_PREFIX  = "/" + NAMESPACE + "/" + GROUP
	GROUPS_PATH = API_PREFIX + "/groups"
	GROUP_PATH  = API_PREFIX + "/group"
	GROUP_NAME  = "Name"
)

type Handler interface {
	// 查询
	NewGetWebService() *restful.WebService
}

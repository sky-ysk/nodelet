package node

import (
	"github.com/emicklei/go-restful/v3"
)

const (
	NAMESPACE  = "framework"
	GROUP      = "v1"
	TAG        = "Node"
	API_PREFIX = "/" + NAMESPACE + "/" + GROUP
	NODES_PATH = API_PREFIX + "/nodes"
	NODE_PATH  = API_PREFIX + "/node"
	NODE_NAME  = "Name"
	NAME_SPACE = "Namespace"
)

type Handler interface {
	// 查询
	NewGetWebService() *restful.WebService
}

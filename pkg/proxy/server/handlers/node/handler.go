package node

import (
	"github.com/emicklei/go-restful/v3"
)

const (
	NAMESPACE  = "resource"
	GROUP      = "v1"
	TAG        = "Node"
	API_PREFIX = "/" + NAMESPACE + "/" + GROUP
	NODES_PATH = API_PREFIX + "/nodes"
	NODE_PATH  = API_PREFIX + "/node"
	NODE_NAME  = "Name"
)

type Handler interface {
	// 查询
	NewGetWebService() *restful.WebService
}

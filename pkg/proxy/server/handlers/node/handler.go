package node

import (
	"github.com/emicklei/go-restful/v3"
)

const (
	//NAMESPACE  = "framework"
	GROUP = "v1"
	TAG   = "Node"
	//API_PREFIX = "/" + NAMESPACE + "/" + GROUP
	NODES_PATH = "/apis/resources/v1/namespaces/nodes"
	NODE_PATH  = "/apis/resources/v1/namespaces/node"
	NODE_NAME  = "Name"
	NAMESPACE  = "Namespace"
)

type Handler interface {
	// 查询
	NewGetWebService() *restful.WebService
}

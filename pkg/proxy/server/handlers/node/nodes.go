package node

import (
	restfulspec "github.com/emicklei/go-restful-openapi/v2"
	"github.com/emicklei/go-restful/v3"
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/component-base/logs"
	"net/http"
)

type NodesHandler struct{}

var _ Handler = &NodesHandler{}

func NewNodesHandler() *NodesHandler {
	return &NodesHandler{}
}

func GetNodes(request *restful.Request, response *restful.Response) {
	// TODO: 使用client-go实现查询
	node1 := apis.Node{
		Spec: apis.NodeSpec{
			NodeName: "TestNode",
		},
	}

	node2 := apis.Node{
		Spec: apis.NodeSpec{
			NodeName: "TestNode",
		},
	}

	nodes := []apis.Node{node1, node2}

	err := response.WriteEntity(nodes)
	if err != nil {
		response.WriteError(http.StatusInternalServerError, err)
	}
	logs.Debugf("get nodes")
}

func (h *NodesHandler) NewGetWebService() *restful.WebService {
	ws := new(restful.WebService)
	ws.Path(NODES_PATH).
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("").
		//Docs
		Doc("Get all nodes").
		Metadata(restfulspec.KeyOpenAPITags, []string{TAG}).
		To(GetNodes).
		Operation("getNodes").
		Returns(200, "OK", []apis.Node{}).
		Returns(400, "Not Found", nil),
	)
	return ws
}

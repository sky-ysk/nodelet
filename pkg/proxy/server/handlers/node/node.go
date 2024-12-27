package node

import (
	"fmt"
	restfulspec "github.com/emicklei/go-restful-openapi/v2"
	"github.com/emicklei/go-restful/v3"
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/component-base/logs"
	"net/http"
)

type NodeHandler struct {
}

var _ Handler = &NodeHandler{}

func NewNodeHandler() *NodeHandler {
	return &NodeHandler{}
}

func GetNode(request *restful.Request, response *restful.Response) {
	// TODO: 使用client-go实现查询
	name := request.PathParameter(NODE_NAME)
	node := apis.Node{
		Spec: apis.NodeSpec{
			NodeName: name,
		},
	}
	err := response.WriteEntity(node)
	if err != nil {
		response.WriteError(http.StatusInternalServerError, err)
	}
	logs.Debugf("get nodes")
}

func (h *NodeHandler) NewGetWebService() *restful.WebService {
	ws := new(restful.WebService)
	ws.Path(NODE_PATH).
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)
	ws.Route(ws.GET(fmt.Sprintf("/{%s}", NODE_NAME)).
		To(GetNode).
		Doc("Get a node with name").
		Metadata(restfulspec.KeyOpenAPITags, []string{TAG}).
		Operation("getNode").
		Returns(200, "OK", apis.Node{}).
		Returns(400, "Not Found", nil),
	)
	return ws
}

package node

import (
	"context"
	restfulspec "github.com/emicklei/go-restful-openapi/v2"
	"github.com/emicklei/go-restful/v3"
	apis "hit.edu/framework/pkg/apis/cores"
	metav1 "hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/client-go/clients"
	"hit.edu/framework/pkg/client-go/clients/typed/core"
	"hit.edu/framework/pkg/component-base/logs"
	"net/http"
)

type NodesHandler struct {
	client core.NodeInterface
}

var _ Handler = &NodesHandler{}

func NewNodesHandler(clientSet *clients.ClientSet) *NodesHandler {
	c := clientSet.Core().Nodes(apis.NamespaceAll)
	return &NodesHandler{
		client: c,
	}
}

func (h *NodesHandler) GetNodes(request *restful.Request, response *restful.Response) {
	// 使用client-go实现查询
	results, err := h.client.List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		logs.Errorf("Get nodes failed: %v", err)
		response.WriteError(http.StatusInternalServerError, err)
	}

	err = response.WriteEntity(results)
	if err != nil {
		response.WriteError(http.StatusInternalServerError, err)
	}
	logs.Debugf("Get nodes")
}

// TODO: DeleteAll

func (h *NodesHandler) NewGetWebService() *restful.WebService {
	ws := new(restful.WebService)
	ws.Path(NODES_PATH).
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("").
		//Docs
		Doc("Get all nodes").
		Metadata(restfulspec.KeyOpenAPITags, []string{TAG}).
		To(h.GetNodes).
		Operation("Get nodes").
		Returns(200, "OK", []apis.Node{}).
		Returns(400, "Not Found", nil),
	)
	return ws
}

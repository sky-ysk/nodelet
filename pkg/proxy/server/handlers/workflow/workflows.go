package workflow

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

type WorkflowsHandler struct {
	client core.WorkflowInterface
}

var _ Handler = &WorkflowsHandler{}

func NewWorkflowsHandler(clientSet *clients.ClientSet) *WorkflowsHandler {
	c := clientSet.Core().Workflows(apis.NamespaceAll)
	return &WorkflowsHandler{
		client: c,
	}
}

func (h *WorkflowsHandler) GetWorkflows(request *restful.Request, response *restful.Response) {
	// 使用client-go实现查询
	results, err := h.client.List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		logs.Errorf("Get workflows failed: %v", err)
		response.WriteError(http.StatusInternalServerError, err)
	}

	err = response.WriteEntity(results)
	if err != nil {
		response.WriteError(http.StatusInternalServerError, err)
	}
	logs.Debugf("Get workflows")
}

// TODO: DeleteAll

func (h *WorkflowsHandler) NewGetWebService() *restful.WebService {
	ws := new(restful.WebService)
	ws.Path(WorkflowsPath).
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)
	ws.Route(ws.GET("").
		Doc("Get all workflows").
		Metadata(restfulspec.KeyOpenAPITags, []string{TAG}).
		To(h.GetWorkflows).
		Operation("Get workflows").
		Returns(200, "OK", []apis.Workflow{}).
		Returns(400, "Not Found", nil),
	)

	return ws
}

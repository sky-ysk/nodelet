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
	c := clientSet.Core().Workflows("test") //apis.NamespaceAll
	return &WorkflowsHandler{
		client: c,
	}
}

func (h *WorkflowsHandler) GetWorkflows(request *restful.Request, response *restful.Response) {
	results, err := h.client.List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		logs.Errorf("Get workflows failed: %v", err)
		err := response.WriteError(http.StatusInternalServerError, err)
		if err != nil {
			logs.Errorf("failed to return a status code")
			return
		}
	}

	err = response.WriteEntity(results)
	if err != nil {
		err := response.WriteError(http.StatusInternalServerError, err)
		if err != nil {
			logs.Errorf("failed to return a status code")
			return
		}
	}
	logs.Debugf("Get workflows")
}

// DeleteAll
func (h *WorkflowsHandler) DeleteAllWorkflow(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	str := "Spec.Name=" + namespace

	lstOpts := metav1.ListOptions{
		FieldSelector: str,
	}
	err := h.client.DeleteCollection(context.TODO(), metav1.DeleteOptions{}, lstOpts)
	if err != nil {
		logs.Errorf("Delete workflows failed: %v", err)
		err := response.WriteError(http.StatusInternalServerError, err)
		if err != nil {
			logs.Errorf("failed to return a status code")
			return
		}
		return
	}
	// 返回停止成功的状态
	response.WriteHeader(http.StatusOK)
	// 记录日志
	logs.Debugf("delete all workflows ")
}

func (h *WorkflowsHandler) NewGetWebService() *restful.WebService {
	ws := new(restful.WebService)
	ws.Path(WORKFLOWS_PATH).
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/{Namespace}/workflows").
		Doc("Get all workflows").
		Metadata(restfulspec.KeyOpenAPITags, []string{TAG}).
		Param(ws.QueryParameter("Namespace", "The namespace of the workflows").DataType("string")).
		To(h.GetWorkflows).
		Operation("Get workflows").
		Returns(200, "OK", []apis.Workflow{}).
		Returns(400, "Not Found", nil),
	)

	ws.Route(ws.DELETE("/{Namespace}/workflows").
		Doc("Delete all workflows").
		Metadata(restfulspec.KeyOpenAPITags, []string{TAG}).
		Param(ws.QueryParameter("Namespace", "The namespace of the workflows").DataType("string")).
		To(h.DeleteAllWorkflow).
		Operation("Delete workflows").
		Returns(200, "OK", []apis.Workflow{}).
		Returns(400, "Not Found", nil),
	)

	return ws
}

package workflow

import (
	"fmt"
	restfulspec "github.com/emicklei/go-restful-openapi/v2"
	"github.com/emicklei/go-restful/v3"
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/client-go/clients"
	"hit.edu/framework/pkg/client-go/util/manager"
	"hit.edu/framework/pkg/component-base/logs"
	"net/http"
	"sync"
)

type WorkflowsHandler struct {
	clientSet *clients.ClientSet
	manager   *manager.Manager
	mu        sync.Mutex
}

var _ Handler = &WorkflowsHandler{}

// NewWorkflowHandler 创建一个 ActionHandler
func NewWorkflowsHandler(clientSet *clients.ClientSet) *WorkflowsHandler {
	return &WorkflowsHandler{
		manager:   manager.NewManager(clientSet),
		clientSet: clientSet,
	}
}

func (h *WorkflowsHandler) GetWorkflows(request *restful.Request, response *restful.Response) {
	// 从url中获取namespace
	namespace := request.QueryParameter(NAME_SPACE)
	if namespace == "" {
		err := response.WriteError(http.StatusBadRequest, fmt.Errorf("namespace is required"))
		if err != nil {
			logs.Errorf("failed to return a status code ")
			return
		}
		return
	}

	results, err := h.manager.GetWorkflows(namespace)
	if err != nil {
		logs.Errorf("Get workflows failed: %v", err)
		err := response.WriteError(http.StatusInternalServerError, err)
		if err != nil {
			logs.Errorf("failed to return a status code")
			return
		}
		return
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
	// 获取namespace
	namespace := request.QueryParameter(NAME_SPACE)
	if namespace == "" {
		err := response.WriteError(http.StatusBadRequest, fmt.Errorf("namespace is required"))
		if err != nil {
			logs.Errorf("failed to return a status code ")
			return
		}
		return
	}

	err := h.manager.DeleteWorkflows(namespace)
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

	ws.Route(ws.GET(fmt.Sprintf("/")).
		Doc("Get all workflows").
		Metadata(restfulspec.KeyOpenAPITags, []string{TAG}).
		Param(ws.QueryParameter("Namespace", "The namespace of the workflows").DataType("string")).
		To(h.GetWorkflows).
		Operation("Get workflows").
		Returns(200, "OK", []apis.Workflow{}).
		Returns(400, "Not Found", nil),
	)

	ws.Route(ws.DELETE(fmt.Sprintf("/")).
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

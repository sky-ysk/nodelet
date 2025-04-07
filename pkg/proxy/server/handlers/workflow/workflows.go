package workflow

import (
	"context"
	"fmt"
	restfulspec "github.com/emicklei/go-restful-openapi/v2"
	"github.com/emicklei/go-restful/v3"
	apis "hit.edu/framework/pkg/apis/cores"
	metav1 "hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/client-go/clients"
	"hit.edu/framework/pkg/client-go/clients/typed/core"
	"hit.edu/framework/pkg/component-base/logs"
	"net/http"
	"sync"
)

type WorkflowsHandler struct {
	clients   map[string]core.WorkflowInterface
	clientSet *clients.ClientSet
	mu        sync.Mutex
}

type CurrentWorkflowsHandler struct {
	client core.WorkflowInterface
}

var _ Handler = &WorkflowsHandler{}

//func NewWorkflowsHandler(clientSet *clients.ClientSet) *WorkflowsHandler {
//	c := clientSet.Core().Workflows("test") //apis.NamespaceAll
//	return &WorkflowsHandler{
//		client: c,
//	}
//}

// NewWorkflowHandler 创建一个 ActionHandler
func NewWorkflowsHandler(clientSet *clients.ClientSet) *WorkflowsHandler {
	return &WorkflowsHandler{
		clients:   make(map[string]core.WorkflowInterface),
		clientSet: clientSet,
	}
}

// GetClient 根据 namespace 获取 client，如果不存在则创建
func (h *WorkflowsHandler) GetClient(namespace string) *CurrentWorkflowsHandler {
	h.mu.Lock()
	defer h.mu.Unlock()

	// 如果已经存在，直接返回
	if c, exists := h.clients[namespace]; exists {
		return &CurrentWorkflowsHandler{
			client: c,
		}
	}

	// 否则创建新的 client
	newClient := h.clientSet.Core().Workflows(namespace)
	h.clients[namespace] = newClient
	return &CurrentWorkflowsHandler{
		client: newClient,
	}
}

func (h *WorkflowsHandler) GetWorkflows(request *restful.Request, response *restful.Response) {
	c := &CurrentWorkflowsHandler{}
	// 从url中获取namespace
	namespace := request.QueryParameter(NAME_SPACE)
	if namespace == "" {
		err := response.WriteError(http.StatusBadRequest, fmt.Errorf("namespace is required"))
		if err != nil {
			logs.Errorf("failed to return a status code ")
			return
		}
		return
	} else {
		c = h.GetClient(namespace)
	}
	results, err := c.client.List(context.TODO(), metav1.ListOptions{})
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
	c := &CurrentWorkflowsHandler{}
	// 从url中获取namespace
	namespace := request.QueryParameter(NAME_SPACE)
	if namespace == "" {
		err := response.WriteError(http.StatusBadRequest, fmt.Errorf("namespace is required"))
		if err != nil {
			logs.Errorf("failed to return a status code ")
			return
		}
		return
	} else {
		c = h.GetClient(namespace)
	}

	str := "Spec.Name=" + namespace

	lstOpts := metav1.ListOptions{
		FieldSelector: str,
	}
	err := c.client.DeleteCollection(context.TODO(), metav1.DeleteOptions{}, lstOpts)
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

	ws.Route(ws.GET(fmt.Sprintf("/{Namespace}")).
		Doc("Get all workflows").
		Metadata(restfulspec.KeyOpenAPITags, []string{TAG}).
		Param(ws.QueryParameter("Namespace", "The namespace of the workflows").DataType("string")).
		To(h.GetWorkflows).
		Operation("Get workflows").
		Returns(200, "OK", []apis.Workflow{}).
		Returns(400, "Not Found", nil),
	)

	ws.Route(ws.DELETE(fmt.Sprintf("/{Namespace}")).
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

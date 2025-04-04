package action

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

type ActionsHandler struct {
	clients   map[string]core.ActionInterface
	clientSet *clients.ClientSet
	mu        sync.Mutex
}

type CurrentActionsHandler struct {
	client core.ActionInterface
}

var _ Handler = &ActionsHandler{}

//func NewActionsHandler(clientSet *clients.ClientSet) *ActionsHandler {
//	c := clientSet.Core().Actions("test") // apis.NamespaceAll
//	return &ActionsHandler{
//		client: c,
//	}
//}

// NewActionHandler 创建一个 ActionHandler
func NewActionsHandler(clientSet *clients.ClientSet) *ActionsHandler {
	return &ActionsHandler{
		clients:   make(map[string]core.ActionInterface),
		clientSet: clientSet,
	}
}

// GetClient 根据 namespace 获取 client，如果不存在则创建
func (h *ActionsHandler) GetClient(namespace string) *CurrentActionsHandler {
	h.mu.Lock()
	defer h.mu.Unlock()

	// 如果已经存在，直接返回
	if c, exists := h.clients[namespace]; exists {
		return &CurrentActionsHandler{
			client: c,
		}
	}

	// 否则创建新的 client
	newClient := h.clientSet.Core().Actions(namespace)
	h.clients[namespace] = newClient
	return &CurrentActionsHandler{
		client: newClient,
	}
}

func (h *ActionsHandler) GetActions(request *restful.Request, response *restful.Response) {
	c := &CurrentActionsHandler{}
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
		logs.Errorf("Get actions failed: %v", err)
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
	logs.Debugf("Get actions")
}

// TODO: DeleteAll

func (h *ActionsHandler) NewGetWebService() *restful.WebService {
	ws := new(restful.WebService)
	ws.Path(ACTIONS_PATH).
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/").
		Doc("Get all actions").
		Metadata(restfulspec.KeyOpenAPITags, []string{TAG}).
		Param(ws.QueryParameter("Namespace", "The namespace of the action").DataType("string")).
		To(h.GetActions).
		Operation("Get actions").
		Returns(200, "OK", []apis.Action{}).
		Returns(400, "Not Found", nil),
	)

	return ws
}

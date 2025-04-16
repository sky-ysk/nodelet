package runtime

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

type RuntimesHandler struct {
	clients   map[string]core.RuntimeInterface
	clientSet *clients.ClientSet
	mu        sync.Mutex
}

type CurrentRuntimesHandler struct {
	client core.RuntimeInterface
}

var _ Handler = &RuntimesHandler{}

//func NewRuntimesHandler(clientSet *clients.ClientSet) *RuntimesHandler {
//	c := clientSet.Core().Runtimes("test") // apis.NamespaceAll
//	return &RuntimesHandler{
//		client: c,
//	}
//}

// NewRuntimesHandler 创建一个 RuntimesHandler
func NewRuntimesHandler(clientSet *clients.ClientSet) *RuntimesHandler {
	return &RuntimesHandler{
		clients:   make(map[string]core.RuntimeInterface),
		clientSet: clientSet,
	}
}

// GetClient 根据 namespace 获取 client，如果不存在则创建
func (h *RuntimesHandler) GetClient(namespace string) *CurrentRuntimesHandler {
	h.mu.Lock()
	defer h.mu.Unlock()

	// 如果已经存在，直接返回
	if c, exists := h.clients[namespace]; exists {
		return &CurrentRuntimesHandler{
			client: c,
		}
	}

	// 否则创建新的 client
	newClient := h.clientSet.Core().Runtimes(namespace)
	h.clients[namespace] = newClient
	return &CurrentRuntimesHandler{
		client: newClient,
	}
}

func (h *RuntimesHandler) GetRuntimes(request *restful.Request, response *restful.Response) {
	c := &CurrentRuntimesHandler{}
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
		logs.Errorf("Get Runtimes failed: %v", err)
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
	logs.Debugf("Get runtimes")
}

func (h *RuntimesHandler) NewGetWebService() *restful.WebService {
	ws := new(restful.WebService)
	ws.Path(RUNTIMES_PATH).
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/").
		Doc("Get all runtimes").
		Metadata(restfulspec.KeyOpenAPITags, []string{TAG}).
		Param(ws.QueryParameter("Namespace", "The namespace of the runtimes").DataType("string")).
		To(h.GetRuntimes).
		Operation("Get runtimes").
		Returns(200, "OK", []apis.Runtime{}).
		Returns(400, "Not Found", nil),
	)
	return ws
}

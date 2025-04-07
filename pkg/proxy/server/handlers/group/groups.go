package group

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

type GroupsHandler struct {
	clients   map[string]core.GroupInterface
	clientSet *clients.ClientSet
	mu        sync.Mutex
}

type CurrentGroupsHandler struct {
	client core.GroupInterface
}

var _ Handler = &GroupsHandler{}

//func NewGroupsHandler(clientSet *clients.ClientSet) *GroupsHandler {
//	c := clientSet.Core().Groups("test") // apis.NamespaceAll
//	return &GroupsHandler{
//		client: c,
//	}
//}

// NewGroupsHandler 创建一个 GroupsHandler
func NewGroupsHandler(clientSet *clients.ClientSet) *GroupsHandler {
	return &GroupsHandler{
		clients:   make(map[string]core.GroupInterface),
		clientSet: clientSet,
	}
}

// GetClient 根据 namespace 获取 client，如果不存在则创建
func (h *GroupsHandler) GetClient(namespace string) *CurrentGroupsHandler {
	h.mu.Lock()
	defer h.mu.Unlock()

	// 如果已经存在，直接返回
	if c, exists := h.clients[namespace]; exists {
		return &CurrentGroupsHandler{
			client: c,
		}
	}

	// 否则创建新的 client
	newClient := h.clientSet.Core().Groups(namespace)
	h.clients[namespace] = newClient
	return &CurrentGroupsHandler{
		client: newClient,
	}
}

func (h *GroupsHandler) GetGroups(request *restful.Request, response *restful.Response) {
	c := &CurrentGroupsHandler{}
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
		logs.Errorf("Get groups failed: %v", err)
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
	logs.Debugf("Get groups")
}

func (h *GroupsHandler) NewGetWebService() *restful.WebService {
	ws := new(restful.WebService)
	ws.Path(GROUPS_PATH).
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/").
		Doc("Get all groups").
		Metadata(restfulspec.KeyOpenAPITags, []string{TAG}).
		Param(ws.QueryParameter("Namespace", "The namespace of the groups").DataType("string")).
		To(h.GetGroups).
		Operation("Get groups").
		Returns(200, "OK", []apis.Group{}).
		Returns(400, "Not Found", nil),
	)
	return ws
}

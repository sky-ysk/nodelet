package task

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

type TasksHandler struct {
	clients   map[string]core.TaskInterface
	clientSet *clients.ClientSet
	mu        sync.Mutex
}

type CurrentTasksHandler struct {
	client core.TaskInterface
}

var _ Handler = &TasksHandler{}

//func NewTasksHandler(clientSet *clients.ClientSet) *TasksHandler {
//	c := clientSet.Core().Tasks("test") //apis.NamespaceAll
//	return &TasksHandler{
//		client: c,
//	}
//}

// NewTaskHandler 创建一个 TaskHandler
func NewTasksHandler(clientSet *clients.ClientSet) *TasksHandler {
	return &TasksHandler{
		clients:   make(map[string]core.TaskInterface),
		clientSet: clientSet,
	}
}

// GetClient 根据 namespace 获取 client，如果不存在则创建
func (h *TasksHandler) GetClient(namespace string) *CurrentTasksHandler {
	h.mu.Lock()
	defer h.mu.Unlock()

	// 如果已经存在，直接返回   c 和 gc 会一起创建
	c, exists := h.clients[namespace]
	if exists {
		return &CurrentTasksHandler{
			client: c,
		}
	}

	// 否则创建新的 client
	newClient := h.clientSet.Core().Tasks(namespace)
	h.clients[namespace] = newClient
	return &CurrentTasksHandler{
		client: newClient,
	}
}

func (h *TasksHandler) GetTasks(request *restful.Request, response *restful.Response) {
	c := &CurrentTasksHandler{}
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
		logs.Errorf("Get tasks failed : %v", err)
		err := response.WriteHeaderAndEntity(http.StatusInternalServerError, err)
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
	logs.Debugf("Get tasks")
}

func (h *TasksHandler) NewGetWebService() *restful.WebService {
	ws := new(restful.WebService)
	ws.Path(TASKS_PATH).
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("").
		//Docs
		Doc("Get all tasks").
		Metadata(restfulspec.KeyOpenAPITags, []string{TAG}).
		Param(ws.QueryParameter("Namespace", "The namespace of the tasks").DataType("string")).
		To(h.GetTasks).
		Operation("Get tasks").
		Returns(200, "OK", []apis.Task{}).
		Returns(400, "Not Found", nil),
	)

	return ws
}

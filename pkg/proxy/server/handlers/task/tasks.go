package task

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

type TasksHandler struct {
	client core.TaskInterface
}

var _ Handler = &TasksHandler{}

func NewTasksHandler(clientSet *clients.ClientSet) *TasksHandler {
	c := clientSet.Core().Tasks(apis.NamespaceAll)
	return &TasksHandler{
		client: c,
	}
}

func (h *TasksHandler) GetTasks(request *restful.Request, response *restful.Response) {
	// 使用client-go实现查询
	results, err := h.client.List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		logs.Errorf("Get tasks failed: %v", err)
		response.WriteError(http.StatusInternalServerError, err)
	}

	err = response.WriteEntity(results)
	if err != nil {
		response.WriteError(http.StatusInternalServerError, err)
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
		To(h.GetTasks).
		Operation("Get tasks").
		Returns(200, "OK", []apis.Task{}).
		Returns(400, "Not Found", nil),
	)
	return ws
}

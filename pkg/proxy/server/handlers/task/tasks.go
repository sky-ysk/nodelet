package task

import (
	restfulspec "github.com/emicklei/go-restful-openapi/v2"
	"github.com/emicklei/go-restful/v3"
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/client-go/clients"
	"hit.edu/framework/pkg/client-go/util/manager"
	"hit.edu/framework/pkg/component-base/logs"
	"net/http"
	"sync"
)

type TasksHandler struct {
	clientSet *clients.ClientSet
	manager   *manager.Manager
	mu        sync.Mutex
}

var _ Handler = &TasksHandler{}

// NewTaskHandler 创建一个 TaskHandler
func NewTasksHandler(clientSet *clients.ClientSet) *TasksHandler {
	return &TasksHandler{
		manager:   manager.NewManager(clientSet),
		clientSet: clientSet,
	}
}

func (h *TasksHandler) GetTasks(request *restful.Request, response *restful.Response) {
	// 从url中获取namespace
	namespace := request.QueryParameter(NAME_SPACE)
	var results *apis.TaskList
	var err error
	if namespace == "" {
		//err := response.WriteError(http.StatusBadRequest, fmt.Errorf("namespace is required"))
		//if err != nil {
		//	logs.Errorf("failed to return a status code ")
		//	return
		//}
		//return
		results, err = h.manager.GetTasks(namespace)
		if err != nil {
			logs.Errorf("Get tasks with labels failed: %v", err)
			err := response.WriteError(http.StatusInternalServerError, err)
			if err != nil {
				logs.Errorf("failed to return a status code")
				return
			}
		}
	}

	labels := request.QueryParameter("Label")
	//var results *apis.TaskList
	//var err error
	if labels == "" {
		results, err = h.manager.GetTasks(namespace)
		if err != nil {
			logs.Errorf("Get tasks with labels failed: %v", err)
			err := response.WriteError(http.StatusInternalServerError, err)
			if err != nil {
				logs.Errorf("failed to return a status code")
				return
			}
		}
	} else {
		results, err = h.manager.FilterTasks(namespace, labels)
		if err != nil {
			logs.Errorf("Get tasks with labels failed: %v", err)
			err := response.WriteError(http.StatusInternalServerError, err)
			if err != nil {
				logs.Errorf("failed to return a status code")
				return
			}
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
		Param(ws.QueryParameter("Label", "Labels of the tasks (optional)").DataType("string")).
		Param(ws.QueryParameter("Namespace", "The namespace of the tasks").DataType("string")).
		To(h.GetTasks).
		Operation("Get tasks").
		Returns(200, "OK", []apis.Task{}).
		Returns(400, "Not Found", nil),
	)

	return ws
}

package task

import (
	restfulspec "github.com/emicklei/go-restful-openapi/v2"
	"github.com/emicklei/go-restful/v3"
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/component-base/logs"
	"net/http"
)

type TasksHandler struct{}

var _ Handler = &TasksHandler{}

func NewTasksHandler() *TasksHandler {
	return &TasksHandler{}
}

func GetTasks(request *restful.Request, response *restful.Response) {
	// TODO: 使用client-go实现查询
	task1 := apis.Task{
		Spec: apis.TaskSpec{
			Name: "TestTask",
		},
	}

	task2 := apis.Task{
		Spec: apis.TaskSpec{
			Name: "TestTask",
		},
	}

	tasks := []apis.Task{task1, task2}

	err := response.WriteEntity(tasks)
	if err != nil {
		response.WriteError(http.StatusInternalServerError, err)
	}
	logs.Debugf("get tasks")
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
		To(GetTasks).
		Operation("getTasks").
		Returns(200, "OK", []apis.Task{}).
		Returns(400, "Not Found", nil),
	)
	return ws
}

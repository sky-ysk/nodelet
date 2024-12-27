package task

import (
	"fmt"
	restfulspec "github.com/emicklei/go-restful-openapi/v2"
	"github.com/emicklei/go-restful/v3"
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/component-base/logs"
	"net/http"
)

type TaskHandler struct{}

var _ Handler = &TaskHandler{}

func NewTaskHandler() *TaskHandler {
	return &TaskHandler{}
}

func GetTask(request *restful.Request, response *restful.Response) {
	// TODO: 使用client-go实现查询
	name := request.PathParameter(TASK_NAME)
	task := apis.Task{
		Spec: apis.TaskSpec{
			Name: name,
		},
	}
	err := response.WriteEntity(task)
	if err != nil {
		response.WriteError(http.StatusInternalServerError, err)
	}
	logs.Debugf("get tasks")
}

func (h *TaskHandler) NewGetWebService() *restful.WebService {
	ws := new(restful.WebService)
	ws.Path(TASK_PATH).
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)
	ws.Route(ws.GET(fmt.Sprintf("/{%s}", TASK_NAME)).
		To(GetTask).
		Doc("Get a task with name").
		Metadata(restfulspec.KeyOpenAPITags, []string{TAG}).
		Operation("getTask").
		Returns(200, "OK", apis.Task{}).
		Returns(400, "Not Found", nil),
	)
	return ws
}

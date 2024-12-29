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
)

type TaskHandler struct {
	client core.TaskInterface
}

var _ Handler = &TaskHandler{}

func NewTaskHandler(clientSet *clients.ClientSet) *TaskHandler {
	c := clientSet.Core().Tasks(apis.NamespaceAll)
	return &TaskHandler{
		client: c,
	}
}

func (h *TaskHandler) GetTask(request *restful.Request, response *restful.Response) {
	// TODO: 使用client-go实现查询
	name := request.PathParameter(TASK_NAME)
	result, err := h.client.Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		logs.Errorf("Get task %s error: %v", name, err)
		response.WriteError(http.StatusInternalServerError, err)
	}
	err = response.WriteEntity(result)
	if err != nil {
		response.WriteError(http.StatusInternalServerError, err)
	}
	logs.Debugf("Get task")
}

func (h *TaskHandler) CreateTask(request *restful.Request, response *restful.Response) {
	// 先查询Workflow是否存在
	name := request.PathParameter(TASK_NAME)
	result, err := h.client.Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		logs.Errorf("Get task %s error: %v", name, err)
		//response.WriteError(http.StatusInternalServerError, err)
	}
	if result.Name == name {
		logs.Errorf("Create task %s error, task existed: %v", name, result)
		err = fmt.Errorf("Create task %s error, task existed: %v", name, result)
		response.WriteError(http.StatusInternalServerError, err)
		return
	}

	// 解析用户的输入
	ew := &apis.Task{}
	err = request.ReadEntity(ew)
	if err != nil {
		logs.Errorf("Failed to create task %s, error: %v", name, err)
		response.WriteError(http.StatusInternalServerError, err)
		return
	}
	// TODO: 格式校验
	// TODO: 为Workflow分配ID
	logs.Debugf("Create task %s", name)

	// 将Workflow写入数据库中
	result, err = h.client.Create(context.TODO(), ew, metav1.CreateOptions{})
	if err != nil {
		response.WriteError(http.StatusInternalServerError, err)
		logs.Errorf("Create task %s error: %v", name, err)
		return
	}
	// 返回结果
	err = response.WriteEntity(result)
	if err != nil {
		response.WriteError(http.StatusInternalServerError, err)
	}
	logs.Debugf("Create task %v", result)
}

func (h *TaskHandler) NewGetWebService() *restful.WebService {
	ws := new(restful.WebService)
	ws.Path(TASK_PATH).
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET(fmt.Sprintf("/{%s}", TASK_NAME)).
		To(h.GetTask).
		Doc("Get a task with name").
		Metadata(restfulspec.KeyOpenAPITags, []string{TAG}).
		Operation("Get task").
		Returns(200, "OK", apis.Task{}).
		Returns(400, "Not Found", nil),
	)

	ws.Route(ws.POST(fmt.Sprintf("/{%s}", TASK_NAME)).
		To(h.CreateTask).
		Doc("Create a task with name").
		Metadata(restfulspec.KeyOpenAPITags, []string{TAG}).
		Operation("Create task").
		Returns(200, "OK", apis.Task{}).
		Returns(400, "Not Found", nil),
	)
	return ws
}

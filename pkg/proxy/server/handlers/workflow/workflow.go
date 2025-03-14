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
)

type WorkflowHandler struct {
	client core.WorkflowInterface
}

var _ Handler = &WorkflowHandler{}

func NewWorkflowHandler(clientSet *clients.ClientSet) *WorkflowHandler {
	c := clientSet.Core().Workflows(apis.NamespaceAll)
	return &WorkflowHandler{
		client: c,
	}
}

func (h *WorkflowHandler) GetWorkflow(request *restful.Request, response *restful.Response) {
	name := request.PathParameter(WorkflowName)
	result, err := h.client.Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		logs.Errorf("Get workflow %s error: %v", name, err)
		response.WriteError(http.StatusInternalServerError, err)
	}
	err = response.WriteEntity(result)
	if err != nil {
		response.WriteError(http.StatusInternalServerError, err)
	}
	logs.Debugf("Get workflow")
}

func (h *WorkflowHandler) CreateWorkflow(request *restful.Request, response *restful.Response) {
	// 先查询Workflow是否存在
	name := request.PathParameter(WorkflowName)
	result, err := h.client.Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		logs.Errorf("Get workflow %s error: %v", name, err)
		//response.WriteError(http.StatusInternalServerError, err)
	}
	if result.Name == name {
		logs.Errorf("Create workflow %s error, workflow existed: %v", name, result)
		err = fmt.Errorf("Create workflow %s error, workflow existed: %v", name, result)
		response.WriteError(http.StatusInternalServerError, err)
		return
	}

	// 解析用户的输入
	ew := &apis.Workflow{}
	err = request.ReadEntity(ew)
	if err != nil {
		logs.Errorf("Failed to create workflow %s, error: %v", name, err)
		response.WriteError(http.StatusInternalServerError, err)
		return
	}
	// TODO: 格式校验
	// TODO: 为Workflow分配ID
	logs.Debugf("Create workflow %s", name)

	// 将Workflow写入数据库中
	result, err = h.client.Create(context.TODO(), ew, metav1.CreateOptions{})
	if err != nil {
		response.WriteError(http.StatusInternalServerError, err)
		logs.Errorf("Create workflow %s error: %v", name, err)
		return
	}
	// 返回结果
	err = response.WriteEntity(result)
	if err != nil {
		response.WriteError(http.StatusInternalServerError, err)
	}
	logs.Debugf("Create workflow %v", result)
}

// TODO: Update Workflow
// TODO: Patch Workflow
// TODO: Delete Workflow

func (h *WorkflowHandler) NewGetWebService() *restful.WebService {
	ws := new(restful.WebService)
	ws.Path(WorkflowPath).
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET(fmt.Sprintf("/{%s}", WorkflowName)).
		To(h.GetWorkflow).
		Doc("Get a workflow with name").
		Metadata(restfulspec.KeyOpenAPITags, []string{TAG}).
		Operation("Get workflow").
		Returns(200, "OK", apis.Workflow{}).
		Returns(400, "Not Found", nil),
	)

	ws.Route(ws.POST(fmt.Sprintf("/{%s}", WorkflowName)).
		To(h.CreateWorkflow).
		Doc("Create a workflow with name").
		Metadata(restfulspec.KeyOpenAPITags, []string{TAG}).
		Operation("Create workflow").
		Returns(200, "OK", apis.Workflow{}).
		Returns(400, "Not Found", nil),
	)
	return ws
}

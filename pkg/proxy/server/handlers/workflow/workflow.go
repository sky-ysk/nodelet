package workflow

import (
	"errors"
	"fmt"
	restfulspec "github.com/emicklei/go-restful-openapi/v2"
	"github.com/emicklei/go-restful/v3"
	"github.com/google/uuid"
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/client-go/clients"
	"hit.edu/framework/pkg/client-go/util/manager"
	"hit.edu/framework/pkg/component-base/analyzer"
	"hit.edu/framework/pkg/component-base/logs"
	"hit.edu/framework/pkg/proxy/server/handlers/util"
	"net/http"
	"sync"
	"time"
)

// var _ Handler = &WorkflowHandler{}

type WorkflowHandler struct {
	manager   *manager.Manager
	clientSet *clients.ClientSet
	mu        sync.Mutex
}

var _ Handler = &WorkflowHandler{}

// NewWorkflowHandler 创建一个 WorkflowHandler
func NewWorkflowHandler(clientSet *clients.ClientSet) *WorkflowHandler {
	return &WorkflowHandler{
		clientSet: clientSet,
		manager:   manager.NewManager(clientSet),
	}
}

func (h *WorkflowHandler) GetWorkflow(request *restful.Request, response *restful.Response) {
	// 尝试从url中获取参数
	name := request.QueryParameter(WORKFLOW_NAME)
	if name == "" {
		// url中没有获取到name参数，尝试从请求体中获取
		req := &apis.Workflow{}
		err := request.ReadEntity(&req)
		if err != nil || req.Name == "" {
			err := response.WriteError(http.StatusBadRequest, fmt.Errorf("provide workflow name , the key is Name "))
			if err != nil {
				logs.Errorf("failed to return a status code ")
				return
			}
			return
		} else {
			name = req.Name
		}
	}

	// 从url中获取namespace
	namespace := request.QueryParameter(NAME_SPACE)
	if namespace == "" {
		err := response.WriteError(http.StatusBadRequest, fmt.Errorf("namespace is required"))
		if err != nil {
			logs.Errorf("failed to return a status code ")
			return
		}
		return
	}

	result, err := h.manager.GetWorkflow(name, namespace)
	if err != nil {
		logs.Errorf("Get workflow %s error: %v , workflow not exist! ", name, err)
		err := response.WriteError(http.StatusNotFound, err)
		if err != nil {
			logs.Errorf("failed to return a status code")
			return
		}
		return
	}

	if result.Name == name {
		err = response.WriteEntity(result)
		if err != nil {
			err := response.WriteError(http.StatusOK, err)
			if err != nil {
				logs.Errorf("failed to return a status code")
				return
			}
			return
		}
		logs.Debugf("Get workflow")
	}
}

func (h *WorkflowHandler) CreateWorkflow(request *restful.Request, response *restful.Response) {
	// 获取workflow数据
	ew := &apis.Workflow{}
	err := request.ReadEntity(&ew)
	if err != nil {
		logs.Errorf("Failed to deserialize json data, error: %v", err)
		err := response.WriteError(http.StatusBadRequest, err)
		if err != nil {
			logs.Errorf("failed to return a status code")
			return
		}
		return
	}
	logs.Info(*ew)

	// TODO：循环依赖检查
	// 增加一个简易版的依赖检查，无法检查a1->a2->a3->a1这种
	dependencyErr := util.CheckTaskCircularDependency(ew.Spec)
	if dependencyErr != nil {
		err := response.WriteError(http.StatusBadRequest, dependencyErr)
		if err != nil {
			logs.Errorf("failed to return a status code")
			return
		}
		return
	}

	// 获取namespace
	namespace := ew.Namespace
	if namespace == "" {
		logs.Error("namespace is empty")
		err := response.WriteError(http.StatusBadRequest, fmt.Errorf("namespace is empty"))
		if err != nil {
			logs.Errorf("failed to return a status code ")
			return
		}
		return
	}

	// 创建一个随机串UUID，该串由时间+10位随机串构成
	// UUID应该唯一
	// 同一个Workflow/Task/workflow/Action/Runtime的UUID应该相同
	//UUID
	timestamp := time.Now().Format("20060102T150405")
	randomStr := uuid.New().String()[:5]
	UUID := timestamp + "-" + randomStr
	logs.Debugf("Create workflow  success, workflow id : %s ", UUID)

	var result *apis.Workflow
	if ew.Labels == nil {
		// 创建Workflow without labels
		result, err = h.manager.CreateWorkflow(ew.Spec, namespace, UUID)
		if err != nil {
			err1 := response.WriteError(http.StatusInternalServerError, err)
			if err1 != nil {
				logs.Errorf("failed to return a status code ,error: %v", err1)
				return
			}
			logs.Errorf("Create workflow fail ,failed write it to database , error: %v", err)
			return
		}
	} else {
		// 创建Workflow with labels
		result, err = h.manager.CreateWorkflowWithLabels(ew.Spec, namespace, UUID, ew.Labels)
		if err != nil {
			err1 := response.WriteError(http.StatusInternalServerError, err)
			if err1 != nil {
				logs.Errorf("failed to return a status code ,error: %v", err1)
				return
			}
			logs.Errorf("Create workflow with labels fail ,failed write it to database , error: %v", err)
			return
		}
	}

	// 返回结果
	err = response.WriteEntity(result)
	if err != nil {
		err := response.WriteError(http.StatusInternalServerError, err)
		if err != nil {
			logs.Errorf("failed to return a status code")
			return
		}
		return
	}

	err = response.WriteError(http.StatusOK, err)
	if err != nil {
		logs.Errorf("failed to return a status code ")
		return
	}

	logs.Debugf("Create workflow %v unsupport", result)
}

func (h *WorkflowHandler) UpdateWorkflow(request *restful.Request, response *restful.Response) {
	// 获取workflow
	ew := &apis.Workflow{}
	err := request.ReadEntity(&ew)
	if err != nil {
		logs.Errorf("Failed to deserialize json data, error: %v", err)
		err := response.WriteError(http.StatusBadRequest, err)
		if err != nil {
			logs.Errorf("failed to return a status code")
			return
		}
		return
	}

	// 获取name
	name := request.QueryParameter(WORKFLOW_NAME)
	if name == "" {
		name = ew.Name
	}

	// 获取namespace
	namespace := request.QueryParameter(NAME_SPACE)
	if namespace == "" {
		err := response.WriteError(http.StatusBadRequest, fmt.Errorf("namespace is empty"))
		if err != nil {
			logs.Errorf("failed to return a status code")
			return
		}
		return
	}

	// 格式验证
	res, err := analyzer.SerializeToJson(ew)
	_, err = analyzer.Deserialize(res, apis.Workflow{})
	if err != nil {
		err := response.WriteError(http.StatusBadRequest, err)
		if err != nil {
			logs.Errorf("failed to return a status code ")
			return
		}
		return
	}

	// 更新workflow
	updatedWorkflow, updateErr := h.manager.UpdateWorkflow(name, namespace, ew)
	if updateErr != nil {
		logs.Errorf("Update workflow %s error: %v", name, updateErr)
		var err error
		if errors.Is(updateErr, manager.NotFound) {
			err = response.WriteError(http.StatusNotFound, err)
		} else {
			err = response.WriteError(http.StatusInternalServerError, err)
		}
		if err != nil {
			logs.Errorf("failed to return a status code")
			return
		}
	}

	// 返回成功修改的通知
	err = response.WriteHeaderAndEntity(http.StatusOK, updatedWorkflow)
	if err != nil {
		logs.Errorf("failed to return a status code")
		return
	}

	// 记录日志
	logs.Debugf("update workflow : %v", name)
}

func (h *WorkflowHandler) DeleteWorkflow(request *restful.Request, response *restful.Response) {
	// 尝试从url中获取参数
	name := request.QueryParameter(WORKFLOW_NAME)
	if name == "" {
		// url中没有获取到name参数，尝试从请求体中获取
		req := &apis.Workflow{}
		err := request.ReadEntity(&req)
		if err != nil || req.Name == "" {
			err := response.WriteError(http.StatusBadRequest, fmt.Errorf("provide workflow name , the key is Name "))
			if err != nil {
				logs.Errorf("failed to return a status code ")
				return
			}
			return
		} else {
			name = req.Name
		}
	}

	// 获取namespace
	namespace := request.QueryParameter(NAME_SPACE)
	if namespace == "" {
		err := response.WriteError(http.StatusBadRequest, fmt.Errorf("namespace is empty"))
		if err != nil {
			logs.Errorf("failed to return a status code ")
			return
		}
		return
	}

	// 删除workflow
	var err error
	err = h.manager.DeleteWorkflow(name, namespace)
	if err != nil {
		logs.Errorf("delete workflow %s error: %v", name, err)
		if errors.Is(err, manager.NotFound) {
			err = response.WriteError(http.StatusNotFound, err)
		} else {
			err = response.WriteError(http.StatusInternalServerError, err)
		}
		if err != nil {
			logs.Errorf("failed to return a status code")
			return
		}
		return
	}

	// 返回停止成功的状态
	response.WriteHeader(http.StatusOK)

	// 记录日志
	logs.Debugf("delete workflow : %v", name)
}

func (h *WorkflowHandler) PatchWorkflow(request *restful.Request, response *restful.Response) {
	// 获取json
	req := &apis.Workflow{}
	err := request.ReadEntity(&req)
	if err != nil {
		logs.Errorf("Failed to deserialize json data, error: %v", err)
		err := response.WriteError(http.StatusBadRequest, err)
		if err != nil {
			logs.Errorf("failed to return a status code ")
			return
		}
		return
	}

	// 获取name
	name := request.QueryParameter(WORKFLOW_NAME)
	if name == "" {
		if req.Name != "" {
			name = req.Name
		} else {
			err := response.WriteError(http.StatusBadRequest, fmt.Errorf("name is empty"))
			if err != nil {
				logs.Errorf("failed to return a status code ")
				return
			}
			return
		}
	}

	// 获取namespace
	namespace := request.QueryParameter(NAME_SPACE)
	if namespace == "" {
		err := response.WriteError(http.StatusBadRequest, fmt.Errorf("namespace is empty"))
		if err != nil {
			logs.Errorf("failed to return a status code ")
			return
		}
		return
	}

	// 序列化PatchWorkflow
	patchWorkflow, err := analyzer.SerializeToJson(req)
	if err != nil {
		logs.Errorf("Serialize patch workflow error: %v", err)
		err := response.WriteError(http.StatusInternalServerError, err)
		if err != nil {
			logs.Errorf("failed to return a status code")
			return
		}
	}

	patchedWorkflow, patchedErr := h.manager.PatchWorkflow(namespace, name, []byte(patchWorkflow))
	if patchedErr != nil {
		logs.Errorf("patched workflow %s error: %v", name, err)
		var err error
		if errors.Is(patchedErr, manager.NotFound) {
			err = response.WriteError(http.StatusNotFound, err)
		} else {
			err = response.WriteError(http.StatusInternalServerError, err)
		}
		if err != nil {
			logs.Errorf("failed to return a status code")
			return
		}
		return
	}

	// 返回成功修改的通知
	err = response.WriteHeaderAndEntity(http.StatusOK, patchedWorkflow)
	if err != nil {
		logs.Errorf("failed to return a status code")
		return
	}

	// 记录日志
	logs.Debugf("patch workflow : %v", name)
}

func (h *WorkflowHandler) NewGetWebService() *restful.WebService {
	ws := new(restful.WebService)
	ws.Path(WORKFLOW_PATH).
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/").
		To(h.GetWorkflow).
		Doc("Get a workflow with name").
		Metadata(restfulspec.KeyOpenAPITags, []string{TAG}).
		Param(ws.QueryParameter("Name", "The name of the workflow").DataType("string")).
		Param(ws.QueryParameter("Namespace", "The namespace of the workflow").DataType("string")).
		Operation("Get workflow").
		Returns(200, "OK", apis.Workflow{}).
		Returns(400, "Not Found", nil),
	)

	ws.Route(ws.POST("/").
		To(h.CreateWorkflow).
		Doc("Create a workflow with name").
		Metadata(restfulspec.KeyOpenAPITags, []string{TAG}).
		Param(ws.QueryParameter("Namespace", "The namespace of the workflow").DataType("string")).
		Param(ws.BodyParameter("Workflow", "The json string of the Workflow object").DataType("string")).
		Operation("Create workflow").
		Returns(200, "OK", apis.Workflow{}).
		Returns(400, "Not Found", nil),
	)

	ws.Route(ws.PUT("/").
		To(h.UpdateWorkflow).
		Doc("Update a workflow with name").
		Metadata(restfulspec.KeyOpenAPITags, []string{TAG}).
		Param(ws.QueryParameter("Name", "The name of the workflow").DataType("string")).
		Param(ws.QueryParameter("Namespace", "The namespace of the workflow").DataType("string")).
		Param(ws.BodyParameter("Workflow", "The json string of the Workflow object").DataType("string")).
		Operation("Update workflow").
		Returns(200, "OK", apis.Workflow{}).
		Returns(400, "Not Found", nil),
	)

	ws.Route(ws.PATCH("/").
		To(h.PatchWorkflow).
		Doc("Patch a workflow with name").
		Metadata(restfulspec.KeyOpenAPITags, []string{TAG}).
		Param(ws.QueryParameter("Name", "The name of the workflow").DataType("string")).
		Param(ws.QueryParameter("Namespace", "The namespace of the workflow").DataType("string")).
		Param(ws.BodyParameter("Workflow", "The json string of the Workflow field").DataType("string")).
		Operation("Patch workflow").
		Returns(200, "OK", apis.Workflow{}).
		Returns(400, "Not Found", nil),
	)

	ws.Route(ws.DELETE("/").
		To(h.DeleteWorkflow).
		Doc("Delete a workflow with name").
		Metadata(restfulspec.KeyOpenAPITags, []string{TAG}).
		Param(ws.QueryParameter("Name", "The name of the workflow").DataType("string")).
		Param(ws.QueryParameter("Namespace", "The namespace of the workflow").DataType("string")).
		Operation("Delete workflow").
		Returns(200, "OK", apis.Workflow{}).
		Returns(400, "Not Found", nil))

	return ws
}

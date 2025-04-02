package workflow

import (
	"context"
	"fmt"
	restfulspec "github.com/emicklei/go-restful-openapi/v2"
	"github.com/emicklei/go-restful/v3"
	"github.com/google/uuid"
	"hit.edu/framework/pkg/apimachinery/types"
	apis "hit.edu/framework/pkg/apis/cores"
	metav1 "hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/client-go/clients"
	"hit.edu/framework/pkg/client-go/clients/typed/core"
	"hit.edu/framework/pkg/component-base/analyzer"
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
	// 尝试从url中获取参数
	name := request.QueryParameter(WorkflowName)
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

	result, err := h.client.Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		logs.Errorf("Get workflow %s error: %v", name, err)
		err := response.WriteError(http.StatusInternalServerError, err)
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
	// 先查询Workflow是否存在
	// 尝试从url中获取参数
	name := request.QueryParameter(WorkflowName)
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

	result, err := h.client.Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		logs.Errorf("Get workflow %s error: %v , workflow not exist! ", name, err)
	} else if result.Name == name {
		logs.Errorf("Create workflow %s error, workflow existed: %v", name, result)
		err = fmt.Errorf("create workflow %s error, workflow existed: %v", name, result)
		err := response.WriteError(http.StatusConflict, err)
		if err != nil {
			logs.Errorf("failed to return a status code")
			return
		}
		return
	}

	// 不存在，解析用户的输入
	ew := &apis.Workflow{}
	err = request.ReadEntity(ew)
	if err != nil {
		logs.Errorf("Failed to create workflow %s, error: %v", name, err)
		err := response.WriteError(http.StatusInternalServerError, err)
		if err != nil {
			logs.Errorf("failed to return a status code")
			return
		}
		return
	}

	//格式校验
	res, err := analyzer.SerializeToJson(ew)
	_, err = analyzer.Deserialize(res, apis.Workflow{})
	if err != nil {
		err := response.WriteError(http.StatusBadRequest, err)
		if err != nil {
			logs.Errorf("failed to return a status code")
			return
		}
		// return
	}

	// workflow分配ID
	tu := uuid.New().String()
	ew.Status.WorkflowID = tu
	logs.Debugf("Create workflow %s success, workflow id : %s ", name, tu)

	// 将Workflow写入数据库中
	result, err = h.client.Create(context.TODO(), ew, metav1.CreateOptions{})
	if err != nil {
		err := response.WriteError(http.StatusInternalServerError, err)
		if err != nil {
			logs.Errorf("failed to return a status code")
			return
		}
		logs.Errorf("Create workflow %s ,failed write to database ,error: %v", name, err)
		return
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
	logs.Debugf("Create workflow %v", result)
}

func (h *WorkflowHandler) UpdateWorkflow(request *restful.Request, response *restful.Response) {
	// 先检查workflow是否存在
	// 存在：更新
	// 不存在：返回错误
	// 尝试从url中获取参数
	name := request.QueryParameter(WorkflowName)
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

	// 检查workflow是否存在
	workflow, err := h.client.Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		logs.Errorf("Get workflow %s error: %v , workflow not exist! ", name, err)
		err := response.WriteError(http.StatusNotFound, err)
		if err != nil {
			logs.Errorf("failed to return a status code")
			return
		}
		return
	}

	// workflow 存在，更新
	if workflow.Name == name {
		err := request.ReadEntity(&workflow)
		if err != nil {
			err := response.WriteError(http.StatusInternalServerError, err)
			if err != nil {
				logs.Errorf("failed to return a status code")
				return
			}
			return
		}

		// 格式验证
		res, err := analyzer.SerializeToJson(workflow)
		_, err = analyzer.Deserialize(res, apis.Workflow{})
		if err != nil {
			err := response.WriteError(http.StatusBadRequest, err)
			if err != nil {
				logs.Errorf("failed to return a status code ")
				return
			}
			// return
		}

		updatedWorkflow, updateErr := h.client.Update(context.TODO(), workflow, metav1.UpdateOptions{})
		if updateErr != nil {
			logs.Errorf("Update workflow %s error: %v", name, updateErr)
			err := response.WriteError(http.StatusInternalServerError, err)
			if err != nil {
				logs.Errorf("failed to return a status code")
				return
			}
		}

		//返回成功修改的通知
		err = response.WriteHeaderAndEntity(http.StatusOK, updatedWorkflow)
		if err != nil {
			logs.Errorf("failed to return a status code")
			return
		}

		// 记录日志
		logs.Debugf("update workflow : %v", name)

	}
}

func (h *WorkflowHandler) DeleteWorkflow(request *restful.Request, response *restful.Response) {
	// 查看workflow是否存在
	// 如果存在，删除workflow
	// 如果不存在，返回 404 not found
	// 尝试从url中获取参数
	name := request.QueryParameter(WorkflowName)
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

	// 查看workflow是否存在
	workflow, err := h.client.Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		logs.Errorf("Get workflow %s error: %v ， workflow not exist!  ", name, err)
		err := response.WriteError(http.StatusNotFound, err)
		if err != nil {
			logs.Errorf("failed to return a status code")
			return
		}
	}

	// workflow 存在
	if workflow.Name == name {
		err := h.client.Delete(context.TODO(), name, metav1.DeleteOptions{})
		if err != nil {
			logs.Errorf("Delete workflow %s error: %v", name, err)
			err := response.WriteError(http.StatusInternalServerError, err)
			if err != nil {
				logs.Errorf("failed to return a status code")
				return
			}
		}

		// 返回停止成功的状态
		response.WriteHeader(http.StatusOK)

		// 记录日志
		logs.Debugf("delete workflow : %v", name)
	}
}

func (h *WorkflowHandler) PatchWorkflow(request *restful.Request, response *restful.Response) {
	// 先检查资源是否存在
	// 存在：更新
	// 不存在：返回错误
	// 尝试从url中获取参数
	name := request.QueryParameter(WorkflowName)
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

	// 检查workflow是否存在
	workflow, err := h.client.Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		logs.Errorf("Get workflow %s error: %v , workflow not exist! ", name, err)
		err := response.WriteError(http.StatusNotFound, err)
		if err != nil {
			logs.Errorf("failed to return a status code")
			return
		}
		return
	}

	// workflow 存在，部分更新
	if workflow.Name == name {
		err := request.ReadEntity(workflow)
		if err != nil {
			err := response.WriteError(http.StatusInternalServerError, err)
			if err != nil {
				logs.Errorf("failed to return a status code")
				return
			}
			return
		}
		patchWorkflow, err := analyzer.SerializeToJson(workflow)
		if err != nil {
			logs.Errorf("Serialize patch workflow error: %v", err)
			err := response.WriteError(http.StatusInternalServerError, err)
			if err != nil {
				logs.Errorf("failed to return a status code")
				return
			}
		}
		patchedWorkflow, err := h.client.Patch(context.TODO(), name, types.StrategicMergePatchType, []byte(patchWorkflow), metav1.PatchOptions{})
		if err != nil {
			logs.Errorf("Patch: workflow %s error: %v", name, err)
			err := response.WriteError(http.StatusInternalServerError, err)
			if err != nil {
				logs.Errorf("failed to return a status code")
				return
			}
		}

		//返回成功修改的通知
		err = response.WriteHeaderAndEntity(http.StatusOK, patchedWorkflow)
		if err != nil {
			logs.Errorf("failed to return a status code")
			return
		}

		// 记录日志
		logs.Debugf("update workflow : %v", name)

	}
}

func (h *WorkflowHandler) NewGetWebService() *restful.WebService {
	ws := new(restful.WebService)
	ws.Path(WorkflowPath).
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/").
		To(h.GetWorkflow).
		Doc("Get a workflow with name").
		Metadata(restfulspec.KeyOpenAPITags, []string{TAG}).
		Param(ws.QueryParameter("Name", "The name of the workflow").DataType("string")).
		Operation("Get workflow").
		Returns(200, "OK", apis.Workflow{}).
		Returns(400, "Not Found", nil),
	)

	ws.Route(ws.POST("/").
		To(h.CreateWorkflow).
		Doc("Create a workflow with name").
		Metadata(restfulspec.KeyOpenAPITags, []string{TAG}).
		Param(ws.QueryParameter("Name", "The name of the workflow").DataType("string")).
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
		Operation("Delete workflow").
		Returns(200, "OK", apis.Workflow{}).
		Returns(400, "Not Found", nil))

	return ws
}

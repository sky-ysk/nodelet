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
	"hit.edu/framework/pkg/proxy/server/handlers/util"
	"net/http"
	"time"
)

// var _ Handler = &WorkflowHandler{}

type WorkflowHandler struct {
	manager *util.Manager
}

// NewGroupHandler 创建一个 GroupHandler
//
//	func NewWorkflowHandler(clientSet *clients.ClientSet) *util.WorkflowHandler {
//		c := clientSet.Core().Workflows(apis.NamespaceAll)
//		return &util.WorkflowHandler{
//			Client: c,
//		}
//	}
func NewWorkflowHandler(manager *util.Manager) *WorkflowHandler {
	return &WorkflowHandler{
		manager: manager,
	}
}

func GetWorkflow(request *restful.Request, response *restful.Response) {
	// 尝试从url中获取参数
	c := &WorkflowHandler{}
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
	} else {
		c = util.GetWorkflowClient(namespace)
	}

	result, err := c.client.Get(context.TODO(), name, metav1.GetOptions{})
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

func CreateWorkflow(request *restful.Request, response *restful.Response) {
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
	// 同一个Workflow/Task/Group/Action/Runtime的UUID应该相同
	//UUID
	timestamp := time.Now().Format("20060102T150405")
	randomStr := uuid.New().String()[:10]
	UUID := timestamp + "-" + randomStr
	logs.Debugf("Create workflow  success, workflow id : %s ", UUID)

	// 将Workflow写入数据库中
	result, err := util.CreateWorkflow(ew.Spec, namespace, UUID)
	if err != nil || result == nil {
		err1 := response.WriteError(http.StatusBadRequest, err)
		if err1 != nil {
			logs.Errorf("failed to return a status code")
			return
		}
		logs.Infof("Failed Create workflow %s , error : %s ", err)
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

func UpdateWorkflow(request *restful.Request, response *restful.Response) {
	// 先检查workflow是否存在
	// 存在：更新
	// 不存在：返回错误
	// 尝试从url中获取参数
	c := &CurrentWorkflowHandler{}
	name := request.QueryParameter(WORKFLOW_NAME)
	ew := &apis.Workflow{}
	if name == "" {
		// url中没有获取到name参数，尝试从请求体中获取

		err := request.ReadEntity(&ew)
		if err != nil || ew.Name == "" {
			if err != nil {
				logs.Errorf("Failed to deserialize json data, error: %v", err)
				err := response.WriteError(http.StatusBadRequest, err)
				if err != nil {
					logs.Errorf("failed to return a status code")
					return
				}
				return
			} else {
				err := response.WriteError(http.StatusBadRequest, fmt.Errorf("provide workflow name , the key is Name "))
				if err != nil {
					logs.Errorf("failed to return a status code ")
					return
				}
				return
			}
		} else {
			name = ew.Name
		}
	} else {
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
	}

	namespace := request.QueryParameter(NAME_SPACE)
	if namespace == "" {
		err := response.WriteError(http.StatusBadRequest, fmt.Errorf("namespace is empty"))
		if err != nil {
			logs.Errorf("failed to return a status code")
			return
		}
		return
	} else {
		c = h.GetClient(namespace)
	}

	// 检查workflow是否存在
	workflow, err := c.client.Get(context.TODO(), name, metav1.GetOptions{})
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
		//err := request.ReadEntity(&workflow)
		//if err != nil {
		//	err := response.WriteError(http.StatusInternalServerError, err)
		//	if err != nil {
		//		logs.Errorf("failed to return a status code")
		//		return
		//	}
		//	return
		//}

		// 格式验证
		res, err := analyzer.SerializeToJson(ew)
		_, err = analyzer.Deserialize(res, apis.Workflow{})
		if err != nil {
			err := response.WriteError(http.StatusBadRequest, err)
			if err != nil {
				logs.Errorf("failed to return a status code ")
				return
			}
			// return
		}

		updatedWorkflow, updateErr := c.client.Update(context.TODO(), ew, metav1.UpdateOptions{})
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

func DeleteWorkflow(request *restful.Request, response *restful.Response) {
	// 查看workflow是否存在
	// 如果存在，删除workflow
	// 如果不存在，返回 404 not found
	// 尝试从url中获取参数
	c := &CurrentWorkflowHandler{}
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
	} else {
		c = h.GetClient(namespace)
	}

	// 查看workflow是否存在
	workflow, err := c.client.Get(context.TODO(), name, metav1.GetOptions{})
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
		err := c.client.Delete(context.TODO(), name, metav1.DeleteOptions{})
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

func PatchWorkflow(request *restful.Request, response *restful.Response) {
	// 先检查资源是否存在
	// 存在：更新
	// 不存在：返回错误
	// 尝试从url中获取参数
	c := &CurrentWorkflowHandler{}
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
	} else {
		c = h.GetClient(namespace)
	}

	// 检查workflow是否存在
	workflow, err := c.client.Get(context.TODO(), name, metav1.GetOptions{})
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
		patchedWorkflow, err := c.client.Patch(context.TODO(), name, types.StrategicMergePatchType, []byte(patchWorkflow), metav1.PatchOptions{})
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

func NewGetWebService() *restful.WebService {
	ws := new(restful.WebService)
	ws.Path(WORKFLOW_PATH).
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/").
		To(GetWorkflow).
		Doc("Get a workflow with name").
		Metadata(restfulspec.KeyOpenAPITags, []string{TAG}).
		Param(ws.QueryParameter("Name", "The name of the workflow").DataType("string")).
		Param(ws.QueryParameter("Namespace", "The namespace of the task").DataType("string")).
		Operation("Get workflow").
		Returns(200, "OK", apis.Workflow{}).
		Returns(400, "Not Found", nil),
	)

	ws.Route(ws.POST("/").
		To(CreateWorkflow).
		Doc("Create a workflow with name").
		Metadata(restfulspec.KeyOpenAPITags, []string{TAG}).
		Param(ws.QueryParameter("Name", "The name of the workflow").DataType("string")).
		Param(ws.QueryParameter("Namespace", "The namespace of the task").DataType("string")).
		Param(ws.BodyParameter("Workflow", "The json string of the Workflow object").DataType("string")).
		Operation("Create workflow").
		Returns(200, "OK", apis.Workflow{}).
		Returns(400, "Not Found", nil),
	)

	ws.Route(ws.PUT("/").
		To(UpdateWorkflow).
		Doc("Update a workflow with name").
		Metadata(restfulspec.KeyOpenAPITags, []string{TAG}).
		Param(ws.QueryParameter("Name", "The name of the workflow").DataType("string")).
		Param(ws.QueryParameter("Namespace", "The namespace of the task").DataType("string")).
		Param(ws.BodyParameter("Workflow", "The json string of the Workflow object").DataType("string")).
		Operation("Update workflow").
		Returns(200, "OK", apis.Workflow{}).
		Returns(400, "Not Found", nil),
	)

	ws.Route(ws.PATCH("/").
		To(PatchWorkflow).
		Doc("Patch a workflow with name").
		Metadata(restfulspec.KeyOpenAPITags, []string{TAG}).
		Param(ws.QueryParameter("Name", "The name of the workflow").DataType("string")).
		Param(ws.QueryParameter("Namespace", "The namespace of the task").DataType("string")).
		Param(ws.BodyParameter("Workflow", "The json string of the Workflow field").DataType("string")).
		Operation("Patch workflow").
		Returns(200, "OK", apis.Workflow{}).
		Returns(400, "Not Found", nil),
	)

	ws.Route(ws.DELETE("/").
		To(DeleteWorkflow).
		Doc("Delete a workflow with name").
		Metadata(restfulspec.KeyOpenAPITags, []string{TAG}).
		Param(ws.QueryParameter("Name", "The name of the workflow").DataType("string")).
		Param(ws.QueryParameter("Namespace", "The namespace of the task").DataType("string")).
		Operation("Delete workflow").
		Returns(200, "OK", apis.Workflow{}).
		Returns(400, "Not Found", nil))

	return ws
}

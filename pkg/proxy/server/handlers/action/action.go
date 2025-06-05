package action

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

type ActionHandler struct {
	manager   *manager.Manager
	clientSet *clients.ClientSet
	mu        sync.Mutex
}

var _ Handler = &ActionHandler{}

// NewActionHandler 创建一个 ActionHandler
func NewActionHandler(clientSet *clients.ClientSet) *ActionHandler {
	return &ActionHandler{
		clientSet: clientSet,
		manager:   manager.NewManager(clientSet),
	}
}

func (h *ActionHandler) GetAction(request *restful.Request, response *restful.Response) {
	// 获取name
	name := request.QueryParameter(ACTION_NAME)
	if name == "" {
		err := response.WriteError(http.StatusBadRequest, fmt.Errorf("name is required , the key is Name "))
		if err != nil {
			logs.Errorf("failed to return a status code ")
			return
		}
		return
	}

	// 获取namespace
	namespace := request.QueryParameter(NAME_SPACE)
	if namespace == "" {
		err := response.WriteError(http.StatusBadRequest, fmt.Errorf("namespace is required, the key is Namespace "))
		if err != nil {
			logs.Errorf("failed to return a status code ")
			return
		}
		return
	}

	result, err := h.manager.GetAction(name, namespace)
	if err != nil {
		logs.Errorf("Get action %s error: %v , action not exist! ", name, err)
		err := response.WriteError(http.StatusNotFound, err)
		if err != nil {
			logs.Errorf("failed to return a status code")
			return
		}
		return
	}

	// if result.Name == name {}

	err = response.WriteHeaderAndEntity(http.StatusOK, result)
	if err != nil {
		logs.Errorf("failed to return a status code")
		return
	}

	logs.Debugf("Get action success")
}

func (h *ActionHandler) CreateAction(request *restful.Request, response *restful.Response) {
	// 获取action数据
	ew := &apis.Action{}
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

	// TODO: 还是需要检查spec里面的字段，这里就能防止创建无效的任务
	//格式校验
	res, err := analyzer.SerializeToJson(ew)
	_, err = analyzer.Deserialize(res, apis.Action{})
	if err != nil {
		err := response.WriteError(http.StatusBadRequest, err)
		if err != nil {
			logs.Errorf("failed to return a status code")
			return
		}
		return
	}

	// TODO：循环依赖检查
	// 简易版的依赖检查，无法检查a1->a2->a3->a1这种
	dependencyErr := util.CheckRuntimeCircularDependency(ew.Spec)
	if dependencyErr != nil {
		err := response.WriteError(http.StatusBadRequest, dependencyErr)
		if err != nil {
			logs.Errorf("failed to return a status code")
			return
		}
		return
	}

	// 获取 namespace
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

	logs.Info(*ew)

	// 产生UUID
	timestamp := time.Now().Format("20060102T150405")
	randomStr := uuid.New().String()[:5]
	UUID := timestamp + "-" + randomStr

	var result *apis.Action
	result, err = h.manager.CreateAction(ew.Spec, nil, namespace, UUID, "")
	if err != nil {
		err1 := response.WriteError(http.StatusInternalServerError, err)
		if err1 != nil {
			logs.Errorf("failed to return a status code ,error: %v", err1)
			return
		}
		logs.Errorf("Create action fail ,failed write it to database , error: %v", err)
		return
	}

	// TODO: 改成使用spec中的label创建
	// 创建action without labels
	//if ew.Labels == nil {
	//	result, err = h.manager.CreateAction(ew.Spec, nil, namespace, UUID, "")
	//	if err != nil {
	//		err1 := response.WriteError(http.StatusInternalServerError, err)
	//		if err1 != nil {
	//			logs.Errorf("failed to return a status code ,error: %v", err1)
	//			return
	//		}
	//		logs.Errorf("Create action fail ,failed write it to database , error: %v", err)
	//		return
	//	}
	//}
	//
	//// 创建action with labels
	//result, err = h.manager.CreateActionWithLabels(ew.Spec, nil, namespace, UUID, "", ew.Labels)
	//if err != nil {
	//	err1 := response.WriteError(http.StatusInternalServerError, err)
	//	if err1 != nil {
	//		logs.Errorf("failed to return a status code ,error: %v", err1)
	//		return
	//	}
	//	logs.Errorf("Create action with label fail ,failed write it to database , error: %v", err)
	//	return
	//}

	// 返回结果
	err = response.WriteHeaderAndEntity(http.StatusCreated, result)
	if err != nil {
		logs.Errorf("failed to return a status code")
		return
	}

	logs.Debugf("Create action %v success", result)
}

func (h *ActionHandler) UpdateAction(request *restful.Request, response *restful.Response) {
	// 获取action
	ew := &apis.Action{}
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

	// TODO: 还是需要检查spec里面的字段，这里就能防止创建无效的任务
	//格式校验
	res, err := analyzer.SerializeToJson(ew)
	_, err = analyzer.Deserialize(res, apis.Action{})
	if err != nil {
		err := response.WriteError(http.StatusBadRequest, err)
		if err != nil {
			logs.Errorf("failed to return a status code ")
			return
		}
		return
	}

	// TODO：循环依赖检查
	// 简易版的依赖检查，无法检查a1->a2->a3->a1这种
	dependencyErr := util.CheckRuntimeCircularDependency(ew.Spec)
	if dependencyErr != nil {
		err := response.WriteError(http.StatusBadRequest, dependencyErr)
		if err != nil {
			logs.Errorf("failed to return a status code")
			return
		}
		return
	}

	// 获取name
	name := request.QueryParameter(ACTION_NAME)
	if name == "" {
		err := response.WriteError(http.StatusBadRequest, fmt.Errorf("name is empty"))
		if err != nil {
			logs.Errorf("failed to return a status code")
			return
		}
		return
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

	// TODO:修改此处错乱的error
	// 更新action
	updatedAction, updateErr := h.manager.UpdateAction(name, namespace, ew)
	if updateErr != nil {
		logs.Errorf("Update action %s error: %v", name, updateErr)
		var err error
		if errors.Is(updateErr, manager.NotFound) {
			err = response.WriteError(http.StatusNotFound, updateErr)
		} else if errors.Is(updateErr, manager.InternalServerError) {
			err = response.WriteError(http.StatusInternalServerError, updateErr)
		} else {
			err = response.WriteError(http.StatusInternalServerError, updateErr)
		}
		if err != nil {
			logs.Errorf("failed to return a status code")
			return
		}
	}

	// 返回成功修改的通知
	err = response.WriteHeaderAndEntity(http.StatusOK, updatedAction)
	if err != nil {
		logs.Errorf("failed to return a status code")
		return
	}

	// 记录日志
	logs.Debugf("update action : %v", name)

}

func (h *ActionHandler) DeleteAction(request *restful.Request, response *restful.Response) {
	// 获取name
	name := request.QueryParameter(ACTION_NAME)
	if name == "" {
		err := response.WriteError(http.StatusBadRequest, fmt.Errorf("provide action name , the key is Name "))
		if err != nil {
			logs.Errorf("failed to return a status code ")
			return
		}
		return
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

	// 删除action
	var err error
	err = h.manager.DeleteAction(name, namespace)
	if err != nil {
		logs.Errorf("delete action %s error: %v", name, err)
		if errors.Is(err, manager.NotFound) {
			err = response.WriteError(http.StatusNotFound, err)
		} else if errors.Is(err, manager.InternalServerError) {
			err = response.WriteError(http.StatusInternalServerError, err)
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
	logs.Debugf("delete action : %v", name)
}

func (h *ActionHandler) PatchAction(request *restful.Request, response *restful.Response) {
	// TODO：不可更改字段更改检查
	// 获取 json
	req := &apis.Action{}
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
	name := request.QueryParameter(ACTION_NAME)
	if name == "" {
		err := response.WriteError(http.StatusBadRequest, fmt.Errorf("name is empty"))
		if err != nil {
			logs.Errorf("failed to return a status code ")
			return
		}
		return
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

	// 序列化Patch action
	patchAction, err := analyzer.SerializeToJson(req)
	if err != nil {
		logs.Errorf("Serialize patch action error: %v", err)
		err := response.WriteError(http.StatusInternalServerError, err)
		if err != nil {
			logs.Errorf("failed to return a status code")
			return
		}
	}

	patchedAction, patchedErr := h.manager.PatchAction(namespace, name, []byte(patchAction))
	if patchedErr != nil {
		logs.Errorf("patched action %s error: %v", name, err)
		var err error
		if errors.Is(patchedErr, manager.NotFound) {
			err = response.WriteError(http.StatusNotFound, patchedErr)
		} else if errors.Is(patchedErr, manager.InternalServerError) {
			err = response.WriteError(http.StatusInternalServerError, patchedErr)
		} else {
			err = response.WriteError(http.StatusInternalServerError, patchedErr)
		}
		if err != nil {
			logs.Errorf("failed to return a status code")
			return
		}
		return
	}

	// 返回成功修改的通知
	err = response.WriteHeaderAndEntity(http.StatusOK, patchedAction)
	if err != nil {
		logs.Errorf("failed to return a status code")
		return
	}

	// 记录日志
	logs.Debugf("patch action : %v", name)
}

func (h *ActionHandler) NewGetWebService() *restful.WebService {
	ws := new(restful.WebService)
	ws.Path(ACTION_PATH).
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/").
		To(h.GetAction).
		Doc("Get a action with name").
		Metadata(restfulspec.KeyOpenAPITags, []string{TAG}).
		Param(ws.QueryParameter("Name", "The name of the action").DataType("string")).
		Param(ws.QueryParameter("Namespace", "The namespace of the action").DataType("string")).
		Operation("Get action").
		Returns(200, "OK", apis.Action{}).
		Returns(404, "Not Found", nil),
	)

	ws.Route(ws.POST("/").
		To(h.CreateAction).
		Doc("Create a action with namespace").
		Metadata(restfulspec.KeyOpenAPITags, []string{TAG}).
		Param(ws.QueryParameter("Namespace", "The namespace of the action").DataType("string")).
		Param(ws.BodyParameter("Action", "The json string of the Action object").DataType("string")).
		Operation("Create action").
		Returns(200, "OK", apis.Action{}).
		Returns(404, "Not Found", nil),
	)

	ws.Route(ws.PUT("/").
		To(h.UpdateAction).
		Doc("Update a action with name and namespace").
		Metadata(restfulspec.KeyOpenAPITags, []string{TAG}).
		Param(ws.QueryParameter("Name", "The name of the action").DataType("string")).
		Param(ws.QueryParameter("Namespace", "The namespace of the action").DataType("string")).
		Param(ws.BodyParameter("Action", "The json string of the Action object").DataType("string")).
		Operation("Update action").
		Returns(200, "OK", apis.Action{}).
		Returns(404, "Not Found", nil))

	ws.Route(ws.PATCH("/").
		To(h.PatchAction).
		Doc("Patch a action with name and namespace").
		Metadata(restfulspec.KeyOpenAPITags, []string{TAG}).
		Param(ws.QueryParameter("Name", "The name of the action").DataType("string")).
		Param(ws.QueryParameter("Namespace", "The namespace of the action").DataType("string")).
		Param(ws.BodyParameter("Action", "The json string of the Action field").DataType("string")).
		Operation("Patch action").
		Returns(200, "OK", apis.Action{}).
		Returns(404, "Not Found", nil))

	ws.Route(ws.DELETE("/").
		To(h.DeleteAction).
		Doc("Delete a action with name and namespace").
		Metadata(restfulspec.KeyOpenAPITags, []string{TAG}).
		Param(ws.QueryParameter("Name", "The name of the action").DataType("string")).
		Param(ws.QueryParameter("Namespace", "The namespace of the action").DataType("string")).
		Operation("Delete action").
		Returns(200, "OK", apis.Action{}).
		Returns(404, "Not Found", nil))

	return ws
}

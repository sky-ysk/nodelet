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
	// 尝试从url中获取参数
	name := request.QueryParameter(ACTION_NAME)
	if name == "" {
		// url中没有获取到name参数，尝试从请求体中获取
		req := &apis.Action{}
		err := request.ReadEntity(&req)
		if err != nil || req.Name == "" {
			err := response.WriteError(http.StatusBadRequest, fmt.Errorf("provide action name , the key is Name "))
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
		logs.Debugf("Get action")
	}
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

	////格式校验
	//res, err := analyzer.SerializeToJson(ew)
	//_, err = analyzer.Deserialize(res, apis.Action{})
	//if err != nil {
	//	err := response.WriteError(http.StatusBadRequest, err)
	//	if err != nil {
	//		logs.Errorf("failed to return a status code")
	//		return
	//	}
	//	// return
	//}

	// 产生UUID
	timestamp := time.Now().Format("20060102T150405")
	randomStr := uuid.New().String()[:5]
	UUID := timestamp + "-" + randomStr

	var result *apis.Action

	// 创建action without labels
	if ew.Labels == nil {
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
	} else {
		// 创建action with labels
		result, err = h.manager.CreateActionWithLabels(ew.Spec, nil, namespace, UUID, "", ew.Labels)
		if err != nil {
			err1 := response.WriteError(http.StatusInternalServerError, err)
			if err1 != nil {
				logs.Errorf("failed to return a status code ,error: %v", err1)
				return
			}
			logs.Errorf("Create action with label fail ,failed write it to database , error: %v", err)
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

	logs.Debugf("Create action %v unsupport", result)
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

	// 获取name
	name := request.QueryParameter(ACTION_NAME)
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
	//res, err := analyzer.SerializeToJson(ew)
	//_, err = analyzer.Deserialize(res, apis.Action{})
	//if err != nil {
	//	err := response.WriteError(http.StatusBadRequest, err)
	//	if err != nil {
	//		logs.Errorf("failed to return a status code ")
	//		return
	//	}
	//	return
	//}

	// 更新action
	updatedAction, updateErr := h.manager.UpdateAction(name, namespace, ew)
	if updateErr != nil {
		logs.Errorf("Update action %s error: %v", name, updateErr)
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
	err = response.WriteHeaderAndEntity(http.StatusOK, updatedAction)
	if err != nil {
		logs.Errorf("failed to return a status code")
		return
	}

	// 记录日志
	logs.Debugf("update action : %v", name)

}

func (h *ActionHandler) DeleteAction(request *restful.Request, response *restful.Response) {
	// 尝试从url中获取参数
	name := request.QueryParameter(ACTION_NAME)
	if name == "" {
		// url中没有获取到name参数，尝试从请求体中获取
		req := &apis.Action{}
		err := request.ReadEntity(&req)
		if err != nil || req.Name == "" {
			err := response.WriteError(http.StatusBadRequest, fmt.Errorf("provide action name , the key is Name "))
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

	// 删除action
	var err error
	err = h.manager.DeleteAction(namespace, name)
	if err != nil {
		logs.Errorf("delete action %s error: %v", name, err)
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
	logs.Debugf("delete action : %v", name)
}

func (h *ActionHandler) PatchAction(request *restful.Request, response *restful.Response) {
	// 获取json
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

	// 序列化Patchaction
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
		Returns(400, "Not Found", nil),
	)

	ws.Route(ws.POST("/").
		To(h.CreateAction).
		Doc("Create a action").
		Metadata(restfulspec.KeyOpenAPITags, []string{TAG}).
		Param(ws.QueryParameter("Namespace", "The namespace of the action").DataType("string")).
		Param(ws.BodyParameter("Action", "The json string of the Action object").DataType("string")).
		Operation("Create action").
		Returns(200, "OK", apis.Action{}).
		Returns(400, "Not Found", nil),
	)

	ws.Route(ws.PUT("/").
		To(h.UpdateAction).
		Doc("Update a action").
		Metadata(restfulspec.KeyOpenAPITags, []string{TAG}).
		Param(ws.QueryParameter("Name", "The name of the action").DataType("string")).
		Param(ws.QueryParameter("Namespace", "The namespace of the action").DataType("string")).
		Param(ws.BodyParameter("Action", "The json string of the Action object").DataType("string")).
		Operation("Update action").
		Returns(200, "OK", apis.Action{}).
		Returns(400, "Not Found", nil))

	ws.Route(ws.PATCH("/").
		To(h.PatchAction).
		Doc("Patch a action").
		Metadata(restfulspec.KeyOpenAPITags, []string{TAG}).
		Param(ws.QueryParameter("Name", "The name of the action").DataType("string")).
		Param(ws.QueryParameter("Namespace", "The namespace of the action").DataType("string")).
		Param(ws.BodyParameter("Action", "The json string of the Action field").DataType("string")).
		Operation("Patch action").
		Returns(200, "OK", apis.Action{}).
		Returns(400, "Not Found", nil))

	ws.Route(ws.DELETE("/").
		To(h.DeleteAction).
		Doc("Delete a action").
		Metadata(restfulspec.KeyOpenAPITags, []string{TAG}).
		Param(ws.QueryParameter("Name", "The name of the action").DataType("string")).
		Param(ws.QueryParameter("Namespace", "The namespace of the action").DataType("string")).
		Operation("Delete action").
		Returns(200, "OK", apis.Action{}).
		Returns(400, "Not Found", nil))

	return ws
}

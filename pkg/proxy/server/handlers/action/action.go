package action

import (
	"fmt"
	restfulspec "github.com/emicklei/go-restful-openapi/v2"
	"github.com/emicklei/go-restful/v3"
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/client-go/clients"
	"hit.edu/framework/pkg/client-go/util/manager"
	"hit.edu/framework/pkg/component-base/logs"
	"net/http"
	"sync"
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
	//// 先查询action是否存在
	//// 尝试从url中获取参数
	//name := request.QueryParameter(ACTION_NAME)
	//ew := &apis.Action{}
	//if name == "" {
	//	// url中没有获取到name参数，尝试从请求体中获取
	//	err := request.ReadEntity(&ew)
	//	if err != nil || ew.Name == "" {
	//		if err != nil {
	//			logs.Errorf("Failed to deserialize json data, error: %v", err)
	//			err := response.WriteError(http.StatusBadRequest, err)
	//			if err != nil {
	//				logs.Errorf("failed to return a status code")
	//				return
	//			}
	//			return
	//		} else {
	//			err := response.WriteError(http.StatusBadRequest, fmt.Errorf("provide action name , the key is Name "))
	//			if err != nil {
	//				logs.Errorf("failed to return a status code ")
	//				return
	//			}
	//			return
	//		}
	//	} else {
	//		name = ew.Name
	//	}
	//} else {
	//	err := request.ReadEntity(&ew)
	//	if err != nil {
	//		logs.Errorf("Failed to deserialize json data, error: %v", err)
	//		err := response.WriteError(http.StatusBadRequest, err)
	//		if err != nil {
	//			logs.Errorf("failed to return a status code")
	//			return
	//		}
	//		return
	//	}
	//}
	//
	//namespace := ew.Namespace
	//if namespace == "" {
	//	logs.Error("namespace is empty")
	//	err := response.WriteError(http.StatusBadRequest, fmt.Errorf("namespace is empty"))
	//	if err != nil {
	//		logs.Errorf("failed to return a status code ")
	//		return
	//	}
	//	return
	//}
	//
	//logs.Info(*ew)
	//
	//////格式校验
	////res, err := analyzer.SerializeToJson(ew)
	////_, err = analyzer.Deserialize(res, apis.Action{})
	////if err != nil {
	////	err := response.WriteError(http.StatusBadRequest, err)
	////	if err != nil {
	////		logs.Errorf("failed to return a status code")
	////		return
	////	}
	////	// return
	////}
	//
	//// 将action写入数据库中
	//result, err = h.manager.CreateAction()
	//if err != nil {
	//	err1 := response.WriteError(http.StatusInternalServerError, err)
	//	if err1 != nil {
	//		logs.Errorf("failed to return a status code ,error: %v", err1)
	//		return
	//	}
	//	logs.Errorf("Create action %s ,failed write to database , error: %v", name, err)
	//	return
	//}
	//
	//// 返回结果
	//err = response.WriteEntity(result)
	//if err != nil {
	//	err := response.WriteError(http.StatusInternalServerError, err)
	//	if err != nil {
	//		logs.Errorf("failed to return a status code")
	//		return
	//	}
	//	return
	//}
	//
	//err = response.WriteError(http.StatusOK, err)
	//if err != nil {
	//	logs.Errorf("failed to return a status code ")
	//	return
	//}
	//
	//logs.Debugf("Create action %v unsupport", result)
}

func (h *ActionHandler) UpdateAction(request *restful.Request, response *restful.Response) {
	//// 先检查action是否存在
	//// 存在：更新
	//// 不存在：返回错误
	//// 尝试从url中获取参数
	//c := &CurrentActionHandler{}
	//name := request.QueryParameter(ACTION_NAME)
	//ew := &apis.Action{}
	//if name == "" {
	//	// url中没有获取到name参数，尝试从请求体中获取
	//	err := request.ReadEntity(&ew)
	//	if err != nil || ew.Name == "" {
	//		if err != nil {
	//			logs.Errorf("Failed to deserialize json data, error: %v", err)
	//			err := response.WriteError(http.StatusBadRequest, err)
	//			if err != nil {
	//				logs.Errorf("failed to return a status code")
	//				return
	//			}
	//			return
	//		} else {
	//			err := response.WriteError(http.StatusBadRequest, fmt.Errorf("provide action name , the key is Name "))
	//			if err != nil {
	//				logs.Errorf("failed to return a status code ")
	//				return
	//			}
	//			return
	//		}
	//	} else {
	//		name = ew.Name
	//	}
	//} else {
	//	err := request.ReadEntity(&ew)
	//	if err != nil {
	//		logs.Errorf("Failed to deserialize json data, error: %v", err)
	//		err := response.WriteError(http.StatusBadRequest, err)
	//		if err != nil {
	//			logs.Errorf("failed to return a status code")
	//			return
	//		}
	//		return
	//	}
	//}
	//
	//namespace := request.QueryParameter(NAME_SPACE)
	//if namespace == "" {
	//	err := response.WriteError(http.StatusBadRequest, fmt.Errorf("namespace is empty"))
	//	if err != nil {
	//		logs.Errorf("failed to return a status code")
	//		return
	//	}
	//	return
	//} else {
	//	c = h.GetClient(namespace)
	//}
	//
	//// 检查action是否存在
	//action, err := c.client.Get(context.TODO(), name, metav1.GetOptions{})
	//if err != nil {
	//	logs.Errorf("Get action %s error: %v , action not exist !", name, err)
	//	err := response.WriteError(http.StatusNotFound, err)
	//	if err != nil {
	//		logs.Errorf("failed to return a status code")
	//		return
	//	}
	//	return
	//}
	//
	//// action 存在，更新
	//if action.Name == name {
	//	//err := request.ReadEntity(&action)
	//	//if err != nil {
	//	//	err := response.WriteError(http.StatusInternalServerError, err)
	//	//	if err != nil {
	//	//		logs.Errorf("failed to return a status code")
	//	//		return
	//	//	}
	//	//	return
	//	//}
	//
	//	// logs.Debugf("Update action to : %v", ew)
	//
	//	// 格式验证
	//	res, err := analyzer.SerializeToJson(ew)
	//	_, err = analyzer.Deserialize(res, apis.Action{})
	//	if err != nil {
	//		err := response.WriteError(http.StatusBadRequest, err)
	//		if err != nil {
	//			logs.Errorf("failed to return a status code ")
	//			return
	//		}
	//		// return
	//	}
	//
	//	updatedAction, updateErr := c.client.Update(context.TODO(), ew, metav1.UpdateOptions{})
	//	if updateErr != nil {
	//		logs.Errorf("Update action %s error: %v", name, updateErr)
	//		err := response.WriteError(http.StatusInternalServerError, err)
	//		if err != nil {
	//			logs.Errorf("failed to return a status code")
	//			return
	//		}
	//	}
	//
	//	// 返回成功修改的通知
	//	err = response.WriteHeaderAndEntity(http.StatusOK, updatedAction)
	//	if err != nil {
	//		logs.Errorf("failed to return a status code")
	//		return
	//	}
	//
	//	// 记录日志
	//	logs.Debugf("update action : %v", name)
	//
	//}
}

func (h *ActionHandler) DeleteAction(request *restful.Request, response *restful.Response) {
	//// 查看action是否存在
	//// 存在，删除节点
	//// 不存在，返回 404 not found
	//// 尝试从url中获取参数
	//c := &CurrentActionHandler{}
	//name := request.QueryParameter(ACTION_NAME)
	//if name == "" {
	//	// url中没有获取到name参数，尝试从请求体中获取
	//	req := &apis.Action{}
	//	err := request.ReadEntity(&req)
	//	if err != nil || req.Name == "" {
	//		err := response.WriteError(http.StatusBadRequest, fmt.Errorf("provide action name , the key is Name "))
	//		if err != nil {
	//			logs.Errorf("failed to return a status code ")
	//			return
	//		}
	//		return
	//	} else {
	//		name = req.Name
	//	}
	//}
	//
	//// 获取namespace
	//namespace := request.QueryParameter(NAME_SPACE)
	//if namespace == "" {
	//	err := response.WriteError(http.StatusBadRequest, fmt.Errorf("namespace is empty"))
	//	if err != nil {
	//		logs.Errorf("failed to return a status code ")
	//		return
	//	}
	//	return
	//} else {
	//	c = h.GetClient(namespace)
	//}
	//
	//// 查看action是否存在
	//action, err := c.client.Get(context.TODO(), name, metav1.GetOptions{})
	//if err != nil {
	//	logs.Errorf("Get action %s error: %v , action not exist !", name, err)
	//	err := response.WriteError(http.StatusNotFound, err)
	//	if err != nil {
	//		logs.Errorf("failed to return a status code")
	//		return
	//	}
	//}
	//
	//// action 存在
	//if action.Name == name {
	//	err := c.client.Delete(context.TODO(), name, metav1.DeleteOptions{})
	//	if err != nil {
	//		logs.Errorf("Delete action %s error: %v", name, err)
	//		err := response.WriteError(http.StatusInternalServerError, err)
	//		if err != nil {
	//			logs.Errorf("failed to return a status code")
	//			return
	//		}
	//	}
	//
	//	// 返回停止成功的状态
	//	response.WriteHeader(http.StatusOK)
	//
	//	// 记录日志
	//	logs.Debugf("delete action : %v", name)
	//
	//}
}

func (h *ActionHandler) PatchAction(request *restful.Request, response *restful.Response) {
	//// 先检查action是否存在
	//// 存在：部分更新
	//// 不存在：返回错误
	//// 尝试从url中获取参数
	//c := &CurrentActionHandler{}
	//name := request.QueryParameter(ACTION_NAME)
	//if name == "" {
	//	// url中没有获取到name参数，尝试从请求体中获取
	//	req := &apis.Action{}
	//	err := request.ReadEntity(&req)
	//	if err != nil || req.Name == "" {
	//		err := response.WriteError(http.StatusBadRequest, fmt.Errorf("provide action name , the key is Name "))
	//		if err != nil {
	//			logs.Errorf("failed to return a status code ")
	//			return
	//		}
	//		return
	//	} else {
	//		name = req.Name
	//	}
	//}
	//
	//// 获取namespace
	//namespace := request.QueryParameter(NAME_SPACE)
	//if namespace == "" {
	//	err := response.WriteError(http.StatusBadRequest, fmt.Errorf("namespace is empty"))
	//	if err != nil {
	//		logs.Errorf("failed to return a status code ")
	//		return
	//	}
	//	return
	//} else {
	//	c = h.GetClient(namespace)
	//}
	//
	//// 检查action是否存在
	//action, err := c.client.Get(context.TODO(), name, metav1.GetOptions{})
	//if err != nil {
	//	logs.Errorf("Get action %s error: %v , action not exist !", name, err)
	//	err := response.WriteError(http.StatusNotFound, err)
	//	if err != nil {
	//		logs.Errorf("failed to return a status code")
	//		return
	//	}
	//	return
	//}
	//
	//// action 存在，部分更新
	//if action.Name == name {
	//	err := request.ReadEntity(action)
	//	if err != nil {
	//		err := response.WriteError(http.StatusInternalServerError, err)
	//		if err != nil {
	//			logs.Errorf("failed to return a status code")
	//			return
	//		}
	//		return
	//	}
	//
	//	patchAction, err := analyzer.SerializeToJson(action)
	//	if err != nil {
	//		logs.Errorf("Serialize patch action error: %v", err)
	//		err := response.WriteError(http.StatusInternalServerError, err)
	//		if err != nil {
	//			logs.Errorf("failed to return a status code")
	//			return
	//		}
	//	}
	//	patchedAction, err := c.client.Patch(context.TODO(), name, types.StrategicMergePatchType, []byte(patchAction), metav1.PatchOptions{})
	//	if err != nil {
	//		logs.Errorf("Patch action %s error: %v", name, err)
	//		err := response.WriteError(http.StatusInternalServerError, err)
	//		if err != nil {
	//			logs.Errorf("failed to return a status code")
	//			return
	//		}
	//	}
	//
	//	// 返回成功修改的通知
	//	err = response.WriteHeaderAndEntity(http.StatusOK, patchedAction)
	//	if err != nil {
	//		logs.Errorf("failed to return a status code")
	//		return
	//	}
	//
	//	// 记录日志
	//	logs.Debugf("patch action : %v", name)
	//
	//}
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
		Param(ws.QueryParameter("Name", "The name of the action").DataType("string")).
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

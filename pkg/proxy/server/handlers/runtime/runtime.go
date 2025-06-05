package runtime

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
	"io"
	"net/http"
	"sync"
	"time"
)

type RuntimeHandler struct {
	manager   *manager.Manager
	clientSet *clients.ClientSet
	mu        sync.Mutex
}

var _ Handler = &RuntimeHandler{}

// NewRuntimeHandler 创建一个 RuntimeHandler
func NewRuntimeHandler(clientSet *clients.ClientSet) *RuntimeHandler {
	return &RuntimeHandler{
		clientSet: clientSet,
		manager:   manager.NewManager(clientSet),
	}
}

func (h *RuntimeHandler) GetRuntime(request *restful.Request, response *restful.Response) {
	// 获取name
	name := request.QueryParameter(RUNTIME_NAME)
	if name == "" {
		err := response.WriteError(http.StatusBadRequest, fmt.Errorf("provide runtime name , the key is Name "))
		if err != nil {
			logs.Errorf("failed to return a status code ")
			return
		}
		return
	}

	// 获取namespace
	namespace := request.QueryParameter(NAME_SPACE)
	if namespace == "" {
		err := response.WriteError(http.StatusBadRequest, fmt.Errorf("namespace is required"))
		if err != nil {
			logs.Errorf("failed to return a status code ")
			return
		}
		return
	}

	result, err := h.manager.GetRuntime(name, namespace)
	if err != nil {
		logs.Errorf("Get runtime %s error: %v , runtime not exist! ", name, err)
		err := response.WriteError(http.StatusNotFound, err)
		if err != nil {
			logs.Errorf("failed to return a status code")
			return
		}
		return
	}

	//if result.Name == name {
	//	err = response.WriteEntity(result)
	//	if err != nil {
	//		err := response.WriteError(http.StatusOK, err)
	//		if err != nil {
	//			logs.Errorf("failed to return a status code")
	//			return
	//		}
	//		return
	//	}
	//	logs.Debugf("Get runtime")
	//}

	err = response.WriteHeaderAndEntity(http.StatusOK, result)
	if err != nil {
		logs.Errorf("failed to return a status code")
		return
	}

	logs.Debugf("Get runtime success")
}

func (h *RuntimeHandler) CreateRuntime(request *restful.Request, response *restful.Response) {
	// 获取runtime数据
	ew := &apis.Runtime{}
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
	_, err = analyzer.Deserialize(res, apis.Runtime{})
	if err != nil {
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

	// 产生UUID
	timestamp := time.Now().Format("20060102T150405")
	randomStr := uuid.New().String()[:5]
	UUID := timestamp + "-" + randomStr

	var result *apis.Runtime
	result, err = h.manager.CreateRuntime(ew.Spec, nil, namespace, UUID, "")
	if err != nil {
		err1 := response.WriteError(http.StatusInternalServerError, err)
		if err1 != nil {
			logs.Errorf("failed to return a status code ,error: %v", err1)
			return
		}
		logs.Errorf("Create runtime fail ,failed write it to database , error: %v", err)
		return
	}

	// TODO: 改成使用spec中的label创建
	//if ew.Labels == nil {
	//	// 创建runtime without label
	//	result, err = h.manager.CreateRuntime(ew.Spec, nil, namespace, UUID, "")
	//	if err != nil {
	//		err1 := response.WriteError(http.StatusInternalServerError, err)
	//		if err1 != nil {
	//			logs.Errorf("failed to return a status code ,error: %v", err1)
	//			return
	//		}
	//		logs.Errorf("Create runtime fail ,failed write it to database , error: %v", err)
	//		return
	//	}
	//} else {
	//	// 创建runtime with label
	//	result, err = h.manager.CreateRuntimeWithLabels(ew.Spec, nil, namespace, UUID, "", ew.Labels)
	//	if err != nil {
	//		err1 := response.WriteError(http.StatusInternalServerError, err)
	//		if err1 != nil {
	//			logs.Errorf("failed to return a status code ,error: %v", err1)
	//			return
	//		}
	//		logs.Errorf("Create runtime with labels fail ,failed write it to database , error: %v", err)
	//		return
	//	}
	//}

	// 返回结果
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

	// 返回结果
	err = response.WriteHeaderAndEntity(http.StatusCreated, result)
	if err != nil {
		logs.Errorf("failed to return a status code")
		return
	}

	logs.Debugf("Create runtime %v success", result)
}

func (h *RuntimeHandler) UpdateRuntime(request *restful.Request, response *restful.Response) {
	// 获取runtime
	ew := &apis.Runtime{}
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

	// TODO：格式验证
	// 格式验证
	res, err := analyzer.SerializeToJson(ew)
	_, err = analyzer.Deserialize(res, apis.Runtime{})
	if err != nil {
		err := response.WriteError(http.StatusBadRequest, err)
		if err != nil {
			logs.Errorf("failed to return a status code ")
			return
		}
		return
	}

	// 获取name
	name := request.QueryParameter(RUNTIME_NAME)
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

	// 更新runtime
	updatedRuntime, updateErr := h.manager.UpdateRuntime(name, namespace, ew)
	if updateErr != nil {
		logs.Errorf("Update runtime %s error: %v", name, updateErr)
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
	err = response.WriteHeaderAndEntity(http.StatusOK, updatedRuntime)
	if err != nil {
		logs.Errorf("failed to return a status code")
		return
	}

	// 记录日志
	logs.Debugf("update runtime : %v", name)
}

func (h *RuntimeHandler) DeleteRuntime(request *restful.Request, response *restful.Response) {
	// 获取name
	name := request.QueryParameter(RUNTIME_NAME)
	if name == "" {
		err := response.WriteError(http.StatusBadRequest, fmt.Errorf("provide runtime name , the key is Name "))
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

	// 删除runtime
	var err error
	err = h.manager.DeleteRuntime(name, namespace)
	if err != nil {
		logs.Errorf("delete runtime %s error: %v", name, err)
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
	logs.Debugf("delete runtime : %v", name)
}

func (h *RuntimeHandler) PatchRuntime(request *restful.Request, response *restful.Response) {
	// 获取json
	//req := &apis.Runtime{}
	//err := request.ReadEntity(&req)
	//if err != nil {
	//	logs.Errorf("Failed to deserialize json data, error: %v", err)
	//	err := response.WriteError(http.StatusBadRequest, err)
	//	if err != nil {
	//		logs.Errorf("failed to return a status code ")
	//		return
	//	}
	//	return
	//}

	// TODO: 这似乎不是我改的，不过这么写确实也可以哈

	// 获取json
	bodyBytes, err := io.ReadAll(request.Request.Body)
	if err != nil {
		err := response.WriteError(http.StatusBadRequest, err)
		if err != nil {
			logs.Errorf("failed to return a status code ")
			return
		}
		return
	}
	jsonStr := string(bodyBytes)

	// 获取name
	name := request.QueryParameter(RUNTIME_NAME)
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

	// 序列化Patchruntime
	//patchRuntime, err := analyzer.SerializeToJson(req)
	//if err != nil {
	//	logs.Errorf("Serialize patch runtime error: %v", err)
	//	err := response.WriteError(http.StatusInternalServerError, err)
	//	if err != nil {
	//		logs.Errorf("failed to return a status code")
	//		return
	//	}
	//}

	patchedRuntime, patchedErr := h.manager.PatchRuntime(namespace, name, []byte(jsonStr))
	if patchedErr != nil {
		logs.Errorf("patched runtime %s error: %v", name, err)
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
	err = response.WriteHeaderAndEntity(http.StatusOK, patchedRuntime)
	if err != nil {
		logs.Errorf("failed to return a status code")
		return
	}

	// 记录日志
	logs.Debugf("patch runtime : %v", name)
}

func (h *RuntimeHandler) NewGetWebService() *restful.WebService {
	ws := new(restful.WebService)
	ws.Path(RUNTIME_PATH).
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/").
		To(h.GetRuntime).
		Doc("Get a runtime with name").
		Metadata(restfulspec.KeyOpenAPITags, []string{TAG}).
		Param(ws.QueryParameter("Name", "The name of the runtime").DataType("string")).
		Param(ws.QueryParameter("Namespace", "The namespace of the runtime").DataType("string")).
		Operation("Get runtime").
		Returns(200, "OK", apis.Runtime{}).
		Returns(404, "Not Found", nil),
	)

	ws.Route(ws.POST("/").
		To(h.CreateRuntime).
		Doc("Create a runtime").
		Metadata(restfulspec.KeyOpenAPITags, []string{TAG}).
		Param(ws.QueryParameter("Namespace", "The namespace of the runtime").DataType("string")).
		Param(ws.BodyParameter("Runtime", "The json string of the Runtime object").DataType("string")).
		Operation("Create runtime").
		Returns(200, "OK", apis.Runtime{}).
		Returns(404, "Not Found", nil),
	)

	ws.Route(ws.PUT("/").
		To(h.UpdateRuntime).
		Doc("Update a runtime").
		Metadata(restfulspec.KeyOpenAPITags, []string{TAG}).
		Param(ws.QueryParameter("Name", "The name of the runtime").DataType("string")).
		Param(ws.QueryParameter("Namespace", "The namespace of the runtime").DataType("string")).
		Param(ws.BodyParameter("Runtime", "The json string of the Runtime object").DataType("string")).
		Operation("Update runtime").
		Returns(200, "OK", apis.Runtime{}).
		Returns(404, "Not Found", nil))

	ws.Route(ws.PATCH("/").
		To(h.PatchRuntime).
		Doc("Patch a runtime").
		Metadata(restfulspec.KeyOpenAPITags, []string{TAG}).
		Param(ws.QueryParameter("Name", "The name of the runtime").DataType("string")).
		Param(ws.QueryParameter("Namespace", "The namespace of the runtime").DataType("string")).
		Param(ws.BodyParameter("Runtime", "The json string of the Runtime field").DataType("string")).
		Operation("Patch runtime").
		Returns(200, "OK", apis.Runtime{}).
		Returns(404, "Not Found", nil))

	ws.Route(ws.DELETE("/").
		To(h.DeleteRuntime).
		Doc("Delete a runtime").
		Metadata(restfulspec.KeyOpenAPITags, []string{TAG}).
		Param(ws.QueryParameter("Name", "The name of the runtime").DataType("string")).
		Param(ws.QueryParameter("Namespace", "The namespace of the runtime").DataType("string")).
		Operation("Delete runtime").
		Returns(200, "OK", apis.Runtime{}).
		Returns(404, "Not Found", nil))

	return ws
}

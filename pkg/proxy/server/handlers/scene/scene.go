package scene

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

type SceneHandler struct {
	manager   *manager.Manager
	clientSet *clients.ClientSet
	mu        sync.Mutex
}

var _ Handler = &SceneHandler{}

// NewSceneHandler 创建一个 SceneHandler
func NewSceneHandler(clientSet *clients.ClientSet) *SceneHandler {
	return &SceneHandler{
		clientSet: clientSet,
		manager:   manager.NewManager(clientSet),
	}
}

func (h *SceneHandler) GetScene(request *restful.Request, response *restful.Response) {
	// 获取name
	name := request.QueryParameter(SCENE_NAME)
	if name == "" {
		err := response.WriteError(http.StatusBadRequest, fmt.Errorf("provide scene name , the key is Name "))
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

	result, err := h.manager.GetScene(name, namespace)
	if err != nil {
		logs.Errorf("Get scene %s error: %v , scene not exist! ", name, err)
		err := response.WriteError(http.StatusNotFound, err)
		if err != nil {
			logs.Errorf("failed to return a status code")
			return
		}
		return
	}

	err = response.WriteHeaderAndEntity(http.StatusOK, result)
	if err != nil {
		logs.Errorf("failed to return a status code")
		return
	}

	logs.Debugf("Get scene success")
}

func (h *SceneHandler) CreateScene(request *restful.Request, response *restful.Response) {
	// 获取scene数据
	ew := &apis.Scene{}
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

	// TODO：格式校验
	//格式校验
	res, err := analyzer.SerializeToJson(ew)
	_, err = analyzer.Deserialize(res, apis.Scene{})
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

	var result *apis.Scene

	// 创建scene
	result, err = h.manager.CreateScene(ew.Spec, namespace, UUID)
	if err != nil {
		err1 := response.WriteError(http.StatusInternalServerError, err)
		if err1 != nil {
			logs.Errorf("failed to return a status code ,error: %v", err1)
			return
		}
		logs.Errorf("Create scene fail ,failed write it to database , error: %v", err)
		return
	}

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

	logs.Debugf("Create scene %v unsupport", result)
}

func (h *SceneHandler) UpdateScene(request *restful.Request, response *restful.Response) {
	// 获取scene
	ew := &apis.Scene{}
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
	_, err = analyzer.Deserialize(res, apis.Scene{})
	if err != nil {
		err := response.WriteError(http.StatusBadRequest, err)
		if err != nil {
			logs.Errorf("failed to return a status code ")
			return
		}
		return
	}

	// 获取name
	name := request.QueryParameter(SCENE_NAME)
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

	// 更新Scene
	updatedScene, updateErr := h.manager.UpdateScene(name, namespace, ew)
	if updateErr != nil {
		logs.Errorf("Update scene %s error: %v", name, updateErr)
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
	err = response.WriteHeaderAndEntity(http.StatusOK, updatedScene)
	if err != nil {
		logs.Errorf("failed to return a status code")
		return
	}

	// 记录日志
	logs.Debugf("update scene : %v", name)

}

func (h *SceneHandler) DeleteScene(request *restful.Request, response *restful.Response) {
	// 获取name
	name := request.QueryParameter(SCENE_NAME)
	if name == "" {
		err := response.WriteError(http.StatusBadRequest, fmt.Errorf("provide scene name , the key is Name "))
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

	// 删除scene
	var err error
	err = h.manager.DeleteScene(namespace, name)
	if err != nil {
		logs.Errorf("delete scene %s error: %v", name, err)
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
	logs.Debugf("delete scene : %v", name)
}

func (h *SceneHandler) PatchScene(request *restful.Request, response *restful.Response) {
	// 获取json
	req := &apis.Scene{}
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
	name := request.QueryParameter(SCENE_NAME)
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

	// 序列化Patch scene
	patchScene, err := analyzer.SerializeToJson(req)
	if err != nil {
		logs.Errorf("Serialize patch scene error: %v", err)
		err := response.WriteError(http.StatusInternalServerError, err)
		if err != nil {
			logs.Errorf("failed to return a status code")
			return
		}
	}

	patchedScene, patchedErr := h.manager.PatchScene(namespace, name, []byte(patchScene))
	if patchedErr != nil {
		logs.Errorf("patched scene %s error: %v", name, err)
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
	err = response.WriteHeaderAndEntity(http.StatusOK, patchedScene)
	if err != nil {
		logs.Errorf("failed to return a status code")
		return
	}

	// 记录日志
	logs.Debugf("patch scene : %v", name)
}

func (h *SceneHandler) NewGetWebService() *restful.WebService {
	ws := new(restful.WebService)
	ws.Path(SCENE_PATH).
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/").
		To(h.GetScene).
		Doc("Get a scene with name").
		Metadata(restfulspec.KeyOpenAPITags, []string{TAG}).
		Param(ws.QueryParameter("Name", "The name of the scene").DataType("string")).
		Param(ws.QueryParameter("Namespace", "The namespace of the scene").DataType("string")).
		Operation("Get scene").
		Returns(200, "OK", apis.Scene{}).
		Returns(404, "Not Found", nil),
	)

	ws.Route(ws.POST("/").
		To(h.CreateScene).
		Doc("Create a scene").
		Metadata(restfulspec.KeyOpenAPITags, []string{TAG}).
		Param(ws.QueryParameter("Namespace", "The namespace of the scene").DataType("string")).
		Param(ws.BodyParameter("Scene", "The json string of the Scene object").DataType("string")).
		Operation("Create scene").
		Returns(200, "OK", apis.Scene{}).
		Returns(404, "Not Found", nil),
	)

	ws.Route(ws.PUT("/").
		To(h.UpdateScene).
		Doc("Update a scene").
		Metadata(restfulspec.KeyOpenAPITags, []string{TAG}).
		Param(ws.QueryParameter("Name", "The name of the scene").DataType("string")).
		Param(ws.QueryParameter("Namespace", "The namespace of the scene").DataType("string")).
		Param(ws.BodyParameter("Scene", "The json string of the scene object").DataType("string")).
		Operation("Update scene").
		Returns(200, "OK", apis.Scene{}).
		Returns(404, "Not Found", nil))

	ws.Route(ws.PATCH("/").
		To(h.PatchScene).
		Doc("Patch a scene").
		Metadata(restfulspec.KeyOpenAPITags, []string{TAG}).
		Param(ws.QueryParameter("Name", "The name of the scene").DataType("string")).
		Param(ws.QueryParameter("Namespace", "The namespace of the scene").DataType("string")).
		Param(ws.BodyParameter("Scene", "The json string of the Scene field").DataType("string")).
		Operation("Patch scene").
		Returns(200, "OK", apis.Scene{}).
		Returns(404, "Not Found", nil))

	ws.Route(ws.DELETE("/").
		To(h.DeleteScene).
		Doc("Delete a scene").
		Metadata(restfulspec.KeyOpenAPITags, []string{TAG}).
		Param(ws.QueryParameter("Name", "The name of the scene").DataType("string")).
		Param(ws.QueryParameter("Namespace", "The namespace of the scene").DataType("string")).
		Operation("Delete scene").
		Returns(200, "OK", apis.Scene{}).
		Returns(404, "Not Found", nil))

	return ws
}

package group

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

type GroupHandler struct {
	manager   *manager.Manager
	clientSet *clients.ClientSet
	mu        sync.Mutex
}

var _ Handler = &GroupHandler{}

// NewGroupHandler 创建一个 GroupHandler
func NewGroupHandler(clientSet *clients.ClientSet) *GroupHandler {
	return &GroupHandler{
		clientSet: clientSet,
		manager:   manager.NewManager(clientSet),
	}
}

func (h *GroupHandler) GetGroup(request *restful.Request, response *restful.Response) {
	// 获取name
	name := request.QueryParameter(GROUP_NAME)
	if name == "" {
		err := response.WriteError(http.StatusBadRequest, fmt.Errorf("provide group name , the key is Name "))
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

	result, err := h.manager.GetGroup(name, namespace)
	if err != nil {
		logs.Errorf("Get group %s error: %v , group not exist! ", name, err)
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

	logs.Debugf("Get group success")
}

func (h *GroupHandler) CreateGroup(request *restful.Request, response *restful.Response) {
	// 获取group数据
	ew := &apis.Group{}
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

	// TODO: 格式校验
	////格式校验
	//res, err := analyzer.SerializeToJson(ew)
	//_, err = analyzer.Deserialize(res, apis.Group{})
	//if err != nil {
	//	err := response.WriteError(http.StatusBadRequest, err)
	//	if err != nil {
	//		logs.Errorf("failed to return a status code")
	//		return
	//	}
	//	return
	//}

	// TODO：循环依赖检查
	// 简易版的依赖检查，无法检查a1->a2->a3->a1这种
	dependencyErr := util.CheckActionCircularDependency(ew.Spec)
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

	// TODO：使用spec字段中的label创建label
	// 创建group without label
	//if ew.Labels == nil {
	//	// 创建group
	//	result, err = h.manager.CreateGroup(ew.Spec, nil, namespace, UUID, "")
	//	if err != nil {
	//		err1 := response.WriteError(http.StatusInternalServerError, err)
	//		if err1 != nil {
	//			logs.Errorf("failed to return a status code ,error: %v", err1)
	//			return
	//		}
	//		logs.Errorf("Create group fail ,failed write it to database , error: %v", err)
	//		return
	//	}
	//} else {
	//	// 创建group with label
	//	result, err = h.manager.CreateGroupWithLabels(ew.Spec, nil, namespace, UUID, "", ew.Labels)
	//	if err != nil {
	//		err1 := response.WriteError(http.StatusInternalServerError, err)
	//		if err1 != nil {
	//			logs.Errorf("failed to return a status code ,error: %v", err1)
	//			return
	//		}
	//		logs.Errorf("Create group  with label  fail ,failed write it to database , error: %v", err)
	//		return
	//	}
	//
	//}

	var result *apis.Group
	result, err = h.manager.CreateGroup(ew.Spec, nil, namespace, UUID, "")
	if err != nil {
		err1 := response.WriteError(http.StatusInternalServerError, err)
		if err1 != nil {
			logs.Errorf("failed to return a status code ,error: %v", err1)
			return
		}
		logs.Errorf("Create group fail ,failed write it to database , error: %v", err)
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

	err = response.WriteHeaderAndEntity(http.StatusCreated, result)
	if err != nil {
		logs.Errorf("failed to return a status code")
		return
	}

	logs.Debugf("Create group %v success", result)
}

func (h *GroupHandler) UpdateGroup(request *restful.Request, response *restful.Response) {
	// TODO：update方法也需要检查循环依赖的问题，可能出现本来没有循环依赖，更新之后出现循环依赖
	// TODO：对不允许修改的信息检查？Update Patch
	// 获取group
	ew := &apis.Group{}
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

	// TODO: 格式校验
	// 格式校验
	//res, err := analyzer.SerializeToJson(ew)
	//_, err = analyzer.Deserialize(res, apis.Group{})
	//if err != nil {
	//	err := response.WriteError(http.StatusBadRequest, err)
	//	if err != nil {
	//		logs.Errorf("failed to return a status code")
	//		return
	//	}
	//	return
	//}

	// TODO：循环依赖检查
	// 简易版的依赖检查，无法检查a1->a2->a3->a1这种
	dependencyErr := util.CheckActionCircularDependency(ew.Spec)
	if dependencyErr != nil {
		err := response.WriteError(http.StatusBadRequest, dependencyErr)
		if err != nil {
			logs.Errorf("failed to return a status code")
			return
		}
		return
	}

	// 获取name
	name := request.QueryParameter(GROUP_NAME)
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

	// TODO: 完善一下没有返回状态部分的是什么情况 ， 暂时500
	// 更新group
	updatedGroup, updateErr := h.manager.UpdateGroup(name, namespace, ew)
	if updateErr != nil {
		logs.Errorf("Update group %s error: %v", name, updateErr)
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
	err = response.WriteHeaderAndEntity(http.StatusOK, updatedGroup)
	if err != nil {
		logs.Errorf("failed to return a status code")
		return
	}

	// 记录日志
	logs.Debugf("update group : %v", name)
}

func (h *GroupHandler) DeleteGroup(request *restful.Request, response *restful.Response) {
	// 获取name
	name := request.QueryParameter(GROUP_NAME)
	if name == "" {
		err := response.WriteError(http.StatusBadRequest, fmt.Errorf("provide group name , the key is Name "))
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

	// 删除group
	var err error
	err = h.manager.DeleteGroup(name, namespace)
	if err != nil {
		logs.Errorf("delete group %s error: %v", name, err)
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
	logs.Debugf("delete group : %v", name)
}

func (h *GroupHandler) PatchGroup(request *restful.Request, response *restful.Response) {
	// 获取 json
	req := &apis.Group{}
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
	name := request.QueryParameter(GROUP_NAME)
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

	// 序列化Patch group
	patchGroup, err := analyzer.SerializeToJson(req)
	if err != nil {
		logs.Errorf("Serialize patch group error: %v", err)
		err := response.WriteError(http.StatusInternalServerError, err)
		if err != nil {
			logs.Errorf("failed to return a status code")
			return
		}
	}

	patchedGroup, patchedErr := h.manager.PatchGroup(name, namespace, []byte(patchGroup))
	if patchedErr != nil {
		logs.Errorf("patched group %s error: %v", name, err)
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
	err = response.WriteHeaderAndEntity(http.StatusOK, patchedGroup)
	if err != nil {
		logs.Errorf("failed to return a status code")
		return
	}

	// 记录日志
	logs.Debugf("patch group : %v", name)
}

func (h *GroupHandler) NewGetWebService() *restful.WebService {
	ws := new(restful.WebService)
	ws.Path(GROUP_PATH).
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/").
		To(h.GetGroup).
		Doc("Get a group with name").
		Metadata(restfulspec.KeyOpenAPITags, []string{TAG}).
		Param(ws.QueryParameter("Name", "The name of the group").DataType("string")).
		Param(ws.QueryParameter("Namespace", "The namespace of the group").DataType("string")).
		Operation("Get group").
		Returns(200, "OK", apis.Group{}).
		Returns(404, "Not Found", nil),
	)

	ws.Route(ws.POST("/").
		To(h.CreateGroup).
		Doc("Create a group with name").
		Metadata(restfulspec.KeyOpenAPITags, []string{TAG}).
		Param(ws.QueryParameter("Namespace", "The namespace of the group").DataType("string")).
		Param(ws.BodyParameter("Group", "The json string of the group object").DataType("string")).
		Operation("Create group").
		Returns(200, "OK", apis.Group{}).
		Returns(404, "Not Found", nil),
	)

	ws.Route(ws.PUT("/").
		To(h.UpdateGroup).
		Doc("Update a group with name").
		Metadata(restfulspec.KeyOpenAPITags, []string{TAG}).
		Param(ws.QueryParameter("Name", "The name of the group").DataType("string")).
		Param(ws.QueryParameter("Namespace", "The namespace of the group").DataType("string")).
		Param(ws.BodyParameter("Group", "The json string of the Group object").DataType("string")).
		Operation("Update group").
		Returns(200, "OK", apis.Group{}).
		Returns(404, "Not Found", nil))

	ws.Route(ws.PATCH("/").
		To(h.PatchGroup).
		Doc("Patch a group").
		Metadata(restfulspec.KeyOpenAPITags, []string{TAG}).
		Param(ws.QueryParameter("Name", "The name of the group").DataType("string")).
		Param(ws.QueryParameter("Namespace", "The namespace of the group").DataType("string")).
		Param(ws.BodyParameter("Group", "The json string of the Group field").DataType("string")).
		Operation("Patch group").
		Returns(200, "OK", apis.Group{}).
		Returns(404, "Not Found", nil))

	ws.Route(ws.DELETE("/").
		To(h.DeleteGroup).
		Doc("Delete a group").
		Metadata(restfulspec.KeyOpenAPITags, []string{TAG}).
		Param(ws.QueryParameter("Name", "The name of the group").DataType("string")).
		Param(ws.QueryParameter("Namespace", "The namespace of the group").DataType("string")).
		Operation("Delete group").
		Returns(200, "OK", apis.Group{}).
		Returns(404, "Not Found", nil),
	)

	return ws
}

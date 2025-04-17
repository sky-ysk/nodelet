package group

import (
	"context"
	"fmt"
	restfulspec "github.com/emicklei/go-restful-openapi/v2"
	"github.com/emicklei/go-restful/v3"
	"hit.edu/framework/pkg/apimachinery/types"
	apis "hit.edu/framework/pkg/apis/cores"
	metav1 "hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/client-go/clients"
	"hit.edu/framework/pkg/client-go/clients/typed/core"
	"hit.edu/framework/pkg/component-base/analyzer"
	"hit.edu/framework/pkg/component-base/logs"
	"net/http"
	"sync"
)

type GroupHandler struct {
	clients   map[string]core.GroupInterface
	clientSet *clients.ClientSet
	mu        sync.Mutex
}

type CurrentGroupHandler struct {
	client core.GroupInterface
}

var _ Handler = &GroupHandler{}

//func NewGroupHandler(clientSet *clients.ClientSet) *GroupHandler {
//	c := clientSet.Core().Groups("test") //apis.NamespaceAll
//	return &GroupHandler{
//		client: c,
//	}
//}

// NewGroupHandler 创建一个 GroupHandler
func NewGroupHandler(clientSet *clients.ClientSet) *GroupHandler {
	return &GroupHandler{
		clients:   make(map[string]core.GroupInterface),
		clientSet: clientSet,
	}
}

// GetClient 根据 namespace 获取 client，如果不存在则创建
func (h *GroupHandler) GetClient(namespace string) *CurrentGroupHandler {
	h.mu.Lock()
	defer h.mu.Unlock()

	// 如果已经存在，直接返回
	if c, exists := h.clients[namespace]; exists {
		return &CurrentGroupHandler{
			client: c,
		}
	}

	// 否则创建新的 client
	newClient := h.clientSet.Core().Groups(namespace)
	h.clients[namespace] = newClient
	return &CurrentGroupHandler{
		client: newClient,
	}
}
func (h *GroupHandler) GetGroup(request *restful.Request, response *restful.Response) {
	// 尝试从url中获取参数
	c := &CurrentGroupHandler{}
	name := request.QueryParameter(GROUP_NAME)
	if name == "" {
		// url中没有获取到name参数，尝试从请求体中获取
		req := &apis.Group{}
		err := request.ReadEntity(&req)
		if err != nil || req.Name == "" {
			err := response.WriteError(http.StatusBadRequest, fmt.Errorf("provide group name , the key is Name "))
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
		c = h.GetClient(namespace)
	}

	result, err := c.client.Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		logs.Errorf("Get group %s error: %v , group not exist !", name, err)
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
		logs.Debugf("Get group")
	}
}

func (h *GroupHandler) CreateGroup(request *restful.Request, response *restful.Response) {
	// 先查询Group是否存在
	// 尝试从url中获取参数
	c := &CurrentGroupHandler{}
	name := request.QueryParameter(GROUP_NAME)
	ew := &apis.Group{}
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
				err := response.WriteError(http.StatusBadRequest, fmt.Errorf("provide group name , the key is Name "))
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

	namespace := ew.Namespace
	if namespace == "" {
		logs.Error("namespace is empty")
		err := response.WriteError(http.StatusBadRequest, fmt.Errorf("namespace is empty"))
		if err != nil {
			logs.Errorf("failed to return a status code ")
			return
		}
		return
	} else {
		c = h.GetClient(namespace)
	}

	result, err := c.client.Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		logs.Infof("Get group %s error: %v , group not exist! creat it", name, err)
	} else if result.Name == name {
		logs.Errorf("Create group %s error, group existed: %v", name, result)
		err = fmt.Errorf("create group %s error, group existed: %v", name, result)
		err := response.WriteError(http.StatusConflict, err)
		if err != nil {
			logs.Errorf("failed to return a status code")
			return
		}
		return
	}

	// 不存在，解析用户的输入
	//ew := &apis.Group{}
	//err = request.ReadEntity(ew)
	//if err != nil {
	//	logs.Errorf("Failed to create group %s, error: %v", name, err)
	//	err := response.WriteError(http.StatusInternalServerError, err)
	//	if err != nil {
	//		logs.Errorf("failed to return a status code")
	//		return
	//	}
	//	return
	//}

	logs.Info(*ew)

	//格式校验
	res, err := analyzer.SerializeToJson(ew)
	_, err = analyzer.Deserialize(res, apis.Group{})
	if err != nil {
		err := response.WriteError(http.StatusBadRequest, err)
		if err != nil {
			logs.Errorf("failed to return a status code")
			return
		}
		// return
	}

	// TODO: 为Workflow分配ID

	// 将Workflow写入数据库中
	result, err = c.client.Create(context.TODO(), ew, metav1.CreateOptions{})
	if err != nil {
		err1 := response.WriteError(http.StatusInternalServerError, err)
		if err1 != nil {
			logs.Errorf("failed to return a status code ,error %v", err1)
			return
		}
		logs.Errorf("Create group %s ,failed write to database , error: %v", name, err)
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
	logs.Debugf("Create group %v", result)
}

func (h *GroupHandler) UpdateGroup(request *restful.Request, response *restful.Response) {
	// 先检查group是否存在
	// 存在：更新
	// 不存在：返回错误
	// 尝试从url中获取参数
	c := &CurrentGroupHandler{}
	name := request.QueryParameter(GROUP_NAME)
	ew := &apis.Group{}
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
				err := response.WriteError(http.StatusBadRequest, fmt.Errorf("provide group name , the key is Name "))
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

	// 检查group是否存在
	group, err := c.client.Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		logs.Errorf("Get group %s error: %v , group not exist! ", name, err)
		err := response.WriteError(http.StatusNotFound, err)
		if err != nil {
			logs.Errorf("failed to return a status code")
			return
		}
		return
	}

	// group 存在，更新
	if group.Name == name {
		//err := request.ReadEntity(&group)
		//if err != nil {
		//	err := response.WriteError(http.StatusInternalServerError, err)
		//	if err != nil {
		//		logs.Errorf("failed to return a status code")
		//		return
		//	}
		//	return
		//}

		// 格式验证
		res, err := analyzer.SerializeToJson(group)
		_, err = analyzer.Deserialize(res, apis.Group{})
		if err != nil {
			err := response.WriteError(http.StatusBadRequest, err)
			if err != nil {
				logs.Errorf("failed to return a status code ")
				return
			}
			// return
		}

		updatedGroup, updateErr := c.client.Update(context.TODO(), ew, metav1.UpdateOptions{})
		if updateErr != nil {
			logs.Errorf("Update group %s error: %v", name, updateErr)
			err := response.WriteError(http.StatusInternalServerError, err)
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
}

func (h *GroupHandler) DeleteGroup(request *restful.Request, response *restful.Response) {
	// 查看group是否存在
	// 存在，删除节点
	// 不存在，返回 404 not found
	// 尝试从url中获取参数
	c := &CurrentGroupHandler{}
	name := request.QueryParameter(GROUP_NAME)
	if name == "" {
		// url中没有获取到name参数，尝试从请求体中获取
		req := &apis.Group{}
		err := request.ReadEntity(&req)
		if err != nil || req.Name == "" {
			err := response.WriteError(http.StatusBadRequest, fmt.Errorf("provide group name , the key is Name "))
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

	// 查看group是否存在
	group, err := c.client.Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		logs.Errorf("Get group %s error: %v , group not exist! ", name, err)
		err := response.WriteError(http.StatusNotFound, err)
		if err != nil {
			logs.Errorf("failed to return a status code")
			return
		}
	}

	// group 存在
	if group.Name == name {
		err := c.client.Delete(context.TODO(), name, metav1.DeleteOptions{})
		if err != nil {
			logs.Errorf("Delete group %s error: %v", name, err)
			err := response.WriteError(http.StatusInternalServerError, err)
			if err != nil {
				logs.Errorf("failed to return a status code")
				return
			}
		}

		// 返回停止成功的状态
		response.WriteHeader(http.StatusOK)

		// 记录日志
		logs.Debugf("delete group : %v", name)

	}
}

func (h *GroupHandler) PatchGroup(request *restful.Request, response *restful.Response) {
	// 先检查group是否存在
	// 存在：部分更新
	// 不存在：返回错误
	// 尝试从url中获取参数
	c := &CurrentGroupHandler{}
	name := request.QueryParameter(GROUP_NAME)
	if name == "" {
		// url中没有获取到name参数，尝试从请求体中获取
		req := &apis.Group{}
		err := request.ReadEntity(&req)
		if err != nil || req.Name == "" {
			err := response.WriteError(http.StatusBadRequest, fmt.Errorf("provide group name , the key is Name "))
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

	// 检查group是否存在
	group, err := c.client.Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		logs.Errorf("Get group %s error: %v , group not exist !", name, err)
		err := response.WriteError(http.StatusNotFound, err)
		if err != nil {
			logs.Errorf("failed to return a status code")
			return
		}
		return
	}

	// group 存在，部分更新
	if group.Name == name {
		err := request.ReadEntity(group)
		if err != nil {
			err := response.WriteError(http.StatusInternalServerError, err)
			if err != nil {
				logs.Errorf("failed to return a status code")
				return
			}
			return
		}
		patchGroup, err := analyzer.SerializeToJson(group)
		if err != nil {
			logs.Errorf("Serialize patch group error: %v", err)
			err := response.WriteError(http.StatusInternalServerError, err)
			if err != nil {
				logs.Errorf("failed to return a status code")
				return
			}
		}
		patchedGroup, err := c.client.Patch(context.TODO(), name, types.StrategicMergePatchType, []byte(patchGroup), metav1.PatchOptions{})
		if err != nil {
			logs.Errorf("Patch group %s error: %v", name, err)
			err := response.WriteError(http.StatusInternalServerError, err)
			if err != nil {
				logs.Errorf("failed to return a status code")
				return
			}
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
		Returns(400, "Not Found", nil),
	)

	ws.Route(ws.POST("/").
		To(h.CreateGroup).
		Doc("Create a group with name").
		Metadata(restfulspec.KeyOpenAPITags, []string{TAG}).
		Param(ws.QueryParameter("Name", "The name of the group").DataType("string")).
		Param(ws.QueryParameter("Namespace", "The namespace of the group").DataType("string")).
		Param(ws.BodyParameter("Group", "The json string of the group object").DataType("string")).
		Operation("Create group").
		Returns(200, "OK", apis.Group{}).
		Returns(400, "Not Found", nil),
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
		Returns(400, "Not Found", nil))

	ws.Route(ws.PATCH("/").
		To(h.PatchGroup).
		Doc("Patch a group").
		Metadata(restfulspec.KeyOpenAPITags, []string{TAG}).
		Param(ws.QueryParameter("Name", "The name of the group").DataType("string")).
		Param(ws.QueryParameter("Namespace", "The namespace of the group").DataType("string")).
		Param(ws.BodyParameter("Group", "The json string of the Group field").DataType("string")).
		Operation("Patch group").
		Returns(200, "OK", apis.Group{}).
		Returns(400, "Not Found", nil))

	ws.Route(ws.DELETE("/").
		To(h.DeleteGroup).
		Doc("Delete a group").
		Metadata(restfulspec.KeyOpenAPITags, []string{TAG}).
		Param(ws.QueryParameter("Name", "The name of the group").DataType("string")).
		Param(ws.QueryParameter("Namespace", "The namespace of the group").DataType("string")).
		Operation("Delete group").
		Returns(200, "OK", apis.Group{}).
		Returns(400, "Not Found", nil),
	)

	return ws
}

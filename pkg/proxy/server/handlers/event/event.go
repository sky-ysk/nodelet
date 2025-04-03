package event

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

type EventHandler struct {
	clients   map[string]core.EventInterface
	clientSet *clients.ClientSet
	mu        sync.Mutex
}

type CurrentEventHandler struct {
	client core.EventInterface
}

var _ Handler = &EventHandler{}

//func NewEventHandler(clientSet *clients.ClientSet) *EventHandler {
//	c := clientSet.Core().Events("test") // apis.NamespaceAll
//	return &EventHandler{
//		client: c,
//	}
//}

// NewEventHandler 创建一个 EventHandler
func NewEventHandler(clientSet *clients.ClientSet) *EventHandler {
	return &EventHandler{
		clients:   make(map[string]core.EventInterface),
		clientSet: clientSet,
	}
}

// GetClient 根据 namespace 获取 client，如果不存在则创建
func (h *EventHandler) GetClient(namespace string) *CurrentEventHandler {
	h.mu.Lock()
	defer h.mu.Unlock()

	// 如果已经存在，直接返回
	if c, exists := h.clients[namespace]; exists {
		return &CurrentEventHandler{
			client: c,
		}
	}

	// 否则创建新的 client
	newClient := h.clientSet.Core().Events(namespace)
	h.clients[namespace] = newClient
	return &CurrentEventHandler{
		client: newClient,
	}
}

func (h *EventHandler) GetEvent(request *restful.Request, response *restful.Response) {
	// 尝试从url中获取参数
	c := &CurrentEventHandler{}
	name := request.QueryParameter(EVENT_NAME)
	if name == "" {
		// url中没有获取到name参数，尝试从请求体中获取
		req := &apis.Event{}
		err := request.ReadEntity(&req)
		if err != nil || req.Name == "" {
			err := response.WriteError(http.StatusBadRequest, fmt.Errorf("provide event name , the key is Name "))
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
	namespace := request.PathParameter(NAMESPACE)
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
		logs.Errorf("Get event %s error: %v , event not exist! ", name, err)
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
		logs.Debugf("Get event")
	}
}

func (h *EventHandler) CreateEvent(request *restful.Request, response *restful.Response) {
	// 先查询event是否存在
	// 尝试从url中获取参数
	c := &CurrentEventHandler{}
	name := request.QueryParameter(EVENT_NAME)
	ew := &apis.Event{}
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
				err := response.WriteError(http.StatusBadRequest, fmt.Errorf("provide event name , the key is Name "))
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
		logs.Infof("Get event %s error: %v , event not exist! create it", name, err)
	} else if result.Name == name {
		logs.Errorf("Create event %s error, event existed: %v", name, result)
		err = fmt.Errorf("create event %s error, event existed: %v", name, result)
		err := response.WriteError(http.StatusConflict, err)
		if err != nil {
			logs.Errorf("failed to return a status code")
			return
		}
		return
	}

	// 不存在，解析用户的输入
	//ew := &apis.Event{}
	//err = request.ReadEntity(ew)
	//if err != nil {
	//	logs.Errorf("Failed to create event %s, error: %v", name, err)
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
	_, err = analyzer.Deserialize(res, apis.Event{})
	if err != nil {
		err := response.WriteError(http.StatusBadRequest, err)
		if err != nil {
			logs.Errorf("failed to return a status code")
			return
		}
		// return
	}

	// TODO：为Event分配ID

	// 将event写入数据库中
	result, err = c.client.Create(context.TODO(), ew, metav1.CreateOptions{})
	if err != nil {
		err1 := response.WriteError(http.StatusInternalServerError, err)
		if err1 != nil {
			logs.Errorf("failed to return a status code ,error %v", err1)
			return
		}
		logs.Errorf("Create event %s ,failed write to database , error: %v", name, err)
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

	logs.Debugf("Create event %v", result)
}

func (h *EventHandler) UpdateEvent(request *restful.Request, response *restful.Response) {
	// 先检查event是否存在
	// 存在：更新
	// 不存在：返回错误
	// 尝试从url中获取参数
	c := &CurrentEventHandler{}
	name := request.QueryParameter(EVENT_NAME)
	ew := &apis.Event{}
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
				err := response.WriteError(http.StatusBadRequest, fmt.Errorf("provide event name , the key is Name "))
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

	namespace := request.PathParameter(NAMESPACE)
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

	// 检查event是否存在
	event, err := c.client.Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		logs.Errorf("Get event %s error: %v , event not exist !", name, err)
		err := response.WriteError(http.StatusNotFound, err)
		if err != nil {
			logs.Errorf("failed to return a status code")
			return
		}
		return
	}

	// event 存在，更新
	if event.Name == name {
		//err := request.ReadEntity(&event)
		//if err != nil {
		//	err := response.WriteError(http.StatusInternalServerError, err)
		//	if err != nil {
		//		logs.Errorf("failed to return a status code")
		//		return
		//	}
		//	return
		//}
		//
		//logs.Debugf("Update event to : %v", event)

		// 格式验证
		res, err := analyzer.SerializeToJson(event)
		_, err = analyzer.Deserialize(res, apis.Event{})
		if err != nil {
			err := response.WriteError(http.StatusBadRequest, err)
			if err != nil {
				logs.Errorf("failed to return a status code ")
				return
			}
			// return
		}

		updatedEvent, updateErr := c.client.Update(context.TODO(), ew, metav1.UpdateOptions{})
		if updateErr != nil {
			logs.Errorf("Update event %s error: %v", name, updateErr)
			err := response.WriteError(http.StatusInternalServerError, err)
			if err != nil {
				logs.Errorf("failed to return a status code")
				return
			}
		}

		// 返回成功修改的通知
		err = response.WriteHeaderAndEntity(http.StatusOK, updatedEvent)
		if err != nil {
			logs.Errorf("failed to return a status code")
			return
		}

		// 记录日志
		logs.Debugf("update event : %v", name)

	}
}

func (h *EventHandler) DeleteEvent(request *restful.Request, response *restful.Response) {
	// 查看event是否存在
	// 存在，删除节点
	// 不存在，返回 404 not found
	// 尝试从url中获取参数
	c := &CurrentEventHandler{}
	name := request.QueryParameter(EVENT_NAME)
	if name == "" {
		// url中没有获取到name参数，尝试从请求体中获取
		req := &apis.Event{}
		err := request.ReadEntity(&req)
		if err != nil || req.Name == "" {
			err := response.WriteError(http.StatusBadRequest, fmt.Errorf("provide event name , the key is Name "))
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
	namespace := request.PathParameter(NAMESPACE)
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

	// 查看event是否存在
	event, err := c.client.Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		logs.Errorf("Get event %s error: %v , event not exist !", name, err)
		err := response.WriteError(http.StatusNotFound, err)
		if err != nil {
			logs.Errorf("failed to return a status code")
			return
		}
	}

	// event 存在
	if event.Name == name {
		err := c.client.Delete(context.TODO(), name, metav1.DeleteOptions{})
		if err != nil {
			logs.Errorf("Delete event %s error: %v", name, err)
			err := response.WriteError(http.StatusInternalServerError, err)
			if err != nil {
				logs.Errorf("failed to return a status code")
				return
			}
		}

		// 返回停止成功的状态
		response.WriteHeader(http.StatusOK)

		// 记录日志
		logs.Debugf("delete event : %v", name)

	}
}

func (h *EventHandler) PatchEvent(request *restful.Request, response *restful.Response) {
	// 先检查event是否存在
	// 存在：部分更新
	// 不存在：返回错误
	// 尝试从url中获取参数
	c := &CurrentEventHandler{}
	name := request.QueryParameter(EVENT_NAME)
	if name == "" {
		// url中没有获取到name参数，尝试从请求体中获取
		req := &apis.Event{}
		err := request.ReadEntity(&req)
		if err != nil || req.Name == "" {
			err := response.WriteError(http.StatusBadRequest, fmt.Errorf("provide event name , the key is Name "))
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
	namespace := request.PathParameter(NAMESPACE)
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

	// 检查event是否存在
	event, err := c.client.Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		logs.Errorf("Get event %s error: %v , event not exist !", name, err)
		err := response.WriteError(http.StatusNotFound, err)
		if err != nil {
			logs.Errorf("failed to return a status code")
			return
		}
		return
	}

	// event 存在，部分更新
	if event.Name == name {
		err := request.ReadEntity(event)
		if err != nil {
			err := response.WriteError(http.StatusInternalServerError, err)
			if err != nil {
				logs.Errorf("failed to return a status code")
				return
			}
			return
		}

		patchEvent, err := analyzer.SerializeToJson(event)
		if err != nil {
			logs.Errorf("Serialize patch event error: %v", err)
			err := response.WriteError(http.StatusInternalServerError, err)
			if err != nil {
				logs.Errorf("failed to return a status code")
				return
			}
		}
		patchedEvent, err := c.client.Patch(context.TODO(), name, types.StrategicMergePatchType, []byte(patchEvent), metav1.PatchOptions{})
		if err != nil {
			logs.Errorf("Patch event %s error: %v", name, err)
			err := response.WriteError(http.StatusInternalServerError, err)
			if err != nil {
				logs.Errorf("failed to return a status code")
				return
			}
		}

		// 返回成功修改的通知
		err = response.WriteHeaderAndEntity(http.StatusOK, patchedEvent)
		if err != nil {
			logs.Errorf("failed to return a status code")
			return
		}

		// 记录日志
		logs.Debugf("patch event : %v", name)

	}
}

func (h *EventHandler) NewGetWebService() *restful.WebService {
	ws := new(restful.WebService)
	ws.Path(EVENT_PATH).
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/{Namespace}/event").
		To(h.GetEvent).
		Doc("Get a event with name").
		Metadata(restfulspec.KeyOpenAPITags, []string{TAG}).
		Param(ws.QueryParameter("Name", "The name of the event").DataType("string")).
		Param(ws.PathParameter("Namespace", "The namespace of the event").DataType("string")).
		Operation("Get event").
		Returns(200, "OK", apis.Event{}).
		Returns(400, "Not Found", nil),
	)

	ws.Route(ws.POST("/{Namespace}/event").
		To(h.CreateEvent).
		Doc("Create a event").
		Metadata(restfulspec.KeyOpenAPITags, []string{TAG}).
		Param(ws.QueryParameter("Name", "The name of the event").DataType("string")).
		Param(ws.PathParameter("Namespace", "The namespace of the event").DataType("string")).
		Param(ws.BodyParameter("Event", "The json string of the Event object").DataType("string")).
		Operation("Create event").
		Returns(200, "OK", apis.Event{}).
		Returns(400, "Not Found", nil),
	)

	ws.Route(ws.PUT("/{Namespace}/event").
		To(h.UpdateEvent).
		Doc("Update a event").
		Metadata(restfulspec.KeyOpenAPITags, []string{TAG}).
		Param(ws.QueryParameter("Name", "The name of the event").DataType("string")).
		Param(ws.PathParameter("Namespace", "The namespace of the event").DataType("string")).
		Param(ws.BodyParameter("Event", "The json string of the Event object").DataType("string")).
		Operation("Update event").
		Returns(200, "OK", apis.Event{}).
		Returns(400, "Not Found", nil))

	ws.Route(ws.PATCH("/{Namespace}/event").
		To(h.PatchEvent).
		Doc("Patch a event").
		Metadata(restfulspec.KeyOpenAPITags, []string{TAG}).
		Param(ws.QueryParameter("Name", "The name of the event").DataType("string")).
		Param(ws.PathParameter("Namespace", "The namespace of the event").DataType("string")).
		Param(ws.BodyParameter("Event", "The json string of the Event field").DataType("string")).
		Operation("Patch event").
		Returns(200, "OK", apis.Event{}).
		Returns(400, "Not Found", nil))

	ws.Route(ws.DELETE("/{Namespace}/event").
		To(h.DeleteEvent).
		Doc("Delete a event").
		Metadata(restfulspec.KeyOpenAPITags, []string{TAG}).
		Param(ws.QueryParameter("Name", "The name of the event").DataType("string")).
		Param(ws.PathParameter("Namespace", "The namespace of the event").DataType("string")).
		Operation("Delete event").
		Returns(200, "OK", apis.Event{}).
		Returns(400, "Not Found", nil))

	return ws
}

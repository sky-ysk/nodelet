package event

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

type EventsHandler struct {
	clientSet *clients.ClientSet
	manager   *manager.Manager
	mu        sync.Mutex
}

//type CurrentEventsHandler struct {
//	client core.EventInterface
//}

var _ Handler = &EventsHandler{}

//func NewEventsHandler(clientSet *clients.ClientSet) *EventsHandler {
//	c := clientSet.Core().Events("test") //apis.NamespaceAll
//	return &EventsHandler{
//		client: c,
//	}
//}

// NewEventHandler 创建一个 EventHandler
func NewEventsHandler(clientSet *clients.ClientSet) *EventsHandler {
	return &EventsHandler{
		manager:   manager.NewManager(clientSet),
		clientSet: clientSet,
	}
}

// GetClient 根据 namespace 获取 client，如果不存在则创建
//func (h *EventsHandler) GetClient(namespace string) *CurrentEventsHandler {
//	h.mu.Lock()
//	defer h.mu.Unlock()
//
//	// 如果已经存在，直接返回
//	if c, exists := h.clients[namespace]; exists {
//		return &CurrentEventsHandler{
//			client: c,
//		}
//	}
//
//	// 否则创建新的 client
//	newClient := h.clientSet.Core().Events(namespace)
//	h.clients[namespace] = newClient
//	return &CurrentEventsHandler{
//		client: newClient,
//	}
//}

func (h *EventsHandler) GetEvents(request *restful.Request, response *restful.Response) {
	//c := &CurrentEventsHandler{}
	//
	//// 从url中获取namespace
	//namespace := request.QueryParameter(NAME_SPACE)
	////if namespace == "" {
	////	err := response.WriteError(http.StatusBadRequest, fmt.Errorf("namespace is required"))
	////	if err != nil {
	////		logs.Errorf("failed to return a status code ")
	////		return
	////	}
	////	return
	////} else {
	////	c = h.GetClient(namespace)
	////}
	//
	//c = h.GetClient(namespace)
	//results, err := c.client.List(context.TODO(), metav1.ListOptions{})
	//if err != nil {
	//	logs.Errorf("Get events failed: %v", err)
	//	err := response.WriteError(http.StatusInternalServerError, err)
	//	if err != nil {
	//		logs.Errorf("failed to return a status code")
	//		return
	//	}
	//}
	//
	//err = response.WriteEntity(results)
	//if err != nil {
	//	err := response.WriteError(http.StatusInternalServerError, err)
	//	if err != nil {
	//		logs.Errorf("failed to return a status code")
	//		return
	//	}
	//}
	//logs.Debugf("Get events")

	// 获取namespace
	// namespace := request.QueryParameter(NAME_SPACE)

	// 获取name
	name := request.QueryParameter(EVENT_NAME)
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

	var results *apis.EventList
	var err error

	results, err = h.manager.GetEvents(name, namespace)
	if err != nil {
		logs.Errorf("Get events failed: %v", err)
		err := response.WriteError(http.StatusInternalServerError, err)
		if err != nil {
			logs.Errorf("failed to return a status code")
			return
		}
	}

	err = response.WriteHeaderAndEntity(http.StatusOK, results)
	if err != nil {
		logs.Errorf("failed to return a status code")
		return
	}

	logs.Debugf("Get events ")
}

// TODO: DeleteAll

func (h *EventsHandler) NewGetWebService() *restful.WebService {
	ws := new(restful.WebService)
	ws.Path(EVENTS_PATH).
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/").
		Doc("Get all events").
		Metadata(restfulspec.KeyOpenAPITags, []string{TAG}).
		Param(ws.QueryParameter("Name", "The name of the involved object").DataType("string")).
		Param(ws.QueryParameter("Namespace", "The namespace of the events").DataType("string")).
		To(h.GetEvents).
		Operation("Get events").
		Returns(200, "OK", []apis.Event{}).
		Returns(404, "Not Found", nil),
	)

	return ws
}

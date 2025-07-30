package event

import (
	restfulspec "github.com/emicklei/go-restful-openapi/v2"
	"github.com/emicklei/go-restful/v3"
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/client-go/clients"
	"hit.edu/framework/pkg/client-go/util/manager"
	"hit.edu/framework/pkg/component-base/logs"
	"net/http"
	"strconv"
	"sync"
)

type EventsHandler struct {
	clientSet *clients.ClientSet
	manager   *manager.Manager
	mu        sync.Mutex
}

var _ Handler = &EventsHandler{}

// NewEventsHandler 创建一个 EventHandler
func NewEventsHandler(clientSet *clients.ClientSet) *EventsHandler {
	return &EventsHandler{
		manager:   manager.NewManager(clientSet),
		clientSet: clientSet,
	}
}

func (h *EventsHandler) GetEvents(request *restful.Request, response *restful.Response) {
	// 获取selectName
	//selectName := request.QueryParameter(SELECT_NAME)
	//if selectName == "" {
	//	err := response.WriteError(http.StatusBadRequest, fmt.Errorf("name is empty"))
	//	if err != nil {
	//		logs.Errorf("failed to return a status code ")
	//		return
	//	}
	//	return
	//}

	// 获取namespace
	namespace := request.QueryParameter(NAME_SPACE)

	// 分页
	pageStr := request.QueryParameter(PAGE)
	pageSizeStr := request.QueryParameter(PAGE_SIZE)

	page := 0
	pageSize := 10

	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil {
			if p <= 0 {
				err := response.WriteErrorString(400, "Invalid Page parameter")
				if err != nil {
					logs.Error(err.Error())
					return
				}
				return
			}
			page = p
		} else {
			// 可选：返回错误或使用默认值
			err := response.WriteErrorString(400, "Invalid Page parameter")
			if err != nil {
				logs.Errorf("failed to return a status code")
			}
			return
		}
	}

	if pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil {
			if ps < 0 {
				err := response.WriteErrorString(400, "Invalid Page parameter")
				if err != nil {
					logs.Errorf("failed to return a status code")
					return
				}
				return
			}
			pageSize = ps
		} else {
			err := response.WriteErrorString(400, "Invalid PageSize parameter")
			if err != nil {
				logs.Errorf("failed to return a status code")
			}
			return
		}
	}

	var results *apis.EventList
	var err error

	results, err = h.manager.GetEvents(namespace)
	if err != nil {
		logs.Errorf("Get events failed: %v", err)
		err := response.WriteError(http.StatusInternalServerError, err)
		if err != nil {
			logs.Errorf("failed to return a status code")
			return
		}
	}

	if (len(results.Items) <= pageSize && page == 1) || page == 0 {
		err = response.WriteHeaderAndEntity(http.StatusOK, results)
		if err != nil {
			logs.Errorf("failed to return a status code")
			return
		}
		logs.Debugf("Get events ")
	} else {
		if (page-1)*pageSize < len(results.Items) && (page*pageSize)-1 < len(results.Items) {
			// 返回一整页
			tmpRes := results
			tmpRes.Items = results.Items[(page-1)*pageSize : page*pageSize]
			err = response.WriteHeaderAndEntity(http.StatusOK, tmpRes)
			if err != nil {
				logs.Errorf("failed to return a status code")
				return
			}
			logs.Debugf("Get events ")
		} else if (page-1)*pageSize < len(results.Items) {
			// 返回开头到最后
			tmpRes := results
			tmpRes.Items = results.Items[(page-1)*pageSize:]
			err = response.WriteHeaderAndEntity(http.StatusOK, tmpRes)
			if err != nil {
				logs.Errorf("failed to return a status code")
				return
			}
			logs.Debugf("Get events ")
		} else {
			// 页数太大，没有这么多event数据
			err := response.WriteErrorString(400, "该页没有event数据")
			if err != nil {
				logs.Errorf("failed to return a status code")
				return
			}
			return
		}
	}

}

// TODO: DeleteAll
// TODO: 检查满足功能 ， 删除命名空间下指定条件所有的event，未指定命名空间，删除所有符合条件的event

// TODO: 确认这里的事件应该也是删除某一任务相关的所有事件

func (h *EventsHandler) DeleteEvents(request *restful.Request, response *restful.Response) {
	namespace := request.QueryParameter(NAME_SPACE)

	fieldSelector := request.QueryParameter(SELECT_NAME)

	logs.Debugf("delete events namespace: %s, fieldSelector: %s", namespace, fieldSelector)

	err := h.manager.DeleteEvents(namespace, fieldSelector)
	if err != nil {
		logs.Errorf("Delete events failed: %v", err)
		err := response.WriteError(http.StatusInternalServerError, err)
		if err != nil {
			logs.Errorf("failed to return a status code ")
			return
		}
		return
	}

	response.WriteHeader(http.StatusOK)
	logs.Debugf("Delete events ")
}

func (h *EventsHandler) NewGetWebService() *restful.WebService {
	ws := new(restful.WebService)
	ws.Path(EVENTS_PATH).
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/").
		Doc("Get all events with selector").
		Metadata(restfulspec.KeyOpenAPITags, []string{TAG}).
		Param(ws.QueryParameter("Page", "The page of the events(optional)").DataType("int")).
		Param(ws.QueryParameter("PageSize", "The size of the page(optional Default=10)").DataType("int")).
		// Param(ws.QueryParameter("SelectorName", "The name of the involved object").DataType("string")).
		Param(ws.QueryParameter("Namespace", "The namespace of the events").DataType("string")).
		To(h.GetEvents).
		Operation("Get events").
		Returns(200, "OK", []apis.Event{}).
		Returns(404, "Not Found", nil),
	)

	ws.Route(ws.DELETE("/").
		Doc("Delete all events with selector").
		Metadata(restfulspec.KeyOpenAPITags, []string{TAG}).
		Param(ws.QueryParameter("SelectorName", "Selector（optional）").DataType("string")).
		Param(ws.QueryParameter("Namespace", "The namespace of the events").DataType("string")).
		To(h.DeleteEvents).
		Operation("Delete events").
		Returns(200, "OK", []apis.Event{}).
		Returns(404, "Not Found", nil),
	)

	return ws
}

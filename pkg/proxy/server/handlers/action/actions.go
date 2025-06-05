package action

import (
	restfulspec "github.com/emicklei/go-restful-openapi/v2"
	"github.com/emicklei/go-restful/v3"
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/client-go/clients"
	"hit.edu/framework/pkg/client-go/util/manager"
	"hit.edu/framework/pkg/component-base/logs"
	"net/http"
	"sync"
)

type ActionsHandler struct {
	clientSet *clients.ClientSet
	manager   *manager.Manager
	mu        sync.Mutex
}

var _ Handler = &ActionsHandler{}

// NewActionsHandler 创建一个 ActionHandler
func NewActionsHandler(clientSet *clients.ClientSet) *ActionsHandler {
	return &ActionsHandler{
		manager:   manager.NewManager(clientSet),
		clientSet: clientSet,
	}
}

// TODO：删除所有的Action，同一命名空间下所有，不提供命名空间默认全部action

func (h *ActionsHandler) GetActions(request *restful.Request, response *restful.Response) {
	// 获取namespace
	namespace := request.QueryParameter(NAME_SPACE)

	var results *apis.ActionList
	var err error

	labels := request.QueryParameter("Label")
	if labels == "" {
		results, err = h.manager.GetActions(namespace)
		if err != nil {
			logs.Errorf("Get actions failed: %v", err)
			err := response.WriteError(http.StatusInternalServerError, err)
			if err != nil {
				logs.Errorf("failed to return a status code")
				return
			}
		}
	} else {
		results, err = h.manager.FilterActions(namespace, labels)
		if err != nil {
			logs.Errorf("Get actions with labels failed: %v", err)
			err := response.WriteError(http.StatusInternalServerError, err)
			if err != nil {
				logs.Errorf("failed to return a status code")
				return
			}
		}
	}

	//err = response.WriteEntity(results)
	//if err != nil {
	//	err := response.WriteError(http.StatusInternalServerError, err)
	//	if err != nil {
	//		logs.Errorf("failed to return a status code")
	//		return
	//	}
	//}
	//
	//response.WriteHeader(http.StatusOK)

	err = response.WriteHeaderAndEntity(http.StatusOK, results)
	if err != nil {
		logs.Errorf("failed to return a status code")
		return
	}

	logs.Debugf("Get actions ")
}

func (h *ActionsHandler) NewGetWebService() *restful.WebService {
	ws := new(restful.WebService)
	ws.Path(ACTIONS_PATH).
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/").
		Doc("Get all actions (with selector)").
		Metadata(restfulspec.KeyOpenAPITags, []string{TAG}).
		Param(ws.QueryParameter("Label", "Labels of the action (optional)").DataType("string")).
		Param(ws.QueryParameter("Namespace", "The namespace of the action").DataType("string")).
		To(h.GetActions).
		Operation("Get actions").
		Returns(200, "OK", []apis.Action{}).
		Returns(404, "Not Found", nil),
	)

	return ws
}

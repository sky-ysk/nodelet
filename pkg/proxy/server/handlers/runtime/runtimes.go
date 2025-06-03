package runtime

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

type RuntimesHandler struct {
	clientSet *clients.ClientSet
	manager   *manager.Manager
	mu        sync.Mutex
}

var _ Handler = &RuntimesHandler{}

// NewRuntimesHandler 创建一个 RuntimesHandler
func NewRuntimesHandler(clientSet *clients.ClientSet) *RuntimesHandler {
	return &RuntimesHandler{
		manager:   manager.NewManager(clientSet),
		clientSet: clientSet,
	}
}

func (h *RuntimesHandler) GetRuntimes(request *restful.Request, response *restful.Response) {
	// 从url中获取namespace
	namespace := request.QueryParameter(NAME_SPACE)
	var results *apis.RuntimeList
	var err error
	if namespace == "" {
		//err := response.WriteError(http.StatusBadRequest, fmt.Errorf("namespace is required"))
		//if err != nil {
		//	logs.Errorf("failed to return a status code ")
		//	return
		//}
		//return
		results, err = h.manager.GetRuntimes(namespace)
		if err != nil {
			logs.Errorf("Get runtimes failed: %v", err)
			err := response.WriteError(http.StatusInternalServerError, err)
			if err != nil {
				logs.Errorf("failed to return a status code")
				return
			}
		}
	}

	labels := request.QueryParameter("Label")
	//var results *apis.RuntimeList
	//var err error
	if labels == "" {
		results, err = h.manager.GetRuntimes(namespace)
		if err != nil {
			logs.Errorf("Get runtimes failed: %v", err)
			err := response.WriteError(http.StatusInternalServerError, err)
			if err != nil {
				logs.Errorf("failed to return a status code")
				return
			}
		}
	} else {
		results, err = h.manager.FilterRuntimes(namespace, labels)
		if err != nil {
			logs.Errorf("Get runtimes failed: %v", err)
			err := response.WriteError(http.StatusInternalServerError, err)
			if err != nil {
				logs.Errorf("failed to return a status code")
				return
			}
		}
	}

	err = response.WriteEntity(results)
	if err != nil {
		err := response.WriteError(http.StatusInternalServerError, err)
		if err != nil {
			logs.Errorf("failed to return a status code")
			return
		}
	}
	logs.Debugf("Get runtimes")
}

func (h *RuntimesHandler) NewGetWebService() *restful.WebService {
	ws := new(restful.WebService)
	ws.Path(RUNTIMES_PATH).
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/").
		Doc("Get all runtimes").
		Metadata(restfulspec.KeyOpenAPITags, []string{TAG}).
		Param(ws.QueryParameter("Label", "Labels of the runtimes (optional)").DataType("string")).
		Param(ws.QueryParameter("Namespace", "The namespace of the runtimes").DataType("string")).
		To(h.GetRuntimes).
		Operation("Get runtimes").
		Returns(200, "OK", []apis.Runtime{}).
		Returns(400, "Not Found", nil),
	)
	return ws
}

package scene

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

type ScenesHandler struct {
	clientSet *clients.ClientSet
	manager   *manager.Manager
	mu        sync.Mutex
}

var _ Handler = &ScenesHandler{}

// NewScenesHandler 创建一个 SceneHandler
func NewScenesHandler(clientSet *clients.ClientSet) *ScenesHandler {
	return &ScenesHandler{
		manager:   manager.NewManager(clientSet),
		clientSet: clientSet,
	}
}

func (h *ScenesHandler) GetScenes(request *restful.Request, response *restful.Response) {
	// 获取namespace
	namespace := request.QueryParameter(NAME_SPACE)

	var results *apis.SceneList
	var err error

	labels := request.QueryParameter("Label")
	if labels == "" {
		results, err = h.manager.GetScenes(namespace)
		if err != nil {
			logs.Errorf("Get scenes failed: %v", err)
			err := response.WriteError(http.StatusInternalServerError, err)
			if err != nil {
				logs.Errorf("failed to return a status code")
				return
			}
		}
	} else {
		results, err = h.manager.FilterScenes(namespace, labels)
		if err != nil {
			logs.Errorf("Get scenes with labels failed: %v", err)
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

	err = response.WriteHeaderAndEntity(http.StatusOK, results)
	if err != nil {
		logs.Errorf("failed to return a status code")
		return
	}

	logs.Debugf("Get scenes ")
}

func (h *ScenesHandler) NewGetWebService() *restful.WebService {
	ws := new(restful.WebService)
	ws.Path(SCENES_PATH).
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/").
		Doc("Get all scenes").
		Metadata(restfulspec.KeyOpenAPITags, []string{TAG}).
		Param(ws.QueryParameter("Label", "Labels of the Scene (optional)").DataType("string")).
		Param(ws.QueryParameter("Namespace", "The namespace of the Scene").DataType("string")).
		To(h.GetScenes).
		Operation("Get scenes").
		Returns(200, "OK", []apis.Scene{}).
		Returns(404, "Not Found", nil),
	)

	return ws
}

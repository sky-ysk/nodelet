package data

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

type DatasHandler struct {
	clientSet *clients.ClientSet
	manager   *manager.Manager
	mu        sync.Mutex
}

var _ Handler = &DatasHandler{}

// NewDatasHandler 创建一个 DatasHandler
func NewDatasHandler(clientSet *clients.ClientSet) *DatasHandler {
	return &DatasHandler{
		manager:   manager.NewManager(clientSet),
		clientSet: clientSet,
	}
}

func (h *DatasHandler) GetDatas(request *restful.Request, response *restful.Response) {
	// 获取namespace
	namespace := request.QueryParameter(NAME_SPACE)

	var results *apis.DataList
	var err error

	labels := request.QueryParameter("Label")
	if labels == "" {
		results, err = h.manager.GetDatas(namespace)
		if err != nil {
			logs.Errorf("Get actions failed: %v", err)
			err := response.WriteError(http.StatusInternalServerError, err)
			if err != nil {
				logs.Errorf("failed to return a status code")
				return
			}
		}
	} else {
		results, err = h.manager.FilterDatas(namespace, labels)
		if err != nil {
			logs.Errorf("Get actions with labels failed: %v", err)
			err := response.WriteError(http.StatusInternalServerError, err)
			if err != nil {
				logs.Errorf("failed to return a status code")
				return
			}
		}
	}

	err = response.WriteHeaderAndEntity(http.StatusOK, results)
	if err != nil {
		logs.Errorf("failed to return a status code")
		return
	}

	logs.Debugf("Get datas")
}

// TODO: DeleteAll

func (h *DatasHandler) NewGetWebService() *restful.WebService {
	ws := new(restful.WebService)
	ws.Path(DATAS_PATH).
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/").
		Doc("Get all datas").
		Metadata(restfulspec.KeyOpenAPITags, []string{TAG}).
		Param(ws.QueryParameter("Label", "Labels of the datas (optional)").DataType("string")).
		Param(ws.QueryParameter("Namespace", "The namespace of the datas").DataType("string")).
		To(h.GetDatas).
		Operation("Get datas").
		Returns(200, "OK", []apis.Data{}).
		Returns(404, "Not Found", nil),
	)

	return ws
}

package group

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

type GroupsHandler struct {
	clientSet *clients.ClientSet
	manager   *manager.Manager
	mu        sync.Mutex
}

var _ Handler = &GroupsHandler{}

// NewGroupsHandler 创建一个 GroupsHandler
func NewGroupsHandler(clientSet *clients.ClientSet) *GroupsHandler {
	return &GroupsHandler{
		manager:   manager.NewManager(clientSet),
		clientSet: clientSet,
	}
}

func (h *GroupsHandler) GetGroups(request *restful.Request, response *restful.Response) {
	// 从url中获取namespace
	namespace := request.QueryParameter(NAME_SPACE)

	labels := request.QueryParameter("Label")
	var results *apis.GroupList
	var err error
	if labels == "" {
		results, err = h.manager.GetGroups(namespace)
		if err != nil {
			logs.Errorf("Get groups with labels failed: %v", err)
			err := response.WriteError(http.StatusInternalServerError, err)
			if err != nil {
				logs.Errorf("failed to return a status code")
				return
			}
		}
	}

	results, err = h.manager.FilterGroups(namespace, labels)
	if err != nil {
		logs.Errorf("Get groups failed: %v", err)
		err := response.WriteError(http.StatusInternalServerError, err)
		if err != nil {
			logs.Errorf("failed to return a status code")
			return
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
	logs.Debugf("Get groups")
}

func (h *GroupsHandler) NewGetWebService() *restful.WebService {
	ws := new(restful.WebService)
	ws.Path(GROUPS_PATH).
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/").
		Doc("Get all groups").
		Metadata(restfulspec.KeyOpenAPITags, []string{TAG}).
		Param(ws.QueryParameter("Label", "Labels of the groups (optional)").DataType("string")).
		Param(ws.QueryParameter("Namespace", "The namespace of the groups").DataType("string")).
		To(h.GetGroups).
		Operation("Get groups").
		Returns(200, "OK", []apis.Group{}).
		Returns(400, "Not Found", nil),
	)
	return ws
}

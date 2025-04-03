package group

import (
	"context"
	restfulspec "github.com/emicklei/go-restful-openapi/v2"
	"github.com/emicklei/go-restful/v3"
	apis "hit.edu/framework/pkg/apis/cores"
	metav1 "hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/client-go/clients"
	"hit.edu/framework/pkg/client-go/clients/typed/core"
	"hit.edu/framework/pkg/component-base/logs"
	"net/http"
)

type GroupsHandler struct {
	client core.GroupInterface
}

var _ Handler = &GroupsHandler{}

func NewGroupsHandler(clientSet *clients.ClientSet) *GroupsHandler {
	c := clientSet.Core().Groups("test") // apis.NamespaceAll
	return &GroupsHandler{
		client: c,
	}
}

func (h *GroupsHandler) GetGroups(request *restful.Request, response *restful.Response) {
	results, err := h.client.List(context.TODO(), metav1.ListOptions{})
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

	ws.Route(ws.GET("/{Namespace}/groups").
		Doc("Get all groups").
		Metadata(restfulspec.KeyOpenAPITags, []string{TAG}).
		Param(ws.QueryParameter("Namespace", "The namespace of the groups").DataType("string")).
		To(h.GetGroups).
		Operation("Get groups").
		Returns(200, "OK", []apis.Group{}).
		Returns(400, "Not Found", nil),
	)
	return ws
}

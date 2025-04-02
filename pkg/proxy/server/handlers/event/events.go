package event

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

type EventsHandler struct {
	client core.EventInterface
}

var _ Handler = &EventsHandler{}

func NewEventsHandler(clientSet *clients.ClientSet) *EventsHandler {
	c := clientSet.Core().Events(apis.NamespaceAll)
	return &EventsHandler{
		client: c,
	}
}

func (h *EventsHandler) GetEvents(request *restful.Request, response *restful.Response) {
	results, err := h.client.List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		logs.Errorf("Get events failed: %v", err)
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
	logs.Debugf("Get events")
}

// TODO: DeleteAll

func (h *EventsHandler) NewGetWebService() *restful.WebService {
	ws := new(restful.WebService)
	ws.Path(EVENTS_PATH).
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("").
		Doc("Get all events").
		Metadata(restfulspec.KeyOpenAPITags, []string{TAG}).
		To(h.GetEvents).
		Operation("Get events").
		Returns(200, "OK", []apis.Event{}).
		Returns(400, "Not Found", nil),
	)

	return ws
}

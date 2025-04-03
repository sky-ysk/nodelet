package device

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

type DevicesHandler struct {
	client core.DeviceInterface
}

var _ Handler = &DevicesHandler{}

func NewDevicesHandler(clientSet *clients.ClientSet) *DevicesHandler {
	c := clientSet.Core().Devices("test") //apis.NamespaceAll
	return &DevicesHandler{
		client: c,
	}
}

func (h *DevicesHandler) GetDevices(request *restful.Request, response *restful.Response) {
	results, err := h.client.List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		logs.Errorf("Get devices failed: %v", err)
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
	logs.Debugf("Get devices")
}

// TODO: DeleteAll

func (h *DevicesHandler) NewGetWebService() *restful.WebService {
	ws := new(restful.WebService)
	ws.Path(DEVICES_PATH).
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/{Namespace}/devices").
		Doc("Get all devices").
		Metadata(restfulspec.KeyOpenAPITags, []string{TAG}).
		Param(ws.QueryParameter("Namespace", "The namespace of the devices").DataType("string")).
		To(h.GetDevices).
		Operation("Get devices").
		Returns(200, "OK", []apis.Device{}).
		Returns(400, "Not Found", nil),
	)

	return ws
}

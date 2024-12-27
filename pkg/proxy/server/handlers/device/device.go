package device

import (
	"fmt"
	restfulspec "github.com/emicklei/go-restful-openapi/v2"
	"github.com/emicklei/go-restful/v3"
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/component-base/logs"
	"net/http"
)

type DeviceHandler struct{}

var _ Handler = &DeviceHandler{}

func NewDeviceHandler() *DeviceHandler {
	return &DeviceHandler{}
}

func GetDevice(request *restful.Request, response *restful.Response) {
	// TODO: 使用client-go实现查询
	name := request.PathParameter(DEVICE_NAME)
	device := apis.Device{
		Spec: apis.DeviceSpec{
			Name: name,
		},
	}
	err := response.WriteEntity(device)
	if err != nil {
		response.WriteError(http.StatusInternalServerError, err)
	}
	logs.Debugf("get devices")
}

func (h *DeviceHandler) NewGetWebService() *restful.WebService {
	ws := new(restful.WebService)
	ws.Path(DEVICE_PATH).
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)
	ws.Route(ws.GET(fmt.Sprintf("/{%s}", DEVICE_NAME)).
		To(GetDevice).
		Doc("Get a device with name").
		Metadata(restfulspec.KeyOpenAPITags, []string{TAG}).
		Operation("getDevice").
		Returns(200, "OK", apis.Device{}).
		Returns(400, "Not Found", nil),
	)
	return ws
}

package device

import (
	restfulspec "github.com/emicklei/go-restful-openapi/v2"
	"github.com/emicklei/go-restful/v3"
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/component-base/logs"
	"net/http"
)

type DevicesHandler struct{}

var _ Handler = &DevicesHandler{}

func NewDevicesHandler() *DevicesHandler {
	return &DevicesHandler{}
}

func GetDevices(request *restful.Request, response *restful.Response) {
	// TODO: 使用client-go实现查询
	device1 := apis.Device{
		Spec: apis.DeviceSpec{
			Name: "TestDevice",
		},
	}

	device2 := apis.Device{
		Spec: apis.DeviceSpec{
			Name: "TestDevice",
		},
	}

	devices := []apis.Device{device1, device2}

	err := response.WriteEntity(devices)
	if err != nil {
		response.WriteError(http.StatusInternalServerError, err)
	}
	logs.Debugf("get devices")
}

func (h *DevicesHandler) NewGetWebService() *restful.WebService {
	ws := new(restful.WebService)
	ws.Path(DEVICES_PATH).
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("").
		//Docs
		Doc("Get all devices").
		Metadata(restfulspec.KeyOpenAPITags, []string{TAG}).
		To(GetDevices).
		Operation("getDevices").
		Returns(200, "OK", []apis.Device{}).
		Returns(400, "Not Found", nil),
	)
	return ws
}

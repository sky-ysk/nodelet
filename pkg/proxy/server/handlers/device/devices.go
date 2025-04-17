package device

import (
	"context"
	"fmt"
	restfulspec "github.com/emicklei/go-restful-openapi/v2"
	"github.com/emicklei/go-restful/v3"
	apis "hit.edu/framework/pkg/apis/cores"
	metav1 "hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/client-go/clients"
	"hit.edu/framework/pkg/client-go/clients/typed/core"
	"hit.edu/framework/pkg/component-base/logs"
	"net/http"
	"sync"
)

type DevicesHandler struct {
	clients   map[string]core.DeviceInterface
	clientSet *clients.ClientSet
	mu        sync.Mutex
}

type CurrentDevicesHandler struct {
	client core.DeviceInterface
}

var _ Handler = &DevicesHandler{}

//func NewDevicesHandler(clientSet *clients.ClientSet) *DevicesHandler {
//	c := clientSet.Core().Devices("test") //apis.NamespaceAll
//	return &DevicesHandler{
//		client: c,
//	}
//}

// NewDevicesHandler 创建一个 DevicesHandler
func NewDevicesHandler(clientSet *clients.ClientSet) *DevicesHandler {
	return &DevicesHandler{
		clients:   make(map[string]core.DeviceInterface),
		clientSet: clientSet,
	}
}

// GetClient 根据 namespace 获取 client，如果不存在则创建
func (h *DevicesHandler) GetClient(namespace string) *CurrentDevicesHandler {
	h.mu.Lock()
	defer h.mu.Unlock()

	// 如果已经存在，直接返回
	if c, exists := h.clients[namespace]; exists {
		return &CurrentDevicesHandler{
			client: c,
		}
	}

	// 否则创建新的 client
	newClient := h.clientSet.Core().Devices(namespace)
	h.clients[namespace] = newClient
	return &CurrentDevicesHandler{
		client: newClient,
	}
}

func (h *DevicesHandler) GetDevices(request *restful.Request, response *restful.Response) {
	c := &CurrentDevicesHandler{}
	// 从url中获取namespace
	namespace := request.QueryParameter(NAME_SPACE)
	if namespace == "" {
		err := response.WriteError(http.StatusBadRequest, fmt.Errorf("namespace is required"))
		if err != nil {
			logs.Errorf("failed to return a status code ")
			return
		}
		return
	} else {
		c = h.GetClient(namespace)
	}
	results, err := c.client.List(context.TODO(), metav1.ListOptions{})
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

	ws.Route(ws.GET("/").
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

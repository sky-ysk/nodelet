package device

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

type DevicesHandler struct {
	clientSet *clients.ClientSet
	manager   *manager.Manager
	mu        sync.Mutex
}

var _ Handler = &DevicesHandler{}

// NewDevicesHandler 创建一个 DevicesHandler
func NewDevicesHandler(clientSet *clients.ClientSet) *DevicesHandler {
	return &DevicesHandler{
		manager:   manager.NewManager(clientSet),
		clientSet: clientSet,
	}
}

// GetClient 根据 namespace 获取 client，如果不存在则创建
//func (h *DevicesHandler) GetClient(namespace string) *CurrentDevicesHandler {
//	h.mu.Lock()
//	defer h.mu.Unlock()
//
//	// 如果已经存在，直接返回
//	if c, exists := h.clients[namespace]; exists {
//		return &CurrentDevicesHandler{
//			client: c,
//		}
//	}
//
//	// 否则创建新的 client
//	newClient := h.clientSet.Core().Devices(namespace)
//	h.clients[namespace] = newClient
//	return &CurrentDevicesHandler{
//		client: newClient,
//	}
//}

func (h *DevicesHandler) GetDevices(request *restful.Request, response *restful.Response) {
	// 从url中获取namespace
	namespace := request.QueryParameter(NAME_SPACE)
	var results *apis.DeviceList
	var err error

	labels := request.QueryParameter("Label")
	if labels == "" {
		results, err = h.manager.GetDevices(namespace)
		if err != nil {
			logs.Errorf("Get actions failed: %v", err)
			err := response.WriteError(http.StatusInternalServerError, err)
			if err != nil {
				logs.Errorf("failed to return a status code")
				return
			}
		}
	}

	results, err = h.manager.FilterDevices(namespace, labels)
	if err != nil {
		logs.Errorf("Get actions with labels failed: %v", err)
		err := response.WriteError(http.StatusInternalServerError, err)
		if err != nil {
			logs.Errorf("failed to return a status code")
			return
		}
	}

	//devices, err := h.manager.GetDevices(namespace)
	//if namespace == "" {
	//	err := response.WriteError(http.StatusBadRequest, fmt.Errorf("namespace is required"))
	//	if err != nil {
	//		logs.Errorf("failed to return a status code ")
	//		return
	//	}
	//	return
	//} else {
	//	c = h.GetClient(namespace)
	//}
	//c = h.GetClient(namespace)
	//results, err := c.client.List(context.TODO(), metav1.ListOptions{})
	//if err != nil {
	//	logs.Errorf("Get devices failed: %v", err)
	//	err := response.WriteError(http.StatusInternalServerError, err)
	//	if err != nil {
	//		logs.Errorf("failed to return a status code")
	//		return
	//	}
	//}

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

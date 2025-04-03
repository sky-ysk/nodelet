package device

import (
	"github.com/emicklei/go-restful/v3"
)

const (
	//NAMESPACE    = "resource"
	//GROUP        = "v1"
	//TAG          = "Device"
	//API_PREFIX   = "/" + NAMESPACE + "/" + GROUP
	//DEVICES_PATH = API_PREFIX + "/devices"
	//DEVICE_PATH  = API_PREFIX + "/device"
	//DEVICE_NAME  = "Name"

	TAG          = "Device"
	DEVICES_PATH = "/apis/resources/v1/namespaces/devices"
	DEVICE_PATH  = "/apis/resources/v1/namespaces/device"
	DEVICE_NAME  = "Name"
	NAMESPACE    = "Namespace"
)

type Handler interface {
	// 查询
	NewGetWebService() *restful.WebService
}

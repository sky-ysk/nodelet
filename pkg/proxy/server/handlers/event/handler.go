package event

import "github.com/emicklei/go-restful/v3"

const (
	//NAMESPACE   = "framework"
	//GROUP       = "v1"
	//TAG         = "Event"
	//API_PREFIX  = "/" + NAMESPACE + "/" + GROUP
	//EVENTS_PATH = API_PREFIX + "/events"
	//EVENT_PATH  = API_PREFIX + "/event"
	//EVENT_NAME  = "Name"

	TAG         = "Event"
	EVENTS_PATH = "/apis/resources/v1/namespaces/events"
	EVENT_PATH  = "/apis/resources/v1/namespaces/event"
	EVENT_NAME  = "Name"
	NAMESPACE   = "Namespace"
)

type Handler interface {
	// 查询
	NewGetWebService() *restful.WebService
}

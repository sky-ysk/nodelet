package runtime

import "github.com/emicklei/go-restful/v3"

const (
	NAMESPACE     = "framework"
	GROUP         = "v1"
	TAG           = "Runtime"
	API_PREFIX    = "/" + NAMESPACE + "/" + GROUP
	RUNTIMES_PATH = API_PREFIX + "/runtimes"
	RUNTIME_PATH  = API_PREFIX + "/runtime"
	RUNTIME_NAME  = "Name"
	NAME_SPACE    = "Namespace"
)

type Handler interface {
	// 查询
	NewGetWebService() *restful.WebService
}

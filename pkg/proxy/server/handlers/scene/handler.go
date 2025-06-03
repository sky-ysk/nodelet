package scene

import (
	"github.com/emicklei/go-restful/v3"
)

const (
	NAMESPACE   = "framework"
	GROUP       = "v1"
	TAG         = "Scene"
	API_PREFIX  = "/" + NAMESPACE + "/" + GROUP
	SCENES_PATH = API_PREFIX + "/Scenes"
	SCENE_PATH  = API_PREFIX + "/scene"
	SCENE_NAME  = "Name"
	NAME_SPACE  = "Namespace"
)

type Handler interface {
	// 查询
	NewGetWebService() *restful.WebService
}

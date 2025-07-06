package data

import (
	"github.com/emicklei/go-restful/v3"
)

const (
	NAMESPACE  = "framework"
	DATA       = "v1"
	TAG        = "Data"
	API_PREFIX = "/" + NAMESPACE + "/" + DATA
	DATAS_PATH = API_PREFIX + "/datas"
	DATA_PATH  = API_PREFIX + "/data"
	DATA_NAME  = "Name"
	NAME_SPACE = "Namespace"
)

type Handler interface {
	// 查询
	NewGetWebService() *restful.WebService
}

package task

import (
	"github.com/emicklei/go-restful/v3"
)

const (
	NAMESPACE  = "framework"
	GROUP      = "v1"
	TAG        = "Task"
	API_PREFIX = "/" + NAMESPACE + "/" + GROUP
	TASKS_PATH = API_PREFIX + "/tasks"
	TASK_PATH  = API_PREFIX + "/task"
	TASK_NAME  = "Name"
)

type Handler interface {
	// 查询
	NewGetWebService() *restful.WebService
}

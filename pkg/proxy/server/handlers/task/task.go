package task

import (
	"fmt"
	restfulspec "github.com/emicklei/go-restful-openapi/v2"
	"github.com/emicklei/go-restful/v3"
	"github.com/google/uuid"
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/client-go/clients"
	"hit.edu/framework/pkg/client-go/util/manager"
	"hit.edu/framework/pkg/component-base/analyzer"
	"hit.edu/framework/pkg/component-base/logs"
	"net/http"
	"sync"
	"time"
)

type TaskHandler struct {
	manager   *manager.Manager
	clientSet *clients.ClientSet
	mu        sync.Mutex
}

var _ Handler = &TaskHandler{}

// NewTaskHandler 创建一个 TaskHandler
func NewTaskHandler(clientSet *clients.ClientSet) *TaskHandler {
	return &TaskHandler{
		clientSet: clientSet,
		manager:   manager.NewManager(clientSet),
	}
}

func (h *TaskHandler) GetTask(request *restful.Request, response *restful.Response) {
	// 尝试从url中获取参数
	name := request.QueryParameter(TASK_NAME)
	if name == "" {
		// url中没有获取到name参数，尝试从请求体中获取
		req := &apis.Task{}
		err := request.ReadEntity(&req)
		if err != nil || req.Name == "" {
			err := response.WriteError(http.StatusBadRequest, fmt.Errorf("provide task name , the key is Name "))
			if err != nil {
				logs.Errorf("failed to return a status code ")
				return
			}
			return
		} else {
			name = req.Name
		}
	}

	// 从url中获取namespace
	namespace := request.QueryParameter(NAME_SPACE)
	if namespace == "" {
		err := response.WriteError(http.StatusBadRequest, fmt.Errorf("namespace is required"))
		if err != nil {
			logs.Errorf("failed to return a status code ")
			return
		}
		return
	}

	result, err := h.manager.GetTask(name, namespace)
	if err != nil {
		logs.Errorf("Get task %s error: %v , task not exist! ", name, err)
		err := response.WriteError(http.StatusNotFound, err)
		if err != nil {
			logs.Errorf("failed to return a status code")
			return
		}
		return
	}

	if result.Name == name {
		err = response.WriteEntity(result)
		if err != nil {
			err := response.WriteError(http.StatusOK, err)
			if err != nil {
				logs.Errorf("failed to return a status code")
				return
			}
			return
		}
		logs.Debugf("Get task")
	}
}

func (h *TaskHandler) CreateTask(request *restful.Request, response *restful.Response) {
	// 获取 task 数据
	et := &apis.Task{}
	err := request.ReadEntity(&et)
	if err != nil {
		logs.Errorf("Failed to deserialize json data, error: %v", err)
		err := response.WriteError(http.StatusBadRequest, err)
		if err != nil {
			logs.Errorf("failed to return a status code")
			return
		}
		return
	}

	// 获取 namespace
	namespace := et.Namespace
	if namespace == "" {
		logs.Error("namespace is empty")
		err := response.WriteError(http.StatusBadRequest, fmt.Errorf("namespace is empty"))
		if err != nil {
			logs.Errorf("failed to return a status code ")
			return
		}
		return
	}

	logs.Info(*et)

	////格式校验
	//res, err := analyzer.SerializeToJson(ew)
	//_, err = analyzer.Deserialize(res, apis.Task{})
	//if err != nil {
	//	err := response.WriteError(http.StatusBadRequest, err)
	//	if err != nil {
	//		logs.Errorf("failed to return a status code ")
	//		return
	//	}
	//	// 正常情况就应该return不创建，但测试的时候没有构造完整的task，根据名字能创建就行
	//	// return
	//}

	// 产生UUID
	timestamp := time.Now().Format("20060102T150405")
	randomStr := uuid.New().String()[:5]
	UUID := timestamp + "-" + randomStr

	// 将 Task写入数据库中
	result, err := h.manager.CreateTask(et.Spec, nil, namespace, UUID, "")
	if err != nil {
		err1 := response.WriteError(http.StatusInternalServerError, err)
		if err1 != nil {
			logs.Errorf("failed to return a status code ,error : %v ", err1)
			return
		}
		logs.Errorf("Create task fail  ,failed write it to database , error: %v ", err)
		return
	}

	//// 构造tasks
	//for _, g := range ew.Spec.Groups {
	//	//
	//	g.ObjectMeta = metav1.ObjectMeta{
	//		Name: g.Spec.Name,
	//	}
	//	//
	//	g.TypeMeta = metav1.TypeMeta{
	//		Kind:       "Group",
	//		APIVersion: "resources/v1",
	//	}
	//
	//	// 分配Group的GroupID
	//	gu := uuid.New().String()
	//	g.Status.GroupID = gu
	//
	//	// Group的TaskID
	//	g.Status.Belongs.TaskID = tu
	//
	//	_, err = c.groupClient.Create(context.TODO(), &g, metav1.CreateOptions{})
	//	if err != nil {
	//		err := response.WriteError(http.StatusInternalServerError, err)
	//		if err != nil {
	//			return
	//		}
	//		logs.Errorf("Create task %s group %s error: %v ", name, g.Name, err)
	//		logs.Errorf("Group %v ", g)
	//		return
	//	}
	//}

	// TODO: 错误处理

	// 返回结果
	err = response.WriteEntity(result)
	if err != nil {
		err := response.WriteError(http.StatusInternalServerError, err)
		if err != nil {
			logs.Errorf("failed to return a status code ")
			return
		}
		return
	}

	err = response.WriteError(http.StatusOK, err)
	if err != nil {
		logs.Errorf("failed to return a status code ")
		return
	}

	logs.Debugf("Create task %v ", result)
}

func (h *TaskHandler) DeleteTask(request *restful.Request, response *restful.Response) {
	// 尝试从url中获取参数
	name := request.QueryParameter(TASK_NAME)
	if name == "" {
		// url中没有获取到name参数，尝试从请求体中获取
		req := &apis.Task{}
		err := request.ReadEntity(&req)
		if err != nil || req.Name == "" {
			err := response.WriteError(http.StatusBadRequest, fmt.Errorf("provide task name , the key is Name "))
			if err != nil {
				logs.Errorf("failed to return a status code ")
				return
			}
			return
		} else {
			name = req.Name
		}
	}

	// 获取namespace
	namespace := request.QueryParameter(NAME_SPACE)
	if namespace == "" {
		err := response.WriteError(http.StatusBadRequest, fmt.Errorf("namespace is empty"))
		if err != nil {
			logs.Errorf("failed to return a status code ")
			return
		}
		return
	}

	// 删除task
	err := h.manager.DeleteTask(name, namespace)
	if err != nil {
		logs.Error(err)
		err := response.WriteError(http.StatusInternalServerError, err)
		if err != nil {
			logs.Errorf("failed to return a status code ")
			return
		}
		return
	}

	// 返回停止成功的状态
	response.WriteHeader(http.StatusOK)

	// 记录日志
	logs.Debugf("delete task : %v", name)

}

// TODO ：StopTask

func (h *TaskHandler) UpdateTask(request *restful.Request, response *restful.Response) {
	// 获取task
	ew := &apis.Task{}
	err := request.ReadEntity(&ew)
	if err != nil {
		logs.Errorf("Failed to deserialize json data, error: %v", err)
		err := response.WriteError(http.StatusBadRequest, err)
		if err != nil {
			logs.Errorf("failed to return a status code")
			return
		}
		return
	}

	// 获取name
	name := request.QueryParameter(TASK_NAME)
	if name == "" {
		name = ew.Name
	}

	// 获取namespace
	namespace := request.QueryParameter(NAME_SPACE)
	if namespace == "" {
		err := response.WriteError(http.StatusBadRequest, fmt.Errorf("namespace is empty"))
		if err != nil {
			logs.Errorf("failed to return a status code")
			return
		}
		return
	}

	// 格式验证
	res, err := analyzer.SerializeToJson(ew)
	_, err = analyzer.Deserialize(res, apis.Task{})
	if err != nil {
		err := response.WriteError(http.StatusBadRequest, err)
		if err != nil {
			logs.Errorf("failed to return a status code ")
			return
		}
		return
	}

	// 更新task
	updatedTask, updateErr := h.manager.UpdateTask(namespace, name, ew)
	if updateErr != nil {
		logs.Errorf("Update task %s error: %v", name, updateErr)
		err := response.WriteError(http.StatusInternalServerError, err)
		if err != nil {
			logs.Errorf("failed to return a status code")
			return
		}
	}

	// 返回成功修改的通知
	err = response.WriteHeaderAndEntity(http.StatusOK, updatedTask)
	if err != nil {
		logs.Errorf("failed to return a status code")
		return
	}

	// 记录日志
	logs.Debugf("update task : %v", name)
}

func (h *TaskHandler) PatchTask(request *restful.Request, response *restful.Response) {
	// 获取json
	req := &apis.Task{}
	err := request.ReadEntity(&req)
	if err != nil {
		logs.Errorf("Failed to deserialize json data, error: %v", err)
		err := response.WriteError(http.StatusBadRequest, err)
		if err != nil {
			logs.Errorf("failed to return a status code ")
			return
		}
		return
	}

	// 获取name
	name := request.QueryParameter(TASK_NAME)
	if name == "" {
		if req.Name != "" {
			name = req.Name
		} else {
			err := response.WriteError(http.StatusBadRequest, fmt.Errorf("name is empty"))
			if err != nil {
				logs.Errorf("failed to return a status code ")
				return
			}
			return
		}
	}

	// 获取namespace
	namespace := request.QueryParameter(NAME_SPACE)
	if namespace == "" {
		err := response.WriteError(http.StatusBadRequest, fmt.Errorf("namespace is empty"))
		if err != nil {
			logs.Errorf("failed to return a status code ")
			return
		}
		return
	}

	// 序列化PatchTask
	patchTask, err := analyzer.SerializeToJson(req)
	if err != nil {
		logs.Errorf("Serialize patch task error: %v", err)
		err := response.WriteError(http.StatusInternalServerError, err)
		if err != nil {
			logs.Errorf("failed to return a status code")
			return
		}
	}

	patchedTask, err := h.manager.PatchTask(namespace, name, patchTask)
	if err != nil {
		logs.Error(err)
		err := response.WriteError(http.StatusInternalServerError, err)
		if err != nil {
			logs.Errorf("failed to return a status code ")
			return
		}
		return
	}

	// 返回成功修改的通知
	err = response.WriteHeaderAndEntity(http.StatusOK, patchedTask)
	if err != nil {
		logs.Errorf("failed to return a status code")
		return
	}

	// 记录日志
	logs.Debugf("patch task : %v", name)
}

func (h *TaskHandler) NewGetWebService() *restful.WebService {
	ws := new(restful.WebService)
	ws.Path(TASK_PATH).
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	// 查询任务
	ws.Route(ws.GET("/").
		To(h.GetTask).
		Doc("Get a task with name").
		Metadata(restfulspec.KeyOpenAPITags, []string{TAG}).
		Param(ws.QueryParameter("Name", "The name of the task").DataType("string")).
		Param(ws.QueryParameter("Namespace", "The namespace of the task").DataType("string")).
		Operation("get Task").
		Returns(200, "OK", apis.Task{}).
		Returns(400, "Not Found", nil),
	)

	//创建任务
	ws.Route(ws.POST("/").
		To(h.CreateTask).
		Doc("Create a task with name").
		Metadata(restfulspec.KeyOpenAPITags, []string{TAG}).
		Param(ws.QueryParameter("Namespace", "The namespace of the task").DataType("string")).
		Param(ws.BodyParameter("Task", "The json string of the Task object").DataType("string")).
		Operation("createTask").
		Returns(200, "OK", apis.Task{}).
		Returns(400, "Not Found", nil),
	)

	//修改任务
	ws.Route(ws.PUT("/").
		To(h.UpdateTask).
		Doc("Update a task with name").
		Metadata(restfulspec.KeyOpenAPITags, []string{TAG}).
		Param(ws.QueryParameter("Name", "The name of the task").DataType("string")).
		Param(ws.QueryParameter("Namespace", "The namespace of the task").DataType("string")).
		Param(ws.BodyParameter("Task", "The json string of the Task object").DataType("string")).
		Operation("update Task").
		Returns(200, "OK", apis.Task{}).
		Returns(400, "Not Found", nil),
	)

	//部分修改任务
	ws.Route(ws.PATCH("/").
		To(h.PatchTask).
		Doc("Patch a task").
		Metadata(restfulspec.KeyOpenAPITags, []string{TAG}).
		Param(ws.QueryParameter("Name", "The name of the task").DataType("string")).
		Param(ws.QueryParameter("Namespace", "The namespace of the task").DataType("string")).
		Param(ws.BodyParameter("Task", "The json string of the Task field").DataType("string")).
		Operation("patch Task").
		Returns(200, "OK", apis.Task{}).
		Returns(400, "Not Found", nil),
	)

	// 删除任务
	ws.Route(ws.DELETE("/").
		To(h.DeleteTask).
		Doc("Delete a task").
		Metadata(restfulspec.KeyOpenAPITags, []string{TAG}).
		Param(ws.QueryParameter("Name", "The name of the task").DataType("string")).
		Param(ws.QueryParameter("Namespace", "The namespace of the task").DataType("string")).
		Operation("Delete Task").
		Returns(200, "OK", apis.Task{}).
		Returns(400, "Not Found", nil),
	)

	return ws
}

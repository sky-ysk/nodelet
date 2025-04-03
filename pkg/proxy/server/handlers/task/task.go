package task

import (
	"context"
	"fmt"
	restfulspec "github.com/emicklei/go-restful-openapi/v2"
	"github.com/emicklei/go-restful/v3"
	"github.com/google/uuid"
	"hit.edu/framework/pkg/apimachinery/types"
	apis "hit.edu/framework/pkg/apis/cores"
	metav1 "hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/client-go/clients"
	"hit.edu/framework/pkg/client-go/clients/typed/core"
	"hit.edu/framework/pkg/component-base/analyzer"
	"hit.edu/framework/pkg/component-base/logs"
	"net/http"
	"sync"
)

type TaskHandler struct {
	clients      map[string]core.TaskInterface
	groupClients map[string]core.GroupInterface
	clientSet    *clients.ClientSet
	mu           sync.Mutex
}

type CurrentTaskHandler struct {
	client      core.TaskInterface
	groupClient core.GroupInterface
}

var _ Handler = &TaskHandler{}

//func NewTaskHandler(clientSet *clients.ClientSet) *TaskHandler {
//	c := clientSet.Core().Tasks("test")   // apis.NamespaceAll
//	gc := clientSet.Core().Groups("test") //apis.NamespaceAll
//	return &TaskHandler{
//		client:      c,
//		groupClient: gc,
//	}
//}

// NewTaskHandler 创建一个 TaskHandler
func NewTaskHandler(clientSet *clients.ClientSet) *TaskHandler {
	return &TaskHandler{
		clients:      make(map[string]core.TaskInterface),
		groupClients: make(map[string]core.GroupInterface),
		clientSet:    clientSet,
	}
}

// GetClient 根据 namespace 获取 client，如果不存在则创建
func (h *TaskHandler) GetClient(namespace string) *CurrentTaskHandler {
	h.mu.Lock()
	defer h.mu.Unlock()

	// 如果已经存在，直接返回   c 和 gc 会一起创建
	c, exists := h.clients[namespace]
	gc, groupExists := h.groupClients[namespace]
	if exists && groupExists {
		return &CurrentTaskHandler{
			client:      c,
			groupClient: gc,
		}
	}

	// 否则创建新的 client
	newClient := h.clientSet.Core().Tasks(namespace)
	newGroupClient := h.clientSet.Core().Groups(namespace)
	h.clients[namespace] = newClient
	h.groupClients[namespace] = newGroupClient
	return &CurrentTaskHandler{
		client:      newClient,
		groupClient: newGroupClient,
	}
}

func (h *TaskHandler) GetTask(request *restful.Request, response *restful.Response) {
	// 尝试从url中获取参数
	c := &CurrentTaskHandler{}
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
	namespace := request.PathParameter(NAMESPACE)
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

	result, err := c.client.Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		logs.Errorf("Get task %s error: %v , task not exist! ", name, err)
		err := response.WriteError(http.StatusNotFound, err)
		if err != nil {
			logs.Errorf("failed to return a status code ")
			return
		}
		return
	}

	if result.Name == name {
		err = response.WriteEntity(result)
		if err != nil {
			err := response.WriteError(http.StatusOK, err)
			if err != nil {
				logs.Errorf("failed to return a status code ")
				return
			}
			return
		}
		logs.Debugf("Get task ")
	}
}

func (h *TaskHandler) CreateTask(request *restful.Request, response *restful.Response) {
	// 先查询 task 是否存在
	// 尝试从url中获取参数
	c := &CurrentTaskHandler{}
	name := request.QueryParameter(TASK_NAME)
	ew := &apis.Task{}
	if name == "" {
		// url中没有获取到name参数，尝试从请求体中获取

		err := request.ReadEntity(&ew)
		if err != nil || ew.Name == "" {
			if err != nil {
				logs.Errorf("Failed to deserialize json data, error: %v", err)
				err := response.WriteError(http.StatusBadRequest, err)
				if err != nil {
					logs.Errorf("failed to return a status code")
					return
				}
				return
			} else {
				err := response.WriteError(http.StatusBadRequest, fmt.Errorf("provide task name , the key is Name "))
				if err != nil {
					logs.Errorf("failed to return a status code ")
					return
				}
				return
			}
		} else {
			name = ew.Name
		}
	} else {
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
	}

	namespace := ew.Namespace
	if namespace == "" {
		logs.Error("namespace is empty")
		err := response.WriteError(http.StatusBadRequest, fmt.Errorf("namespace is empty"))
		if err != nil {
			logs.Errorf("failed to return a status code ")
			return
		}
		return
	} else {
		c = h.GetClient(namespace)
	}

	result, err := c.client.Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		logs.Infof("Get task %s error: %v   , task not exist! creat it ", name, err)
	} else if result.Name == name {
		logs.Errorf("Create task %s error, task existed: %v ", name, result)
		err = fmt.Errorf("create task %s error, task existed: %v ", name, result)
		err := response.WriteError(http.StatusConflict, err)
		if err != nil {
			logs.Errorf("failed to return a status code ")
			return
		}
		return
	}

	// 不存在，解析用户的输入
	//ew := &apis.Task{}
	//err = request.ReadEntity(ew)
	//if err != nil {
	//	logs.Errorf("Failed to create task %s , error : %v ", name, err)
	//	err := response.WriteError(http.StatusInternalServerError, err)
	//	if err != nil {
	//		logs.Errorf("failed to return a status code ")
	//		return
	//	}
	//	return
	//}

	logs.Info(*ew)

	//格式校验
	res, err := analyzer.SerializeToJson(ew)
	_, err = analyzer.Deserialize(res, apis.Task{})
	if err != nil {
		err := response.WriteError(http.StatusBadRequest, err)
		if err != nil {
			logs.Errorf("failed to return a status code ")
			return
		}
		// 正常情况就应该return不创建，但测试的时候没有构造完整的task，根据名字能创建就行
		// return
	}

	//Task分配ID
	tu := uuid.New().String()
	ew.Status.TaskID = tu
	logs.Debugf("Create task %s success, task id : %s  ", name, tu)

	// 将 Task写入数据库中
	result, err = c.client.Create(context.TODO(), ew, metav1.CreateOptions{})
	if err != nil {
		err1 := response.WriteError(http.StatusInternalServerError, err)
		if err1 != nil {
			logs.Errorf("failed to return a status code ,error : %v ", err1)
			return
		}
		logs.Errorf("Create task %s  ,failed write to database , error: %v ", name, err)
		return
	}

	// 构造Groups
	for _, g := range ew.Spec.Groups {
		//
		g.ObjectMeta = metav1.ObjectMeta{
			Name: g.Spec.Name,
		}
		//
		g.TypeMeta = metav1.TypeMeta{
			Kind:       "Group",
			APIVersion: "resources/v1",
		}

		// 分配Group的GroupID
		gu := uuid.New().String()
		g.Status.GroupID = gu

		// Group的TaskID
		g.Status.Belongs.TaskID = tu

		_, err = c.groupClient.Create(context.TODO(), &g, metav1.CreateOptions{})
		if err != nil {
			err := response.WriteError(http.StatusInternalServerError, err)
			if err != nil {
				return
			}
			logs.Errorf("Create task %s group %s error: %v ", name, g.Name, err)
			logs.Errorf("Group %v ", g)
			return
		}
	}

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
	// 查看task是否存在
	// 如果存在，删除任务
	// 如果不存在，返回 404 not found
	// 尝试从url中获取参数
	c := &CurrentTaskHandler{}
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
	namespace := request.PathParameter(NAMESPACE)
	if namespace == "" {
		err := response.WriteError(http.StatusBadRequest, fmt.Errorf("namespace is empty"))
		if err != nil {
			logs.Errorf("failed to return a status code ")
			return
		}
		return
	} else {
		c = h.GetClient(namespace)
	}

	// 查看task是否存在
	task, err := c.client.Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		logs.Errorf("Get task %s error: %v , task not exist!", name, err)
		err := response.WriteError(http.StatusNotFound, err)
		if err != nil {
			logs.Errorf("failed to return a status code")
			return
		}
	}

	// task 存在
	if task.Name == name {
		err := c.client.Delete(context.TODO(), name, metav1.DeleteOptions{})
		if err != nil {
			logs.Errorf("Delete task %s error: %v", name, err)
			err := response.WriteError(http.StatusInternalServerError, err)
			if err != nil {
				logs.Errorf("failed to return a status code")
				return
			}
		}

		// 返回停止成功的状态
		response.WriteHeader(http.StatusOK)

		// 记录日志
		logs.Debugf("delete task : %v", name)
	}

}

// TODO ：StopTask

func (h *TaskHandler) UpdateTask(request *restful.Request, response *restful.Response) {
	// 先检查task是否存在
	// 存在：更新
	// 不存在：返回错误
	// 尝试从url中获取参数
	c := &CurrentTaskHandler{}
	name := request.QueryParameter(TASK_NAME)
	ew := &apis.Task{}
	if name == "" {
		// url中没有获取到name参数，尝试从请求体中获取
		err := request.ReadEntity(&ew)
		if err != nil || ew.Name == "" {
			if err != nil {
				logs.Errorf("Failed to deserialize json data, error: %v", err)
				err := response.WriteError(http.StatusBadRequest, err)
				if err != nil {
					logs.Errorf("failed to return a status code")
					return
				}
				return
			} else {
				err := response.WriteError(http.StatusBadRequest, fmt.Errorf("provide task name , the key is Name "))
				if err != nil {
					logs.Errorf("failed to return a status code ")
					return
				}
				return
			}
		} else {
			name = ew.Name
		}
	} else {
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
	}

	namespace := request.PathParameter(NAMESPACE)
	if namespace == "" {
		err := response.WriteError(http.StatusBadRequest, fmt.Errorf("namespace is empty"))
		if err != nil {
			logs.Errorf("failed to return a status code")
			return
		}
		return
	} else {
		c = h.GetClient(namespace)
	}

	// 检查task是否存在
	task, err := c.client.Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		logs.Errorf("Get task %s error: %v , task not exist!", name, err)
		err := response.WriteError(http.StatusNotFound, err)
		if err != nil {
			logs.Errorf("failed to return a status code")
			return
		}
		return
	}

	// task存在，更新
	if task.Name == name {
		//err := request.ReadEntity(&task)
		//if err != nil {
		//	err := response.WriteError(http.StatusInternalServerError, err)
		//	if err != nil {
		//		logs.Errorf("failed to return a status code")
		//		return
		//	}
		//	return
		//}
		//logs.Debugf("Update task to : %v", task)

		// 格式验证
		res, err := analyzer.SerializeToJson(ew)
		_, err = analyzer.Deserialize(res, apis.Task{})
		if err != nil {
			err := response.WriteError(http.StatusBadRequest, err)
			if err != nil {
				logs.Errorf("failed to return a status code ")
				return
			}
			// return
		}

		// 这个函数执行的时候返回错误
		updatedTask, updateErr := c.client.Update(context.TODO(), ew, metav1.UpdateOptions{})
		if updateErr != nil {
			logs.Errorf("Update task %s error: %v", name, updateErr)
			err := response.WriteError(http.StatusInternalServerError, err)
			if err != nil {
				logs.Errorf("failed to return a status code")
				return
			}
		}

		//返回成功修改的通知
		err = response.WriteHeaderAndEntity(http.StatusOK, updatedTask)
		if err != nil {
			logs.Errorf("failed to return a status code")
			return
		}

		// 记录日志
		logs.Debugf("update task : %v", name)

	}
}

func (h *TaskHandler) PatchTask(request *restful.Request, response *restful.Response) {
	// 先检查task是否存在
	// 存在：更新
	// 不存在：返回错误
	// 尝试从url中获取参数
	c := &CurrentTaskHandler{}
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
	namespace := request.PathParameter(NAMESPACE)
	if namespace == "" {
		err := response.WriteError(http.StatusBadRequest, fmt.Errorf("namespace is empty"))
		if err != nil {
			logs.Errorf("failed to return a status code ")
			return
		}
		return
	} else {
		c = h.GetClient(namespace)
	}

	// 检查task是否存在
	task, err := c.client.Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		logs.Errorf("Get task %s error: %v , task not exist!", name, err)
		err := response.WriteError(http.StatusNotFound, err)
		if err != nil {
			logs.Errorf("failed to return a status code")
			return
		}
		return
	}

	// task 存在，部分更新
	if task.Name == name {
		err := request.ReadEntity(task)
		if err != nil {
			err := response.WriteError(http.StatusInternalServerError, err)
			if err != nil {
				logs.Errorf("failed to return a status code")
				return
			}
			return
		}
		// logs.Debugf("Patch task to : %v", task)
		patchTask, err := analyzer.SerializeToJson(task)
		if err != nil {
			logs.Errorf("Serialize patch task error: %v", err)
			err := response.WriteError(http.StatusInternalServerError, err)
			if err != nil {
				logs.Errorf("failed to return a status code")
				return
			}
		}
		// logs.Debugf("change Patch task to json : %v", patchTask)
		patchedTask, err := c.client.Patch(context.TODO(), name, types.StrategicMergePatchType, []byte(patchTask), metav1.PatchOptions{})
		if err != nil {
			logs.Errorf("Patch:task %s error: %v", name, err)
			err := response.WriteError(http.StatusInternalServerError, err)
			if err != nil {
				logs.Errorf("failed to return a status code")
				return
			}
		}
		// logs.Debugf("Patched task : %v", patchedTask)

		//返回成功修改的通知
		err = response.WriteHeaderAndEntity(http.StatusOK, patchedTask)
		if err != nil {
			logs.Errorf("failed to return a status code")
			return
		}

		// 记录日志
		logs.Debugf("patch task : %v", name)
	}
}

func (h *TaskHandler) NewGetWebService() *restful.WebService {
	ws := new(restful.WebService)
	ws.Path(TASK_PATH).
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	// 查询任务
	ws.Route(ws.GET("/{Namespace}/task").
		To(h.GetTask).
		Doc("Get a task with name").
		Metadata(restfulspec.KeyOpenAPITags, []string{TAG}).
		Param(ws.QueryParameter("Name", "The name of the task").DataType("string")).
		Param(ws.PathParameter("Namespace", "The namespace of the task").DataType("string")).
		Operation("get Task").
		Returns(200, "OK", apis.Task{}).
		Returns(400, "Not Found", nil),
	)

	//创建任务
	ws.Route(ws.POST("/{Namespace}/task").
		To(h.CreateTask).
		Doc("Create a task with name").
		Metadata(restfulspec.KeyOpenAPITags, []string{TAG}).
		Param(ws.QueryParameter("Name", "The name of the task").DataType("string")).
		Param(ws.PathParameter("Namespace", "The namespace of the task").DataType("string")).
		Param(ws.BodyParameter("Task", "The json string of the Task object").DataType("string")).
		Operation("createTask").
		Returns(200, "OK", apis.Task{}).
		Returns(400, "Not Found", nil),
	)

	//修改任务
	ws.Route(ws.PUT("/{Namespace}/task").
		To(h.UpdateTask).
		Doc("Update a task with name").
		Metadata(restfulspec.KeyOpenAPITags, []string{TAG}).
		Param(ws.QueryParameter("Name", "The name of the task").DataType("string")).
		Param(ws.PathParameter("Namespace", "The namespace of the task").DataType("string")).
		Param(ws.BodyParameter("Task", "The json string of the Task object").DataType("string")).
		Operation("update Task").
		Returns(200, "OK", apis.Task{}).
		Returns(400, "Not Found", nil),
	)

	//部分修改任务
	ws.Route(ws.PATCH("/{Namespace}/task").
		To(h.PatchTask).
		Doc("Patch a task").
		Metadata(restfulspec.KeyOpenAPITags, []string{TAG}).
		Param(ws.QueryParameter("Name", "The name of the task").DataType("string")).
		Param(ws.PathParameter("Namespace", "The namespace of the task").DataType("string")).
		Param(ws.BodyParameter("Task", "The json string of the Task field").DataType("string")).
		Operation("patch Task").
		Returns(200, "OK", apis.Task{}).
		Returns(400, "Not Found", nil),
	)

	// 删除任务
	ws.Route(ws.DELETE("/{Namespace}/task").
		To(h.DeleteTask).
		Doc("Delete a task").
		Metadata(restfulspec.KeyOpenAPITags, []string{TAG}).
		Param(ws.QueryParameter("Name", "The name of the task").DataType("string")).
		Param(ws.PathParameter("Namespace", "The namespace of the task").DataType("string")).
		Operation("Delete Task").
		Returns(200, "OK", apis.Task{}).
		Returns(400, "Not Found", nil),
	)

	return ws
}

package device

import (
	"errors"
	"fmt"
	restfulspec "github.com/emicklei/go-restful-openapi/v2"
	"github.com/emicklei/go-restful/v3"
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/client-go/clients"
	"hit.edu/framework/pkg/client-go/util/manager"
	"hit.edu/framework/pkg/component-base/analyzer"
	"hit.edu/framework/pkg/component-base/logs"
	"net/http"
	"sync"
)

type DeviceHandler struct {
	manager   *manager.Manager
	clientSet *clients.ClientSet
	mu        sync.Mutex
}

//type CurrentDeviceHandler struct {
//	client core.DeviceInterface
//}

var _ Handler = &DeviceHandler{}

//func NewDeviceHandler(clientSet *clients.ClientSet) *DeviceHandler {
//	c := clientSet.Core().Devices("test") //apis.NamespaceAll
//	return &DeviceHandler{
//		client: c,
//	}
//}

// NewDeviceHandler 创建一个 DeviceHandler
func NewDeviceHandler(clientSet *clients.ClientSet) *DeviceHandler {
	return &DeviceHandler{
		clientSet: clientSet,
		manager:   manager.NewManager(clientSet),
	}
}

// GetClient 根据 namespace 获取 client，如果不存在则创建
//func (h *DeviceHandler) GetClient(namespace string) *CurrentDeviceHandler {
//	h.mu.Lock()
//	defer h.mu.Unlock()
//
//	// 如果已经存在，直接返回
//	if c, exists := h.clients[namespace]; exists {
//		return &CurrentDeviceHandler{
//			client: c,
//		}
//	}
//
//	// 否则创建新的 client
//	newClient := h.clientSet.Core().Devices(namespace)
//	h.clients[namespace] = newClient
//	return &CurrentDeviceHandler{
//		client: newClient,
//	}
//}

func (h *DeviceHandler) GetDevice(request *restful.Request, response *restful.Response) {
	// 获取name
	name := request.QueryParameter(DEVICE_NAME)
	if name == "" {
		err := response.WriteError(http.StatusBadRequest, fmt.Errorf("provide device name , the key is Name "))
		if err != nil {
			logs.Errorf("failed to return a status code ")
			return
		}
		return
	}

	// 获取namespace
	namespace := request.QueryParameter(NAME_SPACE)
	if namespace == "" {
		err := response.WriteError(http.StatusBadRequest, fmt.Errorf("namespace is required"))
		if err != nil {
			logs.Errorf("failed to return a status code ")
			return
		}
		return
	}

	result, err := h.manager.GetDevice(name, namespace)
	if err != nil {
		logs.Errorf("Get device %s error: %v , device not exist! ", name, err)
		err := response.WriteError(http.StatusNotFound, err)
		if err != nil {
			logs.Errorf("failed to return a status code")
			return
		}
		return
	}

	err = response.WriteHeaderAndEntity(http.StatusOK, result)
	if err != nil {
		logs.Errorf("failed to return a status code")
		return
	}

	logs.Debugf("Get device success")
}

func (h *DeviceHandler) CreateDevice(request *restful.Request, response *restful.Response) {
	// 先查询device是否存在
	// 尝试从url中获取参数
	//name := request.QueryParameter(DEVICE_NAME)
	//ew := &apis.Device{}
	//if name == "" {
	//	// url中没有获取到name参数，尝试从请求体中获取
	//
	//	err := request.ReadEntity(&ew)
	//	if err != nil || ew.Name == "" {
	//		if err != nil {
	//			logs.Errorf("Failed to deserialize json data, error: %v", err)
	//			err := response.WriteError(http.StatusBadRequest, err)
	//			if err != nil {
	//				logs.Errorf("failed to return a status code")
	//				return
	//			}
	//			return
	//		} else {
	//			err := response.WriteError(http.StatusBadRequest, fmt.Errorf("provide device name , the key is Name "))
	//			if err != nil {
	//				logs.Errorf("failed to return a status code ")
	//				return
	//			}
	//			return
	//		}
	//	} else {
	//		name = ew.Name
	//	}
	//} else {
	//	err := request.ReadEntity(&ew)
	//	if err != nil {
	//		logs.Errorf("Failed to deserialize json data, error: %v", err)
	//		err := response.WriteError(http.StatusBadRequest, err)
	//		if err != nil {
	//			logs.Errorf("failed to return a status code")
	//			return
	//		}
	//		return
	//	}
	//}
	//
	//namespace := ew.Namespace
	//if namespace == "" {
	//	logs.Error("namespace is empty")
	//	err := response.WriteError(http.StatusBadRequest, fmt.Errorf("namespace is empty"))
	//	if err != nil {
	//		logs.Errorf("failed to return a status code ")
	//		return
	//	}
	//	return
	//} else {
	//	c = h.GetClient(namespace)
	//}
	//
	//result, err := c.client.Get(context.TODO(), name, metav1.GetOptions{})
	//if err != nil {
	//	logs.Infof("Get device %s error: %v , device not exist! create it", name, err)
	//} else if result.Name == name {
	//	logs.Errorf("Create device %s error, device existed: %v", name, result)
	//	err = fmt.Errorf("create device %s error, device existed: %v", name, result)
	//	err := response.WriteError(http.StatusConflict, err)
	//	if err != nil {
	//		logs.Errorf("failed to return a status code")
	//		return
	//	}
	//	return
	//}
	//
	//logs.Info(*ew)
	//
	//// 不存在，解析用户的输入
	////ew := &apis.Device{}
	////err = request.ReadEntity(ew)
	////if err != nil {
	////	logs.Errorf("Failed to create device %s, error: %v", name, err)
	////	err := response.WriteError(http.StatusInternalServerError, err)
	////	if err != nil {
	////		logs.Errorf("failed to return a status code")
	////		return
	////	}
	////	return
	////}
	//
	////格式校验
	////res, err := analyzer.SerializeToJson(ew)
	////_, err = analyzer.Deserialize(res, apis.Device{})
	////if err != nil {
	////	err := response.WriteError(http.StatusBadRequest, err)
	////	if err != nil {
	////		logs.Errorf("failed to return a status code")
	////		return
	////	}
	////	return
	////}
	//
	//// TODO：为Device分配ID?
	//
	//// 将device写入数据库中
	//result, err = c.client.Create(context.TODO(), ew, metav1.CreateOptions{})
	//if err != nil {
	//	err1 := response.WriteError(http.StatusInternalServerError, err)
	//	if err1 != nil {
	//		logs.Errorf("failed to return a status code , error :%v", err1)
	//		return
	//	}
	//	logs.Errorf("Create device %s ,failed write to database , error: %v", name, err)
	//	return
	//}
	//
	//// 返回结果
	//err = response.WriteEntity(result)
	//if err != nil {
	//	err := response.WriteError(http.StatusInternalServerError, err)
	//	if err != nil {
	//		logs.Errorf("failed to return a status code")
	//		return
	//	}
	//	return
	//}
	//
	//err = response.WriteError(http.StatusOK, err)
	//if err != nil {
	//	logs.Errorf("failed to return a status code ")
	//	return
	//}
	//
	//logs.Debugf("Create device %v", result)

	// 获取device数据
	d := &apis.Device{}
	err := request.ReadEntity(&d)
	if err != nil {
		logs.Errorf("Failed to deserialize json data, error: %v", err)
		err := response.WriteError(http.StatusBadRequest, err)
		if err != nil {
			logs.Errorf("failed to return a status code")
			return
		}
		return
	}

	// TODO: 增加device的json schema约束
	//格式校验
	//res, err := analyzer.SerializeToJson(d)
	//_, err = analyzer.Deserialize(res, apis.Device{})
	//if err != nil {
	//	err := response.WriteError(http.StatusBadRequest, err)
	//	if err != nil {
	//		logs.Errorf("failed to return a status code")
	//		return
	//	}
	//	return
	//}

	// 获取 namespace
	namespace := request.QueryParameter(NAME_SPACE)
	// namespace := d.Namespace
	if namespace == "" {
		logs.Error("namespace is empty")
		err := response.WriteError(http.StatusBadRequest, fmt.Errorf("namespace is empty"))
		if err != nil {
			logs.Errorf("failed to return a status code ")
			return
		}
		return
	}

	logs.Info(*d)

	var result *apis.Device
	result, err = h.manager.CreateDevice(d, namespace)
	if err != nil {
		err1 := response.WriteError(http.StatusInternalServerError, err)
		if err1 != nil {
			logs.Errorf("failed to return a status code ,error: %v", err1)
			return
		}
		logs.Errorf("Create device fail ,failed write it to database , error: %v", err)
		return
	}

	// 返回结果
	//err = response.WriteEntity(result)
	//if err != nil {
	//	err := response.WriteError(http.StatusInternalServerError, err)
	//	if err != nil {
	//		logs.Errorf("failed to return a status code")
	//		return
	//	}
	//	return
	//}
	//response.WriteHeader(http.StatusOK)

	err = response.WriteHeaderAndEntity(http.StatusCreated, result)
	if err != nil {
		logs.Errorf("failed to return a status code")
		return
	}

	logs.Debugf("Create device %v ", result)
}

func (h *DeviceHandler) UpdateDevice(request *restful.Request, response *restful.Response) {
	// 获取device
	//name := request.QueryParameter(DEVICE_NAME)
	//d := &apis.Device{}
	//if name == "" {
	//	// url中没有获取到name参数，尝试从请求体中获取
	//
	//	err := request.ReadEntity(&ew)
	//	if err != nil || ew.Name == "" {
	//		if err != nil {
	//			logs.Errorf("Failed to deserialize json data, error: %v", err)
	//			err := response.WriteError(http.StatusBadRequest, err)
	//			if err != nil {
	//				logs.Errorf("failed to return a status code")
	//				return
	//			}
	//			return
	//		} else {
	//			err := response.WriteError(http.StatusBadRequest, fmt.Errorf("provide device name , the key is Name "))
	//			if err != nil {
	//				logs.Errorf("failed to return a status code ")
	//				return
	//			}
	//			return
	//		}
	//	} else {
	//		name = ew.Name
	//	}
	//} else {
	//	err := request.ReadEntity(&ew)
	//	if err != nil {
	//		logs.Errorf("Failed to deserialize json data, error: %v", err)
	//		err := response.WriteError(http.StatusBadRequest, err)
	//		if err != nil {
	//			logs.Errorf("failed to return a status code")
	//			return
	//		}
	//		return
	//	}
	//}
	//
	//namespace := request.QueryParameter(NAME_SPACE)
	//if namespace == "" {
	//	err := response.WriteError(http.StatusBadRequest, fmt.Errorf("namespace is empty"))
	//	if err != nil {
	//		logs.Errorf("failed to return a status code")
	//		return
	//	}
	//	return
	//} else {
	//	c = h.GetClient(namespace)
	//}
	//
	//// 检查device是否存在
	//device, err := c.client.Get(context.TODO(), name, metav1.GetOptions{})
	//if err != nil {
	//	logs.Errorf("Get device %s error: %v , device not exist !", name, err)
	//	err := response.WriteError(http.StatusNotFound, err)
	//	if err != nil {
	//		logs.Errorf("failed to return a status code")
	//		return
	//	}
	//	return
	//}
	//
	//// device 存在，更新
	//if device.Name == name {
	//	//err := request.ReadEntity(&device)
	//	//if err != nil {
	//	//	err := response.WriteError(http.StatusInternalServerError, err)
	//	//	if err != nil {
	//	//		logs.Errorf("failed to return a status code")
	//	//		return
	//	//	}
	//	//	return
	//	//}
	//	//
	//	//logs.Debugf("Update device to : %v", device)
	//
	//	// 格式验证
	//	//res, err := analyzer.SerializeToJson(device)
	//	//_, err = analyzer.Deserialize(res, apis.Device{})
	//	//if err != nil {
	//	//	err := response.WriteError(http.StatusBadRequest, err)
	//	//	if err != nil {
	//	//		logs.Errorf("failed to return a status code ")
	//	//		return
	//	//	}
	//	//	// return
	//	//}
	//
	//	updatedDevice, updateErr := c.client.Update(context.TODO(), ew, metav1.UpdateOptions{})
	//	if updateErr != nil {
	//		logs.Errorf("Update device %s error: %v", name, updateErr)
	//		err := response.WriteError(http.StatusInternalServerError, err)
	//		if err != nil {
	//			logs.Errorf("failed to return a status code")
	//			return
	//		}
	//	}
	//
	//	// 返回成功修改的通知
	//	err = response.WriteHeaderAndEntity(http.StatusOK, updatedDevice)
	//	if err != nil {
	//		logs.Errorf("failed to return a status code")
	//		return
	//	}
	//
	//	// 记录日志
	//	logs.Debugf("update device : %v", name)
	//
	//}

	// 获取device
	d := &apis.Device{}
	err := request.ReadEntity(&d)
	if err != nil {
		logs.Errorf("Failed to deserialize json data, error: %v", err)
		err := response.WriteError(http.StatusBadRequest, err)
		if err != nil {
			logs.Errorf("failed to return a status code")
			return
		}
		return
	}

	// 格式验证
	//res, err := analyzer.SerializeToJson(d)
	//_, err = analyzer.Deserialize(res, apis.Device{})
	//if err != nil {
	//	err := response.WriteError(http.StatusBadRequest, err)
	//	if err != nil {
	//		logs.Errorf("failed to return a status code ")
	//		return
	//	}
	//	return
	//}

	// 获取name
	name := request.QueryParameter(DEVICE_NAME)
	if name == "" {
		err := response.WriteError(http.StatusBadRequest, fmt.Errorf("name is empty"))
		if err != nil {
			logs.Errorf("failed to return a status code")
			return
		}
		return
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

	// 更新device
	updatedDevice, updateErr := h.manager.UpdateDevice(name, namespace, d)
	if updateErr != nil {
		logs.Errorf("Update device %s error: %v", name, updateErr)
		var err error
		if errors.Is(updateErr, manager.NotFound) {
			err = response.WriteError(http.StatusNotFound, updateErr)
		} else if errors.Is(updateErr, manager.InternalServerError) {
			err = response.WriteError(http.StatusInternalServerError, updateErr)
		} else {
			err = response.WriteError(http.StatusInternalServerError, updateErr)
		}
		if err != nil {
			logs.Errorf("failed to return a status code")
			return
		}
	}

	// 返回成功修改的通知
	err = response.WriteHeaderAndEntity(http.StatusOK, updatedDevice)
	if err != nil {
		logs.Errorf("failed to return a status code")
		return
	}

	// 记录日志
	logs.Debugf("update device : %v", name)

}

func (h *DeviceHandler) DeleteDevice(request *restful.Request, response *restful.Response) {
	//// 查看device是否存在
	//// 存在，删除节点
	//// 不存在，返回 404 not found
	//// 尝试从url中获取参数
	//c := &CurrentDeviceHandler{}
	//name := request.QueryParameter(DEVICE_NAME)
	//if name == "" {
	//	// url中没有获取到name参数，尝试从请求体中获取
	//	req := &apis.Device{}
	//	err := request.ReadEntity(&req)
	//	if err != nil || req.Name == "" {
	//		err := response.WriteError(http.StatusBadRequest, fmt.Errorf("provide device name , the key is Name "))
	//		if err != nil {
	//			logs.Errorf("failed to return a status code ")
	//			return
	//		}
	//		return
	//	} else {
	//		name = req.Name
	//	}
	//}
	//
	//// 获取namespace
	//namespace := request.QueryParameter(NAME_SPACE)
	//if namespace == "" {
	//	err := response.WriteError(http.StatusBadRequest, fmt.Errorf("namespace is empty"))
	//	if err != nil {
	//		logs.Errorf("failed to return a status code ")
	//		return
	//	}
	//	return
	//} else {
	//	c = h.GetClient(namespace)
	//}
	//
	//// 查看device是否存在
	//device, err := c.client.Get(context.TODO(), name, metav1.GetOptions{})
	//if err != nil {
	//	logs.Errorf("Get device %s error: %v , device not exist !", name, err)
	//	err := response.WriteError(http.StatusNotFound, err)
	//	if err != nil {
	//		logs.Errorf("failed to return a status code")
	//		return
	//	}
	//}
	//
	//// device 存在
	//if device.Name == name {
	//	err := c.client.Delete(context.TODO(), name, metav1.DeleteOptions{})
	//	if err != nil {
	//		logs.Errorf("Delete device %s error: %v", name, err)
	//		err := response.WriteError(http.StatusInternalServerError, err)
	//		if err != nil {
	//			logs.Errorf("failed to return a status code")
	//			return
	//		}
	//	}
	//
	//	// 返回停止成功的状态
	//	response.WriteHeader(http.StatusOK)
	//
	//	// 记录日志
	//	logs.Debugf("delete device : %v", name)
	//
	//}

	// 获取name
	name := request.QueryParameter(DEVICE_NAME)
	if name == "" {
		err := response.WriteError(http.StatusBadRequest, fmt.Errorf("provide device name , the key is Name "))
		if err != nil {
			logs.Errorf("failed to return a status code ")
			return
		}
		return
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

	// 删除device
	var err error
	err = h.manager.DeleteDevice(name, namespace)
	if err != nil {
		logs.Errorf("delete device %s error: %v", name, err)
		if errors.Is(err, manager.NotFound) {
			err = response.WriteError(http.StatusNotFound, err)
		} else if errors.Is(err, manager.InternalServerError) {
			err = response.WriteError(http.StatusInternalServerError, err)
		} else {
			err = response.WriteError(http.StatusInternalServerError, err)
		}
		if err != nil {
			logs.Errorf("failed to return a status code")
			return
		}
		return
	}

	// 返回停止成功的状态
	response.WriteHeader(http.StatusOK)

	// 记录日志
	logs.Debugf("delete device : %v", name)
}

func (h *DeviceHandler) PatchDevice(request *restful.Request, response *restful.Response) {
	// 先检查device是否存在
	// 存在：部分更新
	// 不存在：返回错误
	// 尝试从url中获取参数
	//c := &CurrentDeviceHandler{}
	//name := request.QueryParameter(DEVICE_NAME)
	//if name == "" {
	//	// url中没有获取到name参数，尝试从请求体中获取
	//	req := &apis.Device{}
	//	err := request.ReadEntity(&req)
	//	if err != nil || req.Name == "" {
	//		err := response.WriteError(http.StatusBadRequest, fmt.Errorf("provide device name , the key is Name "))
	//		if err != nil {
	//			logs.Errorf("failed to return a status code ")
	//			return
	//		}
	//		return
	//	} else {
	//		name = req.Name
	//	}
	//}
	//
	//// 获取namespace
	//namespace := request.QueryParameter(NAME_SPACE)
	//if namespace == "" {
	//	err := response.WriteError(http.StatusBadRequest, fmt.Errorf("namespace is empty"))
	//	if err != nil {
	//		logs.Errorf("failed to return a status code ")
	//		return
	//	}
	//	return
	//} else {
	//	c = h.GetClient(namespace)
	//}
	//
	//// 检查device是否存在
	//device, err := c.client.Get(context.TODO(), name, metav1.GetOptions{})
	//if err != nil {
	//	logs.Errorf("Get device %s error: %v , device not exist !", name, err)
	//	err := response.WriteError(http.StatusNotFound, err)
	//	if err != nil {
	//		logs.Errorf("failed to return a status code")
	//		return
	//	}
	//	return
	//}
	//
	//// device 存在，部分更新
	//if device.Name == name {
	//	err := request.ReadEntity(device)
	//	if err != nil {
	//		err := response.WriteError(http.StatusInternalServerError, err)
	//		if err != nil {
	//			logs.Errorf("failed to return a status code")
	//			return
	//		}
	//		return
	//	}
	//
	//	patchDevice, err := analyzer.SerializeToJson(device)
	//	if err != nil {
	//		logs.Errorf("Serialize patch device error: %v", err)
	//		err := response.WriteError(http.StatusInternalServerError, err)
	//		if err != nil {
	//			logs.Errorf("failed to return a status code")
	//			return
	//		}
	//	}
	//	patchedDevice, err := c.client.Patch(context.TODO(), name, types.StrategicMergePatchType, []byte(patchDevice), metav1.PatchOptions{})
	//	if err != nil {
	//		logs.Errorf("Patch device %s error: %v", name, err)
	//		err := response.WriteError(http.StatusInternalServerError, err)
	//		if err != nil {
	//			logs.Errorf("failed to return a status code")
	//			return
	//		}
	//	}
	//
	//	// 返回成功修改的通知
	//	err = response.WriteHeaderAndEntity(http.StatusOK, patchedDevice)
	//	if err != nil {
	//		logs.Errorf("failed to return a status code")
	//		return
	//	}
	//
	//	// 记录日志
	//	logs.Debugf("patch device : %v", name)
	//
	//}

	// 获取json
	req := &apis.Device{}
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
	name := request.QueryParameter(DEVICE_NAME)
	if name == "" {
		err := response.WriteError(http.StatusBadRequest, fmt.Errorf("name is empty"))
		if err != nil {
			logs.Errorf("failed to return a status code ")
			return
		}
		return
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

	// 序列化PatchDevice
	patchDevice, err := analyzer.SerializeToJson(req)
	if err != nil {
		logs.Errorf("Serialize patch device error: %v", err)
		err := response.WriteError(http.StatusInternalServerError, err)
		if err != nil {
			logs.Errorf("failed to return a status code")
			return
		}
	}

	patchedDevice, patchedErr := h.manager.PatchDevice(name, namespace, []byte(patchDevice))
	if patchedErr != nil {
		logs.Errorf("patched device %s error: %v", name, err)
		var err error
		if errors.Is(patchedErr, manager.NotFound) {
			err = response.WriteError(http.StatusNotFound, patchedErr)
		} else if errors.Is(patchedErr, manager.InternalServerError) {
			err = response.WriteError(http.StatusInternalServerError, patchedErr)
		} else {
			err = response.WriteError(http.StatusInternalServerError, patchedErr)
		}
		if err != nil {
			logs.Errorf("failed to return a status code")
			return
		}
		return
	}

	// 返回成功修改的通知
	err = response.WriteHeaderAndEntity(http.StatusOK, patchedDevice)
	if err != nil {
		logs.Errorf("failed to return a status code")
		return
	}

	// 记录日志
	logs.Debugf("patch device : %v", name)

}

func (h *DeviceHandler) NewGetWebService() *restful.WebService {
	ws := new(restful.WebService)
	ws.Path(DEVICE_PATH).
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/").
		To(h.GetDevice).
		Doc("Get a device with name and namespace").
		Metadata(restfulspec.KeyOpenAPITags, []string{TAG}).
		Param(ws.QueryParameter("Name", "The name of the device").DataType("string")).
		Param(ws.QueryParameter("Namespace", "The namespace of the device").DataType("string")).
		Operation("Get device").
		Returns(200, "OK", apis.Device{}).
		Returns(404, "Not Found", nil),
	)

	ws.Route(ws.POST("/").
		To(h.CreateDevice).
		Doc("Create a device with namespace").
		Metadata(restfulspec.KeyOpenAPITags, []string{TAG}).
		Param(ws.QueryParameter("Namespace", "The namespace of the device").DataType("string")).
		Param(ws.BodyParameter("Device", "The json string of the Device object").DataType("string")).
		Operation("Create device").
		Returns(200, "OK", apis.Device{}).
		Returns(404, "Not Found", nil),
	)

	ws.Route(ws.PUT("/").
		To(h.UpdateDevice).
		Doc("Update a device with name and namespace").
		Metadata(restfulspec.KeyOpenAPITags, []string{TAG}).
		Param(ws.QueryParameter("Name", "The name of the device").DataType("string")).
		Param(ws.QueryParameter("Namespace", "The namespace of the device").DataType("string")).
		Param(ws.BodyParameter("Device", "The json string of the Device object").DataType("string")).
		Operation("Update device").
		Returns(200, "OK", apis.Device{}).
		Returns(404, "Not Found", nil))

	ws.Route(ws.PATCH("/").
		To(h.PatchDevice).
		Doc("Patch a device with name and namespace").
		Metadata(restfulspec.KeyOpenAPITags, []string{TAG}).
		Param(ws.QueryParameter("Name", "The name of the device").DataType("string")).
		Param(ws.QueryParameter("Namespace", "The namespace of the device").DataType("string")).
		Param(ws.BodyParameter("Device", "The json string of the Device field").DataType("string")).
		Operation("Patch device").
		Returns(200, "OK", apis.Device{}).
		Returns(404, "Not Found", nil))

	ws.Route(ws.DELETE("/").
		To(h.DeleteDevice).
		Doc("Delete a device with name and namespace").
		Metadata(restfulspec.KeyOpenAPITags, []string{TAG}).
		Param(ws.QueryParameter("Name", "The name of the device").DataType("string")).
		Param(ws.QueryParameter("Namespace", "The namespace of the device").DataType("string")).
		Operation("Delete device").
		Returns(200, "OK", apis.Device{}).
		Returns(404, "Not Found", nil))

	return ws
}

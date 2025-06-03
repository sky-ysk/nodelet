package node

import (
	"errors"
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

type NodeHandler struct {
	manager   *manager.Manager
	clientSet *clients.ClientSet
	mu        sync.Mutex
}

//type CurrentNodeHandler struct {
//	client core.NodeInterface
//}

var _ Handler = &NodeHandler{}

//func NewNodeHandler(clientSet *clients.ClientSet) *NodeHandler {
//	c := clientSet.Core().Nodes(apis.NamespaceAll) //"test"
//	return &NodeHandler{
//		client: c,
//	}
//}

// NewNodeHandler 创建一个 NodeHandler
func NewNodeHandler(clientSet *clients.ClientSet) *NodeHandler {
	return &NodeHandler{
		clientSet: clientSet,
		manager:   manager.NewManager(clientSet),
	}
}

// GetClient 根据 namespace 获取 client，如果不存在则创建
//func (h *NodeHandler) GetClient(namespace string) *CurrentNodeHandler {
//	h.mu.Lock()
//	defer h.mu.Unlock()
//
//	// 如果已经存在，直接返回
//	if c, exists := h.clients[namespace]; exists {
//		return &CurrentNodeHandler{
//			client: c,
//		}
//	}
//
//	// 否则创建新的 client
//	newClient := h.clientSet.Core().Nodes(namespace)
//	h.clients[namespace] = newClient
//	return &CurrentNodeHandler{
//		client: newClient,
//	}
//}

func (h *NodeHandler) GetNode(request *restful.Request, response *restful.Response) {
	// 尝试从url中获取参数
	name := request.QueryParameter(NODE_NAME)
	if name == "" {
		// url中没有获取到name参数，尝试从请求体中获取
		req := &apis.Node{}
		err := request.ReadEntity(&req)
		if err != nil || req.Name == "" {
			err := response.WriteError(http.StatusBadRequest, fmt.Errorf("provide node name , the key is Name "))
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

	result, err := h.manager.GetNode(name, namespace)
	if err != nil {
		logs.Errorf("Get node %s error: %v , node not exist! ", name, err)
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
		logs.Debugf("Get node")
	}
}

func (h *NodeHandler) CreateNode(request *restful.Request, response *restful.Response) {
	// 先查询Node是否存在
	// 尝试从url中获取参数
	//c := &CurrentNodeHandler{}
	//name := request.QueryParameter(NODE_NAME)
	//ew := &apis.Node{}
	//if name == "" {
	//	// url中没有获取到name参数，尝试从请求体中获取
	//	// 解析用户输入
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
	//			err := response.WriteError(http.StatusBadRequest, fmt.Errorf("provide node name , the key is Name "))
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
	//	logs.Infof("Get node %s error: %v , node not exist ! creat it ", name, err)
	//} else if result.Name == name {
	//	logs.Errorf("Create node %s error, node existed: %v", name, result)
	//	err = fmt.Errorf("create node %s error, node existed: %v", name, result)
	//	err := response.WriteError(http.StatusConflict, err)
	//	if err != nil {
	//		logs.Errorf("failed to return a status code")
	//		return
	//	}
	//	return
	//}
	//
	//// 不存在，解析用户的输入
	////err = request.ReadEntity(&ew)
	////if err != nil {
	////	logs.Errorf("Failed to create node %s, error: %v", name, err)
	////	err := response.WriteError(http.StatusInternalServerError, err)
	////	if err != nil {
	////		logs.Errorf("failed to return a status code")
	////		return
	////	}
	////	return
	////}
	//
	//logs.Info(*ew)
	//
	//// 格式校验
	//res, err := analyzer.SerializeToJson(ew)
	//_, err = analyzer.Deserialize(res, apis.Node{})
	//if err != nil {
	//	err := response.WriteError(http.StatusBadRequest, err)
	//	if err != nil {
	//		logs.Errorf("failed to return a status code")
	//		return
	//	}
	//	return
	//}
	//
	//// TODO：为NODE分配MachineID?
	//
	//// 将Node写入数据库中
	//result, err = c.client.Create(context.TODO(), ew, metav1.CreateOptions{})
	//if err != nil {
	//	err1 := response.WriteError(http.StatusInternalServerError, err)
	//	if err1 != nil {
	//		logs.Errorf("failed to return a status code ,error %v", err1)
	//		return
	//	}
	//	logs.Errorf("Create node %s ，failed write to database ,error: %v", name, err)
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
	//logs.Debugf("Create node %v", result)

	// 获取node数据
	n := &apis.Node{}
	err := request.ReadEntity(&n)
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
	namespace := n.Namespace
	if namespace == "" {
		logs.Error("namespace is empty")
		err := response.WriteError(http.StatusBadRequest, fmt.Errorf("namespace is empty"))
		if err != nil {
			logs.Errorf("failed to return a status code ")
			return
		}
		return
	}

	logs.Info(*n)

	//格式校验
	res, err := analyzer.SerializeToJson(n)
	_, err = analyzer.Deserialize(res, apis.Node{})
	if err != nil {
		err := response.WriteError(http.StatusBadRequest, err)
		if err != nil {
			logs.Errorf("failed to return a status code")
			return
		}
		return
	}

	// 产生UUID
	timestamp := time.Now().Format("20060102T150405")
	randomStr := uuid.New().String()[:5]
	UUID := timestamp + "-" + randomStr

	var result *apis.Node

	// 创建node without labels
	result, err = h.manager.CreateNode(n.Spec, namespace, UUID)
	if err != nil {
		err1 := response.WriteError(http.StatusInternalServerError, err)
		if err1 != nil {
			logs.Errorf("failed to return a status code ,error: %v", err1)
			return
		}
		logs.Errorf("Create node fail ,failed write it to database , error: %v", err)
		return
	}

	// 返回结果
	err = response.WriteEntity(result)
	if err != nil {
		err := response.WriteError(http.StatusInternalServerError, err)
		if err != nil {
			logs.Errorf("failed to return a status code")
			return
		}
		return
	}

	err = response.WriteError(http.StatusOK, err)
	if err != nil {
		logs.Errorf("failed to return a status code ")
		return
	}

	logs.Debugf("Create node %v unsupport", result)
}

func (h *NodeHandler) UpdateNode(request *restful.Request, response *restful.Response) {
	// 先检查Node是否存在
	// 存在：更新
	// 不存在：返回错误
	//c := &CurrentNodeHandler{}
	//name := request.QueryParameter(NODE_NAME)
	//ew := &apis.Node{}
	//if name == "" {
	//	// url中没有获取到name参数，尝试从请求体中获取
	//	// 解析用户输入
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
	//			err := response.WriteError(http.StatusBadRequest, fmt.Errorf("provide node name , the key is Name "))
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
	//// 检查Node是否存在
	//node, err := c.client.Get(context.TODO(), name, metav1.GetOptions{})
	//if err != nil {
	//	logs.Errorf("Get node %s error: %v , node not exist! ", name, err)
	//	err := response.WriteError(http.StatusNotFound, err)
	//	if err != nil {
	//		logs.Errorf("failed to return a status code")
	//		return
	//	}
	//	return
	//}
	//
	//// Node 存在，更新
	//if node.Name == name {
	//	//err := request.ReadEntity(&node)
	//	//if err != nil {
	//	//	err := response.WriteError(http.StatusInternalServerError, err)
	//	//	if err != nil {
	//	//		logs.Errorf("failed to return a status code")
	//	//		return
	//	//	}
	//	//	return
	//	//}
	//
	//	// 格式验证
	//	res, err := analyzer.SerializeToJson(ew)
	//	_, err = analyzer.Deserialize(res, apis.Node{})
	//	if err != nil {
	//		err := response.WriteError(http.StatusBadRequest, err)
	//		if err != nil {
	//			logs.Errorf("failed to return a status code ")
	//			return
	//		}
	//		// return
	//	}
	//
	//	updatedNode, updateErr := c.client.Update(context.TODO(), ew, metav1.UpdateOptions{})
	//	if updateErr != nil {
	//		logs.Errorf("Update node %s error: %v", name, updateErr)
	//		err := response.WriteError(http.StatusInternalServerError, err)
	//		if err != nil {
	//			logs.Errorf("failed to return a status code")
	//			return
	//		}
	//	}
	//
	//	//返回成功修改的通知
	//	err = response.WriteHeaderAndEntity(http.StatusOK, updatedNode)
	//	if err != nil {
	//		logs.Errorf("failed to return a status code")
	//		return
	//	}
	//
	//	// 记录日志
	//	logs.Debugf("update node : %v", name)
	//
	//}

	// 获取node
	ew := &apis.Node{}
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
	name := request.QueryParameter(NODE_NAME)
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
	_, err = analyzer.Deserialize(res, apis.Node{})
	if err != nil {
		err := response.WriteError(http.StatusBadRequest, err)
		if err != nil {
			logs.Errorf("failed to return a status code ")
			return
		}
		return
	}

	// 更新node
	updatedNode, updateErr := h.manager.UpdateNode(name, namespace, ew)
	if updateErr != nil {
		logs.Errorf("Update node %s error: %v", name, updateErr)
		var err error
		if errors.Is(updateErr, manager.NotFound) {
			err = response.WriteError(http.StatusNotFound, err)
		} else {
			err = response.WriteError(http.StatusInternalServerError, err)
		}
		if err != nil {
			logs.Errorf("failed to return a status code")
			return
		}
	}

	// 返回成功修改的通知
	err = response.WriteHeaderAndEntity(http.StatusOK, updatedNode)
	if err != nil {
		logs.Errorf("failed to return a status code")
		return
	}

	// 记录日志
	logs.Debugf("update node : %v", name)
}

func (h *NodeHandler) DeleteNode(request *restful.Request, response *restful.Response) {
	// 查看Node是否存在
	// 如果存在，删除节点
	// 如果不存在，返回 404 not found
	//c := &CurrentNodeHandler{}
	//name := request.QueryParameter(NODE_NAME)
	//if name == "" {
	//	// url中没有获取到name参数，尝试从请求体中获取
	//	req := &apis.Node{}
	//	err := request.ReadEntity(&req)
	//	if err != nil || req.Name == "" {
	//		err := response.WriteError(http.StatusBadRequest, fmt.Errorf("provide node name , the key is Name "))
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
	//// 查看node是否存在
	//node, err := c.client.Get(context.TODO(), name, metav1.GetOptions{})
	//if err != nil {
	//	logs.Errorf("Get node %s error: %v , node not exist!", name, err)
	//	err := response.WriteError(http.StatusNotFound, err)
	//	if err != nil {
	//		logs.Errorf("failed to return a status code")
	//		return
	//	}
	//}
	//
	//// node 存在
	//if node.Name == name {
	//	err := c.client.Delete(context.TODO(), name, metav1.DeleteOptions{})
	//	if err != nil {
	//		logs.Errorf("Delete node %s error: %v", name, err)
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
	//	logs.Debugf("delete node : %v", name)
	//}

	// 尝试从url中获取参数
	name := request.QueryParameter(NODE_NAME)
	if name == "" {
		// url中没有获取到name参数，尝试从请求体中获取
		req := &apis.Node{}
		err := request.ReadEntity(&req)
		if err != nil || req.Name == "" {
			err := response.WriteError(http.StatusBadRequest, fmt.Errorf("provide node name , the key is Name "))
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

	// 删除node
	var err error
	err = h.manager.DeleteNode(namespace, name)
	if err != nil {
		logs.Errorf("delete node %s error: %v", name, err)
		if errors.Is(err, manager.NotFound) {
			err = response.WriteError(http.StatusNotFound, err)
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
	logs.Debugf("delete node : %v", name)
}

func (h *NodeHandler) PatchNode(request *restful.Request, response *restful.Response) {
	// 先检查Node是否存在
	// 存在：部分更新
	// 不存在：返回错误
	//c := &CurrentNodeHandler{}
	//name := request.QueryParameter(NODE_NAME)
	//if name == "" {
	//	// url中没有获取到name参数，尝试从请求体中获取
	//	req := &apis.Node{}
	//	err := request.ReadEntity(&req)
	//	if err != nil || req.Name == "" {
	//		err := response.WriteError(http.StatusBadRequest, fmt.Errorf("provide node name , the key is Name "))
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
	//// 检查Node是否存在
	//node, err := c.client.Get(context.TODO(), name, metav1.GetOptions{})
	//if err != nil {
	//	logs.Errorf("Get node %s error: %v , node not exist! ", name, err)
	//	err := response.WriteError(http.StatusNotFound, err)
	//	if err != nil {
	//		logs.Errorf("failed to return a status code")
	//		return
	//	}
	//	return
	//}
	//
	//// Node 存在，部分更新
	//if node.Name == name {
	//	err := request.ReadEntity(node)
	//	if err != nil {
	//		err := response.WriteError(http.StatusInternalServerError, err)
	//		if err != nil {
	//			logs.Errorf("failed to return a status code")
	//			return
	//		}
	//		return
	//	}
	//
	//	patchNode, err := analyzer.SerializeToJson(node)
	//	if err != nil {
	//		logs.Errorf("Serialize patch node error: %v", err)
	//		err := response.WriteError(http.StatusInternalServerError, err)
	//		if err != nil {
	//			logs.Errorf("failed to return a status code")
	//			return
	//		}
	//	}
	//	logs.Debugf("patch node : %v", patchNode)
	//	patchedNode, err := c.client.Patch(context.TODO(), name, types.StrategicMergePatchType, []byte(patchNode), metav1.PatchOptions{})
	//	if err != nil {
	//		logs.Errorf("Patch node %s error: %v", name, err)
	//		err := response.WriteError(http.StatusInternalServerError, err)
	//		if err != nil {
	//			logs.Errorf("failed to return a status code")
	//			return
	//		}
	//	}
	//
	//	//返回成功修改的通知
	//	err = response.WriteHeaderAndEntity(http.StatusOK, patchedNode)
	//	if err != nil {
	//		logs.Errorf("failed to return a status code")
	//		return
	//	}
	//
	//	// 记录日志
	//	logs.Debugf("update node : %v", name)
	//
	//}

	// 获取json
	req := &apis.Node{}
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
	name := request.QueryParameter(NODE_NAME)
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

	// 序列化PatchNode
	patchNode, err := analyzer.SerializeToJson(req)
	if err != nil {
		logs.Errorf("Serialize patch node error: %v", err)
		err := response.WriteError(http.StatusInternalServerError, err)
		if err != nil {
			logs.Errorf("failed to return a status code")
			return
		}
	}

	patchedNode, patchedErr := h.manager.PatchNode(namespace, name, []byte(patchNode))
	if patchedErr != nil {
		logs.Errorf("patched Node %s error: %v", name, err)
		var err error
		if errors.Is(patchedErr, manager.NotFound) {
			err = response.WriteError(http.StatusNotFound, err)
		} else {
			err = response.WriteError(http.StatusInternalServerError, err)
		}
		if err != nil {
			logs.Errorf("failed to return a status code")
			return
		}
		return
	}

	// 返回成功修改的通知
	err = response.WriteHeaderAndEntity(http.StatusOK, patchedNode)
	if err != nil {
		logs.Errorf("failed to return a status code")
		return
	}

	// 记录日志
	logs.Debugf("patch node : %v", name)
}

func (h *NodeHandler) NewGetWebService() *restful.WebService {
	ws := new(restful.WebService)
	ws.Path(NODE_PATH).
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET(fmt.Sprint("/")).
		To(h.GetNode).
		Doc("Get a node with name").
		Metadata(restfulspec.KeyOpenAPITags, []string{TAG}).
		Param(ws.QueryParameter("Name", "The name of the node").DataType("string")).
		Param(ws.QueryParameter("Namespace", "The namespace of the node").DataType("string")).
		Operation("Get node").
		Returns(200, "OK", apis.Node{}).
		Returns(400, "Not Found", nil),
	)

	ws.Route(ws.POST(fmt.Sprint("/")).
		To(h.CreateNode).
		Doc("Create a node with name").
		Metadata(restfulspec.KeyOpenAPITags, []string{TAG}).
		Operation("Create node").
		Param(ws.QueryParameter("Name", "The name of the node").DataType("string")).
		Param(ws.QueryParameter("Namespace", "The namespace of the node").DataType("string")).
		Param(ws.BodyParameter("Node", "The json string of the Node object").DataType("string")).
		Returns(200, "OK", apis.Node{}).
		Returns(400, "Not Found", nil),
	)

	ws.Route(ws.PUT(fmt.Sprintf("/")).
		To(h.UpdateNode).
		Doc("Update a node with name").
		Metadata(restfulspec.KeyOpenAPITags, []string{TAG}).
		Param(ws.QueryParameter("Name", "The name of the node").DataType("string")).
		Param(ws.QueryParameter("Namespace", "The namespace of the node").DataType("string")).
		Param(ws.BodyParameter("Node", "The json string of the Node object").DataType("string")).
		Operation("Update node").
		Returns(200, "OK", apis.Node{}).
		Returns(400, "Not Found", nil),
	)

	ws.Route(ws.PATCH(fmt.Sprintf("/")).
		To(h.PatchNode).
		Doc("Patch a node with name").
		Metadata(restfulspec.KeyOpenAPITags, []string{TAG}).
		Param(ws.QueryParameter("Name", "The name of the node").DataType("string")).
		Param(ws.QueryParameter("Namespace", "The namespace of the node").DataType("string")).
		Param(ws.BodyParameter("Node", "The json string of the Node object").DataType("string")).
		Operation("Patch node").
		Returns(200, "OK", apis.Node{}).
		Returns(400, "Not Found", nil),
	)

	ws.Route(ws.DELETE(fmt.Sprintf("/")).
		To(h.DeleteNode).
		Doc("Delete a node with name").
		Metadata(restfulspec.KeyOpenAPITags, []string{TAG}).
		Param(ws.QueryParameter("Name", "The name of the node").DataType("string")).
		Param(ws.QueryParameter("Namespace", "The namespace of the node").DataType("string")).
		Operation("Delete node").
		Returns(200, "OK", apis.Node{}).
		Returns(400, "Not Found", nil))

	return ws
}

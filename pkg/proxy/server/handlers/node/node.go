package node

import (
	"context"
	"fmt"
	restfulspec "github.com/emicklei/go-restful-openapi/v2"
	"github.com/emicklei/go-restful/v3"
	"hit.edu/framework/pkg/apimachinery/types"
	apis "hit.edu/framework/pkg/apis/cores"
	metav1 "hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/client-go/clients"
	"hit.edu/framework/pkg/client-go/clients/typed/core"
	"hit.edu/framework/pkg/component-base/analyzer"
	"hit.edu/framework/pkg/component-base/logs"
	"net/http"
)

type NodeHandler struct {
	client core.NodeInterface
}

var _ Handler = &NodeHandler{}

func NewNodeHandler(clientSet *clients.ClientSet) *NodeHandler {
	c := clientSet.Core().Nodes(apis.NamespaceAll)
	return &NodeHandler{
		client: c,
	}
}

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

	result, err := h.client.Get(context.TODO(), name, metav1.GetOptions{})
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

	result, err := h.client.Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		logs.Errorf("Get node %s error: %v , node not exist ! creating node ", name, err)
	} else if result.Name == name {
		logs.Errorf("Create node %s error, node existed: %v", name, result)
		err = fmt.Errorf("create node %s error, node existed: %v", name, result)
		err := response.WriteError(http.StatusConflict, err)
		if err != nil {
			logs.Errorf("failed to return a status code")
			return
		}
		return
	}

	// 不存在，解析用户的输入
	ew := &apis.Node{}
	err = request.ReadEntity(ew)
	if err != nil {
		logs.Errorf("Failed to create node %s, error: %v", name, err)
		err := response.WriteError(http.StatusInternalServerError, err)
		if err != nil {
			logs.Errorf("failed to return a status code")
			return
		}
		return
	}

	// 格式校验
	res, err := analyzer.SerializeToJson(ew)
	_, err = analyzer.Deserialize(res, apis.Node{})
	if err != nil {
		err := response.WriteError(http.StatusBadRequest, err)
		if err != nil {
			logs.Errorf("failed to return a status code")
			return
		}
		//return
	}

	// TODO：为NODE分配MachineID?

	// 将Node写入数据库中
	result, err = h.client.Create(context.TODO(), ew, metav1.CreateOptions{})
	if err != nil {
		err := response.WriteError(http.StatusInternalServerError, err)
		if err != nil {
			logs.Errorf("failed to return a status code")
			return
		}
		logs.Errorf("Create node %s ，failed write to database ,error: %v", name, err)
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
	logs.Debugf("Create node %v", result)
}

func (h *NodeHandler) UpdateNode(request *restful.Request, response *restful.Response) {
	// 先检查Node是否存在
	// 存在：更新
	// 不存在：返回错误
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

	// 检查Node是否存在
	node, err := h.client.Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		logs.Errorf("Get node %s error: %v , node not exist! ", name, err)
		err := response.WriteError(http.StatusNotFound, err)
		if err != nil {
			logs.Errorf("failed to return a status code")
			return
		}
		return
	}

	// Node 存在，更新
	if node.Name == name {
		err := request.ReadEntity(&node)
		if err != nil {
			err := response.WriteError(http.StatusInternalServerError, err)
			if err != nil {
				logs.Errorf("failed to return a status code")
				return
			}
			return
		}

		// 格式验证
		res, err := analyzer.SerializeToJson(node)
		_, err = analyzer.Deserialize(res, apis.Node{})
		if err != nil {
			err := response.WriteError(http.StatusBadRequest, err)
			if err != nil {
				logs.Errorf("failed to return a status code ")
				return
			}
			// return
		}

		updatedNode, updateErr := h.client.Update(context.TODO(), node, metav1.UpdateOptions{})
		if updateErr != nil {
			logs.Errorf("Update node %s error: %v", name, updateErr)
			err := response.WriteError(http.StatusInternalServerError, err)
			if err != nil {
				logs.Errorf("failed to return a status code")
				return
			}
		}

		//返回成功修改的通知
		err = response.WriteHeaderAndEntity(http.StatusOK, updatedNode)
		if err != nil {
			logs.Errorf("failed to return a status code")
			return
		}

		// 记录日志
		logs.Debugf("update node : %v", name)

	}
}

func (h *NodeHandler) DeleteNode(request *restful.Request, response *restful.Response) {
	// 查看Node是否存在
	// 如果存在，删除节点
	// 如果不存在，返回 404 not found
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

	// 查看node是否存在
	node, err := h.client.Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		logs.Errorf("Get node %s error: %v , node not exist!", name, err)
		err := response.WriteError(http.StatusNotFound, err)
		if err != nil {
			logs.Errorf("failed to return a status code")
			return
		}
	}

	// node 存在
	if node.Name == name {
		err := h.client.Delete(context.TODO(), name, metav1.DeleteOptions{})
		if err != nil {
			logs.Errorf("Delete node %s error: %v", name, err)
			err := response.WriteError(http.StatusInternalServerError, err)
			if err != nil {
				logs.Errorf("failed to return a status code")
				return
			}
		}

		// 返回停止成功的状态
		response.WriteHeader(http.StatusOK)

		// 记录日志
		logs.Debugf("delete node : %v", name)
	}
}

func (h *NodeHandler) PatchNode(request *restful.Request, response *restful.Response) {
	// 先检查Node是否存在
	// 存在：部分更新
	// 不存在：返回错误
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

	// 检查Node是否存在
	node, err := h.client.Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		logs.Errorf("Get node %s error: %v , node not exist! ", name, err)
		err := response.WriteError(http.StatusNotFound, err)
		if err != nil {
			logs.Errorf("failed to return a status code")
			return
		}
		return
	}

	// Node 存在，部分更新
	if node.Name == name {
		err := request.ReadEntity(node)
		if err != nil {
			err := response.WriteError(http.StatusInternalServerError, err)
			if err != nil {
				logs.Errorf("failed to return a status code")
				return
			}
			return
		}

		patchNode, err := analyzer.SerializeToJson(node)
		if err != nil {
			logs.Errorf("Serialize patch node error: %v", err)
			err := response.WriteError(http.StatusInternalServerError, err)
			if err != nil {
				logs.Errorf("failed to return a status code")
				return
			}
		}
		logs.Debugf("patch node : %v", patchNode)
		patchedNode, err := h.client.Patch(context.TODO(), name, types.StrategicMergePatchType, []byte(patchNode), metav1.PatchOptions{})
		if err != nil {
			logs.Errorf("Patch node %s error: %v", name, err)
			err := response.WriteError(http.StatusInternalServerError, err)
			if err != nil {
				logs.Errorf("failed to return a status code")
				return
			}
		}

		//返回成功修改的通知
		err = response.WriteHeaderAndEntity(http.StatusOK, patchedNode)
		if err != nil {
			logs.Errorf("failed to return a status code")
			return
		}

		// 记录日志
		logs.Debugf("update node : %v", name)

	}
}

func (h *NodeHandler) NewGetWebService() *restful.WebService {
	ws := new(restful.WebService)
	ws.Path(NODE_PATH).
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/").
		To(h.GetNode).
		Doc("Get a node with name").
		Metadata(restfulspec.KeyOpenAPITags, []string{TAG}).
		Param(ws.QueryParameter("Name", "The name of the node").DataType("string")).
		Operation("Get node").
		Returns(200, "OK", apis.Node{}).
		Returns(400, "Not Found", nil),
	)

	ws.Route(ws.POST("/").
		To(h.CreateNode).
		Doc("Create a node with name").
		Metadata(restfulspec.KeyOpenAPITags, []string{TAG}).
		Operation("Create node").
		Param(ws.QueryParameter("Name", "The name of the node").DataType("string")).
		Param(ws.BodyParameter("Node", "The json string of the Node object").DataType("string")).
		Returns(200, "OK", apis.Node{}).
		Returns(400, "Not Found", nil),
	)

	ws.Route(ws.PUT("/").
		To(h.UpdateNode).
		Doc("Update a node with name").
		Metadata(restfulspec.KeyOpenAPITags, []string{TAG}).
		Param(ws.QueryParameter("Name", "The name of the node").DataType("string")).
		Param(ws.BodyParameter("Node", "The json string of the Node object").DataType("string")).
		Operation("Update node").
		Returns(200, "OK", apis.Node{}).
		Returns(400, "Not Found", nil),
	)

	ws.Route(ws.PATCH("/").
		To(h.PatchNode).
		Doc("Patch a node with name").
		Metadata(restfulspec.KeyOpenAPITags, []string{TAG}).
		Param(ws.QueryParameter("Name", "The name of the node").DataType("string")).
		Param(ws.BodyParameter("Node", "The json string of the Node object").DataType("string")).
		Operation("Patch node").
		Returns(200, "OK", apis.Node{}).
		Returns(400, "Not Found", nil),
	)

	ws.Route(ws.DELETE("/").
		To(h.DeleteNode).
		Doc("Delete a node with name").
		Metadata(restfulspec.KeyOpenAPITags, []string{TAG}).
		Param(ws.QueryParameter("Name", "The name of the node").DataType("string")).
		Operation("Delete node").
		Returns(200, "OK", apis.Node{}).
		Returns(400, "Not Found", nil))

	return ws
}

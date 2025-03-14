package node

import (
	"context"
	"fmt"
	restfulspec "github.com/emicklei/go-restful-openapi/v2"
	"github.com/emicklei/go-restful/v3"
	apis "hit.edu/framework/pkg/apis/cores"
	metav1 "hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/client-go/clients"
	"hit.edu/framework/pkg/client-go/clients/typed/core"
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
	name := request.PathParameter(NODE_NAME)
	result, err := h.client.Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		logs.Errorf("Get node %s error: %v", name, err)
		response.WriteError(http.StatusInternalServerError, err)
	}
	err = response.WriteEntity(result)
	if err != nil {
		response.WriteError(http.StatusInternalServerError, err)
	}
	logs.Debugf("Get node")
}

func (h *NodeHandler) CreateNode(request *restful.Request, response *restful.Response) {
	// 先查询Workflow是否存在
	name := request.PathParameter(NODE_NAME)
	result, err := h.client.Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		logs.Errorf("Get node %s error: %v", name, err)
		//response.WriteError(http.StatusInternalServerError, err)
	}
	if result.Name == name {
		logs.Errorf("Create node %s error, node existed: %v", name, result)
		err = fmt.Errorf("Create node %s error, node existed: %v", name, result)
		response.WriteError(http.StatusInternalServerError, err)
		return
	}

	// 解析用户的输入
	ew := &apis.Node{}
	err = request.ReadEntity(ew)
	if err != nil {
		logs.Errorf("Failed to create node %s, error: %v", name, err)
		response.WriteError(http.StatusInternalServerError, err)
		return
	}
	// TODO: 格式校验
	// TODO: 为Workflow分配ID
	logs.Debugf("Create node %s", name)

	// 将Workflow写入数据库中
	result, err = h.client.Create(context.TODO(), ew, metav1.CreateOptions{})
	if err != nil {
		response.WriteError(http.StatusInternalServerError, err)
		logs.Errorf("Create node %s error: %v", name, err)
		return
	}
	// 返回结果
	err = response.WriteEntity(result)
	if err != nil {
		response.WriteError(http.StatusInternalServerError, err)
	}
	logs.Debugf("Create node %v", result)
}

// TODO: Update Workflow
// TODO: Patch Workflow
// TODO: Delete Workflow

func (h *NodeHandler) NewGetWebService() *restful.WebService {
	ws := new(restful.WebService)
	ws.Path(NODE_PATH).
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET(fmt.Sprintf("/{%s}", NODE_NAME)).
		To(h.GetNode).
		Doc("Get a node with name").
		Metadata(restfulspec.KeyOpenAPITags, []string{TAG}).
		Operation("Get node").
		Returns(200, "OK", apis.Node{}).
		Returns(400, "Not Found", nil),
	)

	ws.Route(ws.POST(fmt.Sprintf("/{%s}", NODE_NAME)).
		To(h.CreateNode).
		Doc("Create a node with name").
		Metadata(restfulspec.KeyOpenAPITags, []string{TAG}).
		Operation("Create node").
		Returns(200, "OK", apis.Node{}).
		Returns(400, "Not Found", nil),
	)
	return ws
}

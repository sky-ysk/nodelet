package node

import (
	restfulspec "github.com/emicklei/go-restful-openapi/v2"
	"github.com/emicklei/go-restful/v3"
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/client-go/clients"
	"hit.edu/framework/pkg/client-go/util/manager"
	"hit.edu/framework/pkg/component-base/logs"
	"net/http"
	"sync"
)

type NodesHandler struct {
	clientSet *clients.ClientSet
	manager   *manager.Manager
	mu        sync.Mutex
}

//type CurrentNodesHandler struct {
//	client core.NodeInterface
//}

var _ Handler = &NodesHandler{}

//func NewNodesHandler(clientSet *clients.ClientSet) *NodesHandler {
//	c := clientSet.Core().Nodes("test") // apis.NamespaceAll
//	return &NodesHandler{
//		client: c,
//	}
//}

// NewNodeHandler 创建一个 NodeHandler
func NewNodesHandler(clientSet *clients.ClientSet) *NodesHandler {
	return &NodesHandler{
		manager:   manager.NewManager(clientSet),
		clientSet: clientSet,
	}
}

// GetClient 根据 namespace 获取 client，如果不存在则创建
//func (h *NodesHandler) GetClient(namespace string) *CurrentNodesHandler {
//	h.mu.Lock()
//	defer h.mu.Unlock()
//
//	// 如果已经存在，直接返回
//	if c, exists := h.clients[namespace]; exists {
//		return &CurrentNodesHandler{
//			client: c,
//		}
//	}
//
//	// 否则创建新的 client
//	newClient := h.clientSet.Core().Nodes(namespace)
//	h.clients[namespace] = newClient
//	return &CurrentNodesHandler{
//		client: newClient,
//	}
//}

func (h *NodesHandler) GetNodes(request *restful.Request, response *restful.Response) {

	//// 从url中获取namespace
	//c := &CurrentNodesHandler{}
	//namespace := request.QueryParameter(NAME_SPACE)
	////if namespace == "" {
	////	//err := response.WriteError(http.StatusBadRequest, fmt.Errorf("namespace is required"))
	////	//if err != nil {
	////	//	logs.Errorf("failed to return a status code ")
	////	//	return
	////	//}
	////	//return
	////	c = h.GetClient(namespace)
	////
	////} else {
	////	c = h.GetClient(namespace)
	////}
	//
	//c = h.GetClient(namespace)
	//results, err := c.client.List(context.TODO(), metav1.ListOptions{})
	//if err != nil {
	//	logs.Errorf("Get nodes failed: %v", err)
	//	err := response.WriteError(http.StatusInternalServerError, err)
	//	if err != nil {
	//		logs.Errorf("failed to return a status code")
	//		return
	//	}
	//}
	//
	//err = response.WriteEntity(results)
	//if err != nil {
	//	err := response.WriteError(http.StatusInternalServerError, err)
	//	if err != nil {
	//		logs.Errorf("failed to return a status code")
	//		return
	//	}
	//}
	//logs.Debugf("Get nodes")

	// 从url中获取namespace
	namespace := request.QueryParameter(NAME_SPACE)
	var results *apis.NodeList
	var err error

	labels := request.QueryParameter("Label")
	if labels == "" {
		results, err = h.manager.GetNodes(namespace)
		if err != nil {
			logs.Errorf("Get nodes failed: %v", err)
			err := response.WriteError(http.StatusInternalServerError, err)
			if err != nil {
				logs.Errorf("failed to return a status code")
				return
			}
		}
	}

	results, err = h.manager.FilterNodes(namespace, labels)
	if err != nil {
		logs.Errorf("Get nodes with labels failed: %v", err)
		err := response.WriteError(http.StatusInternalServerError, err)
		if err != nil {
			logs.Errorf("failed to return a status code")
			return
		}
	}

	err = response.WriteEntity(results)
	if err != nil {
		err := response.WriteError(http.StatusInternalServerError, err)
		if err != nil {
			logs.Errorf("failed to return a status code")
			return
		}
	}
	logs.Debugf("Get nodes ")
}

// TODO: DeleteAll

func (h *NodesHandler) NewGetWebService() *restful.WebService {
	ws := new(restful.WebService)
	ws.Path(NODES_PATH).
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/").
		Doc("Get all nodes").
		Metadata(restfulspec.KeyOpenAPITags, []string{TAG}).
		Param(ws.QueryParameter("Namespace", "The namespace of the nodes").DataType("string")).
		To(h.GetNodes).
		Operation("Get nodes").
		Returns(200, "OK", []apis.Node{}).
		Returns(400, "Not Found", nil),
	)

	return ws
}

package group

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

type GroupHandler struct {
	client core.GroupInterface
}

var _ Handler = &GroupHandler{}

func NewGroupHandler(clientSet *clients.ClientSet) *GroupHandler {
	c := clientSet.Core().Groups(apis.NamespaceAll)
	return &GroupHandler{
		client: c,
	}
}

func (h *GroupHandler) GetGroup(request *restful.Request, response *restful.Response) {
	// TODO: 使用client-go实现查询
	name := request.PathParameter(GROUP_NAME)
	result, err := h.client.Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		logs.Errorf("Get group %s error: %v", name, err)
		response.WriteError(http.StatusInternalServerError, err)
	}
	err = response.WriteEntity(result)
	if err != nil {
		response.WriteError(http.StatusInternalServerError, err)
	}
	logs.Debugf("Get group")
}

func (h *GroupHandler) CreateGroup(request *restful.Request, response *restful.Response) {
	// 先查询Workflow是否存在
	name := request.PathParameter(GROUP_NAME)
	result, err := h.client.Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		logs.Errorf("Get group %s error: %v", name, err)
		//response.WriteError(http.StatusInternalServerError, err)
	}
	if result.Name == name {
		logs.Errorf("Create group %s error, group existed: %v", name, result)
		err = fmt.Errorf("Create group %s error, group existed: %v", name, result)
		response.WriteError(http.StatusInternalServerError, err)
		return
	}

	// 解析用户的输入
	ew := &apis.Group{}
	err = request.ReadEntity(ew)
	if err != nil {
		logs.Errorf("Failed to create group %s, error: %v", name, err)
		response.WriteError(http.StatusInternalServerError, err)
		return
	}
	// TODO: 格式校验
	// TODO: 为Workflow分配ID
	logs.Debugf("Create group %s", name)

	// 将Workflow写入数据库中
	result, err = h.client.Create(context.TODO(), ew, metav1.CreateOptions{})
	if err != nil {
		response.WriteError(http.StatusInternalServerError, err)
		logs.Errorf("Create group %s error: %v", name, err)
		return
	}
	// 返回结果
	err = response.WriteEntity(result)
	if err != nil {
		response.WriteError(http.StatusInternalServerError, err)
	}
	logs.Debugf("Create group %v", result)
}

func (h *GroupHandler) NewGetWebService() *restful.WebService {
	ws := new(restful.WebService)
	ws.Path(GROUP_PATH).
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET(fmt.Sprintf("/{%s}", GROUP_NAME)).
		To(h.GetGroup).
		Doc("Get a group with name").
		Metadata(restfulspec.KeyOpenAPITags, []string{TAG}).
		Operation("Get group").
		Returns(200, "OK", apis.Group{}).
		Returns(400, "Not Found", nil),
	)

	ws.Route(ws.POST(fmt.Sprintf("/{%s}", GROUP_NAME)).
		To(h.CreateGroup).
		Doc("Create a group with name").
		Metadata(restfulspec.KeyOpenAPITags, []string{TAG}).
		Operation("Create group").
		Returns(200, "OK", apis.Group{}).
		Returns(400, "Not Found", nil),
	)
	return ws
}

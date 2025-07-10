package group

import (
	restfulspec "github.com/emicklei/go-restful-openapi/v2"
	"github.com/emicklei/go-restful/v3"
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/client-go/clients"
	"hit.edu/framework/pkg/client-go/util/manager"
	"hit.edu/framework/pkg/component-base/logs"
	"net/http"
	"strconv"
	"sync"
)

type GroupsHandler struct {
	clientSet *clients.ClientSet
	manager   *manager.Manager
	mu        sync.Mutex
}

var _ Handler = &GroupsHandler{}

// NewGroupsHandler 创建一个 GroupsHandler
func NewGroupsHandler(clientSet *clients.ClientSet) *GroupsHandler {
	return &GroupsHandler{
		manager:   manager.NewManager(clientSet),
		clientSet: clientSet,
	}
}

func (h *GroupsHandler) GetGroups(request *restful.Request, response *restful.Response) {
	// TODO: 检查某些namespace为空判断是否考虑namespace为空默认所有namespace资源的情况
	// 获取namespace
	namespace := request.QueryParameter(NAME_SPACE)

	// 分页
	pageStr := request.QueryParameter(PAGE)
	pageSizeStr := request.QueryParameter(PAGE_SIZE)

	page := 0
	pageSize := 10

	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil {
			if p <= 0 {
				err := response.WriteErrorString(400, "Invalid Page parameter")
				if err != nil {
					logs.Error(err.Error())
					return
				}
				return
			}
			page = p
		} else {
			// 可选：返回错误或使用默认值
			err := response.WriteErrorString(400, "Invalid Page parameter")
			if err != nil {
				logs.Errorf("failed to return a status code")
			}
			return
		}
	}

	if pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil {
			if ps < 0 {
				err := response.WriteErrorString(400, "Invalid Page parameter")
				if err != nil {
					logs.Errorf("failed to return a status code")
					return
				}
				return
			}
			pageSize = ps
		} else {
			err := response.WriteErrorString(400, "Invalid PageSize parameter")
			if err != nil {
				logs.Errorf("failed to return a status code")
			}
			return
		}
	}

	var results *apis.GroupList
	var err error

	labels := request.QueryParameter("Label")
	if labels == "" {
		results, err = h.manager.GetGroups(namespace)
		if err != nil {
			logs.Errorf("Get groups with labels failed: %v", err)
			err := response.WriteError(http.StatusInternalServerError, err)
			if err != nil {
				logs.Errorf("failed to return a status code")
				return
			}
		}
	} else {
		results, err = h.manager.FilterGroups(namespace, labels)
		if err != nil {
			logs.Errorf("Get groups failed: %v", err)
			err := response.WriteError(http.StatusInternalServerError, err)
			if err != nil {
				logs.Errorf("failed to return a status code")
				return
			}
		}
	}

	if page == 0 {
		err = response.WriteHeaderAndEntity(http.StatusOK, results)
		if err != nil {
			logs.Errorf("failed to return a status code")
			return
		}
		logs.Debugf("Get events ")
	} else {
		if (page-1)*pageSize < len(results.Items) && (page*pageSize)-1 < len(results.Items) {
			// 返回一整页
			tmpRes := results
			tmpRes.Items = results.Items[(page-1)*pageSize : page*pageSize]
			err = response.WriteHeaderAndEntity(http.StatusOK, tmpRes)
			if err != nil {
				logs.Errorf("failed to return a status code")
				return
			}
			logs.Debugf("Get events ")
		} else if (page-1)*pageSize < len(results.Items) {
			// 返回开头到最后
			tmpRes := results
			tmpRes.Items = results.Items[(page-1)*pageSize:]
			err = response.WriteHeaderAndEntity(http.StatusOK, tmpRes)
			if err != nil {
				logs.Errorf("failed to return a status code")
				return
			}
			logs.Debugf("Get groups ")
		} else {
			// 页数太大，没有这么多group数据
			err := response.WriteErrorString(400, "该页没有group数据")
			if err != nil {
				logs.Errorf("failed to return a status code")
				return
			}
			return
		}
	}

	//err = response.WriteEntity(results)
	//if err != nil {
	//	err := response.WriteError(http.StatusInternalServerError, err)
	//	if err != nil {
	//		logs.Errorf("failed to return a status code")
	//		return
	//	}
	//}

	//err = response.WriteHeaderAndEntity(http.StatusOK, results)
	//if err != nil {
	//	logs.Errorf("failed to return a status code")
	//	return
	//}

	logs.Debugf("Get groups")
}

func (h *GroupsHandler) NewGetWebService() *restful.WebService {
	ws := new(restful.WebService)
	ws.Path(GROUPS_PATH).
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/").
		Doc("Get all groups").
		Metadata(restfulspec.KeyOpenAPITags, []string{TAG}).
		Param(ws.QueryParameter("Page", "The page of the groups(optional)").DataType("int")).
		Param(ws.QueryParameter("PageSize", "The size of the page(optional Default=10)").DataType("int")).
		Param(ws.QueryParameter("Label", "Labels of the groups (optional)").DataType("string")).
		Param(ws.QueryParameter("Namespace", "The namespace of the groups").DataType("string")).
		To(h.GetGroups).
		Operation("Get groups").
		Returns(200, "OK", []apis.Group{}).
		Returns(404, "Not Found", nil),
	)
	return ws
}

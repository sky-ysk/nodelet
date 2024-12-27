package server

import (
	"fmt"
	"github.com/emicklei/go-restful/v3"
	"go.uber.org/zap"
	"hit.edu/framework/pkg/apiserver/endpoints/request"
	"hit.edu/framework/pkg/apiserver/util/notfoundhandler"
	"hit.edu/framework/pkg/component-base/logs"
	"k8s.io/apimachinery/pkg/util/sets"
	
	"net/http"
	"strings"
	"time"
)

const (
	defaultKeepAlivePeriod = 3 * time.Minute
)

// 处理不同的HTTP请求
type APIServerHandler struct {
	//
	FullHandlerChain http.Handler
	
	//
	GoRestfulContainer *restful.Container
	
	//
	Director http.Handler
}

type director struct {
	name               string
	goRestfulContainer *restful.Container
	notFoundHandler    http.Handler
}

func NewAPIServerHandler(name string) *APIServerHandler {
	// TODO：基础Handler
	// TODO: Not Found Handler
	notFoundHandler := notfoundhandler.New()
	
	goRestfulContainer := restful.NewContainer()
	goRestfulContainer.Router(restful.CurlyRouter{}) // e.g. for proxy/{kind}/{name}/{*}
	goRestfulContainer.RecoverHandler(func(panicReason interface{}, httpWriter http.ResponseWriter) {
		// TODO: 处理函数
	})
	goRestfulContainer.ServiceErrorHandler(func(serviceErr restful.ServiceError, request *restful.Request, response *restful.Response) {
		// TODO: 处理函数
	})
	
	director := director{
		name:               name,
		goRestfulContainer: goRestfulContainer,
		notFoundHandler:    notFoundHandler,
	}
	
	return &APIServerHandler{
		FullHandlerChain:   director,
		GoRestfulContainer: goRestfulContainer,
		Director:           director,
	}
	
}

func (d director) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	
	//对请求进行预处理
	req = preProcess(w, req)
	path := req.URL.Path
	//logs.Info("Get Http Request with Path\t", path)
	logs.Info("Get Http Request with Path", zap.String("path", path))
	
	// 使用webservice处理/apis目录下的请求
	// 遍历Webservice,如果找到对应的前缀，则将请求路由到该WebService下
	for _, ws := range d.goRestfulContainer.RegisteredWebServices() {
		switch {
		case ws.RootPath() == "/apis":
			// 处理apis的情况
			if path == "/apis" || path == "/apis/" {
				//logs.Info("Http Requests Start With /apis")
				logs.Info("Http Requests Start With /apis")
				d.goRestfulContainer.Dispatch(w, req)
				return
			}
		
		case strings.HasPrefix(path, ws.RootPath()):
			//
			if len(path) == len(ws.RootPath()) || path[len(ws.RootPath())] == '/' {
				//
				d.goRestfulContainer.Dispatch(w, req)
				return
			}
		}
	}
	
	//TODO: 处理不正确路径的情况
	d.notFoundHandler.ServeHTTP(w, req)
}

func preProcess(w http.ResponseWriter, req *http.Request) *http.Request {
	//获取requestInfo,注入到req的上下文中
	resolver := &request.RequestInfoFactory{
		APIPrefixes: sets.NewString("apis"),
	}
	requestInfo, err := resolver.NewRequestInfo(req)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to parse request info: %v", err), http.StatusBadRequest)
		return nil
	}
	ctx := request.WithRequestInfo(req.Context(), requestInfo)
	req = req.WithContext(ctx)
	return req
}

func (a *APIServerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.FullHandlerChain.ServeHTTP(w, r)
}

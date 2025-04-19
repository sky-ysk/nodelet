package server

import (
	"bytes"
	"fmt"
	"github.com/emicklei/go-restful/v3"
	"go.uber.org/zap"
	apierrors "hit.edu/framework/pkg/apimachinery/errors"
	"hit.edu/framework/pkg/apimachinery/runtime"
	"hit.edu/framework/pkg/apimachinery/runtime/schema"
	"hit.edu/framework/pkg/apiserver/endpoints/handler/responsewriters"
	"hit.edu/framework/pkg/apiserver/endpoints/request"
	"hit.edu/framework/pkg/apiserver/util/notfoundhandler"
	"hit.edu/framework/pkg/component-base/logs"
	"k8s.io/apimachinery/pkg/util/sets"
	"net/http"
	rt "runtime"
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

func NewAPIServerHandler(name string, s runtime.NegotiatedSerializer) *APIServerHandler {
	notFoundHandler := notfoundhandler.New()

	goRestfulContainer := restful.NewContainer()
	goRestfulContainer.Router(restful.CurlyRouter{}) // e.g. for proxy/{kind}/{name}/{*}
	goRestfulContainer.RecoverHandler(func(panicReason interface{}, httpWriter http.ResponseWriter) {
		logStackOnRecover(s, panicReason, httpWriter)
	})
	goRestfulContainer.ServiceErrorHandler(func(serviceErr restful.ServiceError, request *restful.Request, response *restful.Response) {
		serviceErrorHandler(s, serviceErr, request, response)
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
	//req = preProcess(w, req)
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
				//logs.Info("Http Requests Start With /apis")]

				logs.Infof("%v: %v %q satisfied by gorestful with webservice %v", d.name, req.Method, path, ws.RootPath())

				d.goRestfulContainer.Dispatch(w, req)
				return
			}

		case strings.HasPrefix(path, ws.RootPath()):
			if len(path) == len(ws.RootPath()) || path[len(ws.RootPath())] == '/' {
				logs.Infof("%v: %v %q satisfied by gorestful with webservice %v", d.name, req.Method, path, ws.RootPath())
				d.goRestfulContainer.Dispatch(w, req)
				return
			}
		}
	}

	//TODO: 处理不正确路径的情况
	logs.Infof("%v: %v %q satisfied by notFoundHandler", d.name, req.Method, path)
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

// logStackOnRecover 捕获panic，记录到日志并返回给客户端
func logStackOnRecover(s runtime.NegotiatedSerializer, panicReason interface{}, w http.ResponseWriter) {
	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("recover from panic situation: - %v\r\n", panicReason))
	for i := 2; ; i++ {
		_, file, line, ok := rt.Caller(i)
		if !ok {
			break
		}
		buffer.WriteString(fmt.Sprintf("    %s:%d\r\n", file, line))
	}
	logs.Error(buffer.String())

	headers := http.Header{}
	if ct := w.Header().Get("Content-Type"); len(ct) > 0 {
		headers.Set("Accept", ct)
	}
	responsewriters.ErrorNegotiated(apierrors.NewGenericServerResponse(http.StatusInternalServerError, "", schema.GroupResource{}, "", "", 0, false), s, schema.GroupVersion{}, w, &http.Request{Header: headers})
}

// serviceErrorHandler 处理服务错误
func serviceErrorHandler(s runtime.NegotiatedSerializer, serviceErr restful.ServiceError, request *restful.Request, resp *restful.Response) {
	logs.Errorf("Service error:%v,URL:%s", serviceErr.Message, request.Request.URL.Path)
	responsewriters.ErrorNegotiated(
		apierrors.NewGenericServerResponse(serviceErr.Code, "", schema.GroupResource{}, "", serviceErr.Message, 0, false),
		s,
		schema.GroupVersion{},
		resp,
		request.Request,
	)
}

func (a *APIServerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.FullHandlerChain.ServeHTTP(w, r)
}

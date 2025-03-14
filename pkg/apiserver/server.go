package apiserver

import (
	"context"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"hit.edu/framework/pkg/apimachinery/runtime"
	"hit.edu/framework/pkg/apimachinery/runtime/schema"
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/apiserver/endpoints/handler/filters"
	corerest "hit.edu/framework/pkg/apiserver/registry/core/rest"
	genericregistry "hit.edu/framework/pkg/apiserver/registry/generic"
	"hit.edu/framework/pkg/apiserver/registry/rest"
	"hit.edu/framework/pkg/component-base/logs"
	"k8s.io/apimachinery/pkg/util/sets"
	"log"
	"net/http"
	"sync"

	"hit.edu/framework/pkg/apiserver/endpoints"
	"hit.edu/framework/pkg/apiserver/server"
	"strings"
	"time"
)

// TOOO: 替换K8s相关组件

// 模块名称
const (
	APIServerName = "api-server"
)

type APIServerInterface interface {
	//
	PrepareRun() PreparedAPIServer

	//
	Destroy()
}

type APIServer struct {
	APIServerInterface

	// 处理Server HTTP请求
	Handler *server.APIServerHandler

	// 服务相关信息
	ServingInfo *server.ServingInfo

	// 服务器配置信息

	//给定资源的RESTOptions
	RESTOptionsGetter genericregistry.RESTOptionsGetter

	// PostStart、PreShutDown钩子函数配置
	postStartHookLock      sync.Mutex
	postStartHooks         map[string]postStartHookEntry
	postStartHooksCalled   bool
	disabledPostStartHooks sets.String

	preShutdownHookLock    sync.Mutex
	preShutdownHooks       map[string]preShutdownHookEntry
	preShutdownHooksCalled bool

	// 关闭延时
	ShutdownTimeout time.Duration

	// minRequestTimeout is how short the request timeout can be.  This is used to build the RESTHandler
	minRequestTimeout time.Duration
}

// 需要安装Rest等配置
// 因此配置PreparedAPIServer
type PreparedAPIServer struct {
	*APIServer
}

func NewAPIServer(cfg *Config) *APIServer {
	// TODO: 创建Handler
	// TODO: 创建Serving
	// TODO: 参数配置
	// 创建Serving
	apiServerHandler := server.NewAPIServerHandler(APIServerName, cfg.Serializer)

	//构建handler链
	logs.Info("start to build handler chain")
	apiServerHandler.FullHandlerChain = HandlerWithFilters(apiServerHandler.FullHandlerChain, cfg)

	s := &APIServer{
		ServingInfo:            cfg.ServingInfo,
		Handler:                apiServerHandler,
		RESTOptionsGetter:      cfg.RESTOptionsGetter,
		postStartHooks:         map[string]postStartHookEntry{},
		preShutdownHooks:       map[string]preShutdownHookEntry{},
		disabledPostStartHooks: sets.NewString(),
		ShutdownTimeout:        cfg.ShutdownTimeout,
		minRequestTimeout:      time.Duration(cfg.MinRequestTimeout) * time.Second,
	}
	//TODO:调用s.AddPostStartHook、s.AddPreShutDownHook添加钩子函数
	return s
}

// HandlerWithFilters 过滤http请求
// TODO:访问控制
func HandlerWithFilters(apiHandler http.Handler, c *Config) http.Handler {
	//注入requestInfo到请求上下文
	logs.Debug("handler register with RequestInfo filter")
	apiHandler = filters.WithRequestInfo(apiHandler, c.RequestInfoResolver)
	return apiHandler
}

func (s *APIServer) PrepareRun() PreparedAPIServer {
	//创建核心组apiGroupInfo并注入各资源RESTStorage
	logs.Info("creating core apiGroupInfo")
	apiGroupInfo, err := corerest.NewRESTStorage(s.RESTOptionsGetter)
	if err != nil {
		logs.Error("Core apiGroupInfo created failed", zap.Error(err))
	}

	//装载各Group的资源RESTStorage
	logs.Info("installing core apiGroupInfo")
	if err := s.InstallAPIGroup(&apiGroupInfo); err != nil {
		logs.Error("Install Core Resources failed", zap.Error(err))
		return PreparedAPIServer{}
	}

	return PreparedAPIServer{s}
}

func (s *APIServer) Destroy() {
	//TODO: 清空资源
}

func (s *PreparedAPIServer) RunWithContext(ctx context.Context) error {
	logs.Info("Running API Server")
	// TODO: 实现运行逻辑
	// TODO: channel配置

	shutdownTimeout := s.ShutdownTimeout

	stopHTTPServerCtx, stopHTTPServer := context.WithCancelCause(context.WithoutCancel(ctx))
	go func() {
		defer stopHTTPServer(errors.New("time to stop HTTP server"))
	}()

	stoppedCh, listenerStoppedCh, err := s.NonBlockingRunWithContext(stopHTTPServerCtx, shutdownTimeout)
	if err != nil {
		return err
	}

	<-listenerStoppedCh
	<-stoppedCh

	//TODO:资源清理，优雅推出
	func() {
		defer func() {
			log.Print("[graceful-termination] pre-shutdown hooks completed")
		}()
		err = s.RunPreShutdownHooks()
	}()
	if err != nil {
		return err
	}

	//logs.Info("Stopping API Server")
	logs.Info("Stopping API Server")
	return nil
}

func (s PreparedAPIServer) NonBlockingRunWithContext(ctx context.Context, shutdownTimeout time.Duration) (<-chan struct{}, <-chan struct{}, error) {
	internalStopCh := make(chan struct{})
	var stoppedCh <-chan struct{}
	var listenerStoppedCh <-chan struct{}

	// TODO: 安全认证相关

	// 开启HTTP Server
	if s.ServingInfo != nil && s.Handler != nil {
		var err error
		stoppedCh, listenerStoppedCh, err = s.ServingInfo.Serve(s.Handler, shutdownTimeout, internalStopCh)
		if err != nil {
			close(internalStopCh)
			return nil, nil, err
		}
	}

	go func() {
		<-ctx.Done()
		close(internalStopCh)
	}()

	// 运行PostStartHooks
	s.RunPostStartHooks(ctx)

	return stoppedCh, listenerStoppedCh, nil
}

// InstallAPIGroup Install API Resources
func (s *APIServer) InstallAPIGroup(apiGroupInfo *server.APIGroupInfo) error {
	groupVersion := apis.SchemeGroupVersion
	// 这里安装REST相关组件
	logs.Debug("creating apiGroupVersion with RESTStorage", zap.String("group", groupVersion.Group), zap.String("version", groupVersion.Version))
	apiGroupVersion, err := s.getAPIGroupVersion(apiGroupInfo, groupVersion, APIGroupPrefix)
	if err != nil {
		logs.Error("create APIGroupVersion failed", zap.Error(err))
		return err
	}

	err = apiGroupVersion.InstallREST(s.Handler.GoRestfulContainer)
	if err != nil {
		logs.Error("install REST failed", zap.Error(err))
		return err
	}

	return nil
}

// From K8s
func (s *APIServer) getAPIGroupVersion(apiGroupInfo *server.APIGroupInfo, groupVersion schema.GroupVersion, apiPrefix string) (*endpoints.APIGroupVersion, error) {
	storage := make(map[string]rest.Storage)
	for k, v := range apiGroupInfo.VersionedResourcesStorageMap[groupVersion.Version] {
		if strings.ToLower(k) != k {
			return nil, fmt.Errorf("resource names must be lowercase only, not %q", k)
		}
		storage[k] = v
	}
	version := s.newAPIGroupVersion(apiGroupInfo, groupVersion)
	version.Root = apiPrefix
	version.Storage = storage
	return version, nil
}

// From K8s
func (s *APIServer) newAPIGroupVersion(apiGroupInfo *server.APIGroupInfo, groupVersion schema.GroupVersion) *endpoints.APIGroupVersion {

	allServedVersionsByResource := map[string][]string{}
	for version, resourcesInVersion := range apiGroupInfo.VersionedResourcesStorageMap {
		for resource := range resourcesInVersion {
			if len(groupVersion.Group) == 0 {
				allServedVersionsByResource[resource] = append(allServedVersionsByResource[resource], version)
			} else {
				allServedVersionsByResource[resource] = append(allServedVersionsByResource[resource], fmt.Sprintf("%s/%s", groupVersion.Group, version))
			}
		}
	}

	return &endpoints.APIGroupVersion{
		GroupVersion:      groupVersion,
		MetaGroupVersion:  apiGroupInfo.MetaGroupVersion,
		Serializer:        apiGroupInfo.NegotiatedSerializer,
		Typer:             apiGroupInfo.Scheme,
		Creater:           apiGroupInfo.Scheme,
		Convertor:         apiGroupInfo.Scheme,
		Defaulter:         apiGroupInfo.Scheme,
		Namer:             runtime.Namer(meta.NewAccessor()),
		MinRequestTimeout: s.minRequestTimeout,
	}
}

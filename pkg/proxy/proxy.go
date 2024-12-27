package proxy

import (
	"context"
	"hit.edu/framework/pkg/client-go/clients"
	"hit.edu/framework/pkg/component-base/logs"
	"hit.edu/framework/pkg/proxy/server"
	"net/http"
	"time"
)

type Proxy struct {
	ProxyInterface

	// 服务相关信息
	ServingInfo *server.ServingInfo

	// ClientSets
	ClientSet *clients.ClientSet

	// 关闭延时
	ShutdownTimeout time.Duration

	//
	minRequestTimeout time.Duration
}

type ProxyInterface interface {
	//
	Run(ctx context.Context) error
	//
	Destroy()
}

func NewProxy(cfg *Config) (*Proxy, error) {
	// TODO: 创建Handler
	// TODO: 创建Serving
	// TODO: 参数配置

	// 创建ClientSets
	clientSet, err := clients.NewForConfig(cfg.HandlerConfig.GetRestConfig())
	if err != nil {
		logs.Errorf("Failed to create clientSet: %v", err)
		return nil, err
	}

	// 创建Serving
	return &Proxy{
		ClientSet:         clientSet,
		ServingInfo:       cfg.ServingInfo,
		ShutdownTimeout:   cfg.ShutdownTimeout,
		minRequestTimeout: time.Duration(cfg.MinRequestTimeout) * time.Second,
	}, nil
}

func (r *Proxy) Destroy() {
	//TODO: 清空资源
}

func (r *Proxy) Run(ctx context.Context) error {
	logs.Info("Running API Proxy")
	// TODO: 优雅退出
	//shutdownTimeout := r.ShutdownTimeout
	//
	//stopHTTPServerCtx, stopHTTPServer := context.WithCancelCause(context.WithoutCancel(ctx))
	//go func() {
	//	defer stopHTTPServer(errors.New("time to stop HTTP s"))
	//}()
	logs.Infof("Server started, address %s", r.ServingInfo.Listener.Addr().String())

	handler := server.NewServer(r.ClientSet)
	// TODO: 服务器配置，如超时时间等

	// Openapi文档
	logs.Info("Get the API using http://localhost:8899/apidocs.json")
	logs.Info("Open Swagger UI using http://localhost:8899/apidocs")

	err := http.Serve(r.ServingInfo.Listener, &handler)
	if err != nil {
		logs.Fatalf("Failed to start server: %v", err)
	}

	logs.Info("Stopping API Proxy")
	return nil
}

package registry

import (
	"context"
	"hit.edu/framework/pkg/component-base/logs"
	"hit.edu/framework/pkg/registry/server"
	"net/http"
	"time"
)

// TODO: 文件查询功能（检索）
// TODO: 文件删除功能
// TODO: 文件版本（文件名 版本 tag一类）
// TODO: 文件重复时的处理
// TODO: 稳定性测试，测试多连接（并发），处理连接断开的情况，服务器优雅退出等
// TODO: 文件信息缓存（后面实现对应的Controller，将信息存入资源总线中，供快速查询使用）
// TODO: 身份认证 https（最后实现）

// 模块名称
const (
	RegistryName = "registry"
)

type Registry struct {
	RegistryInterface
	
	// 服务相关信息
	ServingInfo *server.ServingInfo
	
	// 关闭延时
	ShutdownTimeout time.Duration
	
	//
	minRequestTimeout time.Duration
}

type RegistryInterface interface {
	//
	Run(ctx context.Context) error
	//
	Destroy()
}

func NewRegistry(cfg *Config) (*Registry, error) {
	// TODO: 创建Handler
	// TODO: 创建Serving
	// TODO: 参数配置
	// 创建Serving
	return &Registry{
		ServingInfo:       cfg.ServingInfo,
		ShutdownTimeout:   cfg.ShutdownTimeout,
		minRequestTimeout: time.Duration(cfg.MinRequestTimeout) * time.Second,
	}, nil
}

func (r *Registry) Destroy() {
	//TODO: 清空资源
}

func (r *Registry) Run(ctx context.Context) error {
	logs.Info("Running Registry")
	// TODO: 实现运行逻辑
	// TODO: channel配置
	
	//shutdownTimeout := r.ShutdownTimeout
	//
	//stopHTTPServerCtx, stopHTTPServer := context.WithCancelCause(context.WithoutCancel(ctx))
	//go func() {
	//	defer stopHTTPServer(errors.New("time to stop HTTP server"))
	//}()
	
	//
	err := http.Serve(r.ServingInfo.Listener, nil)
	if err != nil {
		logs.Fatalf("Failed to start server: %v", err)
	}
	
	logs.Info("Stopping Registry")
	return nil
}

package proxy

import (
	"hit.edu/framework/pkg/apimachinery/runtime/schema"
	"hit.edu/framework/pkg/component-base/logs"
	"hit.edu/framework/pkg/proxy/server"
	"hit.edu/framework/pkg/proxy/server/handlers"
	"net"
	"strconv"
	"time"
)

// Config为对外暴露的配置，可以简化
type Config struct {
	// Serving Info
	ServingInfo *server.ServingInfo
	
	// Handler Config
	HandlerConfig *handlers.ClientSetConfig
	
	// 关闭延迟
	ShutdownTimeout time.Duration
	
	// minRequestTimeout is how short the request timeout can be.  This is used to build the RESTHandler
	MinRequestTimeout time.Duration
	
	Extra
}

type Extra struct {
	// TODO:
}

func NewConfig() *Config {
	//FIXME: 如果端口已经被占用，错误处理
	
	// 在这里设置各种选项
	servingOptions := server.NewServingOptions()
	
	// 创建Serving Info
	servingInfo := NewServingInfo(servingOptions)
	if servingInfo == nil {
		logs.Error("Serving Info is Empty")
		return nil
	}
	
	// HandlersConfig
	// TODO: 临时性配置
	handlerConfig := handlers.NewForConfig(
		"127.0.0.1",
		8120,
		schema.GroupVersion{
			"resources",
			"v1",
		},
	)
	
	// 完成所有参数配置
	c := &Config{
		ServingInfo:       servingInfo,
		HandlerConfig:     handlerConfig,
		ShutdownTimeout:   time.Duration(60) * time.Second,
		Extra:             Extra{},
		MinRequestTimeout: 1800,
	}
	
	return c
}

func NewServingInfo(s *server.ServingOptions) *server.ServingInfo {
	// 格式检查
	if s == nil {
		return nil
	}
	if s.BindPort <= 0 && s.Listener == nil {
		return nil
	}
	
	if s.Listener == nil {
		
		var err error
		addr := net.JoinHostPort(s.BindAddress.String(), strconv.Itoa(s.BindPort))
		
		c := net.ListenConfig{}
		
		s.Listener, s.BindPort, err = server.CreateListener(s.BindNetwork, addr, c)
		if err != nil {
			return nil
		}
	}
	
	return &server.ServingInfo{
		Listener: s.Listener,
	}
}

package apiserver

import (
	"fmt"
	"hit.edu/framework/pkg/apimachinery/runtime"
	"hit.edu/framework/pkg/apis/legacyscheme"
	apirequest "hit.edu/framework/pkg/apiserver/endpoints/request"
	"hit.edu/framework/pkg/apiserver/options"
	genericregistry "hit.edu/framework/pkg/apiserver/registry/generic"
	"hit.edu/framework/pkg/apiserver/server"
	"hit.edu/framework/pkg/component-base/logs"
	"k8s.io/apimachinery/pkg/util/sets"
	"net"
	"strconv"
	"time"
)

const (
	APIGroupPrefix = "/apis"
)

// Config为对外暴露的配置，可以简化
type Config struct {
	Option *options.Options
	// TODO: 其他配置

	// Serving Info
	ServingInfo *server.ServingInfo

	// 关闭延迟
	ShutdownTimeout time.Duration

	// minRequestTimeout is how short the request timeout can be.  This is used to build the RESTHandler
	MinRequestTimeout time.Duration

	//给定资源的RESTOptions
	RESTOptionsGetter genericregistry.RESTOptionsGetter

	//提供对象序列化
	Serializer runtime.NegotiatedSerializer

	// RequestInfoResolver 请求解析器
	RequestInfoResolver apirequest.RequestInfoResolver

	Extra
}

type Extra struct {
	// TODO:
}

func NewConfig(opts *options.Options) *Config {
	//FIXME: 如果端口已经被占用，错误处理

	// 在这里设置各种选项
	// 创建Serving Info
	servingInfo := NewServingInfo(opts.ServingOptions)
	if servingInfo == nil {
		//logs.Error("Serving Info is Empty")
		logs.Error("Serving Info is Empty")
		return nil
	}

	// 完成所有参数配置
	c := &Config{
		Option: opts,
		//
		ServingInfo:       servingInfo,
		ShutdownTimeout:   time.Duration(60) * time.Second,
		Extra:             Extra{},
		MinRequestTimeout: 1800,
	}

	//opts.EtcdOptions.ApplyStorageFactoryToConfig(c)
	storageFactory := &options.SimpleStorageFactory{Options: opts.EtcdOptions, StorageConfig: opts.EtcdOptions.StorageConfig}
	c.RESTOptionsGetter = storageFactory
	c.Serializer = legacyscheme.Codecs
	c.RequestInfoResolver = NewRequestInfoResolver()
	logs.Info("Config Completed")

	return c
}

func NewServingInfo(s *options.ServingOptions) *server.ServingInfo {

	if s == nil {
		return nil
	}
	if s.BindPort <= 0 && s.Listener == nil {
		return nil
	}

	//处理端口占用
	address := fmt.Sprintf(":%d", s.BindPort)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		//logs.Error("BindPort has been used")
		logs.Error("BindPort has been used")
		return nil
	}
	_ = listener.Close()

	if s.Listener == nil {
		var err error
		addr := net.JoinHostPort(s.BindAddress.String(), strconv.Itoa(s.BindPort))

		c := net.ListenConfig{}

		s.Listener, s.BindPort, err = server.CreateListener(s.BindNetwork, addr, c)
		if err != nil {
			return nil
		}
	} else {
		if _, ok := s.Listener.Addr().(*net.TCPAddr); !ok {
			return nil
		}
		s.BindPort = s.Listener.Addr().(*net.TCPAddr).Port
		s.BindAddress = s.Listener.Addr().(*net.TCPAddr).IP
	}

	return &server.ServingInfo{
		Listener: s.Listener,
		//HTTP2MaxStreamsPerConnection: s.HTTP2MaxStreamsPerConnection,
		//DisableHTTP2:                 s.DisableHTTP2Serving,
	}
}

func NewRequestInfoResolver() *apirequest.RequestInfoFactory {
	return &apirequest.RequestInfoFactory{
		APIPrefixes: sets.NewString("apis"),
	}
}

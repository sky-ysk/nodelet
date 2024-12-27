package server

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"golang.org/x/net/http2"
	"hit.edu/framework/pkg/component-base/logs"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"net"
	"net/http"
	"time"
)

// 配置运行时的Serving信息
// Serving信息应该从Option中获取
type ServingInfo struct {
	// HTTP Server Listener
	Listener net.Listener
	
	// HTTP2连接最大的流数目
	// 如果值为0,则使用HTTP2的默认配置
	HTTP2MaxStreamsPerConnection int
	
	// 是否启用HTTP2
	DisableHTTP2 bool
}

func CreateListener(network, addr string, config net.ListenConfig) (net.Listener, int, error) {
	if len(network) == 0 {
		network = "tcp"
	}
	
	ln, err := config.Listen(context.TODO(), network, addr)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to listen on %v: %v", addr, err)
	}
	
	// get port
	tcpAddr, ok := ln.Addr().(*net.TCPAddr)
	if !ok {
		ln.Close()
		return nil, 0, fmt.Errorf("invalid listen address: %q", ln.Addr().String())
	}
	
	return ln, tcpAddr.Port, nil
}

// From K8s

// Serve runs the secure http server. It fails only if certificates cannot be loaded or the initial listen call fails.
// The actual server loop (stoppable by closing stopCh) runs in a go routine, i.e. Serve does not block.
// It returns a stoppedCh that is closed when all non-hijacked active requests have been processed.
// It returns a listenerStoppedCh that is closed when the underlying http Server has stopped listening.
func (s *ServingInfo) Serve(handler http.Handler, shutdownTimeout time.Duration, stopCh <-chan struct{}) (<-chan struct{}, <-chan struct{}, error) {
	if s.Listener == nil {
		return nil, nil, fmt.Errorf("listener must not be nil")
	}
	
	server := &http.Server{
		Addr:           s.Listener.Addr().String(),
		Handler:        handler,
		MaxHeaderBytes: 1 << 20,
		
		IdleTimeout:       90 * time.Second, // matches http.DefaultTransport keep-alive timeout
		ReadHeaderTimeout: 32 * time.Second, // just shy of requestTimeoutUpperBound
	}
	
	// At least 99% of serialized resources in surveyed clusters were smaller than 256kb.
	// This should be big enough to accommodate most API POST requests in a single frame,
	// and small enough to allow a per connection buffer of this size multiplied by `MaxConcurrentStreams`.
	const resourceBody99Percentile = 256 * 1024
	
	http2Options := &http2.Server{
		IdleTimeout: 90 * time.Second, // matches http.DefaultTransport keep-alive timeout
	}
	
	// shrink the per-stream buffer and max framesize from the 1MB default while still accommodating most API POST requests in a single frame
	http2Options.MaxUploadBufferPerStream = resourceBody99Percentile
	http2Options.MaxReadFrameSize = resourceBody99Percentile
	
	// use the overridden concurrent streams setting or make the default of 250 explicit so we can size MaxUploadBufferPerConnection appropriately
	if s.HTTP2MaxStreamsPerConnection > 0 {
		http2Options.MaxConcurrentStreams = uint32(s.HTTP2MaxStreamsPerConnection)
	} else {
		// match http2.initialMaxConcurrentStreams used by clients
		// this makes it so that a malicious client can only open 400 streams before we forcibly close the connection
		// https://github.com/golang/net/commit/b225e7ca6dde1ef5a5ae5ce922861bda011cfabd
		http2Options.MaxConcurrentStreams = 100
	}
	
	// increase the connection buffer size from the 1MB default to handle the specified number of concurrent streams
	http2Options.MaxUploadBufferPerConnection = http2Options.MaxUploadBufferPerStream * int32(http2Options.MaxConcurrentStreams)
	
	if !s.DisableHTTP2 {
		// apply settings to the server
		if err := http2.ConfigureServer(server, http2Options); err != nil {
			return nil, nil, fmt.Errorf("error configuring http2: %v", err)
		}
	}
	
	//logs.Infof("Serving on %s", server.Addr)
	logs.Info("Start Serving", zap.String("server Address", server.Addr))
	return RunServer(server, s.Listener, shutdownTimeout, stopCh)
}

// RunServer spawns a go-routine continuously serving until the stopCh is
// closed.
// It returns a stoppedCh that is closed when all non-hijacked active requests
// have been processed.
// This function does not block
// TODO: make private when insecure serving is gone from the kube-apiserver
func RunServer(
	server *http.Server,
	ln net.Listener,
	shutDownTimeout time.Duration,
	stopCh <-chan struct{},
) (<-chan struct{}, <-chan struct{}, error) {
	if ln == nil {
		return nil, nil, fmt.Errorf("listener must not be nil")
	}
	
	// Shutdown server gracefully.
	serverShutdownCh, listenerStoppedCh := make(chan struct{}), make(chan struct{})
	// TODO: 设计信号
	//go func() {
	//	defer close(serverShutdownCh)
	//	<-stopCh
	//	ctx, cancel := context.WithTimeout(context.Background(), shutDownTimeout)
	//	server.Shutdown(ctx)
	//	cancel()
	//}()
	
	go func() {
		defer utilruntime.HandleCrash()
		defer close(listenerStoppedCh)
		
		var listener net.Listener
		listener = tcpKeepAliveListener{ln}
		//if server.TLSConfig != nil {
		//	listener = tls.NewListener(listener, server.TLSConfig)
		//}
		
		err := server.Serve(listener)
		if err != nil {
			//logs.Errorf("Server Shutdown With Error\t", err.Error())
			logs.Error("Server Shutdown With Error", zap.String("error", err.Error()))
		}
		
		msg := fmt.Sprintf("Stopped listening on %s", ln.Addr().String())
		select {
		case <-stopCh:
			//	logs.Error(msg)
			logs.Error(msg)
		default:
			panic(fmt.Sprintf("%s due to error: %v", msg, err))
		}
		
	}()
	
	return serverShutdownCh, listenerStoppedCh, nil
}

type tcpKeepAliveListener struct {
	net.Listener
}

func (ln tcpKeepAliveListener) Accept() (net.Conn, error) {
	c, err := ln.Listener.Accept()
	if err != nil {
		return nil, err
	}
	if tc, ok := c.(*net.TCPConn); ok {
		tc.SetKeepAlive(true)
		tc.SetKeepAlivePeriod(defaultKeepAlivePeriod)
	}
	return c, nil
}

package main

import (
	"flag"
	"fmt"
	"net/http"
	"time"

	"hit.edu/framework/pkg/apimachinery/runtime"
	"hit.edu/framework/pkg/apimachinery/runtime/schema"
	"hit.edu/framework/pkg/apimachinery/runtime/serializer"
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/client-go/rest"
	"hit.edu/framework/pkg/component-base/logs"
	"hit.edu/framework/test/etcd_sync/informer"
)

func main() {

	var apiserverAddress string
	var serverAddress string
	var apiserverPort int
	var serverPort int
	flag.StringVar(&apiserverAddress, "apiserver-address", "0.0.0.0", "apiserver 监听的 IP 地址 (default 0.0.0.0)")
	flag.IntVar(&apiserverPort, "apiserver-port", 10000, "apiserver 监听的 端口号 (default 10000)")

	flag.StringVar(&serverAddress, "bind-address", "0.0.0.0", "syncserver 监听的 IP 地址 (default 0.0.0.0)")
	flag.IntVar(&serverPort, "bind-port", 14399, "syncserver 监听的端口号 (default 14399)")

	flag.Parse()

	apiserverhost := fmt.Sprintf("http://%s:%d", apiserverAddress, apiserverPort)
	serverhost := fmt.Sprintf("%s:%d", serverAddress, serverPort)

	// fmt.Println(apiserverhost)
	// fmt.Println(serverhost)
	// return
	logs.Init("sync_server")
	scheme := runtime.NewScheme()
	apis.AddToScheme(scheme)
	c := &rest.Config{
		Host:    apiserverhost,
		APIPath: "/apis/resources/v1",
		ContentConfig: rest.ContentConfig{
			AcceptContentTypes: "application/json; charset=UTF-8",
			ContentType:        "application/json; charset=UTF-8",
			GroupVersion: &schema.GroupVersion{
				Group:   "resources",
				Version: "v1",
			},
			NegotiatedSerializer: serializer.NewCodecFactory(scheme),
		},
		UserAgent: "defaultUserAgent",
		Transport: &http.Transport{
			MaxIdleConns:        100,              // 最大空闲连接数
			IdleConnTimeout:     90 * time.Second, // 空闲连接超时时间
			TLSHandshakeTimeout: 10 * time.Second, // TLS 握手超时时间
		},
		Timeout: 10 * time.Second,
	}
	server := &informer.Sever_Config{
		Client: c,
	}
	logs.Infof("apisever run on %s", apiserverhost)
	server.CreateWebHandler(c, serverhost) //本地服务暴露的位置
}

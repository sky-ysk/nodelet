package informer

import (
	"time"
)

func Synctest(config *Config, URLs string) {

	SyncConfig = *config
	URL = URLs
	//TODO： 完成侦听端口，等待其他域的修改请求的功能
	go CreateWebHandler()

	// scheme := runtime.NewScheme()
	// apis.AddToScheme(scheme)

	// c := &rest.Config{
	// 	Host:    "http://localhost:10000",
	// 	APIPath: "/apis/resources/v1",
	// 	ContentConfig: rest.ContentConfig{
	// 		AcceptContentTypes: "application/json; charset=UTF-8",
	// 		ContentType:        "application/json; charset=UTF-8",
	// 		GroupVersion: &schema.GroupVersion{
	// 			Group:   "resources",
	// 			Version: "v1",
	// 		},
	// 		NegotiatedSerializer: serializer.NewCodecFactory(scheme),
	// 	},
	// 	UserAgent: "defaultUserAgent",
	// 	Transport: &http.Transport{
	// 		MaxIdleConns:        100,              // 最大空闲连接数
	// 		IdleConnTimeout:     90 * time.Second, // 空闲连接超时时间
	// 		TLSHandshakeTimeout: 10 * time.Second, // TLS 握手超时时间
	// 	},
	// 	Timeout: 1000 * time.Second,
	// }

	// 创建ClientSet
	// clientSet, err := clients.NewForConfig(config.Client)
	// if err != nil {
	// 	logs.Error(err)
	// }

	// go CreateController(clientSet, &apis.Node{}, "Test")
	// go CreateController(clientSet, &apis.Task{}, "Test")
	time.Sleep(5 * time.Second)
	//EventTestSender("http://127.0.0.1:14399")
	NodeCreateRequestSender()
	NodeGetRequestSender()
	//CreateTestTaskEvents(clientSet)
	//CreateTestNodeEvents(clientSet)
	select {}

}

package main

import (
	"net/http"
	"time"

	"hit.edu/framework/pkg/apimachinery/runtime"
	"hit.edu/framework/pkg/apimachinery/runtime/schema"
	"hit.edu/framework/pkg/apimachinery/runtime/serializer"
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/client-go/rest"
	"hit.edu/framework/test/etcd_sync/informer"
)

func main() {
	// 目前已经实现了域内的监视，只要把需要转发到的域的IP和端口写在这里，然后里面加上域和资源信息即可
	// TODO：把得到的消息转发出去
	//informer.Synctest("http://broker.registry-svc.test.svc.clusterset.local:3001/forward?target=127.0.0.1:14399")
	//informer.Synctest("127.0.0.1:14399")

	//注册资源
	scheme := runtime.NewScheme()
	apis.AddToScheme(scheme)

	c := &rest.Config{
		Host:    "http://localhost:10000",
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
			MaxIdleConns:        10000,
			IdleConnTimeout:     90 * time.Second,
			TLSHandshakeTimeout: 10 * time.Second,
		},
		Timeout: 1000 * time.Second,
	}

	config := &informer.Config{Client: c}

	informer.Synctest(config, "http://127.0.0.1:14399")

	// informer.AddToTargets(config, "broker", "registry-svc.test.svc.clusterset.local", 3001, "172.100.0.109", 4399)
	// informer.AddToTargets(config, "pve2", "registry-svc.test.svc.clusterset.local", 3001, "172.110.0.109", 4399)
	// informer.UpdateTarget(config, "pve2", "registry-svc.test.svc.clusterset.local", 3001, "172.110.0.109", 4299)
	// //informer.DeleteTarget(config, "broker")
	// informer.PrintTargets(*config)
}

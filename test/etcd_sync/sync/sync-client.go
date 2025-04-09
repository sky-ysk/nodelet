package main

import (
	"fmt"
	"net"
	"net/http"
	"time"

	"hit.edu/framework/pkg/apimachinery/runtime"
	"hit.edu/framework/pkg/apimachinery/runtime/schema"
	"hit.edu/framework/pkg/apimachinery/runtime/serializer"
	"hit.edu/framework/pkg/client-go/rest"
)

type Target struct {
	ClusterID   string //域名
	ServiceIP   string //跨域服务的URL
	ServicePort int    //跨域服务的端口
	EtcdIP      string //域内etcdIP
	EtcdPort    int    //域内etcd端口
}

func CreateTarget(clusterID, serviceIP string, servicePort int, etcdIP string, etcdPort int) (*Target, error) {
	if clusterID == "" {
		return nil, fmt.Errorf("ClusterID 不能为空")
	}
	if net.ParseIP(etcdIP) == nil {
		return nil, fmt.Errorf("EtcdIP 不是有效的 IP 地址: %s", etcdIP)
	}
	isValidPort := func(port int) bool {
		return port > 0 && port <= 65535
	}
	if !isValidPort(servicePort) || !isValidPort(etcdPort) {
		return nil, fmt.Errorf("Port 必须在 1 到 65535 之间: %d", servicePort)
	}
	target := &Target{
		ClusterID:   clusterID,
		ServiceIP:   serviceIP,
		ServicePort: servicePort,
		EtcdIP:      etcdIP,
		EtcdPort:    etcdPort,
	}

	return target, nil
}
func CreateConfig(target *Target, scheme *runtime.Scheme) *rest.Config {
	URL := CreateURL(target)
	c := &rest.Config{
		Host:    URL,
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
	return c

}

func CreateURL(target *Target) string {
	url := fmt.Sprintf("http://%s.%s:%d/forward?target=%s:%d", target.ClusterID, target.ServiceIP, target.ServicePort, target.EtcdIP, target.EtcdPort)
	return url
}

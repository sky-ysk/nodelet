package handlers

import (
	"hit.edu/framework/pkg/apimachinery/runtime/schema"
	"hit.edu/framework/pkg/apimachinery/runtime/serializer"
	"hit.edu/framework/pkg/client-go/clients/scheme"
	"hit.edu/framework/pkg/client-go/rest"
	"net/http"
	"time"
)

type ClientSetConfig struct {
	// Host
	Host string
	// Port
	Port int
	// Group Version
	GroupVersion schema.GroupVersion
	//其他配置
}

func NewForConfig(host string, port int, version schema.GroupVersion) *ClientSetConfig {
	return &ClientSetConfig{
		host,
		port,
		version,
	}
}

func (c *ClientSetConfig) GetRestConfig() *rest.Config {
	//host := fmt.Sprintf("https://%s:%d", c.Host, c.Port)
	//apiPath := fmt.Sprintf("/apis/%s/%s", c.GroupVersion.Group, c.GroupVersion.Version)
	//r := &rest.Config{
	//	Host:    host,
	//	APIPath: apiPath,
	//	ContentConfig: rest.ContentConfig{
	//		AcceptContentTypes:   "application/json; charset=UTF-8", //text/plain; charset=UTF-8
	//		ContentType:          "application/json; charset=UTF-8", //application/json; charset=UTF-8
	//		GroupVersion:         &c.GroupVersion,
	//		NegotiatedSerializer: serializer.NewCodecFactory(scheme.Scheme),
	//	},
	//	UserAgent: "defaultUserAgent",
	//	Transport: &http.Transport{
	//		MaxIdleConns:        100,              // 最大空闲连接数
	//		IdleConnTimeout:     90 * time.Second, // 空闲连接超时时间
	//		TLSHandshakeTimeout: 10 * time.Second, // TLS 握手超时时间
	//	},
	//	Timeout: 10 * time.Second,
	//}
	r := &rest.Config{
		Host:    "http://localhost:10000",
		APIPath: "/apis/resources/v1",
		ContentConfig: rest.ContentConfig{
			AcceptContentTypes: "application/json; charset=UTF-8", //text/plain; charset=UTF-8
			ContentType:        "application/json; charset=UTF-8", //application/json; charset=UTF-8
			GroupVersion: &schema.GroupVersion{
				Group:   "resources",
				Version: "v1",
			},
			NegotiatedSerializer: serializer.NewCodecFactory(scheme.Scheme),
		},
		UserAgent: "defaultUserAgent",
		Transport: &http.Transport{
			MaxIdleConns:        100,              // 最大空闲连接数
			IdleConnTimeout:     90 * time.Second, // 空闲连接超时时间
			TLSHandshakeTimeout: 10 * time.Second, // TLS 握手超时时间
		},
		Timeout: 10 * time.Second,
	}
	
	return r
}

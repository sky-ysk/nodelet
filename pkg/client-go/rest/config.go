package rest

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"hit.edu/framework/pkg/apimachinery/runtime"
	"hit.edu/framework/pkg/apimachinery/runtime/schema"
)

type Config struct {
	// host:port pair
	Host string

	//
	APIPath string

	//
	ContentConfig

	// UserAgent is an optional field that specifies the caller of this request.
	UserAgent string

	// From K8s
	// Transport may be used for custom HTTP behavior. This attribute may not
	// be specified with the TLS client certificate options. Use WrapTransport
	// to provide additional per-server middleware behavior.
	Transport http.RoundTripper

	// The maximum length of time to wait before giving up on a server request. A value of zero means no timeout.
	Timeout time.Duration
}

// From K8s
type ContentConfig struct {
	// AcceptContentTypes specifies the types the client will accept and is optional.
	// If not set, ContentType will be used to define the Accept header
	AcceptContentTypes string
	// ContentType specifies the wire format used to communicate with the server.
	// This value will be set as the Accept header on requests made to the server, and
	// as the default content type on any object sent to the server. If not set,
	// "application/json" is used.
	ContentType string
	// GroupVersion is the API version to talk to. Must be provided when initializing
	// a RESTClient directly. When initializing a Client, will be set with the default
	// code version.
	GroupVersion *schema.GroupVersion

	//NegotiatedSerializer用于为多种支持的媒体类型获取编码器和解码器。
	NegotiatedSerializer runtime.NegotiatedSerializer

	// 以下用于跨域header配置
	FlowType  string
	ClusterID string
}

// //先用着，待修改
//RESTClientForConfigAndClient返回一个满足客户端Config对象所请求属性的RESTClient。
//与RESTClientFor不同，RESTClientForConfigAndClient允许传递http。在所有API组和版本之间共享的客户端。
//请注意，http客户端优先于配置的传输值。
//http客户端默认为`http。DefaultClient `if nil。

func RESTClientForConfigAndClient(config *Config, httpClient *http.Client) (*RESTClient, error) {
	// 检查必要参数是否为空
	if config.GroupVersion == nil {
		return nil, fmt.Errorf("GroupVersion is required when initializing a RESTClient")
	}
	if config.NegotiatedSerializer == nil {
		return nil, fmt.Errorf("NegotiatedSerializer is required when initializing a RESTClient")
	}

	//todo: 设置基础 URL 和 API 路径
	baseURL, err := url.Parse(config.Host) // Kubernetes API 服务器的基础地址
	if err != nil {
		return nil, fmt.Errorf("failed to parse base URL: %w", err)
	}
	versionedAPIPath := config.APIPath

	// 配置内容设置，主要用于序列化和数据格式
	clientContent := ClientContentConfig{
		ContentType:  config.ContentConfig.ContentType, // 默认使用 JSON 格式
		GroupVersion: *config.GroupVersion,
		Negotiator:   runtime.NewClientNegotiator(config.NegotiatedSerializer, *config.GroupVersion),
		FlowType:     config.FlowType,
		ClusterID:    config.ClusterID,
	}

	// 创建 RESTClient，处理和 API 服务器的通信
	restClient, err := NewRESTClient(baseURL, versionedAPIPath, clientContent, httpClient)
	if err != nil {
		return nil, fmt.Errorf("failed to create RESTClient: %w", err)
	}
	return restClient, nil
}

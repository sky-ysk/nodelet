package rest

import (
	"hit.edu/framework/pkg/apimachinery/types"
	"net/http"
	"net/url"
	"strings"
	"sync/atomic"

	"hit.edu/framework/pkg/apimachinery/runtime"
	"hit.edu/framework/pkg/apimachinery/runtime/schema"
)

// 基础的REST Client
// 执行Post、Put、Get、Delelte等基础操作

// 定义Rest APIs基本的交互功能
type Interface interface {
	Verb(verb string) *Request
	Post() *Request
	Put() *Request
	Get() *Request
	Delete() *Request
	Patch(pt types.PatchType) *Request
	APIVersion() schema.GroupVersion

	// TODO: 拓展动词
}

type ClientContentConfig struct {
	// TODO: 客户端Content配置参数填写

	//
	AcceptContentTypes string

	//
	GroupVersion schema.GroupVersion

	//
	ContentType string

	// Negotiator is used for obtaining encoders and decoders for multiple
	// supported media types.
	//在request中的body()方法中用到
	Negotiator runtime.ClientNegotiator
}

// 管理基础资源的客户端
type RESTClient struct {

	// 所有请求的根地址
	base *url.URL

	// baseURL后对应的资源组访问地址
	versionedAPIPath string

	// 客户端配置
	// 描述RestClient如何编码和解码Response信息
	content ClientContentConfig

	// HTTP客户端
	Client *http.Client
}

// 创建一个新的REST客户端
func NewRESTClient(baseURL *url.URL, versionedAPIPath string, config ClientContentConfig, client *http.Client) (*RESTClient, error) {
	base := *baseURL
	if !strings.HasSuffix(base.Path, "/") {
		base.Path += "/"
	}
	base.RawQuery = ""
	base.Fragment = ""

	return &RESTClient{
		base:             &base,
		versionedAPIPath: versionedAPIPath,
		content:          config,
		Client:           client,
	}, nil
}

func (c *RESTClient) Verb(verb string) *Request {
	return NewRequest(c).Verb(verb)
}

// Post begins a POST request. Short for c.Verb("POST").
func (c *RESTClient) Post() *Request {
	return c.Verb("POST")
}

// Put begins a PUT request. Short for c.Verb("PUT").
func (c *RESTClient) Put() *Request {
	return c.Verb("PUT")
}

// Patch begins a PATCH request. Short for c.Verb("Patch").
func (c *RESTClient) Patch(pt types.PatchType) *Request {
	return c.Verb("PATCH").SetHeader("Content-Type", string(pt))
}

// Get begins a GET request. Short for c.Verb("GET").
func (c *RESTClient) Get() *Request {
	return c.Verb("GET")
}

// Delete begins a DELETE request. Short for c.Verb("DELETE").
func (c *RESTClient) Delete() *Request {
	return c.Verb("DELETE")
}

// APIVersion returns the APIVersion this RESTClient is expected to use.
func (c *RESTClient) APIVersion() schema.GroupVersion {
	return c.content.GroupVersion
}

// UnsupportedMediaType reports that the server has responded to a request with HTTP 415 Unsupported
// Media Type.
func (p *ClientContentConfig) UnsupportedMediaType(requestContentType string) {
	//TODO： 待完善与检查
}

// requestClientContentConfigProvider observes HTTP 415 (Unsupported Media Type) responses to detect
// that the server does not understand CBOR. Once this has happened, future requests are forced to
// use JSON so they can succeed. This is convenient for client users that want to prefer CBOR, but
// also need to interoperate with older servers so requests do not permanently fail. The clients
// will not default to using CBOR until at least all supported kube-apiservers have enable-CBOR
// locked to true, so this path will be rarely taken. Additionally, all generated clients accessing
// built-in kube resources are forced to protobuf, so those will not degrade to JSON.
type requestClientContentConfigProvider struct {
	base ClientContentConfig

	// Becomes permanently true if a server responds with HTTP 415 (Unsupported Media Type) to a
	// request with "Content-Type" header containing the CBOR media type.
	sawUnsupportedMediaTypeForCBOR atomic.Bool
}

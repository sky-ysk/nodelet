package rest

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"os"
	"path"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"hit.edu/framework/pkg/apimachinery/errors"
	"hit.edu/framework/pkg/apimachinery/runtime"
	"hit.edu/framework/pkg/apimachinery/runtime/schema"
	"hit.edu/framework/pkg/apimachinery/runtime/serializer/streaming"
	"hit.edu/framework/pkg/apimachinery/watch"
	metav1 "hit.edu/framework/pkg/apis/meta"
	restclientwatch "hit.edu/framework/pkg/client-go/rest/watch"
	"hit.edu/framework/pkg/component-base/logs"
	"k8s.io/utils/clock"
)

// 在tryThrottleWithInfo方法中会用到
var (
	// longThrottleLatency defines threshold for logging requests. All requests being
	// throttled (via the provided rateLimiter) for more than longThrottleLatency will
	// be logged.
	longThrottleLatency = 50 * time.Millisecond

	// extraLongThrottleLatency defines the threshold for logging requests at log level 2.
	extraLongThrottleLatency = 1 * time.Second
)

// tryThrottleWithInfo中会用到的
var globalThrottledLogger = &throttledLogger{
	clock: clock.RealClock{},
	settings: []*throttleSettings{
		{
			minLogInterval: 1 * time.Second,
		}, {
			minLogInterval: 10 * time.Second,
		},
	},
}

// Rest请求格式

type Request struct {
	//
	c *RESTClient

	warningHandler WarningHandler

	//因为body()方法加的contentConfig
	contentConfig ClientContentConfig

	timeout    time.Duration
	maxRetries int

	// 基础组件
	verb       string
	pathPrefix string
	subPath    string
	params     url.Values
	headers    http.Header

	// output
	err error

	//
	body      io.Reader
	bodyBytes []byte

	// 资源相关
	namespace    string
	namespaceSet bool
	resource     string
	resourceName string
	subresource  string

	retryFn requestRetryFunc
	// TODO: 参数补全
}

type requestRetryFunc func(maxRetries int) WithRetry

func NewRequest(c *RESTClient) *Request {

	var pathPrefix string
	if c.base != nil {
		pathPrefix = path.Join("/", c.base.Path, c.versionedAPIPath)
	} else {
		pathPrefix = path.Join("/", c.versionedAPIPath)
	}

	var timeout time.Duration
	if c.Client != nil {
		timeout = c.Client.Timeout
	}

	r := &Request{
		c:             c,
		timeout:       timeout,
		pathPrefix:    pathPrefix,
		maxRetries:    10,
		retryFn:       defaultRequestRetryFn,
		contentConfig: c.content,
	}

	switch {
	case len(c.content.AcceptContentTypes) > 0:
		r.SetHeader("Accept", c.content.AcceptContentTypes)
	case len(c.content.ContentType) > 0:
		r.SetHeader("Accept", c.content.ContentType+", */*")
	}
	if len(c.content.FlowType) > 0 {
		r.SetHeader("FlowType", c.content.FlowType)
	}
	if len(c.content.ClusterID) > 0 {
		r.SetHeader("ClusterID", c.content.ClusterID)
	}
	return r
}

func defaultRequestRetryFn(maxRetries int) WithRetry {
	return &withRetry{maxRetries: maxRetries}
}

func (r *Request) SetHeader(key string, values ...string) *Request {
	if r.headers == nil {
		r.headers = http.Header{}
	}
	r.headers.Del(key)
	for _, value := range values {
		r.headers.Add(key, value)
	}
	return r
}

func (r *Request) Verb(verb string) *Request {
	r.verb = verb
	return r
}

func (r *Request) Prefix(segments ...string) *Request {
	if r.err != nil {
		return r
	}
	r.pathPrefix = path.Join(r.pathPrefix, path.Join(segments...))
	return r
}

func (r *Request) Name(resourceName string) *Request {
	// TODO: 格式检查
	if r.err != nil {
		return r
	}
	r.resourceName = resourceName
	return r
}

func (r *Request) Namespace(namespace string) *Request {
	// TODO: 检查Namespace是否为空
	// TODO: 格式检查
	if r.err != nil {
		return r
	}

	if r.namespaceSet {
		r.err = fmt.Errorf("namespace already set")
		return r
	}

	r.namespaceSet = true
	r.namespace = namespace
	return r
}

func (r *Request) Resource(resource string) *Request {
	// TODO: 格式检查
	if r.err != nil {
		return r
	}

	r.resource = resource
	return r
}

func (r *Request) Do(ctx context.Context) Result {
	// TODO: 格式检查
	// TODO: 实现Request协议检查
	var result Result

	//连接服务器的过程
	err := r.request(ctx, func(req *http.Request, resp *http.Response) {
		result = r.transformResponse(ctx, resp, req)
	})

	if err != nil {
		return Result{err: err}
	}
	return result
}

func (r *Request) Timeout(timeout time.Duration) *Request {
	if r.err != nil {
		return r
	}
	r.timeout = timeout
	return r
}

func (r *Request) Watch(ctx context.Context) (watch.Interface, error) {
	w, _, e := r.watchInternal(ctx)
	return w, e
}

func (r *Request) watchInternal(ctx context.Context) (watch.Interface, runtime.Decoder, error) {
	// 检查基础条件
	if r.err != nil {
		return nil, nil, r.err
	}

	client := r.c.Client
	if client == nil {
		client = http.DefaultClient
	}

	// 构建请求
	req, err := r.newHTTPRequest(ctx)
	if err != nil {
		return nil, nil, err
	}
	logs.Trace("client watch req.URL:", req.URL.String())
	// 执行请求
	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, err
	}

	// 检查响应状态
	if resp.StatusCode == http.StatusOK {
		// 返回 Watcher，用于处理流事件
		return r.newStreamWatcher(resp)
	}

	// 处理失败的情况
	if resp != nil {
		return nil, nil, fmt.Errorf("failed to watch: %s, status: %v", r.URL().String(), resp.StatusCode)
	}
	return nil, nil, err
}

func (r *Request) newStreamWatcher(resp *http.Response) (watch.Interface, runtime.Decoder, error) {
	// 获取并解析 Content-Type
	contentType := resp.Header.Get("Content-Type")
	mediaType, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		logs.Error("Unexpected content type from the server")
	}

	// 获取解码器和流式序列化器
	objectDecoder, streamingSerializer, framer, err := r.contentConfig.Negotiator.StreamDecoder(mediaType, params)
	if err != nil {
		return nil, nil, err
	}

	// 构建 Watcher
	frameReader := framer.NewFrameReader(resp.Body)
	watchEventDecoder := streaming.NewDecoder(frameReader, streamingSerializer)
	return watch.NewStreamWatcher(
		restclientwatch.NewDecoder(watchEventDecoder, objectDecoder),
		errors.NewClientErrorReporter(http.StatusInternalServerError, r.verb, "ClientWatchDecoding"),
	), objectDecoder, nil
}

func (r Result) Error() error {
	// TODO: 逻辑实现
	return r.err
}

func (r *Request) Body(obj interface{}) *Request {
	//fmt.Println("Body starting")
	if r.err != nil {
		return r
	}
	switch t := obj.(type) {
	case string:
		data, err := os.ReadFile(t)
		if err != nil {
			r.err = err
			return r
		}
		r.body = nil
		r.bodyBytes = data
	case []byte:
		r.body = nil
		r.bodyBytes = t
	case io.Reader:
		r.body = t
		r.bodyBytes = nil
	case runtime.Object:
		// callers may pass typed interface pointers, therefore we must check nil with reflection
		if reflect.ValueOf(t).IsNil() {
			return r
		}
		encoder, err := r.contentConfig.Negotiator.Encoder(r.contentConfig.ContentType, nil)

		if err != nil {
			r.err = err
			return r
		}
		data, err := runtime.Encode(encoder, t)
		if err != nil {
			r.err = err
			return r
		}
		r.body = nil

		r.bodyBytes = data
		r.SetHeader("Content-Type", r.contentConfig.ContentType)
	default:
		r.err = fmt.Errorf("unknown type used for body: %+v", obj)
	}
	// 暂时隐藏---2005.03.04--何涨溢
	//logs.Trace("Body to string:", string(r.bodyBytes))
	return r
}

// 请求连接到服务器，并在收到服务器响应时调用提供的函数。
func (r *Request) request(ctx context.Context, fn func(*http.Request, *http.Response)) error {
	client := r.c.Client
	if client == nil {
		client = http.DefaultClient
	}

	// 创建请求
	//这里可点开，便于查看运行时的Url
	req, err := r.newHTTPRequest(ctx)
	if err != nil {
		return err
	}
	// 暂时隐藏---2005.03.04--何涨溢
	//logs.Trace(req.URL.String())

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	// 处理响应
	defer readAndCloseResponseBody(resp)
	fn(req, resp)

	return nil
}

// request方法中会用到的
// finalURLTemplate is similar to URL(), but will make all specific parameter values equal
// - instead of name or namespace, "{name}" and "{namespace}" will be used, and all query
// parameters will be reset. This creates a copy of the url so as not to change the
// underlying object.
func (r Request) finalURLTemplate() url.URL {
	newParams := url.Values{}
	v := []string{"{value}"}
	for k := range r.params {
		newParams[k] = v
	}
	r.params = newParams
	u := r.URL()
	if u == nil {
		return url.URL{}
	}

	segments := strings.Split(u.Path, "/")
	groupIndex := 0
	index := 0
	trimmedBasePath := ""
	if r.c.base != nil && strings.Contains(u.Path, r.c.base.Path) {
		p := strings.TrimPrefix(u.Path, r.c.base.Path)
		if !strings.HasPrefix(p, "/") {
			p = "/" + p
		}
		// store the base path that we have trimmed so we can append it
		// before returning the URL
		trimmedBasePath = r.c.base.Path
		segments = strings.Split(p, "/")
		groupIndex = 1
	}
	if len(segments) <= 2 {
		return *u
	}

	const CoreGroupPrefix = "api"
	const NamedGroupPrefix = "apis"
	isCoreGroup := segments[groupIndex] == CoreGroupPrefix
	isNamedGroup := segments[groupIndex] == NamedGroupPrefix
	if isCoreGroup {
		// checking the case of core group with /api/v1/... format
		index = groupIndex + 2
	} else if isNamedGroup {
		// checking the case of named group with /apis/apps/v1/... format
		index = groupIndex + 3
	} else {
		// this should not happen that the only two possibilities are /api... and /apis..., just want to put an
		// outlet here in case more API groups are added in future if ever possible:
		// https://kubernetes.io/docs/concepts/overview/kubernetes-api/#api-groups
		// if a wrong API groups name is encountered, return the {prefix} for url.Path
		u.Path = "/{prefix}"
		u.RawQuery = ""
		return *u
	}
	// switch segLength := len(segments) - index; segLength {
	switch {
	// case len(segments) - index == 1:
	// resource (with no name) do nothing
	case len(segments)-index == 2:
		// /$RESOURCE/$NAME: replace $NAME with {name}
		segments[index+1] = "{name}"
	case len(segments)-index == 3:
		if segments[index+2] == "finalize" || segments[index+2] == "status" {
			// /$RESOURCE/$NAME/$SUBRESOURCE: replace $NAME with {name}
			segments[index+1] = "{name}"
		} else {
			// /namespace/$NAMESPACE/$RESOURCE: replace $NAMESPACE with {namespace}
			segments[index+1] = "{namespace}"
		}
	case len(segments)-index >= 4:
		segments[index+1] = "{namespace}"
		// /namespace/$NAMESPACE/$RESOURCE/$NAME: replace $NAMESPACE with {namespace},  $NAME with {name}
		if segments[index+3] != "finalize" && segments[index+3] != "status" {
			// /$RESOURCE/$NAME/$SUBRESOURCE: replace $NAME with {name}
			segments[index+3] = "{name}"
		}
	}
	u.Path = path.Join(trimmedBasePath, path.Join(segments...))
	return *u
}

// finalURLTemplate 中会用到的
// URL returns the current working URL. Check the result of Error() to ensure
// that the returned URL is valid.
func (r *Request) URL() *url.URL {
	p := r.pathPrefix
	//todo: 这里可以使用命名空间作为url的一部分,可以设置namespace 与 resource
	if r.namespaceSet && len(r.namespace) > 0 {
		p = path.Join(p, "namespaces/", r.namespace)
	}
	if len(r.resource) != 0 {
		p = path.Join(p, strings.ToLower(r.resource))
	}

	// Join trims trailing slashes, so preserve r.pathPrefix's trailing slash for backwards compatibility if nothing was changed
	if len(r.resourceName) != 0 || len(r.subresource) != 0 {
		p = path.Join(p, r.resourceName)
	}

	finalURL := &url.URL{}

	if r.c.base != nil {
		*finalURL = *r.c.base
	}

	finalURL.Path = p
	query := url.Values{}
	for key, values := range r.params {
		for _, value := range values {
			query.Add(key, value)
		}
	}

	// timeout is handled specially here.
	if r.timeout != 0 {
		query.Set("timeout", r.timeout.String())
	}
	//单独处理需要跨域的请求
	if r.c.content.FlowType == "etcd" {
		finalURL.Path = "/forward"
		rawquery := "target=" + r.c.content.TargetURL + p + "?"

		finalURL.RawQuery = rawquery + query.Encode()
		return finalURL

	}
	finalURL.RawQuery = query.Encode()
	return finalURL
}

// tryThrottle方法中会用到的
func (r *Request) tryThrottleWithInfo(ctx context.Context, retryInfo string) error {
	now := time.Now()

	latency := time.Since(now)

	var message string
	switch {
	case len(retryInfo) > 0:
		message = fmt.Sprintf("Waited for %v, %s - request: %s:%s", latency, retryInfo, r.verb, r.URL().String())
	default:
		message = fmt.Sprintf("Waited for %v due to client-side throttling, not priority and fairness, request: %s:%s", latency, r.verb, r.URL().String())
	}

	if latency > longThrottleLatency {
		logs.Error(message)
	}
	if latency > extraLongThrottleLatency {
		// If the rate limiter latency is very high, the log message should be printed at a higher log level,
		// but we use a throttled logger to prevent spamming.
		logs.Error(message)
	}

	return nil
}

type throttledLogger struct {
	clock    clock.PassiveClock
	settings []*throttleSettings
}

type throttleSettings struct {
	minLogInterval time.Duration

	lastLogTime time.Time
	lock        sync.RWMutex
}

func (r *Request) newHTTPRequest(ctx context.Context) (*http.Request, error) {
	var body io.Reader
	switch {
	case r.body != nil && r.bodyBytes != nil:
		return nil, fmt.Errorf("cannot set both body and bodyBytes")
	case r.body != nil:
		body = r.body
	case r.bodyBytes != nil:
		// Create a new reader specifically for this request.
		// Giving each request a dedicated reader allows retries to avoid races resetting the request body.
		body = bytes.NewReader(r.bodyBytes)
	}

	//在r.URL()中可以设置url的命名空间、资源、资源名称，例如resorceName 在查找资源时会用到，但是在创建资源时不会用到
	url := r.URL().String()
	// 暂时隐藏 --2025.3.4 hzy
	//logs.Tracef("url:", url)
	// 暂时隐藏 --2025.3.4 hzy
	//logs.Debugf("body to string:", string(r.bodyBytes))
	req, err := http.NewRequestWithContext(httptrace.WithClientTrace(ctx, newDNSMetricsTrace(ctx)), r.verb, url, body)
	if err != nil {
		return nil, err
	}
	req.Header = r.headers
	return req, nil
}

// newDNSMetricsTrace returns an HTTP trace that tracks time spent on DNS lookups per host.
// This metric is available in client as "rest_client_dns_resolution_duration_seconds".
func newDNSMetricsTrace(ctx context.Context) *httptrace.ClientTrace {
	type dnsMetric struct {
		start time.Time
		host  string
		sync.Mutex
	}
	dns := &dnsMetric{}
	return &httptrace.ClientTrace{
		DNSStart: func(info httptrace.DNSStartInfo) {
			dns.Lock()
			defer dns.Unlock()
			dns.start = time.Now()
			dns.host = info.Host
		},
		DNSDone: func(info httptrace.DNSDoneInfo) {
			dns.Lock()
			defer dns.Unlock()
		},
	}
}

// transformResponse converts an API response into a structured API object
func (r *Request) transformResponse(ctx context.Context, resp *http.Response, req *http.Request) Result {
	// Read the response body
	body, err := io.ReadAll(resp.Body)

	if err != nil {
		// Handle error while reading response body
		logs.Errorf("Error reading response body: %v", err)
		return Result{
			err: fmt.Errorf("unexpected error reading response body: %w", err),
		}
	}

	// Check and decode the response based on content type
	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = r.contentConfig.ContentType
	}

	// Parse content type and get a decoder
	mediaType, params, err := mime.ParseMediaType(contentType)
	//fmt.Println("mediaType:", mediaType)
	if err != nil {
		return Result{err: err}
	}

	// If the content type is "text/plain", we don't need a decoder, just return the body as is
	if mediaType == "text/plain" {
		//fmt.Println("transformResponse 中 mediaType == text/plain，直接返回body")
		return Result{
			body:        body, // Directly return the raw body (bytes)
			contentType: contentType,
			statusCode:  resp.StatusCode,
			warnings:    handleWarnings(resp.Header, r.warningHandler),
		}
	}

	decoder, err := r.contentConfig.Negotiator.Decoder(mediaType, params)
	if err != nil {
		// if we fail to negotiate a decoder, treat this as an unstructured error
		switch {
		case resp.StatusCode == http.StatusSwitchingProtocols:
			// no-op, we've been upgraded
		case resp.StatusCode < http.StatusOK || resp.StatusCode > http.StatusPartialContent:
			return Result{err: r.transformUnstructuredResponseError(resp, req, body)}
		}
		return Result{
			body:        body,
			contentType: contentType,
			statusCode:  resp.StatusCode,
			warnings:    handleWarnings(resp.Header, r.warningHandler),
		}
	}
	switch {
	case resp.StatusCode == http.StatusSwitchingProtocols:
		// no-op, we've been upgraded
	case resp.StatusCode < http.StatusOK || resp.StatusCode > http.StatusPartialContent:
		// calculate an unstructured error from the response which the Result object may use if the caller
		// did not return a structured error.
		retryAfter, _ := retryAfterSeconds(resp)
		err := r.newUnstructuredResponseError(body, isTextResponse(resp), resp.StatusCode, req.Method, retryAfter)
		return Result{
			body:        body,
			contentType: contentType,
			statusCode:  resp.StatusCode,
			decoder:     decoder,
			err:         err,
			warnings:    handleWarnings(resp.Header, r.warningHandler),
		}
	}
	// Return structured result with body and decoder
	return Result{
		body:        body,
		contentType: contentType,
		statusCode:  resp.StatusCode,
		decoder:     decoder,
		warnings:    handleWarnings(resp.Header, r.warningHandler),
	}
}

// TODO: introduce transformation of generic http.Client.Do() errors that separates 4.
func (r *Request) transformUnstructuredResponseError(resp *http.Response, req *http.Request, body []byte) error {
	if body == nil && resp.Body != nil {
		if data, err := io.ReadAll(&io.LimitedReader{R: resp.Body, N: maxUnstructuredResponseTextBytes}); err == nil {
			body = data
		}
	}
	retryAfter, _ := retryAfterSeconds(resp)
	return r.newUnstructuredResponseError(body, isTextResponse(resp), resp.StatusCode, req.Method, retryAfter)
}

// retryAfterSeconds returns the value of the Retry-After header and true, or 0 and false if
// the header was missing or not a valid number.
func retryAfterSeconds(resp *http.Response) (int, bool) {
	if h := resp.Header.Get("Retry-After"); len(h) > 0 {
		if i, err := strconv.Atoi(h); err == nil {
			return i, true
		}
	}
	return 0, false
}

const maxUnstructuredResponseTextBytes = 2048

// newUnstructuredResponseError instantiates the appropriate generic error for the provided input. It also logs the body.
// todo: 错误信息需要完善
func (r *Request) newUnstructuredResponseError(body []byte, isTextResponse bool, statusCode int, method string, retryAfter int) error {
	// cap the amount of output we create
	if len(body) > maxUnstructuredResponseTextBytes {
		body = body[:maxUnstructuredResponseTextBytes]
	}

	message := "unknown"
	if isTextResponse {
		message = strings.TrimSpace(string(body))
	}
	var groupResource schema.GroupResource
	if len(r.resource) > 0 {
		groupResource.Group = r.contentConfig.GroupVersion.Group
		groupResource.Resource = r.resource
	}
	return errors.NewGenericServerResponse(
		statusCode,
		method,
		groupResource,
		r.resourceName,
		message,
		retryAfter,
		true,
	)
}

// 在transformResponse中会用到的
// isTextResponse returns true if the response appears to be a textual media type.
func isTextResponse(resp *http.Response) bool {
	contentType := resp.Header.Get("Content-Type")
	if len(contentType) == 0 {
		return true
	}
	media, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return false
	}
	return strings.HasPrefix(media, "text/")
}

// VersionedParams将获取所提供的对象，使用隐式RESTClient API版本和默认参数编解码器将其序列化为map[string][]string，
// 然后将其作为参数添加到请求中。使用此选项可从客户端库中提供版本化的查询参数。
// VersionedParams不会写入设置了omitempty且为空的查询参数。如果已经设置了参数，则将其附加到（Params和VersionedParams是可加的）。
func (r *Request) VersionedParams(obj runtime.Object, codec runtime.ParameterCodec) *Request {
	return r.SpecificallyVersionedParams(obj, codec, r.contentConfig.GroupVersion)
}

func (r *Request) SpecificallyVersionedParams(obj runtime.Object, codec runtime.ParameterCodec, version schema.GroupVersion) *Request {
	if r.err != nil {
		return r
	}
	params, err := codec.EncodeParameters(obj, version)
	if err != nil {
		r.err = err
		return r
	}
	for k, v := range params {
		if r.params == nil {
			r.params = make(url.Values)
		}
		r.params[k] = append(r.params[k], v...)
	}
	return r
}

// Get将结果作为对象返回，这意味着它通过了解码器。
// 如果返回的对象是Status类型并且具有。状态！=StatusSuccess，Status中的附加信息将用于丰富错误。
func (r Result) Get() (runtime.Object, error) {
	if r.err != nil {
		// Check whether the result has a Status object in the body and prefer that.
		return nil, r.Error()
	}
	if r.decoder == nil {
		return nil, fmt.Errorf("serializer for %s doesn't exist", r.contentType)
	}

	out, _, err := r.decoder.Decode(r.body, nil, nil)

	if err != nil {
		fmt.Println("decode err:", err)
		return nil, err
	}
	switch t := out.(type) {
	case *metav1.Status:
		// any status besides StatusSuccess is considered an error.
		if t.Status != metav1.StatusSuccess {
			logs.Info("出现了除StatusSuccess之外的状态")
		}
	}
	return out, nil
}

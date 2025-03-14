package request

import (
	"context"
	"go.uber.org/zap"
	"hit.edu/framework/pkg/apis/meta"
	metainternalversion "hit.edu/framework/pkg/apis/meta/internalversion"
	metainternalversionscheme "hit.edu/framework/pkg/apis/meta/internalversion/scheme"
	"hit.edu/framework/pkg/component-base/logs"
	"k8s.io/apimachinery/pkg/api/validation/path"
	"k8s.io/apimachinery/pkg/util/sets"
	"net/http"
	"strings"
)

// namespaceSubresources 包含 namespace 的子资源 此列表允许解析器区分命名空间子资源和命名空间资源
var namespaceSubresources = sets.NewString("status")

// verbsWithSelectors 是支持 fieldSelector 和 labelSelector 参数的动词列表
var verbsWithSelectors = sets.NewString("list", "watch", "deletecollection")

type RequestInfoResolver interface {
	NewRequestInfo(req *http.Request) (*RequestInfo, error)
}

// RequestInfo 保存从http请求解析的信息。
type RequestInfo struct {
	// Namespace 命名空间
	Namespace string
	// IsResourceRequest 指示请求是不是针对API资源（或子资源）
	IsResourceRequest bool
	// Path 请求URL路径
	Path string
	// Verb 是与API请求关联的动词，而不是http方法。
	Verb string

	APIPrefix  string
	APIGroup   string
	APIVersion string
	// Resource 请求的资源类型。非Kind
	Resource string
	// Subresource 子资源
	Subresource string
	// Name 对于某些动词为空，但如果请求直接指示一个名称（而不是在正文内容中），则填充此字段。
	Name string
	// Parts 是请求的路径部分，总是以/{resource}/{name}开头
	Parts []string

	// FieldSelector 包含请求中未解析的字段选择器。 它仅在 apiserver 遵循与此请求关联的谓词的字段选择器。
	FieldSelector string
	// LabelSelector 包含请求中未解析的字段选择器。 它仅在 apiserver 遵循与此请求关联的谓词的字段选择器。
	LabelSelector string
}

type RequestInfoFactory struct {
	APIPrefixes sets.String // without leading and trailing slashes
}

// NewRequestInfo 解析来自http请求的消息，为每个请求填写相关信息
// 有效输入:
// /apis/{api-group}/{version}/{resource}
// /apis/{api-group}/{version}/{resource}/{resourceName}
// /apis/{api-group}/{version}/{resource}/{resourceName}/{subresourceName}
// /apis/{api-group}/{version}/namespaces/{namespace}/{resource}
// /apis/{api-group}/{version}/namespaces/{namespace}/{resource}/{resourceName}
// /apis/{api-group}/{version}/namespaces/{namespace}/{resource}/{resourceName}/{subresourceName}

func (r *RequestInfoFactory) NewRequestInfo(req *http.Request) (*RequestInfo, error) {
	requestInfo := RequestInfo{
		IsResourceRequest: false,
		Path:              req.URL.Path,
		Verb:              strings.ToLower(req.Method),
	}

	currentParts := splitPath(req.URL.Path)
	if len(currentParts) < 4 {
		return &requestInfo, nil
	}
	if !r.APIPrefixes.Has(currentParts[0]) {
		return &requestInfo, nil
	}

	requestInfo.APIPrefix = currentParts[0]
	currentParts = currentParts[1:]

	requestInfo.APIGroup = currentParts[0]
	currentParts = currentParts[1:]

	requestInfo.IsResourceRequest = true
	requestInfo.APIVersion = currentParts[0]
	currentParts = currentParts[1:]

	switch req.Method {
	case "POST":
		requestInfo.Verb = "create"
	case "GET":
		requestInfo.Verb = "get"
	case "PUT":
		requestInfo.Verb = "update"
	case "PATCH":
		requestInfo.Verb = "patch"
	case "DELETE":
		requestInfo.Verb = "delete"
	default:
		requestInfo.Verb = ""
	}

	// URL forms: /namespaces/{namespace}/{kind}/*, where parts are adjusted to be relative to kind
	if currentParts[0] == "namespaces" {
		if len(currentParts) > 1 {
			requestInfo.Namespace = currentParts[1]

			// if there is another step after the namespace name and it is not a known namespace subresource
			// move currentParts to include it as a resource in its own right
			if len(currentParts) > 2 && !namespaceSubresources.Has(currentParts[2]) {
				currentParts = currentParts[2:]
			}
		}
	} else {
		requestInfo.Namespace = meta.NamespaceNone
	}

	requestInfo.Parts = currentParts
	switch {
	case len(requestInfo.Parts) >= 3:
		requestInfo.Subresource = requestInfo.Parts[2]
		fallthrough
	case len(requestInfo.Parts) >= 2:
		requestInfo.Name = requestInfo.Parts[1]
		fallthrough
	case len(requestInfo.Parts) >= 1:
		requestInfo.Resource = requestInfo.Parts[0]
	}

	if len(requestInfo.Name) == 0 && requestInfo.Verb == "get" {
		opts := metainternalversion.ListOptions{}
		if err := metainternalversionscheme.ParameterCodec.DecodeParameters(req.URL.Query(), meta.SchemeGroupVersion, &opts); err != nil {
			logs.Error("Couldn't parse request", zap.String("request", req.URL.Query().Encode()))
			opts = metainternalversion.ListOptions{}
			if values := req.URL.Query()["watch"]; len(values) > 0 {
				switch strings.ToLower(values[0]) {
				case "false", "0":
				default:
					opts.Watch = true
				}
			}
		}

		if opts.Watch {
			requestInfo.Verb = "watch"
		} else {
			requestInfo.Verb = "list"
		}

		if opts.FieldSelector != nil {
			if name, ok := opts.FieldSelector.RequiresExactMatch("metadata.name"); ok {
				if len(path.IsValidPathSegmentName(name)) == 0 {
					requestInfo.Name = name
				}
			}
		}
	}

	if len(requestInfo.Name) == 0 && requestInfo.Verb == "delete" {
		requestInfo.Verb = "deletecollection"
	}

	if verbsWithSelectors.Has(requestInfo.Verb) {
		// interestingly these are parsed above, but the current structure there means that if one (or anything) in the
		// listOptions fails to decode, the field and label selectors are lost.
		// therefore, do the straight query param read here.
		if vals := req.URL.Query()["fieldSelector"]; len(vals) > 0 {
			requestInfo.FieldSelector = vals[0]
		}
		if vals := req.URL.Query()["labelSelector"]; len(vals) > 0 {
			requestInfo.LabelSelector = vals[0]
		}
	}
	return &requestInfo, nil
}

type requestInfoKeyType int

const requestInfoKey requestInfoKeyType = iota

// WithRequestInfo 返回父节点的副本，其中设置了请求信息值
func WithRequestInfo(parent context.Context, info *RequestInfo) context.Context {
	return WithValue(parent, requestInfoKey, info)
}

// RequestInfoFrom 返回ctx上的RequestInfo键的值
func RequestInfoFrom(ctx context.Context) (*RequestInfo, bool) {
	info, ok := ctx.Value(requestInfoKey).(*RequestInfo)
	return info, ok
}

// splitPath 返回URL路径的分段
func splitPath(path string) []string {
	path = strings.Trim(path, "/")
	if path == "" {
		return []string{}
	}
	return strings.Split(path, "/")
}

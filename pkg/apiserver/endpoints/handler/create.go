package handler

import (
	"context"
	"hit.edu/framework/pkg/apimachinery/errors"
	"hit.edu/framework/pkg/apimachinery/runtime"
	"hit.edu/framework/pkg/apis/meta"
	metainternalversionscheme "hit.edu/framework/pkg/apis/meta/internalversion/scheme"
	negotiation "hit.edu/framework/pkg/apiserver/endpoints/handler/negotitation"
	"hit.edu/framework/pkg/apiserver/endpoints/handler/responsewriters"
	"hit.edu/framework/pkg/apiserver/registry/rest"
	"net/http"
)

// createHandler 创建资源Handler实现
func createHandler(r rest.NamedCreater, scope *RequestScope, includeName bool) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		
		name, err := scope.Namer.Name(req)
		if err != nil {
			if includeName {
				scope.err(err, w, req)
				return
			}
		}
		//ctx, cancel := context.WithTimeout(ctx, requestTimeoutUpperBound)
		//defer cancel()
		
		s, err := negotiation.NegotiateInputSerializer(req, false, scope.Serializer)
		if err != nil {
			scope.err(err, w, req)
			return
		}
		
		body, err := limitedReadBody(req, 0)
		if err != nil {
			scope.err(err, w, req)
			return
		}
		defaultGVK := scope.Kind
		
		original := r.New()
		
		decodeSerializer := s.Serializer
		decoder := scope.Serializer.DecoderToVersion(decodeSerializer, scope.HubGroupVersion)
		obj, _, err := decoder.Decode(body, &defaultGVK, original)
		
		if len(name) == 0 {
			name, _ = scope.Namer.ObjectName(obj)
		}
		
		options := &meta.CreateOptions{}
		values := req.URL.Query()
		if err := metainternalversionscheme.ParameterCodec.DecodeParameters(values, meta.SchemeGroupVersion, options); err != nil {
			err = errors.NewBadRequest(err.Error())
			scope.err(err, w, req)
			return
		}
		options.TypeMeta.SetGroupVersionKind(meta.SchemeGroupVersion.WithKind("CreateOptions"))
		
		requestFunc := func() (runtime.Object, error) {
			return r.Create(ctx, name, obj, rest.ValidateAllObjectFunc, options)
		}
		result, err := requestFunc()
		if err != nil {
			scope.err(err, w, req)
			return
		}
		code := http.StatusCreated
		status, ok := result.(*meta.Status)
		if ok && status.Code == 0 {
			status.Code = int32(code)
		}
		responsewriters.WriteObjectNegotiated(scope.Serializer, scope, scope.Kind.GroupVersion(), w, req, code, result, false)
	}
}

type namedCreaterAdapter struct {
	rest.Creater
}

func (c *namedCreaterAdapter) Create(ctx context.Context, name string, obj runtime.Object, createValidating rest.ValidateObjectFunc, options *meta.CreateOptions) (runtime.Object, error) {
	return c.Creater.Create(ctx, obj, createValidating)
}

func CreateNamedResource(r rest.NamedCreater, scope *RequestScope) http.HandlerFunc {
	return createHandler(r, scope, true)
}

// CreateResource 返回一个处理资源创建的函数。
func CreateResource(r rest.Creater, scope *RequestScope) http.HandlerFunc {
	return createHandler(&namedCreaterAdapter{r}, scope, false)
}

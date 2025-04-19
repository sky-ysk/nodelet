package handler

import (
	"context"
	"go.uber.org/zap"
	"hit.edu/framework/pkg/apimachinery/errors"
	"hit.edu/framework/pkg/apimachinery/runtime"
	"hit.edu/framework/pkg/apimachinery/runtime/schema"
	"hit.edu/framework/pkg/apis/meta"
	metainternalversionscheme "hit.edu/framework/pkg/apis/meta/internalversion/scheme"
	"hit.edu/framework/pkg/apiserver/endpoints/handler/finisher"
	negotiation "hit.edu/framework/pkg/apiserver/endpoints/handler/negotitation"
	"hit.edu/framework/pkg/apiserver/endpoints/handler/responsewriters"
	"hit.edu/framework/pkg/apiserver/endpoints/request"
	"hit.edu/framework/pkg/apiserver/registry/rest"
	"hit.edu/framework/pkg/component-base/logs"
	"net/http"
)

var namespaceGVR = schema.GroupVersionResource{Group: "resources", Version: "v1", Resource: "namespaces"}

// createHandler 创建资源Handler实现
func createHandler(r rest.NamedCreater, scope *RequestScope, includeName bool) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()

		namespace, name, err := scope.Namer.Name(req)
		if err != nil {
			if includeName {
				// name 是必需的，返回
				logs.Error("name must be specified", zap.Error(err))
				scope.err(err, w, req)
				return
			}

			// 否则，尝试查找命名空间
			namespace, err = scope.Namer.Namespace(req)
			if err != nil {
				logs.Error("get namespace from requestInfo failed", zap.Error(err))
				scope.err(err, w, req)
				return
			}
		}
		ctx, cancel := context.WithTimeout(ctx, requestTimeoutUpperBound)
		defer cancel()

		s, err := negotiation.NegotiateInputSerializer(req, false, scope.Serializer)
		if err != nil {
			logs.Error("get input serializer failed", zap.Error(err))
			scope.err(err, w, req)
			return
		}

		body, err := limitedReadBody(req, 0)
		if err != nil {
			logs.Error("limitedReadBody failed:", err.Error())
			scope.err(err, w, req)
			return
		}
		logs.Info("limitedReadBody succeed", zap.String("len(Body)", string(len(body))))

		defaultGVK := scope.Kind

		original := r.New()

		decodeSerializer := s.Serializer
		decoder := scope.Serializer.DecoderToVersion(decodeSerializer, scope.HubGroupVersion)
		obj, _, err := decoder.Decode(body, &defaultGVK, original)

		if len(name) == 0 {
			_, name, _ = scope.Namer.ObjectName(obj)
		}
		if len(namespace) == 0 && scope.Resource == namespaceGVR {
			namespace = name
		}
		ctx = request.WithNamespace(ctx, namespace)

		if objectMeta, err := meta.Accessor(obj); err == nil {
			//preserveObjectMetaSystemFields := false
			//if c, ok := r.(rest.SubresourceObjectMetaPreserver); ok && len(scope.Subresource) > 0 {
			//	preserveObjectMetaSystemFields = c.PreserveRequestObjectMetaSystemFieldsOnSubresourceCreate()
			//}
			//if !preserveObjectMetaSystemFields {
			//	rest.WipeObjectMetaSystemFields(objectMeta)
			//}
			if err := EnsureObjectNamespaceMatchesRequestNamespace(ExpectedNamespaceForResource(namespace, scope.Resource), objectMeta); err != nil {
				logs.Error(zap.Error(err))
				scope.err(err, w, req)
				return
			}
		}

		options := &meta.CreateOptions{}
		values := req.URL.Query()
		if err := metainternalversionscheme.ParameterCodec.DecodeParameters(values, meta.SchemeGroupVersion, options); err != nil {
			logs.Error("decode CreateOptions failed:", err.Error())
			err = errors.NewBadRequest(err.Error())
			scope.err(err, w, req)
			return
		}
		options.TypeMeta.SetGroupVersionKind(meta.SchemeGroupVersion.WithKind("CreateOptions"))
		logs.Info("decode CreateOptions succeed,about to store object in database")

		requestFunc := func() (runtime.Object, error) {
			return r.Create(ctx, name, obj, rest.ValidateAllObjectFunc, options)
		}
		result, err := finisher.FinishRequest(ctx, requestFunc)
		//result, err := requestFunc()
		if err != nil {
			logs.Error("store object in database failed:", err.Error())
			scope.err(err, w, req)
			return
		}
		logs.Info("store object in database done", zap.String("kind", result.GetObjectKind().GroupVersionKind().Kind))
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

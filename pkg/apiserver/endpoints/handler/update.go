package handler

import (
	"context"
	"go.uber.org/zap"
	"hit.edu/framework/pkg/apimachinery/runtime"
	"hit.edu/framework/pkg/apis/meta"
	metainternalversionscheme "hit.edu/framework/pkg/apis/meta/internalversion/scheme"
	negotiation "hit.edu/framework/pkg/apiserver/endpoints/handler/negotitation"
	"hit.edu/framework/pkg/apiserver/endpoints/handler/responsewriters"
	"hit.edu/framework/pkg/apiserver/endpoints/request"
	"hit.edu/framework/pkg/apiserver/registry/rest"
	"hit.edu/framework/pkg/component-base/logs"
	"k8s.io/apimachinery/pkg/api/errors"
	"net/http"
)

// UpdateResource 返回一个处理资源更新的函数
func UpdateResource(r rest.Updater, scope *RequestScope) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		namespace, name, err := scope.Namer.Name(req)
		if err != nil {
			logs.Error("get name from requestInfo failed", zap.Error(err))
			scope.err(err, w, req)
			return
		}

		ctx, cancel := context.WithTimeout(ctx, requestTimeoutUpperBound)
		defer cancel()
		ctx = request.WithNamespace(ctx, namespace)

		body, err := limitedReadBody(req, 0)
		if err != nil {
			logs.Error("limitedReadBody failed:", err.Error())
			scope.err(err, w, req)
			return
		}
		logs.Info("limitedReadBody succeed", zap.String("len(Body)", string(len(body))))

		options := &meta.UpdateOptions{}
		if err := metainternalversionscheme.ParameterCodec.DecodeParameters(req.URL.Query(), scope.MetaGroupVersion, options); err != nil {
			err = errors.NewBadRequest(err.Error())
			logs.Error("decode UpdateOptions failed:", err.Error())
			scope.err(err, w, req)
			return
		}
		options.TypeMeta.SetGroupVersionKind(meta.SchemeGroupVersion.WithKind("UpdateOptions"))
		logs.Info("decode UpdateOptions succeed,about to update object in database")

		s, err := negotiation.NegotiateInputSerializer(req, false, scope.Serializer)
		if err != nil {
			logs.Error("get input serializer failed", zap.Error(err))
			scope.err(err, w, req)
			return
		}

		defaultGVK := scope.Kind

		original := r.New()

		decodeSerializer := s.Serializer
		decoder := scope.Serializer.DecoderToVersion(decodeSerializer, scope.HubGroupVersion)
		obj, _, err := decoder.Decode(body, &defaultGVK, original)

		if objectMeta, err := meta.Accessor(obj); err == nil {
			// 确保对象上的 namespace 正确无误，如果在对象中设置了冲突的 namespace 则出错
			if err := EnsureObjectNamespaceMatchesRequestNamespace(ExpectedNamespaceForResource(namespace, scope.Resource), objectMeta); err != nil {
				scope.err(err, w, req)
				return
			}
		}

		if err := checkName(obj, name, namespace, scope.Namer); err != nil {
			logs.Error("error occur while checking name", zap.Error(err))
			scope.err(err, w, req)
			return
		}

		transformers := []rest.TransformFunc{}
		transformers = append(transformers, func(_ context.Context, newObj, liveObj runtime.Object) (runtime.Object, error) {
			return newObj, nil
		})
		wasCreated := false
		requestFunc := func() (runtime.Object, error) {
			result, created, err := r.Update(ctx, name, rest.DefaultUpdatedObjectInfo(obj, transformers...), rest.ValidateAllObjectFunc, rest.ValidateAllObjectUpdateFunc, false, options)
			if err != nil {
				return nil, err
			}
			wasCreated = created
			return result, err
		}
		result, err := requestFunc()
		if err != nil {
			logs.Error("update object in database failed:", err.Error())
			scope.err(err, w, req)
			return
		}
		logs.Info("update object in database done", zap.String("kind", result.GetObjectKind().GroupVersionKind().Kind))
		status := http.StatusOK
		if wasCreated {
			status = http.StatusCreated
		}
		responsewriters.WriteObjectNegotiated(scope.Serializer, scope, scope.Kind.GroupVersion(), w, req, status, result, false)
	}
}

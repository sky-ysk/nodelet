package handler

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"hit.edu/framework/pkg/apimachinery/errors"
	"hit.edu/framework/pkg/apimachinery/runtime"
	"hit.edu/framework/pkg/apis/meta"
	metainternalversion "hit.edu/framework/pkg/apis/meta/internalversion"
	metainternalversionscheme "hit.edu/framework/pkg/apis/meta/internalversion/scheme"
	negotiation "hit.edu/framework/pkg/apiserver/endpoints/handler/negotitation"
	"hit.edu/framework/pkg/apiserver/endpoints/handler/responsewriters"
	"hit.edu/framework/pkg/apiserver/endpoints/request"
	"hit.edu/framework/pkg/apiserver/registry/rest"
	"hit.edu/framework/pkg/component-base/logs"
	"net/http"
)

func DeleteResource(r rest.GracefulDeleter, allowsOptions bool, scope *RequestScope) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()

		//尝试获取资源的名称
		namespace, name, err := scope.Namer.Name(req)
		if err != nil {
			logs.Error("get name from requestInfo failed", zap.Error(err))
			scope.err(err, w, req)
			return
		}
		ctx, cancel := context.WithTimeout(ctx, requestTimeoutUpperBound)
		defer cancel()

		ctx = request.WithNamespace(ctx, namespace)

		options := &meta.DeleteOptions{}
		if allowsOptions {
			//从body体中或者url query中解析DeleteOptions
			body, err := limitedReadBody(req, 0)
			if err != nil {
				logs.Error("limitedReadBody failed:", err.Error())
				scope.err(err, w, req)
				return
			}
			logs.Info("limitedReadBody succeed", zap.String("len(Body)", string(len(body))))
			if len(body) > 0 {
				s, err := negotiation.NegotiateInputSerializer(req, false, scope.Serializer)
				if err != nil {
					logs.Error("get input serializer failed", zap.Error(err))
					scope.err(err, w, req)
					return
				}
				defaultGVK := scope.MetaGroupVersion.WithKind("DeleteOptions")
				obj, _, err := scope.Serializer.DecoderToVersion(s.Serializer, defaultGVK.GroupVersion()).Decode(body, &defaultGVK, options)
				if err != nil {
					logs.Error("get decoder failed", zap.Error(err))
					scope.err(err, w, req)
					return
				}
				if obj != options {
					logs.Error("decoded object cannot be converted to DeleteOptions")
					scope.err(fmt.Errorf("decoded object cannot be converted to DeleteOptions"), w, req)
					return
				}
			} else {
				if err := metainternalversionscheme.ParameterCodec.DecodeParameters(req.URL.Query(), meta.SchemeGroupVersion, options); err != nil {
					logs.Error("decode DeleteOptions failed:", err.Error())
					err = errors.NewBadRequest(err.Error())
					scope.err(err, w, req)
					return
				}
			}
		}
		options.TypeMeta.SetGroupVersionKind(meta.SchemeGroupVersion.WithKind("DeleteOptions"))
		logs.Info("decode DeleteOptions succeed,about to delete object in database")

		//s, err := negotiation.NegotiateInputSerializer(req, false, scope.Serializer)
		//if err != nil {
		//	scope.err(err, w, req)
		//	return
		//}

		wasDeleted := false
		requestFunc := func() (runtime.Object, error) {
			result, deleted, err := r.Delete(ctx, name, rest.ValidateAllObjectFunc, options)
			wasDeleted = deleted
			if err != nil {
				return nil, err
			}
			return result, nil
		}
		result, err := requestFunc()
		if err != nil {
			logs.Error("delete object in database failed:", err.Error())
			scope.err(err, w, req)
			return
		}
		logs.Info("delete object in database done", zap.String("kind", result.GetObjectKind().GroupVersionKind().Kind))

		status := http.StatusOK
		if !wasDeleted {
			status = http.StatusAccepted
		}
		if result == nil {
			result = &meta.Status{
				Status: meta.StatusSuccess,
				Code:   int32(status),
				Details: &meta.StatusDetails{
					Name: name,
					Kind: scope.Kind.Kind,
				},
			}
		}
		responsewriters.WriteObjectNegotiated(scope.Serializer, scope, scope.Kind.GroupVersion(), w, req, status, result, false)
	}
}

func DeleteCollection(r rest.CollectionDeleter, checkBody bool, scope *RequestScope) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()

		listOptions := metainternalversion.ListOptions{}
		if err := metainternalversionscheme.ParameterCodec.DecodeParameters(req.URL.Query(), meta.SchemeGroupVersion, &listOptions); err != nil {
			err = errors.NewBadRequest(err.Error())
			scope.err(err, w, req)
			return
		}

		namespace, err := scope.Namer.Namespace(req)
		if err != nil {
			scope.err(err, w, req)
			return
		}

		ctx = request.WithNamespace(ctx, namespace)

		options := &meta.DeleteOptions{}
		if checkBody {
			//从body体中或者url query中解析DeleteOptions
			body, err := limitedReadBody(req, 0)
			if err != nil {
				scope.err(err, w, req)
				return
			}
			logs.Info("limitedReadBody succeed", zap.String("len(Body)", string(len(body))))
			if len(body) > 0 {
				s, err := negotiation.NegotiateInputSerializer(req, false, scope.Serializer)
				if err != nil {
					logs.Error("get input serializer failed", zap.Error(err))
					scope.err(err, w, req)
					return
				}
				defaultGVK := scope.MetaGroupVersion.WithKind("DeleteOptions")
				obj, _, err := scope.Serializer.DecoderToVersion(s.Serializer, defaultGVK.GroupVersion()).Decode(body, &defaultGVK, options)
				if err != nil {
					logs.Error("get decoder failed", zap.Error(err))
					scope.err(err, w, req)
					return
				}
				if obj != options {
					scope.err(fmt.Errorf("decoded object cannot be converted to DeleteOptions"), w, req)
					return
				}
			} else {
				if err := metainternalversionscheme.ParameterCodec.DecodeParameters(req.URL.Query(), meta.SchemeGroupVersion, options); err != nil {
					logs.Error("decode url parameters failed", zap.Error(err))
					err = errors.NewBadRequest(err.Error())
					scope.err(err, w, req)
					return
				}
			}
		}
		options.TypeMeta.SetGroupVersionKind(meta.SchemeGroupVersion.WithKind("DeleteOptions"))
		logs.Info("decode DeleteOptions succeed,about to delete collection in database")

		//s, err := negotiation.NegotiateInputSerializer(req, false, scope.Serializer)
		//if err != nil {
		//	scope.err(err, w, req)
		//	return
		//}

		requestFunc := func() (runtime.Object, error) {
			return r.DeleteCollection(ctx, rest.ValidateAllObjectFunc, options, &listOptions)
		}
		result, err := requestFunc()
		if err != nil {
			logs.Error("delete collection in database", zap.Error(err))
			scope.err(err, w, req)
			return
		}
		logs.Info("delete collection in database done", zap.String("kind", result.GetObjectKind().GroupVersionKind().Kind))
		status := http.StatusOK
		if result == nil {
			result = &meta.Status{
				Status: meta.StatusSuccess,
				Code:   int32(status),
				Details: &meta.StatusDetails{
					Kind: scope.Kind.Kind,
				},
			}
		}
		responsewriters.WriteObjectNegotiated(scope.Serializer, scope, scope.Kind.GroupVersion(), w, req, status, result, false)
	}
}

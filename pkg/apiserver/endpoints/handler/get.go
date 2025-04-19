package handler

import (
	"context"
	"go.uber.org/zap"
	"hit.edu/framework/pkg/apimachinery/errors"
	"hit.edu/framework/pkg/apimachinery/fields"
	"hit.edu/framework/pkg/apimachinery/runtime"
	"hit.edu/framework/pkg/apis/meta"
	metainternalversion "hit.edu/framework/pkg/apis/meta/internalversion"
	metainternalversionscheme "hit.edu/framework/pkg/apis/meta/internalversion/scheme"
	negotiation "hit.edu/framework/pkg/apiserver/endpoints/handler/negotitation"
	"hit.edu/framework/pkg/apiserver/endpoints/handler/responsewriters"
	"hit.edu/framework/pkg/apiserver/endpoints/request"
	"hit.edu/framework/pkg/apiserver/registry/rest"
	"hit.edu/framework/pkg/component-base/logs"
	"math/rand"
	"net/http"
	"time"
)

// getterFunc 执行Storage Get请求
type getterFunc func(ctx context.Context, name string, req *http.Request) (runtime.Object, error)

// getResourceHandler Get请求http Handler
func getResourceHandler(scope *RequestScope, getter getterFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		req = req.WithContext(ctx)
		namespace, name, err := scope.Namer.Name(req)
		if err != nil {
			logs.Error("get name from requestInfo failed", zap.Error(err))
			scope.err(err, w, req)
			return
		}
		ctx = request.WithNamespace(ctx, namespace)
		result, err := getter(ctx, name, req)
		if err != nil {
			logs.Error("Get from storage failed:", err.Error())
			scope.err(err, w, req)
			return
		}
		code := http.StatusOK
		status, ok := result.(*meta.Status)
		if ok && status.Code == 0 {
			status.Code = int32(code)
		}
		logs.Info("About to write a response")

		defer logs.Info("Writing http response done")
		responsewriters.WriteObjectNegotiated(scope.Serializer, scope, scope.Kind.GroupVersion(), w, req, code, result, false)
	}
}

// GetResource 查找一个资源
func GetResource(r rest.Getter, scope *RequestScope) http.HandlerFunc {
	return getResourceHandler(scope,
		func(ctx context.Context, name string, req *http.Request) (runtime.Object, error) {
			options := meta.GetOptions{}
			if values := req.URL.Query(); len(values) > 0 {
				if err := metainternalversionscheme.ParameterCodec.DecodeParameters(values, scope.MetaGroupVersion, &options); err != nil {
					logs.Error("decode GetOptions failed:", err.Error())
					err = errors.NewBadRequest(err.Error())
					return nil, err
				}
				logs.Info("decode GetOptions succeed,about to Get from storage")
			}

			return r.Get(ctx, name, &options)
		})
}

// ListResource 获取资源列表或监听资源
func ListResource(r rest.Lister, rw rest.Watcher, scope *RequestScope, minRequestTimeout time.Duration) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		req = req.WithContext(ctx)

		namespace, err := scope.Namer.Namespace(req)
		if err != nil {
			logs.Error("get name from requestInfo failed", zap.Error(err))
			scope.err(err, w, req)
			return
		}

		hasName := true
		_, name, err := scope.Namer.Name(req)
		if err != nil {
			hasName = false
		}
		ctx = request.WithNamespace(ctx, namespace)

		opts := metainternalversion.ListOptions{}
		if err := metainternalversionscheme.ParameterCodec.DecodeParameters(req.URL.Query(), scope.MetaGroupVersion, &opts); err != nil {
			logs.Error("decode ListOptions failed:", err.Error())
			err = errors.NewBadRequest(err.Error())
			scope.err(err, w, req)
			return
		}
		logs.Info("decode ListOptions succeed")

		//metainternalversion.SetListOptionsDefaults(&opts, utilfeature.DefaultFeatureGate.Enabled(features.WatchList))

		if hasName {
			logs.Info("name specified", zap.String("name", name))
			nameSelector := fields.OneTermEqualSelector("metadata.name", name)

			if opts.FieldSelector != nil && !opts.FieldSelector.Empty() {
				selectedName, ok := opts.FieldSelector.RequiresExactMatch("metadata.name")
				if !ok || name != selectedName {
					scope.err(errors.NewBadRequest("fieldSelector metadata.name doesn't match requested name"), w, req)
					return
				}
			} else {
				opts.FieldSelector = nameSelector
			}
		}

		outputMediaType, _, err := negotiation.NegotiateOutputMediaType(req, scope.Serializer, scope)
		if err != nil {
			logs.Error("get output serializer failed", zap.Error(err))
			scope.err(err, w, req)
			return
		}
		if opts.Watch {
			logs.Info("It's a watch request")
			if rw == nil {
				scope.err(errors.NewMethodNotSupported(scope.Resource.GroupResource(), "watch"), w, req)
				return
			}
			timeout := time.Duration(0)
			if opts.TimeoutSeconds != nil {
				timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
			}
			if timeout == 0 && minRequestTimeout > 0 {
				timeout = time.Duration(float64(minRequestTimeout) * (rand.Float64() + 1.0))
			}

			var emptyVersionedList runtime.Object
			//if isListWatchRequest(opts) {
			//	emptyVersionedList, err = scope.Convertor.ConvertToVersion(r.NewList(), scope.Kind.GroupVersion())
			//	if err != nil {
			//		scope.err(errors.NewInternalError(err), w, req)
			//		return
			//	}
			//}

			logs.Info("Starting watch", zap.String("path", req.URL.Path), zap.String("resourceVersion", opts.ResourceVersion), zap.String("labels", opts.LabelSelector.String()), zap.String("fields", opts.FieldSelector.String()), zap.String("timeout", timeout.String()))
			ctx, cancel := context.WithTimeout(ctx, timeout)
			defer func() { cancel() }()
			watcher, err := rw.Watch(ctx, &opts)
			if err != nil {
				scope.err(err, w, req)
				return
			}

			handler, err := serveWatchHandler(watcher, scope, outputMediaType, req, w, timeout, emptyVersionedList)
			if err != nil {
				logs.Error("error occur while serving watch handler", zap.Error(err))
				scope.err(err, w, req)
				return
			}
			// Invalidate cancel() to defer until serve() is complete.
			deferredCancel := cancel
			serve := func() {
				defer deferredCancel()
				defer watcher.Stop()
				handler.ServeHTTP(w, req)
			}
			serve()
			return
		}
		logs.Info("It's a list request,about to List from storage")
		result, err := r.List(ctx, &opts)
		if err != nil {
			logs.Error("list from storage failed", zap.Error(err))
			scope.err(err, w, req)
			return
		}
		logs.Info("About to write a response")
		defer logs.Info("Writing http response done")
		responsewriters.WriteObjectNegotiated(scope.Serializer, scope, scope.Kind.GroupVersion(), w, req, http.StatusOK, result, false)
	}
}

//func isListWatchRequest(opts meta.ListOptions) bool {
//	return utilfeature.DefaultFeatureGate.Enabled(features.WatchList) && ptr.Deref(opts.SendInitialEvents, false) && opts.AllowWatchBookmarks
//}

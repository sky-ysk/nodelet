package handler

import (
	"context"
	"hit.edu/framework/pkg/apimachinery/errors"
	"hit.edu/framework/pkg/apimachinery/fields"
	"hit.edu/framework/pkg/apimachinery/runtime"
	"hit.edu/framework/pkg/apis/meta"
	metainternalversion "hit.edu/framework/pkg/apis/meta/internalversion"
	metainternalversionscheme "hit.edu/framework/pkg/apis/meta/internalversion/scheme"
	negotiation "hit.edu/framework/pkg/apiserver/endpoints/handler/negotitation"
	"hit.edu/framework/pkg/apiserver/endpoints/handler/responsewriters"
	"hit.edu/framework/pkg/apiserver/registry/rest"
	"k8s.io/apiserver/pkg/features"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	"k8s.io/utils/ptr"
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
		name, err := scope.Namer.Name(req)
		if err != nil {
			scope.err(err, w, req)
			return
		}
		result, err := getter(ctx, name, req)
		if err != nil {
			scope.err(err, w, req)
			return
		}
		code := http.StatusOK
		status, ok := result.(*meta.Status)
		if ok && status.Code == 0 {
			status.Code = int32(code)
		}
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
					err = errors.NewBadRequest(err.Error())
					return nil, err
				}
			}
			
			//TODO:临时测试使用
			//out := &apis.Node{}
			//key := fmt.Sprintf("/%s/%s", scope.Resource.Resource, name)
			//
			//err := r.Get(ctx, key, storage.GetOptions{}, out)
			//if err != nil {
			//	return nil, err
			//}
			
			return r.Get(ctx, name, &options)
		})
}

// ListResource 获取资源列表或监听资源
func ListResource(r rest.Lister, rw rest.Watcher, scope *RequestScope, minRequestTimeout time.Duration) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		req = req.WithContext(ctx)
		
		hasName := true
		name, err := scope.Namer.Name(req)
		if err != nil {
			hasName = false
		}
		
		opts := metainternalversion.ListOptions{}
		if err := metainternalversionscheme.ParameterCodec.DecodeParameters(req.URL.Query(), scope.MetaGroupVersion, &opts); err != nil {
			err = errors.NewBadRequest(err.Error())
			scope.err(err, w, req)
			return
		}
		//metainternalversion.SetListOptionsDefaults(&opts, utilfeature.DefaultFeatureGate.Enabled(features.WatchList))
		
		if hasName {
			//From k8s
			//TODO:指定名称，watch单个资源
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
			scope.err(err, w, req)
			return
		}
		if opts.Watch {
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
			
			//ctx, cancel := context.WithTimeout(ctx, timeout)
			//defer func() { cancel() }()
			watcher, err := rw.Watch(ctx, &opts)
			if err != nil {
				scope.err(err, w, req)
				return
			}
			
			handler, err := serveWatchHandler(watcher, scope, outputMediaType, req, w, timeout, emptyVersionedList)
			if err != nil {
				scope.err(err, w, req)
				return
			}
			// Invalidate cancel() to defer until serve() is complete.
			// deferredCancel := cancel
			serve := func() {
				//defer deferredCancel()
				//defer watcher.Stop()
				handler.ServeHTTP(w, req)
			}
			serve()
			return
		}
		result, err := r.List(ctx, &opts)
		if err != nil {
			scope.err(err, w, req)
			return
		}
		responsewriters.WriteObjectNegotiated(scope.Serializer, scope, scope.Kind.GroupVersion(), w, req, http.StatusOK, result, false)
	}
}

func isListWatchRequest(opts meta.ListOptions) bool {
	return utilfeature.DefaultFeatureGate.Enabled(features.WatchList) && ptr.Deref(opts.SendInitialEvents, false) && opts.AllowWatchBookmarks
}

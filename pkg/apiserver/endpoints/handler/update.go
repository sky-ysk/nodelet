package handler

import (
	"context"
	"hit.edu/framework/pkg/apimachinery/runtime"
	"hit.edu/framework/pkg/apis/meta"
	metainternalversionscheme "hit.edu/framework/pkg/apis/meta/internalversion/scheme"
	negotiation "hit.edu/framework/pkg/apiserver/endpoints/handler/negotitation"
	"hit.edu/framework/pkg/apiserver/endpoints/handler/responsewriters"
	"hit.edu/framework/pkg/apiserver/registry/rest"
	"k8s.io/apimachinery/pkg/api/errors"
	"net/http"
)

// UpdateResource 返回一个处理资源更新的函数
func UpdateResource(r rest.Updater, scope *RequestScope) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		name, err := scope.Namer.Name(req)
		if err != nil {
			scope.err(err, w, req)
			return
		}
		
		ctx, cancel := context.WithTimeout(ctx, requestTimeoutUpperBound)
		defer cancel()
		
		body, err := limitedReadBody(req, 0)
		if err != nil {
			scope.err(err, w, req)
			return
		}
		
		options := &meta.UpdateOptions{}
		if err := metainternalversionscheme.ParameterCodec.DecodeParameters(req.URL.Query(), scope.MetaGroupVersion, options); err != nil {
			err = errors.NewBadRequest(err.Error())
			scope.err(err, w, req)
			return
		}
		options.TypeMeta.SetGroupVersionKind(meta.SchemeGroupVersion.WithKind("UpdateOptions"))
		
		s, err := negotiation.NegotiateInputSerializer(req, false, scope.Serializer)
		if err != nil {
			scope.err(err, w, req)
			return
		}
		
		defaultGVK := scope.Kind
		
		original := r.New()
		
		decodeSerializer := s.Serializer
		decoder := scope.Serializer.DecoderToVersion(decodeSerializer, scope.HubGroupVersion)
		obj, _, err := decoder.Decode(body, &defaultGVK, original)
		
		if err := checkName(obj, name, scope.Namer); err != nil {
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
			scope.err(err, w, req)
			return
		}
		status := http.StatusOK
		if wasCreated {
			status = http.StatusCreated
		}
		responsewriters.WriteObjectNegotiated(scope.Serializer, scope, scope.Kind.GroupVersion(), w, req, status, result, false)
	}
}

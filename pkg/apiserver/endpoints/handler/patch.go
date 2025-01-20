package handler

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	jsonpatch "gopkg.in/evanphx/json-patch.v4"
	"hit.edu/framework/pkg/apimachinery/errors"
	"hit.edu/framework/pkg/apimachinery/runtime"
	"hit.edu/framework/pkg/apimachinery/runtime/schema"
	"hit.edu/framework/pkg/apis/meta"
	metainternalversionscheme "hit.edu/framework/pkg/apis/meta/internalversion/scheme"
	negotiation "hit.edu/framework/pkg/apiserver/endpoints/handler/negotitation"
	"hit.edu/framework/pkg/apiserver/endpoints/handler/responsewriters"
	"hit.edu/framework/pkg/apiserver/endpoints/handler/types"
	"hit.edu/framework/pkg/apiserver/endpoints/request"
	"hit.edu/framework/pkg/apiserver/registry/rest"
	"hit.edu/framework/pkg/component-base/logs"

	"net/http"
)

const (
	// maximum number of operations a single json patch may contain.
	maxJSONPatchOperations = 10000
)

func PatchResource(r rest.Patcher, scope *RequestScope, patchTypes []string) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()

		contentType := req.Header.Get("Content-Type")
		patchType := types.PatchType(contentType)
		supportPatchType := false
		for _, patchtype := range patchTypes {
			if patchtype == string(patchType) {
				supportPatchType = true
				break
			}
		}
		if !supportPatchType {
			scope.err(negotiation.NewUnsupportedMediaTypeError(patchTypes), w, req)
		}

		namespace, name, err := scope.Namer.Name(req)
		if err != nil {
			scope.err(err, w, req)
			return
		}
		ctx = request.WithNamespace(ctx, namespace)

		ctx, cancel := context.WithTimeout(ctx, requestTimeoutUpperBound)
		defer cancel()

		body, err := limitedReadBody(req, 0)
		if err != nil {
			scope.err(err, w, req)
			return
		}
		logs.Info("limitedReadBody succeed", zap.String("len(Body)", string(len(body))))

		options := &meta.PatchOptions{}
		if err := metainternalversionscheme.ParameterCodec.DecodeParameters(req.URL.Query(), meta.SchemeGroupVersion, options); err != nil {
			scope.err(err, w, req)
			return
		}
		logs.Info("decode PatchOptions succeed,about to patch object in database")
		//TODO:Options验证
		options.TypeMeta.SetGroupVersionKind(meta.SchemeGroupVersion.WithKind("PatchOptions"))

		baseContentType := runtime.ContentTypeJSON
		s, ok := runtime.SerializerInfoForMediaType(scope.Serializer.SupportedMediaTypes(), baseContentType)
		if !ok {
			scope.err(fmt.Errorf("no serializer defined for %v", baseContentType), w, req)
			return
		}
		gv := scope.Kind.GroupVersion()
		decodeSerializer := s.Serializer
		codec := runtime.NewCodec(
			scope.Serializer.EncoderForVersion(s.Serializer, gv),
			scope.Serializer.DecoderToVersion(decodeSerializer, scope.HubGroupVersion),
		)

		p := patcher{
			namer:            scope.Namer,
			typer:            scope.Typer,
			unsafeConvertor:  scope.Convertor,
			resource:         scope.Resource,
			kind:             scope.Kind,
			subresource:      scope.Subresource,
			dryRun:           len(options.DryRun) > 0,
			hubGroupVersion:  scope.HubGroupVersion,
			createValidation: rest.ValidateAllObjectFunc,
			updateValidation: rest.ValidateAllObjectUpdateFunc,
			codec:            codec,
			options:          options,
			restPatcher:      r,
			name:             name,
			patchType:        patchType,
			patchBytes:       body,
		}

		result, wasCreated, err := p.patchResource(ctx, scope)
		if err != nil {
			scope.err(err, w, req)
		}
		logs.Info("patch object in database done", zap.String("kind", result.GetObjectKind().GroupVersionKind().Kind))
		status := http.StatusOK
		if wasCreated {
			status = http.StatusCreated
		}
		responsewriters.WriteObjectNegotiated(scope.Serializer, scope, scope.Kind.GroupVersion(), w, req, status, result, false)
	}
}

type patcher struct {
	namer           ScopeNamer
	typer           runtime.ObjectTyper
	unsafeConvertor runtime.ObjectConvertor
	resource        schema.GroupVersionResource
	kind            schema.GroupVersionKind
	subresource     string
	dryRun          bool
	hubGroupVersion schema.GroupVersion

	createValidation rest.ValidateObjectFunc
	updateValidation rest.ValidateObjectUpdateFunc

	codec runtime.Codec

	options *meta.PatchOptions

	// Operation information
	restPatcher rest.Patcher
	name        string
	patchType   types.PatchType
	patchBytes  []byte

	namespace         string
	updatedObjectInfo rest.UpdatedObjectInfo
	mechanism         patchMechanism
	forceAllowCreate  bool
}

type patchMechanism interface {
	applyPatchToCurrentObject(requextContext context.Context, currentObject runtime.Object) (runtime.Object, error)
	createNewObject(requestContext context.Context) (runtime.Object, error)
}

func (p *patcher) patchResource(ctx context.Context, scope *RequestScope) (runtime.Object, bool, error) {
	p.namespace = request.NamespaceValue(ctx)
	switch p.patchType {
	case types.JSONPatchType, types.MergePatchType:
		p.mechanism = &jsonPatcher{
			patcher: p,
		}
	default:
		return nil, false, fmt.Errorf("%v: unimplemented patch type", p.patchType)
	}

	transformers := []rest.TransformFunc{p.applyPatch}

	wasCreated := false
	p.updatedObjectInfo = rest.DefaultUpdatedObjectInfo(nil, transformers...)
	requestFunc := func() (runtime.Object, error) {
		options := patchToUpdateOptions(p.options)
		updateObject, created, updateErr := p.restPatcher.Update(ctx, p.name, p.updatedObjectInfo, p.createValidation, p.updateValidation, p.forceAllowCreate, options)
		wasCreated = created
		return updateObject, updateErr
	}

	result, err := requestFunc()
	return result, wasCreated, err
}

func (p *patcher) applyPatch(ctx context.Context, _, currentObject runtime.Object) (objToUpdate runtime.Object, patchErr error) {
	currentObjectHasUID, err := hasUID(currentObject)
	if err != nil {
		return nil, err
	} else if !currentObjectHasUID {
		objToUpdate, patchErr = p.mechanism.createNewObject(ctx)
	} else {
		objToUpdate, patchErr = p.mechanism.applyPatchToCurrentObject(ctx, currentObject)
	}

	if patchErr != nil {
		return nil, patchErr
	}

	objToUpdateHasUID, err := hasUID(objToUpdate)
	if err != nil {
		return nil, err
	}
	if objToUpdateHasUID && !currentObjectHasUID {
		accessor, err := meta.Accessor(objToUpdate)
		if err != nil {
			return nil, err
		}
		return nil, errors.NewConflict(p.resource.GroupResource(), p.name, fmt.Errorf("uid mismatch: the provided object specified uid %s, and no existing object was found", accessor.GetUID()))
	}

	if objectMeta, err := meta.Accessor(objToUpdate); err == nil {
		// ensure namespace on the object is correct, or error if a conflicting namespace was set in the object
		if err := EnsureObjectNamespaceMatchesRequestNamespace(ExpectedNamespaceForResource(p.namespace, p.resource), objectMeta); err != nil {
			return nil, err
		}
	}

	if err := checkName(objToUpdate, p.name, p.namespace, p.namer); err != nil {
		return nil, err
	}
	return objToUpdate, nil
}

type jsonPatcher struct {
	*patcher
}

func (p *jsonPatcher) applyPatchToCurrentObject(requestContext context.Context, currentObject runtime.Object) (runtime.Object, error) {
	currentObjJS, err := runtime.Encode(p.codec, currentObject)
	if err != nil {
		return nil, err
	}

	// Apply the patch.
	patchedObjJS, _, err := p.applyJSPatch(currentObjJS)
	if err != nil {
		return nil, err
	}

	objToUpdate := p.restPatcher.New()

	if err := runtime.DecodeInto(p.codec, patchedObjJS, objToUpdate); err != nil {
		return nil, err
	}

	if p.options == nil {
		// Provide a more informative error for the crash that would
		// happen on the next line
		panic("PatchOptions required but not provided")
	}
	return objToUpdate, nil
}

func (p *jsonPatcher) createNewObject(_ context.Context) (runtime.Object, error) {
	return nil, errors.NewNotFound(p.resource.GroupResource(), p.name)
}

func (p *jsonPatcher) applyJSPatch(versionedJS []byte) (patchedJS []byte, strictErrors []error, retErr error) {
	switch p.patchType {
	case types.JSONPatchType:

		patchObj, err := jsonpatch.DecodePatch(p.patchBytes)
		if err != nil {
			return nil, nil, errors.NewBadRequest(err.Error())
		}
		if len(patchObj) > maxJSONPatchOperations {
			return nil, nil, errors.NewRequestEntityTooLargeError(
				fmt.Sprintf("The allowed maximum operations in a JSON patch is %d, got %d",
					maxJSONPatchOperations, len(patchObj)))
		}
		patchedJS, err := patchObj.Apply(versionedJS)
		if err != nil {
			return nil, nil, errors.NewGenericServerResponse(http.StatusUnprocessableEntity, "", schema.GroupResource{}, "", err.Error(), 0, false)
		}
		return patchedJS, strictErrors, nil
	case types.MergePatchType:
		patchedJS, retErr = jsonpatch.MergePatch(versionedJS, p.patchBytes)
		if retErr == jsonpatch.ErrBadJSONPatch {
			return nil, nil, errors.NewBadRequest(retErr.Error())
		}
		return patchedJS, strictErrors, retErr
	default:
		return nil, nil, fmt.Errorf("unknown Content-Type header for patch: %v", p.patchType)
	}
}

// patchToUpdateOptions 创建一个 UpdateOptions，其字段值与提供的 PatchOptions 相同。
func patchToUpdateOptions(po *meta.PatchOptions) *meta.UpdateOptions {
	if po == nil {
		return nil
	}
	uo := &meta.UpdateOptions{
		DryRun:          po.DryRun,
		FieldValidation: po.FieldValidation,
	}
	uo.TypeMeta.SetGroupVersionKind(meta.SchemeGroupVersion.WithKind("UpdateOptions"))
	return uo
}

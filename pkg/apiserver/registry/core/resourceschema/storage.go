package resourceschema

import (
	"context"
	"fmt"

	"hit.edu/framework/pkg/apimachinery/fields"
	"hit.edu/framework/pkg/apimachinery/labels"

	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/apis/meta"

	"hit.edu/framework/pkg/apimachinery/runtime"
	"hit.edu/framework/pkg/apiserver/registry/generic"
	genericregistry "hit.edu/framework/pkg/apiserver/registry/generic/registry"
	"hit.edu/framework/pkg/apiserver/registry/rest"
	"hit.edu/framework/pkg/apiserver/registry/storage"
)

type REST struct {
	*genericregistry.Store
}

type StatusREST struct {
	*genericregistry.Store
}

type ResourceSchemaStorage struct {
	ResourceSchema *REST
	Status         *StatusREST
}

func (r *StatusREST) New() runtime.Object {
	return &apis.ResourceSchema{}
}

func (r *StatusREST) Destroy() {
}

func (r *StatusREST) Get(ctx context.Context, name string, options *meta.GetOptions) (runtime.Object, error) {
	return r.Store.Get(ctx, name, options)
}

func (r *StatusREST) Update(ctx context.Context, name string, objInfo rest.UpdatedObjectInfo, createValidation rest.ValidateObjectFunc, updateValidation rest.ValidateObjectUpdateFunc, forceAllowCreate bool, options *meta.UpdateOptions) (runtime.Object, bool, error) {
	return r.Store.Update(ctx, name, objInfo, createValidation, updateValidation, false, options)
}

func NewFunc() runtime.Object {
	return &apis.ResourceSchema{}
}

func NewListFunc() runtime.Object {
	return &apis.ResourceSchemaList{}
}

// GetAttrs 从传入的资源对象中提取标签和字段
func GetAttrs(obj runtime.Object) (labels.Set, fields.Set, error) {
	rs, ok := obj.(*apis.ResourceSchema)
	if !ok {
		return nil, nil, fmt.Errorf("not a ResourceSchema")
	}
	return labels.Set(rs.ObjectMeta.Labels), generic.ObjectMetaFieldsSet(&rs.ObjectMeta, true), nil
}

func Match(label labels.Selector, field fields.Selector) storage.SelectionPredicate {
	return storage.SelectionPredicate{
		Label:    label,
		Field:    field,
		GetAttrs: GetAttrs,
	}
}

func NewResourceSchemaStorage(optsGetter generic.RESTOptionsGetter) (ResourceSchemaStorage, error) {
	store := &genericregistry.Store{
		NewFunc:                   NewFunc,
		NewListFunc:               NewListFunc,
		PredicateFunc:             Match,
		DefaultQualifiedResource:  apis.Resource("resourceschemas"),
		SingularQualifiedResource: apis.Resource("resourceschema"),

		CreateStrategy: thisStrategy,
		UpdateStrategy: thisStrategy,
		DeleteStrategy: thisStrategy,
	}

	options := &generic.StoreOptions{
		RESTOptions: optsGetter,
		AttrFunc:    GetAttrs,
	}

	if err := store.CompleteWithOptions(options); err != nil {
		return ResourceSchemaStorage{}, err
	}

	statusStore := *store
	statusStore.UpdateStrategy = thisStrategy

	resourceSchemaREST := &REST{Store: store}
	statusREST := &StatusREST{Store: &statusStore}

	return ResourceSchemaStorage{
		ResourceSchema: resourceSchemaREST,
		Status:         statusREST,
	}, nil
}

package node

import (
	"context"
	"fmt"
	"hit.edu/framework/pkg/apimachinery/fields"
	"hit.edu/framework/pkg/apimachinery/labels"
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/apis/meta"
	
	"hit.edu/framework/pkg/apimachinery/runtime"
	//"hit.edu/framework/pkg/apiserver/registry/core/rest"
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

type NodeStorage struct {
	Node   *REST
	Status *StatusREST
	Spec   *SpecREST
}
type SpecREST struct {
	*genericregistry.Store
}

// 对节点状态进行操作
func (r *SpecREST) New() runtime.Object {
	return &apis.Node{}
}
func (r *SpecREST) Destroy() {
}
func (r *SpecREST) Get(ctx context.Context, name string, options *meta.GetOptions) (runtime.Object, error) {
	return r.Store.Get(ctx, name, options)
}
func (r *SpecREST) Update(ctx context.Context, name string, objInfo rest.UpdatedObjectInfo, createValidation rest.ValidateObjectFunc, updateValidation rest.ValidateObjectUpdateFunc, forceAllowCreate bool, options *meta.UpdateOptions) (runtime.Object, bool, error) {
	return r.Store.Update(ctx, name, objInfo, createValidation, updateValidation, false, options)
}
func (r *StatusREST) New() runtime.Object {
	return &apis.Node{}
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
	return &apis.Node{}
}
func NewListFunc() runtime.Object {
	return &apis.NodeList{}
}

// GetAttrs 从传入的资源对象中提取标签和字段
func GetAttrs(obj runtime.Object) (labels.Set, fields.Set, error) {
	node, ok := obj.(*apis.Node)
	if !ok {
		return nil, nil, fmt.Errorf("not a node")
	}
	return labels.Set(node.ObjectMeta.Labels), generic.ObjectMetaFieldsSet(&node.ObjectMeta), nil
}
func Match(label labels.Selector, field fields.Selector) storage.SelectionPredicate {
	return storage.SelectionPredicate{
		Label:    label,
		Field:    field,
		GetAttrs: GetAttrs,
	}
}

func NewNodeStorage(optsGetter generic.RESTOptionsGetter) (NodeStorage, error) {
	
	store := &genericregistry.Store{
		NewFunc:                   NewFunc,
		NewListFunc:               NewListFunc,
		PredicateFunc:             Match,
		DefaultQualifiedResource:  apis.Resource("nodes"),
		SingularQualifiedResource: apis.Resource("node"),
		
		CreateStrategy: thisStrategy,
		UpdateStrategy: thisStrategy,
		DeleteStrategy: thisStrategy,
		//Storage:
	}
	options := &generic.StoreOptions{
		RESTOptions: optsGetter,
		AttrFunc:    GetAttrs,
	}
	if err := store.CompleteWithOptions(options); err != nil {
		return NodeStorage{}, err
	}
	statusStore := *store
	statusStore.UpdateStrategy = thisStrategy
	
	specStore := *store
	specStore.UpdateStrategy = thisStrategy
	
	nodeREST := &REST{Store: store}
	statusREST := &StatusREST{Store: &statusStore}
	specREST := &SpecREST{Store: &specStore}
	
	return NodeStorage{
		Node:   nodeREST,
		Status: statusREST,
		Spec:   specREST,
	}, nil
}

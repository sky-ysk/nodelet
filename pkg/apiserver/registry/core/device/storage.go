package device

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
type DeviceStorage struct {
	Device *REST
	Status *StatusREST
	Spec   *SpecREST
}
type SpecREST struct {
	*genericregistry.Store
}

func (r *SpecREST) New() runtime.Object {
	return &apis.Device{}
}
func (r *SpecREST) Destroy() {
}
func (r *SpecREST) Get(ctx context.Context, name string, options *meta.GetOptions) (runtime.Object, error) {
	return r.Store.Get(ctx, name, options)
}
func (r *SpecREST) Update(ctx context.Context, name string, objInfo rest.UpdatedObjectInfo, createValidation rest.ValidateObjectFunc, updateValidation rest.ValidateObjectUpdateFunc, forceAllowCreate bool, options *meta.UpdateOptions) (runtime.Object, bool, error) {
	return r.Store.Update(ctx, name, objInfo, createValidation, updateValidation, false, options)
}

// 对节点状态进行操作
func (r *StatusREST) New() runtime.Object {
	return &apis.Device{}
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
	return &apis.Device{}
}

func NewListFunc() runtime.Object {
	return &apis.DeviceList{}
}

// GetAttrs 从传入的资源对象中提取标签和字段
func GetAttrs(obj runtime.Object) (labels.Set, fields.Set, error) {
	Device, ok := obj.(*apis.Device)
	if !ok {
		return nil, nil, fmt.Errorf("not a Device")
	}
	return labels.Set(Device.ObjectMeta.Labels), generic.ObjectMetaFieldsSet(&Device.ObjectMeta, true), nil
}
func Match(label labels.Selector, field fields.Selector) storage.SelectionPredicate {
	return storage.SelectionPredicate{
		Label:    label,
		Field:    field,
		GetAttrs: GetAttrs,
	}
}

func NewDeviceStorage(optsGetter generic.RESTOptionsGetter) (DeviceStorage, error) {

	store := &genericregistry.Store{
		NewFunc:                   NewFunc,
		NewListFunc:               NewListFunc,
		PredicateFunc:             Match,
		DefaultQualifiedResource:  apis.Resource("Devices"),
		SingularQualifiedResource: apis.Resource("Device"),

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
		return DeviceStorage{}, err
	}
	statusStore := *store
	statusStore.UpdateStrategy = thisStrategy
	specStore := *store
	specStore.UpdateStrategy = thisStrategy

	DeviceREST := &REST{Store: store}
	statusREST := &StatusREST{Store: &statusStore}
	specREST := &SpecREST{Store: &specStore}
	return DeviceStorage{
		Device: DeviceREST,
		Status: statusREST,
		Spec:   specREST,
	}, nil
}

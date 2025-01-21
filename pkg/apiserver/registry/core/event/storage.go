package event

import (
	"fmt"

	"hit.edu/framework/pkg/apimachinery/fields"
	"hit.edu/framework/pkg/apimachinery/labels"

	apis "hit.edu/framework/pkg/apis/cores"

	"hit.edu/framework/pkg/apimachinery/runtime"
	//"hit.edu/framework/pkg/apiserver/registry/core/rest"
	"hit.edu/framework/pkg/apiserver/registry/generic"
	genericregistry "hit.edu/framework/pkg/apiserver/registry/generic/registry"
	"hit.edu/framework/pkg/apiserver/registry/storage"
)

type REST struct {
	*genericregistry.Store
}

type EventStorage struct {
	Event *REST
}

func NewFunc() runtime.Object {
	return &apis.Event{}
}

func NewListFunc() runtime.Object {
	return &apis.EventList{}
}

// GetAttrs 从传入的资源对象中提取标签和字段
func GetAttrs(obj runtime.Object) (labels.Set, fields.Set, error) {
	Event, ok := obj.(*apis.Event)
	if !ok {
		return nil, nil, fmt.Errorf("not a Event")
	}
	return labels.Set(Event.ObjectMeta.Labels), generic.ObjectMetaFieldsSet(&Event.ObjectMeta, true), nil
}
func Match(label labels.Selector, field fields.Selector) storage.SelectionPredicate {
	return storage.SelectionPredicate{
		Label:    label,
		Field:    field,
		GetAttrs: GetAttrs,
	}
}

func NewEventStorage(optsGetter generic.RESTOptionsGetter) (EventStorage, error) {

	store := &genericregistry.Store{
		NewFunc:                   NewFunc,
		NewListFunc:               NewListFunc,
		PredicateFunc:             Match,
		DefaultQualifiedResource:  apis.Resource("Events"),
		SingularQualifiedResource: apis.Resource("Event"),

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
		return EventStorage{}, err
	}
	statusStore := *store
	statusStore.UpdateStrategy = thisStrategy
	specStore := *store
	specStore.UpdateStrategy = thisStrategy

	EventREST := &REST{Store: store}
	return EventStorage{
		Event: EventREST,
	}, nil
}

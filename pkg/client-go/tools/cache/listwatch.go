package cache

import (
	"context"
	"hit.edu/framework/pkg/apimachinery/fields"
	"hit.edu/framework/pkg/apimachinery/runtime"
	"hit.edu/framework/pkg/apimachinery/watch"
	metav1 "hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/client-go/clients/scheme"
	"hit.edu/framework/pkg/client-go/rest"
)

type Lister interface {
	// TODO: List对象配置
	// 返回一个List类型的Obejct
	List(options metav1.ListOptions) (runtime.Object, error)
}

type Watcher interface {
	//
	Watch(options metav1.ListOptions) (watch.Interface, error)
}

type ListerWatcher interface {
	Lister
	Watcher
}

// ListFunc knows how to list resources
type ListFunc func(options metav1.ListOptions) (runtime.Object, error)

// WatchFunc knows how to watch resources
type WatchFunc func(options metav1.ListOptions) (watch.Interface, error)

type ListWatch struct {
	ListFunc        ListFunc
	WatchFunc       WatchFunc
	DisableChunking bool
}

// TODO: 增加List Options

// List a set of apiserver resources
func (lw *ListWatch) List(options metav1.ListOptions) (runtime.Object, error) {
	// ListWatch is used in Reflector, which already supports pagination.
	// Don't paginate here to avoid duplication.
	return lw.ListFunc(options)
}

// Watch a set of apiserver resources
func (lw *ListWatch) Watch(options metav1.ListOptions) (watch.Interface, error) {
	return lw.WatchFunc(options)
}

// Getter interface knows how to access Get method from RESTClient.
type Getter interface {
	Get() *rest.Request
}

// NewListWatchFromClient creates a new ListWatch from the specified client, resource, namespace and field selector.
func NewListWatchFromClient(c Getter, resource string, namespace string, fieldSelector fields.Selector) *ListWatch {
	optionsModifier := func(options *metav1.ListOptions) {
		options.FieldSelector = fieldSelector.String()
	}
	return NewFilteredListWatchFromClient(c, resource, namespace, optionsModifier)
}

func NewFilteredListWatchFromClient(c Getter, resource string, namespace string, optionsModifier func(options *metav1.ListOptions)) *ListWatch {
	listFunc := func(options metav1.ListOptions) (runtime.Object, error) {
		optionsModifier(&options)
		return c.Get().
			Namespace(namespace).
			Resource(resource).
			VersionedParams(&options, scheme.ParameterCodec).
			Do(context.TODO()).
			Get()
	}
	watchFunc := func(options metav1.ListOptions) (watch.Interface, error) {
		options.Watch = true
		options.ProgressNotify = true
		optionsModifier(&options)
		return c.Get().
			Namespace(namespace).
			Resource(resource).
			VersionedParams(&options, scheme.ParameterCodec).
			Watch(context.TODO())
	}
	return &ListWatch{ListFunc: listFunc, WatchFunc: watchFunc}
}

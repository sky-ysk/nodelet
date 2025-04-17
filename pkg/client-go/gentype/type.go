package gentype

import (
	"context"
	"time"

	"hit.edu/framework/pkg/apimachinery/runtime"
	"hit.edu/framework/pkg/apimachinery/types"
	"hit.edu/framework/pkg/apimachinery/watch"
	"hit.edu/framework/pkg/apis/meta"
	metav1 "hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/client-go/rest"
)

// 各类资源的形式化生成过程描述

type objectWithMeta interface {
	runtime.Object
	meta.Object
}

type Client[T objectWithMeta] struct {
	resource  string
	client    rest.Interface
	namespace string
	newObject func() T

	parameterCodec runtime.ParameterCodec
}

// ClientWithList represents a client with support for lists.
type ClientWithList[T objectWithMeta, L runtime.Object] struct {
	*Client[T]
	alsoLister[T, L]
}

// Helper types for composition
type alsoLister[T objectWithMeta, L runtime.Object] struct {
	client  *Client[T]
	newList func() L
}

func NewClient[T objectWithMeta](
	resource string, client rest.Interface, namespace string, parameterCodec runtime.ParameterCodec, objectCreator func() T) *Client[T] {
	return &Client[T]{
		resource:       resource,
		client:         client,
		namespace:      namespace,
		newObject:      objectCreator,
		parameterCodec: parameterCodec,
	}
}

// NewClientWithList constructs a namespaced client with support for lists.
func NewClientWithList[T objectWithMeta, L runtime.Object](
	resource string, client rest.Interface, namespace string, parameterCodec runtime.ParameterCodec, objectCreator func() T, listCreator func() L) *ClientWithList[T, L] {
	typeClient := NewClient[T](resource, client, namespace, parameterCodec, objectCreator)
	return &ClientWithList[T, L]{
		typeClient,
		alsoLister[T, L]{typeClient, listCreator},
	}
}

// Get 获取资源的名称，并返回相应的对象，如果有错误，则返回错误。
func (c *Client[T]) Get(ctx context.Context, name string, options metav1.GetOptions) (T, error) {
	result := c.newObject()
	err := c.client.Get().
		Namespace(c.namespace).
		Resource(c.resource).
		Name(name).
		VersionedParams(&options, c.parameterCodec).
		Do(ctx).
		Into(result)
	return result, err
}

func (c *Client[T]) Watch(ctx context.Context, options metav1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if options.TimeoutSeconds != nil {
		timeout = time.Duration(*options.TimeoutSeconds) * time.Second
	}
	options.Watch = true
	options.ProgressNotify = true
	return c.client.Get().
		Namespace(c.namespace).
		Resource(c.resource).
		VersionedParams(&options, c.parameterCodec).
		Timeout(timeout).
		Watch(ctx)
}

func (c *Client[T]) Create(ctx context.Context, obj T, options metav1.CreateOptions) (T, error) {
	result := c.newObject()
	err := c.client.Post().
		Namespace(c.namespace).
		Resource(c.resource).
		VersionedParams(&options, c.parameterCodec).
		Body(obj).
		Do(ctx).
		Into(result)
	return result, err
}

// TODO: VersionedParams的实现，以及result定义时应定义为 c.newList()
func (l *alsoLister[T, L]) List(ctx context.Context, options metav1.ListOptions) (L, error) {
	list := l.newList()
	var timeout time.Duration
	if options.TimeoutSeconds != nil {
		timeout = time.Duration(*options.TimeoutSeconds) * time.Second
	}
	err := l.client.client.Get().
		Namespace(l.client.namespace).
		Resource(l.client.resource).
		VersionedParams(&options, l.client.parameterCodec).
		Timeout(timeout).
		Do(ctx).
		Into(list)
	return list, err
}

// TODO: VersionedParams 与 body 的实现
func (c *Client[T]) Update(ctx context.Context, obj T, options metav1.UpdateOptions) (T, error) {
	result := c.newObject()
	err := c.client.Put().
		Namespace(c.namespace).
		Resource(c.resource).
		Name(obj.GetName()).
		VersionedParams(&options, c.parameterCodec).
		Body(obj).
		Do(ctx).
		Into(result)
	return result, err
}

func (c *Client[T]) Delete(ctx context.Context, name string, options metav1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.namespace).
		Resource(c.resource).
		Name(name).
		Body(&options).
		Do(ctx).
		Error()
}

// DeleteCollection deletes a collection of objects.
func (l *alsoLister[T, L]) DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	var timeout time.Duration
	if listOpts.TimeoutSeconds != nil {
		timeout = time.Duration(*listOpts.TimeoutSeconds) * time.Second
	}
	return l.client.client.Delete().
		Namespace(l.client.namespace).
		Resource(l.client.resource).
		VersionedParams(&listOpts, l.client.parameterCodec).
		Timeout(timeout).
		Body(&opts).
		Do(ctx).
		Error()
}

// Patch applies the patch and returns the patched resource.
func (c *Client[T]) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (T, error) {
	result := c.newObject()
	err := c.client.Patch(pt).
		Namespace(c.namespace).
		Resource(c.resource).
		Name(name).
		VersionedParams(&opts, c.parameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return result, err
}

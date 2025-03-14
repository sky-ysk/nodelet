package request

import (
	"context"
)

// The key type is unexported to prevent collisions
type key int

const (
	// namespaceKey is the context key for the request namespace.
	namespaceKey key = iota

	// userKey is the context key for the request user.
	userKey
)

// NewContext 实例化基本上下文对象
func NewContext() context.Context {
	return context.TODO()
}

// WithValue 返回父节点的副本，其中与key关联的值为val。
func WithValue(parent context.Context, key interface{}, val interface{}) context.Context {
	return context.WithValue(parent, key, val)
}

// WithNamespace returns a copy of parent in which the namespace value is set
func WithNamespace(parent context.Context, namespace string) context.Context {
	return WithValue(parent, namespaceKey, namespace)
}

// NamespaceFrom returns the value of the namespace key on the ctx
func NamespaceFrom(ctx context.Context) (string, bool) {
	namespace, ok := ctx.Value(namespaceKey).(string)
	return namespace, ok
}

// NamespaceValue returns the value of the namespace key on the ctx, or the empty string if none
func NamespaceValue(ctx context.Context) string {
	namespace, _ := NamespaceFrom(ctx)
	return namespace
}

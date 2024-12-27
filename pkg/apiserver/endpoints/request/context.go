package request

import "context"

// NewContext 实例化基本上下文对象
func NewContext() context.Context {
	return context.TODO()
}

// WithValue 返回父节点的副本，其中与key关联的值为val。
func WithValue(parent context.Context, key interface{}, val interface{}) context.Context {
	return context.WithValue(parent, key, val)
}

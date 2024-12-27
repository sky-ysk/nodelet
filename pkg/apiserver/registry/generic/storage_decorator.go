package generic

import (
	"hit.edu/framework/pkg/apimachinery/runtime"
	"hit.edu/framework/pkg/apiserver/registry/storage"
	"hit.edu/framework/pkg/apiserver/registry/storage/storagebackend"
)

// StorageDecorator 是一个用于创建storage.Interface和清理函数的函数类型
type StorageDecorator func(
	config *storagebackend.ConfigForResource,
	resourcePrefix string,
	keyFunc func(obj runtime.Object) (string, error),
	newFunc func() runtime.Object,
	newListFunc func() runtime.Object,
	getAttrsFunc storage.AttrFunc) (storage.Interface, storagebackend.DestroyFunc, error)

// UndecoratedStorage 返回未经装饰的存储实例
func UndecoratedStorage(
	config *storagebackend.ConfigForResource,
	resourcePrefix string,
	keyFunc func(obj runtime.Object) (string, error),
	newFunc func() runtime.Object,
	newListFunc func() runtime.Object,
	getAttrsFunc storage.AttrFunc) (storage.Interface, storagebackend.DestroyFunc, error) {
	return NewRawStorage(config, newFunc, newListFunc, resourcePrefix)
}

// NewRawStorage 创建实际的存储接口
func NewRawStorage(config *storagebackend.ConfigForResource, newFunc, newListFunc func() runtime.Object, resourcePrefix string) (storage.Interface, storagebackend.DestroyFunc, error) {
	return storagebackend.Create(*config, newFunc, newListFunc, resourcePrefix)
}

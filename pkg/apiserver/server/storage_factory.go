package server

import (
	"hit.edu/framework/pkg/apimachinery/runtime"
	"hit.edu/framework/pkg/apimachinery/runtime/schema"
	"hit.edu/framework/pkg/apiserver/registry/storage/storagebackend"
)

// StorageFactory 是用于查找给定 GroupResource 的存储的接口
type StorageFactory interface {
	// NewConfig 查找给定组和资源的存储目标.如果组未配置存储目标，它会返回错误.
	NewConfig(groupResource schema.GroupResource, example runtime.Object) (*storagebackend.ConfigForResource, error)
	
	// ResourcePrefix 返回 GroupResource 的重写资源前缀
	ResourcePrefix(groupResource schema.GroupResource) string
}

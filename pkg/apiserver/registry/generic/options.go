// options.go提供了资源的存储配置和选项，支持了资源注册的灵活配置
package generic

import (
	"hit.edu/framework/pkg/apimachinery/runtime/schema"
	"time"
	
	"hit.edu/framework/pkg/apimachinery/runtime"
	
	"hit.edu/framework/pkg/apiserver/registry/storage"
	"hit.edu/framework/pkg/apiserver/registry/storage/storagebackend"
	
	flowcontrolrequest "k8s.io/apiserver/pkg/util/flowcontrol/request"
)

// RESTOptions 提供了资源的通用存储选项，定义了与etcd交互的行为
type RESTOptions struct {
	StorageConfig *storagebackend.ConfigForResource
	Decorator     StorageDecorator
	
	EnableGarbageCollection   bool
	DeleteCollectionWorkers   int
	ResourcePrefix            string
	CountMetricPollPeriod     time.Duration
	StorageObjectCountTracker flowcontrolrequest.StorageObjectCountTracker
}

// 实现了RESTOptionsGetter接口，用于返回当前的RESTOptions
func (opts RESTOptions) GetRESTOptions(schema.GroupResource, runtime.Object) (RESTOptions, error) {
	return opts, nil
}

type RESTOptionsGetter interface {
	GetRESTOptions(resource schema.GroupResource, example runtime.Object) (RESTOptions, error)
}

// StoreOptions 定义了存储的附加配置
type StoreOptions struct {
	RESTOptions RESTOptionsGetter
	AttrFunc    storage.AttrFunc
}

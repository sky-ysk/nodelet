package storagebackend

import (
	"context"
	"fmt"
	
	clientv3 "go.etcd.io/etcd/client/v3"
	"hit.edu/framework/pkg/apimachinery/runtime"
	"hit.edu/framework/pkg/apiserver/registry/storage"
)

type DestroyFunc func()

// Create creates a storage backend based on given config.
func Create(c ConfigForResource, newFunc, newListFunc func() runtime.Object, resourcePrefix string) (storage.Interface, DestroyFunc, error) {
	switch c.Type {
	case StorageTypeETCD2:
		return nil, nil, fmt.Errorf("%s is no longer a supported storage backend", c.Type)
	case StorageTypeUnset, StorageTypeETCD3:
		return newETCD3Storage(c, newFunc, newListFunc, resourcePrefix)
	default:
		return nil, nil, fmt.Errorf("unknown storage type: %s", c.Type)
	}
}

// 创建客户端
func CreateEtcdClient(c ConfigForResource) (*clientv3.Client, error) {
	switch c.Type {
	case StorageTypeETCD2:
		return nil, fmt.Errorf("%s is no longer a supported storage backend", c.Type)
	case StorageTypeUnset, StorageTypeETCD3:
		return NewETCD3Client(c.Transport)
	default:
		return nil, fmt.Errorf("unknown storage type: %s", c.Type)
	}
}

// 创建检查
func CreateHealthCheck(c Config, stopCh <-chan struct{}) (func() error, error) {
	switch c.Type {
	case StorageTypeETCD2:
		return nil, fmt.Errorf("%s is no longer a supported storage backend", c.Type)
	case StorageTypeUnset, StorageTypeETCD3:
		return newETCD3HealthCheck(c, stopCh)
	default:
		return nil, fmt.Errorf("unknown storage type: %s", c.Type)
	}
}

func CreateReadyCheck(c Config, stopCh <-chan struct{}) (func() error, error) {
	switch c.Type {
	case StorageTypeETCD2:
		return nil, fmt.Errorf("%s is no longer a supported storage backend", c.Type)
	case StorageTypeUnset, StorageTypeETCD3:
		return newETCD3ReadyCheck(c, stopCh)
	default:
		return nil, fmt.Errorf("unknown storage type: %s", c.Type)
	}
}

// 创建探测
func CreateProber(c Config) (Prober, error) {
	switch c.Type {
	case StorageTypeETCD2:
		return nil, fmt.Errorf("%s is no longer a supported storage backend", c.Type)
	case StorageTypeUnset, StorageTypeETCD3:
		return newETCD3ProberMonitor(c)
	default:
		return nil, fmt.Errorf("unknown storage type: %s", c.Type)
	}
}

//创建监视
// func CreateMonitor(c Config) (metrics.Monitor, error) {
// 	switch c.Type {
// 	case storagebackend.StorageTypeETCD2:
// 		return nil, fmt.Errorf("%s is no longer a supported storage backend", c.Type)
// 	case storagebackend.StorageTypeUnset, storagebackend.StorageTypeETCD3:
// 		return newETCD3ProberMonitor(c)
// 	default:
// 		return nil, fmt.Errorf("unknown storage type: %s", c.Type)
// 	}
// }

// Prober is an interface that defines the Probe function for doing etcd readiness/liveness checks.
type Prober interface {
	Probe(ctx context.Context) error
	Close() error
}

// healthcheck.go提供了检查etcd健康状态的函数
package etcd3

import (
	"encoding/json"
	"fmt"
)

// etcdHealth encodes data returned from etcd /healthz handler.
type etcdHealth struct {
	// Note this has to be public so the json library can modify it.
	Health string `json:"health"`
}

// EtcdHealthCheck decodes data returned from etcd /healthz handler.
// Deprecated: Validate health by passing storagebackend.Config directly to storagefactory.CreateProber.
func EtcdHealthCheck(data []byte) error {
	obj := etcdHealth{}
	if err := json.Unmarshal(data, &obj); err != nil {
		return err
	}
	if obj.Health != "true" {
		return fmt.Errorf("Unhealthy status: %s", obj.Health)
	}
	return nil
}

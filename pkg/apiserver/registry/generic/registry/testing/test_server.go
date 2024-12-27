package testing

import (
	"os"
	"path"
	"testing"

	clientv3 "go.etcd.io/etcd/client/v3"
	"hit.edu/framework/pkg/apiserver/registry/storage/etcd3/testserver"
	"hit.edu/framework/pkg/apiserver/registry/storage/storagebackend"
)

// EtcdTestServer encapsulates the datastructures needed to start local instance for testing
type EtcdTestServer struct {
	V3Client *clientv3.Client
}

func (e *EtcdTestServer) Terminate(t *testing.T) {
	// no-op, server termination moved to test cleanup
}

// NewUnsecuredEtcd3TestClientServer creates a new client and server for testing
func NewUnsecuredEtcd3TestClientServer(t *testing.T) (*EtcdTestServer, *storagebackend.Config) {
	server := &EtcdTestServer{}
	server.V3Client = testserver.RunEtcd(t, nil)
	config := &storagebackend.Config{
		Type:   "etcd3",
		Prefix: PathPrefix(),
		Transport: storagebackend.TransportConfig{
			ServerList: server.V3Client.Endpoints(),
		},
	}
	//fmt.Println(server.V3Client.Endpoints())
	return server, config
}

// PathPrefix returns the prefix set via the ETCD_PREFIX environment variable (if any).
func PathPrefix() string {
	pref := os.Getenv("ETCD_PREFIX")
	if pref == "" {
		pref = "registry"
	}
	return path.Join("/", pref)
}

// AddPrefix adds the ETCD_PREFIX to the provided key
func AddPrefix(in string) string {
	return path.Join(PathPrefix(), in)
}

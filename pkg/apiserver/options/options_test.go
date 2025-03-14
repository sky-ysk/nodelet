package options

import (
	"github.com/google/go-cmp/cmp"
	"github.com/spf13/pflag"
	noopoteltrace "go.opentelemetry.io/otel/trace/noop"
	"hit.edu/framework/pkg/apis/legacyscheme"
	"hit.edu/framework/pkg/apiserver/registry/storage/etcd3"
	"hit.edu/framework/pkg/apiserver/registry/storage/storagebackend"
	netutils "k8s.io/utils/net"
	"reflect"
	"testing"
)

func TestAddFlags(t *testing.T) {
	fs := pflag.NewFlagSet("addflagstest", pflag.PanicOnError)
	s := NewOptions()
	s.AddFlags(fs)

	args := []string{
		"--bind-address=192.168.10.10",
		"--bind-port=12345",
		"--etcd-servers=example1.com,example2.com",
		"--etcd-prefix=testPrefix/",
		"--etcd-keyfile=/var/run/kubernetes/etcd.key",
		"--etcd-certfile=/var/run/kubernetes/etcdce.crt",
		"--etcd-cafile=/var/run/kubernetes/etcdca.crt",
		"--delete-collection-workers=2",
		"--enable-garbage-collector=true",
		"--watch-cache=true",
	}
	fs.Parse(args)

	expected := &Options{
		ServingOptions: &ServingOptions{
			BindAddress: netutils.ParseIPSloppy("192.168.10.10"),
			BindPort:    12345,
		},
		EtcdOptions: &EtcdOptions{
			StorageConfig: storagebackend.Config{
				Type:   "",
				Prefix: "testPrefix/",
				Transport: storagebackend.TransportConfig{
					ServerList:     []string{"example1.com", "example2.com"},
					KeyFile:        "/var/run/kubernetes/etcd.key",
					CertFile:       "/var/run/kubernetes/etcdce.crt",
					TrustedCAFile:  "/var/run/kubernetes/etcdca.crt",
					TracerProvider: noopoteltrace.NewTracerProvider(),
				},
				Codec:                legacyscheme.Codecs.LegacyCodec(),
				CompactionInterval:   storagebackend.DefaultCompactInterval,
				DBMetricPollInterval: storagebackend.DefaultDBMetricPollInterval,
				HealthcheckTimeout:   storagebackend.DefaultHealthcheckTimeout,
				ReadycheckTimeout:    storagebackend.DefaultReadinessTimeout,
				LeaseManagerConfig:   etcd3.NewDefaultLeaseManagerConfig(),
			},
			DefaultStorageMediaType: "application/json",
			DeleteCollectionWorkers: 2,
			EnableGarbageCollection: true,
			DefaultWatchCacheSize:   100,
			EnableWatchCache:        true,
		},
	}

	if !reflect.DeepEqual(expected, s) {
		t.Errorf("Got different run options than expected.\nDifference detected on:\n%s", cmp.Diff(expected, s))
	}
}

package options

import (
	"hit.edu/framework/pkg/apiserver/registry/storage/storagebackend"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"strings"
	"testing"
	"time"
)

func TestEtcdOptionsValidate(t *testing.T) {
	testCases := []struct {
		name        string
		testOptions *EtcdOptions
		expectErr   string
	}{
		{
			name: "test when ServerList is not specified",
			testOptions: &EtcdOptions{
				StorageConfig: storagebackend.Config{
					Type:   "etcd3",
					Prefix: "/registry",
					Transport: storagebackend.TransportConfig{
						ServerList:    nil,
						KeyFile:       "/var/run/kubernetes/etcd.key",
						TrustedCAFile: "/var/run/kubernetes/etcdca.crt",
						CertFile:      "/var/run/kubernetes/etcdce.crt",
					},
					CompactionInterval:    storagebackend.DefaultCompactInterval,
					CountMetricPollPeriod: time.Minute,
				},
				DefaultStorageMediaType: "application/json",
				DeleteCollectionWorkers: 1,
				EnableGarbageCollection: true,
				EnableWatchCache:        true,
				DefaultWatchCacheSize:   100,
			},
			expectErr: "--etcd-servers must be specified",
		},
		{
			name: "test when storage-backend is invalid",
			testOptions: &EtcdOptions{
				StorageConfig: storagebackend.Config{
					Type:   "etcd4",
					Prefix: "/registry",
					Transport: storagebackend.TransportConfig{
						ServerList:    []string{"http://127.0.0.1"},
						KeyFile:       "/var/run/kubernetes/etcd.key",
						TrustedCAFile: "/var/run/kubernetes/etcdca.crt",
						CertFile:      "/var/run/kubernetes/etcdce.crt",
					},
					CompactionInterval:    storagebackend.DefaultCompactInterval,
					CountMetricPollPeriod: time.Minute,
				},
				DefaultStorageMediaType: "application/json",
				DeleteCollectionWorkers: 1,
				EnableGarbageCollection: true,
				EnableWatchCache:        true,
				DefaultWatchCacheSize:   100,
			},
			expectErr: "--storage-backend invalid, allowed values: etcd3. If not specified, it will default to 'etcd3'",
		},
		{
			name: "test when EtcdOptions is valid",
			testOptions: &EtcdOptions{
				StorageConfig: storagebackend.Config{
					Type:   "etcd3",
					Prefix: "/registry",
					Transport: storagebackend.TransportConfig{
						ServerList:    []string{"http://127.0.0.1"},
						KeyFile:       "/var/run/kubernetes/etcd.key",
						TrustedCAFile: "/var/run/kubernetes/etcdca.crt",
						CertFile:      "/var/run/kubernetes/etcdce.crt",
					},
					CompactionInterval:    storagebackend.DefaultCompactInterval,
					CountMetricPollPeriod: time.Minute,
				},
				DefaultStorageMediaType: "application/json",
				DeleteCollectionWorkers: 1,
				EnableGarbageCollection: true,
				EnableWatchCache:        true,
				DefaultWatchCacheSize:   100,
			},
		},
		{
			name: "empty storage-media-type",
			testOptions: &EtcdOptions{
				StorageConfig: storagebackend.Config{
					Transport: storagebackend.TransportConfig{
						ServerList: []string{"http://127.0.0.1"},
					},
				},
				DefaultStorageMediaType: "",
			},
		},
		{
			name: "recognized storage-media-type",
			testOptions: &EtcdOptions{
				StorageConfig: storagebackend.Config{
					Transport: storagebackend.TransportConfig{
						ServerList: []string{"http://127.0.0.1"},
					},
				},
				DefaultStorageMediaType: "application/json",
			},
		},
		{
			name: "unrecognized storage-media-type",
			testOptions: &EtcdOptions{
				StorageConfig: storagebackend.Config{
					Transport: storagebackend.TransportConfig{
						ServerList: []string{"http://127.0.0.1"},
					},
				},
				DefaultStorageMediaType: "foo/bar",
			},
			expectErr: `--storage-media-type "foo/bar" invalid, allowed values: application/json`,
		},
	}

	for _, testcase := range testCases {
		t.Run(testcase.name, func(t *testing.T) {
			errs := testcase.testOptions.Validate()
			if len(testcase.expectErr) != 0 && !strings.Contains(utilerrors.NewAggregate(errs).Error(), testcase.expectErr) {
				t.Errorf("got err: %v, expected err: %s", errs, testcase.expectErr)
			}
			if len(testcase.expectErr) == 0 && len(errs) != 0 {
				t.Errorf("got err: %s, expected err nil", errs)
			}
		})
	}
}

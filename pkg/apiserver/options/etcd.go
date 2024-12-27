package options

import (
	"fmt"
	"github.com/spf13/pflag"
	"hit.edu/framework/pkg/apimachinery/runtime"
	"hit.edu/framework/pkg/apimachinery/runtime/schema"
	"hit.edu/framework/pkg/apiserver/registry/generic"
	"hit.edu/framework/pkg/apiserver/registry/storage/storagebackend"
	"k8s.io/apimachinery/pkg/util/sets"
	"strings"
)

// ETCD连接相关参数

type EtcdOptions struct {
	// 存储后端配置
	StorageConfig storagebackend.Config
	// 存储数据的默认媒体类型
	DefaultStorageMediaType string
	// 删除集合时的并发工作线程数量
	DeleteCollectionWorkers int
	// 是否启用垃圾回收
	EnableGarbageCollection bool
	// 是否启用 watch 缓存
	EnableWatchCache bool
	// 配置默认的 watch 缓存大小
	DefaultWatchCacheSize int
	// 为特定资源配置 watch 缓存大小
	WatchCacheSizes []string
}

func NewEtcdOptions(backendConfig *storagebackend.Config) *EtcdOptions {
	options := &EtcdOptions{
		StorageConfig:           *backendConfig,
		DefaultStorageMediaType: "application/json",
		DeleteCollectionWorkers: 3,
		EnableGarbageCollection: true,
		EnableWatchCache:        true,
		DefaultWatchCacheSize:   100,
	}
	return options
}

var storageTypes = sets.NewString(
	storagebackend.StorageTypeETCD3,
)

var storageMediaTypes = sets.New(
	runtime.ContentTypeJSON,
)

func (s *EtcdOptions) Validate() []error {
	if s == nil {
		return nil
	}
	
	allErrors := []error{}
	if len(s.StorageConfig.Transport.ServerList) == 0 {
		allErrors = append(allErrors, fmt.Errorf("--etcd-servers must be specified"))
	}
	
	if s.StorageConfig.Type != storagebackend.StorageTypeUnset && !storageTypes.Has(s.StorageConfig.Type) {
		allErrors = append(allErrors, fmt.Errorf("--storage-backend invalid, allowed values: %s. If not specified, it will default to 'etcd3'", strings.Join(storageTypes.List(), ", ")))
	}
	
	if s.DefaultStorageMediaType != "" && !storageMediaTypes.Has(s.DefaultStorageMediaType) {
		allErrors = append(allErrors, fmt.Errorf("--storage-media-type %q invalid, allowed values: %s", s.DefaultStorageMediaType, strings.Join(sets.List(storageMediaTypes), ", ")))
	}
	
	return allErrors
}

// AddFlags adds flags related to etcd storage for a specific APIServer to the specified FlagSet
func (s *EtcdOptions) AddFlags(fs *pflag.FlagSet) {
	if s == nil {
		return
	}
	fs.StringSliceVar(&s.StorageConfig.Transport.ServerList, "etcd-servers", s.StorageConfig.Transport.ServerList,
		"要连接的 etcd 服务器列表 (scheme://ip:port), 逗号分隔")
	
	fs.StringVar(&s.StorageConfig.Prefix, "etcd-prefix", s.StorageConfig.Prefix,
		"在 etcd 中所有资源路径前附加的前缀")
	
	fs.StringVar(&s.StorageConfig.Transport.KeyFile, "etcd-keyfile", s.StorageConfig.Transport.KeyFile,
		"用于保护 etcd 通信的 SSL 密钥文件。")
	
	fs.StringVar(&s.StorageConfig.Transport.CertFile, "etcd-certfile", s.StorageConfig.Transport.CertFile,
		"用于保护 etcd 通信的 SSL 证书文件。")
	
	fs.StringVar(&s.StorageConfig.Transport.TrustedCAFile, "etcd-cafile", s.StorageConfig.Transport.TrustedCAFile,
		"用于保护 etcd 通信的 SSL 证书颁发机构文件。")
	
	fs.StringVar(&s.DefaultStorageMediaType, "storage-media-type", s.DefaultStorageMediaType, ""+
		"用于在存储中存储对象的媒体类型 "+
		"支持的媒体类型: [application/json]")
	
	fs.IntVar(&s.DeleteCollectionWorkers, "delete-collection-workers", s.DeleteCollectionWorkers,
		"为 DeleteCollection 调用生成的工作线程数。用于加速清理资源集合。")
	
	fs.BoolVar(&s.EnableGarbageCollection, "enable-garbage-collector", s.EnableGarbageCollection, ""+
		"启用泛型垃圾回收器")
	
	fs.BoolVar(&s.EnableWatchCache, "watch-cache", s.EnableWatchCache,
		"启用 watch 缓存")
	
	//fs.IntVar(&s.DefaultWatchCacheSize, "default-watch-cache-size", s.DefaultWatchCacheSize,
	//	"Default watch cache size. If zero, watch cache will be disabled for resources that do not have a default watch size set.")
	
	//fs.StringSliceVar(&s.WatchCacheSizes, "watch-cache-sizes", s.WatchCacheSizes, ""+
	//	"Watch cache size settings for some resources (pods, nodes, etc.), comma separated. "+
	//	"The individual setting format: resource[.group]#size, where resource is lowercase plural (no version), "+
	//	"group is omitted for resources of apiVersion v1 (the legacy core API) and included for others, "+
	//	"and size is a number. This option is only meaningful for resources built into the apiserver, "+
	//	"not ones defined by CRDs or aggregated from external servers, and is only consulted if the "+
	//	"watch-cache is enabled. The only meaningful size setting to supply here is zero, which means to "+
	//	"disable watch caching for the associated resource; all non-zero values are equivalent and mean "+
	//	"to not disable watch caching for that resource")
	
	//fs.DurationVar(&s.StorageConfig.CompactionInterval, "etcd-compaction-interval", s.StorageConfig.CompactionInterval,
	//	"The interval of compaction requests. If 0, the compaction request from apiserver is disabled.")
	//
	//fs.DurationVar(&s.StorageConfig.CountMetricPollPeriod, "etcd-count-metric-poll-period", s.StorageConfig.CountMetricPollPeriod, ""+
	//	"Frequency of polling etcd for number of resources per type. 0 disables the metric collection.")
	//
	//fs.DurationVar(&s.StorageConfig.DBMetricPollInterval, "etcd-db-metric-poll-interval", s.StorageConfig.DBMetricPollInterval,
	//	"The interval of requests to poll etcd and update metric. 0 disables the metric collection")
	//
	//fs.DurationVar(&s.StorageConfig.HealthcheckTimeout, "etcd-healthcheck-timeout", s.StorageConfig.HealthcheckTimeout,
	//	"The timeout to use when checking etcd health.")
	//
	//fs.DurationVar(&s.StorageConfig.ReadycheckTimeout, "etcd-readycheck-timeout", s.StorageConfig.ReadycheckTimeout,
	//	"The timeout to use when checking etcd readiness")
	//
	//fs.Int64Var(&s.StorageConfig.LeaseManagerConfig.ReuseDurationSeconds, "lease-reuse-duration-seconds", s.StorageConfig.LeaseManagerConfig.ReuseDurationSeconds,
	//	"The time in seconds that each lease is reused. A lower value could avoid large number of objects reusing the same lease. Notice that a too small value may cause performance problems at storage layer.")
	
}

type SimpleStorageFactory struct {
	Options       *EtcdOptions
	StorageConfig storagebackend.Config
}

func (s *SimpleStorageFactory) NewConfig(resource schema.GroupResource, example runtime.Object) (*storagebackend.ConfigForResource, error) {
	return s.StorageConfig.ForResource(resource), nil
}

func (s *SimpleStorageFactory) ResourcePrefix(resource schema.GroupResource) string {
	return resource.Group + "/" + resource.Resource
	//return resource.Resource
}
func (s *SimpleStorageFactory) GetRESTOptions(resource schema.GroupResource, example runtime.Object) (generic.RESTOptions, error) {
	storageConfig, err := s.NewConfig(resource, example)
	if err != nil {
		return generic.RESTOptions{}, fmt.Errorf("unable to find storage destination for %v, due to %v", resource, err.Error())
	}
	
	ret := generic.RESTOptions{
		StorageConfig:           storageConfig,
		Decorator:               generic.UndecoratedStorage,
		DeleteCollectionWorkers: s.Options.DeleteCollectionWorkers,
		EnableGarbageCollection: s.Options.EnableGarbageCollection,
		ResourcePrefix:          s.ResourcePrefix(resource),
		CountMetricPollPeriod:   s.Options.StorageConfig.CountMetricPollPeriod,
	}
	return ret, nil
}

//func (s *EtcdOptions) ApplyStorageFactoryToConfig(c *apiserver.Config) {
//	storageFactory := &SimpleStorageFactory{Options: s, StorageConfig: s.StorageConfig}
//	c.RESTOptionsGetter = storageFactory
//}

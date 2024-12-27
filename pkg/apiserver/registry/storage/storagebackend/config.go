package storagebackend

import (
	"time"
	
	//clientv3 "go.etcd.io/etcd/client/v3"
	oteltrace "go.opentelemetry.io/otel/trace"
	noopoteltrace "go.opentelemetry.io/otel/trace/noop"
	
	"hit.edu/framework/pkg/apimachinery/runtime"
	"hit.edu/framework/pkg/apimachinery/runtime/schema"
	"hit.edu/framework/pkg/apiserver/registry/storage/etcd3"
	"hit.edu/framework/pkg/apiserver/registry/storage/value"
)

const (
	//存储后端的类型
	StorageTypeUnset = ""
	StorageTypeETCD2 = "etcd2"
	StorageTypeETCD3 = "etcd3"
	//Config缺省值
	DefaultCompactInterval      = 5 * time.Minute
	DefaultDBMetricPollInterval = 30 * time.Second
	DefaultHealthcheckTimeout   = 2 * time.Second
	DefaultReadinessTimeout     = 2 * time.Second
)

// 与存储服务器的连接信息
type TransportConfig struct {
	// 存储服务器的地址列表
	ServerList []string
	// TLS认证文件路径
	KeyFile       string
	CertFile      string
	TrustedCAFile string
	
	// The TracerProvider can add tracing the connection
	TracerProvider oteltrace.TracerProvider
}

// 存储后端的配置
type Config struct {
	// 存储后端的类型，默认值为etcd3
	Type string
	// 使用的前缀
	Prefix string
	// 连接信息
	Transport TransportConfig
	
	Codec runtime.Codec
	// EncodeVersioner is the same groupVersioner used to build the
	// storage encoder. Given a list of kinds the input object might belong
	// to, the EncodeVersioner outputs the gvk the object will be
	// converted to before persisted in etcd.
	EncodeVersioner runtime.GroupVersioner
	// Transformer allows the value to be transformed prior to persisting into etcd.
	Transformer value.Transformer
	
	// 请求 apiserver 压缩的间隔时间,0 表示不进行压缩
	CompactionInterval time.Duration
	// 计数指标更新间隔时间
	CountMetricPollPeriod time.Duration
	// 存储后端指标更新频率
	DBMetricPollInterval time.Duration
	// 健康检查超时时间
	HealthcheckTimeout time.Duration
	// 就绪检查超时时间
	ReadycheckTimeout time.Duration
	// TODO:租约管理器的配置
	LeaseManagerConfig etcd3.LeaseManagerConfig
	
	// TODO:跟踪每个资源在存储中的对象总数
	//StorageObjectCountTracker flowcontrolrequest.StorageObjectCountTracker
}

type ConfigForResource struct {
	Config
	GroupResource schema.GroupResource
}

// ForResource specializes to the given resource
func (config *Config) ForResource(resource schema.GroupResource) *ConfigForResource {
	return &ConfigForResource{
		Config:        *config,
		GroupResource: resource,
	}
}

func NewDefaultConfig(prefix string, codec runtime.Codec) *Config {
	return &Config{
		Prefix:               prefix,
		Codec:                codec,
		CompactionInterval:   DefaultCompactInterval,
		DBMetricPollInterval: DefaultDBMetricPollInterval,
		HealthcheckTimeout:   DefaultHealthcheckTimeout,
		ReadycheckTimeout:    DefaultReadinessTimeout,
		LeaseManagerConfig:   etcd3.NewDefaultLeaseManagerConfig(),
		Transport:            TransportConfig{TracerProvider: noopoteltrace.NewTracerProvider()},
	}
}

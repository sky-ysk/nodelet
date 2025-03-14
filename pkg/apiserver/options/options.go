package options

import (
	"github.com/spf13/pflag"
	"go.uber.org/zap"
	"hit.edu/framework/pkg/apis/legacyscheme"
	"hit.edu/framework/pkg/apiserver/registry/storage/etcd3/testserver"
	"hit.edu/framework/pkg/apiserver/registry/storage/storagebackend"
	"hit.edu/framework/pkg/component-base/logs"
	"testing"
)

//Config与Option区别
// Options对应默认参数，运行前配置
// Config对应初始化后的配置和运行中的配置

type Options struct {
	ServingOptions *ServingOptions
	EtcdOptions    *EtcdOptions
}

// 配置所有的选项
func NewOptions() *Options {
	s := Options{
		ServingOptions: NewServingOptions(),
		EtcdOptions:    NewEtcdOptions(storagebackend.NewDefaultConfig("registry/", legacyscheme.Codecs.LegacyCodec())),
	}
	return &s
}

func (s *Options) AddFlags(fs *pflag.FlagSet) {
	s.ServingOptions.AddFlags(fs)
	s.EtcdOptions.AddFlags(fs)
}

func (s *Options) Validate() []error {
	var errs []error

	//TODO:启动本地etcd服务器，仅用于测试
	if len(s.EtcdOptions.StorageConfig.Transport.ServerList) == 0 {
		var t testing.T
		etcdClient := testserver.RunEtcd(&t, nil)
		s.EtcdOptions.StorageConfig.Transport.ServerList = etcdClient.Endpoints()
		logs.Info("start local embed etcd", zap.String(
			"endpoints",
			etcdClient.Endpoints()[0],
		))
	}

	errs = append(errs, s.EtcdOptions.Validate()...)

	return errs
}

package options

import (
	"github.com/spf13/pflag"
	netutils "k8s.io/utils/net"
	"net"
)

// TODO: 将K8s模块替换成我们自己的模块
// 只有一个API Server
// 负责获取资源，与Etcd模块通信
// 不支持用户自定义扩展

// 设置HTTP Server相关参数
type ServingOptions struct {
	// 连接地址
	BindAddress net.IP
	// HTTP连接端口
	BindPort int
	// HTTP连接网络,tcp4 or tcp6
	BindNetwork string
	// TODO: 本地Hostname
	BindHost string
	// 对外广播的地址，用于服务发现
	ExternalAddress net.IP
	// HTTP Lisntener
	Listener net.Listener

	//TODO: HTTP2相关配置
	//TODO: 网络环境优化
}

func NewServingOptions() *ServingOptions {
	return &ServingOptions{
		BindAddress: netutils.ParseIPSloppy("0.0.0.0"),
		BindPort:    10000,
	}
}

// TODO: 参数验证
// TODO: 参数获取

func (s *ServingOptions) AddFlags(fs *pflag.FlagSet) {
	if s == nil {
		return
	}

	// 绑定IP地址
	fs.IPVar(&s.BindAddress, "bind-address", s.BindAddress, "API Server 监听的IP地址")

	// 绑定端口
	fs.IntVar(&s.BindPort, "bind-port", s.BindPort, "API Server 监听的端口号")
}

package device

// Config 表示DeviceExporter配置
type Config struct {
	// 需要配置的选项包括
	//  需要监控的内容，需要启用哪些Collector  (例如：[]string{"state", "status"})
	EnabledCollectors []string
	//  资源访问方式 （例如：资源访问协议或 API 的 URL）
	ResourceAccessMethod string
	//  ......
}

// NewConfig 表示新建一个Config结构
func NewConfig(enabledCollectors []string, resourceAccessMethod string) *Config {
	// 如果资源访问方式为空，设置默认值
	if resourceAccessMethod == "" {
		resourceAccessMethod = "local" // 默认资源访问方式
	}
	return &Config{
		EnabledCollectors:    enabledCollectors,
		ResourceAccessMethod: resourceAccessMethod,
	}
}

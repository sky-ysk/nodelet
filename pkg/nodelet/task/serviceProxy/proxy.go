package serviceProxy

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/stretchr/testify/assert/yaml"
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/client-go/util/manager"
	"hit.edu/framework/pkg/component-base/logs"
)

// ServiceProxy,用于调用ServiceProxy的接口。本身包含服务迁移组件的地址，
type ServiceProxy struct {
	clientsManager *manager.Manager
	ProxyAddr      string
	ProxyIp        string
	ProxyPort      string
}

// ServiceInfo 存储服务信息
type ServiceProxyConfig struct {
	ServiceProxyAddr string `yaml:"ServiceProxyAddr"`
}

func NewServiceProxy(clientsManager *manager.Manager) *ServiceProxy {
	addr := GetServiceProxyAddr()
	ip, port, err := parseAddr(addr)
	if err != nil {
		logs.Errorf("Failed to parse ServiceProxy address: %v", err)
	}
	return &ServiceProxy{
		clientsManager: clientsManager,
		ProxyAddr:      addr,
		ProxyIp:        ip,
		ProxyPort:      port,
	}
}

func parseAddr(addr string) (string, string, error) {
	parts := strings.Split(addr, ":")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid address format: expected 'ip:port', got %q", addr)
	}
	ip := parts[0]
	port := parts[1]
	if ip == "" || port == "" {
		return "", "", fmt.Errorf("invalid address format: empty ip or port in %q", addr)
	}
	return ip, port, nil
}

// 获取ServiceProxy的地址
func GetServiceProxyAddr() string {
	config := newConfig()
	return config.ServiceProxyAddr
}

func newConfig() *ServiceProxyConfig {
	var config *ServiceProxyConfig
	var err error
	// logs.Info("framework-conf ", configPath)
	//没有指定配置文件位置，则去默认位置加载

	logs.Info("ConfigPath is empty, using default")
	fileName := "frameworkConf.yaml"
	// 获取当前文件绝对路径
	_, currentFilePath, _, _ := runtime.Caller(0)
	// 计算项目根目录路径
	projectRoot := filepath.Join(filepath.Dir(currentFilePath), "..", "..", "..", "..")
	// 构建配置文件的绝对路径
	configPath := filepath.Join(projectRoot, fileName)

	logs.Infof("configPath:%v", configPath)
	// 验证路径有效性
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		logs.Errorf("配置文件不存在于：%s", configPath)
	}
	config, err = loadConfig(configPath)
	if err != nil {
		logs.Errorf("frameworkConf.yaml load failed: %e", err)
		config = &ServiceProxyConfig{} // 使用空配置
	}

	// 获取必要配置项（环境变量优先于配置文件）
	serviceProxy := GetServiceProxy(config)
	logs.Infof("serviceProxyConfig:%v===========", config)
	logs.Infof("serviceProxy:%v==============", serviceProxy)

	return &ServiceProxyConfig{
		ServiceProxyAddr: serviceProxy,
	}
}

// 配置加载函数
func loadConfig(path string) (*ServiceProxyConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %v", err)
	}

	var config ServiceProxyConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("解析YAML失败: %v", err)
	}

	return &config, nil
}

func GetServiceProxy(config *ServiceProxyConfig) string {
	if config.ServiceProxyAddr != "" {
		return config.ServiceProxyAddr
	}
	return "http://localhost:8921"
}

// 在这里写解析IP和PORT、服务名的函数
func (sp *ServiceProxy) ExtractFromArgs(args []string) (string, string, string) {
	var ip, port, serviceName string
	for i, arg := range args {
		if arg == "--ip" && i+1 < len(args) {
			ip = args[i+1]
		}
		if arg == "--port" && i+1 < len(args) {
			port = args[i+1]
		}
		if arg == "--serviceName" && i+1 < len(args) {
			serviceName = args[i+1]
		}
	}
	return ip, port, serviceName
}

// 在这里进行args的修改，ip和port进行替换
func (sp *ServiceProxy) ModifyArgs(args []string) {
	// modifiedArgs := make([]string, len(args))
	// copy(modifiedArgs, args)

	for i, arg := range args {
		if arg == "--ip" && i+1 < len(args) {
			args[i+1] = sp.ProxyIp
		}
		if arg == "--port" && i+1 < len(args) {
			args[i+1] = sp.ProxyPort
		}
	}
}

// 迁移服务
func (sp *ServiceProxy) MigrateService(group *apis.Group, groupNamespace string, runtime *apis.Runtime, args []string) error {

	logs.Infof("runtime.Spec.IsHttpService:%v, runtime.Spec.IsHttpClient:%v", runtime.Spec.IsHttpService, runtime.Spec.IsHttpClient)

	// 说明是迁移后的copy服务端，需要调用服务迁移的接口进行服务迁移
	_, port, serviceName := sp.ExtractFromArgs(args)
	// 获取当前group所在的node的IP信息
	nodeName := *group.Status.Node
	node, err := sp.clientsManager.GetNode(nodeName, groupNamespace)
	if err != nil {
		logs.Infof("fail to get node:%s", nodeName)
	}
	nodeIp := node.Spec.HostIp
	logs.Infof("New nodeIp:%s", nodeIp)

	// BUG:需要先保证新进程已经启动成功，才能进行迁移，否则会在迁移期间报错服务不可用或者无响应
	// TODO:杀死之前的进程，否则端口一直被占用
	// 在这里处理：先检查新runtime的running是否，再继续后续的操作，最后再想办法用kill杀掉原进程
	cnt := 0
	for {
		cnt++
		// 100s如果还没拉起，则认为迁移出错；
		//后续可能优化一下，尝试重新调用迁移接口
		if cnt >= 1000 {
			logs.Infof("100s waiting already, migrated server still not running, migrated Error!")
			return fmt.Errorf("migrated Error for waiting too long time!")
		}
		newRuntime, err := sp.clientsManager.GetRuntime(runtime.Name, runtime.Namespace)
		if err != nil {
			logs.Errorf("GetRuntime Err")
		}
		if newRuntime.Status.Phase == apis.Running {
			logs.Infof("migrated server runtime already running, start migrate interface!")
			break
		} else {
			logs.Warnf("migrated server runtime still pending, wait for running!")
			time.Sleep(100 * time.Millisecond)
		}
	}

	// 向组件发送迁移服务的请求
	url := fmt.Sprintf("http://%s/migrate?name=%s&host=%s&port=%s", sp.ProxyAddr, serviceName, nodeIp, port)
	resp, err := http.Post(url, "application/x-www-form-urlencoded", nil)
	if err != nil {
		return fmt.Errorf("failed to migrate service: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("migration failed with status %d: %s", resp.StatusCode, string(body))
	}

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("Migration response: %s\n", string(body))
	logs.Infof("MigrateService success: %s %s", serviceName, nodeIp)
	return nil
}

// 注册服务
func (sp *ServiceProxy) RegisterService(group *apis.Group, groupNamespace string, runtime *apis.Runtime, args []string) error {

	_, port, serviceName := sp.ExtractFromArgs(args)
	logs.Infof("server's port:%s, name:%s", port, serviceName)
	// 获取当前group所在的node的IP信息
	nodeName := *group.Status.Node
	node, err := sp.clientsManager.GetNode(nodeName, groupNamespace)
	if err != nil {
		logs.Infof("fail to get node:%s", nodeName)
	}
	nodeIp := node.Spec.HostIp
	logs.Infof("nodeIp:%s", nodeIp)

	url := fmt.Sprintf("http://%s/register?name=%s&host=%s&port=%s", sp.ProxyAddr, serviceName, nodeIp, port)
	resp, err := http.Post(url, "application/x-www-form-urlencoded", nil)
	if err != nil {
		return fmt.Errorf("failed to register service: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("registration failed with status %d: %s", resp.StatusCode, string(body))
	}

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("Service registration response: %s\n", string(body))
	return nil
}

package serviceProxy

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	"github.com/stretchr/testify/assert/yaml"
	"hit.edu/framework/pkg/component-base/logs"
)

// ServiceInfo 存储服务信息
type ServiceProxyConfig struct {
	address string
}

// 获取ServiceProxy的地址
func GetServiceProxyAddr() string {
	config := newConfig()
	return config.address
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
	projectRoot := filepath.Join(filepath.Dir(currentFilePath), "..", "..")
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
	serviceProxy := config.address
	logs.Infof("serviceProxy:%v==============", serviceProxy)

	return &ServiceProxyConfig{
		address: serviceProxy,
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

// 在这里写解析IP和PORT、服务名的函数
func ExtractFromArgs(args []string) (string, string, string) {
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

// 迁移服务
func MigrateService(proxyAddr, serviceName, newHost, newPort string) error {
	url := fmt.Sprintf("http://%s/migrate?name=%s&host=%s&port=%s", proxyAddr, serviceName, newHost, newPort)
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
	return nil
}

// 注册服务
func RegisterService(proxyAddr, serviceName, host, port string) error {
	url := fmt.Sprintf("http://%s/register?name=%s&host=%s&port=%s", proxyAddr, serviceName, host, port)
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

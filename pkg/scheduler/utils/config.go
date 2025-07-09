package utils

import (
	"errors"
	"fmt"
	"gopkg.in/yaml.v3"
	"hit.edu/framework/pkg/component-base/logs"
	"hit.edu/framework/pkg/nodelet"
	"os"
	"path/filepath"
	run "runtime"
	"sync"
)

func GetNamespace() string {
	logs.Info("ConfigPath is empty, using default")
	fileName := "frameworkConf.yaml"
	// 获取当前文件绝对路径
	_, currentFilePath, _, _ := run.Caller(0)
	// 计算项目根目录路径
	projectRoot := filepath.Join(filepath.Dir(currentFilePath), "..", "..", "..")
	// 构建配置文件的绝对路径
	configPath := filepath.Join(projectRoot, fileName)
	logs.Infof("configPath:%v", configPath)
	// 验证路径有效性
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		logs.Errorf("配置文件不存在于：%s", configPath)
		panic(err)
	}
	cf, err := nodelet.LoadConfig(configPath)
	if err != nil {
		panic(err)
	}
	fmt.Println("the config is ", cf.Namespace)
	return cf.Namespace
}

var (
	apiServerHost string
	once          sync.Once
	initErr       error
)

// APIConfig 定义YAML配置结构
type APIConfig struct {
	APIServerHost string `yaml:"api_server_host"`
}

// GetAPIServerHost 获取API服务器地址（线程安全）
func GetAPIServerHost() string {
	once.Do(func() {
		initErr = initializeConfig()
	})

	if initErr != nil {
		panic(initErr)
	}

	return apiServerHost
}

// initializeConfig 初始化配置（私有方法）
func initializeConfig() error {
	// 获取当前可执行文件所在目录
	exeDir, err := os.Executable()
	if err != nil {
		return fmt.Errorf("获取可执行文件路径失败: %w", err)
	}
	exeDir = filepath.Dir(exeDir)

	// 构建配置文件的绝对路径（相对于项目根目录上三层）
	configPath := filepath.Join(exeDir, "../../../config.yaml")

	// 读取YAML文件
	data, err := os.ReadFile(configPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// 文件不存在时使用默认值
			apiServerHost = "http://localhost:10000"
			return nil
		}
		return fmt.Errorf("读取配置文件失败: %w", err)
	}

	// 解析YAML
	var cfg APIConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("解析配置文件失败: %w", err)
	}

	// 设置API服务器地址
	if cfg.APIServerHost != "" {
		apiServerHost = cfg.APIServerHost
	} else {
		apiServerHost = "http://localhost:10000" // 默认值
	}

	return nil
}

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

type Utils struct {
	configPaht string
}

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
	APIServerHost string `yaml:"ApiServerAddr"`
}

func Initialize(configPath string) {
	once.Do(func() {
		initErr = initializeConfig(configPath)
	})
	if initErr != nil {
		panic(initErr)
	}
}

// GetAPIServerHost 获取API服务器地址（线程安全）
func GetAPIServerHost() string {
	return apiServerHost
}

// initializeConfig 初始化配置（私有方法）
func initializeConfig(configPath string) error {
	if configPath == "" {
		logs.Info("ConfigPath is empty, using default")
		fileName := "frameworkConf.yaml"
		// 获取当前文件绝对路径
		_, currentFilePath, _, _ := run.Caller(0)
		// 计算项目根目录路径
		projectRoot := filepath.Join(filepath.Dir(currentFilePath), "..", "..", "..")
		// 构建配置文件的绝对路径
		configPath = filepath.Join(projectRoot, fileName)
	}
	logs.Infof("configPath:%v", configPath)

	// 构建配置文件的绝对路径（相对于项目根目录上三层）
	// 读取YAML文件
	data, err := os.ReadFile(configPath)
	if err != nil {
		panic(err)
		if errors.Is(err, os.ErrNotExist) {
			// 文件不存在时使用默认值
			apiServerHost = "http://localhost:8120"
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
		apiServerHost = "http://localhost:8120" // 默认值
	}

	return nil
}

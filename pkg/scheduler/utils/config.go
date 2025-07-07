package utils

import (
	"fmt"
	"hit.edu/framework/pkg/component-base/logs"
	"hit.edu/framework/pkg/nodelet"
	"os"
	"path/filepath"
	run "runtime"
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

func GetAPIServerHost() string {
	if host := os.Getenv("API_SERVER_HOST"); host != "" {
		return host
	}
	return "http://localhost:10000"
}

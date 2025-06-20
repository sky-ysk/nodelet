package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// 配置参数
const (
	podName       = "grpc-server-pod"
	namespace     = "switch"
	outputFile    = "switch3.txt"
	retryInterval = 3 // Pod 不存在时的重试间隔（秒）
	logLineMatch  = "Starting server"
)

func main() {
	for {
		if !podExists(podName, namespace) {
			fmt.Printf("Pod %s 不存在，等待 %d 秒后重试...\n", podName, retryInterval)
			time.Sleep(time.Duration(retryInterval) * time.Second)
			continue
		} else {
			fmt.Printf("Pod存在!")
		}

		// 捕获日志流
		cmd := exec.Command("kubectl", "logs", "-f", podName, "-n", namespace, "--since=0s")
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			fmt.Printf("创建管道失败: %v\n", err)
			time.Sleep(time.Duration(retryInterval) * time.Second)
			continue
		}

		if err := cmd.Start(); err != nil {
			fmt.Printf("启动日志捕获失败: %v\n", err)
			time.Sleep(time.Duration(retryInterval) * time.Second)
			continue
		}

		// 读取日志流
		reader := bufio.NewReader(stdout)
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				// 日志流中断，重新开始循环
				cmd.Process.Kill()
				break
			}

			// 匹配目标日志行
			if strings.Contains(line, logLineMatch) {
				// 提取时间戳（假设时间戳是日志行的前两个字段，格式为 YYYY/MM/DD HH:MM:SS.MICROSECONDS）
				timestamp := extractTimestamp(line)
				if timestamp != "" {
					fmt.Printf("已记录事件：%s\n", timestamp)
					return 
				}
			}

			// 等待重试间隔
			time.Sleep(time.Duration(retryInterval) * time.Second)
		}

		// 等待重试间隔
		time.Sleep(time.Duration(retryInterval) * time.Second)
	}
}

// 检查 Pod 是否存在
func podExists(podName, namespace string) bool {
	cmd := exec.Command("kubectl", "get", "pod", podName, "-n", namespace)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	return err == nil
}

// 提取时间戳（时间戳是日志行的前两个字段）
func extractTimestamp(line string) string {
	parts := strings.Fields(line)
	if len(parts) >= 2 {
		return parts[0] + " " + parts[1]
	}
	return ""
}

package main

import (
	"fmt"
	"os"
	"time"

	"hit.edu/framework/pkg/nodelet/task/monitor"
)

func main() {
	// 示例：监控当前进程
	pid := os.Getpid()
	monitor := monitor.NewProcessMonitor(monitor.ResourceIndex{PID: string(pid)})
	go busy()

	for {
		stats, err := monitor.Collect()
		if err != nil {
			fmt.Printf("监控错误: %v\n", err)
			break
		}

		fmt.Printf("PID: %d  CPU: %.2f%%  Memory: %.2f MB\n",
			stats.PID,
			stats.CPUPercent,
			stats.MemoryUsageMB)

		time.Sleep(5 * time.Second)
	}
}

func busy() {
	var i int
	for {
		i += 100
		time.Sleep(100 * time.Microsecond)
	}
}

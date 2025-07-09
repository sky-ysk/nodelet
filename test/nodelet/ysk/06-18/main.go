package main

 import (
    "fmt"
    "log"
    "time"

    "github.com/shirou/gopsutil/v3/process"
)

func main() {
    pid := 1975 // 替换为你要监控的进程 ID
    p, err := process.NewProcess(int32(pid))
    if err != nil {
        log.Fatalf("Failed to get process: %v", err)
    }

    // 获取 CPU 使用率    
	cpuPercent, err := p.Percent(time.Second)
    if err != nil {
        log.Fatalf("Failed to get CPU usage: %v", err)
    }
    fmt.Printf("CPU Usage: %.2f%%\n", cpuPercent)

    // 获取内存信息
    memInfo, err := p.MemoryInfo()
    if err != nil {
        log.Fatalf("Failed to get memory info: %v", err)
    }
    memMB := float64(memInfo.RSS) / (1024 * 1024)
    fmt.Printf("Memory Usage: %.2f MB\n", memMB)
}
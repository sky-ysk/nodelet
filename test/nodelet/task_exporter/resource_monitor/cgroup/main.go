package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/containerd/cgroups/v3"
	"github.com/containerd/cgroups/v3/cgroup2"

	"hit.edu/framework/pkg/component-base/logs"
)

var cgroupPath = "/sys/fs/cgroup/my_task_group"

func main() {
	logs.Init("test")
	// err := containedCgroups()
	// if err != nil {
	// 	logs.Error(err)
	// }

	testForCustomed()

}

func containedCgroups() error {
	isVersion_2 := cgroupsVersion_2()
	if !isVersion_2 {
		return errors.New("only supports cgroup version V2")
	}
	res := &cgroup2.Resources{}
	// // 定义资源限制
	// res := &specs.LinuxResources{
	// 	CPU: &specs.LinuxCPU{
	// 		Shares: cgroup2.Uint64Ptr(512),  // CPU 权重
	// 		Quota:  cgroup2.Int64Ptr(50000), // CPU 配额（单位：微秒）
	// 	},
	// 	Memory: &specs.LinuxMemory{
	// 		Limit: cgroup2.Int64Ptr(1024 * 1024 * 100), // 内存限制 100MB
	// 	},
	// }
	// dummy PID of -1 is used for creating a "general slice" to be used as a parent cgroup.
	// see https://github.com/containerd/cgroups/blob/1df78138f1e1e6ee593db155c6b369466f577651/v2/manager.go#L732-L735
	logs.Info("new cgroup")
	// cm, err := cgroup2.NewSystemd("/", "my.slice", -1, res)
	// cm, err := cgroup2.NewSystemd("/", "task-group-action-runtime.slice", -1, res)
	_, err := cgroup2.NewManager("/", "/my.slice/cgroup.slice", res)
	// defer cm.Delete()
	if err != nil {
		logs.Error(err)
		return err
	}
	logs.Info("cgroup2.NewSystemd successsed")
	// cm.AddProc()
	return nil
}

// func NewCgroupManager() (interface{}, error) {
//     if cgroup2.IsCgroup2UnifiedMode() {
//         return cgroup2.NewManager() // v2 逻辑
//     }
//     return cgroup1.New() // v1 逻辑
// }

func addProcessToCgroup(mgr *cgroup2.Manager, pid int) error {
	return mgr.AddProc(uint64(pid))
}

// func getCgroupStats(mgr *cgroup2.Manager) (*stats.Metrics, error) {
// 	stats, err := mgr.Stat()
// 	if err != nil {
// 		return nil, err
// 	}
// 	fmt.Printf("CPU Usage: %v ns\n", stats.CPU.Usage.Total)
// 	fmt.Printf("Memory Usage: %v bytes\n", stats.Memory.Usage.Usage)
// 	return stats, nil
// }

func deleteCgroup(mgr *cgroup2.Manager) error {
	return mgr.Delete()
}

func cgroupsVersion_2() bool {
	var cgroupV2 bool
	if cgroups.Mode() == cgroups.Unified {
		cgroupV2 = true
	} else {
		cgroupV2 = false
	}
	return cgroupV2
}

func testForCustomed() {
	cmd := LaunchTask()
	defer cmd.Process.Kill()
	mmint64, err := GetMemoryUsage(cgroupPath)
	if err != nil {
		logs.Error(err)
	}
	str := strconv.FormatUint(mmint64, 10)
	fmt.Println(str)

	err = cmd.Wait()
	if err != nil {
		logs.Error(err)
	}
	prompt()
}

// 创建 cgroup 并将进程 PID 加入其中
func AssignProcessToCgroup(pid int, cgroupPath string) error {
	// 创建 cgroup 目录
	if err := os.MkdirAll(cgroupPath, 0755); err != nil {
		return err
	}

	// 将 PID 写入 cgroup.procs
	pidStr := strconv.Itoa(pid)
	// logs.Info(pidStr)
	return os.WriteFile(filepath.Join(cgroupPath, "cgroup.procs"), []byte(pidStr), 0644)
}

// 在启动任务时调用
func LaunchTask() *exec.Cmd {
	cmd := exec.Command("python3", "/home/kcm/workspace/migration-demo-0116/yolo-runner.py")
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	if err := cmd.Start(); err != nil {
		panic(err)
	}

	// 将进程 PID 分配到 cgroup
	if err := AssignProcessToCgroup(cmd.Process.Pid, cgroupPath); err != nil {
		logs.Error(err)
		panic(err)
	}
	return cmd
}

func GetMemoryUsage(cgroupPath string) (uint64, error) {
	data, err := os.ReadFile(filepath.Join(cgroupPath, "memory.current"))
	if err != nil {
		logs.Error(err)
		return 0, err
	}
	logs.Info(data)
	return strconv.ParseUint(strings.TrimSpace(string(data)), 10, 64)
}

func prompt() {
	fmt.Printf("-> Press Return key to continue.")
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		break
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}
	logs.Info()
}

type CPUStats struct {
	LastUsage uint64 // 上一次的 CPU 使用时间（纳秒）
	LastTime  time.Time
}

func (s *CPUStats) getCPUUsage(cgroupPath string) (usagePercent float64) {
	// 读取当前CPU 使用时间
	currentUsage := readCgroupCPUUsage(cgroupPath) // ！注意单位是纳秒

	// 计算时间差和用量差
	currentTime := time.Now()
	timeDiff := currentTime.Sub(s.LastTime).Seconds()
	usageDiff := currentUsage - s.LastUsage

	// 计算使用率（CPU 核心数为 N）
	// 公式：(usage_diff / time_diff) / (核心数 * 1e9) * 100%
	cores := getCPUCores()
	usagePercent = float64(usageDiff) / (timeDiff * 1e9) / float64(cores) * 100

	// 更新
	s.LastUsage = currentUsage
	s.LastTime = currentTime
	return
}

// CPU 核心数
func getCPUCores() int {
	data, _ := os.ReadFile("/proc/cpuinfo")
	return strings.Count(string(data), "processor\t:")
}

// 读取 cgroup CPU 使用时间
func readCgroupCPUUsage(cgroupPath string) uint64 {
	data, _ := os.ReadFile(filepath.Join(cgroupPath, "cpu.stat"))
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "usage_usec") {
			fields := strings.Fields(line)
			val, _ := strconv.ParseUint(fields[1], 10, 64)
			return val * 1000 // 微秒转纳秒
		}
	}
	return 0
}

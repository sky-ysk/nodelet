package memory

import "github.com/shirou/gopsutil/mem"

type MemoryInfo struct {
	Total     uint64
	Available uint64
}
type MemoryUsage struct {
	Usage float64
}

type MemoryInfoProvider interface {
	GetMemoryInfo() (*MemoryInfo, error)
	GetMemoryUsage() (*MemoryUsage, error)
}

// Linux 平台的内存信息提供者
type LinuxMemInfoProvider struct{}

func NewLinuxMemInfoProvider() *LinuxMemInfoProvider {
	return &LinuxMemInfoProvider{}
}

func (l LinuxMemInfoProvider) GetMemoryInfo() (*MemoryInfo, error) {
	memStat, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
	}
	return &MemoryInfo{
		Total:     memStat.Total,
		Available: memStat.Available,
	}, nil
}

func (l LinuxMemInfoProvider) GetMemoryUsage() (*MemoryUsage, error) {
	memStat, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
	}
	return &MemoryUsage{
		Usage: memStat.UsedPercent,
	}, nil
}

// Windows 平台的内存信息提供者
type WindowsMemInfoProvider struct{}

func NewWindowsMemInfoProvider() *WindowsMemInfoProvider {
	return &WindowsMemInfoProvider{}
}

func (w *WindowsMemInfoProvider) GetMemoryInfo() (*MemoryInfo, error) {
	memStat, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
	}
	return &MemoryInfo{
		Total:     memStat.Total,
		Available: memStat.Available,
	}, nil
}

func (w *WindowsMemInfoProvider) GetMemoryUsage() (*MemoryUsage, error) {
	memStat, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
	}
	return &MemoryUsage{
		Usage: memStat.UsedPercent,
	}, nil
}

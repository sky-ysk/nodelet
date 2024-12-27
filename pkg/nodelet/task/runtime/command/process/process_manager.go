package process

import (
	"os/exec"
	"sync"
)

type ProcessManager struct {
	lock      sync.Mutex
	processes map[string]*exec.Cmd // 任务名称映射到 exec.Cmd 对象
	successProcess map[string]*exec.Cmd
}

func NewProcessManager() *ProcessManager {
	return &ProcessManager{
		processes: make(map[string]*exec.Cmd),
		successProcess: make(map[string]*exec.Cmd),
	}
}

// 管理进程的添加

func (pm *ProcessManager) AddProcess(name string, cmd *exec.Cmd) {
	pm.lock.Lock()
	defer pm.lock.Unlock()
	pm.processes[name] = cmd
}

// 管理进程的删除

func (pm *ProcessManager) RemoveProcess(name string) {
	pm.lock.Lock()
	defer pm.lock.Unlock()
	delete(pm.processes, name)
}

func (pm *ProcessManager) RemoveProcessFromSuccess(name string) {
	pm.lock.Lock()
	defer pm.lock.Unlock()
	delete(pm.successProcess, name)
}

func (pm *ProcessManager) GetProcess(name string) (*exec.Cmd, bool) {
	pm.lock.Lock()
	defer pm.lock.Unlock()
	cmd, ok := pm.processes[name]
	return cmd, ok
}

func (pm *ProcessManager) MoveProcessToSucess(name string) {
	pm.lock.Lock()
	defer pm.lock.Unlock()
	pm.successProcess[name] = pm.processes[name]
	pm.RemoveProcessFromSuccess(name)
}
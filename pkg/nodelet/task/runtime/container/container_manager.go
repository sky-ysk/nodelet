package container

import (
	"context"
	"strings"
	"sync"

	"github.com/docker/docker/api/types/container"
	client "github.com/docker/docker/client"
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/component-base/logs"
)

// 功能：管理容器的创建、删除和查询等操作
type ContainerManager struct {
	lock    sync.Mutex
	client  *client.Client    // Docker client for managing containers
	manager map[string]string // 用于runtime名称和容器ID的映射
}

func NewContainerManager(dockerClient *client.Client) *ContainerManager {
	return &ContainerManager{
		client:  dockerClient,
		manager: make(map[string]string),
	}
}

// 向manager里加入runtime映射
func (cm *ContainerManager) AddRuntimeMapping(runtimeName string, containerID string) bool {
	cm.lock.Lock()
	defer cm.lock.Unlock()
	if _, exists := cm.manager[runtimeName]; exists {
		logs.Errorf("Error: runtime %s already exists in manager", runtimeName)
		return false
	}
	cm.manager[runtimeName] = containerID
	logs.Infof("Added runtime mapping: %s -> %s", runtimeName, containerID)
	return true
}

// 从manager里删除runtime映射
func (cm *ContainerManager) DeleteRuntimeMapping(runtimeName string) bool {
	cm.lock.Lock()
	defer cm.lock.Unlock()
	if _, exists := cm.manager[runtimeName]; !exists {
		logs.Errorf("Error: runtime %s does not exist in manager", runtimeName)
		return false
	}
	delete(cm.manager, runtimeName)
	logs.Infof("Removed runtime mapping: %s", runtimeName)
	return true
}

// 获取所有容器列表
func (cm *ContainerManager) GetContainerList() (map[string]string, error) {
	cm.lock.Lock()
	defer cm.lock.Unlock()
	ctx := context.Background()
	containers, err := cm.client.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		logs.Errorf("Error:get containers failed. %v\n", err)
		return nil, err
	}
	container_list := make(map[string]string)
	// 遍历容器列表，存入所有容器信息
	for _, container := range containers {
		for _, name := range container.Names {
			container_list[name] = container.ID
			// logs.Infof("Container ID: %s, Container NAME: %s\n", container.ID, name)
			// logs.Infof("Container Image: %s, Container Status: %s\n", container.Image, container.Status)
		}
	}

	return container_list, nil
}

// 根据name查询ID
func (cm *ContainerManager) GetContainerID(name string) string {
	// 遍历所有容器，查找匹配的名称
	// 获取所有容器，使用docker的包
	// logs.Infof("iiiiiiiiiiDDDDDDDDDDDDDDDD")
	name = "/" + name // docker容器名称前面有一个斜杠
	containerList, err := cm.GetContainerList()
	if err != nil {
		logs.Errorf("Failed to get container list: %v", err)
		return ""
	}
	for containerName, containerID := range containerList {
		if containerName == name {
			return containerID
		}
	}
	logs.Errorf("Container with name %s not found", name)
	return ""
}

// 根据runtime名称获取容器ID（线程安全）
func (cm *ContainerManager) getContainerIDByRuntimeName(runtimeName string) (string, bool) {
	cm.lock.Lock()
	defer cm.lock.Unlock()

	containerID, exists := cm.manager[runtimeName]
	return containerID, exists
}

// 获取单个容器信息
func (cm *ContainerManager) GetContainer(id string) (*container.InspectResponse, error) {
	cm.lock.Lock()
	defer cm.lock.Unlock()
	ctx := context.Background()

	// 获取单个container，根据ID
	container, err := cm.client.ContainerInspect(ctx, id)
	if err != nil {
		logs.Errorf("Error:get container info failed. %v\n", err)
		return nil, err
	}
	return &container, nil
}

// 根据名称和镜像创建容器，后续可以扩展更多配置选项
func (cm *ContainerManager) CreateContainer(name string, runtime *apis.Runtime) (bool, string) {
	cm.lock.Lock()
	defer cm.lock.Unlock()
	ctx := context.Background()
	newCmd := append([]string{runtime.Spec.Command[0]}, runtime.Spec.Args...)
	// 使用的是group的文件夹作为工作目录
	testFolder := runtime.Status.Directory
	config := &container.Config{
		Image:      runtime.Spec.Image,
		Cmd:        newCmd,     // 示例命令
		WorkingDir: testFolder, // 设置工作目录
	}
	logs.Infof("docker container config:%v\n", config)

	hostConfig := &container.HostConfig{
		// 可以在这里添加更多的主机配置选项，例如端口映射、目录挂载
		// 目前主要使用到目录挂载，将runtime.directory挂载到容器的/tmp目录
		// 将runtime目录挂载到容器的对应目录,必须要文件仓库支持。:rw表示读写权限
		Binds: []string{testFolder + ":" + testFolder + ":rw"},
	}
	// 如果包含/，代表是绝对路径，特殊处理
	if strings.Contains(runtime.Spec.Args[0], "/") {
		execPath := runtime.Spec.Args[0]
		// 如果不适配文件仓库，那么需要先找到执行文件路径args[0]所在的目录
		if len(runtime.Spec.Args) > 0 {
			// 这里假设runtime.Spec.Args[0]是可执行文件的路径
			// （需要确保这个路径在容器内是可访问的？不确定，先测试）
			// 找到可执行文件路径，去掉可执行文件本身的字符串，例如/tmp/a.txt需要去掉/a.txt
			if lastSlash := strings.LastIndex(execPath, "/"); lastSlash != -1 {
				execPath = execPath[:lastSlash] // 截取到最后一个斜杠之前的部分
			}

			logs.Infof("Executable path in args: %s\n", execPath)
			// // 这里可以添加更多的配置选项，例如环境变量、工作目录等
			// config.Entrypoint = []string{newCmd} // 设置容器的入口
			// config.WorkingDir =
		}

		hostConfig = &container.HostConfig{
			// 可以在这里添加更多的主机配置选项，例如端口映射、目录挂载
			Binds: append(hostConfig.Binds, execPath+":"+execPath+":rw"),
		}
	}
	// 注：当不使用文件仓库的时候，需要挂载两个目录，一个是Directory对Directory，这是工作目录，保证生成文件最后是在Directory目录下
	// 另外一个挂载则是execPath对execPath，这是为了让容器能够访问这个绝对地址下的文件并且运行
	logs.Infof("config:%v, hostConfig:%v\n", config, hostConfig)

	resp, err := cm.client.ContainerCreate(ctx, config, hostConfig, nil, nil, name)
	if err != nil {
		logs.Errorf("Error:create container failed. %v\n", err)
		return false, ""
	}
	logs.Infof("Container created: %s\n", resp.ID)
	return true, resp.ID
}

// 启动容器
func (cm *ContainerManager) StartContainer(id string) bool {
	cm.lock.Lock()
	defer cm.lock.Unlock()
	ctx := context.Background()
	err := cm.client.ContainerStart(ctx, id, container.StartOptions{})
	if err != nil {
		logs.Errorf("Error: start container failed. %v\n", err)
		return false
	}
	logs.Infof("Container started: %s\n", id)
	return true
}

// 停止容器
func (cm *ContainerManager) StopContainer(id string) bool {
	cm.lock.Lock()
	defer cm.lock.Unlock()
	ctx := context.Background()
	err := cm.client.ContainerStop(ctx, id, container.StopOptions{})
	if err != nil {
		logs.Errorf("Error: stop container failed. %v\n", err)
		return false
	}
	logs.Infof("Container stopped: %s\n", id)
	return true
}

// 根据名称和镜像删除容器
// 注意：删除容器前需要先停止它
func (cm *ContainerManager) RemoveContainer(id string) (bool, error) {
	cm.lock.Lock()
	defer cm.lock.Unlock()
	ctx := context.Background()
	err := cm.client.ContainerRemove(ctx, id, container.RemoveOptions{Force: true})
	if err != nil {
		logs.Errorf("Error: remove container failed. %v\n", err)
		return false, err
	}
	logs.Infof("Container removed: %s\n", id)
	return true, nil
}

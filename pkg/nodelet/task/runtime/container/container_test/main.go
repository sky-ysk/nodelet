package main

import (
	"fmt"
	"time"

	"github.com/docker/docker/api/types/container"
	client "github.com/docker/docker/client"
	"golang.org/x/net/context"
	"hit.edu/framework/pkg/component-base/logs"
)

func main() {
	// 初始化Docker客户端
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// 列出所有容器
	time1 := time.Now()
	containers, err := cli.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	time2 := time.Now()
	fmt.Printf("Time taken to list containers: %v\n", time2.Sub(time1))

	time3 := time.Now()
	// 遍历容器列表，查找指定名称的容器
	for _, container := range containers {
		for _, name := range container.Names {
			if name == "/T1.G1.A1.R1-0197ed71-b0da-72a7-9fd6-450e6c192ee5" {
				// 打印容器的详细信息
				fmt.Printf("Container ID: %s, Container NAME: %s\n", container.ID, name)
				fmt.Printf("Container Image: %s, Container Status: %s\n", container.Image, container.Status)
				fmt.Printf("Container State: %+v\n", container.State)
				containerinfo, err := cli.ContainerInspect(ctx, container.ID)
				if err != nil {
					logs.Errorf("Error:get container info failed. %v\n", err)
					return 
				}
				fmt.Printf("Container Srate Status:%s\n", containerinfo.State.Status)
			}

		}
	}
	time4 := time.Now()
	fmt.Printf("Time taken to print container details: %v\n", time4.Sub(time3))

}

// func main1() {
// 	// 初始化Docker客户端
// 	ctx := context.Background()
// 	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
// 	if err != nil {
// 		fmt.Printf("Error initializing Docker client: %v\n", err)
// 		return
// 	}

// 	// 定义容器配置
// 	config := &container.Config{
// 		Image: "python:3.10-slim",
// 		// ExposedPorts: map[nat.Port]struct{}{
// 		// 	"80/tcp": {},
// 		// },
// 		// Cmd: []string{"nginx", "-g", "daemon off;"},
// 	}

// 	hostConfig := &container.HostConfig{
// 		// PortBindings: map[nat.Port][]nat.PortBinding{
// 		// 	"80/tcp": {{HostIP: "", HostPort: "8080"}},
// 		// },
// 	}

// 	// 创建容器
// 	resp, err := cli.ContainerCreate(ctx, config, hostConfig, nil, nil, "my-sky")
// 	if err != nil {
// 		fmt.Printf("Error creating container: %v\n", err)
// 		return
// 	}
// 	fmt.Printf("Container created with ID: %s\n", resp.ID)

// 	// 启动容器
// 	err = cli.ContainerStart(ctx, resp.ID, container.StartOptions{})
// 	if err != nil {
// 		fmt.Printf("Error starting container: %v\n", err)
// 		return
// 	}
// 	fmt.Printf("Container started with ID: %s\n", resp.ID)

// 	// 等待容器运行一段时间后关闭（可选）
// 	time.Sleep(20 * time.Second)

// 	// 停止容器
// 	err = cli.ContainerStop(ctx, resp.ID, container.StopOptions{})
// 	if err != nil {
// 		fmt.Printf("Error stopping container: %v\n", err)
// 		return
// 	}
// 	fmt.Printf("Container stopped with ID: %s\n", resp.ID)

// 	// 删除容器
// 	err = cli.ContainerRemove(ctx, resp.ID, container.RemoveOptions{Force: true})
// 	if err != nil {
// 		fmt.Printf("Error removing container: %v\n", err)
// 		return
// 	}
// 	fmt.Printf("Container removed with ID: %s\n", resp.ID)
// }

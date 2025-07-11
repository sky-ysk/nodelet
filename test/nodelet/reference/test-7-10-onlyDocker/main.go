package main

import (
	"fmt"
	"os"
	"os/exec"
)

func main() {

	// // 用户输入镜像名称
	// fmt.Print("请输入要拉取的 Docker 镜像名称（例如：nginx）：")
	// var imageName string
	// fmt.Scanln(&imageName)

	// // 拉取镜像
	// fmt.Printf("正在拉取镜像 %s...\n", imageName)
	// output, err := runCommand(fmt.Sprintf("docker pull %s", imageName))
	// if err != nil {
	// 	fmt.Printf("拉取镜像失败：\n%s\n", output)
	// 	return
	// }
	// fmt.Printf("镜像拉取成功：\n%s\n", output)

	// 启动容器
	// fmt.Printf("正在启动容器...\n")
	// cmd := "docker"
	// homeDir := os.Getenv("HOME")
	// fmt.Printf("Home Directory: %s\n", homeDir)
	// args := []string{"run", "-itd", "-v", "/dev/shm:/dev/shm", "--ipc=host", "-e", "RMW_IMPLEMENTATION=rmw_fastrtps_cpp", "--volume", homeDir + "/.Xauthority:/root/.Xauthority", "--volume", "/tmp/.X11-unix:/tmp/.X11-unix", "--volume", "/dev/dri:/dev/dri", "--device=/dev/snd", "--device=/dev/dri", "--env", "QT_X11_NO_MITSHM=1", "--env", "DISPLAY", "--env", "ROS_DOMAIN_ID=41", "--network=host", "--entrypoint=/mechmind_yolo.sh", "--name=camera_container", "74cea28bb666"}
	// // args := []string{"run", "-id", "--name=sky", "-v", "/home:/home", "-v", homeDir+"/tmp:/root/tmp", "python:3.10-slim"}
	// CMD := exec.Command(cmd, args...)
	// CMD.Env = os.Environ()
	// if err := CMD.Run(); err != nil {
	// 	fmt.Printf("启动容器失败：%v\n", err)
	// 	return
	// }
	// fmt.Printf("容器启动成功\n")

	// args := []string{"run", "-itd", "-v", "/dev/shm:/dev/shm", "--ipc=host", "-e", "RMW_IMPLEMENTATION=rmw_fastrtps_cpp", "--volume", "$HOME/.Xauthority:/root/.Xauthority", "--volume", "/tmp/.X11-unix:/tmp/.X11-unix", "--volume", "/dev/dri:/dev/dri", "--device=/dev/snd", "--device=/dev/dri", "--env", "QT_X11_NO_MITSHM=1", "--env", "DISPLAY", "--env", "ROS_DOMAIN_ID=41", "--network=host", "--entrypoint=/mechmind_yolo.sh", "--name=camera_container", "74cea28bb666"}
	// // 解析其中的环境变量，使用os.enpandenv
	// // before
	// fmt.Printf("Before expansion: %s\n", args)
	// for i, arg := range args {
	// 	args[i] = os.ExpandEnv(arg)
	// }
	// //打印结果
	// fmt.Printf("Docker run command: docker %s\n", args)

	// docker run -it -v $HOME:/home -v $HOME/tmp:/root/tmp python:3.10-slim
	cmd := "docker"
	// args := []string{"run", "-it", "-v", "$HOME:/home", "-v", "$HOME/tmp:/root/tmp", "--name=sky", "python:3.10-slim"}
	args := []string{"run", "-itd", "-v", "$HOME:/home", "--name=sky", "python:3.10-slim"}
	for i, arg := range args {
		args[i] = os.ExpandEnv(arg)
	}
	fmt.Printf("Docker run command: %s %v\n", cmd, args)
	CMD := exec.Command(cmd, args...)
	CMD.Env = os.Environ() // 获取当前环境的环境变量
	if err := CMD.Run(); err != nil {
		fmt.Printf("启动容器失败：%v\n", err)
		return
	}
	fmt.Printf("容器启动成功\n")

}

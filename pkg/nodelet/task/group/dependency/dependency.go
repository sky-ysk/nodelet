// Dependency作用是检查程序的包依赖是否满足（目前针对python任务），具体功能有：
// check:检查设备上的虚拟环境与提供的requirememts.txt是否能够满足
// setUpEnv:根据requirememts.txt创建新的虚拟环境，一般在不满足依赖的情况下调用
// inputEnv:注入环境变量，将系统自身的Env加上指定虚拟环境的python的环境变量加入到PATH之后，注入到runtime的Env中，在command.go中执行的时候直接加入到cmd.Env即可
package dependency

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"time"

	"hit.edu/framework/pkg/component-base/logs"
)

type Requirement struct {
	Name    string
	Version string
}

// 查找当前conda的  所有虚拟环境，判断是否有虚拟环境满足这个requirements.txt的依赖
func CheckEnvironmentSatisfy(requirementsPath string) (string, bool) {
	//获取requirements里面包含的所有package以及version
	requirements, err := ParseRequirements(requirementsPath)
	if err != nil {
		fmt.Printf("Error reading requirements file: %v\n", err)
		return "", false
	}

	//获取每一个虚拟环境里面包含的所有package以及version
	//获取所有虚拟环境的名称
	envName, err := GetAllCondaEnv()
	if err != nil {
		fmt.Printf("Error retrieving installed packages: %v\n", err)
		return "", false
	}
	//for遍历所有虚拟环境
	for _, envname := range envName {
		installed, err := GetInstalledPackages(envname)
		if err != nil {
			fmt.Printf("Error retrieving installed packages: %v\n", err)
			return "", false
		}

		if CheckRequirements(requirements, installed, envname) {
			fmt.Printf("All requirements are satisfied.")
			return envname, true
		} else {
			fmt.Printf("Some requirements are not satisfied. EnvName:%v", envname)
		}
	}
	return "", false
}

// 解析requirements.txt文件，返回包的切片
func ParseRequirements(filePath string) ([]Requirement, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var requirements []Requirement
	re := regexp.MustCompile(`(?P<Name>[a-zA-Z0-9_-]+)(==(?P<Version>[0-9\.]+))?`)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue // Skip empty lines and comments
		}
		match := re.FindStringSubmatch(line)
		if match != nil {
			req := Requirement{
				Name:    match[1],
				Version: match[3],
			}
			requirements = append(requirements, req)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return requirements, nil
}

// 使用pip list获取pip freeze格式（numpy==1.23.1这种）的字符数组
func GetInstalledPackages(envName string) (map[string]string, error) {

	cmd := exec.Command("conda", "run", "-n", envName, "pip", "list", "--format=freeze")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	installed := make(map[string]string)
	re := regexp.MustCompile(`(?P<Name>[a-zA-Z0-9_-]+)==(?P<Version>[0-9\.]+)`)
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		line := scanner.Text()
		match := re.FindStringSubmatch(line)
		if match != nil {
			installed[match[1]] = match[2]
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return installed, nil
}

// 将获取到的requirements.txt的内容与已有的installed的内容比较
func CheckRequirements(requirements []Requirement, installed map[string]string, envName string) bool {
	allSatisfied := true
	for _, req := range requirements {
		installedVersion, found := installed[req.Name]
		if !found {
			logs.Info("Package %v is not installed.", req.Name)
			allSatisfied = false
			return allSatisfied
		} else if req.Version != "" && installedVersion < req.Version {
			//找到对应的package但是版本落后
			logs.Info("Package %v version mismatch: required %v, installed %v.", req.Name, req.Version, installedVersion)
			allSatisfied = false
			return allSatisfied
		} else {
			// fmt.Printf("Package %s is satisfied.\n", req.Name)
		}
	}
	logs.Info("requirements satisfied envName: %v", envName)
	return allSatisfied
}

// 根据已有的requirements.txt创建新的conda环境
// 有些包可能使用conda create -n --file requirements.txt无法从conda官网安装，这里采取先新创建空的
// conda虚拟环境，然后再执行pip install，这是手动配置环境的正常流程
// 可能存在一个问题，创建新的虚拟环境的试时候需要指定python版本，目前默认3.9版本，后续需要调整
func SetupEnvironment(reqFile string, runtiemName string) bool {
	pythonVersion := "python=3.9"
	nowtime := time.Now().Format("2006-01-02 15:04:05")
	newEnvName := runtiemName + nowtime
	newEnvName = strings.Replace(newEnvName, ":", "-", -1)
	newEnvName = strings.Replace(newEnvName, " ", "-", -1)
	// reqFile := "requirements.txt"

	if _, err := os.Stat(reqFile); os.IsNotExist(err) {
		logs.Error("requirements.txt not found: %v", err)
		return false
	}

	createEnvCmd := "conda"
	createEnvArgs := []string{"create", "-n", newEnvName, pythonVersion}
	CMDCreate := exec.Command(createEnvCmd, createEnvArgs...)
	logs.Info("Creating a new Conda environment...")
	if err := CMDCreate.Start(); err != nil {
		logs.Error("Failed to start create Conda environment: %v", err)
		return false
	}

	if err := CMDCreate.Wait(); err != nil {
		logs.Info("Failed to finish create Conda environment: %v", err)
		return false
	}

	//需要指定run -n的名称
	installEnvCmd := "conda"
	installEnvArgs := []string{"run", "-n", newEnvName, "pip", "install", "-r", reqFile}
	CMDInstall := exec.Command(installEnvCmd, installEnvArgs...)
	logs.Info("Installing for new Conda environment...")
	if err := CMDInstall.Start(); err != nil {
		logs.Error("Failed to start install dependency requirements.txt: %v", err)
		return false
	}

	if err := CMDInstall.Wait(); err != nil {
		logs.Info("Failed to finish install dependency requirements.txt: %v", err)
		return false
	}

	logs.Info("Environment '%v' created successfully with requirements from %v", newEnvName, reqFile)
	return true
}

// conda remove --name ENV_NAME --all
// TODO 存在一种场景，有其他程序正在使用这个虚拟环境，这时候删除会出现未知的结果，可能是删除失败
func RemoveEnvironment(envName string) bool {
	removeEnvCmd := "conda"
	removeEnvArgs := []string{"remove", "-n", envName, "--all"}
	logs.Info("Removing Conda environment:%v", envName)
	if err := runCommand(removeEnvCmd, removeEnvArgs...); err != nil {
		logs.Error("Failed to remove Conda environment:%v. err info: %v", envName, err)
		return false
	}

	logs.Info("Environment '%v' removed successfully", envName)
	return true
}

func runCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("Command output: %s\n", out.String())
		fmt.Printf("Command error: %s\n", stderr.String())
		return fmt.Errorf("command execution failed: %v", err)
	}

	fmt.Printf("Command output: %s\n", out.String())
	return nil
}

// 根据传入的虚拟环境名称，生成一个用于runtime的ENV属性的Value并返回
func EnvForInput(envName string) (string, bool) {
	// pythonPath := "/home/ysk/miniconda3/envs/py38/bin" // 替换为Python 安装路径
	// pythonPath := "" // 替换为Python 安装路径
	// cmd := exec.Command("/bin/bash","-c","source /home/ysk/miniconda3/etc/profile.d/conda.sh;conda activate py38;" + "which python;python " + "predict.py")   // 启动 bash 终端
	// cmd := exec.Command("/bin/bash","-c","echo $PATH;which python;python " + "predict.py")   // 启动 bash 终端
	// 获取当前环境变量

	//先找到这个名字的虚拟环境的path
	pythonPath, err := GetCondaEnvPath(envName)

	if err != nil {
		logs.Errorf("get conda Env err! envName: %v", envName)
		return "", false
	}
	//填充虚拟环境路径后的/bin，python解释器在这里
	pythonPath = pythonPath + "/bin/python"
	// env := os.Environ()

	// // 初始化新的环境变量列表,两种方法，删除base加入新的虚拟环境；
	// //第二种：在PATH最前面拼接新的虚拟环境，这样就不会被base替代---目前是这种
	// var updatedEnv string

	// for _, e := range env {
	// 	if strings.HasPrefix(e, "PATH=") {
	// 		// 获取当前 PATH 的值
	// 		currentPath := strings.TrimPrefix(e, "PATH=")
	// 		// fmt.Printf("e: %v", e)
	// 		// var newPath = "PATH=" + pythonPath + currentPath
	// 		//下面的写法是只保留PATH=之后的value的内容，后续可以根据需要增加其他环境变量
	// 		//export PATH=$(echo "$PATH" | awk -v RS=: '!a[$0]++' | tr '\n' :)  去除PATH中重复的路径，暂时认为重复会不会有影响
	// 		var newPath = pythonPath + ":" + currentPath
	// 		// updatedEnv = append(updatedEnv, newPath)
	// 		updatedEnv = newPath
	// 	} else {
	// 		// 其他环境变量保持不变
	// 		//下面注释掉的话，updatedEnv只会保留PATH这一条环境变量
	// 		// updatedEnv = append(updatedEnv, e)
	// 	}
	// }
	// return updatedEnv
	return pythonPath, true
}

func GetAllCondaEnv() ([]string, error) {
	cmd := exec.Command("conda", "info", "--envs")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		fmt.Printf("Error executing conda command: %v\n", err)
		return nil, err
	}

	// 解析输出
	output := out.String()
	lines := strings.Split(output, "\n")

	// 存储环境名称
	var envNames []string

	// 遍历每一行，提取虚拟环境名称
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) > 1 && !strings.HasPrefix(fields[0], "#") {
			envNames = append(envNames, fields[0]) // 添加环境名称
		}
	}
	return envNames, nil
}

// GetCondaEnvPath 根据传入的虚拟环境名称，返回虚拟环境的路径
func GetCondaEnvPath(envName string) (string, error) {
	// 执行 "conda info --envs" 命令，获取所有虚拟环境的信息  命令conda env list 也可以
	cmd := exec.Command("conda", "info", "--envs")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("failed to execute conda command: %v", err)
	}

	// 解析输出
	output := out.String()
	lines := strings.Split(output, "\n")

	// 遍历每一行，查找目标环境
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) > 1 && fields[0] == envName {
			return fields[len(fields)-1], nil // 返回路径
		}
	}

	// 如果没有找到环境
	return "", fmt.Errorf("environment '%s' not found", envName)
}

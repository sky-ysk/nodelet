package dependency

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"hit.edu/framework/pkg/component-base/logs"
	// "time"
)

// 查找当前conda的所有虚拟环境，判断是否有虚拟环境满足这个requirements.txt的依赖
func checkEnvironmentSatisfy(requirementsPath string) (string, error) {
	// 列出所有环境
	listCmd := exec.Command("conda", "env", "list")
	var listOut bytes.Buffer
	var listErr bytes.Buffer
	listCmd.Stdout = &listOut
	listCmd.Stderr = &listErr
	if err := listCmd.Run(); err != nil {
		fmt.Printf("Failed to list Conda environments: %s\n", listErr.String())
		return "", fmt.Errorf("failed to list Conda environments: %v", err)
	}

	envLines := strings.Split(listOut.String(), "\n")
	envNames := []string{}
	for _, line := range envLines {
		fields := strings.Fields(line)
		if len(fields) > 0 && !strings.HasPrefix(fields[0], "#") {
			envNames = append(envNames, fields[0])
		}
	}

	//依次检查是否满足
	for _, envName := range envNames {
		fmt.Printf("Checking environment: %s\n", envName)
		checkCmd := exec.Command("conda", "run", "--name", envName, "pip", "check", "--requirement", requirementsPath)
		var checkOut bytes.Buffer
		var checkErr bytes.Buffer
		checkCmd.Stdout = &checkOut
		checkCmd.Stderr = &checkErr
		if err := checkCmd.Run(); err == nil {
			fmt.Printf("Environment '%s' satisfies the requirements.\n", envName)
			return envName, nil
		} else {
			fmt.Printf("Environment '%s' does not satisfy the requirements: %s\n", envName, checkErr.String())
		}
	}

	return "", nil //没有满足的
}

// 直接run test即可根据已有的requirements.txt创建新的conda环境
func setupCondaEnvironment(envName, reqFile string) {
	// envName := "test_env"
	// reqFile := "requirements.txt"

	if _, err := os.Stat(reqFile); os.IsNotExist(err) {
		logs.Error("requirements.txt not found: %v", err)
	}

	createEnvCmd := "conda"
	createEnvArgs := []string{"create", "-n", envName, "python=3.9", "--file", reqFile}
	logs.Info("Creating new Conda environment...")
	if err := runCommand(createEnvCmd, createEnvArgs...); err != nil {
		logs.Error("Failed to create Conda environment: %v", err)
	}

	logs.Info("Environment '%s' created successfully with requirements from %s", envName, reqFile)
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

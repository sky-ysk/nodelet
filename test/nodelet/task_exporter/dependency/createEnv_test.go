package main

import (
	"fmt"
	"os"
	"testing"
	// "time"
)

// 直接run test即可根据已有的requirements.txt创建新的conda环境
func TestSetupCondaEnvironment(t *testing.T) {
	printEnv()
	envName := "test_env"
	reqFile := "requirements.txt"

	if _, err := os.Stat(reqFile); os.IsNotExist(err) {
		t.Fatalf("requirements.txt not found: %v", err)
	}

	createEnvCmd := "conda"
	createEnvArgs := []string{"create", "-n", envName, "python=3.9", "--file", reqFile}
	// createEnvArgs := []string{"env", "list"}
	// createEnvArgs := []string{"--version"}
	t.Log("Creating new Conda environment...")
	if err := runCommand(createEnvCmd, createEnvArgs...); err != nil {
		t.Fatalf("Failed to create Conda environment: %v", err)
	}

	//其他conda命令，可以用于激活已有的conda环境然后再安装requirements,但是有顺序要求，activate之前要init
	//目前在一个cmd里面先后执行两个命令还需要测试

	// initEnvCmd := "conda"
	// initEnvArgs := []string{"init", "--all"}
	// t.Log("activating new Conda environment...")
	// if err := runCommand(initEnvCmd, initEnvArgs...); err != nil {
	// 	t.Fatalf("Failed to activate Conda environment: %v", err)
	// }

	// activateEnvCmd := "conda"
	// activateEnvArgs := []string{"activate", envName}
	// t.Log("activating new Conda environment...")
	// if err := runCommand(activateEnvCmd, activateEnvArgs...); err != nil {
	// 	t.Fatalf("Failed to activate Conda environment: %v", err)
	// }

	// installEnvCmd := "conda"
	// installEnvArgs := []string{"install", "--file", reqFile}
	// t.Log("installing new Conda environment...")
	// if err := runCommand(installEnvCmd, installEnvArgs...); err != nil {
	// 	t.Fatalf("Failed to install Conda environment: %v", err)
	// }

	t.Logf("Environment '%s' created successfully with requirements from %s", envName, reqFile)
}

func printEnv() {
	for _, env := range os.Environ() {
		fmt.Println(env)
	}
}

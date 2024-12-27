package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestSetupOrReuseCondaEnvironment(t *testing.T) {
	envName := "test_env"
	reqFile := "requirements.txt"

	if _, err := os.Stat(reqFile); os.IsNotExist(err) {
		t.Fatalf("requirements.txt not found: %v", err)
	}

	t.Log("Checking existing Conda environments...")
	existingEnv, err := findExistingEnvironment(envName, reqFile)
	if err != nil {
		t.Fatalf("Failed to check existing environments: %v", err)
	}

	if existingEnv {
		t.Logf("An existing environment '%s' satisfies the requirements.", envName)
		return
	}

	t.Log("Creating a new Conda environment...")
	createEnvCmd := "conda"
	createEnvArgs := []string{"create", "--name", envName, "--file", reqFile, "-y"}
	if err := runCommand(createEnvCmd, createEnvArgs...); err != nil {
		t.Fatalf("Failed to create Conda environment: %v", err)
	}

	t.Logf("Environment '%s' created successfully with requirements from %s", envName, reqFile)
}

//检查指定虚拟环境的包是否满足依赖
func findExistingEnvironment(envName, reqFile string) (bool, error) {

	listCmd := exec.Command("conda", "env", "list")
	var out bytes.Buffer
	listCmd.Stdout = &out
	if err := listCmd.Run(); err != nil {
		return false, fmt.Errorf("failed to list Conda environments: %v", err)
	}

	// 环境是否存在
	envs := out.String()
	if !strings.Contains(envs, envName) {
		return false, nil
	}

	// 检查已有的依赖
	tempCmd := exec.Command("conda", "run", "--name", envName, "pip", "check","--requirement", reqFile)
	var checkOut bytes.Buffer
	tempCmd.Stdout = &checkOut
	tempCmd.Stderr = &checkOut
	if err := tempCmd.Run(); err != nil {
		return false, nil // 不满足依赖需求返回false
	}

	return true, nil
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

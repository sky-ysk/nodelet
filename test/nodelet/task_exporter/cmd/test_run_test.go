package run_test

import (
	"fmt"
	"os/exec"
	"testing"
	"os"
)

func TestRun(t *testing.T) {
	cmd := "cmd"
	args := []string{"/c", "python resourcelet\\test\\nodelet\\task_exporter\\cmd\\train.py"}
	CMD := exec.Command(cmd, args...)
	out, err := CMD.CombinedOutput()
	if err != nil {
		fmt.Println("Failed to run cmd:", cmd, args)
		return
	}

	CMD.Stdin = os.Stdin
	CMD.Stdout = os.Stdout

	fmt.Printf("combined out:\n%s\n", string(out))
}

func TestRun1(t *testing.T) {
	cmd := "ls"
	args := "-a"
	CMD := exec.Command(cmd, args)
	err := CMD.Run()
	if err != nil {
		fmt.Println("Failed to run cmd:", cmd, args)
		return
	}
}

func TestMain(t *testing.T) {
    cmd := exec.Command("cmd", "ls", "-a")
    out, err := cmd.CombinedOutput()
    if err != nil {
        fmt.Println("cmd-Run() failed with err")
    }
    fmt.Printf("combined out:\n%s\n", string(out))
}
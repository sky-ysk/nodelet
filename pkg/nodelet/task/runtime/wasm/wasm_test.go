package wasm

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/apis/meta"

	"hit.edu/framework/pkg/component-base/logs"
)

func TestForWasm(t *testing.T) {
	newRuntime := apis.Runtime{
		Name:    "CMD",
		Image:   "/home/hzy/goproject/new-wasm/resourcelet/test/wasm/wasm_task/printf.wasm", //暂时以文件本地地址进行测试
		Type:    apis.ByWasm,
		Command: []string{},
		Args:    []string{},
	}
	newAction := apis.Action{
		Spec: apis.ActionSpec{
			Name: "TestAction",
			Runtimes: []apis.Runtime{
				newRuntime,
			},
		},
	}
	newGroup := apis.Group{
		ObjectMeta: meta.ObjectMeta{Name: "cmd_yolo_train"},
		Spec: apis.GroupSpec{
			Name:    "TestGroup",
			Parents: make([]string, 0),
			Actions: []apis.Action{
				newAction,
			},
		},
		Status: apis.GroupStatus{
			GroupID: "test-train-group",
		},
	}

	logs.Info("---TestForWasm---")
	wasm_runtime := NewWasmRuntime()
	// 运行时会跑在独立的进程,所以程序结束一定要kill进程,不然始终在运行
	defer wasm_runtime.StopCMD()

	time.Sleep(1 * time.Second)
	wasm_runtime.Run(&newGroup, &newAction, &newRuntime)
	time.Sleep(1 * time.Second)
	wasm_runtime.Destory()
}

func TestForAIWasm(t *testing.T) {
	// 下载wasm toolchain

	newRuntime := apis.Runtime{
		Name:    "wasm-test-ai-task",
		Image:   "/tmp/wasm/onnx.wasm", //暂时以文件本地地址进行测试
		Type:    apis.ByWasm,
		Command: []string{},
		Args:    []string{},
		// EnvVar: []apis.EnvVar{
		// 	{Name: "FIXTURES_DIR", Value: "/home/kcm/tmp/wasm/fixtures"},
		// },
	}
	newAction := apis.Action{
		Spec: apis.ActionSpec{
			Name: "TestAction",
			Runtimes: []apis.Runtime{
				newRuntime,
			},
		},
	}
	newGroup := apis.Group{
		ObjectMeta: meta.ObjectMeta{Name: "wasm_inference"},
		Spec: apis.GroupSpec{
			Name:    "TestGroup",
			Parents: make([]string, 0),
			Actions: []apis.Action{
				newAction,
			},
		},
		Status: apis.GroupStatus{
			GroupID: "test-wasm-inference-group",
		},
	}

	logs.Info("---TestForAIWasm---")
	wasm_runtime := NewWasmRuntime()
	// 运行时会跑在独立的进程,所以程序结束一定要kill进程,不然始终在运行
	defer wasm_runtime.StopCMD()

	wasm_runtime.Run(&newGroup, &newAction, &newRuntime)
	time.Sleep(1 * time.Second)
	wasm_runtime.Destory()

}

func DownloadFile(url string, filepath string) error {
	response, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("获取文件失败: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("服务器返回状态: %s", response.Status)
	}
	// 获得响应reader
	reader := bufio.NewReaderSize(response.Body, 32*1024)

	// 创建目标文件，获得writer
	outFile, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("创建文件失败: %v", err)
	}
	defer outFile.Close()
	writer := bufio.NewWriter(outFile)

	// 将响应体写入目标文件
	_, err = io.Copy(writer, reader)
	if err != nil {
		return fmt.Errorf("写入文件失败: %v", err)
	}

	return nil
}

func TestForDownload(t *testing.T) {
	ip := "127.0.0.1"
	projectName := "/tmp/wasm/toolchain"
	fileName := "server"
	// filepath := "wasm-llvm"

	url := fmt.Sprintf("http://%s/get?projectname=%s&filename=%s", ip, projectName, fileName)
	filepath := "downloaded_file.txt" // 指定下载后保存的文件路径

	// 下载文件
	if err := DownloadFile(url, filepath); err != nil {
		fmt.Println("下载错误:", err)
		return
	}

	fmt.Println("文件下载成功:", filepath)
}

package main

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"hit.edu/framework/pkg/apimachinery/runtime"
	"hit.edu/framework/pkg/apimachinery/runtime/schema"
	"hit.edu/framework/pkg/apimachinery/runtime/serializer"
	"hit.edu/framework/pkg/apimachinery/watch"
	apis "hit.edu/framework/pkg/apis/cores"
	metav1 "hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/client-go/clients"
	"hit.edu/framework/pkg/client-go/rest"
	"hit.edu/framework/pkg/component-base/logs"
)

// 创建一个Rest Client
// 验证xxx动词
// 与API Server通信，并执行基础操作

func main() {
	//---临时参数:以cmd任务形式运行wasm任务
	// 目前需保证/tmp/wasm的前缀不可改变
	// cmd := []string{"/tmp/wasm/toolchain/wa2x-wasi-nn"}
	// arg := []string{"/tmp/wasm/onnx.so"}
	//---
	moduleName := "testModule"
	logs.Init(moduleName)
	scheme := runtime.NewScheme()
	apis.AddToScheme(scheme)
	logs.Info(scheme)
	//参数配置
	//部分参数之后可以在core_client等 编写setConfigDefaults函数进行填充
	c := &rest.Config{
		Host:    "http://localhost:10000",
		APIPath: "/apis/resources/v1",
		ContentConfig: rest.ContentConfig{
			AcceptContentTypes: "application/json; charset=UTF-8", //text/plain; charset=UTF-8
			ContentType:        "application/json; charset=UTF-8", //application/json; charset=UTF-8
			GroupVersion: &schema.GroupVersion{
				Group:   "resources",
				Version: "v1",
			},
			NegotiatedSerializer: serializer.NewCodecFactory(scheme),
		},
		UserAgent: "defaultUserAgent",
		Transport: &http.Transport{
			MaxIdleConns:        100,              // 最大空闲连接数
			IdleConnTimeout:     90 * time.Second, // 空闲连接超时时间
			TLSHandshakeTimeout: 10 * time.Second, // TLS 握手超时时间
		},
		Timeout: 10 * time.Second,
	}

	//创建ClientSet
	clientSet, err := clients.NewForConfig(c)
	if err != nil {
		panic(err)
	}
	// 资源定义在 pkg/apis/xxx/type.go 下
	// 这里以访问资源Task为例，
	// 获取访问Task的客户端
	// 默认访问的Namespace是 ""

	tasksClient := clientSet.Core().Tasks("test")
	groupsClient := clientSet.Core().Groups("test")

	var runtimeCommand = apis.Runtime{
		Name:                         "yolo-cmd",
		Type:                         apis.ByCommand,
		Command:                      []string{"python3"},
		Args:                         []string{"/home/kcm/workplace/migration-demo-0116/yolo-runner.py"},
		EnableFineGrainedControl:     true,
		EnableFineGrainedControlPort: "5123",
	}

	// var runtimeWasm = apis.Runtime{
	// 	Name:                         "wasm-test-ai-task",
	// 	Image:                        "/tmp/wasm/onnx.wasm", //暂时以文件本地地址进行测试
	// 	Type:                         apis.ByWasm,
	// 	Command:                      []string{},
	// 	Args:                         []string{},
	// 	EnableFineGrainedControl:     true,
	// 	EnableFineGrainedControlPort: "8080",
	// 	// EnvVar: []apis.EnvVar{
	// 	// 	{Name: "FIXTURES_DIR", Value: "/home/kcm/tmp/wasm/fixtures"},
	// 	// },
	// }

	task := &apis.Task{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "TestTask-wasm",
			Namespace: "test",
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Task",
			APIVersion: "resources/v1",
		},
		Spec: apis.TaskSpec{
			Name: "demo-task",
			Groups: []apis.Group{
				apis.Group{
					ObjectMeta: metav1.ObjectMeta{Name: "TestGroup-wasm", Namespace: "test"},
					TypeMeta:   metav1.TypeMeta{Kind: "Group", APIVersion: "resources/v1"},
					Spec: apis.GroupSpec{
						Name:    "TestGroup-wasm",
						Parents: make([]string, 0),
						Actions: []apis.Action{
							apis.Action{
								ObjectMeta: metav1.ObjectMeta{Name: "wasm_action"},
								Spec: apis.ActionSpec{
									Name: "wasm_action",
									Runtimes: []apis.Runtime{
										runtimeCommand,
									},
								},
								Status: apis.ActionStatus{
									ActionID: "wasm_action:TestGroup-wasm:test-task", // ActionID =ActionName + GroupID
									Phase:    apis.Unknown,
									RuntimeStatus: []apis.RuntimeStatus{
										apis.RuntimeStatus{
											RuntimeID: "wasm-cmd:wasm_action:TestGroup-wasm:test-task", // RuntimeID = RuntimeName + ActionID
											Phase:     apis.Unknown,
										},
									},
								},
							},
						},
					},
					Status: apis.GroupStatus{
						GroupID: "TestGroup-wasm:test-task",
						ActionStatus: []apis.ActionStatus{
							apis.ActionStatus{
								ActionID: "wasm_action:TestGroup-wasm:test-task",
								RuntimeStatus: []apis.RuntimeStatus{
									apis.RuntimeStatus{
										RuntimeID: "wasm-cmd:wasm_action:TestGroup-wasm:test-task",
										Phase:     apis.Unknown,
									},
								},
								Phase: apis.Unknown,
							},
						},
						Belongs: apis.IDRef{TaskID: "test-task"},
					},
				},
			},
		},
		Status: apis.TaskStatus{
			TaskID: "test-task",
			Phase:  apis.Unknown,
			GroupStatus: []apis.GroupStatus{
				apis.GroupStatus{
					GroupID: "TestGroup-wasm:test-task",
					ActionStatus: []apis.ActionStatus{
						apis.ActionStatus{
							ActionID: "wasm_action:TestGroup-wasm:test-task",
							RuntimeStatus: []apis.RuntimeStatus{
								apis.RuntimeStatus{
									RuntimeID: "wasm-cmd:wasm_action:TestGroup-wasm:test-task",
									Phase:     apis.Unknown,
								},
							},
							Phase: apis.Unknown,
						},
					},
					Belongs: apis.IDRef{TaskID: "test-task"},
				},
			},
		},
	}

	group1 := &apis.Group{
		ObjectMeta: metav1.ObjectMeta{Name: "TestGroup-wasm", Namespace: "test"},
		TypeMeta:   metav1.TypeMeta{Kind: "Group", APIVersion: "resources/v1"},
		Spec: apis.GroupSpec{
			Name:    "TestGroup-wasm",
			Parents: make([]string, 0),
			Actions: []apis.Action{
				apis.Action{
					ObjectMeta: metav1.ObjectMeta{Name: "wasm_action"},
					Spec: apis.ActionSpec{
						Name: "wasm_action",
						Runtimes: []apis.Runtime{
							runtimeCommand,
						},
					},
					Status: apis.ActionStatus{
						ActionID: "wasm_action:TestGroup-wasm:test-task", // ActionID =ActionName + GroupID
						Phase:    apis.Unknown,
						RuntimeStatus: []apis.RuntimeStatus{
							apis.RuntimeStatus{
								RuntimeID: "wasm-cmd:wasm_action:TestGroup-wasm:test-task", // RuntimeID = RuntimeName + ActionID
								Phase:     apis.Unknown,
							},
						},
					},
				},
			},
		},
		Status: apis.GroupStatus{
			GroupID: "TestGroup-wasm:test-task",
			ActionStatus: []apis.ActionStatus{
				apis.ActionStatus{
					ActionID: "wasm_action:TestGroup-wasm:test-task",
					RuntimeStatus: []apis.RuntimeStatus{
						apis.RuntimeStatus{
							RuntimeID: "wasm-cmd:wasm_action:TestGroup-wasm:test-task",
							Phase:     apis.Unknown,
						},
					},
					Phase: apis.Unknown,
				},
			},
			Belongs: apis.IDRef{TaskID: "test-task"},
		},
	}

	//监听事件并打印  监听resources/v1/tasks
	go func() {
		logs.Infof("watching")
		watchOptions := metav1.ListOptions{}

		watcher, err := tasksClient.Watch(context.TODO(), watchOptions)
		if err != nil {
			panic(err)
		}
		defer watcher.Stop() // 确保 watcher 被停止

		// 获取事件通道
		watchChan := watcher.ResultChan()

		for {
			select {
			case event, ok := <-watchChan:
				if !ok {
					logs.Infof("watchChan closed")
					return
				}

				// 打印事件类型和对象的相关信息
				logs.Infof("接收到事件类型: %v\n", event.Type)
				switch event.Type {
				case watch.Added:
					logs.Infof("资源被添加: ", event.Object)
				case watch.Modified:
					logs.Infof("资源被修改: ", event.Object)
				case watch.Deleted:
					logs.Infof("资源被删除: ", event.Object)
				case watch.Error:
					logs.Infof("发生错误: ", event.Object)
				default:
					logs.Infof("未识别的事件类型: ", event.Type)
				}
			}
		}
	}()

	//如果已经存在，先删掉

	err = tasksClient.Delete(context.TODO(), "TestTask-wasm", metav1.DeleteOptions{})
	err1 := groupsClient.Delete(context.TODO(), "TestGroup-wasm", metav1.DeleteOptions{})
	// Create一个Task
	logs.Infof("creating")
	results, err := tasksClient.Create(context.TODO(), task, metav1.CreateOptions{})
	results2, err1 := groupsClient.Create(context.TODO(), group1, metav1.CreateOptions{})

	if err != nil {
		logs.Errorf("Failed to create task: %v", err)
		panic(err)
	}
	if err1 != nil {
		logs.Errorf("Failed to create task: %v", err)
		panic(err)
	}
	logs.Infof("Created task ", results)
	logs.Infof("Created group1 ", results2)
	prompt()
}

// From K8s
func prompt() {
	fmt.Printf("-> Press Return key to continue.")
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		break
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}
	logs.Info()
}

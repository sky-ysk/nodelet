package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"hit.edu/framework/pkg/apimachinery/runtime"
	"hit.edu/framework/pkg/apimachinery/runtime/schema"
	"hit.edu/framework/pkg/apimachinery/runtime/serializer"
	"hit.edu/framework/pkg/apimachinery/types"
	"hit.edu/framework/pkg/apimachinery/watch"
	apis "hit.edu/framework/pkg/apis/cores"
	metav1 "hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/client-go/clients"
	"hit.edu/framework/pkg/client-go/rest"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// 创建一个Rest Client
// 验证xxx动词
// 与API Server通信，并执行基础操作

func main() {
	//启动模拟 HTTP 服务器
	go func() {
		server := createMockAPIServer()
		fmt.Println("Starting mock API server on :10000...")
		if err := server.ListenAndServe(); err != nil {
			fmt.Printf("Failed to start mock API server: %v\n", err)
		}
	}()
	time.Sleep(1 * time.Second) // 等待服务器启动

	scheme := runtime.NewScheme()
	apis.AddToScheme(scheme)
	//fmt.Println(scheme)
	//参数配置
	// TODO: 填写参数
	//部分参数之后可以在core_client等 编写setConfigDefaults函数进行填充

	c := &rest.Config{
		Host:    "http://localhost:10000",
		APIPath: "/apis/resources/v1",
		ContentConfig: rest.ContentConfig{
			AcceptContentTypes: "application/json; charset=UTF-8", //text/plain; charset=UTF-8
			ContentType:        "application/json; charset=UTF-8", //application/json; charset=UTF-8
			GroupVersion: &schema.GroupVersion{
				Group:   "hit.edu",
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
	// 默认访问的Namespace是""

	tasksClient := clientSet.Core().Tasks("")

	task1 := &apis.Task{
		ObjectMeta: metav1.ObjectMeta{
			Name: "demo-tasks",
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Task",
			APIVersion: "resources/v1",
		},
	}
	patchTask, err := json.Marshal(map[string]interface{}{
		"Spec": map[string]interface{}{
			"Name": "demo-task",
		},
	})

	// Create一个Task
	fmt.Println("creating")
	results, err := tasksClient.Create(context.TODO(), task1, metav1.CreateOptions{})

	if err != nil {
		panic(err)
	}
	fmt.Println("Created task ", results)

	prompt()

	//Patch 一个Task
	fmt.Println("patching")
	patchResult, err := tasksClient.Patch(context.TODO(), "demo-tasks", types.StrategicMergePatchType, []byte(patchTask), metav1.PatchOptions{})
	fmt.Println("patchResult: ", patchResult)
	fmt.Println("patch Done")

	prompt()

	//Update一个Task

	fmt.Println("updating")
	// 部分更改一个参数
	// 先Get一个Task ,更改Task的参数, UpdateTask

	result, getErr := tasksClient.Get(context.TODO(), "demo-tasks", metav1.GetOptions{})
	if getErr != nil {
		panic(fmt.Errorf("Failed to get : %v", getErr))
	}

	fmt.Println("get result", result)
	fmt.Println("修改前的result.Spec.TaskName：", result.Spec.Name)

	result.Spec.Name = "updatedTaskName"
	_, updateErr := tasksClient.Update(context.TODO(), result, metav1.UpdateOptions{})
	if updateErr != nil {
		panic(fmt.Errorf("Update failed: %v", updateErr))
	}

	fmt.Println("修改后的result.Spec.TaskName：", result.Spec.Name)
	fmt.Println("Updated task...")
	prompt()

	// List所有Node
	fmt.Println("listing")
	lstOpts := metav1.ListOptions{
		FieldSelector: "ObjectMeta.Name=demo-tasks",
	}
	list, err := tasksClient.List(context.TODO(), lstOpts)
	if err != nil {
		panic(err)
	}
	for _, d := range list.Items {
		fmt.Println(d)
	}

	fmt.Println("listing done")
	prompt()

	// Delete一个Task
	fmt.Println("deleting")
	err = tasksClient.Delete(context.TODO(), "demo-tasks", metav1.DeleteOptions{})
	if err != nil {
		panic(err)
	}
	fmt.Println("Deleted task...")
	prompt()

	// Delete 之后再次 List所有Node
	fmt.Println("listing")
	lstOpts = metav1.ListOptions{
		FieldSelector: "ObjectMeta.Name=demo-nodes",
	}
	list, err = tasksClient.List(context.TODO(), lstOpts)
	if err != nil {
		panic(err)
	}
	for _, d := range list.Items {
		fmt.Println(d)
	}

	fmt.Println("listing done")
}

func int32Ptr(i int) {
	panic("unimplemented")
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
	fmt.Println()
}

var (
	mu sync.Mutex // 用于保护 watchChans 的并发访问
)

func createMockAPIServer() *http.Server {
	mux := http.NewServeMux()

	// 模拟存储节点的内存数据库
	tasks := make(map[string]apis.Task)

	// 处理task的集合操作（POST 创建,List 和 Watch）
	mux.HandleFunc("/apis/resources/v1/tasks", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		iswatch := query.Get("watch")
		fieldSelector := query.Get("fieldSelector")
		switch r.Method {
		case http.MethodGet:
			if iswatch == "true" { // 判断是否是Watch请求
				// 解析 fieldSelector 并筛选节点
				filteredTasks := make([]apis.Task, 0)
				for _, task := range tasks {
					if taskMatchesFieldSelector(task, fieldSelector) {
						filteredTasks = append(filteredTasks, task)
					}
				}
				watchChan := make(chan watch.Event)
				mu.Lock()
				watchChans = append(watchChans, watchChan)
				mu.Unlock()

				// 设置响应头以支持流式传输
				w.Header().Set("Content-Type", "application/json; charset=UTF-8")
				w.WriteHeader(http.StatusOK)

				// 持续监听通道，发送事件
				encoder := json.NewEncoder(w)
				for event := range watchChan {
					if err := encoder.Encode(event); err != nil {
						fmt.Printf("Error encoding watch event: %v\n", err)
						return
					}
					w.(http.Flusher).Flush()           // 确保事件及时推送到客户端
					time.Sleep(500 * time.Millisecond) // 模拟一定的延迟
				}
				return
			} else { //否则为List请求
				// 解析 fieldSelector 并筛选节点
				taskList := apis.TaskList{}
				filteredTasks := make([]apis.Task, 0)
				for _, task := range tasks {
					if taskMatchesFieldSelector(task, fieldSelector) {
						filteredTasks = append(filteredTasks, task)
					}
				}
				taskList.Items = filteredTasks
				// 返回筛选后的节点列表
				w.Header().Set("Content-Type", "application/json; charset=UTF-8")
				w.WriteHeader(http.StatusOK)
				if err := json.NewEncoder(w).Encode(taskList); err != nil {
					http.Error(w, "Error encoding response", http.StatusInternalServerError)
				}
			}
			// 处理 GET 请求
		case http.MethodPost:
			// 检查 Content-Type 是否是 application/json; charset=UTF-8
			if r.Header.Get("Content-Type") != "application/json; charset=UTF-8" {
				http.Error(w, "Unsupported Content-Type", http.StatusUnsupportedMediaType)
				return
			}

			// 解析请求体中的 Task 数据
			newTask := &apis.Task{}
			if err := json.NewDecoder(r.Body).Decode(&newTask); err != nil {
				http.Error(w, "Invalid request body", http.StatusBadRequest)
				return
			}

			// 模拟存储节点
			tasks[newTask.ObjectMeta.Name] = *newTask

			// 推送 Watch 事件
			fmt.Println("newTask:", newTask)
			event := watch.Event{
				Type:   "ADDED",
				Object: newTask,
			}
			notifyWatchers(event)

			// 返回创建成功的响应
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.WriteHeader(http.StatusCreated) // 状态码 201 Created
			if err := json.NewEncoder(w).Encode(newTask); err != nil {
				http.Error(w, "Error encoding response", http.StatusInternalServerError)
			}
		default:
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	})
	// 处理task的单个操作（单个的GET 查询和 PUT 更新）
	mux.HandleFunc("/apis/resources/v1/tasks/demo-tasks", func(w http.ResponseWriter, r *http.Request) {
		taskName := "demo-tasks" // 固定为 demo-tasks

		switch r.Method {
		case http.MethodGet: // GET 查询
			// 查询内存数据库中的节点
			task, exists := tasks[taskName]
			if !exists {
				http.Error(w, "Task not found", http.StatusNotFound)
				return
			}

			// 返回节点信息
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.WriteHeader(http.StatusOK) // 状态码 200 OK
			if err := json.NewEncoder(w).Encode(task); err != nil {
				http.Error(w, "Error encoding response", http.StatusInternalServerError)
			}

		case http.MethodPut: // PUT 更新
			// 检查 Content-Type 是否是 application/json; charset=UTF-8
			if r.Header.Get("Content-Type") != "application/json; charset=UTF-8" {
				http.Error(w, "Unsupported Content-Type", http.StatusUnsupportedMediaType)
				return
			}

			// 解析请求体中的 Task 数据
			updatedTask := &apis.Task{}
			if err := json.NewDecoder(r.Body).Decode(&updatedTask); err != nil {
				http.Error(w, "Invalid request body", http.StatusBadRequest)
				return
			}

			// 检查节点是否存在
			_, exists := tasks[taskName]
			if !exists {
				http.Error(w, "Task not found", http.StatusNotFound)
				return
			}

			// 更新节点信息
			tasks[taskName] = *updatedTask

			// 推送 Watch 事件：MODIFIED
			event := watch.Event{
				Type:   "MODIFIED",
				Object: updatedTask,
			}
			notifyWatchers(event)

			// 返回更新成功的响应
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.WriteHeader(http.StatusOK) // 状态码 200 OK
			if err := json.NewEncoder(w).Encode(updatedTask); err != nil {
				http.Error(w, "Error encoding response", http.StatusInternalServerError)
			}
		case http.MethodPatch: // PATCH 部分更新
			// 检查 Content-Type 是否是 application/json; charset=UTF-8 或 application/merge-patch+json
			contentType := r.Header.Get("Content-Type")
			if contentType != "application/json; charset=UTF-8" && contentType != "application/strategic-merge-patch+json" {
				http.Error(w, "Unsupported Content-Type", http.StatusUnsupportedMediaType)
				return
			}

			// 查询内存数据库中的节点
			task, exists := tasks[taskName]
			if !exists {
				http.Error(w, "Task not found", http.StatusNotFound)
				return
			}

			// 解析 Patch 数据
			patchData, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "Error reading request body", http.StatusBadRequest)
				return
			}

			// 应用 Patch 到现有的节点
			updatedTask := task
			if err := json.Unmarshal(patchData, &updatedTask); err != nil {
				http.Error(w, "Invalid Patch format", http.StatusBadRequest)
				return
			}

			// 更新内存中的节点
			tasks[taskName] = updatedTask

			// 推送 Watch 事件：MODIFIED
			event := watch.Event{
				Type:   "MODIFIED",
				Object: &updatedTask,
			}
			notifyWatchers(event)

			// 返回更新成功的响应
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.WriteHeader(http.StatusOK) // 状态码 200 OK
			if err := json.NewEncoder(w).Encode(updatedTask); err != nil {
				http.Error(w, "Error encoding response", http.StatusInternalServerError)
			}
		case http.MethodDelete: // 删除节点
			deletetask, exists := tasks[taskName]
			if !exists {
				http.Error(w, "Task not found", http.StatusNotFound)
				return
			}
			deletedTask := &deletetask
			// 删除节点
			delete(tasks, taskName)

			// 推送 Watch 事件：DELETED
			event := watch.Event{
				Type:   "DELETED",
				Object: deletedTask,
			}
			notifyWatchers(event)

			// 返回删除成功的响应
			w.WriteHeader(http.StatusNoContent) // 状态码 204 No Content

		default:
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}

	})

	// 创建一个 HTTP 服务器
	server := &http.Server{
		Addr:    ":10000", // 监听端口 1000010000
		Handler: mux,
	}

	return server
}

func taskMatchesFieldSelector(task apis.Task, fieldSelector string) bool {
	if fieldSelector == "" {
		return true // 如果没有指定 fieldSelector，匹配所有节点
	}

	// 示例：支持解析 "ObjectMeta.Name=demo-tasks" 的 fieldSelector
	parts := strings.Split(fieldSelector, "=")
	if len(parts) != 2 {
		return false
	}

	key, value := parts[0], parts[1]
	switch key {
	case "ObjectMeta.Name":
		return task.ObjectMeta.Name == value
	// 可扩展其他字段匹配
	default:
		return false
	}
}

var (
	tasks      = make(map[string]apis.Task)  // 模拟存储节点的内存数据库
	watchChans = make([]chan watch.Event, 0) // 维护所有watch监听的通道
)

// 推送事件
func notifyWatchers(event watch.Event) {
	for _, ch := range watchChans {
		fmt.Println("推送的event:", event)
		ch <- event
	}
}

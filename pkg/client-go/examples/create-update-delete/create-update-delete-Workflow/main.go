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
	"hit.edu/framework/pkg/component-base/logs"
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
	logs.Init("create-update-delete-Workflow")
	scheme := runtime.NewScheme()
	//apis.AddToScheme(scheme)
	////fmt.Println(scheme)
	////启动模拟 HTTP 服务器
	//go func() {
	//	server := createMockAPIServer()
	//	fmt.Println("Starting mock API server on :10000...")
	//	if err := server.ListenAndServe(); err != nil {
	//		fmt.Printf("Failed to start mock API server: %v\n", err)
	//	}
	//}()
	//time.Sleep(1 * time.Second) // 等待服务器启动

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
	// 这里以访问资源Workflow为例，
	// 获取访问Workflow的客户端
	// 默认访问的Namespace是 ""

	workflowsClient := clientSet.Core().Workflows("")

	workflow1 := &apis.Workflow{
		ObjectMeta: metav1.ObjectMeta{
			Name: "demo-workflows",
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Workflow",
			APIVersion: "resources/v1",
		},
	}
	workflow2 := &apis.Workflow{
		ObjectMeta: metav1.ObjectMeta{
			Name: "demo-workflow2",
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Workflow",
			APIVersion: "resources/v1",
		},
		Spec: apis.WorkflowSpec{
			Name: "demo-workflow",
		},
	}
	workflow3 := &apis.Workflow{
		ObjectMeta: metav1.ObjectMeta{
			Name: "demo-workflow3",
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Workflow",
			APIVersion: "resources/v1",
		},
		Spec: apis.WorkflowSpec{
			Name: "demo-workflow",
		},
	}
	patchWorkflow, err := json.Marshal(map[string]interface{}{
		"Spec": map[string]interface{}{
			"Name": "demo-workflow",
		},
	})

	// Create一个Workflow
	fmt.Println("creating")
	results, err := workflowsClient.Create(context.TODO(), workflow1, metav1.CreateOptions{})

	if err != nil {
		panic(err)
	}
	_, _ = workflowsClient.Create(context.TODO(), workflow2, metav1.CreateOptions{})
	_, _ = workflowsClient.Create(context.TODO(), workflow3, metav1.CreateOptions{})
	fmt.Println("Created workflow ", results)

	prompt()

	//Patch 一个Workflow
	fmt.Println("patching")
	patchResult, err := workflowsClient.Patch(context.TODO(), "demo-workflows", types.StrategicMergePatchType, []byte(patchWorkflow), metav1.PatchOptions{})
	fmt.Println("patchResult: ", patchResult)
	fmt.Println("patch Done")

	//Update一个Workflow

	fmt.Println("updating")
	// 部分更改一个参数
	// 先Get一个Workflow ,更改Workflow的参数, UpdateWorkflow

	result, getErr := workflowsClient.Get(context.TODO(), "demo-workflows", metav1.GetOptions{})
	if getErr != nil {
		panic(fmt.Errorf("Failed to get : %v", getErr))
	}

	fmt.Println("get result", result)
	fmt.Println("修改前的result.Spec.Name：", result.Spec.Name)

	result.Spec.Name = "updatedName"
	_, updateErr := workflowsClient.Update(context.TODO(), result, metav1.UpdateOptions{})
	if updateErr != nil {
		panic(fmt.Errorf("Update failed: %v", updateErr))
	}

	fmt.Println("修改后的result.Spec.Name：", result.Spec.Name)
	fmt.Println("Updated workflow...")
	prompt()

	// List所有Workflow
	fmt.Println("listing")
	lstOpts := metav1.ListOptions{}
	list, err := workflowsClient.List(context.TODO(), lstOpts)
	if err != nil {
		panic(err)
	}
	for _, d := range list.Items {
		fmt.Println(d)
	}

	fmt.Println("listing done")
	prompt()

	// Delete一个Workflow
	fmt.Println("deleting")
	err = workflowsClient.Delete(context.TODO(), "demo-workflows", metav1.DeleteOptions{})
	if err != nil {
		panic(err)
	}
	fmt.Println("Deleted workflow...")
	prompt()

	// Delete 之后再次 List所有Workflow
	fmt.Println("listing")
	lstOpts = metav1.ListOptions{}
	list, err = workflowsClient.List(context.TODO(), lstOpts)
	if err != nil {
		panic(err)
	}
	for _, d := range list.Items {
		fmt.Println(d)
	}
	fmt.Println("listing done")

	//DeleteCollection 删除所有Spec.WorkflowName=demo-workflow的Workflow
	fmt.Println("deleting collection")
	lstOpts = metav1.ListOptions{
		FieldSelector: "Spec.Name=demo-workflow",
	}
	err = workflowsClient.DeleteCollection(context.TODO(), metav1.DeleteOptions{}, lstOpts)
	if err != nil {
		panic(err)
	}
	fmt.Println("Deleted collection...")

	// DeleteCollection 之后再次 List所有Workflow
	fmt.Println("listing")
	lstOpts = metav1.ListOptions{}
	list, err = workflowsClient.List(context.TODO(), lstOpts)
	if err != nil {
		panic(err)
	}
	for _, d := range list.Items {
		fmt.Println(d)
	}
	fmt.Println("listing done")
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
	workflows := make(map[string]apis.Workflow)

	// 处理workflow的集合操作（POST 创建,List 和 Watch）
	mux.HandleFunc("/apis/resources/v1/workflows", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		iswatch := query.Get("watch")
		fieldSelector := query.Get("fieldSelector")
		switch r.Method {
		case http.MethodGet:
			fmt.Println("handle get")
			if iswatch == "true" { // 判断是否是Watch请求
				// 解析 fieldSelector 并筛选节点
				filteredWorkflows := make([]apis.Workflow, 0)
				for _, workflow := range workflows {
					if workflowMatchesFieldSelector(workflow, fieldSelector) {
						filteredWorkflows = append(filteredWorkflows, workflow)
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
				workflowList := apis.WorkflowList{}
				filteredWorkflows := make([]apis.Workflow, 0)
				for _, workflow := range workflows {
					if workflowMatchesFieldSelector(workflow, fieldSelector) {
						filteredWorkflows = append(filteredWorkflows, workflow)
					}
				}
				workflowList.Items = filteredWorkflows
				// 返回筛选后的节点列表
				w.Header().Set("Content-Type", "application/json; charset=UTF-8")
				w.WriteHeader(http.StatusOK)
				if err := json.NewEncoder(w).Encode(workflowList); err != nil {
					http.Error(w, "Error encoding response", http.StatusInternalServerError)
				}
			}
		case http.MethodPost:
			// 检查 Content-Type 是否是 application/json; charset=UTF-8
			if r.Header.Get("Content-Type") != "application/json; charset=UTF-8" {
				fmt.Println("content-type is not application/json; charset=UTF-8::::", r.Header.Get("Content-Type"))
				http.Error(w, "Unsupported Content-Type", http.StatusUnsupportedMediaType)
				return
			}

			// 解析请求体中的 Workflow 数据
			newWorkflow := &apis.Workflow{}
			if err := json.NewDecoder(r.Body).Decode(&newWorkflow); err != nil {
				http.Error(w, "Invalid request body", http.StatusBadRequest)
				return
			}

			// 模拟存储节点
			workflows[newWorkflow.ObjectMeta.Name] = *newWorkflow

			// 推送 Watch 事件
			fmt.Println("newWorkflow:", newWorkflow)
			event := watch.Event{
				Type:   "ADDED",
				Object: newWorkflow,
			}
			notifyWatchers(event)

			// 返回创建成功的响应
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.WriteHeader(http.StatusCreated) // 状态码 201 Created
			if err := json.NewEncoder(w).Encode(newWorkflow); err != nil {
				http.Error(w, "Error encoding response", http.StatusInternalServerError)
			}
		case http.MethodDelete: // 删除节点
			fieldSelector := r.URL.Query().Get("fieldSelector")
			// 用于存储被删除的节点
			var deletedWorkflows []apis.Workflow
			// 遍历所有节点，筛选符合条件的节点
			for workflowName, workflow := range workflows {
				if workflowMatchesFieldSelector(workflow, fieldSelector) {
					deletedWorkflows = append(deletedWorkflows, workflow)
					delete(workflows, workflowName) // 从内存中删除节点
				}
			}
			// 推送 Watch 事件：DELETED
			for _, deletedWorkflow := range deletedWorkflows {
				event := watch.Event{
					Type:   "DELETED",
					Object: &deletedWorkflow,
				}
				notifyWatchers(event)
			}

			// 返回删除成功的响应
			w.WriteHeader(http.StatusNoContent) // 状态码 204 No Content
		default:
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	})
	// 处理workflow的单个操作（单个的GET 查询和 PUT 更新）
	mux.HandleFunc("/apis/resources/v1/workflows/demo-workflows", func(w http.ResponseWriter, r *http.Request) {
		workflowName := "demo-workflows" // 固定为 demo-workflows

		switch r.Method {
		case http.MethodGet: // GET 查询
			// 查询内存数据库中的节点
			workflow, exists := workflows[workflowName]
			if !exists {
				http.Error(w, "Workflow not found", http.StatusNotFound)
				return
			}

			// 返回节点信息
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.WriteHeader(http.StatusOK) // 状态码 200 OK
			if err := json.NewEncoder(w).Encode(workflow); err != nil {
				http.Error(w, "Error encoding response", http.StatusInternalServerError)
			}

		case http.MethodPut: // PUT 更新
			// 检查 Content-Type 是否是 application/json; charset=UTF-8
			if r.Header.Get("Content-Type") != "application/json; charset=UTF-8" {
				http.Error(w, "Unsupported Content-Type", http.StatusUnsupportedMediaType)
				return
			}

			// 解析请求体中的 Workflow 数据
			updatedWorkflow := &apis.Workflow{}
			if err := json.NewDecoder(r.Body).Decode(&updatedWorkflow); err != nil {
				http.Error(w, "Invalid request body", http.StatusBadRequest)
				return
			}

			// 检查节点是否存在
			_, exists := workflows[workflowName]
			if !exists {
				http.Error(w, "Workflow not found", http.StatusNotFound)
				return
			}

			// 更新节点信息
			workflows[workflowName] = *updatedWorkflow

			// 推送 Watch 事件：MODIFIED
			event := watch.Event{
				Type:   "MODIFIED",
				Object: updatedWorkflow,
			}
			notifyWatchers(event)

			// 返回更新成功的响应
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.WriteHeader(http.StatusOK) // 状态码 200 OK
			if err := json.NewEncoder(w).Encode(updatedWorkflow); err != nil {
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
			workflow, exists := workflows[workflowName]
			if !exists {
				http.Error(w, "Workflow not found", http.StatusNotFound)
				return
			}

			// 解析 Patch 数据
			patchData, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "Error reading request body", http.StatusBadRequest)
				return
			}

			// 应用 Patch 到现有的节点
			updatedWorkflow := workflow
			if err := json.Unmarshal(patchData, &updatedWorkflow); err != nil {
				http.Error(w, "Invalid Patch format", http.StatusBadRequest)
				return
			}

			// 更新内存中的节点
			workflows[workflowName] = updatedWorkflow

			// 推送 Watch 事件：MODIFIED
			event := watch.Event{
				Type:   "MODIFIED",
				Object: &updatedWorkflow,
			}
			notifyWatchers(event)

			// 返回更新成功的响应
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.WriteHeader(http.StatusOK) // 状态码 200 OK
			if err := json.NewEncoder(w).Encode(updatedWorkflow); err != nil {
				http.Error(w, "Error encoding response", http.StatusInternalServerError)
			}
		case http.MethodDelete: // 删除节点
			deleteworkflow, exists := workflows[workflowName]
			if !exists {
				http.Error(w, "Workflow not found", http.StatusNotFound)
				return
			}
			deletedWorkflow := &deleteworkflow
			// 删除节点
			delete(workflows, workflowName)

			// 推送 Watch 事件：DELETED
			event := watch.Event{
				Type:   "DELETED",
				Object: deletedWorkflow,
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
		Addr:    ":10000", // 监听端口 10000
		Handler: mux,
	}

	return server
}

func workflowMatchesFieldSelector(workflow apis.Workflow, fieldSelector string) bool {
	if fieldSelector == "" {
		return true // 如果没有指定 fieldSelector，匹配所有节点
	}

	// 示例：支持解析 "ObjectMeta.Name=demo-workflows" 的 fieldSelector
	parts := strings.Split(fieldSelector, "=")
	if len(parts) != 2 {
		return false
	}

	key, value := parts[0], parts[1]
	switch key {
	case "Spec.Name":
		return workflow.Spec.Name == value
	// 可扩展其他字段匹配
	default:
		return false
	}
}

var (
	workflows  = make(map[string]apis.Workflow) // 模拟存储节点的内存数据库
	watchChans = make([]chan watch.Event, 0)    // 维护所有watch监听的通道
)

// 推送事件
func notifyWatchers(event watch.Event) {
	for _, ch := range watchChans {
		fmt.Println("推送的event:", event)
		ch <- event
	}
}

package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"hit.edu/framework/pkg/apimachinery/runtime"
	"hit.edu/framework/pkg/apimachinery/runtime/schema"
	"hit.edu/framework/pkg/apimachinery/runtime/serializer"
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

// 尚未与api server测试
func main() {
	logs.Init("main")
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
	fmt.Println(scheme)
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
	// 这里以访问资源ResourceRequirement为例，
	// 获取访问ResourceRequirement的客户端
	// 默认访问的Namespace是 ""

	resourceRequirementsClient := clientSet.Core().ResourceRequirements("test")

	resourceRequirement1 := &apis.ResourceRequirement{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "demo-resourceRequirements",
			Namespace: "test",
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "ResourceRequirement",
			APIVersion: "resources/v1",
		},
	}

	// Create一个ResourceRequirement
	fmt.Println("creating")
	results, err := resourceRequirementsClient.Create(context.TODO(), resourceRequirement1, metav1.CreateOptions{})

	if err != nil {
		panic(err)
	}
	fmt.Println("Created resourceRequirement ", results)

	prompt()

	// todo:update 和 patch

	// List所有ResourceRequirement
	fmt.Println("listing")
	lstOpts := metav1.ListOptions{
		FieldSelector: "ObjectMeta.Name=demo-resourceRequirements",
	}
	list, err := resourceRequirementsClient.List(context.TODO(), lstOpts)
	if err != nil {
		panic(err)
	}
	for _, d := range list.Items {
		fmt.Println(d)
	}

	fmt.Println("listing done")
	prompt()

	// Delete一个ResourceRequirement
	fmt.Println("deleting")
	err = resourceRequirementsClient.Delete(context.TODO(), "demo-resourceRequirements", metav1.DeleteOptions{})
	if err != nil {
		panic(err)
	}
	fmt.Println("Deleted resourceRequirement...")
	prompt()

	// Delete 之后再次 List所有ResourceRequirement
	fmt.Println("listing")
	lstOpts = metav1.ListOptions{
		FieldSelector: "ObjectMeta.Name=demo-resourceRequirements",
	}
	list, err = resourceRequirementsClient.List(context.TODO(), lstOpts)
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
	resourceRequirements := make(map[string]apis.ResourceRequirement)

	// 处理resourceRequirement的集合操作（POST 创建,List 和 Watch）
	mux.HandleFunc("/apis/resources/v1/namespaces/test/resourcerequirements", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Received request:", r.Method, r.URL.Path)
		query := r.URL.Query()
		iswatch := query.Get("watch")
		fieldSelector := query.Get("fieldSelector")
		switch r.Method {
		case http.MethodGet:
			if iswatch == "true" { // 判断是否是Watch请求
				// 解析 fieldSelector 并筛选节点
				filteredResourceRequirements := make([]apis.ResourceRequirement, 0)
				for _, resourceRequirement := range resourceRequirements {
					if resourceRequirementMatchesFieldSelector(resourceRequirement, fieldSelector) {
						filteredResourceRequirements = append(filteredResourceRequirements, resourceRequirement)
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
				resourceRequirementsList := apis.ResourceRequirementList{}
				filteredResourceRequirements := make([]apis.ResourceRequirement, 0)
				for _, resourceRequirement := range resourceRequirements {
					if resourceRequirementMatchesFieldSelector(resourceRequirement, fieldSelector) {
						filteredResourceRequirements = append(filteredResourceRequirements, resourceRequirement)
					}
				}
				resourceRequirementsList.Items = filteredResourceRequirements
				// 返回筛选后的节点列表
				w.Header().Set("Content-Type", "application/json; charset=UTF-8")
				w.WriteHeader(http.StatusOK)
				if err := json.NewEncoder(w).Encode(resourceRequirementsList); err != nil {
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
			// 解析请求体中的 ResourceRequirement 数据
			newResourceRequirement := &apis.ResourceRequirement{}
			if err := json.NewDecoder(r.Body).Decode(&newResourceRequirement); err != nil {
				http.Error(w, "Invalid request body", http.StatusBadRequest)
				return
			}

			// 模拟存储节点
			resourceRequirements[newResourceRequirement.ObjectMeta.Name] = *newResourceRequirement

			// 推送 Watch 事件
			fmt.Println("newResourceRequirement:", newResourceRequirement)
			event := watch.Event{
				Type:   "ADDED",
				Object: newResourceRequirement,
			}
			notifyWatchers(event)

			// 返回创建成功的响应
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.WriteHeader(http.StatusCreated) // 状态码 201 Created
			if err := json.NewEncoder(w).Encode(newResourceRequirement); err != nil {
				http.Error(w, "Error encoding response", http.StatusInternalServerError)
			}
		default:
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	})
	// 处理resourceRequirement的单个操作（单个的GET 查询和 PUT、Patch 更新）
	mux.HandleFunc("/apis/resources/v1/namespaces/test/resourcerequirements/demo-resourceRequirements", func(w http.ResponseWriter, r *http.Request) {
		resourceRequirementName := "demo-resourceRequirements" // 固定为 demo-resourceRequirements

		switch r.Method {
		case http.MethodGet: // GET 查询
			// 查询内存数据库中的节点
			resourceRequirement, exists := resourceRequirements[resourceRequirementName]
			if !exists {
				http.Error(w, "ResourceRequirement not found", http.StatusNotFound)
				return
			}

			// 返回节点信息
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.WriteHeader(http.StatusOK) // 状态码 200 OK
			if err := json.NewEncoder(w).Encode(resourceRequirement); err != nil {
				http.Error(w, "Error encoding response", http.StatusInternalServerError)
			}

		case http.MethodPut: // PUT 更新
			// 检查 Content-Type 是否是 application/json; charset=UTF-8
			if r.Header.Get("Content-Type") != "application/json; charset=UTF-8" {
				http.Error(w, "Unsupported Content-Type", http.StatusUnsupportedMediaType)
				return
			}

			// 解析请求体中的 ResourceRequirement 数据
			updatedResourceRequirement := &apis.ResourceRequirement{}
			if err := json.NewDecoder(r.Body).Decode(&updatedResourceRequirement); err != nil {
				http.Error(w, "Invalid request body", http.StatusBadRequest)
				return
			}

			// 检查节点是否存在
			_, exists := resourceRequirements[resourceRequirementName]
			if !exists {
				http.Error(w, "ResourceRequirement not found", http.StatusNotFound)
				return
			}

			// 更新节点信息
			resourceRequirements[resourceRequirementName] = *updatedResourceRequirement

			// 推送 Watch 事件：MODIFIED
			event := watch.Event{
				Type:   "MODIFIED",
				Object: updatedResourceRequirement,
			}
			notifyWatchers(event)

			// 返回更新成功的响应
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.WriteHeader(http.StatusOK) // 状态码 200 OK
			if err := json.NewEncoder(w).Encode(updatedResourceRequirement); err != nil {
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
			resourceRequirement, exists := resourceRequirements[resourceRequirementName]
			if !exists {
				http.Error(w, "ResourceRequirement not found", http.StatusNotFound)
				return
			}

			// 解析 Patch 数据
			patchData, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "Error reading request body", http.StatusBadRequest)
				return
			}

			// 应用 Patch 到现有的节点
			updatedResourceRequirement := resourceRequirement
			if err := json.Unmarshal(patchData, &updatedResourceRequirement); err != nil {
				http.Error(w, "Invalid Patch format", http.StatusBadRequest)
				return
			}

			// 更新内存中的节点
			resourceRequirements[resourceRequirementName] = updatedResourceRequirement

			// 推送 Watch 事件：MODIFIED
			event := watch.Event{
				Type:   "MODIFIED",
				Object: &updatedResourceRequirement,
			}
			notifyWatchers(event)

			// 返回更新成功的响应
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.WriteHeader(http.StatusOK) // 状态码 200 OK
			if err := json.NewEncoder(w).Encode(updatedResourceRequirement); err != nil {
				http.Error(w, "Error encoding response", http.StatusInternalServerError)
			}
		case http.MethodDelete: // 删除节点
			deleteresourceRequirement, exists := resourceRequirements[resourceRequirementName]
			if !exists {
				http.Error(w, "ResourceRequirement not found", http.StatusNotFound)
				return
			}
			deletedResourceRequirement := &deleteresourceRequirement
			// 删除节点
			delete(resourceRequirements, resourceRequirementName)

			// 推送 Watch 事件：DELETED
			event := watch.Event{
				Type:   "DELETED",
				Object: deletedResourceRequirement,
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

func resourceRequirementMatchesFieldSelector(resourceRequirement apis.ResourceRequirement, fieldSelector string) bool {
	if fieldSelector == "" {
		return true // 如果没有指定 fieldSelector，匹配所有节点
	}

	// 示例：支持解析 "ObjectMeta.Name=demo-resourceRequirements" 的 fieldSelector
	parts := strings.Split(fieldSelector, "=")
	if len(parts) != 2 {
		return false
	}

	key, value := parts[0], parts[1]
	switch key {
	case "ObjectMeta.Name":
		return resourceRequirement.ObjectMeta.Name == value
	// 可扩展其他字段匹配
	default:
		return false
	}
}

var (
	resourceRequirements = make(map[string]apis.ResourceRequirement) // 模拟存储节点的内存数据库
	watchChans           = make([]chan watch.Event, 0)               // 维护所有watch监听的通道
)

// 推送事件
func notifyWatchers(event watch.Event) {
	for _, ch := range watchChans {
		fmt.Println("推送的event:", event)
		ch <- event
	}
}

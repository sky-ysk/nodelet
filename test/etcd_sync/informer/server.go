package informer

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	"hit.edu/framework/pkg/component-base/logs"
)

func printHTTPRequest(r *http.Request) {
	// 打印请求方法
	fmt.Printf("Request Method: %s\n", r.Method)
	// 打印请求的完整URL
	fmt.Printf("Request URL: %s\n", r.URL.String())
	// 打印请求头
	//fmt.Printf("Request Headers: %v\n", r.Header)
	// 打印查询参数
	//fmt.Printf("Query Parameters: %v\n", r.URL.Query())

	// 如果请求体存在，读取并打印请求体
	if r.Body != nil {
		defer r.Body.Close()
		_, err := ioutil.ReadAll(r.Body)
		if err != nil {
			//fmt.Printf("Failed to read request body: %v", err)
		} else {
			//fmt.Printf("Request Body: %s", string(bodyBytes))
		}
	}
}

// TODO： 消除label
func handler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	resource := vars["resource"]
	//fmt.Printf("Received request for resource: %s\n", resource)
	printHTTPRequest(r)
	//TODO: 转发给apiserver
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := fmt.Sprintf(`{"message": "Received request for resource: %s"}`, resource)
	w.Write([]byte(response))
	_, err := w.Write([]byte(response))
	if err != nil {
		logs.Errorf("Failed to write response: %v", err)
	}
}

func healthzHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "OK\n")
}

func CreateWebHandler() {
	//http.HandleFunc("/apis/resources/v1/{resource}/{name}", handler)
	http.HandleFunc("/apis/resources/v1/{resource}/{namespace}/{name}", handler)
	http.HandleFunc("/apis/resources/v1/{resource}", handler)
	http.HandleFunc("/healthz", healthzHandler)

	port := "127.0.0.1:14399" // 监听 14399 端口
	fmt.Println("Starting server on port", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		fmt.Println("Failed to start server:", err)
	}
}

package informer

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"hit.edu/framework/pkg/client-go/rest"
	"hit.edu/framework/pkg/component-base/logs"
)

func printHTTPRequest(r *http.Request) {
	// 打印请求方法
	fmt.Printf("Request Method: %s\n", r.Method)
	// 打印请求的完整URL
	fmt.Printf("Request URL: %s\n", r.URL.String())
	// 打印请求头
	fmt.Printf("Request Headers: %v\n", r.Header)
	// 打印查询参数
	fmt.Printf("Query Parameters: %v\n", r.URL.Query())

	// 如果请求体存在，读取并打印请求体
	if r.Body != nil {
		defer r.Body.Close()
		_, err := ioutil.ReadAll(r.Body)
		if err != nil {
			fmt.Printf("Failed to read request body: %v", err)
		} else {
			//fmt.Printf("Request Body: %s", string(bodyBytes))
		}
	}
}

func healthzHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "OK\n")
}

func (c Sever_Config) CreateWebHandler(config *rest.Config, listen_ip string) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		requestURI := r.RequestURI
		segments := strings.Split(requestURI, "/")
		method := segments[4]

		switch method {
		case "POST":
			{
				logs.Infof("receive a POST")
				obj, err := c.PostHandler(r)
				if err != nil {
					logs.Errorf("Failed to create resource: %v", err)
					http.Error(w, fmt.Sprintf("Failed to create resource: %v", err), http.StatusBadRequest)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusCreated)

				err = json.NewEncoder(w).Encode(obj)
				if err != nil {
					logs.Errorf("Failed to serialize response: %v", err)
					http.Error(w, fmt.Sprintf("Failed to serialize response: %v", err), http.StatusInternalServerError)
				}
			}
		case "PUT":
			{
				obj, err := c.PutHandler(r)
				if err != nil {
					logs.Errorf("Failed to update resource: %v", err)
					http.Error(w, fmt.Sprintf("Failed to update resource: %v", err), http.StatusBadRequest)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusCreated)

				err = json.NewEncoder(w).Encode(obj)
				if err != nil {
					logs.Errorf("Failed to serialize response: %v", err)
					http.Error(w, fmt.Sprintf("Failed to serialize response: %v", err), http.StatusInternalServerError)
				}
			}
		case "DELETE":
			{
				err := c.DeleteHandler(r)
				if err != nil {
					logs.Errorf("Failed to delete resource: %v", err)
					http.Error(w, fmt.Sprintf("Failed to delete resource: %v", err), http.StatusBadRequest)
					return
				}
			}

		case "GET":
			{
				obj, err := c.GetHandler(r)
				if err != nil {
					logs.Errorf("Failed to get resource: %v", err)
					http.Error(w, fmt.Sprintf("Failed to get resource: %v", err), http.StatusBadRequest)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusCreated)

				err = json.NewEncoder(w).Encode(obj)
				if err != nil {
					logs.Errorf("Failed to serialize response: %v", err)
					http.Error(w, fmt.Sprintf("Failed to serialize response: %v", err), http.StatusInternalServerError)
				}
			}
		default:
			{
				logs.Errorf("Unsupported method: %s", method)
				http.Error(w, fmt.Sprintf("Unsupported method: %s", r.Method), http.StatusMethodNotAllowed)
			}

		}

	}
	//http.HandleFunc("/apis/resources/v1/{resource}/{name}", handler)
	http.HandleFunc("/apis/resources/v1/{method}/{resource}/{namespace}/{name}", handler)
	http.HandleFunc("/apis/resources/v1/{method}/{resource}", handler)
	http.HandleFunc("/healthz", healthzHandler)

	logs.Infof("Starting server on %s", listen_ip)
	if err := http.ListenAndServe(listen_ip, nil); err != nil {
		logs.Errorf("Failed to start server:%v", err)
	}
}

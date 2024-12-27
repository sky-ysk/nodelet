package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() { //linux
	// 日志打印每个请求信息
	fmt.Println("Starting server...")
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		clientIP := r.RemoteAddr
		host := r.Host
		log.Printf("Received request: %s %s from IP: %s, Host: %s\n", r.Method, r.URL.Path, clientIP, host)
		// 转发请求给实际的文件服务器
		http.StripPrefix("/files/", http.FileServer(http.Dir("/tmp/nodelet/hengtong"))).ServeHTTP(w, r)
	})

	// 启动 HTTP 文件服务器
	fmt.Println("Starting file server on :8080...")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}

//func main() { //Windows
//	// 日志打印每个请求信息
//	fmt.Println("Starting server...")
//	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
//		clientIP := r.RemoteAddr
//		host := r.Host
//		log.Printf("Received request: %s %s from IP: %s, Host: %s\n", r.Method, r.URL.Path, clientIP, host)
//		// 转发请求给实际的文件服务器
//		http.StripPrefix("/files/", http.FileServer(http.Dir("C:\\Users\\hzy1207\\Desktop\\heongtong_yolo\\heongtong_yolo"))).ServeHTTP(w, r)
//	})
//
//	// 启动 HTTP 文件服务器
//	fmt.Println("Starting file server on :8080...")
//	err := http.ListenAndServe(":8080", nil)
//	if err != nil {
//		log.Fatalf("Error starting server: %v", err)
//	}
//}

//// 日志中间件，用于打印 HTTP 请求信息
//func loggingMiddleware(next http.Handler) http.Handler {
//	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//		// 打印请求路径和远程地址
//		fmt.Printf("Received request: %s from %s\n", r.URL.Path, r.RemoteAddr)
//		// 继续处理请求
//		next.ServeHTTP(w, r)
//	})
//}

//func main() { //windows
//
//	fmt.Println("Starting server...")
//
//	// 创建文件服务器
//	fileServer := http.FileServer(http.Dir("C:\\Users\\hzy1207\\Desktop\\heongtong_yolo\\heongtong_yolo"))
//
//	// 使用 StripPrefix 去掉 /files/ 前缀
//	stripPrefixHandler := http.StripPrefix("/files/", fileServer)
//
//	// 包裹文件服务器处理器，添加日志中间件
//	http.Handle("/files/", loggingMiddleware(stripPrefixHandler))
//
//	// 启动 HTTP 服务器
//	err := http.ListenAndServe(":8080", nil)
//	if err != nil {
//		fmt.Println(err)
//	}
//}

//func main() { //Linux
//	fmt.Println("Starting server...")
//
//	// 使用 StripPrefix 去掉 /files/ 前缀
//	stripPrefixHandler := http.StripPrefix("/files/", http.FileServer(http.Dir("/tmp/nodelet/hengtong")))
//
//	// 包裹文件服务器处理器，添加日志中间件
//	http.Handle("/files/", loggingMiddleware(stripPrefixHandler))
//
//	// 启动 HTTP 服务器
//	err := http.ListenAndServe(":8080", nil) //0.0.0.0:8080
//	if err != nil {
//		fmt.Println(err)
//	}
//}

////版本2-linux
//

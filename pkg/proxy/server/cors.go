package server

import (
	"github.com/emicklei/go-restful/v3"
	"net/http"
)

func main() {
	// 创建容器并注册全局过滤器
	container := restful.NewContainer()
	container.Filter(globalCORS) // 全局跨域处理

	// 示例路由
	ws := new(restful.WebService)
	ws.Route(ws.GET("/hello").To(helloHandler))
	container.Add(ws)

	// 启动 HTTP 服务（非 HTTPS）
	http.ListenAndServe(":8080", container)
}

// 处理函数
func helloHandler(req *restful.Request, resp *restful.Response) {
	resp.WriteAsJson(map[string]string{"message": "Hello, CORS enabled!"})
}

// 全局 CORS 过滤器
func globalCORS(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
	// 设置 CORS 头
	resp.Header().Set("Access-Control-Allow-Origin", "*")
	resp.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	resp.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	resp.Header().Set("Access-Control-Expose-Headers", "Content-Length")
	resp.Header().Set("Access-Control-Allow-Credentials", "false")

	// 处理 OPTIONS 预检请求[5](@ref)
	if req.Request.Method == "OPTIONS" {
		resp.WriteHeader(http.StatusOK)
		chain.ProcessFilter(req, resp)
		return
	}

	// 继续处理其他请求
	chain.ProcessFilter(req, resp)
}

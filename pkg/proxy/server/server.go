package server

import (
	restfulspec "github.com/emicklei/go-restful-openapi/v2"
	restful "github.com/emicklei/go-restful/v3"
	"hit.edu/framework/pkg/client-go/clients"
	"hit.edu/framework/pkg/proxy/server/handlers"
	"hit.edu/framework/pkg/proxy/server/openapi"
	"net/http"
)

type Server struct {
	//Handlers
	handlers *handlers.Handlers

	container *restful.Container
}

func (s *Server) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	h := s.container
	h.ServeHTTP(writer, request)
}

// TODO: 安全认证

func NewServer(clientSet *clients.ClientSet) Server {
	hs := handlers.NewHandlers(clientSet)

	server := Server{
		handlers:  hs,
		container: restful.NewContainer(),
	}

	// 安装各类Handlers
	server.InstallDefaultHandlers()

	cors := restful.CrossOriginResourceSharing{
		ExposeHeaders:  []string{"*"},
		AllowedDomains: []string{"localhost", "wangwanu.com"},
		AllowedHeaders: []string{"Content-Type", "Accept"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "HEAD"},
		CookiesAllowed: false,
		Container:      server.container}
	server.container.Filter(cors.Filter)

	// Add container filter to respond to OPTIONS
	server.container.Filter(server.container.OPTIONSFilter)

	return server
}

func (s *Server) InstallDefaultHandlers() {
	// Workflow相关
	s.handlers.InstallWorkflowHandlers(s.container)
	// Task相关
	s.handlers.InstallTaskHandlers(s.container)
	// Group相关
	s.handlers.InstallGroupHandlers(s.container)

	// Resource相关
	// Node
	s.handlers.InstallNodeHandlers(s.container)

	// Logs相关

	// Openapi相关
	s.InstallOpenapiHandlers()

	// +Optional SwaggerUI
	s.InstallSwaggerUI()

}

func (s *Server) InstallOpenapiHandlers() {
	config := restfulspec.Config{
		WebServices:                   s.container.RegisteredWebServices(), // you control what services are visible
		APIPath:                       "/apidocs.json",
		PostBuildSwaggerObjectHandler: openapi.EnrichSwaggerObject}
	s.container.Add(restfulspec.NewOpenAPIService(config))
}

func (s *Server) InstallSwaggerUI() {
	s.container.Handle("/apidocs/", http.StripPrefix("/apidocs/", http.FileServer(http.Dir("/tmp/swagger-ui/dist"))))
}

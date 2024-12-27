package handler

import (
	"fmt"
	"hit.edu/framework/pkg/apimachinery/runtime"
	"hit.edu/framework/pkg/apiserver/endpoints/request"
	"k8s.io/apimachinery/pkg/api/errors"
	"net/http"
)

// ScopeNamer 处理从请求和对象访问名称
type ScopeNamer interface {
	// Name 返回来自请求的名称
	Name(req *http.Request) (name string, err error)
	// ObjectName 返回对象的名称，如果不不支持名称则返回错误
	ObjectName(obj runtime.Object) (name string, err error)
}

type ContextBasedNaming struct {
	Namer runtime.Namer
}

// ContextBasedNaming implements ScopeNamer
var _ ScopeNamer = ContextBasedNaming{}

func (n ContextBasedNaming) Name(req *http.Request) (name string, err error) {
	requestInfo, ok := request.RequestInfoFrom(req.Context())
	if !ok {
		return "", fmt.Errorf("missing requestInfo")
	}
	
	if len(requestInfo.Name) == 0 {
		return "", errEmptyName
	}
	return requestInfo.Name, nil
}

func (n ContextBasedNaming) ObjectName(obj runtime.Object) (name string, err error) {
	name, err = n.Namer.Name(obj)
	if err != nil {
		return "", err
	}
	if len(name) == 0 {
		return "", errEmptyName
	}
	return name, err
}

// errEmptyName is returned when API requests do not fill the name section of the path.
var errEmptyName = errors.NewBadRequest("name must be provided")

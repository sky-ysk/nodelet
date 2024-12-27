package handler

import (
	"fmt"
	"hit.edu/framework/pkg/apimachinery/errors"
	"hit.edu/framework/pkg/apimachinery/runtime"
	"hit.edu/framework/pkg/apimachinery/runtime/schema"
	"hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/apiserver/endpoints/handler/responsewriters"
	"io"
	"io/ioutil"
	
	"net/http"
	"time"
)

const (
	requestTimeoutUpperBound = 34 * time.Second
)

// RequestScope 封装RESTful处理程序方法的公共字段。
type RequestScope struct {
	Namer ScopeNamer
	
	Serializer runtime.NegotiatedSerializer
	runtime.ParameterCodec
	
	Creater   runtime.ObjectCreater
	Defaulter runtime.ObjectDefaulter
	Convertor runtime.ObjectConvertor
	Typer     runtime.ObjectTyper
	
	Resource schema.GroupVersionResource
	Kind     schema.GroupVersionKind
	
	Subresource string
	
	MetaGroupVersion schema.GroupVersion
	
	// HubGroupVersion 指示从etcd或传入请求中读取的版本对象应转换为内存中处理的版本对象。
	HubGroupVersion schema.GroupVersion
	
	MaxRequestBodyBytes int64
}

func (scope *RequestScope) AllowsMediaTypeTransform(mimeType, mimeSubType string, target *schema.GroupVersionKind) bool {
	return true
}

func (scope *RequestScope) AllowsServerVersion(version string) bool {
	return true
}

func (scope *RequestScope) AllowsStreamSchema(schema string) bool {
	return true
}

func (scope *RequestScope) err(err error, w http.ResponseWriter, req *http.Request) {
	responsewriters.ErrorNegotiated(err, scope.Serializer, scope.Kind.GroupVersion(), w, req)
}

func limitedReadBody(req *http.Request, limit int64) ([]byte, error) {
	defer req.Body.Close()
	if limit <= 0 {
		return ioutil.ReadAll(req.Body)
	}
	lr := &io.LimitedReader{
		R: req.Body,
		N: limit + 1,
	}
	data, err := ioutil.ReadAll(lr)
	if err != nil {
		return nil, err
	}
	if lr.N <= 0 {
		return nil, errors.NewRequestEntityTooLargeError(fmt.Sprintf("limit is %d", limit))
	}
	return data, nil
}

func hasUID(obj runtime.Object) (bool, error) {
	if obj == nil {
		return false, nil
	}
	accessor, err := meta.Accessor(obj)
	if err != nil {
		return false, errors.NewInternalError(err)
	}
	if len(accessor.GetUID()) == 0 {
		return false, nil
	}
	return true, nil
}

// checkName 根据请求检查所提供的名称
func checkName(obj runtime.Object, name string, namer ScopeNamer) error {
	objName, err := namer.ObjectName(obj)
	if err != nil {
		return errors.NewBadRequest(fmt.Sprintf(
			"the name of the object (%s based on URL) was undeterminable: %v", name, err))
	}
	if objName != name {
		return errors.NewBadRequest(fmt.Sprintf(
			"the name of the object (%s) does not match the name on the URL (%s)", objName, name))
	}
	
	return nil
}

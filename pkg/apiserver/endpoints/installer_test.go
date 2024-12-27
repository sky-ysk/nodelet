package endpoints

import (
	"hit.edu/framework/pkg/apimachinery/runtime/schema"
	"hit.edu/framework/pkg/apis/meta"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

func TestRegisterResourceHandler(t *testing.T) {
	restStorage, embedEtcdServer := getTestRESTStorage(t)
	defer embedEtcdServer.Terminate(t)
	apiInstaller := &APIInstaller{
		group: &APIGroupVersion{
			Storage:           restStorage,
			Root:              "/apis",
			GroupVersion:      schema.GroupVersion{Group: "resources", Version: "v1"},
			MetaGroupVersion:  &meta.SchemeGroupVersion,
			ParameterCodec:    meta.ParameterCodec,
			Serializer:        codecs,
			Typer:             scheme,
			Creater:           scheme,
			Namer:             namer,
			Convertor:         scheme,
			Defaulter:         scheme,
			MinRequestTimeout: 1800,
		},
		prefix:            "/apis/resources/v1",
		minRequestTimeout: 1800,
	}
	ws, errs := apiInstaller.Install()
	if errs != nil {
		t.Errorf("Install failed")
	}

	expectedRoutes := []struct {
		Method string
		Path   string
		Func   string
	}{
		{"POST", "/apis/resources/v1/nodes", "restfulCreateResource"},
		{"DELETE", "/apis/resources/v1/nodes", "restfulDeleteCollection"},
		{"GET", "/apis/resources/v1/nodes/{name}", "restfulGetResource"},
		{"DELETE", "/apis/resources/v1/nodes/{name}", "restfulDeleteResource"},
		{"PUT", "/apis/resources/v1/nodes/{name}", "restfulUpdateResource"},
		{"PATCH", "/apis/resources/v1/nodes/{name}", "restfulPatchResource"},
		{"GET", "/apis/resources/v1/nodes", "restfulListResource"},
	}

	for _, expectedRoute := range expectedRoutes {
		foundPath := false
		isExpectedFunc := false
		for _, route := range ws.Routes() {
			routeFuncName := getFunctionName(route.Function)
			if route.Method == expectedRoute.Method && route.Path == expectedRoute.Path {
				foundPath = true
				if strings.Contains(routeFuncName, expectedRoute.Func) {
					isExpectedFunc = true
					break
				}
			}
		}
		if !foundPath {
			t.Errorf("Method:%s,Path:%s not register in WebService", expectedRoute.Method, expectedRoute.Path)
		}
		if !isExpectedFunc {
			t.Errorf("Func:%s not register to route Method:%s,Path:%s", expectedRoute.Func, expectedRoute.Method, expectedRoute.Path)
		}
	}

}

func getFunctionName(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}

package filters

import (
	"hit.edu/framework/pkg/apiserver/endpoints/request"
	"k8s.io/apimachinery/pkg/util/sets"
)

func newTestRequestInfoResolver() *request.RequestInfoFactory {
	return &request.RequestInfoFactory{
		APIPrefixes: sets.NewString("apis"),
	}
}

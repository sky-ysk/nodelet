package install

import (
	"hit.edu/framework/pkg/apimachinery/runtime"
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/apis/legacyscheme"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
)

func init() {
	Install(legacyscheme.Scheme)
}

// Install 注册core资源
func Install(scheme *runtime.Scheme) {
	utilruntime.Must(apis.AddToScheme(scheme))
}

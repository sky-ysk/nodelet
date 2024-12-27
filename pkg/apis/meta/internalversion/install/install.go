package install

import (
	"hit.edu/framework/pkg/apimachinery/runtime"
	"hit.edu/framework/pkg/apis/legacyscheme"
	metainternalversion "hit.edu/framework/pkg/apis/meta/internalversion"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
)

func init() {
	Install(legacyscheme.Scheme)
}

// Install 注册core资源
func Install(scheme *runtime.Scheme) {
	utilruntime.Must(metainternalversion.AddToScheme(scheme))
}

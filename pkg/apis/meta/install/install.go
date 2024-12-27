package install

import (
	"hit.edu/framework/pkg/apimachinery/runtime"
	"hit.edu/framework/pkg/apis/legacyscheme"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	
	meta "hit.edu/framework/pkg/apis/meta/internalversion"
)

func init() {
	Install(legacyscheme.Scheme)
}

// Install 注册core资源
func Install(scheme *runtime.Scheme) {
	utilruntime.Must(meta.AddToScheme(scheme))
}

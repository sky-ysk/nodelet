package scheme

import (
	"hit.edu/framework/pkg/apimachinery/runtime"
	"hit.edu/framework/pkg/apimachinery/runtime/serializer"
	"hit.edu/framework/pkg/apis/meta/internalversion"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
)

var Scheme = runtime.NewScheme()

var Codecs = serializer.NewCodecFactory(Scheme)

var ParameterCodec = runtime.NewParameterCodec(Scheme)

func init() {
	utilruntime.Must(internalversion.AddToScheme(Scheme))
}

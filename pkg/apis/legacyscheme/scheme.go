package legacyscheme

import (
	"hit.edu/framework/pkg/apimachinery/runtime"
	"hit.edu/framework/pkg/apimachinery/runtime/serializer"
)

var (
	Scheme = runtime.NewScheme()
	
	Codecs = serializer.NewCodecFactory(Scheme)
	
	ParameterCodec = runtime.NewParameterCodec(Scheme)
)

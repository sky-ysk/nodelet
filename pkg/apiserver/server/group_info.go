package server

import (
	"hit.edu/framework/pkg/apimachinery/runtime"
	"hit.edu/framework/pkg/apimachinery/runtime/schema"
	"hit.edu/framework/pkg/apimachinery/runtime/serializer"
	"hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/apiserver/registry/rest"
)

// API Group
type APIGroupInfo struct {
	//
	VersionedResourcesStorageMap map[string]map[string]rest.Storage
	
	MetaGroupVersion *schema.GroupVersion
	
	// From K8s
	// NegotiatedSerializer controls how this group encodes and decodes data
	NegotiatedSerializer runtime.NegotiatedSerializer
	// ParameterCodec performs conversions for query parameters passed to API calls
	ParameterCodec runtime.ParameterCodec
	
	Scheme *runtime.Scheme
}

func NewDefaultAPIGroupInfo(Scheme *runtime.Scheme, parameterCodec runtime.ParameterCodec, codecs serializer.CodecFactory) APIGroupInfo {
	return APIGroupInfo{
		VersionedResourcesStorageMap: map[string]map[string]rest.Storage{},
		MetaGroupVersion:             &meta.SchemeGroupVersion,
		// TODO unhardcode this.  It was hardcoded before, but we need to re-evaluate
		Scheme:               Scheme,
		ParameterCodec:       parameterCodec,
		NegotiatedSerializer: codecs,
	}
}

func (a *APIGroupInfo) destroyStorage() {
	for _, stores := range a.VersionedResourcesStorageMap {
		for _, store := range stores {
			store.Destroy()
		}
	}
}

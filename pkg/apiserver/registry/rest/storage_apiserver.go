package rest

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apiserver/pkg/registry/generic"
	"k8s.io/apiserver/pkg/registry/rest"
	genericapiserver "k8s.io/apiserver/pkg/server"
	serverstorage "k8s.io/apiserver/pkg/server/storage"
	
	"k8s.io/kube-aggregator/pkg/apis/apiregistration"
	v1 "k8s.io/kube-aggregator/pkg/apis/apiregistration/v1"
	aggregatorscheme "k8s.io/kube-aggregator/pkg/apiserver/scheme"
	apiservicestorage "k8s.io/kube-aggregator/pkg/registry/apiservice/etcd"
)

// NewRESTStorage returns an APIGroupInfo object that will work against apiservice.
func NewRESTStorage(apiResourceConfigSource serverstorage.APIResourceConfigSource, restOptionsGetter generic.RESTOptionsGetter, shouldServeBeta bool) genericapiserver.APIGroupInfo {
	apiGroupInfo := genericapiserver.NewDefaultAPIGroupInfo(apiregistration.GroupName, aggregatorscheme.Scheme, metav1.ParameterCodec, aggregatorscheme.Codecs)
	
	storage := map[string]rest.Storage{}
	
	if resource := "apiservices"; apiResourceConfigSource.ResourceEnabled(v1.SchemeGroupVersion.WithResource(resource)) {
		apiServiceREST := apiservicestorage.NewREST(aggregatorscheme.Scheme, restOptionsGetter)
		storage[resource] = apiServiceREST
		storage[resource+"/status"] = apiservicestorage.NewStatusREST(aggregatorscheme.Scheme, apiServiceREST)
	}
	
	if len(storage) > 0 {
		apiGroupInfo.VersionedResourcesStorageMap["v1"] = storage
	}
	
	return apiGroupInfo
}

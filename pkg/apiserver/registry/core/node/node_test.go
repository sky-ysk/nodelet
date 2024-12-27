package node

import (
	"context"
	"fmt"
	"testing"
	
	"hit.edu/framework/pkg/apis/meta"
	//coretesting "hit.edu/framework/pkg/apiserver/registry/core/pod/testing"
	registryrest "hit.edu/framework/pkg/apiserver/registry/rest"
	
	"hit.edu/framework/pkg/apimachinery/runtime"
	//"hit.edu/framework/pkg/apiserver/registry/core/rest"
	"hit.edu/framework/pkg/apiserver/registry/generic"
	"hit.edu/framework/pkg/apimachinery/runtime/schema"
	//"k8s.io/apimachinery/pkg/api/meta"
	//metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/apiserver/registry/rest"
	"hit.edu/framework/pkg/apiserver/registry/storage/storagebackend"
	"k8s.io/apimachinery/pkg/api/apitesting"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	genericapirequest "k8s.io/apiserver/pkg/endpoints/request"
	//genericregistrytest "hit.edu/framework/pkg/apiserver/registry/generic/registry/testing"
)

var scheme = runtime.NewScheme()
var codecs = serializer.NewCodecFactory(scheme)

const GroupName = "etcd3test"

var SchemeGroupVersion = schema.GroupVersion{Group: GroupName, Version: "v1"}

func init() {
	//meta.AddToGroupVersion(scheme, meta.SchemeGroupVersion)
	NodeObject := []runtime.Object{
		&apis.Node{},
		&apis.NodeList{},
	}
	addKnownTypes := func(scheme *runtime.Scheme) error {
		scheme.AddKnownTypes(SchemeGroupVersion, NodeObject...)
		
		if err := meta.RegisterConversions(scheme); err != nil {
			panic(err)
		}
		return nil
	}
	
	addUnversionedTypes := func(scheme *runtime.Scheme) error {
		scheme.AddUnversionedTypes(SchemeGroupVersion, NodeObject...)
		return nil
	}
	
	SchemeBuilder := runtime.NewSchemeBuilder(addKnownTypes, addUnversionedTypes)
	AddToScheme := SchemeBuilder.AddToScheme
	utilruntime.Must(AddToScheme(scheme))
	
	// scheme.AddUnversionedTypes(SchemeGroupVersion, NodeObject...)
	// meta.AddToScheme(scheme)
	thisStrategy = Strategy{scheme}
}

func NewEtcdStorage(t *testing.T, group string) (*storagebackend.ConfigForResource, *EtcdTestServer) {
	return NewEtcdStorageForResource(t, schema.GroupResource{Group: group, Resource: "any"})
}

func NewEtcdStorageForResource(t *testing.T, resource schema.GroupResource) (*storagebackend.ConfigForResource, *EtcdTestServer) {
	t.Helper()
	server, config := NewUnsecuredEtcd3TestClientServer(t)
	testcodec := apitesting.TestStorageCodec(codecs, SchemeGroupVersion)
	config.Codec = testcodec
	resourceConfig := &storagebackend.ConfigForResource{
		Config:        *config,
		GroupResource: resource,
	}
	return resourceConfig, server
}

func newStorage(t *testing.T) (*NodeStorage, *EtcdTestServer) {
	//启动etcd服务器
	etcdStorage, server := NewEtcdStorage(t, "")
	restOptions := generic.RESTOptions{
		StorageConfig:           etcdStorage,
		Decorator:               generic.UndecoratedStorage,
		DeleteCollectionWorkers: 3,
		ResourcePrefix:          "nodes",
	}
	storage, err := NewNodeStorage(restOptions)
	storage1, _, _ := generic.NewRawStorage(etcdStorage, nil, nil, "nodes")
	storage.Node.Storage = storage1
	if err != nil {
		t.Fatalf("unexpected error from REST storage: %v", err)
	}
	return &storage, server
	
}
func TestCreate(t *testing.T) {
	nodestorage, server := newStorage(t)
	defer server.Terminate(t)
	defer nodestorage.Node.Store.DestroyFunc()
	nodeA := &apis.Node{
		ObjectMeta: meta.ObjectMeta{Name: "foo"},
		Spec:       apis.NodeSpec{NodeName: "test", HostName: "testhost", Unschedulable: false},
		Status:     apis.NodeStatus{},
	}
	nodeA1 := &apis.Node{
		ObjectMeta: meta.ObjectMeta{Name: "foo"},
		Spec:       apis.NodeSpec{NodeName: "test1", HostName: "testhost1", Unschedulable: false},
		Status:     apis.NodeStatus{},
	}
	testContext := genericapirequest.NewContext()
	
	_, err := nodestorage.Node.Create(testContext, nodeA, registryrest.ValidateAllObjectFunc)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	
	_, _, err = nodestorage.Node.Update(testContext, nodeA.Name, rest.DefaultUpdatedObjectInfo(nodeA1), rest.ValidateAllObjectFunc, rest.ValidateAllObjectUpdateFunc, false, &meta.UpdateOptions{})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	obj, err := nodestorage.Node.Get(testContext, nodeA.Name, &meta.GetOptions{})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	
	accessor, _ := meta.Accessor(obj)
	//resourceVersion := accessor.GetResourceVersion()
	Name := accessor.GetName()
	if Name != "foo" {
		t.Errorf("name not foo: %s", Name)
	}
}

func denyCreateValidation(ctx context.Context, obj runtime.Object) error {
	return fmt.Errorf("admission denied")
}
func denyUpdateValidation(ctx context.Context, obj, old runtime.Object) error {
	return fmt.Errorf("admission denied")
}

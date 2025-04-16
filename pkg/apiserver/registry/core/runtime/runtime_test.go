package runtime

import (
	"context"
	"fmt"
	"testing"

	"hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/component-base/logs"
	"k8s.io/apimachinery/pkg/selection"

	//coretesting "hit.edu/framework/pkg/apiserver/registry/core/pod/testing"
	"hit.edu/framework/pkg/apimachinery/fields"
	"hit.edu/framework/pkg/apimachinery/labels"
	"hit.edu/framework/pkg/apimachinery/runtime"
	registryrest "hit.edu/framework/pkg/apiserver/registry/rest"

	//"hit.edu/framework/pkg/apiserver/registry/core/rest"
	"hit.edu/framework/pkg/apimachinery/runtime/schema"
	"hit.edu/framework/pkg/apiserver/registry/generic"

	//"k8s.io/apimachinery/pkg/api/meta"
	//metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/apiserver/registry/rest"
	"hit.edu/framework/pkg/apiserver/registry/storage"
	"hit.edu/framework/pkg/apiserver/registry/storage/storagebackend"

	//"k8s.io/apimachinery/pkg/api/apitesting"
	"hit.edu/framework/pkg/apimachinery/runtime/serializer"
	genericapirequest "hit.edu/framework/pkg/apiserver/endpoints/request"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	//genericregistrytest "hit.edu/framework/pkg/apiserver/registry/generic/registry/testing"
)

var scheme = runtime.NewScheme()
var codecs = serializer.NewCodecFactory(scheme)

const GroupName = "etcd3test"

var SchemeGroupVersion = schema.GroupVersion{Group: GroupName, Version: "v1"}

func init() {
	logs.Init("etcd")
	logs.Info("this is a test in Runtime")
	//meta.AddToGroupVersion(scheme, meta.SchemeGroupVersion)
	RuntimeObject := []runtime.Object{
		&apis.Runtime{},
		&apis.RuntimeList{},
	}
	addKnownTypes := func(scheme *runtime.Scheme) error {
		scheme.AddKnownTypes(SchemeGroupVersion, RuntimeObject...)

		if err := meta.RegisterConversions(scheme); err != nil {
			panic(err)
		}
		return nil
	}

	addUnversionedTypes := func(scheme *runtime.Scheme) error {
		scheme.AddUnversionedTypes(SchemeGroupVersion, RuntimeObject...)
		return nil
	}

	SchemeBuilder := runtime.NewSchemeBuilder(addKnownTypes, addUnversionedTypes)
	AddToScheme := SchemeBuilder.AddToScheme
	utilruntime.Must(AddToScheme(scheme))

	// scheme.AddUnversionedTypes(SchemeGroupVersion, RuntimeObject...)
	// meta.AddToScheme(scheme)
	thisStrategy = &Strategy{scheme}
}

func NewEtcdStorage(t *testing.T, group string) (*storagebackend.ConfigForResource, *EtcdTestServer) {
	return NewEtcdStorageForResource(t, schema.GroupResource{Group: group, Resource: "any"})
}

func NewEtcdStorageForResource(t *testing.T, resource schema.GroupResource) (*storagebackend.ConfigForResource, *EtcdTestServer) {
	t.Helper()
	server, config := NewUnsecuredEtcd3TestClientServer(t)
	testcodec := serializer.NewCodecFactory(scheme).LegacyCodec()
	//testcodec := apitesting.TestStorageCodec(codecs, SchemeGroupVersion)
	config.Codec = testcodec
	resourceConfig := &storagebackend.ConfigForResource{
		Config:        *config,
		GroupResource: resource,
	}
	return resourceConfig, server
}

func newStorage(t *testing.T) (*RuntimeStorage, *EtcdTestServer) {
	//启动etcd服务器
	etcdStorage, server := NewEtcdStorage(t, "")
	restOptions := generic.RESTOptions{
		StorageConfig:           etcdStorage,
		Decorator:               generic.UndecoratedStorage,
		DeleteCollectionWorkers: 3,
		ResourcePrefix:          "runtimes",
	}
	storage, err := NewRuntimeStorage(restOptions)
	storage1, _, _ := generic.NewRawStorage(etcdStorage, nil, nil, "runtimes")
	storage.Runtime.Storage = storage1
	if err != nil {
		t.Fatalf("unexpected error from REST storage: %v", err)
	}
	return &storage, server

}

func getRuntimeapp(obj runtime.Object) (labels.Set, fields.Set, error) {
	Runtime := obj.(*apis.Runtime)
	return labels.Set{"app": Runtime.Labels["app"]}, nil, nil
}

func matchapp(names ...string) storage.SelectionPredicate {
	l, err := labels.NewRequirement("app", selection.In, names)
	if err != nil {
		panic("Labels set not successful")
	}
	return storage.SelectionPredicate{
		Label:    labels.Everything().Add(*l),
		Field:    fields.Everything(),
		GetAttrs: getRuntimeapp,
	}
}

func TestCreate(t *testing.T) {
	Runtimestorage, server := newStorage(t)
	defer server.Terminate(t)
	defer Runtimestorage.Runtime.Store.DestroyFunc()
	RuntimeA := &apis.Runtime{
		ObjectMeta: meta.ObjectMeta{Name: "foo"},
		Spec:       apis.RuntimeSpec{Name: "test", Image: "testhost"},
		Status:     apis.RuntimeStatus{},
	}
	RuntimeA1 := &apis.Runtime{
		ObjectMeta: meta.ObjectMeta{Name: "foo"},
		Spec:       apis.RuntimeSpec{Name: "test1", Image: "testhost1"},
		Status:     apis.RuntimeStatus{},
	}
	testContext := genericapirequest.WithNamespace(genericapirequest.NewContext(), "aaa")

	runtime1, err := Runtimestorage.Runtime.Create(testContext, RuntimeA, registryrest.ValidateAllObjectFunc)
	if err != nil {

		t.Errorf("Unexpected error: %v", err)
	}

	runtime1, _, err = Runtimestorage.Runtime.Update(testContext, RuntimeA.Name, rest.DefaultUpdatedObjectInfo(RuntimeA1), rest.ValidateAllObjectFunc, rest.ValidateAllObjectUpdateFunc, false, &meta.UpdateOptions{})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	obj, err := Runtimestorage.Runtime.Get(testContext, RuntimeA.Name, &meta.GetOptions{})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	fmt.Println(runtime1)
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

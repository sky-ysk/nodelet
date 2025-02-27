package event

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
	logs.Info("this is a test in Event")
	//meta.AddToGroupVersion(scheme, meta.SchemeGroupVersion)
	EventObject := []runtime.Object{
		&apis.Event{},
		&apis.EventList{},
	}
	addKnownTypes := func(scheme *runtime.Scheme) error {
		scheme.AddKnownTypes(SchemeGroupVersion, EventObject...)

		if err := meta.RegisterConversions(scheme); err != nil {
			panic(err)
		}
		return nil
	}

	addUnversionedTypes := func(scheme *runtime.Scheme) error {
		scheme.AddUnversionedTypes(SchemeGroupVersion, EventObject...)
		return nil
	}

	SchemeBuilder := runtime.NewSchemeBuilder(addKnownTypes, addUnversionedTypes)
	AddToScheme := SchemeBuilder.AddToScheme
	utilruntime.Must(AddToScheme(scheme))

	// scheme.AddUnversionedTypes(SchemeGroupVersion, EventObject...)
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

func newStorage(t *testing.T) (*EventStorage, *EtcdTestServer) {
	//启动etcd服务器
	etcdStorage, server := NewEtcdStorage(t, "")
	restOptions := generic.RESTOptions{
		StorageConfig:           etcdStorage,
		Decorator:               generic.UndecoratedStorage,
		DeleteCollectionWorkers: 3,
		ResourcePrefix:          "Events",
	}
	storage, err := NewEventStorage(restOptions)
	storage1, _, _ := generic.NewRawStorage(etcdStorage, nil, nil, "Events")
	storage.Event.Storage = storage1
	if err != nil {
		t.Fatalf("unexpected error from REST storage: %v", err)
	}
	return &storage, server

}

func getEventapp(obj runtime.Object) (labels.Set, fields.Set, error) {
	Event := obj.(*apis.Event)
	return labels.Set{"app": Event.Labels["app"]}, nil, nil
}

func matchapp(names ...string) storage.SelectionPredicate {
	l, err := labels.NewRequirement("app", selection.In, names)
	if err != nil {
		panic("Labels set not successful")
	}
	return storage.SelectionPredicate{
		Label:    labels.Everything().Add(*l),
		Field:    fields.Everything(),
		GetAttrs: getEventapp,
	}
}

func TestCreate(t *testing.T) {
	Eventstorage, server := newStorage(t)
	defer server.Terminate(t)
	defer Eventstorage.Event.Store.DestroyFunc()
	EventA := &apis.Event{
		ObjectMeta: meta.ObjectMeta{Name: "foo", Namespace: "aaa"},
		Reason:     "aaa",
		Message:    "aaa",
		Type:       "aaa",
	}
	EventA1 := &apis.Event{
		ObjectMeta: meta.ObjectMeta{Name: "foo", Namespace: "aaa"},
		Reason:     "aaa1",
		Message:    "aaa1",
		Type:       "aaa1",
	}
	testContext := genericapirequest.NewContext()

	_, err := Eventstorage.Event.Create(testContext, EventA, registryrest.ValidateAllObjectFunc)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	_, _, err = Eventstorage.Event.Update(testContext, EventA.Name, rest.DefaultUpdatedObjectInfo(EventA1), rest.ValidateAllObjectFunc, rest.ValidateAllObjectUpdateFunc, false, &meta.UpdateOptions{})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	obj, err := Eventstorage.Event.Get(testContext, EventA.Name, &meta.GetOptions{})
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

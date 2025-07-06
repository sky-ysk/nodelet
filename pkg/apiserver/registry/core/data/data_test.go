package data

import (
	"context"
	"fmt"
	"testing"

	"hit.edu/framework/pkg/apis/meta"
	storagetesting "hit.edu/framework/pkg/apiserver/registry/generic/registry/testing"
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
	logs.Info("this is a test in Data")
	//meta.AddToGroupVersion(scheme, meta.SchemeGroupVersion)
	DataObject := []runtime.Object{
		&apis.Data{},
		&apis.DataList{},
	}
	addKnownTypes := func(scheme *runtime.Scheme) error {
		scheme.AddKnownTypes(SchemeGroupVersion, DataObject...)

		if err := meta.RegisterConversions(scheme); err != nil {
			panic(err)
		}
		return nil
	}

	addUnversionedTypes := func(scheme *runtime.Scheme) error {
		scheme.AddUnversionedTypes(SchemeGroupVersion, DataObject...)
		return nil
	}

	SchemeBuilder := runtime.NewSchemeBuilder(addKnownTypes, addUnversionedTypes)
	AddToScheme := SchemeBuilder.AddToScheme
	utilruntime.Must(AddToScheme(scheme))

	// scheme.AddUnversionedTypes(SchemeGroupVersion, DataObject...)
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

func newStorage(t *testing.T) (*DataStorage, *EtcdTestServer) {
	//启动etcd服务器
	etcdStorage, server := NewEtcdStorage(t, "")
	restOptions := generic.RESTOptions{
		StorageConfig:           etcdStorage,
		Decorator:               generic.UndecoratedStorage,
		DeleteCollectionWorkers: 3,
		ResourcePrefix:          "Datas",
	}
	storage, err := NewDataStorage(restOptions)
	storage1, _, _ := generic.NewRawStorage(etcdStorage, nil, nil, "Datas")
	storage.Data.Storage = storage1
	if err != nil {
		t.Fatalf("unexpected error from REST storage: %v", err)
	}
	return &storage, server

}

func getDataapp(obj runtime.Object) (labels.Set, fields.Set, error) {
	Data := obj.(*apis.Data)
	return labels.Set{"app": Data.Labels["app"]}, nil, nil
}

func matchapp(names ...string) storage.SelectionPredicate {
	l, err := labels.NewRequirement("app", selection.In, names)
	if err != nil {
		panic("Labels set not successful")
	}
	return storage.SelectionPredicate{
		Label:    labels.Everything().Add(*l),
		Field:    fields.Everything(),
		GetAttrs: getDataapp,
	}
}

func TestLabels(t *testing.T) {
	Datastorage, server := newStorage(t)
	defer server.Terminate(t)
	defer Datastorage.Data.Store.DestroyFunc()
	testContext := genericapirequest.WithNamespace(genericapirequest.NewContext(), "aaa")
	DataA := &apis.Data{
		ObjectMeta: meta.ObjectMeta{Name: "fooA", Namespace: "aaa", Labels: map[string]string{"app": "myapp"}},
		Spec:       apis.DataSpec{Name: "testA", BelongNode: "testhost"},
		Status:     apis.DataStatus{},
	}
	DataA1 := &apis.Data{
		ObjectMeta: meta.ObjectMeta{Name: "fooA1", Namespace: "aaa", Labels: map[string]string{"app": "myapp"}},
		Spec:       apis.DataSpec{Name: "testA1", BelongNode: "testhost1"},
		Status:     apis.DataStatus{},
	}
	DataB := &apis.Data{
		ObjectMeta: meta.ObjectMeta{Name: "fooB", Namespace: "aaa", Labels: map[string]string{"app": "myapp1"}},
		Spec:       apis.DataSpec{Name: "testB", BelongNode: "testhost"},
		Status:     apis.DataStatus{},
	}
	DataB1 := &apis.Data{
		ObjectMeta: meta.ObjectMeta{Name: "fooB1", Namespace: "aaa", Labels: map[string]string{"app": "myapp1"}},
		Spec:       apis.DataSpec{Name: "testB1", BelongNode: "testhost1"},
		Status:     apis.DataStatus{},
	}

	in := &apis.DataList{Items: []apis.Data{*DataA, *DataA1, *DataB, *DataB1}}

	if err := storagetesting.CreateList("/Datas", Datastorage.Data.Storage, in); err != nil {
		t.Fatalf("Unexpected error %v", err)
	}

	table := map[string]struct {
		m       storage.SelectionPredicate
		out     runtime.Object
		context context.Context
	}{
		"notFound": {
			m:   matchapp("myapp2"),
			out: &apis.DataList{Items: []apis.Data{}},
		},
		"matchmyapp": {
			m:   matchapp("myapp"),
			out: &apis.DataList{Items: []apis.Data{*DataA, *DataA1}},
		},
		"matchmyapp1": {
			m:   matchapp("myapp1"),
			out: &apis.DataList{Items: []apis.Data{*DataB, *DataB1}},
		},
	}
	for name, item := range table {
		t.Run(name, func(t *testing.T) {
			list, err := Datastorage.Data.ListPredicate(testContext, item.m, nil)
			if err != nil {
				t.Fatalf("Unexpected error %v", err)
			}
			if list != item.out {
			}
		})
	}
}

func TestCreate(t *testing.T) {
	Datastorage, server := newStorage(t)
	defer server.Terminate(t)
	defer Datastorage.Data.Store.DestroyFunc()
	DataA := &apis.Data{
		ObjectMeta: meta.ObjectMeta{Name: "foo"},
		Spec:       apis.DataSpec{Name: "test", BelongNode: "testhost"},
		Status:     apis.DataStatus{},
	}
	DataA1 := &apis.Data{
		ObjectMeta: meta.ObjectMeta{Name: "foo"},
		Spec:       apis.DataSpec{Name: "test1", BelongNode: "testhost1"},
		Status:     apis.DataStatus{},
	}
	testContext := genericapirequest.WithNamespace(genericapirequest.NewContext(), "aaa")

	_, err := Datastorage.Data.Create(testContext, DataA, registryrest.ValidateAllObjectFunc)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	_, _, err = Datastorage.Data.Update(testContext, DataA.Name, rest.DefaultUpdatedObjectInfo(DataA1), rest.ValidateAllObjectFunc, rest.ValidateAllObjectUpdateFunc, false, &meta.UpdateOptions{})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	obj, err := Datastorage.Data.Get(testContext, DataA.Name, &meta.GetOptions{})
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

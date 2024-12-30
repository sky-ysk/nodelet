package registry

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"hit.edu/framework/pkg/apimachinery/runtime"
	"hit.edu/framework/pkg/apimachinery/runtime/schema"
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/apis/meta"
	genericapirequest "hit.edu/framework/pkg/apiserver/endpoints/request"
	storagetesting "hit.edu/framework/pkg/apiserver/registry/generic/registry/testing"
	"hit.edu/framework/pkg/apiserver/registry/rest"
	"hit.edu/framework/pkg/apiserver/registry/storage"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
)

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
}

func TestStoreCreate(t *testing.T) {
	nodeA := &apis.Node{
		ObjectMeta: meta.ObjectMeta{Name: "foo", Namespace: "aaa"},
	}
	testContext := genericapirequest.WithNamespace(genericapirequest.NewContext(), "aaa")
	destroyFunc, registry := NewTestGenericStoreRegistry(t)
	defer destroyFunc()
	defaultDeleteStrategy := testRESTStrategy{scheme, true}
	registry.DeleteStrategy = testGracefulStrategy{defaultDeleteStrategy}
	_, err := registry.Create(testContext, nodeA, denyCreateValidation)
	if err == nil {
		t.Errorf("Expected admission error: %v", err)
	}
	objA, err := registry.Create(testContext, nodeA, rest.ValidateAllObjectFunc)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	checkobj, err := registry.Get(testContext, nodeA.Name, &meta.GetOptions{})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if e, a := objA, checkobj; !reflect.DeepEqual(e, a) {
		t.Errorf("Expected %#v, got %#v", e, a)
	}
}

// 测试ListPredicate能否完成GetList操作，List并根据标签筛选
func TestStoreList(t *testing.T) {
	nodeA := &apis.Node{
		ObjectMeta: meta.ObjectMeta{Name: "bar", Namespace: "aaa"},
		TypeMeta:   meta.TypeMeta{Kind: "Node", APIVersion: "etcd3test/v1"},
	}
	nodeB := &apis.Node{
		ObjectMeta: meta.ObjectMeta{Name: "foo", Namespace: "aaa"},
		TypeMeta:   meta.TypeMeta{Kind: "Node", APIVersion: "etcd3test/v1"},
	}
	testContext := genericapirequest.WithNamespace(genericapirequest.NewContext(), "aaa")
	table := map[string]struct {
		in      *apis.NodeList
		m       storage.SelectionPredicate
		out     runtime.Object
		context context.Context
	}{
		"notFound": {
			in:  nil,
			m:   matchEverything(),
			out: &apis.NodeList{Items: []apis.Node{}},
		},
		"normal": {
			in:  &apis.NodeList{Items: []apis.Node{*nodeA, *nodeB}},
			m:   matchEverything(),
			out: &apis.NodeList{Items: []apis.Node{*nodeA, *nodeB}},
		},
		"normalFiltered": {
			in:  &apis.NodeList{Items: []apis.Node{*nodeA, *nodeB}},
			m:   matchNodeName("foo"),
			out: &apis.NodeList{Items: []apis.Node{*nodeB}},
		},
		"normalFilteredMatchMultiple": {
			in:  &apis.NodeList{Items: []apis.Node{*nodeA, *nodeB}},
			m:   matchNodeName("foo", "makeMatchSingleReturnFalse"),
			out: &apis.NodeList{Items: []apis.Node{*nodeB}},
		},
	}

	for name, item := range table {
		t.Run(name, func(t *testing.T) {
			ctx := testContext
			if item.context != nil {
				ctx = item.context
			}
			destroyFunc, registry := NewTestGenericStoreRegistry(t)
			defer destroyFunc()

			if item.in != nil {
				if err := storagetesting.CreateList("/nodes", registry.Storage, item.in); err != nil {
					t.Fatalf("Unexpected error %v", err)
				}
			}

			list, err := registry.ListPredicate(ctx, item.m, nil)
			if err != nil {
				t.Fatalf("Unexpected error %v", err)
			}
			// fmt.Println(list)
			// fmt.Println(item.out)
			if e, a := item.out, list; !compareNodeList(e, a) {
				t.Fatalf("%v: Expected %#v, got %#v", name, e, a)
			}
		})
	}
}

// 测试Update能否正确处理资源的更新
func TestStoreUpdate(t *testing.T) {
	podA := &apis.Node{
		ObjectMeta: meta.ObjectMeta{Name: "foo", Namespace: "aaa"},
	}
	podB := &apis.Node{
		ObjectMeta: meta.ObjectMeta{Name: "foo", Namespace: "aaa"},
		Spec:       apis.NodeSpec{NodeName: "machine"},
	}
	testContext := genericapirequest.WithNamespace(genericapirequest.NewContext(), "aaa")
	destroyFunc, registry := NewTestGenericStoreRegistry(t)
	defer destroyFunc()

	// try to update a non-existing node
	_, _, err := registry.Update(testContext, podA.Name, rest.DefaultUpdatedObjectInfo(podA), rest.ValidateAllObjectFunc, rest.ValidateAllObjectUpdateFunc, false, &meta.UpdateOptions{})
	if err == nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// create the object
	_, err = registry.Create(testContext, podA, rest.ValidateAllObjectFunc)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// normal update
	_, _, err = registry.Update(testContext, podB.Name, rest.DefaultUpdatedObjectInfo(podB), rest.ValidateAllObjectFunc, rest.ValidateAllObjectUpdateFunc, false, &meta.UpdateOptions{})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if !updateAndVerify(t, testContext, registry, podA) {
		t.Errorf("Unexpected error updating podA")
	}
	if !updateAndVerify(t, testContext, registry, podB) {
		t.Errorf("Unexpected error updating podB")
	}
	if !updateAndVerify(t, testContext, registry, podA) {
		t.Errorf("Unexpected error updating podA")
	}
}

func TestStoreGet(t *testing.T) {
	podA := &apis.Node{
		ObjectMeta: meta.ObjectMeta{Name: "foo", Namespace: "aaa"},
		Spec:       apis.NodeSpec{NodeName: "machine"},
	}

	testContext := genericapirequest.WithNamespace(genericapirequest.NewContext(), "aaa")
	destroyFunc, registry := NewTestGenericStoreRegistry(t)
	defer destroyFunc()

	_, err := registry.Get(testContext, podA.Name, &meta.GetOptions{})
	if err == nil {
		fmt.Printf("expect not found")
	}

	// create the object
	_, err = registry.Create(testContext, podA, rest.ValidateAllObjectFunc)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if !updateAndVerify(t, testContext, registry, podA) {
		t.Errorf("Unexpected error creating podA")
	}
}

func TestStoreDelete(t *testing.T) {
	podA := &apis.Node{
		ObjectMeta: meta.ObjectMeta{Name: "foo", Namespace: "aaa"},
		Spec:       apis.NodeSpec{NodeName: "machine"},
	}

	testContext := genericapirequest.WithNamespace(genericapirequest.NewContext(), "aaa")
	destroyFunc, registry := NewTestGenericStoreRegistry(t)
	defer destroyFunc()

	// test failure condition
	_, _, err := registry.Delete(testContext, podA.Name, rest.ValidateAllObjectFunc, nil)
	if err == nil {
		fmt.Printf("expect not found")
	}

	// create pod
	_, err = registry.Create(testContext, podA, rest.ValidateAllObjectFunc)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// delete object
	_, wasDeleted, err := registry.Delete(testContext, podA.Name, rest.ValidateAllObjectFunc, nil)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !wasDeleted {
		t.Errorf("unexpected, pod %s should have been deleted immediately", podA.Name)
	}

	// try to get a item which should be deleted
	_, err = registry.Get(testContext, podA.Name, &meta.GetOptions{})
	if err == nil {
		fmt.Printf("expect not found")
	}
}

func TestStoreWatch(t *testing.T) {
	testContext := genericapirequest.WithNamespace(genericapirequest.NewContext(), "aaa")
	Context1 := genericapirequest.NewContext()

	table := map[string]struct {
		selectPred storage.SelectionPredicate
		context    context.Context
	}{
		"single": {
			selectPred: matchNodeName("foo"),
		},
		"multi": {
			selectPred: matchNodeName("foo", "bar"),
		},
		"singleNoNamespace": {
			selectPred: matchNodeName("foo"),
			context:    Context1,
		},
	}

	for name, m := range table {
		t.Run(name, func(t *testing.T) {
			ctx := testContext
			if m.context != nil {
				ctx = m.context
			}
			podA := &apis.Node{
				ObjectMeta: meta.ObjectMeta{
					Name:      "foo",
					Namespace: "aaa",
				},
				Spec: apis.NodeSpec{NodeName: "machine"},
			}

			destroyFunc, registry := NewTestGenericStoreRegistry(t)
			defer destroyFunc()
			wi, err := registry.WatchPredicate(ctx, m.selectPred, "0", nil, false)
			if err != nil {
				t.Errorf("%v: unexpected error: %v", name, err)
			} else {
				obj, err := registry.Create(testContext, podA, rest.ValidateAllObjectFunc)
				if err != nil {
					got, open := <-wi.ResultChan()
					if !open {
						t.Errorf("%v: unexpected channel close", name)
					} else {
						if e, a := obj, got.Object; !reflect.DeepEqual(e, a) {
							t.Errorf("Expected %#v, got %#v", e, a)
						}
					}
				}
				wi.Stop()
			}
		})
	}
}

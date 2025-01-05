package registry

import (
	"context"
	"fmt"
	"path"
	"testing"

	"hit.edu/framework/pkg/apimachinery/fields"
	"hit.edu/framework/pkg/apimachinery/labels"
	"hit.edu/framework/pkg/apimachinery/runtime"
	"hit.edu/framework/pkg/apimachinery/runtime/schema"
	"hit.edu/framework/pkg/apimachinery/runtime/serializer"
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/apiserver/registry/generic"
	storagetesting "hit.edu/framework/pkg/apiserver/registry/generic/registry/testing"
	"hit.edu/framework/pkg/apiserver/registry/storage"
	"hit.edu/framework/pkg/apiserver/registry/storage/storagebackend"
)

func NewTestGenericStoreRegistry(t *testing.T) (storagebackend.DestroyFunc, *Store) {
	return newTestGenericStoreRegistry(t, scheme)
}

func newTestGenericStoreRegistry(t *testing.T, scheme *runtime.Scheme) (storagebackend.DestroyFunc, *Store) {
	podPrefix := "/nodes"
	server, sc := storagetesting.NewUnsecuredEtcd3TestClientServer(t)
	strategy := &testRESTStrategy{scheme, true}

	newFunc := func() runtime.Object { return &apis.Node{} }
	newListFunc := func() runtime.Object { return &apis.NodeList{} }

	sc.Codec = serializer.NewCodecFactory(scheme).LegacyCodec()
	s, dFunc, err := storagebackend.Create(*sc.ForResource(schema.GroupResource{Resource: "nodes"}), newFunc, newListFunc, "/nodes")
	if err != nil {
		t.Fatalf("Error creating storage: %v", err)
	}
	destroyFunc := func() {
		dFunc()
		server.Terminate(t)
	}
	return destroyFunc, &Store{
		NewFunc:                   func() runtime.Object { return &apis.Node{} },
		NewListFunc:               func() runtime.Object { return &apis.NodeList{} },
		DefaultQualifiedResource:  apis.Resource("nodes"),
		SingularQualifiedResource: apis.Resource("node"),
		CreateStrategy:            strategy,
		UpdateStrategy:            strategy,
		DeleteStrategy:            strategy,
		KeyRootFunc: func(ctx context.Context) string {
			return podPrefix
		},
		KeyFunc: func(ctx context.Context, id string) (string, error) {
			return path.Join(podPrefix, id), nil
		},
		ObjectNameFunc: func(obj runtime.Object) (string, error) { return obj.(*apis.Node).Name, nil },
		PredicateFunc: func(label labels.Selector, field fields.Selector) storage.SelectionPredicate {
			return storage.SelectionPredicate{
				Label: label,
				Field: field,
				GetAttrs: func(obj runtime.Object) (labels.Set, fields.Set, error) {
					node, ok := obj.(*apis.Node)
					if !ok {
						return nil, nil, fmt.Errorf("not a node")
					}
					return labels.Set(node.ObjectMeta.Labels), generic.ObjectMetaFieldsSet(&node.ObjectMeta, true), nil
				},
			}
		},
		Storage: s,
	}
}

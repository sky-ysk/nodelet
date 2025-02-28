package etcd3

import (
	"context"
	//"fmt"
	"testing"

	"hit.edu/framework/pkg/apimachinery/fields"
	"hit.edu/framework/pkg/apimachinery/labels"
	"hit.edu/framework/pkg/apimachinery/runtime"
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/apis/meta"
	genericapirequest "hit.edu/framework/pkg/apiserver/endpoints/request"
	"hit.edu/framework/pkg/apiserver/registry/storage"
	"k8s.io/apimachinery/pkg/watch"
)

func TestNamespaceScopedWatch(t *testing.T) {
	ctx, store, _ := testSetup(t)
	RunTestNamespaceScopedWatch(ctx, t, store)
}

type testWatchStruct struct {
	obj         *apis.Node
	expectEvent bool
	watchType   watch.EventType
}

// It tests watch of namespace-scoped resources.
func RunTestNamespaceScopedWatch(ctx context.Context, t *testing.T, store storage.Interface) {
	tests := []struct {
		name               string
		requestedName      string
		requestedNamespace string
		recursive          bool
		fieldSelector      fields.Selector
		indexFields        []string
		watchTests         []*testWatchStruct
	}{
		{
			name:          "namespaced watch, request without name, request without namespace, without field selector",
			recursive:     true,
			fieldSelector: fields.Everything(),
			watchTests: []*testWatchStruct{
				{baseNamespacedPod("t1-foo1", "t1-ns1"), true, watch.Added},
				{baseNamespacedPod("t1-foo2", "t1-ns2"), true, watch.Added},
				{baseNamespacedPodUpdated("t1-foo1", "t1-ns1"), true, watch.Modified},
				{baseNamespacedPodUpdated("t1-foo2", "t1-ns2"), true, watch.Modified},
			},
		},
		{
			name:               "namespaced watch, request without name, request with namespace, without field selector",
			requestedNamespace: "t2-ns1",
			recursive:          true,
			fieldSelector:      fields.Everything(),
			watchTests: []*testWatchStruct{
				{baseNamespacedPod("t2-foo1", "t2-ns1"), true, watch.Added},
				{baseNamespacedPod("t2-foo1", "t2-ns2"), false, ""},
				{baseNamespacedPod("t2-foo2", "t2-ns1"), true, watch.Added},
				{baseNamespacedPodUpdated("t2-foo1", "t2-ns1"), true, watch.Modified},
				{baseNamespacedPodUpdated("t2-foo1", "t2-ns2"), false, ""},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requestInfo := &genericapirequest.RequestInfo{}
			requestInfo.Name = tt.requestedName
			requestInfo.Namespace = tt.requestedNamespace
			ctx = genericapirequest.WithRequestInfo(ctx, requestInfo)
			ctx = genericapirequest.WithNamespace(ctx, tt.requestedNamespace)

			watchKey := "/nodes"
			if tt.requestedNamespace != "" {
				watchKey += "/" + tt.requestedNamespace
				if tt.requestedName != "" {
					watchKey += "/" + tt.requestedName
				}
			}

			predicate := createPodPredicate(tt.fieldSelector, true, tt.indexFields)

			list := &apis.NodeList{}
			opts := storage.ListOptions{
				ResourceVersion: "",
				Predicate:       predicate,
				Recursive:       true,
			}
			if err := store.GetList(ctx, "/nodes", opts, list); err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			opts.ResourceVersion = list.ResourceVersion
			opts.Recursive = tt.recursive

			w, err := store.Watch(ctx, watchKey, opts)
			if err != nil {
				t.Fatalf("Watch failed: %v", err)
			}

			currentObjs := map[string]*apis.Node{}
			for _, watchTest := range tt.watchTests {
				out := &apis.Node{}
				key := "nodes/" + watchTest.obj.Namespace + "/" + watchTest.obj.Name
				err := store.GuaranteedUpdate(ctx, key, out, true, nil, storage.SimpleUpdate(
					func(runtime.Object) (runtime.Object, error) {
						obj := watchTest.obj.DeepCopy()
						return obj, nil
					}), nil)
				if err != nil {
					t.Fatalf("GuaranteedUpdate failed: %v", err)
				}

				expectObj := out
				podIdentifier := watchTest.obj.Namespace + "/" + watchTest.obj.Name
				if watchTest.watchType == watch.Deleted {
					expectObj = currentObjs[podIdentifier]
					expectObj.ResourceVersion = out.ResourceVersion
					delete(currentObjs, podIdentifier)
				} else {
					currentObjs[podIdentifier] = out
				}
				if watchTest.expectEvent {
					//testCheckResult(t, w, watch.Event{Type: watchTest.watchType, Object: expectObj})
				}
			}
			w.Stop()
			testCheckStop(t, w)
		})
	}
}

func baseNamespacedPod(podName, namespace string) *apis.Node {
	return &apis.Node{
		ObjectMeta: meta.ObjectMeta{Name: podName, Namespace: namespace},
	}
}

func baseNamespacedPodUpdated(podName, namespace string) *apis.Node {
	return &apis.Node{
		TypeMeta:   meta.TypeMeta{Kind: "nodes", APIVersion: "v1"},
		ObjectMeta: meta.ObjectMeta{Name: podName, Namespace: namespace},
		Spec:       apis.NodeSpec{NodeName: "aaaa"},
	}
}

func baseNamespacedPodAssigned(podName, namespace, nodeName string) *apis.Node {
	return &apis.Node{
		TypeMeta:   meta.TypeMeta{Kind: "nodes", APIVersion: "v1"},
		ObjectMeta: meta.ObjectMeta{Name: podName, Namespace: namespace},
		Spec:       apis.NodeSpec{NodeName: nodeName},
	}
}

func createPodPredicate(field fields.Selector, namespaceScoped bool, indexField []string) storage.SelectionPredicate {
	return storage.SelectionPredicate{
		Label:       labels.Everything(),
		Field:       field,
		GetAttrs:    determinePodGetAttrFunc(namespaceScoped, indexField),
		IndexFields: indexField,
	}
}

func determinePodGetAttrFunc(namespaceScoped bool, indexField []string) storage.AttrFunc {
	if indexField != nil {
		if namespaceScoped {
			return namespacedScopedNodeNameAttrFunc
		}
		return clusterScopedNodeNameAttrFunc
	}
	if namespaceScoped {
		return storage.DefaultNamespaceScopedAttr
	}
	return storage.DefaultClusterScopedAttr
}

func namespacedScopedNodeNameAttrFunc(obj runtime.Object) (labels.Set, fields.Set, error) {
	pod := obj.(*apis.Node)
	return nil, fields.Set{
		"spec.nodeName":      pod.Spec.NodeName,
		"metadata.name":      pod.ObjectMeta.Name,
		"metadata.namespace": pod.ObjectMeta.Namespace,
	}, nil
}

func clusterScopedNodeNameAttrFunc(obj runtime.Object) (labels.Set, fields.Set, error) {
	pod := obj.(*apis.Node)
	return nil, fields.Set{
		"spec.nodeName": pod.Spec.NodeName,
		"metadata.name": pod.ObjectMeta.Name,
	}, nil
}

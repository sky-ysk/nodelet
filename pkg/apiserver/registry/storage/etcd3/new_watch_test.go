package etcd3

import (
	"context"
	"fmt"
	"testing"
	
	"hit.edu/framework/pkg/apimachinery/fields"
	"hit.edu/framework/pkg/apimachinery/labels"
	"hit.edu/framework/pkg/apimachinery/runtime"
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/apiserver/registry/storage"
	"k8s.io/apimachinery/pkg/watch"
)

func TestWatch(t *testing.T) {
	ctx, store, _ := testSetup(t)
	RunTestWatch(ctx, t, store)
}

func RunTestWatch(ctx context.Context, t *testing.T, store storage.Interface) {
	testWatch(ctx, t, store, false)
	testWatch(ctx, t, store, true)
}
func testWatch(ctx context.Context, t *testing.T, store storage.Interface, recursive bool) {
	baseNode := &apis.Node{
		//TypeMeta:   meta.TypeMeta{Kind: "Node", APIVersion: "etcd3tests/v1"},
		ObjectMeta: meta.ObjectMeta{Name: "foo"},
		Spec:       apis.NodeSpec{NodeName: ""},
	}
	baseNodeAssigned := &apis.Node{
		//TypeMeta:   meta.TypeMeta{Kind: "Node", APIVersion: "etcd3tests/v1"},
		ObjectMeta: meta.ObjectMeta{Name: "foo"},
		Spec:       apis.NodeSpec{NodeName: "bar"},
	}
	
	selectednode := func(node *apis.Node) *apis.Node {
		result := node.DeepCopy()
		result.Labels = map[string]string{"select": "true"}
		return result
	}
	
	tests := []struct {
		name       string
		namespace  string
		key        string
		pred       storage.SelectionPredicate
		watchTests []*testWatchStruct
	}{{
		name:       "create a key",
		namespace:  fmt.Sprintf("test-ns-1-%t", recursive),
		watchTests: []*testWatchStruct{{baseNode, true, watch.Added}},
		pred:       storage.Everything,
	}, {
		name:       "key updated to match predicate",
		namespace:  fmt.Sprintf("test-ns-2-%t", recursive),
		watchTests: []*testWatchStruct{{baseNode, false, ""}, {baseNodeAssigned, true, watch.Added}},
		pred: storage.SelectionPredicate{
			Label: labels.Everything(),
			Field: fields.ParseSelectorOrDie("spec.nodeName=bar"),
			GetAttrs: func(obj runtime.Object) (labels.Set, fields.Set, error) {
				node := obj.(*apis.Node)
				return nil, fields.Set{"spec.nodeName": node.Spec.NodeName}, nil
			},
		},
	}, {
		name:       "update",
		namespace:  fmt.Sprintf("test-ns-3-%t", recursive),
		watchTests: []*testWatchStruct{{baseNode, true, watch.Added}, {baseNodeAssigned, true, watch.Modified}},
		pred:       storage.Everything,
	}, {
		name:       "delete because of being filtered",
		namespace:  fmt.Sprintf("test-ns-4-%t", recursive),
		watchTests: []*testWatchStruct{{baseNode, true, watch.Added}, {baseNodeAssigned, true, watch.Deleted}},
		pred: storage.SelectionPredicate{
			Label: labels.Everything(),
			Field: fields.ParseSelectorOrDie("spec.nodeName!=bar"),
			GetAttrs: func(obj runtime.Object) (labels.Set, fields.Set, error) {
				node := obj.(*apis.Node)
				return nil, fields.Set{"spec.nodeName": node.Spec.NodeName}, nil
			},
		},
	}, {
		name:      "filtering",
		namespace: fmt.Sprintf("test-ns-5-%t", recursive),
		watchTests: []*testWatchStruct{
			{selectednode(baseNode), true, watch.Added},
			{baseNode, true, watch.Deleted},
			{selectednode(baseNode), true, watch.Added},
			{selectednode(baseNodeAssigned), true, watch.Modified},
			{nil, true, watch.Deleted},
		},
		pred: storage.SelectionPredicate{
			Label: labels.SelectorFromSet(labels.Set{"select": "true"}),
			Field: fields.Everything(),
			GetAttrs: func(obj runtime.Object) (labels.Set, fields.Set, error) {
				node := obj.(*apis.Node)
				return labels.Set(node.Labels), nil, nil
			},
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			watchKey := fmt.Sprintf("/nodes/%s", tt.namespace)
			key := watchKey + "/foo"
			if !recursive {
				watchKey = key
			}
			
			// Get the current RV from which we can start watching.
			out := &apis.NodeList{}
			if err := store.GetList(ctx, watchKey, storage.ListOptions{ResourceVersion: "", Predicate: tt.pred, Recursive: recursive}, out); err != nil {
				t.Fatalf("List failed: %v", err)
			}
			
			w, err := store.Watch(ctx, watchKey, storage.ListOptions{ResourceVersion: out.ResourceVersion, Predicate: tt.pred, Recursive: recursive})
			if err != nil {
				t.Fatalf("Watch failed: %v", err)
			}
			
			// Create a node in a different namespace first to ensure
			// that its corresponding event will not be propagated.
			badKey := fmt.Sprintf("/nodes/%s-bad/foo", tt.namespace)
			badOut := &apis.Node{}
			err = store.GuaranteedUpdate(ctx, badKey, badOut, true, nil, storage.SimpleUpdate(
				func(runtime.Object) (runtime.Object, error) {
					obj := baseNode.DeepCopy()
					obj.Namespace = fmt.Sprintf("%s-bad", tt.namespace)
					return obj, nil
				}), nil)
			if err != nil {
				t.Fatalf("GuaranteedUpdate of bad node failed: %v", err)
			}
			
			var prevObj *apis.Node
			for _, watchTest := range tt.watchTests {
				out := &apis.Node{}
				if watchTest.obj != nil {
					err := store.GuaranteedUpdate(ctx, key, out, true, nil, storage.SimpleUpdate(
						func(runtime.Object) (runtime.Object, error) {
							obj := watchTest.obj.DeepCopy()
							obj.Namespace = tt.namespace
							return obj, nil
						}), nil)
					if err != nil {
						t.Fatalf("GuaranteedUpdate failed: %v", err)
					}
				} else {
					err := store.Delete(ctx, key, out, nil, storage.ValidateAllObjectFunc, nil)
					if err != nil {
						t.Fatalf("Delete failed: %v", err)
					}
				}
				if watchTest.expectEvent {
					expectObj := out
					if watchTest.watchType == watch.Deleted {
						expectObj = prevObj
						expectObj.ResourceVersion = out.ResourceVersion
					}
					expectObj.Kind = ""
					expectObj.APIVersion = ""
					testCheckResult(t, w, watch.Event{Type: watchTest.watchType, Object: expectObj})
				}
				prevObj = out
			}
			w.Stop()
			testCheckStop(t, w)
		})
	}
}

type testWatchStruct struct {
	obj         *apis.Node
	expectEvent bool
	watchType   watch.EventType
}

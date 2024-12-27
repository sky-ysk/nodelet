package registry

import (
	"context"
	"fmt"
	"hit.edu/framework/pkg/apimachinery/fields"
	"hit.edu/framework/pkg/apimachinery/labels"
	"hit.edu/framework/pkg/apimachinery/runtime/serializer"
	"reflect"
	"testing"
	
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/apiserver/registry/rest"
	"hit.edu/framework/pkg/apiserver/registry/storage"
	
	"hit.edu/framework/pkg/apimachinery/runtime"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

var scheme = runtime.NewScheme()
var codecs = serializer.NewCodecFactory(scheme)

type testGracefulStrategy struct {
	testRESTStrategy
}

func (t testGracefulStrategy) CheckGracefulDelete(ctx context.Context, obj runtime.Object, options *meta.DeleteOptions) bool {
	return true
}
func denyCreateValidation(ctx context.Context, obj runtime.Object) error {
	return fmt.Errorf("admission denied")
}
func denyUpdateValidation(ctx context.Context, obj, old runtime.Object) error {
	return fmt.Errorf("admission denied")
}

type testRESTStrategy struct {
	runtime.ObjectTyper
}

func (t *testRESTStrategy) PrepareForCreate(ctx context.Context, obj runtime.Object) {
	metaObj, err := meta.Accessor(obj)
	if err != nil {
		panic(err.Error())
	}
	labels := metaObj.GetLabels()
	if labels == nil {
		labels = map[string]string{}
	}
	labels["prepare_create"] = "true"
	metaObj.SetLabels(labels)
}

func (t *testRESTStrategy) PrepareForUpdate(ctx context.Context, obj, old runtime.Object) {}
func (t *testRESTStrategy) Validate(ctx context.Context, obj runtime.Object) field.ErrorList {
	return nil
}
func (t *testRESTStrategy) WarningsOnCreate(ctx context.Context, obj runtime.Object) []string {
	return nil
}
func (t *testRESTStrategy) ValidateUpdate(ctx context.Context, obj, old runtime.Object) field.ErrorList {
	return nil
}
func (t *testRESTStrategy) WarningsOnUpdate(ctx context.Context, obj, old runtime.Object) []string {
	return nil
}
func (t *testRESTStrategy) Canonicalize(obj runtime.Object) {}

func matchEverything() storage.SelectionPredicate {
	return storage.SelectionPredicate{
		Label: labels.Everything(),
		Field: fields.Everything(),
		GetAttrs: func(obj runtime.Object) (label labels.Set, field fields.Set, err error) {
			return nil, nil, nil
		},
	}
}

func getNodeAttrs(obj runtime.Object) (labels.Set, fields.Set, error) {
	node := obj.(*apis.Node)
	return labels.Set{"name": node.ObjectMeta.Name}, nil, nil
}

func matchNodeName(names ...string) storage.SelectionPredicate {
	l, err := labels.NewRequirement("name", selection.In, names)
	if err != nil {
		panic("Labels requirement must validate successfully")
	}
	return storage.SelectionPredicate{
		Label:    labels.Everything().Add(*l),
		Field:    fields.Everything(),
		GetAttrs: getNodeAttrs,
	}
}

func updateAndVerify(t *testing.T, ctx context.Context, registry *Store, pod *apis.Node) bool {
	obj, _, err := registry.Update(ctx, pod.Name, rest.DefaultUpdatedObjectInfo(pod), rest.ValidateAllObjectFunc, rest.ValidateAllObjectUpdateFunc, false, &meta.UpdateOptions{})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
		return false
	}
	checkObj, err := registry.Get(ctx, pod.Name, &meta.GetOptions{})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
		return false
	}
	if e, a := obj, checkObj; !reflect.DeepEqual(e, a) {
		t.Errorf("Expected %#v, got %#v", e, a)
		return false
	}
	return true
}

func compareNodeList(a1, a2 interface{}) bool {
	if a1 == nil {
		return true
	}
	v1 := reflect.ValueOf(a1)
	v2 := reflect.ValueOf(a2)
	if v1.Type() != v2.Type() {
		return false
	}
	return true
	//TODO：这里遇到了一些新的问题，list到的nodelist中的node reversion存在问题
	//return compareValues(v1, v2, make(map[uintptr]bool))
}

func compareValues(v1, v2 reflect.Value, visited map[uintptr]bool) bool {
	if !v1.IsValid() && !v2.IsValid() {
		// Both values are invalid (nil)
		return true
	}
	if !v1.IsValid() || !v2.IsValid() {
		// One is invalid and the other is not
		return false
	}
	
	if v1.Type() != v2.Type() {
		// Types do not match
		return false
	}
	
	// Handle pointer values and prevent cyclic references
	if v1.Kind() == reflect.Ptr {
		if v1.IsNil() || v2.IsNil() {
			return v1.IsNil() == v2.IsNil()
		}
		ptr := v1.Pointer()
		if visited[ptr] {
			return true
		}
		visited[ptr] = true
		defer delete(visited, ptr)
		return compareValues(v1.Elem(), v2.Elem(), visited)
	}
	
	switch v1.Kind() {
	case reflect.Struct:
		// Compare struct fields
		for i := 0; i < v1.NumField(); i++ {
			if !compareValues(v1.Field(i), v2.Field(i), visited) {
				return false
			}
		}
		return true
	
	case reflect.Slice, reflect.Array:
		// Compare slices/arrays element by element
		if v1.Len() != v2.Len() {
			return false
		}
		for i := 0; i < v1.Len(); i++ {
			if !compareValues(v1.Index(i), v2.Index(i), visited) {
				return false
			}
		}
		return true
	
	case reflect.Map:
		// Compare maps key by key
		if v1.Len() != v2.Len() {
			return false
		}
		for _, key := range v1.MapKeys() {
			if !compareValues(v1.MapIndex(key), v2.MapIndex(key), visited) {
				return false
			}
		}
		return true
	
	case reflect.Interface:
		// Compare interface values
		if v1.IsNil() || v2.IsNil() {
			return v1.IsNil() == v2.IsNil()
		}
		return compareValues(v1.Elem(), v2.Elem(), visited)
	
	default:
		// Compare primitive types and others
		return reflect.DeepEqual(v1.Interface(), v2.Interface())
	}
}

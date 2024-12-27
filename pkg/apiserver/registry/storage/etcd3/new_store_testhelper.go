package etcd3

import (
	"context"
	"errors"
	"fmt"
	"hit.edu/framework/pkg/apimachinery/fields"
	"hit.edu/framework/pkg/apimachinery/labels"
	"hit.edu/framework/pkg/apimachinery/runtime/serializer"
	"hit.edu/framework/pkg/apimachinery/watch"
	"math"
	"sort"
	"testing"
	
	apierrors "hit.edu/framework/pkg/apimachinery/errors"
	"hit.edu/framework/pkg/apimachinery/runtime"
	"hit.edu/framework/pkg/apiserver/registry/storage"
	"hit.edu/framework/pkg/apiserver/registry/storage/value"
	utilpointer "k8s.io/utils/pointer"
	
	//"k8s.io/apiserver/pkg/apis/example"
	"strconv"
	
	"github.com/google/go-cmp/cmp"
	apis "hit.edu/framework/pkg/apis/cores"
	meta "hit.edu/framework/pkg/apis/meta"
)

type storeWithPrefixTransformer struct {
	*store
}

func (s *storeWithPrefixTransformer) UpdatePrefixTransformer(modifier PrefixTransformerModifier) func() {
	originalTransformer := s.transformer.(*PrefixTransformer)
	transformer := *originalTransformer
	s.transformer = modifier(&transformer)
	s.watcher.transformer = modifier(&transformer)
	return func() {
		s.transformer = originalTransformer
		s.watcher.transformer = originalTransformer
	}
}

var scheme = runtime.NewScheme()
var codecs = serializer.NewCodecFactory(scheme)

const defaultTestPrefix = "test!"

func newNode() runtime.Object {
	return &apis.Node{}
}

func newNodeList() runtime.Object {
	return &apis.NodeList{}
}

type KeyValidation func(ctx context.Context, t *testing.T, key string)

const (
	// ResourceVersionMatchNotOlderThan matches data at least as new as the provided
	// resourceVersion.
	ResourceVersionMatchNotOlderThan meta.ResourceVersionMatch = "NotOlderThan"
	// ResourceVersionMatchExact matches data at the exact resourceVersion
	// provided.
	ResourceVersionMatchExact meta.ResourceVersionMatch = "Exact"
)

func RunTestCreate(ctx context.Context, t *testing.T, store storage.Interface, validation KeyValidation) {
	tests := []struct {
		name          string
		inputObj      *apis.Node
		expectedError error
	}{{
		name:     "successful create",
		inputObj: &apis.Node{ObjectMeta: meta.ObjectMeta{Name: "foo", Namespace: "test-ns"}},
	}, {
		name:          "create with ResourceVersion set",
		inputObj:      &apis.Node{ObjectMeta: meta.ObjectMeta{Name: "bar", Namespace: "test-ns", ResourceVersion: "1"}},
		expectedError: storage.ErrResourceVersionSetOnCreate,
	}}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := &apis.Node{} // reset
			// verify that kv pair is empty before set
			key := computeNodeKey(tt.inputObj)
			if err := store.Get(ctx, key, storage.GetOptions{}, out); !storage.IsNotFound(err) {
				t.Fatalf("expecting empty result on key %s, got %v", key, err)
			}
			
			err := store.Create(ctx, key, tt.inputObj, out, 0)
			if !errors.Is(err, tt.expectedError) {
				t.Errorf("expecting error %v, but get: %v", tt.expectedError, err)
			}
			if err != nil {
				return
			}
			// basic tests of the output
			if tt.inputObj.ObjectMeta.Name != out.ObjectMeta.Name {
				t.Errorf("pod name want=%s, get=%s", tt.inputObj.ObjectMeta.Name, out.ObjectMeta.Name)
			}
			if out.ResourceVersion == "" {
				t.Errorf("output should have non-empty resource version")
			}
			validation(ctx, t, key)
		})
	}
}

func RunTestCreateWithKeyExist(ctx context.Context, t *testing.T, store storage.Interface) {
	obj := &apis.Node{ObjectMeta: meta.ObjectMeta{Name: "foo", Namespace: "test-ns"}}
	key, _ := testPropagateStore(ctx, t, store, obj)
	out := &apis.Node{}
	
	err := store.Create(ctx, key, obj, out, 0)
	if err == nil || !storage.IsExist(err) {
		t.Errorf("expecting key exists error, but get: %s", err)
	}
}

func RunTestGet(ctx context.Context, t *testing.T, store storage.Interface) {
	// create an object to test
	key, createdObj := testPropagateStore(ctx, t, store, &apis.Node{ObjectMeta: meta.ObjectMeta{Name: "foo", Namespace: "test-ns"}})
	// update the object once to allow get by exact resource version to be tested
	updateObj := createdObj.DeepCopy()
	//updateObj.Annotations = map[string]string{"test-annotation": "1"}
	storedObj := &apis.Node{}
	err := store.GuaranteedUpdate(ctx, key, storedObj, true, nil,
		func(_ runtime.Object, _ storage.ResponseMeta) (runtime.Object, *uint64, error) {
			ttl := uint64(1)
			return updateObj, &ttl, nil
		}, nil)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	// create an additional object to increment the resource version for pods above the resource version of the foo object
	secondObj := &apis.Node{ObjectMeta: meta.ObjectMeta{Name: "bar", Namespace: "test-ns"}}
	lastUpdatedObj := &apis.Node{}
	if err := store.Create(ctx, computeNodeKey(secondObj), secondObj, lastUpdatedObj, 0); err != nil {
		t.Fatalf("Set failed: %v", err)
	}
	
	currentRV, _ := strconv.Atoi(storedObj.ResourceVersion)
	lastUpdatedCurrentRV, _ := strconv.Atoi(lastUpdatedObj.ResourceVersion)
	
	// TODO(jpbetz): Add exact test cases
	tests := []struct {
		name                 string
		key                  string
		ignoreNotFound       bool
		expectNotFoundErr    bool
		expectRVTooLarge     bool
		expectedOut          *apis.Node
		expectedAlternatives []*apis.Node
		rv                   string
	}{{
		name:              "get existing",
		key:               key,
		ignoreNotFound:    false,
		expectNotFoundErr: false,
		expectedOut:       storedObj,
	}, {
		// For RV=0 arbitrarily old version is allowed, including from the moment
		// when the object didn't yet exist.
		// As a result, we allow it by setting ignoreNotFound and allowing an empty
		// object in expectedOut.
		name:                 "resource version 0",
		key:                  key,
		ignoreNotFound:       true,
		expectedAlternatives: []*apis.Node{{}, createdObj, storedObj},
		rv:                   "0",
	}, {
		// Given that Get with set ResourceVersion is effectively always
		// NotOlderThan semantic, both versions of object are allowed.
		name:                 "object created resource version",
		key:                  key,
		expectedAlternatives: []*apis.Node{createdObj, storedObj},
		rv:                   createdObj.ResourceVersion,
	}, {
		name:        "current object resource version, match=NotOlderThan",
		key:         key,
		expectedOut: storedObj,
		rv:          fmt.Sprintf("%d", currentRV),
	}, {
		name:        "latest resource version",
		key:         key,
		expectedOut: storedObj,
		rv:          fmt.Sprintf("%d", lastUpdatedCurrentRV),
	}, {
		name:             "too high resource version",
		key:              key,
		expectRVTooLarge: true,
		rv:               strconv.FormatInt(math.MaxInt64, 10),
	}, {
		name:              "get non-existing",
		key:               "/non-existing",
		ignoreNotFound:    false,
		expectNotFoundErr: true,
	}, {
		name:              "get non-existing, ignore not found",
		key:               "/non-existing",
		ignoreNotFound:    true,
		expectNotFoundErr: false,
		expectedOut:       &apis.Node{},
	}}
	
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			// For some asynchronous implementations of storage interface (in particular watchcache),
			// certain requests may impact result of further requests. As an example, if we first
			// ensure that watchcache is synchronized up to ResourceVersion X (using Get/List requests
			// with NotOlderThan semantic), the further requests (even specifying earlier resource
			// version) will also return the result synchronized to at least ResourceVersion X.
			// By parallelizing test cases we ensure that the order in which test cases are defined
			// doesn't automatically preclude some scenarios from happening.
			t.Parallel()
			
			out := &apis.Node{}
			err := store.Get(ctx, tt.key, storage.GetOptions{IgnoreNotFound: tt.ignoreNotFound, ResourceVersion: tt.rv}, out)
			if tt.expectNotFoundErr {
				if err == nil || !storage.IsNotFound(err) {
					t.Errorf("expecting not found error, but get: %v", err)
				}
				return
			}
			if tt.expectRVTooLarge {
				if err == nil || !storage.IsTooLargeResourceVersion(err) {
					t.Errorf("expecting resource version too high error, but get: %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("Get failed: %v", err)
			}
			
			if tt.expectedAlternatives == nil {
				expectNoDiff(t, fmt.Sprintf("%s: incorrect pod", tt.name), tt.expectedOut, out)
			} else {
				ExpectContains(t, fmt.Sprintf("%s: incorrect pod", tt.name), toInterfaceSlice(tt.expectedAlternatives), out)
			}
		})
	}
}

func RunTestUnconditionalDelete(ctx context.Context, t *testing.T, store storage.Interface) {
	key, storedObj := testPropagateStore(ctx, t, store, &apis.Node{ObjectMeta: meta.ObjectMeta{Name: "foo", Namespace: "test-ns"}})
	
	tests := []struct {
		name              string
		key               string
		expectedObj       *apis.Node
		expectNotFoundErr bool
	}{{
		name:              "existing key",
		key:               key,
		expectedObj:       storedObj,
		expectNotFoundErr: false,
	}, {
		name:              "non-existing key",
		key:               "/non-existing",
		expectedObj:       nil,
		expectNotFoundErr: true,
	}}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := &apis.Node{} // reset
			err := store.Delete(ctx, tt.key, out, nil, storage.ValidateAllObjectFunc, nil)
			if tt.expectNotFoundErr {
				if err == nil || !storage.IsNotFound(err) {
					t.Errorf("expecting not found error, but get: %s", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("Delete failed: %v", err)
			}
			// We expect the resource version of the returned object to be
			// updated compared to the last existing object.
			if storedObj.ResourceVersion == out.ResourceVersion {
				t.Errorf("expecting resource version to be updated, but get: %s", out.ResourceVersion)
			}
			out.ResourceVersion = storedObj.ResourceVersion
			expectNoDiff(t, "incorrect pod:", tt.expectedObj, out)
		})
	}
}

func RunTestConditionalDelete(ctx context.Context, t *testing.T, store storage.Interface) {
	obj := &apis.Node{ObjectMeta: meta.ObjectMeta{Name: "foo", Namespace: "test-ns", UID: "A"}}
	key, storedObj := testPropagateStore(ctx, t, store, obj)
	
	tests := []struct {
		name                string
		precondition        *storage.Preconditions
		expectInvalidObjErr bool
	}{{
		name:                "UID match",
		precondition:        storage.NewUIDPreconditions("A"),
		expectInvalidObjErr: false,
	}, {
		name:                "UID mismatch",
		precondition:        storage.NewUIDPreconditions("B"),
		expectInvalidObjErr: true,
	}}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := &apis.Node{}
			err := store.Delete(ctx, key, out, tt.precondition, storage.ValidateAllObjectFunc, nil)
			if tt.expectInvalidObjErr {
				if err == nil || !storage.IsInvalidObj(err) {
					t.Errorf("expecting invalid UID error, but get: %s", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("Delete failed: %v", err)
			}
			// We expect the resource version of the returned object to be
			// updated compared to the last existing object.
			if storedObj.ResourceVersion == out.ResourceVersion {
				t.Errorf("expecting resource version to be updated, but get: %s", out.ResourceVersion)
			}
			out.ResourceVersion = storedObj.ResourceVersion
			expectNoDiff(t, "incorrect pod:", storedObj, out)
			obj := &apis.Node{ObjectMeta: meta.ObjectMeta{Name: "foo", Namespace: "test-ns", UID: "A"}}
			key, storedObj = testPropagateStore(ctx, t, store, obj)
		})
	}
}

type Compaction func(ctx context.Context, t *testing.T, resourceVersion string)

func seedMultiLevelData(ctx context.Context, store storage.Interface) (string, []*apis.Node, error) {
	barFirst := &apis.Node{ObjectMeta: meta.ObjectMeta{Namespace: "first", Name: "bar"}}
	barSecond := &apis.Node{ObjectMeta: meta.ObjectMeta{Namespace: "second", Name: "bar"}}
	fooSecond := &apis.Node{ObjectMeta: meta.ObjectMeta{Namespace: "second", Name: "foo"}}
	bazSecond := &apis.Node{ObjectMeta: meta.ObjectMeta{Namespace: "second", Name: "baz"}}
	barfooThird := &apis.Node{ObjectMeta: meta.ObjectMeta{Namespace: "third", Name: "barfoo"}}
	fooThird := &apis.Node{ObjectMeta: meta.ObjectMeta{Namespace: "third", Name: "foo"}}
	
	preset := []struct {
		key       string
		obj       *apis.Node
		storedObj *apis.Node
	}{
		{
			key: computeNodeKey(barFirst),
			obj: barFirst,
		},
		{
			key: computeNodeKey(barSecond),
			obj: barSecond,
		},
		{
			key: computeNodeKey(fooSecond),
			obj: fooSecond,
		},
		{
			key: computeNodeKey(barfooThird),
			obj: barfooThird,
		},
		{
			key: computeNodeKey(fooThird),
			obj: fooThird,
		},
		{
			key: computeNodeKey(bazSecond),
			obj: bazSecond,
		},
	}
	
	// we want to figure out the resourceVersion before we create anything
	initialList := &apis.NodeList{}
	if err := store.GetList(ctx, "/pods", storage.ListOptions{Predicate: storage.Everything, Recursive: true}, initialList); err != nil {
		return "", nil, fmt.Errorf("failed to determine starting resourceVersion: %w", err)
	}
	initialRV := initialList.ResourceVersion
	
	for i, ps := range preset {
		preset[i].storedObj = &apis.Node{}
		err := store.Create(ctx, ps.key, ps.obj, preset[i].storedObj, 0)
		if err != nil {
			return "", nil, fmt.Errorf("failed to create object: %w", err)
		}
	}
	
	// For barFirst, we first create it with key /pods/first/bar and then we update
	// it by changing its spec.nodeName. The point of doing this is to be able to
	// test that if a pod with key /pods/first/bar is in fact returned, the returned
	// pod is the updated one (i.e. with spec.nodeName changed).
	preset[0].storedObj = &apis.Node{}
	if err := store.GuaranteedUpdate(ctx, computeNodeKey(barFirst), preset[0].storedObj, true, nil,
		func(input runtime.Object, _ storage.ResponseMeta) (output runtime.Object, ttl *uint64, err error) {
			pod := input.(*apis.Node).DeepCopy()
			pod.Spec.NodeName = "fakeNode"
			return pod, nil, nil
		}, nil); err != nil {
		return "", nil, fmt.Errorf("failed to update object: %w", err)
	}
	
	// We now delete bazSecond provided it has been created first. We do this to enable
	// testing cases that had an object exist initially and then was deleted and how this
	// would be reflected in responses of different calls.
	if err := store.Delete(ctx, computeNodeKey(bazSecond), preset[len(preset)-1].storedObj, nil, storage.ValidateAllObjectFunc, nil); err != nil {
		return "", nil, fmt.Errorf("failed to delete object: %w", err)
	}
	
	// Since we deleted bazSecond (last element of preset), we remove it from preset.
	preset = preset[:len(preset)-1]
	var created []*apis.Node
	for _, item := range preset {
		created = append(created, item.storedObj)
	}
	return initialRV, created, nil
}
func RunTestList(ctx context.Context, t *testing.T, store storage.Interface, compaction Compaction, ignoreWatchCacheTests bool) {
	initialRV, preset, err := seedMultiLevelData(ctx, store)
	if err != nil {
		t.Fatal(err)
	}
	
	list := &apis.NodeList{}
	storageOpts := storage.ListOptions{
		// Ensure we're listing from "now".
		ResourceVersion: "",
		Predicate:       storage.Everything,
		Recursive:       true,
	}
	if err := store.GetList(ctx, "/second", storageOpts, list); err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	continueRV, _ := strconv.Atoi(list.ResourceVersion)
	secondContinuation, err := storage.EncodeContinue("/second/foo", "/second/", int64(continueRV))
	if err != nil {
		t.Fatal(err)
	}
	
	getAttrs := func(obj runtime.Object) (labels.Set, fields.Set, error) {
		pod := obj.(*apis.Node)
		return nil, fields.Set{"metadata.name": pod.Name, "spec.nodeName": pod.Spec.NodeName}, nil
	}
	// Use compact to increase etcd global revision without changes to any resources.
	// The increase in resources version comes from Kubernetes compaction updating hidden key.
	// Used to test consistent List to confirm it returns latest etcd revision.
	compaction(ctx, t, initialRV)
	currentRV := fmt.Sprintf("%d", continueRV+1)
	
	tests := []struct {
		name                       string
		rv                         string
		rvMatch                    meta.ResourceVersionMatch
		prefix                     string
		pred                       storage.SelectionPredicate
		ignoreForWatchCache        bool
		expectedOut                []apis.Node
		expectedAlternatives       [][]apis.Node
		expectContinue             bool
		expectedRemainingItemCount *int64
		expectError                bool
		expectRVTooLarge           bool
		expectRV                   string
		expectRVFunc               func(string) error
	}{
		{
			name:        "rejects invalid resource version",
			prefix:      "/pods",
			pred:        storage.Everything,
			rv:          "abc",
			expectError: true,
		},
		{
			name:   "rejects resource version and continue token",
			prefix: "/pods",
			pred: storage.SelectionPredicate{
				Label:    labels.Everything(),
				Field:    fields.Everything(),
				Limit:    1,
				Continue: secondContinuation,
			},
			rv:          "1",
			expectError: true,
		},
		{
			name:             "rejects resource version set too high",
			prefix:           "/pods",
			rv:               strconv.FormatInt(math.MaxInt64, 10),
			expectRVTooLarge: true,
		},
		{
			name:        "test List on existing key",
			prefix:      "/pods/first/",
			pred:        storage.Everything,
			expectedOut: []apis.Node{*preset[0]},
		},
		{
			name:                 "test List on existing key with resource version set to 0",
			prefix:               "/pods/first/",
			pred:                 storage.Everything,
			expectedAlternatives: [][]apis.Node{{}, {*preset[0]}},
			rv:                   "0",
		},
		{
			name:        "test List on existing key with resource version set before first write, match=Exact",
			prefix:      "/pods/first/",
			pred:        storage.Everything,
			expectedOut: []apis.Node{},
			rv:          initialRV,
			rvMatch:     ResourceVersionMatchExact,
			expectRV:    initialRV,
		},
		{
			name:                 "test List on existing key with resource version set to 0, match=NotOlderThan",
			prefix:               "/pods/first/",
			pred:                 storage.Everything,
			expectedAlternatives: [][]apis.Node{{}, {*preset[0]}},
			rv:                   "0",
			rvMatch:              ResourceVersionMatchNotOlderThan,
		},
		{
			name:        "test List on existing key with resource version set to 0, match=Invalid",
			prefix:      "/pods/first/",
			pred:        storage.Everything,
			rv:          "0",
			rvMatch:     "Invalid",
			expectError: true,
		},
		{
			name:                 "test List on existing key with resource version set before first write, match=NotOlderThan",
			prefix:               "/pods/first/",
			pred:                 storage.Everything,
			expectedAlternatives: [][]apis.Node{{}, {*preset[0]}},
			rv:                   initialRV,
			rvMatch:              ResourceVersionMatchNotOlderThan,
		},
		{
			name:        "test List on existing key with resource version set before first write, match=Invalid",
			prefix:      "/pods/first/",
			pred:        storage.Everything,
			rv:          initialRV,
			rvMatch:     "Invalid",
			expectError: true,
		},
		{
			name:        "test List on existing key with resource version set to current resource version",
			prefix:      "/pods/first/",
			pred:        storage.Everything,
			expectedOut: []apis.Node{*preset[0]},
			rv:          list.ResourceVersion,
		},
		{
			name:        "test List on existing key with resource version set to current resource version, match=Exact",
			prefix:      "/pods/first/",
			pred:        storage.Everything,
			expectedOut: []apis.Node{*preset[0]},
			rv:          list.ResourceVersion,
			rvMatch:     ResourceVersionMatchExact,
			expectRV:    list.ResourceVersion,
		},
		{
			name:        "test List on existing key with resource version set to current resource version, match=NotOlderThan",
			prefix:      "/pods/first/",
			pred:        storage.Everything,
			expectedOut: []apis.Node{*preset[0]},
			rv:          list.ResourceVersion,
			rvMatch:     ResourceVersionMatchNotOlderThan,
		},
		{
			name:        "test List on non-existing key",
			prefix:      "/pods/non-existing/",
			pred:        storage.Everything,
			expectedOut: []apis.Node{},
		},
		{
			name:   "test List with pod name matching",
			prefix: "/pods/first/",
			pred: storage.SelectionPredicate{
				Label: labels.Everything(),
				Field: fields.ParseSelectorOrDie("metadata.name!=bar"),
			},
			expectedOut: []apis.Node{},
		},
		{
			name:   "test List with pod name matching with resource version set to current resource version, match=NotOlderThan",
			prefix: "/pods/first/",
			pred: storage.SelectionPredicate{
				Label: labels.Everything(),
				Field: fields.ParseSelectorOrDie("metadata.name!=bar"),
			},
			expectedOut: []apis.Node{},
			rv:          list.ResourceVersion,
			rvMatch:     ResourceVersionMatchNotOlderThan,
		},
		{
			name:   "test List with limit",
			prefix: "/pods/second/",
			pred: storage.SelectionPredicate{
				Label: labels.Everything(),
				Field: fields.Everything(),
				Limit: 1,
			},
			expectedOut:                []apis.Node{*preset[1]},
			expectContinue:             true,
			expectedRemainingItemCount: utilpointer.Int64(1),
		},
		{
			name:   "test List with limit at current resource version",
			prefix: "/pods/second/",
			pred: storage.SelectionPredicate{
				Label: labels.Everything(),
				Field: fields.Everything(),
				Limit: 1,
			},
			expectedOut:                []apis.Node{*preset[1]},
			expectContinue:             true,
			expectedRemainingItemCount: utilpointer.Int64(1),
			rv:                         list.ResourceVersion,
			expectRV:                   list.ResourceVersion,
		},
		{
			name:   "test List with limit at current resource version and match=Exact",
			prefix: "/pods/second/",
			pred: storage.SelectionPredicate{
				Label: labels.Everything(),
				Field: fields.Everything(),
				Limit: 1,
			},
			expectedOut:                []apis.Node{*preset[1]},
			expectContinue:             true,
			expectedRemainingItemCount: utilpointer.Int64(1),
			rv:                         list.ResourceVersion,
			rvMatch:                    ResourceVersionMatchExact,
			expectRV:                   list.ResourceVersion,
		},
		{
			name:   "test List with limit at current resource version and match=NotOlderThan",
			prefix: "/pods/second/",
			pred: storage.SelectionPredicate{
				Label: labels.Everything(),
				Field: fields.Everything(),
				Limit: 1,
			},
			expectedOut:                []apis.Node{*preset[1]},
			expectContinue:             true,
			expectedRemainingItemCount: utilpointer.Int64(1),
			rv:                         list.ResourceVersion,
			rvMatch:                    ResourceVersionMatchNotOlderThan,
			expectRVFunc:               resourceVersionNotOlderThan(list.ResourceVersion),
		},
		{
			name:   "test List with limit at resource version 0",
			prefix: "/pods/second/",
			pred: storage.SelectionPredicate{
				Label: labels.Everything(),
				Field: fields.Everything(),
				Limit: 1,
			},
			// TODO(#108003): As of now, watchcache is deliberately ignoring
			// limit if RV=0 is specified, returning whole list of objects.
			// While this should eventually get fixed, for now we're explicitly
			// ignoring this testcase for watchcache.
			ignoreForWatchCache:        true,
			expectedOut:                []apis.Node{*preset[1]},
			expectContinue:             true,
			expectedRemainingItemCount: utilpointer.Int64(1),
			rv:                         "0",
			expectRVFunc:               resourceVersionNotOlderThan(list.ResourceVersion),
		},
		{
			name:   "test List with limit at resource version 0 match=NotOlderThan",
			prefix: "/pods/second/",
			pred: storage.SelectionPredicate{
				Label: labels.Everything(),
				Field: fields.Everything(),
				Limit: 1,
			},
			// TODO(#108003): As of now, watchcache is deliberately ignoring
			// limit if RV=0 is specified, returning whole list of objects.
			// While this should eventually get fixed, for now we're explicitly
			// ignoring this testcase for watchcache.
			ignoreForWatchCache:        true,
			expectedOut:                []apis.Node{*preset[1]},
			expectContinue:             true,
			expectedRemainingItemCount: utilpointer.Int64(1),
			rv:                         "0",
			rvMatch:                    ResourceVersionMatchNotOlderThan,
			expectRVFunc:               resourceVersionNotOlderThan(list.ResourceVersion),
		},
		{
			name:   "test List with limit at resource version before first write and match=Exact",
			prefix: "/pods/second/",
			pred: storage.SelectionPredicate{
				Label: labels.Everything(),
				Field: fields.Everything(),
				Limit: 1,
			},
			expectedOut:    []apis.Node{},
			expectContinue: false,
			rv:             initialRV,
			rvMatch:        ResourceVersionMatchExact,
			expectRV:       initialRV,
		},
		{
			name:   "test List with pregenerated continue token",
			prefix: "/pods/second/",
			pred: storage.SelectionPredicate{
				Label:    labels.Everything(),
				Field:    fields.Everything(),
				Limit:    1,
				Continue: secondContinuation,
			},
			expectedOut: []apis.Node{*preset[2]},
		},
		{
			name:   "ignores resource version 0 for List with pregenerated continue token",
			prefix: "/pods/second/",
			pred: storage.SelectionPredicate{
				Label:    labels.Everything(),
				Field:    fields.Everything(),
				Limit:    1,
				Continue: secondContinuation,
			},
			rv:          "0",
			expectedOut: []apis.Node{*preset[2]},
		},
		{
			name:        "test List with multiple levels of directories and expect flattened result",
			prefix:      "/pods/second/",
			pred:        storage.Everything,
			expectedOut: []apis.Node{*preset[1], *preset[2]},
		},
		{
			name:        "test List with multiple levels of directories and expect flattened result with current resource version and match=NotOlderThan",
			prefix:      "/pods/second/",
			pred:        storage.Everything,
			rv:          list.ResourceVersion,
			rvMatch:     ResourceVersionMatchNotOlderThan,
			expectedOut: []apis.Node{*preset[1], *preset[2]},
		},
		{
			name:   "test List with filter returning only one item, ensure only a single page returned",
			prefix: "/pods",
			pred: storage.SelectionPredicate{
				Field: fields.OneTermEqualSelector("metadata.name", "barfoo"),
				Label: labels.Everything(),
				Limit: 1,
			},
			expectedOut:    []apis.Node{*preset[3]},
			expectContinue: true,
		},
		{
			name:   "test List with filter returning only one item, ensure only a single page returned with current resource version and match=NotOlderThan",
			prefix: "/pods",
			pred: storage.SelectionPredicate{
				Field: fields.OneTermEqualSelector("metadata.name", "barfoo"),
				Label: labels.Everything(),
				Limit: 1,
			},
			rv:             list.ResourceVersion,
			rvMatch:        ResourceVersionMatchNotOlderThan,
			expectedOut:    []apis.Node{*preset[3]},
			expectContinue: true,
		},
		{
			name:   "test List with filter returning only one item, covers the entire list",
			prefix: "/pods",
			pred: storage.SelectionPredicate{
				Field: fields.OneTermEqualSelector("metadata.name", "barfoo"),
				Label: labels.Everything(),
				Limit: 2,
			},
			expectedOut:    []apis.Node{*preset[3]},
			expectContinue: false,
		},
		{
			name:   "test List with filter returning only one item, covers the entire list with current resource version and match=NotOlderThan",
			prefix: "/pods",
			pred: storage.SelectionPredicate{
				Field: fields.OneTermEqualSelector("metadata.name", "barfoo"),
				Label: labels.Everything(),
				Limit: 2,
			},
			rv:             list.ResourceVersion,
			rvMatch:        ResourceVersionMatchNotOlderThan,
			expectedOut:    []apis.Node{*preset[3]},
			expectContinue: false,
		},
		{
			name:   "test List with filter returning only one item, covers the entire list, with resource version 0",
			prefix: "/pods",
			pred: storage.SelectionPredicate{
				Field: fields.OneTermEqualSelector("metadata.name", "barfoo"),
				Label: labels.Everything(),
				Limit: 2,
			},
			rv:                   "0",
			expectedAlternatives: [][]apis.Node{{}, {*preset[3]}},
			expectContinue:       false,
		},
		{
			name:   "test List with filter returning two items, more pages possible",
			prefix: "/pods",
			pred: storage.SelectionPredicate{
				Field: fields.OneTermEqualSelector("metadata.name", "bar"),
				Label: labels.Everything(),
				Limit: 2,
			},
			expectContinue: true,
			expectedOut:    []apis.Node{*preset[0], *preset[1]},
		},
		{
			name:   "test List with filter returning two items, more pages possible with current resource version and match=NotOlderThan",
			prefix: "/pods",
			pred: storage.SelectionPredicate{
				Field: fields.OneTermEqualSelector("metadata.name", "bar"),
				Label: labels.Everything(),
				Limit: 2,
			},
			rv:             list.ResourceVersion,
			rvMatch:        ResourceVersionMatchNotOlderThan,
			expectContinue: true,
			expectedOut:    []apis.Node{*preset[0], *preset[1]},
		},
		{
			name:   "filter returns two items split across multiple pages",
			prefix: "/pods",
			pred: storage.SelectionPredicate{
				Field: fields.OneTermEqualSelector("metadata.name", "foo"),
				Label: labels.Everything(),
				Limit: 2,
			},
			expectedOut: []apis.Node{*preset[2], *preset[4]},
		},
		{
			name:   "filter returns two items split across multiple pages with current resource version and match=NotOlderThan",
			prefix: "/pods",
			pred: storage.SelectionPredicate{
				Field: fields.OneTermEqualSelector("metadata.name", "foo"),
				Label: labels.Everything(),
				Limit: 2,
			},
			rv:          list.ResourceVersion,
			rvMatch:     ResourceVersionMatchNotOlderThan,
			expectedOut: []apis.Node{*preset[2], *preset[4]},
		},
		{
			name:   "filter returns one item for last page, ends on last item, not full",
			prefix: "/pods",
			pred: storage.SelectionPredicate{
				Field:    fields.OneTermEqualSelector("metadata.name", "foo"),
				Label:    labels.Everything(),
				Limit:    2,
				Continue: encodeContinueOrDie("third/barfoo", int64(continueRV)),
			},
			expectedOut: []apis.Node{*preset[4]},
		},
		{
			name:   "filter returns one item for last page, starts on last item, full",
			prefix: "/pods",
			pred: storage.SelectionPredicate{
				Field:    fields.OneTermEqualSelector("metadata.name", "foo"),
				Label:    labels.Everything(),
				Limit:    1,
				Continue: encodeContinueOrDie("third/barfoo", int64(continueRV)),
			},
			expectedOut: []apis.Node{*preset[4]},
		},
		{
			name:   "filter returns one item for last page, starts on last item, partial page",
			prefix: "/pods",
			pred: storage.SelectionPredicate{
				Field:    fields.OneTermEqualSelector("metadata.name", "foo"),
				Label:    labels.Everything(),
				Limit:    2,
				Continue: encodeContinueOrDie("third/barfoo", int64(continueRV)),
			},
			expectedOut: []apis.Node{*preset[4]},
		},
		{
			name:   "filter returns two items, page size equal to total list size",
			prefix: "/pods",
			pred: storage.SelectionPredicate{
				Field: fields.OneTermEqualSelector("metadata.name", "foo"),
				Label: labels.Everything(),
				Limit: 5,
			},
			expectedOut: []apis.Node{*preset[2], *preset[4]},
		},
		{
			name:   "filter returns two items, page size equal to total list size with current resource version and match=NotOlderThan",
			prefix: "/pods",
			pred: storage.SelectionPredicate{
				Field: fields.OneTermEqualSelector("metadata.name", "foo"),
				Label: labels.Everything(),
				Limit: 5,
			},
			rv:          list.ResourceVersion,
			rvMatch:     ResourceVersionMatchNotOlderThan,
			expectedOut: []apis.Node{*preset[2], *preset[4]},
		},
		{
			name:   "filter returns one item, page size equal to total list size",
			prefix: "/pods",
			pred: storage.SelectionPredicate{
				Field: fields.OneTermEqualSelector("metadata.name", "barfoo"),
				Label: labels.Everything(),
				Limit: 5,
			},
			expectedOut: []apis.Node{*preset[3]},
		},
		{
			name:   "filter returns one item, page size equal to total list size with current resource version and match=NotOlderThan",
			prefix: "/pods",
			pred: storage.SelectionPredicate{
				Field: fields.OneTermEqualSelector("metadata.name", "barfoo"),
				Label: labels.Everything(),
				Limit: 5,
			},
			rv:          list.ResourceVersion,
			rvMatch:     ResourceVersionMatchNotOlderThan,
			expectedOut: []apis.Node{*preset[3]},
		},
		{
			name:        "list all items",
			prefix:      "/pods",
			pred:        storage.Everything,
			expectedOut: []apis.Node{*preset[0], *preset[1], *preset[2], *preset[3], *preset[4]},
		},
		{
			name:        "list all items with current resource version and match=NotOlderThan",
			prefix:      "/pods",
			pred:        storage.Everything,
			rv:          list.ResourceVersion,
			rvMatch:     ResourceVersionMatchNotOlderThan,
			expectedOut: []apis.Node{*preset[0], *preset[1], *preset[2], *preset[3], *preset[4]},
		},
		{
			name:   "verify list returns updated version of object; filter returns one item, page size equal to total list size with current resource version and match=NotOlderThan",
			prefix: "/pods",
			pred: storage.SelectionPredicate{
				Field: fields.OneTermEqualSelector("spec.nodeName", "fakeNode"),
				Label: labels.Everything(),
				Limit: 5,
			},
			rv:          list.ResourceVersion,
			rvMatch:     ResourceVersionMatchNotOlderThan,
			expectedOut: []apis.Node{*preset[0]},
		},
		{
			name:   "verify list does not return deleted object; filter for deleted object, page size equal to total list size with current resource version and match=NotOlderThan",
			prefix: "/pods",
			pred: storage.SelectionPredicate{
				Field: fields.OneTermEqualSelector("metadata.name", "baz"),
				Label: labels.Everything(),
				Limit: 5,
			},
			rv:          list.ResourceVersion,
			rvMatch:     ResourceVersionMatchNotOlderThan,
			expectedOut: []apis.Node{},
		},
		{
			name:        "test consistent List",
			prefix:      "/pods/empty",
			pred:        storage.Everything,
			rv:          "",
			expectRV:    currentRV,
			expectedOut: []apis.Node{},
		},
		{
			name:         "test non-consistent List",
			prefix:       "/pods/empty",
			pred:         storage.Everything,
			rv:           "0",
			expectRVFunc: resourceVersionNotOlderThan(list.ResourceVersion),
			expectedOut:  []apis.Node{},
		},
	}
	
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			// For some asynchronous implementations of storage interface (in particular watchcache),
			// certain requests may impact result of further requests. As an example, if we first
			// ensure that watchcache is synchronized up to ResourceVersion X (using Get/List requests
			// with NotOlderThan semantic), the further requests (even specifying earlier resource
			// version) will also return the result synchronized to at least ResourceVersion X.
			// By parallelizing test cases we ensure that the order in which test cases are defined
			// doesn't automatically preclude some scenarios from happening.
			t.Parallel()
			
			if ignoreWatchCacheTests && tt.ignoreForWatchCache {
				t.Skip()
			}
			
			if tt.pred.GetAttrs == nil {
				tt.pred.GetAttrs = getAttrs
			}
			
			out := &apis.NodeList{}
			storageOpts := storage.ListOptions{
				ResourceVersion:      tt.rv,
				ResourceVersionMatch: tt.rvMatch,
				Predicate:            tt.pred,
				Recursive:            true,
			}
			err := store.GetList(ctx, tt.prefix, storageOpts, out)
			if tt.expectRVTooLarge {
				if err == nil || !apierrors.IsTimeout(err) || !storage.IsTooLargeResourceVersion(err) {
					t.Fatalf("expecting resource version too high error, but get: %s", err)
				}
				return
			}
			
			if err != nil {
				if !tt.expectError {
					t.Fatalf("GetList failed: %v", err)
				}
				return
			}
			if tt.expectError {
				t.Fatalf("expected error but got none")
			}
			if (len(out.Continue) > 0) != tt.expectContinue {
				t.Errorf("unexpected continue token: %q", out.Continue)
			}
			
			// If a client requests an exact resource version, it must be echoed back to them.
			if tt.expectRV != "" {
				if tt.expectRV != out.ResourceVersion {
					t.Errorf("resourceVersion in list response want=%s, got=%s", tt.expectRV, out.ResourceVersion)
				}
			}
			if tt.expectRVFunc != nil {
				if err := tt.expectRVFunc(out.ResourceVersion); err != nil {
					t.Errorf("resourceVersion in list response invalid: %v", err)
				}
			}
			
			if tt.expectedAlternatives == nil {
				sort.Sort(sortableNodeList(tt.expectedOut))
				expectNoDiff(t, "incorrect list pods", tt.expectedOut, out.Items)
			} else {
				ExpectContains(t, "incorrect list pods", toInterfaceSlice(tt.expectedAlternatives), out.Items)
			}
			if !cmp.Equal(tt.expectedRemainingItemCount, out.RemainingItemCount) {
				t.Fatalf("unexpected remainingItemCount, diff: %s", cmp.Diff(tt.expectedRemainingItemCount, out.RemainingItemCount))
			}
		})
	}
}

func RunTestGuaranteedUpdate(ctx context.Context, t *testing.T, store InterfaceWithPrefixTransformer, validation KeyValidation) {
	inputObj := &apis.Node{ObjectMeta: meta.ObjectMeta{Name: "foo", Namespace: "test-ns", UID: "A"}}
	key := computeNodeKey(inputObj)
	
	tests := []struct {
		name                string
		key                 string
		ignoreNotFound      bool
		precondition        *storage.Preconditions
		expectNotFoundErr   bool
		expectInvalidObjErr bool
		expectNoUpdate      bool
		transformStale      bool
		hasSelfLink         bool
	}{{
		name:                "non-existing key, ignoreNotFound=false",
		key:                 "/non-existing",
		ignoreNotFound:      false,
		precondition:        nil,
		expectNotFoundErr:   true,
		expectInvalidObjErr: false,
		expectNoUpdate:      false,
	}, {
		name:                "non-existing key, ignoreNotFound=true",
		key:                 "/non-existing",
		ignoreNotFound:      true,
		precondition:        nil,
		expectNotFoundErr:   false,
		expectInvalidObjErr: false,
		expectNoUpdate:      false,
	}, {
		name:                "existing key",
		key:                 key,
		ignoreNotFound:      false,
		precondition:        nil,
		expectNotFoundErr:   false,
		expectInvalidObjErr: false,
		expectNoUpdate:      false,
	}, {
		name:                "same data",
		key:                 key,
		ignoreNotFound:      false,
		precondition:        nil,
		expectNotFoundErr:   false,
		expectInvalidObjErr: false,
		expectNoUpdate:      true,
	}, {
		name:                "same data, a selfLink",
		key:                 key,
		ignoreNotFound:      false,
		precondition:        nil,
		expectNotFoundErr:   false,
		expectInvalidObjErr: false,
		expectNoUpdate:      true,
		hasSelfLink:         true,
	}, {
		name:                "same data, stale",
		key:                 key,
		ignoreNotFound:      false,
		precondition:        nil,
		expectNotFoundErr:   false,
		expectInvalidObjErr: false,
		expectNoUpdate:      false,
		transformStale:      true,
	}, {
		name:                "UID match",
		key:                 key,
		ignoreNotFound:      false,
		precondition:        storage.NewUIDPreconditions("A"),
		expectNotFoundErr:   false,
		expectInvalidObjErr: false,
		expectNoUpdate:      true,
	}, {
		name:                "UID mismatch",
		key:                 key,
		ignoreNotFound:      false,
		precondition:        storage.NewUIDPreconditions("B"),
		expectNotFoundErr:   false,
		expectInvalidObjErr: true,
		expectNoUpdate:      true,
	}}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, storeObj := testPropagateStore(ctx, t, store, inputObj)
			
			out := &apis.Node{}
			newname := "newnametest"
			if tt.expectNoUpdate {
				newname = ""
			}
			
			if tt.transformStale {
				revertTransformer := store.UpdatePrefixTransformer(
					func(transformer *PrefixTransformer) value.Transformer {
						transformer.stale = true
						return transformer
					})
				defer revertTransformer()
			}
			
			version := storeObj.ResourceVersion
			err := store.GuaranteedUpdate(ctx, tt.key, out, tt.ignoreNotFound, tt.precondition,
				storage.SimpleUpdate(func(obj runtime.Object) (runtime.Object, error) {
					if tt.expectNotFoundErr && tt.ignoreNotFound {
						if node := obj.(*apis.Node); node.Name != "" {
							t.Errorf("%s: expecting zero value, but get=%#v", tt.name, node)
						}
					}
					node := *storeObj
					if tt.hasSelfLink {
						//node.SelfLink = "testlink"
					}
					node.Spec.NodeName = newname
					//pod.Annotations = annotations
					return &node, nil
				}), nil)
			
			if tt.expectNotFoundErr {
				if err == nil || !storage.IsNotFound(err) {
					t.Errorf("%s: expecting not found error, but get: %v", tt.name, err)
				}
				return
			}
			if tt.expectInvalidObjErr {
				if err == nil || !storage.IsInvalidObj(err) {
					t.Errorf("%s: expecting invalid UID error, but get: %s", tt.name, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("%s: GuaranteedUpdate failed: %v", tt.name, err)
			}
			validation(ctx, t, key)
			
			switch tt.expectNoUpdate {
			case true:
				if version != out.ResourceVersion {
					t.Errorf("%s: expect no version change, before=%s, after=%s", tt.name, version, out.ResourceVersion)
				}
			case false:
				if version == out.ResourceVersion {
					t.Errorf("%s: expect version change, but get the same version=%s", tt.name, version)
				}
			}
		})
	}
}

func RunTestGuaranteedUpdateWithTTL(ctx context.Context, t *testing.T, store storage.Interface) {
	input := &apis.Node{ObjectMeta: meta.ObjectMeta{Name: "foo", Namespace: "test-ns"}}
	key := computeNodeKey(input)
	
	out := &apis.Node{}
	err := store.GuaranteedUpdate(ctx, key, out, true, nil,
		func(_ runtime.Object, _ storage.ResponseMeta) (runtime.Object, *uint64, error) {
			ttl := uint64(1)
			return input, &ttl, nil
		}, nil)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	
	w, err := store.Watch(ctx, key, storage.ListOptions{ResourceVersion: out.ResourceVersion, Predicate: storage.Everything})
	if err != nil {
		t.Fatalf("Watch failed: %v", err)
	}
	testCheckEventType(t, w, watch.Deleted)
}

func RunTestGuaranteedUpdateChecksStoredData(ctx context.Context, t *testing.T, store InterfaceWithPrefixTransformer) {
	input := &apis.Node{ObjectMeta: meta.ObjectMeta{Name: "foo", Namespace: "test-ns"}}
	key := computeNodeKey(input)
	
	// serialize input into etcd with data that would be normalized by a write -
	// in this case, leading whitespace
	revertTransformer := store.UpdatePrefixTransformer(
		func(transformer *PrefixTransformer) value.Transformer {
			transformer.prefix = []byte(string(transformer.prefix) + " ")
			return transformer
		})
	_, initial := testPropagateStore(ctx, t, store, input)
	revertTransformer()
	
	// this update should write the canonical value to etcd because the new serialization differs
	// from the stored serialization
	input.ResourceVersion = initial.ResourceVersion
	out := &apis.Node{}
	err := store.GuaranteedUpdate(ctx, key, out, true, nil,
		func(_ runtime.Object, _ storage.ResponseMeta) (runtime.Object, *uint64, error) {
			return input, nil, nil
		}, input)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if out.ResourceVersion == initial.ResourceVersion {
		t.Errorf("guaranteed update should have updated the serialized data, got %#v", out)
	}
	
	lastVersion := out.ResourceVersion
	
	// this update should not write to etcd because the input matches the stored data
	input = out
	out = &apis.Node{}
	err = store.GuaranteedUpdate(ctx, key, out, true, nil,
		func(_ runtime.Object, _ storage.ResponseMeta) (runtime.Object, *uint64, error) {
			return input, nil, nil
		}, input)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if out.ResourceVersion != lastVersion {
		t.Errorf("guaranteed update should have short-circuited write, got %#v", out)
	}
	
	revertTransformer = store.UpdatePrefixTransformer(
		func(transformer *PrefixTransformer) value.Transformer {
			transformer.stale = true
			return transformer
		})
	defer revertTransformer()
	
	// this update should write to etcd because the transformer reported stale
	err = store.GuaranteedUpdate(ctx, key, out, true, nil,
		func(_ runtime.Object, _ storage.ResponseMeta) (runtime.Object, *uint64, error) {
			return input, nil, nil
		}, input)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if out.ResourceVersion == lastVersion {
		t.Errorf("guaranteed update should have written to etcd when transformer reported stale, got %#v", out)
	}
}

func RunTestCount(ctx context.Context, t *testing.T, store storage.Interface) {
	resourceA := "/foo.bar.io/abc"
	
	// resourceA is intentionally a prefix of resourceB to ensure that the count
	// for resourceA does not include any objects from resourceB.
	resourceB := fmt.Sprintf("%sdef", resourceA)
	
	resourceACountExpected := 5
	for i := 1; i <= resourceACountExpected; i++ {
		obj := &apis.Node{ObjectMeta: meta.ObjectMeta{Name: fmt.Sprintf("foo-%d", i)}}
		
		key := fmt.Sprintf("%s/%d", resourceA, i)
		if err := store.Create(ctx, key, obj, nil, 0); err != nil {
			t.Fatalf("Create failed: %v", err)
		}
	}
	
	resourceBCount := 4
	for i := 1; i <= resourceBCount; i++ {
		obj := &apis.Node{ObjectMeta: meta.ObjectMeta{Name: fmt.Sprintf("foo-%d", i)}}
		
		key := fmt.Sprintf("%s/%d", resourceB, i)
		if err := store.Create(ctx, key, obj, nil, 0); err != nil {
			t.Fatalf("Create failed: %v", err)
		}
	}
	
	resourceACountGot, err := store.Count(resourceA)
	if err != nil {
		t.Fatalf("store.Count failed: %v", err)
	}
	
	// count for resourceA should not include the objects for resourceB
	// even though resourceA is a prefix of resourceB.
	if int64(resourceACountExpected) != resourceACountGot {
		t.Fatalf("store.Count for resource %s: expected %d but got %d", resourceA, resourceACountExpected, resourceACountGot)
	}
}

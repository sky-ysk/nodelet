package etcd3

import (
	"context"
	"testing"
	
	clientv3 "go.etcd.io/etcd/client/v3"
	apis "hit.edu/framework/pkg/apis/cores"
	
	"hit.edu/framework/pkg/apimachinery/runtime"
	"hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/apimachinery/runtime/schema"
	//"hit.edu/framework/pkg/apiserver/registry/storage"
	"hit.edu/framework/pkg/apiserver/registry/storage"
	"hit.edu/framework/pkg/apiserver/registry/storage/etcd3/testserver"
	"hit.edu/framework/pkg/apiserver/registry/storage/value"
	"k8s.io/apimachinery/pkg/api/apitesting"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
)

const GroupName = "etcd3test"

var SchemeGroupVersion = schema.GroupVersion{Group: GroupName, Version: "v1"}

func init() {
	//metav1.AddToGroupVersion(scheme, metav1.SchemeGroupVersion)
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

func NewFunc() runtime.Object {
	return &apis.Node{}
}
func NewListFunc() runtime.Object {
	return &apis.NodeList{}
}

type setupOptions struct {
	client         func(testing.TB) *clientv3.Client
	codec          runtime.Codec
	newFunc        func() runtime.Object
	newListFunc    func() runtime.Object
	prefix         string
	resourcePrefix string
	groupResource  schema.GroupResource
	transformer    value.Transformer
	leaseConfig    LeaseManagerConfig
	
	recorderEnabled bool
}

func withDefaults(options *setupOptions) {
	options.client = func(t testing.TB) *clientv3.Client {
		return testserver.RunEtcd(t, nil)
	}
	options.codec = apitesting.TestCodec(codecs, schema.GroupVersion{Group: GroupName, Version: "v1"})
	options.newFunc = newNode
	options.newListFunc = newNodeList
	options.prefix = ""
	options.resourcePrefix = "/nodes"
	options.groupResource = schema.GroupResource{Resource: "nodes"}
	options.transformer = newTestTransformer()
	options.leaseConfig = newTestLeaseManagerConfig()
}

type setupOption func(*setupOptions)

var _ setupOption = withDefaults

type clientRecorder struct {
	reads uint64
	clientv3.KV
}

func newTestTransformer() value.Transformer {
	return NewPrefixTransformer([]byte(defaultTestPrefix), false)
}
func newTestLeaseManagerConfig() LeaseManagerConfig {
	cfg := NewDefaultLeaseManagerConfig()
	// As 30s is the default timeout for testing in global configuration,
	// we cannot wait longer than that in a single time: change it to 1s
	// for testing purposes. See wait.ForeverTestTimeout
	cfg.ReuseDurationSeconds = 1
	return cfg
}
func testSetup(t testing.TB, opts ...setupOption) (context.Context, *store, *clientv3.Client) {
	setupOpts := setupOptions{}
	opts = append([]setupOption{withDefaults}, opts...)
	for _, opt := range opts {
		opt(&setupOpts)
	}
	client := setupOpts.client(t)
	if setupOpts.recorderEnabled {
		client.KV = &clientRecorder{KV: client.KV}
	}
	store := newStore(
		client,
		setupOpts.codec,
		setupOpts.newFunc,
		setupOpts.newListFunc,
		setupOpts.prefix,
		setupOpts.resourcePrefix,
		setupOpts.groupResource,
		setupOpts.transformer,
		setupOpts.leaseConfig,
	)
	ctx := context.Background()
	return ctx, store, client
}

func checkStorageInvariants(etcdClient *clientv3.Client, codec runtime.Codec) KeyValidation {
	return func(ctx context.Context, t *testing.T, key string) {
		getResp, err := etcdClient.KV.Get(ctx, key)
		if err != nil {
			t.Fatalf("etcdClient.KV.Get failed: %v", err)
		}
		if len(getResp.Kvs) == 0 {
			t.Fatalf("expecting non empty result on key: %s", key)
		}
		decoded, err := runtime.Decode(codec, getResp.Kvs[0].Value[len(defaultTestPrefix):])
		if err != nil {
			t.Fatalf("expecting successful decode of object from %v\n%v", err, string(getResp.Kvs[0].Value))
		}
		obj := decoded.(*apis.Node)
		if obj.ResourceVersion != "" {
			t.Errorf("stored object should have empty resource version")
		}
	}
}

func TestCreate(t *testing.T) {
	ctx, store, etcdClient := testSetup(t)
	RunTestCreate(ctx, t, store, checkStorageInvariants(etcdClient, store.codec))
}

func TestCreateWithKeyExist(t *testing.T) {
	ctx, store, _ := testSetup(t)
	RunTestCreateWithKeyExist(ctx, t, store)
}

func TestGet(t *testing.T) {
	ctx, store, _ := testSetup(t)
	RunTestGet(ctx, t, store)
}

func TestUnconditionalDelete(t *testing.T) {
	ctx, store, _ := testSetup(t)
	RunTestUnconditionalDelete(ctx, t, store)
}

func TestConditionalDelete(t *testing.T) {
	ctx, store, _ := testSetup(t)
	RunTestConditionalDelete(ctx, t, store)
}

func compactStorage(etcdClient *clientv3.Client) Compaction {
	return func(ctx context.Context, t *testing.T, resourceVersion string) {
		versioner := storage.APIObjectVersioner{}
		rv, err := versioner.ParseResourceVersion(resourceVersion)
		if err != nil {
			t.Fatal(err)
		}
		if _, _, err = compact(ctx, etcdClient, 0, int64(rv)); err != nil {
			t.Fatalf("Unable to compact, %v", err)
		}
	}
}
func TestList(t *testing.T) {
	ctx, store, client := testSetup(t)
	RunTestList(ctx, t, store, compactStorage(client), false)
}

func TestGuaranteedUpdate(t *testing.T) {
	ctx, store, etcdClient := testSetup(t)
	RunTestGuaranteedUpdate(ctx, t, &storeWithPrefixTransformer{store}, checkStorageInvariants(etcdClient, store.codec))
}

func TestGuaranteedUpdateWithTTL(t *testing.T) {
	ctx, store, _ := testSetup(t)
	RunTestGuaranteedUpdateWithTTL(ctx, t, store)
}

func TestGuaranteedUpdateChecksStoredData(t *testing.T) {
	ctx, store, _ := testSetup(t)
	RunTestGuaranteedUpdateChecksStoredData(ctx, t, &storeWithPrefixTransformer{store})
}

func TestCount(t *testing.T) {
	ctx, store, _ := testSetup(t)
	RunTestCount(ctx, t, store)
}

package endpoints

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/emicklei/go-restful/v3"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
	"hit.edu/framework/pkg/apimachinery/runtime"
	"hit.edu/framework/pkg/apimachinery/runtime/schema"
	"hit.edu/framework/pkg/apimachinery/watch"
	apis "hit.edu/framework/pkg/apis/cores"
	_ "hit.edu/framework/pkg/apis/cores/install"
	"hit.edu/framework/pkg/apis/legacyscheme"
	"hit.edu/framework/pkg/apis/meta"
	_ "hit.edu/framework/pkg/apis/meta/install"
	_ "hit.edu/framework/pkg/apis/meta/internalversion/install"
	"hit.edu/framework/pkg/apiserver/endpoints/handler/responsewriters"
	"hit.edu/framework/pkg/apiserver/endpoints/request"
	corerest "hit.edu/framework/pkg/apiserver/registry/core/rest"
	genericregistry "hit.edu/framework/pkg/apiserver/registry/generic"
	"hit.edu/framework/pkg/apiserver/registry/rest"
	"hit.edu/framework/pkg/apiserver/registry/storage/etcd3/testserver"
	"hit.edu/framework/pkg/apiserver/registry/storage/storagebackend"
	"hit.edu/framework/pkg/component-base/logs"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/util/sets"
	"net/http/httptest"
	"net/url"
	"os"
	"path"
	"sync"

	"net/http"
	"testing"
	"time"
)

var minRequestTimeout = time.Second * 34

var testPrefix = "apis"
var testAPIGroup = "resources"
var testAPIVersion = "v1"
var testGroupVersion = schema.GroupVersion{Group: testAPIGroup, Version: testAPIVersion}

var scheme = legacyscheme.Scheme
var codecs = legacyscheme.Codecs
var codec = codecs.LegacyCodec()
var accessor = meta.NewAccessor()
var namer runtime.Namer = accessor

func init() {
}

type defaultAPIServer struct {
	http.Handler
	container *restful.Container
}

// uses the default settings
func handle(storage map[string]rest.Storage) http.Handler {
	return handleInternal(storage)
}

func handleInternal(storage map[string]rest.Storage) http.Handler {
	container := restful.NewContainer()
	container.Router(restful.CurlyRouter{})
	mux := container.ServeMux

	template := APIGroupVersion{
		Storage: storage,

		Creater:           scheme,
		Convertor:         scheme,
		ParameterCodec:    legacyscheme.ParameterCodec,
		Serializer:        codecs,
		Typer:             scheme,
		Namer:             namer,
		MetaGroupVersion:  &schema.GroupVersion{Group: "meta", Version: "v1"},
		MinRequestTimeout: minRequestTimeout,
	}
	//  group:resources version:v1 Install
	{
		group := template
		group.Root = "/" + testPrefix
		group.GroupVersion = testGroupVersion
		group.Serializer = codecs
		if err := (&group).InstallREST(container); err != nil {
			panic(fmt.Sprintf("unable to install container %s: %v", group.GroupVersion, err))
		}
	}
	handler := withRequestInfo(mux, testRequestInfoResolver())
	return &defaultAPIServer{handler, container}
}

type TestRESTStorage rest.Storage

func getTestRESTStorage(t *testing.T) (map[string]rest.Storage, *EtcdTestServer) {
	t.Helper()
	server, config := NewUnsecuredEtcd3TestClientServer(t)
	config.Codec = codec
	resourceConfig := &storagebackend.ConfigForResource{
		Config:        *config,
		GroupResource: schema.GroupResource{Group: "resources", Resource: "nodes"},
	}
	restOptions := genericregistry.RESTOptions{
		StorageConfig:           resourceConfig,
		Decorator:               genericregistry.UndecoratedStorage,
		DeleteCollectionWorkers: 3,
		ResourcePrefix:          "nodes",
	}
	apiGroupInfo, err := corerest.NewRESTStorage(restOptions)
	if err != nil {
		logs.Error("get coreStorage failed", zap.Error(err))
	}
	return apiGroupInfo.VersionedResourcesStorageMap["v1"], server
}

func testRequestInfoResolver() *request.RequestInfoFactory {
	return &request.RequestInfoFactory{
		APIPrefixes: sets.NewString("apis"),
	}
}

// WithRequestInfo 对http请求预处理
func withRequestInfo(handler http.Handler, resolver request.RequestInfoResolver) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		info, err := resolver.NewRequestInfo(req)
		if err != nil {
			responsewriters.InternalError(w, req, fmt.Errorf("failed to create RequestInfo: %v", err))
			return
		}

		req = req.WithContext(request.WithRequestInfo(ctx, info))

		handler.ServeHTTP(w, req)
	})
}

func extractBody(response *http.Response, object runtime.Object) (string, error) {
	return extractBodyDecoder(response, object, codec)
}

func extractBodyDecoder(response *http.Response, object runtime.Object, decoder runtime.Decoder) (string, error) {
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return string(body), err
	}
	return string(body), runtime.DecodeInto(decoder, body, object)
}

// EtcdTestServer encapsulates the datastructures needed to start local instance for testing
type EtcdTestServer struct {
	V3Client *clientv3.Client
}

func (e *EtcdTestServer) Terminate(t *testing.T) {
	e.V3Client.Close()
	// no-op, server termination moved to test cleanup
}

// NewUnsecuredEtcd3TestClientServer creates a new client and server for testing
func NewUnsecuredEtcd3TestClientServer(t *testing.T) (*EtcdTestServer, *storagebackend.Config) {
	server := &EtcdTestServer{}
	server.V3Client = testserver.RunEtcd(t, nil)
	config := &storagebackend.Config{
		Type:   "etcd3",
		Prefix: PathPrefix(),
		Transport: storagebackend.TransportConfig{
			ServerList: server.V3Client.Endpoints(),
		},
	}
	return server, config
}

// PathPrefix returns the prefix set via the ETCD_PREFIX environment variable (if any).
func PathPrefix() string {
	pref := os.Getenv("ETCD_PREFIX")
	if pref == "" {
		pref = "registry"
	}
	return path.Join("/", pref)
}

func TestNotFound(t *testing.T) {
	type T struct {
		Method string
		Path   string
		Status int
	}
	cases := map[string]T{
		"GET NotFound": {"GET", "/notfound/" + testGroupVersion.Group + "/" + testGroupVersion.Version + "/simpleroots", http.StatusNotFound},
	}
	restStorage, embedEtcdServer := getTestRESTStorage(t)
	defer embedEtcdServer.Terminate(t)
	handler := handle(restStorage)
	server := httptest.NewServer(handler)
	defer server.Close()
	client := http.Client{}
	for k, v := range cases {
		request, err := http.NewRequest(v.Method, server.URL+v.Path, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		response, err := client.Do(request)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if response.StatusCode != v.Status {
			t.Errorf("Expected %d for %s (%s), Got %#v", v.Status, v.Method, k, response)
		}
	}
}

func TestCreate(t *testing.T) {
	restStorage, embedEtcdServer := getTestRESTStorage(t)
	defer embedEtcdServer.Terminate(t)
	handler := handle(restStorage)
	server := httptest.NewServer(handler)
	defer server.Close()
	client := http.Client{}

	simple := &apis.Node{
		TypeMeta:   meta.TypeMeta{Kind: "Node", APIVersion: "resources/v1"},
		ObjectMeta: meta.ObjectMeta{Name: "foo"},
		Spec: apis.NodeSpec{
			NodeName:      "123",
			HostName:      "456",
			Unschedulable: false,
		},
		Status: apis.NodeStatus{
			Capacity:    nil,
			Allocatable: nil,
			Images:      nil,
			Wasms:       nil,
			Addresses:   apis.NodeAddress{},
			NodeInfo:    apis.NodeSystemInfo{},
		},
	}
	data, err := runtime.Encode(codec, simple)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	bytesBuffer := bytes.NewBuffer(data)
	request, err := http.NewRequest("POST", server.URL+"/"+testPrefix+"/"+testGroupVersion.Group+"/"+testGroupVersion.Version+"/nodes", bytesBuffer)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	wg := sync.WaitGroup{}
	wg.Add(1)
	var response *http.Response
	go func() {
		response, err = client.Do(request)
		wg.Done()
	}()
	wg.Wait()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	var itemOut apis.Node
	_, err = extractBody(response, &itemOut)
	if err != nil {
		t.Errorf("unexpected error: %v %#v", err, response)
	}
	if response.StatusCode != http.StatusCreated {
		t.Errorf("Unexpected status: %d, Expected: %d, %#v", response.StatusCode, http.StatusCreated, response)
	}
}

func TestDelete(t *testing.T) {
	restStorage, embedEtcdServer := getTestRESTStorage(t)
	defer embedEtcdServer.Terminate(t)
	handler := handle(restStorage)
	ID := "foo"
	server := httptest.NewServer(handler)
	defer server.Close()
	client := http.Client{}

	//创建一个Node资源
	simple := &apis.Node{
		TypeMeta:   meta.TypeMeta{Kind: "Node", APIVersion: "resources/v1"},
		ObjectMeta: meta.ObjectMeta{Name: "foo"},
		Spec: apis.NodeSpec{
			NodeName:      "123",
			HostName:      "456",
			Unschedulable: false,
		},
		Status: apis.NodeStatus{
			Capacity:    nil,
			Allocatable: nil,
			Images:      nil,
			Wasms:       nil,
			Addresses:   apis.NodeAddress{},
			NodeInfo:    apis.NodeSystemInfo{},
		},
	}
	data, err := runtime.Encode(codec, simple)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	bytesBuffer := bytes.NewBuffer(data)
	req, err := http.NewRequest("POST", server.URL+"/"+testPrefix+"/"+testGroupVersion.Group+"/"+testGroupVersion.Version+"/nodes", bytesBuffer)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	wg := sync.WaitGroup{}
	wg.Add(1)
	var response *http.Response
	go func() {
		response, err = client.Do(req)
		wg.Done()
	}()
	wg.Wait()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	var itemOut apis.Node
	_, err = extractBody(response, &itemOut)
	if err != nil {
		t.Errorf("unexpected error: %v %#v", err, response)
	}
	if response.StatusCode != http.StatusCreated {
		t.Errorf("Unexpected status: %d, Expected: %d, %#v", response.StatusCode, http.StatusCreated, response)
	}

	//测试删除
	request, err := http.NewRequest("DELETE", server.URL+"/"+testPrefix+"/"+testGroupVersion.Group+"/"+testGroupVersion.Version+"/nodes/"+ID, nil)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	res, err := client.Do(request)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.StatusCode != http.StatusOK {
		t.Errorf("unexpected response: %#v", res)
	}
}

func TestGet(t *testing.T) {
	restStorage, embedEtcdServer := getTestRESTStorage(t)
	defer embedEtcdServer.Terminate(t)
	handler := handle(restStorage)
	ID := "foo"
	server := httptest.NewServer(handler)
	defer server.Close()

	client := http.Client{}

	//创建一个Node资源
	simple := &apis.Node{
		TypeMeta:   meta.TypeMeta{Kind: "Node", APIVersion: "resources/v1"},
		ObjectMeta: meta.ObjectMeta{Name: "foo"},
		Spec: apis.NodeSpec{
			NodeName:      "123",
			HostName:      "456",
			Unschedulable: false,
		},
		Status: apis.NodeStatus{
			Capacity:    nil,
			Allocatable: nil,
			Images:      nil,
			Wasms:       nil,
			Addresses:   apis.NodeAddress{},
			NodeInfo:    apis.NodeSystemInfo{},
		},
	}
	data, err := runtime.Encode(codec, simple)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	bytesBuffer := bytes.NewBuffer(data)
	req, err := http.NewRequest("POST", server.URL+"/"+testPrefix+"/"+testGroupVersion.Group+"/"+testGroupVersion.Version+"/nodes", bytesBuffer)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	wg := sync.WaitGroup{}
	wg.Add(1)
	var response *http.Response
	go func() {
		response, err = client.Do(req)
		wg.Done()
	}()
	wg.Wait()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if err != nil {
		t.Errorf("unexpected error: %v %#v", err, response)
	}
	if response.StatusCode != http.StatusCreated {
		t.Errorf("Unexpected status: %d, Expected: %d, %#v", response.StatusCode, http.StatusCreated, response)
	}

	//测试查询
	resp, err := http.Get(server.URL + "/" + testPrefix + "/" + testGroupVersion.Group + "/" + testGroupVersion.Version + "/nodes/" + ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected response: %#v", resp)
	}
	var itemOut apis.Node
	body, err := extractBody(resp, &itemOut)
	if err != nil {
		t.Errorf("unexpected error: %v %#v", err, resp)
	}
	if itemOut.Name != simple.Name {
		t.Errorf("Unexpected data: %#v, expected %#v (%s)", itemOut, simple, string(body))
	}
}

func TestUpdate(t *testing.T) {
	restStorage, embedEtcdServer := getTestRESTStorage(t)
	defer embedEtcdServer.Terminate(t)
	handler := handle(restStorage)
	ID := "foo"
	server := httptest.NewServer(handler)
	defer server.Close()
	client := http.Client{}

	//创建一个Node资源
	original := &apis.Node{
		TypeMeta:   meta.TypeMeta{Kind: "Node", APIVersion: "resources/v1"},
		ObjectMeta: meta.ObjectMeta{Name: "foo"},
		Spec: apis.NodeSpec{
			NodeName:      "123",
			HostName:      "456",
			Unschedulable: false,
		},
		Status: apis.NodeStatus{
			Capacity:    nil,
			Allocatable: nil,
			Images:      nil,
			Wasms:       nil,
			Addresses:   apis.NodeAddress{},
			NodeInfo:    apis.NodeSystemInfo{},
		},
	}
	data, err := runtime.Encode(codec, original)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	bytesBuffer := bytes.NewBuffer(data)
	req, err := http.NewRequest("POST", server.URL+"/"+testPrefix+"/"+testGroupVersion.Group+"/"+testGroupVersion.Version+"/nodes", bytesBuffer)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	wg := sync.WaitGroup{}
	wg.Add(1)
	var response *http.Response
	go func() {
		response, err = client.Do(req)
		wg.Done()
	}()
	wg.Wait()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if err != nil {
		t.Errorf("unexpected error: %v %#v", err, response)
	}
	if response.StatusCode != http.StatusCreated {
		t.Errorf("Unexpected status: %d, Expected: %d, %#v", response.StatusCode, http.StatusCreated, response)
	}
	//测试更新
	simple := &apis.Node{
		TypeMeta:   meta.TypeMeta{Kind: "Node", APIVersion: "resources/v1"},
		ObjectMeta: meta.ObjectMeta{Name: "foo"},
		Spec: apis.NodeSpec{
			NodeName: "updated",
			HostName: "updated",
		},
	}
	data, err = runtime.Encode(codec, simple)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	request, err := http.NewRequest("PUT", server.URL+"/"+testPrefix+"/"+testGroupVersion.Group+"/"+testGroupVersion.Version+"/nodes/"+ID, bytes.NewBuffer(data))
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	wg = sync.WaitGroup{}
	wg.Add(1)
	go func() {
		response, err = client.Do(request)
		wg.Done()
	}()
	wg.Wait()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	var itemOut apis.Node
	_, err = extractBody(response, &itemOut)
	if err != nil {
		t.Errorf("unexpected error: %v %#v", err, response)
	}
	if response.StatusCode != http.StatusOK {
		t.Errorf("Unexpected status: %d, Expected: %d, %#v", response.StatusCode, http.StatusOK, response)
	}
	if itemOut.Spec.HostName != simple.Spec.HostName || itemOut.Spec.NodeName != simple.Spec.NodeName {
		t.Errorf("Unexpected HostName:%s,NodeName:%s,Expected HostName:%s,NodeName:%s,", itemOut.Spec.HostName, itemOut.Spec.NodeName, simple.Spec.HostName, simple.Spec.NodeName)
	}
}

func TestDeleteCollection(t *testing.T) {
	restStorage, embedEtcdServer := getTestRESTStorage(t)
	defer embedEtcdServer.Terminate(t)
	handler := handle(restStorage)
	server := httptest.NewServer(handler)
	defer server.Close()
	client := http.Client{}

	//创建两个Node资源
	node1 := &apis.Node{
		TypeMeta:   meta.TypeMeta{Kind: "Node", APIVersion: "resources/v1"},
		ObjectMeta: meta.ObjectMeta{Name: "foo"},
		Spec: apis.NodeSpec{
			NodeName:      "123",
			HostName:      "456",
			Unschedulable: false,
		},
		Status: apis.NodeStatus{
			Capacity:    nil,
			Allocatable: nil,
			Images:      nil,
			Wasms:       nil,
			Addresses:   apis.NodeAddress{},
			NodeInfo:    apis.NodeSystemInfo{},
		},
	}
	data, err := runtime.Encode(codec, node1)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	bytesBuffer := bytes.NewBuffer(data)
	req, err := http.NewRequest("POST", server.URL+"/"+testPrefix+"/"+testGroupVersion.Group+"/"+testGroupVersion.Version+"/nodes", bytesBuffer)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	wg := sync.WaitGroup{}
	wg.Add(1)
	var response *http.Response
	go func() {
		response, err = client.Do(req)
		wg.Done()
	}()
	wg.Wait()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if err != nil {
		t.Errorf("unexpected error: %v %#v", err, response)
	}
	if response.StatusCode != http.StatusCreated {
		t.Errorf("Unexpected status: %d, Expected: %d, %#v", response.StatusCode, http.StatusCreated, response)
	}

	node2 := &apis.Node{
		TypeMeta:   meta.TypeMeta{Kind: "Node", APIVersion: "resources/v1"},
		ObjectMeta: meta.ObjectMeta{Name: "foo1"},
		Spec: apis.NodeSpec{
			NodeName:      "123",
			HostName:      "456",
			Unschedulable: false,
		},
		Status: apis.NodeStatus{
			Capacity:    nil,
			Allocatable: nil,
			Images:      nil,
			Wasms:       nil,
			Addresses:   apis.NodeAddress{},
			NodeInfo:    apis.NodeSystemInfo{},
		},
	}
	data, err = runtime.Encode(codec, node2)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	bytesBuffer = bytes.NewBuffer(data)
	req, err = http.NewRequest("POST", server.URL+"/"+testPrefix+"/"+testGroupVersion.Group+"/"+testGroupVersion.Version+"/nodes", bytesBuffer)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	wg = sync.WaitGroup{}
	wg.Add(1)
	go func() {
		response, err = client.Do(req)
		wg.Done()
	}()
	wg.Wait()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if err != nil {
		t.Errorf("unexpected error: %v %#v", err, response)
	}
	if response.StatusCode != http.StatusCreated {
		t.Errorf("Unexpected status: %d, Expected: %d, %#v", response.StatusCode, http.StatusCreated, response)
	}

	//测试批量删除
	request, err := http.NewRequest("DELETE", server.URL+"/"+testPrefix+"/"+testGroupVersion.Group+"/"+testGroupVersion.Version+"/nodes", nil)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	res, err := client.Do(request)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.StatusCode != http.StatusOK {
		t.Errorf("unexpected response: %#v", res)
	}
}

func TestList(t *testing.T) {
	restStorage, embedEtcdServer := getTestRESTStorage(t)
	defer embedEtcdServer.Terminate(t)
	handler := handle(restStorage)
	server := httptest.NewServer(handler)
	defer server.Close()

	client := http.Client{}

	//创建两个Node资源
	node1 := &apis.Node{
		TypeMeta:   meta.TypeMeta{Kind: "Node", APIVersion: "resources/v1"},
		ObjectMeta: meta.ObjectMeta{Name: "foo"},
		Spec: apis.NodeSpec{
			NodeName:      "123",
			HostName:      "456",
			Unschedulable: false,
		},
		Status: apis.NodeStatus{
			Capacity:    nil,
			Allocatable: nil,
			Images:      nil,
			Wasms:       nil,
			Addresses:   apis.NodeAddress{},
			NodeInfo:    apis.NodeSystemInfo{},
		},
	}
	data, err := runtime.Encode(codec, node1)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	bytesBuffer := bytes.NewBuffer(data)
	req, err := http.NewRequest("POST", server.URL+"/"+testPrefix+"/"+testGroupVersion.Group+"/"+testGroupVersion.Version+"/nodes", bytesBuffer)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	wg := sync.WaitGroup{}
	wg.Add(1)
	var response *http.Response
	go func() {
		response, err = client.Do(req)
		wg.Done()
	}()
	wg.Wait()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if err != nil {
		t.Errorf("unexpected error: %v %#v", err, response)
	}
	if response.StatusCode != http.StatusCreated {
		t.Errorf("Unexpected status: %d, Expected: %d, %#v", response.StatusCode, http.StatusCreated, response)
	}

	node2 := &apis.Node{
		TypeMeta:   meta.TypeMeta{Kind: "Node", APIVersion: "resources/v1"},
		ObjectMeta: meta.ObjectMeta{Name: "foo1"},
		Spec: apis.NodeSpec{
			NodeName:      "123",
			HostName:      "456",
			Unschedulable: false,
		},
		Status: apis.NodeStatus{
			Capacity:    nil,
			Allocatable: nil,
			Images:      nil,
			Wasms:       nil,
			Addresses:   apis.NodeAddress{},
			NodeInfo:    apis.NodeSystemInfo{},
		},
	}
	data, err = runtime.Encode(codec, node2)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	bytesBuffer = bytes.NewBuffer(data)
	req, err = http.NewRequest("POST", server.URL+"/"+testPrefix+"/"+testGroupVersion.Group+"/"+testGroupVersion.Version+"/nodes", bytesBuffer)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	wg = sync.WaitGroup{}
	wg.Add(1)
	go func() {
		response, err = client.Do(req)
		wg.Done()
	}()
	wg.Wait()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if err != nil {
		t.Errorf("unexpected error: %v %#v", err, response)
	}
	if response.StatusCode != http.StatusCreated {
		t.Errorf("Unexpected status: %d, Expected: %d, %#v", response.StatusCode, http.StatusCreated, response)
	}

	//测试List
	url := server.URL + "/" + testPrefix + "/" + testGroupVersion.Group + "/" + testGroupVersion.Version + "/nodes"

	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("unexpected status: %d from url: %s, Expected: %d, %#v", resp.StatusCode, url, http.StatusOK, resp)
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		t.Logf("body: %s", string(body))
	}

	var nodeList apis.NodeList
	_, err = extractBody(resp, &nodeList)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	for _, item := range nodeList.Items {
		if item.Name != "foo" && item.Name != "foo1" {
			t.Errorf("get unexpected item :%s", item.Name)
		}
	}

}

func TestPatch(t *testing.T) {
	restStorage, embedEtcdServer := getTestRESTStorage(t)
	defer embedEtcdServer.Terminate(t)
	handler := handle(restStorage)
	server := httptest.NewServer(handler)
	defer server.Close()
	client := http.Client{}
	//创建一个Node资源
	simple := &apis.Node{
		TypeMeta:   meta.TypeMeta{Kind: "Node", APIVersion: "resources/v1"},
		ObjectMeta: meta.ObjectMeta{Name: "foo"},
		Spec: apis.NodeSpec{
			NodeName:      "123",
			HostName:      "456",
			Unschedulable: false,
		},
		Status: apis.NodeStatus{
			Capacity:    nil,
			Allocatable: nil,
			Images:      nil,
			Wasms:       nil,
			Addresses:   apis.NodeAddress{},
			NodeInfo:    apis.NodeSystemInfo{},
		},
	}
	data, err := runtime.Encode(codec, simple)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	bytesBuffer := bytes.NewBuffer(data)
	req, err := http.NewRequest("POST", server.URL+"/"+testPrefix+"/"+testGroupVersion.Group+"/"+testGroupVersion.Version+"/nodes", bytesBuffer)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	wg := sync.WaitGroup{}
	wg.Add(1)
	var response *http.Response
	go func() {
		response, err = client.Do(req)
		wg.Done()
	}()
	wg.Wait()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if err != nil {
		t.Errorf("unexpected error: %v %#v", err, response)
	}
	if response.StatusCode != http.StatusCreated {
		t.Errorf("Unexpected status: %d, Expected: %d, %#v", response.StatusCode, http.StatusCreated, response)
	}

	//测试Patch
	jsonPatchBytes := []byte(`[
		{ "op": "replace", "path": "/Spec/NodeName", "value": "boo" },
		{ "op": "add", "path": "/Status/Addresses/Type", "value": "IPv6" },
		{ "op": "remove", "path": "/Status/NodeInfo/BootID" }
	]`)
	mergePatchBytes := []byte(`{
 		"Spec": {
   		"NodeName": "boo"
		},
 		"Status": {
   		"Addresses": {
     			"Type": "IPv6"
			},
   		"NodeInfo": {
     			"BootID": null
   		}
 		}
	}`)

	testCases := []struct {
		ID         string
		patchType  string
		patchBytes []byte
	}{
		{
			ID:         "foo",
			patchType:  "application/json-patch+json",
			patchBytes: jsonPatchBytes,
		},
		{
			ID:         "foo",
			patchType:  "application/merge-patch+json",
			patchBytes: mergePatchBytes,
		},
	}

	for _, tc := range testCases {
		client := http.Client{}
		request, err := http.NewRequest("PATCH", server.URL+"/"+testPrefix+"/"+testGroupVersion.Group+"/"+testGroupVersion.Version+"/nodes/"+tc.ID, bytes.NewReader(tc.patchBytes))
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		request.Header.Set("Content-Type", tc.patchType)
		response, err := client.Do(request)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if response.StatusCode != http.StatusOK {
			t.Errorf("Unexpected status: %d, Expected: %d, %#v", response.StatusCode, http.StatusOK, response)
		}
		var itemOut apis.Node
		_, err = extractBody(response, &itemOut)
		if err != nil {
			t.Errorf("unexpected error: %v %#v", err, response)
		}
		if itemOut.Spec.NodeName != "boo" || itemOut.Status.Addresses.Type != "IPv6" || itemOut.Status.NodeInfo.BootID != "" {
			t.Errorf("Unexpected NodeName,Addresses,BootID:%s,%s,%s  Expected:\"boo\",\"IPv6\",\"\"", itemOut.Spec.NodeName, itemOut.Status.Addresses.Type, itemOut.Status.NodeInfo.BootID)
		}
		if itemOut.Spec.HostName != simple.Spec.HostName {
			t.Errorf("Unexpected modified field: Spec.HostName")
		}
	}
}

func TestWatchList(t *testing.T) {
	restStorage, embedEtcdServer := getTestRESTStorage(t)
	defer embedEtcdServer.Terminate(t)
	handler := handle(restStorage)
	server := httptest.NewServer(handler)

	//查询NodeList最新的ResourceVersion
	client := http.Client{}
	dest, _ := url.Parse(server.URL)
	dest.Path = "/" + testPrefix + "/" + testGroupVersion.Group + "/" + testGroupVersion.Version + "/nodes"
	reqList, _ := http.NewRequest("GET", dest.String(), nil)
	responseList, err := client.Do(reqList)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if responseList.StatusCode != http.StatusOK {
		t.Errorf("Unexpected status: %d, Expected: %d, %#v", responseList.StatusCode, http.StatusOK, responseList)
	}
	var nodeList apis.NodeList
	_, err = extractBody(responseList, &nodeList)
	resourceVersion := nodeList.ResourceVersion

	//发送Watch请求
	dest, _ = url.Parse(server.URL)
	dest.Path = "/" + testPrefix + "/" + testGroupVersion.Group + "/" + testGroupVersion.Version + "/nodes"
	dest.RawQuery = "watch=true&sendInitialEvents=false&resourceVersionMatch=NotOlderThan&resourceVersion=" + resourceVersion
	req, _ := http.NewRequest("GET", dest.String(), nil)
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Unexpected status: %d, Expected: %d, %#v", resp.StatusCode, http.StatusOK, resp)
	}

	//创建一个TestNode
	simple := &apis.Node{
		TypeMeta:   meta.TypeMeta{Kind: "Node", APIVersion: "resources/v1"},
		ObjectMeta: meta.ObjectMeta{Name: "watchTestNode"},
		Spec: apis.NodeSpec{
			NodeName:      "123",
			HostName:      "456",
			Unschedulable: false,
		},
		Status: apis.NodeStatus{
			Capacity:    nil,
			Allocatable: nil,
			Images:      nil,
			Wasms:       nil,
			Addresses:   apis.NodeAddress{},
			NodeInfo:    apis.NodeSystemInfo{},
		},
	}
	data, err := runtime.Encode(codec, simple)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	bytesBuffer := bytes.NewBuffer(data)
	req, err = http.NewRequest("POST", server.URL+"/"+testPrefix+"/"+testGroupVersion.Group+"/"+testGroupVersion.Version+"/nodes", bytesBuffer)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	response, err := client.Do(req)
	if response.StatusCode != http.StatusCreated {
		t.Errorf("Unexpected status: %d, Expected: %d, %#v", response.StatusCode, http.StatusOK, response)
	}
	response.Body.Close()

	var dataWatch []byte
	scanner := bufio.NewScanner(resp.Body)
	scanner.Scan()
	dataWatch = scanner.Bytes()

	//检查watchEvent
	var got watchJSON
	err = json.Unmarshal(dataWatch, &got)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if got.Type != watch.Added {
		t.Errorf("Unexpected event type: %s, Expected: %s", got.Type, watch.Added)
	}
	var itemOut apis.Node
	err = json.Unmarshal(got.Object, &itemOut)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if itemOut.Name != "watchTestNode" {
		t.Errorf("Unexpected event Node name:%s,Expected:%s", itemOut.Name, "watchTestNode")
	}
	resp.Body.Close()
	server.Close()
}

func TestWatchSingleResource(t *testing.T) {
	restStorage, embedEtcdServer := getTestRESTStorage(t)
	defer embedEtcdServer.Terminate(t)
	handler := handle(restStorage)
	server := httptest.NewServer(handler)

	client := http.Client{}

	//创建一个Node资源
	original := &apis.Node{
		TypeMeta:   meta.TypeMeta{Kind: "Node", APIVersion: "resources/v1"},
		ObjectMeta: meta.ObjectMeta{Name: "watchTestNode"},
		Spec: apis.NodeSpec{
			NodeName:      "123",
			HostName:      "456",
			Unschedulable: false,
		},
		Status: apis.NodeStatus{
			Capacity:    nil,
			Allocatable: nil,
			Images:      nil,
			Wasms:       nil,
			Addresses:   apis.NodeAddress{},
			NodeInfo:    apis.NodeSystemInfo{},
		},
	}
	data, err := runtime.Encode(codec, original)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	bytesBuffer := bytes.NewBuffer(data)
	req, err := http.NewRequest("POST", server.URL+"/"+testPrefix+"/"+testGroupVersion.Group+"/"+testGroupVersion.Version+"/nodes", bytesBuffer)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	wg := sync.WaitGroup{}
	wg.Add(1)
	var response *http.Response
	go func() {
		response, err = client.Do(req)
		wg.Done()
	}()
	wg.Wait()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if err != nil {
		t.Errorf("unexpected error: %v %#v", err, response)
	}
	if response.StatusCode != http.StatusCreated {
		t.Errorf("Unexpected status: %d, Expected: %d, %#v", response.StatusCode, http.StatusCreated, response)
	}

	//查询NodeList最新的ResourceVersion
	dest, _ := url.Parse(server.URL)
	dest.Path = "/" + testPrefix + "/" + testGroupVersion.Group + "/" + testGroupVersion.Version + "/nodes"
	reqList, _ := http.NewRequest("GET", dest.String(), nil)
	responseList, err := client.Do(reqList)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if responseList.StatusCode != http.StatusOK {
		t.Errorf("Unexpected status: %d, Expected: %d, %#v", responseList.StatusCode, http.StatusOK, responseList)
	}
	var nodeList apis.NodeList
	_, err = extractBody(responseList, &nodeList)
	resourceVersion := nodeList.ResourceVersion

	//发送Watch请求
	dest, _ = url.Parse(server.URL)
	dest.Path = "/" + testPrefix + "/" + testGroupVersion.Group + "/" + testGroupVersion.Version + "/nodes"
	dest.RawQuery = "watch=true&fieldSelector=metadata.name=watchTestNode&sendInitialEvents=false&resourceVersionMatch=NotOlderThan&resourceVersion=" + resourceVersion
	req, _ = http.NewRequest("GET", dest.String(), nil)
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Unexpected status: %d, Expected: %d, %#v", resp.StatusCode, http.StatusOK, resp)
	}

	//新增一个TestNode，测试是否只监听单个资源
	anotherNode := &apis.Node{
		TypeMeta:   meta.TypeMeta{Kind: "Node", APIVersion: "resources/v1"},
		ObjectMeta: meta.ObjectMeta{Name: "anotherNode"},
		Spec: apis.NodeSpec{
			NodeName:      "123",
			HostName:      "456",
			Unschedulable: false,
		},
		Status: apis.NodeStatus{
			Capacity:    nil,
			Allocatable: nil,
			Images:      nil,
			Wasms:       nil,
			Addresses:   apis.NodeAddress{},
			NodeInfo:    apis.NodeSystemInfo{},
		},
	}
	data, err = runtime.Encode(codec, anotherNode)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	bytesBuffer = bytes.NewBuffer(data)
	req, err = http.NewRequest("POST", server.URL+"/"+testPrefix+"/"+testGroupVersion.Group+"/"+testGroupVersion.Version+"/nodes", bytesBuffer)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	response1, err := client.Do(req)
	if response1.StatusCode != http.StatusCreated {
		t.Errorf("Unexpected status: %d, Expected: %d, %#v", response.StatusCode, http.StatusCreated, response)
	}
	response1.Body.Close()

	//更新一个TestNode
	simple := &apis.Node{
		TypeMeta:   meta.TypeMeta{Kind: "Node", APIVersion: "resources/v1"},
		ObjectMeta: meta.ObjectMeta{Name: "watchTestNode"},
		Spec: apis.NodeSpec{
			NodeName: "watchTestNode",
		},
	}
	data, err = runtime.Encode(codec, simple)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	bytesBuffer = bytes.NewBuffer(data)
	req, err = http.NewRequest("PUT", server.URL+"/"+testPrefix+"/"+testGroupVersion.Group+"/"+testGroupVersion.Version+"/nodes/"+simple.Name, bytesBuffer)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	response, err = client.Do(req)
	if response.StatusCode != http.StatusOK {
		t.Errorf("Unexpected status: %d, Expected: %d, %#v", response.StatusCode, http.StatusOK, response)
	}
	response.Body.Close()

	var dataWatch []byte
	scanner := bufio.NewScanner(resp.Body)
	scanner.Scan()
	dataWatch = scanner.Bytes()

	//检查watchEvent
	var got watchJSON
	err = json.Unmarshal(dataWatch, &got)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if got.Type != watch.Modified {
		t.Errorf("Unexpected event type: %s, Expected: %s", got.Type, watch.Modified)
	}
	var itemOut apis.Node
	err = json.Unmarshal(got.Object, &itemOut)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if itemOut.Spec.NodeName != "watchTestNode" {
		t.Errorf("Unexpected Spec NodeName:%s,Expected:%s", itemOut.Spec.NodeName, "watchTestNode")
	}

	resp.Body.Close()
	server.Close()
}

func TestNamespacedCreate(t *testing.T) {
	restStorage, embedEtcdServer := getTestRESTStorage(t)
	defer embedEtcdServer.Terminate(t)
	handler := handle(restStorage)
	server := httptest.NewServer(handler)
	defer server.Close()
	client := http.Client{}

	namespace := "my-namespace"

	simple := &apis.Workflow{
		TypeMeta: meta.TypeMeta{
			Kind:       "Workflow",
			APIVersion: "resources/v1",
		},
		ObjectMeta: meta.ObjectMeta{
			Name: "foo",
		},
		Spec: apis.WorkflowSpec{
			Name: "workflow-foo",
		},
		Status: apis.WorkflowStatus{
			WorkflowID: "123",
		},
	}
	data, err := runtime.Encode(codec, simple)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	bytesBuffer := bytes.NewBuffer(data)
	request, err := http.NewRequest("POST", server.URL+"/"+testPrefix+"/"+testGroupVersion.Group+"/"+testGroupVersion.Version+"/namespaces/"+namespace+"/workflows", bytesBuffer)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	wg := sync.WaitGroup{}
	wg.Add(1)
	var response *http.Response
	go func() {
		response, err = client.Do(request)
		wg.Done()
	}()
	wg.Wait()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	var itemOut apis.Workflow
	_, err = extractBody(response, &itemOut)
	if err != nil {
		t.Errorf("unexpected error: %v %#v", err, response)
	}
	if response.StatusCode != http.StatusCreated {
		t.Errorf("Unexpected status: %d, Expected: %d, %#v", response.StatusCode, http.StatusCreated, response)
	}
}

func TestNamespacedDelete(t *testing.T) {
	restStorage, embedEtcdServer := getTestRESTStorage(t)
	defer embedEtcdServer.Terminate(t)
	handler := handle(restStorage)
	ID := "foo"
	server := httptest.NewServer(handler)
	defer server.Close()
	client := http.Client{}

	//创建一个Workflow资源
	namespace := "my-namespace"
	simple := &apis.Workflow{
		TypeMeta: meta.TypeMeta{
			Kind:       "Workflow",
			APIVersion: "resources/v1",
		},
		ObjectMeta: meta.ObjectMeta{
			Name: "foo",
		},
		Spec: apis.WorkflowSpec{
			Name: "workflow-foo",
		},
		Status: apis.WorkflowStatus{
			WorkflowID: "123",
		},
	}
	data, err := runtime.Encode(codec, simple)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	bytesBuffer := bytes.NewBuffer(data)
	req, err := http.NewRequest("POST", server.URL+"/"+testPrefix+"/"+testGroupVersion.Group+"/"+testGroupVersion.Version+"/namespaces/"+namespace+"/workflows", bytesBuffer)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	wg := sync.WaitGroup{}
	wg.Add(1)
	var response *http.Response
	go func() {
		response, err = client.Do(req)
		wg.Done()
	}()
	wg.Wait()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	var itemOut apis.Workflow
	_, err = extractBody(response, &itemOut)
	if err != nil {
		t.Errorf("unexpected error: %v %#v", err, response)
	}
	if response.StatusCode != http.StatusCreated {
		t.Errorf("Unexpected status: %d, Expected: %d, %#v", response.StatusCode, http.StatusCreated, response)
	}

	//测试删除
	request, err := http.NewRequest("DELETE", server.URL+"/"+testPrefix+"/"+testGroupVersion.Group+"/"+testGroupVersion.Version+"/namespaces/"+namespace+"/workflows/"+ID, nil)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	res, err := client.Do(request)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.StatusCode != http.StatusOK {
		t.Errorf("unexpected response: %#v", res)
	}
}

func TestNamespacedGet(t *testing.T) {
	restStorage, embedEtcdServer := getTestRESTStorage(t)
	defer embedEtcdServer.Terminate(t)
	handler := handle(restStorage)
	ID := "foo"
	server := httptest.NewServer(handler)
	defer server.Close()

	client := http.Client{}

	//创建一个Workflow资源
	namespace := "my-namespace"
	simple := &apis.Workflow{
		TypeMeta: meta.TypeMeta{
			Kind:       "Workflow",
			APIVersion: "resources/v1",
		},
		ObjectMeta: meta.ObjectMeta{
			Name: "foo",
		},
		Spec: apis.WorkflowSpec{
			Name: "workflow-foo",
		},
		Status: apis.WorkflowStatus{
			WorkflowID: "123",
		},
	}
	data, err := runtime.Encode(codec, simple)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	bytesBuffer := bytes.NewBuffer(data)
	req, err := http.NewRequest("POST", server.URL+"/"+testPrefix+"/"+testGroupVersion.Group+"/"+testGroupVersion.Version+"/namespaces/"+namespace+"/workflows", bytesBuffer)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	wg := sync.WaitGroup{}
	wg.Add(1)
	var response *http.Response
	go func() {
		response, err = client.Do(req)
		wg.Done()
	}()
	wg.Wait()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if err != nil {
		t.Errorf("unexpected error: %v %#v", err, response)
	}
	if response.StatusCode != http.StatusCreated {
		t.Errorf("Unexpected status: %d, Expected: %d, %#v", response.StatusCode, http.StatusCreated, response)
	}

	//测试查询
	resp, err := http.Get(server.URL + "/" + testPrefix + "/" + testGroupVersion.Group + "/" + testGroupVersion.Version + "/namespaces/" + namespace + "/workflows/" + ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected response: %#v", resp)
	}
	var itemOut apis.Workflow
	body, err := extractBody(resp, &itemOut)
	if err != nil {
		t.Errorf("unexpected error: %v %#v", err, resp)
	}
	if itemOut.Name != simple.Name {
		t.Errorf("Unexpected data: %#v, expected %#v (%s)", itemOut, simple, string(body))
	}
}

func TestNamespacedUpdate(t *testing.T) {
	restStorage, embedEtcdServer := getTestRESTStorage(t)
	defer embedEtcdServer.Terminate(t)
	handler := handle(restStorage)
	ID := "foo"
	server := httptest.NewServer(handler)
	defer server.Close()
	client := http.Client{}

	//创建一个Workflow资源
	namespace := "my-namespace"
	simple := &apis.Workflow{
		TypeMeta: meta.TypeMeta{
			Kind:       "Workflow",
			APIVersion: "resources/v1",
		},
		ObjectMeta: meta.ObjectMeta{
			Name: "foo",
		},
		Spec: apis.WorkflowSpec{
			Name: "workflow-foo",
		},
		Status: apis.WorkflowStatus{
			WorkflowID: "123",
		},
	}
	data, err := runtime.Encode(codec, simple)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	bytesBuffer := bytes.NewBuffer(data)
	req, err := http.NewRequest("POST", server.URL+"/"+testPrefix+"/"+testGroupVersion.Group+"/"+testGroupVersion.Version+"/namespaces/"+namespace+"/workflows", bytesBuffer)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	wg := sync.WaitGroup{}
	wg.Add(1)
	var response *http.Response
	go func() {
		response, err = client.Do(req)
		wg.Done()
	}()
	wg.Wait()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if err != nil {
		t.Errorf("unexpected error: %v %#v", err, response)
	}
	if response.StatusCode != http.StatusCreated {
		t.Errorf("Unexpected status: %d, Expected: %d, %#v", response.StatusCode, http.StatusCreated, response)
	}
	//测试更新
	updateObj := &apis.Workflow{
		TypeMeta: meta.TypeMeta{
			Kind:       "Workflow",
			APIVersion: "resources/v1",
		},
		ObjectMeta: meta.ObjectMeta{
			Name: "foo",
		},
		Spec: apis.WorkflowSpec{
			Name: "updated",
		},
		Status: apis.WorkflowStatus{
			WorkflowID: "updated",
		},
	}
	data, err = runtime.Encode(codec, updateObj)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	request, err := http.NewRequest("PUT", server.URL+"/"+testPrefix+"/"+testGroupVersion.Group+"/"+testGroupVersion.Version+"/namespaces/"+namespace+"/workflows/"+ID, bytes.NewBuffer(data))
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	wg = sync.WaitGroup{}
	wg.Add(1)
	go func() {
		response, err = client.Do(request)
		wg.Done()
	}()
	wg.Wait()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	var itemOut apis.Workflow
	_, err = extractBody(response, &itemOut)
	if err != nil {
		t.Errorf("unexpected error: %v %#v", err, response)
	}
	if response.StatusCode != http.StatusOK {
		t.Errorf("Unexpected status: %d, Expected: %d, %#v", response.StatusCode, http.StatusOK, response)
	}
	if itemOut.Spec.Name != updateObj.Spec.Name || itemOut.Status.WorkflowID != updateObj.Status.WorkflowID {
		t.Errorf("Unexpected SpecName:%s,StatusWorkflowID:%s,Expected SpecName:%s,StatusWorkflowID:%s,", itemOut.Spec.Name, itemOut.Spec.Name, updateObj.Status.WorkflowID, updateObj.Status.WorkflowID)
	}
}

func TestNamespacedDeleteCollection(t *testing.T) {
	restStorage, embedEtcdServer := getTestRESTStorage(t)
	defer embedEtcdServer.Terminate(t)
	handler := handle(restStorage)
	server := httptest.NewServer(handler)
	defer server.Close()
	client := http.Client{}

	//创建两个Workflow资源
	namespace := "my-namespace"
	workflow1 := &apis.Workflow{
		TypeMeta: meta.TypeMeta{
			Kind:       "Workflow",
			APIVersion: "resources/v1",
		},
		ObjectMeta: meta.ObjectMeta{
			Name: "foo",
		},
		Spec: apis.WorkflowSpec{
			Name: "workflow-foo",
		},
		Status: apis.WorkflowStatus{
			WorkflowID: "123",
		},
	}
	data, err := runtime.Encode(codec, workflow1)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	bytesBuffer := bytes.NewBuffer(data)
	req, err := http.NewRequest("POST", server.URL+"/"+testPrefix+"/"+testGroupVersion.Group+"/"+testGroupVersion.Version+"/namespaces/"+namespace+"/workflows", bytesBuffer)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	wg := sync.WaitGroup{}
	wg.Add(1)
	var response *http.Response
	go func() {
		response, err = client.Do(req)
		wg.Done()
	}()
	wg.Wait()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if err != nil {
		t.Errorf("unexpected error: %v %#v", err, response)
	}
	if response.StatusCode != http.StatusCreated {
		t.Errorf("Unexpected status: %d, Expected: %d, %#v", response.StatusCode, http.StatusCreated, response)
	}

	workflow2 := &apis.Workflow{
		TypeMeta: meta.TypeMeta{
			Kind:       "Workflow",
			APIVersion: "resources/v1",
		},
		ObjectMeta: meta.ObjectMeta{
			Name: "foo1",
		},
		Spec: apis.WorkflowSpec{
			Name: "workflow-foo",
		},
		Status: apis.WorkflowStatus{
			WorkflowID: "123",
		},
	}
	data, err = runtime.Encode(codec, workflow2)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	bytesBuffer = bytes.NewBuffer(data)
	req, err = http.NewRequest("POST", server.URL+"/"+testPrefix+"/"+testGroupVersion.Group+"/"+testGroupVersion.Version+"/namespaces/"+namespace+"/workflows", bytesBuffer)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	wg = sync.WaitGroup{}
	wg.Add(1)
	go func() {
		response, err = client.Do(req)
		wg.Done()
	}()
	wg.Wait()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if err != nil {
		t.Errorf("unexpected error: %v %#v", err, response)
	}
	if response.StatusCode != http.StatusCreated {
		t.Errorf("Unexpected status: %d, Expected: %d, %#v", response.StatusCode, http.StatusCreated, response)
	}

	//测试批量删除
	request, err := http.NewRequest("DELETE", server.URL+"/"+testPrefix+"/"+testGroupVersion.Group+"/"+testGroupVersion.Version+"/namespaces/"+namespace+"/workflows", nil)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	res, err := client.Do(request)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.StatusCode != http.StatusOK {
		t.Errorf("unexpected response: %#v", res)
	}
}

func TestNamespacedList(t *testing.T) {
	restStorage, embedEtcdServer := getTestRESTStorage(t)
	defer embedEtcdServer.Terminate(t)
	handler := handle(restStorage)
	server := httptest.NewServer(handler)
	defer server.Close()

	client := http.Client{}

	//创建两个Workflow资源
	namespace := "my-namespace"
	workflow1 := &apis.Workflow{
		TypeMeta: meta.TypeMeta{
			Kind:       "Workflow",
			APIVersion: "resources/v1",
		},
		ObjectMeta: meta.ObjectMeta{
			Name: "foo",
		},
		Spec: apis.WorkflowSpec{
			Name: "workflow-foo",
		},
		Status: apis.WorkflowStatus{
			WorkflowID: "123",
		},
	}
	data, err := runtime.Encode(codec, workflow1)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	bytesBuffer := bytes.NewBuffer(data)
	req, err := http.NewRequest("POST", server.URL+"/"+testPrefix+"/"+testGroupVersion.Group+"/"+testGroupVersion.Version+"/namespaces/"+namespace+"/workflows", bytesBuffer)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	wg := sync.WaitGroup{}
	wg.Add(1)
	var response *http.Response
	go func() {
		response, err = client.Do(req)
		wg.Done()
	}()
	wg.Wait()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if err != nil {
		t.Errorf("unexpected error: %v %#v", err, response)
	}
	if response.StatusCode != http.StatusCreated {
		t.Errorf("Unexpected status: %d, Expected: %d, %#v", response.StatusCode, http.StatusCreated, response)
	}

	workflow2 := &apis.Workflow{
		TypeMeta: meta.TypeMeta{
			Kind:       "Workflow",
			APIVersion: "resources/v1",
		},
		ObjectMeta: meta.ObjectMeta{
			Name: "foo1",
		},
		Spec: apis.WorkflowSpec{
			Name: "workflow-foo",
		},
		Status: apis.WorkflowStatus{
			WorkflowID: "123",
		},
	}
	data, err = runtime.Encode(codec, workflow2)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	bytesBuffer = bytes.NewBuffer(data)
	req, err = http.NewRequest("POST", server.URL+"/"+testPrefix+"/"+testGroupVersion.Group+"/"+testGroupVersion.Version+"/namespaces/"+namespace+"/workflows", bytesBuffer)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	wg = sync.WaitGroup{}
	wg.Add(1)
	go func() {
		response, err = client.Do(req)
		wg.Done()
	}()
	wg.Wait()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if err != nil {
		t.Errorf("unexpected error: %v %#v", err, response)
	}
	if response.StatusCode != http.StatusCreated {
		t.Errorf("Unexpected status: %d, Expected: %d, %#v", response.StatusCode, http.StatusCreated, response)
	}

	//测试List
	url := server.URL + "/" + testPrefix + "/" + testGroupVersion.Group + "/" + testGroupVersion.Version + "/nodes"

	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("unexpected status: %d from url: %s, Expected: %d, %#v", resp.StatusCode, url, http.StatusOK, resp)
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		t.Logf("body: %s", string(body))
	}

	var nodeList apis.NodeList
	_, err = extractBody(resp, &nodeList)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	for _, item := range nodeList.Items {
		if item.Name != "foo" && item.Name != "foo1" {
			t.Errorf("get unexpected item :%s", item.Name)
		}
	}

}

func TestNamespacedPatch(t *testing.T) {
	restStorage, embedEtcdServer := getTestRESTStorage(t)
	defer embedEtcdServer.Terminate(t)
	handler := handle(restStorage)
	server := httptest.NewServer(handler)
	defer server.Close()
	client := http.Client{}
	//创建一个Workflow资源
	namespace := "my-namespace"
	simple := &apis.Workflow{
		TypeMeta: meta.TypeMeta{
			Kind:       "Workflow",
			APIVersion: "resources/v1",
		},
		ObjectMeta: meta.ObjectMeta{
			Name: "foo",
		},
		Spec: apis.WorkflowSpec{
			Name: "workflow-foo",
			Desc: apis.Description{
				Docs: "docs",
			},
		},
		Status: apis.WorkflowStatus{
			WorkflowID: "123",
		},
	}
	data, err := runtime.Encode(codec, simple)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	bytesBuffer := bytes.NewBuffer(data)
	req, err := http.NewRequest("POST", server.URL+"/"+testPrefix+"/"+testGroupVersion.Group+"/"+testGroupVersion.Version+"/namespaces/"+namespace+"/workflows", bytesBuffer)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	wg := sync.WaitGroup{}
	wg.Add(1)
	var response *http.Response
	go func() {
		response, err = client.Do(req)
		wg.Done()
	}()
	wg.Wait()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if err != nil {
		t.Errorf("unexpected error: %v %#v", err, response)
	}
	if response.StatusCode != http.StatusCreated {
		t.Errorf("Unexpected status: %d, Expected: %d, %#v", response.StatusCode, http.StatusCreated, response)
	}

	//测试Patch
	jsonPatchBytes := []byte(`[
		{ "op": "replace", "path": "/Spec/Name", "value": "boo" },
		{ "op": "remove", "path": "/Status/WorkflowID" }
	]`)
	mergePatchBytes := []byte(`{
 		"Spec": {
   		"Name": "boo"
		},
 		"Status": {
   		"WorkflowID": ""
 		}
	}`)

	testCases := []struct {
		ID         string
		patchType  string
		patchBytes []byte
	}{
		{
			ID:         "foo",
			patchType:  "application/json-patch+json",
			patchBytes: jsonPatchBytes,
		},
		{
			ID:         "foo",
			patchType:  "application/merge-patch+json",
			patchBytes: mergePatchBytes,
		},
	}

	for _, tc := range testCases {
		client := http.Client{}
		request, err := http.NewRequest("PATCH", server.URL+"/"+testPrefix+"/"+testGroupVersion.Group+"/"+testGroupVersion.Version+"/namespaces/"+namespace+"/workflows/"+tc.ID, bytes.NewReader(tc.patchBytes))
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		request.Header.Set("Content-Type", tc.patchType)
		response, err := client.Do(request)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if response.StatusCode != http.StatusOK {
			t.Errorf("Unexpected status: %d, Expected: %d, %#v", response.StatusCode, http.StatusOK, response)
		}
		var itemOut apis.Workflow
		_, err = extractBody(response, &itemOut)
		if err != nil {
			t.Errorf("unexpected error: %v %#v", err, response)
		}
		if itemOut.Spec.Name != "boo" || itemOut.Status.WorkflowID != "" {
			t.Errorf("Unexpected SpecName,StatusWorkflowID:%s,%s  Expected:\"boo\",\"\"", itemOut.Spec.Name, itemOut.Status.WorkflowID)
		}
		if itemOut.Spec.Desc.Docs != simple.Spec.Desc.Docs {
			t.Errorf("Unexpected modified field: Spec.Desc.Docs")
		}
	}
}

func TestNamespacedWatchList(t *testing.T) {
	restStorage, embedEtcdServer := getTestRESTStorage(t)
	defer embedEtcdServer.Terminate(t)
	handler := handle(restStorage)
	server := httptest.NewServer(handler)

	//查询WorkflowList最新的ResourceVersion
	namespace := "my-namespace"
	client := http.Client{}
	dest, _ := url.Parse(server.URL)
	dest.Path = "/" + testPrefix + "/" + testGroupVersion.Group + "/" + testGroupVersion.Version + "/namespaces/" + namespace + "/workflows"
	reqList, _ := http.NewRequest("GET", dest.String(), nil)
	responseList, err := client.Do(reqList)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if responseList.StatusCode != http.StatusOK {
		t.Errorf("Unexpected status: %d, Expected: %d, %#v", responseList.StatusCode, http.StatusOK, responseList)
	}
	var workflowList apis.WorkflowList
	_, err = extractBody(responseList, &workflowList)
	resourceVersion := workflowList.ResourceVersion

	//发送Watch请求
	dest, _ = url.Parse(server.URL)
	dest.Path = "/" + testPrefix + "/" + testGroupVersion.Group + "/" + testGroupVersion.Version + "/namespaces/" + namespace + "/workflows"
	dest.RawQuery = "watch=true&sendInitialEvents=false&resourceVersionMatch=NotOlderThan&resourceVersion=" + resourceVersion
	req, _ := http.NewRequest("GET", dest.String(), nil)
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Unexpected status: %d, Expected: %d, %#v", resp.StatusCode, http.StatusOK, resp)
	}

	//创建一个Workflow资源
	simple := &apis.Workflow{
		TypeMeta: meta.TypeMeta{
			Kind:       "Workflow",
			APIVersion: "resources/v1",
		},
		ObjectMeta: meta.ObjectMeta{
			Name: "watchTestWorkflow",
		},
		Spec: apis.WorkflowSpec{
			Name: "workflow-foo",
		},
		Status: apis.WorkflowStatus{
			WorkflowID: "123",
		},
	}
	data, err := runtime.Encode(codec, simple)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	bytesBuffer := bytes.NewBuffer(data)
	req, err = http.NewRequest("POST", server.URL+"/"+testPrefix+"/"+testGroupVersion.Group+"/"+testGroupVersion.Version+"/namespaces/"+namespace+"/workflows", bytesBuffer)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	wg := sync.WaitGroup{}
	wg.Add(1)
	var response *http.Response
	go func() {
		response, err = client.Do(req)
		wg.Done()
	}()
	wg.Wait()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if response.StatusCode != http.StatusCreated {
		t.Errorf("Unexpected status: %d, Expected: %d, %#v", response.StatusCode, http.StatusCreated, response)
	}
	response.Body.Close()

	var dataWatch []byte
	scanner := bufio.NewScanner(resp.Body)
	scanner.Scan()
	dataWatch = scanner.Bytes()

	//检查watchEvent
	var got watchJSON
	err = json.Unmarshal(dataWatch, &got)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if got.Type != watch.Added {
		t.Errorf("Unexpected event type: %s, Expected: %s", got.Type, watch.Added)
	}
	var itemOut apis.Workflow
	err = json.Unmarshal(got.Object, &itemOut)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if itemOut.Name != "watchTestWorkflow" {
		t.Errorf("Unexpected event Node name:%s,Expected:%s", itemOut.Name, "watchTestWorkflow")
	}
	resp.Body.Close()
	server.Close()
}

func TestNamespacedWatchSingleResource(t *testing.T) {
	restStorage, embedEtcdServer := getTestRESTStorage(t)
	defer embedEtcdServer.Terminate(t)
	handler := handle(restStorage)
	server := httptest.NewServer(handler)

	client := http.Client{}

	//创建一个Workflow资源
	namespace := "my-namespace"
	simple := &apis.Workflow{
		TypeMeta: meta.TypeMeta{
			Kind:       "Workflow",
			APIVersion: "resources/v1",
		},
		ObjectMeta: meta.ObjectMeta{
			Name: "watchTestWorkflow",
		},
		Spec: apis.WorkflowSpec{
			Name: "workflow-foo",
		},
		Status: apis.WorkflowStatus{
			WorkflowID: "123",
		},
	}
	data, err := runtime.Encode(codec, simple)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	bytesBuffer := bytes.NewBuffer(data)
	req, err := http.NewRequest("POST", server.URL+"/"+testPrefix+"/"+testGroupVersion.Group+"/"+testGroupVersion.Version+"/namespaces/"+namespace+"/workflows", bytesBuffer)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	wg := sync.WaitGroup{}
	wg.Add(1)
	var response *http.Response
	go func() {
		response, err = client.Do(req)
		wg.Done()
	}()
	wg.Wait()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if response.StatusCode != http.StatusCreated {
		t.Errorf("Unexpected status: %d, Expected: %d, %#v", response.StatusCode, http.StatusCreated, response)
	}

	//查询WorkflowList最新的ResourceVersion
	dest, _ := url.Parse(server.URL)
	dest.Path = "/" + testPrefix + "/" + testGroupVersion.Group + "/" + testGroupVersion.Version + "/namespaces/" + namespace + "/workflows"
	reqList, _ := http.NewRequest("GET", dest.String(), nil)
	responseList, err := client.Do(reqList)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if responseList.StatusCode != http.StatusOK {
		t.Errorf("Unexpected status: %d, Expected: %d, %#v", responseList.StatusCode, http.StatusOK, responseList)
	}
	var workflowList apis.WorkflowList
	_, err = extractBody(responseList, &workflowList)
	resourceVersion := workflowList.ResourceVersion

	//发送Watch请求
	dest, _ = url.Parse(server.URL)
	dest.Path = "/" + testPrefix + "/" + testGroupVersion.Group + "/" + testGroupVersion.Version + "/namespaces/" + namespace + "/workflows"
	dest.RawQuery = "watch=true&fieldSelector=metadata.name=watchTestWorkflow&sendInitialEvents=false&resourceVersionMatch=NotOlderThan&resourceVersion=" + resourceVersion
	req, _ = http.NewRequest("GET", dest.String(), nil)
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Unexpected status: %d, Expected: %d, %#v", resp.StatusCode, http.StatusOK, resp)
	}

	//新增一个TestWorkflow，测试是否只监听单个资源
	simple1 := &apis.Workflow{
		TypeMeta: meta.TypeMeta{
			Kind:       "Workflow",
			APIVersion: "resources/v1",
		},
		ObjectMeta: meta.ObjectMeta{
			Name: "watchTestWorkflow1",
		},
		Spec: apis.WorkflowSpec{
			Name: "workflow-foo",
		},
		Status: apis.WorkflowStatus{
			WorkflowID: "123",
		},
	}
	data, err = runtime.Encode(codec, simple1)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	bytesBuffer = bytes.NewBuffer(data)
	req, err = http.NewRequest("POST", server.URL+"/"+testPrefix+"/"+testGroupVersion.Group+"/"+testGroupVersion.Version+"/namespaces/"+namespace+"/workflows", bytesBuffer)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	wg = sync.WaitGroup{}
	wg.Add(1)
	go func() {
		response, err = client.Do(req)
		wg.Done()
	}()
	wg.Wait()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if response.StatusCode != http.StatusCreated {
		t.Errorf("Unexpected status: %d, Expected: %d, %#v", response.StatusCode, http.StatusCreated, response)
	}
	//更新TestWorkflow
	simple2 := &apis.Workflow{
		TypeMeta: meta.TypeMeta{
			Kind:       "Workflow",
			APIVersion: "resources/v1",
		},
		ObjectMeta: meta.ObjectMeta{
			Name: "watchTestWorkflow",
		},
		Status: apis.WorkflowStatus{
			WorkflowID: "456",
		},
	}
	data, err = runtime.Encode(codec, simple2)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	bytesBuffer = bytes.NewBuffer(data)
	req, err = http.NewRequest("PUT", server.URL+"/"+testPrefix+"/"+testGroupVersion.Group+"/"+testGroupVersion.Version+"/namespaces/"+namespace+"/workflows/"+simple.Name, bytesBuffer)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	response, err = client.Do(req)
	if response.StatusCode != http.StatusOK {
		t.Errorf("Unexpected status: %d, Expected: %d, %#v", response.StatusCode, http.StatusOK, response)
	}
	response.Body.Close()

	var dataWatch []byte
	scanner := bufio.NewScanner(resp.Body)
	scanner.Scan()
	dataWatch = scanner.Bytes()

	//检查watchEvent
	var got watchJSON
	err = json.Unmarshal(dataWatch, &got)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if got.Type != watch.Modified {
		t.Errorf("Unexpected event type: %s, Expected: %s", got.Type, watch.Modified)
	}
	var itemOut apis.Workflow
	err = json.Unmarshal(got.Object, &itemOut)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if itemOut.Status.WorkflowID != "456" {
		t.Errorf("Unexpected Spec StatusWorkflowID:%s,Expected:%s", itemOut.Status.WorkflowID, "456")
	}

	resp.Body.Close()
	server.Close()
}

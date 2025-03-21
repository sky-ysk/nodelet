package informer

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"sync"

	"hit.edu/framework/pkg/apimachinery/runtime"
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/client-go/clients"
	"hit.edu/framework/pkg/client-go/tools/cache"
	"hit.edu/framework/pkg/component-base/logs"
)

const ()

// 转发目的地的地址列表
var URL string

var SyncConfig Config

type Controller struct {
	indexer  cache.Indexer
	informer cache.Controller
}

func NewController(indexer cache.Indexer, informer cache.Controller) *Controller {
	return &Controller{
		informer: informer,
		indexer:  indexer,
	}
}
func (c *Controller) Run(workers int, stopCh chan struct{}) {
	go c.informer.Run(stopCh)
	if !cache.WaitForCacheSync(stopCh, c.informer.HasSynced) {
		return
	}
	<-stopCh
}

func writeToFile(data string) error {
	file, err := os.OpenFile("a.txt", os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(data)
	if err != nil {
		return err
	}

	return nil
}

func JudgeResourceType(obj interface{}) (runtime.Object, string, error) {
	switch obj.(type) {
	case *apis.Node:
		return &apis.Node{}, "nodes", nil
	case *apis.Event:
		return &apis.Event{}, "events", nil
	case *apis.Data:
		return &apis.Data{}, "datas", nil
	case *apis.Workflow:
		return &apis.Workflow{}, "workflows", nil
	case *apis.Task:
		return &apis.Task{}, "tasks", nil
	case *apis.Group:
		return &apis.Group{}, "groups", nil
	case *apis.Action:
		return &apis.Action{}, "actions", nil
	case *apis.Scene:
		return &apis.Scene{}, "acenes", nil
	case *apis.Device:
		return &apis.Device{}, "devices", nil
	case *apis.Resource_NodeList:
		return &apis.Resource_NodeList{}, "resources", nil
	default:
		return nil, "", fmt.Errorf("unknown resource type")

	}

}

func EncodeResourceType(obj interface{}) (runtime.Object, string, error) {

	switch obj.(type) {
	case *apis.Node:
		return obj.(*apis.Node), "nodes", nil
	case *apis.Event:
		return obj.(*apis.Event), "events", nil
	case *apis.Data:
		return obj.(*apis.Data), "datas", nil
	case *apis.Workflow:
		return obj.(*apis.Workflow), "workflows", nil
	case *apis.Task:
		return obj.(*apis.Task), "tasks", nil
	case *apis.Group:
		return obj.(*apis.Group), "groups", nil
	case *apis.Action:
		return obj.(*apis.Action), "actions", nil
	case *apis.Scene:
		return obj.(*apis.Scene), "acenes", nil
	case *apis.Device:
		return obj.(*apis.Device), "devices", nil
	case *apis.Resource_Node:
		return obj.(*apis.Resource_Node), "resources", nil
	default:
		return nil, "", fmt.Errorf("unknown resource type")

	}

}

func RemoveSyncLabel(obj runtime.Object) error {
	metaObj, ok := obj.(meta.Object)
	if !ok {
		return fmt.Errorf("object does not implement metav1.Object")
	}

	labels := metaObj.GetLabels()
	if labels == nil {
		return nil
	}
	delete(labels, "sync")

	metaObj.SetLabels(labels)
	return nil
}

// 三种事件的处理
func AddFunc(obj interface{}) {
	key, err := cache.MetaNamespaceKeyFunc(obj)
	if err != nil {
		logs.Infof("Failed to deal object: %v", err)
	}
	writeToFile(fmt.Sprintf("AddFunc:%s\n", key))
	object, types, err := EncodeResourceType(obj)
	RemoveSyncLabel(object)
	if err != nil {
		logs.Errorf("Error while Encode %v", err)
	}
	data, err := runtime.Encode(codec, object)
	if err != nil {
		logs.Errorf("Error while Encode %v", err)
	}

	bytesBuffer := bytes.NewBuffer(data)
	request, err := http.NewRequest("POST", URL+"/"+testPrefix+"/"+testGroupVersion.Group+"/"+testGroupVersion.Version+"/"+types, bytesBuffer)
	if err != nil {
		logs.Errorf("unexpected error: %v", err)
	}
	request.Header.Set("FlowType", "etcd")
	request.Header.Set("ClusterID", "pve2")
	client := &http.Client{}
	wg := sync.WaitGroup{}
	wg.Add(1)
	var response *http.Response
	go func() {
		response, err = client.Do(request)
		wg.Done()
	}()
	wg.Wait()
	if err != nil {
		logs.Errorf("unexpected error: %v", err)
	}

	repo, err := extractSimpleDecoder(response)
	if err != nil {
		fmt.Println(repo)
		logs.Errorf("unexpected error: %v %#v", err, response)
	}
	if response.StatusCode != http.StatusCreated {
		logs.Errorf("Unexpected status: %d, Expected: %d, %#v", response.StatusCode, http.StatusCreated, response)
	}

}
func UpdateFunc(old interface{}, new interface{}) {
	//TODO: 处理假更新
	key, err := cache.MetaNamespaceKeyFunc(new)
	if err != nil {
		logs.Infof("Failed to deal object: %v", err)
	}
	writeToFile(fmt.Sprintf("UpdateFunc:%s\n", key))

	object, types, err := EncodeResourceType(new)
	RemoveSyncLabel(object)
	if err != nil {
		logs.Errorf("Error while Encode %v", err)
	}
	data, err := runtime.Encode(codec, object)
	if err != nil {
		logs.Errorf("Error while Encode %v", err)
	}

	request, err := http.NewRequest("PUT", URL+"/"+testPrefix+"/"+testGroupVersion.Group+"/"+testGroupVersion.Version+"/"+types+"/"+key, bytes.NewBuffer(data))
	if err != nil {
		logs.Errorf("unexpected error: %v", err)
	}
	request.Header.Set("FlowType", "etcd")
	client := &http.Client{}
	wg := sync.WaitGroup{}
	wg.Add(1)
	var response *http.Response
	go func() {
		response, err = client.Do(request)
		wg.Done()
	}()
	wg.Wait()
	if err != nil {
		logs.Errorf("unexpected error: %v", err)
	}
	_, err = extractSimpleDecoder(response)
	if err != nil {
		logs.Errorf("unexpected error: %v %#v", err, response)
	}
	if response.StatusCode != http.StatusOK {
		logs.Errorf("Unexpected status: %d, Expected: %d, %#v", response.StatusCode, http.StatusOK, response)
	}
}
func DeleteFunc(obj interface{}) {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	if err != nil {
		logs.Infof("Failed to deal object: %v", err)
	}
	_, types, err := JudgeResourceType(obj.(cache.DeletedFinalStateUnknown).Obj)
	if err != nil {
		logs.Errorf("unexpected error: %v", err)
	}
	writeToFile(fmt.Sprintf("DeleteFunc:%s\n", key))
	request, err := http.NewRequest("DELETE", URL+"/"+testPrefix+"/"+testGroupVersion.Group+"/"+testGroupVersion.Version+"/"+types+"/"+key, nil)
	if err != nil {
		logs.Errorf("unexpected error: %v", err)
	}
	request.Header.Set("FlowType", "etcd")
	client := &http.Client{}
	wg := sync.WaitGroup{}
	wg.Add(1)
	var response *http.Response
	go func() {
		response, err = client.Do(request)
		wg.Done()
	}()
	wg.Wait()
	if err != nil {
		logs.Errorf("unexpected error: %v", err)
	}
	_, err = extractSimpleDecoder(response)
	if err != nil {
		logs.Errorf("unexpected error: %v %#v", err, response)
	}
	if response.StatusCode != http.StatusOK {
		logs.Errorf("Unexpected status: %d, Expected: %d, %#v", response.StatusCode, http.StatusOK, response)
	}
}

func CreateController(clientSet *clients.ClientSet, resource interface{}, namespace string) {
	//TODO：同时监视所有namespace
	lstOpts := func(options *meta.ListOptions) {
		options.LabelSelector = "sync=yes"
	}

	resourcetype, resourcename, err := JudgeResourceType(resource)
	if err != nil {
		logs.Errorf("unexpected error: %v ", err)
	}
	ListWatcher := cache.NewFilteredListWatchFromClient(clientSet.Core().RESTClient(), resourcename, namespace, lstOpts)
	sourceEventHandler := cache.ResourceEventHandlerFuncs{
		AddFunc:    AddFunc,
		UpdateFunc: UpdateFunc,
		DeleteFunc: DeleteFunc,
	}
	options := cache.InformerOptions{
		ListerWatcher: ListWatcher,
		ObjectType:    resourcetype,
		Handler:       sourceEventHandler,
		ResyncPeriod:  0,
		Indexers:      cache.Indexers{},
	}
	indexer, informer := cache.NewInformerWithOptions(options)
	// 创建Controller
	controller := NewController(indexer, informer)
	stop := make(chan struct{})
	defer close(stop)
	go controller.Run(1, stop)
	select {}
}

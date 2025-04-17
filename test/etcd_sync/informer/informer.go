package informer

import (
	"bytes"
	"fmt"
	"net/http"
	"sync"

	"hit.edu/framework/pkg/apimachinery/runtime"
	"hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/client-go/clients"
	"hit.edu/framework/pkg/client-go/tools/cache"
	"hit.edu/framework/pkg/component-base/logs"
)

const ()

// 转发目的地的地址列表
//var URL string

//var SyncConfig Config

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
func (t *Target[T]) AddFunc(obj interface{}) {
	key, err := cache.MetaNamespaceKeyFunc(obj)
	if err != nil {
		logs.Errorf("Failed to deal object: %v", err)
	}
	logs.Infof(fmt.Sprintf("AddFunc:%s\n", key))
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
	request, err := http.NewRequest("POST", t.URL+"/"+testPrefix+"/"+testGroupVersion.Group+"/"+testGroupVersion.Version+"/"+types, bytesBuffer)
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

	_, err = extractSimpleDecoder(response)
	if err != nil {
		logs.Errorf("unexpected error: %v %#v", err, response)
	}
	if response.StatusCode != http.StatusCreated {
		logs.Errorf("Unexpected status: %d, Expected: %d, %#v", response.StatusCode, http.StatusCreated, response)
	}

}
func (t *Target[T]) UpdateFunc(old interface{}, new interface{}) {
	//TODO: 处理假更新
	key, err := cache.MetaNamespaceKeyFunc(new)
	if err != nil {
		logs.Errorf("Failed to deal object: %v", err)
	}
	logs.Infof(fmt.Sprintf("UpdateFunc:%s\n", key))

	object, types, err := EncodeResourceType(new)
	RemoveSyncLabel(object)
	if err != nil {
		logs.Errorf("Error while Encode %v", err)
	}
	data, err := runtime.Encode(codec, object)
	if err != nil {
		logs.Errorf("Error while Encode %v", err)
	}

	request, err := http.NewRequest("PUT", t.URL+"/"+testPrefix+"/"+testGroupVersion.Group+"/"+testGroupVersion.Version+"/"+types+"/"+key, bytes.NewBuffer(data))
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
func (t *Target[T]) DeleteFunc(obj interface{}) {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	if err != nil {
		logs.Errorf("Failed to deal object: %v", err)
	}
	_, types, err := JudgeResourceType(obj.(cache.DeletedFinalStateUnknown).Obj)
	if err != nil {
		logs.Errorf("unexpected error: %v", err)
	}
	logs.Infof(fmt.Sprintf("DeleteFunc:%s\n", key))
	request, err := http.NewRequest("DELETE", t.URL+"/"+testPrefix+"/"+testGroupVersion.Group+"/"+testGroupVersion.Version+"/"+types+"/"+key, nil)
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

func (t *Target[T]) CreateController(clientSet *clients.ClientSet, resource interface{}, namespace string) {
	lstOpts := func(options *meta.ListOptions) {
		options.LabelSelector = "sync=yes"
	}

	resourcetype, resourcename, err := JudgeResourceType(resource)
	if err != nil {
		logs.Errorf("unexpected error: %v ", err)
	}
	ListWatcher := cache.NewFilteredListWatchFromClient(clientSet.Core().RESTClient(), resourcename, namespace, lstOpts)
	sourceEventHandler := cache.ResourceEventHandlerFuncs{
		AddFunc:    t.AddFunc,
		UpdateFunc: t.UpdateFunc,
		DeleteFunc: t.DeleteFunc,
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
	logs.Infof("Create Controller")
	controller := NewController(indexer, informer)
	stop := make(chan struct{})
	defer close(stop)
	go controller.Run(1, stop)
	select {}
}

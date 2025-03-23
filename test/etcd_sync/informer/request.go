package informer

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"sync"

	"hit.edu/framework/pkg/apimachinery/runtime"
	metav1 "hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/client-go/tools/cache"
	"hit.edu/framework/pkg/component-base/logs"
)

func Create(ctx context.Context, obj interface{}, opts metav1.CreateOptions) (interface{}, error) {
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
	return nil, nil
}
func Update(ctx context.Context, obj interface{}, opts metav1.UpdateOptions) (interface{}, error) {
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
	return nil, nil
}

func Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error {
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
	return nil
}
func Get(ctx context.Context, name string, opts metav1.GetOptions) (interface{}, error) {
	key := fmt.Sprintf("%s/%s", namespace, name)
	types := "Nodes"
	writeToFile(fmt.Sprintf("GetFunc: %s\n", key))

	// 构造 GET 请求
	request, err := http.NewRequest("GET", URL+"/"+testPrefix+"/"+testGroupVersion.Group+"/"+testGroupVersion.Version+"/"+types+"/"+key, nil)
	if err != nil {
		logs.Errorf("Failed to create HTTP request: %v", err)
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
		logs.Errorf("HTTP request failed: %v", err)
	}
	defer response.Body.Close()

	// 解析响应
	data, err := extractSimpleDecoder(response)
	if err != nil {
		logs.Errorf("Failed to parse response: %v %#v", err, response)
	}

	if response.StatusCode != http.StatusOK {
		logs.Errorf("Unexpected status: %d, Expected: %d, %#v", response.StatusCode, http.StatusOK, response)
	}

	return nil, nil
}

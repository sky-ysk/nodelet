package informer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	metav1 "hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/client-go/tools/cache"
	"hit.edu/framework/pkg/component-base/logs"
)

func (t *Target[T]) Create(ctx context.Context, obj T, opts metav1.CreateOptions) (T, error) {

	var result T
	data, err := json.Marshal(obj)
	if err != nil {
		logs.Errorf("Error while encoding object: %v", err)
		return result, err
	}
	logs.Infof("AddFunc to:%s\n", t.RemoteClusterID)
	_, types, err := EncodeResourceType(obj)
	if err != nil {
		logs.Errorf("Error while Encode %v", err)
	}

	bytesBuffer := bytes.NewBuffer(data)
	request, err := http.NewRequest("POST", t.URL+"/"+testPrefix+"/"+testGroupVersion.Group+"/"+testGroupVersion.Version+"/POST"+"/"+types, bytesBuffer)
	if err != nil {
		logs.Errorf("unexpected error: %v", err)
	}
	request.Header.Set("FlowType", "etcd")
	request.Header.Set("ClusterID", t.RemoteClusterID)
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

	if response.StatusCode != http.StatusCreated {
		logs.Errorf("Unexpected status: %d, Expected: %d, %#v", response.StatusCode, http.StatusCreated, response)
	}
	err = json.NewDecoder(response.Body).Decode(&result)
	if err != nil {
		logs.Errorf("Failed to decode response: %v", err)
		return result, err
	}

	return result, err
}

func (t *Target[T]) Update(ctx context.Context, new T, opts metav1.UpdateOptions) (T, error) {
	var result T
	key, err := cache.MetaNamespaceKeyFunc(new)
	if err != nil {
		logs.Infof("Failed to deal object: %v", err)
	}
	data, err := json.Marshal(new)
	if err != nil {
		logs.Errorf("Error while encoding object: %v", err)
		return result, err
	}
	_, types, err := EncodeResourceType(new)
	// RemoveSyncLabel(object)
	if err != nil {
		logs.Errorf("Error while Encode %v", err)
	}
	request, err := http.NewRequest("PUT", t.URL+"/"+testPrefix+"/"+testGroupVersion.Group+"/"+testGroupVersion.Version+"/PUT"+"/"+types+"/"+key, bytes.NewBuffer(data))
	if err != nil {
		logs.Errorf("unexpected error: %v", err)
	}
	request.Header.Set("FlowType", "etcd")
	request.Header.Set("ClusterID", t.RemoteClusterID)
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

	if response.StatusCode != http.StatusCreated {
		logs.Errorf("Unexpected status: %d, Expected: %d, %#v", response.StatusCode, http.StatusCreated, response)
	}
	err = json.NewDecoder(response.Body).Decode(&result)
	if err != nil {
		logs.Errorf("Failed to decode response: %v", err)
		return result, err
	}
	return result, nil
}

func (t *Target[T]) Delete(ctx context.Context, namespace string, name string, types string, opts metav1.DeleteOptions) error {
	logs.Infof("DeleteFunc:%s/%s\n", namespace, name)
	request, err := http.NewRequest("DELETE", t.URL+"/"+testPrefix+"/"+testGroupVersion.Group+"/"+testGroupVersion.Version+"/DELETE"+"/"+types+"/"+namespace+"/"+name, nil)
	if err != nil {
		logs.Errorf("unexpected error: %v", err)
	}
	request.Header.Set("FlowType", "etcd")
	request.Header.Set("ClusterID", t.RemoteClusterID)
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
	_, err = extractSimpleDecoder(response)
	if err != nil {
		logs.Errorf("unexpected error: %v %#v", err, response)
	}
	if response.StatusCode != http.StatusOK {
		logs.Errorf("Unexpected status: %d, Expected: %d, %#v", response.StatusCode, http.StatusOK, response)
	}
	return nil
}

func (t *Target[T]) Get(ctx context.Context, namespace string, name string, types string, opts metav1.GetOptions) (T, error) {
	var result T
	key := fmt.Sprintf("%s/%s", namespace, name)
	logs.Infof("GetFunc: %s\n", key)
	request, err := http.NewRequest("GET", t.URL+"/"+testPrefix+"/"+testGroupVersion.Group+"/"+testGroupVersion.Version+"/GET"+"/"+types+"/"+key, nil)
	if err != nil {
		logs.Errorf("Failed to create HTTP request: %v", err)
	}
	request.Header.Set("FlowType", "etcd")
	request.Header.Set("ClusterID", t.RemoteClusterID)
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
	if response.StatusCode != http.StatusCreated {
		logs.Errorf("Unexpected status: %d, Expected: %d, %#v", response.StatusCode, http.StatusCreated, response)
	}
	err = json.NewDecoder(response.Body).Decode(&result)
	if err != nil {
		logs.Errorf("Failed to decode response: %v", err)
		return result, err
	}
	return result, nil
}

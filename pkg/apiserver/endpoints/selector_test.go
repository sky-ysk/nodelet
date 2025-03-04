package endpoints

import (
	"bytes"
	"fmt"
	"hit.edu/framework/pkg/apimachinery/runtime"
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/apis/meta"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"
)

func TestLabelSelector(t *testing.T) {
	restStorage, embedEtcdServer := getTestRESTStorage(t)
	defer embedEtcdServer.Terminate(t)
	handler := handle(restStorage)
	server := httptest.NewServer(handler)
	defer server.Close()

	client := http.Client{}

	//创建三个Workflow资源
	namespace := "my-namespace"
	workflow1 := &apis.Workflow{
		TypeMeta: meta.TypeMeta{
			Kind:       "Workflow",
			APIVersion: "resources/v1",
		},
		ObjectMeta: meta.ObjectMeta{
			Name: "workflow1",
			Labels: map[string]string{
				"app": "app1",
				"foo": "foo1",
			},
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
			Name: "workflow2",
			Labels: map[string]string{
				"app": "app2",
				"foo": "foo2",
			},
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

	workflow3 := &apis.Workflow{
		TypeMeta: meta.TypeMeta{
			Kind:       "Workflow",
			APIVersion: "resources/v1",
		},
		ObjectMeta: meta.ObjectMeta{
			Name: "workflow3",
			Labels: map[string]string{
				"app": "app3",
				"foo": "foo3",
			},
		},
	}
	data, err = runtime.Encode(codec, workflow3)
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

	workflow4 := &apis.Workflow{
		TypeMeta: meta.TypeMeta{
			Kind:       "Workflow",
			APIVersion: "resources/v1",
		},
		ObjectMeta: meta.ObjectMeta{
			Name: "workflow4",
			Labels: map[string]string{
				"app": "app4",
				"foo": "foo3",
				"x":   "y",
			},
		},
	}
	data, err = runtime.Encode(codec, workflow4)
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

	workflow5 := &apis.Workflow{
		TypeMeta: meta.TypeMeta{
			Kind:       "Workflow",
			APIVersion: "resources/v1",
		},
		ObjectMeta: meta.ObjectMeta{
			Name: "workflow5",
			Labels: map[string]string{
				"app": "app5",
				"foo": "foo2",
				"x":   "y",
			},
		},
	}
	data, err = runtime.Encode(codec, workflow5)
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

	testCases := []struct {
		labelSelectorParams string
		expectedList        apis.WorkflowList
	}{
		{"app == app1", apis.WorkflowList{Items: []apis.Workflow{*workflow1}}},
		{"foo != foo2", apis.WorkflowList{Items: []apis.Workflow{*workflow1, *workflow3, *workflow4}}},
		{"app in ( app1 , app2 )", apis.WorkflowList{Items: []apis.Workflow{*workflow1, *workflow2}}},
		{"foo notin ( foo2,foo3)", apis.WorkflowList{Items: []apis.Workflow{*workflow1}}},
		{"x", apis.WorkflowList{Items: []apis.Workflow{*workflow4, *workflow5}}},
		{"!x", apis.WorkflowList{Items: []apis.Workflow{*workflow1, *workflow2, *workflow3}}},
		{"!x , app in ( app1 , app2 ),app==app2", apis.WorkflowList{Items: []apis.Workflow{*workflow2}}},
	}

	for _, tc := range testCases {
		labelSelector := tc.labelSelectorParams
		encodedSelector := url.QueryEscape(labelSelector)
		dest := fmt.Sprintf("%s/%s/%s/%s/namespaces/%s/workflows?labelSelector=%s", server.URL, testPrefix, testGroupVersion.Group, testGroupVersion.Version, namespace, encodedSelector)
		req, _ := http.NewRequest("GET", dest, nil)
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Errorf("unexpected status: %d from url: %s, Expected: %d, %#v", resp.StatusCode, dest, http.StatusOK, resp)
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			t.Logf("body: %s", string(body))
		}

		var workflowList apis.WorkflowList
		_, err = extractBody(resp, &workflowList)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		//先判断数组元素长度是否相同
		if len(workflowList.Items) != len(tc.expectedList.Items) {
			t.Errorf("unexpected error: %v", err)
		}

		for _, expectedItem := range tc.expectedList.Items {
			hasExpected := false
			for _, item := range workflowList.Items {
				if item.Name == expectedItem.Name {
					hasExpected = true
					break
				}
			}
			if !hasExpected {
				t.Errorf("workflowList not has the expected workflow: %v", expectedItem.Name)
			}
		}
	}

}

func TestFieldSelector(t *testing.T) {
	restStorage, embedEtcdServer := getTestRESTStorage(t)
	defer embedEtcdServer.Terminate(t)
	handler := handle(restStorage)
	server := httptest.NewServer(handler)
	defer server.Close()

	client := http.Client{}

	//创建三个Workflow资源
	namespace := "my-namespace"
	workflow1 := &apis.Workflow{
		TypeMeta: meta.TypeMeta{
			Kind:       "Workflow",
			APIVersion: "resources/v1",
		},
		ObjectMeta: meta.ObjectMeta{
			Name: "workflow1",
			Labels: map[string]string{
				"app": "app1",
				"foo": "foo1",
			},
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
			Name: "workflow2",
			Labels: map[string]string{
				"app": "app2",
				"foo": "foo2",
			},
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

	workflow3 := &apis.Workflow{
		TypeMeta: meta.TypeMeta{
			Kind:       "Workflow",
			APIVersion: "resources/v1",
		},
		ObjectMeta: meta.ObjectMeta{
			Name: "workflow3",
			Labels: map[string]string{
				"app": "app3",
				"foo": "foo3",
			},
		},
	}
	data, err = runtime.Encode(codec, workflow3)
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

	namespace = "my-namespace1"
	workflow4 := &apis.Workflow{
		TypeMeta: meta.TypeMeta{
			Kind:       "Workflow",
			APIVersion: "resources/v1",
		},
		ObjectMeta: meta.ObjectMeta{
			Name: "workflow4",
			Labels: map[string]string{
				"app": "app4",
				"foo": "foo3",
				"x":   "y",
			},
		},
	}
	data, err = runtime.Encode(codec, workflow4)
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

	workflow5 := &apis.Workflow{
		TypeMeta: meta.TypeMeta{
			Kind:       "Workflow",
			APIVersion: "resources/v1",
		},
		ObjectMeta: meta.ObjectMeta{
			Name: "workflow5",
			Labels: map[string]string{
				"app": "app5",
				"foo": "foo2",
				"x":   "y",
			},
		},
	}
	data, err = runtime.Encode(codec, workflow5)
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

	testCases := []struct {
		labelSelectorParams string
		expectedList        apis.WorkflowList
	}{
		{"metadata.name = workflow1", apis.WorkflowList{Items: []apis.Workflow{*workflow1}}},
		{"metadata.namespace = my-namespace1", apis.WorkflowList{Items: []apis.Workflow{*workflow4, *workflow5}}},
		{"metadata.namespace != my-namespace1", apis.WorkflowList{Items: []apis.Workflow{*workflow1, *workflow2, *workflow3}}},
		{"metadata.name=workflow1,metadata.namespace==my-namespace1", apis.WorkflowList{Items: []apis.Workflow{}}},
		{"metadata.name = workflow1, metadata.namespace == my-namespace ", apis.WorkflowList{Items: []apis.Workflow{*workflow1}}},
	}

	for _, tc := range testCases {
		labelSelector := tc.labelSelectorParams
		encodedSelector := url.QueryEscape(labelSelector)
		dest := fmt.Sprintf("%s/%s/%s/%s/workflows?fieldSelector=%s", server.URL, testPrefix, testGroupVersion.Group, testGroupVersion.Version, encodedSelector)
		req, _ := http.NewRequest("GET", dest, nil)
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Errorf("unexpected status: %d from url: %s, Expected: %d, %#v", resp.StatusCode, dest, http.StatusOK, resp)
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			t.Logf("body: %s", string(body))
		}

		var workflowList apis.WorkflowList
		_, err = extractBody(resp, &workflowList)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		//先判断数组元素长度是否相同
		if len(workflowList.Items) != len(tc.expectedList.Items) {
			t.Errorf("unexpected error: %v", err)
		}

		for _, expectedItem := range tc.expectedList.Items {
			hasExpected := false
			for _, item := range workflowList.Items {
				if item.Name == expectedItem.Name {
					hasExpected = true
					break
				}
			}
			if !hasExpected {
				t.Errorf("workflowList not has the expected workflow: %v", expectedItem.Name)
			}
		}
	}

}

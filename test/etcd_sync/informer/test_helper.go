package informer

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"hit.edu/framework/pkg/apimachinery/runtime"
	"hit.edu/framework/pkg/apimachinery/runtime/schema"
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/apis/legacyscheme"
	metav1 "hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/client-go/clients"
	"hit.edu/framework/pkg/component-base/logs"
)

func CreateTestTaskEvents(clientSet *clients.ClientSet) {
	tasksClient := clientSet.Core().Tasks("Test")

	task1 := &apis.Task{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "demo_task1",
			Namespace: "Test",
			Labels: map[string]string{
				"sync": "yes",
			},
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Task",
			APIVersion: "resources/v1",
		},
		Spec: apis.TaskSpec{
			Name: "demo-task",
		},
	}

	_, err := tasksClient.Create(context.TODO(), task1, metav1.CreateOptions{})
	if err != nil {
		logs.Infof("Failed to create task: %v", err)
	}
	time.Sleep(5 * time.Second)
	tasksClient.Delete(context.TODO(), "demo_task1", metav1.DeleteOptions{})
}

func CreateTestNodeEvents(clientSet *clients.ClientSet) {
	nodesClient := clientSet.Core().Nodes("Test")
	node := &apis.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "demo-nodes",
			Namespace: "Test",
			Labels: map[string]string{
				"sync": "yes",
			},
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Node",
			APIVersion: "resources/v1",
		},
		Spec: apis.NodeSpec{
			NodeName: "demo-node",
			HostName: "master",
		},
	}

	node2 := &apis.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "demo-node2",
			Namespace: "Test",
			Labels: map[string]string{
				"sync": "yes",
			},
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Node",
			APIVersion: "resources/v1",
		},
		Spec: apis.NodeSpec{
			NodeName: "demo-node",
		},
	}
	node3 := &apis.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "demo-node3",
			Namespace: "Test",
			Labels: map[string]string{
				"sync": "yes",
			},
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Node",
			APIVersion: "resources/v1",
		},
		Spec: apis.NodeSpec{
			NodeName: "demo-node",
		},
	}

	_, err := nodesClient.Create(context.TODO(), node, metav1.CreateOptions{})
	_, _ = nodesClient.Create(context.TODO(), node2, metav1.CreateOptions{})
	_, _ = nodesClient.Create(context.TODO(), node3, metav1.CreateOptions{})

	if err != nil {
		logs.Infof("Failed to create node: %v", err)
	}
	result, getErr := nodesClient.Get(context.TODO(), "demo-nodes", metav1.GetOptions{})
	if getErr != nil {
		logs.Errorf("Failed to get : %v", getErr)
	}

	result.Spec.NodeName = "updatedNodeName"
	_, updateErr := nodesClient.Update(context.TODO(), result, metav1.UpdateOptions{})
	if updateErr != nil {
		logs.Infof("Update failed: %v", updateErr)
	}
	time.Sleep(5 * time.Second)
	nodesClient.Delete(context.TODO(), "demo-nodes", metav1.DeleteOptions{})
	time.Sleep(5 * time.Second)
	nodesClient.Delete(context.TODO(), "demo-node2", metav1.DeleteOptions{})
	time.Sleep(5 * time.Second)
	nodesClient.Delete(context.TODO(), "demo-node3", metav1.DeleteOptions{})

}

var codecs = legacyscheme.Codecs
var codec = codecs.LegacyCodec()
var testPrefix = "apis"
var testAPIGroup = "resources"
var testAPIVersion = "v1"
var testGroupVersion = schema.GroupVersion{Group: testAPIGroup, Version: testAPIVersion}

func EventTestSender(URLs string) {
	node := &apis.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "demo-nodesaaa",
			Namespace: "Test",
			Labels: map[string]string{
				"sync": "true",
			},
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Node",
			APIVersion: "resources/v1",
		},
		Spec: apis.NodeSpec{
			NodeName: "demo-node",
			HostName: "master",
		},
	}
	data, err := runtime.Encode(codec, node)
	if err != nil {
		logs.Errorf("unexpected error: %v", err)
	}
	bytesBuffer := bytes.NewBuffer(data)
	request, err := http.NewRequest("POST", URL+"/"+testPrefix+"/"+testGroupVersion.Group+"/"+testGroupVersion.Version+"/nodes", bytesBuffer)
	if err != nil {
		logs.Errorf("unexpected error: %v", err)
	}
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
	var itemOut apis.Node
	_, err = extractBodyDecoder(response, &itemOut, codec)
	if err != nil {
		logs.Errorf("unexpected error: %v %#v", err, response)
	}
	if response.StatusCode != http.StatusCreated {
		logs.Errorf("Unexpected status: %d, Expected: %d, %#v", response.StatusCode, http.StatusCreated, response)
	}
}

func extractBodyDecoder(response *http.Response, object runtime.Object, decoder runtime.Decoder) (string, error) {
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return string(body), err
	}
	return string(body), runtime.DecodeInto(decoder, body, object)
}

func extractSimpleDecoder(response *http.Response) (string, error) {
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

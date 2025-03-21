package informer

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	apis "hit.edu/framework/pkg/apis/cores"
	metav1 "hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/client-go/clients"
)

func PutHandler(r *http.Request) error {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf("Failed to read request body")
	}
	obj := &apis.Node{}
	err = json.Unmarshal(body, obj)
	if err != nil {
		return fmt.Errorf("Failed to unmarshal request body")
	}

	clientSet, err := clients.NewForConfig(SyncConfig.Client)
	if err != nil {
		return fmt.Errorf("Failed to create client set: %v", err)
	}

	nodeClient := clientSet.Core().Nodes("Test")
	_, err = nodeClient.Update(context.TODO(), obj, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("Failed to update Node: %v", err)
	}
	return nil
}

func PostHandler(r *http.Request) error {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf("Failed to read request body")
	}
	obj := &apis.Node{}
	err = json.Unmarshal(body, obj)
	if err != nil {
		return fmt.Errorf("Failed to read request body")
	}

	clientSet, err := clients.NewForConfig(SyncConfig.Client)
	if err != nil {
		return fmt.Errorf("Failed to read request body")
	}

	nodeClient := clientSet.Core().Nodes("Test")
	_, err = nodeClient.Create(context.TODO(), obj, metav1.CreateOptions{})
	return nil
}

func DeleteHandler(r *http.Request) error {
	nodeName := r.URL.Query().Get("name")
	if nodeName == "" {
		return fmt.Errorf("Missing 'name' parameter in request URL")
	}

	clientSet, err := clients.NewForConfig(SyncConfig.Client)
	if err != nil {
		return fmt.Errorf("Failed to create client set: %v", err)
	}

	nodeClient := clientSet.Core().Nodes("Test")
	err = nodeClient.Delete(context.TODO(), nodeName, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("Failed to delete Node: %v", err)
	}
	return nil
}

func GetHandler(r *http.Request) (*apis.Node, error) {
	nodeName := r.URL.Query().Get("name")
	if nodeName == "" {
		return nil, fmt.Errorf("Missing 'name' parameter in request URL")
	}

	clientSet, err := clients.NewForConfig(SyncConfig.Client)
	if err != nil {
		return nil, fmt.Errorf("Failed to create client set: %v", err)
	}

	nodeClient := clientSet.Core().Nodes("Test")
	node, err := nodeClient.Get(context.TODO(), nodeName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("Failed to get Node: %v", err)
	}
	return node, nil
}

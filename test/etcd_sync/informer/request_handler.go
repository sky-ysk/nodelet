package informer

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"hit.edu/framework/pkg/client-go/clients"
)

func (c Sever_Config) PutHandler(r *http.Request) (interface{}, error) {
	// 获取请求URI并解析出资源类型
	requestURI := r.RequestURI
	segments := strings.Split(requestURI, "/")
	types := segments[5]

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("Failed to read request body: %v", err)
	}

	obj, err := GetResourceObj(types)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(body, obj)
	if err != nil {
		return nil, fmt.Errorf("Failed to unmarshal request body: %v", err)
	}

	clientSet, err := clients.NewForConfig(c.Client)
	if err != nil {
		return nil, fmt.Errorf("Failed to create client set: %v", err)
	}
	updatedObj, err := UpdateResource(clientSet, obj, types)
	if err != nil {
		return nil, fmt.Errorf("Failed to update resource: %v", err)
	}

	return updatedObj, nil
}

func (c Sever_Config) PostHandler(r *http.Request) (interface{}, error) {

	requestURI := r.RequestURI

	segments := strings.Split(requestURI, "/")
	types := segments[5]
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("Failed to read request body")
	}

	obj, err := GetResourceObj(types)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(body, obj)
	if err != nil {
		return nil, fmt.Errorf("Failed to read request body")
	}

	clientSet, err := clients.NewForConfig(c.Client)
	if err != nil {
		return nil, fmt.Errorf("Failed to read request body")
	}

	return CreateResource(clientSet, obj, types)

}

func (c Sever_Config) DeleteHandler(r *http.Request) error {
	requestURI := r.RequestURI
	segments := strings.Split(requestURI, "/")

	if len(segments) < 7 {
		return fmt.Errorf("invalid URL format, expected at least 7 segments, got %d", len(segments))
	}
	types := segments[5]
	namespace := segments[6]
	resourceName := segments[7]

	clientSet, err := clients.NewForConfig(c.Client)
	if err != nil {
		return fmt.Errorf("failed to create client set: %v", err)
	}

	err = DeleteResource(clientSet, types, namespace, resourceName)
	if err != nil {
		return fmt.Errorf("failed to delete %s: %v", types, err)
	}

	return nil
}

func (c Sever_Config) GetHandler(r *http.Request) (interface{}, error) {
	requestURI := r.RequestURI
	segments := strings.Split(requestURI, "/")

	if len(segments) < 7 {
		return nil, fmt.Errorf("invalid URL format, expected at least 7 segments, got %d", len(segments))
	}

	types := segments[5]
	namespace := segments[6]
	resourceName := segments[7]

	if resourceName == "" {
		return nil, fmt.Errorf("missing 'path' parameter in request URL")
	}
	clientSet, err := clients.NewForConfig(c.Client)
	if err != nil {
		return nil, fmt.Errorf("failed to create client set: %v", err)
	}

	resource, err := GetResource(clientSet, types, namespace, resourceName)
	if err != nil {
		return nil, fmt.Errorf("failed to get %s: %v", types, err)
	}

	return resource, nil
}

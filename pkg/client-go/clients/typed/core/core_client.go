package core

import (
	"net/http"

	"hit.edu/framework/pkg/client-go/rest"
)

type CoreInterface interface {
	RESTClient() rest.Interface
	NodesGetter
	WorkflowsGetter
	TasksGetter
	GroupsGetter
	ActionsGetter
	DevicesGetter
	DatasGetter
	ScenesGetter
}

type CoreClient struct {
	restClient rest.Interface
}

func (c *CoreClient) Nodes(namespace string) NodeInterface {
	return newNodes(c, namespace)
}

func (c *CoreClient) Workflows(namespace string) WorkflowInterface {
	// TODO:
	return newWorkflows(c, namespace)
}

func (c *CoreClient) Tasks(namespace string) TaskInterface {
	return newTasks(c, namespace)
}

func (c *CoreClient) Groups(namespace string) GroupInterface {
	return newGroups(c, namespace)
}

func (c *CoreClient) Actions(namespace string) ActionInterface {
	return newActions(c, namespace)
}

func (c *CoreClient) Devices(namespace string) DeviceInterface {
	return newDevices(c, namespace)
}

func (c *CoreClient) Datas(namespace string) DataInterface {
	return newDatas(c, namespace)
}

func (c *CoreClient) Scenes(namespace string) SceneInterface {
	return newScenes(c, namespace)
}

func (c *CoreClient) RESTClient() rest.Interface {
	if c == nil {
		return nil
	}
	return c.restClient
}

//在clientset中需要调用的两个newfor

func NewForConfig(c *rest.Config) (*CoreClient, error) {
	config := *c
	httpClient, err := rest.HTTPClientFor(&config)
	if err != nil {
		return nil, err
	}
	return NewForConfigAndClient(&config, httpClient)
}

// NewForConfigAndClient为给定的配置和http客户端创建一个新的CoreClient。
// 请注意，提供的http客户端优先于配置的传输值。
func NewForConfigAndClient(c *rest.Config, h *http.Client) (*CoreClient, error) {
	config := *c
	client, err := rest.RESTClientForConfigAndClient(&config, h)
	if err != nil {
		return nil, err
	}
	return &CoreClient{client}, nil
}

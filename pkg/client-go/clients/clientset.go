package clients

import (
	"net/http"

	"hit.edu/framework/pkg/client-go/clients/typed/core"
	"hit.edu/framework/pkg/client-go/rest"
)

// Group下所有资源的访问接口
type Interface interface {
	// Core
	Core() core.CoreInterface
}

// Group下所有的客户端
type ClientSet struct {
	// Core
	core *core.CoreClient
}

func (c *ClientSet) Core() core.CoreInterface {
	return c.core
}

func NewForConfig(c *rest.Config) (*ClientSet, error) {
	configShallowCopy := *c

	httpClient, err := rest.HTTPClientFor(&configShallowCopy)

	if err != nil {
		return nil, err
	}
	return NewForConfigAndClient(&configShallowCopy, httpClient)
}

func NewForConfigAndClient(c *rest.Config, httpClient *http.Client) (*ClientSet, error) {
	//configShallowCopy := *c

	var cs ClientSet
	var err error
	// TODO: ClientSet下子资源的Config配置
	cs.core, err = core.NewForConfigAndClient(c, httpClient)
	if err != nil {
		return nil, err
	}

	return &cs, err
}

// // New creates a new Clientset for the given RESTClient.
// func New(c rest.Interface) *ClientSet {
// 	var cs ClientSet
// 	cs.core = core.New(c)
// 	return &cs
// }

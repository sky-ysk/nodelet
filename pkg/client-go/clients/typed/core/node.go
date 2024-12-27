package core

import (
	"context"
	"hit.edu/framework/pkg/apimachinery/types"
	"hit.edu/framework/pkg/apimachinery/watch"
	metav1 "hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/client-go/clients/scheme"
	
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/client-go/gentype"
)

// TODO: 替换K8s相关组件

const (
	NodeResource string = "nodes"
)

type NodesGetter interface {
	Nodes(namespace string) NodeInterface
}

type NodeInterface interface {
	Create(ctx context.Context, node *apis.Node, opts metav1.CreateOptions) (*apis.Node, error)
	Update(ctx context.Context, node *apis.Node, opts metav1.UpdateOptions) (*apis.Node, error)
	//UpdateStatus(ctx context.Context, node *apis.Node, opts runtime.UpdateOptions) (*apis.Node, error)
	Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error
	Get(ctx context.Context, name string, opts metav1.GetOptions) (*apis.Node, error)
	List(ctx context.Context, opts metav1.ListOptions) (*apis.NodeList, error)
	Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *apis.Node, err error)
	// TODO: Apply
	// TODO: ApplyStatus
}

type nodes struct {
	*gentype.ClientWithList[*apis.Node, *apis.NodeList]
	// TODO: 完善逻辑
}

// newNodes returns a Nodes
func newNodes(c *CoreClient, namespace string) *nodes {
	// TODO: 完善逻辑
	return &nodes{
		gentype.NewClientWithList[*apis.Node, *apis.NodeList](
			NodeResource,
			c.RESTClient(),
			namespace,
			scheme.ParameterCodec,
			func() *apis.Node { return &apis.Node{} },
			func() *apis.NodeList { return &apis.NodeList{} },
		),
	}
}

// TODO: 使用Gentype抽象描述生成逻辑
// func (n nodes) Create(ctx context.Context, node *apis.Node, opts runtime.CreateOptions) (*apis.Node, error) {
// 	//TODO implement me
// 	panic("implement me")
// }

//
//func (n nodes) Update(ctx context.Context, node *apis.Node, opts metav1.UpdateOptions) (*apis.Node, error) {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (n nodes) UpdateStatus(ctx context.Context, node *apis.Node, opts metav1.UpdateOptions) (*apis.Node, error) {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (n nodes) Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (n nodes) DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (n nodes) Get(ctx context.Context, name string, opts metav1.GetOptions) (*apis.Node, error) {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (n nodes) List(ctx context.Context, opts metav1.ListOptions) (*apis.NodeList, error) {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (n nodes) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
//	//TODO implement me
//	panic("implement me")
//}

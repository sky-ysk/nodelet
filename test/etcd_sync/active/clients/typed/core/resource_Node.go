package core

import (
	"context"

	"hit.edu/framework/pkg/apimachinery/types"
	"hit.edu/framework/pkg/apimachinery/watch"
	metav1 "hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/client-go/clients/scheme"

	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/test/etcd_sync/active/gentype"
)

// TODO: 替换K8s相关组件

const (
	Resource_NodeResource string = "resource_Nodes"
)

type Resource_NodesGetter interface {
	Resource_Nodes(namespace string) Resource_NodeInterface
}

type Resource_NodeInterface interface {
	Create(ctx context.Context, resource_Node *apis.Resource_Node, opts metav1.CreateOptions) (*apis.Resource_Node, error)
	Update(ctx context.Context, resource_Node *apis.Resource_Node, opts metav1.UpdateOptions) (*apis.Resource_Node, error)
	//UpdateStatus(ctx context.Context, resource_Node *apis.Resource_Node, opts runtime.UpdateOptions) (*apis.Resource_Node, error)
	Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error
	Get(ctx context.Context, name string, opts metav1.GetOptions) (*apis.Resource_Node, error)
	List(ctx context.Context, opts metav1.ListOptions) (*apis.Resource_NodeList, error)
	Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *apis.Resource_Node, err error)

	// TODO: Apply
	// TODO: ApplyStatus
}

type resource_Nodes struct {
	*gentype.ClientWithList[*apis.Resource_Node, *apis.Resource_NodeList]
	// TODO: 完善逻辑
}

// newResource_Nodes returns a Resource_Nodes
func newResource_Nodes(c *CoreClient, namespace string) *resource_Nodes {
	// TODO: 完善逻辑
	return &resource_Nodes{
		gentype.NewClientWithList[*apis.Resource_Node, *apis.Resource_NodeList](
			Resource_NodeResource,
			c.RESTClient(),
			namespace,
			scheme.ParameterCodec,
			func() *apis.Resource_Node { return &apis.Resource_Node{} },
			func() *apis.Resource_NodeList { return &apis.Resource_NodeList{} },
		),
	}
}

// TODO: 使用Gentype抽象描述生成逻辑
// func (n resource_Nodes) Create(ctx context.Context, resource_Node *apis.Resource_Node, opts runtime.CreateOptions) (*apis.Resource_Node, error) {
// 	//TODO implement me
// 	panic("implement me")
// }

//
//func (n resource_Nodes) Update(ctx context.Context, resource_Node *apis.Resource_Node, opts metav1.UpdateOptions) (*apis.Resource_Node, error) {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (n resource_Nodes) UpdateStatus(ctx context.Context, resource_Node *apis.Resource_Node, opts metav1.UpdateOptions) (*apis.Resource_Node, error) {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (n resource_Nodes) Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (n resource_Nodes) DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (n resource_Nodes) Get(ctx context.Context, name string, opts metav1.GetOptions) (*apis.Resource_Node, error) {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (n resource_Nodes) List(ctx context.Context, opts metav1.ListOptions) (*apis.Resource_NodeList, error) {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (n resource_Nodes) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
//	//TODO implement me
//	panic("implement me")
//}

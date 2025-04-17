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
	RuntimeResource string = "runtimes"
)

type RuntimesGetter interface {
	Runtimes(namespace string) RuntimeInterface
}

type RuntimeInterface interface {
	Create(ctx context.Context, runtime *apis.Runtime, opts metav1.CreateOptions) (*apis.Runtime, error)
	Update(ctx context.Context, runtime *apis.Runtime, opts metav1.UpdateOptions) (*apis.Runtime, error)
	//UpdateStatus(ctx context.Context, runtime *apis.Runtime, opts runtime.UpdateOptions) (*apis.Runtime, error)
	Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error
	Get(ctx context.Context, name string, opts metav1.GetOptions) (*apis.Runtime, error)
	List(ctx context.Context, opts metav1.ListOptions) (*apis.RuntimeList, error)
	Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *apis.Runtime, err error)

	// TODO: Apply
	// TODO: ApplyStatus
}

type runtimes struct {
	*gentype.ClientWithList[*apis.Runtime, *apis.RuntimeList]
	// TODO: 完善逻辑
}

// newRuntimes returns a Runtimes
func newRuntimes(c *CoreClient, namespace string) *runtimes {
	// TODO: 完善逻辑
	return &runtimes{
		gentype.NewClientWithList[*apis.Runtime, *apis.RuntimeList](
			RuntimeResource,
			c.RESTClient(),
			namespace,
			scheme.ParameterCodec,
			func() *apis.Runtime { return &apis.Runtime{} },
			func() *apis.RuntimeList { return &apis.RuntimeList{} },
		),
	}
}

// TODO: 使用Gentype抽象描述生成逻辑
// func (n runtimes) Create(ctx context.Context, runtime *apis.Runtime, opts runtime.CreateOptions) (*apis.Runtime, error) {
// 	//TODO implement me
// 	panic("implement me")
// }

//
//func (n runtimes) Update(ctx context.Context, runtime *apis.Runtime, opts metav1.UpdateOptions) (*apis.Runtime, error) {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (n runtimes) UpdateStatus(ctx context.Context, runtime *apis.Runtime, opts metav1.UpdateOptions) (*apis.Runtime, error) {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (n runtimes) Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (n runtimes) DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (n runtimes) Get(ctx context.Context, name string, opts metav1.GetOptions) (*apis.Runtime, error) {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (n runtimes) List(ctx context.Context, opts metav1.ListOptions) (*apis.RuntimeList, error) {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (n runtimes) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
//	//TODO implement me
//	panic("implement me")
//}

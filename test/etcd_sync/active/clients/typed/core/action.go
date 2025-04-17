package core

import (
	"context"

	"hit.edu/framework/pkg/apimachinery/types"
	"hit.edu/framework/pkg/apimachinery/watch"
	metav1 "hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/apis/meta/internalversion/scheme"

	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/test/etcd_sync/active/gentype"
)

// TODO: 替换K8s相关组件

const (
	ActionResource string = "actions"
)

type ActionsGetter interface {
	Actions(namespace string) ActionInterface
}

type ActionInterface interface {
	Create(ctx context.Context, action *apis.Action, opts metav1.CreateOptions) (*apis.Action, error)
	Update(ctx context.Context, action *apis.Action, opts metav1.UpdateOptions) (*apis.Action, error)
	//UpdateStatus(ctx context.Context, action *apis.Action, opts runtime.UpdateOptions) (*apis.Action, error)
	Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error
	Get(ctx context.Context, name string, opts metav1.GetOptions) (*apis.Action, error)
	List(ctx context.Context, opts metav1.ListOptions) (*apis.ActionList, error)
	Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *apis.Action, err error)

	// TODO: Apply
	// TODO: ApplyStatus
}

type actions struct {
	*gentype.ClientWithList[*apis.Action, *apis.ActionList]
	// TODO: 完善逻辑
}

// newActions returns a Actions
func newActions(c *CoreClient, namespace string) *actions {
	// TODO: 完善逻辑
	return &actions{
		gentype.NewClientWithList[*apis.Action, *apis.ActionList](
			ActionResource,
			c.RESTClient(),
			namespace,
			scheme.ParameterCodec,
			func() *apis.Action { return &apis.Action{} },
			func() *apis.ActionList { return &apis.ActionList{} },
		),
	}
}

// TODO: 使用Gentype抽象描述生成逻辑
// func (n actions) Create(ctx context.Context, action *apis.Action, opts runtime.CreateOptions) (*apis.Action, error) {
// 	//TODO implement me
// 	panic("implement me")
// }

//
//func (n actions) Update(ctx context.Context, action *apis.Action, opts metav1.UpdateOptions) (*apis.Action, error) {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (n actions) UpdateStatus(ctx context.Context, action *apis.Action, opts metav1.UpdateOptions) (*apis.Action, error) {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (n actions) Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (n actions) DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (n actions) Get(ctx context.Context, name string, opts metav1.GetOptions) (*apis.Action, error) {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (n actions) List(ctx context.Context, opts metav1.ListOptions) (*apis.ActionList, error) {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (n actions) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
//	//TODO implement me
//	panic("implement me")
//}

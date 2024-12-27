package core

import (
	"context"
	"hit.edu/framework/pkg/apimachinery/types"
	"hit.edu/framework/pkg/apimachinery/watch"
	metav1 "hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/apis/meta/internalversion/scheme"
	
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/client-go/gentype"
)

// TODO: 替换K8s相关组件

const (
	GroupResource string = "groups"
)

type GroupsGetter interface {
	Groups(namespace string) GroupInterface
}

type GroupInterface interface {
	Create(ctx context.Context, group *apis.Group, opts metav1.CreateOptions) (*apis.Group, error)
	Update(ctx context.Context, group *apis.Group, opts metav1.UpdateOptions) (*apis.Group, error)
	//UpdateStatus(ctx context.Context, group *apis.Group, opts runtime.UpdateOptions) (*apis.Group, error)
	Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error
	Get(ctx context.Context, name string, opts metav1.GetOptions) (*apis.Group, error)
	List(ctx context.Context, opts metav1.ListOptions) (*apis.GroupList, error)
	Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *apis.Group, err error)
	
	// TODO: Apply
	// TODO: ApplyStatus
}

type groups struct {
	*gentype.ClientWithList[*apis.Group, *apis.GroupList]
	// TODO: 完善逻辑
}

// newGroups returns a Groups
func newGroups(c *CoreClient, namespace string) *groups {
	// TODO: 完善逻辑
	return &groups{
		gentype.NewClientWithList[*apis.Group, *apis.GroupList](
			GroupResource,
			c.RESTClient(),
			namespace,
			scheme.ParameterCodec,
			func() *apis.Group { return &apis.Group{} },
			func() *apis.GroupList { return &apis.GroupList{} },
		),
	}
}

// TODO: 使用Gentype抽象描述生成逻辑
// func (n groups) Create(ctx context.Context, group *apis.Group, opts runtime.CreateOptions) (*apis.Group, error) {
// 	//TODO implement me
// 	panic("implement me")
// }

//
//func (n groups) Update(ctx context.Context, group *apis.Group, opts metav1.UpdateOptions) (*apis.Group, error) {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (n groups) UpdateStatus(ctx context.Context, group *apis.Group, opts metav1.UpdateOptions) (*apis.Group, error) {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (n groups) Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (n groups) DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (n groups) Get(ctx context.Context, name string, opts metav1.GetOptions) (*apis.Group, error) {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (n groups) List(ctx context.Context, opts metav1.ListOptions) (*apis.GroupList, error) {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (n groups) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
//	//TODO implement me
//	panic("implement me")
//}

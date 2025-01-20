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
	ResourceRequirementResource string = "resourceRequirements"
)

type ResourceRequirementsGetter interface {
	ResourceRequirements(namespace string) ResourceRequirementInterface
}

type ResourceRequirementInterface interface {
	Create(ctx context.Context, resourceRequirement *apis.ResourceRequirement, opts metav1.CreateOptions) (*apis.ResourceRequirement, error)
	Update(ctx context.Context, resourceRequirement *apis.ResourceRequirement, opts metav1.UpdateOptions) (*apis.ResourceRequirement, error)
	//UpdateStatus(ctx context.Context, resourceRequirement *apis.ResourceRequirement, opts runtime.UpdateOptions) (*apis.ResourceRequirement, error)
	Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error
	Get(ctx context.Context, name string, opts metav1.GetOptions) (*apis.ResourceRequirement, error)
	List(ctx context.Context, opts metav1.ListOptions) (*apis.ResourceRequirementList, error)
	Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *apis.ResourceRequirement, err error)

	// TODO: Apply
	// TODO: ApplyStatus
}

type resourceRequirements struct {
	*gentype.ClientWithList[*apis.ResourceRequirement, *apis.ResourceRequirementList]
	// TODO: 完善逻辑
}

// newResourceRequirements returns a ResourceRequirements
func newResourceRequirements(c *CoreClient, namespace string) *resourceRequirements {
	// TODO: 完善逻辑
	return &resourceRequirements{
		gentype.NewClientWithList[*apis.ResourceRequirement, *apis.ResourceRequirementList](
			ResourceRequirementResource,
			c.RESTClient(),
			namespace,
			scheme.ParameterCodec,
			func() *apis.ResourceRequirement { return &apis.ResourceRequirement{} },
			func() *apis.ResourceRequirementList { return &apis.ResourceRequirementList{} },
		),
	}
}

// TODO: 使用Gentype抽象描述生成逻辑
// func (n resourceRequirements) Create(ctx context.Context, resourceRequirement *apis.ResourceRequirement, opts runtime.CreateOptions) (*apis.ResourceRequirement, error) {
// 	//TODO implement me
// 	panic("implement me")
// }

//
//func (n resourceRequirements) Update(ctx context.Context, resourceRequirement *apis.ResourceRequirement, opts metav1.UpdateOptions) (*apis.ResourceRequirement, error) {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (n resourceRequirements) UpdateStatus(ctx context.Context, resourceRequirement *apis.ResourceRequirement, opts metav1.UpdateOptions) (*apis.ResourceRequirement, error) {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (n resourceRequirements) Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (n resourceRequirements) DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (n resourceRequirements) Get(ctx context.Context, name string, opts metav1.GetOptions) (*apis.ResourceRequirement, error) {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (n resourceRequirements) List(ctx context.Context, opts metav1.ListOptions) (*apis.ResourceRequirementList, error) {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (n resourceRequirements) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
//	//TODO implement me
//	panic("implement me")
//}

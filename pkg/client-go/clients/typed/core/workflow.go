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
	WorkflowResource string = "workflows"
)

type WorkflowsGetter interface {
	Workflows(namespace string) WorkflowInterface
}

type WorkflowInterface interface {
	Create(ctx context.Context, workflow *apis.Workflow, opts metav1.CreateOptions) (*apis.Workflow, error)
	Update(ctx context.Context, workflow *apis.Workflow, opts metav1.UpdateOptions) (*apis.Workflow, error)
	//UpdateStatus(ctx context.Context, workflow *apis.Workflow, opts runtime.UpdateOptions) (*apis.Workflow, error)
	Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error
	Get(ctx context.Context, name string, opts metav1.GetOptions) (*apis.Workflow, error)
	List(ctx context.Context, opts metav1.ListOptions) (*apis.WorkflowList, error)
	Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *apis.Workflow, err error)
	
	// TODO: Apply
	// TODO: ApplyStatus
}

type workflows struct {
	*gentype.ClientWithList[*apis.Workflow, *apis.WorkflowList]
	// TODO: 完善逻辑
}

// newWorkflows returns a Workflows
func newWorkflows(c *CoreClient, namespace string) *workflows {
	// TODO: 完善逻辑
	return &workflows{
		gentype.NewClientWithList[*apis.Workflow, *apis.WorkflowList](
			WorkflowResource,
			c.RESTClient(),
			namespace,
			scheme.ParameterCodec,
			func() *apis.Workflow { return &apis.Workflow{} },
			func() *apis.WorkflowList { return &apis.WorkflowList{} },
		),
	}
}

// TODO: 使用Gentype抽象描述生成逻辑
// func (n workflows) Create(ctx context.Context, workflow *apis.Workflow, opts runtime.CreateOptions) (*apis.Workflow, error) {
// 	//TODO implement me
// 	panic("implement me")
// }

//
//func (n workflows) Update(ctx context.Context, workflow *apis.Workflow, opts metav1.UpdateOptions) (*apis.Workflow, error) {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (n workflows) UpdateStatus(ctx context.Context, workflow *apis.Workflow, opts metav1.UpdateOptions) (*apis.Workflow, error) {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (n workflows) Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (n workflows) DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (n workflows) Get(ctx context.Context, name string, opts metav1.GetOptions) (*apis.Workflow, error) {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (n workflows) List(ctx context.Context, opts metav1.ListOptions) (*apis.WorkflowList, error) {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (n workflows) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
//	//TODO implement me
//	panic("implement me")
//}

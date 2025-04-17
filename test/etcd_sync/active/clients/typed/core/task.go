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
	TaskResource string = "tasks"
)

type TasksGetter interface {
	Tasks(namespace string) TaskInterface
}

type TaskInterface interface {
	Create(ctx context.Context, task *apis.Task, opts metav1.CreateOptions) (*apis.Task, error)
	Update(ctx context.Context, task *apis.Task, opts metav1.UpdateOptions) (*apis.Task, error)
	//UpdateStatus(ctx context.Context, task *apis.Task, opts runtime.UpdateOptions) (*apis.Task, error)
	Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error
	Get(ctx context.Context, name string, opts metav1.GetOptions) (*apis.Task, error)
	List(ctx context.Context, opts metav1.ListOptions) (*apis.TaskList, error)
	Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *apis.Task, err error)

	// TODO: Apply
	// TODO: ApplyStatus
}

type tasks struct {
	*gentype.ClientWithList[*apis.Task, *apis.TaskList]
	// TODO: 完善逻辑
}

// newTasks returns a Tasks
func newTasks(c *CoreClient, namespace string) *tasks {
	// TODO: 完善逻辑
	return &tasks{
		gentype.NewClientWithList[*apis.Task, *apis.TaskList](
			TaskResource,
			c.RESTClient(),
			namespace,
			scheme.ParameterCodec,
			func() *apis.Task { return &apis.Task{} },
			func() *apis.TaskList { return &apis.TaskList{} },
		),
	}
}

// TODO: 使用Gentype抽象描述生成逻辑
// func (n tasks) Create(ctx context.Context, task *apis.Task, opts runtime.CreateOptions) (*apis.Task, error) {
// 	//TODO implement me
// 	panic("implement me")
// }

//
//func (n tasks) Update(ctx context.Context, task *apis.Task, opts metav1.UpdateOptions) (*apis.Task, error) {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (n tasks) UpdateStatus(ctx context.Context, task *apis.Task, opts metav1.UpdateOptions) (*apis.Task, error) {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (n tasks) Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (n tasks) DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (n tasks) Get(ctx context.Context, name string, opts metav1.GetOptions) (*apis.Task, error) {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (n tasks) List(ctx context.Context, opts metav1.ListOptions) (*apis.TaskList, error) {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (n tasks) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
//	//TODO implement me
//	panic("implement me")
//}

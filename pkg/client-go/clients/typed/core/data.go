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
	DataResource string = "datas"
)

type DatasGetter interface {
	Datas(namespace string) DataInterface
}

type DataInterface interface {
	Create(ctx context.Context, data *apis.Data, opts metav1.CreateOptions) (*apis.Data, error)
	Update(ctx context.Context, data *apis.Data, opts metav1.UpdateOptions) (*apis.Data, error)
	//UpdateStatus(ctx context.Context, data *apis.Data, opts runtime.UpdateOptions) (*apis.Data, error)
	Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error
	Get(ctx context.Context, name string, opts metav1.GetOptions) (*apis.Data, error)
	List(ctx context.Context, opts metav1.ListOptions) (*apis.DataList, error)
	Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *apis.Data, err error)

	// TODO: Apply
	// TODO: ApplyStatus
}

type datas struct {
	*gentype.ClientWithList[*apis.Data, *apis.DataList]
	// TODO: 完善逻辑
}

// newDatas returns a Datas
func newDatas(c *CoreClient, namespace string) *datas {
	// TODO: 完善逻辑
	return &datas{
		gentype.NewClientWithList[*apis.Data, *apis.DataList](
			DataResource,
			c.RESTClient(),
			namespace,
			scheme.ParameterCodec,
			func() *apis.Data { return &apis.Data{} },
			func() *apis.DataList { return &apis.DataList{} },
		),
	}
}

// TODO: 使用Gentype抽象描述生成逻辑
// func (n datas) Create(ctx context.Context, data *apis.Data, opts runtime.CreateOptions) (*apis.Data, error) {
// 	//TODO implement me
// 	panic("implement me")
// }

//
//func (n datas) Update(ctx context.Context, data *apis.Data, opts metav1.UpdateOptions) (*apis.Data, error) {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (n datas) UpdateStatus(ctx context.Context, data *apis.Data, opts metav1.UpdateOptions) (*apis.Data, error) {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (n datas) Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (n datas) DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (n datas) Get(ctx context.Context, name string, opts metav1.GetOptions) (*apis.Data, error) {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (n datas) List(ctx context.Context, opts metav1.ListOptions) (*apis.DataList, error) {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (n datas) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
//	//TODO implement me
//	panic("implement me")
//}

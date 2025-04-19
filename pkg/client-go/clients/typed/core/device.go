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
	DeviceResource string = "devices"
)

type DevicesGetter interface {
	Devices(namespace string) DeviceInterface
}

type DeviceInterface interface {
	Create(ctx context.Context, device *apis.Device, opts metav1.CreateOptions) (*apis.Device, error)
	Update(ctx context.Context, device *apis.Device, opts metav1.UpdateOptions) (*apis.Device, error)
	//UpdateStatus(ctx context.Context, device *apis.Device, opts runtime.UpdateOptions) (*apis.Device, error)
	Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error
	Get(ctx context.Context, name string, opts metav1.GetOptions) (*apis.Device, error)
	List(ctx context.Context, opts metav1.ListOptions) (*apis.DeviceList, error)
	Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *apis.Device, err error)

	// TODO: Apply
	// TODO: ApplyStatus
}

type devices struct {
	*gentype.ClientWithList[*apis.Device, *apis.DeviceList]
	// TODO: 完善逻辑
}

// newDevices returns a Devices
func newDevices(c *CoreClient, namespace string) *devices {
	// TODO: 完善逻辑
	return &devices{
		gentype.NewClientWithList[*apis.Device, *apis.DeviceList](
			DeviceResource,
			c.RESTClient(),
			namespace,
			scheme.ParameterCodec,
			func() *apis.Device { return &apis.Device{} },
			func() *apis.DeviceList { return &apis.DeviceList{} },
		),
	}
}

// TODO: 使用Gentype抽象描述生成逻辑
// func (n devices) Create(ctx context.Context, device *apis.Device, opts runtime.CreateOptions) (*apis.Device, error) {
// 	//TODO implement me
// 	panic("implement me")
// }

//
//func (n devices) Update(ctx context.Context, device *apis.Device, opts metav1.UpdateOptions) (*apis.Device, error) {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (n devices) UpdateStatus(ctx context.Context, device *apis.Device, opts metav1.UpdateOptions) (*apis.Device, error) {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (n devices) Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (n devices) DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (n devices) Get(ctx context.Context, name string, opts metav1.GetOptions) (*apis.Device, error) {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (n devices) List(ctx context.Context, opts metav1.ListOptions) (*apis.DeviceList, error) {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (n devices) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
//	//TODO implement me
//	panic("implement me")
//}

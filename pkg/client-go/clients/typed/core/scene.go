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
	SceneResource string = "scenes"
)

type ScenesGetter interface {
	Scenes(namespace string) SceneInterface
}

type SceneInterface interface {
	Create(ctx context.Context, scene *apis.Scene, opts metav1.CreateOptions) (*apis.Scene, error)
	Update(ctx context.Context, scene *apis.Scene, opts metav1.UpdateOptions) (*apis.Scene, error)
	//UpdateStatus(ctx context.Context, scene *apis.Scene, opts runtime.UpdateOptions) (*apis.Scene, error)
	Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error
	Get(ctx context.Context, name string, opts metav1.GetOptions) (*apis.Scene, error)
	List(ctx context.Context, opts metav1.ListOptions) (*apis.SceneList, error)
	Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *apis.Scene, err error)

	// TODO: Apply
	// TODO: ApplyStatus
}

type scenes struct {
	*gentype.ClientWithList[*apis.Scene, *apis.SceneList]
	// TODO: 完善逻辑
}

// newScenes returns a Scenes
func newScenes(c *CoreClient, namespace string) *scenes {
	// TODO: 完善逻辑
	return &scenes{
		gentype.NewClientWithList[*apis.Scene, *apis.SceneList](
			SceneResource,
			c.RESTClient(),
			namespace,
			scheme.ParameterCodec,
			func() *apis.Scene { return &apis.Scene{} },
			func() *apis.SceneList { return &apis.SceneList{} },
		),
	}
}

// TODO: 使用Gentype抽象描述生成逻辑
// func (n scenes) Create(ctx context.Context, scene *apis.Scene, opts runtime.CreateOptions) (*apis.Scene, error) {
// 	//TODO implement me
// 	panic("implement me")
// }

//
//func (n scenes) Update(ctx context.Context, scene *apis.Scene, opts metav1.UpdateOptions) (*apis.Scene, error) {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (n scenes) UpdateStatus(ctx context.Context, scene *apis.Scene, opts metav1.UpdateOptions) (*apis.Scene, error) {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (n scenes) Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (n scenes) DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (n scenes) Get(ctx context.Context, name string, opts metav1.GetOptions) (*apis.Scene, error) {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (n scenes) List(ctx context.Context, opts metav1.ListOptions) (*apis.SceneList, error) {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (n scenes) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
//	//TODO implement me
//	panic("implement me")
//}

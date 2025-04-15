package action

import (
	"context"

	"hit.edu/framework/pkg/apis/legacyscheme"
	"hit.edu/framework/pkg/apiserver/registry/storage/field"

	"hit.edu/framework/pkg/apimachinery/runtime"
)

type Strategy struct {
	runtime.ObjectTyper
}

var thisStrategy = &Strategy{legacyscheme.Scheme}

func (t Strategy) PrepareForCreate(ctx context.Context, obj runtime.Object) {
	//action, ok := obj.(*apis.Action)
	//if !ok {
	//	return
	//}
	////生成 actionid
	//if action.Status.ActionID == "" {
	//	action.Status.ActionID = string(action.ObjectMeta.UID)
	//}
}

func (t Strategy) PrepareForUpdate(ctx context.Context, obj, old runtime.Object) {}
func (t Strategy) Validate(ctx context.Context, obj runtime.Object) field.ErrorList {
	return nil
}
func (t Strategy) ValidateUpdate(ctx context.Context, obj, old runtime.Object) field.ErrorList {
	return nil
}
func (t Strategy) Canonicalize(obj runtime.Object) {}

// 资源的命名空间级别
func (t Strategy) NamespaceScoped() bool {
	// true表示资源必须配备命名空间信息
	return true
}

package workflow

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

func (t Strategy) PrepareForCreate(ctx context.Context, obj runtime.Object)      {}
func (t Strategy) PrepareForUpdate(ctx context.Context, obj, old runtime.Object) {}
func (t Strategy) Validate(ctx context.Context, obj runtime.Object) field.ErrorList {
	return nil
}
func (t Strategy) ValidateUpdate(ctx context.Context, obj, old runtime.Object) field.ErrorList {
	return nil
}
func (t Strategy) Canonicalize(obj runtime.Object) {}
func (t Strategy) NamespaceScoped() bool {
	return true
}

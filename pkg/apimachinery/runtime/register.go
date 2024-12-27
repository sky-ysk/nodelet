package runtime

import (
	"hit.edu/framework/pkg/apimachinery/runtime/schema"
)

// SetGroupVersionKind satisfies the ObjectKind interface for all objects that embed TypeMeta
func (obj *TypeMeta) SetGroupVersionKind(gvk schema.GroupVersionKind) {
	obj.APIVersion, obj.Kind = gvk.ToAPIVersionAndKind()
}

// GroupVersionKind satisfies the ObjectKind interface for all objects that embed TypeMeta
func (obj *TypeMeta) GroupVersionKind() schema.GroupVersionKind {
	return schema.FromAPIVersionAndKind(obj.APIVersion, obj.Kind)
}

const APIInternalVersion = "__internal"

func (obj *TypeMeta) GetObjectKind() schema.ObjectKind { return obj }

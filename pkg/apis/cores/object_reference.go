package apis

import "hit.edu/framework/pkg/apimachinery/runtime/schema"

func (obj ObjectReference) SetGroupVersionKind(gvk schema.GroupVersionKind) {
	obj.APIVersion, obj.Kind = gvk.ToAPIVersionAndKind()
}

func (obj ObjectReference) GroupVersionKind() schema.GroupVersionKind {
	return schema.FromAPIVersionAndKind(obj.APIVersion, obj.Kind)
}

func (obj ObjectReference) GetObjectKind() schema.ObjectKind { return obj }

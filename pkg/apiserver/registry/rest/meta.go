package rest

import (
	"github.com/google/uuid"
	"hit.edu/framework/pkg/apimachinery/errors"
	"hit.edu/framework/pkg/apimachinery/runtime/schema"
	"hit.edu/framework/pkg/apis/meta"
)

// WipeObjectMetaSystemFields 清除ObjectMeta中系统自动管理的字段
func WipeObjectMetaSystemFields(metas meta.Object) {
	metas.SetCreationTimestamp(meta.Time{})
	metas.SetUID("")
	metas.SetDeletionTimestamp(nil)
	metas.SetDeletionGracePeriodSeconds(nil)
	//metas.SetSelfLink("")
}

// EnsureObjectNamespaceMatchesRequestNamespace returns an error if obj.Namespace and requestNamespace
// are both populated and do not match. If either is unpopulated, it modifies obj as needed to ensure
// obj.GetNamespace() == requestNamespace.
func EnsureObjectNamespaceMatchesRequestNamespace(requestNamespace string, obj meta.Object) error {
	objNamespace := obj.GetNamespace()
	switch {
	case objNamespace == requestNamespace:
		// already matches, no-op
		return nil

	case objNamespace == "":
		// unset, default to request namespace
		obj.SetNamespace(requestNamespace)
		return nil

	case requestNamespace == "":
		// cluster-scoped, clear namespace
		obj.SetNamespace("")
		return nil

	default:
		// mismatch, error
		return errors.NewBadRequest("the namespace of the provided object does not match the namespace sent on the request")
	}
}

func ExpectedNamespaceForScope(requestNamespace string, namespaceScoped bool) string {
	if namespaceScoped {
		return requestNamespace
	}
	return ""
}

func ExpectedNamespaceForResource(requestNamespace string, resource schema.GroupVersionResource) string {
	if resource.Resource == "namespaces" && resource.Group == "" {
		return ""
	}
	return requestNamespace
}

func WipeObjectMetaSystemFields1(metas meta.Object) {
	metas.SetCreationTimestamp(meta.Time{})
	metas.SetUID("")
	metas.SetDeletionTimestamp(nil)
	metas.SetDeletionGracePeriodSeconds(nil)
}

// FillObjectMetaSystemFields 填充ObjectMeta中系统自动管理的字段
func FillObjectMetaSystemFields(metas meta.Object) {
	metas.SetCreationTimestamp(meta.Now())
	metas.SetUID(meta.UID(uuid.NewString()))
}
func FillObjectMetaSystemFields1(metas meta.Object) {
	metas.SetCreationTimestamp(meta.Now())
	uid1 := meta.UID(uuid.NewString())
	uid := meta.UID(string(uid1)) // 转换为 meta.UID 类型的值
	metas.SetUID(uid)
}

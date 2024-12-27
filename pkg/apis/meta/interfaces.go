package meta

import (
	"hit.edu/framework/pkg/apimachinery/runtime"
)

type List ListInterface

type MetadataAccessor interface {
	APIVersion(obj runtime.Object) (string, error)
	SetAPIVersion(obj runtime.Object, version string) error

	Kind(obj runtime.Object) (string, error)
	SetKind(obj runtime.Object, kind string) error

	Namespace(obj runtime.Object) (string, error)
	SetNamespace(obj runtime.Object, namespace string) error

	Name(obj runtime.Object) (string, error)
	SetName(obj runtime.Object, name string) error

	GenerateName(obj runtime.Object) (string, error)
	SetGenerateName(obj runtime.Object, name string) error

	UID(obj runtime.Object) (UID, error)
	SetUID(obj runtime.Object, uid UID) error

	Labels(obj runtime.Object) (map[string]string, error)
	SetLabels(obj runtime.Object, labels map[string]string) error
}

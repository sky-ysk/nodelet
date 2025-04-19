package meta

import (
	"hit.edu/framework/pkg/apimachinery/runtime"
	"hit.edu/framework/pkg/apimachinery/runtime/schema"

	"fmt"
)

func (obj *TypeMeta) GetObjectKind() schema.ObjectKind { return obj }

// SetGroupVersionKind 实现GetObjectKind接口
func (obj *TypeMeta) SetGroupVersionKind(gvk schema.GroupVersionKind) {
	obj.APIVersion, obj.Kind = gvk.ToAPIVersionAndKind()
}

// GroupVersionKind 实现GetObjectKind接口
func (obj *TypeMeta) GroupVersionKind() schema.GroupVersionKind {
	return schema.FromAPIVersionAndKind(obj.APIVersion, obj.Kind)
}

type ObjectMetaAccessor interface {
	GetObjectMeta() Object
}

type Object interface {
	GetNamespace() string
	SetNamespace(namespace string)
	GetName() string
	SetName(name string)
	GetGenerateName() string
	SetGenerateName(name string)
	GetUID() UID
	SetUID(uid UID)
	GetResourceVersion() string
	SetResourceVersion(version string)
	GetGeneration() int64
	SetGeneration(generation int64)
	GetCreationTimestamp() Time
	SetCreationTimestamp(timestamp Time)
	GetDeletionTimestamp() *Time
	SetDeletionTimestamp(timestamp *Time)
	GetDeletionGracePeriodSeconds() *int64
	SetDeletionGracePeriodSeconds(*int64)
	GetLabels() map[string]string
	SetLabels(labels map[string]string)
	GetAnnotations() map[string]string
	SetAnnotations(annotations map[string]string)
}

func (obj *ObjectMeta) GetObjectMeta() Object { return obj }

//允许快速、直接地访问API对象的元数据字段。

func (meta *ObjectMeta) GetNamespace() string                { return meta.Namespace }
func (meta *ObjectMeta) SetNamespace(namespace string)       { meta.Namespace = namespace }
func (meta *ObjectMeta) GetName() string                     { return meta.Name }
func (meta *ObjectMeta) SetName(name string)                 { meta.Name = name }
func (meta *ObjectMeta) GetGenerateName() string             { return meta.GenerateName }
func (meta *ObjectMeta) SetGenerateName(generateName string) { meta.GenerateName = generateName }
func (meta *ObjectMeta) GetUID() UID                         { return meta.UID }
func (meta *ObjectMeta) SetUID(uid UID)                      { meta.UID = uid }
func (meta *ObjectMeta) GetResourceVersion() string          { return meta.ResourceVersion }
func (meta *ObjectMeta) SetResourceVersion(version string)   { meta.ResourceVersion = version }
func (meta *ObjectMeta) GetGeneration() int64                { return meta.Generation }
func (meta *ObjectMeta) SetGeneration(generation int64)      { meta.Generation = generation }

func (meta *ObjectMeta) GetCreationTimestamp() Time { return meta.CreationTimestamp }
func (meta *ObjectMeta) SetCreationTimestamp(creationTimestamp Time) {
	meta.CreationTimestamp = creationTimestamp
}
func (meta *ObjectMeta) GetDeletionTimestamp() *Time { return meta.DeletionTimestamp }
func (meta *ObjectMeta) SetDeletionTimestamp(deletionTimestamp *Time) {
	meta.DeletionTimestamp = deletionTimestamp
}
func (meta *ObjectMeta) GetDeletionGracePeriodSeconds() *int64 {
	return meta.DeletionGracePeriodSeconds
}
func (meta *ObjectMeta) SetDeletionGracePeriodSeconds(deletionGracePeriodSeconds *int64) {
	meta.DeletionGracePeriodSeconds = deletionGracePeriodSeconds
}
func (meta *ObjectMeta) GetLabels() map[string]string                 { return meta.Labels }
func (meta *ObjectMeta) SetLabels(labels map[string]string)           { meta.Labels = labels }
func (meta *ObjectMeta) GetAnnotations() map[string]string            { return meta.Annotations }
func (meta *ObjectMeta) SetAnnotations(annotations map[string]string) { meta.Annotations = annotations }

type ListMetaAccessor interface {
	GetListMeta() ListInterface
}

type ListInterface interface {
	GetResourceVersion() string
	SetResourceVersion(version string)
	GetContinue() string
	SetContinue(c string)
	GetRemainingItemCount() *int64
	SetRemainingItemCount(c *int64)
}

var _ ListInterface = &ListMeta{}

func (obj *ListMeta) GetListMeta() ListInterface { return obj }

func (meta *ListMeta) GetResourceVersion() string        { return meta.ResourceVersion }
func (meta *ListMeta) SetResourceVersion(version string) { meta.ResourceVersion = version }
func (meta *ListMeta) GetContinue() string               { return meta.Continue }
func (meta *ListMeta) SetContinue(c string)              { meta.Continue = c }
func (meta *ListMeta) GetRemainingItemCount() *int64     { return meta.RemainingItemCount }
func (meta *ListMeta) SetRemainingItemCount(c *int64)    { meta.RemainingItemCount = c }

type objectAccessor struct {
	runtime.Object
}

func (obj objectAccessor) GetKind() string {
	return obj.GetObjectKind().GroupVersionKind().Kind
}

func (obj objectAccessor) SetKind(kind string) {
	gvk := obj.GetObjectKind().GroupVersionKind()
	gvk.Kind = kind
	obj.GetObjectKind().SetGroupVersionKind(gvk)
}

func (obj objectAccessor) GetAPIVersion() string {
	return obj.GetObjectKind().GroupVersionKind().GroupVersion().String()
}

func (obj objectAccessor) SetAPIVersion(version string) {
	gvk := obj.GetObjectKind().GroupVersionKind()
	gv, err := schema.ParseGroupVersion(version)
	if err != nil {
		gv = schema.GroupVersion{Version: version}
	}
	gvk.Group, gvk.Version = gv.Group, gv.Version
	obj.GetObjectKind().SetGroupVersionKind(gvk)
}

func NewAccessor() MetadataAccessor {
	return resourceAccessor{}
}

// resourceAccessor implements ResourceVersioner and SelfLinker.
type resourceAccessor struct{}

func (resourceAccessor) Kind(obj runtime.Object) (string, error) {
	return objectAccessor{obj}.GetKind(), nil
}

func (resourceAccessor) SetKind(obj runtime.Object, kind string) error {
	objectAccessor{obj}.SetKind(kind)
	return nil
}

func (resourceAccessor) APIVersion(obj runtime.Object) (string, error) {
	return objectAccessor{obj}.GetAPIVersion(), nil
}

func (resourceAccessor) SetAPIVersion(obj runtime.Object, version string) error {
	objectAccessor{obj}.SetAPIVersion(version)
	return nil
}

func (resourceAccessor) Namespace(obj runtime.Object) (string, error) {
	accessor, err := Accessor(obj)
	if err != nil {
		return "", err
	}
	return accessor.GetNamespace(), nil
}

func (resourceAccessor) SetNamespace(obj runtime.Object, namespace string) error {
	accessor, err := Accessor(obj)
	if err != nil {
		return err
	}
	accessor.SetNamespace(namespace)
	return nil
}

func (resourceAccessor) Name(obj runtime.Object) (string, error) {
	accessor, err := Accessor(obj)
	if err != nil {
		return "", err
	}
	return accessor.GetName(), nil
}

func (resourceAccessor) SetName(obj runtime.Object, name string) error {
	accessor, err := Accessor(obj)
	if err != nil {
		return err
	}
	accessor.SetName(name)
	return nil
}

func (resourceAccessor) GenerateName(obj runtime.Object) (string, error) {
	accessor, err := Accessor(obj)
	if err != nil {
		return "", err
	}
	return accessor.GetGenerateName(), nil
}

func (resourceAccessor) SetGenerateName(obj runtime.Object, name string) error {
	accessor, err := Accessor(obj)
	if err != nil {
		return err
	}
	accessor.SetGenerateName(name)
	return nil
}

func (resourceAccessor) UID(obj runtime.Object) (UID, error) {
	accessor, err := Accessor(obj)
	if err != nil {
		return "", err
	}
	return accessor.GetUID(), nil
}

func (resourceAccessor) SetUID(obj runtime.Object, uid UID) error {
	accessor, err := Accessor(obj)
	if err != nil {
		return err
	}
	accessor.SetUID(uid)
	return nil
}

func (resourceAccessor) Labels(obj runtime.Object) (map[string]string, error) {
	accessor, err := Accessor(obj)
	if err != nil {
		return nil, err
	}
	return accessor.GetLabels(), nil
}

func (resourceAccessor) SetLabels(obj runtime.Object, labels map[string]string) error {
	accessor, err := Accessor(obj)
	if err != nil {
		return err
	}
	accessor.SetLabels(labels)
	return nil
}

func Accessor(obj interface{}) (Object, error) {
	switch t := obj.(type) {
	case Object:
		return t, nil
	case ObjectMetaAccessor:
		if m := t.GetObjectMeta(); m != nil {
			return m, nil
		}
		return nil, errNotObject
	default:
		return nil, errNotObject
	}
}

func ListAccessor(obj interface{}) (List, error) {
	switch t := obj.(type) {
	case List:
		return t, nil
	case ListMetaAccessor:
		if m := t.GetListMeta(); m != nil {
			return m, nil
		}
		return nil, errNotList
	default:
		return nil, errNotList
	}
}

var errNotObject = fmt.Errorf("object does not implement the Object interfaces")
var errNotList = fmt.Errorf("object does not implement the List interfaces")

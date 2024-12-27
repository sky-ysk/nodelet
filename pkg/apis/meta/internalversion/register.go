package internalversion

import (
	"hit.edu/framework/pkg/apimachinery/runtime"
	"hit.edu/framework/pkg/apimachinery/runtime/schema"
	"hit.edu/framework/pkg/apis/meta"
)

const GroupName = "meta"

var (
	SchemeBuilder      runtime.SchemeBuilder
	localSchemeBuilder = &SchemeBuilder
	AddToScheme        = localSchemeBuilder.AddToScheme
)

var SchemeGroupVersion = schema.GroupVersion{Group: GroupName, Version: runtime.APIVersionInternal}

func Kind(kind string) schema.GroupKind {
	return SchemeGroupVersion.WithKind(kind).GroupKind()
}

// addToGroupVersion 将常见的元类型注册到 Schema 中。
func addToGroupVersion(scheme *runtime.Scheme) error {
	if err := scheme.AddIgnoredConversionType(&meta.TypeMeta{}, &meta.TypeMeta{}); err != nil {
		return err
	}
	scheme.AddKnownTypes(SchemeGroupVersion,
		&ListOptions{},
	)
	err := meta.AddTypes(scheme)
	if err != nil {
		return err
	}
	err = RegisterConversions(scheme)
	if err != nil {
		return err
	}
	return nil
}

func init() {
	localSchemeBuilder.Register(addToGroupVersion)
}

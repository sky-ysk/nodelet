package meta

import (
	"fmt"
	metav1 "hit.edu/framework/pkg/apis/meta"
)

// Accessor takes an arbitrary object pointer and returns meta.Interface.
// obj must be a pointer to an API type. An error is returned if the minimum
// required fields are missing. Fields that are not required return the default
// value and are a no-op if set.
func Accessor(obj interface{}) (metav1.Object, error) {
	switch t := obj.(type) {
	case metav1.Object:
		return t, nil
	case metav1.ObjectMetaAccessor:
		if m := t.GetObjectMeta(); m != nil {
			return m, nil
		}
		return nil, errNotObject
	default:
		return nil, errNotObject
	}
}

// errNotObject is returned when an object implements the List style interfaces but not the Object style
// interfaces.
var errNotObject = fmt.Errorf("object does not implement the Object interfaces")

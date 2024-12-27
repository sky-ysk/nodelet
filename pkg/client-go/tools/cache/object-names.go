package cache

import "hit.edu/framework/pkg/apimachinery/types"

// ObjectName is a reference to an object of some implicit kind
type ObjectName struct {
	Namespace string
	Name      string
}

// NewObjectName constructs a new one
func NewObjectName(namespace, name string) ObjectName {
	return ObjectName{Namespace: namespace, Name: name}
}

// Parts is the inverse of the constructor
func (objName ObjectName) Parts() (namespace, name string) {
	return objName.Namespace, objName.Name
}

// String returns the standard string encoding,
// which is designed to match the historical behavior of MetaNamespaceKeyFunc.
// Note this behavior is different from the String method of types.NamespacedName.
func (objName ObjectName) String() string {
	if len(objName.Namespace) > 0 {
		return objName.Namespace + "/" + objName.Name
	}
	return objName.Name
}

// ParseObjectName tries to parse the standard encoding
func ParseObjectName(str string) (ObjectName, error) {
	var objName ObjectName
	var err error
	objName.Namespace, objName.Name, err = SplitMetaNamespaceKey(str)
	return objName, err
}

// NamespacedNameAsObjectName rebrands the given NamespacedName as an ObjectName
func NamespacedNameAsObjectName(nn types.NamespacedName) ObjectName {
	return NewObjectName(nn.Namespace, nn.Name)
}

// AsNamespacedName rebrands as a NamespacedName
func (objName ObjectName) AsNamespacedName() types.NamespacedName {
	return types.NamespacedName{Namespace: objName.Namespace, Name: objName.Name}
}

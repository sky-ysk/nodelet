package schema

// From K8s
type ObjectKind interface {
	SetGroupVersionKind(kind GroupVersionKind)
	GroupVersionKind() GroupVersionKind
}

// From K8s
type GroupVersionKind struct {
	Group   string
	Version string
	Kind    string
}

// EmptyObjectKind implements the ObjectKind interface as a noop
var EmptyObjectKind = emptyObjectKind{}

type emptyObjectKind struct{}

// SetGroupVersionKind implements the ObjectKind interface
func (emptyObjectKind) SetGroupVersionKind(gvk GroupVersionKind) {}

// GroupVersionKind implements the ObjectKind interface
func (emptyObjectKind) GroupVersionKind() GroupVersionKind { return GroupVersionKind{} }

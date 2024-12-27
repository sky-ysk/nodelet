package types

type NamespacedName struct {
	Namespace string
	Name      string
}

const (
	Separator = '/'
)

// String returns the general purpose string representation
func (n NamespacedName) String() string {
	return n.Namespace + string(Separator) + n.Name
}

// MarshalLog emits a struct containing required key/value pair
func (n NamespacedName) MarshalLog() interface{} {
	return struct {
		Name      string `json:"name"`
		Namespace string `json:"namespace,omitempty"`
	}{
		Name:      n.Name,
		Namespace: n.Namespace,
	}
}

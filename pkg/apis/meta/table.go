// from k8s
package meta

import "hit.edu/framework/pkg/apimachinery/runtime"

// TODO: Table does not generate to protobuf because of the interface{} - fix protobuf
//   generation to support a meta type that can accept any valid JSON. This can be introduced
//   in a v1 because clients a) receive an error if they try to access proto today, and b)
//   once introduced they would be able to gracefully switch over to using it.

// Table is a tabular representation of a set of API resources. The server transforms the
// object into a set of preferred columns for quickly reviewing the objects.
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +protobuf=false
type Table struct {
	TypeMeta `json:",inline"`
	// Standard list metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds
	// +optional
	ListMeta `json:"metadata,omitempty"`
	
	// columnDefinitions describes each column in the returned items array. The number of cells per row
	// will always match the number of column definitions.
	// +listType=atomic
	ColumnDefinitions []TableColumnDefinition `json:"columnDefinitions"`
	// rows is the list of items in the table.
	// +listType=atomic
	Rows []TableRow `json:"rows"`
}

// TableColumnDefinition contains information about a column returned in the Table.
// +protobuf=false
type TableColumnDefinition struct {
	// name is a human readable name for the column.
	Name string `json:"name"`
	// type is an OpenAPI type definition for this column, such as number, integer, string, or
	// array.
	// See https://github.com/OAI/OpenAPI-Specification/blob/master/versions/2.0.md#data-types for more.
	Type string `json:"type"`
	// format is an optional OpenAPI type modifier for this column. A format modifies the type and
	// imposes additional rules, like date or time formatting for a string. The 'name' format is applied
	// to the primary identifier column which has type 'string' to assist in clients identifying column
	// is the resource name.
	// See https://github.com/OAI/OpenAPI-Specification/blob/master/versions/2.0.md#data-types for more.
	Format string `json:"format"`
	// description is a human readable description of this column.
	Description string `json:"description"`
	// priority is an integer defining the relative importance of this column compared to others. Lower
	// numbers are considered higher priority. Columns that may be omitted in limited space scenarios
	// should be given a higher priority.
	Priority int32 `json:"priority"`
}

// TableRow is an individual row in a table.
// +protobuf=false
type TableRow struct {
	// cells will be as wide as the column definitions array and may contain strings, numbers (float64 or
	// int64), booleans, simple maps, lists, or null. See the type field of the column definition for a
	// more detailed description.
	// +listType=atomic
	Cells []interface{} `json:"cells"`
	// conditions describe additional status of a row that are relevant for a human user. These conditions
	// apply to the row, not to the object, and will be specific to table output. The only defined
	// condition type is 'Completed', for a row that indicates a resource that has run to completion and
	// can be given less visual priority.
	// +optional
	// +listType=atomic
	Conditions []TableRowCondition `json:"conditions,omitempty"`
	// This field contains the requested additional information about each object based on the includeObject
	// policy when requesting the Table. If "None", this field is empty, if "Object" this will be the
	// default serialization of the object for the current API version, and if "Metadata" (the default) will
	// contain the object metadata. Check the returned kind and apiVersion of the object before parsing.
	// The media type of the object will always match the enclosing list - if this as a JSON table, these
	// will be JSON encoded objects.
	// +optional
	Object runtime.RawExtension `json:"object,omitempty"`
}

// TableRowCondition allows a row to be marked with additional information.
// +protobuf=false
type TableRowCondition struct {
	// Type of row condition. The only defined value is 'Completed' indicating that the
	// object this row represents has reached a completed state and may be given less visual
	// priority than other rows. Clients are not required to honor any conditions but should
	// be consistent where possible about handling the conditions.
	Type RowConditionType `json:"type"`
	// Status of the condition, one of True, False, Unknown.
	Status ConditionStatus `json:"status"`
	// (brief) machine readable reason for the condition's last transition.
	// +optional
	Reason string `json:"reason,omitempty"`
	// Human readable message indicating details about last transition.
	// +optional
	Message string `json:"message,omitempty"`
}

type RowConditionType string

// These are valid conditions of a row. This list is not exhaustive and new conditions may be
// included by other resources.
const (
	// RowCompleted means the underlying resource has reached completion and may be given less
	// visual priority than other resources.
	RowCompleted RowConditionType = "Completed"
)

type ConditionStatus string

// These are valid condition statuses. "ConditionTrue" means a resource is in the condition.
// "ConditionFalse" means a resource is not in the condition. "ConditionUnknown" means kubernetes
// can't decide if a resource is in the condition or not. In the future, we could add other
// intermediate conditions, e.g. ConditionDegraded.
const (
	ConditionTrue    ConditionStatus = "True"
	ConditionFalse   ConditionStatus = "False"
	ConditionUnknown ConditionStatus = "Unknown"
)

// IncludeObjectPolicy controls which portion of the object is returned with a Table.
type IncludeObjectPolicy string

const (
	// IncludeNone returns no object.
	IncludeNone IncludeObjectPolicy = "None"
	// IncludeMetadata serializes the object containing only its metadata field.
	IncludeMetadata IncludeObjectPolicy = "Metadata"
	// IncludeObject contains the full object.
	IncludeObject IncludeObjectPolicy = "Object"
)

// TableOptions are used when a Table is requested by the caller.
// +k8s:conversion-gen:explicit-from=net/url.Values
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type TableOptions struct {
	TypeMeta `json:",inline"`
	
	// NoHeaders is only exposed for internal callers. It is not included in our OpenAPI definitions
	// and may be removed as a field in a future release.
	NoHeaders bool `json:"-"`
	
	// includeObject decides whether to include each object along with its columnar information.
	// Specifying "None" will return no object, specifying "Object" will return the full object contents, and
	// specifying "Metadata" (the default) will return the object's metadata in the PartialObjectMetadata kind
	// in version v1beta1 of the meta.k8s.io API group.
	IncludeObject IncludeObjectPolicy `json:"includeObject,omitempty" protobuf:"bytes,1,opt,name=includeObject,casttype=IncludeObjectPolicy"`
}

// PartialObjectMetadata is a generic representation of any object with ObjectMeta. It allows clients
// to get access to a particular ObjectMeta schema without knowing the details of the version.
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type PartialObjectMetadata struct {
	TypeMeta `json:",inline"`
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	// +optional
	ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
}

// PartialObjectMetadataList contains a list of objects containing only their metadata
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type PartialObjectMetadataList struct {
	TypeMeta `json:",inline"`
	// Standard list metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds
	// +optional
	ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	
	// items contains each of the included items.
	Items []PartialObjectMetadata `json:"items" protobuf:"bytes,2,rep,name=items"`
}

// Condition contains details for one aspect of the current state of this API Resource.
// ---
// This struct is intended for direct use as an array at the field path .status.conditions.  For example,
//
//	type FooStatus struct{
//	    // Represents the observations of a foo's current state.
//	    // Known .status.conditions.type are: "Available", "Progressing", and "Degraded"
//	    // +patchMergeKey=type
//	    // +patchStrategy=merge
//	    // +listType=map
//	    // +listMapKey=type
//	    Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,1,rep,name=conditions"`
//
//	    // other fields
//	}
type Condition struct {
	// type of condition in CamelCase or in foo.example.com/CamelCase.
	// ---
	// Many .condition.type values are consistent across resources like Available, but because arbitrary conditions can be
	// useful (see .node.status.conditions), the ability to deconflict is important.
	// The regex it matches is (dns1123SubdomainFmt/)?(qualifiedNameFmt)
	// +required
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern=`^([a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*/)?(([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9])$`
	// +kubebuilder:validation:MaxLength=316
	Type string `json:"type" protobuf:"bytes,1,opt,name=type"`
	// status of the condition, one of True, False, Unknown.
	// +required
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=True;False;Unknown
	Status ConditionStatus `json:"status" protobuf:"bytes,2,opt,name=status"`
	// observedGeneration represents the .metadata.generation that the condition was set based upon.
	// For instance, if .metadata.generation is currently 12, but the .status.conditions[x].observedGeneration is 9, the condition is out of date
	// with respect to the current state of the instance.
	// +optional
	// +kubebuilder:validation:Minimum=0
	ObservedGeneration int64 `json:"observedGeneration,omitempty" protobuf:"varint,3,opt,name=observedGeneration"`
	// lastTransitionTime is the last time the condition transitioned from one status to another.
	// This should be when the underlying condition changed.  If that is not known, then using the time when the API field changed is acceptable.
	// +required
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Type=string
	// +kubebuilder:validation:Format=date-time
	LastTransitionTime Time `json:"lastTransitionTime" protobuf:"bytes,4,opt,name=lastTransitionTime"`
	// reason contains a programmatic identifier indicating the reason for the condition's last transition.
	// Producers of specific condition types may define expected values and meanings for this field,
	// and whether the values are considered a guaranteed API.
	// The value should be a CamelCase string.
	// This field may not be empty.
	// +required
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MaxLength=1024
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:Pattern=`^[A-Za-z]([A-Za-z0-9_,:]*[A-Za-z0-9_])?$`
	Reason string `json:"reason" protobuf:"bytes,5,opt,name=reason"`
	// message is a human readable message indicating details about the transition.
	// This may be an empty string.
	// +required
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MaxLength=32768
	Message string `json:"message" protobuf:"bytes,6,opt,name=message"`
}

package internalversion

import (
	"hit.edu/framework/pkg/apimachinery/fields"
	"hit.edu/framework/pkg/apimachinery/labels"
	"hit.edu/framework/pkg/apimachinery/runtime"
	"hit.edu/framework/pkg/apis/meta"
)

// ListOptions 处理rest请求时用到的列表查询选项
type ListOptions struct {
	meta.TypeMeta
	
	LabelSelector labels.Selector
	
	FieldSelector fields.Selector
	
	Watch bool
	
	AllowWatchBookmarks bool
	
	ResourceVersion string
	
	ResourceVersionMatch meta.ResourceVersionMatch
	
	TimeoutSeconds *int64
	
	Limit int64
	
	Continue string
	
	SendInitialEvents *bool
	
	ProgressNotify bool
}

type List struct {
	meta.TypeMeta
	// +optional
	meta.ListMeta
	
	Items []runtime.Object
}

package event

import (
	"hit.edu/framework/pkg/apimachinery/fields"
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/apiserver/registry/generic"
)

func ToSelectableFields(event *apis.Event) fields.Set {
	objectMetaFieldsSet := generic.ObjectMetaFieldsSet(&event.ObjectMeta, true)
	specificFieldsSet := fields.Set{
		//在这里定义需要设置的请求的字符串
		"reason":              event.Reason,
		"involvedObject.name": event.InvolvedObject.Name,
	}
	return generic.MergeFieldsSets(objectMetaFieldsSet, specificFieldsSet)
}

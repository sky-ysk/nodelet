package rest

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"hit.edu/framework/pkg/apis/meta"
	genericapirequest "k8s.io/apiserver/pkg/endpoints/request"

	"hit.edu/framework/pkg/apimachinery/runtime"
	"hit.edu/framework/pkg/apimachinery/runtime/schema"
)

type defaultTableConvertor struct {
	defaultQualifiedResource schema.GroupResource
}

// NewDefaultTableConvertor 创建默认的转换器
func NewDefaultTableConvertor(defaultQualifiedResource schema.GroupResource) TableConvertor {
	return defaultTableConvertor{defaultQualifiedResource: defaultQualifiedResource}
}

//var swaggerMetadataDescriptions = meta.ObjectMeta{}.SwaggerDoc()

// ConvertToTable将资源对象转换成表格形式，并生成metav1.Table对象
func (c defaultTableConvertor) ConvertToTable(ctx context.Context, object runtime.Object, tableOptions runtime.Object) (*meta.Table, error) {
	var table meta.Table
	fn := func(obj runtime.Object) error {
		m, err := meta.Accessor(obj)
		if err != nil {
			resource := c.defaultQualifiedResource
			if info, ok := genericapirequest.RequestInfoFrom(ctx); ok {
				resource = schema.GroupResource{Group: info.APIGroup, Resource: info.Resource}
			}
			return errNotAcceptable{resource: resource}
		}
		table.Rows = append(table.Rows, meta.TableRow{
			Cells:  []interface{}{m.GetName(), m.GetCreationTimestamp().Time.UTC().Format(time.RFC3339)},
			Object: runtime.RawExtension{Object: obj},
		})
		return nil
	}
	switch {
	case meta.IsListType(object):
		if err := meta.EachListItem(object, fn); err != nil {
			return nil, err
		}
	default:
		if err := fn(object); err != nil {
			return nil, err
		}
	}
	if m, err := meta.ListAccessor(object); err == nil {
		table.ResourceVersion = m.GetResourceVersion()
		table.Continue = m.GetContinue()
		table.RemainingItemCount = m.GetRemainingItemCount()
	}
	return &table, nil
}

// errNotAcceptable 表示资源不支持转换成表格
type errNotAcceptable struct {
	resource schema.GroupResource
}

func (e errNotAcceptable) Error() string {
	return fmt.Sprintf("the resource %s does not support being converted to a Table", e.resource)
}

func (e errNotAcceptable) Status() meta.Status {
	return meta.Status{
		Status:  meta.StatusFailure,
		Code:    http.StatusNotAcceptable,
		Reason:  meta.StatusReason("NotAcceptable"),
		Message: e.Error(),
	}
}

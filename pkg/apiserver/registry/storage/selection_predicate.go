// selection_predicate.go文件定义了与对象选择、筛选的操作，用于支持复杂的查询
// 目前基本上是从apiserver中照搬过来的，可能需要进行部分修改
package storage

import (
	"context"
	"hit.edu/framework/pkg/apimachinery/fields"
	"hit.edu/framework/pkg/apimachinery/labels"
	"hit.edu/framework/pkg/apiserver/endpoints/request"
	
	"hit.edu/framework/pkg/apimachinery/runtime"
	"hit.edu/framework/pkg/apis/meta"
)

// AttrFunc 提取对象的labels和fields
type AttrFunc func(obj runtime.Object) (labels.Set, fields.Set, error)

// FieldMutationFunc 对输入的fieldSet进行操作
type FieldMutationFunc func(obj runtime.Object, fieldSet fields.Set) error

// 默认的集群对象属性选择函数，提取对象的name放入fieldsSet中
func DefaultClusterScopedAttr(obj runtime.Object) (labels.Set, fields.Set, error) {
	metadata, err := meta.Accessor(obj)
	if err != nil {
		return nil, nil, err
	}
	fieldSet := fields.Set{
		"metadata.name": metadata.GetName(),
	}
	
	return labels.Set(metadata.GetLabels()), fieldSet, nil
}

// TODO:根据实际情况修改默认的属性选择函数

// 默认的命名空间级别对象属性选择函数，提取name和namespace
func DefaultNamespaceScopedAttr(obj runtime.Object) (labels.Set, fields.Set, error) {
	metadata, err := meta.Accessor(obj)
	if err != nil {
		return nil, nil, err
	}
	fieldSet := fields.Set{
		"metadata.name":      metadata.GetName(),
		"metadata.namespace": metadata.GetNamespace(),
	}
	
	return labels.Set(metadata.GetLabels()), fieldSet, nil
}

// 扩展AtrrFunc的功能，使用fieldMutator变更fieldSet
func (f AttrFunc) WithFieldMutation(fieldMutator FieldMutationFunc) AttrFunc {
	return func(obj runtime.Object) (labels.Set, fields.Set, error) {
		labelSet, fieldSet, err := f(obj)
		if err != nil {
			return nil, nil, err
		}
		if err := fieldMutator(obj, fieldSet); err != nil {
			return nil, nil, err
		}
		return labelSet, fieldSet, nil
	}
}

// SelectionPredicate 是从存储中选择对象时的筛选条件
type SelectionPredicate struct {
	Label               labels.Selector
	Field               fields.Selector
	GetAttrs            AttrFunc
	IndexLabels         []string
	IndexFields         []string
	Limit               int64
	Continue            string
	AllowWatchBookmarks bool
}

// Matches 检查对象是否符合SelectionPredicate中定义的选择条件
func (s *SelectionPredicate) Matches(obj runtime.Object) (bool, error) {
	if s.Empty() {
		return true, nil
	}
	labels, fields, err := s.GetAttrs(obj)
	if err != nil {
		return false, err
	}
	matched := s.Label.Matches(labels)
	if matched && s.Field != nil {
		matched = matched && s.Field.Matches(fields)
	}
	return matched, nil
}

// Matches 检查标签是否符合SelectionPredicate中定义的选择条件
// match s.Label and s.Field.
func (s *SelectionPredicate) MatchesObjectAttributes(l labels.Set, f fields.Set) bool {
	if s.Label.Empty() && s.Field.Empty() {
		return true
	}
	matched := s.Label.Matches(l)
	if matched && s.Field != nil {
		matched = (matched && s.Field.Matches(f))
	}
	return matched
}

// MatchesSingleNamespace 匹配namespace属性
func (s *SelectionPredicate) MatchesSingleNamespace() (string, bool) {
	if len(s.Continue) > 0 || s.Field == nil {
		return "", false
	}
	if namespace, ok := s.Field.RequiresExactMatch("metadata.namespace"); ok {
		return namespace, true
	}
	return "", false
}

// MatchesSingle 匹配name属性
func (s *SelectionPredicate) MatchesSingle() (string, bool) {
	if len(s.Continue) > 0 || s.Field == nil {
		return "", false
	}
	// TODO: should be namespace.name
	if name, ok := s.Field.RequiresExactMatch("metadata.name"); ok {
		return name, true
	}
	return "", false
}

// Empty 判断SelectionPredicate是否不做任何筛选
func (s *SelectionPredicate) Empty() bool {
	return s.Label.Empty() && s.Field.Empty()
}

// 将选择器下所有的字段和标签生成索引值
func (s *SelectionPredicate) MatcherIndex(ctx context.Context) []MatchValue {
	var result []MatchValue
	for _, field := range s.IndexFields {
		if value, ok := s.Field.RequiresExactMatch(field); ok {
			result = append(result, MatchValue{IndexName: FieldIndex(field), Value: value})
		} else if field == "metadata.namespace" {
			// list pods in the namespace. i.e. /api/v1/namespaces/default/pods
			if namespace, isNamespaceScope := isNamespaceScopedRequest(ctx); isNamespaceScope {
				result = append(result, MatchValue{IndexName: FieldIndex(field), Value: namespace})
			}
		}
	}
	for _, label := range s.IndexLabels {
		if value, ok := s.Label.RequiresExactMatch(label); ok {
			result = append(result, MatchValue{IndexName: LabelIndex(label), Value: value})
		}
	}
	return result
}

func isNamespaceScopedRequest(ctx context.Context) (string, bool) {
	re, _ := request.RequestInfoFrom(ctx)
	if re == nil || len(re.Namespace) == 0 {
		return "", false
	}
	return re.Namespace, true
}

// LabelIndex add prefix for label index.
func LabelIndex(label string) string {
	return "l:" + label
}

// FiledIndex add prefix for field index.
func FieldIndex(field string) string {
	return "f:" + field
}

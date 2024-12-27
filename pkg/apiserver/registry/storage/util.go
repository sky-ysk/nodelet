// util.go定义了与对象的生命周期管理相关的工具函数
// 目前基本上是从apiserver中照搬过来的，可能需要进行部分修改
package storage

import (
	"context"
	"fmt"
	"hit.edu/framework/pkg/apimachinery/fields"
	"hit.edu/framework/pkg/apimachinery/labels"
	"strconv"
	"sync/atomic"
	
	"hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/apiserver/registry/storage/helper"
	
	"hit.edu/framework/pkg/apimachinery/runtime"
)

type SimpleUpdateFunc func(runtime.Object) (runtime.Object, error)

// SimpleUpdateFunc converts SimpleUpdateFunc into UpdateFunc
func SimpleUpdate(fn SimpleUpdateFunc) UpdateFunc {
	return func(input runtime.Object, _ ResponseMeta) (runtime.Object, *uint64, error) {
		out, err := fn(input)
		return out, nil, err
	}
}

// NamespaceKeyFunc基于namespace和name生成键
func NamespaceKeyFunc(prefix string, obj runtime.Object) (string, error) {
	meta, err := meta.Accessor(obj)
	if err != nil {
		return "", err
	}
	name := meta.GetName()
	if msgs := helper.IsValidPathSegmentName(name); len(msgs) != 0 {
		return "", fmt.Errorf("invalid name: %v", msgs)
	}
	return prefix + "/" + meta.GetNamespace() + "/" + name, nil
}

// NoNamespaceKeyFunc基于name生成键
func NoNamespaceKeyFunc(prefix string, obj runtime.Object) (string, error) {
	meta, err := meta.Accessor(obj)
	if err != nil {
		return "", err
	}
	name := meta.GetName()
	if msgs := helper.IsValidPathSegmentName(name); len(msgs) != 0 {
		return "", fmt.Errorf("invalid name: %v", msgs)
	}
	return prefix + "/" + name, nil
}

// HighWaterMark 用于追踪版本的最新值
type HighWaterMark int64

// Update 确保HighWaterMark更新为最新值
func (hwm *HighWaterMark) Update(current int64) bool {
	for {
		old := atomic.LoadInt64((*int64)(hwm))
		if current <= old {
			return false
		}
		if atomic.CompareAndSwapInt64((*int64)(hwm), old, current) {
			return true
		}
	}
}

// GetCurrentResourceVersionFromStorage 从存储中获取当前资源的ResourceVersion
func GetCurrentResourceVersionFromStorage(ctx context.Context, storage Interface, newListFunc func() runtime.Object, resourcePrefix, objectType string) (uint64, error) {
	if storage == nil {
		return 0, fmt.Errorf("storage wasn't provided for %s", objectType)
	}
	if newListFunc == nil {
		return 0, fmt.Errorf("newListFunction wasn't provided for %s", objectType)
	}
	emptyList := newListFunc()
	pred := SelectionPredicate{
		Label: labels.Everything(),
		Field: fields.Everything(),
		Limit: 1, // just in case we actually hit something
	}
	
	err := storage.GetList(ctx, resourcePrefix, ListOptions{Predicate: pred}, emptyList)
	if err != nil {
		return 0, err
	}
	emptyListAccessor, err := meta.ListAccessor(emptyList)
	if err != nil {
		return 0, err
	}
	if emptyListAccessor == nil {
		return 0, fmt.Errorf("unable to extract a list accessor from %T", emptyList)
	}
	
	currentResourceVersion, err := strconv.Atoi(emptyListAccessor.GetResourceVersion())
	if err != nil {
		return 0, err
	}
	
	if currentResourceVersion == 0 {
		return 0, fmt.Errorf("the current resource version must be greater than 0")
	}
	return uint64(currentResourceVersion), nil
}

// AnnotateInitialEventsEndBookmark 向对象添加注解，表示初始事件已发生
// func AnnotateInitialEventsEndBookmark(obj runtime.Object) error {
// 	objMeta, err := meta.Accessor(obj)
// 	if err != nil {
// 		return err
// 	}
// 	objAnnotations := objMeta.GetAnnotations()
// 	if objAnnotations == nil {
// 		objAnnotations = map[string]string{}
// 	}
// 	objAnnotations[metav1.InitialEventsAnnotationKey] = "true"
// 	objMeta.SetAnnotations(objAnnotations)
// 	return nil
// }

// // HasInitialEventsEndBookmarkAnnotation 检查对象中是否包含初始事件已发送的注解
// func HasInitialEventsEndBookmarkAnnotation(obj runtime.Object) (bool, error) {
// 	objMeta, err := meta.Accessor(obj)
// 	if err != nil {
// 		return false, err
// 	}
// 	objAnnotations := objMeta.GetAnnotations()
// 	return objAnnotations[metav1.InitialEventsAnnotationKey] == "true", nil
// }

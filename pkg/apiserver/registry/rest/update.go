package rest

import (
	"context"
	"fmt"
	"hit.edu/framework/pkg/apimachinery/errors"
	"hit.edu/framework/pkg/apiserver/registry/storage/field"
	
	"hit.edu/framework/pkg/apimachinery/runtime"
	"hit.edu/framework/pkg/apis/meta"
)

// RESTUpdateStrategy 定义了更新策略的最小验证
type RESTUpdateStrategy interface {
	runtime.ObjectTyper
	// PrepareForUpdate 在创建之前调用，以实现标准化
	PrepareForUpdate(ctx context.Context, obj, old runtime.Object)
	// ValidateUpdate 在默认字段填充好，还未持久化前调用，返回验证错误
	ValidateUpdate(ctx context.Context, obj, old runtime.Object) field.ErrorList
	// Canonicalize 将对象转化成标准化的形式，用于比较操作
	Canonicalize(obj runtime.Object)
}

// validateCommonFields 验证对象的元数据合乎预期
func validateCommonFields(obj, old runtime.Object, strategy RESTUpdateStrategy) (field.ErrorList, error) {
	allErrs := field.ErrorList{}
	_, err := meta.Accessor(obj)
	//objectMeta, err := meta.Accessor(obj)
	if err != nil {
		return nil, fmt.Errorf("failed to get new object metadata: %v", err)
	}
	_, err = meta.Accessor(old)
	//objectMeta, err := meta.Accessor(obj)
	if err != nil {
		return nil, fmt.Errorf("failed to get old object metadata: %v", err)
	}
	//allErrs = append(allErrs, genericvalidation.ValidateObjectMetaAccessor(objectMeta, false, path.ValidatePathSegmentName, field.NewPath("metadata"))...)
	//allErrs = append(allErrs, genericvalidation.ValidateObjectMetaAccessorUpdate(objectMeta, oldObjectMeta, field.NewPath("metadata"))...)
	
	return allErrs, nil
}

// BeforeUpdate ensures that common operations for all resources are performed on update. It only returns
// errors that can be converted to api.Status. It will invoke update validation with the provided existing
// and updated objects.
// It sets zero values only if the object does not have a zero value for the respective field.
func BeforeUpdate(strategy RESTUpdateStrategy, ctx context.Context, obj, old runtime.Object) error {
	objectMeta, kind, kerr := objectMetaAndKind(strategy, obj)
	if kerr != nil {
		return kerr
	}
	// 确保请求不更新generation字段
	oldMeta, err := meta.Accessor(old)
	if err != nil {
		return err
	}
	objectMeta.SetGeneration(oldMeta.GetGeneration())
	
	strategy.PrepareForUpdate(ctx, obj, old)
	
	// 未提供UID则使用旧对象的UID
	if len(objectMeta.GetUID()) == 0 {
		objectMeta.SetUID(oldMeta.GetUID())
	}
	// 忽略时间戳变化，使用旧对象的创建时间戳
	if oldCreationTime := oldMeta.GetCreationTimestamp(); !oldCreationTime.IsZero() {
		objectMeta.SetCreationTimestamp(oldMeta.GetCreationTimestamp())
	}
	// 更新操作不应该修改时间戳信息
	if !oldMeta.GetDeletionTimestamp().IsZero() {
		objectMeta.SetDeletionTimestamp(oldMeta.GetDeletionTimestamp())
	}
	// 确保公共字段对所有资源都进行了验证
	errs, err := validateCommonFields(obj, old, strategy)
	if err != nil {
		return errors.NewInternalError(err)
	}
	// 执行更新校验
	errs = append(errs, strategy.ValidateUpdate(ctx, obj, old)...)
	if len(errs) > 0 {
		return errors.NewInvalid(kind.GroupKind(), objectMeta.GetName(), errs)
	}
	strategy.Canonicalize(obj)
	
	return nil
}

// TransformFunc 用于转换并返回新对象
type TransformFunc func(ctx context.Context, newObj runtime.Object, oldObj runtime.Object) (transformedNewObj runtime.Object, err error)

// defaultUpdatedObjectInfo 实现了UpdatedObjectInfo接口
type defaultUpdatedObjectInfo struct {
	// obj 是更新后的对象
	obj runtime.Object
	
	// transformers 是一个可选的转换函数列表，用于修改更新后的对象
	transformers []TransformFunc
}

// DefaultUpdatedObjectInfo 根据对象返回UpdatedObjectInfo实现
func DefaultUpdatedObjectInfo(obj runtime.Object, transformers ...TransformFunc) UpdatedObjectInfo {
	return &defaultUpdatedObjectInfo{obj, transformers}
}

// Preconditions 实现了UpdatedObjectInfo接口
func (i *defaultUpdatedObjectInfo) Preconditions() *meta.Preconditions {
	// 从对象中获取UID
	accessor, err := meta.Accessor(i.obj)
	if err != nil {
		return nil
	}
	// UID为空，不需要预条件
	uid := accessor.GetUID()
	if len(uid) == 0 {
		return nil
	}
	return &meta.Preconditions{UID: &uid}
}

// UpdatedObject 实现了UpdatedObjectInfo 接口
func (i *defaultUpdatedObjectInfo) UpdatedObject(ctx context.Context, oldObj runtime.Object) (runtime.Object, error) {
	var err error
	newObj := i.obj
	if newObj != nil {
		newObj = newObj.DeepCopyObject()
	}
	// 允许所有的转换器来更新新对象
	for _, transformer := range i.transformers {
		newObj, err = transformer(ctx, newObj, oldObj)
		if err != nil {
			return nil, err
		}
	}
	return newObj, nil
}

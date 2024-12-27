package rest

import (
	"context"
	"fmt"
	"hit.edu/framework/pkg/apimachinery/errors"
	"hit.edu/framework/pkg/apiserver/registry/storage/field"
	"time"
	
	"hit.edu/framework/pkg/apimachinery/runtime"
	"hit.edu/framework/pkg/apimachinery/runtime/schema"
	"hit.edu/framework/pkg/apis/meta"
	utilpointer "k8s.io/utils/pointer"
)

// RESTDeleteStrategy 定义了对象删除时的行为
type RESTDeleteStrategy interface {
	runtime.ObjectTyper
}

type GarbageCollectionPolicy string

const (
	// DeleteDependents指示删除时同时删除所有依赖对象
	DeleteDependents GarbageCollectionPolicy = "DeleteDependents"
	// OrphanDependents指示删除时保留所有依赖对象
	OrphanDependents GarbageCollectionPolicy = "OrphanDependents"
	// Unsupported 表示资源不支持回收，也不应设置finalizer
	Unsupported GarbageCollectionPolicy = "Unsupported"
)

// GarbageCollectionDeleteStrategy 在注册时必须实现，默认使用OrphanDependents
type GarbageCollectionDeleteStrategy interface {
	// DefaultGarbageCollectionPolicy 返回默认的回收行为
	DefaultGarbageCollectionPolicy(ctx context.Context) GarbageCollectionPolicy
}

// RESTGracefulDeleteStrategy 在支持graceful delete的注册时必须实现
type RESTGracefulDeleteStrategy interface {
	// CheckGracefulDelete 在对象可以优雅删除的情况下返回true
	CheckGracefulDelete(ctx context.Context, obj runtime.Object, options *meta.DeleteOptions) bool
}

func ValidateDeleteOptions(options *meta.DeleteOptions) field.ErrorList {
	allErrs := field.ErrorList{}
	//lint:file-ignore SA1019 Keep validation for deprecated OrphanDependents option until it's being removed
	if options.OrphanDependents != nil && options.PropagationPolicy != nil {
		allErrs = append(allErrs, field.Invalid(field.NewPath("propagationPolicy"), options.PropagationPolicy, "orphanDependents and deletionPropagation cannot be both set"))
	}
	if options.PropagationPolicy != nil &&
		*options.PropagationPolicy != meta.DeletePropagationForeground &&
		*options.PropagationPolicy != meta.DeletePropagationBackground &&
		*options.PropagationPolicy != meta.DeletePropagationOrphan {
		allErrs = append(allErrs, field.NotSupported(field.NewPath("propagationPolicy"), options.PropagationPolicy, []string{string(meta.DeletePropagationForeground), string(meta.DeletePropagationBackground), string(meta.DeletePropagationOrphan), "nil"}))
	}
	return allErrs
}

// BeforeDelete 检查对象是否支持优雅删除，同时完善相关设置
func BeforeDelete(strategy RESTDeleteStrategy, ctx context.Context, obj runtime.Object, options *meta.DeleteOptions) (graceful, gracefulPending bool, err error) {
	objectMeta, gvk, kerr := objectMetaAndKind(strategy, obj)
	if kerr != nil {
		return false, false, kerr
	}
	// 校验DeleteOptions是否有效
	if errs := ValidateDeleteOptions(options); len(errs) > 0 {
		return false, false, errors.NewInvalid(schema.GroupKind{Group: meta.GroupName, Kind: "DeleteOptions"}, "", errs)
	}
	// 检查前置条件
	if options.Preconditions != nil {
		if options.Preconditions.UID != nil && string(*options.Preconditions.UID) != string(objectMeta.GetUID()) {
			return false, false, errors.NewConflict(schema.GroupResource{Group: gvk.Group, Resource: gvk.Kind}, objectMeta.GetName(), fmt.Errorf("the UID in the precondition (%s) does not match the UID in record (%s). The object might have been deleted and then recreated", *options.Preconditions.UID, objectMeta.GetUID()))
		}
		if options.Preconditions.ResourceVersion != nil && *options.Preconditions.ResourceVersion != objectMeta.GetResourceVersion() {
			return false, false, errors.NewConflict(schema.GroupResource{Group: gvk.Group, Resource: gvk.Kind}, objectMeta.GetName(), fmt.Errorf("the ResourceVersion in the precondition (%s) does not match the ResourceVersion in record (%s). The object might have been modified", *options.Preconditions.ResourceVersion, objectMeta.GetResourceVersion()))
		}
	}
	
	// 将负值设置为1秒
	if gracePeriodSeconds := options.GracePeriodSeconds; gracePeriodSeconds != nil && *gracePeriodSeconds < 0 {
		options.GracePeriodSeconds = utilpointer.Int64(1)
	}
	if deletionGracePeriodSeconds := objectMeta.GetDeletionGracePeriodSeconds(); deletionGracePeriodSeconds != nil && *deletionGracePeriodSeconds < 0 {
		objectMeta.SetDeletionGracePeriodSeconds(utilpointer.Int64(1))
	}
	
	// 检查对象是否实现了RESTGracefulDeleteStrategy
	gracefulStrategy, ok := strategy.(RESTGracefulDeleteStrategy)
	if !ok {
		// 不支持优雅删除，也不需要更新generation
		return false, false, nil
	}
	// 已有删除时间戳，正在进行删除
	if objectMeta.GetDeletionTimestamp() != nil {
		if objectMeta.GetDeletionGracePeriodSeconds() == nil || *objectMeta.GetDeletionGracePeriodSeconds() == 0 {
			return false, false, nil
		}
		// 宽限期只会被缩短
		if options.GracePeriodSeconds != nil {
			period := int64(*options.GracePeriodSeconds)
			if period >= *objectMeta.GetDeletionGracePeriodSeconds() {
				return false, true, nil
			}
			newDeletionTimestamp := meta.NewTime(
				objectMeta.GetDeletionTimestamp().Add(-time.Second * time.Duration(*objectMeta.GetDeletionGracePeriodSeconds())).
					Add(time.Second * time.Duration(*options.GracePeriodSeconds)))
			objectMeta.SetDeletionTimestamp(&newDeletionTimestamp)
			objectMeta.SetDeletionGracePeriodSeconds(&period)
			return true, false, nil
		}
		options.GracePeriodSeconds = objectMeta.GetDeletionGracePeriodSeconds()
		return false, true, nil
	}
	
	// 检查是否可以优雅删除
	if !gracefulStrategy.CheckGracefulDelete(ctx, obj, options) {
		return false, false, nil
	}
	
	if options.GracePeriodSeconds == nil {
		return false, false, errors.NewInternalError(fmt.Errorf("options.GracePeriodSeconds should not be nil"))
	}
	
	now := meta.NewTime(meta.Now().Add(time.Second * time.Duration(*options.GracePeriodSeconds)))
	objectMeta.SetDeletionTimestamp(&now)
	objectMeta.SetDeletionGracePeriodSeconds(options.GracePeriodSeconds)
	// 对于第一次优雅删除，需要设置对象的Generation
	if objectMeta.GetGeneration() > 0 {
		objectMeta.SetGeneration(objectMeta.GetGeneration() + 1)
	}
	
	return true, false, nil
}

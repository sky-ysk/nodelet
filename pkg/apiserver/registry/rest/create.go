package rest

import (
	"context"
	"fmt"

	"hit.edu/framework/pkg/apimachinery/errors"
	"hit.edu/framework/pkg/apiserver/registry/storage/field"

	"hit.edu/framework/pkg/apimachinery/runtime"
	"hit.edu/framework/pkg/apimachinery/runtime/schema"
	"hit.edu/framework/pkg/apis/meta"
	genericapirequest "hit.edu/framework/pkg/apiserver/endpoints/request"
)

// RESTCreateStrategy 定义了创建策略的最小验证
type RESTCreateStrategy interface {
	runtime.ObjectTyper
	// NamespaceScoped判断资源是否支持NameSpace
	NamespaceScoped() bool
	// PrepareForCreate 在创建之前调用，以实现标准化
	PrepareForCreate(ctx context.Context, obj runtime.Object)
	// Validate 在默认字段填充好，还未持久化前调用，返回验证错误
	Validate(ctx context.Context, obj runtime.Object) field.ErrorList
	// Canonicalize 将对象转化成标准化的形式，用于比较操作
	Canonicalize(obj runtime.Object)
}

// BeforeCreate 在资源对象创建之前执行通用的操作，以确保资源对象符合创建要求
func BeforeCreate(strategy RESTCreateStrategy, ctx context.Context, obj runtime.Object) error {
	objectMeta, kind, kerr := objectMetaAndKind(strategy, obj)
	if kerr != nil {
		return kerr
	}

	// 确保关键元数据已填充
	// if !metav1.HasObjectMetaSystemFieldValues(objectMeta) {
	// 	return errors.NewInternalError(fmt.Errorf("system metadata was not initialized"))
	// }
	if len(objectMeta.GetName()) == 0 {
		return errors.NewInternalError(fmt.Errorf("metadata.name was not generated"))
	}

	requestNamespace, _ := genericapirequest.NamespaceFrom(ctx)
	//if !ok {
	//	return errors.NewInternalError(fmt.Errorf("no namespace information found in request context"))
	//}
	if err := EnsureObjectNamespaceMatchesRequestNamespace(ExpectedNamespaceForScope(requestNamespace, strategy.NamespaceScoped()), objectMeta); err != nil {
		return err
	}
	strategy.PrepareForCreate(ctx, obj)

	if errs := strategy.Validate(ctx, obj); len(errs) > 0 {
		return errors.NewInvalid(kind.GroupKind(), objectMeta.GetName(), errs)
	}
	// 执行通用的对象元数据验证
	// if errs := genericvalidation.ValidateObjectMetaAccessor(objectMeta, false, path.ValidatePathSegmentName, field.NewPath("metadata")); len(errs) > 0 {
	// 	return errors.NewInvalid(kind.GroupKind(), objectMeta.GetName(), errs)
	// }
	//执行标准化
	strategy.Canonicalize(obj)

	return nil
}

// CheckGeneratedNameError 检查资源创建时是否因名称冲突导致错误
func CheckGeneratedNameError(ctx context.Context, strategy RESTCreateStrategy, err error, obj runtime.Object) error {
	if !errors.IsAlreadyExists(err) {
		return err
	}

	objectMeta, gvk, kerr := objectMetaAndKind(strategy, obj)
	if kerr != nil {
		return kerr
	}
	// 从上下文获取资源信息
	gr := schema.GroupResource{}
	if requestInfo, found := genericapirequest.RequestInfoFrom(ctx); found {
		gr = schema.GroupResource{Group: gvk.Group, Resource: requestInfo.Resource}
	}
	return errors.NewGenerateNameConflict(gr, objectMeta.GetName(), 1)
}

// objectMetaAndKind 从资源对象中提取元数据ObjectMeta和类型信息
func objectMetaAndKind(typer runtime.ObjectTyper, obj runtime.Object) (meta.Object, schema.GroupVersionKind, error) {
	objectMeta, err := meta.Accessor(obj)
	if err != nil {
		return nil, schema.GroupVersionKind{}, errors.NewInternalError(err)
	}
	kinds, _, err := typer.ObjectKinds(obj)
	if err != nil {
		return nil, schema.GroupVersionKind{}, errors.NewInternalError(err)
	}
	return objectMeta, kinds[0], nil
}

// NamespaceScopedStrategy 指示对象是否必须在某个namespace下
type NamespaceScopedStrategy interface {
	NamespaceScoped() bool
}

package rest

import (
	"context"
	"hit.edu/framework/pkg/apimachinery/watch"
	
	//metainternalversion "k8s.io/apimachinery/pkg/apis/meta/internalversion"
	"hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/apis/meta/internalversion"
	
	"hit.edu/framework/pkg/apimachinery/runtime"
)

// Storage 是用于RESTful存储服务的通用接口，apiserver中的资源必须实现这个接口
type Storage interface {
	// New 返回一个空的对象，被用于处理Create和Update请求
	New() runtime.Object
	// Destroy 在系统关闭时清理相关的资源，必须是线程安全的，并且能够处理被多次调用的情况
	Destroy()
}

// StorageWithReadiness 在Storage的基础上扩展了ReadinessCheck
type StorageWithReadiness interface {
	Storage
	// ReadinessCheck 用于检查Storage的就绪状态
	ReadinessCheck() error
}

// SingularNameProvider 返回资源的单数名称
type SingularNameProvider interface {
	GetSingularName() string
}

// Lister 实现List操作，用于检索符合特定标签条件的资源
type Lister interface {
	// 返回一个可以被List调用的空对象
	NewList() runtime.Object
	// List 完成List操作，其中options可以为nil
	List(ctx context.Context, options *internalversion.ListOptions) (runtime.Object, error)
	// TableConvertor 确保List同时实现了TableConvertor
	TableConvertor
}

// Getter 用于检索指定名称的对象
type Getter interface {
	// Get 根据名称查找资源并返回
	Get(ctx context.Context, name string, options *meta.GetOptions) (runtime.Object, error)
}

// GetterWithOptions 在检索指定名称的对象时扩展了附加选项
type GetterWithOptions interface {
	// Get 根据名称查找资源并返回
	Get(ctx context.Context, name string, options runtime.Object) (runtime.Object, error)
	// NewGetOptions 返回选项对象，并被传递给Get方法
	NewGetOptions() (runtime.Object, bool, string)
}

type TableConvertor interface {
	ConvertToTable(ctx context.Context, object runtime.Object, tableOptions runtime.Object) (*meta.Table, error)
}

// GracefulDeleter 允许延迟并删除对象
type GracefulDeleter interface {
	// Delete 查找存储中的资源并将其删除
	Delete(ctx context.Context, name string, deleteValidation ValidateObjectFunc, options *meta.DeleteOptions) (runtime.Object, bool, error)
}

// CollectionDeleter 用于删除一组资源对象
type CollectionDeleter interface {
	// DeleteCollection 根据listOptions筛选资源并完成删除
	DeleteCollection(ctx context.Context, deleteValidation ValidateObjectFunc, options *meta.DeleteOptions, listOptions *internalversion.ListOptions) (runtime.Object, error)
}

// Creater 用于创建资源对象
type Creater interface {
	// New 返回用于创建资源的空对象
	New() runtime.Object
	// Create 创建一个新版本的资源
	Create(ctx context.Context, obj runtime.Object, createValidation ValidateObjectFunc) (runtime.Object, error)
}

// NamedCreater 使用名称参数创建资源对象
type NamedCreater interface {
	// New 返回用于创建资源的空对象
	New() runtime.Object
	// Create 创建包含名称参数的资源
	Create(ctx context.Context, name string, obj runtime.Object, createValidation ValidateObjectFunc, options *meta.CreateOptions) (runtime.Object, error)
}

// UpdatedObjectInfo 提供了关于更新对象的信息，用于Updater
type UpdatedObjectInfo interface {
	//返回从更新后的对象构建的前提条件
	Preconditions() *meta.Preconditions
	// UpdatedObject 返回更新后的对象
	UpdatedObject(ctx context.Context, oldObj runtime.Object) (newObj runtime.Object, err error)
}

// ValidateObjectFunc 对对象进行验证，不改变提供的对象
type ValidateObjectFunc func(ctx context.Context, obj runtime.Object) error

// ValidateAllObjectFunc 不进行任何验证操作
func ValidateAllObjectFunc(ctx context.Context, obj runtime.Object) error {
	return nil
}

// ValidateObjectUpdateFunc 验证对象在更新中的合法性
type ValidateObjectUpdateFunc func(ctx context.Context, obj, old runtime.Object) error

// ValidateAllObjectUpdateFunc 不进行任何验证操作
func ValidateAllObjectUpdateFunc(ctx context.Context, obj, old runtime.Object) error {
	return nil
}

// Updater 更新对象实例
type Updater interface {
	// New 返回用于更新操作的对象
	New() runtime.Object
	// Update 查找资源并完成更新
	Update(ctx context.Context, name string, objInfo UpdatedObjectInfo, createValidation ValidateObjectFunc, updateValidation ValidateObjectUpdateFunc, forceAllowCreate bool, options *meta.UpdateOptions) (runtime.Object, bool, error)
}

// CreaterUpdater 是一个支持存储和更新的存储对象
type CreaterUpdater interface {
	Creater
	Update(ctx context.Context, name string, objInfo UpdatedObjectInfo, createValidation ValidateObjectFunc, updateValidation ValidateObjectUpdateFunc, forceAllowCreate bool, options *meta.UpdateOptions) (runtime.Object, bool, error)
}

// CreaterUpdater 必须满足Updater接口
var _ Updater = CreaterUpdater(nil)

type Watcher interface {
	Watch(ctx context.Context, options *internalversion.ListOptions) (watch.Interface, error)
}

type Patcher interface {
	Getter
	Updater
}

// StandardStorage 是一个涵盖常见存储操作的接口
type StandardStorage interface {
	Getter
	Lister
	CreaterUpdater
	GracefulDeleter
	CollectionDeleter
	Watcher
	
	// Destroy 在系统关闭时清理相关的资源，必须是线程安全的，并且能够处理被多次调用的情况
	Destroy()
}

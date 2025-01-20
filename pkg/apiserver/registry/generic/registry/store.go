package registry

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"hit.edu/framework/pkg/apimachinery/fields"
	"hit.edu/framework/pkg/apimachinery/labels"
	"hit.edu/framework/pkg/apimachinery/util/wait"
	"hit.edu/framework/pkg/apimachinery/watch"

	//"k8s.io/apimachinery/pkg/api/validation"
	"hit.edu/framework/pkg/apimachinery/util/validation"
	"hit.edu/framework/pkg/apiserver/registry/generic"
	"hit.edu/framework/pkg/apiserver/registry/rest"
	"hit.edu/framework/pkg/apiserver/registry/storage"
	"hit.edu/framework/pkg/component-base/logs"
	"k8s.io/apimachinery/pkg/api/validation/path"

	apierrors "hit.edu/framework/pkg/apimachinery/errors"
	"hit.edu/framework/pkg/apimachinery/runtime"
	"hit.edu/framework/pkg/apimachinery/runtime/schema"
	"hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/apis/meta/internalversion"
	genericapirequest "hit.edu/framework/pkg/apiserver/endpoints/request"
	storeerr "hit.edu/framework/pkg/apiserver/registry/storage/errors"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apiserver/pkg/storage/etcd3/metrics"
	flowcontrolrequest "k8s.io/apiserver/pkg/util/flowcontrol/request"
	//"k8s.io/klog/v2"
)

// FinishFunc is a function returned by Begin hooks to complete an operation.
type FinishFunc func(ctx context.Context, success bool)

// AfterDeleteFunc is the type used for the Store.AfterDelete hook.
type AfterDeleteFunc func(obj runtime.Object, options *meta.DeleteOptions)

// BeginCreateFunc is the type used for the Store.BeginCreate hook.
type BeginCreateFunc func(ctx context.Context, obj runtime.Object, options *meta.CreateOptions) (FinishFunc, error)

// AfterCreateFunc is the type used for the Store.AfterCreate hook.
type AfterCreateFunc func(obj runtime.Object, options *meta.CreateOptions)

// BeginUpdateFunc is the type used for the Store.BeginUpdate hook.
type BeginUpdateFunc func(ctx context.Context, obj, old runtime.Object, options *meta.UpdateOptions) (FinishFunc, error)

// AfterUpdateFunc is the type used for the Store.AfterUpdate hook.
type AfterUpdateFunc func(obj runtime.Object, options *meta.UpdateOptions)

// GenericStore interface can be used for type assertions when we need to access the underlying strategies.
type GenericStore interface {
	GetCreateStrategy() rest.RESTCreateStrategy
	GetUpdateStrategy() rest.RESTUpdateStrategy
	GetDeleteStrategy() rest.RESTDeleteStrategy
}

func init() {
	logs.Init("etcd")
}

// Store 实现了hit.edu/framework/pkg/apiserver/registry/rest.StandardStorage接口，为REST存储提供基础操作
// Store 用于嵌入到RESTStorage中，为RESTStorage实现与底层存储相关的细节
type Store struct {
	// NewFunc 返回一个资源对象的实例，用于单个对象的Get请求
	NewFunc func() runtime.Object
	// NewListFunc 返回资源对象列表的实例，用于List请求
	NewListFunc func() runtime.Object
	// DefaultQualifiedResource 是资源的复数名称，用于标识资源类型
	DefaultQualifiedResource schema.GroupResource
	// SingularQualifiedResource 是资源的单数名称
	SingularQualifiedResource schema.GroupResource

	// KeyRootFunc 返回资源在etcd中的根键
	KeyRootFunc func(ctx context.Context) string
	// KeyFunc 返回对象具体的键，用于更新、删除等操作
	// KeyFunc and KeyRootFunc 搭配提供精确的存储路径
	KeyFunc func(ctx context.Context, name string) (string, error)
	// ObjectNameFunc 返回对象的名称或错误
	ObjectNameFunc func(obj runtime.Object) (string, error)
	// TTLFunc 返回对象的TTL
	// existing表示当前TTL或操作的默认值，update表示是否操作现有对象
	TTLFunc func(obj runtime.Object, existing uint64, update bool) (uint64, error)

	// PredicateFunc 根据label和field返回筛选函数，用于与对象进行匹配
	PredicateFunc func(label labels.Selector, field fields.Selector) storage.SelectionPredicate

	// EnableGarbageCollection 决定是否启用垃圾回收机制，允许通过finalizers对象在删除前进行清理工作
	EnableGarbageCollection bool

	// DeleteCollectionWorkers 是集合删除调用时最大的并发工作线程数
	DeleteCollectionWorkers int

	// CreateStrategy 是资源创建时的特定行为
	CreateStrategy rest.RESTCreateStrategy
	// UpdateStrategy 实现更新时的特定行为，如数据一致性验证等
	UpdateStrategy rest.RESTUpdateStrategy
	// DeleteStrategy 实现资源删除时的特定行为
	DeleteStrategy rest.RESTDeleteStrategy
	// TableConvertor 是可选接口，用于将对象转换成表格输出
	// 若未设置，则使用默认实现
	//TableConvertor rest.TableConvertor

	// Storage 是与etcd具体交互的接口
	Storage storage.Interface
	// StorageVersioner 输出对象在存储到 etcd 前被转换为的 <group/version/kind>
	StorageVersioner runtime.GroupVersioner

	// ReadinessCheckFunc 用于检查存储是否可以接受请求
	ReadinessCheckFunc func() error
	// DestroyFunc 用于清理底层存储使用的客户端，用于资源释放
	DestroyFunc func()
}

// Note: the rest.StandardStorage interface aggregates the common REST verbs
var _ rest.StandardStorage = &Store{}
var _ rest.StorageWithReadiness = &Store{}
var _ rest.TableConvertor = &Store{}
var _ GenericStore = &Store{}

var _ rest.SingularNameProvider = &Store{}

const (
	OptimisticLockErrorMsg        = "the object has been modified; please apply your changes to the latest version and try again"
	resourceCountPollPeriodJitter = 1.2
)

// KeyFunc构建键，格式为prefix/name
func defaultKeyFunc(ctx context.Context, prefix string, name string) (string, error) {
	if len(name) == 0 {
		return "", apierrors.NewBadRequest("Name parameter required.")
	}
	if msgs := path.IsValidPathSegmentName(name); len(msgs) != 0 {
		return "", apierrors.NewBadRequest(fmt.Sprintf("Name parameter invalid: %q: %s", name, strings.Join(msgs, ";")))
	}
	key := prefix + "/" + name
	return key, nil
}

// New 实现了 RESTStorage.New.
func (e *Store) New() runtime.Object {
	return e.NewFunc()
}

// ReadinessCheck 检查就绪状态
func (e *Store) ReadinessCheck() error {
	if e.ReadinessCheckFunc != nil {
		return e.ReadinessCheckFunc()
	}
	return nil
}

// Destroy 在销毁时清理存储资源
func (e *Store) Destroy() {
	if e.DestroyFunc != nil {
		e.DestroyFunc()
	}
}

// NewList 实现了rest.Lister接口中的NewList方法，用于返回新的资源列表对象
func (e *Store) NewList() runtime.Object {
	return e.NewListFunc()
}

// NamespaceScoped 判断资源是否支持namespace
func (e *Store) NamespaceScoped() bool {
	if e.CreateStrategy != nil {
		return e.CreateStrategy.NamespaceScoped()
	}
	if e.UpdateStrategy != nil {
		return e.UpdateStrategy.NamespaceScoped()
	}
	logs.Error("no Strategy for resource")
	return false
}
func (e *Store) GetCreateStrategy() rest.RESTCreateStrategy {
	return e.CreateStrategy
}
func (e *Store) GetUpdateStrategy() rest.RESTUpdateStrategy {
	return e.UpdateStrategy
}
func (e *Store) GetDeleteStrategy() rest.RESTDeleteStrategy {
	return e.DeleteStrategy
}

// List 调用List方法，并根据labels和fields进行筛选
func (e *Store) List(ctx context.Context, options *internalversion.ListOptions) (runtime.Object, error) {
	label := labels.Everything()
	if options != nil && options.LabelSelector != nil {
		label = options.LabelSelector
	}
	field := fields.Everything()
	if options != nil && options.FieldSelector != nil {
		field = options.FieldSelector
	}
	out, err := e.ListPredicate(ctx, e.PredicateFunc(label, field), options)
	if err != nil {
		logs.Error("error occur when list", err.Error())
		return nil, err
	}
	return out, nil
}

// ListPredicate 根据给定的筛选条件返回符合条件的资源列表
func (e *Store) ListPredicate(ctx context.Context, p storage.SelectionPredicate, options *internalversion.ListOptions) (runtime.Object, error) {
	if options == nil {
		// 空options，不指定资源版本
		options = &internalversion.ListOptions{ResourceVersion: ""}
	}
	// 设置分页信息
	p.Limit = options.Limit
	p.Continue = options.Continue
	list := e.NewListFunc()
	qualifiedResource := e.DefaultQualifiedResource
	storageOpts := storage.ListOptions{
		ResourceVersion:      options.ResourceVersion,
		ResourceVersionMatch: options.ResourceVersionMatch,
		Predicate:            p,
		Recursive:            true,
	}

	if requestNamespace, _ := genericapirequest.NamespaceFrom(ctx); len(requestNamespace) == 0 {
		if selectorNamespace, ok := p.MatchesSingleNamespace(); ok {
			if len(ValidateNamespaceName(selectorNamespace, false)) == 0 {
				ctx = genericapirequest.WithNamespace(ctx, selectorNamespace)
			}
		}
	}

	// 获取单个资源
	if name, ok := p.MatchesSingle(); ok {
		if key, err := e.KeyFunc(ctx, name); err == nil {
			storageOpts.Recursive = false
			err := e.Storage.GetList(ctx, key, storageOpts, list)
			return list, storeerr.InterpretListError(err, qualifiedResource)
		}
	}
	// 默认情况，获取所有资源
	err := e.Storage.GetList(ctx, e.KeyRootFunc(ctx), storageOpts, list)
	return list, storeerr.InterpretListError(err, qualifiedResource)
}

func ValidateNamespaceName(name string, prefix bool) []string {
	if prefix {
		name = maskTrailingDash(name)
	}
	return validation.IsDNS1035Label(name)
}
func maskTrailingDash(name string) string {
	if len(name) > 1 && strings.HasSuffix(name, "-") {
		return name[:len(name)-2] + "a"
	}
	return name
}

// finishNothing 是什么也不做的 FinishFunc，用于占位
func finishNothing(context.Context, bool) {}

// Create 完成Create操作
func (e *Store) Create(ctx context.Context, obj runtime.Object, createValidation rest.ValidateObjectFunc) (runtime.Object, error) {
	return e.create(ctx, obj, createValidation)
}

// create完成创建操作
func (e *Store) create(ctx context.Context, obj runtime.Object, createValidation rest.ValidateObjectFunc) (runtime.Object, error) {
	//logs.Info("now in create")
	var finishCreate FinishFunc = finishNothing
	if objectMeta, err := meta.Accessor(obj); err != nil {
		return nil, err
	} else {
		rest.FillObjectMetaSystemFields1(objectMeta)
	}
	if err := rest.BeforeCreate(e.CreateStrategy, ctx, obj); err != nil {
		logs.Error("error occur before create", err.Error())
		return nil, err
	}
	// 根据提供的函数验证对象的合法性
	if createValidation != nil {
		if err := createValidation(ctx, obj.DeepCopyObject()); err != nil {
			logs.Error("error occur before create", err.Error())
			return nil, err
		}
	}
	name, err := e.ObjectNameFunc(obj)
	if err != nil {
		logs.Error("error occur before create", err.Error())
		return nil, err
	}
	key, err := e.KeyFunc(ctx, name)
	if err != nil {
		logs.Error("error occur before create", err.Error())
		return nil, err
	}
	// 创建对象的TTL
	qualifiedResource := e.DefaultQualifiedResource
	ttl, err := e.calculateTTL(obj, 0, false)
	if err != nil {
		logs.Error("error occur before create", err.Error())
		return nil, err
	}
	out := e.NewFunc()
	if err := e.Storage.Create(ctx, key, obj, out, ttl); err != nil {
		err = storeerr.InterpretCreateError(err, qualifiedResource, name)
		err = rest.CheckGeneratedNameError(ctx, e.CreateStrategy, err, obj)
		if !apierrors.IsAlreadyExists(err) {
			logs.Error("error occur while create", err.Error())
			return nil, err
		}
		if errGet := e.Storage.Get(ctx, key, storage.GetOptions{}, out); errGet != nil {
			logs.Error("error occur while Get", errGet.Error())
			return nil, err
		}
		accessor, errGetAcc := meta.Accessor(out)
		if errGetAcc != nil {
			return nil, err
		}
		if accessor.GetDeletionTimestamp() != nil {
			msg := &err.(*apierrors.StatusError).ErrStatus.Message
			*msg = fmt.Sprintf("object is being deleted: %s", *msg)
		}
		logs.Error("error occur while create", err.Error())
		return nil, err
	}
	fn := finishCreate
	finishCreate = finishNothing
	fn(ctx, true)
	return out, nil
}

// Update 用于执行对象的更新
func (e *Store) Update(ctx context.Context, name string, objInfo rest.UpdatedObjectInfo, createValidation rest.ValidateObjectFunc, updateValidation rest.ValidateObjectUpdateFunc, forceAllowCreate bool, options *meta.UpdateOptions) (runtime.Object, bool, error) {
	// 生成存储键
	key, err := e.KeyFunc(ctx, name)
	if err != nil {
		logs.Error("error occur before update", err.Error())
		return nil, false, err
	}
	// 预先条件检查，确保更新符合用户期望
	qualifiedResource := e.DefaultQualifiedResource
	storagePreconditions := &storage.Preconditions{}
	if preconditions := objInfo.Preconditions(); preconditions != nil {
		uid := meta.UID(string(*preconditions.UID)) // 转换为 meta.UID 类型的值
		storagePreconditions.UID = &uid
		storagePreconditions.ResourceVersion = preconditions.ResourceVersion
	}

	out := e.NewFunc()
	err = e.Storage.GuaranteedUpdate(ctx, key, out, true, storagePreconditions, func(existing runtime.Object, res storage.ResponseMeta) (runtime.Object, *uint64, error) {
		// 获取已有对象的版本信息
		existingResourceVersion, err := e.Storage.Versioner().ObjectResourceVersion(existing)
		if err != nil {
			logs.Error("error occur while update", err.Error())
			return nil, nil, err
		}
		if existingResourceVersion == 0 {
			logs.Error("error occur while update")
			return nil, nil, apierrors.NewNotFound(qualifiedResource, name)
		}
		// 获取更新后的对象
		obj, err := objInfo.UpdatedObject(ctx, existing)
		if err != nil {
			logs.Error("error occur while update", err.Error())
			return nil, nil, err
		}
		newResourceVersion, err := e.Storage.Versioner().ObjectResourceVersion(obj)
		if err != nil {
			logs.Error("error occur while update", err.Error())
			return nil, nil, err
		}
		doUnconditionalUpdate := newResourceVersion == 0
		if doUnconditionalUpdate {
			// 更新最新的资源版本
			err = e.Storage.Versioner().UpdateObject(obj, res.ResourceVersion)
			if err != nil {
				logs.Error("error occur while update", err.Error())
				return nil, nil, err
			}
		} else {
			if newResourceVersion != existingResourceVersion {
				logs.Error("error occur while update", OptimisticLockErrorMsg)
				return nil, nil, apierrors.NewConflict(qualifiedResource, name, fmt.Errorf(OptimisticLockErrorMsg))
			}
		}
		var finishUpdate FinishFunc = finishNothing

		if err := rest.BeforeUpdate(e.UpdateStrategy, ctx, obj, existing); err != nil {
			logs.Error("error occur before update", err.Error())
			return nil, nil, err
		}
		// 对更新后的对象执行验证
		if updateValidation != nil {
			if err := updateValidation(ctx, obj.DeepCopyObject(), existing.DeepCopyObject()); err != nil {
				return nil, nil, err
			}
		}
		ttl, err := e.calculateTTL(obj, res.TTL, true)
		if err != nil {
			logs.Error("error occur while update", err.Error())
			return nil, nil, err
		}
		fn := finishUpdate
		finishUpdate = finishNothing
		fn(ctx, true)

		if int64(ttl) != res.TTL {
			return obj, &ttl, nil
		}
		return obj, nil, nil
	}, nil)

	if err != nil {
		logs.Error("error occur while update", err.Error())
		err = storeerr.InterpretUpdateError(err, qualifiedResource, name)
		return nil, false, err
	}

	return out, false, nil
}

// Get 从存储中获取对象
func (e *Store) Get(ctx context.Context, name string, options *meta.GetOptions) (runtime.Object, error) {
	obj := e.NewFunc()
	key, err := e.KeyFunc(ctx, name)
	if err != nil {
		logs.Error("error occur before get", err.Error())
		return nil, err
	}
	if err := e.Storage.Get(ctx, key, storage.GetOptions{ResourceVersion: options.ResourceVersion}, obj); err != nil {
		logs.Error("error occur while get", err.Error())
		return nil, storeerr.InterpretGetError(err, e.DefaultQualifiedResource, name)
	}
	return obj, nil
}

// Delete 从存储中移除对象
func (e *Store) Delete(ctx context.Context, name string, deleteValidation rest.ValidateObjectFunc, options *meta.DeleteOptions) (runtime.Object, bool, error) {
	key, err := e.KeyFunc(ctx, name)
	if err != nil {
		logs.Error("error occur before delete", err.Error())
		return nil, false, err
	}
	obj := e.NewFunc()
	qualifiedResource := e.DefaultQualifiedResource
	if err = e.Storage.Get(ctx, key, storage.GetOptions{}, obj); err != nil {
		logs.Error("error occur before delete", err.Error())
		return nil, false, storeerr.InterpretDeleteError(err, qualifiedResource, name)
	}
	if options == nil {
		sec := int64(0)
		options = &meta.DeleteOptions{GracePeriodSeconds: &sec}
		//options = meta.NewDeleteOptions(0)
	}
	var preconditions storage.Preconditions
	if options.Preconditions != nil {
		uid := meta.UID(string(*preconditions.UID)) // 转换为 meta.UID 类型的值
		preconditions.UID = &uid
		preconditions.ResourceVersion = options.Preconditions.ResourceVersion
	}
	_, pendingGraceful, err := rest.BeforeDelete(e.DeleteStrategy, ctx, obj, options)
	if err != nil {
		logs.Error("error occur before delete", err.Error())
		return nil, false, err
	}
	// 删除已在进行中
	if pendingGraceful {
		out, err := e.finalizeDelete(ctx, obj, false, options)
		return out, false, err
	}
	var ignoreNotFound bool
	var lastExisting, out runtime.Object
	out = e.NewFunc()
	if err := e.Storage.Delete(ctx, key, out, &preconditions, storage.ValidateObjectFunc(deleteValidation), nil); err != nil {
		logs.Error("error occur while delete", err.Error())
		if storage.IsNotFound(err) && ignoreNotFound && lastExisting != nil {
			out, err := e.finalizeDelete(ctx, lastExisting, true, options)
			return out, true, err
		}
		return nil, false, storeerr.InterpretDeleteError(err, qualifiedResource, name)
	}
	out, err = e.finalizeDelete(ctx, out, true, options)
	return out, true, err
}

// deleteCollectionPageSize 设置DeleteCollection时的默认分页大小
var deleteCollectionPageSize = int64(10000)

// DeleteCollection 根据List列出所有符合条件的对象，逐一调用Delete方法删除
func (e *Store) DeleteCollection(ctx context.Context, deleteValidation rest.ValidateObjectFunc, options *meta.DeleteOptions, listOptions *internalversion.ListOptions) (runtime.Object, error) {
	if listOptions == nil {
		listOptions = &internalversion.ListOptions{}
	} else {
		listOptions = listOptions.DeepCopy()
	}

	var items []runtime.Object

	// 初始化并发工作线程
	workersNumber := e.DeleteCollectionWorkers
	if workersNumber < 1 {
		workersNumber = 1
	}
	wg := sync.WaitGroup{}
	// 设置toProcess通道大小
	chanSize := 2 * workersNumber
	if chanSize < 256 {
		chanSize = 256
	}
	toProcess := make(chan runtime.Object, chanSize)
	errs := make(chan error, workersNumber+1)
	workersExited := make(chan struct{})

	wg.Add(workersNumber)
	for i := 0; i < workersNumber; i++ {
		go func() {
			// 启动删除线程
			defer utilruntime.HandleCrash(func(panicReason interface{}) {
				errs <- fmt.Errorf("DeleteCollection goroutine panicked: %v", panicReason)
			})
			defer wg.Done()

			for item := range toProcess {
				accessor, err := meta.Accessor(item)
				if err != nil {
					errs <- err
					return
				}
				if _, _, err := e.Delete(ctx, accessor.GetName(), deleteValidation, options.DeepCopy()); err != nil && !apierrors.IsNotFound(err) {
					logs.Error("error occur in deletecollection", err.Error())
					errs <- err
					return
				}
			}
		}()
	}
	// 启动线程退出监控
	go func() {
		defer utilruntime.HandleCrash(func(panicReason interface{}) {
			errs <- fmt.Errorf("DeleteCollection workers closer panicked: %v", panicReason)
		})
		wg.Wait()
		// 完成后关闭此通道，主进程继续执行
		close(workersExited)
	}()

	hasLimit := listOptions.Limit > 0
	if listOptions.Limit == 0 {
		listOptions.Limit = deleteCollectionPageSize
	}

	// 分页列出对象并完成分发
	listObj, err := func() (runtime.Object, error) {
		defer close(toProcess)

		processedItems := 0
		var originalList runtime.Object
		for {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			default:
			}

			listObj, err := e.List(ctx, listOptions)
			if err != nil {
				logs.Error("error occur while list", err.Error())
				return nil, err
			}

			newItems, err := meta.ExtractList(listObj)
			if err != nil {
				logs.Error("error occur while list", err.Error())
				return nil, err
			}
			items = append(items, newItems...)

			for i := 0; i < len(newItems); i++ {
				select {
				case toProcess <- newItems[i]:
				case <-workersExited:
					select {
					case err := <-errs:
						return nil, err
					default:
						return nil, fmt.Errorf("all DeleteCollection workers exited")
					}
				}
			}
			processedItems += len(newItems)

			if hasLimit {
				return listObj, nil
			}

			if originalList == nil {
				originalList = listObj
				meta.SetList(originalList, nil)
			}

			// If there are no more items, return the list.
			m, err := meta.ListAccessor(listObj)
			if err != nil {
				logs.Error("error occur while list", err.Error())
				return nil, err
			}
			if len(m.GetContinue()) == 0 {
				meta.SetList(originalList, items)
				return originalList, nil
			}

			// Set up the next loop.
			listOptions.Continue = m.GetContinue()
			listOptions.ResourceVersion = ""
			listOptions.ResourceVersionMatch = ""
		}
	}()
	if err != nil {
		return nil, err
	}

	// 等待所有线程完成
	<-workersExited

	select {
	case err := <-errs:
		return nil, err
	default:
		return listObj, nil
	}
}

// finalizeDelete 调用AfterDelete钩子，返回状态信息
func (e *Store) finalizeDelete(ctx context.Context, obj runtime.Object, runHooks bool, options *meta.DeleteOptions) (runtime.Object, error) {
	accessor, err := meta.Accessor(obj)
	if err != nil {
		logs.Error("error occur before delete", err.Error())
		return nil, err
	}
	qualifiedResource := e.DefaultQualifiedResource

	uid := meta.UID(string(accessor.GetUID())) // 转换为 meta.UID 类型的值
	details := &meta.StatusDetails{
		Name:  accessor.GetName(),
		Group: qualifiedResource.Group,
		Kind:  qualifiedResource.Resource,
		UID:   uid,
	}
	status := &meta.Status{Status: meta.StatusSuccess, Details: details}
	return status, nil
}

// Watch 创建筛选器，并调用WatchPredicate启动资源监听
func (e *Store) Watch(ctx context.Context, options *internalversion.ListOptions) (watch.Interface, error) {
	label := labels.Everything()
	if options != nil && options.LabelSelector != nil {
		label = options.LabelSelector
	}
	field := fields.Everything()
	if options != nil && options.FieldSelector != nil {
		field = options.FieldSelector
	}
	predicate := e.PredicateFunc(label, field)

	resourceVersion := ""
	if options != nil {
		resourceVersion = options.ResourceVersion
		predicate.AllowWatchBookmarks = options.AllowWatchBookmarks
	}
	return e.WatchPredicate(ctx, predicate, resourceVersion, options.SendInitialEvents, options.ProgressNotify)
}

// WatchPredicate 基于给定的筛选条件启动对资源的监听
func (e *Store) WatchPredicate(ctx context.Context, p storage.SelectionPredicate, resourceVersion string, sendInitialEvents *bool, progressNotify bool) (watch.Interface, error) {
	storageOpts := storage.ListOptions{ResourceVersion: resourceVersion, Predicate: p, Recursive: true, SendInitialEvents: sendInitialEvents, ProgressNotify: progressNotify}
	if requestNamespace, _ := genericapirequest.NamespaceFrom(ctx); len(requestNamespace) == 0 {
		if selectorNamespace, ok := p.MatchesSingleNamespace(); ok {
			if len(ValidateNamespaceName(selectorNamespace, false)) == 0 {
				ctx = genericapirequest.WithNamespace(ctx, selectorNamespace)
			}
		}
	}
	key := e.KeyRootFunc(ctx)
	if name, ok := p.MatchesSingle(); ok {
		if k, err := e.KeyFunc(ctx, name); err == nil {
			key = k
			storageOpts.Recursive = false
		}
	}

	w, err := e.Storage.Watch(ctx, key, storageOpts)
	if err != nil {
		logs.Error("error occur while watch", err.Error())
		return nil, err
	}
	return w, nil
}

// calculateTTL 用于计算并更新对象的TTL
func (e *Store) calculateTTL(obj runtime.Object, defaultTTL int64, update bool) (ttl uint64, err error) {
	if defaultTTL < 0 {
		defaultTTL = 1
	}
	ttl = uint64(defaultTTL)
	if e.TTLFunc != nil {
		ttl, err = e.TTLFunc(obj, ttl, update)
	}
	return ttl, err
}

// CompleteWithOptions 使用提供的选项和默认字段完善Store
func (e *Store) CompleteWithOptions(options *generic.StoreOptions) error {
	if e.DefaultQualifiedResource.Empty() {
		return fmt.Errorf("store %#v must have a non-empty qualified resource", e)
	}
	if e.SingularQualifiedResource.Empty() {
		return fmt.Errorf("store %#v must have a non-empty singular qualified resource", e)
	}
	if e.DefaultQualifiedResource.Group != e.SingularQualifiedResource.Group {
		return fmt.Errorf("store for %#v, singular and plural qualified resource's group name's must match", e)
	}
	if e.NewFunc == nil {
		return fmt.Errorf("store for %s must have NewFunc set", e.DefaultQualifiedResource.String())
	}
	if e.NewListFunc == nil {
		return fmt.Errorf("store for %s must have NewListFunc set", e.DefaultQualifiedResource.String())
	}
	if (e.KeyRootFunc == nil) != (e.KeyFunc == nil) {
		return fmt.Errorf("store for %s must set both KeyRootFunc and KeyFunc or neither", e.DefaultQualifiedResource.String())
	}

	// if e.TableConvertor == nil {
	// 	return fmt.Errorf("store for %s must set TableConvertor; rest.NewDefaultTableConvertor(e.DefaultQualifiedResource) can be used to output just name/creation time", e.DefaultQualifiedResource.String())
	// }
	var isNamespaced bool
	switch {
	case e.CreateStrategy != nil:
		isNamespaced = e.CreateStrategy.NamespaceScoped()
	case e.UpdateStrategy != nil:
		isNamespaced = e.UpdateStrategy.NamespaceScoped()
	default:
		return fmt.Errorf("store for %s must have CreateStrategy or UpdateStrategy set", e.DefaultQualifiedResource.String())
	}

	if e.DeleteStrategy == nil {
		return fmt.Errorf("store for %s must have DeleteStrategy set", e.DefaultQualifiedResource.String())
	}

	if options.RESTOptions == nil {
		return fmt.Errorf("options for %s must have RESTOptions set", e.DefaultQualifiedResource.String())
	}

	attrFunc := options.AttrFunc
	if attrFunc == nil {
		if isNamespaced {
			attrFunc = storage.DefaultNamespaceScopedAttr
		} else {
			attrFunc = storage.DefaultClusterScopedAttr
		}
	}
	if e.PredicateFunc == nil {
		e.PredicateFunc = func(label labels.Selector, field fields.Selector) storage.SelectionPredicate {
			return storage.SelectionPredicate{
				Label:    label,
				Field:    field,
				GetAttrs: attrFunc,
			}
		}
	}

	opts, err := options.RESTOptions.GetRESTOptions(e.DefaultQualifiedResource, e.NewFunc())
	if err != nil {
		return err
	}

	prefix := opts.ResourcePrefix
	if !strings.HasPrefix(prefix, "/") {
		prefix = "/" + prefix
	}
	if prefix == "/" {
		return fmt.Errorf("store for %s has an invalid prefix %q", e.DefaultQualifiedResource.String(), opts.ResourcePrefix)
	}
	if e.KeyRootFunc == nil && e.KeyFunc == nil {
		if isNamespaced {
			e.KeyRootFunc = func(ctx context.Context) string {
				return NamespaceKeyRootFunc(ctx, prefix)
			}
			e.KeyFunc = func(ctx context.Context, name string) (string, error) {
				return NamespaceKeyFunc(ctx, prefix, name)
			}
		} else {
			e.KeyRootFunc = func(ctx context.Context) string {
				return prefix
			}
			e.KeyFunc = func(ctx context.Context, name string) (string, error) {
				return NoNamespaceKeyFunc(ctx, prefix, name)
			}
		}
	}

	// 封装Store的keyFunc，使其适配Storage的keyFunc
	keyFunc := func(obj runtime.Object) (string, error) {
		accessor, err := meta.Accessor(obj)
		if err != nil {
			return "", err
		}
		if isNamespaced {
			return e.KeyFunc(genericapirequest.WithNamespace(genericapirequest.NewContext(), accessor.GetNamespace()), accessor.GetName())
		}
		return e.KeyFunc(genericapirequest.NewContext(), accessor.GetName())
	}

	if e.DeleteCollectionWorkers == 0 {
		e.DeleteCollectionWorkers = opts.DeleteCollectionWorkers
	}

	e.EnableGarbageCollection = opts.EnableGarbageCollection

	if e.ObjectNameFunc == nil {
		e.ObjectNameFunc = func(obj runtime.Object) (string, error) {
			accessor, err := meta.Accessor(obj)
			if err != nil {
				return "", err
			}
			return accessor.GetName(), nil
		}
	}

	if e.Storage == nil {
		var err error
		e.Storage, e.DestroyFunc, err = opts.Decorator(
			opts.StorageConfig,
			prefix,
			keyFunc,
			e.NewFunc,
			e.NewListFunc,
			attrFunc,
		)
		if err != nil {
			return err
		}
		e.StorageVersioner = opts.StorageConfig.EncodeVersioner

		if opts.CountMetricPollPeriod > 0 {
			stopFunc := e.startObservingCount(opts.CountMetricPollPeriod, opts.StorageObjectCountTracker)
			previousDestroy := e.DestroyFunc
			var once sync.Once
			e.DestroyFunc = func() {
				once.Do(func() {
					stopFunc()
					if previousDestroy != nil {
						previousDestroy()
					}
				})
			}
		}
	}
	if e.Storage != nil {
		e.ReadinessCheckFunc = e.Storage.ReadinessCheck
	}

	return nil
}

// 生成带命名空间的键前缀
func NamespaceKeyRootFunc(ctx context.Context, prefix string) string {
	key := prefix
	ns, ok := genericapirequest.NamespaceFrom(ctx)
	if ok && len(ns) > 0 {
		key = key + "/" + ns
	}
	return key
}

// 生成带命名空间的键
func NamespaceKeyFunc(ctx context.Context, prefix string, name string) (string, error) {
	key := NamespaceKeyRootFunc(ctx, prefix)
	ns, ok := genericapirequest.NamespaceFrom(ctx)
	if !ok || len(ns) == 0 {
		return "", apierrors.NewBadRequest("Namespace parameter required.")
	}
	if len(name) == 0 {
		return "", apierrors.NewBadRequest("Name parameter required.")
	}
	if msgs := path.IsValidPathSegmentName(name); len(msgs) != 0 {
		return "", apierrors.NewBadRequest(fmt.Sprintf("Name parameter invalid: %q: %s", name, strings.Join(msgs, ";")))
	}
	key = key + "/" + name
	return key, nil
}

// 生成不带命名空间的键
func NoNamespaceKeyFunc(ctx context.Context, prefix string, name string) (string, error) {
	if len(name) == 0 {
		return "", apierrors.NewBadRequest("Name parameter required.")
	}
	if msgs := path.IsValidPathSegmentName(name); len(msgs) != 0 {
		return "", apierrors.NewBadRequest(fmt.Sprintf("Name parameter invalid: %q: %s", name, strings.Join(msgs, ";")))
	}
	key := prefix + "/" + name
	return key, nil
}

// startObservingCount 用于定期监控某资源的对象数量，更新数量并返回停止监控的函数
func (e *Store) startObservingCount(period time.Duration, objectCountTracker flowcontrolrequest.StorageObjectCountTracker) func() {
	prefix := e.KeyRootFunc(genericapirequest.NewContext())
	resourceName := e.DefaultQualifiedResource.String()
	stopCh := make(chan struct{})
	go wait.JitterUntil(func() {
		count, err := e.Storage.Count(prefix)
		if err != nil {
			logs.Error("error occur while count", err.Error())
			count = -1
		}

		metrics.UpdateObjectCount(resourceName, count)
		if objectCountTracker != nil {
			objectCountTracker.Set(resourceName, count)
		}
	}, period, resourceCountPollPeriodJitter, true, stopCh)
	return func() { close(stopCh) }
}

func (e *Store) ConvertToTable(ctx context.Context, object runtime.Object, tableOptions runtime.Object) (*meta.Table, error) {
	// if e.TableConvertor != nil {
	// 	return e.TableConvertor.ConvertToTable(ctx, object, tableOptions)
	// }
	return rest.NewDefaultTableConvertor(e.DefaultQualifiedResource).ConvertToTable(ctx, object, tableOptions)
}

func (e *Store) StorageVersion() runtime.GroupVersioner {
	return e.StorageVersioner
}

// 返回资源的单数名称
func (e *Store) GetSingularName() string {
	return e.SingularQualifiedResource.Resource
}

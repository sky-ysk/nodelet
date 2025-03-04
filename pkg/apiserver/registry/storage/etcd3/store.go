// store.go定义了与ETCD进行实际交互的结构体和相关的方法
package etcd3

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"path"
	"reflect"
	"strings"

	"hit.edu/framework/pkg/apimachinery/watch"

	clientv3 "go.etcd.io/etcd/client/v3"
	"hit.edu/framework/pkg/apiserver/registry/storage"
	"hit.edu/framework/pkg/apiserver/registry/storage/value"

	"hit.edu/framework/pkg/apimachinery/runtime"
	"hit.edu/framework/pkg/apimachinery/runtime/schema"
	"hit.edu/framework/pkg/component-base/logs"
)

const (
	// 获取对象时允许的最大分页数
	// 若请求的分页较小，则会根据maxLimit增加页的大小，确保分页数小于maxLimit
	// 若请求的limit参数超过maxLimit则不会做修改
	maxLimit = 10000
)

// authenticatedDataString 满足了 the value.Context interface.
// 使用key来进行数据加密
// TODO:这段是直接搬上来的，加密认证还没做
type authenticatedDataString string

// AuthenticatedData 实现了 value.Context 接口
func (d authenticatedDataString) AuthenticatedData() []byte {
	return []byte(string(d))
}

var _ value.Context = authenticatedDataString("")

type store struct {
	client              *clientv3.Client
	codec               runtime.Codec
	versioner           storage.Versioner
	transformer         value.Transformer
	pathPrefix          string
	groupResource       schema.GroupResource
	groupResourceString string
	watcher             *watcher
	leaseManager        *leaseManager
}

// 获取watch进度
func (s *store) RequestWatchProgress(ctx context.Context) error {
	return s.client.RequestProgress(s.watchContext(ctx))
}

type objState struct {
	obj   runtime.Object
	meta  *storage.ResponseMeta
	rev   int64
	data  []byte
	stale bool
}

// New 返回Store实现的实际接口
func New(c *clientv3.Client, codec runtime.Codec, newFunc, newListFunc func() runtime.Object, prefix, resourcePrefix string, groupResource schema.GroupResource, transformer value.Transformer, leaseManagerConfig LeaseManagerConfig) storage.Interface {
	return newStore(c, codec, newFunc, newListFunc, prefix, resourcePrefix, groupResource, transformer, leaseManagerConfig)
}

func newStore(c *clientv3.Client, codec runtime.Codec, newFunc, newListFunc func() runtime.Object, prefix, resourcePrefix string, groupResource schema.GroupResource, transformer value.Transformer, leaseManagerConfig LeaseManagerConfig) *store {
	logs.Init("etcd")
	logs.Info("Initializing store with groupResource: " + groupResource.String())
	versioner := storage.APIObjectVersioner{}
	pathPrefix := path.Join("/", prefix)
	if !strings.HasSuffix(pathPrefix, "/") {
		pathPrefix += "/"
	}

	w := &watcher{
		client:        c,
		codec:         codec,
		newFunc:       newFunc,
		groupResource: groupResource,
		versioner:     versioner,
		transformer:   transformer,
	}
	if newFunc == nil {
		w.objectType = "<unknown>"
	} else {
		w.objectType = reflect.TypeOf(newFunc()).String()
	}
	logs.Info("Watcher initialized with objectType:  " + w.objectType)
	s := &store{
		client:              c,
		codec:               codec,
		versioner:           versioner,
		transformer:         transformer,
		pathPrefix:          pathPrefix,
		groupResource:       groupResource,
		groupResourceString: groupResource.String(),
		watcher:             w,
		leaseManager:        newDefaultLeaseManager(c, leaseManagerConfig),
	}
	w.getCurrentStorageRV = func(ctx context.Context) (uint64, error) {
		return storage.GetCurrentResourceVersionFromStorage(ctx, s, newListFunc, resourcePrefix, w.objectType)
	}
	return s
}

// Versioner 实现了 storage.Interface.Versioner接口
func (s *store) Versioner() storage.Versioner {
	return s.versioner
}

// Get 实现了 storage.Interface.Get接口
// todo: metrics实现指标监控
func (s *store) Get(ctx context.Context, key string, opts storage.GetOptions, out runtime.Object) error {
	preparedKey, err := s.prepareKey(key)
	if err != nil {
		return err
	}
	getResp, err := s.client.KV.Get(ctx, preparedKey)
	if err != nil {
		logs.Error("err when Get:", err.Error())
		return err
	}
	if err = s.validateMinimumResourceVersion(opts.ResourceVersion, uint64(getResp.Header.Revision)); err != nil {
		return err
	}

	if len(getResp.Kvs) == 0 {
		if opts.IgnoreNotFound {
			return runtime.SetZeroValue(out)
		}
		return storage.NewKeyNotFoundError(preparedKey, 0)
	}
	kv := getResp.Kvs[0]

	data, _, err := s.transformer.TransformFromStorage(ctx, kv.Value, authenticatedDataString(preparedKey))
	if err != nil {
		return err
	}

	err = decode(s.codec, s.versioner, data, out, kv.ModRevision)
	if err != nil {
		recordDecodeError(s.groupResourceString, preparedKey)
		return err
	}
	return nil
}

// Create 实现了 storage.Interface.Create接口
func (s *store) Create(ctx context.Context, key string, obj, out runtime.Object, ttl uint64) error {
	//logs.Info("this is also a test")
	preparedKey, err := s.prepareKey(key)
	if err != nil {
		return err
	}
	if version, err := s.versioner.ObjectResourceVersion(obj); err == nil && version != 0 {
		return storage.ErrResourceVersionSetOnCreate
	}
	if err := s.versioner.PrepareObjectForStorage(obj); err != nil {
		return fmt.Errorf("PrepareObjectForStorage failed: %v", err)
	}
	data, err := runtime.Encode(s.codec, obj)
	if err != nil {
		return err
	}
	opts, err := s.ttlOpts(ctx, int64(ttl))
	if err != nil {
		return err
	}

	newData, err := s.transformer.TransformToStorage(ctx, data, authenticatedDataString(preparedKey))
	if err != nil {
		return storage.NewInternalError(err.Error())
	}
	txnResp, err := s.client.KV.Txn(ctx).If(
		notFound(preparedKey),
	).Then(
		clientv3.OpPut(preparedKey, string(newData), opts...),
	).Commit()
	if err != nil {
		return err
	}

	if !txnResp.Succeeded {
		return storage.NewKeyExistsError(preparedKey, 0)
	}

	if out != nil {
		putResp := txnResp.Responses[0].GetResponsePut()
		err = decode(s.codec, s.versioner, data, out, putResp.Header.Revision)
		if err != nil {
			recordDecodeError(s.groupResourceString, preparedKey)
			return err
		}
	}
	return nil
}

// Delete 实现了 storage.Interface.Delete接口
func (s *store) Delete(
	ctx context.Context, key string, out runtime.Object, preconditions *storage.Preconditions,
	validateDeletion storage.ValidateObjectFunc, cachedExistingObject runtime.Object) error {
	preparedKey, err := s.prepareKey(key)
	if err != nil {
		return err
	}
	v, err := EnforcePtr(out)
	if err != nil {
		return fmt.Errorf("unable to convert output object to pointer: %v", err)
	}
	return s.conditionalDelete(ctx, preparedKey, out, v, preconditions, validateDeletion, cachedExistingObject)
}

func (s *store) conditionalDelete(
	ctx context.Context, key string, out runtime.Object, v reflect.Value, preconditions *storage.Preconditions,
	validateDeletion storage.ValidateObjectFunc, cachedExistingObject runtime.Object) error {
	getCurrentState := s.getCurrentState(ctx, key, v, false)

	var origState *objState
	var err error
	var origStateIsCurrent bool
	if cachedExistingObject != nil {
		origState, err = s.getStateFromObject(cachedExistingObject)
	} else {
		origState, err = getCurrentState()
		origStateIsCurrent = true
	}
	if err != nil {
		return err
	}

	for {
		if preconditions != nil {
			if err := preconditions.Check(key, origState.obj); err != nil {
				if origStateIsCurrent {
					return err
				}
				// 记录当前数据的版本
				cachedRev := origState.rev
				cachedUpdateErr := err

				// 获取当前状态
				origState, err = getCurrentState()
				if err != nil {
					return err
				}
				origStateIsCurrent = true
				// 版本没有过期
				if cachedRev == origState.rev {
					return cachedUpdateErr
				}
				continue
			}
		}
		if err := validateDeletion(ctx, origState.obj); err != nil {
			if origStateIsCurrent {
				return err
			}
			cachedRev := origState.rev
			cachedUpdateErr := err
			origState, err = getCurrentState()
			if err != nil {
				return err
			}
			origStateIsCurrent = true
			if cachedRev == origState.rev {
				return cachedUpdateErr
			}
			continue
		}

		txnResp, err := s.client.KV.Txn(ctx).If(
			clientv3.Compare(clientv3.ModRevision(key), "=", origState.rev),
		).Then(
			clientv3.OpDelete(key),
		).Else(
			clientv3.OpGet(key),
		).Commit()
		if err != nil {
			return err
		}
		if !txnResp.Succeeded {
			logs.Info("delete" + key + "failed,retry")
			getResp := (*clientv3.GetResponse)(txnResp.Responses[0].GetResponseRange())
			origState, err = s.getState(ctx, getResp, key, v, false)
			if err != nil {
				return err
			}
			origStateIsCurrent = true
			continue
		}

		if len(txnResp.Responses) == 0 || txnResp.Responses[0].GetResponseDeleteRange() == nil {
			return errors.New(fmt.Sprintf("invalid DeleteRange response: %v", txnResp.Responses))
		}
		deleteResp := txnResp.Responses[0].GetResponseDeleteRange()
		if deleteResp.Header == nil {
			return errors.New("invalid DeleteRange response - nil header")
		}
		err = decode(s.codec, s.versioner, origState.data, out, deleteResp.Header.Revision)
		if err != nil {
			recordDecodeError(s.groupResourceString, key)
			return err
		}
		return nil
	}
}

// GuaranteedUpdate 实现了 storage.Interface.GuaranteedUpdate接口
func (s *store) GuaranteedUpdate(
	ctx context.Context, key string, destination runtime.Object, ignoreNotFound bool,
	preconditions *storage.Preconditions, tryUpdate storage.UpdateFunc, cachedExistingObject runtime.Object) error {
	preparedKey, err := s.prepareKey(key)
	if err != nil {
		return err
	}

	v, err := EnforcePtr(destination)
	if err != nil {
		return fmt.Errorf("unable to convert output object to pointer: %v", err)
	}

	getCurrentState := s.getCurrentState(ctx, preparedKey, v, ignoreNotFound)

	var origState *objState
	var origStateIsCurrent bool
	if cachedExistingObject != nil {
		origState, err = s.getStateFromObject(cachedExistingObject)
	} else {
		origState, err = getCurrentState()
		origStateIsCurrent = true
	}
	if err != nil {
		return err
	}

	transformContext := authenticatedDataString(preparedKey)
	for {
		if err := preconditions.Check(preparedKey, origState.obj); err != nil {
			// 已为最新版本，返回错误信息
			if origStateIsCurrent {
				return err
			}
			origState, err = getCurrentState()
			if err != nil {
				return err
			}
			origStateIsCurrent = true
			continue
		}

		ret, ttl, err := s.updateState(origState, tryUpdate)
		if err != nil {
			if origStateIsCurrent {
				return err
			}
			cachedRev := origState.rev
			cachedUpdateErr := err
			origState, err = getCurrentState()
			if err != nil {
				return err
			}
			origStateIsCurrent = true
			if cachedRev == origState.rev {
				return cachedUpdateErr
			}
			continue
		}

		data, err := runtime.Encode(s.codec, ret)
		if err != nil {
			return err
		}
		if !origState.stale && bytes.Equal(data, origState.data) {
			// 在循环中如果跳过了Get操作，需要重新获取以确保最新
			if !origStateIsCurrent {
				origState, err = getCurrentState()
				if err != nil {
					return err
				}
				origStateIsCurrent = true
				if !bytes.Equal(data, origState.data) {
					// 原始数据发生了变化
					continue
				}
			}
			// 从etcd获取的数据没有过期
			if !origState.stale {
				err = decode(s.codec, s.versioner, origState.data, destination, origState.rev)
				if err != nil {
					recordDecodeError(s.groupResourceString, preparedKey)
					return err
				}
				return nil
			}
		}

		newData, err := s.transformer.TransformToStorage(ctx, data, transformContext)
		if err != nil {
			return storage.NewInternalError(err.Error())
		}
		opts, err := s.ttlOpts(ctx, int64(ttl))
		if err != nil {
			return err
		}

		txnResp, err := s.client.KV.Txn(ctx).If(
			clientv3.Compare(clientv3.ModRevision(preparedKey), "=", origState.rev),
		).Then(
			clientv3.OpPut(preparedKey, string(newData), opts...),
		).Else(
			clientv3.OpGet(preparedKey),
		).Commit()

		if err != nil {
			return err
		}
		if !txnResp.Succeeded {
			getResp := (*clientv3.GetResponse)(txnResp.Responses[0].GetResponseRange())
			origState, err = s.getState(ctx, getResp, preparedKey, v, ignoreNotFound)
			logs.Info("update" + key + "failed,retry")
			if err != nil {
				return err
			}
			origStateIsCurrent = true
			continue
		}
		putResp := txnResp.Responses[0].GetResponsePut()

		err = decode(s.codec, s.versioner, data, destination, putResp.Header.Revision)
		if err != nil {
			recordDecodeError(s.groupResourceString, preparedKey)
			return err
		}
		return nil
	}
}

// 返回用于生成runtime.Object类型的工厂函数
func getNewItemFunc(listObj runtime.Object, v reflect.Value) func() runtime.Object {
	// TODO:处理不定结构的列表
	// Otherwise just instantiate an empty item
	elem := v.Type().Elem()
	return func() runtime.Object {
		return reflect.New(elem).Interface().(runtime.Object)
	}
}

func (s *store) Count(key string) (int64, error) {
	preparedKey, err := s.prepareKey(key)
	if err != nil {
		return 0, err
	}
	// 以/结尾确保计数准确
	if !strings.HasSuffix(preparedKey, "/") {
		preparedKey += "/"
	}

	getResp, err := s.client.KV.Get(context.Background(), preparedKey, clientv3.WithRange(clientv3.GetPrefixRangeEnd(preparedKey)), clientv3.WithCountOnly())
	if err != nil {
		return 0, err
	}
	return getResp.Count, nil
}

// ReadinessCheck 实现了 storage.Interface接口
func (s *store) ReadinessCheck() error {
	return nil
}

// resolveGetListRev 被GetList调用，用于确定在执行 client.KV.Get 请求时应该使用的资源版本（rev）
func (s *store) resolveGetListRev(continueKey string, continueRV int64, opts storage.ListOptions) (int64, error) {
	var withRev int64
	// 对于继续请求使用continueRV
	if len(continueKey) > 0 {
		if len(opts.ResourceVersion) > 0 && opts.ResourceVersion != "0" {
			return withRev, NewBadRequest("specifying resource version is not allowed when using continue")
		}
		// continueRV > 0, LIST请求需要特定的资源版本
		// continueRV==0 无效
		// continueRV < 0, 请求最新资源版本
		if continueRV > 0 {
			withRev = continueRV
		}
		return withRev, nil
	}
	// ResourceVersion 不指定则返回0
	if len(opts.ResourceVersion) == 0 {
		return withRev, nil
	}
	parsedRV, err := s.versioner.ParseResourceVersion(opts.ResourceVersion)
	if err != nil {
		return withRev, NewBadRequest(fmt.Sprintf("invalid resource version: %v", err))
	}

	switch string(opts.ResourceVersionMatch) {
	case "NotOlderThan":
		// The not older than constraint is checked after we get a response from etcd,
		// and returnedRV is then set to the revision we get from the etcd response.
	case "Exact":
		withRev = int64(parsedRV)
	case "": // legacy case
		if opts.Recursive && opts.Predicate.Limit > 0 && parsedRV > 0 {
			withRev = int64(parsedRV)
		}
	default:
		return withRev, fmt.Errorf("unknown ResourceVersionMatch value: %v", opts.ResourceVersionMatch)
	}
	return withRev, nil
}

// GetList 实现了 storage.Interface接口
func (s *store) GetList(ctx context.Context, key string, opts storage.ListOptions, listObj runtime.Object) error {
	preparedKey, err := s.prepareKey(key)
	if err != nil {
		return err
	}
	listPtr, err := GetItemsPtr(listObj)
	if err != nil {
		return err
	}
	v, err := EnforcePtr(listPtr)
	if err != nil || v.Kind() != reflect.Slice {
		return fmt.Errorf("need ptr to slice: %v", err)
	}
	if opts.Recursive && !strings.HasSuffix(preparedKey, "/") {
		preparedKey += "/"
	}
	keyPrefix := preparedKey
	var limitOption *clientv3.OpOption
	limit := opts.Predicate.Limit
	var paging bool
	options := make([]clientv3.OpOption, 0, 4)
	if opts.Predicate.Limit > 0 {
		paging = true
		options = append(options, clientv3.WithLimit(limit))
		limitOption = &options[len(options)-1]
	}

	if opts.Recursive {
		rangeEnd := clientv3.GetPrefixRangeEnd(keyPrefix)
		options = append(options, clientv3.WithRange(rangeEnd))
	}

	newItemFunc := getNewItemFunc(listObj, v)

	var continueRV, withRev int64
	var continueKey string
	if opts.Recursive && len(opts.Predicate.Continue) > 0 {
		continueKey, continueRV, err = storage.DecodeContinue(opts.Predicate.Continue, keyPrefix)
		if err != nil {
			return NewBadRequest(fmt.Sprintf("invalid continue token: %v", err))
		}
		preparedKey = continueKey
	}
	if withRev, err = s.resolveGetListRev(continueKey, continueRV, opts); err != nil {
		return err
	}

	if withRev != 0 {
		options = append(options, clientv3.WithRev(withRev))
	}

	var lastKey []byte
	var hasMore bool
	var getResp *clientv3.GetResponse
	var numFetched int
	var numEvald int

	for {
		getResp, err = s.client.KV.Get(ctx, preparedKey, options...)
		if err != nil {
			return interpretListError(err, len(opts.Predicate.Continue) > 0, continueKey, keyPrefix)
		}
		numFetched += len(getResp.Kvs)
		if err = s.validateMinimumResourceVersion(opts.ResourceVersion, uint64(getResp.Header.Revision)); err != nil {
			return err
		}
		hasMore = getResp.More

		if len(getResp.Kvs) == 0 && getResp.More {
			return fmt.Errorf("no results were found, but etcd indicated there were more values remaining")
		}
		// indicate to the client which resource version was returned, and use the same resource version for subsequent requests.
		if withRev == 0 {
			withRev = getResp.Header.Revision
			options = append(options, clientv3.WithRev(withRev))
		}

		// avoid small allocations for the result slice, since this can be called in many
		// different contexts and we don't know how significantly the result will be filtered
		if opts.Predicate.Empty() {
			growSlice(v, len(getResp.Kvs))
		} else {
			growSlice(v, 2048, len(getResp.Kvs))
		}

		// take items from the response until the bucket is full, filtering as we go
		for i, kv := range getResp.Kvs {
			if paging && int64(v.Len()) >= opts.Predicate.Limit {
				hasMore = true
				break
			}
			lastKey = kv.Key

			data, _, err := s.transformer.TransformFromStorage(ctx, kv.Value, authenticatedDataString(kv.Key))
			if err != nil {
				return storage.NewInternalErrorf("unable to transform key %q: %v", kv.Key, err)
			}

			// Check if the request has already timed out before decode object
			select {
			case <-ctx.Done():
				// parent context is canceled or timed out, no point in continuing
				return storage.NewTimeoutError(string(kv.Key), "request did not complete within requested timeout")
			default:
			}

			obj, err := decodeListItem(ctx, data, uint64(kv.ModRevision), s.codec, s.versioner, newItemFunc)
			if err != nil {
				recordDecodeError(s.groupResourceString, string(kv.Key))
				return err
			}

			// being unable to set the version does not prevent the object from being extracted
			if matched, err := opts.Predicate.Matches(obj); err == nil && matched {
				v.Set(reflect.Append(v, reflect.ValueOf(obj).Elem()))
			}

			numEvald++

			// free kv early. Long lists can take O(seconds) to decode.
			getResp.Kvs[i] = nil
		}

		// no more results remain or we didn't request paging
		if !hasMore || !paging {
			break
		}
		// we're paging but we have filled our bucket
		if int64(v.Len()) >= opts.Predicate.Limit {
			break
		}

		if limit < maxLimit {
			// We got incomplete result due to field/label selector dropping the object.
			// Double page size to reduce total number of calls to etcd.
			limit *= 2
			if limit > maxLimit {
				limit = maxLimit
			}
			*limitOption = clientv3.WithLimit(limit)
		}
		preparedKey = string(lastKey) + "\x00"
	}

	if v.IsNil() {
		// Ensure that we never return a nil Items pointer in the result for consistency.
		v.Set(reflect.MakeSlice(v.Type(), 0, 0))
	}

	continueValue, remainingItemCount, err := storage.PrepareContinueToken(string(lastKey), keyPrefix, withRev, getResp.Count, hasMore, opts)
	if err != nil {
		return err
	}
	return s.versioner.UpdateList(listObj, uint64(withRev), continueValue, remainingItemCount)
}

// growSlice 扩展切片的容量，直到达到指定的尺寸或maxCapacity
func growSlice(v reflect.Value, maxCapacity int, sizes ...int) {
	cap := v.Cap()
	max := cap
	for _, size := range sizes {
		if size > max {
			max = size
		}
	}
	if len(sizes) == 0 || max > maxCapacity {
		max = maxCapacity
	}
	if max <= cap {
		return
	}
	if v.Len() > 0 {
		extra := reflect.MakeSlice(v.Type(), v.Len(), max)
		reflect.Copy(extra, v)
		v.Set(extra)
	} else {
		extra := reflect.MakeSlice(v.Type(), 0, max)
		v.Set(extra)
	}
}

// Watch 实现了 storage.Interface.Watch接口
func (s *store) Watch(ctx context.Context, key string, opts storage.ListOptions) (watch.Interface, error) {
	preparedKey, err := s.prepareKey(key)
	if err != nil {
		return nil, err
	}
	rev, err := s.versioner.ParseResourceVersion(opts.ResourceVersion)
	if err != nil {
		return nil, err
	}
	return s.watcher.Watch(s.watchContext(ctx), preparedKey, int64(rev), opts)
}

func (s *store) watchContext(ctx context.Context) context.Context {
	// Leader失效情况下的处理
	return clientv3.WithRequireLeader(ctx)
}

// 从etcd中获取指定键的状态
func (s *store) getCurrentState(ctx context.Context, key string, v reflect.Value, ignoreNotFound bool) func() (*objState, error) {
	return func() (*objState, error) {
		getResp, err := s.client.KV.Get(ctx, key)
		if err != nil {
			return nil, err
		}
		return s.getState(ctx, getResp, key, v, ignoreNotFound)
	}
}

func (s *store) getState(ctx context.Context, getResp *clientv3.GetResponse, key string, v reflect.Value, ignoreNotFound bool) (*objState, error) {
	state := &objState{
		meta: &storage.ResponseMeta{},
	}

	if u, ok := v.Addr().Interface().(runtime.Unstructured); ok {
		state.obj = u.NewEmptyInstance()
	} else {
		state.obj = reflect.New(v.Type()).Interface().(runtime.Object)
	}

	if len(getResp.Kvs) == 0 {
		if !ignoreNotFound {
			return nil, storage.NewKeyNotFoundError(key, 0)
		}
		if err := runtime.SetZeroValue(state.obj); err != nil {
			return nil, err
		}
	} else {
		data, stale, err := s.transformer.TransformFromStorage(ctx, getResp.Kvs[0].Value, authenticatedDataString(key))
		if err != nil {
			return nil, storage.NewInternalError(err.Error())
		}
		state.rev = getResp.Kvs[0].ModRevision
		state.meta.ResourceVersion = uint64(state.rev)
		state.data = data
		state.stale = stale
		if err := decode(s.codec, s.versioner, state.data, state.obj, state.rev); err != nil {
			recordDecodeError(s.groupResourceString, key)
			return nil, err
		}
	}
	return state, nil
}

func (s *store) getStateFromObject(obj runtime.Object) (*objState, error) {
	state := &objState{
		obj:  obj,
		meta: &storage.ResponseMeta{},
	}

	rv, err := s.versioner.ObjectResourceVersion(obj)
	if err != nil {
		return nil, fmt.Errorf("couldn't get resource version: %v", err)
	}
	state.rev = int64(rv)
	state.meta.ResourceVersion = uint64(state.rev)

	// Compute the serialized form - for that we need to temporarily clean
	// its resource version field (those are not stored in etcd).
	if err := s.versioner.PrepareObjectForStorage(obj); err != nil {
		return nil, fmt.Errorf("PrepareObjectForStorage failed: %v", err)
	}
	state.data, err = runtime.Encode(s.codec, obj)
	if err != nil {
		return nil, err
	}
	if err := s.versioner.UpdateObject(state.obj, uint64(rv)); err != nil {
		logs.Error("err when update version:", err.Error())
	}
	return state, nil
}

// updateState更新对象
func (s *store) updateState(st *objState, userUpdate storage.UpdateFunc) (runtime.Object, uint64, error) {
	ret, ttlPtr, err := userUpdate(st.obj, *st.meta)
	if err != nil {
		return nil, 0, err
	}

	if err := s.versioner.PrepareObjectForStorage(ret); err != nil {
		return nil, 0, fmt.Errorf("PrepareObjectForStorage failed: %v", err)
	}
	var ttl uint64
	if ttlPtr != nil {
		ttl = *ttlPtr
	}
	return ret, ttl, nil
}

// ttlOpts根据给定的生存时间ttl生成客户端操作选项
func (s *store) ttlOpts(ctx context.Context, ttl int64) ([]clientv3.OpOption, error) {
	if ttl == 0 {
		return nil, nil
	}
	id, err := s.leaseManager.GetLease(ctx, ttl)
	if err != nil {
		return nil, err
	}
	return []clientv3.OpOption{clientv3.WithLease(id)}, nil
}

// validateMinimumResourceVersion 验证传入的版本是否最新
func (s *store) validateMinimumResourceVersion(minimumResourceVersion string, actualRevision uint64) error {
	if minimumResourceVersion == "" {
		return nil
	}
	minimumRV, err := s.versioner.ParseResourceVersion(minimumResourceVersion)
	if err != nil {
		return NewBadRequest(fmt.Sprintf("invalid resource version: %v", err))
	}
	if minimumRV > actualRevision {
		return storage.NewTooLargeResourceVersionError(minimumRV, actualRevision, 0)
	}
	return nil
}

func (s *store) prepareKey(key string) (string, error) {
	if key == ".." ||
		strings.HasPrefix(key, "../") ||
		strings.HasSuffix(key, "/..") ||
		strings.Contains(key, "/../") {
		return "", fmt.Errorf("invalid key: %q", key)
	}
	if key == "." ||
		strings.HasPrefix(key, "./") ||
		strings.HasSuffix(key, "/.") ||
		strings.Contains(key, "/./") {
		return "", fmt.Errorf("invalid key: %q", key)
	}
	if key == "" || key == "/" {
		return "", fmt.Errorf("empty key: %q", key)
	}
	startIndex := 0
	if key[0] == '/' {
		startIndex = 1
	}
	return s.pathPrefix + key[startIndex:], nil
}

// decode 将单个对象的字节数据value转换到目标对象
func decode(codec runtime.Codec, versioner storage.Versioner, value []byte, objPtr runtime.Object, rev int64) error {
	if _, err := EnforcePtr(objPtr); err != nil {
		return fmt.Errorf("unable to convert output object to pointer: %v", err)
	}
	_, _, err := codec.Decode(value, nil, objPtr)
	if err != nil {
		return err
	}
	if err := versioner.UpdateObject(objPtr, uint64(rev)); err != nil {
		logs.Error("err when update version:", err.Error())
	}
	return nil
}

func decodeListItem(ctx context.Context, data []byte, rev uint64, codec runtime.Codec, versioner storage.Versioner, newItemFunc func() runtime.Object) (runtime.Object, error) {
	obj, _, err := codec.Decode(data, nil, newItemFunc())
	if err != nil {
		return nil, err
	}

	if err := versioner.UpdateObject(obj, rev); err != nil {
		logs.Error("err when update version:", err.Error())
	}

	return obj, nil
}

func recordDecodeError(resource string, key string) {
	logs.Error("Decoding " + resource + "," + key + "failed")
}

func notFound(key string) clientv3.Cmp {
	return clientv3.Compare(clientv3.ModRevision(key), "=", 0)
}

// getTypeName 获取obj的类型名
func getTypeName(obj interface{}) string {
	return reflect.TypeOf(obj).String()
}

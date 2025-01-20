package etcd3

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"hit.edu/framework/pkg/apimachinery/watch"
	"hit.edu/framework/pkg/component-base/logs"

	clientv3 "go.etcd.io/etcd/client/v3"
	grpccodes "google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"

	apierrors "hit.edu/framework/pkg/apimachinery/errors"
	"hit.edu/framework/pkg/apimachinery/runtime"
	"hit.edu/framework/pkg/apimachinery/runtime/schema"
	"hit.edu/framework/pkg/apiserver/registry/storage"
	"hit.edu/framework/pkg/apiserver/registry/storage/value"

	utilflowcontrol "k8s.io/apiserver/pkg/util/flowcontrol"
)

// 缓冲区大小设置
const (
	incomingBufSize         = 100
	outgoingBufSize         = 100
	processEventConcurrency = 10
)

var defaultWatcherMaxLimit int64 = maxLimit

// fatalOnDecodeError 用于在解码错误时触发panic
var fatalOnDecodeError = false

func init() {
	TestOnlySetFatalOnDecodeError(true)
	fatalOnDecodeError, _ = strconv.ParseBool(os.Getenv("KUBE_PANIC_WATCH_DECODE_ERROR"))
}

func TestOnlySetFatalOnDecodeError(b bool) {
	fatalOnDecodeError = b
}

type watcher struct {
	client              *clientv3.Client
	codec               runtime.Codec
	newFunc             func() runtime.Object
	objectType          string
	groupResource       schema.GroupResource
	versioner           storage.Versioner
	transformer         value.Transformer
	getCurrentStorageRV func(context.Context) (uint64, error)
}

// watchChan 实现了 watch.Interface.
type watchChan struct {
	watcher           *watcher
	key               string
	initialRev        int64
	recursive         bool
	progressNotify    bool
	internalPred      storage.SelectionPredicate
	ctx               context.Context
	cancel            context.CancelFunc
	incomingEventChan chan *event
	resultChan        chan watch.Event
	errChan           chan error
}

// Watch 监视指定的key，并返回相关事件的接口
// rev参数用于指定修订的版本，为0时使用最新版本
func (w *watcher) Watch(ctx context.Context, key string, rev int64, opts storage.ListOptions) (watch.Interface, error) {
	if opts.Recursive && !strings.HasSuffix(key, "/") {
		key += "/"
	}
	if opts.ProgressNotify && w.newFunc == nil {
		return nil, apierrors.NewInternalError(errors.New("progressNotify for watch is unsupported by the etcd storage because no newFunc was provided"))
	}
	startWatchRV, err := w.getStartWatchResourceVersion(ctx, rev, opts)
	if err != nil {
		return nil, err
	}
	logs.Info("start watch on : " + key)
	wc := w.createWatchChan(ctx, key, startWatchRV, opts.Recursive, opts.ProgressNotify, opts.Predicate)
	go wc.run(isInitialEventsEndBookmarkRequired(opts), areInitialEventsRequired(rev, opts))

	utilflowcontrol.WatchInitialized(ctx)

	return wc, nil
}

func (w *watcher) createWatchChan(ctx context.Context, key string, rev int64, recursive, progressNotify bool, pred storage.SelectionPredicate) *watchChan {
	wc := &watchChan{
		watcher:           w,
		key:               key,
		initialRev:        rev,
		recursive:         recursive,
		progressNotify:    progressNotify,
		internalPred:      pred,
		incomingEventChan: make(chan *event, incomingBufSize),
		resultChan:        make(chan watch.Event, outgoingBufSize),
		errChan:           make(chan error, 1),
	}
	if pred.Empty() {
		// The filter doesn't filter out any object.
		wc.internalPred = storage.Everything
	}
	wc.ctx, wc.cancel = context.WithCancel(ctx)
	return wc
}

// getStartWatchResourceVersion 返回开始watch的ResourceVersion
func (w *watcher) getStartWatchResourceVersion(ctx context.Context, resourceVersion int64, opts storage.ListOptions) (int64, error) {
	if resourceVersion > 0 {
		return resourceVersion, nil
	}
	if opts.SendInitialEvents == nil || *opts.SendInitialEvents {
		return 0, nil
	}
	currentStorageRV, err := w.getCurrentStorageRV(ctx)
	if err != nil {
		return 0, err
	}
	return int64(currentStorageRV), nil
}

func isInitialEventsEndBookmarkRequired(opts storage.ListOptions) bool {
	return opts.SendInitialEvents != nil && *opts.SendInitialEvents && opts.Predicate.AllowWatchBookmarks
}

// areInitialEventsRequired 判断是否要返回初始事件
func areInitialEventsRequired(resourceVersion int64, opts storage.ListOptions) bool {
	if opts.SendInitialEvents == nil && resourceVersion == 0 {
		return true
	}
	return opts.SendInitialEvents != nil && *opts.SendInitialEvents
}

type etcdError interface {
	Code() grpccodes.Code
	Error() string
}

type grpcError interface {
	GRPCStatus() *grpcstatus.Status
}

func isCancelError(err error) bool {
	if err == nil {
		return false
	}
	if err == context.Canceled {
		return true
	}
	if etcdErr, ok := err.(etcdError); ok && etcdErr.Code() == grpccodes.Canceled {
		return true
	}
	if grpcErr, ok := err.(grpcError); ok && grpcErr.GRPCStatus().Code() == grpccodes.Canceled {
		return true
	}
	return false
}

// run启动watch，并处理监听到的事件
func (wc *watchChan) run(initialEventsEndBookmarkRequired, forceInitialEvents bool) {
	watchClosedCh := make(chan struct{})
	go wc.startWatching(watchClosedCh, initialEventsEndBookmarkRequired, forceInitialEvents)

	var resultChanWG sync.WaitGroup
	wc.processEvents(&resultChanWG)

	select {
	case err := <-wc.errChan:
		if isCancelError(err) {
			break
		}
		errResult := transformErrorToEvent(err)
		if errResult != nil {
			select {
			case wc.resultChan <- *errResult:
			case <-wc.ctx.Done():
			}
		}
	case <-watchClosedCh:
	case <-wc.ctx.Done(): // user cancel
	}

	wc.cancel()

	resultChanWG.Wait()
	close(wc.resultChan)
}

func (wc *watchChan) Stop() {
	wc.cancel()
}

func (wc *watchChan) ResultChan() <-chan watch.Event {
	return wc.resultChan
}

func (wc *watchChan) RequestWatchProgress() error {
	return wc.watcher.client.RequestProgress(wc.ctx)
}

// sync 同步数据并发送到后续事件处理中
func (wc *watchChan) sync() error {
	opts := []clientv3.OpOption{}
	if wc.recursive {
		opts = append(opts, clientv3.WithLimit(defaultWatcherMaxLimit))
		rangeEnd := clientv3.GetPrefixRangeEnd(wc.key)
		opts = append(opts, clientv3.WithRange(rangeEnd))
	}

	var err error
	var lastKey []byte
	var withRev int64
	var getResp *clientv3.GetResponse

	preparedKey := wc.key

	for {
		getResp, err = wc.watcher.client.KV.Get(wc.ctx, preparedKey, opts...)
		if err != nil {
			return interpretListError(err, true, preparedKey, wc.key)
		}

		if len(getResp.Kvs) == 0 && getResp.More {
			return fmt.Errorf("no results were found, but etcd indicated there were more values remaining")
		}

		// 发送所有etcd返回的响应
		for i, kv := range getResp.Kvs {
			lastKey = kv.Key
			wc.sendEvent(parseKV(kv))
			getResp.Kvs[i] = nil
		}

		if withRev == 0 {
			wc.initialRev = getResp.Header.Revision
		}

		if !getResp.More {
			return nil
		}

		preparedKey = string(lastKey) + "\x00"
		if withRev == 0 {
			withRev = getResp.Header.Revision
			opts = append(opts, clientv3.WithRev(withRev))
		}
	}
}

func logWatchChannelErr(err error) {
	switch {
	case strings.Contains(err.Error(), "mvcc: required revision has been compacted"):
		logs.Error("watch chan error:", err.Error())
	case isCancelError(err):
		logs.Error("watch chan error:", err.Error())
	}
}

func Jitter(duration time.Duration, maxFactor float64) time.Duration {
	if maxFactor <= 0.0 {
		maxFactor = 1.0
	}
	wait := duration + time.Duration(rand.Float64()*maxFactor*float64(duration))
	return wait
}

// startWatching启动对指定键的监听，并处理事件和错误
func (wc *watchChan) startWatching(watchClosedCh chan struct{}, initialEventsEndBookmarkRequired, forceInitialEvents bool) {
	if wc.initialRev > 0 && forceInitialEvents {
		currentStorageRV, err := wc.watcher.getCurrentStorageRV(wc.ctx)
		if err != nil {
			wc.sendError(err)
			return
		}
		if uint64(wc.initialRev) > currentStorageRV {
			wc.sendError(storage.NewTooLargeResourceVersionError(uint64(wc.initialRev), currentStorageRV, int(Jitter(1*time.Second, 3).Seconds())))
			return
		}
	}

	//同步初始事件
	if forceInitialEvents {
		if err := wc.sync(); err != nil {
			logs.Error("fail in sync:", err.Error())
			wc.sendError(err)
			return
		}
	}
	//初始事件的bookmark
	if initialEventsEndBookmarkRequired {
		wc.sendEvent(func() *event {
			e := progressNotifyEvent(wc.initialRev)
			e.isInitialEventsEndBookmark = true
			return e
		}())
	}
	opts := []clientv3.OpOption{clientv3.WithRev(wc.initialRev + 1), clientv3.WithPrevKV()}
	if wc.recursive {
		opts = append(opts, clientv3.WithPrefix())
	}
	if wc.progressNotify {
		opts = append(opts, clientv3.WithProgressNotify())
	}
	wch := wc.watcher.client.Watch(wc.ctx, wc.key, opts...)
	for wres := range wch {
		if wres.Err() != nil {
			err := wres.Err()
			// 因为发生错误而关闭
			logWatchChannelErr(err)
			wc.sendError(err)
			return
		}
		// 回送bookmark
		if wres.IsProgressNotify() {
			wc.sendEvent(progressNotifyEvent(wres.Header.GetRevision()))
			continue
		}

		for _, e := range wres.Events {
			parsedEvent, err := parseEvent(e)
			if err != nil {
				logWatchChannelErr(err)
				wc.sendError(err)
				return
			}
			wc.sendEvent(parsedEvent)
		}
	}
	// 结束监视
	close(watchClosedCh)
}

// processEvents 从etcdwatcher处理事件并发送到resultChan
func (wc *watchChan) processEvents(wg *sync.WaitGroup) {
	wg.Add(1)
	go wc.serialProcessEvents(wg)
}
func (wc *watchChan) serialProcessEvents(wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case e := <-wc.incomingEventChan:
			res := wc.transform(e)
			if res == nil {
				continue
			}
			select {
			case wc.resultChan <- *res:
			case <-wc.ctx.Done():
				return
			}
		case <-wc.ctx.Done():
			return
		}
	}
}

// concurrentProcessEvents并发地处理从wc.incomingEventChan中接收到的事件
func (wc *watchChan) concurrentProcessEvents(wg *sync.WaitGroup) {
	p := concurrentOrderedEventProcessing{
		input:           wc.incomingEventChan,
		processFunc:     wc.transform,
		output:          wc.resultChan,
		processingQueue: make(chan chan *watch.Event, processEventConcurrency-1),

		objectType:    wc.watcher.objectType,
		groupResource: wc.watcher.groupResource,
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		p.scheduleEventProcessing(wc.ctx, wg)
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		p.collectEventProcessing(wc.ctx)
	}()
}

type concurrentOrderedEventProcessing struct {
	input       chan *event
	processFunc func(*event) *watch.Event
	output      chan watch.Event

	processingQueue chan chan *watch.Event
	// Metadata for logging
	objectType    string
	groupResource schema.GroupResource
}

func (p *concurrentOrderedEventProcessing) scheduleEventProcessing(ctx context.Context, wg *sync.WaitGroup) {
	var e *event
	for {
		select {
		case <-ctx.Done():
			return
		case e = <-p.input:
		}
		processingResponse := make(chan *watch.Event, 1)
		select {
		case <-ctx.Done():
			return
		case p.processingQueue <- processingResponse:
		}
		wg.Add(1)
		go func(e *event, response chan<- *watch.Event) {
			defer wg.Done()
			select {
			case <-ctx.Done():
			case response <- p.processFunc(e):
			}
		}(e, processingResponse)
	}
}

func (p *concurrentOrderedEventProcessing) collectEventProcessing(ctx context.Context) {
	var processingResponse chan *watch.Event
	var e *watch.Event
	for {
		select {
		case <-ctx.Done():
			return
		case processingResponse = <-p.processingQueue:
		}
		select {
		case <-ctx.Done():
			return
		case e = <-processingResponse:
		}
		if e == nil {
			continue
		}
		select {
		case <-ctx.Done():
			return
		case p.output <- *e:
		}
	}
}

func (wc *watchChan) filter(obj runtime.Object) bool {
	if wc.internalPred.Empty() {
		return true
	}
	matched, err := wc.internalPred.Matches(obj)
	return err == nil && matched
}

func (wc *watchChan) acceptAll() bool {
	return wc.internalPred.Empty()
}

// transform 将事件转换成结果
func (wc *watchChan) transform(e *event) (res *watch.Event) {
	curObj, oldObj, err := wc.prepareObjs(e)
	if err != nil {
		logs.Error("error in prepare object", err.Error())
		wc.sendError(err)
		return nil
	}

	switch {
	case e.isProgressNotify:
		object := wc.watcher.newFunc()
		if err := wc.watcher.versioner.UpdateObject(object, uint64(e.rev)); err != nil {
			logs.Error("error in update object", err.Error())
			return nil
		}
		//if e.isInitialEventsEndBookmark {
		//	if err := storage.AnnotateInitialEventsEndBookmark(object); err != nil {
		//		wc.sendError(fmt.Errorf("error while accessing object's metadata gr: %v, type: %v, obj: %#v, err: %v", wc.watcher.groupResource, wc.watcher.objectType, object, err))
		//		return nil
		//	}
		//}
		res = &watch.Event{
			Type:   watch.Bookmark,
			Object: object,
		}
	case e.isDeleted:
		if !wc.filter(oldObj) {
			return nil
		}
		res = &watch.Event{
			Type:   watch.Deleted,
			Object: oldObj,
		}
	case e.isCreated:
		if !wc.filter(curObj) {
			return nil
		}
		res = &watch.Event{
			Type:   watch.Added,
			Object: curObj,
		}
	default:
		if wc.acceptAll() {
			res = &watch.Event{
				Type:   watch.Modified,
				Object: curObj,
			}
			return res
		}
		curObjPasses := wc.filter(curObj)
		oldObjPasses := wc.filter(oldObj)
		switch {
		case curObjPasses && oldObjPasses:
			res = &watch.Event{
				Type:   watch.Modified,
				Object: curObj,
			}
		case curObjPasses && !oldObjPasses:
			res = &watch.Event{
				Type:   watch.Added,
				Object: curObj,
			}
		case !curObjPasses && oldObjPasses:
			res = &watch.Event{
				Type:   watch.Deleted,
				Object: oldObj,
			}
		}
	}
	return res
}

func transformErrorToEvent(err error) *watch.Event {
	err = interpretWatchError(err)
	if _, ok := err.(apierrors.APIStatus); !ok {
		err = apierrors.NewInternalError(err)
	}
	status := err.(apierrors.APIStatus).Status()
	return &watch.Event{
		Type:   watch.Error,
		Object: &status,
	}
}

func (wc *watchChan) sendError(err error) {
	select {
	case wc.errChan <- err:
	case <-wc.ctx.Done():
	}
}

func (wc *watchChan) sendEvent(e *event) {
	select {
	case wc.incomingEventChan <- e:
	case <-wc.ctx.Done():
	}
}

func (wc *watchChan) prepareObjs(e *event) (curObj runtime.Object, oldObj runtime.Object, err error) {
	if e.isProgressNotify {
		// progressNotify 事件不包含当前或之前的对象版本
		return nil, nil, nil
	}

	if !e.isDeleted {
		data, _, err := wc.watcher.transformer.TransformFromStorage(wc.ctx, e.value, authenticatedDataString(e.key))
		if err != nil {
			return nil, nil, err
		}
		curObj, err = decodeObj(wc.watcher.codec, wc.watcher.versioner, data, e.rev)
		if err != nil {
			return nil, nil, err
		}
	}
	if len(e.prevValue) > 0 && (e.isDeleted || !wc.acceptAll()) {
		data, _, err := wc.watcher.transformer.TransformFromStorage(wc.ctx, e.prevValue, authenticatedDataString(e.key))
		if err != nil {
			return nil, nil, err
		}
		oldObj, err = decodeObj(wc.watcher.codec, wc.watcher.versioner, data, e.rev)
		if err != nil {
			return nil, nil, err
		}
	}
	return curObj, oldObj, nil
}

func decodeObj(codec runtime.Codec, versioner storage.Versioner, data []byte, rev int64) (_ runtime.Object, err error) {
	obj, err := runtime.Decode(codec, []byte(data))
	if err != nil {
		if fatalOnDecodeError {
			logs.Error(err)
		}
		return nil, err
	}
	if err := versioner.UpdateObject(obj, uint64(rev)); err != nil {
		return nil, fmt.Errorf("failure to version api object (%d) %#v: %v", rev, obj, err)
	}
	return obj, nil
}

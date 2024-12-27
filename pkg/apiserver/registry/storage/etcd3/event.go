// event.go定义了与事件相关的内容，将对etcd的操作转变为event进行处理
package etcd3

import (
	"fmt"

	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type event struct {
	key              string
	value            []byte
	prevValue        []byte
	rev              int64
	isDeleted        bool
	isCreated        bool
	isProgressNotify bool
	// isInitialEventsEndBookmark 用于跟踪是否已发送bookmark event
	isInitialEventsEndBookmark bool
}

// parseKV 将etcd返回的键值对转换成表示创建事件的event
func parseKV(kv *mvccpb.KeyValue) *event {
	return &event{
		key:       string(kv.Key),
		value:     kv.Value,
		prevValue: nil,
		rev:       kv.ModRevision,
		isDeleted: false,
		isCreated: true,
	}
}

// parseEvent 将etcd的Event转换成event
func parseEvent(e *clientv3.Event) (*event, error) {
	if !e.IsCreate() && e.PrevKv == nil {
		// If the previous value is nil, error. One example of how this is possible is if the previous value has been compacted already.
		return nil, fmt.Errorf("etcd event received with PrevKv=nil (key=%q, modRevision=%d, type=%s)", string(e.Kv.Key), e.Kv.ModRevision, e.Type.String())

	}
	ret := &event{
		key:       string(e.Kv.Key),
		value:     e.Kv.Value,
		rev:       e.Kv.ModRevision,
		isDeleted: e.Type == clientv3.EventTypeDelete,
		isCreated: e.IsCreate(),
	}
	if e.PrevKv != nil {
		ret.prevValue = e.PrevKv.Value
	}
	return ret, nil
}

func progressNotifyEvent(rev int64) *event {
	return &event{
		rev:              rev,
		isProgressNotify: true,
	}
}

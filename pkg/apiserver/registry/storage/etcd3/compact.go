package etcd3

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"hit.edu/framework/pkg/component-base/logs"
)

const (
	compactRevKey = "compact_rev_key"
)

var (
	endpointsMapMu sync.Mutex
	// 记录客户端到endpoints建立的连接的映射
	endpointsMap map[string]struct{}
)

func init() {
	endpointsMap = make(map[string]struct{})
}

func ensureCompactRevKeyExists(ctx context.Context, client *clientv3.Client) error {
	resp, err := client.Get(ctx, compactRevKey)
	if err != nil {
		return err
	}
	if len(resp.Kvs) == 0 {
		_, err := client.Put(ctx, compactRevKey, "0")
		if err != nil {
			return fmt.Errorf("failed to initialize compactRevKey: %v", err)
		}
		logs.Info("initialized compactRevKey in etcd")
	} else {
		logs.Info("compactRevKey already exists in etcd")
	}
	return nil
}

func StartCompactor(ctx context.Context, client *clientv3.Client, compactInterval time.Duration) {
	endpointsMapMu.Lock()
	defer endpointsMapMu.Unlock()

	for _, ep := range client.Endpoints() {
		if _, ok := endpointsMap[ep]; ok {
			logs.Info("compactore already exist")
			return
		}
	}
	for _, ep := range client.Endpoints() {
		endpointsMap[ep] = struct{}{}
	}

	if compactInterval != 0 {
		go compactor(ctx, client, compactInterval)
	}
}

// compactor 定期压缩历史版本的功能，用于清理etcd中的旧版本键值对，只保留最新的键值对
func compactor(ctx context.Context, client *clientv3.Client, interval time.Duration) {
	var compactTime int64
	var rev int64
	var err error
	if err := ensureCompactRevKeyExists(ctx, client); err != nil {
		logs.Error("failed to initialize compactRevKey: ", err)
		return
	}
	for {
		select {
		case <-time.After(interval):
		case <-ctx.Done():
			return
		}

		compactTime, rev, err = compact(ctx, client, compactTime, rev)
		if err != nil {
			logs.Error("compact failed")
			continue
		}
	}
}

// compact 完成对etcd的压缩，并返回当前的版本
func compact(ctx context.Context, client *clientv3.Client, t, rev int64) (int64, int64, error) {
	logs.Info("start compact")
	resp, err := client.KV.Txn(ctx).If(
		clientv3.Compare(clientv3.Version(compactRevKey), "=", t),
	).Then(
		clientv3.OpPut(compactRevKey, strconv.FormatInt(rev, 10)),
	).Else(
		clientv3.OpGet(compactRevKey),
	).Commit()
	if err != nil {
		return t, rev, err
	}

	curRev := resp.Header.Revision

	if !resp.Succeeded {
		kvs := resp.Responses[0].GetResponseRange().Kvs
		if len(kvs) == 0 {
			logs.Info("etcd compact failed since compactRevKey not found in etcd")
			return 0, 0, fmt.Errorf("compactRevKey not found in etcd")
		}
		curTime := kvs[0].Version
		logs.Infof("compact skipped, curTime=%d, rev=%d", curTime, curRev)
		return curTime, curRev, nil
	}
	curTime := t + 1

	if rev == 0 {
		logs.Infof("compact skipped due to rev=0, curTime=%d, curRev=%d", curTime, curRev)
		return curTime, curRev, nil
	}
	if _, err = client.Compact(ctx, rev); err != nil {
		return curTime, curRev, err
	}
	logs.Info("etcd compact rev" + strconv.FormatInt(rev, 10))
	return curTime, curRev, nil
}

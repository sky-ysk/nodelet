package databus

import (
	"context"

	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// Interface定义了与etcd交互时的最小客户端接口
// 以下是标准的etcd操作，不过经过调整使其更适合需求
// 目前这一部分直接挪用了etcd的k8s部分，后续需要再做调整
type Interface interface {
	// Get 从etcd中检索单个键值对
	//
	// 若opts.Revision被设置为非0值，则会检索指定的版本
	// 若该版本已被压缩，则操作失败并返回ErrCompacted错误
	Get(ctx context.Context, key string, opts GetOptions) (GetResponse, error)

	// List 返回某指定前缀下所有键值对，并按字典序排列
	//
	// 若opts.Revision被设置为非0值，则会检索指定的版本
	// 若该版本已被压缩，则操作失败并返回ErrCompacted错误
	// opts.Limit大于零，则限制返回个数
	// opts.Continue 不为空，则从该键的下一个开始列出
	// Continue应当是上一个分页中返回的最后一个键
	List(ctx context.Context, prefix string, opts ListOptions) (ListResponse, error)

	// Count 返回某前缀下键数
	//
	Count(ctx context.Context, prefix string, opts CountOptions) (int64, error)

	// OptimisticPut 在键于 expectedRevision 指定的版本后若未被修改则进行创建或更新
	//
	// 若被修改则操作失败
	OptimisticPut(ctx context.Context, key string, value []byte, expectedRevision int64, opts PutOptions) (PutResponse, error)

	// OptimisticPut 在键于 expectedRevision 指定的版本后若未被修改则进行删除
	//
	// 若被修改则操作失败
	OptimisticDelete(ctx context.Context, key string, expectedRevision int64, opts DeleteOptions) (DeleteResponse, error)
}

type GetOptions struct {
	// Revision 是用于Get操作的版本号
	// 若 Revision 为0，则返回最新的值
	Revision int64
}

type ListOptions struct {
	// Revision 是用于Get操作的版本号
	// 若 Revision 为0，则返回最新的值
	Revision int64

	// Limit 是 List 操作的最大返回数
	// Limite 为0则不进行限制
	Limit int64

	// Continue 是用于继续 List 操作的键，从指定键的下一个键开始。
	// 当进行分页操作时，应将其设置为上一个 ListResponse 中返回的最后一个键。
	Continue string
}

// 占位符，将来可能用到
type CountOptions struct{}

type PutOptions struct {
	// GetOnFailure 指定如果由于版本不匹配导致 Put 操作失败，是否返回修改后的键值对。
	GetOnFailure bool

	// LeaseID 是与键关联的租约 ID，租约到期后会自动删除该键（基于租约的 TTL，生存时间）。
	// 已废弃：当 Interface 开始为每个对象使用一个租约时，应该使用 TTL 替代。
	LeaseID clientv3.LeaseID
}

type DeleteOptions struct {
	// GetOnFailure 指定如果由于版本不匹配导致 Delete 操作失败，是否返回修改后的键值对。
	GetOnFailure bool
}

type GetResponse struct {
	// KV 是从 etcd 检索到的键值对。
	KV *mvccpb.KeyValue

	// Revision 是用于Get操作的版本号
	Revision int64
}

type ListResponse struct {
	// KV 是从 etcd 检索到的键值对，使用字典序排列
	Kvs []*mvccpb.KeyValue

	// Count 是具有指定前缀的键的总数，即使由于限制而未返回所有键。
	Count int64

	// Revision 是执行 List 操作时键值存储的修订版本。
	Revision int64
}

type PutResponse struct {
	// KV 是已创建或更新的键值对。如果 Put 操作失败并且 GetOnFailure 为 true，
	// 这将是导致失败的已修改键值对。
	KV *mvccpb.KeyValue

	// Succeeded 表示 Put 操作是否成功。
	Succeeded bool

	// Revision 是执行 Put 操作后键值存储的修订版本。
	Revision int64
}

type DeleteResponse struct {
	// KV 是已删除的键值对。如果 Delete 操作失败并且 GetOnFailure 为 true，
	// 这将是导致失败的已修改键值对。
	KV *mvccpb.KeyValue

	// Succeeded 表示 Delete 操作是否成功。
	Succeeded bool

	// Revision 是执行 Delete 操作后键值存储的修订版本。
	Revision int64
}

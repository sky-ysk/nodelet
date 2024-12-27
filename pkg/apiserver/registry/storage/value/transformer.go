// package value 提供了辅助存储值转换的方法
package value

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"hit.edu/framework/pkg/apimachinery/runtime/schema"
	"k8s.io/apimachinery/pkg/util/errors"
	genericapirequest "k8s.io/apiserver/pkg/endpoints/request"
	"k8s.io/klog/v2"
)

func init() {
	//RegisterMetrics()
}

// Context 存储转换可能需要的附加信息，用于验证静态数据。
type Context interface {
	// AuthenticatedData 应该返回描述当前值的字节数组
	// 为了额外验证，将其设置为能强烈识别该值的数据，例如存储数据的密钥和创建版本
	AuthenticatedData() []byte
}

type Read interface {
	// TransformFromStorage 可以将提供的数据从其底层存储表示转换过来，或者返回一个错误。
	// 如果磁盘上的对象已过时，并且即使对象内容没有变化，也应向 etcd 进行写操作，则 stale 为 true
	TransformFromStorage(ctx context.Context, data []byte, dataCtx Context) (out []byte, stale bool, err error)
}

type Write interface {
	// TransformToStorage 可以将提供的数据转换为存储中的适当形式
	TransformToStorage(ctx context.Context, data []byte, dataCtx Context) (out []byte, err error)
}

// Transformer 允许在从底层存储读取或写入之前对值进行转换
type Transformer interface {
	Read
	Write
}

// ResourceTransformers 返回一个用于指定资源的转换器
type ResourceTransformers interface {
	TransformerForResource(resource schema.GroupResource) Transformer
}

// DefaultContext 是Context的简单实现
type DefaultContext []byte

func (c DefaultContext) AuthenticatedData() []byte { return c }

// PrefixTransformer 包含一个转换器接口和该转换器所在的前缀
type PrefixTransformer struct {
	Prefix      []byte
	Transformer Transformer
}

type prefixTransformers struct {
	transformers []PrefixTransformer
	err          error
}

var _ Transformer = &prefixTransformers{}

// NewPrefixTransformers 将传入的数据与提供的前缀进行匹配。
func NewPrefixTransformers(err error, transformers ...PrefixTransformer) Transformer {
	if err == nil {
		err = fmt.Errorf("the provided value does not match any of the supported transformers")
	}
	return &prefixTransformers{
		transformers: transformers,
		err:          err,
	}
}

// TransformFromStorage 用于找到第一个具有匹配前缀的转换器，并返回转换后的结果。
func (t *prefixTransformers) TransformFromStorage(ctx context.Context, data []byte, dataCtx Context) ([]byte, bool, error) {
	start := time.Now()
	var errs []error
	for i, transformer := range t.transformers {
		if bytes.HasPrefix(data, transformer.Prefix) {
			result, stale, err := transformer.Transformer.TransformFromStorage(ctx, data[len(transformer.Prefix):], dataCtx)
			if len(transformer.Prefix) == 0 && err != nil {
				continue
			}
			if len(transformer.Prefix) == 0 {
				RecordTransformation("from_storage", "identity", time.Since(start), err)
			} else {
				RecordTransformation("from_storage", string(transformer.Prefix), time.Since(start), err)
			}

			if err != nil {
				errs = append(errs, err)
				continue
			}

			return result, stale || i != 0, err
		}
	}
	if err := errors.Reduce(errors.NewAggregate(errs)); err != nil {
		logTransformErr(ctx, err, "failed to decrypt data")
		return nil, false, err
	}
	RecordTransformation("from_storage", "unknown", time.Since(start), t.err)
	return nil, false, t.err
}

// TransformToStorage 用第一个转换器并在数据前加上其前缀
func (t *prefixTransformers) TransformToStorage(ctx context.Context, data []byte, dataCtx Context) ([]byte, error) {
	start := time.Now()
	transformer := t.transformers[0]
	result, err := transformer.Transformer.TransformToStorage(ctx, data, dataCtx)
	RecordTransformation("to_storage", string(transformer.Prefix), time.Since(start), err)
	if err != nil {
		logTransformErr(ctx, err, "failed to encrypt data")
		return nil, err
	}
	prefixedData := make([]byte, len(transformer.Prefix), len(result)+len(transformer.Prefix))
	copy(prefixedData, transformer.Prefix)
	prefixedData = append(prefixedData, result...)
	return prefixedData, nil
}

func logTransformErr(ctx context.Context, err error, message string) {
	requestInfo := getRequestInfoFromContext(ctx)
	if klogLevel6 := klog.V(6); klogLevel6.Enabled() {
		klogLevel6.InfoSDepth(
			1,
			message,
			"err", err,
			"group", requestInfo.APIGroup,
			"version", requestInfo.APIVersion,
			"resource", requestInfo.Resource,
			"subresource", requestInfo.Subresource,
			"verb", requestInfo.Verb,
			"namespace", requestInfo.Namespace,
			"name", requestInfo.Name,
		)

		return
	}

	klog.ErrorSDepth(1, err, message)
}

// getRequestInfoFromContext从给定的上下文中提取请求信息
func getRequestInfoFromContext(ctx context.Context) *genericapirequest.RequestInfo {
	if reqInfo, found := genericapirequest.RequestInfoFrom(ctx); found {
		return reqInfo
	}
	return &genericapirequest.RequestInfo{}
}

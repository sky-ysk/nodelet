package value

import (
	"errors"
	"sync"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"k8s.io/component-base/metrics"
	"k8s.io/component-base/metrics/legacyregistry"
)

const (
	namespace = "apiserver"
	subsystem = "storage"
)

var (
	transformerLatencies = metrics.NewHistogramVec(
		&metrics.HistogramOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "transformation_duration_seconds",
			Help:      "Latencies in seconds of value transformation operations.",
			// In-process transformations (ex. AES CBC) complete on the order of 20 microseconds. However, when
			// external KMS is involved latencies may climb into hundreds of milliseconds.
			Buckets:        metrics.ExponentialBuckets(5e-6, 2, 25),
			StabilityLevel: metrics.ALPHA,
		},
		[]string{"transformation_type", "transformer_prefix"},
	)

	transformerOperationsTotal = metrics.NewCounterVec(
		&metrics.CounterOpts{
			Namespace:      namespace,
			Subsystem:      subsystem,
			Name:           "transformation_operations_total",
			Help:           "Total number of transformations. Successful transformation will have a status 'OK' and a varied status string when the transformation fails. This status and transformation_type fields may be used for alerting on encryption/decryption failure using transformation_type from_storage for decryption and to_storage for encryption",
			StabilityLevel: metrics.ALPHA,
		},
		[]string{"transformation_type", "transformer_prefix", "status"},
	)

	envelopeTransformationCacheMissTotal = metrics.NewCounter(
		&metrics.CounterOpts{
			Namespace:      namespace,
			Subsystem:      subsystem,
			Name:           "envelope_transformation_cache_misses_total",
			Help:           "Total number of cache misses while accessing key decryption key(KEK).",
			StabilityLevel: metrics.ALPHA,
		},
	)

	dataKeyGenerationLatencies = metrics.NewHistogram(
		&metrics.HistogramOpts{
			Namespace:      namespace,
			Subsystem:      subsystem,
			Name:           "data_key_generation_duration_seconds",
			Help:           "Latencies in seconds of data encryption key(DEK) generation operations.",
			Buckets:        metrics.ExponentialBuckets(5e-6, 2, 14),
			StabilityLevel: metrics.ALPHA,
		},
	)

	dataKeyGenerationFailuresTotal = metrics.NewCounter(
		&metrics.CounterOpts{
			Namespace:      namespace,
			Subsystem:      subsystem,
			Name:           "data_key_generation_failures_total",
			Help:           "Total number of failed data encryption key(DEK) generation operations.",
			StabilityLevel: metrics.ALPHA,
		},
	)
)

var registerMetrics sync.Once

func RegisterMetrics() {
	registerMetrics.Do(func() {
		legacyregistry.MustRegister(transformerLatencies)
		legacyregistry.MustRegister(transformerOperationsTotal)
		legacyregistry.MustRegister(envelopeTransformationCacheMissTotal)
		legacyregistry.MustRegister(dataKeyGenerationLatencies)
		legacyregistry.MustRegister(dataKeyGenerationFailuresTotal)
	})
}

// RecordTransformation 记录 TransformFromStorage 和 TransformToStorage 操作的延迟和数量。
func RecordTransformation(transformationType, transformerPrefix string, elapsed time.Duration, err error) {
	transformerOperationsTotal.WithLabelValues(transformationType, transformerPrefix, getErrorCode(err)).Inc()

	if err == nil {
		transformerLatencies.WithLabelValues(transformationType, transformerPrefix).Observe(elapsed.Seconds())
	}
}

// RecordDataKeyGeneration 记录数据加密密钥生成操作的延迟和数量。

func RecordDataKeyGeneration(start time.Time, err error) {
	if err != nil {
		dataKeyGenerationFailuresTotal.Inc()
		return
	}

	dataKeyGenerationLatencies.Observe(sinceInSeconds(start))
}

// sinceInSeconds 获取自指定开始时间以来的时间
func sinceInSeconds(start time.Time) float64 {
	return time.Since(start).Seconds()
}

type gRPCError interface {
	GRPCStatus() *status.Status
}

func getErrorCode(err error) string {
	if err == nil {
		return codes.OK.String()
	}
	var s gRPCError
	if errors.As(err, &s) {
		return s.GRPCStatus().Code().String()
	}
	return "unknown-non-grpc"
}

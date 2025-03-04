package storagebackend

import (
	"context"
	"fmt"
	"log"
	"os"
	"path"
	"sync"
	"time"

	grpcprom "github.com/grpc-ecosystem/go-grpc-prometheus"
	"go.etcd.io/etcd/client/pkg/v3/logutil"
	"go.etcd.io/etcd/client/pkg/v3/transport"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/time/rate"
	"google.golang.org/grpc"
	"hit.edu/framework/pkg/component-base/logs"
	//utilruntime "k8s.io/apimachinery/pkg/util/runtime"
)

const (
	// 保持连接的时间间隔和超时时间
	keepaliveTime    = 30 * time.Second
	keepaliveTimeout = 10 * time.Second
	// 建立连接的超时时间
	dialTimeout = 20 * time.Second
	// 数据库指标检测抖动时间，避免过于频繁的指标查询
	dbMetricsMonitorJitter = 0.5
)

var etcd3ClientLogger *zap.Logger

// 创建Zap日志记录器，并设置日志级别，默认使用InfoLevel
func init() {
	dbMetricsMonitors = make(map[string]struct{}) //用于完成指标监控
	l, err := logutil.CreateDefaultZapLogger(etcdClientDebugLevel())
	if err != nil {
		l = zap.NewNop()
	}
	etcd3ClientLogger = l.Named("etcd-client")
}
func etcdClientDebugLevel() zapcore.Level {
	envLevel := os.Getenv("ETCD_CLIENT_DEBUG")
	if envLevel == "" || envLevel == "true" {
		return zapcore.InfoLevel
	}
	var l zapcore.Level
	//无法识别提供的日志级别
	if err := l.Set(envLevel); err == nil {
		log.Printf("Deprecated env ETCD_CLIENT_DEBUG value. Using default level: 'info'")
		return zapcore.InfoLevel
	}
	return l
}

// 健康检查
func newETCD3HealthCheck(c Config, stopCh <-chan struct{}) (func() error, error) {
	timeout := DefaultHealthcheckTimeout
	if c.HealthcheckTimeout != time.Duration(0) {
		timeout = c.HealthcheckTimeout
	}
	return newETCD3Check(c, timeout, stopCh)
}

// 就绪检查
func newETCD3ReadyCheck(c Config, stopCh <-chan struct{}) (func() error, error) {
	timeout := DefaultReadinessTimeout
	if c.ReadycheckTimeout != time.Duration(0) {
		timeout = c.ReadycheckTimeout
	}
	return newETCD3Check(c, timeout, stopCh)
}

// 实现原子性存储错误信息
type atomicLastError struct {
	mu        sync.RWMutex
	err       error
	timestamp time.Time
}

// 操作错误信息
func (a *atomicLastError) Store(err error, t time.Time) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.timestamp.IsZero() || a.timestamp.Before(t) {
		a.err = err
		a.timestamp = t
	}
}
func (a *atomicLastError) Load() error {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.err
}

// 定期检查
func PollUntil(interval time.Duration, condition func() (bool, error), stopCh <-chan struct{}) error {
	for {
		select {
		case <-stopCh:
			return fmt.Errorf("polling stopped")
		default:
			done, err := condition()
			if err != nil {
				return err
			}
			if done {
				return nil
			}
			time.Sleep(interval)
		}
	}
}

func newETCD3Check(c Config, timeout time.Duration, stopCh <-chan struct{}) (func() error, error) {

	lock := sync.RWMutex{}
	var prober *etcd3ProberMonitor
	clientErr := fmt.Errorf("etcd client connection not yet established")

	//每秒执行轮询
	go PollUntil(time.Second, func() (bool, error) {
		lock.Lock()
		defer lock.Unlock()
		//尝试创建客户端
		newProber, err := newETCD3ProberMonitor(c)
		select {
		case <-stopCh:
			if err == nil {
				newProber.Close()
			}
			return true, nil
		default:
		}
		if err != nil {
			clientErr = err
			return false, nil
		}
		prober = newProber
		clientErr = nil
		return true, nil
	}, stopCh)

	// 等待关闭信号并关闭客户端
	go func() {
		//捕获并处理协程的恐慌
		//defer utilruntime.HandleCrash()
		<-stopCh

		lock.Lock()
		defer lock.Unlock()
		if prober != nil {
			prober.Close()
			clientErr = fmt.Errorf("server is shutting down")
		}
	}()

	// 限流器，限制健康检查频率和最多发送数量
	limiter := rate.NewLimiter(rate.Every(timeout/2), 1)
	// 初始状态设置
	lastError := &atomicLastError{err: fmt.Errorf("etcd client connection not yet established")}
	// 返回用于健康检查的函数
	return func() error {
		lock.RLock()
		defer lock.RUnlock()

		if clientErr != nil {
			return clientErr
		}
		if limiter.Allow() == false {
			return lastError.Load()
		}
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		now := time.Now()
		err := prober.Probe(ctx)
		lastError.Store(err, now)
		return err
	}, nil
}

func newETCD3ProberMonitor(c Config) (*etcd3ProberMonitor, error) {
	client, err := NewETCD3Client(c.Transport)
	if err != nil {
		return nil, err
	}
	return &etcd3ProberMonitor{
		client:    client.Client,
		prefix:    c.Prefix,
		endpoints: c.Transport.ServerList,
	}, nil
}

type etcd3ProberMonitor struct {
	prefix    string
	endpoints []string

	mux    sync.RWMutex
	client *clientv3.Client
	closed bool
}

func (t *etcd3ProberMonitor) Close() error {
	t.mux.Lock()
	defer t.mux.Unlock()
	if !t.closed {
		t.closed = true
		return t.client.Close()
	}
	return fmt.Errorf("closed")
}

func (t *etcd3ProberMonitor) Probe(ctx context.Context) error {
	t.mux.RLock()
	defer t.mux.RUnlock()
	if t.closed {
		return fmt.Errorf("closed")
	}
	_, err := t.client.Get(ctx, path.Join("/", t.prefix, "health"))
	if err != nil {
		return fmt.Errorf("error getting data from etcd: %w", err)
	}
	return nil
}

//TODO: metrics 指标监测
// func (t *etcd3ProberMonitor) Monitor(ctx context.Context) (metrics.StorageMetrics, error) {
// 	t.mux.RLock()
// 	defer t.mux.RUnlock()
// 	if t.closed {
// 		return metrics.StorageMetrics{}, fmt.Errorf("closed")
// 	}
// 	status, err := t.client.Status(ctx, t.endpoints[rand.Int()%len(t.endpoints)])
// 	if err != nil {
// 		return metrics.StorageMetrics{}, err
// 	}
// 	return metrics.StorageMetrics{
// 		Size: status.DbSize,
// 	}, nil
// }

func NewETCD3Client(c TransportConfig) (*Client, error) {
	tlsInfo := transport.TLSInfo{
		CertFile:      c.CertFile,
		KeyFile:       c.KeyFile,
		TrustedCAFile: c.TrustedCAFile,
	}
	tlsConfig, err := tlsInfo.ClientConfig()
	if err != nil {
		return nil, err
	}
	// 使用非安全连接
	if len(c.CertFile) == 0 && len(c.KeyFile) == 0 && len(c.TrustedCAFile) == 0 {
		tlsConfig = nil
	}
	dialOptions := []grpc.DialOption{
		grpc.WithBlock(),
		grpc.WithChainUnaryInterceptor(grpcprom.UnaryClientInterceptor),
		grpc.WithChainStreamInterceptor(grpcprom.StreamClientInterceptor),
	}
	cfg := clientv3.Config{
		DialTimeout:          dialTimeout,
		DialKeepAliveTime:    keepaliveTime,
		DialKeepAliveTimeout: keepaliveTimeout,
		//TODO:关注拨号的参数设置
		DialOptions: dialOptions,
		Endpoints:   c.ServerList,
		TLS:         tlsConfig,
		Logger:      etcd3ClientLogger,
	}
	logs.Info("etcd3 client create successfully")
	return New(cfg)
}

func New(cfg clientv3.Config) (*Client, error) {
	c, err := clientv3.New(cfg)
	if err != nil {
		return nil, err
	}
	kc := &Client{
		Client: c,
	}
	kc.ViewsOptions = kc
	return kc, nil
}

// TODO: etcd compact DBMonitor
type runningCompactor struct {
	interval time.Duration
	cancel   context.CancelFunc
	client   *clientv3.Client
	refs     int
}

var (
	// 压缩器实例
	compactorsMu sync.Mutex
	compactors   = map[string]*runningCompactor{}
	// 数据库监控
	dbMetricsMonitorsMu sync.Mutex
	dbMetricsMonitors   map[string]struct{}
)

// TODO: 修改补充newStorage
